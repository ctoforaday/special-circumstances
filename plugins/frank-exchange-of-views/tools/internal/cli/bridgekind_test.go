package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// SUPPORTS_WITH_BRIDGE NAMES ITS BRIDGE, through the verb a lens types: refused without --bridge,
// recorded with it, refused beside any other outcome, and shown on the evidence row and the
// verb's own line. Deleting the flag's read in writeVerify fails the second case; deleting the
// write path's check fails the first.
func TestVerifyNamesWhatItBridged(t *testing.T) {
	runDir, anchor := citedClause(t)
	if out, err := verifyAs(t, runDir, anchor, "supports_with_bridge"); !refusedSaying(out, err, "supports_with_bridge requires --bridge") {
		t.Fatalf("supports_with_bridge with no --bridge: err=%v\n%s\nwant the refusal naming --bridge", err, out)
	}
	if out, err := run(t, "verify", "--run", runDir, "--seat-id", lensSeat, "--anchor", anchor,
		"--as", "supports", "--bridge", "summary", "--confidence", "high", "--reason", "read at the leaf"); !refusedSaying(out, err, "drop --bridge") {
		t.Fatalf("--bridge beside supports: err=%v\n%s\nwant it refused", err, out)
	}
	if out, err := run(t, "verify", "--run", runDir, "--seat-id", lensSeat, "--anchor", anchor,
		"--as", "supports_with_bridge", "--bridge", "nearly", "--confidence", "medium", "--reason", "read at the leaf"); err == nil {
		t.Fatalf("--bridge nearly was admitted:\n%s", out)
	}
	out, err := run(t, "verify", "--run", runDir, "--seat-id", lensSeat, "--anchor", anchor,
		"--as", "supports_with_bridge", "--bridge", "near_restatement", "--confidence", "medium", "--reason", "read at the leaf")
	if err != nil {
		t.Fatalf("verify with --bridge: %v\n%s", err, out)
	}
	if !strings.Contains(out, "bridged: near_restatement") {
		t.Errorf("verify's own line does not say what was bridged:\n%s", out)
	}
	if got := firstVerify(t, runDir).GetBridge(); got != recordpb.BridgeKind_BRIDGE_KIND_NEAR_RESTATEMENT {
		t.Errorf("recorded bridge = %v, want near_restatement", got)
	}
	ev, err := run(t, "show", "evidence", "--run", runDir, "--seat-id", lensSeat, "--json")
	if err != nil {
		t.Fatalf("show evidence: %v\n%s", err, ev)
	}
	var got struct {
		Sources []struct {
			Verified []struct {
				Outcome string `json:"outcome"`
				Bridge  string `json:"bridge"`
			} `json:"verified"`
		} `json:"sources"`
	}
	if err := json.Unmarshal([]byte(ev), &got); err != nil {
		t.Fatalf("show evidence: %v\n%s", err, ev)
	}
	if len(got.Sources) != 1 || len(got.Sources[0].Verified) != 1 || got.Sources[0].Verified[0].Bridge != "near_restatement" {
		t.Errorf("the evidence row does not carry the bridge: %+v", got)
	}
}

// corroborate SHARES THE AXIS: its supports_with_bridge needs a bridge too.
func TestCorroborateNamesWhatItBridged(t *testing.T) {
	runDir, _ := citedClause(t)
	out, err := run(t, "corroborate", "--run", runDir, "--seat-id", lensSeat, "--url", "https://sky/v", "--title", "Sky",
		"--quote", verifiedSentence, "--as", "supports_with_bridge", "--confidence", "medium", "--reason", "read at the leaf")
	if !refusedSaying(out, err, "supports_with_bridge requires --bridge") {
		t.Fatalf("corroborate supports_with_bridge with no --bridge: err=%v\n%s", err, out)
	}
}

// refusedSaying reports whether a verb was refused with want in what it printed or returned.
func refusedSaying(out string, err error, want string) bool {
	return err != nil && strings.Contains(out+err.Error(), want)
}
