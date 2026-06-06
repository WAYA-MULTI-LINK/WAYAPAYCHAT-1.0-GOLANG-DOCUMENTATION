package wayapay

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// TransactionsService covers transaction verification and history.
type TransactionsService struct{ client *Client }

// Transaction is the status of a single transaction. Trust Status over your own
// assumptions: SUCCESS is settled; anything else means keep waiting or dig in.
type Transaction struct {
	TransactionReference string  `json:"transactionReference"`
	MerchantReference    string  `json:"merchantReference"`
	Status               string  `json:"status"`
	Amount               float64 `json:"amount"`
	Currency             string  `json:"currency"`
	Channel              string  `json:"channel"`
	CustomerEmail        string  `json:"customerEmail"`
	PaidAt               string  `json:"paidAt"`
	CreatedAt            string  `json:"createdAt"`
}

// Verify retrieves the current status of a transaction by reference. Use it
// after a payout, or to confirm a collection landed. This is a GET, so it is
// retried automatically on a transient failure.
//
// GET /transaction/verify?reference=...
func (s *TransactionsService) Verify(ctx context.Context, reference string) (*Transaction, error) {
	if reference == "" {
		return nil, errors.New("wayapay: reference is required")
	}

	q := url.Values{}
	q.Set("reference", reference)

	var out Transaction
	if err := s.client.do(ctx, http.MethodGet, "/transaction/verify", q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// HistoryParams filters the transaction history. Zero value fields are omitted
// from the query, so the server defaults apply (page 0 onward).
type HistoryParams struct {
	Page   int        // zero based page index
	Size   int        // items per page
	Status string     // e.g. SUCCESS
	From   *time.Time // start of range; sent as RFC3339 UTC
	To     *time.Time // end of range; sent as RFC3339 UTC
}

// TransactionHistory is a paginated transaction list. Walk pages with Page
// until you reach TotalPages; TotalElements is the full count behind the
// filter so you can size your loop before you start.
type TransactionHistory struct {
	Items         []Transaction `json:"items"`
	Page          int           `json:"page"`
	Size          int           `json:"size"`
	TotalElements int64         `json:"totalElements"`
	TotalPages    int           `json:"totalPages"`
}

// History returns a paginated, filterable list of your transactions. This is
// the endpoint for reconciliation and dashboards.
//
// GET /transaction/history
func (s *TransactionsService) History(ctx context.Context, p HistoryParams) (*TransactionHistory, error) {
	q := url.Values{}
	if p.Page > 0 {
		q.Set("page", strconv.Itoa(p.Page))
	}
	if p.Size > 0 {
		q.Set("size", strconv.Itoa(p.Size))
	}
	if p.Status != "" {
		q.Set("status", p.Status)
	}
	if p.From != nil {
		q.Set("from", p.From.UTC().Format(time.RFC3339))
	}
	if p.To != nil {
		q.Set("to", p.To.UTC().Format(time.RFC3339))
	}

	var out TransactionHistory
	if err := s.client.do(ctx, http.MethodGet, "/transaction/history", q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
