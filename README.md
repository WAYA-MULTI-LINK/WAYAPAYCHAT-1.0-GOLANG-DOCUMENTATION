# WayaQuick Go SDK

Go client for the **WayaQuick Merchant API v2**. Collect payments, send payouts, verify bank accounts, and run BVN identity checks — all in Nigeria.

One client, four services, a single transport that handles auth headers and the shared response envelope so you never parse `success`/`code` by hand. No dependencies outside the standard library. **Server-side only** — your secret key must never leave your server.

## Install

```bash
go get github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick
```

Rename the module path in `go.mod` if you host this somewhere else.

## Quickstart

```go
import wayaquick "github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick"

client := wayaquick.New(
    "MER_...",                 // from the dashboard
    "WAYASECK_TEST_...",       // swap for WAYASECK_... on live
)
```

The client targets the production base URL. Every method takes a `context.Context` first and returns a typed struct plus an error. The client is safe to share across goroutines, so build it once at startup.

## List banks

```go
banks, err := client.Payout.ListBanks(ctx)
// []wayaquick.Bank — each has .Code and .Name
```

## Verify an account

Always verify before sending a payout — confirms the account exists and returns the registered name.

```go
acct, err := client.Payout.VerifyAccount(ctx, wayaquick.VerifyAccountRequest{
    AccountNumber: "0123456789",
    BankCode:      "044",                       // required when EnquiryType is OTHERS
    EnquiryType:   wayaquick.EnquiryTypeOthers,   // EnquiryTypeWayaBank for intra-bank
})
fmt.Println(acct.AccountName) // "JOHN DOE"
```

## Initiate a payout

```go
payout, err := client.Payout.Initiate(ctx, wayaquick.PayoutRequest{
    Amount:        25000,
    Currency:      "NGN",
    AccountNumber: acct.AccountNumber,
    BankCode:      "058",
    AccountName:   acct.AccountName,
    Reference:     wayaquick.GenerateReference("PAYOUT"),
    Narration:     "April salary",
})
// payout.Status == wayaquick.StatusProcessing means accepted, not yet settled

// Reconcile by the reference you sent at initiation:
st, err := client.Payout.Status(ctx, payout.MerchantReference)
// branch on st.ParsedStatus().Outcome() / .IsTerminal()
```

`GenerateReference` produces a timestamped, collision-resistant key (`PAYOUT-1748160000000-A1B2C3D4`). Generate a fresh one per operation and reuse the same one on retries.

## Collect a payment

```go
link, err := client.Collect.Initiate(ctx, wayaquick.CollectRequest{
    PaymentLinkType: wayaquick.PaymentLinkTypeOneTime,
    PaymentLinkName: "Order #1234",
    Description:     "Order #1234 - 2 items",
    PayableAmount:   1500,
    Currency:        "NGN",
    RedirectLink:    "https://merchant.example.com/callback",
})
// Send the customer to link.ShortURL (or link.CustomerPaymentLink) to pay.
// Confirm the result on your server before fulfilling the order.

// Reconcile a deposit by its refNo (the gateway transactionId / webhook OrderId):
cs, err := client.Collect.Status(ctx, refNo)
// branch on cs.ParsedStatus().Outcome() / .IsTerminal()
```

`Collect.Initiate` fails unless you have whitelisted your server IPs and configured payment preferences on the dashboard.

## BVN identity check

```go
rec, err := client.Identity.VerifyBVN(ctx, "22500809037") // exactly 11 digits — validated locally
fmt.Printf("%s %s\n", rec.FirstName, rec.LastName)
if rec.IsWatchListed() {
    // handle a flagged BVN
}
```

BVN data is sensitive personal information. Store, transmit, and log it only as your data-protection obligations allow.

A payout returning `PROCESSING` is accepted, not settled. Poll `Payout.Status` with the reference until it reaches a terminal status.

## The services

| Service | Method | Endpoint |
|---|---|---|
| `client.Payout` | `ListBanks` | `GET /get-bank-list` |
| `client.Payout` | `VerifyAccount` | `POST /verify-account` |
| `client.Payout` | `Initiate` | `POST /payment-payout/initiate` |
| `client.Payout` | `Status` | `GET /payment-payout/status/{reference}` |
| `client.Collect` | `Initiate` | `POST /payment-collect/initiate` |
| `client.Collect` | `Status` | `GET /payment-collect/status/{refNo}` |
| `client.Identity` | `VerifyBVN` | `POST /identity-verification/bvn` |
| `client.Webhooks` | `ConstructEvent` / `VerifySignature` | — (verifies inbound webhooks) |

## Error handling

Anything other than code `00` (or a non-envelope body) comes back as `*APIError`. Branch on the code:

```go
acct, err := client.Payout.VerifyAccount(ctx, req)
if err != nil {
    if ae, ok := wayaquick.AsAPIError(err); ok {
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
client := wayaquick.New(merchant, secret,
    wayaquick.WithMaxRetries(3),
    wayaquick.WithHTTPClient(&http.Client{Timeout: 15 * time.Second}),
)
```

## Design notes worth knowing

**This is a server-side SDK.** Your `WAYASECK_...` key is a wallet with the PIN written on the back. Never bundle it in a mobile app or browser. Keep payout authorization behind your own backend, full stop.

**Money is modelled as `float64`.** The API declares amounts as `number` in the major unit, so this maps cleanly for whole-naira values. If you handle sub-unit precision and want to kill float rounding risk entirely, swap those fields for a decimal type (`shopspring/decimal`) or integer minor units in your own layer.

**References are your idempotency key.** v2 dropped the separate `idempotencyKey`. The unique `Reference` on payout and `PaymentLinkReference` on collect are how retries map to the original record. `GenerateReference` gives you one per operation.

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
