package capture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// seedClosures writes a mint+close pair per closure into the record, so the attestation audit
// reads the fields it reconciles rather than a markdown line it has to split apart.
//
// ONE shard for the seat, not one per closure: a second nonce for the same seat is a RETRY in
// this record's semantics, and replay keeps a single shard per seat — seeding one file per
// event silently drops all but one closure.
func seedClosures(t *testing.T, runDir string, cl []Claim, carried bool) {
	t.Helper()
	recs := filepath.Join(runDir, "records")
	if err := os.MkdirAll(recs, 0o755); err != nil {
		t.Fatal(err)
	}
	const seat, nonce = "red-merge-r1", "10000000"
	var lines []string
	for i, c := range cl {
		closure := record.NewPayload().Set("gap_id", c.ID).Set("closure_class", "closed").Set("prose", "verified")
		if carried {
			closure.Set("carried_from", "1")
		} else {
			closure.Set("anchor_seat", c.Seat).Set("anchor_tool", c.Tool).Set("anchor_target", c.Target)
		}
		for _, e := range []record.Event{
			{Seq: i * 2, SeatID: seat, Nonce: nonce, Round: 1, Type: "mint", Key: seat + ":mint:" + c.ID,
				Payload: record.NewPayload().Set("gap_id", c.ID).Set("problem", "p").Set("acceptance_check", "c").
					Set("check_kind", "document").Set("class", "x").Set("likelihood", "medium").Set("impact", "medium")},
			{Seq: i*2 + 1, SeatID: seat, Nonce: nonce, Round: 1, Type: "close", Key: seat + ":close:" + c.ID, Payload: closure},
		} {
			b, err := record.MarshalEvent(e)
			if err != nil {
				t.Fatal(err)
			}
			lines = append(lines, string(b))
		}
	}
	write(t, filepath.Join(recs, "events-"+seat+"-"+nonce+".jsonl"), strings.Join(lines, "\n")+"\n")
}

// THE AUDIT READS THE RECORD.
//
// It used to read red/archive.md — a file `setup` stopped creating when the archive became a
// projection — so it returned SKIP "no archive records to reconcile" on every record-mode run,
// reconciling nothing in silence for as long as the record tier has existed.
//
// The old path was the class in miniature: a close event carries anchor_seat, anchor_tool and
// anchor_target as FIELDS; archiveMD concatenates them into a pipe-delimited line; and the
// audit split that line back apart and re-derived the gap id by regex.
func TestAttestationAudit(t *testing.T) {
	run, tr := t.TempDir(), t.TempDir()
	seedClosures(t, run, []Claim{
		{ID: "R1-1", Seat: "L1", Tool: "Read", Target: "report.md#SectionTwo"},
		{ID: "R1-2", Seat: "L5", Tool: "git show", Target: "7bc501e:plans/record-tool.md"},
	}, false)
	write(t, filepath.Join(tr, "agent-a.jsonl"),
		`{"message":{"role":"assistant","content":[{"type":"tool_use","name":"Read","input":{"file_path":"report.md#SectionTwo"}}]}}`+"\n")

	got := AttestationAudit(run, tr, []string{"agent-a.jsonl"}, 5)
	if got.Verdict != "FAIL" {
		t.Errorf("a claimed act with no matching tool call: want FAIL, got %s (%s)", got.Verdict, got.Detail)
	}
	if len(got.Unreconciled) != 1 || got.Unreconciled[0].ID != "R1-2" {
		t.Errorf("the unsupported claim R1-2 must be the one named: %+v", got.Unreconciled)
	}
	if !strings.Contains(got.Detail, "RECORD IS HONEST, never on the merits") {
		t.Errorf("the separation must be stated in the finding: %s", got.Detail)
	}
}

// A CARRIED closure asserts no fresh act, so there is nothing to reconcile.
func TestCarriedClosuresAreNotAttested(t *testing.T) {
	run, tr := t.TempDir(), t.TempDir()
	seedClosures(t, run, []Claim{{ID: "R2-1"}}, true)
	write(t, filepath.Join(tr, "agent-a.jsonl"), `{"message":{"role":"assistant"}}`+"\n")
	if got := AttestationAudit(run, tr, []string{"agent-a.jsonl"}, 5).Verdict; got != "SKIP" {
		t.Errorf("all-carried closures: want SKIP, got %s", got)
	}
}

// An empty record SKIPs rather than reporting a clean reconciliation of nothing — which is
// exactly how the markdown version failed for as long as it did.
func TestNoClosuresSkips(t *testing.T) {
	tr := t.TempDir()
	write(t, filepath.Join(tr, "agent-a.jsonl"), `{"message":{"role":"assistant"}}`+"\n")
	if got := AttestationAudit(t.TempDir(), tr, []string{"agent-a.jsonl"}, 5).Verdict; got != "SKIP" {
		t.Errorf("empty record: want SKIP, got %s", got)
	}
}
