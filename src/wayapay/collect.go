package wayapay

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// Payment link types.
const (
	PaymentLinkTypeOneTime = "ONE_TIME_PAYMENT_LINK"
)

// CollectService covers payment link creation.
type CollectService struct{ client *Client }

// CollectRequest creates a payment link your customer can pay through.
type CollectRequest struct {
	PaymentLinkType  string         `json:"paymentLinkType"`
	PaymentLinkName  string         `json:"paymentLinkName"`
	Description      string         `json:"description"`
	PayableAmount    float64        `json:"payableAmount"` // major unit
	Currency         string         `json:"currency"`
	SuccessMessage   string         `json:"successMessage,omitempty"`
	PhoneNumber      string         `json:"phoneNumber,omitempty"`
	RedirectLink     string         `json:"redirectLink"`
	CustomURL        string         `json:"customURL,omitempty"`
	TotalCount       int            `json:"totalCount,omitempty"`
	ChargeInterval   string         `json:"chargeInterval,omitempty"` // e.g. MONTHLY, for subscriptions
	PlanID           string         `json:"planId,omitempty"`
	ExpiryDate       string         `json:"expiryDate,omitempty"` // required when LinkCanExpire is true
	LinkCanExpire    bool           `json:"linkCanExpire,omitempty"`
	OtherDetailsJSON map[string]any `json:"otherDetailsJSON,omitempty"`
}

// PaymentLink is the created link. Send the customer to CustomerPaymentLink or
// the tidier ShortURL; keep PaymentLinkReference to reconcile later.
// MerchantKeyMode tells you whether the link was created with a TEST or live
// key, so you never confuse a sandbox link for a real one.
type PaymentLink struct {
	MerchantID                string  `json:"merchantId"`
	PaymentLinkID             string  `json:"paymentLinkId"`
	PaymentLinkType           string  `json:"paymentLinkType"`
	PaymentLinkName           string  `json:"paymentLinkName"`
	Description               string  `json:"description"`
	PayableAmount             float64 `json:"payableAmount"`
	Currency                  string  `json:"currency"`
	AmountText                string  `json:"amountText"`
	SuccessMessage            string  `json:"successMessage"`
	RedirectLink              string  `json:"redirectLink"`
	CustomerPaymentLink       string  `json:"customerPaymentLink"`
	ShortURL                  string  `json:"shortUrl"`
	Status                    string  `json:"status"`
	Deleted                   bool    `json:"deleted"`
	MerchantKeyMode           string  `json:"merchantKeyMode"`
	PaymentLinkReference      string  `json:"paymentLinkReference"`
	ExpiryDate                string  `json:"expiryDate"`
	TotalCount                int     `json:"totalCount"`
	LinkCanExpire             bool    `json:"linkCanExpire"`
	IsSubscriptionPaymentLink bool    `json:"isSubscriptionPaymentLink"`
	CreatedBy                 int64   `json:"createdBy"`
	CreatedAt                 string  `json:"createdAt"`
}

// Initiate creates a payment link.
//
// This call fails unless you have whitelisted your server IPs and configured
// payment preferences on the merchant dashboard.
//
// POST /payment-collect/initiate
func (s *CollectService) Initiate(ctx context.Context, req CollectRequest) (*PaymentLink, error) {
	switch {
	case req.PaymentLinkType == "":
		return nil, errors.New("wayapay: paymentLinkType is required")
	case req.PaymentLinkName == "":
		return nil, errors.New("wayapay: paymentLinkName is required")
	case req.Description == "":
		return nil, errors.New("wayapay: description is required")
	case req.PayableAmount <= 0:
		return nil, errors.New("wayapay: payableAmount must be greater than zero")
	case req.Currency == "":
		return nil, errors.New("wayapay: currency is required")
	case req.RedirectLink == "":
		return nil, errors.New("wayapay: redirectLink is required")
	case req.LinkCanExpire && req.ExpiryDate == "":
		return nil, errors.New("wayapay: expiryDate is required when linkCanExpire is true")
	}

	var out PaymentLink
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
		return nil, errors.New("wayapay: refNo is required")
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
