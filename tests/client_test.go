package wayapay_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	wayapay "github.com/wayapaychat/wayapay-go/src/wayapay"
)

func TestNew_WiresUpServices(t *testing.T) {
	c := wayapay.New("MER_X", "WAYASECK_X")
	if c.Banks == nil || c.Accounts == nil || c.Identity == nil ||
		c.Payout == nil || c.Collect == nil || c.Transactions == nil {
		t.Fatal("expected every service to be initialised")
	}
}

func TestRequest_SendsAuthAndMerchantHeaders(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK, body: `{"success":true,"code":"00","data":[]}`}
	c := newClient(cap)

	if _, err := c.Banks.List(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := cap.lastReq.Header.Get("Authorization"); got != "Bearer WAYASECK_TEST_key" {
		t.Errorf("Authorization = %q", got)
	}
	if got := cap.lastReq.Header.Get("X-Merchant-Id"); got != "MER_TEST" {
		t.Errorf("X-Merchant-Id = %q", got)
	}
	if got := cap.lastReq.Header.Get("Accept"); got != "application/json" {
		t.Errorf("Accept = %q", got)
	}
}

func TestRequest_SetsContentTypeOnlyForWrites(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK, body: `{"success":true,"code":"00","data":{}}`}
	c := newClient(cap)

	// GET: no body, no Content-Type.
	_, _ = c.Banks.List(context.Background())
	if ct := cap.lastReq.Header.Get("Content-Type"); ct != "" {
		t.Errorf("GET Content-Type = %q, want empty", ct)
	}

	// POST: Content-Type is set.
	_, _ = c.Identity.VerifyBVN(context.Background(), "22500809037")
	if ct := cap.lastReq.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("POST Content-Type = %q", ct)
	}
}

func TestRequest_ReturnsAPIError_WhenSuccessFalse(t *testing.T) {
	c := errStub(http.StatusBadRequest, "57", "IP is not whitelisted")

	_, err := c.Banks.List(context.Background())
	ae, ok := wayapay.AsAPIError(err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if ae.Code != "57" || ae.HTTPStatus != http.StatusBadRequest {
		t.Errorf("got code=%q status=%d", ae.Code, ae.HTTPStatus)
	}
	if !strings.Contains(ae.Message, "whitelisted") {
		t.Errorf("message = %q", ae.Message)
	}
}

func TestRequest_ReturnsAPIError_OnNonEnvelopeBody(t *testing.T) {
	c := stubClient(http.StatusBadGateway, "<html>502 Bad Gateway</html>")

	_, err := c.Banks.List(context.Background())
	ae, ok := wayapay.AsAPIError(err)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if !strings.Contains(ae.Message, "unexpected response") {
		t.Errorf("message = %q", ae.Message)
	}
}

func TestRequest_RetriesGetOnTransientStatus(t *testing.T) {
	// First two calls 503, third succeeds. With MaxRetries=2 we should land it.
	var n int
	rt := roundTripFunc(func(*http.Request) (*http.Response, error) {
		n++
		if n < 3 {
			return jsonResponse(http.StatusServiceUnavailable, `{"success":false,"code":"99"}`), nil
		}
		return jsonResponse(http.StatusOK, `{"success":true,"code":"00","data":[]}`), nil
	})
	c := newClient(rt, wayapay.WithMaxRetries(2))

	if _, err := c.Banks.List(context.Background()); err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3 attempts, got %d", n)
	}
}

func TestRequest_DoesNotRetryWrites(t *testing.T) {
	cap := &capturingTransport{status: http.StatusServiceUnavailable, body: `{"success":false,"code":"99"}`}
	c := newClient(cap, wayapay.WithMaxRetries(5))

	_, err := c.Payout.Initiate(context.Background(), validPayout())
	if err == nil {
		t.Fatal("expected error")
	}
	if cap.Calls() != 1 {
		t.Errorf("writes must not retry: got %d calls", cap.Calls())
	}
}

func TestRequest_PropagatesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, r.Context().Err()
	})
	c := newClient(rt, wayapay.WithMaxRetries(3))

	_, err := c.Banks.List(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestGenerateReference(t *testing.T) {
	if ref := wayapay.GenerateReference("PAYOUT"); !strings.HasPrefix(ref, "PAYOUT-") {
		t.Errorf("ref = %q, want PAYOUT- prefix", ref)
	}
	if ref := wayapay.GenerateReference(""); !strings.HasPrefix(ref, "WP-") {
		t.Errorf("empty prefix should default to WP-, got %q", ref)
	}
	if a, b := wayapay.GenerateReference("X"), wayapay.GenerateReference("X"); a == b {
		t.Errorf("consecutive references collided: %q", a)
	}
}
