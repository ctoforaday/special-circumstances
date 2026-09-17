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

// AN ARCHIVED REGISTER REPAIRS NO SITTING (epoch 7 -> 8). No run before epoch 8 could say a register
// was a sitting-record repair, so the translation writes every register as one that opens a sitting
// of its own, whatever its row carries.
func TestAnArchivedRegisterRepairsNoSitting(t *testing.T) {
	dst := runtest.New(t, recordtest.TmpRun(t))
	entry, ok := migrate.Entries()["register"]
	if !ok {
		t.Fatal("no register translation: epoch 8 moved the register's shape and migrate has no step for it")
	}
	bodies, err := entry.Translate(migrate.OldEvent{ID: 1, SeatID: "blue-respond", TS: ts1, Word: "register",
		Fields: map[string]any{"agent_id": "agent_blue_respond_sitting_record", "repairs_sitting": "blue-respond:register:#3"}}, dst)
	if err != nil {
		t.Fatal(err)
	}
	r, ok := bodies[0].(*recordpb.Register)
	if !ok || len(bodies) != 1 {
		t.Fatalf("a register translated to %d bodies, the first %T", len(bodies), bodies[0])
	}
	if r.RepairsSitting != nil || r.GetAgentId() != "agent_blue_respond_sitting_record" {
		t.Errorf("translated register = %v, want its agent kept and no repair", r)
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
	if got := strings.Join(types, ","); got != "register,log" || len(ss[2].Open) != 0 {
		t.Errorf("blue sitting 3 = open %v, acts %s; want nothing open and its own register and log — the re-prompt repairs nothing once migrated", ss[2].Open, got)
	}
}
