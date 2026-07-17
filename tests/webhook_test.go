package wayaquick_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	wayaquick "github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick"
)

const webhookSecret = "WAYASECK_TEST_demo_webhook_secret"

// signWebhook computes base64(HMAC-SHA256(secret, "{ts}.{body}")).
func signWebhook(ts, body, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "." + body))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func nowMillis() string { return strconv.FormatInt(time.Now().UnixMilli(), 10) }

// Mixed PascalCase/camelCase wire body with a nested customer; description and
// branchCategory are omitted to exercise zero-value defaults.
const webhookBody = `{"OrderId":"177966","Amount":1500.00,"Fee":15.00,` +
	`"Currency":"NGN","Status":"SUCCESSFUL","productName":"CARD",` +
	`"customer":{"name":"John Doe","email":"john@example.com","phoneNumber":"0801"},` +
	`"merchantId":"MER_xyz","recurrentPayment":false}`

func TestWebhook_ValidSignatureParses(t *testing.T) {
	ts := nowMillis()
	sig := signWebhook(ts, webhookBody, webhookSecret)

	evt, err := wayaquick.ConstructEvent(webhookBody, ts, sig, webhookSecret, wayaquick.DefaultWebhookTolerance)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if evt.OrderID != "177966" || evt.Amount != 1500.00 || evt.Fee != 15.00 {
		t.Errorf("decoded wrong: %+v", evt)
	}
	if evt.Customer == nil || evt.Customer.Email != "john@example.com" || evt.Customer.PhoneNumber != "0801" {
		t.Errorf("customer decoded wrong: %+v", evt.Customer)
	}
	if evt.Description != "" || evt.BranchCategory != "" {
		t.Errorf("omitted fields not zero-valued: %+v", evt)
	}
	if !evt.ShouldFulfil() {
		t.Errorf("ShouldFulfil = false for SUCCESSFUL")
	}
}

func TestWebhook_WrongSecretRejected(t *testing.T) {
	ts := nowMillis()
	sig := signWebhook(ts, webhookBody, webhookSecret)

	if _, err := wayaquick.ConstructEvent(webhookBody, ts, sig, "OTHER_SECRET", wayaquick.DefaultWebhookTolerance); err == nil ||
		!strings.Contains(err.Error(), "signature") {
		t.Errorf("wrong secret: got %v", err)
	}
}

func TestWebhook_TamperedBodyRejected(t *testing.T) {
	ts := nowMillis()
	sig := signWebhook(ts, webhookBody, webhookSecret)
	tampered := strings.Replace(webhookBody, "1500.00", "9999.00", 1)

	if _, err := wayaquick.ConstructEvent(tampered, ts, sig, webhookSecret, wayaquick.DefaultWebhookTolerance); err == nil ||
		!strings.Contains(err.Error(), "signature") {
		t.Errorf("tampered body: got %v", err)
	}
}

func TestWebhook_StaleTimestampRejected(t *testing.T) {
	ts := strconv.FormatInt(time.Now().Add(-10*time.Minute).UnixMilli(), 10)
	sig := signWebhook(ts, webhookBody, webhookSecret)

	if _, err := wayaquick.ConstructEvent(webhookBody, ts, sig, webhookSecret, wayaquick.DefaultWebhookTolerance); err == nil ||
		!strings.Contains(err.Error(), "tolerance") {
		t.Errorf("stale timestamp: got %v", err)
	}
}

func TestWebhook_DisabledReplayAcceptsStale(t *testing.T) {
	ts := strconv.FormatInt(time.Now().Add(-10*time.Minute).UnixMilli(), 10)
	sig := signWebhook(ts, webhookBody, webhookSecret)

	// A negative tolerance disables the timestamp check.
	if _, err := wayaquick.ConstructEvent(webhookBody, ts, sig, webhookSecret, -1); err != nil {
		t.Errorf("disabled replay should accept stale: %v", err)
	}
}

func TestWebhook_VerifySignature(t *testing.T) {
	ts := nowMillis()
	sig := signWebhook(ts, webhookBody, webhookSecret)

	if !wayaquick.VerifySignature(webhookBody, ts, sig, webhookSecret) {
		t.Errorf("valid signature returned false")
	}
	// Missing inputs.
	if wayaquick.VerifySignature(webhookBody, "", sig, webhookSecret) {
		t.Errorf("empty timestamp should be false")
	}
	if wayaquick.VerifySignature(webhookBody, ts, "", webhookSecret) {
		t.Errorf("empty signature should be false")
	}
	if wayaquick.VerifySignature(webhookBody, ts, sig, "") {
		t.Errorf("empty secret should be false")
	}
	// Malformed base64.
	if wayaquick.VerifySignature(webhookBody, ts, "not!!base64", webhookSecret) {
		t.Errorf("malformed base64 should be false")
	}
}

func TestWebhook_InvalidJSONWithValidSignature(t *testing.T) {
	body := `{"OrderId":` // truncated, invalid JSON
	ts := nowMillis()
	sig := signWebhook(ts, body, webhookSecret)

	if _, err := wayaquick.ConstructEvent(body, ts, sig, webhookSecret, wayaquick.DefaultWebhookTolerance); err == nil ||
		!strings.Contains(err.Error(), "JSON") {
		t.Errorf("invalid JSON: got %v", err)
	}
}

func TestParseWebhookStatus(t *testing.T) {
	cases := []struct {
		raw  string
		want wayaquick.WebhookStatusCode
	}{
		{"SUCCESSFUL", wayaquick.WebhookStatusSuccessful},
		{" successful ", wayaquick.WebhookStatusSuccessful},
		{"PARTIAL", wayaquick.WebhookStatusPartial},
		{"FAILED", wayaquick.WebhookStatusFailed},
		{"nope", wayaquick.WebhookStatusUnknown},
		{"", wayaquick.WebhookStatusUnknown},
	}
	for _, tc := range cases {
		if got := wayaquick.ParseWebhookStatus(tc.raw); got != tc.want {
			t.Errorf("ParseWebhookStatus(%q) = %v, want %v", tc.raw, got, tc.want)
		}
	}
}

func TestWebhookService_ConfiguredSecret(t *testing.T) {
	rt := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{}`), nil
	})
	c := newClient(rt, wayaquick.WithWebhookSecret(webhookSecret))

	ts := nowMillis()
	sig := signWebhook(ts, webhookBody, webhookSecret)

	evt, err := c.Webhooks.ConstructEvent(webhookBody, ts, sig, wayaquick.DefaultWebhookTolerance)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if evt.OrderID != "177966" {
		t.Errorf("decoded wrong: %+v", evt)
	}
	if !c.Webhooks.VerifySignature(webhookBody, ts, sig) {
		t.Errorf("VerifySignature via service returned false")
	}
}

func TestWebhookService_NoSecretConfigured(t *testing.T) {
	rt := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusOK, `{}`), nil
	})
	c := newClient(rt) // no WithWebhookSecret

	ts := nowMillis()
	sig := signWebhook(ts, webhookBody, webhookSecret)

	if _, err := c.Webhooks.ConstructEvent(webhookBody, ts, sig, wayaquick.DefaultWebhookTolerance); err == nil ||
		!strings.Contains(err.Error(), "secret") {
		t.Errorf("no secret configured: got %v", err)
	}
	if c.Webhooks.VerifySignature(webhookBody, ts, sig) {
		t.Errorf("VerifySignature should be false with no secret configured")
	}
}
