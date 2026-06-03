package wayapay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// envelope is the shape every endpoint returns.
type envelope struct {
	Success   bool            `json:"success"`
	Code      string          `json:"code"`
	Message   string          `json:"message"`
	Data      json.RawMessage `json:"data"`
	Timestamp string          `json:"timestamp"`
}

// do executes a request against the API and decodes the standard envelope.
//
// It returns an *APIError when success is false, when code is not "00", or when
// the body is not the expected envelope. When out is non nil the envelope data
// is unmarshalled into it.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	full := c.baseURL + path
	if len(query) > 0 {
		full += "?" + query.Encode()
	}

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("wayapay: encode request: %w", err)
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, full, reader)
	if err != nil {
		return fmt.Errorf("wayapay: build request: %w", err)
	}

	req.Header.Set("X-Merchant-Id", c.merchantID)
	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("wayapay: http call: %w", err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("wayapay: read response: %w", err)
	}

	var env envelope
	if err := json.Unmarshal(payload, &env); err != nil {
		// Body was not the expected envelope: gateway error, HTML page, etc.
		return &APIError{
			HTTPStatus: resp.StatusCode,
			Message:    fmt.Sprintf("unexpected response: %s", truncate(string(payload), 500)),
		}
	}

	if !env.Success || env.Code != codeSuccess {
		return &APIError{
			HTTPStatus: resp.StatusCode,
			Code:       env.Code,
			Message:    env.Message,
			Timestamp:  env.Timestamp,
		}
	}

	if out != nil && len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return fmt.Errorf("wayapay: decode data: %w", err)
		}
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
