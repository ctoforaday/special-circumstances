package setup

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claudedir"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// WHAT THIS RUN IS BILLED TO, recorded because the run had no way to say.
//
// run-config.json records the topic, the tiers, the run's terms and which hook binaries were
// installed. It did not record the one fact that decides who PAYS, so "which account is this on"
// was unanswerable from the record and surfaced only as a failure.
//
// MEASURED, and it cost real money. `scripts/universe.sh` isolates CLAUDE_CONFIG_DIR so a smoke
// run cannot install its plugins into the operator's real config — correct — but the credential
// lives in that directory too, so the SUBSCRIPTION was isolated with it and `claude -p` fell
// through to a console profile in a different organization, billed in credits. Every run reports
// a dollar figure whichever way it authenticated, so ~$21 a run read as telemetry rather than as a
// bill for months. It became visible when a run died mid-debate on "Credit balance is too low",
// thirty-five minutes and $7.09 in, with no outcome row. The identity is a field of the run's
// configuration, record.BillingIdentity.

// ReadBillingIdentity resolves the account the run will answer to, from the same files the CLI
// authenticates with. It never fails the setup: a run whose billing cannot be read is a run that
// says so, not one that refuses to start.
func ReadBillingIdentity() record.BillingIdentity {
	b := record.BillingIdentity{}
	home, _ := os.UserHomeDir()
	cfg := claudedir.ConfigDir(os.Getenv, home)
	if os.Getenv("CLAUDE_CONFIG_DIR") != "" {
		b.ConfigDir, b.Isolated = cfg, cfg != claudedir.ConfigDir(func(string) string { return "" }, home)
	}

	// The credential says which PLAN is in force. Only the non-secret fields are read.
	var creds struct {
		OAuth struct {
			SubscriptionType string `json:"subscriptionType"`
			RateLimitTier    string `json:"rateLimitTier"`
		} `json:"claudeAiOauth"`
	}
	if raw, err := os.ReadFile(filepath.Join(cfg, ".credentials.json")); err != nil {
		b.Unreadable = "no credential in the config directory: " + err.Error()
	} else if err := json.Unmarshal(raw, &creds); err != nil {
		b.Unreadable = "the credential could not be parsed: " + err.Error()
	} else {
		b.SubscriptionType, b.RateLimitTier = creds.OAuth.SubscriptionType, creds.OAuth.RateLimitTier
	}

	// The account file says which ORGANIZATION and how it is paid for.
	var acct struct {
		OAuthAccount struct {
			AccountUUID      string `json:"accountUuid"`
			OrganizationUUID string `json:"organizationUuid"`
			BillingType      string `json:"billingType"`
		} `json:"oauthAccount"`
	}
	if raw, err := os.ReadFile(filepath.Join(cfg, ".claude.json")); err == nil {
		if json.Unmarshal(raw, &acct) == nil {
			b.Account = acct.OAuthAccount.AccountUUID
			b.Organization = acct.OAuthAccount.OrganizationUUID
			b.BillingType = acct.OAuthAccount.BillingType
		}
	}
	return b
}
