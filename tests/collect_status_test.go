package wayapay_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	wayapay "github.com/wayapaychat/wayapay-go/src/wayapay"
)

func TestCollectStatus_RequiresRefNo(t *testing.T) {
	c := okStub(`{}`)
	if _, err := c.Collect.Status(context.Background(), ""); err == nil ||
		!strings.Contains(err.Error(), "refNo is required") {
		t.Errorf("empty refNo: got %v", err)
	}
}

func TestCollectStatus_EscapesPathAndUsesGET(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK,
		body: `{"success":true,"code":"00","data":{"status":"SUCCESSFUL"}}`}
	c := newClient(cap)

	// A refNo containing a slash must be percent-escaped into the path.
	if _, err := c.Collect.Status(context.Background(), "abc/def"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.lastReq.Method != http.MethodGet {
		t.Errorf("method = %q, want GET", cap.lastReq.Method)
	}
	// url.PathEscape keeps the raw escaped form on EscapedPath / URL.String.
	if got := cap.lastReq.URL.EscapedPath(); got != "/payment-collect/status/abc%2Fdef" {
		t.Errorf("escaped path = %q", got)
	}
}

func TestCollectStatus_DecodesSuccessful(t *testing.T) {
	c := okStub(`{"refNo":"177966","tranId":"GUID-1","status":"SUCCESSFUL",
	             "amount":"1500.00","amountPaid":"1500.00","currencyCode":"NGN"}`)

	out, err := c.Collect.Status(context.Background(), "177966")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.RefNo != "177966" || out.Status != "SUCCESSFUL" || out.AmountPaid != "1500.00" {
		t.Errorf("decoded wrong: %+v", out)
	}
	if out.ParsedStatus() != wayapay.CollectStatusSuccessful {
		t.Errorf("parsed = %v", out.ParsedStatus())
	}
}

func TestCollectStatus_OutcomeAndTerminal(t *testing.T) {
	cases := []struct {
		raw      string
		outcome  wayapay.CollectionOutcome
		terminal bool
	}{
		{"PENDING", wayapay.CollectionInFlight, false},
		{"PARTIAL", wayapay.CollectionInFlight, false},
		{"SUCCESSFUL", wayapay.CollectionSucceeded, true},
		{"REFUNDED", wayapay.CollectionRefunded, true},
		{"DECLINED", wayapay.CollectionNotDebited, true},
		{"BANK_ERROR", wayapay.CollectionIndeterminate, true},
		{"something-new", wayapay.CollectionIndeterminate, false},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			code := wayapay.ParseCollectionStatus(tc.raw)
			if got := code.Outcome(); got != tc.outcome {
				t.Errorf("Outcome(%q) = %v, want %v", tc.raw, got, tc.outcome)
			}
			if got := code.IsTerminal(); got != tc.terminal {
				t.Errorf("IsTerminal(%q) = %v, want %v", tc.raw, got, tc.terminal)
			}
		})
	}
}
