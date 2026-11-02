package wayaquick_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestCollect_Validates(t *testing.T) {
	c := okStub(`{}`)

	// missing required transactionId
	{
		req := validCollect()
		req.TransactionID = ""
		if _, err := c.Collect.Initiate(context.Background(), req); err == nil ||
			!strings.Contains(err.Error(), "transactionId is required") {
			t.Errorf("missing transactionId: got %v", err)
		}
	}
	// missing amount
	{
		req := validCollect()
		req.Amount = ""
		if _, err := c.Collect.Initiate(context.Background(), req); err == nil ||
			!strings.Contains(err.Error(), "amount is required") {
			t.Errorf("missing amount: got %v", err)
		}
	}
	// missing email
	{
		req := validCollect()
		req.Email = ""
		if _, err := c.Collect.Initiate(context.Background(), req); err == nil ||
			!strings.Contains(err.Error(), "email is required") {
			t.Errorf("missing email: got %v", err)
		}
	}
}

func TestCollect_DecodesResponse(t *testing.T) {
	c := okStub(`{
		"uniqueId": "TXN-V2-20260604-01",
		"transactionId": "178056425312256824",
		"email": "john@example.com",
		"amount": "1500.00",
		"checkOutUrl": "https://pay.wayaquick.com/?_tranId=178056425312256824",
		"merchantId": "MER_LOfV71779980928158w7MEX",
		"expiresAt": "2026-06-04T09:40:00Z"
	}`)

	out, err := c.Collect.Initiate(context.Background(), validCollect())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.CheckOutURL != "https://pay.wayaquick.com/?_tranId=178056425312256824" || out.TransactionID != "178056425312256824" {
		t.Errorf("decoded wrong: %+v", out)
	}
}

func TestCollect_PostsToCorrectPath(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK,
		body: `{"success":true,"code":"00","data":{"checkOutUrl":"https://pay.wayaquick.com/?_tranId=123"}}`}
	c := newClient(cap)

	if _, err := c.Collect.Initiate(context.Background(), validCollect()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.lastReq.URL.Path != "/payment-collect/initiate" {
		t.Errorf("path = %q", cap.lastReq.URL.Path)
	}
	if !strings.Contains(cap.lastBody, `"transactionId":"TXN-V2-20260604-01"`) {
		t.Errorf("expected transactionId in body, got: %s", cap.lastBody)
	}
}
