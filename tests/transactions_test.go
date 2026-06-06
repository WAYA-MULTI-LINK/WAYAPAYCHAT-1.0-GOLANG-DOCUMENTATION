package wayapay_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	wayapay "github.com/wayapaychat/wayapay-go/src/wayapay"
)

func TestTransactionsVerify_RequiresReference(t *testing.T) {
	c := okStub(`{}`)
	if _, err := c.Transactions.Verify(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty reference")
	}
}

func TestTransactionsVerify_DecodesAndSendsQuery(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK,
		body: `{"success":true,"code":"00","data":{"transactionReference":"PYT-99","status":"SUCCESS","amount":5000}}`}
	c := newClient(cap)

	out, err := c.Transactions.Verify(context.Background(), "PYT-99")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Status != wayapay.StatusSuccess || out.Amount != 5000 {
		t.Errorf("decoded wrong: %+v", out)
	}
	if cap.lastReq.URL.Path != "/transaction/verify" {
		t.Errorf("path = %q", cap.lastReq.URL.Path)
	}
	if got := cap.lastReq.URL.Query().Get("reference"); got != "PYT-99" {
		t.Errorf("reference query = %q", got)
	}
}

func TestTransactionsHistory_BuildsQuery(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK,
		body: `{"success":true,"code":"00","data":{"items":[],"page":2,"size":20,"totalElements":0,"totalPages":0}}`}
	c := newClient(cap)

	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	_, err := c.Transactions.History(context.Background(), wayapay.HistoryParams{
		Page:   2,
		Size:   20,
		Status: wayapay.StatusSuccess,
		From:   &from,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q := cap.lastReq.URL.Query()
	if q.Get("page") != "2" || q.Get("size") != "20" || q.Get("status") != "SUCCESS" {
		t.Errorf("query = %v", q)
	}
	if q.Get("from") != "2026-05-01T00:00:00Z" {
		t.Errorf("from = %q", q.Get("from"))
	}
}

func TestTransactionsHistory_OmitsZeroValues(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK,
		body: `{"success":true,"code":"00","data":{"items":[]}}`}
	c := newClient(cap)

	if _, err := c.Transactions.History(context.Background(), wayapay.HistoryParams{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw := cap.lastReq.URL.RawQuery; raw != "" {
		t.Errorf("expected empty query for zero params, got %q", raw)
	}
}
