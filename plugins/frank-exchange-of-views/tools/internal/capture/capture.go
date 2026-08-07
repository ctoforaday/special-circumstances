// Package capture is the Go port of capture-research-run.mjs: the mechanical half of
// /research's run-record step plus the mechanized post-hoc auditor. It recomputes counts from
// the git-tracked FILES and diffs them against the envelopes' self-reports, writes cost.md and
// the transcript tarball, appends each chair's scorecard, harvests precedents into law/proposed,
// removes the run-live marker, and writes run-record-audit.md — exit 2 on any audit FAIL.
//
// The three record-backed audits (telemetry, friction-parity, record-parity) read the record
// IN-PROCESS via record.BoardState → DebateJSONOf/FrictionJSONOf, never by spawning `merge show`.
// The JS no-`--bin` SKIP path has no Go analogue: the binary always holds the record. Real runs
// always passed --bin, so this is behavior-preserving — the differential in the port PR pins it.
package capture

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cost"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
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

// numOrUndefined renders a value as JS string-interpolation would: a number as its integer text,
// an absent/non-number field as the literal "undefined" (the `${obj.missing}` coercion).
func numOrUndefined(m map[string]any, k string) string {
	if f, ok := numOf(m[k]); ok {
		return strconv.FormatInt(int64(f), 10)
	}
	return "undefined"
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
func ReadJournal(transcriptDir string) (results []map[string]any, friction []string) {
	results = []map[string]any{}
	friction = []string{}
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
		if fr, ok := r["friction"].([]any); ok {
			for _, f := range fr {
				friction = append(friction, jsString(f))
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

// TelemetryAudit checks one telemetry round per red round. redRounds comes from the debate view
// (rounds with a red sitting), read in-process. Telemetry is now DERIVED from the record via the
// shared view library (never a materialized file), so it cannot be "absent": the check is whether
// the derived series covers every red round. SKIP when there were no red rounds.
func TelemetryAudit(runDir string, redRounds int) Audit {
	if redRounds == 0 {
		return Audit{Check: "telemetry", Verdict: "SKIP", Detail: "the debate record shows 0 red round(s)"}
	}
	lines, err := view.Telemetry(runDir)
	if err != nil {
		return Audit{Check: "telemetry", Verdict: "FAIL", Detail: fmt.Sprintf("telemetry could not be computed from the record: %v", err)}
	}
	rounds := map[string]bool{}
	for _, j := range lines {
		if r, ok := j["round"]; ok && r != nil {
			rounds[jsString(r)] = true
		}
	}
	v := "FAIL"
	if len(rounds) >= redRounds {
		v = "PASS"
	}
	return Audit{Check: "telemetry", Verdict: v, Detail: fmt.Sprintf("%d telemetry round(s) vs %d red round(s) on the record", len(rounds), redRounds)}
}

// ---- AUDIT 2: shard self-report vs files ----

// ShardAudit compares the merge's SELF-REPORTED counts against the board.
//
// It read red/ledger.md and red/archive.md — both files `setup` stopped creating when they
// became projections — and returned SKIP "no ledger/archive (pre-sharding run)" on every
// record-mode run. That is the FOURTH audit this slice found reading a file that no longer
// exists, and I missed it in the first pass of this same sweep: the sibling-sweep note on the
// previous commit claimed no markdown reads remained under capture/, and this was still here.
//
// The self-report it grades is real and worth keeping — a haiku merge once reported
// archive_blocks:22 in a round where the true archived count was 0 — so the counts now come
// from the board, which is the thing the self-report is supposed to agree with.
func ShardAudit(runDir string, results []map[string]any) Audit {
	board, err := record.BoardState(runDir)
	if err != nil {
		return Audit{Check: "shards", Verdict: "SKIP", Detail: "the record could not be read: " + err.Error()}
	}
	closed := 0
	for _, id := range board.GapOrder {
		g := board.Gaps[id]
		if g == nil {
			continue
		}
		// The closure INDEX carried one line per closed gap; the ARCHIVE one block per closure.
		// Both are the closed set, which is what the merge was reporting on.
		if !g.Open {
			closed++
		}
	}
	if len(board.GapOrder) == 0 {
		return Audit{Check: "shards", Verdict: "SKIP", Detail: "no gaps on the record"}
	}
	var lastRed map[string]any
	for i := len(results) - 1; i >= 0; i-- {
		if _, ok := numOf(results[i]["ledger_closure_lines"]); ok {
			lastRed = results[i]
			break
		}
		if _, ok := numOf(results[i]["archive_blocks"]); ok {
			lastRed = results[i]
			break
		}
	}
	self := "no envelope self-report found in journal"
	consistent := lastRed == nil
	if lastRed != nil {
		self = fmt.Sprintf("self-reported %s/%s", numOrUndefined(lastRed, "ledger_closure_lines"), numOrUndefined(lastRed, "archive_blocks"))
		lcl, okL := numOf(lastRed["ledger_closure_lines"])
		ab, okA := numOf(lastRed["archive_blocks"])
		consistent = okL && okA && int(lcl) == closed && int(ab) == closed
	}
	v := "FAIL"
	if consistent {
		v = "PASS"
	}
	return Audit{Check: "shards", Verdict: v, Detail: fmt.Sprintf("the record holds %d closed gap(s); %s", closed, self)}
}

// ---- AUDIT 3: friction parity ----

// FrictionAudit checks every envelope-self-reported friction reached the record. onRecord is the
// friction view's texts, read in-process. 60-char tolerant match in either direction.
func FrictionAudit(envelopeFriction, onRecord []string) Audit {
	present := func(f string) bool {
		for _, t := range onRecord {
			if strings.Contains(t, jsSlice(f, 60)) || strings.Contains(f, jsSlice(t, 60)) {
				return true
			}
		}
		return false
	}
	var missing []string
	for _, f := range envelopeFriction {
		if !present(f) {
			missing = append(missing, f)
		}
	}
	if len(missing) > 0 {
		lines := make([]string, len(missing))
		for i, m := range missing {
			lines[i] = "    - " + jsSlice(m, 120)
		}
		return Audit{Check: "friction-parity", Verdict: "FAIL",
			Detail: fmt.Sprintf("%d envelope friction entr%s never reached the record (should have been recorded via the friction verb during the run):\n%s",
				len(missing), plural(len(missing), "y", "ies"), strings.Join(lines, "\n"))}
	}
	return Audit{Check: "friction-parity", Verdict: "PASS",
		Detail: fmt.Sprintf("%d envelope friction entr%s all present on the record", len(envelopeFriction), plural(len(envelopeFriction), "y", "ies"))}
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

// AssemblyScreen SHOULD catch a report that still carries a citation red refuted. It cannot
// today, and it now SAYS SO instead of reporting PASS.
//
// THE SIGNAL WAS DESTROYED, not merely relocated. Until the citation ledger moved onto the
// record, red typed its verdict into the confidence column as prose — "LOW — REFUTED: closed
// as duplicate", "LOW (absent from abs and html)" — and this screen regex-scanned that column
// for REFUTED|ABSENT. Real ledgers carried 5-11 such rows (2026-07-12 through 2026-07-20).
// Then `lens cite --confidence` became a CLOSED ENUM of high|medium|low, so there is no longer
// any field in which red can record that verification FAILED, and red/citation-ledger.md
// became a `setup` stub nothing writes (46 bytes on the 2026-08-05 run).
//
// So the screen has been reading an empty file and reporting
// "PASS — 0 REFUTED-row token(s) screened; no hits" on every record-mode run.
//
// Porting the regex onto the rendered projection would reproduce the same nothing. What the
// concept needs is a FIELD — a verification verdict on the cite event, distinct from
// confidence, which grades how much a SUPPORTING source is trusted rather than recording that
// a source does not support the claim at all. That is a schema decision with #247 in play, so
// it is tracked (#296) rather than invented here. Until then this reports the gap.
func AssemblyScreen(runDir string) Audit {
	board, err := record.BoardState(runDir)
	if err != nil {
		return Audit{Check: "assembly-screen", Verdict: "SKIP", Detail: "the record could not be read: " + err.Error()}
	}
	cites := 0
	for _, e := range board.Events {
		if e.Type == "cite" {
			cites++
		}
	}
	if cites == 0 {
		return Audit{Check: "assembly-screen", Verdict: "SKIP", Detail: "no citations on the record — nothing to screen"}
	}
	return Audit{Check: "assembly-screen", Verdict: "SKIP",
		Detail: fmt.Sprintf("%d citation(s) on the record, and NONE can be screened: a cite carries confidence (high|medium|low) but no verification VERDICT, so \"red checked this source and it does not support the claim\" is unrecordable. This check reported PASS on an empty stub before saying so. Tracked: #296", cites)}
}

// ---- AUDIT 6: record parity ----

// RecordParityAudit checks each red round has a blue sitting and a blue ROUND RECORD, floor
// redRounds-1. redRounds/blueBlocks come from the debate view, read in-process.
//
// THE ROUND RECORD IS COUNTED FROM `revision` EVENTS, not from headings in blue/CHANGELOG.md.
// The file is authored by hand while the event is emitted by the tool, so counting the file
// audited the seat's typing rather than the record — and the two disagree: the 2026-08-05 run
// carried a 6,847-byte CHANGELOG and exactly ONE revision event, from one of three eligible
// blue seats. The old regex pair (`^#+.*Round \d+` then `Round (\d+)`) read the plausible
// number and passed; the record shows the round records were never filed. See #268.
func RecordParityAudit(runDir string, redRounds, blueBlocks int) Audit {
	if redRounds == 0 {
		return Audit{Check: "record-parity", Verdict: "SKIP", Detail: "no red rounds on record"}
	}
	rounds := map[int]bool{}
	if board, err := record.BoardState(runDir); err == nil {
		for _, e := range board.Events {
			if e.Type == "revision" {
				rounds[e.Round] = true
			}
		}
	}
	clRounds := len(rounds)
	ok := blueBlocks >= redRounds-1 && clRounds >= redRounds-1
	v := "FAIL"
	if ok {
		v = "PASS"
	}
	return Audit{Check: "record-parity", Verdict: v,
		Detail: fmt.Sprintf("%d red round(s) vs %d blue sitting(s) and %d recorded round record(s) (floor: redRounds-1 — a PASS exit has no final blue response)",
			redRounds, blueBlocks, clRounds)}
}

// ---- AUDIT 7: record-join ----

type recEvent struct {
	SeatID string
	Type   string
	Key    string
}

var (
	reInvocBin = regexp.MustCompile(`feov-record"?\s+(?:lens|merge|blue|bench)\s+([a-z-]+)[\s\S]*?--seat-id\s+(\S+)`)
	reInvocMjs = regexp.MustCompile(`(?:red-lens|red-merge|blue|bench)\.mjs\s+([a-z-]+)[\s\S]*?--seat-id\s+(\S+)`)
)

func verbOf(t string) string {
	if t == "class-new" {
		return "mint"
	}
	return t
}

func RecordJoinAudit(runDir, transcriptDir string, agentFiles []string) Audit {
	recDir := filepath.Join(runDir, "records")
	entries, err := os.ReadDir(recDir)
	if err != nil {
		return Audit{Check: "record-join", Verdict: "SKIP", Detail: "no records/ (pre-record-tool run)"}
	}
	var events []recEvent
	for _, e := range entries {
		n := e.Name()
		if !strings.HasPrefix(n, "events-") || !strings.HasSuffix(n, ".jsonl") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(recDir, n))
		if err != nil {
			continue
		}
		for _, l := range strings.Split(string(b), "\n") {
			if l == "" {
				continue
			}
			var ev map[string]any
			if json.Unmarshal([]byte(l), &ev) != nil {
				continue
			}
			if strOf(ev["type"]) == "register" {
				continue
			}
			events = append(events, recEvent{SeatID: strOf(ev["seatId"]), Type: strOf(ev["type"]), Key: strOf(ev["key"])})
		}
	}
	if len(events) == 0 {
		return Audit{Check: "record-join", Verdict: "SKIP", Detail: "records/ empty"}
	}
	invocations := map[string]bool{}
	var tailWarns []string
	for _, f := range agentFiles {
		b, err := os.ReadFile(filepath.Join(transcriptDir, f))
		if err != nil {
			continue
		}
		var cmds []int
		toolCallCount := 0
		for _, l := range strings.Split(string(b), "\n") {
			if !strings.Contains(l, "tool_use") {
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
				toolCallCount++
				cmd := ""
				if inp, ok := c["input"].(map[string]any); ok {
					cmd = strOf(inp["command"])
				}
				var mm []string
				if mm = reInvocBin.FindStringSubmatch(cmd); mm == nil {
					mm = reInvocMjs.FindStringSubmatch(cmd)
				}
				if mm != nil {
					invocations[mm[2]+"::"+mm[1]] = true
					cmds = append(cmds, toolCallCount)
				}
			}
		}
		if len(cmds) > 5 && cmds[0] > toolCallCount-len(cmds)-2 {
			tailWarns = append(tailWarns, fmt.Sprintf("%s: %d record invocations all tail-clustered (back-fill pattern — parity may be vacuous for this seat)", f, len(cmds)))
		}
	}
	var orphans []recEvent
	for _, e := range events {
		if !invocations[e.SeatID+"::"+verbOf(e.Type)] {
			orphans = append(orphans, e)
		}
	}
	parts := []string{fmt.Sprintf("%d event(s) across shards; %d distinct (seat, verb) invocations in transcripts", len(events), len(invocations))}
	if len(orphans) > 0 {
		lim := orphans
		if len(lim) > 10 {
			lim = lim[:10]
		}
		ol := make([]string, len(lim))
		for i, e := range lim {
			ol[i] = fmt.Sprintf("    - %s %s (%s)", e.SeatID, e.Type, e.Key)
		}
		parts = append(parts, fmt.Sprintf("FLAGGED %d event(s) with no matching transcript invocation:\n%s", len(orphans), strings.Join(ol, "\n")))
	} else {
		parts = append(parts, "every event traces to a transcript invocation")
	}
	for _, w := range tailWarns {
		parts = append(parts, "WARN "+w)
	}
	v := "PASS"
	if len(orphans) > 0 {
		v = "FAIL"
	} else if len(tailWarns) > 0 {
		v = "WARN"
	}
	return Audit{Check: "record-join", Verdict: v, Detail: strings.Join(parts, "\n  ")}
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
func AttestationAudit(runDir, transcriptDir string, agentFiles []string, sampleFloor int) Audit {
	board, err := record.BoardState(runDir)
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
		if g.Closure.Str("carried_from") != "" {
			continue
		}
		seat, tool, target := g.Closure.Str("anchor_seat"), g.Closure.Str("anchor_tool"), g.Closure.Str("anchor_target")
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
	var unreconciled []Claim
	for _, c := range sampled {
		needle := mostDistinctive(c.Target)
		if needle == "" {
			c.Why = "anchor target too vague to reconcile"
			unreconciled = append(unreconciled, c)
			continue
		}
		found := false
		for _, k := range calls {
			if strings.Contains(k, needle) {
				found = true
				break
			}
		}
		if !found {
			c.Why = "no tool call in any transcript touches " + needle
			unreconciled = append(unreconciled, c)
		}
	}
	if len(unreconciled) > 0 {
		ul := make([]string, len(unreconciled))
		for i, u := range unreconciled {
			ul[i] = fmt.Sprintf("    - %s (%s | %s): %s", u.ID, u.Seat, u.Tool, u.Why)
		}
		detail := fmt.Sprintf("%d/%d sampled closure(s) NOT reconcilable against the trajectories — a claimed act with no matching tool call:\n%s\n    This rules on whether the RECORD IS HONEST, never on the merits of the gap.",
			len(unreconciled), len(sampled), strings.Join(ul, "\n"))
		return Audit{Check: "attestation-integrity", Verdict: "FAIL", Detail: detail, Unreconciled: unreconciled}
	}
	return Audit{Check: "attestation-integrity", Verdict: "PASS",
		Detail: fmt.Sprintf("%d/%d anchored closure(s) sampled and reconciled against actual tool calls", len(sampled), len(claims))}
}

// splitAfter mirrors JS `archive.split(/^## /m).slice(1)`: split on the delimiter, drop the head.
func splitAfter(s string, re *regexp.Regexp) []string {
	if s == "" {
		return nil
	}
	locs := re.FindAllStringIndex(s, -1)
	if len(locs) == 0 {
		return nil
	}
	var out []string
	for i, loc := range locs {
		start := loc[1]
		end := len(s)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		out = append(out, s[start:end])
	}
	return out
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

func ModelTierAudit(runDir, transcriptDir string, agentFiles []string) Audit {
	model, judgmentModel := cost.TierConfig(runDir)
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
	tiers := fmt.Sprintf("bulk=%s, judgment=%s", orQ(model), orQ(judgmentModel))
	var detail string
	if len(findings) > 0 {
		fs := make([]string, len(findings))
		for i, f := range findings {
			fs[i] = f.Verdict + " " + f.Why
		}
		detail = tiers + "; " + strings.Join(fs, "; ")
	} else {
		detail = "every seat ran on its configured tier (" + tiers + ")"
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

// ---- precedent harvest ----

// HarvestResult is the writeScorecards-style report struct for the harvest.
type HarvestResult struct {
	Written bool
	Count   int
	Path    string
	Reason  string
}

var (
	reTrailingSep = regexp.MustCompile(`[\\/]+$`)
	reSep         = regexp.MustCompile(`[\\/]`)
)

func slugOf(runDir string) string {
	s := reTrailingSep.ReplaceAllString(runDir, "")
	parts := reSep.Split(s, -1)
	return parts[len(parts)-1]
}

type ruling struct {
	kind, gapID, petitioner, disposition, rationale string
}

func HarvestPrecedents(runDir string, results []map[string]any, lawDir string) HarvestResult {
	slug := slugOf(runDir)
	var rulings []ruling
	for _, r := range results {
		if rs, ok := r["resolutions"].([]any); ok {
			for _, x := range rs {
				if m, ok := x.(map[string]any); ok {
					rulings = append(rulings, ruling{kind: "docket", gapID: strOf(m["gap_id"]), disposition: strOf(m["resolution"]), rationale: strOf(m["rationale"])})
				}
			}
		}
		if rs, ok := r["rulings"].([]any); ok {
			for _, x := range rs {
				if m, ok := x.(map[string]any); ok {
					rulings = append(rulings, ruling{kind: "petition", petitioner: strOf(m["petitioner"]), disposition: strOf(m["ruling"]), rationale: strOf(m["opinion"])})
				}
			}
		}
	}
	if len(rulings) == 0 {
		return HarvestResult{Written: false, Count: 0}
	}
	if _, err := os.Stat(lawDir); err != nil {
		return HarvestResult{Written: false, Count: len(rulings), Reason: "no law/ dir at repo root"}
	}
	_ = os.MkdirAll(filepath.Join(lawDir, "proposed"), 0o755)
	out := filepath.Join(lawDir, "proposed", slug+".md")
	var body []string
	body = append(body, "# proposed holdings — "+slug+" [ALL PERSUASIVE — awaiting human review per law/README.md]", "")
	for i, r := range rulings {
		q := "disposition of " + r.gapID
		src := slug + ", " + r.gapID
		if r.kind == "petition" {
			q = "petition by " + r.petitioner
			src = slug + ", petition:" + r.petitioner
		}
		body = append(body, strings.Join([]string{
			"## " + slug + "-" + strconv.Itoa(i+1) + " [PERSUASIVE]",
			"facts: <reviewer: fill from the cited record — the harvest never invents>",
			"question: " + q,
			"holding: " + r.disposition,
			"rationale: " + r.rationale,
			"scope-limits: <reviewer: state what this assumed>",
			"source: " + src,
			"",
		}, "\n"))
	}
	if err := os.WriteFile(out, []byte(strings.Join(body, "\n")), 0o644); err != nil {
		return HarvestResult{Written: false, Count: len(rulings), Reason: "no law/ dir at repo root"}
	}
	return HarvestResult{Written: true, Count: len(rulings), Path: out}
}

// ---- scorecards ----

// ScorecardResult mirrors the writeScorecards return.
type ScorecardResult struct {
	Written bool
	Rows    int
	Chairs  int
	Reason  string
}

func WriteScorecards(runDir string, results []map[string]any, memoryDir string, board *record.Board) ScorecardResult {
	if _, err := os.Stat(memoryDir); err != nil {
		return ScorecardResult{Written: false, Reason: fmt.Sprintf("no %s — scorecards need the tracked memory dir", memoryDir)}
	}
	cards := scorecard.Compute(runDir, results, board)
	label := labelOf(runDir)
	rows := 0
	chairs := 0
	for chair, chairRows := range cards {
		p := filepath.Join(memoryDir, chair+"-scorecard.md")
		head := scorecard.ChairHeader(chair)
		if b, err := os.ReadFile(p); err == nil {
			head = string(b)
		}
		body := trimTrailingNewlines(head) + "\n" + scorecard.RenderChair(chair, chairRows, label)
		_ = os.WriteFile(p, []byte(body), 0o644)
		rows += len(chairRows)
		chairs++
	}
	return ScorecardResult{Written: true, Rows: rows, Chairs: chairs}
}

var reTrailNL = regexp.MustCompile(`\n+$`)

// trimTrailingNewlines mirrors JS `head.replace(/\n+$/, '\n')` — collapse a trailing newline run
// to a single '\n'.
func trimTrailingNewlines(s string) string {
	return reTrailNL.ReplaceAllString(s, "\n")
}

// labelOf mirrors JS `runDir.split(/[\\/]/).filter(Boolean).pop() || 'run'`.
func labelOf(runDir string) string {
	parts := reSep.Split(runDir, -1)
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

// appendCostToReport folds the per-seat-round cost table from cost.md into report.md as a ## Cost
// section, so the final report carries the cost breakdown. No-op (returns "") when report.md is
// absent (a pre-record run), already carries a ## Cost section (idempotent across re-runs), or
// cost.md has no table. cost.md itself is left byte-identical.
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
	body := strings.TrimRight(string(report), "\n") + "\n\n## Cost\n\n" + table + "\n"
	if err := os.WriteFile(reportPath, []byte(body), 0o644); err != nil {
		return "report.md: cost append FAILED — " + jsSlice(err.Error(), 200)
	}
	return "report.md: cost breakdown folded in (## Cost)"
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
func Run(runDir, transcriptDir string) (audits []Audit, report string, exitFail bool, err error) {
	var lines []string

	// Mechanics: journal copy, transcript tarball, cost.md.
	if err = copyFile(filepath.Join(transcriptDir, "journal.jsonl"), filepath.Join(runDir, "trajectories", "journal.jsonl")); err != nil {
		return nil, "", false, err
	}
	agentFiles, err := listAgentFiles(transcriptDir)
	if err != nil {
		return nil, "", false, err
	}
	tarPath := filepath.Join(runDir, "trajectories", "agent-transcripts.tar.gz")
	if terr := writeTarball(transcriptDir, tarPath, agentFiles); terr != nil {
		lines = append(lines, "tarball: FAILED — "+jsSlice(terr.Error(), 200))
	} else {
		lines = append(lines, fmt.Sprintf("tarball: %d transcript(s)", len(agentFiles)))
	}

	costF, cerr := os.Create(filepath.Join(runDir, "cost.md"))
	if cerr != nil {
		return nil, "", false, cerr
	}
	if rerr := cost.Report(transcriptDir, runDir, costF); rerr != nil {
		costF.Close()
		lines = append(lines, "cost.md: FAILED — "+jsSlice(rerr.Error(), 200))
	} else {
		costF.Close()
		lines = append(lines, "cost.md: written (telemetry join included)")
		// Fold the per-seat-round cost table into report.md as a ## Cost section — the final
		// report carries the cost breakdown. report.md is assembled mid-run WITHOUT transcript
		// access (the transcript dir reaches only capture), so this is the one stage that can.
		// Slices the already-rendered cost.md (kept byte-identical) rather than re-generate.
		if msg := appendCostToReport(filepath.Join(runDir, "report.md"), filepath.Join(runDir, "cost.md")); msg != "" {
			lines = append(lines, msg)
		}
	}

	results, friction := ReadJournal(filepath.Join(runDir, "trajectories"))

	// Record-backed reads, in-process (the JS spawned `merge show` views).
	board, _ := record.BoardState(runDir)
	redRounds, blueBlocks := 0, 0
	onRecord := []string{}
	if board != nil {
		dj := record.DebateJSONOf(board)
		for _, r := range dj.Rounds {
			if len(r.Red) > 0 {
				redRounds++
			}
			if len(r.Blue) > 0 {
				blueBlocks++
			}
		}
		for _, fr := range record.FrictionJSONOf(board).Friction {
			onRecord = append(onRecord, fr.Text)
		}
	}

	audits = []Audit{
		TelemetryAudit(runDir, redRounds),
		ShardAudit(runDir, results),
		FrictionAudit(friction, onRecord),
		ContextUse(transcriptDir, agentFiles),
		AssemblyScreen(runDir),
		RecordParityAudit(runDir, redRounds, blueBlocks),
		RecordJoinAudit(runDir, transcriptDir, agentFiles),
		AttestationAudit(runDir, transcriptDir, agentFiles, 5),
		ModelTierAudit(runDir, transcriptDir, agentFiles),
	}

	cwd, _ := os.Getwd()
	sc := WriteScorecards(runDir, results, filepath.Join(cwd, "feov-memory"), board)
	if sc.Written {
		lines = append(lines, fmt.Sprintf("scorecards: %d row(s) across %d chair(s) -> feov-memory/", sc.Rows, sc.Chairs))
	} else {
		lines = append(lines, "scorecards: "+sc.Reason)
	}

	prec := HarvestPrecedents(runDir, results, filepath.Join(cwd, "law"))
	switch {
	case prec.Written:
		lines = append(lines, fmt.Sprintf("precedent harvest: %d ruling(s) -> %s (PERSUASIVE, awaiting review)", prec.Count, prec.Path))
	case prec.Count > 0:
		lines = append(lines, fmt.Sprintf("precedent harvest: %d ruling(s), %s", prec.Count, prec.Reason))
	default:
		lines = append(lines, "precedent harvest: no rulings this run")
	}

	marker := filepath.Join(cwd, ".claude", "run-live.json")
	if _, serr := os.Stat(marker); serr == nil {
		_ = os.Remove(marker)
		lines = append(lines, "run-live marker: removed")
	}

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
	if werr := os.WriteFile(filepath.Join(runDir, "run-record-audit.md"), []byte(report), 0o644); werr != nil {
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
