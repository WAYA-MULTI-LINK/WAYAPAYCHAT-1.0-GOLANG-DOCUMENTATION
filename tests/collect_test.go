package wayaquick_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestCollect_Validates(t *testing.T) {
	c := okStub(`{}`)

	// missing required field
	{
		req := validCollect()
		req.RedirectLink = ""
		if _, err := c.Collect.Initiate(context.Background(), req); err == nil ||
			!strings.Contains(err.Error(), "redirectLink is required") {
			t.Errorf("missing redirectLink: got %v", err)
		}
	}
	// payable amount <= 0
	{
		req := validCollect()
		req.PayableAmount = 0
		if _, err := c.Collect.Initiate(context.Background(), req); err == nil ||
			!strings.Contains(err.Error(), "payableAmount must be greater than zero") {
			t.Errorf("amount=0: got %v", err)
		}
	}
	// expiring link without an expiry date
	{
		req := validCollect()
		req.LinkCanExpire = true
		if _, err := c.Collect.Initiate(context.Background(), req); err == nil ||
			!strings.Contains(err.Error(), "expiryDate is required") {
			t.Errorf("expiring link without date: got %v", err)
		}
	}
}

func TestCollect_DecodesLink(t *testing.T) {
	c := okStub(`{"paymentLinkId":"PL-1","shortUrl":"https://pay.test/abc",
	             "customerPaymentLink":"https://pay.test/full/abc","status":"ACTIVE",
	             "paymentLinkReference":"PLR-1","merchantKeyMode":"TEST"}`)

	out, err := c.Collect.Initiate(context.Background(), validCollect())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ShortURL != "https://pay.test/abc" || out.PaymentLinkReference != "PLR-1" {
		t.Errorf("decoded wrong: %+v", out)
	}
}

func TestCollect_PostsToCorrectPath(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK,
		body: `{"success":true,"code":"00","data":{"shortUrl":"https://pay.test/x"}}`}
	c := newClient(cap)

	if _, err := c.Collect.Initiate(context.Background(), validCollect()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.lastReq.URL.Path != "/payment-collect/initiate" {
		t.Errorf("path = %q", cap.lastReq.URL.Path)
	}
}
