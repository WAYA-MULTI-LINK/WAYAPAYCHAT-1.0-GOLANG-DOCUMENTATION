package wayapay_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	wayapay "github.com/wayapaychat/wayapay-go/src/wayapay"
)

func TestVerify_ErrorsWhenBankCodeMissingForOthers(t *testing.T) {
	c := okStub(`{}`)
	req := validVerify()
	req.BankCode = ""

	_, err := c.Accounts.Verify(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "bankCode is required") {
		t.Fatalf("expected bankCode validation error, got %v", err)
	}
}

func TestVerify_AllowsWayaBankWithoutBankCode(t *testing.T) {
	c := okStub(`{"successful":true,"accountNumber":"0123456789","accountName":"JOHN DOE"}`)

	out, err := c.Accounts.Verify(context.Background(), wayapay.VerifyAccountRequest{
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

func TestVerify_DecodesResolvedAccount(t *testing.T) {
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

	out, err := c.Accounts.Verify(context.Background(), validVerify())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.Successful || out.AccountName != "JOHN DOE" || out.BankName != "Access Bank" {
		t.Errorf("decoded wrong: %+v", out)
	}
}

func TestVerify_PostsToCorrectPathWithBody(t *testing.T) {
	cap := &capturingTransport{status: http.StatusOK,
		body: `{"success":true,"code":"00","data":{"accountName":"JOHN DOE"}}`}
	c := newClient(cap)

	if _, err := c.Accounts.Verify(context.Background(), validVerify()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.lastReq.Method != http.MethodPost {
		t.Errorf("method = %s, want POST", cap.lastReq.Method)
	}
	if cap.lastReq.URL.Path != "/account-enquiry/verify-account" {
		t.Errorf("path = %q", cap.lastReq.URL.Path)
	}
	if !strings.Contains(cap.lastBody, `"accountNumber":"0123456789"`) {
		t.Errorf("body missing account number: %s", cap.lastBody)
	}
}

func TestCreateDynamicAccount_Validates(t *testing.T) {
	c := okStub(`{}`)
	_, err := c.Accounts.CreateDynamicAccount(context.Background(), wayapay.DynamicAccountRequest{})
	if err == nil || !strings.Contains(err.Error(), "accountName is required") {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestCreateDynamicAccount_Decodes(t *testing.T) {
	c := okStub(`{"id":42,"virtualAccountNumber":"9900112233","nubanNumber":"9900112233",
	             "accountName":"ACME LTD","canReceivePayments":true,"currentBalance":0}`)

	out, err := c.Accounts.CreateDynamicAccount(context.Background(), wayapay.DynamicAccountRequest{
		AccountName: "ACME LTD",
		CustomerID:  "CUST-1",
		ReferenceID: "REF-1",
		Purpose:     "order",
		Mode:        wayapay.AccountModeOneTime,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.VirtualAccountNumber != "9900112233" || !out.CanReceivePayments {
		t.Errorf("decoded wrong: %+v", out)
	}
}
