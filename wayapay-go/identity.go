package wayapay

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// IdentityService covers identity verification (BVN).
type IdentityService struct{ client *Client }

// BVNRecord is the identity detail returned for a BVN. Treat it as sensitive
// personal data: store, transmit, and log it only as your data protection
// obligations allow.
type BVNRecord struct {
	BVN                string `json:"bvn"`
	FirstName          string `json:"firstName"`
	MiddleName         string `json:"middleName"`
	LastName           string `json:"lastName"`
	DateOfBirth        string `json:"dateOfBirth"`
	PhoneNumber1       string `json:"phoneNumber1"`
	RegistrationDate   string `json:"registrationDate"`
	Gender             string `json:"gender"`
	LGAOfOrigin        string `json:"lgaOfOrigin"`
	LGAOfResidence     string `json:"lgaOfResidence"`
	MaritalStatus      string `json:"maritalStatus"`
	Nationality        string `json:"nationality"`
	ResidentialAddress string `json:"residentialAddress"`
	StateOfOrigin      string `json:"stateOfOrigin"`
	WatchListed        string `json:"watchListed"`
}

// IsWatchListed reports whether the BVN is flagged. The API returns the string
// "False" when clear; anything else (case insensitive) is treated as listed.
func (b *BVNRecord) IsWatchListed() bool {
	return !strings.EqualFold(strings.TrimSpace(b.WatchListed), "False")
}

// VerifyBVN confirms a customer identity against their 11 digit Bank
// Verification Number. Useful for KYC, onboarding, and fraud checks.
//
// POST /identity-verification/bvn
func (s *IdentityService) VerifyBVN(ctx context.Context, bvn string) (*BVNRecord, error) {
	if len(bvn) != 11 {
		return nil, errors.New("wayapay: bvn must be 11 digits")
	}

	body := map[string]string{"bvn": bvn}
	var out BVNRecord
	if err := s.client.do(ctx, http.MethodPost, "/identity-verification/bvn", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
