package wayaquick

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// Transaction and payout statuses.
const (
	StatusProcessing = "PROCESSING" // accepted, not yet settled
	StatusSuccess    = "SUCCESS"    // settled
	StatusFailed     = "FAILED"
)

// Enquiry types for account verification.
const (
	EnquiryTypeOthers   = "OTHERS"   // any external bank
	EnquiryTypeWayaBank = "WAYABANK" // internal WayaBank account
)

// PayoutService covers bank lookups, account verification, and outbound transfers.
type PayoutService struct{ client *Client }

// Bank is a supported bank entry. Use Code in account verification and payout.
type Bank struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	ID     string `json:"id"`
	Status bool   `json:"status"` // whether the bank is currently reachable
}

// ListBanks returns the supported banks with their codes. This is a GET, so it
// is retried automatically on a transient failure.
//
// GET /get-bank-list
func (s *PayoutService) ListBanks(ctx context.Context) ([]Bank, error) {
	var out []Bank
	if err := s.client.do(ctx, http.MethodGet, "/get-bank-list", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

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

// VerifyAccount resolves an account number to its account name. Always verify a
// destination before you pay it.
//
// POST /verify-account
func (s *PayoutService) VerifyAccount(ctx context.Context, req VerifyAccountRequest) (*AccountVerification, error) {
	if req.AccountNumber == "" {
		return nil, errors.New("wayaquick: accountNumber is required")
	}
	if req.EnquiryType == "" {
		return nil, errors.New("wayaquick: enquiryType is required")
	}
	if req.EnquiryType == EnquiryTypeOthers && req.BankCode == "" {
		return nil, errors.New("wayaquick: bankCode is required when enquiryType is OTHERS")
	}

	var out AccountVerification
	if err := s.client.do(ctx, http.MethodPost, "/verify-account", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PayoutRequest sends funds from your merchant balance to a bank account.
// Reference is your dedup and tracking key, so make it unique per payout.
type PayoutRequest struct {
	Amount        float64 `json:"amount"` // major unit, e.g. 25000 = NGN 25,000
	Currency      string  `json:"currency"`
	AccountNumber string  `json:"accountNumber"`
	BankCode      string  `json:"bankCode"`
	AccountName   string  `json:"accountName"` // match the verified name
	Reference     string  `json:"reference"`
	Narration     string  `json:"narration"`
}

// PayoutResult is the accepted payout acknowledgement. A PROCESSING status
// means accepted, not settled: poll Status with the reference you sent at
// initiation to confirm the money actually landed.
type PayoutResult struct {
	PayoutReference   string `json:"payoutReference"`
	MerchantReference string `json:"merchantReference"`
	Status            string `json:"status"`
	Message           string `json:"message"`
}

// Initiate sends a payout. Verify the destination with VerifyAccount first,
// and supply a fresh unique Reference per payout so retries map cleanly to the
// original record (GenerateReference is a convenient source).
//
// This is a write: it is never auto retried, so the funds move at most once per
// call.
//
// POST /payment-payout/initiate
func (s *PayoutService) Initiate(ctx context.Context, req PayoutRequest) (*PayoutResult, error) {
	switch {
	case req.Amount <= 0:
		return nil, errors.New("wayaquick: amount must be greater than zero")
	case req.Currency == "":
		return nil, errors.New("wayaquick: currency is required")
	case req.AccountNumber == "":
		return nil, errors.New("wayaquick: accountNumber is required")
	case req.BankCode == "":
		return nil, errors.New("wayaquick: bankCode is required")
	case req.AccountName == "":
		return nil, errors.New("wayaquick: accountName is required")
	case req.Reference == "":
		return nil, errors.New("wayaquick: reference is required")
	case req.Narration == "":
		return nil, errors.New("wayaquick: narration is required")
	}

	var out PayoutResult
	if err := s.client.do(ctx, http.MethodPost, "/payment-payout/initiate", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PayoutStatus is the latest state of a single payout (disbursement) by the
// reference you sent at initiation. Use TransactionReference as the idempotency
// key. Interpret Status with ParsedStatus / ParsePayoutStatus.
type PayoutStatus struct {
	TransactionReference     string `json:"transactionReference"`
	Status                   string `json:"status"`
	Amount                   string `json:"amount"`
	DestinationAccountNumber string `json:"destinationAccountNumber"`
	DestinationAccountName   string `json:"destinationAccountName"`
	DestinationBankName      string `json:"destinationBankName"`
	Narration                string `json:"narration"`
	CreatedAt                string `json:"createdAt"`
}

// Status returns the latest status of a payout by the reference you sent at
// initiation. It is scoped to the authenticated merchant: a reference belonging
// to another merchant (or a different environment) returns a 404. This is a
// GET, so it is retried automatically on a transient failure.
//
// GET /payment-payout/status/{reference}
func (s *PayoutService) Status(ctx context.Context, reference string) (*PayoutStatus, error) {
	if reference == "" {
		return nil, errors.New("wayaquick: reference is required")
	}

	var out PayoutStatus
	if err := s.client.do(ctx, http.MethodGet, "/payment-payout/status/"+url.PathEscape(reference), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ParsedStatus parses the raw Status string into a PayoutStatusCode.
func (s *PayoutStatus) ParsedStatus() PayoutStatusCode {
	return ParsePayoutStatus(s.Status)
}
