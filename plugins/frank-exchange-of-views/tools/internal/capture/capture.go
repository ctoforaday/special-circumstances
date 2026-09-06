// Package capture is the Go port of capture-research-run.mjs: the mechanical half of
// /research's run-record step plus the mechanized post-hoc auditor. It recomputes counts from
// the git-tracked FILES and diffs them against the envelopes' self-reports, writes cost.md and
// the transcript tarball, appends each chair's scorecard, harvests precedents into law/proposed,
// removes the run-live marker, and writes run-record-audit.md — exit 2 on any audit FAIL.
//
// The three record-backed audits (telemetry, log-parity, record-parity) read the record
// IN-PROCESS via record.BoardState → DebateJSONOf/FrictionJSONOf, never by spawning `merge show`.
// The PRECEDENT HARVEST reads it too, never the envelopes' self-reported ruling arrays: a bench
// that under-reports would promote less than it ruled, and one that reported nothing would
// promote nothing, indistinguishably from a run with nothing to promote.
package capture

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/runlive"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cost"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/modeltier"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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

// numOf returns a numeric value as float64; ok=false for a non-number (JS `typeof x==='number'`).
func numOf(v any) (float64, bool) {
	switch x := v.(type) {
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	case float64:
		return x, true
	}
	return 0, false
}

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
// UTF-8 bytes but 1 UTF-16 unit, so a byte cut lands short — and this feeds the friction
// 60-char prefix MATCH (present()), so it can change the missing SET, not just the display.
// Real capture prose carries em-dashes; the real-data differential caught it. Runes above the
// BMP (>0xFFFF) are a surrogate pair = 2 units.
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

// ReadJournal walks journal.jsonl tolerantly: every result object and the friction arrays inside.
func ReadJournal(transcriptDir string) (results []map[string]any, friction []EnvelopeLog) {
	results = []map[string]any{}
	friction = []EnvelopeLog{}
	b, err := os.ReadFile(filepath.Join(transcriptDir, "journal.jsonl"))
	if err != nil {
		return results, friction
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
		r, ok := j["result"].(map[string]any)
		if !ok {
			continue
		}
		results = append(results, r)
		if fr, ok := r["log"].([]any); ok {
			// THE AGENT HANDLE TRAVELS WITH THE TEXT. It is the only thing on this line that can
			// be joined to a seat, and dropping it here is what left the parity check comparing
			// prose to prose. See LogAudit.
			agent := jsString(j["agentId"])
			for _, f := range fr {
				friction = append(friction, EnvelopeLog{AgentID: agent, Text: jsString(f)})
			}
		}
	}
	return results, friction
}

// jsString mirrors JS String(x) for the values friction arrays carry (strings pass through).
func jsString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	case nil:
		return "null"
	case bool:
		if x {
			return "true"
		}
		return "false"
	default:
		b, _ := json.Marshal(x)
		return string(b)
	}
}

// ---- AUDIT 1: telemetry presence ----

// LivenessAudit records whether the run reached capture UNDER ITS OWN POWER or was terminated.
//
// THE TWO ARE OPPOSITE FACTS AND WERE THE SAME OUTPUT. "UNVERIFIED by judged deadlock" means the
// debate ran its course; "UNVERIFIED because the process was killed at round 1" means it never
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
	verdict := record.TerminalVerdict(run)
	l := record.Assess(run, now, verdict != "")
	if verdict != "" {
		return Audit{Check: "liveness", Verdict: "PASS", Detail: fmt.Sprintf(
			"the run reached a terminal outcome on the record (%s) — it finished under its own power", verdict)}
	}
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

// TelemetryAudit checks one telemetry round per red round. redRounds comes from the debate view
// (rounds with a red sitting), read in-process. Telemetry is now DERIVED from the record via the
// shared view library (never a materialized file), so it cannot be "absent": the check is whether
// the derived series covers every red round. SKIP when there were no red rounds.
func TelemetryAudit(run record.Run, redRounds int) Audit {
	if redRounds == 0 {
		return Audit{Check: "telemetry", Verdict: "SKIP", Detail: "the debate record shows 0 red round(s)"}
	}
	lines, err := view.Telemetry(run)
	if err != nil {
		return Audit{Check: "telemetry", Verdict: "FAIL", Detail: fmt.Sprintf("telemetry could not be computed from the record: %v", err)}
	}
	// A ROW WITHOUT A ROUND IS NOT COUNTED, and with the row typed that is now the only way to
	// miss: the field is optional in the schema, so presence is the question, and a key that was
	// never written can no longer masquerade as one that was absent.
	rounds := map[int32]bool{}
	for _, j := range lines {
		if j.Round != nil {
			rounds[j.GetRound()] = true
		}
	}
	v := "FAIL"
	if len(rounds) >= redRounds {
		v = "PASS"
	}
	return Audit{Check: "telemetry", Verdict: v, Detail: fmt.Sprintf("%d telemetry round(s) vs %d red round(s) on the record", len(rounds), redRounds)}
}

// AUDIT 2 IS GONE: "shard self-report vs files".
//
// It read red/ledger.md and red/archive.md and compared their line counts against the merge's
// self-reported ledger_closure_lines / archive_blocks. BOTH SIDES OF THAT COMPARISON ARE GONE —
// the envelope counts were removed 2026-07-19 for comparing numbers the merge made up (a haiku
// smoke self-reported archive_blocks: 22 in a round whose true archived count was 0), and the
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
// The duty it was reaching for is not lost, and is not restored here. It lives where the numbers
// cannot be authored: record.SpotCheckAudit computes the archive's size at round start by replay,
// and AttestationAudit reconciles each anchored closure against real tool calls. Both read the
// board. A self-report has no place on either side of that comparison.

// ---- AUDIT 3: friction parity ----

// LogAudit checks every envelope-self-reported friction reached the record. onRecord is the
// friction view's texts, read in-process. 60-char tolerant match in either direction.
// EnvelopeLog is one friction entry as a seat reported it in its RETURN ENVELOPE, carrying
// the harness agent that wrote it — the only handle on that line a record can be joined to.
type EnvelopeLog struct {
	AgentID, Text string
}

// LogAudit checks that every seat which reported friction in its envelope also opened the
// channel on the record. It joins on the SEAT, through the agent binding `register` writes.
//
// IT COMPARED PROSE TO PROSE, and reported 5 failures out of 5 on a run where every one of them
// was on the record. Measured 2026-08-22, research/2026-08-22_is-7-prime:
//
//	envelope   "blue-synthesize: citation-hygiene: candidate draft lane-1 provides authoritative
//	            source names (…) but lacks full URLs with coordinates required by research protocol"
//	record     "Citation-URL resolution: candidate draft lane-1 provides source names (…) but not
//	            full URLs with coordinates required by citation protocol"
//
// Same complaint, different wording — a seat that did its duty and then paraphrased itself into
// its return value. The other four were the empty case ("no capability gaps encountered"), also
// present, also reworded. A seat behaving correctly tripped this every time.
//
// The record carries seat_id as a FIELD. The envelope side glues the seat name onto the front of
// the prose ("blue-synthesize: …") when it remembers to, so the join had to be recovered by
// splitting a string — and could not be, so the check fell back to a 60-character substring
// comparison in either direction. record.SeatOfAgent already existed for exactly this, and its
// own doc comment is an argument against what this function was doing.
//
// WHEN THE JOIN IS ABSENT, SAY SO. A run whose PreToolUse hook never fired carries no agent_id on
// any register event — deliberately, so "not measured" stays legible rather than reading as an
// agent whose handle is the empty string. Those entries are counted out loud and are NOT findings:
// an unjoinable entry is a thing this audit cannot see, not a duty a seat skipped.
func LogAudit(run record.Run, envelope []EnvelopeLog, onRecord []record.LogEntryJSON) Audit {
	wroteToRecord := map[string]bool{}
	for _, fr := range onRecord {
		wroteToRecord[fr.SeatID] = true
	}
	var silent, unjoinable []string
	for _, e := range envelope {
		seat, found, err := record.SeatOfAgent(run, e.AgentID)
		if err != nil || !found || seat == "" {
			unjoinable = append(unjoinable, jsSlice(e.Text, 90))
			continue
		}
		if !wroteToRecord[seat] {
			silent = append(silent, seat+": "+jsSlice(e.Text, 90))
		}
	}
	note := ""
	if len(unjoinable) > 0 {
		note = fmt.Sprintf("\n    %d envelope entr%s COULD NOT BE JOINED to a seat and %s not judged "+
			"(no agent_id on the record — this run's PreToolUse hook did not fire, so the binding "+
			"`register` writes was never supplied):\n    - %s",
			len(unjoinable), plural(len(unjoinable), "y", "ies"), plural(len(unjoinable), "was", "were"),
			strings.Join(unjoinable, "\n    - "))
	}
	if len(silent) > 0 {
		return Audit{Check: "log-parity", Verdict: "FAIL",
			Detail: fmt.Sprintf("%d seat(s) reported friction in the envelope and opened no channel on the record "+
				"(it should have been recorded via the friction verb during the run):\n    - %s%s",
				len(silent), strings.Join(silent, "\n    - "), note)}
	}
	judged := len(envelope) - len(unjoinable)
	return Audit{Check: "log-parity", Verdict: "PASS",
		Detail: fmt.Sprintf("%d envelope entr%s joined to a seat, and every one of those seats is on the record "+
			"(%d friction entr%s recorded in total)%s",
			judged, plural(judged, "y", "ies"), len(onRecord), plural(len(onRecord), "y", "ies"), note)}
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// ---- AUDIT 4: context-use ----

var reHaiku = regexp.MustCompile(`haiku`)

type peak struct {
	agent  string
	peak   int
	window int
}

func ContextUse(transcriptDir string, agentFiles []string) Audit {
	var peaks []peak
	for _, f := range agentFiles {
		b, err := os.ReadFile(filepath.Join(transcriptDir, f))
		if err != nil {
			continue
		}
		pk, model := 0, ""
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
			m, _ := j["message"].(map[string]any)
			if m == nil {
				continue
			}
			u, _ := m["usage"].(map[string]any)
			if u == nil {
				continue
			}
			ctx := intField(u, "input_tokens") + intField(u, "cache_read_input_tokens") + intField(u, "cache_creation_input_tokens")
			if ctx > pk {
				pk = ctx
				if mm, ok := m["model"].(string); ok && mm != "" {
					model = mm
				}
			}
		}
		if pk > 0 {
			window := 1_000_000
			if reHaiku.MatchString(model) {
				window = 200_000
			}
			peaks = append(peaks, peak{agent: stripAgent(f), peak: pk, window: window})
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

func intField(m map[string]any, k string) int {
	if f, ok := numOf(m[k]); ok {
		return int(f)
	}
	return 0
}

var reAgentStrip = regexp.MustCompile(`^agent-|\.jsonl$`)

func stripAgent(f string) string { return reAgentStrip.ReplaceAllString(f, "") }

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
// while 6 proof anchors became `[^P1]`–`[^P6]` references with NO definitions and appendix
// headings written as `### [^P1] …` — which markdown reads as a second reference, not a
// heading. Both reports shipped and rendered literal `[^P1]` at every site: 12 broken
// references in one, 4 in the other.
//
// Nothing caught it. The blue-side detectors (`unbacked_citations`, `dropped_finding_markers`)
// read blue/report.md, where the anchors are all present and correct, so they reported 0 while
// the DELIVERABLE was broken. This audit reads the assembled artifact, which is the only
// surface where the weave's output can be judged.
//
// The rule is markdown's own and needs no heuristic: a definition is `[^id]:` at the start of a
// line; every other `[^id]` is a reference; a reference with no definition renders as its own
// source text. That precision matters — a reference may legitimately be followed by a colon in
// prose ("…artifacts exist[^P1]: `skills/…`"), which is a reference, not a definition.
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
		r, d, bad := footnotesIn(string(body))
		refs += r
		defs += d
		for _, id := range bad {
			dangling = append(dangling, name+":[^"+id+"]")
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

// footnotesIn counts one document's references and definitions and returns the ids referenced
// in it with no definition in it.
func footnotesIn(body string) (refs, defs int, dangling []string) {
	// CODE IS NOT PROSE, and a renderer does not look for footnotes inside it. A report that
	// quotes a proof's output in a fenced block, or writes `[^P…]` inline while DESCRIBING the
	// footnote surface, is not referencing anything — scanning those would report a dangling
	// footnote at exactly the sites that talk about footnotes.
	scanned := stripCode(body)
	defined := map[string]bool{}
	for _, m := range footnoteDef.FindAllStringSubmatch(scanned, -1) {
		defined[m[1]] = true
	}
	referenced := map[string]bool{}
	for _, loc := range footnoteRef.FindAllStringSubmatchIndex(scanned, -1) {
		id := scanned[loc[2]:loc[3]]
		// A definition matches the reference pattern too; tell them apart by position, exactly
		// as a markdown renderer does — line start, immediately followed by a colon.
		atLineStart := loc[0] == 0 || scanned[loc[0]-1] == '\n'
		followedByColon := loc[1] < len(scanned) && scanned[loc[1]] == ':'
		if atLineStart && followedByColon {
			continue
		}
		referenced[id] = true
	}
	for id := range referenced {
		if !defined[id] {
			dangling = append(dangling, id)
		}
	}
	return len(referenced), len(defined), dangling
}

var (
	footnoteDef = regexp.MustCompile(`(?m)^\[\^([^\]]+)\]:`)
	footnoteRef = regexp.MustCompile(`\[\^([^\]]+)\]`)
	fencedBlock = regexp.MustCompile("(?s)```.*?```")
	inlineCode  = regexp.MustCompile("`[^`\n]*`")
)

// stripCode blanks fenced and inline code, preserving newlines so line-start positions (which
// separate a definition from a reference) still mean what they meant.
func stripCode(s string) string {
	blank := func(m string) string {
		out := []rune(m)
		for i, r := range out {
			if r != '\n' {
				out[i] = ' '
			}
		}
		return string(out)
	}
	s = fencedBlock.ReplaceAllStringFunc(s, blank)
	return inlineCode.ReplaceAllStringFunc(s, blank)
}

func AssemblyScreen(run record.Run) Audit {
	board, err := record.BoardState(run)
	if err != nil {
		return Audit{Check: "assembly-screen", Verdict: "SKIP", Detail: "the record could not be read: " + err.Error()}
	}
	ev := record.EvidenceJSONOf(board)
	if ev.Counts.Sources == 0 {
		return Audit{Check: "assembly-screen", Verdict: "SKIP", Detail: "no citations on the record — nothing to screen"}
	}

	// THE WHOLE SET, not report.md. A refuted source cited in the debate transcript points the
	// reader at rejected evidence exactly as one in the analysis does, and a screen that reads
	// a seventh of the deliverable reports the other six as clean.
	var assembled strings.Builder
	for _, name := range reportdoc.Files() {
		body, rerr := os.ReadFile(filepath.Join(run.Dir(), name))
		if rerr != nil {
			continue
		}
		assembled.Write(body)
	}
	if assembled.Len() == 0 {
		// Before assembly there is nothing to screen, and saying so beats a PASS that means
		// "the file I was looking for was not there".
		return Audit{Check: "assembly-screen", Verdict: "SKIP",
			Detail: fmt.Sprintf("%d citation(s) on the record and no assembled document yet — the screen runs against the assembled artifact", ev.Counts.Sources)}
	}

	var carried []string
	for _, s := range ev.Sources {
		found := false
		for _, v := range s.Verified {
			if v.Refuted() {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		// The assembled report has had its anchors woven into visible footnotes, so the c- id is
		// screened against the PRE-assembly anchor AND the source's own url, which survives into
		// the composed bibliography. Either surviving means the reader is still being pointed at
		// a source red found against.
		if s.URL != "" && strings.Contains(assembled.String(), s.URL) {
			carried = append(carried, fmt.Sprintf("%s (%s)", s.Anchor, s.URL))
		}
	}

	if len(carried) > 0 {
		return Audit{Check: "assembly-screen", Verdict: "FAIL",
			Detail: fmt.Sprintf("%d source(s) red found AGAINST are still cited in the assembled report: %s. A refuted or absent source in the bibliography points a reader at evidence the audit rejected",
				len(carried), strings.Join(carried, ", "))}
	}
	return Audit{Check: "assembly-screen", Verdict: "PASS",
		Detail: fmt.Sprintf("%d citation(s), %d found against by red (refutes|absent), none still cited in the assembled report; %d cited source(s) nobody has checked",
			ev.Counts.Sources, ev.Counts.SourcesRefuted, ev.Counts.SourcesUnverified)}
}

// ---- AUDIT 6: record parity ----

// repoRootOf walks up from the run directory to the repository root — the directory holding
// `plugins/`. Derived rather than passed: capture is invoked with a run directory and a
// transcript directory, and threading a third path through every caller for one audit would be
// more surface than the audit is worth.
func repoRootOf(run record.Run) string {
	dir, err := filepath.Abs(run.Dir())
	if err != nil {
		return ""
	}
	for i := 0; i < 12; i++ {
		if st, e := os.Stat(filepath.Join(dir, "plugins")); e == nil && st.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
	return ""
}

// StrayRecordsAudit finds event shards written OUTSIDE any run directory.
//
// # What a stray is, and why it is not simply "another run"
//
// The repository holds many run directories, each with its own `records/`. What distinguishes a
// STRAY is that its parent is not a run: no `inputs/run-config.json`, no `blue/`. Nothing set it
// up; something wrote into it.
//
// Measured (#358): during research/2026-08-10_dual-read-vs-migration a seat whose shell cwd was
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

func StrayRecordsAudit(repoRoot, runDir string) Audit {
	if repoRoot == "" {
		return Audit{Check: "stray-records", Verdict: "SKIP", Detail: "no repository root to walk"}
	}
	captured, _ := filepath.Abs(runDir)
	var strays []string
	_ = filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
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
			rel, rerr := filepath.Rel(repoRoot, path)
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

// RecordParityAudit checks each red round has a blue sitting and a blue ROUND RECORD, floor
// redRounds-1. redRounds/blueBlocks come from the debate view, read in-process.
//
// THE ROUND RECORD IS COUNTED FROM `revision` EVENTS, not from headings in blue/CHANGELOG.md.
// The file is authored by hand while the event is emitted by the tool, so counting the file
// audited the seat's typing rather than the record — and the two disagree: the 2026-08-05 run
// carried a 6,847-byte CHANGELOG and exactly ONE revision event, from one of three eligible
// blue seats. The old regex pair (`^#+.*Round \d+` then `Round (\d+)`) read the plausible
// number and passed; the record shows the round records were never filed. See #268.
func RecordParityAudit(run record.Run, redRounds, blueBlocks int) Audit {
	if redRounds == 0 {
		return Audit{Check: "record-parity", Verdict: "SKIP", Detail: "no red rounds on record"}
	}
	clRounds := record.RoundsWithRevision(run)
	ok := blueBlocks >= redRounds-1 && clRounds >= redRounds-1
	v := "FAIL"
	if ok {
		v = "PASS"
	}
	return Audit{Check: "record-parity", Verdict: v,
		Detail: fmt.Sprintf("%d red round(s) vs %d blue sitting(s) and %d recorded round record(s) (floor: redRounds-1 — a PASS exit has no final blue response)",
			redRounds, blueBlocks, clRounds)}
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
// when emitted as a side effect of mint/prove/cite); and `[\s\S]*?` spanned invocations, so one
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
	board, err := record.BoardState(run)
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
	for _, e := range board.Events {
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

// Claim is one anchored closure (seat|tool|target) spot-checked against tool calls.
type Claim struct {
	ID, Seat, Tool, Target, Why string
}

// AttestationAudit spot-checks anchored closures against actual tool-call inputs.
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
func AttestationAudit(run record.Run, transcriptDir string, agentFiles []string, sampleFloor int) Audit {
	board, err := record.BoardState(run)
	if err != nil {
		return Audit{Check: "attestation-integrity", Verdict: "SKIP", Detail: "the record could not be read: " + err.Error()}
	}
	var claims []Claim
	for _, id := range board.GapOrder {
		g := board.Gaps[id]
		if g == nil || g.Closure == nil {
			continue
		}
		// A CARRIED closure attests nothing — it is last round's verification restated, and
		// holding a seat to a tool call it never made this round is a false finding.
		if g.Closure.GetCarriedFrom() != "" {
			continue
		}
		seat, tool, target := g.Closure.GetAnchorSeat(), g.Closure.GetAnchorTool(), g.Closure.GetAnchorTarget()
		if seat != "" && tool != "" && target != "" {
			claims = append(claims, Claim{ID: g.ID, Seat: seat, Tool: tool, Target: target})
		}
	}
	if len(claims) == 0 {
		return Audit{Check: "attestation-integrity", Verdict: "SKIP", Detail: "no anchored closures (all carried, or anchors absent)"}
	}
	var calls []string // JSON.stringify(input) of every tool call
	for _, f := range agentFiles {
		b, err := os.ReadFile(filepath.Join(transcriptDir, f))
		if err != nil {
			continue
		}
		for _, l := range strings.Split(string(b), "\n") {
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
			for _, ci := range content {
				c, ok := ci.(map[string]any)
				if !ok || strOf(c["type"]) != "tool_use" {
					continue
				}
				input := c["input"]
				if input == nil {
					input = map[string]any{}
				}
				calls = append(calls, jsStringify(input))
			}
		}
	}
	sampled := sampleClaims(claims, sampleFloor)
	var unreconciled, unmeasured []Claim
	for _, c := range sampled {
		needles, byID := needlesFor(c)
		if len(needles) == 0 {
			// NOT MEASURED IS NOT A FINDING, and folding it into one is how this audit came to
			// accuse honest seats. See needlesFor.
			c.Why = "no anchor id in the tool and no distinctive fragment in the target — nothing to reconcile AGAINST"
			unmeasured = append(unmeasured, c)
			continue
		}
		missing := ""
		for _, n := range needles {
			found := false
			for _, k := range calls {
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
				c.Why = "the closure cites " + missing + " but no tool call in any transcript carries it"
			} else {
				c.Why = "no tool call in any transcript touches " + missing
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
		detail := fmt.Sprintf("%d/%d sampled closure(s) NOT reconcilable against the trajectories — a claimed act with no matching tool call:\n%s\n    This rules on whether the RECORD IS HONEST, never on the merits of the gap.%s",
			len(unreconciled), len(sampled), strings.Join(ul, "\n"), note)
		return Audit{Check: "attestation-integrity", Verdict: "FAIL", Detail: detail, Unreconciled: unreconciled}
	}
	return Audit{Check: "attestation-integrity", Verdict: "PASS",
		Detail: fmt.Sprintf("%d/%d anchored closure(s) sampled and reconciled against actual tool calls%s",
			len(sampled)-len(unmeasured), len(claims), note)}
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

// anchorID matches the record's OWN identifiers — a citation, finding, proof or anchor id, each
// unique by construction rather than assembled from words.
//
// MEASURED 2026-08-22, and it is the whole reason needlesFor exists. Two closures anchored
// `report text at line 103` and `report text at line 62` — precise, human-meaningful, naming the
// exact line. mostDistinctive found no fragment longer than six characters in either (`report` is
// exactly six), returned "", and both were filed as claims with no matching tool call, under an
// audit whose own message says "This rules on whether the RECORD IS HONEST". The third sampled
// closure reconciled only because its target happened to contain the ten-letter word
// `projection`. Honesty was being decided by word length.
//
// The record was carrying the answer the whole time: anchor_tool reads
// `show report --anchor f-0dd40334`, and that id is exact. Prose is what a human reads; the id is
// what a machine joins on — see [[facts-are-fields]] clause 5, "never re-derive from the assembled
// form what the record could simply carry".
var anchorID = regexp.MustCompile(`\b[a-z]{1,2}-[0-9a-f]{8}\b`)

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
	if ids := anchorID.FindAllString(c.Tool, -1); len(ids) > 0 {
		return ids, true
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
// And it no longer has to be inferred from a price at all. `register` records served_model on the
// record — measured from the seat's own trajectory, where the harness DECLARES a swap by naming
// both ends — so a declared substitution is a fact this reads rather than a tier it deduces. The
// transcript pass stays for what the record cannot cover: runs recorded before the field existed,
// and seats that never registered.
func recordTierFindings(run record.Run, model, judgmentModel string) (findings []string, measured, total int) {
	board, err := record.BoardState(run)
	if err != nil || board == nil {
		return nil, 0, 0
	}
	for _, sm := range record.SeatModels(board) {
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

func ModelTierAudit(run record.Run, transcriptDir string, agentFiles []string) Audit {
	model, judgmentModel := cost.TierConfig(run)
	if model == "" && judgmentModel == "" {
		return Audit{Check: "model-tier", Verdict: "SKIP", Detail: "no run-config models (pre-#111 run)"}
	}
	var rows []cost.Row
	for _, f := range agentFiles {
		b, err := os.ReadFile(filepath.Join(transcriptDir, f))
		if err != nil {
			continue
		}
		rows = append(rows, cost.ScanTranscript(string(b)))
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
func closeRunLiveMarker(cwd, runDir string) string {
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
	// the next run was eleven minutes into its first round. Caught by reading this function
	// rather than by anything in it.
	//
	// A path comparison guarded that. Per-run rows REMOVE it: the row this capture owns is the
	// only one it can take, so there is no longer a case where taking the wrong one is possible
	// and a check is what stands between (#529).
	if _, err := os.Stat(marker); err != nil {
		if os.IsNotExist(err) {
			return "run-live marker: none at " + marker +
				" — either it was already cleared, or capture is running from a directory that is not the project root, in which case the real marker is still there"
		}
		return "run-live marker: could not be read at " + marker + ": " + err.Error()
	}
	before := len(runlive.ReadRunLive(cwd))
	found, remaining := runlive.RemoveRunLiveEntry(cwd, runDir)
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
	// EnvelopeClaimed is how many rulings the SEATS said they made. It is a cross-check, not
	// a source: a field rather than a sentence, so a reader can compare it to Count without
	// parsing prose (facts-are-fields).
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
// DELIBERATELY NOT HARVESTED: `motion-rule` on a grade or direction. Those are rulings, and a
// grade holding ("disclosure does not lower likelihood") reads like a rule — but the harvest
// template's `question:` would have to be the motion's ask, which lives on the separate `motion`
// event, and promoting a ruling without the ask it answered is how a holding loses its scope.
// That is a change to WHAT law/proposed contains, not to where the harvest reads; it wants its
// own decision. Named here so the smaller scope is visible rather than silent.
func rulingsFromRecord(board *record.Board) []ruling {
	if board == nil {
		return nil
	}
	// A petition ruling names the motion, not the filer. The filer is the seat that appended
	// the `motion` event, which is the only place that fact exists.
	filedBy := map[string]string{}
	for _, e := range board.Events {
		if m, ok := recordpb.BodyAs[*recordpb.Motion](e); ok {
			filedBy[m.GetMotionId()] = e.GetSeatId()
		}
	}

	var out []ruling
	for _, e := range board.Events {
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_OPINION:
			o, ok := recordpb.BodyAs[*recordpb.Opinion](e)
			if !ok {
				continue
			}
			out = append(out, ruling{
				kind:        "docket",
				gapID:       o.GetGapId(),
				disposition: recordpb.Word(o.GetDisposition()),
				rationale:   o.GetRationale(),
			})
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
			// Only the PETITION subject reaches the precedent harvest: a grade or direction
			// ruling answers a motion the motions view renders with its ask beside it.
			//
			// THE SUBJECT IS NOW THE ONEOF, not a string compare. A petition ruling carries its
			// verdict on the `petition` arm, so reading the arm IS the subject test — a grade
			// ruling cannot answer it, where the old `Str("subject") != "petition"` depended on a
			// field that could disagree with the value beside it.
			r, ok := recordpb.BodyAs[*recordpb.MotionRule](e)
			if !ok || r.GetSubject() != recordpb.MotionSubject_MOTION_SUBJECT_PETITION {
				continue
			}
			out = append(out, ruling{
				kind:        "petition",
				petitioner:  filedBy[r.GetMotionId()],
				disposition: recordpb.Word(r.GetPetition()),
				rationale:   r.GetOpinion(),
			})
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
		for _, key := range []string{"resolutions", "rulings"} {
			if rs, ok := r[key].([]any); ok {
				n += len(rs)
			}
		}
	}
	return n
}

func HarvestPrecedents(run record.Run, results []map[string]any, lawDir string, board *record.Board) HarvestResult {
	slug := slugOf(run)
	rulings := rulingsFromRecord(board)
	claimed := rulingsClaimedByEnvelopes(results)
	if len(rulings) == 0 {
		// STATED, not implied. "0 rulings" is otherwise the output of both an honest quiet run
		// and a harvest that cannot see the rulings in front of it. If the envelopes claim
		// rulings the record does not hold, that is the second case and it says so.
		if claimed > 0 {
			return HarvestResult{Written: false, Count: 0, EnvelopeClaimed: claimed,
				Reason: fmt.Sprintf("the record holds NO rulings while the envelopes claim %d — the record is the source, so nothing is promoted; this divergence is the finding", claimed)}
		}
		return HarvestResult{Written: false, Count: 0}
	}
	if _, err := os.Stat(lawDir); err != nil {
		return HarvestResult{Written: false, Count: len(rulings), EnvelopeClaimed: claimed, Reason: "no law/ dir at repo root"}
	}
	_ = os.MkdirAll(filepath.Join(lawDir, "proposed"), 0o755)
	out := filepath.Join(lawDir, "proposed", slug+".md")
	var body []string
	body = append(body, "# proposed holdings — "+slug+" [ALL PERSUASIVE — awaiting human review per law/README.md]",
		"",
		"Each entry below is a RULING the bench recorded, not yet a holding. `facts`, `holding` and",
		"`scope-limits` are the reviewer's to write from the cited record: the harvest states what it",
		"observed and never invents the rule. A ruling promoted with its placeholders intact is not",
		"citable — law/README.md: a holding without its factual predicate is not citable.", "")
	for i, r := range rulings {
		q := "disposition of " + r.gapID
		src := slug + ", " + r.gapID
		if r.kind == "petition" {
			q = "petition by " + r.petitioner
			src = slug + ", petition:" + r.petitioner
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
		return HarvestResult{Written: false, Count: len(rulings), EnvelopeClaimed: claimed,
			Reason: "could not write " + out + ": " + err.Error()}
	}
	return HarvestResult{Written: true, Count: len(rulings), EnvelopeClaimed: claimed, Path: out}
}

// ---- scorecards ----

// ScorecardResult mirrors the writeScorecards return.
type ScorecardResult struct {
	Written bool
	Rows    int
	Chairs  int
	Reason  string
}

func WriteScorecards(run record.Run, results []map[string]any, memoryDir string, board *record.Board) ScorecardResult {
	if _, err := os.Stat(memoryDir); err != nil {
		return ScorecardResult{Written: false, Reason: fmt.Sprintf("no %s — scorecards need the tracked memory dir", memoryDir)}
	}
	cards := scorecard.Compute(run, results, board)
	label := labelOf(run)
	rows := 0
	chairs := 0
	var failed []string

	// Sorted, so a partial failure names the same chairs in the same order every run: a
	// reason that reshuffles between runs is one a reader cannot diff.
	names := make([]string, 0, len(cards))
	for chair := range cards {
		names = append(names, chair)
	}
	sort.Strings(names)

	for _, chair := range names {
		chairRows := cards[chair]
		p := filepath.Join(memoryDir, chair+"-scorecard.md")
		head := scorecard.ChairHeader(chair)
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
			failed = append(failed, chair+": cannot read "+p+" ("+err.Error()+") — left untouched rather than overwritten with a fresh header")
			continue
		}
		body := trimTrailingNewlines(head) + "\n" + scorecard.RenderChair(chair, chairRows, label)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			// The write's error was discarded while the result reported Written: true —
			// the same shape the law-harvest write above refuses by name.
			failed = append(failed, chair+": cannot write "+p+" ("+err.Error()+")")
			continue
		}
		rows += len(chairRows)
		chairs++
	}

	res := ScorecardResult{Written: len(failed) == 0, Rows: rows, Chairs: chairs}
	if len(failed) > 0 {
		res.Reason = fmt.Sprintf("%d of %d chair(s) NOT written: %s", len(failed), len(cards), strings.Join(failed, "; "))
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

// ---- tarball ----

func writeTarball(transcriptDir, outPath string, agentFiles []string) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	sorted := append([]string(nil), agentFiles...)
	sort.Strings(sorted)
	for _, name := range sorted {
		b, err := os.ReadFile(filepath.Join(transcriptDir, name))
		if err != nil {
			return err
		}
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(b))}); err != nil {
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

// appendCostToReport folds the per-seat-round cost table from cost.md into the run document as
// a ## Cost section. No-op (returns "") when the target is absent (a run that never reached
// assembly), already carries a ## Cost section (idempotent across re-runs), or cost.md has no
// table. cost.md itself is left byte-identical.
//
// THE HEADING THE TABLE BRINGS IS DEMOTED, NOT DUPLICATED. The first cut wrote "## Cost" and
// then pasted a slice that opens with its own "## Per seat-round" — so every archived report
// carries a Cost section of exactly nine bytes: a heading, and nothing under it. An empty
// section reads as a section that found nothing, which is a different claim from one whose
// content is sitting immediately below it under another name.
func appendCostToReport(reportPath, costPath string) string {
	report, err := os.ReadFile(reportPath)
	if err != nil {
		return ""
	}
	if strings.Contains(string(report), "\n## Cost\n") {
		return ""
	}
	costMd, err := os.ReadFile(costPath)
	if err != nil {
		return ""
	}
	table := perSeatRoundTable(string(costMd))
	if table == "" {
		return ""
	}
	table = strings.Replace(table, "## Per seat-round", "### Per seat-round", 1)
	body := strings.TrimRight(string(report), "\n") + "\n\n## Cost\n\n" + table + "\n"
	if err := os.WriteFile(reportPath, []byte(body), 0o644); err != nil {
		return filepath.Base(reportPath) + ": cost append FAILED — " + jsSlice(err.Error(), 200)
	}
	return filepath.Base(reportPath) + ": cost breakdown folded in (## Cost)"
}

// perSeatRoundTable slices the "## Per seat-round" table out of a rendered cost.md — from that
// heading up to the "## Notes" section — or "" if the markers are absent.
func perSeatRoundTable(costMd string) string {
	start := strings.Index(costMd, "## Per seat-round")
	if start < 0 {
		return ""
	}
	rest := costMd[start:]
	if end := strings.Index(rest, "\n## Notes"); end >= 0 {
		rest = rest[:end]
	}
	return strings.TrimRight(rest, "\n")
}

// Run executes capture: mechanics + the nine audits, writes run-record-audit.md, and returns the
// audits, the report string, and whether any audit FAILed (exit 2). cwd-rooted side effects
// (feov-memory, law, .claude/run-live.json) resolve from os.Getwd(), exactly as the JS used
// process.cwd().
// `now` is injected so the one non-deterministic input — how long ago the record stopped moving —
// is controllable in a test, the same way WriteRunLiveMarker takes its clock.
func Run(run record.Run, transcriptDir string, now time.Time) (audits []Audit, report string, exitFail bool, err error) {
	var lines []string

	// Mechanics: journal copy, transcript tarball, cost.md.
	if err = copyFile(filepath.Join(transcriptDir, "journal.jsonl"), filepath.Join(run.Dir(), "trajectories", "journal.jsonl")); err != nil {
		return nil, "", false, err
	}
	agentFiles, err := listAgentFiles(transcriptDir)
	if err != nil {
		return nil, "", false, err
	}
	tarPath := filepath.Join(run.Dir(), "trajectories", "agent-transcripts.tar.gz")
	if terr := writeTarball(transcriptDir, tarPath, agentFiles); terr != nil {
		lines = append(lines, "tarball: FAILED — "+jsSlice(terr.Error(), 200))
	} else {
		lines = append(lines, fmt.Sprintf("tarball: %d transcript(s)", len(agentFiles)))
	}
	// THE RECORD ITSELF, to the one place that outlives the container. Reported either way: an
	// archive that silently did not happen is a run whose evidence is gone the next time the
	// container is reclaimed, and nothing would have said so.
	if arc, aerr := ArchiveRecord(run, repoRootOf(run)); aerr != nil {
		lines = append(lines, "record archive: FAILED — "+jsSlice(aerr.Error(), 200))
	} else {
		rel := arc
		if r, rerr := filepath.Rel(repoRootOf(run), arc); rerr == nil {
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

	costF, cerr := os.Create(filepath.Join(run.Dir(), "cost.md"))
	if cerr != nil {
		return nil, "", false, cerr
	}
	if rerr := cost.Report(transcriptDir, run, costF); rerr != nil {
		costF.Close()
		lines = append(lines, "cost.md: FAILED — "+jsSlice(rerr.Error(), 200))
	} else {
		costF.Close()
		lines = append(lines, "cost.md: written (telemetry join included)")
		// Fold the per-seat-round cost table into run.md as a ## Cost section — what the run
		// cost is a fact about the RUN, and run.md is the document whose subject that is. The
		// set is assembled mid-run WITHOUT transcript access (the transcript dir reaches only
		// capture), so this is the one stage that can. Slices the already-rendered cost.md
		// (kept byte-identical) rather than re-generate.
		if msg := appendCostToReport(filepath.Join(run.Dir(), reportdoc.FileRun), filepath.Join(run.Dir(), "cost.md")); msg != "" {
			lines = append(lines, msg)
		}
	}

	results, friction := ReadJournal(filepath.Join(run.Dir(), "trajectories"))

	// Record-backed reads, in-process (the JS spawned `merge show` views).
	board, _ := record.BoardState(run)
	redRounds, blueBlocks := 0, 0
	onRecord := []record.LogEntryJSON{}
	if board != nil {
		dj := record.DebateJSONOfEvents(board.Events)
		for _, r := range dj.Rounds {
			if len(r.Red) > 0 {
				redRounds++
			}
			if len(r.Blue) > 0 {
				blueBlocks++
			}
		}
		// ONE ARM NOW, AND IT STILL COUNTS THE ATTESTED EMPTY CASE. The clean sitting used to be a
		// second event type appended separately; it is a `nominal` ENTRY on the same list, so a
		// seat that files one has used the channel exactly as the duty asks and needs no special
		// collection to be seen doing it.
		fj := record.LogJSONOf(board.Events)
		onRecord = append(onRecord, fj.Log...)
	}

	audits = []Audit{
		LivenessAudit(run, now),
		TelemetryAudit(run, redRounds),
		LogAudit(run, friction, onRecord),
		ContextUse(transcriptDir, agentFiles),
		AssemblyScreen(run),
		FootnoteIntegrity(run),
		StrayRecordsAudit(repoRootOf(run), run.Dir()),
		RecordParityAudit(run, redRounds, blueBlocks),
		BackfillAudit(run),
		AttestationAudit(run, transcriptDir, agentFiles, 5),
		ModelTierAudit(run, transcriptDir, agentFiles),
		// THE LAST AUDIT, because it is the only one that EXECUTES anything: a bounded sample of
		// the run's own proofs, re-run and compared. See proofrerun.go for why a seat's spot-check
		// could not be the thing that does this.
		ProofRerunAudit(run, proofRerunSample),
		// Round 0's declared breadth against the lane seats that actually took theirs — the one
		// run-config field nothing reconciled. See lanecoverage.go.
		LaneCoverageAudit(run),
	}

	cwd, _ := os.Getwd()
	sc := WriteScorecards(run, results, filepath.Join(cwd, "feov-memory"), board)
	// BOTH LINES WHEN BOTH ARE TRUE. The old form printed the counts OR the reason, so a
	// partial failure — some chairs written, one unwritable — showed the cheerful line and
	// swallowed the reason entirely.
	if sc.Chairs > 0 || sc.Reason == "" {
		lines = append(lines, fmt.Sprintf("scorecards: %d row(s) across %d chair(s) -> feov-memory/", sc.Rows, sc.Chairs))
	}
	if sc.Reason != "" {
		lines = append(lines, "scorecards: "+sc.Reason)
	}

	prec := HarvestPrecedents(run, results, filepath.Join(cwd, "law"), board)
	// The divergence has to REACH the report. Computing it and then printing "no rulings this
	// run" would rebuild the same defect one level up: the miss folded back into the zero.
	divergence := ""
	if prec.EnvelopeClaimed != prec.Count {
		divergence = fmt.Sprintf(" [the envelopes claim %d — the record is the source]", prec.EnvelopeClaimed)
	}
	switch {
	case prec.Written:
		lines = append(lines, fmt.Sprintf("precedent harvest: %d ruling(s) -> %s (PERSUASIVE, awaiting review)%s", prec.Count, prec.Path, divergence))
	case prec.Count > 0:
		lines = append(lines, fmt.Sprintf("precedent harvest: %d ruling(s), %s%s", prec.Count, prec.Reason, divergence))
	case prec.Reason != "":
		lines = append(lines, "precedent harvest: "+prec.Reason)
	default:
		lines = append(lines, "precedent harvest: no rulings this run")
	}

	// THE CLASS HARVEST, beside the precedent one and for the same reason: a run's coinage that
	// reaches nothing outside the run directory is a run's coinage lost (#515).
	cls := HarvestClasses(run, filepath.Join(cwd, "law"), board)
	switch {
	case cls.Written:
		lines = append(lines, fmt.Sprintf("class harvest: %d class(es) -> %s (PROPOSED, not staged into any run until adopted)", cls.Count, cls.Path))
	case cls.Count > 0:
		lines = append(lines, fmt.Sprintf("class harvest: %d class(es), %s", cls.Count, cls.Reason))
	case cls.Reason != "":
		lines = append(lines, "class harvest: "+cls.Reason)
	default:
		lines = append(lines, "class harvest: no classes coined this run")
	}

	lines = append(lines, closeRunLiveMarker(cwd, run.Dir()))

	var out []string
	out = append(out,
		"# capture-audit — mechanized post-hoc checks (capture-research-run.mjs)",
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
	if werr := os.WriteFile(filepath.Join(run.Dir(), "run-record-audit.md"), []byte(report), 0o644); werr != nil {
		return nil, "", false, werr
	}
	for _, a := range audits {
		if a.Verdict == "FAIL" {
			exitFail = true
		}
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

func listAgentFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		n := e.Name()
		if strings.HasPrefix(n, "agent-") && strings.HasSuffix(n, ".jsonl") {
			files = append(files, n)
		}
	}
	sort.Strings(files)
	return files, nil
}

// ---- record archive ----

// ArchiveRecord writes the run's RAW RECORD to run-archive/<slug>.tar.gz at the repository root,
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
// WHAT IS IN IT, AND WHAT IS NOT. records/ and proofs/: 49 KB gzipped for a fifteen-seat run, so
// keeping every run costs about a megabyte per twenty. NOT cache/ — 7.3 MB of fetched source
// bytes for the same run, re-fetchable, and content-addressed by a sha256 the citation events
// already carry, so its integrity stays checkable without it. NOT the agent transcripts: they are
// an order of magnitude larger than the record and answer a different question (what a seat DID
// rather than what it RECORDED); keeping them is a separate decision and is named here rather
// than made quietly.
//
// It does not reintroduce ffc4bf4's hazard: seats read their own run directory and the inputs
// mirrored into it, never run-archive/, and a gzipped tar is not prose a glob can wander into.
func ArchiveRecord(run record.Run, repoRoot string) (string, error) {
	recs := run.Records()
	var files []struct {
		name string
		path string
	}
	add := func(root, prefix string) error {
		return filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil // a missing proofs/ is ordinary; a missing records/ is caught below
			}
			rel, rerr := filepath.Rel(root, p)
			if rerr != nil {
				return rerr
			}
			files = append(files, struct{ name, path string }{prefix + filepath.ToSlash(rel), p})
			return nil
		})
	}
	if err := add(recs, "records/"); err != nil {
		return "", err
	}
	shards := len(files)
	_ = add(filepath.Join(run.Dir(), "proofs"), "proofs/")
	// AN EMPTY ARCHIVE IS A REFUSAL, not a small file. A tarball with no shards in it is exactly
	// what a run whose record never resolved would produce, and it would sit in run-archive/
	// looking like a kept run forever.
	if shards == 0 {
		return "", fmt.Errorf("no event shards under %s — refusing to write an archive that would "+
			"preserve nothing while looking like a preserved run", recs)
	}
	outDir := filepath.Join(repoRoot, "run-archive")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	out := filepath.Join(outDir, filepath.Base(filepath.Clean(run.Dir()))+".tar.gz")
	f, err := os.Create(out)
	if err != nil {
		return "", err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })
	for _, fl := range files {
		b, rerr := os.ReadFile(fl.path)
		if rerr != nil {
			return "", rerr
		}
		if err := tw.WriteHeader(&tar.Header{Name: fl.name, Mode: 0o644, Size: int64(len(b))}); err != nil {
			return "", err
		}
		if _, err := tw.Write(b); err != nil {
			return "", err
		}
	}
	if err := tw.Close(); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
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
// its last round: it is the freshest thing in the directory and cannot be the oldest.
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
// refreshes its mirror every round, so what crosses this line stopped writing weeks ago.
const mirrorOrphanDays = 30

// ---- gap-class harvest ----

// coinedClass is one class a seat minted this run, and where it was first used.
type coinedClass struct {
	slug, definition, neighbor, distinguisher string
	seatID, firstGap                          string
}

// classesFromRecord reads the classes THE RECORD holds, for the same reason rulingsFromRecord
// does: `class new` refuses a coinage missing its definition, neighbour or distinguisher
// (validateClassNew), so what arrives here is well-formed by construction rather than by hope.
//
// FirstGap is the join a reviewer needs and the record already holds: a class exists to
// discriminate, and the gap it was first minted against is the concrete case that motivated it.
// Empty means coined and never used, which is itself worth seeing on the proposal.
func classesFromRecord(board *record.Board) []coinedClass {
	if board == nil {
		return nil
	}
	var out []coinedClass
	for _, e := range board.Events {
		cn := e.GetClassNew()
		if cn == nil {
			continue
		}
		c := coinedClass{
			slug: cn.GetSlug(), definition: cn.GetDefinition(),
			neighbor: cn.GetNeighbor(), distinguisher: cn.GetDistinguisher(),
			seatID: e.GetSeatId(),
		}
		for _, m := range board.Events {
			if mint := m.GetMint(); mint != nil && mint.GetClass() == c.slug {
				c.firstGap = mint.GetGapId()
				break
			}
		}
		out = append(out, c)
	}
	return out
}

// HarvestClasses promotes every class a run coined into law/proposed, and NEVER into the registry.
//
// WHY PROPOSED AND NOT ADOPTED (#515). feov-memory/class-registry.json is what every `--class` is
// validated against, and the whole value of that gate is that a mint is REFUSED rather than waved
// through. A class nobody reviewed, validating a future mint, is the registry losing the only
// thing it means. So this mirrors the precedent flow exactly: the harvest states what the run
// coined, and adoption is a human act.
//
// WHY IT EXISTS AT ALL. Before this, a coined class lived in exactly two places — the run's event
// record and the run's staged copy of the registry — both inside a gitignored run directory. It
// reached run-archive/*.tar.gz and nothing read that back. Measured 2026-08-22: red coined
// `silent-no-match-probe` mid-run, with a sharper distinguisher than anything already registered,
// and the next run would have staged the same 38 classes and made red discover it again.
//
// ONE FILE PER CLASS PER RUN, keyed on both. Two runs coining the same slug with different
// definitions then land as two files a reviewer sees side by side, rather than one silently
// overwriting the other — the collision has an arbiter, and the arbiter is a person.
func HarvestClasses(run record.Run, lawDir string, board *record.Board) HarvestResult {
	slug := slugOf(run)
	classes := classesFromRecord(board)
	if len(classes) == 0 {
		return HarvestResult{Written: false, Count: 0}
	}
	if _, err := os.Stat(lawDir); err != nil {
		return HarvestResult{Written: false, Count: len(classes), Reason: "no law/ dir at repo root"}
	}
	if err := os.MkdirAll(filepath.Join(lawDir, "proposed"), 0o755); err != nil {
		return HarvestResult{Written: false, Count: len(classes), Reason: err.Error()}
	}
	var written []string
	for _, c := range classes {
		out := filepath.Join(lawDir, "proposed", "class-"+c.slug+"--"+slug+".md")
		used := c.firstGap
		if used == "" {
			used = "(coined and not minted against this run — the class has no worked case yet)"
		}
		body := strings.Join([]string{
			"# proposed gap class — `" + c.slug + "` [PROPOSED — not in the registry until adopted]",
			"",
			"Coined by `" + c.seatID + "` during `" + slug + "`. It is NOT staged into any later run:",
			"an unreviewed class validating a future `--class` is the registry losing the only thing it",
			"means. Adopting it means adding the slug to `feov-memory/class-registry.json` by hand.",
			"",
			"- **definition**: " + c.definition,
			"- **neighbour**: `" + c.neighbor + "`",
			"- **distinguisher**: " + c.distinguisher,
			"- **first used on**: " + used,
			"",
			"The three fields above are the seat's own words, refused at the write if any were missing",
			"(`record.validateClassNew`), so this proposal is well-formed by construction. What a",
			"reviewer decides is whether it DISCRIMINATES — whether the distinguisher actually separates",
			"it from its neighbour on a case where both are arguable.",
			"",
		}, "\n")
		if err := os.WriteFile(out, []byte(body), 0o644); err != nil {
			return HarvestResult{Written: len(written) > 0, Count: len(classes), Reason: err.Error()}
		}
		written = append(written, out)
	}
	return HarvestResult{Written: true, Count: len(classes), Path: filepath.Join(lawDir, "proposed")}
}
