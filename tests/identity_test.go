package wayaquick_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	wayaquick "github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick"
)

func TestInitiateBVN_RejectsBadFormatBeforeNetwork(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK, body: `{"success":true,"code":"00","data":{}}`}
	c := newClient(cap)

	for _, bad := range []string{"", "123", "2250080903X", "225008090377"} {
		if _, err := c.Identity.InitiateBVN(context.Background(), bad); err == nil {
			t.Errorf("InitiateBVN(%q) = nil error, want validation failure", bad)
		}
	}
	if cap.Calls() != 0 {
		t.Errorf("invalid BVN must not hit the network, got %d calls", cap.Calls())
	}
}

func TestInitiateBVN_Success(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK,
		body: `{"success":true,"code":"00","message":"OTP sent","data":{"reference":"ref-123","phoneNumber":"234****0000","expiresAt":"2026-10-13T14:32:10.482"}}`}
	c := newClient(cap)

	res, err := c.Identity.InitiateBVN(context.Background(), "22222222222")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Reference != "ref-123" || res.PhoneNumber != "234****0000" {
		t.Errorf("decoded wrong: %+v", res)
	}
	if cap.lastReq.URL.Path != "/identity-verification/bvn/initiate" {
		t.Errorf("path = %q", cap.lastReq.URL.Path)
	}
	if !strings.Contains(cap.lastBody, `"bvn":"22222222222"`) {
		t.Errorf("body = %s", cap.lastBody)
	}
}

func TestCompleteBVN_RequiresReferenceAndOtp(t *testing.T) {
	c := okStub(`{}`)
	if _, err := c.Identity.CompleteBVN(context.Background(), "", "123456"); err == nil {
		t.Error("expected error for empty reference")
	}
	if _, err := c.Identity.CompleteBVN(context.Background(), "ref-123", ""); err == nil {
		t.Error("expected error for empty otp")
	}
}

func TestCompleteBVN_Success(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK,
		body: `{"success":true,"code":"00","message":"completed","data":{"bvn":"22222222222","firstName":"TEST","lastName":"USER","watchListed":"SANDBOX"}}`}
	c := newClient(cap)

	rec, err := c.Identity.CompleteBVN(context.Background(), "ref-123", "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.FirstName != "TEST" || rec.LastName != "USER" || rec.WatchListed != "SANDBOX" {
		t.Errorf("decoded wrong: %+v", rec)
	}
	if cap.lastReq.URL.Path != "/identity-verification/bvn/complete" {
		t.Errorf("path = %q", cap.lastReq.URL.Path)
	}
	if !strings.Contains(cap.lastBody, `"reference":"ref-123"`) || !strings.Contains(cap.lastBody, `"otp":"123456"`) {
		t.Errorf("body = %s", cap.lastBody)
	}
}

func TestVerifyBVN_DecodesRecord(t *testing.T) {
	c := okStub(`{"bvn":"22500809037","firstName":"JOHN","lastName":"DOE","watchListed":"False"}`)

	rec, err := c.Identity.VerifyBVN(context.Background(), "22500809037")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.FirstName != "JOHN" || rec.LastName != "DOE" {
		t.Errorf("decoded wrong: %+v", rec)
	}
	if rec.IsWatchListed() {
		t.Error("expected not watchlisted for \"False\"")
	}
}

func TestVerifyBVN_PostsToCorrectPath(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK,
		body: `{"success":true,"code":"00","data":{"bvn":"22500809037"}}`}
	c := newClient(cap)

	if _, err := c.Identity.VerifyBVN(context.Background(), "22500809037"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.lastReq.URL.Path != "/identity-verification/bvn" {
		t.Errorf("path = %q", cap.lastReq.URL.Path)
	}
	if !strings.Contains(cap.lastBody, `"bvn":"22500809037"`) {
		t.Errorf("body = %s", cap.lastBody)
	}
}

func TestBVNRecord_IsWatchListed(t *testing.T) {
	cases := map[string]bool{
		"False":   false,
		"false":   false,
		" False ": false,
		"True":    true,
		"":        true,
		"YES":     true,
	}
	for in, want := range cases {
		rec := wayaquick.BVNRecord{WatchListed: in}
		if rec.IsWatchListed() != want {
			t.Errorf("IsWatchListed(%q) = %v, want %v", in, rec.IsWatchListed(), want)
		}
	}
}
