package wayapay_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestPayout_Validates(t *testing.T) {
	c := okStub(`{}`)

	// amount <= 0
	{
		req := validPayout()
		req.Amount = 0
		if _, err := c.Payout.Initiate(context.Background(), req); err == nil ||
			!strings.Contains(err.Error(), "amount must be greater than zero") {
			t.Errorf("amount=0: got %v", err)
		}
	}
	// missing reference
	{
		req := validPayout()
		req.Reference = ""
		if _, err := c.Payout.Initiate(context.Background(), req); err == nil ||
			!strings.Contains(err.Error(), "reference is required") {
			t.Errorf("missing reference: got %v", err)
		}
	}
	// missing narration
	{
		req := validPayout()
		req.Narration = ""
		if _, err := c.Payout.Initiate(context.Background(), req); err == nil ||
			!strings.Contains(err.Error(), "narration is required") {
			t.Errorf("missing narration: got %v", err)
		}
	}
}

func TestPayout_DecodesProcessingResult(t *testing.T) {
	c := okStub(`{"payoutReference":"PYT-99","merchantReference":"REF-001","status":"PROCESSING","message":"accepted"}`)

	out, err := c.Payout.Initiate(context.Background(), validPayout())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.PayoutReference != "PYT-99" || out.Status != "PROCESSING" {
		t.Errorf("decoded wrong: %+v", out)
	}
}

func TestPayout_SendsBody(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK,
		body: `{"success":true,"code":"00","data":{"status":"PROCESSING"}}`}
	c := newClient(cap)

	if _, err := c.Payout.Initiate(context.Background(), validPayout()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.lastReq.URL.Path != "/payment-payout/initiate" {
		t.Errorf("path = %q", cap.lastReq.URL.Path)
	}
	if !strings.Contains(cap.lastBody, `"amount":5000`) ||
		!strings.Contains(cap.lastBody, `"reference":"REF-001"`) {
		t.Errorf("body = %s", cap.lastBody)
	}
}
