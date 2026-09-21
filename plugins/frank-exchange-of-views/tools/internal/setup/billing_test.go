package setup

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeCfg(t *testing.T, dir, creds, acct string) {
	t.Helper()
	if creds != "" {
		if err := os.WriteFile(filepath.Join(dir, ".credentials.json"), []byte(creds), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if acct != "" {
		if err := os.WriteFile(filepath.Join(dir, ".claude.json"), []byte(acct), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// THE RUN RECORDS WHICH BILL IT IS ON, and no part of the credential comes with it.
func TestBillingIdentityRecordsThePlanAndNoSecret(t *testing.T) {
	dir := t.TempDir()
	writeCfg(t, dir,
		`{"claudeAiOauth":{"accessToken":"SECRET-TOKEN-VALUE","refreshToken":"SECRET-REFRESH","subscriptionType":"max","rateLimitTier":"default_claude_max_20x"}}`,
		`{"oauthAccount":{"accountUuid":"acct-1","organizationUuid":"org-1","billingType":"stripe_subscription"}}`)
	t.Setenv("CLAUDE_CONFIG_DIR", dir)

	b := ReadBillingIdentity()
	if !b.Subscription() {
		t.Errorf("a stripe_subscription account does not read as a subscription: %+v", b)
	}
	for _, want := range []string{b.SubscriptionType, b.RateLimitTier, b.Organization, b.BillingType} {
		if want == "" {
			t.Errorf("a field the console page is found by is empty: %+v", b)
		}
	}
	// THE CREDENTIAL MUST NOT TRAVEL. run-config.json is written into a run directory that is
	// archived and re-read by every later audit, so a token landing here would outlive the
	// container it was issued in.
	blob := b.ConfigDir + b.BillingType + b.SubscriptionType + b.RateLimitTier + b.Organization + b.Account + b.Unreadable
	for _, secret := range []string{"SECRET-TOKEN-VALUE", "SECRET-REFRESH"} {
		if strings.Contains(blob, secret) {
			t.Errorf("the recorded identity carries %q — a credential must never become a record", secret)
		}
	}
}

// AN ISOLATED CONFIG WITH NO CREDENTIAL IS THE DEFECT, and it must be loud rather than absent.
// This is the shape scripts/universe.sh had: plugins isolated correctly, the subscription isolated
// with them, and the CLI quietly authenticating against a console profile billed in credits.
func TestAnIsolatedConfigWithNoSubscriptionSaysSoLoudly(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", dir) // no credential at all

	b := ReadBillingIdentity()
	if b.Subscription() {
		t.Error("a config with no credential reports as a subscription — the safe direction is not knowing")
	}
	if !b.Isolated {
		t.Errorf("an explicit CLAUDE_CONFIG_DIR does not read as isolated: %+v", b)
	}
	if b.Unreadable == "" {
		t.Error("nothing says WHY no account was found, so the absence reads like a blank field")
	}

	var out, errb bytes.Buffer
	reportBilling(b, &out, &errb)
	msg := errb.String()
	for _, want := range []string{"NOT ON A SUBSCRIPTION", "CREDITS", dir} {
		if !strings.Contains(msg, want) {
			t.Errorf("the warning does not say %q, so an operator learns it from the bill instead:\n%s", want, msg)
		}
	}
	if out.Len() > 0 {
		t.Errorf("the not-a-subscription case wrote to stdout, where a log tail may not show it:\n%s", out.String())
	}
}

// AND THE HEALTHY CASE STILL SAYS WHAT IT IS. A run that prints nothing leaves the operator with
// the same question the silence caused.
func TestASubscriptionRunStatesWhatItBillsTo(t *testing.T) {
	dir := t.TempDir()
	writeCfg(t, dir, `{"claudeAiOauth":{"subscriptionType":"max","rateLimitTier":"t"}}`,
		`{"oauthAccount":{"billingType":"stripe_subscription"}}`)
	t.Setenv("CLAUDE_CONFIG_DIR", dir)

	var out, errb bytes.Buffer
	reportBilling(ReadBillingIdentity(), &out, &errb)
	if !strings.Contains(out.String(), "stripe_subscription") {
		t.Errorf("a subscription run does not say what it bills to:\n%s", out.String())
	}
	if errb.Len() > 0 {
		t.Errorf("a healthy run warned anyway, which is how a warning stops being read:\n%s", errb.String())
	}
}
