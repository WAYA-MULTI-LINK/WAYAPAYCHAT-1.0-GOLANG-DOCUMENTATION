# wayapay-go

Go client for the **WayaPay Merchant API v2**. Collect payments, send payouts, mint virtual accounts, verify bank accounts, run BVN identity checks, and reconcile transactions — all in Nigeria.

One client, six services, a single transport that handles auth headers and the shared response envelope so you never parse `success`/`code` by hand. No dependencies outside the standard library. **Server-side only** — your secret key must never leave your server.

## Install

```bash
go get github.com/wayapaychat/wayapay-go/src/wayapay
```

Rename the module path in `go.mod` if you host this somewhere else.

## Quickstart

```go
import wayapay "github.com/wayapaychat/wayapay-go/src/wayapay"

client := wayapay.New(
    "MER_...",                 // from the dashboard
    "WAYASECK_TEST_...",       // swap for WAYASECK_... on live
)
```

The client targets the production base URL. Every method takes a `context.Context` first and returns a typed struct plus an error. The client is safe to share across goroutines, so build it once at startup.

## List banks

```go
banks, err := client.Banks.List(ctx)
// []wayapay.Bank — each has .Code and .Name
```

## Verify an account

Always verify before sending a payout — confirms the account exists and returns the registered name.

```go
acct, err := client.Accounts.Verify(ctx, wayapay.VerifyAccountRequest{
    AccountNumber: "0123456789",
    BankCode:      "044",                       // required when EnquiryType is OTHERS
    EnquiryType:   wayapay.EnquiryTypeOthers,   // EnquiryTypeWayaBank for intra-bank
})
fmt.Println(acct.AccountName) // "JOHN DOE"
```

## Initiate a payout

```go
payout, err := client.Payout.Initiate(ctx, wayapay.PayoutRequest{
    Amount:        25000,
    Currency:      "NGN",
    AccountNumber: acct.AccountNumber,
    BankCode:      "058",
    AccountName:   acct.AccountName,
    Reference:     wayapay.GenerateReference("PAYOUT"),
    Narration:     "April salary",
})
// payout.Status == wayapay.StatusProcessing means accepted, not yet settled
```

`GenerateReference` produces a timestamped, collision-resistant key (`PAYOUT-1748160000000-A1B2C3D4`). Generate a fresh one per operation and reuse the same one on retries.

## Collect a payment

```go
link, err := client.Collect.Initiate(ctx, wayapay.CollectRequest{
    PaymentLinkType: wayapay.PaymentLinkTypeOneTime,
    PaymentLinkName: "Order #1234",
    Description:     "Order #1234 - 2 items",
    PayableAmount:   1500,
    Currency:        "NGN",
    RedirectLink:    "https://merchant.example.com/callback",
})
// Send the customer to link.ShortURL (or link.CustomerPaymentLink) to pay.
// Confirm the result on your server before fulfilling the order.
```

`Collect.Initiate` fails unless you have whitelisted your server IPs and configured payment preferences on the dashboard.

## Mint a virtual account

```go
va, err := client.Accounts.CreateDynamicAccount(ctx, wayapay.DynamicAccountRequest{
    AccountName: "ACME LTD",
    CustomerID:  "CUST-1",
    ReferenceID: wayapay.GenerateReference("VA"),
    Purpose:     "Order #1234",
    Mode:        wayapay.AccountModeOneTime,
})
// Hand va.VirtualAccountNumber to the customer; watch va.CanReceivePayments.
```

## BVN identity check

```go
rec, err := client.Identity.VerifyBVN(ctx, "22500809037") // exactly 11 digits — validated locally
fmt.Printf("%s %s\n", rec.FirstName, rec.LastName)
if rec.IsWatchListed() {
    // handle a flagged BVN
}
```

BVN data is sensitive personal information. Store, transmit, and log it only as your data-protection obligations allow.

## Verify a transaction / pull history

```go
txn, err := client.Transactions.Verify(ctx, payout.PayoutReference)
// txn.Status == wayapay.StatusSuccess means settled

hist, err := client.Transactions.History(ctx, wayapay.HistoryParams{
    Page: 0, Size: 20, Status: wayapay.StatusSuccess,
})
```

A payout returning `PROCESSING` is accepted, not settled. Poll `Transactions.Verify` with the `PayoutReference` until you see `SUCCESS`.

## The services

| Service | Method | Endpoint |
|---|---|---|
| `client.Banks` | `List` | `GET /account-enquiry/get-bank-list` |
| `client.Accounts` | `Verify` | `POST /account-enquiry/verify-account` |
| `client.Accounts` | `CreateDynamicAccount` | `POST /account-enquiry/create-dynamic-account` |
| `client.Identity` | `VerifyBVN` | `POST /identity-verification/bvn` |
| `client.Payout` | `Initiate` | `POST /payment-payout/initiate` |
| `client.Collect` | `Initiate` | `POST /payment-collect/initiate` |
| `client.Transactions` | `Verify` | `GET /transaction/verify` |
| `client.Transactions` | `History` | `GET /transaction/history` |

## Error handling

Anything other than code `00` (or a non-envelope body) comes back as `*APIError`. Branch on the code:

```go
acct, err := client.Accounts.Verify(ctx, req)
if err != nil {
    if ae, ok := wayapay.AsAPIError(err); ok {
        switch ae.Code {
        case "07":
            // invalid account number — surface a friendly message
        default:
            log.Printf("waya error %s: %s", ae.Code, ae.Message)
        }
    }
    return err
}
```

Input validation errors (missing required fields, malformed BVN, missing `BankCode`) are returned as plain errors **before any network call is made**.

## Configuration options

- `WithBaseURL(url)` — override `BaseURLProduction` (the default), e.g. to point at a mock.
- `WithHTTPClient(h)` — bring your own `*http.Client` for custom timeouts, proxies, transports, or tracing.
- `WithUserAgent(ua)` — set a custom User-Agent.
- `WithMaxRetries(n)` — retries on GET only (default 2). Timeouts, network errors, 429, and 5xx. **Writes never auto-retry.**

```go
client := wayapay.New(merchant, secret,
    wayapay.WithMaxRetries(3),
    wayapay.WithHTTPClient(&http.Client{Timeout: 15 * time.Second}),
)
```

## Design notes worth knowing

**This is a server-side SDK.** Your `WAYASECK_...` key is a wallet with the PIN written on the back. Never bundle it in a mobile app or browser. Keep payout authorization behind your own backend, full stop.

**Money is modelled as `float64`.** The API declares amounts as `number` in the major unit, so this maps cleanly for whole-naira values. If you handle sub-unit precision and want to kill float rounding risk entirely, swap those fields for a decimal type (`shopspring/decimal`) or integer minor units in your own layer.

**References are your idempotency key.** v2 dropped the separate `idempotencyKey`. The unique `Reference` on payout, `ReferenceID` on dynamic account, and `PaymentLinkReference` on collect are how retries map to the original record. `GenerateReference` gives you one per operation.

## Full example

See [samples/main.go](samples/main.go) for a runnable end-to-end demo covering all the operations.

```bash
WAYA_MERCHANT_ID=MER_... WAYA_SECRET_KEY=WAYASECK_TEST_... go run ./samples
```

## Going live

On the merchant dashboard: finish KYC, grab your Merchant ID, generate your secret key under **Settings → API Keys and Webhooks**, and whitelist your server IPs. Swap `WAYASECK_TEST_...` for `WAYASECK_...` — the rest of your code stays the same.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT
