package wayaquick

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

// BVNInitiateResult is returned by Step 1 of BVN verification.
// It carries the session reference and the masked phone number the OTP was sent to.
type BVNInitiateResult struct {
	Reference   string `json:"reference"`
	PhoneNumber string `json:"phoneNumber"`
	ExpiresAt   string `json:"expiresAt"`
}

// InitiateBVN starts a two-step BVN verification. It validates that bvn is exactly
// 11 digits and requests an OTP to be sent to the phone number registered against the BVN.
//
// POST /identity-verification/bvn/initiate
func (s *IdentityService) InitiateBVN(ctx context.Context, bvn string) (*BVNInitiateResult, error) {
	if !isElevenDigits(bvn) {
		return nil, errors.New("wayaquick: bvn must be exactly 11 digits")
	}

	body := map[string]string{"bvn": bvn}
	var out BVNInitiateResult
	if err := s.client.do(ctx, http.MethodPost, "/identity-verification/bvn/initiate", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CompleteBVN completes Step 2 of BVN verification by submitting the session reference
// from InitiateBVN and the OTP received on the registered phone number.
//
// POST /identity-verification/bvn/complete
func (s *IdentityService) CompleteBVN(ctx context.Context, reference, otp string) (*BVNRecord, error) {
	reference = strings.TrimSpace(reference)
	otp = strings.TrimSpace(otp)
	if reference == "" {
		return nil, errors.New("wayaquick: reference is required")
	}
	if otp == "" {
		return nil, errors.New("wayaquick: otp is required")
	}

	body := map[string]string{
		"reference": reference,
		"otp":       otp,
	}
	var out BVNRecord
	if err := s.client.do(ctx, http.MethodPost, "/identity-verification/bvn/complete", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// VerifyBVN confirms a customer identity against their 11 digit Bank
// Verification Number.
//
// Deprecated: WayaQuick BVN verification requires a two-step OTP verification flow.
// Use InitiateBVN followed by CompleteBVN instead.
func (s *IdentityService) VerifyBVN(ctx context.Context, bvn string) (*BVNRecord, error) {
	if !isElevenDigits(bvn) {
		return nil, errors.New("wayaquick: bvn must be exactly 11 digits")
	}

	body := map[string]string{"bvn": bvn}
	var out BVNRecord
	if err := s.client.do(ctx, http.MethodPost, "/identity-verification/bvn", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
