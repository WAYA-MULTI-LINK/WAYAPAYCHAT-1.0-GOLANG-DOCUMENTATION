# wayapay-go

A Go client for the **WayaPay Merchant API v2** covering collect, payout, accounts, and identity. One client, six services, a single transport that handles auth headers and the shared response envelope so you never parse `success`/`code` by hand.

## Structure

```
wayapay-go/
├── go.mod                # module github.com/wayapaychat/wayapay-go
├── doc.go                # package documentation
├── wayapay.go            # Client, functional options, New()
├── transport.go          # request building, auth headers, envelope decoding
├── errors.go             # APIError + AsAPIError helper
├── banks.go              # Banks.List
├── accounts.go           # Accounts.Verify, Accounts.CreateDynamicAccount
├── identity.go           # Identity.VerifyBVN
├── payout.go             # Payout.Initiate
├── collect.go            # Collect.Initiate
├── transactions.go       # Transactions.Verify, Transactions.History
└── examples/
    └── main.go           # runnable end to end walkthrough
```

## Install

```bash
go get github.com/wayapaychat/wayapay-go
```

Rename the module path in `go.mod` if you host this somewhere else.

## Quickstart

```go
import wayapay "github.com/wayapaychat/wayapay-go"

client := wayapay.New(
    "MER_lCquc1779095889226CVfl7",
    "WAYASECK_TEST_0x3a93476a20d347d6847b62665e9ecb4b",
    wayapay.WithBaseURL(wayapay.BaseURLStaging), // drop this for production
)

banks, err := client.Banks.List(context.Background())
```

Every method takes a `context.Context` first and returns a typed struct plus an error. The client is safe to share across goroutines, so build it once at startup.

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

Anything other than code `00` (or a non envelope body) comes back as `*APIError`. Branch on the code:

```go
acct, err := client.Accounts.Verify(ctx, req)
if err != nil {
    if ae, ok := wayapay.AsAPIError(err); ok {
        switch ae.Code {
        case "07":
            // invalid account number, surface a friendly message
        default:
            log.Printf("waya error %s: %s", ae.Code, ae.Message)
        }
    }
    return err
}
```

## Configuration options

- `WithBaseURL(url)` switch between `BaseURLStaging` and `BaseURLProduction`, or point at a mock.
- `WithHTTPClient(h)` bring your own `*http.Client` for custom timeouts, retries, proxies, or tracing.
- `WithUserAgent(ua)` set a custom User-Agent.

## Design notes worth knowing

**This is a server side SDK.** Your `WAYASECK_...` key is a wallet with the PIN written on the back. Never bundle it in a mobile app or browser. Keep payout authorization behind your own backend, full stop.

**Money is modelled as `float64`.** The API declares amounts as `number` in the major unit, so this maps cleanly for whole naira values. If you handle sub unit precision and want to kill float rounding risk entirely, swap those fields for a decimal type (`shopspring/decimal`) or integer minor units in your own layer. Flagging it up front rather than letting it bite you at reconciliation time.

**References are your idempotency key.** v2 dropped the separate `idempotencyKey`. The unique `Reference` on payout, `ReferenceID` on dynamic account, and `PaymentLinkReference` on collect are how retries map to the original record. Generate a fresh unique one per logical operation.

**Collect has prerequisites.** `Collect.Initiate` fails unless you have whitelisted server IPs and configured payment preferences on the dashboard.

**Polling, not assuming.** A payout returning `PROCESSING` is accepted, not settled. Poll `Transactions.Verify` with the `PayoutReference` until you see `SUCCESS`.

## Not yet included

Webhook signature verification. The v2 spec mentions webhooks but does not publish a signing scheme, so there is nothing to verify against yet. Drop it in here once that contract is documented.
