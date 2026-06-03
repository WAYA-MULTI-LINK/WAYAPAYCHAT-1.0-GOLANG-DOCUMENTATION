package wayapay

import (
	"net/http"
	"strings"
	"time"
)

const (
	// BaseURLProduction is the live environment base URL.
	BaseURLProduction = "https://services.wayapay.ng/merchant-middleware/api/v2"
	// BaseURLStaging is the staging/sandbox environment base URL.
	BaseURLStaging = "https://services.staging.wayapay.ng/merchant-middleware/api/v2"

	defaultTimeout = 30 * time.Second
	defaultUA      = "wayapay-go/1.0"
)

// Client is the WayaPay Merchant API v2 client. Build it once with New and
// reuse it across goroutines; it is safe for concurrent use.
type Client struct {
	merchantID string
	secretKey  string
	baseURL    string
	userAgent  string
	httpClient *http.Client

	Banks        *BanksService
	Accounts     *AccountsService
	Identity     *IdentityService
	Payout       *PayoutService
	Collect      *CollectService
	Transactions *TransactionsService
}

// Option configures a Client at construction time.
type Option func(*Client)

// WithBaseURL overrides the API base URL. Pass BaseURLStaging while integrating.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(url, "/") }
}

// WithHTTPClient swaps the underlying *http.Client so you can control timeouts,
// proxies, retries, or tracing.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.httpClient = h
		}
	}
}

// WithUserAgent sets a custom User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Client) {
		if ua != "" {
			c.userAgent = ua
		}
	}
}

// New builds a Client. merchantID is your MER_... id from the dashboard and
// secretKey is your WAYASECK_... key. It defaults to production; pass
// WithBaseURL(BaseURLStaging) to target staging.
func New(merchantID, secretKey string, opts ...Option) *Client {
	c := &Client{
		merchantID: merchantID,
		secretKey:  secretKey,
		baseURL:    BaseURLProduction,
		userAgent:  defaultUA,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
	for _, o := range opts {
		o(c)
	}

	c.Banks = &BanksService{client: c}
	c.Accounts = &AccountsService{client: c}
	c.Identity = &IdentityService{client: c}
	c.Payout = &PayoutService{client: c}
	c.Collect = &CollectService{client: c}
	c.Transactions = &TransactionsService{client: c}
	return c
}
