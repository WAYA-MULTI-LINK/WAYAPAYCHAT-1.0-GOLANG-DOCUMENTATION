package wayapay_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	wayapay "github.com/wayapaychat/wayapay-go/src/wayapay"
)

func TestListBanks_ReturnsBanks(t *testing.T) {
	c := okStub(`[{"code":"044","name":"Access Bank","id":"044","status":true},
	             {"code":"058","name":"GTBank","id":"058","status":true}]`)

	banks, err := c.Payout.ListBanks(context.Background())
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

func TestListBanks_HitsCorrectEndpoint(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK, body: `{"success":true,"code":"00","data":[]}`}
	c := newClient(cap)

	if _, err := c.Payout.ListBanks(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.lastReq.Method != http.MethodGet {
		t.Errorf("method = %s, want GET", cap.lastReq.Method)
	}
	if got := cap.lastReq.URL.Path; got != "/get-bank-list" {
		t.Errorf("path = %q", got)
	}
}

func TestVerifyAccount_ErrorsWhenBankCodeMissingForOthers(t *testing.T) {
	c := okStub(`{}`)
	req := validVerify()
	req.BankCode = ""

	_, err := c.Payout.VerifyAccount(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "bankCode is required") {
		t.Fatalf("expected bankCode validation error, got %v", err)
	}
}

func TestVerifyAccount_AllowsWayaBankWithoutBankCode(t *testing.T) {
	c := okStub(`{"successful":true,"accountNumber":"0123456789","accountName":"JOHN DOE"}`)

	out, err := c.Payout.VerifyAccount(context.Background(), wayapay.VerifyAccountRequest{
		AccountNumber: "0123456789",
		EnquiryType:   wayapay.EnquiryTypeWayaBank,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.AccountName != "JOHN DOE" {
		t.Errorf("accountName = %q", out.AccountName)
	}
}

func TestVerifyAccount_DecodesResolvedAccount(t *testing.T) {
	c := okStub(`{
		"successful": true,
		"responseCode": "00",
		"responseMessage": "Approved",
		"accountNumber": "0123456789",
		"accountName": "JOHN DOE",
		"bankCode": "044",
		"bankName": "Access Bank",
		"enquiryType": "OTHERS"
	}`)

	out, err := c.Payout.VerifyAccount(context.Background(), validVerify())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.Successful || out.AccountName != "JOHN DOE" || out.BankName != "Access Bank" {
		t.Errorf("decoded wrong: %+v", out)
	}
}

func TestVerifyAccount_PostsToCorrectPathWithBody(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK,
		body: `{"success":true,"code":"00","data":{"accountName":"JOHN DOE"}}`}
	c := newClient(cap)

	if _, err := c.Payout.VerifyAccount(context.Background(), validVerify()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.lastReq.Method != http.MethodPost {
		t.Errorf("method = %s, want POST", cap.lastReq.Method)
	}
	if cap.lastReq.URL.Path != "/verify-account" {
		t.Errorf("path = %q", cap.lastReq.URL.Path)
	}
	if !strings.Contains(cap.lastBody, `"accountNumber":"0123456789"`) {
		t.Errorf("body missing account number: %s", cap.lastBody)
	}
}

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
