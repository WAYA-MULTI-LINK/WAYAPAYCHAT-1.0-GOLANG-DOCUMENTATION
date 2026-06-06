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
- `BaseURLProduction` / `BaseURLStaging` constants.
- `Banks.List` — returns all supported banks and their CBN codes.
- `Accounts.Verify` — resolves an account number to its registered name; requires `BankCode` when `EnquiryType` is `OTHERS`.
- `Accounts.CreateDynamicAccount` — mints a virtual NUBAN for inbound collection.
- `Identity.VerifyBVN` — verifies a BVN with a local 11-digit format check before the network call; `BVNRecord.IsWatchListed` helper.
- `Payout.Initiate` — initiates a bank transfer; `PROCESSING` means accepted, not settled.
- `Collect.Initiate` — creates a payment link and returns the checkout URLs.
- `Transactions.Verify` / `Transactions.History` — transaction status lookup and paginated, filterable history.
- `GenerateReference(prefix)` — timestamped, collision-resistant idempotency key.
- `APIError` + `AsAPIError` — typed errors carrying the API code, message, and HTTP status.
- Automatic retry with exponential backoff on GET requests (timeouts, network errors, 429, 5xx); writes never auto-retry.
- Per-request validation that fails before any network call (missing required fields, malformed BVN, missing `BankCode`).
- `context.Context` support on every method.
- Stub-handler unit test suite plus build-tagged (`live`) integration tests.
