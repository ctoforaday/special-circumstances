package checkpointseal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/statefile"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/stopnudge"
)

// The three states nudge.json can be in, and the three DIFFERENT rows they must produce.
//
// Criterion 6 compares the nudge-on population against the nudge-off one, and criterion 4
// reads the counters. A row that reports "off" when it means "could not tell" files an
// unreadable state into the control group, which is a manufactured baseline — the same
// defect this design removed from live_handles, note_age_turns and growth in turn.
func TestTheRowDistinguishesNudgeOffFromNudgeUnreadable(t *testing.T) {
	for _, tc := range []struct {
		name          string
		write         func(t *testing.T, root string)
		wantMeasured  bool
		wantEnabled   bool
		wantCounters  bool
		wantEmissions int
	}{
		{
			name:         "absent is an honest off",
			write:        func(*testing.T, string) {},
			wantMeasured: true, wantEnabled: false, wantCounters: false,
		},
		{
			name: "present carries the counters criterion 4 is decided by",
			write: func(t *testing.T, root string) {
				if err := statefile.Write(stopnudge.StatePath(root), stopnudge.State{
					SessionID: "s1", Emissions: 3, EmissionBytes: 174,
				}); err != nil {
					t.Fatal(err)
				}
			},
			wantMeasured: true, wantEnabled: true, wantCounters: true, wantEmissions: 3,
		},
		{
			name: "unreadable is NOT off",
			write: func(t *testing.T, root string) {
				p := stopnudge.StatePath(root)
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p, []byte("{truncated"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			wantMeasured: false, wantEnabled: false, wantCounters: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			tc.write(t, root)

			var row sealRow
			readNudgeState(root, &row)

			if row.NudgeMeasured != tc.wantMeasured {
				t.Errorf("nudge_measured = %v, want %v", row.NudgeMeasured, tc.wantMeasured)
			}
			if row.NudgeEnabled != tc.wantEnabled {
				t.Errorf("nudge_enabled = %v, want %v", row.NudgeEnabled, tc.wantEnabled)
			}
			if got := row.EmissionsThisSession != nil; got != tc.wantCounters {
				t.Errorf("counters present = %v, want %v", got, tc.wantCounters)
			}
			if tc.wantCounters && *row.EmissionsThisSession != tc.wantEmissions {
				t.Errorf("emissions_this_session = %d, want %d", *row.EmissionsThisSession, tc.wantEmissions)
			}
		})
	}
}

// A ZERO EMISSION COUNT IS A REAL READING and must survive to the record. It is what a
// well-behaved session looks like, so omitempty on a plain int would have deleted exactly
// the rows criterion 4 most wants to count — leaving the query to average only the
// sessions that emitted.
func TestAZeroEmissionCountIsWrittenRatherThanOmitted(t *testing.T) {
	root := t.TempDir()
	if err := statefile.Write(stopnudge.StatePath(root), stopnudge.State{SessionID: "s1"}); err != nil {
		t.Fatal(err)
	}
	var row sealRow
	readNudgeState(root, &row)

	b, err := json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	v, ok := m["emissions_this_session"]
	if !ok {
		t.Fatal("a measured zero was omitted from the row; the query cannot tell it from an unmeasured session")
	}
	if v != float64(0) {
		t.Errorf("emissions_this_session = %v, want 0", v)
	}
}
