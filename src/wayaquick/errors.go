package wayaquick

import (
	"errors"
	"fmt"
)

// codeSuccess is the only response code that means all is well.
const codeSuccess = "00"

// APIError is returned when the API responds with success=false, a non "00"
// code, or a body that is not the standard envelope. Branch on Code to react
// to specific failures.
type APIError struct {
	HTTPStatus int    // underlying HTTP status code
	Code       string // API code, e.g. "07"; "00" means success
	Message    string // human readable message from the API
	Timestamp  string // server timestamp, when present
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("wayaquick: api error %s: %s (http %d)", e.Code, e.Message, e.HTTPStatus)
	}
	return fmt.Sprintf("wayaquick: api error: %s (http %d)", e.Message, e.HTTPStatus)
}

// AsAPIError extracts an *APIError from err, if present, so you can read Code
// and Message without a manual type assertion.
//
//	if ae, ok := wayaquick.AsAPIError(err); ok && ae.Code == "07" {
//		// handle invalid account number
//	}
func AsAPIError(err error) (*APIError, bool) {
	var ae *APIError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
