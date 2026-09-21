package setup

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// WHAT THIS RUN IS BILLED TO, recorded because the run had no way to say.
//
// run-config.json records the topic, the tiers, the budgets and which hook binaries were
// installed. It did not record the one fact that decides who PAYS, so "which account is this on"
// was unanswerable from the record and surfaced only as a failure.
//
// MEASURED, and it cost real money. `scripts/universe.sh` isolates CLAUDE_CONFIG_DIR so a smoke
// run cannot install its plugins into the operator's real config — correct — but the credential
// lives in that directory too, so the SUBSCRIPTION was isolated with it and `claude -p` fell
// through to a console profile in a different organization, billed in credits. Every run reports
// a dollar figure whichever way it authenticated, so ~$21 a run read as telemetry rather than as a
// bill for months. It became visible when a run died mid-debate on "Credit balance is too low",
// thirty-five minutes and $7.09 in, with no outcome row.
//
// NO SECRET IS RECORDED. Tokens, refresh tokens and their expiries are read past: what goes on the
// record is the ACCOUNT the run answers to — billing type, subscription tier, organization — which
// is what a reader needs and what a credential must never become.
//
// ABSENT IS NOT "the subscription". A config with no account recorded is one whose identity could
// not be read, and it stays legible as unknown rather than defaulting to the answer an operator
// would hope for — the whole defect being that a silent fallback still runs and just bills
// elsewhere.
type BillingIdentity struct {
	// ConfigDir is the CLAUDE_CONFIG_DIR in force, or "" for the operator's default. An isolated
	// one is the case that went wrong, so a reader sees it beside the account rather than having
	// to infer it.
	ConfigDir string `json:"configDir,omitempty"`
	// Isolated says the run used a config directory that is not the operator's default.
	Isolated bool `json:"isolated"`
	// BillingType is what the account is paid for by — `stripe_subscription` for a subscription,
	// absent where no account could be read.
	BillingType string `json:"billingType,omitempty"`
	// SubscriptionType and RateLimitTier come from the credential in force: `max`,
	// `default_claude_max_20x` and so on. Absent means the credential named none, which is the
	// signature of an API-billed profile rather than a subscription.
	SubscriptionType string `json:"subscriptionType,omitempty"`
	RateLimitTier    string `json:"rateLimitTier,omitempty"`
	// Organization and Account are uuids, not names or addresses: enough to tell one bill from
	// another and to match a console page, without putting an operator's identity in an archive
	// that outlives the container.
	Organization string `json:"organization,omitempty"`
	Account      string `json:"account,omitempty"`
	// Unreadable says why nothing could be determined, when nothing could. A blank block and a
	// block that says it could not look are different facts.
	Unreadable string `json:"unreadable,omitempty"`
}

// ReadBillingIdentity resolves the account the run will answer to, from the same files the CLI
// authenticates with. It never fails the setup: a run whose billing cannot be read is a run that
// says so, not one that refuses to start.
func ReadBillingIdentity() BillingIdentity {
	b := BillingIdentity{}
	cfg := os.Getenv("CLAUDE_CONFIG_DIR")
	home, _ := os.UserHomeDir()
	if cfg != "" {
		b.ConfigDir, b.Isolated = cfg, cfg != filepath.Join(home, ".claude")
	} else {
		cfg = filepath.Join(home, ".claude")
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

// Subscription reports whether this run is on a subscription rather than billed per call. A run
// that could not read its identity answers false: the safe direction is to say "not known to be a
// subscription" rather than to assume the cheaper one.
func (b BillingIdentity) Subscription() bool {
	return b.BillingType == "stripe_subscription" || b.SubscriptionType != ""
}
