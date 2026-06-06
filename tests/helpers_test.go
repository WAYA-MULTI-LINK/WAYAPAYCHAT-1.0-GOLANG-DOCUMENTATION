package wayapay_test

import (
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	wayapay "github.com/wayapaychat/wayapay-go/src/wayapay"
)

// roundTripFunc adapts a function into an http.RoundTripper so tests can stand
// in for the network without an httptest server.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// jsonResponse builds an *http.Response carrying a JSON body.
func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

// stubClient returns a client whose every request resolves to the given status
// and body. Use it for happy-path and error-path assertions.
func stubClient(status int, body string) *wayapay.Client {
	rt := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(status, body), nil
	})
	return newClient(rt)
}

// okStub wraps data in the success envelope the API returns.
func okStub(data string) *wayapay.Client {
	return stubClient(http.StatusOK, `{"success":true,"code":"00","data":`+data+`}`)
}

// errStub wraps an error envelope the API returns when success is false.
func errStub(status int, code, message string) *wayapay.Client {
	return stubClient(status, `{"success":false,"code":"`+code+`","message":"`+message+`"}`)
}

// capturingTransport records the last request it saw and a copy of its body so
// tests can assert on method, path, headers, and payload shape.
type capturingTransport struct {
	status   int
	body     string
	calls    int32
	lastReq  *http.Request
	lastBody string
}

func (c *capturingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	atomic.AddInt32(&c.calls, 1)
	c.lastReq = r
	if r.Body != nil {
		raw, _ := io.ReadAll(r.Body)
		c.lastBody = string(raw)
	}
	return jsonResponse(c.status, c.body), nil
}

func (c *capturingTransport) Calls() int { return int(atomic.LoadInt32(&c.calls)) }

func newClient(rt http.RoundTripper, opts ...wayapay.Option) *wayapay.Client {
	base := []wayapay.Option{
		wayapay.WithBaseURL("https://api.test"),
		wayapay.WithHTTPClient(&http.Client{Transport: rt}),
		wayapay.WithMaxRetries(0), // deterministic by default; opt back in per test
	}
	return wayapay.New("MER_TEST", "WAYASECK_TEST_key", append(base, opts...)...)
}
