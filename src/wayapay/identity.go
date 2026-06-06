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
	Email              string `json:"email"`
	Gender             string `json:"gender"`
	LGAOfOrigin        string `json:"lgaOfOrigin"`
	LGAOfResidence     string `json:"lgaOfResidence"`
	MaritalStatus      string `json:"maritalStatus"`
	Nationality        string `json:"nationality"`
	ResidentialAddress string `json:"residentialAddress"`
	StateOfOrigin      string `json:"stateOfOrigin"`
	WatchListed        string `json:"watchListed"`
	Base64Image        string `json:"base64Image"`
}

// IsWatchListed reports whether the BVN is flagged. The API returns the string
// "False" when clear; anything else (case insensitive) is treated as listed.
func (b *BVNRecord) IsWatchListed() bool {
	return !strings.EqualFold(strings.TrimSpace(b.WatchListed), "False")
}

// isElevenDigits reports whether s is exactly 11 ASCII digits.
func isElevenDigits(s string) bool {
	if len(s) != 11 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// VerifyBVN confirms a customer identity against their 11 digit Bank
// Verification Number. The BVN is validated locally as exactly 11 digits
// before the network call. Useful for KYC, onboarding, and fraud checks.
//
// POST /identity-verification/bvn
func (s *IdentityService) VerifyBVN(ctx context.Context, bvn string) (*BVNRecord, error) {
	if !isElevenDigits(bvn) {
		return nil, errors.New("wayapay: bvn must be exactly 11 digits")
	}

	body := map[string]string{"bvn": bvn}
	var out BVNRecord
	if err := s.client.do(ctx, http.MethodPost, "/identity-verification/bvn", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
