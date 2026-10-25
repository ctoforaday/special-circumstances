package capture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// CAPTURE READS THE TRANSCRIPTS THE RECORD NAMES, AND NAMES EVERY SEAT IT CANNOT READ. A file in a
// directory that no stop names is not read; a named path gone from the box, a stop with no path, and
// an agent with no stop each appear in the census line under their seat, because each is a seat whose
// spend and turns are missing from cost.md and the audits.
func TestCaptureReadsTheTranscriptsTheRecordNamesAndNamesEachMiss(t *testing.T) {
	dir := t.TempDir()
	here := filepath.Join(dir, "anything.jsonl")
	write(t, here, `{"agentId":"L1","message":{"id":"m","model":"claude-haiku-4-5","usage":{"input_tokens":1}}}`+"\n")
	// A transcript-shaped file no stop names: what a directory listing would have read.
	write(t, filepath.Join(dir, "agent-STRAY.jsonl"), `{"agentId":"STRAY"}`+"\n")
	gone := filepath.Join(dir, "gone.jsonl")
	stop := func(agent, path string) *recordpb.Event {
		cl := &recordpb.SittingClose{AgentId: proto.String(agent)}
		if path != "" {
			cl.AgentTranscriptPath = proto.String(path)
		}
		return recordtest.Event(t, record.HarnessSeat, cl)
	}
	reg := func(seat, agent string) *recordpb.Event {
		return recordtest.Event(t, seat, &recordpb.Register{AgentId: proto.String(agent)})
	}
	runDir := recordtest.TmpRun(t)
	recordtest.Seed(t, runDir,
		reg("red-lens-evidence", "L1"), stop("L1", here),
		reg("red-chair", "C1"), stop("C1", gone),
		reg("frontier", "F1"), stop("F1", ""),
		reg("judge", "J1"),
	)
	run := runtest.Open(t, runDir)
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}

	got, line := seatTranscripts(run, &fam)
	if len(got) != 1 || got[0].AgentID != "L1" || got[0].Path != here {
		t.Fatalf("transcripts = %+v, want only L1 at the path its stop names", got)
	}
	for _, want := range []string{
		"2 named by the record's sitting closes, 1 readable here",
		"1 NOT ON THIS BOX", "red-chair (agent C1): " + gone,
		"1 stopped with NO PATH", "frontier (agent F1)",
		"1 opened a sitting with NO STOP", "judge (agent J1)",
	} {
		if !strings.Contains(line, want) {
			t.Errorf("the census line lacks %q:\n%s", want, line)
		}
	}
	if strings.Contains(line, "STRAY") {
		t.Errorf("a file no stop names is not the run's: %s", line)
	}

	if _, line := seatTranscripts(run, nil); !strings.Contains(line, "NONE READ") {
		t.Errorf("an unreadable record must say no transcript was read: %s", line)
	}
}

// CAPTURE ITSELF READS WHAT THE RECORD NAMES: the transcript directory it is handed holds the
// journal and nothing it reads as a seat, and the one transcript a stop names reaches the census,
// the tarball, the turn ingest and cost.md.
func TestCaptureRunReadsTheSeatTranscriptsFromTheRecord(t *testing.T) {
	home := t.TempDir() // the mirror purge reads it
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	p := filepath.Join(t.TempDir(), "trajectory-of-L1.jsonl")
	write(t, p, `{"agentId":"L1","timestamp":"2026-09-03T03:00:00.000Z","message":{"id":"m","model":"claude-haiku-4-5","usage":{"input_tokens":10,"output_tokens":1}}}`+"\n")
	wf := t.TempDir()
	write(t, filepath.Join(wf, "agent-STRAY.jsonl"), `{"agentId":"STRAY","message":{"id":"s","model":"claude-haiku-4-5","usage":{"input_tokens":10}}}`+"\n")
	runDir := recordtest.TmpRun(t)
	recordtest.Seed(t, runDir,
		recordtest.Event(t, "red-lens-evidence", &recordpb.Register{AgentId: proto.String("L1")}),
		recordtest.Event(t, record.HarnessSeat, &recordpb.SittingClose{AgentId: proto.String("L1"), AgentTranscriptPath: proto.String(p)}),
	)
	if err := os.MkdirAll(filepath.Join(runDir, "trajectories"), 0o755); err != nil {
		t.Fatal(err)
	}
	run := runtest.Open(t, runDir)
	_, report, _, err := Run(run, wf, t.TempDir(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"transcripts: 1 named by the record's sitting closes, 1 readable here", "tarball: 1 transcript(s)", "seat turns: 1 row(s) from 1 transcript(s)"} {
		if !strings.Contains(report, want) {
			t.Errorf("capture's report lacks %q:\n%s", want, report)
		}
	}
	costMD, err := os.ReadFile(filepath.Join(runDir, "cost.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(costMD), "Measured from 1 per-agent API transcripts") {
		t.Errorf("cost.md must read the one named transcript and not the directory's stray:\n%s", costMD)
	}
}

// THE TARBALL PACKS WHAT THE RECORD NAMES, under each file's own name.
func TestTheTarballPacksTheNamedTranscripts(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "sub", "agent-a1.jsonl")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, p, "{}\n")
	out := filepath.Join(dir, "t.tar.gz")
	if err := writeTarball(out, []record.Transcript{{AgentID: "a1", Path: p}}); err != nil {
		t.Fatal(err)
	}
	if names := tarNames(t, out); len(names) != 1 || names[0] != "agent-a1.jsonl" {
		t.Errorf("tarball entries = %v, want [agent-a1.jsonl]", names)
	}
}
