//go:build live

// Live integration tests. These hit the real WayaQuick API and are excluded from
// the default test run. Enable them with the "live" build tag and real creds:
//
//	export WAYA_MERCHANT_ID=MER_...
//	export WAYA_SECRET_KEY=WAYASECK_TEST_...
//	go test -tags live ./tests/...
//
// They run against the production API; use test credentials.
package wayaquick_test

import (
	"context"
	"os"
	"testing"
	"time"

	wayaquick "github.com/WAYA-MULTI-LINK/WAYAPAYCHAT-1.0-GOLANG-DOCUMENTATION/src/wayaquick"
)

func liveClient(t *testing.T) *wayaquick.Client {
	t.Helper()
	merchant := os.Getenv("WAYA_MERCHANT_ID")
	secret := os.Getenv("WAYA_SECRET_KEY")
	if merchant == "" || secret == "" {
		t.Skip("set WAYA_MERCHANT_ID and WAYA_SECRET_KEY to run live tests")
	}

	return wayaquick.New(merchant, secret)
}

func TestLive_BanksList(t *testing.T) {
	c := liveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	banks, err := c.Payout.ListBanks(ctx)
	if err != nil {
		t.Fatalf("bank list: %v", err)
	}
	if len(banks) == 0 {
		t.Fatal("expected at least one bank")
	}
	t.Logf("got %d banks; first: %s (%s)", len(banks), banks[0].Name, banks[0].Code)
}

func TestLive_VerifyAccount(t *testing.T) {
	c := liveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	acct, err := c.Payout.VerifyAccount(ctx, wayaquick.VerifyAccountRequest{
		AccountNumber: os.Getenv("WAYA_TEST_ACCOUNT"),
		BankCode:      os.Getenv("WAYA_TEST_BANK_CODE"),
		EnquiryType:   wayaquick.EnquiryTypeOthers,
	})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	t.Logf("resolved: %s", acct.AccountName)
}
