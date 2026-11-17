package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE BENCH'S DISPOSITION IS A PROPOSED ACT AND ITS ANSWER. A seat files the act it wants on a gap
// (`motion docket file regrade|close|supersede`) once the gap has had one exchange; the bench, in the
// docket sitting the dispatch convenes for it, accepts, rejects or remands it. These helpers let the
// fixtures below read as one act: the exchange, the filing, the convening, the ruling.

// exchangeOn gives gap its one counted exchange, the cheapest the record counts: the chair engages
// blue on it, and blue sits and sits again having recorded nothing on it — a null turn, which ends
// the exchange and stalls it.
func exchangeOn(t *testing.T, runDir, gap string) {
	t.Helper()
	ensureCast(t, runDir)
	// A handle (G1) names the id it resolves to in this run, as runAt resolves a call's.
	gap = handleText(t, runDir, gap)
	r := runtest.Open(t, runDir)
	if _, err := record.Append(record.Identity{Run: r, SeatID: "red-chair"},
		&recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("blue-respond"), GapIds: []string{gap}}); err != nil {
		t.Fatalf("engaging blue on %s: %v", gap, err)
	}
	for range 2 {
		if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-respond"); err != nil {
			t.Fatalf("blue's sitting on %s: %v", gap, err)
		}
	}
}

// docketFile proposes a close of gap, as filer, after its one exchange, and returns the motion id the
// tool minted: class is the disposition the close proposes.
//
// THE ID IS READ FROM A FIELD, not from the sentence. `--json` puts `motion_id` on the result
// envelope; recovering "M1" from "motion M1 filed (docket)" by string shape would be a fact
// carried in prose, and a reword would return the empty string as confidently as a real id.
func docketFile(t *testing.T, runDir, filer, gapID, basis string) string {
	t.Helper()
	exchangeOn(t, runDir, gapID)
	return proposeClose(t, runDir, filer, gapID, "not_a_defect", basis)
}

// proposeClose files a close proposal and returns its motion id.
func proposeClose(t *testing.T, runDir, filer, gapID, class, basis string) string {
	t.Helper()
	args := []string{"motion", "docket", "file", "close", "--run", runDir, "--seat-id", filer,
		"--id", gapID, "--as", class, "--verified-by", filer, "--verified-with", "the report read",
		"--verified-against", "the report's sentence", "--reason", basis, "--json"}
	out, err := runAt(t, args...)
	// A GAP THAT MOVED SINCE THE FILER'S SITTING OPENED refuses the filing once, and the refusal moves
	// the filer's read point past what it shows: the same filing lands when it is made again.
	if err != nil && strings.Contains(err.Error()+out, "changed after your sitting opened") {
		out, err = runAt(t, args...)
	}
	if err != nil {
		t.Fatalf("motion docket file close %s: %v", gapID, err)
	}
	return filedMotionID(t, out)
}

// filedMotionID is the motion id a filing's --json envelope carries.
func filedMotionID(t *testing.T, out string) string {
	t.Helper()
	var env struct {
		OK     bool `json:"ok"`
		Result struct {
			MotionID string `json:"motion_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("the filing's envelope is not JSON (%v): %s", err, out)
	}
	if !env.OK || env.Result.MotionID == "" {
		t.Fatalf("the filing minted no motion id — every later ruling joins on it: %s", out)
	}
	return env.Result.MotionID
}

// conveneBench readies the bench on gaps as `dispatch next` does — a bench row naming them, written
// by the chair — and opens the docket sitting that row convenes, which holds every docket motion on
// those gaps filed before it.
func conveneBench(t *testing.T, runDir string, gaps ...string) {
	t.Helper()
	ensureCast(t, runDir)
	ids := make([]string, len(gaps))
	for i, g := range gaps {
		ids[i] = handleText(t, runDir, g)
	}
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: "red-chair"},
		&recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("judge"), GapIds: ids}); err != nil {
		t.Fatalf("convening the bench on %v: %v", gaps, err)
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "judge", "--occasion", "docket"); err != nil {
		t.Fatalf("the bench's docket sitting: %v", err)
	}
}

// benchRuleArgs is the ruling half's invocation, less the run and the seat. `as` is the
// disposition; the rest is the reasoning a bench owes with it. A remand states its direction; every
// other disposition says nothing would reopen the gap.
func benchRuleArgs(motionID, as, principle string) []string {
	reopen := []string{"--final"}
	if as == "remanded" {
		reopen = []string{"--direction", "a reproduction on the shipped binary"}
	}
	return append([]string{"motion", "docket", "rule", "--id", motionID, "--as", as,
		"--principle", principle,
		"--tension", "thoroughness against ceremony",
		"--review-flag", "no — the ruling is mechanical",
		"--settled", "the proposition this ruling bars",
		"--reason", "the ruling as reasoned"}, reopen...)
}

// benchDisposes is the whole act: the gap's exchange, the chair's proposal of a close with the
// disposition as (a remand is ruled on a not_a_defect proposal), the bench convened on it, and its
// ruling — accepting the close it proposes, or remanding it.
func benchDisposes(t *testing.T, runDir, gapID, as, principle string) {
	t.Helper()
	class := as
	if as == "remanded" {
		class = "not_a_defect"
	}
	exchangeOn(t, runDir, gapID)
	id := proposeClose(t, runDir, "red-chair", gapID, class, "contested, and not mine to close")
	conveneBench(t, runDir, gapID)
	args := append([]string{}, benchRuleArgs(id, as, principle)...)
	args = append([]string{args[0], args[1], args[2], "--run", runDir, "--seat-id", "judge"}, args[3:]...)
	if _, err := runAt(t, args...); err != nil {
		t.Fatalf("motion docket rule %s on %s: %v", as, gapID, err)
	}
}

// dispatchOf is the chair's dispatch row engaging seat on gaps.
func dispatchOf(seat string, gaps ...string) *recordpb.Dispatch {
	return &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String(seat), GapIds: gaps}
}

// ensureCast seeds the cast setup writes (seedCast) on a record that holds none: a dispatch row is
// admitted only against a cast.
func ensureCast(t *testing.T, runDir string) {
	t.Helper()
	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range m.Events {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_CAST {
			return
		}
	}
	seedCast(t, runDir)
}
