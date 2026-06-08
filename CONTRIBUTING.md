# Contributing

## Requirements

- Go 1.21 or newer (`go version`)

## Project layout

```
src/
  wayapay/                  # The library — package wayapay, this is what callers import
    doc.go                  # Package documentation
    wayapay.go              # Client, functional options, New(), GenerateReference()
    transport.go            # Request building, auth headers, envelope decoding, GET retry
    errors.go               # APIError + AsAPIError helper
    banks.go                # Banks.List
    accounts.go             # Accounts.Verify, Accounts.CreateDynamicAccount
    identity.go             # Identity.VerifyBVN
    payout.go               # Payout.Initiate
    collect.go              # Collect.Initiate
    transactions.go         # Transactions.Verify, Transactions.History

tests/                      # External (black-box) test package: wayapay_test
    helpers_test.go         # RoundTripper stubs + capturing transport, client builder
    factories_test.go       # Valid request builders shared across tests
    client_test.go          # Construction, headers, envelope, errors, retry, GenerateReference
    banks_test.go           # ... one file per service
    accounts_test.go
    identity_test.go
    payout_test.go
    collect_test.go
    transactions_test.go
    live_test.go            # //go:build live — hits the real API, excluded by default

samples/
    main.go                 # Runnable end-to-end demo — kept in sync with the API
```

The package lives under `src/wayapay`, so the import path is
`github.com/wayapaychat/wayapay-go/src/wayapay`.

## Build

```bash
go build ./...
go vet ./...
gofmt -l src tests samples   # prints files that need formatting; should be empty
```

## Run unit tests

```bash
# All unit tests (no network, no credentials)
go test ./tests/...

# Verbose
go test -v ./tests/...

# A single test
go test ./tests/... -run TestPayout_DecodesProcessingResult

# With the race detector
go test -race ./tests/...
```

Unit tests run entirely against stubbed `http.RoundTripper` implementations
injected via `WithHTTPClient`. No credentials, no network.

## Run live integration tests

Live tests are guarded by the `live` build tag and are excluded from the default
run. They call the real WayaPay API, so you need valid credentials.

```bash
export WAYA_MERCHANT_ID=MER_...
export WAYA_SECRET_KEY=WAYASECK_TEST_...
# live tests run against the production API; use test credentials

go test -tags live ./tests/... -run TestLive -v
```

Live tests are intentionally not run in CI to avoid flakiness from network
conditions or credential availability.

## Run the sample

```bash
WAYA_MERCHANT_ID=MER_... WAYA_SECRET_KEY=WAYASECK_TEST_... go run ./samples
```

## Adding a new feature

1. Add request/response types and the method to the relevant service file under `src/wayapay/`.
2. Validate required fields at the boundary and return a plain error before the network call.
3. Add unit tests covering the happy path, error path, correct HTTP method/path, and request body shape.
4. Update `samples/main.go` if the feature is user-facing.
5. Update `CHANGELOG.md` under the relevant version.

## Versioning

This project follows [Semantic Versioning](https://semver.org). Releases are cut
by pushing a `v*.*.*` tag.

## Code style

- Run `gofmt` (or `goimports`) before committing — CI rejects unformatted code.
- One service per file; the service struct holds only `*Client`.
- Validate at the boundary with plain `errors.New("wayapay: ...")` messages, prefixed `wayapay:`.
- Return `*APIError` for anything the server rejects, so callers can branch on `Code`.
- No comments explaining *what* the code does — only add one when the *why* is non-obvious.
