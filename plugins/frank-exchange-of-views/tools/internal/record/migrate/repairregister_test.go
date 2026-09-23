package migrate_test

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// AN ARCHIVED REGISTER OPENS A SITTING (epoch 7 -> 8).
//
// THE ROW IS THE ARCHIVE'S. is-91-prime-b's register table is
// `event_id, tool_version, agent_id, run_via, agent_type` — there is no repairs_sitting column and
// no run before epoch 8 could record one, so the shape a real row arrives in is the shape this
// asserts on. Fed a column no epoch-7 row can carry, the earlier form of this test measured the
// translation against data migrate will never see.
//
// A ROW THAT DOES CARRY IT IS REFUSED, and that arm is what makes the step more than a comment: a
// silent drop reads identically whether the source was an honest archive or a record claiming a
// field its epoch could not hold. It is not an epoch-7 row — that is the point of refusing it.
func TestAnArchivedRegisterOpensASitting(t *testing.T) {
	dst := runtest.New(t, recordtest.TmpRun(t))
	entry, ok := migrate.Entries()["register"]
	if !ok {
		t.Fatal("no register translation: epoch 8 moved the register's shape and migrate has no step for it")
	}

	archived := migrate.OldEvent{ID: 1, SeatID: "blue-respond", TS: ts1, Word: "register", Key: "blue-respond:register:#1",
		Fields: map[string]any{
			"tool_version": "e6b33cf+dirty",
			"agent_id":     "agent_blue_respond_sitting_record",
			"run_via":      "injected",
			"agent_type":   "frank-exchange-of-views:blue-researcher",
		}}
	bodies, err := entry.Translate(archived, dst)
	if err != nil {
		t.Fatal(err)
	}
	r, isRegister := bodies[0].(*recordpb.Register)
	if !isRegister || len(bodies) != 1 {
		t.Fatalf("a register translated to %d bodies, the first %T", len(bodies), bodies[0])
	}
	if r.RepairsSitting != nil {
		t.Errorf("an archived register claims to repair %q — every one of them opens a sitting of its own", r.GetRepairsSitting())
	}
	if r.GetToolVersion() != "e6b33cf+dirty" || r.GetAgentId() != "agent_blue_respond_sitting_record" ||
		r.GetRunVia() != "injected" || r.GetAgentType() != "frank-exchange-of-views:blue-researcher" {
		t.Errorf("translated register = %v, want every archived column kept word for word", r)
	}

	forged := archived
	forged.Fields = map[string]any{"agent_id": "a", "repairs_sitting": "blue-respond:register:#3"}
	switch _, err := entry.Translate(forged, dst); {
	case err == nil:
		t.Error("a row carrying a column its epoch could not hold was translated rather than refused")
	case !strings.Contains(err.Error(), "repairs_sitting") || !strings.Contains(err.Error(), "blue-respond:register:#1"):
		t.Errorf("the refusal must name the field it found and the row it found it on: %v", err)
	}
}

// THE ARCHIVE IS THE FIXTURE. is-91-prime-b's blue-respond sat for G2 (register, log), and the
// engine's sitting-record re-prompt registered again under agent_blue_respond_sitting_record and filed
// a position, a revision and a log. Migrated, that register repairs nothing, so its acts are no
// sitting's: the third blue sitting holds its own register and log, and record-parity is unmoved
// because that sitting found G2 closed before it sat.
func TestIs91PrimeBsReprompt(t *testing.T) {
	tarball := archivePath(t, "2026-09-09_is-91-prime-b.tar.gz")
	runDir := t.TempDir()
	untar(t, tarball, runDir)
	toDir := recordtest.TmpRun(t)
	res, merr := migrate.Migrate(runDir, toDir, migrate.Entries(), migrate.Options{})
	if merr != nil || res == nil || len(res.Refusals) > 0 {
		t.Fatalf("migrate: %v (%+v)", merr, res)
	}
	fam, err := record.FamilyOf(runtest.Open(t, toDir))
	if err != nil {
		t.Fatal(err)
	}
	reprompts := 0
	for _, e := range fam.Events {
		r, ok := recordpb.BodyAs[*recordpb.Register](e)
		if !ok {
			continue
		}
		if r.RepairsSitting != nil {
			t.Errorf("an archived register claims a repair: %v", e)
		}
		if strings.HasSuffix(r.GetAgentId(), "_sitting_record") {
			reprompts++
		}
	}
	if reprompts != 1 {
		t.Fatalf("found %d sitting-record re-prompt registers, want the one this archive carries", reprompts)
	}
	ss := record.BlueSittings(fam.Events, record.AfterTheRun)
	if len(ss) != 3 {
		t.Fatalf("blue sittings = %d, want 3", len(ss))
	}
	var types []string
	for _, e := range ss[2].Acts {
		types = append(types, recordpb.Word(e.GetType()))
	}
	// ITS OWN REGISTER, AND NO LOG. The entry that sitting filed asserted a clean sitting, and
	// that type is retired — clean is derived from a bracketed sitting that filed nothing — so the
	// migration carries the register forward and drops the entry.
	if got := strings.Join(types, ","); got != "register" || len(ss[2].Open) != 0 {
		t.Errorf("blue sitting 3 = open %v, acts %s; want nothing open and its own register — the re-prompt repairs nothing once migrated", ss[2].Open, got)
	}
}
