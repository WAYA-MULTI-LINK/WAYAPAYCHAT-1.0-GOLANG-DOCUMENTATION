package wayapay_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	wayapay "github.com/wayapaychat/wayapay-go/src/wayapay"
)

func TestVerifyBVN_RejectsBadFormatBeforeNetwork(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK, body: `{"success":true,"code":"00","data":{}}`}
	c := newClient(cap)

	for _, bad := range []string{"", "123", "2250080903X", "225008090377"} {
		if _, err := c.Identity.VerifyBVN(context.Background(), bad); err == nil {
			t.Errorf("VerifyBVN(%q) = nil error, want validation failure", bad)
		}
	}
	if cap.Calls() != 0 {
		t.Errorf("invalid BVN must not hit the network, got %d calls", cap.Calls())
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
		rec := wayapay.BVNRecord{WatchListed: in}
		if rec.IsWatchListed() != want {
			t.Errorf("IsWatchListed(%q) = %v, want %v", in, rec.IsWatchListed(), want)
		}
	}
}
