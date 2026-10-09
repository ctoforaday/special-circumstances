package record

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"google.golang.org/protobuf/proto"
)

func TestRoleOf(t *testing.T) {
	cases := map[string]string{
		// A live seat id carries its AREA, which is the identity a finding is credited to.
		"red-lens-adversary": "adversary",
		"red-lens-evidence":  "evidence",
		"red-lens-dark-side": "dark-side",
		// THE NUMERIC FORM IS NOT READ LIVE. Records that carry it are migrated to the area form
		// (migrate/remap.go, plans/roundless.md §III.A.5); a live seat id shaped that way is no lens.
		"red-lens-L1": "",
		"red-lens-L2": "",
		"red-chair":   "", // no lens role
		"blue-lane-1": "",
		"judge":       "",
	}
	for seat, want := range cases {
		if got := AreaOf(seat); got != want {
			t.Errorf("AreaOf(%q) = %q, want %q", seat, got, want)
		}
	}
}

// A crash-retried finding (same --key) returns the id of the finding it already made, so the verb
// appends no second event — the mint-parity idempotency.
func TestFindingByKeyIsIdempotent(t *testing.T) {
	runDir := newRun(t)
	seat := "red-lens-evidence"

	// No prior finding under this key.
	if got, err := FindingByKey(mustRun(t, runDir), seat, "F1"); err != nil || got != "" {
		t.Fatalf("empty run: got %q, %v; want \"\"", got, err)
	}

	// Record one under key F1 with id F-f0000001.
	f := &recordpb.Finding{Id: proto.String("F-f0000001"), FindingKey: proto.String("F1"), Text: proto.String("x")}
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: seat}, f); err != nil {
		t.Fatal(err)
	}
	if got, _ := FindingByKey(mustRun(t, runDir), seat, "F1"); got != "F-f0000001" {
		t.Errorf("retry lookup = %q, want F-f0000001", got)
	}
	// A different key on the same seat, and the same key on a different seat, do not match.
	if got, _ := FindingByKey(mustRun(t, runDir), seat, "F2"); got != "" {
		t.Errorf("unrelated key matched: %q", got)
	}
	if got, _ := FindingByKey(mustRun(t, runDir), "red-lens-adversary", "F1"); got != "" {
		t.Errorf("another seat's key matched: %q", got)
	}
	// No key names no finding: a finding recorded without one is never the answer to a keyless retry.
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: seat}, &recordpb.Finding{Id: proto.String("F-f0000002"), Text: proto.String("y")}); err != nil {
		t.Fatal(err)
	}
	if got, err := FindingByKey(mustRun(t, runDir), seat, ""); err != nil || got != "" {
		t.Errorf("the empty key matched: (%q, %v)", got, err)
	}
}

// A FINDING HAS ONE NAME, AND THE VIEWS SAY WHO FILED IT BESIDE IT. The id says nothing of the lens,
// so the findings view and the board each carry the id and the area as two fields, a finding with
// no id is refused at the write, and text names one as `<id> (<area>)`.
func TestAFindingIsNamedByItsIDWithItsAreaBeside(t *testing.T) {
	runDir := newRun(t)
	run := mustRun(t, runDir)
	if _, err := Append(Identity{Run: run, SeatID: "red-lens-dark-side"}, &recordpb.Finding{Text: proto.String("x")}); err == nil || !strings.Contains(err.Error(), "finding requires --id") || !strings.Contains(err.Error(), "can never be credited") {
		t.Fatalf("a finding with no id = %v, want the refusal — nothing can credit it", err)
	}
	id := NewID("finding")
	if _, err := Append(Identity{Run: run, SeatID: "red-lens-dark-side"}, &recordpb.Finding{Id: proto.String(id), Text: proto.String("x")}); err != nil {
		t.Fatal(err)
	}

	out, err := FindingsJSONBytes(run)
	if err != nil {
		t.Fatal(err)
	}
	var fj struct {
		Findings []map[string]any `json:"findings"`
	}
	if err := json.Unmarshal(out, &fj); err != nil {
		t.Fatal(err)
	}
	if len(fj.Findings) != 1 {
		t.Fatalf("show findings holds %d findings, want 1: %s", len(fj.Findings), out)
	}
	f := fj.Findings[0]
	if f["id"] != id || f["area"] != "dark-side" || f["seat_id"] != "red-lens-dark-side" {
		t.Errorf("show findings = id %v, area %v, seat_id %v; want %s, dark-side, red-lens-dark-side", f["id"], f["area"], f["seat_id"], id)
	}
	for _, gone := range []string{"label", "anchor", "role", "finding_id"} {
		if _, has := f[gone]; has {
			t.Errorf("show findings carries %q beside the id — a finding has one name", gone)
		}
	}

	obs := mustBoardJSONT(t, run).Observations
	if len(obs) != 1 || obs[0].ID != id || obs[0].Area != "dark-side" || obs[0].SeatID != "red-lens-dark-side" {
		t.Errorf("the board's observations = %+v, want the one finding %s filed by dark-side", obs, id)
	}
	if got, want := FindingRef(id, obs[0].Area), id+" (dark-side)"; got != want {
		t.Errorf("FindingRef = %q, want %q", got, want)
	}
}
