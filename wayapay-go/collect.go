package wayapay

import (
	"context"
	"errors"
	"net/http"
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
