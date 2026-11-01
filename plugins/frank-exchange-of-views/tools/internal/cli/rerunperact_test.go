package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/archive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/capture"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/lens"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// ONE STANDING RE-RUN PER BUNDLE, JUDGED PER ACT (plans/compute-and-show.md §III.1.4, §V V1a.4 and
// V1a.5). These drive one program whose output each test steers from outside its bundle: it reads a
// control file, and prints what the file says — or, where the file says `count`, prints a counter
// file's value and increments it. Both files sit beside the run and are named absolutely in the
// command, so they are part of the bundle's identity and every act on it is the same bundle; the
// program writes what it printed to out/f.txt too.

const countProgram = `const fs = require("fs");
const [ctl, ctr] = process.argv.slice(2);
let c = fs.readFileSync(ctl, "utf8").trim();
if (c === "count") { const n = +fs.readFileSync(ctr, "utf8"); fs.writeFileSync(ctr, String(n + 1)); c = String(n); }
fs.writeFileSync("out/f.txt", c);
console.log(c);
`

// perAct is a run whose report holds sentences, with the steerable program drafted in its scratch/.
type perAct struct {
	t                       *testing.T
	runDir, ctl, ctr, draft string
}

func newPerAct(t *testing.T, sentences ...string) *perAct {
	t.Helper()
	runDir := newRun(t)
	computeSeat(t, runDir, "# H\n\n"+strings.Join(sentences, "\n\n")+"\n")
	f := &perAct{t: t, runDir: runDir, ctl: filepath.Join(runDir, "ctl.txt"), ctr: filepath.Join(runDir, "ctr.txt")}
	f.draft = script(t, runDir, "count.js", countProgram)
	return f
}

// set writes a control or counter file.
func (f *perAct) set(path, body string) {
	f.t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		f.t.Fatal(err)
	}
}

func (f *perAct) read(path string) string {
	f.t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		f.t.Fatal(err)
	}
	return string(b)
}

// compute records the program against a sentence, with an invariant where one is given, and returns
// the computation's anchor.
func (f *perAct) compute(quote, variance string) string {
	f.t.Helper()
	args := []string{"compute", "--run", f.runDir, "--seat-id", "blue-respond", "--quote", quote, "--dir", f.draft, "--reason", "r"}
	if variance != "" {
		args = append(args, "--variance", variance)
	}
	if out, err := run(f.t, append(args, "--", "node", "count.js", f.ctl, f.ctr)...); err != nil {
		f.t.Fatalf("compute %q: %v\n%s", quote, err, out)
	}
	return lastBody(f.t, f.runDir, &recordpb.Computation{}).GetComputationId()
}

// reproduce judges an anchor's sentence as a lens, returning the envelope's result.
func (f *perAct) reproduce(anchor, as string) (map[string]any, error) {
	f.t.Helper()
	out, err := run(f.t, "reproduce", "--run", f.runDir, "--seat-id", "red-lens-logic", "--anchor", anchor, "--as", as,
		"--reason", "it prints the count the sentence states", "--json")
	if err != nil {
		return nil, err
	}
	var env struct {
		Result map[string]any `json:"result"`
	}
	if jerr := json.Unmarshal([]byte(out), &env); jerr != nil {
		f.t.Fatalf("reproduce %s: %v\n%s", anchor, jerr, out)
	}
	return env.Result, nil
}

// bundle is the steerable program's bundle sha, off the record.
func (f *perAct) bundle() string {
	f.t.Helper()
	for _, p := range mustComputations(f.t, f.runDir) {
		if slices.Contains(p.Argv, f.ctl) {
			return p.SHA
		}
	}
	f.t.Fatal("no computation of the steerable program on the record")
	return ""
}

// reproduces is every reproduce on the record, in record order.
func reproduces(t *testing.T, runDir string) []*recordpb.Reproduce {
	t.Helper()
	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var out []*recordpb.Reproduce
	for _, e := range m.Events {
		if r, ok := recordpb.BodyAs[*recordpb.Reproduce](e); ok {
			out = append(out, r)
		}
	}
	return out
}

// treeHash is the sha256 of a directory's listing: each file's path and bytes, in walk order. An
// absent directory hashes as empty.
func treeHash(t *testing.T, dir string) string {
	t.Helper()
	h := sha256.New()
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		h.Write([]byte(filepath.ToSlash(rel) + "\x00"))
		if d.Type().IsRegular() {
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			h.Write(b)
		}
		return nil
	})
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// staging is every .tmp-* directory in a bundle's store directory.
func staging(t *testing.T, runDir, bundle string) []string {
	t.Helper()
	m, err := filepath.Glob(filepath.Join(runDir, compute.StoreDir, bundle, ".tmp-*"))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// reruns is every directory under a bundle's rerun/.
func reruns(t *testing.T, runDir, bundle string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(runDir, compute.StoreDir, bundle, "rerun"))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	var out []string
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out
}

// evidenceSections is evidence.md's computation sections, keyed by heading number (K1, K2, …).
func evidenceSections(t *testing.T, runDir string) map[string]string {
	t.Helper()
	docs, err := shippedDocs(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	ev := ""
	for _, d := range docs {
		if d.File == report.FileEvidence {
			ev = d.Body
		}
	}
	out := map[string]string{}
	parts := strings.Split(ev, "\n### ")
	for _, p := range parts[1:] {
		num, _, _ := strings.Cut(p, " ")
		out[num] = p
	}
	return out
}

// fence is a body as evidence.md fences an output.
func fence(body string) string { return "```\n" + body + "\n```" }

// V1a.4. Acts A (recorded 1), B (recorded 3) and D (recorded 2) on one `variant` bundle, E on the
// same bundle recorded `reproducible`, and F on a bundle nobody executes. The counter is set to 3 and
// A's reproduce executes the bundle once; every later act compares against that execution with its
// own recorded output.
func TestEachActIsComparedAgainstTheBundlesOneReRun(t *testing.T) {
	f := newPerAct(t, "The first count is one.", "The second count is three.", "The third count is two.",
		"The fourth value is e.", "The fixed program prints f.")
	f.set(f.ctl, "count")
	f.set(f.ctr, "1")
	a := f.compute("The first count is one.", "a non-negative count")
	b := f.compute("The second count is three.", "a non-negative count")
	f.set(f.ctr, "2")
	d := f.compute("The third count is two.", "a non-negative count")
	f.set(f.ctl, "e")
	e := f.compute("The fourth value is e.", "")
	f.set(f.ctl, "count")
	fixed := script(t, f.runDir, "fixed.js", `console.log("f");`)
	if out, err := run(t, "compute", "--run", f.runDir, "--seat-id", "blue-respond", "--quote", "The fixed program prints f.",
		"--dir", fixed, "--reason", "r", "--", "node", "fixed.js"); err != nil {
		t.Fatalf("compute: %v\n%s", err, out)
	}
	fx := lastBody(t, f.runDir, &recordpb.Computation{}).GetComputationId()
	bundle := f.bundle()
	for id, want := range map[string]string{a: "1\n", b: "3\n", d: "2\n", e: "e\n"} {
		if got, _ := compute.ActOutput(f.runDir, bundle, id); string(got) != want {
			t.Fatalf("%s recorded %q, want %q", id, got, want)
		}
	}
	actA := treeHash(t, compute.ActPath(f.runDir, bundle, a))
	store := filepath.Join(f.runDir, compute.StoreDir)

	// capture's audit, before any seat executed the bundle: one execution, each act graded by its
	// own basis, and nothing stored.
	audit := func() capture.Audit {
		t.Helper()
		f.set(f.ctr, "3")
		before := treeHash(t, store)
		got := capture.ComputationRerunAudit(runtest.Open(t, f.runDir), 10)
		if after := treeHash(t, store); after != before {
			t.Errorf("capture's audit changed the computation store")
		}
		if n := f.read(f.ctr); n != "4" {
			t.Errorf("capture's audit executed the bundle %s time(s) past the counter's 3, want once", n)
		}
		return got
	}
	first := audit()
	for _, want := range []string{"moved " + a + " ", "moved " + d + " ", "FAIL " + e + " ", "backs acts recorded reproducible and variant"} {
		if !strings.Contains(first.Detail, want) {
			t.Errorf("capture's audit does not say %q:\n%s", want, first.Detail)
		}
	}
	if strings.Contains(first.Detail, b) || first.Verdict != "FAIL" {
		t.Errorf("capture's audit graded B, which reproduced, or did not fail E: %s\n%s", first.Verdict, first.Detail)
	}

	// A executes the bundle: its own comparison differs, and the counter moves once.
	f.set(f.ctr, "3")
	res, err := f.reproduce(a, "sound")
	if err != nil {
		t.Fatalf("reproduce A: %v", err)
	}
	if res["rerun"] != "differs" || res["ran"] != true || f.read(f.ctr) != "4" {
		t.Errorf("A's reproduce: %v, counter %s — want an execution, once, that differs", res, f.read(f.ctr))
	}
	standing := reproduces(t, f.runDir)[0].GetRerunSha()
	if got, err := compute.ReadRerun(f.runDir, bundle, standing, 0); err != nil || got.Output != "3\n" || len(got.Out) != 1 || string(got.Out[0].Body) != "3" {
		t.Errorf("the standing re-run's directory holds %+v (%v), want the execution's 3 and out/f.txt", got, err)
	}

	// B executes nothing: compared with its own recorded 3, it reproduces.
	res, err = f.reproduce(b, "sound")
	if err != nil {
		t.Fatalf("reproduce B: %v", err)
	}
	if res["rerun"] != "reproduced" || res["ran"] != false || res["rerun_sha"] != standing || res["rerun_by"] != "red-lens-logic" || f.read(f.ctr) != "4" {
		t.Errorf("B's reproduce: %v, counter %s — want no execution, reproduced, against A's", res, f.read(f.ctr))
	}
	// D executes nothing: its own recorded 2 differs from the execution's 3.
	if res, err = f.reproduce(d, "unsound"); err != nil || res["rerun"] != "differs" || !strings.Contains(res["drift"].(string), "on re-run the program wrote: 3") {
		t.Errorf("D's reproduce: %v (%v), want differs against its own 2", res, err)
	}
	if got := treeHash(t, compute.ActPath(f.runDir, bundle, a)); got != actA {
		t.Error("A's acts/ changed after it was written")
	}

	for _, r := range reproduces(t, f.runDir) {
		if r.GetRerunSha() != standing {
			t.Errorf("%s compared against %q, want the standing %s", r.GetComputationId(), r.GetRerunSha(), standing)
		}
		if r.GetRerun() == recordpb.Rerun_RERUN_DIFFERS && r.GetDrift() == "" {
			t.Errorf("%s differs and carries no drift", r.GetComputationId())
		}
		if r.GetRan() != (r.GetRerunExit() == 0 && r.RerunExit != nil) {
			t.Errorf("%s ran=%v carries rerun_exit %v", r.GetComputationId(), r.GetRan(), r.RerunExit)
		}
	}

	rows := map[string]record.EvidenceComputationJSON{}
	for _, p := range evidenceJSON(t, f.runDir).Computations {
		rows[p.Anchor] = p
	}
	for id, want := range map[string]string{a: "differs", b: "reproduced", d: "differs", e: record.RerunNotCompared, fx: record.RerunNone} {
		if rows[id].Rerun != want {
			t.Errorf("%s's evidence row reads rerun %q, want %q", id, rows[id].Rerun, want)
		}
	}
	for _, id := range []string{a, b, d, e} {
		if got := rows[id].BundleBases; !slices.Equal(got, []string{"reproducible", "variant"}) {
			t.Errorf("%s's row names its bundle's bases %q", id, got)
		}
	}
	if rows[fx].BundleBases != nil {
		t.Errorf("F's bundle backs one basis and its row names %q", rows[fx].BundleBases)
	}

	// evidence.md shows each act's own output, and beside a comparison that differs the output of
	// the execution that comparison read.
	sec := evidenceSections(t, f.runDir)
	if s := sec["K1"]; !strings.Contains(s, fence("1")) || !strings.Contains(s, "Output on re-run:\n\n"+fence("3")) {
		t.Errorf("A's section does not show its own 1 beside the re-run's 3:\n%s", s)
	}
	if s := sec["K2"]; !strings.Contains(s, fence("3")) || strings.Contains(s, fence("1")) || strings.Contains(s, "Output on re-run") {
		t.Errorf("B's section does not show its own 3 alone:\n%s", s)
	}
	if s := sec["K3"]; !strings.Contains(s, fence("2")) || !strings.Contains(s, "Output on re-run:\n\n"+fence("3")) || strings.Contains(s, fence("1")) {
		t.Errorf("D's section does not show its own 2 beside the re-run's 3, or shows A's 1:\n%s", s)
	}
	if s := sec["K4"]; !strings.Contains(s, "**not compared**") {
		t.Errorf("E's section does not say it was not compared:\n%s", s)
	}
	if s := sec["K5"]; !strings.Contains(s, "**not re-run**") {
		t.Errorf("F's section does not say it was not re-run:\n%s", s)
	}

	// capture's audit reads no re-run the seats stored: a tampered one changes nothing.
	if err := os.WriteFile(filepath.Join(compute.RerunPath(f.runDir, bundle, standing), "output"), []byte("tampered\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// The count of bundles no seat executed moves with A's reproduce; every verdict stays.
	verdicts := func(a capture.Audit) string {
		_, notes, _ := strings.Cut(a.Detail, "by any seat; ")
		return a.Verdict + " " + notes
	}
	if again := audit(); verdicts(again) != verdicts(first) {
		t.Errorf("capture's audit read the stored re-run:\nbefore: %+v\nafter:  %+v", first, again)
	}
}

// A LATER EXECUTION RECORDED BETWEEN AN ACT'S READ AND ITS WRITE REFUSES THE ACT. The act read the
// standing re-run and compared against it; another checker's execution of the bundle lands before
// the act is appended, and the write — which re-reads the standing re-run — refuses the act.
func TestAComparisonAgainstAReRunThatStoppedStandingIsRefused(t *testing.T) {
	f := newPerAct(t, "The value is x.", "The value is still x.")
	f.set(f.ctl, "x")
	p := f.compute("The value is x.", "")
	q := f.compute("The value is still x.", "")
	if _, err := f.reproduce(p, "sound"); err != nil {
		t.Fatal(err)
	}
	bundle := f.bundle()
	later := &recordpb.Reproduce{ComputationId: proto.String(p), BundleSha: proto.String(bundle), Ran: proto.Bool(true),
		Rerun: recordpb.Rerun_RERUN_REPRODUCED.Enum(), RerunSha: proto.String(strings.Repeat("f", 64)), RerunExit: proto.Int32(0),
		Soundness: recordpb.Soundness_SOUNDNESS_SOUND.Enum(), Note: proto.String("another checker's execution")}
	restore := lens.SetBeforeReproduceAppendForTest(func() error {
		_, err := record.Append(record.Identity{Run: runtest.Open(t, f.runDir), SeatID: "red-lens-computation"}, later)
		return err
	})
	before := countType(t, f.runDir, recordpb.EventType_EVENT_TYPE_REPRODUCE)
	_, err := f.reproduce(q, "sound")
	restore()
	if err == nil || !strings.Contains(err.Error(), "no longer stands") {
		t.Fatalf("an act whose standing re-run moved before its write was recorded: %v", err)
	}
	if n := countType(t, f.runDir, recordpb.EventType_EVENT_TYPE_REPRODUCE); n != before+1 {
		t.Errorf("%d reproduce(s) recorded, want the later execution alone", n-before)
	}
}

// V1a.4: A `reproducible` act whose OWN comparison differs takes no `sound`, whatever the standing
// re-run's own act showed. P recorded x and Q recorded y on one bundle; Q executes it, printing y,
// and reproduces; P compared against that execution differs.
func TestASoundVerdictOverAnActsOwnDifferingComparisonIsRefused(t *testing.T) {
	f := newPerAct(t, "The value is x.", "The value is y.")
	f.set(f.ctl, "x")
	p := f.compute("The value is x.", "")
	f.set(f.ctl, "y")
	q := f.compute("The value is y.", "")
	if res, err := f.reproduce(q, "sound"); err != nil || res["rerun"] != "reproduced" || res["ran"] != true {
		t.Fatalf("Q's reproduce: %v (%v), want its own execution, reproduced", res, err)
	}
	before := countType(t, f.runDir, recordpb.EventType_EVENT_TYPE_REPRODUCE)
	_, err := f.reproduce(p, "sound")
	if err == nil || !strings.Contains(err.Error(), "did NOT reproduce this computation's recorded result") || !strings.Contains(err.Error(), "recorded: x") {
		t.Fatalf("a sound verdict over P's own differing comparison: %v", err)
	}
	if n := countType(t, f.runDir, recordpb.EventType_EVENT_TYPE_REPRODUCE); n != before {
		t.Errorf("the refused verdict recorded %d event(s)", n-before)
	}
	if res, err := f.reproduce(p, "unsound"); err != nil || res["rerun"] != "differs" || res["ran"] != false {
		t.Errorf("P's unsound verdict: %v (%v), want differs, executing nothing", res, err)
	}
}

// V1a.5 (a), (b), (e), (f): the re-run store's lifecycle. A refused verdict stores nothing; an
// append that fails after the rename leaves a directory no event names, and the next reproduce
// executes again, finds that directory holding its bytes, and stands; the archive holds each
// recorded re-run and nothing else; and a migration that renumbers the `ran` event leaves its
// directory as it was, resolved from the record.
func TestAReRunIsStoredBeforeItsEventAndOnlyWhenTheActIsRecorded(t *testing.T) {
	f := newPerAct(t, "The value is x.")
	f.set(f.ctl, "x")
	p := f.compute("The value is x.", "")
	bundle := f.bundle()
	f.set(f.ctl, "z")
	before := countType(t, f.runDir, recordpb.EventType_EVENT_TYPE_REPRODUCE)

	// (a) a `sound` over a reproducible act whose re-run differs is refused, and stores nothing.
	if _, err := f.reproduce(p, "sound"); err == nil {
		t.Fatal("a sound verdict over a re-run that differs was recorded")
	}
	if r, s := reruns(t, f.runDir, bundle), staging(t, f.runDir, bundle); len(r) != 0 || len(s) != 0 {
		t.Fatalf("the refused verdict left rerun/ %q and staging %q", r, s)
	}

	// (b) the append fails after the rename: a directory no event names.
	restore := lens.SetBeforeReproduceAppendForTest(func() error { return errors.New("the append fails here") })
	_, err := f.reproduce(p, "unsound")
	restore()
	if err == nil || !strings.Contains(err.Error(), "the append fails here") {
		t.Fatalf("the seam did not fail the append: %v", err)
	}
	orphan := reruns(t, f.runDir, bundle)
	if len(orphan) != 1 || countType(t, f.runDir, recordpb.EventType_EVENT_TYPE_REPRODUCE) != before || len(staging(t, f.runDir, bundle)) != 0 {
		t.Fatalf("after the failed append: rerun/ %q, %d event(s), staging %q — want one directory, no event, no staging",
			orphan, countType(t, f.runDir, recordpb.EventType_EVENT_TYPE_REPRODUCE)-before, staging(t, f.runDir, bundle))
	}
	orphanBytes := treeHash(t, compute.RerunPath(f.runDir, bundle, orphan[0]))

	// The next reproduce executes, finds the directory present with its bytes, and stands.
	res, err := f.reproduce(p, "unsound")
	if err != nil || res["ran"] != true || res["rerun_sha"] != orphan[0] {
		t.Fatalf("the next reproduce: %v (%v), want an execution standing on %s", res, err, orphan[0])
	}
	ev := lastBody(t, f.runDir, &recordpb.Reproduce{})
	if ev.RerunExit == nil || ev.GetRerunSha() != orphan[0] {
		t.Errorf("the recorded execution carries rerun_exit %v and rerun_sha %q", ev.RerunExit, ev.GetRerunSha())
	}
	if r := reruns(t, f.runDir, bundle); len(r) != 1 || treeHash(t, compute.RerunPath(f.runDir, bundle, r[0])) != orphanBytes {
		t.Errorf("rerun/ holds %q after the execution that stands, want the one directory unchanged", r)
	}
	got, err := compute.ReadRerun(f.runDir, bundle, ev.GetRerunSha(), int(ev.GetRerunExit()))
	if err != nil || got.Output != "z\n" || len(got.Out) != 1 || string(got.Out[0].Body) != "z" {
		t.Errorf("the stored execution reads %+v (%v), want its output and out/", got, err)
	}

	// (e) the archive holds each recorded re-run, and no staging or directory no event names.
	unnamed := compute.RerunPath(f.runDir, bundle, strings.Repeat("0", 64))
	tmp := filepath.Join(f.runDir, compute.StoreDir, bundle, ".tmp-left")
	for _, d := range []string{unnamed, tmp} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, "output"), []byte("left\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	tarball, err := capture.ArchiveRecord(runtest.Open(t, f.runDir), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	unpacked := t.TempDir()
	if err := archive.Extract(tarball, unpacked); err != nil {
		t.Fatal(err)
	}
	at := filepath.Join(unpacked, compute.StoreDir, bundle)
	for _, want := range []string{filepath.Join("rerun", orphan[0], "output"), filepath.Join("rerun", orphan[0], "out", "f.txt"),
		filepath.Join("acts", p, "output"), filepath.Join("dir", "count.js")} {
		if _, err := os.Stat(filepath.Join(at, want)); err != nil {
			t.Errorf("the archive does not hold %s: %v", want, err)
		}
	}
	for _, not := range []string{filepath.Join("rerun", strings.Repeat("0", 64)), ".tmp-left"} {
		if _, err := os.Stat(filepath.Join(at, not)); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("the archive holds %s, which no event names", not)
		}
	}
	if err := os.RemoveAll(unnamed); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(tmp); err != nil {
		t.Fatal(err)
	}

	// (f) a migration that renumbers the `ran` event: its directory is unchanged and the record
	// still resolves it.
	ranID := func(dir string) int64 {
		t.Helper()
		m, err := record.MergedEvents(runtest.Open(t, dir))
		if err != nil {
			t.Fatal(err)
		}
		ids := m.At.IDs(m.Events)
		for i, e := range m.Events {
			if r, ok := recordpb.BodyAs[*recordpb.Reproduce](e); ok && r.GetRan() {
				return ids[i]
			}
		}
		t.Fatalf("no `ran` reproduce in %s", dir)
		return 0
	}
	to := filepath.Join(t.TempDir(), "migrated")
	t.Cleanup(func() { _ = recordsql.CloseUnder(to) })
	if res, err := migrate.Migrate(f.runDir, to, migrate.Entries(), migrate.Options{}); err != nil {
		t.Fatalf("migrate: %v (%+v)", err, res)
	}
	if ranID(to) == ranID(f.runDir) {
		t.Fatal("the replay did not renumber the `ran` event, so nothing here exercised the store's naming")
	}
	if got := treeHash(t, compute.RerunPath(to, bundle, orphan[0])); got != orphanBytes {
		t.Errorf("the migrated run's rerun/%s is not the source's", orphan[0][:12])
	}
	k, err := record.ComputationAt(runtest.Open(t, to), p)
	if err != nil {
		t.Fatal(err)
	}
	if k.Standing == nil || k.Standing.RerunSha != orphan[0] {
		t.Fatalf("the migrated record's standing re-run is %+v, want %s", k.Standing, orphan[0])
	}
	if _, err := compute.ReadRerun(to, bundle, k.Standing.RerunSha, k.Standing.Exit); err != nil {
		t.Errorf("the migrated record's re-run does not resolve: %v", err)
	}
}

// A REFUSED APPEND LEAVES NO ACT (V1a.5 (e)). The quote is checked where the event is placed, after
// the store is written: the bundle's dir/ stays, content-addressed and shared, and the act's
// directory is taken back, so no acts/ entry is left that no event names.
func TestARefusedComputeLeavesNoActDirectory(t *testing.T) {
	runDir := newRun(t)
	computeSeat(t, runDir, "# H\n\nSeven is prime.\n")
	s := script(t, runDir, "p.js", `console.log("7")`)
	if _, err := run(t, "compute", "--run", runDir, "--seat-id", "blue-respond", "--quote", "Nine is prime.", "--dir", s, "--reason", "r", "--", "node", "p.js"); err == nil {
		t.Fatal("a quote the report does not hold was recorded")
	}
	dirs, _ := filepath.Glob(filepath.Join(runDir, compute.StoreDir, "*", "dir"))
	if len(dirs) != 1 {
		t.Fatalf("the refusal came before the store was written (dir/ %q), so nothing here tests the take-back", dirs)
	}
	acts, _ := filepath.Glob(filepath.Join(runDir, compute.StoreDir, "*", "acts", "*"))
	tmp, _ := filepath.Glob(filepath.Join(runDir, compute.StoreDir, "*", ".tmp-*"))
	if len(acts) != 0 || len(tmp) != 0 {
		t.Errorf("the refused act left acts/ %q and staging %q", acts, tmp)
	}
}
