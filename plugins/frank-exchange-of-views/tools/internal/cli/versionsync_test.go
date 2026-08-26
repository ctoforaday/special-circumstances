package cli

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// THE BINARY AND THE PLUGIN READ THE SAME EPOCH, or setup refuses every binary — including
// the right one.
//
// scripts/schemagen generates record.EventSchema from requirements.json and gates the two in
// CI, so this is the in-module half of that contract: it fails here, at the package that
// ships the constant, rather than only in a gate somebody has to be running.
//
// What it deliberately no longer asserts is a RELEASE number. The predecessor kept a semver
// constant in step with the manifest, which meant every release had to remember to move a
// second number, and the check tracked shipping rather than compatibility. The epoch moves
// only when the event shape does.
func TestTheCompiledEpochMatchesTheManifest(t *testing.T) {
	req, err := repotree.Plugin("requirements.json")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(req)
	if err != nil {
		t.Fatalf("cannot read requirements.json, which is what setup preflights against: %v", err)
	}
	var m struct {
		EventSchema *int `json:"eventSchema"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	// A POINTER, so an ABSENT field cannot read as zero. Absent means the manifest stopped
	// declaring the epoch, and setup would then have nothing to compare the binary against.
	if m.EventSchema == nil {
		t.Fatal("requirements.json declares no eventSchema — setup then has nothing to check the binary against")
	}
	if *m.EventSchema != record.EventSchema {
		t.Errorf("record.EventSchema = %d but requirements.json reads %d.\n"+
			"These are generated together by scripts/schemagen; regenerate with "+
			"`(cd scripts && go run ./schemagen)`. While they disagree, setup refuses every "+
			"binary or none, depending on which one moved.",
			record.EventSchema, *m.EventSchema)
	}
}

// The plugin manifest must NOT carry the epoch as well: .claude-plugin/plugin.json has a
// fixed published schema, so an unknown field there makes `claude plugin validate` warn on
// every run — and a permanently expected warning is how a validator stops being read.
func TestThePluginManifestDoesNotCarryTheEpoch(t *testing.T) {
	manifest, err := repotree.Plugin(".claude-plugin", "plugin.json")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	var p map[string]any
	if err := json.Unmarshal(b, &p); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"eventSchema", "recordToolVersion"} {
		if _, ok := p[k]; ok {
			t.Errorf("%s is in .claude-plugin/plugin.json. It belongs in requirements.json alone: "+
				"plugin.json has a fixed published schema, so an unknown field there makes "+
				"`claude plugin validate` warn on every run.", k)
		}
	}
}
