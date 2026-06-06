package wayapay

import (
	"context"
	"errors"
	"net/http"
)

// Transaction and payout statuses.
const (
	StatusProcessing = "PROCESSING" // accepted, not yet settled
	StatusSuccess    = "SUCCESS"    // settled
	StatusFailed     = "FAILED"
)

// PayoutService covers outbound transfers.
type PayoutService struct{ client *Client }

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
// means accepted, not settled: poll Transactions.Verify with PayoutReference
// to confirm the money actually landed.
type PayoutResult struct {
	PayoutReference   string `json:"payoutReference"`
	MerchantReference string `json:"merchantReference"`
	Status            string `json:"status"`
	Message           string `json:"message"`
}

// Initiate sends a payout. Verify the destination with Accounts.Verify first,
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
		return nil, errors.New("wayapay: amount must be greater than zero")
	case req.Currency == "":
		return nil, errors.New("wayapay: currency is required")
	case req.AccountNumber == "":
		return nil, errors.New("wayapay: accountNumber is required")
	case req.BankCode == "":
		return nil, errors.New("wayapay: bankCode is required")
	case req.AccountName == "":
		return nil, errors.New("wayapay: accountName is required")
	case req.Reference == "":
		return nil, errors.New("wayapay: reference is required")
	case req.Narration == "":
		return nil, errors.New("wayapay: narration is required")
	}

	var out PayoutResult
	if err := s.client.do(ctx, http.MethodPost, "/payment-payout/initiate", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
