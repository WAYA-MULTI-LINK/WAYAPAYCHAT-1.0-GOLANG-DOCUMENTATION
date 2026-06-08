// Package wayapay is a Go client for the WayaPay Merchant API v2
// (collect, payout, accounts, identity, and transactions).
//
// Construct a client with New. It targets the production base URL; override it
// with WithBaseURL if you need to. Every call takes a context.Context and
// returns a typed struct or an *APIError carrying the API code and message.
//
//	client := wayapay.New(
//		"MER_lCquc1779095889226CVfl7",
//		"WAYASECK_TEST_0x3a93476a20d347d6847b62665e9ecb4b",
//	)
//
//	banks, err := client.Banks.List(context.Background())
//
// This SDK is built for server side use. Your secret key must never ship in a
// mobile app, browser bundle, or any client the public can crack open; payout
// authorization belongs behind your own backend.
package wayapay
