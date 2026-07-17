package wayaquick

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	// BaseURLProduction is the live environment base URL.
	BaseURLProduction = "https://services.wayapay.ng/merchant-middleware/api/v2"

	defaultTimeout    = 30 * time.Second
	defaultUA         = "wayaquick-go/1.0"
	defaultMaxRetries = 2
)

// Client is the WayaQuick Merchant API v2 client. Build it once with New and
// reuse it across goroutines; it is safe for concurrent use.
type Client struct {
	merchantID    string
	secretKey     string
	webhookSecret string
	baseURL       string
	userAgent     string
	maxRetries    int
	httpClient    *http.Client

	Identity *IdentityService
	Payout   *PayoutService
	Collect  *CollectService
	Webhooks *WebhookService
}

// Option configures a Client at construction time.
type Option func(*Client)

// WithBaseURL overrides the API base URL. Defaults to BaseURLProduction.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		if url != "" {
			c.baseURL = strings.TrimRight(url, "/")
		}
	}
}

// WithHTTPClient swaps the underlying *http.Client so you can control timeouts,
// proxies, custom transports, or tracing.
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

// WithWebhookSecret configures the secret used by client.Webhooks to verify
// incoming webhook signatures. It is your merchantSecretTestKey for a TEST
// transaction or your merchantProductionSecretKey for a PRODUCTION one. For
// per-environment routing use the package-level ConstructEvent / VerifySignature
// with an explicit secret instead.
func WithWebhookSecret(secret string) Option {
	return func(c *Client) {
		c.webhookSecret = secret
	}
}

// WithMaxRetries sets how many times a GET is retried on a transient failure
// (network error, timeout, HTTP 429, or 5xx). Writes are never auto retried.
// The default is 2; pass 0 to disable retries entirely.
func WithMaxRetries(n int) Option {
	return func(c *Client) {
		if n >= 0 {
			c.maxRetries = n
		}
	}
}

// New builds a Client. merchantID is your MER_... id from the dashboard and
// secretKey is your WAYASECK_... key. It targets the production base URL; pass
// WithBaseURL to override it.
func New(merchantID, secretKey string, opts ...Option) *Client {
	c := &Client{
		merchantID: merchantID,
		secretKey:  secretKey,
		baseURL:    BaseURLProduction,
		userAgent:  defaultUA,
		maxRetries: defaultMaxRetries,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
	for _, o := range opts {
		o(c)
	}

	c.Identity = &IdentityService{client: c}
	c.Payout = &PayoutService{client: c}
	c.Collect = &CollectService{client: c}
	c.Webhooks = &WebhookService{client: c}
	return c
}

// GenerateReference produces a timestamped, collision resistant reference you
// can use as an idempotency and reconciliation key. The shape is
// "<prefix>-<unixMillis>-<hex>", e.g. "PAYOUT-1748160000000-A1B2C3D4".
//
// Generate one fresh reference per logical operation and reuse the same one on
// retries so the gateway can fold a retry into the original record.
func GenerateReference(prefix string) string {
	if prefix == "" {
		prefix = "WP"
	}
	ms := time.Now().UTC().UnixMilli()

	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand failing is effectively impossible on supported platforms;
		// fall back to the millisecond clock so we still return a usable value.
		return fmt.Sprintf("%s-%d-%08X", prefix, ms, ms&0xFFFFFFFF)
	}
	return fmt.Sprintf("%s-%d-%s", prefix, ms, strings.ToUpper(hex.EncodeToString(b[:])))
}
