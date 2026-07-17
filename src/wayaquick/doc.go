// Package wayaquick is a Go client for the WayaQuick Merchant API v2
// (payout, collect, identity, and webhooks).
//
// Construct a client with New. It targets the production base URL; override it
// with WithBaseURL if you need to. Every call takes a context.Context and
// returns a typed struct or an *APIError carrying the API code and message.
//
//	client := wayaquick.New(
//		"MER_lCquc1779095889226CVfl7",
//		"WAYASECK_TEST_0x3a93476a20d347d6847b62665e9ecb4b",
//	)
//
//	banks, err := client.Payout.ListBanks(context.Background())
//
// Poll the state of a deposit or transfer with Collect.Status and
// Payout.Status, then branch on the parsed code's Outcome and IsTerminal (see
// ParseCollectionStatus and ParsePayoutStatus).
//
// Incoming transaction webhooks are verified offline (no network call) with
// ConstructEvent / VerifySignature, or via client.Webhooks once a secret is set
// with WithWebhookSecret. The signature is
// base64(HMAC-SHA256(secret, "{timestamp}.{payload}")); ConstructEvent rejects
// a bad signature or a stale timestamp before returning a *WebhookEvent.
//
// This SDK is built for server side use. Your secret key must never ship in a
// mobile app, browser bundle, or any client the public can crack open; payout
// authorization belongs behind your own backend.
package wayaquick
