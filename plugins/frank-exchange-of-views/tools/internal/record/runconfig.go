package record

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// RunConfig is inputs/run-config.json: the ONE type setup marshals and every reader in the record
// decodes, through readRunConfig. One type, so a key setup writes is a key the record declares by
// construction rather than by a second hand-kept copy.
//
// THE FILE IS READ STRICTLY: anything but one JSON object, a key RunConfig does not declare, or
// content after the object, is an error, never ignored. A term a reader skipped would read exactly
// like a term the run was never given, so a key no field declares — a term the engine no longer
// has, or a misspelt one — refuses the read rather than vanishing.
//
// A VALUE IS TYPED HERE ONLY WHERE THE RECORD OWNS ITS SHAPE. The record's epoch gate
// (EventSchema, feov-migrate) is what keeps an archived run readable when a shape moves, and it
// covers only shapes the record declares. A value whose shape another package owns — `hooks`,
// hookprovenance.Provenance, which moves with no epoch bump — is held as raw JSON and decoded by its
// owner (hookprovenance.Decode): typed here, a change in that package would refuse every archived
// run's configuration with no migration able to bring it forward. The TOP-LEVEL key set stays
// strict either way; `billing` is BillingIdentity, which the record declares, so it is typed.
type RunConfig struct {
	Topic string `json:"topic"`
	// RunDir is the ABSOLUTE path of this run.
	//
	// It is recorded because the run directory reaches a seat as a STRING each seat resolves
	// against its own working directory, and a relative one resolves differently per invocation.
	// Measured (#358): a seat whose shell cwd was the `tools/` directory resolved
	// `research/<slug>/` from there and built a second blackboard — the lane's entire draft, its
	// own shards, clock and locks — while the real run's candidates directory stayed empty. Two
	// shards of one seat class existed in both places.
	//
	// A seat can no longer CREATE a run directory (RegisterSeat refuses one that does not exist),
	// which turns that failure loud. This field is the other half: the operator and every
	// post-hoc reader can see which absolute path the run was set up at, rather than inferring it
	// from wherever they happen to be standing.
	RunDir        string `json:"runDir"`
	Model         string `json:"model"`
	JudgmentModel string `json:"judgmentModel"`
	// Lanes is the `--lanes` flag as typed, nil where the operator named none (LanesDeclared).
	Lanes *string `json:"lanes"`
	// The run's terms (Params says what each bounds). Setup always writes them; each is nil where the
	// file does not state it, which RunParams reads as setup's default.
	K                   *int     `json:"k"`
	KMax                *int     `json:"kMax"`
	ConvergenceFraction *float64 `json:"convergenceFraction"`
	MaxEpochs           *int     `json:"maxEpochs"`
	// MaxSittingCalls is read by the PreToolUse hook through sittingcap.Limit, not by the record:
	// the hook may not link the record. The key is sittingcap.ConfigKey.
	MaxSittingCalls int `json:"maxSittingCalls"`
	// Hooks is what was INSTALLED on the hook side when this run was set up (#751). The record
	// binary is baked from the working tree and the hooks come from the version-gated install
	// cache, so the two can be days apart — and a run whose hooks were stale is otherwise
	// byte-identical on the record to one whose hooks were current.
	//
	// It says what was installed, never what RAN: a cache entry on disk does not prove the
	// harness invoked it. The liveness half is #555's. The key is hookprovenance.ConfigKey, and
	// run.md reads it back through hookprovenance.Decode. RAW, because hookprovenance owns its shape
	// (see the type's comment): setup marshals hookprovenance.Provenance into it.
	Hooks json.RawMessage `json:"hooks"`
	// EventSchema is the event-shape EPOCH this run's events were written under.
	//
	// A run directory is created by `setup` and does not outlive the schema that made it — but
	// `run-archive/` does, and is re-read by every later audit. When the record's storage moved to
	// a database, six archived runs became unreadable by the current binary, and nothing in them
	// said what shape they were in — so a later reader could only guess which binary to build.
	//
	// It is the EPOCH and deliberately not a binary version (#597): a release number moves for
	// reasons the event shape does not care about, so recording one would tell a future reader
	// what shipped rather than what it can read. This is the same number the setup preflight
	// just checked the binary against, so the file states a fact the run has already verified.
	EventSchema int `json:"eventSchema,omitempty"`
	// AllowModelSubstitution records that the OPERATOR accepted, before the run, an environment
	// that may answer with a model other than the configured tier. Absent means no: a run whose
	// config does not state it never consented, and the failing direction of a gate has to be the
	// safe one.
	//
	// It is written by setup and read by SeatModels, which marks a substitution it allows, which is
	// what makes the consent a FIELD an operator set rather than a flag a seat could type for itself.
	AllowModelSubstitution bool `json:"allowModelSubstitution,omitempty"`
	// Billing is WHAT THIS RUN IS BILLED TO — the account, its organization and whether it is a
	// subscription. See BillingIdentity for the measurement that put it here.
	Billing BillingIdentity `json:"billing"`
}

// BillingIdentity is the account a run is billed to, as setup read it from the files the CLI
// authenticates with (setup.ReadBillingIdentity).
//
// It exists because a run had no way to say which bill it was on: an isolated config silently
// authenticated against a different organization and spent credits for months while reporting a
// dollar figure that read as telemetry.
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

// Subscription reports whether this run is on a subscription rather than billed per call. A run
// that could not read its identity answers false: the safe direction is to say "not known to be a
// subscription" rather than to assume the cheaper one.
func (b BillingIdentity) Subscription() bool {
	return b.BillingType == "stripe_subscription" || b.SubscriptionType != ""
}

// ReadRunConfig is the configuration of run, read strictly (readRunConfig): nil and no error for a
// run with no run-config.json, an error for one that is there and does not read. It is the reader
// for a caller outside the record that shows or checks the run's configuration.
func ReadRunConfig(run Run) (*RunConfig, error) { return readRunConfig(run.Dir()) }

// readRunConfig reads the configuration of the run in directory dir; a run with no run-config.json
// has none (nil, nil). A file that is there and cannot be read, holds anything but one JSON object,
// holds a key RunConfig does not declare, or holds anything after its object, is an error, never "no
// configuration": the terms it holds are what the run's limits are checked against. feov-migrate's
// copy of the file (translateRunConfig) refuses the same files.
func readRunConfig(dir string) (*RunConfig, error) {
	b, err := os.ReadFile(filepath.Join(dir, "inputs", "run-config.json"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("inputs/run-config.json could not be read: %w", err)
	}
	rc, err := DecodeRunConfig(b)
	if err != nil {
		return nil, fmt.Errorf("inputs/run-config.json %w", err)
	}
	return rc, nil
}

// DecodeRunConfig is the one reading of run-config.json's bytes: one object, nothing after it, and
// no key RunConfig does not declare. readRunConfig reads a run's file through it, and migrate holds
// the file it writes to it, so the two cannot disagree on what a configuration is.
func DecodeRunConfig(b []byte) (*RunConfig, error) {
	raw, err := RunConfigObject(b)
	if err != nil {
		return nil, err
	}
	var rc RunConfig
	strict := json.NewDecoder(bytes.NewReader(raw))
	strict.DisallowUnknownFields()
	if err := strict.Decode(&rc); err != nil {
		return nil, fmt.Errorf("does not parse as the run's configuration: %w", err)
	}
	return &rc, nil
}

// RunConfigObject is b's one JSON object, refused when b holds anything else: migrate walks an
// archived file's keys from it before the keys the record no longer declares are dropped.
func RunConfigObject(b []byte) (json.RawMessage, error) {
	dec := json.NewDecoder(bytes.NewReader(b))
	var raw json.RawMessage
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("does not parse as the run's configuration: %w", err)
	}
	// ONE OBJECT AND NOTHING AFTER IT. Decode stops at the end of the first value, so a second
	// object — a concatenation, a half-applied rewrite — would otherwise be read past, and its terms
	// with it.
	if tok, err := dec.Token(); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("found %v", tok)
		}
		return nil, fmt.Errorf("holds content after the run's configuration: %w", err)
	}
	// AN OBJECT, NOT ANY VALUE. `null` decodes into a struct as no terms at all, without an error,
	// and would read exactly like a run that stated none.
	if raw[0] != '{' {
		return nil, fmt.Errorf("is not a JSON object: it holds %s", jsonKind(raw))
	}
	return raw, nil
}

// jsonKind names the kind of the JSON value raw holds, for a refusal: "null", "an array", "a
// string", "a number", "a boolean" or "an object". raw is one valid value, as a decoder returns it.
func jsonKind(raw []byte) string {
	switch raw[0] {
	case '{':
		return "an object"
	case '[':
		return "an array"
	case '"':
		return "a string"
	case 'n':
		return "null"
	case 't', 'f':
		return "a boolean"
	default:
		return "a number"
	}
}
