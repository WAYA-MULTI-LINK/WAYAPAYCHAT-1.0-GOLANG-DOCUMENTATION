package wayapay

import (
	"context"
	"errors"
	"net/http"
)

// Enquiry types for account verification.
const (
	EnquiryTypeOthers   = "OTHERS"   // any external bank
	EnquiryTypeWayaBank = "WAYABANK" // internal WayaBank account
)

// Dynamic account modes.
const (
	AccountModeOneTime = "ONE_TIME"
)

// AccountsService covers account verification and dynamic account creation.
type AccountsService struct{ client *Client }

// VerifyAccountRequest resolves an account number to its registered name.
type VerifyAccountRequest struct {
	AccountNumber string `json:"accountNumber"`
	BankCode      string `json:"bankCode,omitempty"` // not required when EnquiryType is WAYABANK
	EnquiryType   string `json:"enquiryType"`
}

// AccountVerification is the resolved account detail.
type AccountVerification struct {
	Successful      bool   `json:"successful"`
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	AccountNumber   string `json:"accountNumber"`
	AccountName     string `json:"accountName"`
	BankCode        string `json:"bankCode"`
	BankName        string `json:"bankName"`
	EnquiryType     string `json:"enquiryType"`
}

// Verify resolves an account number to its account name. Always verify a
// destination before you pay it.
//
// POST /account-enquiry/verify-account
func (s *AccountsService) Verify(ctx context.Context, req VerifyAccountRequest) (*AccountVerification, error) {
	if req.AccountNumber == "" {
		return nil, errors.New("wayapay: accountNumber is required")
	}
	if req.EnquiryType == "" {
		return nil, errors.New("wayapay: enquiryType is required")
	}
	if req.EnquiryType == EnquiryTypeOthers && req.BankCode == "" {
		return nil, errors.New("wayapay: bankCode is required when enquiryType is OTHERS")
	}

	var out AccountVerification
	if err := s.client.do(ctx, http.MethodPost, "/account-enquiry/verify-account", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DynamicAccountRequest mints a virtual NUBAN for inbound collection.
type DynamicAccountRequest struct {
	AccountName string `json:"accountName"`
	CustomerID  string `json:"customerId"`
	ReferenceID string `json:"referenceId"`
	Purpose     string `json:"purpose"`
	Mode        string `json:"mode"` // e.g. AccountModeOneTime
}

// DynamicAccount is a minted virtual account. Hand VirtualAccountNumber (same
// as NubanNumber) to the customer; watch CurrentBalance and CanReceivePayments
// to know it is live, and ExpiresAt to know when a ONE_TIME account closes.
type DynamicAccount struct {
	ID                   int64   `json:"id"`
	VirtualAccountNumber string  `json:"virtualAccountNumber"`
	NubanNumber          string  `json:"nubanNumber"`
	AccountName          string  `json:"accountName"`
	CustomerID           string  `json:"customerId"`
	AccountType          string  `json:"accountType"`
	Status               string  `json:"status"`
	IsActive             bool    `json:"isActive"`
	CanReceivePayments   bool    `json:"canReceivePayments"`
	ReferenceID          string  `json:"referenceId"`
	Metadata             string  `json:"metadata"`
	TotalLimit           float64 `json:"totalLimit"`
	CurrentBalance       float64 `json:"currentBalance"`
	AssignedAt           string  `json:"assignedAt"`
	ExpiresAt            string  `json:"expiresAt"`
	CreatedAt            string  `json:"createdAt"`
}

// CreateDynamicAccount mints a virtual NUBAN account a customer can pay into.
// Pair one per order or per customer, watch the inflow, reconcile cleanly.
//
// POST /account-enquiry/create-dynamic-account
func (s *AccountsService) CreateDynamicAccount(ctx context.Context, req DynamicAccountRequest) (*DynamicAccount, error) {
	switch {
	case req.AccountName == "":
		return nil, errors.New("wayapay: accountName is required")
	case req.CustomerID == "":
		return nil, errors.New("wayapay: customerId is required")
	case req.ReferenceID == "":
		return nil, errors.New("wayapay: referenceId is required")
	case req.Purpose == "":
		return nil, errors.New("wayapay: purpose is required")
	case req.Mode == "":
		return nil, errors.New("wayapay: mode is required")
	}

	var out DynamicAccount
	if err := s.client.do(ctx, http.MethodPost, "/account-enquiry/create-dynamic-account", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
