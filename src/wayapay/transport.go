package wayapay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"
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
//
// GET requests are retried up to c.maxRetries times on a transient failure
// (transport error, HTTP 429, or 5xx) with exponential backoff. Writes are
// never auto retried, so a payout is only ever sent once per call.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body, out any) error {
	full := c.baseURL + path
	if len(query) > 0 {
		full += "?" + query.Encode()
	}

	var raw []byte
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("wayapay: encode request: %w", err)
		}
		raw = encoded
	}

	retryable := method == http.MethodGet
	ceiling := 0
	if retryable {
		ceiling = c.maxRetries
	}

	var lastErr error
	for attempt := 0; ; attempt++ {
		if attempt > 0 {
			if err := backoff(ctx, attempt); err != nil {
				return err
			}
		}

		var reader io.Reader
		if raw != nil {
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
		if raw != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			// Honour an explicit cancellation/deadline rather than retrying it.
			if ctx.Err() != nil {
				return fmt.Errorf("wayapay: http call: %w", err)
			}
			lastErr = fmt.Errorf("wayapay: http call: %w", err)
			if retryable && attempt < ceiling {
				continue
			}
			return lastErr
		}

		payload, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = fmt.Errorf("wayapay: read response: %w", readErr)
			if retryable && attempt < ceiling {
				continue
			}
			return lastErr
		}

		if isTransient(resp.StatusCode) && retryable && attempt < ceiling {
			lastErr = &APIError{HTTPStatus: resp.StatusCode, Message: "transient upstream failure"}
			continue
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
}

// isTransient reports whether an HTTP status is worth retrying.
func isTransient(status int) bool {
	return status == http.StatusTooManyRequests || status >= 500
}

// backoff sleeps for an exponentially growing, jittered delay, returning early
// if the context is cancelled.
func backoff(ctx context.Context, attempt int) error {
	base := 1000 * (1 << (attempt - 1)) // 1s, 2s, 4s, ...
	if base > 4000 {
		base = 4000
	}
	delay := time.Duration(base+rand.Intn(200)) * time.Millisecond

	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
