# WayaQuick Golang SDK Documentation (v1.0.2)

Official integration guide for the **WayaQuick Merchant API v2** using the official Go SDK.

---

## 1. Package Overview

* **Module Path**: `github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick`
* **Current Version**: `v1.0.2`
* **Go Support**: Go 1.20 or higher
* **Dependencies**: Standard library only (zero external dependencies)
* **Concurrency**: `*wayaquick.Client` is thread-safe and safe for concurrent use across multiple goroutines.
* **Runnable Example**: [`samples/main.go`](samples/main.go)

### Installation

```bash
go get github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick@v1.0.2
```

Import statement in your code:

```go
import wayaquick "github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick"
```

---

## 2. Initialization & Authentication

All API operations require your **Merchant ID** (`MER_...`) and **Secret Key** (`WAYASECK_TEST_...` for test sandbox, `WAYASECK_...` for live production):

```go
package main

import (
    "os"
    "time"
    wayaquick "github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick"
)

func main() {
    client := wayaquick.New(
        os.Getenv("WAYA_MERCHANT_ID"), // MER_...
        os.Getenv("WAYA_SECRET_KEY"),  // WAYASECK_TEST_... or WAYASECK_...
        // Optional configuration options:
        // wayaquick.WithMaxRetries(2),
        // wayaquick.WithBaseURL(wayaquick.BaseURLProduction),
    )
    _ = client
}
```

> **Security Rule**: The Secret Key allows debiting your merchant account. Never embed this key in frontend clients or mobile applications.

---

## 3. Core Features

### 3.1. Reference Generation (Idempotency)

Use the built-in reference helper to generate timestamped, collision-resistant identifiers:

```go
ref := wayaquick.GenerateReference("PAYOUT")
// Produces: PAYOUT-1748160000000-A1B2C3D4
```

---

### 3.2. List Banks (`GET /get-bank-list`)

Fetch supported banks and financial institutions in Nigeria with their CBN bank codes:

```go
banks, err := client.Payout.ListBanks(ctx)
if err != nil {
    log.Fatalf("failed to list banks: %v", err)
}

for _, b := range banks {
    fmt.Printf("%s (%s)\n", b.Name, b.Code)
}
```

---

### 3.3. Account Verification / Name Enquiry (`POST /verify-account`)

Validate the bank account number and retrieve the registered account name:

```go
// Inter-bank enquiry (requires BankCode):
account, err := client.Payout.VerifyAccount(ctx, wayaquick.VerifyAccountRequest{
    AccountNumber: "0123456789",
    BankCode:      "044",                       // Access Bank
    EnquiryType:   wayaquick.EnquiryTypeOthers, // Default: "OTHERS"
})
if err != nil {
    log.Fatalf("verify failed: %v", err)
}
fmt.Println("Resolved Account Name:", account.AccountName)

// Intra-bank enquiry (Waya Bank to Waya Bank):
wayaAccount, err := client.Payout.VerifyAccount(ctx, wayaquick.VerifyAccountRequest{
    AccountNumber: "0123456789",
    EnquiryType:   wayaquick.EnquiryTypeWayaBank, // BankCode not required
})
```

---

### 3.4. BVN Identity Verification (Two-Step OTP Flow)

In compliance with data privacy regulations, BVN verification uses a 2-step OTP flow:

```go
// Step 1: Initiate verification (sends SMS OTP to customer's BVN phone number)
session, err := client.Identity.InitiateBVN(ctx, "22222222222") // exactly 11 digits
if err != nil {
    log.Fatalf("failed to initiate BVN check: %v", err)
}
fmt.Printf("OTP sent to: %s (Session: %s)\n", session.PhoneNumber, session.Reference)

// Step 2: Complete verification with the received OTP
profile, err := client.Identity.CompleteBVN(ctx, session.Reference, "123456")
if err != nil {
    log.Fatalf("failed to complete BVN check: %v", err)
}
fmt.Printf("Verified: %s %s\n", profile.FirstName, profile.LastName)
fmt.Println("Watchlisted:", profile.IsWatchListed())
```

---

### 3.5. Disburse Payouts (`POST /payment-payout/initiate`)

Send payouts from your merchant balance to any bank account:

```go
// 1. Submit payout request
payout, err := client.Payout.Initiate(ctx, wayaquick.PayoutRequest{
    Amount:        25000,
    Currency:      "NGN",
    AccountNumber: "0123456789",
    BankCode:      "044",
    AccountName:   "JOHN DOE",
    Reference:     wayaquick.GenerateReference("PAYOUT"),
    Narration:     "Salary Disbursement",
})
if err != nil {
    log.Fatalf("payout failed: %v", err)
}

fmt.Println("Payout status:", payout.Status) // "PROCESSING"

// 2. Reconcile payout status using reference:
ref := payout.MerchantReference
if ref == "" {
    ref = payout.PayoutReference
}

status, err := client.Payout.Status(ctx, ref)
if err != nil {
    log.Fatalf("status query failed: %v", err)
}

switch status.ParsedStatus().Outcome() {
case wayaquick.PayoutSucceeded:
    fmt.Println("Funds delivered.")
case wayaquick.PayoutReversed:
    fmt.Println("Transfer failed; wallet re-credited.")
default:
    fmt.Println("Transfer is in flight; poll again later.")
}
```

---

### 3.6. Payment Collections / Hosted Checkout (`POST /payment-collect/initiate`)

Collect customer payments via WayaQuick Hosted Checkout:

```go
// 1. Create Checkout Session
checkout, err := client.Collect.Initiate(ctx, wayaquick.CollectRequest{
    TransactionID: wayaquick.GenerateReference("TXN"),
    Amount:        "1500.00",
    Currency:      "NGN",
    FirstName:     "John",
    LastName:      "Doe",
    Email:         "customer@example.com",
    Phone:         "08012345678",
    Description:   "Order #1001",
})
if err != nil {
    log.Fatalf("checkout initiation failed: %v", err)
}

fmt.Println("Redirect customer to checkout URL:", checkout.CheckOutURL)

// 2. Reconcile Status by TransactionID
collectStatus, err := client.Collect.Status(ctx, checkout.TransactionID)
if err != nil {
    log.Fatalf("collection status check failed: %v", err)
}

if collectStatus.ParsedStatus() == wayaquick.CollectStatusSuccessful {
    fmt.Printf("Funds confirmed! Amount received: %s. Fulfil order.\n", collectStatus.AmountPaid)
}
```

---

### 3.7. Inbound Webhook Verification

Verify HMAC-SHA256 signatures over raw inbound HTTP requests to prevent replay and forgery attacks:

```go
package main

import (
    "io"
    "net/http"
    "os"
    wayaquick "github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick"
)

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    // Read raw body bytes
    bodyBytes, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "bad request", http.StatusBadRequest)
        return
    }
    rawBody := string(bodyBytes)

    timestamp := r.Header.Get("x-waya-timestamp")
    signature := r.Header.Get("x-waya-signature")
    secretKey := os.Getenv("WAYA_SECRET_KEY")

    event, err := wayaquick.ConstructEvent(
        rawBody,
        timestamp,
        signature,
        secretKey,
        wayaquick.DefaultWebhookTolerance, // 5-minute replay window
    )
    if err != nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    if event.ShouldFulfil() {
        // Fulfil order keyed by event.OrderID
    }

    w.WriteHeader(http.StatusOK)
}
```

---

## 4. Error Handling

API errors return a `*wayaquick.APIError`. Inspect the code using `wayaquick.AsAPIError(err)`:

```go
acct, err := client.Payout.VerifyAccount(ctx, req)
if err != nil {
    if apiErr, ok := wayaquick.AsAPIError(err); ok {
        fmt.Printf("WayaQuick Error [Code %s]: %s (Status: %d)\n",
            apiErr.Code, apiErr.Message, apiErr.StatusCode)
    } else {
        fmt.Printf("Network / Transport Error: %v\n", err)
    }
    return
}
```

---

## 5. Runnable Example Document

The SDK includes a complete runnable demonstration:

* **File Location**: [`samples/main.go`](samples/main.go)
* **Run Command**:
  ```bash
  WAYA_MERCHANT_ID=MER_... WAYA_SECRET_KEY=WAYASECK_TEST_... go run ./samples
  ```

---

## 6. Parseable AI Text (Copy & Paste for LLMs)

Copy the prompt block below to give any AI coding tool complete knowledge of how to use this library:

```xml
<wayaquick_golang_sdk_context version="1.0.2">
  <system_role>
    You are an expert Go backend engineer integrating the official WayaQuick Go SDK (`github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick`).
    Target Go 1.20+. Adhere strictly to the v2 API contract and method signatures below.
  </system_role>

  <rules>
    1. INITIALIZATION: `client := wayaquick.New(merchantID, secretKey)`. Client is thread-safe.
    2. BVN 2-STEP OTP: BVN verification requires `client.Identity.InitiateBVN(ctx, bvn)` (sends OTP) followed by `client.Identity.CompleteBVN(ctx, reference, otp)`. Single-step verification is deprecated.
    3. ACCOUNT ENQUIRY: `client.Payout.VerifyAccount(ctx, wayaquick.VerifyAccountRequest{ ... })`. If `EnquiryType == wayaquick.EnquiryTypeWayaBank`, `BankCode` is not required.
    4. HOSTED CHECKOUT: `client.Collect.Initiate(ctx, wayaquick.CollectRequest{ ... })`. Redirect customer to `checkout.CheckOutURL`.
    5. PAYOUTS ARE ASYNC: `client.Payout.Initiate(...)` returns status `PROCESSING`. Reconcile by polling `client.Payout.Status(ctx, ref)` or via webhooks.
    6. WEBHOOK RAW BODY: `wayaquick.ConstructEvent(rawBody, timestamp, signature, secret, tolerance)`. Verify against the unparsed raw string.
  </rules>

  <package_info>
    <module_path>github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick</module_path>
    <version>v1.0.2</version>
    <install>go get github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick@v1.0.2</install>
    <sample_file>samples/main.go</sample_file>
  </package_info>

  <code_template>
    <![CDATA[
import (
    "context"
    "os"
    wayaquick "github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick"
)

client := wayaquick.New(os.Getenv("WAYA_MERCHANT_ID"), os.Getenv("WAYA_SECRET_KEY"))
ctx := context.Background()

// 1. List Banks
banks, err := client.Payout.ListBanks(ctx)

// 2. Account Name Enquiry
verified, err := client.Payout.VerifyAccount(ctx, wayaquick.VerifyAccountRequest{
    AccountNumber: "0123456789",
    BankCode:      "044",
    EnquiryType:   wayaquick.EnquiryTypeOthers,
})

// 3. BVN 2-Step OTP
bvnSession, err := client.Identity.InitiateBVN(ctx, "22222222222")
bvnProfile, err := client.Identity.CompleteBVN(ctx, bvnSession.Reference, "123456")

// 4. Initiate Payout
payout, err := client.Payout.Initiate(ctx, wayaquick.PayoutRequest{
    Amount:        5000,
    Currency:      "NGN",
    AccountNumber: verified.AccountNumber,
    BankCode:      "044",
    AccountName:   verified.AccountName,
    Reference:     wayaquick.GenerateReference("PAYOUT"),
    Narration:     "Transfer payment",
})
payoutStatus, err := client.Payout.Status(ctx, payout.MerchantReference)

// 5. Payment Collection
checkout, err := client.Collect.Initiate(ctx, wayaquick.CollectRequest{
    TransactionID: wayaquick.GenerateReference("TXN"),
    Amount:        "1500.00",
    Currency:      "NGN",
    FirstName:     "John",
    LastName:      "Doe",
    Email:         "customer@example.com",
    Phone:         "08012345678",
    Description:   "Order #1001",
})
// Redirect customer to: checkout.CheckOutURL
collectStatus, err := client.Collect.Status(ctx, checkout.TransactionID)
    ]]>
  </code_template>
</wayaquick_golang_sdk_context>
```

