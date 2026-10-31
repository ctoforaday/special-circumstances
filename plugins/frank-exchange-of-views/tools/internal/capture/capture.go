// Package capture is `feov-record capture`: the mechanical half of /research's run-record step
// plus the mechanized post-hoc auditor. It recomputes counts from
// the git-tracked FILES and diffs them against the envelopes' self-reports, writes cost.md and
// the transcript tarball, appends each card's scorecard, harvests precedents into law/proposed,
// removes the run-live marker, and writes run-record-audit.md — exit 2 on any audit FAIL.
//
// The record-backed audits (telemetry, record-parity) read the record IN-PROCESS, never by
// spawning `chair show`. A seat's log entries are on the record and nowhere else, so no audit
// here reconciles a second copy of them: the log is read with `ops log`.
// The PRECEDENT HARVEST reads it too, never the envelopes' self-reported ruling arrays: a bench
// that under-reports would promote less than it ruled, and one that reported nothing would
// promote nothing, indistinguishably from a run with nothing to promote.
package capture

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/archive"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cost"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/modeltier"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatturn"
	// Aliased: this package's own `Run` returns a named result called `report`, and the two
	// spellings would shadow each other at exactly the call site that needs the constant.
	reportdoc "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/scorecard"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// Audit is one check's result. Unreconciled is attestation-only (surfaced for the table test).
type Audit struct {
	Check        string
	Verdict      string
	Detail       string
	Unreconciled []Claim
}

// ---- small JS-coercion helpers (mirror the mjs number/string handling) ----

func strOf(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// jsToFixed0 renders a float as JS `x.toFixed(0)`.
func jsToFixed0(x float64) string { return strconv.FormatFloat(x, 'f', 0, 64) }

// jsSlice returns the prefix of s spanning the first n UTF-16 code units, matching JS
// String.slice(0,n). Counting BYTES diverges on non-ASCII prose: an em-dash (U+2014) is 3
// UTF-8 bytes but 1 UTF-16 unit, so a byte cut lands short. Real capture prose carries
// em-dashes; the real-data differential caught it. Runes above the BMP (>0xFFFF) are a
// surrogate pair = 2 units.
func jsSlice(s string, n int) string {
	units := 0
	for i, r := range s {
		w := 1
		if r > 0xFFFF {
			w = 2
		}
		if units+w > n {
			return s[:i]
		}
		units += w
	}
	return s
}

// ---- journal ----

// ReadJournal walks journal.jsonl tolerantly and returns every result object — the envelopes the
// seats returned, which the dispatch-parity audit holds to the record.
func ReadJournal(transcriptDir string) []map[string]any {
	results := []map[string]any{}
	b, err := os.ReadFile(filepath.Join(transcriptDir, "journal.jsonl"))
	if err != nil {
		return results
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		dec := json.NewDecoder(strings.NewReader(line))
		dec.UseNumber()
		var j map[string]any
		if dec.Decode(&j) != nil {
			continue
		}
		if r, ok := j["result"].(map[string]any); ok {
			results = append(results, r)
		}
	}
	return results
}

// ---- AUDIT 1: telemetry presence ----

// LivenessAudit records whether the run reached capture UNDER ITS OWN POWER or was terminated.
//
// THE TWO ARE OPPOSITE FACTS AND WERE THE SAME OUTPUT. "UNVERIFIED — nobody was ready" means the
// debate ran its course; "UNVERIFIED because the process was killed in epoch 1" means it never
// got to. Both rendered as an ordinary capture, so the run record could not tell a reader which
// one it was holding — and neither could a later run reading the corpus.
//
// Capture is allowed to proceed on a terminated run: the events are real and the report assembled
// from them is worth having. What is not allowed is for the artifact to be silent about it. A
// STALE record at capture time means the last seat stopped writing long before anyone ran this,
// which is the signature of a workflow that died rather than finished.
func LivenessAudit(run record.Run, now time.Time) Audit {
	// QUIET IS NOT THE SIGNAL — a finished run is quiet BECAUSE it finished, and the first draft
	// of this audit failed both of the day's real runs for exactly that. Silence separates
	// nothing; what separates them is whether the bench ever recorded an outcome.
	//
	//	quiet + an outcome on the record  -> it ran its course and was captured later
	//	still moving, no outcome yet      -> captured mid-run, unusual but the operator's call
	//	quiet + NO outcome                -> it stopped without finishing: TERMINATED
	verdict, err := record.TerminalVerdict(run)
	if err != nil {
		// Unreadable is neither finished nor terminated, and it is not SKIP's silent cousin: the
		// detail carries the read error so the operator sees what could not be read.
		return Audit{Check: "liveness", Verdict: "SKIP", Detail: "cannot tell whether this run finished or was terminated — the record could not be read: " + err.Error()}
	}
	if verdict != "" {
		return Audit{Check: "liveness", Verdict: "PASS", Detail: fmt.Sprintf(
			"the run reached a terminal outcome on the record (%s) — it finished under its own power", verdict)}
	}
	// THE ENGINE'S OWN STOP IS ON THE RECORD, and it is not a run in flight however recently it was
	// written: the engine threw, the lead wrote the error (`engine-stopped`), and nothing sat after it.
	// The record's own movement would otherwise read that last entry as activity and pass the run as
	// captured mid-run.
	if stop, err := engineStop(run); err != nil {
		return Audit{Check: "liveness", Verdict: "SKIP", Detail: "cannot tell whether this run finished or was terminated — the record could not be read: " + err.Error()}
	} else if stop != "" {
		return Audit{Check: "liveness", Verdict: "FAIL", Detail: "NO terminal outcome on the record, and the ENGINE STOPPED the run — its error, as the lead recorded it: " + stop}
	}
	// THE RECORD'S ACTIVITY, NOT THE RUN'S DISPOSITION: this audit runs before capture writes the
	// mark that ends the run, and again on a run captured before, and either way it asks whether the
	// record stopped moving before anyone captured it.
	l := record.AssessActivity(run, now)
	switch l.State {
	case record.StateStale:
		return Audit{Check: "liveness", Verdict: "FAIL", Detail: fmt.Sprintf(
			"NO terminal outcome on the record, and it went quiet %s before this capture (last: %s, %s) — "+
				"past the %s this run's own cadence implies. This run was TERMINATED, not finished, and its "+
				"verdict must be read that way: an idle SIGTERM kills the workflow, which then cannot lift its "+
				"own marker, so every other instrument goes on reporting a run in flight. %s",
			l.Age.Round(time.Second), l.Last.Seat, l.Last.Type, l.Threshold.Round(time.Second), l.Basis)}
	case record.StateUnmeasured:
		return Audit{Check: "liveness", Verdict: "SKIP", Detail: "cannot tell whether this run finished or was terminated — " + l.Basis}
	default:
		return Audit{Check: "liveness", Verdict: "PASS", Detail: fmt.Sprintf(
			"no terminal outcome yet and the record was still moving %s before capture (last: %s) — captured mid-run",
			l.Age.Round(time.Second), l.Last.Seat)}
	}
}

// engineStop is the error the engine stopped the run on, as the lead recorded it — the latest
// `failure` entry under the harness seat, which is the stop as fields (`engine-stopped`) — where no
// seat opened a sitting after it, with the sittings that failed (record.FailedSittingsOf) named
// after it; "" when the record holds no stop, or a resume sat past it.
func engineStop(run record.Run) (string, error) {
	m, err := record.MergedEvents(run)
	if err != nil {
		return "", err
	}
	stop := -1
	for i, e := range m.Events {
		if _, opens := m.At.Opens(e); opens {
			stop = -1
			continue
		}
		if l, ok := recordpb.BodyAs[*recordpb.Log](e); ok && e.GetSeatId() == record.HarnessSeat && l.GetType() == recordpb.LogType_LOG_TYPE_FAILURE {
			stop = i
		}
	}
	if stop < 0 {
		return "", nil
	}
	text := m.Events[stop].GetLog().GetText()
	failed := record.FailedSittingsOf(m.Events, m.At)
	if len(failed) == 0 {
		return text, nil
	}
	named := make([]string, len(failed))
	for i, f := range failed {
		named[i] = fmt.Sprintf("%s sitting %d (epoch %d)", f.SeatID, f.Sitting, f.Epoch)
	}
	return fmt.Sprintf("%s — and %d sitting(s) closed with no admitted envelope: %s", text, len(failed), strings.Join(named, "; ")), nil
}

// TelemetryAudit checks that the telemetry series derived from the record has a row for every red
// epoch.
//
// EVERY EPOCH IS A RED EPOCH: the record numbers its epochs by the chair's sittings (events_w), so
// the red epochs are 1..N, where N is the last epoch with work in it (WindowIndex.CurrentEpoch). The
// count is read off that field, never off the chair's positions: a run that ends in PASS files no
// position, and a count of positions reads 0 on a run of seven epochs — a zero indistinguishable
// from a run where the chair never sat.
//
// famErr is FamilyOf's error. A record that cannot be read is NOT MEASURED, never 0 red epochs.
func TelemetryAudit(run record.Run, fam record.Family, famErr error) Audit {
	if famErr != nil {
		return Audit{Check: "telemetry", Verdict: "SKIP", Detail: "NOT MEASURED — the record could not be read, so its red epochs are unknown (not a run with none): " + famErr.Error()}
	}
	redEpochs := fam.At.CurrentEpoch(fam.Events)
	if redEpochs == 0 {
		return Audit{Check: "telemetry", Verdict: "SKIP", Detail: "the record holds no epoch with work in it: the chair never sat"}
	}
	lines, err := view.Telemetry(run)
	if err != nil {
		return Audit{Check: "telemetry", Verdict: "FAIL", Detail: fmt.Sprintf("telemetry could not be computed from the record: %v", err)}
	}
	// A ROW WITHOUT AN EPOCH IS NOT COUNTED: the field is optional in the schema, so presence is the
	// question, and a key that was never written cannot masquerade as one that was absent.
	covered := map[int32]bool{}
	for _, j := range lines {
		if j.Epoch != nil {
			covered[j.GetEpoch()] = true
		}
	}
	var missing []string
	for r := 1; r <= redEpochs; r++ {
		if !covered[int32(r)] {
			missing = append(missing, strconv.Itoa(r))
		}
	}
	if len(missing) > 0 {
		return Audit{Check: "telemetry", Verdict: "FAIL", Detail: fmt.Sprintf("telemetry has no row for red epoch(s) %s of the %d on the record", strings.Join(missing, ", "), redEpochs)}
	}
	return Audit{Check: "telemetry", Verdict: "PASS", Detail: fmt.Sprintf("telemetry covers all %d red epoch(s) on the record", redEpochs)}
}

// AUDIT 2 IS GONE: "shard self-report vs files".
//
// It read red/ledger.md and red/archive.md and compared their line counts against the chair's
// self-reported ledger_closure_lines / archive_blocks. BOTH SIDES OF THAT COMPARISON ARE GONE —
// the envelope counts were removed 2026-07-19 for comparing numbers the chair made up (a haiku
// smoke self-reported archive_blocks: 22 in an epoch whose true archived count was 0), and the
// files stopped being written when the ledger and archive became rendered projections.
//
// It never said so. With no files it returned SKIP, "no ledger/archive (pre-sharding run)" — a
// benign, plausible reading, and wrong in the one way that mattered: those were the NEWEST runs.
// Every 2026-08 capture reports it, and the audit had been measuring nothing for months while
// reading as an antique politely declining to judge a modern run. Had the files existed, the
// other half was dead too: no envelope carries those counts, so `lastRed` stays nil and the
// verdict is a hardcoded PASS.
//
// Its unit test passed throughout, because the test WROTE the two files first. That is what kept
// it looking alive: the only place in the system where those files still existed was the test
// that proved the audit could read them.
//
// The numbers it was reaching for live where they cannot be authored: record.ArchiveOf
// computes the archive's size at epoch start by replay, and AttestationAudit reconciles each
// anchored closure against real tool calls. Both read the board. A self-report has no place on
// either side of that comparison.

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// ---- AUDIT 4: context-use ----

type peak struct {
	agent  string
	peak   int
	window int
}

// ContextUse is each seat's largest prompt against its model's context window. The prompt is a
// REQUEST's — read by seatturn.Parse, the reader seat_turn and cost.md share — and the window is the
// model's own, from the cost package's model table: Haiku 5.5's is 1M where Haiku 4.5's was 200K, and
// a window guessed from the tier name put a 118K Haiku 5.5 seat at 59% of a window it does not have.
func ContextUse(transcripts []record.Transcript) Audit {
	var peaks []peak
	for _, tr := range transcripts {
		b, err := os.ReadFile(tr.Path)
		if err != nil {
			continue
		}
		pk, model := 0, ""
		_, turns := seatturn.Parse(string(b))
		for _, t := range turns {
			if t.Prompt() > pk {
				pk = t.Prompt()
				if t.Model != "" {
					model = t.Model
				}
			}
		}
		if pk > 0 {
			peaks = append(peaks, peak{agent: tr.AgentID, peak: pk, window: cost.WindowOf(model)})
		}
	}
	if len(peaks) == 0 {
		return Audit{Check: "context-use", Verdict: "SKIP", Detail: "no usage records found"}
	}
	sort.SliceStable(peaks, func(i, j int) bool {
		return float64(peaks[i].peak)/float64(peaks[i].window) > float64(peaks[j].peak)/float64(peaks[j].window)
	})
	top := peaks[0]
	var flagged []peak
	for _, p := range peaks {
		if float64(p.peak)/float64(p.window) > 0.5 {
			flagged = append(flagged, p)
		}
	}
	v := "PASS"
	tail := ""
	if len(flagged) > 0 {
		v = "WARN"
		var fl []string
		for _, p := range flagged {
			fl = append(fl, fmt.Sprintf("    - %s: %sk / %dk", p.agent, jsToFixed0(float64(p.peak)/1000), p.window/1000))
		}
		tail = ":\n" + strings.Join(fl, "\n")
	}
	detail := fmt.Sprintf("peak %sk = %s%% of its %dk window (agent %s); %d seats measured; %d over the 50%% tripwire%s",
		jsToFixed0(float64(top.peak)/1000), jsToFixed0(float64(top.peak)/float64(top.window)*100), top.window/1000, top.agent,
		len(peaks), len(flagged), tail)
	return Audit{Check: "context-use", Verdict: v, Detail: detail}
}

// ---- AUDIT 5: assembly regression screen ----

// AssemblyScreen catches a report that still carries a citation red found AGAINST.
//
// # It spent two releases unable to
//
// THE SIGNAL WAS DESTROYED, not merely relocated. Until the citation ledger moved onto the
// record, red typed its verdict into the confidence column as prose — "LOW — REFUTED: closed
// as duplicate", "LOW (absent from abs and html)" — and this screen regex-scanned that column
// for REFUTED|ABSENT. Real ledgers carried 5-11 such rows (2026-07-12 through 2026-07-20).
// Then the grade became a CLOSED ENUM of high|medium|low — three values that all mean the
// source SUPPORTS the claim — so there was no field in which red could say a verification
// failed, and `red/citation-ledger.md` became a `setup` stub nothing writes (46 bytes on the
// 2026-08-05 run). The screen read that empty file and reported
// "PASS — 0 REFUTED-row token(s) screened; no hits" on every record-mode run.
//
// It was then made honest — SKIP, naming the gap — which stopped the false PASS and still left
// the check unable to check anything.
//
// # What made it possible again
//
// `lens verify --as` grew its negative half (#296): `refutes` (the source contradicts the claim)
// and `absent` (the source is silent on it) are outcomes red can record, and `--anchor` (#382)
// says WHICH citation each verdict is about. So the screen is now a join on fields rather than a
// regex over prose: for every source with a verification that found against it, is the anchor
// still in the assembled report?
//
// The prose match is gone with it. The old check keyed on the strings REFUTED|ABSENT appearing
// in a rendered row, which was a hope about string shape — and would have gone on returning zero
// the moment anyone reworded the renderer.
// FootnoteIntegrity checks that the assembled report DEFINES every footnote it references.
//
// # The defect this exists for, and why AssemblyScreen could not see it
//
// Assembly does not preserve blue's `<!--cite:…-->` anchors — it WEAVES them into visible
// footnotes, which is the design and is why AssemblyScreen screens the url rather than the
// anchor. The 2026-08-23 runs prove that weave can be LOSSY on one class and sound on another
// in the same pass: 11 citation anchors became `[^1]`–`[^11]` with a complete bibliography,
// while 6 computation anchors became `[^P1]`–`[^P6]` references with NO definitions and appendix
// headings written as `### [^P1] …` — which markdown reads as a second reference, not a
// heading. Both reports shipped and rendered literal `[^P1]` at every site: 12 broken
// references in one, 4 in the other.
//
// Nothing caught it. The blue-side detectors (`unbacked_citations`, `dropped_finding_markers`)
// read blue/report.md, where the anchors are all present and correct, so they reported 0 while
// the DELIVERABLE was broken. This audit reads the assembled artifact, which is the only
// surface where the weave's output can be judged.
//
// The rule is markdown's own, so a markdown parser reads it (report.FootnotesIn): a reference
// with no definition renders as its own source text, code is no footnote, and a reference may
// legitimately be followed by a colon in prose ("…artifacts exist[^P1]: `skills/…`").
// PER DOCUMENT, AND THAT IS NOT A DETAIL. A footnote definition cannot cross a file boundary,
// so once the report is a set, a definition in evidence.md does nothing for a reference in
// report.md — and an audit that concatenated the set before scanning would report a clean
// board for exactly the defect this check exists to catch.
func FootnoteIntegrity(run record.Run) Audit {
	refs, defs := 0, 0
	var dangling []string
	read := 0
	for _, name := range reportdoc.Files() {
		body, err := os.ReadFile(filepath.Join(run.Dir(), name))
		if err != nil {
			continue
		}
		read++
		r, d, bad := reportdoc.FootnotesIn(string(body))
		refs += r
		defs += d
		for _, id := range bad {
			// AS CODE: this detail is printed in the run document, where a bare label would be the
			// next audit's dangling reference.
			dangling = append(dangling, "`"+name+":[^"+id+"]`")
		}
	}
	if read == 0 {
		return Audit{Check: "footnote-integrity", Verdict: "SKIP",
			Detail: "no assembled document to read — this audit judges the weave's output, so before assembly there is nothing to judge"}
	}
	sort.Strings(dangling)
	if len(dangling) > 0 {
		return Audit{Check: "footnote-integrity", Verdict: "FAIL",
			Detail: fmt.Sprintf("%d footnote(s) referenced with no definition in the document that carries them: %s. They render as literal source text to every reader, and the evidence they were meant to carry is unreachable from the deliverable",
				len(dangling), strings.Join(dangling, ", "))}
	}
	return Audit{Check: "footnote-integrity", Verdict: "PASS",
		Detail: fmt.Sprintf("%d document(s) scanned; %d footnote(s) referenced, %d defined, none dangling", read, refs, defs)}
}

func AssemblyScreen(run record.Run) Audit {
	fam, err := record.FamilyOf(run)
	if err != nil {
		return Audit{Check: "assembly-screen", Verdict: "SKIP", Detail: "the record could not be read: " + err.Error()}
	}
	ev := record.EvidenceJSONOf(fam.Events, fam.At)
	if ev.Counts.Sources == 0 {
		return Audit{Check: "assembly-screen", Verdict: "SKIP", Detail: "no citations on the record — nothing to screen"}
	}

	// THE SCREEN NEEDS THE DELIVERABLE: before assembly there is nothing a reader is pointed at,
	// and saying so beats a PASS that means "the file I was looking for was not there".
	assembled := false
	for _, name := range reportdoc.Files() {
		if body, rerr := os.ReadFile(filepath.Join(run.Dir(), name)); rerr == nil && len(body) > 0 {
			assembled = true
			break
		}
	}
	if !assembled {
		return Audit{Check: "assembly-screen", Verdict: "SKIP",
			Detail: fmt.Sprintf("%d citation(s) on the record and no assembled document yet — the screen runs against the assembled artifact", ev.Counts.Sources)}
	}

	// EACH CITATION BY ITS STANDING, KEYED BY ITS ANCHOR. The assembled report weaves every anchor
	// the report carries into a footnote, so a citation still cited is one the record has not
	// retired — and the ledger that says so is the one that says whether a check of its sentence
	// stands. A refutation of a sentence blue has since rewritten, and then checked again, is
	// answered; one standing against the sentence the report carries is not. The url is no key:
	// two citations of one url back two sentences, and a screen keyed by it failed the run whose
	// refuted sentence had been fixed and re-verified (m30, C-4f4b8cfa).
	urlOf := map[string]string{}
	for _, s := range ev.Sources {
		urlOf[s.Anchor] = s.URL
	}
	var carried []string
	for _, a := range record.CitationsFoundAgainst(fam.Events, fam.At) {
		carried = append(carried, fmt.Sprintf("%s (%s)", a, urlOf[a]))
	}

	if len(carried) > 0 {
		return Audit{Check: "assembly-screen", Verdict: "FAIL",
			Detail: fmt.Sprintf("%d source(s) red finds AGAINST the sentence the report carries are still cited in the assembled report: %s. A refuted or absent source in the bibliography points a reader at evidence the audit rejected",
				len(carried), strings.Join(carried, ", "))}
	}
	// EVER FOUND AGAINST is this audit's own count, read off the verification rows: sources_refuted
	// is the standing set, which a PASS here has just shown empty.
	ever := 0
	for _, s := range ev.Sources {
		if slices.ContainsFunc(s.Verified, record.EvidenceVerificationJSON.Refuted) {
			ever++
		}
	}
	return Audit{Check: "assembly-screen", Verdict: "PASS",
		Detail: fmt.Sprintf("%d citation(s), none with a standing check against the sentence it backs (refutes|absent); %d ever found against by red, each since re-checked, reopened by an edit, or retired; %d cited source(s) no reading of the sentence they back has reached",
			ev.Counts.Sources, ever, ev.Counts.SourcesUnverified)}
}

// ---- AUDIT 6: record parity ----

// StrayRecordsAudit finds event shards written OUTSIDE any run directory.
//
// # What a stray is, and why it is not simply "another run"
//
// The repository holds many run directories, each with its own `records/`. What distinguishes a
// STRAY is that its parent is not a run: no `inputs/run-config.json`, no `blue/`. Nothing set it
// up; something wrote into it.
//
// Measured (#358): during a research run of 2026-08-10 a seat whose shell cwd was
// the `tools/` directory resolved the relative `research/<slug>/` from there, and the tool built
// a whole second blackboard — the lane's entire 13.7 KB draft, its own shards, clock and locks —
// while the real run's `blue/candidates/` stayed empty for the run. TWO shards of one seat class
// existed in both places, so the resolution differed per INVOCATION, not per seat. The run
// survived, which is the problem: work landing outside the run is indistinguishable from a seat
// that produced nothing.
//
// `RegisterSeat` refuses a run directory that does not exist now, so the creation path is closed.
// This audit covers what that cannot: a stray already on disk from an older binary, and the
// narrower case where the wrongly-resolved directory happens to exist already.
// THE DISCARDED-EVENTS AUDIT IS DELETED, and the deletion is the point rather than a tidy-up.
//
// It failed a run whose replay threw away work. Multi-nonce was normal — a `register` rotates the
// nonce so a crash re-dispatch writes a fresh shard, and the retry rewrites the same idempotency
// keys, so nothing is lost. The lossy case was one seat id used for two different SITTINGS, where
// the losing shard held keys that survived nowhere. The replay had reported both in one sentence
// ("multi-nonce seat X: N dispatches") and nothing gated on either, so a run that lost a whole
// bench sitting reported success (#394). This audit was the gate that fixed that.
//
// There is no losing shard now. Both sittings' events are ROWS, told apart by `nonce`, and no
// winner is selected — so the loss this audit reported is not merely absent, it is unrepresentable.
// Kept as a permanent PASS it would answer "no seat's replay discarded recorded work" on every run,
// in the exact words it used when the check was real, which is the failure #394 was.
//
// What it was protecting is still protected, one layer down and by construction: every event
// written is a row, and `UNIQUE (seat_id, nonce, seq)` is what stops two writers claiming one slot.

func StrayRecordsAudit(project, runDir string) Audit {
	if project == "" {
		return Audit{Check: "stray-records", Verdict: "SKIP", Detail: "no project directory to walk"}
	}
	captured, _ := filepath.Abs(runDir)
	var strays []string
	_ = filepath.WalkDir(project, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil //nolint:nilerr // an unreadable subtree is not this audit's finding
		}
		switch d.Name() {
		case ".git", "node_modules", "cache":
			return filepath.SkipDir
		case "records":
		default:
			return nil
		}
		parent := filepath.Dir(path)
		if abs, _ := filepath.Abs(parent); abs == captured {
			return filepath.SkipDir // the run being captured
		}
		// A REAL RUN DIRECTORY, just not this one. Every past run has records/, and reporting
		// those would bury the finding under the corpus.
		if _, e := os.Stat(filepath.Join(parent, "inputs", "run-config.json")); e == nil {
			return filepath.SkipDir
		}
		if _, e := os.Stat(filepath.Join(parent, "blue")); e == nil {
			return filepath.SkipDir
		}
		shards, _ := filepath.Glob(filepath.Join(path, "events-*.jsonl"))
		if len(shards) > 0 {
			rel, rerr := filepath.Rel(project, path)
			if rerr != nil {
				rel = path
			}
			strays = append(strays, fmt.Sprintf("%s (%d shard(s))", filepath.ToSlash(rel), len(shards)))
		}
		return filepath.SkipDir
	})
	sort.Strings(strays)
	if len(strays) > 0 {
		return Audit{Check: "stray-records", Verdict: "FAIL",
			Detail: fmt.Sprintf("%d event shard tree(s) outside any run directory — seat work that did not reach the run it was dispatched into: %s. A relative --run resolves against the SEAT's working directory; the engine passes the absolute path recorded in inputs/run-config.json",
				len(strays), strings.Join(strays, ", "))}
	}
	return Audit{Check: "stray-records", Verdict: "PASS",
		Detail: "no event shards outside a run directory"}
}

// RecordParityAudit holds what the record owes the bench to the record: the chair's POSITION at each
// sitting that withheld a PASS its plan permitted (record.PositionSittings), and each party's CLOSING
// on each docketing (record.ClosingDuties) — the gap's minting lens and blue-respond.
//
// THIS AUDIT IS THE READER THAT REPORTS AN ARGUMENT THAT WAS NOT FILED. The transcript prints the
// positions and closings that were, and gives a sitting that holds none no section.
//
// THIS IS THE ONE PLACE A CHAIR SITTING WITHOUT ITS POSITION IS REPORTED. The chair's work list holds
// no item for it, so its position is held by its prompt and by this audit. The chair's first sitting
// after the bench overruled its position owes the verdict instead (gblock, 2026-10-30): its work list
// says so, and this audit fails one that closed without it. A party's closing is on its work list as well, as the PASS
// gate's blocker, and this audit is the after-the-run reading of the same duty.
//
// A CLOSING IS SHORT ONLY WHERE ITS PARTY SAT. A party that sat through a sitting opened after the
// docketing, and before the ruling, and filed none is short. One that never sat in that window —
// the run ended, or the terminal bench ruled first — had no sitting to file in, and is named, not
// failed.
//
// THE SITTINGS ARE READ AFTER THE RUN (record.AfterTheRun). Capture reads a finished run, so no
// sitting is still in flight and the end of the record closes every sitting nothing else closed: a
// last sitting that filed no position before the run ended filed none, though a headless agent
// left no stop to close it.
func RecordParityAudit(run record.Run) Audit {
	fam, err := record.FamilyOf(run)
	if err != nil {
		return Audit{Check: "record-parity", Verdict: "FAIL", Detail: "the record could not be read: " + err.Error()}
	}
	return recordParityAt(fam.Events, fam.At, record.AfterTheRun)
}

// recordParityAt is the audit at a stated reading.
//
// A SITTING READ AS UNRESOLVED IS NEVER JUDGED. Its missing act may belong to a sitting still in
// flight, so it is NOT MEASURED, never a finding: with no closed sitting short, the audit is SKIP
// and names it, and a FAIL names it too. Read after the run, no sitting is unresolved.
func recordParityAt(evs []*record.Event, win record.WindowIndex, when record.ReadWhen) Audit {
	rows := record.PositionSittings(evs, win, when)
	duties := record.ClosingDuties(evs, win)
	if len(rows) == 0 && len(duties) == 0 {
		return Audit{Check: "record-parity", Verdict: "SKIP", Detail: "no chair sitting and no docketed gap on record"}
	}
	owed, closings, notOwed := 0, 0, 0
	var short, unmeasured, unreached []string
	for _, p := range rows {
		name := fmt.Sprintf("red-chair sitting %d, which withheld the PASS its plan permitted, ", p.Ordinal)
		switch p.State {
		case record.PositionNotOwed:
			continue
		case record.PositionUnresolved:
			unmeasured = append(unmeasured, name+"shows no position so far — NOT MEASURED: the record cannot close it (red-chair has not opened a sitting since and its agent's stop is not on the record)")
		case record.PositionMissing:
			short = append(short, name+"filed no position")
		case record.PositionVerdictMissing:
			short = append(short, fmt.Sprintf("red-chair sitting %d, its first after the bench overruled its position %s, closed with no verdict — the overruled objection is decided, and that sitting owed the verdict its plan permitted", p.Ordinal, p.Overruled))
		}
		owed++
	}
	for _, d := range duties {
		name := fmt.Sprintf("%s's closing on %s (docket motion %s)", d.Party, d.GapID, d.Motion)
		switch d.State {
		case record.ClosingNotOwed:
			notOwed++
		case record.ClosingFiled:
			closings++
		case record.ClosingMissing:
			short = append(short, name+" was not filed, and "+d.Party+" sat since the docketing")
		case record.ClosingAwaited:
			unreached = append(unreached, name+": "+d.Party+" sat through no sitting between the docketing and the ruling or the run's end")
		}
	}
	if len(short) > 0 {
		return Audit{Check: "record-parity", Verdict: "FAIL", Detail: strings.Join(append(short, unmeasured...), "; ")}
	}
	if len(unmeasured) > 0 {
		return Audit{Check: "record-parity", Verdict: "SKIP", Detail: strings.Join(unmeasured, "; ")}
	}
	detail := fmt.Sprintf("%d chair sitting(s), %d of them owing a position for a PASS withheld, each with its position; %d closing(s) owed on docketed gaps, each filed; %d not owed (the gap closed after its docketing, or its minter is no cast lens)",
		len(rows), owed, closings, notOwed)
	if len(unreached) > 0 {
		detail += "; not reached: " + strings.Join(unreached, "; ")
	}
	return Audit{Check: "record-parity", Verdict: "PASS", Detail: detail}
}

// ---- AUDIT 7: back-fill ----

// BackfillAudit asks whether a seat RECORDED AS IT WORKED or narrated at the end, and it asks the
// RECORD, not a transcript.
//
// WHAT THIS REPLACES, AND WHY IT IS A DELETION RATHER THAN A REWRITE. `RecordJoinAudit` scraped
// Bash command strings out of agent transcripts with a regex and flagged every event whose
// (seat, verb) pair it failed to reconstruct. It had FIVE independent ways to be wrong (#223):
// `motion` was missing from its role alternation, so 100% of the adjudication mechanism was
// invisible; `(\S+)` captured `--seat-id "x"` with the quotes; FindStringSubmatch saw only the
// first invocation in a batched command, which debate.js encourages; the event type is not the
// verb for six types (friction-none, anchor, blue_edit, the two motion subcommands, and friction
// when emitted as a side effect of mint/compute/cite); and `[\s\S]*?` spanned invocations, so one
// call's verb could pair with another call's seat and MANUFACTURE an invocation that never
// happened — masking a genuine orphan. Four modes fail loudly; the fifth fails silently, which is
// the one that matters.
//
// The check it was straining for could not pay for that. `feov-record` is the sole validated
// writer, so an event exists only because the command ran — fabrication is not a reachable state,
// and the join spent all its complexity re-deriving from prose what the record already holds as
// fields. This is `facts-are-fields`: the fact was composed into a command string at one end and
// recovered by regex at the other, and the recovery failed silently in both directions.
//
// WHAT SURVIVED. Exactly one signal: back-fill. A seat that does its work and then dumps every
// record call in a burst at the end has produced a retrospective narration, and any audit that
// treats those events as contemporaneous evidence is reading a story. That signal does not need
// the transcript, because `ts` is stamped BY THE TOOL at write time and `register` is
// every seat's first record action — so the record already knows when a seat started and when it
// wrote.
//
// AND IT IS NOT THE SAME INSTRUMENT, WHICH IS STATED HERE RATHER THAN GLOSSED. The old warning
// measured POSITION among a seat's tool calls ("all its record calls are in the final block");
// this measures ELAPSED TIME between registering and recording. They detect the same behaviour
// from different evidence, and this one is blind to a seat that back-filled inside a window too
// short to resolve. It is also blind in the other direction the old one was not: a seat that made
// no non-record tool calls at all looks identical to a seat that worked and narrated. Where the
// two disagree, believe neither without looking at the run.
const (
	// backfillMinEvents mirrors the old check's floor: below a handful of events, "clustered"
	// and "this seat only had three things to say" are the same shape.
	backfillMinEvents = 5
	// backfillSpanRatio is the fraction of a seat's lifetime its recording burst must fit
	// inside before the burst reads as narration rather than as work. A seat that registers,
	// works for nine minutes and records for one is at 0.1.
	backfillSpanRatio = 0.15
)

func BackfillAudit(run record.Run) Audit {
	fam, err := record.FamilyOf(run)
	if err != nil {
		// ABSENT IS NOT CLEAN. A run whose record cannot be read has not been audited, and
		// saying so is the whole point — the shape this audit replaced reported a plausible
		// zero when its input went missing.
		return Audit{Check: "backfill", Verdict: "SKIP", Detail: "the record could not be read: " + err.Error()}
	}
	type seatSpan struct {
		registered            time.Time
		firstWrite, lastWrite time.Time
		n                     int
	}
	seats := map[string]*seatSpan{}
	order := []string{}
	unparsed := 0
	for _, e := range fam.Events {
		ts, perr := record.ParseStamp(e.GetTs())
		if perr != nil {
			unparsed++
			continue
		}
		s := seats[e.GetSeatId()]
		if s == nil {
			s = &seatSpan{}
			seats[e.GetSeatId()] = s
			order = append(order, e.GetSeatId())
		}
		if e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER {
			// A seat can register more than once — a re-dispatch rotates the nonce
			// (RegisterSeat). The EARLIEST register is when this seat first existed.
			if s.registered.IsZero() || ts.Before(s.registered) {
				s.registered = ts
			}
			continue
		}
		if s.firstWrite.IsZero() || ts.Before(s.firstWrite) {
			s.firstWrite = ts
		}
		if ts.After(s.lastWrite) {
			s.lastWrite = ts
		}
		s.n++
	}
	if len(seats) == 0 {
		return Audit{Check: "backfill", Verdict: "SKIP", Detail: "no events on the record"}
	}
	var warns []string
	measured := 0
	for _, id := range order {
		s := seats[id]
		if s.n < backfillMinEvents || s.registered.IsZero() {
			continue
		}
		lifetime := s.lastWrite.Sub(s.registered)
		if lifetime <= 0 {
			continue
		}
		measured++
		burst := s.lastWrite.Sub(s.firstWrite)
		if float64(burst) <= backfillSpanRatio*float64(lifetime) {
			warns = append(warns, fmt.Sprintf("%s: %d event(s) written in %s at the end of a %s sitting — recorded after the fact, so these events are narration, not contemporaneous evidence",
				id, s.n, burst.Round(time.Millisecond), lifetime.Round(time.Millisecond)))
		}
	}
	parts := []string{fmt.Sprintf("%d seat(s) on the record; %d with enough events (>=%d) and a known start to measure", len(seats), measured, backfillMinEvents)}
	if unparsed > 0 {
		// Loud, not folded into the pass: an unparseable stamp means this audit did not see
		// that event at all, and a quieter version of this line is how a broken clock reads
		// as a clean board.
		parts = append(parts, fmt.Sprintf("NOT MEASURED: %d event(s) carried an unparseable ts", unparsed))
	}
	switch {
	case len(warns) > 0:
		parts = append(parts, warns...)
	case measured == 0:
		parts = append(parts, "nothing to measure — no seat wrote enough events to distinguish a burst from a short sitting")
	default:
		parts = append(parts, "every measured seat recorded across its sitting rather than in a closing burst")
	}
	v := "PASS"
	if len(warns) > 0 {
		v = "WARN"
	}
	return Audit{Check: "backfill", Verdict: v, Detail: strings.Join(parts, "\n  ")}
}

// ---- AUDIT 8: attestation integrity ----

// Claim is one anchored closure (seat|tool|target) spot-checked against the closer's own tool calls.
type Claim struct {
	ID, Seat, Tool, Target, Why string
	// Anchors are the ids the closure names as verified with (Close.anchors).
	Anchors []string
	// Agent is the harness id of the agent whose sitting recorded the closure — the agent that opened
	// the sitting the close event is stored in. "" when the close is in no sitting, or its opening
	// names no agent.
	Agent string
	// Closes are the stamps of every close the closer's sitting recorded: each is a call the audit
	// does not count, because naming an id in a close is the claim, not the read behind it.
	Closes []time.Time
}

// AttestationAudit holds each anchored closure to the acts of the sitting that made it.
//
// IT READS THE RECORD, not red/archive.md. The file was the source until the archive became a
// rendered projection, and `setup` stopped creating it — so this audit had been reading "" and
// returning SKIP: "no archive records to reconcile" on every record-mode run. It reconciled
// nothing, in silence, for as long as the record tier has existed.
//
// The old path was also the class in miniature: a close event carries anchor_seat, anchor_tool
// and anchor_target as FIELDS; archiveMD concatenates them into a pipe-delimited line; and this
// function split that line back apart on "|" and re-derived the gap id with `^(R\d+-\d+)`.
// Fields to string to regex to fields, with a markdown file in the middle purely as a courier.
//
// THE CLOSER'S SITTING, AND NOT THE CLOSE. Each anchor must appear in a call of the closer's own
// sitting — the transcript of the agent that opened the sitting the close is stored in — that READ
// it (it came back in the call's result) or NAMED it (it is in the call's input), other than the
// calls that recorded the sitting's closes. The audit searched every call of every transcript, and
// the close itself carries every anchor it names, so no anchored closure could fail: m30's
// G-968365ad named two computations and reconciled on its own `close --anchor`. A close is located in its
// sitting's transcript by its stamp, inside the window from the call to its result; a close no
// call's window holds cannot be told apart from the reads, and its closure is NOT MEASURED.
func AttestationAudit(run record.Run, transcripts []record.Transcript, sampleFloor int) Audit {
	fam, err := record.FamilyOf(run)
	if err != nil {
		return Audit{Check: "attestation-integrity", Verdict: "SKIP", Detail: "the record could not be read: " + err.Error()}
	}
	byID := map[int64]*record.Event{}
	closesIn := map[int64][]time.Time{} // sitting → the stamps of the closes recorded in it
	closeEvent := map[*recordpb.Close]*record.Event{}
	for _, e := range fam.Events {
		w := fam.At.Of(e)
		byID[w.ID] = e
		if c := e.GetClose(); c != nil {
			closeEvent[c] = e
			// An unreadable stamp is kept as the zero time, which sittingReads refuses to measure:
			// dropped, its close would go uncounted and the close would reconcile itself again.
			ts, _ := record.ParseStamp(e.GetTs())
			closesIn[w.SittingID] = append(closesIn[w.SittingID], ts)
		}
	}
	var claims []Claim
	for _, g := range fam.Gaps {
		if g.Closure == nil {
			continue
		}
		// A CARRIED closure attests nothing — it is an earlier sitting's verification restated,
		// and holding a seat to a tool call it never made in this sitting is a false finding.
		if g.Closure.GetCarriedFrom() != "" {
			continue
		}
		seat, tool, target := g.Closure.GetAnchorSeat(), g.Closure.GetAnchorTool(), g.Closure.GetAnchorTarget()
		if seat == "" || tool == "" || target == "" {
			continue
		}
		c := Claim{ID: g.ID, Seat: seat, Tool: tool, Target: target, Anchors: g.Closure.GetAnchors()}
		if e := closeEvent[g.Closure]; e != nil {
			sitting := fam.At.Of(e).SittingID
			if opener := byID[sitting]; opener != nil {
				c.Agent = recordpb.AgentOpening(opener)
			}
			c.Closes = closesIn[sitting]
		}
		claims = append(claims, c)
	}
	if len(claims) == 0 {
		return Audit{Check: "attestation-integrity", Verdict: "SKIP", Detail: "no anchored closures (all carried, or anchors absent)"}
	}
	pathOf := map[string]string{}
	for _, tr := range transcripts {
		pathOf[tr.AgentID] = tr.Path
	}
	sampled := sampleClaims(claims, sampleFloor)
	var unreconciled, unmeasured []Claim
	for _, c := range sampled {
		needles, byID := needlesFor(c)
		if len(needles) == 0 {
			// NOT MEASURED IS NOT A FINDING, and folding it into one is how this audit came to
			// accuse honest seats. See needlesFor.
			c.Why = "no anchor named and no distinctive fragment in the target — nothing to reconcile AGAINST"
			unmeasured = append(unmeasured, c)
			continue
		}
		reads, why := sittingReads(c, pathOf)
		if why != "" {
			c.Why = why
			unmeasured = append(unmeasured, c)
			continue
		}
		missing := ""
		for _, n := range needles {
			found := false
			for _, k := range reads {
				if strings.Contains(k, n) {
					found = true
					break
				}
			}
			if !found {
				missing = n
				break
			}
		}
		if missing != "" {
			if byID {
				c.Why = "the closure cites " + missing + " but no call of the closer's sitting, other than its closes, read or named it"
			} else {
				c.Why = "no call of the closer's sitting, other than its closes, touches " + missing
			}
			unreconciled = append(unreconciled, c)
		}
	}
	// THE UNMEASURED ARE COUNTED OUT LOUD, in every verdict, because the alternative is the
	// failure this audit is for one level up: a sample of 3 where 2 could not be measured reads
	// as "1/1 reconciled" unless the other 2 are on the page.
	note := ""
	if len(unmeasured) > 0 {
		ml := make([]string, len(unmeasured))
		for i, u := range unmeasured {
			ml[i] = fmt.Sprintf("    - %s (%s | %s): %s", u.ID, u.Seat, u.Tool, u.Why)
		}
		note = fmt.Sprintf("\n    %d of the %d sampled could NOT BE MEASURED either way (not a finding against them):\n%s",
			len(unmeasured), len(sampled), strings.Join(ml, "\n"))
	}
	if len(unreconciled) > 0 {
		ul := make([]string, len(unreconciled))
		for i, u := range unreconciled {
			ul[i] = fmt.Sprintf("    - %s (%s | %s): %s", u.ID, u.Seat, u.Tool, u.Why)
		}
		detail := fmt.Sprintf("%d/%d sampled closure(s) NOT reconcilable against the closer's sitting — a claimed act with no matching tool call:\n%s\n    This rules on whether the RECORD IS HONEST, never on the merits of the gap.%s",
			len(unreconciled), len(sampled), strings.Join(ul, "\n"), note)
		return Audit{Check: "attestation-integrity", Verdict: "FAIL", Detail: detail, Unreconciled: unreconciled}
	}
	return Audit{Check: "attestation-integrity", Verdict: "PASS",
		Detail: fmt.Sprintf("%d/%d anchored closure(s) sampled and reconciled against the closer's own tool calls%s",
			len(sampled)-len(unmeasured), len(claims), note)}
}

// sittingReads is the text of every call of c's sitting other than its closes — each call's input
// and its result — or why the sitting cannot be read: no agent to name it, no transcript of it
// readable here, or a close of the sitting no call's window holds.
func sittingReads(c Claim, pathOf map[string]string) (reads []string, why string) {
	if c.Agent == "" {
		return nil, "the close is in no sitting whose opening names an agent, so no transcript is the closer's"
	}
	path, ok := pathOf[c.Agent]
	if !ok {
		return nil, "the closer's sitting (agent " + c.Agent + ") has no transcript readable here"
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, "the closer's transcript (agent " + c.Agent + ") is not readable: " + err.Error()
	}
	calls := callsOf(string(b))
	closing := make([]bool, len(calls))
	for _, ts := range c.Closes {
		if ts.IsZero() {
			return nil, "a close the sitting recorded carries no readable stamp, so its closes cannot be told from its reads"
		}
		held := false
		for i, k := range calls {
			if k.holds(ts) {
				closing[i], held = true, true
			}
		}
		if !held {
			return nil, fmt.Sprintf("a close the sitting recorded at %s is in no call of agent %s's transcript, so its closes cannot be told from its reads", ts.Format(time.RFC3339Nano), c.Agent)
		}
	}
	for i, k := range calls {
		if !closing[i] {
			reads = append(reads, k.input, k.result)
		}
	}
	return reads, ""
}

// toolCall is one tool call of a transcript: its input and its result as JSON text, and the window
// from the line that made the call to the line that carried its result.
type toolCall struct {
	input, result string
	// from is the stamp of the call's line; to is its result's, zero while no result line names
	// the call — a call still running holds every later moment.
	from, to time.Time
}

// holds reports whether ts falls inside the call's window.
func (k toolCall) holds(ts time.Time) bool {
	if k.from.IsZero() || ts.Before(k.from) {
		return false
	}
	return k.to.IsZero() || !ts.After(k.to)
}

// callsOf reads a transcript's tool calls, in the order it made them, each joined to its result by
// the result's tool_use_id. A line with no timestamp leaves its stamp zero: a call so stamped holds
// no moment, and a result so stamped leaves its call's window open.
func callsOf(transcript string) []toolCall {
	var calls []toolCall
	at := map[string]int{}
	for _, l := range strings.Split(transcript, "\n") {
		if strings.TrimSpace(l) == "" {
			continue
		}
		var j map[string]any
		if json.Unmarshal([]byte(l), &j) != nil {
			continue
		}
		m, _ := j["message"].(map[string]any)
		if m == nil {
			continue
		}
		content, ok := m["content"].([]any)
		if !ok {
			continue
		}
		stamp, _ := time.Parse(time.RFC3339Nano, strOf(j["timestamp"]))
		for _, ci := range content {
			c, ok := ci.(map[string]any)
			if !ok {
				continue
			}
			switch strOf(c["type"]) {
			case "tool_use":
				// ONE CALL PER ID. The harness writes each request's payload into the transcript again
				// (an `api-request-blob` line, with no timestamp), so a call and its result recur;
				// the first line carrying the call is the one that made it.
				id := strOf(c["id"])
				if _, seen := at[id]; seen && id != "" {
					continue
				}
				input := c["input"]
				if input == nil {
					input = map[string]any{}
				}
				at[id] = len(calls)
				calls = append(calls, toolCall{input: jsStringify(input), from: stamp})
			case "tool_result":
				if i, ok := at[strOf(c["tool_use_id"])]; ok && calls[i].to.IsZero() {
					calls[i].result, calls[i].to = jsStringify(c["content"]), stamp
				}
			}
		}
	}
	return calls
}

// sampleClaims mirrors JS: all when ≤ floor, else index i where i % ceil(n/floor) === 0.
func sampleClaims(claims []Claim, floor int) []Claim {
	if len(claims) <= floor {
		return claims
	}
	step := ceilDiv(len(claims), floor)
	var out []Claim
	for i, c := range claims {
		if i%step == 0 {
			out = append(out, c)
		}
	}
	return out
}

func ceilDiv(a, b int) int { return (a + b - 1) / b }

var reNonToken = regexp.MustCompile(`[\s|]+`)

// THE ANCHORS ARE A FIELD, AND THE AUDIT JOINS ON IT.
//
// MEASURED 2026-08-22, and it is the whole reason needlesFor exists. Two closures anchored
// `report text at line 103` and `report text at line 62` — precise, human-meaningful, naming the
// exact line. mostDistinctive found no fragment longer than six characters in either (`report` is
// exactly six), returned "", and both were filed as claims with no matching tool call, under an
// audit whose own message says "This rules on whether the RECORD IS HONEST". Honesty was being
// decided by word length.
//
// An anchor id is exact, and the closure now carries the ones it was verified with as a field the
// write refused unless the record holds each (Close.anchors). Prose is what a human reads; the id
// is what a machine joins on — never re-derived from --verified-with's words, where an id typed in
// another shape fell through to the prose path and could pass on an unrelated long word.

// needlesFor returns the tokens a claim can be reconciled by, and whether they came from the
// record's identifiers or from prose.
//
// EVERY CITED ID MUST APPEAR, not merely one: a seat naming an id it never touched is exactly the
// dishonesty this audit looks for, and ids are exact enough that requiring all of them cannot
// produce the false accusations the prose path did.
//
// The prose fallback keeps its six-character floor deliberately. Lowering it to admit `report`
// would make the needle match every `show report` call in the run — turning a false FAIL into a
// false PASS, which is the worse direction for an honesty check. Prose that yields no needle is
// reported as NOT MEASURED instead.
func needlesFor(c Claim) (needles []string, fromID bool) {
	if len(c.Anchors) > 0 {
		return c.Anchors, true
	}
	if n := mostDistinctive(c.Target); n != "" {
		return []string{n}, false
	}
	return nil, false
}

// mostDistinctive mirrors JS: split target on /[\s|]+/, keep fragments longer than 6, longest first.
func mostDistinctive(target string) string {
	frags := reNonToken.Split(target, -1)
	best := ""
	for _, f := range frags {
		if len(f) > 6 && len(f) > len(best) {
			best = f
		}
	}
	return best
}

// jsStringify serializes a value as JSON with HTML escaping OFF, matching JS JSON.stringify
// (used only as a haystack for Contains, so key order is not load-bearing here).
func jsStringify(v any) string {
	var sb strings.Builder
	enc := json.NewEncoder(&sb)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
	return strings.TrimRight(sb.String(), "\n")
}

// ---- AUDIT 9: model-tier ----

// THE RECORD'S OWN ANSWER OUTRANKS THE PRICE HEURISTIC.
//
// This audit used to be transcript-only: classify each seat from its prompt head, price its turns,
// and compare the resulting TIER to the configured one. That reading graded the measured incident
// a WARNING and let capture exit 0 — `claude-fable-5` was configured and `claude-opus-4-8` was
// served, which is CHEAPER, and cheaper had been filed as "verification may be discounted".
//
// The framing was wrong for a research debate. What a run's tier decides is not the bill, it is
// how strong the adversary arguing each side actually was; a weaker model silently standing in is
// the worse of the two failures, not the softer one. So a substitution FAILS in either direction.
//
// And it no longer has to be inferred from a price at all. `register` records each seat's AGENT ID,
// which names its trajectory, and the trajectory is where the harness DECLARES a swap by naming
// both ends — so a declared substitution is a fact this reads rather than a tier it deduces. The
// transcript pass stays for what the record cannot cover: runs whose trajectories are gone,
// and seats that never registered.
func recordTierFindings(run record.Run, model, judgmentModel string) (findings []string, measured, total int) {
	fam, err := record.FamilyOf(run)
	if err != nil {
		return nil, 0, 0
	}
	for _, sm := range record.SeatModels(run, fam) {
		if sm.Class == "" {
			continue
		}
		total++
		if !sm.Measured() {
			continue
		}
		measured++
		configured := model
		if sm.Class == "judgment" {
			configured = judgmentModel
		}
		if configured == "" {
			continue
		}
		want, got := modeltier.Of(configured), modeltier.Of(sm.Served)
		switch {
		case sm.Substituted():
			findings = append(findings, fmt.Sprintf("FAIL %s (%s) asked for %s and was answered by %s — the harness declared the substitution",
				sm.SeatID, sm.Class, sm.Requested, sm.Served))
		case want != got:
			findings = append(findings, fmt.Sprintf("FAIL %s (%s) was answered by %s (%s tier) against a configured %s (%s tier)",
				sm.SeatID, sm.Class, sm.Served, got, configured, want))
		}
	}
	return findings, measured, total
}

func ModelTierAudit(run record.Run, transcripts []record.Transcript) Audit {
	model, judgmentModel := cost.TierConfig(run)
	if model == "" && judgmentModel == "" {
		return Audit{Check: "model-tier", Verdict: "SKIP", Detail: "the run config names no models, so there is no configured tier to check the served one against"}
	}
	// THE EPOCH COMES FROM THE RECORD, per transcript: the agent whose stop named it joins to the
	// act that bound it to its seat (record.SeatBindings), and that act's window says which chair
	// sitting it sat in. Without it every sitting of a seat folds into epoch 0 and DedupTierFindings
	// reports one finding for what may be several — so a record that cannot be read is a FAIL, not
	// a table of epoch-0 rows.
	bindings, err := record.SeatBindings(run)
	if err != nil {
		return Audit{Check: "model-tier", Verdict: "FAIL", Detail: "the record could not be read: " + err.Error()}
	}
	var rows []cost.Row
	for _, tr := range transcripts {
		b, err := os.ReadFile(tr.Path)
		if err != nil {
			continue
		}
		row := cost.ScanTranscript(string(b))
		row.Epoch = bindings[tr.AgentID].Epoch
		rows = append(rows, row)
	}
	findings := cost.DedupTierFindings(cost.TierMismatch(rows, model, judgmentModel))
	fails, warns := 0, 0
	for _, f := range findings {
		if f.Verdict == "FAIL" {
			fails++
		} else if f.Verdict == "WARN" {
			warns++
		}
	}
	recFindings, measured, seats := recordTierFindings(run, model, judgmentModel)
	fails += len(recFindings)

	tiers := fmt.Sprintf("configured bulk=%s, judgment=%s", orQ(model), orQ(judgmentModel))
	// THE COVERAGE LINE IS NOT DECORATION. "every seat ran on its configured tier" over a run
	// where nothing was measured is the plausible zero this audit exists to stop printing, so the
	// served side always says how many seats it actually looked at.
	served := fmt.Sprintf("served measured on %d of %d tier-bound seat(s) on the record", measured, seats)
	if seats > 0 && measured == 0 {
		served = fmt.Sprintf("served NOT MEASURED on any of %d tier-bound seat(s) — this run predates the field or its seats' trajectories were unreadable, and that is NOT the same as a run that matched its configuration", seats)
	}
	var fs []string
	for _, f := range findings {
		fs = append(fs, f.Verdict+" "+f.Why)
	}
	fs = append(fs, recFindings...)
	detail := tiers + "; " + served
	if len(fs) > 0 {
		detail += "; " + strings.Join(fs, "; ")
	} else if measured > 0 {
		detail += "; every seat was answered by its configured tier"
	}
	v := "PASS"
	if fails > 0 {
		v = "FAIL"
	} else if warns > 0 {
		v = "WARN"
	}
	return Audit{Check: "model-tier", Verdict: v, Detail: detail}
}

func orQ(s string) string {
	if s == "" {
		return "?"
	}
	return s
}

// closeRunLiveMarker removes the run-live marker and SAYS WHAT IT DID, including when it did
// nothing.
//
// Reporting only on removal makes "there was no marker" and "I looked in the wrong place" print
// the same thing: nothing at all. The path is cwd-rooted, so capture invoked from a subdirectory
// finds no marker, removes nothing, and would report a clean run — leaving a file that tells
// every later un-flagged verb it is still inside this run.
//
// Capture is the step that CLOSES a run. It is the wrong place to be quiet about the one piece of
// state that says the run is open.
//
// An error means the marker still holds a row capture meant to take; the line says so and capture
// exits failed.
func closeRunLiveMarker(cwd, runDir string) (string, error) {
	marker := filepath.Join(cwd, ".claude", "run-live.json")
	// THIS RUN'S ROW, AND ONLY THIS RUN'S.
	//
	// The marker was a singleton: one file naming the one open run. Capture removed it by PATH,
	// never asking which run it named — so capturing run A while run B was live lifted B's
	// marker and reported "removed", which reads exactly like the correct outcome. Every later
	// un-flagged verb in B then inferred no run at all, and B's own capture found nothing to
	// close and said so in the words it uses for a run that was already clean.
	//
	// NEARLY DONE, 2026-08-22: a discarded run was about to be captured for its evidence while
	// the next run was eleven minutes into its first epoch. Caught by reading this function
	// rather than by anything in it.
	//
	// A path comparison guarded that. Per-run rows REMOVE it: the row this capture owns is the
	// only one it can take, so there is no longer a case where taking the wrong one is possible
	// and a check is what stands between (#529).
	if _, err := os.Stat(marker); err != nil {
		if os.IsNotExist(err) {
			return "run-live marker: none at " + marker +
				" — either it was already cleared, or capture is running from a directory that is not the project root, in which case the real marker is still there", nil
		}
		return "run-live marker: could not be read at " + marker + ": " + err.Error(), err
	}
	before := len(runlive.ReadRunLive(cwd))
	found, remaining, err := runlive.RemoveRunLiveEntry(cwd, runDir)
	if err != nil {
		// THE RUN IS STILL LIVE, and saying "removed" here is the silence this function exists to end:
		// every hook in the project keeps attributing to a run its capture closed.
		return "run-live marker: this run's entry could NOT be removed from " + marker + ": " + err.Error() +
			" — the run still counts as live; remove its row by hand or re-run capture", err
	}
	// AND EVERY OTHER ROW THAT IS NO LONGER LIVE — a run whose outcome is recorded or whose directory
	// is gone — pruned AFTER this run's own row is taken, so this run is never reported as pruned
	// rather than closed. Named, because a row removed in silence is a run nobody is told about.
	pruned, err := runlive.PruneRunLive(cwd)
	if err != nil {
		return closedLine(found, before, remaining, marker, runDir) + "; the rows no longer live could NOT be pruned (" +
			err.Error() + ") and are still in the marker: " + prunedRuns(pruned), err
	}
	remaining -= len(pruned)
	return closedLine(found, before, remaining, marker, runDir) + prunedLine(pruned), nil
}

// prunedLine names the rows capture pruned besides its own, or nothing when there were none.
func prunedLine(pruned []runlive.Pruned) string {
	if len(pruned) == 0 {
		return ""
	}
	return "; pruned " + plural(len(pruned), "the row", "the rows") + " no longer live: " + prunedRuns(pruned)
}

// prunedRuns names each row and why it is no longer live.
func prunedRuns(pruned []runlive.Pruned) string {
	parts := make([]string, 0, len(pruned))
	for _, p := range pruned {
		parts = append(parts, p.Row.RunDir+" ("+p.State.Why()+")")
	}
	return strings.Join(parts, ", ")
}

// closedLine says what happened to this run's own row.
func closedLine(found bool, before, remaining int, marker, runDir string) string {
	switch {
	case !found && remaining > 0:
		// The file is there and does not mention this run. Not an error and not a clean close:
		// something else is open and this run was never registered in it.
		return "run-live marker: this run has no entry — " + otherRuns(remaining) +
			" still open, and none of them is " + runDir +
			". Nothing was removed; the open run(s) are unaffected"
	case !found && before == 0:
		// PRESENT AND NAMING NOTHING is its own answer, and it must not read like a clean close.
		// It is what an unreadable marker looks like from here, and reporting "nothing to
		// remove" would let a file that no reader can parse pass for an absent one — the same
		// bytes for a broken state and a healthy one.
		return "run-live marker: the file at " + marker + " exists but names NO open run — it is " +
			"empty or unreadable, so nothing could be removed and no verb can infer a run from it. " +
			"Delete it if it is residue"
	case !found:
		return "run-live marker: this run had no entry to remove"
	case remaining > 0:
		return "run-live marker: this run's entry removed — " + otherRuns(remaining) + " still open, so the marker file remains"
	default:
		return "run-live marker: removed"
	}
}

// otherRuns names how many OTHER runs remain open, for a message that has to read naturally
// either way. The package's own plural() takes two literals; this is the one place where the
// count itself has to appear inside them.
func otherRuns(n int) string {
	return fmt.Sprintf("%d other run%s", n, plural(n, "", "s"))
}

// ---- precedent harvest ----

// HarvestResult is the writeScorecards-style report struct for the harvest.
type HarvestResult struct {
	Written bool
	Count   int
	Path    string
	Reason  string
	// Rulings is how many docket and petition rulings the record holds — what the bench decided,
	// as distinct from Count, the `declare` constructions the harvest files. The precedent harvest
	// alone sets it.
	Rulings int
	// EnvelopeClaimed is how many rulings the SEATS said they made: the envelopes' dispositions
	// and petition rulings, which have no field for a declaration. It is a cross-check, not a
	// source, and it is compared to Rulings, the same set on the record — a field rather than a
	// sentence, so a reader compares numbers without parsing prose (facts-are-fields).
	EnvelopeClaimed int
}

var (
	reTrailingSep = regexp.MustCompile(`[\\/]+$`)
	reSep         = regexp.MustCompile(`[\\/]`)
)

func slugOf(run record.Run) string {
	s := reTrailingSep.ReplaceAllString(run.Dir(), "")
	parts := reSep.Split(s, -1)
	return parts[len(parts)-1]
}

type ruling struct {
	kind, gapID, petitioner, disposition, rationale string
}

// rulingsFromRecord reads the rulings THE RECORD HOLDS.
//
// NOT from the harness journal's envelopes — a seat's own account of what it ruled, composed as
// free-text arrays at the end of its sitting and validated by nothing that knows what happened.
// The record holds the same rulings as events, each written through a verb that refused them if
// they were malformed. Asking the less reliable of the two is the defect.
//
// It failed the way this class always fails. A bench that ruled six gaps and listed four
// promoted four, and nothing noticed. A bench that omitted the array promoted nothing, and
// capture reported "precedent harvest: 0 ruling(s)" — the same bytes as an honest run where the
// bench genuinely ruled nothing. And `bench declare` (#361), which has no envelope field at all,
// was unharvestable by construction rather than by omission: the one verb whose entire purpose
// is to state a holding could not reach the place holdings are promoted from.
//
// DELIBERATELY NOT HARVESTED: `motion-rule` on a grade or an inquiry. Those are rulings, and a
// grade holding ("disclosure does not lower likelihood") reads like a rule — but the harvest
// template's `question:` would have to be the motion's ask, which lives on the separate `motion`
// event, and promoting a ruling without the ask it answered is how a holding loses its scope.
// That is a change to WHAT law/proposed contains, not to where the harvest reads; it wants its
// own decision. Named here so the smaller scope is visible rather than silent.
func rulingsFromRecord(evs []*record.Event) []ruling {
	if evs == nil {
		return nil
	}
	// A petition ruling names the motion, not the filer. The filer is the seat that appended
	// the `motion` event, which is the only place that fact exists.
	filedBy := map[string]string{}
	for _, e := range evs {
		if m, ok := recordpb.BodyAs[*recordpb.Motion](e); ok {
			filedBy[m.GetMotionId()] = e.GetSeatId()
		}
	}

	// THE GAP A DISPOSITION SETTLES RIDES THE FILING, never the ruling — a `motion-rule` carries
	// the ask's id and nothing else that identifies its subject matter. record.DocketGapByMotion is
	// that filing join, the one closure_act makes in SQL; harvesting the gap any other way keys every
	// disposition on the empty string, which is the shape of the defect this function was written
	// to escape.
	docketGapOf := record.DocketGapByMotion(evs)

	var out []ruling
	for _, e := range evs {
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_DECLARE:
			// No gap and no fate — that is the point of the verb. The holding IS the ruling.
			d, ok := recordpb.BodyAs[*recordpb.Declare](e)
			if !ok {
				continue
			}
			out = append(out, ruling{
				kind:        "declaration",
				disposition: "declared",
				rationale:   d.GetHolding(),
			})
		case recordpb.EventType_EVENT_TYPE_MOTION_RULE:
			// Only the DOCKET and PETITION subjects reach the precedent harvest: a grade or
			// inquiry ruling answers a motion the motions view renders with its ask beside it.
			//
			// THE SUBJECT IS THE ONEOF, not a string compare. Each subject's ruling carries its
			// verdict on its own arm, so reading the arm IS the subject test — a grade ruling
			// cannot answer either of these, where the old `Str("subject") != "petition"`
			// depended on a field that could disagree with the value beside it.
			r, ok := recordpb.BodyAs[*recordpb.MotionRule](e)
			if !ok {
				continue
			}
			switch ruled := r.GetRuling().(type) {
			case *recordpb.MotionRule_Docket:
				// The bench disposing of a docketed gap. Its argument is `MotionRule.opinion` —
				// the prose channel every subject's ruling carries — and the gap comes from the
				// join above, because the ruling does not hold it.
				out = append(out, ruling{
					kind:        "docket",
					gapID:       docketGapOf[r.GetMotionId()],
					disposition: recordpb.Word(ruled.Docket.GetDisposition()),
					rationale:   r.GetOpinion(),
				})
			case *recordpb.MotionRule_Petition:
				out = append(out, ruling{
					kind:        "petition",
					petitioner:  filedBy[r.GetMotionId()],
					disposition: recordpb.Word(ruled.Petition),
					rationale:   r.GetOpinion(),
				})
			}
		}
	}
	return out
}

// rulingsClaimedByEnvelopes counts what the seats SAID they ruled. Kept only as a cross-check
// against the record: a divergence is a finding about the envelopes, the way record-parity
// already treats one. It is never the harvest's source.
func rulingsClaimedByEnvelopes(results []map[string]any) int {
	n := 0
	for _, r := range results {
		for _, key := range []string{"dispositions", "rulings"} {
			if rs, ok := r[key].([]any); ok {
				n += len(rs)
			}
		}
	}
	return n
}

func HarvestPrecedents(run record.Run, results []map[string]any, lawDir string, evs []*record.Event) HarvestResult {
	slug := slugOf(run)
	// ONLY A `declare` IS A CANDIDATE HOLDING, and every other ruling is dropped here.
	//
	// The harvest used to file EVERY bench ruling into law/proposed as a PERSUASIVE candidate. The
	// corpus that produced: 76 entries, of which 68 stated no rule at all — 35 whose `holding` was
	// a disposition word (`closed`, `carried`) and 33 carrying a placeholder asking a reviewer to
	// supply the rule the harvest could not invent. Six were real. A queue where nine of ten
	// entries cannot be promoted is not a backlog; it is noise that hides the ones that can, and
	// the review gate silts up behind it.
	//
	// The placeholder was itself a fix — it replaced `holding: closed` so an unpromotable ruling
	// would LOOK unpromotable — and it was the right repair to the wrong layer. Making the noise
	// legible does not stop it being filed. A ruling that disposes of a gap is on the record
	// already, where `inquest debate` renders it; copying it into law/ adds a claim that it is
	// candidate LAW, which is the part that was never true.
	//
	// `declare` is different and always was: that verb exists to state a construction, so its text
	// IS the holding and no placeholder is honest there. New rules arrive that way, deliberately.
	// Everything else a run learns about the law lands as a SCOPE-LIMIT proposal against an
	// existing holding (law/README.md) — written by a human who read the case, because that is the
	// one part of this the harvest genuinely cannot do.
	//
	// TWO SETS, AND THE ENVELOPES ARE CHECKED AGAINST THE ONE THEY DESCRIBE. An envelope lists the
	// bench's dispositions and petition rulings; it has no field for a declaration. Compared to the
	// declarations, a run whose bench ruled one docket and declared nothing read "the record holds
	// NO rulings while the envelopes claim 1" — universe m18, with the ruling at event 460.
	all := rulingsFromRecord(evs)
	rulings := declarationsOnly(all)
	onRecord := len(all) - len(rulings)
	claimed := rulingsClaimedByEnvelopes(results)
	res := HarvestResult{Count: len(rulings), Rulings: onRecord, EnvelopeClaimed: claimed}
	if len(rulings) == 0 {
		return res
	}
	if _, err := os.Stat(lawDir); err != nil {
		res.Reason = "no law/ dir at repo root"
		return res
	}
	_ = os.MkdirAll(filepath.Join(lawDir, "proposed"), 0o755)
	out := filepath.Join(lawDir, "proposed", slug+".md")
	var body []string
	body = append(body, "# proposed holdings — "+slug+" [ALL PERSUASIVE — awaiting human review per law/README.md]",
		"",
		"Each entry is the bench's `declare` — a construction the bench stated deliberately, so its text IS",
		"the holding. `facts` and `scope-limits` remain the reviewer's to write from the cited record.",
		"Ordinary rulings are NOT here: they dispose of a gap and are on the record, and a run's other",
		"lessons about the law arrive as scope-limit proposals against an existing holding.", "")
	for i, r := range rulings {
		// A DECLARATION DISPOSES OF NOTHING, so it does not get a disposition's framing. This read
		// "disposition of " with an empty gap id and a source ending in a bare comma — harmless
		// while dispositions filled the file and the only visible shape once they stopped.
		q := "construction declared this run"
		src := slug
		if r.gapID != "" {
			q = "construction declared beside " + r.gapID
			src = slug + ", " + r.gapID
		}
		// A DISPOSITION IS NOT A HOLDING, and writing one into that field made every harvested
		// ruling unpromotable while looking complete.
		//
		// law/README.md asks for `holding: <the rule applied>` and warns that a holding without
		// its factual predicate is not citable. This wrote `holding: closed` — the docket STATUS
		// — so nine harvested rulings across two runs all read "disposition of R1-n / closed",
		// which tells a later bench nothing it could apply to a different gap. The rule the bench
		// actually reasoned to was sitting in `rationale`, unextracted, one line below.
		//
		// The harvest cannot fix that by synthesising a rule: inventing the holding is exactly
		// what it must not do. So it states the disposition as the disposition, and leaves the
		// holding as a placeholder beside `facts` and `scope-limits` — the three things only a
		// human reviewing the record can supply. An unpromotable ruling now LOOKS unpromotable.
		//
		// A `declare` is the exception and always was: that verb's whole purpose is to state a
		// holding, so its text IS the rule and no placeholder is honest there.
		holding := "<reviewer: state the rule this ruling applied — the harvest never invents; " +
			"the reasoning is in `rationale` below>"
		fields := []string{
			"## " + slug + "-" + strconv.Itoa(i+1) + " [PERSUASIVE]",
			"facts: <reviewer: fill from the cited record — the harvest never invents>",
			"question: " + q,
		}
		if r.kind == "declaration" {
			fields = append(fields, "holding: "+r.rationale)
		} else {
			fields = append(fields,
				"holding: "+holding,
				"disposition: "+r.disposition,
				"rationale: "+r.rationale)
		}
		fields = append(fields,
			"scope-limits: <reviewer: state what this assumed>",
			"source: "+src,
			"")
		body = append(body, strings.Join(fields, "\n"))
	}
	if err := os.WriteFile(out, []byte(strings.Join(body, "\n")), 0o644); err != nil {
		// This branch is a WRITE failure and must say so. Reusing the branch above's reason
		// ("no law/ dir at repo root") would send a reader looking for a directory that exists.
		res.Reason = "could not write " + out + ": " + err.Error()
		return res
	}
	res.Written, res.Path = true, out
	return res
}

// precedentLine is the harvest's line in capture's report: what was filed, what the bench ruled,
// and whether the envelopes agree about the rulings.
//
// THE DIVERGENCE HAS TO REACH THE REPORT, and it is a statement about RULINGS. "0 filed" is the
// output of an honest quiet run, of a run whose bench ruled and declared nothing, and of a harvest
// that cannot see the rulings in front of it; the line tells the three apart by saying how many
// rulings the record holds beside how many the envelopes claim.
func precedentLine(prec HarvestResult) string {
	divergence := ""
	if prec.EnvelopeClaimed != prec.Rulings {
		divergence = fmt.Sprintf(" [the record holds %d docket and petition ruling(s) and the envelopes claim %d — the record is the source, and this divergence is the finding]",
			prec.Rulings, prec.EnvelopeClaimed)
	}
	switch {
	case prec.Written:
		return fmt.Sprintf("precedent harvest: %d declaration(s) -> %s (PERSUASIVE, awaiting review)%s", prec.Count, prec.Path, divergence)
	case prec.Count > 0:
		return fmt.Sprintf("precedent harvest: %d declaration(s), %s%s", prec.Count, prec.Reason, divergence)
	case prec.Rulings > 0:
		// A DISPOSITION IS NOT A HOLDING (HarvestPrecedents says why), so a bench that ruled and
		// declared nothing files nothing — and the line says the rulings are there.
		return fmt.Sprintf("precedent harvest: no declaration this run — the record holds %d docket and petition ruling(s), each disposing of its gap or petition and stating no rule, so none is filed%s",
			prec.Rulings, divergence)
	default:
		return "precedent harvest: no rulings this run" + divergence
	}
}

// ---- scorecards ----

// ScorecardResult mirrors the writeScorecards return.
type ScorecardResult struct {
	Written bool
	Rows    int
	Cards   int
	Reason  string
}

func WriteScorecards(run record.Run, results []map[string]any, memoryDir string, fam *record.Family) ScorecardResult {
	if _, err := os.Stat(memoryDir); err != nil {
		return ScorecardResult{Written: false, Reason: fmt.Sprintf("no %s — scorecards need the tracked memory dir", memoryDir)}
	}
	cards := scorecard.Compute(run, results, fam, record.AfterTheRun)
	label := labelOf(run)
	rows := 0
	written := 0
	var failed []string

	// Sorted, so a partial failure names the same cards in the same order every run: a
	// reason that reshuffles between runs is one a reader cannot diff.
	names := make([]string, 0, len(cards))
	for card := range cards {
		names = append(names, card)
	}
	sort.Strings(names)

	for _, card := range names {
		cardRows := cards[card]
		p := filepath.Join(memoryDir, card+"-scorecard.md")
		head := scorecard.CardHeader(card)
		b, err := os.ReadFile(p)
		switch {
		case err == nil:
			head = string(b)
		case !errors.Is(err, fs.ErrNotExist):
			// A CARD THAT EXISTS AND CANNOT BE READ MUST NOT BE OVERWRITTEN. The read
			// failure used to fall through to the fresh header, and the write below then
			// replaced the file with it — so one unreadable moment (a permission change, an
			// I/O error) silently discarded every earlier run's rows. That is the opposite
			// of what this file is for: the series is the memory, and TestWriteScorecardsAppends
			// asserts it is "appended, never overwritten". Absence is the only reason to
			// start from a header, so absence is the only error handled that way.
			failed = append(failed, card+": cannot read "+p+" ("+err.Error()+") — left untouched rather than overwritten with a fresh header")
			continue
		}
		body := trimTrailingNewlines(head) + "\n" + scorecard.RenderCard(cardRows, label)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			// The write's error was discarded while the result reported Written: true —
			// the same shape the law-harvest write above refuses by name.
			failed = append(failed, card+": cannot write "+p+" ("+err.Error()+")")
			continue
		}
		rows += len(cardRows)
		written++
	}

	res := ScorecardResult{Written: len(failed) == 0, Rows: rows, Cards: written}
	if len(failed) > 0 {
		res.Reason = fmt.Sprintf("%d of %d card(s) NOT written: %s", len(failed), len(cards), strings.Join(failed, "; "))
	}
	return res
}

var reTrailNL = regexp.MustCompile(`\n+$`)

// trimTrailingNewlines mirrors JS `head.replace(/\n+$/, '\n')` — collapse a trailing newline run
// to a single '\n'.
func trimTrailingNewlines(s string) string {
	return reTrailNL.ReplaceAllString(s, "\n")
}

// labelOf mirrors JS `runDir.split(/[\\/]/).filter(Boolean).pop() || 'run'`.
func labelOf(run record.Run) string {
	parts := reSep.Split(run.Dir(), -1)
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return "run"
}

// ---- the seats' transcripts ----

// seatTranscripts is every seat transcript the record names that is readable here, and the line
// capture prints about the census.
//
// THE RECORD SAYS WHICH FILES ARE THE RUN'S SEATS. Each agent's stop names its transcript, so
// capture reads those paths and lists no directory: a directory's `agent-*.jsonl` is whatever the
// harness left there, and a resumed run's seats sit in more than one.
//
// EVERY SEAT IT CANNOT READ IS NAMED, with why — a path gone from this box, a stop that named no
// path, an agent with no stop at all. Each is a seat whose spend and turns are missing from what
// follows, and a smaller count with nothing said reads as a smaller run.
func seatTranscripts(run record.Run, fam *record.Family) ([]record.Transcript, string) {
	if fam == nil {
		return nil, "transcripts: NONE READ — the record could not be read, and it is what names each seat's transcript"
	}
	census := record.TranscriptsOf(*fam)
	bindings, _ := record.SeatBindings(run)
	who := func(agent string) string {
		if b, ok := bindings[agent]; ok && b.SeatID != "" {
			return b.SeatID + " (agent " + agent + ")"
		}
		return "agent " + agent + " (bound to no seat)"
	}
	var present []record.Transcript
	var gone []string
	for _, tr := range census.Named {
		if err := tr.Missing(); err != nil {
			gone = append(gone, who(tr.AgentID)+": "+tr.Path)
			continue
		}
		present = append(present, tr)
	}
	line := fmt.Sprintf("transcripts: %d named by the record's sitting closes, %d readable here", len(census.Named), len(present))
	if len(gone) > 0 {
		line += fmt.Sprintf("; %d NOT ON THIS BOX (transcripts are not archived with the record): %s", len(gone), strings.Join(gone, "; "))
	}
	if len(census.Unnamed) > 0 {
		var ws []string
		for _, a := range census.Unnamed {
			ws = append(ws, who(a))
		}
		line += fmt.Sprintf("; %d stopped with NO PATH on the stop: %s", len(ws), strings.Join(ws, "; "))
	}
	if len(census.Sitting) > 0 {
		var ws []string
		for _, a := range census.Sitting {
			ws = append(ws, who(a))
		}
		line += fmt.Sprintf("; %d opened a sitting with NO STOP on the record (killed mid-sitting, or run as a main session, which fires no SubagentStop), so no transcript is named: %s", len(ws), strings.Join(ws, "; "))
	}
	return present, line
}

// ---- tarball ----

// writeTarball packs each transcript under its own file name; an agent's id makes the name unique.
func writeTarball(outPath string, transcripts []record.Transcript) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for _, tr := range transcripts {
		b, err := os.ReadFile(tr.Path)
		if err != nil {
			return err
		}
		if err := tw.WriteHeader(&tar.Header{Name: filepath.Base(tr.Path), Mode: 0o644, Size: int64(len(b))}); err != nil {
			return err
		}
		if _, err := tw.Write(b); err != nil {
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}

// ---- orchestration ----

// ingestSeatTurns reads each seat transcript into per-turn rows on the record and returns the
// line capture prints about it.
//
// IDENTITY IS NOT COPIED IN HERE. A row carries the harness agent id and nothing else; which seat
// that agent is bound to is already on the record — its latest register, else a bracket that names
// a seat — so the views join to seat_of_agent rather than this function restating it. The agent id comes off the transcript
// LINE, which states which agent produced it — see seatturn.Parse.
func ingestSeatTurns(run record.Run, transcripts []record.Transcript) string {
	seats, rows, failed := 0, 0, 0
	for _, tr := range transcripts {
		b, err := os.ReadFile(tr.Path)
		if err != nil {
			failed++
			continue
		}
		agentID, turns := seatturn.Parse(string(b))
		if len(turns) == 0 {
			continue
		}
		n, err := record.AppendSeatTurns(run, agentID, turns)
		if err != nil {
			failed++
			continue
		}
		seats++
		rows += n
	}
	// THE RECORD'S COUNT, NOT THIS INGEST'S. A sitting's close already files its turns, so by
	// capture nearly every row is on the record and an ingest adds none: printing what this pass
	// added read "0 row(s)" on a record holding 596 (universe-m33, m34).
	held, err := record.CountSeatTurns(run)
	if err != nil {
		return fmt.Sprintf("seat turns: NOT MEASURED — the record's rows could not be counted (%v); this capture added %d from %d transcript(s)", err, rows, seats)
	}
	line := fmt.Sprintf("seat turns: %d row(s) on the record, %d of them added by this capture from %d transcript(s)", held, rows, seats)
	if failed > 0 {
		// NAMED, because the alternative is a smaller number that looks like a smaller run.
		line += fmt.Sprintf(" — %d NOT INGESTED (unreadable or refused)", failed)
	}
	return line
}

// Run executes capture: mechanics + the nine audits, writes run-record-audit.md, and returns the
// audits, the report string, and whether any audit FAILed (exit 2).
//
// project is the directory every project-side effect lands in — run-archive/, feov-memory/,
// law/, .claude/run-live.json — and the tree the stray-records audit walks. It is the caller's to
// state: the command passes its working directory, the project the run was set up from, where
// setup wrote the run-live row capture lifts. It is never derived from the run directory's
// ancestors: a directory name such as plugins/ is not unique to this repository (~/.claude holds
// one), and a run under a scratch directory beneath it archived into ~/.claude/run-archive/.
// `now` is injected so the one non-deterministic input — how long ago the record stopped moving —
// is controllable in a test, the same way WriteRunLiveMarker takes its clock.
//
// transcriptDir is the Workflow's directory, read for its journal alone: the seats' transcripts
// are the ones the record names (seatTranscripts).
func Run(run record.Run, transcriptDir, project string, now time.Time) (audits []Audit, report string, exitFail bool, err error) {
	if project == "" {
		return nil, "", false, errors.New("capture: no project directory given — run-archive/, feov-memory/, law/ and .claude/run-live.json have nowhere to go")
	}
	var lines []string

	// Mechanics: journal copy, transcript tarball, cost.md.
	//
	// A MISSING JOURNAL IS A FACT ABOUT HOW THE RUN WAS DRIVEN, NOT A REASON TO LEAVE IT OPEN.
	// This copy used to be the first thing capture did and a hard error when the file was absent,
	// while ReadJournal below already treated the same absence as an empty journal. Capture is the
	// only code that clears a run's live marker, so a run driven outside the Workflow tool — a
	// headless harness, or a workflow killed before it wrote its journal — could never be closed:
	// three such runs sat "live" in .claude/run-live.json on 2026-09-10, blocking plugin updates.
	//
	// Absent is not the same as empty, and the difference is carried rather than inferred:
	// dispatch-parity, the one audit that reads the envelopes, states that the relayed plans were
	// not compared instead of reading an empty list as a clean pass.
	journalPresent := true
	src := filepath.Join(transcriptDir, "journal.jsonl")
	if _, serr := os.Stat(src); os.IsNotExist(serr) {
		journalPresent = false
		lines = append(lines, "journal: none at "+src+" — this run was not driven by the Workflow tool, or its journal was never written; the relayed plans are NOT COMPARED with the record")
	} else if err = copyFile(src, filepath.Join(run.Dir(), "trajectories", "journal.jsonl")); err != nil {
		return nil, "", false, err
	}
	var onRecord *record.Family
	if f, ferr := record.FamilyOf(run); ferr == nil {
		onRecord = &f
	}
	transcripts, census := seatTranscripts(run, onRecord)
	lines = append(lines, census)
	tarPath := filepath.Join(run.Dir(), "trajectories", "agent-transcripts.tar.gz")
	if terr := writeTarball(tarPath, transcripts); terr != nil {
		lines = append(lines, "tarball: FAILED — "+jsSlice(terr.Error(), 200))
	} else {
		lines = append(lines, fmt.Sprintf("tarball: %d transcript(s)", len(transcripts)))
	}
	// THE RECORD ITSELF, to the one place that outlives the container. Reported either way: an
	// archive that silently did not happen is a run whose evidence is gone the next time the
	// container is reclaimed, and nothing would have said so.
	if arc, aerr := ArchiveRecord(run, project); aerr != nil {
		lines = append(lines, "record archive: FAILED — "+jsSlice(aerr.Error(), 200))
	} else {
		rel := arc
		if r, rerr := filepath.Rel(project, arc); rerr == nil {
			rel = r
		}
		if st, serr := os.Stat(arc); serr == nil {
			lines = append(lines, fmt.Sprintf("record archive: %s (%d KB) — the raw record, which research/ does not keep", rel, st.Size()/1024))
		} else {
			lines = append(lines, "record archive: "+rel)
		}
	}

	if line := reapOrphanMirrors(now); line != "" {
		lines = append(lines, line)
	}

	// A MIGRATED RECORD SAYS SO WHERE RUNS ARE GRADED (plans/replay-migration.md §III.5).
	// Three states, three different bytes: a native run carries no line at all; a migrated
	// run names its source and translator; a run whose manifest is present and unreadable
	// reports THAT, because a run claiming to be a migration that cannot say of what must
	// not read as native.
	if line := migrationLine(run.Dir()); line != "" {
		lines = append(lines, line)
	}

	// PER-TURN TELEMETRY ONTO THE RECORD, so the questions asked about it become SQL instead of
	// another transcript scan (#684 F16). Three readers walk these files today for three slices
	// of one question; the turn goes in once and seat_metrics, the time decomposition and cost
	// become views over it.
	//
	// REPORTED, NEVER FATAL. Capture's job is the audits, and a run whose telemetry could not be
	// read is still a run whose integrity must be checked — but a silent skip would leave an
	// empty table that reads exactly like a run with no turns, so the count is always stated.
	lines = append(lines, ingestSeatTurns(run, transcripts))

	costF, cerr := os.Create(filepath.Join(run.Dir(), "cost.md"))
	if cerr != nil {
		return nil, "", false, cerr
	}
	if rerr := cost.Report(transcripts, run, costF); rerr != nil {
		costF.Close()
		lines = append(lines, "cost.md: FAILED — "+jsSlice(rerr.Error(), 200))
	} else {
		costF.Close()
		lines = append(lines, "cost.md: written (telemetry join included)")
	}

	// THE DOCUMENTS ARE RE-ASSEMBLED FROM THE RECORD AS IT STANDS AT THE CLOSE. The bench assembles
	// inside its own sitting, and the run document's tier table reads each seat's trajectory at the
	// path its stop names: the bench's own stop, and its turns, are not on the record until its
	// sitting stops, and a seat whose trajectory cannot be opened from where capture runs is measured
	// from its turns alone. Re-assembling here, after the turns
	// are ingested, is what makes that table the run's and not the moment's. The audits below then
	// screen the documents that ship. A run that never reached assembly has no run.md and is not
	// assembled here: the bench's act is what makes a run's documents exist.
	if written, werr := reportdoc.AssemblyWritten(run); werr != nil {
		lines = append(lines, "documents: whether the bench assembled them could not be read — "+jsSlice(werr.Error(), 200)+"; they are not re-assembled")
	} else if written {
		if a, aerr := reportdoc.Assemble(run); aerr != nil {
			lines = append(lines, "documents: re-assembly FAILED — "+jsSlice(aerr.Error(), 200)+"; they stand as the bench assembled them")
		} else {
			lines = append(lines, fmt.Sprintf("documents: %d re-assembled from the record at the close", a.Documents))
		}
	}

	results := ReadJournal(filepath.Join(run.Dir(), "trajectories"))

	// Record-backed reads, in-process (the JS spawned `chair show` views).
	// The family off the record; a run whose record cannot be read audits with fam == nil,
	// exactly as the nil board did.
	var fam *record.Family
	f, ferr := record.FamilyOf(run)
	if ferr == nil {
		fam = &f
	}

	audits = []Audit{
		LivenessAudit(run, now),
		TelemetryAudit(run, f, ferr),
		ContextUse(transcripts),
		AssemblyScreen(run),
		FootnoteIntegrity(run),
		StrayRecordsAudit(project, run.Dir()),
		RecordParityAudit(run),
		DispatchParityAudit(run, results, journalPresent),
		// Which dispatches engaged a lens before blue had answered it. See lensturn.go.
		LensTurnAudit(run),
		DocketOrderAudit(run),
		// Which sittings the hook stopped at the per-sitting tool-call limit. See sittinglimit.go.
		SittingLimitAudit(run),
		// Which sittings closed with no admitted envelope. See failedsittings.go.
		FailedSittingsAudit(run),
		BackfillAudit(run),
		AttestationAudit(run, transcripts, 5),
		ModelTierAudit(run, transcripts),
		// THE LAST AUDIT, because it is the only one that EXECUTES anything: a bounded sample of
		// the run's own computations, re-run and compared. See computationrerun.go for why a seat's spot-check
		// could not be the thing that does this.
		ComputationRerunAudit(run, computationRerunSample),
		// Synthesis's declared breadth against the lane seats that actually took theirs — the one
		// run-config field nothing reconciled. See lanecoverage.go.
		LaneCoverageAudit(run),
	}

	sc := WriteScorecards(run, results, filepath.Join(project, "feov-memory"), fam)
	// BOTH LINES WHEN BOTH ARE TRUE. The old form printed the counts OR the reason, so a
	// partial failure — some cards written, one unwritable — showed the cheerful line and
	// swallowed the reason entirely.
	if sc.Cards > 0 || sc.Reason == "" {
		lines = append(lines, fmt.Sprintf("scorecards: %d row(s) across %d card(s) -> feov-memory/", sc.Rows, sc.Cards))
	}
	if sc.Reason != "" {
		lines = append(lines, "scorecards: "+sc.Reason)
	}

	prec := HarvestPrecedents(run, results, filepath.Join(project, "law"), famEvents(fam))
	lines = append(lines, precedentLine(prec))

	markerLine, markerErr := closeRunLiveMarker(project, run.Dir())
	lines = append(lines, markerLine)

	var out []string
	out = append(out,
		"# capture-audit — mechanized post-hoc checks (feov-record capture)",
		"",
		"Presence/consistency tier only: these checks catch a missing line and a self-inconsistent",
		"self-report; a plausible-but-wrong value is vacuity, whose auditor is the next run/",
		"retrospective over these same git-tracked artifacts.",
		"",
	)
	for _, a := range audits {
		out = append(out, fmt.Sprintf("- **%s: %s** — %s", a.Check, a.Verdict, a.Detail))
	}
	out = append(out, "")
	for _, l := range lines {
		out = append(out, "- "+l)
	}
	out = append(out, "")
	report = strings.Join(out, "\n")
	auditPath := filepath.Join(run.Dir(), "run-record-audit.md")
	if werr := os.WriteFile(auditPath, []byte(report), 0o644); werr != nil {
		return nil, "", false, werr
	}
	// THE FOLD COMES AFTER THE AUDIT FILE IS WRITTEN, because the audits are one of the sections it
	// folds. Run before it, it read the file the PREVIOUS capture left: a first capture's run
	// document had no integrity audits at all, and a second carried the first one's. What run.md
	// was given is then said in the audit file and the report, which the fold has already read.
	if msg := foldCaptureArtifacts(filepath.Join(run.Dir(), reportdoc.FileRun), filepath.Join(run.Dir(), "cost.md"), auditPath); msg != "" {
		report = strings.TrimRight(report, "\n") + "\n- " + msg + "\n"
		if werr := os.WriteFile(auditPath, []byte(report), 0o644); werr != nil {
			return nil, "", false, werr
		}
	}
	// THE RUN WAS CAPTURED (runlive.CapturedMark), written after capture's last write to the record: from
	// here the record takes nothing. A mark that could not be written is a failed capture, as an
	// unclosed run-live row is, and the report says which.
	if merr := record.MarkCaptured(run, now.UTC().Format(time.RFC3339)); merr != nil {
		report = strings.TrimRight(report, "\n") + "\n- the run's capture could not be marked (" + merr.Error() + "): its record still takes writes\n"
		_ = os.WriteFile(auditPath, []byte(report), 0o644)
		exitFail = true
	}
	for _, a := range audits {
		if a.Verdict == "FAIL" {
			exitFail = true
		}
	}
	// A run capture could not close is a failed capture: the line above names why.
	if markerErr != nil {
		exitFail = true
	}
	return audits, report, exitFail, nil
}

func copyFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, b, 0o644)
}

// ---- record archive ----

// ArchiveRecord writes the run's RAW RECORD to run-archive/<slug>.tar.gz in the project directory,
// which is the only part of a run that outlives the container.
//
// THE EVIDENCE WAS THE ONE THING NOT BEING KEPT. `research/*/` is gitignored — deliberately, and
// with a commit behind it (ffc4bf4 deleted a tracked research corpus because seats were reading
// test data as prior work). So a run directory survives a SIGTERM and a fast resume, because the
// disk does, and does NOT survive the container being reclaimed. What survived that was only what
// capture promoted: cost.md, the scorecards, the law harvest. Every one of those is a DERIVED
// artifact. The record they were derived from — the thing each audit re-reads, and the only thing
// that can answer a question nobody has thought of yet — was left behind.
//
// WHAT IS IN IT, AND WHAT IS NOT. records/ and computations/: 49 KB gzipped for a fifteen-seat run, so
// keeping every run costs about a megabyte per twenty. NOT cache/ — 7.3 MB of fetched source
// bytes for the same run, re-fetchable, and content-addressed by a sha256 the citation events
// already carry, so its integrity stays checkable without it. NOT the agent transcripts: they are
// an order of magnitude larger than the record and answer a different question (what a seat DID
// rather than what it RECORDED); keeping them is a separate decision and is named here rather
// than made quietly.
//
// It does not reintroduce ffc4bf4's hazard: seats read their own run directory and the inputs
// mirrored into it, never run-archive/, and a gzipped tar is not prose a glob can wander into.
func ArchiveRecord(run record.Run, project string) (string, error) {
	recs := run.Records()
	// SEAL BEFORE WALKING, because the archive is the copy nobody can re-derive. The record runs
	// journal_mode=WAL and nothing checkpoints it during the run, so a third of a finished run's
	// events sit in record.db-wal. Shipping the two files side by side means any later reader that
	// takes only the database — the natural thing to do with a tarball — gets a run that is
	// smaller, self-consistent and wrong. recordsql.SealedBytes folds the log in, refuses where it
	// did not take, and reads the database with this process's connection RELEASED: a descriptor
	// opened beside a live connection drops that connection's POSIX locks on the file when it
	// closes — the -shm locks everywhere, the database file's off Linux. So the
	// database's bytes come from there, and the walk below never opens a database file itself —
	// the sealed log is empty and the shared memory is process scratch, so neither is shipped.
	dbPath := filepath.Join(recs, "record.db")
	var dbBytes []byte
	if _, err := os.Stat(dbPath); err == nil {
		b, err := recordsql.SealedBytes(dbPath)
		if err != nil {
			return "", fmt.Errorf("refusing to archive %s: %w", dbPath, err)
		}
		dbBytes = b
	}
	type archived = archive.Entry
	var files []archived
	add := func(root, prefix string) error {
		return filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil // a missing computations/ is ordinary; a missing records/ is caught below
			}
			rel, rerr := filepath.Rel(root, p)
			if rerr != nil {
				return rerr
			}
			if p == dbPath || p == dbPath+"-wal" || p == dbPath+"-shm" {
				return nil
			}
			files = append(files, archived{Name: prefix + filepath.ToSlash(rel), Path: p})
			return nil
		})
	}
	if dbBytes != nil {
		files = append(files, archived{Name: "records/record.db", Data: dbBytes})
	}
	if err := add(recs, "records/"); err != nil {
		return "", err
	}
	shards := len(files)
	// THE COMPUTATION STORE AS THE RECORD NAMES IT (archive.Store): its staging and any directory no
	// event names stay behind.
	store, err := archive.Store(run)
	if err != nil {
		return "", fmt.Errorf("listing the computation store: %w", err)
	}
	files = append(files, store...)
	// THE RUN'S TERMS, because without them the archive is BROKEN rather than merely thin.
	//
	// inputs/run-config.json is not staged reference material a seat read once. It is read at READ
	// time by a dozen packages: record.RunParams hands k-max, mint-budget, max-epochs and the
	// convergence fraction to the dispatch verb AND the write path; sittingcap reads the
	// per-sitting call limit; modeltier and cost read the configured tiers; capture's own lane
	// coverage reads the declared lane count. Measured on an archived run against the same run
	// live: `ops tiers` printed configuredBulk and configuredJudgment from the live directory and
	// simply OMITTED both from the archive — so a run whose configuration was thrown away reads
	// exactly like a run that had none.
	//
	// It is 1,886 bytes against 49 KB of records, so the cost is not the question. The reason it
	// was missing is that the exclusions above were argued (cache/ re-fetchable, transcripts a
	// different question) and this file was never considered at all.
	//
	// The bulky inputs stay out on the stated line: run-config.json is read to OPERATE on the
	// record, while the law mirror is read by seats during the run and is provenance rather than
	// function. Carrying provenance is a separate decision, and naming it here is how it stays one.
	if cfg := filepath.Join(run.Dir(), "inputs", "run-config.json"); func() bool { st, e := os.Stat(cfg); return e == nil && !st.IsDir() }() {
		files = append(files, archived{Name: "inputs/run-config.json", Path: cfg})
	}

	// THE STAGED CORPUS IS HASHED, NOT CARRIED (gblock, 2026-09-19).
	//
	// The law mirror is provenance rather than function: read by seats during the run, by nothing
	// after it. Hashing is not a weaker answer here, because feov-memory/ is tracked in the
	// repository. The digest names the bytes a run saw, and git already holds them; an audit that
	// needs the prose can recover it, and one that only needs to know WHICH corpus primed the run
	// has the sha256 without paying five times the archive for it. A pointer with no hash would not
	// do — it would name a path whose content can change underneath the claim.
	if dg, err := corpusDigest(run.Dir()); err == nil && len(dg) > 0 {
		digestPath := filepath.Join(run.Dir(), "inputs", corpusDigestName)
		if b, mErr := json.MarshalIndent(dg, "", "  "); mErr == nil {
			if os.WriteFile(digestPath, b, 0o644) == nil {
				files = append(files, archived{Name: "inputs/" + corpusDigestName, Path: digestPath})
			}
		}
	}
	// AN EMPTY ARCHIVE IS A REFUSAL, not a small file. A tarball with no shards in it is exactly
	// what a run whose record never resolved would produce, and it would sit in run-archive/
	// looking like a kept run forever.
	if shards == 0 {
		return "", fmt.Errorf("no event shards under %s — refusing to write an archive that would "+
			"preserve nothing while looking like a preserved run", recs)
	}
	outDir := filepath.Join(project, "run-archive")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	out := filepath.Join(outDir, filepath.Base(filepath.Clean(run.Dir()))+".tar.gz")
	if err := archive.Write(out, files); err != nil {
		return "", err
	}
	return out, nil
}

// reapOrphanMirrors removes the checkpoint mirrors of runs that STOPPED, and reports only
// when it has something to say.
//
// AND NEVER THIS RUN'S. plans/record-tool.md has capture DELETE the mirror "after records/ is
// committed" — but capture does not commit. It writes an untracked tarball into the working
// tree and a human commits it later, so at the moment capture finishes, the record's only
// durable copy does not exist yet. Dropping the mirror here would remove the recovery path
// exactly when its replacement is most exposed to the add -A / checkout / stash classes the
// mirror was built to survive. So the reap is by AGE, and this run's mirror was rewritten at
// its last epoch: it is the freshest thing in the directory and cannot be the oldest.
//
// It belongs here because run-setup was the only caller, and nobody runs run-setup between
// research runs — a crashed run's mirror sat until someone happened to start a new one.
// Capture happens once per run, at its end.
//
// A BARE ZERO IS TWO DIFFERENT ANSWERS, so the one that means "not checked" is spoken and the
// one that means "nothing stale" is not: an unresolvable root reports NOT RUN, a clean board
// reports nothing, and those can no longer be read as the same line.
func reapOrphanMirrors(now time.Time) string {
	root, err := record.MirrorRoot()
	if err != nil {
		return "mirror purge: NOT RUN — " + jsSlice(err.Error(), 200)
	}
	if n := record.PurgeStaleMirrors(root, now, mirrorOrphanDays); n > 0 {
		return fmt.Sprintf("mirror purge: %d stale checkpoint mirror(s) removed", n)
	}
	return ""
}

// mirrorOrphanDays is longer than any run, which is the only property it needs: a live run
// refreshes its mirror every epoch, so what crosses this line stopped writing weeks ago.
const mirrorOrphanDays = 30

// famEvents is the stream, or nil for a record that could not be read — the same nil the
// harvests always treated as "no rulings this run".
func famEvents(fam *record.Family) []*record.Event {
	if fam == nil {
		return nil
	}
	return fam.Events
}

// corpusDigestName is the file the archive carries INSTEAD of the bulky staged corpora.
const corpusDigestName = "corpus-sha256.json"

// CorpusFile is one staged corpus input, named by its run-relative path and its content hash.
type CorpusFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int64  `json:"bytes"`
}

// corpusDigest hashes the staged corpora an archive does not carry, so a reader can say WHICH
// bytes primed a run without the archive paying for them.
//
// THE HASH IS THE CLAIM, AND THE PATH ALONE WOULD NOT BE. feov-memory/ is tracked in the
// repository, so the content behind these paths is recoverable — but it also CHANGES between
// runs, which is the whole reason a run's own copy was staged. A pointer would name a file whose
// content can move underneath it; the sha256 pins the version this run actually read.
//
// Absent files are skipped rather than reported as empty: a run set up before a corpus existed
// staged nothing, and an entry claiming a zero-byte corpus would be a different statement from
// one saying nothing was staged.
func corpusDigest(runDir string) ([]CorpusFile, error) {
	var out []CorpusFile
	// THE RECORDED PATH IS RECORD CONTENT, NOT A FILESYSTEM PATH, so it is written with forward
	// slashes on every platform. filepath.Join gives `inputs\\law\\precedents.md` on Windows, and a
	// digest whose keys change shape with the OS cannot be compared across the machines that read
	// an archive — the run that wrote it and the audit that re-reads it are rarely the same box.
	// Caught by Windows CI, which is the only place the two forms differ.
	add := func(rel string) {
		p := filepath.Join(runDir, filepath.FromSlash(rel))
		st, err := os.Stat(p)
		if err != nil || st.IsDir() {
			return
		}
		b, rErr := os.ReadFile(p)
		if rErr != nil {
			return
		}
		sum := sha256.Sum256(b)
		out = append(out, CorpusFile{Path: filepath.ToSlash(rel), SHA256: hex.EncodeToString(sum[:]), Bytes: st.Size()})
	}
	lawDir := filepath.Join(runDir, "inputs", "law")
	if entries, err := os.ReadDir(lawDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				add(path.Join("inputs", "law", e.Name()))
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// declarationsOnly keeps the rulings that STATE A RULE — the `bench declare` constructions — and
// drops the rest.
//
// A disposition is not a holding. The distinction is the whole reason the law corpus silted up:
// every ruling looked like a candidate because every ruling was filed as one, and a reviewer facing
// 76 entries could not see the six that stated a rule. Filtering at the harvest is what makes the
// queue's depth mean something again.
func declarationsOnly(rs []ruling) []ruling {
	out := rs[:0:0]
	for _, r := range rs {
		if r.kind == "declaration" {
			out = append(out, r)
		}
	}
	return out
}
