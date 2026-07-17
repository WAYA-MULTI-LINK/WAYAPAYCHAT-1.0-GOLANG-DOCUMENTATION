# Contributing

## Requirements

- Go 1.21 or newer (`go version`)

## Project layout

```
src/
  wayaquick/                  # The library — package wayaquick, this is what callers import
    doc.go                  # Package documentation
    wayaquick.go              # Client, functional options, New(), GenerateReference()
    transport.go            # Request building, auth headers, envelope decoding, GET retry
    errors.go               # APIError + AsAPIError helper
    payout.go               # Payout.ListBanks, VerifyAccount, Initiate, Status
    collect.go              # Collect.Initiate, Collect.Status
    identity.go             # Identity.VerifyBVN
    webhook.go              # Webhooks.ConstructEvent, VerifySignature

tests/                      # External (black-box) test package: wayaquick_test
    helpers_test.go         # RoundTripper stubs + capturing transport, client builder
    factories_test.go       # Valid request builders shared across tests
    client_test.go          # Construction, headers, envelope, errors, retry, GenerateReference
    payout_test.go          # ... one file per service
    identity_test.go
    collect_test.go
    webhook_test.go
    live_test.go            # //go:build live — hits the real API, excluded by default

samples/
    main.go                 # Runnable end-to-end demo — kept in sync with the API
```

The package lives under `src/wayaquick`, so the import path is
`github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick`.

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
run. They call the real WayaQuick API, so you need valid credentials.

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

1. Add request/response types and the method to the relevant service file under `src/wayaquick/`.
2. Validate required fields at the boundary and return a plain error before the network call.
3. Add unit tests covering the happy path, error path, correct HTTP method/path, and request body shape.
4. Update `samples/main.go` if the feature is user-facing.
5. Update `CHANGELOG.md` under the relevant version.

## Versioning

This project follows [Semantic Versioning](https://semver.org). Releases are cut
by pushing a `v*.*.*` tag.

## Releasing & publishing

Unlike the Java (Maven Central) and .NET (nuget.org) SDKs, Go has no package
registry to push to — consumers fetch the module straight from this GitHub
repository, and [pkg.go.dev](https://pkg.go.dev) indexes it automatically the
first time someone requests it. Publishing a release is just tagging.

### Cutting a release

1. Add an entry to `CHANGELOG.md` (there is no version field to bump — the git
   tag *is* the version).
2. Commit everything, then tag and push:

   ```bash
   git tag v<x.y.z>
   git push origin main v<x.y.z>
   ```

3. Verify consumers can fetch it:

   ```bash
   go get github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick@v<x.y.z>
   ```

   and import it as:

   ```go
   import wayaquick "github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick"
   ```

> **Caution — published versions are permanent.** The Go module proxy
> (proxy.golang.org) caches a tagged version forever the first time anyone
> fetches it; deleting or re-pointing a tag on GitHub does **not** remove it
> from the proxy. Never reuse a version number — always cut a new patch tag.
> The repository must stay **public**, and the `module` line in `go.mod` must
> always match this repo's URL exactly, or `go get` refuses to fetch it. For
> the same reason, renaming the repository is a breaking change for every
> consumer's import lines — avoid it.

## Code style

- Run `gofmt` (or `goimports`) before committing — CI rejects unformatted code.
- One service per file; the service struct holds only `*Client`.
- Validate at the boundary with plain `errors.New("wayaquick: ...")` messages, prefixed `wayaquick:`.
- Return `*APIError` for anything the server rejects, so callers can branch on `Code`.
- No comments explaining *what* the code does — only add one when the *why* is non-obvious.
