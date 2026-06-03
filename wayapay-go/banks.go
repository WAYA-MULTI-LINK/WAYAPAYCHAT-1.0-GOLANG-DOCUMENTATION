package wayapay

import (
	"context"
	"net/http"
)

// BanksService covers bank list lookups.
type BanksService struct{ client *Client }

// Bank is a supported bank entry. Use Code in account verification and payout.
type Bank struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	ID     string `json:"id"`
	Status bool   `json:"status"` // whether the bank is currently reachable
}

// List returns the supported banks with their codes.
//
// GET /account-enquiry/get-bank-list
func (s *BanksService) List(ctx context.Context) ([]Bank, error) {
	var out []Bank
	if err := s.client.do(ctx, http.MethodGet, "/account-enquiry/get-bank-list", nil, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
