package wayaquick

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// CollectService covers payment collections and checkout URLs.
type CollectService struct{ client *Client }

// CollectRequest starts a payment collection and returns a checkout URL.
type CollectRequest struct {
	TransactionID string         `json:"transactionId"` // Unique reference / idempotency key
	Amount        string         `json:"amount"`        // Quoted string, e.g. "1500.00"
	Currency      string         `json:"currency"`      // e.g. "NGN"
	FirstName     string         `json:"firstName"`     // Customer first name (2-100 characters)
	LastName      string         `json:"lastName"`      // Customer last name (2-100 characters)
	Email         string         `json:"email"`         // Customer email
	Phone         string         `json:"phone"`         // Customer phone number
	Description   string         `json:"description,omitempty"`
	Meta          map[string]any `json:"meta,omitempty"`
}

// CollectionResult is the response from POST /payment-collect/initiate.
// Redirect the customer to CheckOutURL to complete payment.
type CollectionResult struct {
	UniqueID      string `json:"uniqueId"`
	TransactionID string `json:"transactionId"`
	Email         string `json:"email,omitempty"`
	Amount        string `json:"amount,omitempty"`
	CheckOutURL   string `json:"checkOutUrl"`
	MerchantID    string `json:"merchantId,omitempty"`
	ExpiresAt     string `json:"expiresAt,omitempty"`
}

// PaymentLink is an alias for CollectionResult for backwards compatibility.
type PaymentLink = CollectionResult

// Initiate starts a payment collection and returns a checkout URL you redirect your customer to.
//
// POST /payment-collect/initiate
func (s *CollectService) Initiate(ctx context.Context, req CollectRequest) (*CollectionResult, error) {
	switch {
	case req.TransactionID == "":
		return nil, errors.New("wayaquick: transactionId is required")
	case req.Amount == "":
		return nil, errors.New("wayaquick: amount is required")
	case req.Currency == "":
		req.Currency = "NGN"
	case req.FirstName == "":
		return nil, errors.New("wayaquick: firstName is required")
	case req.LastName == "":
		return nil, errors.New("wayaquick: lastName is required")
	case req.Email == "":
		return nil, errors.New("wayaquick: email is required")
	case req.Phone == "":
		return nil, errors.New("wayaquick: phone is required")
	}

	var out CollectionResult
	if err := s.client.do(ctx, http.MethodPost, "/payment-collect/initiate", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CollectionStatus is the current state of a single collection (deposit)
// transaction. Use RefNo as the idempotency key when fulfilling a SUCCESSFUL
// payment. Interpret Status with ParsedStatus / ParseCollectionStatus.
type CollectionStatus struct {
	RefNo            string `json:"refNo"`
	TranID           string `json:"tranId"`
	MerchantID       string `json:"merchantId"`
	Amount           string `json:"amount"`
	CustomerEmail    string `json:"customerEmail"`
	AmountPaid       string `json:"amountPaid"`
	Fee              string `json:"fee"`
	CurrencyCode     string `json:"currencyCode"`
	Status           string `json:"status"`
	SettlementStatus string `json:"settlementStatus"`
	Channel          string `json:"channel"`
	ProcessedBy      string `json:"processedBy"`
	Description      string `json:"description"`
	Environment      string `json:"environment"`
	TranDate         string `json:"tranDate"`
}

// Status returns the current state of a deposit by its refNo (the gateway
// transactionId / webhook OrderId). The deposit webhook is the primary signal;
// this is the pull / safety-net path for reconciliation. This is a GET, so it
// is retried automatically on a transient failure.
//
// GET /payment-collect/status/{refNo}
func (s *CollectService) Status(ctx context.Context, refNo string) (*CollectionStatus, error) {
	if refNo == "" {
		return nil, errors.New("wayaquick: refNo is required")
	}

	var out CollectionStatus
	if err := s.client.do(ctx, http.MethodGet, "/payment-collect/status/"+url.PathEscape(refNo), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ParsedStatus parses the raw Status string into a CollectionStatusCode.
func (s *CollectionStatus) ParsedStatus() CollectionStatusCode {
	return ParseCollectionStatus(s.Status)
}
