package wayapay_test

import (
	"context"
	"net/http"
	"testing"
)

func TestBanksList_ReturnsBanks(t *testing.T) {
	c := okStub(`[{"code":"044","name":"Access Bank","id":"044","status":true},
	             {"code":"058","name":"GTBank","id":"058","status":true}]`)

	banks, err := c.Banks.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(banks) != 2 {
		t.Fatalf("got %d banks, want 2", len(banks))
	}
	if banks[0].Code != "044" || banks[0].Name != "Access Bank" || !banks[0].Status {
		t.Errorf("first bank decoded wrong: %+v", banks[0])
	}
}

func TestBanksList_HitsCorrectEndpoint(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK, body: `{"success":true,"code":"00","data":[]}`}
	c := newClient(cap)

	if _, err := c.Banks.List(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.lastReq.Method != http.MethodGet {
		t.Errorf("method = %s, want GET", cap.lastReq.Method)
	}
	if got := cap.lastReq.URL.Path; got != "/account-enquiry/get-bank-list" {
		t.Errorf("path = %q", got)
	}
}
