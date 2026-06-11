// Command sample exercises the WayaPay Go SDK end to end against production.
//
// Run it with your own credentials:
//
//	WAYA_MERCHANT_ID=MER_... WAYA_SECRET_KEY=WAYASECK_TEST_... go run ./samples
package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	wayapay "github.com/wayapaychat/wayapay-go/src/wayapay"
)

func main() {
	merchant := os.Getenv("WAYA_MERCHANT_ID")
	secret := os.Getenv("WAYA_SECRET_KEY")
	if merchant == "" || secret == "" {
		log.Fatal("set WAYA_MERCHANT_ID and WAYA_SECRET_KEY")
	}

	client := wayapay.New(merchant, secret)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Bank list (GET — auto retried on transient failures).
	banks, err := client.Payout.ListBanks(ctx)
	if err != nil {
		log.Fatalf("bank list: %v", err)
	}
	fmt.Printf("got %d banks; first: %s (%s)\n", len(banks), banks[0].Name, banks[0].Code)

	// 2. Verify a destination before paying it.
	acct, err := client.Payout.VerifyAccount(ctx, wayapay.VerifyAccountRequest{
		AccountNumber: "0123456789",
		BankCode:      "044",
		EnquiryType:   wayapay.EnquiryTypeOthers,
	})
	if err != nil {
		log.Fatalf("verify account: %v", err)
	}
	fmt.Printf("resolved account: %s\n", acct.AccountName)

	// 3. Create a payment link to collect from a customer.
	link, err := client.Collect.Initiate(ctx, wayapay.CollectRequest{
		PaymentLinkType: wayapay.PaymentLinkTypeOneTime,
		PaymentLinkName: "Order #1234",
		Description:     "Order #1234 - 2 items",
		PayableAmount:   1500,
		Currency:        "NGN",
		RedirectLink:    "https://merchant.example.com/callback",
		OtherDetailsJSON: map[string]any{
			"orderId": "1234",
			"sku":     "SKU-001",
		},
	})
	if err != nil {
		log.Fatalf("collect: %v", err)
	}
	fmt.Printf("pay here: %s\n", link.ShortURL)

	// 3b. Check the collection status — the pull/safety-net path alongside the
	// deposit webhook. Reconcile by the link reference.
	collStatus, err := client.Collect.Status(ctx, link.PaymentLinkReference)
	if err != nil {
		log.Fatalf("collect status: %v", err)
	}
	fmt.Printf("collection %s -> %s\n", collStatus.Status, collStatus.ParsedStatus().Outcome())
	if collStatus.ParsedStatus() == wayapay.CollectStatusSuccessful {
		fmt.Printf("funds confirmed; fulfil order using refNo %s\n", collStatus.RefNo)
	}

	// 4. Initiate a payout with a fresh, collision-resistant reference.
	payout, err := client.Payout.Initiate(ctx, wayapay.PayoutRequest{
		Amount:        25000,
		Currency:      "NGN",
		AccountNumber: acct.AccountNumber,
		BankCode:      "058",
		AccountName:   acct.AccountName,
		Reference:     wayapay.GenerateReference("PAYOUT"),
		Narration:     "Salary payment",
	})
	if err != nil {
		log.Fatalf("payout: %v", err)
	}
	fmt.Printf("payout %s status=%s\n", payout.PayoutReference, payout.Status)

	// 4b. Check the payout status by the reference you sent at initiation, then
	// branch on the parsed outcome.
	ref := payout.MerchantReference
	if ref == "" {
		ref = payout.PayoutReference
	}
	poStatus, err := client.Payout.Status(ctx, ref)
	if err != nil {
		log.Fatalf("payout status: %v", err)
	}
	switch poStatus.ParsedStatus().Outcome() {
	case wayapay.PayoutSucceeded:
		fmt.Println("payout delivered")
	case wayapay.PayoutReversed:
		fmt.Println("payout reversed — wallet re-credited")
	default:
		fmt.Println("payout still reconciling — check again later")
	}

	// 5. Verify a webhook (offline demo). In production WayaPay POSTs this to
	// your HTTPS endpoint; here we sign a sample body locally to show the flow.
	const webhookSecret = "WAYASECK_TEST_demo_webhook_secret"
	const rawBody = `{"OrderId":"1779662251460508970","Amount":1500.00,"Fee":15.00,` +
		`"Currency":"NGN","Status":"SUCCESSFUL","productName":"CARD",` +
		`"customer":{"email":"john@example.com"},"merchantId":"MER_xyz","recurrentPayment":false}`

	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write([]byte(ts + "." + rawBody))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	evt, err := wayapay.ConstructEvent(rawBody, ts, sig, webhookSecret, wayapay.DefaultWebhookTolerance)
	if err != nil {
		log.Fatalf("webhook: %v", err)
	}
	fmt.Printf("webhook verified: %s — %s (%.2f %s)\n", evt.OrderID, evt.Status, evt.Amount, evt.Currency)
	if evt.ShouldFulfil() {
		fmt.Printf("fulfil order — idempotency key %s\n", evt.OrderID)
	}
}
