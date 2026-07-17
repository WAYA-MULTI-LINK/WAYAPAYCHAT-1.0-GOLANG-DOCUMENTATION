package wayaquick

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Webhook header and tolerance constants.
const (
	// WebhookTimestampHeader carries the epoch-millisecond timestamp that is
	// signed alongside the body.
	WebhookTimestampHeader = "X-Waya-Timestamp"
	// WebhookSignatureHeader carries the base64 HMAC-SHA256 signature.
	WebhookSignatureHeader = "X-Waya-Signature"
)

// DefaultWebhookTolerance is the default replay-protection window. Webhooks
// older or newer than this are rejected. Pass a negative tolerance to disable
// the timestamp check (not recommended outside tests).
const DefaultWebhookTolerance = 5 * time.Minute

// WebhookEvent is a transaction webhook delivered by WayaQuick when a payment
// becomes SUCCESSFUL, PARTIAL, or FAILED. Construct one only via ConstructEvent,
// which verifies the signature first. Use OrderID as your idempotency key — the
// same OrderID may fire more than once (e.g. a PARTIAL followed by a SUCCESSFUL).
//
// The wire contract mixes casing: the first fields are PascalCase (OrderId,
// Amount, ...) while newer fields are camelCase (customer, merchantId, ...).
// encoding/json is case-insensitive on unmarshal, so the canonical json tags
// below bind both.
type WebhookEvent struct {
	OrderID          string           `json:"orderId"`
	Amount           float64          `json:"amount"`
	Description      string           `json:"description"`
	Fee              float64          `json:"fee"`
	Currency         string           `json:"currency"`
	Status           string           `json:"status"`
	TranTime         string           `json:"tranTime"`
	TransactionDate  string           `json:"transactionDate"`
	ProductName      string           `json:"productName"`
	BusinessName     string           `json:"businessName"`
	Customer         *WebhookCustomer `json:"customer"`
	MerchantID       string           `json:"merchantId"`
	BranchCategory   string           `json:"branchCategory"`
	RecurrentPayment bool             `json:"recurrentPayment"`
}

// WebhookCustomer is the paying customer embedded in a WebhookEvent.
type WebhookCustomer struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
	CustomerID  string `json:"customerId"`
}

// WebhookError is returned when a webhook fails signature verification, replay
// checks, or cannot be parsed. It mirrors the .NET WayaQuickWebhookException.
type WebhookError struct {
	Message string
	Err     error // wrapped cause, when present (e.g. a JSON error)
}

func (e *WebhookError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("wayaquick: %s: %v", e.Message, e.Err)
	}
	return "wayaquick: " + e.Message
}

func (e *WebhookError) Unwrap() error { return e.Err }

// VerifySignature reports whether signature equals
// base64(HMAC-SHA256(key=secret, message="{timestamp}.{payload}")). It does NOT
// check the replay window — prefer ConstructEvent. The comparison is
// constant-time. It returns false on any empty input or a malformed base64
// signature.
func VerifySignature(payload, timestamp, signature, secret string) bool {
	if payload == "" || timestamp == "" || signature == "" || secret == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "." + payload))
	expected := mac.Sum(nil)

	provided, err := base64.StdEncoding.DecodeString(signature)
	if err != nil || len(provided) != len(expected) {
		return false
	}
	return hmac.Equal(expected, provided)
}

// ConstructEvent verifies the signature and replay window, then parses the body
// into a *WebhookEvent. It returns a *WebhookError if verification fails — it
// never returns an unverified event.
//
// payload is the exact raw request body, timestamp is the
// WebhookTimestampHeader value (epoch milliseconds), signature is the
// WebhookSignatureHeader value, and secret is the merchant secret for this
// event's environment (TEST or PRODUCTION).
//
// If tolerance >= 0 the timestamp must be within tolerance of now; a negative
// tolerance disables the replay check (not recommended outside tests).
func ConstructEvent(payload, timestamp, signature, secret string, tolerance time.Duration) (*WebhookEvent, error) {
	if secret == "" {
		return nil, &WebhookError{Message: "webhook secret is required"}
	}

	if !VerifySignature(payload, timestamp, signature, secret) {
		return nil, &WebhookError{Message: "webhook signature verification failed"}
	}

	if tolerance >= 0 {
		tsMs, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			return nil, &WebhookError{Message: "webhook timestamp is not a valid epoch-millisecond value", Err: err}
		}
		skew := time.Now().UTC().Sub(time.UnixMilli(tsMs).UTC())
		if math.Abs(float64(skew)) > float64(tolerance) {
			return nil, &WebhookError{Message: fmt.Sprintf(
				"webhook timestamp is outside the %.0fs tolerance window (possible replay)", tolerance.Seconds())}
		}
	}

	var evt WebhookEvent
	if err := json.Unmarshal([]byte(payload), &evt); err != nil {
		return nil, &WebhookError{Message: "webhook body is not valid JSON", Err: err}
	}
	return &evt, nil
}

// WebhookStatusCode is a recognised value of WebhookEvent.Status. Parse a raw
// string with ParseWebhookStatus; unrecognised values map to
// WebhookStatusUnknown.
type WebhookStatusCode string

const (
	// WebhookStatusUnknown means the status string was not recognised. Don't
	// fulfil; reconcile.
	WebhookStatusUnknown WebhookStatusCode = "UNKNOWN"
	// WebhookStatusSuccessful means the customer paid in full — fulfil.
	WebhookStatusSuccessful WebhookStatusCode = "SUCCESSFUL"
	// WebhookStatusPartial means the customer underpaid — hold fulfilment.
	WebhookStatusPartial WebhookStatusCode = "PARTIAL"
	// WebhookStatusFailed means funds never moved — no fulfilment.
	WebhookStatusFailed WebhookStatusCode = "FAILED"
)

// ParseWebhookStatus parses a raw webhook status string. The input is trimmed
// and upper-cased; unrecognised values return WebhookStatusUnknown.
func ParseWebhookStatus(s string) WebhookStatusCode {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "SUCCESSFUL":
		return WebhookStatusSuccessful
	case "PARTIAL":
		return WebhookStatusPartial
	case "FAILED":
		return WebhookStatusFailed
	default:
		return WebhookStatusUnknown
	}
}

// ParsedStatus parses the raw Status string into a WebhookStatusCode.
func (e *WebhookEvent) ParsedStatus() WebhookStatusCode {
	return ParseWebhookStatus(e.Status)
}

// ShouldFulfil reports whether the customer paid in full — safe to fulfil the
// order, after an idempotency check on OrderID. True only when the status is
// SUCCESSFUL.
func (e *WebhookEvent) ShouldFulfil() bool {
	return e.ParsedStatus() == WebhookStatusSuccessful
}

// WebhookService verifies and parses incoming transaction webhooks using the
// secret configured with WithWebhookSecret. It is a thin wrapper over the
// package-level ConstructEvent / VerifySignature; for per-environment routing
// call those directly with an explicit secret.
type WebhookService struct{ client *Client }

// ConstructEvent verifies the signature and replay window using the configured
// webhook secret, then parses the body. It returns an error if no secret is
// configured, or a *WebhookError if verification fails.
func (s *WebhookService) ConstructEvent(payload, timestamp, signature string, tolerance time.Duration) (*WebhookEvent, error) {
	if s.client.webhookSecret == "" {
		return nil, &WebhookError{Message: "no webhook secret configured: use WithWebhookSecret, or call the package-level ConstructEvent with an explicit secret"}
	}
	return ConstructEvent(payload, timestamp, signature, s.client.webhookSecret, tolerance)
}

// VerifySignature performs a signature-only check (no replay window) using the
// configured webhook secret. It returns false when no secret is configured.
func (s *WebhookService) VerifySignature(payload, timestamp, signature string) bool {
	if s.client.webhookSecret == "" {
		return false
	}
	return VerifySignature(payload, timestamp, signature, s.client.webhookSecret)
}
