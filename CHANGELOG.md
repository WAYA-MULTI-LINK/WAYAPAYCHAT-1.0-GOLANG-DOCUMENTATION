# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org).

## [1.0.0] - 2026-06-06

First release. A Go client for the WayaPay Merchant API v2 with no dependencies
outside the standard library.

### Added

- `New(merchantID, secretKey, ...Option)` — builds a concurrency-safe client; defaults to production.
- Functional options: `WithBaseURL`, `WithHTTPClient`, `WithUserAgent`, `WithMaxRetries`.
- `BaseURLProduction` constant (the default base URL; override with `WithBaseURL`).
- Four services mirroring the .NET library: `Payout`, `Collect`, `Identity`, `Webhooks`.
- `Payout.ListBanks` — returns all supported banks and their CBN codes.
- `Payout.VerifyAccount` — resolves an account number to its registered name; requires `BankCode` when `EnquiryType` is `OTHERS`.
- `Payout.Initiate` — initiates a bank transfer; `PROCESSING` means accepted, not settled.
- `Payout.Status` — returns the latest status of a payout by the reference you sent at initiation.
- `Collect.Initiate` — creates a payment link and returns the checkout URLs.
- `Collect.Status` — returns the current state of a deposit by its `refNo`.
- `Identity.VerifyBVN` — verifies a BVN with a local 11-digit format check before the network call; `BVNRecord.IsWatchListed` helper.
- `Webhooks.ConstructEvent` / `Webhooks.VerifySignature` — verify and parse inbound transaction webhooks.
- `GenerateReference(prefix)` — timestamped, collision-resistant idempotency key.
- `APIError` + `AsAPIError` — typed errors carrying the API code, message, and HTTP status.
- Automatic retry with exponential backoff on GET requests (timeouts, network errors, 429, 5xx); writes never auto-retry.
- Per-request validation that fails before any network call (missing required fields, malformed BVN, missing `BankCode`).
- `context.Context` support on every method.
- Stub-handler unit test suite plus build-tagged (`live`) integration tests.
