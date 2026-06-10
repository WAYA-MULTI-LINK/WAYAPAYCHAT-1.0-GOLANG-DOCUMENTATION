package wayapay_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	wayapay "github.com/wayapaychat/wayapay-go/src/wayapay"
)

func TestPayoutStatus_RequiresReference(t *testing.T) {
	c := okStub(`{}`)
	if _, err := c.Payout.Status(context.Background(), ""); err == nil ||
		!strings.Contains(err.Error(), "reference is required") {
		t.Errorf("empty reference: got %v", err)
	}
}

func TestPayoutStatus_EscapesPathAndUsesGET(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK,
		body: `{"success":true,"code":"00","data":{"status":"SUCCESS"}}`}
	c := newClient(cap)

	if _, err := c.Payout.Status(context.Background(), "PAYOUT/001"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.lastReq.Method != http.MethodGet {
		t.Errorf("method = %q, want GET", cap.lastReq.Method)
	}
	if got := cap.lastReq.URL.EscapedPath(); got != "/payment-payout/status/PAYOUT%2F001" {
		t.Errorf("escaped path = %q", got)
	}
}

func TestPayoutStatus_DecodesSuccess(t *testing.T) {
	c := okStub(`{"transactionReference":"PAYOUT-001","status":"SUCCESS",
	             "amount":"500.00","destinationAccountName":"JOHN DOE"}`)

	out, err := c.Payout.Status(context.Background(), "PAYOUT-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.TransactionReference != "PAYOUT-001" || out.Status != "SUCCESS" {
		t.Errorf("decoded wrong: %+v", out)
	}
	if out.ParsedStatus() != wayapay.PayoutStatusSuccess {
		t.Errorf("parsed = %v", out.ParsedStatus())
	}
}

func TestPayoutStatus_OutcomeAndTerminal(t *testing.T) {
	cases := []struct {
		raw      string
		outcome  wayapay.PayoutOutcome
		terminal bool
	}{
		{"PENDING", wayapay.PayoutReconciling, false},
		{"SUCCESS", wayapay.PayoutSucceeded, true},
		{"REVERSED", wayapay.PayoutReversed, true},
		{"weird", wayapay.PayoutReconciling, false},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			code := wayapay.ParsePayoutStatus(tc.raw)
			if got := code.Outcome(); got != tc.outcome {
				t.Errorf("Outcome(%q) = %v, want %v", tc.raw, got, tc.outcome)
			}
			if got := code.IsTerminal(); got != tc.terminal {
				t.Errorf("IsTerminal(%q) = %v, want %v", tc.raw, got, tc.terminal)
			}
		})
	}
}
