// Command example exercises the WayaPay Go SDK against staging.
//
// Run it with your own credentials:
//
//	WAYA_MERCHANT_ID=MER_... WAYA_SECRET_KEY=WAYASECK_TEST_... go run ./examples
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	wayapay "github.com/wayapaychat/wayapay-go"
)

func main() {
	client := wayapay.New(
		os.Getenv("WAYA_MERCHANT_ID"),
		os.Getenv("WAYA_SECRET_KEY"),
		wayapay.WithBaseURL(wayapay.BaseURLStaging),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Bank list.
	banks, err := client.Banks.List(ctx)
	if err != nil {
		log.Fatalf("bank list: %v", err)
	}
	fmt.Printf("got %d banks; first: %s (%s)\n", len(banks), banks[0].Name, banks[0].Code)

	// 2. Verify a destination before paying it.
	acct, err := client.Accounts.Verify(ctx, wayapay.VerifyAccountRequest{
		AccountNumber: "0123456789",
		BankCode:      "044",
		EnquiryType:   wayapay.EnquiryTypeOthers,
	})
	if err != nil {
		log.Fatalf("verify account: %v", err)
	}
	fmt.Printf("resolved account: %s\n", acct.AccountName)

	// 3. Create a payment link.
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

	// 4. Initiate a payout.
	payout, err := client.Payout.Initiate(ctx, wayapay.PayoutRequest{
		Amount:        25000,
		Currency:      "NGN",
		AccountNumber: acct.AccountNumber,
		BankCode:      "058",
		AccountName:   acct.AccountName,
		Reference:     "PAYOUT-20260523-001",
		Narration:     "Salary payment May 2026",
	})
	if err != nil {
		log.Fatalf("payout: %v", err)
	}
	fmt.Printf("payout %s status=%s\n", payout.PayoutReference, payout.Status)

	// 5. Verify the payout settled, branching on the typed API error.
	txn, err := client.Transactions.Verify(ctx, payout.PayoutReference)
	if err != nil {
		if ae, ok := wayapay.AsAPIError(err); ok {
			log.Fatalf("verify failed with code %s: %s", ae.Code, ae.Message)
		}
		log.Fatalf("verify: %v", err)
	}
	fmt.Printf("transaction status: %s\n", txn.Status)

	// 6. Pull a page of history for reconciliation.
	from := time.Now().AddDate(0, 0, -30)
	hist, err := client.Transactions.History(ctx, wayapay.HistoryParams{
		Page:   0,
		Size:   20,
		Status: wayapay.StatusSuccess,
		From:   &from,
	})
	if err != nil {
		log.Fatalf("history: %v", err)
	}
	fmt.Printf("history: %d of %d total across %d pages\n",
		len(hist.Items), hist.TotalElements, hist.TotalPages)
}
