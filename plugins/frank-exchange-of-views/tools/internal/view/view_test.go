package view

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// ---- fixtures ----

// writeShard writes a run's shard file directly (events-<seat>-<nonce>.jsonl), the
// low-level form the view's replay reads. It uses record.MarshalEvent for the exact
// on-disk line encoding.
func writeShard(t *testing.T, runDir, seatID, nonce string, evs []record.Event) {
	t.Helper()
	recs := filepath.Join(runDir, "records")
	if err := os.MkdirAll(recs, 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for _, e := range evs {
		line, err := record.MarshalEvent(e)
		if err != nil {
			t.Fatal(err)
		}
		b.Write(line)
		b.WriteByte('\n')
	}
	p := filepath.Join(recs, "events-"+seatID+"-"+nonce+".jsonl")
	if err := os.WriteFile(p, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

func ev(seatID, nonce string, seq, round int, typ, key string, p *record.Payload) record.Event {
	if p == nil {
		p = record.NewPayload()
	}
	return record.Event{Seq: seq, SeatID: seatID, Nonce: nonce, Round: round, Type: typ, Key: key, Payload: p}
}

// md is view.Markdown or a fatal.
func md(t *testing.T, runDir, name string) string {
	t.Helper()
	b, err := Markdown(runDir, name, "")
	if err != nil {
		t.Fatalf("Markdown(%q): %v", name, err)
	}
	return string(b)
}

// ---- helper-function tests (moved from record/render_test.go with the funcs) ----

func TestTruncateCountsUTF16CodeUnits(t *testing.T) {
	cases := []struct {
		name string
		in   string
		n    int
		want string
	}{
		{"ascii under the limit is returned whole", "hello", 10, "hello"},
		{"ascii at exactly the limit", "hello", 5, "hello"},
		{"ascii over the limit", "hello world", 5, "hello"},
		{"zero cuts everything", "hello", 0, ""},
		{"empty input", "", 5, ""},
		{"em-dash costs one unit, not three bytes", "a—b—c", 3, "a—b"},
		{"cjk costs one unit each", "日本語です", 3, "日本語"},
		{"cjk under the limit", "日本語", 10, "日本語"},
		{"an astral emoji costs two units", "😀😀😀", 4, "😀😀"},
		{"cutting mid-surrogate keeps the pair count honest", "a😀b", 2, "a\uFFFD"},
		{"mixed scripts, cutting into the surrogate pair", "aé日😀", 4, "aé日\uFFFD"},
		{"mixed scripts, cutting before the pair", "aé日😀", 3, "aé日"},
		{"combining marks are counted separately, as JS does", "e" + string(rune(0x0301)) + "x", 2, "e" + string(rune(0x0301))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := truncate(tc.in, tc.n)
			if got != tc.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tc.in, tc.n, got, tc.want)
			}
			if u := len(utf16.Encode([]rune(got))); u > tc.n {
				t.Errorf("result is %d code units, over the limit of %d", u, tc.n)
			}
		})
	}
}

func TestTruncateNeverExceedsTheLimitForAnyScript(t *testing.T) {
	inputs := []string{
		"plain ascii prose that runs on for a while",
		"em—dashes · and ✓ marks throughout the sentence",
		"日本語のタイトルを引用しています",
		"😀😀😀 astral plane emoji sequence 😀😀😀",
		"mixed aé日😀 together in one string",
		strings.Repeat("é", 200),
	}
	for _, in := range inputs {
		total := len(utf16.Encode([]rune(in)))
		for n := 0; n <= total+2; n++ {
			got := truncate(in, n)
			if u := len(utf16.Encode([]rune(got))); u > n {
				t.Fatalf("truncate(%q, %d) returned %d units", in, n, u)
			}
			if n >= total && got != in {
				t.Fatalf("truncate(%q, %d) shortened a string already within the limit", in, n)
			}
		}
	}
}

func TestJsTextMirrorsTemplateLiteralInterpolation(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
	}{
		{"nil is the literal word", nil, "undefined"},
		{"a string is itself", "high", "high"},
		{"an empty string stays empty, unlike nil", "", ""},
		{"true", true, "true"},
		{"false", false, "false"},
		{"an int", 3, "3"},
		{"a json.Number keeps its text", json.Number("2.50"), "2.50"},
		{"an integral float drops the decimal point", 46.0, "46"},
		{"a fractional float", 0.5, "0.5"},
		{"a negative integral float", -3.0, "-3"},
		{"zero", 0.0, "0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := jsText(tc.in); got != tc.want {
				t.Errorf("jsText(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestUndefStrDistinguishesAbsentFromEmpty(t *testing.T) {
	p := record.NewPayload().Set("present", "v").Set("empty", "").Set("null", nil)
	if got := undefStr(p, "missing"); got != "undefined" {
		t.Errorf("an absent key renders %q, want \"undefined\"", got)
	}
	if got := undefStr(p, "empty"); got != "" {
		t.Errorf("a present-but-empty key renders %q, want empty", got)
	}
	if got := undefStr(p, "null"); got != "undefined" {
		t.Errorf("an explicit null renders %q, want \"undefined\"", got)
	}
	if got := undefStr(p, "present"); got != "v" {
		t.Errorf("a present key renders %q", got)
	}
}

func TestJsNum(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{46, "46"}, {0, "0"}, {-1, "-1"}, {0.5, "0.5"}, {2.25, "2.25"},
		{-0.75, "-0.75"}, {12.25, "12.25"},
		{math.NaN(), "null"}, {math.Inf(1), "null"}, {math.Inf(-1), "null"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			if got := jsNum(tc.in); got != tc.want {
				t.Errorf("jsNum(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestRound2(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{0.6, 0.6}, {1.0, 1}, {0.666666, 0.67}, {0.664, 0.66}, {1.005, 1.0},
		{2.0 / 3.0, 0.67}, {-0.666, -0.67}, {0, 0}, {100, 100},
	}
	for _, tc := range cases {
		t.Run(jsNum(tc.in), func(t *testing.T) {
			if got := round2(tc.in); got != tc.want {
				t.Errorf("round2(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestMassSumIgnoresUngradedGaps(t *testing.T) {
	gaps := []*record.Gap{
		{Likelihood: "medium", Impact: "high"},
		{Likelihood: "low", Impact: "low"},
		{Likelihood: nil, Impact: "high"},
		{Likelihood: true, Impact: "high"},
		{Likelihood: "realized", Impact: "high"},
	}
	if got, want := massSum(gaps), 7.0; got != want {
		t.Errorf("massSum = %v, want %v", got, want)
	}
	if got := massSum(nil); got != 0 {
		t.Errorf("massSum(nil) = %v, want 0", got)
	}
}

// ---- projection tests (relocated: call view.Markdown / TelemetryJSONL / Counts) ----

func TestMarkdownOnAnEmptyRun(t *testing.T) {
	runDir := t.TempDir()
	open, closed, anomalies, err := Counts(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if open != 0 || closed != 0 || anomalies != 0 {
		t.Errorf("empty run reported open=%d closed=%d anomalies=%d", open, closed, anomalies)
	}
	for _, name := range []string{"ledger", "archive", "debate", "changelog", "lines-of-inquiry", "citation-ledger", "changes"} {
		if _, err := Markdown(runDir, name, ""); err != nil {
			t.Errorf("projection %s errored on an empty run: %v", name, err)
		}
	}
	// Telemetry with no rounds is EMPTY, not a blank line.
	b, err := TelemetryJSONL(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 0 {
		t.Errorf("empty telemetry = %q, want zero bytes", b)
	}
}

func TestMarkdownLedgerAndArchive(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	writeShard(t, runDir, seatID, "aaaaaaaa", []record.Event{
		ev(seatID, "aaaaaaaa", 0, 1, "mint", seatID+":mint:R1-1", record.NewPayload().
			Set("gap_id", "R1-1").Set("problem", "an open problem").Set("location", "§2").
			Set("class", "scope-creep").Set("required_fix", "do the thing").
			Set("acceptance_check", "run the check").
			Set("severity", "high").Set("likelihood", "medium").Set("impact", "high")),
		ev(seatID, "aaaaaaaa", 1, 1, "mint", seatID+":mint:R1-2", record.NewPayload().
			Set("gap_id", "R1-2").Set("problem", "a closed problem").Set("location", "§3")),
		ev(seatID, "aaaaaaaa", 2, 1, "close", seatID+":close:R1-2", record.NewPayload().
			Set("gap_id", "R1-2").Set("closure_class", "closed_with_regression").
			Set("successor", "R2-1").
			Set("anchor_seat", "L1").Set("anchor_tool", "git show").Set("anchor_target", "7bc501e:f")),
		ev(seatID, "aaaaaaaa", 3, 1, "mint", seatID+":mint:R1-3", record.NewPayload().
			Set("gap_id", "R1-3").Set("problem", "an unclassed problem").Set("location", "§4")),
	})
	open, closed, _, err := Counts(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if open != 2 || closed != 1 {
		t.Errorf("counts = %d open / %d closed, want 2/1", open, closed)
	}
	ledger := md(t, runDir, "ledger")
	if !strings.Contains(ledger, "## OPEN GAPS (2)") {
		t.Errorf("open-gap heading missing:\n%s", ledger)
	}
	if !strings.Contains(ledger, "### R1-1 — an open problem") {
		t.Errorf("the open gap is not in the ledger:\n%s", ledger)
	}
	if strings.Contains(ledger, "### R1-2") {
		t.Error("a CLOSED gap appears in the open section")
	}
	if !strings.Contains(ledger, "cx undefined") {
		t.Errorf("a gap minted without --cx must render \"cx undefined\":\n%s", ledger)
	}
	if !strings.Contains(ledger, "class undefined") {
		t.Errorf("a gap minted without a class must render \"class undefined\":\n%s", ledger)
	}
	if !strings.Contains(ledger, "acceptance_check: undefined") {
		t.Errorf("a gap minted without --check must render \"acceptance_check: undefined\":\n%s", ledger)
	}
	if !strings.Contains(ledger, "class scope-creep") {
		t.Errorf("a gap WITH a class must render it:\n%s", ledger)
	}
	if !strings.Contains(ledger, "R1-2 | closed_with_regression | a closed problem | R2-1") {
		t.Errorf("closure index row is wrong:\n%s", ledger)
	}

	archive := md(t, runDir, "archive")
	if !strings.Contains(archive, "verification anchor: L1 | git show | 7bc501e:f") {
		t.Errorf("the anchor triple is not in the archive:\n%s", archive)
	}
	if !strings.Contains(archive, "successor: R2-1") {
		t.Errorf("the successor is not in the archive:\n%s", archive)
	}
	if strings.Contains(archive, "R1-1") {
		t.Error("an OPEN gap appears in the archive")
	}
}

func TestMarkdownArchiveShowsCarriedClosures(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r2"
	writeShard(t, runDir, seatID, "aaaaaaaa", []record.Event{
		ev(seatID, "aaaaaaaa", 0, 2, "mint", seatID+":mint:R2-1", record.NewPayload().Set("gap_id", "R2-1").Set("problem", "p")),
		ev(seatID, "aaaaaaaa", 1, 2, "close", seatID+":close:R2-1", record.NewPayload().
			Set("gap_id", "R2-1").Set("carried_from", "1")),
	})
	archive := md(t, runDir, "archive")
	if !strings.Contains(archive, "verification anchor: CARRIED from round 1") {
		t.Errorf("a carried closure must say so:\n%s", archive)
	}
	if strings.Contains(archive, "undefined | undefined") {
		t.Errorf("a carried closure printed an empty anchor triple:\n%s", archive)
	}
}

func TestMarkdownSurfacesAnomaliesAndUndisposedObservations(t *testing.T) {
	runDir := t.TempDir()
	lens := "red-lens-r1-L1"
	writeShard(t, runDir, lens, "aaaaaaaa", []record.Event{
		ev(lens, "aaaaaaaa", 0, 1, "finding", lens+":finding:F1", record.NewPayload().
			Set("label", "F1").Set("text", strings.Repeat("long prose ", 40))),
	})
	writeShard(t, runDir, lens, "bbbbbbbb", []record.Event{
		ev(lens, "bbbbbbbb", 0, 1, "finding", lens+":finding:F2", record.NewPayload().Set("label", "F2").Set("text", "short")),
	})
	_, _, anomalies, err := Counts(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if anomalies == 0 {
		t.Fatal("a double dispatch produced no anomaly count")
	}
	ledger := md(t, runDir, "ledger")
	if !strings.Contains(ledger, "## render anomalies (never silently normalized)") {
		t.Errorf("the anomaly footer is missing:\n%s", ledger)
	}
	if !strings.Contains(ledger, "multi-nonce seat "+lens) {
		t.Errorf("the anomaly does not name the seat:\n%s", ledger)
	}
	if !strings.Contains(ledger, "## undisposed lens observations") {
		t.Errorf("the undisposed footer is missing:\n%s", ledger)
	}
	for _, line := range strings.Split(ledger, "\n") {
		if strings.HasPrefix(line, "- "+lens+" F") {
			prose := line[strings.Index(line, ": ")+2:]
			if len(utf16.Encode([]rune(prose))) > 120 {
				t.Errorf("undisposed prose was not truncated: %d units", len(utf16.Encode([]rune(prose))))
			}
		}
	}
}

func TestTelemetryIsComputed(t *testing.T) {
	runDir := t.TempDir()
	r1 := "red-merge-r1"
	r2 := "red-merge-r2"
	writeShard(t, runDir, r1, "aaaaaaaa", []record.Event{
		ev(r1, "aaaaaaaa", 0, 1, "mint", r1+":mint:R1-1", record.NewPayload().
			Set("gap_id", "R1-1").Set("problem", "p1").
			Set("severity", "high").Set("likelihood", "medium").Set("impact", "high")),
		ev(r1, "aaaaaaaa", 1, 1, "mint", r1+":mint:R1-2", record.NewPayload().
			Set("gap_id", "R1-2").Set("problem", "p2").
			Set("severity", "low").Set("likelihood", "low").Set("impact", "low")),
	})
	writeShard(t, runDir, r2, "bbbbbbbb", []record.Event{
		ev(r2, "bbbbbbbb", 0, 2, "mint", r2+":mint:R2-1", record.NewPayload().
			Set("gap_id", "R2-1").Set("problem", "p3").
			Set("severity", "low").Set("likelihood", "low").Set("impact", "low").
			Set("supersedes", []string{"R1-1"})),
		ev(r2, "bbbbbbbb", 1, 2, "close", r2+":close:R1-1", record.NewPayload().Set("gap_id", "R1-1")),
	})
	raw, err := TelemetryJSONL(runDir)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("%d telemetry lines, want one per round:\n%s", len(lines), raw)
	}
	// Key order is the contract, not merely the key set.
	if !strings.HasPrefix(lines[0], `{"round":1,"mapping_version":"v1","open_count":`) {
		t.Errorf("telemetry key order changed:\n%s", lines[0])
	}

	var r1line map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &r1line); err != nil {
		t.Fatal(err)
	}
	if got := r1line["open_count"]; got != float64(2) {
		t.Errorf("round 1 open_count = %v, want 2", got)
	}
	if got := r1line["max_severity"]; got != "high" {
		t.Errorf("round 1 max_severity = %v, want high", got)
	}
	if got := r1line["mass"]; got != float64(7) {
		t.Errorf("round 1 mass = %v, want 7", got)
	}

	var r2line map[string]any
	if err := json.Unmarshal([]byte(lines[1]), &r2line); err != nil {
		t.Fatal(err)
	}
	rr := r2line["repair_regression"].(map[string]any)
	if rr["closures"] != float64(1) || rr["lineage_mints"] != float64(1) || rr["ratio"] != float64(1) {
		t.Errorf("round 2 repair_regression = %v", rr)
	}
	ed := r2line["edge_deltas"].(map[string]any)
	if ed["down_mass"] != float64(5) || ed["up_mass"] != float64(0) {
		t.Errorf("round 2 edge_deltas = %v, want down 5 / up 0", ed)
	}
}

// THE STOPPING SIGNAL (#284). Severity answers "how bad"; it can sit flat for rounds while
// the KIND of defect being found moves — and that phase change is what tells the bench the
// remaining work is cheaper to shake out in execution than in review. The distribution and
// the repeat rate against the previous round are what make "the findings changed character"
// readable without re-reading every gap's prose.
func TestTelemetryCarriesTheClassDistributionAndRepeatRate(t *testing.T) {
	runDir := t.TempDir()
	mint := func(seat, nonce, id, class string, i, round int) record.Event {
		return ev(seat, nonce, i, round, "mint", seat+":mint:"+id, record.NewPayload().
			Set("gap_id", id).Set("problem", "p").Set("class", class).
			Set("severity", "low").Set("likelihood", "low").Set("impact", "low"))
	}
	r1, r2 := "red-merge-r1", "red-merge-r2"
	writeShard(t, runDir, r1, "aaaaaaaa", []record.Event{
		mint(r1, "aaaaaaaa", "R1-1", "evidence-claim-not-documented", 0, 1),
		mint(r1, "aaaaaaaa", "R1-2", "evidence-claim-not-documented", 1, 1),
		mint(r1, "aaaaaaaa", "R1-3", "scope-closure-missing", 2, 1),
	})
	// Round 2: one class REPEATS from round 1, one is fresh. Repeat rate = 1/2.
	writeShard(t, runDir, r2, "bbbbbbbb", []record.Event{
		mint(r2, "bbbbbbbb", "R2-1", "scope-closure-missing", 0, 2),
		mint(r2, "bbbbbbbb", "R2-2", "co-resident-rules-disagree", 1, 2),
	})

	raw, err := TelemetryJSONL(runDir)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("%d telemetry lines, want 2:\n%s", len(lines), raw)
	}

	decode := func(s string) map[string]any {
		var m map[string]any
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			t.Fatal(err)
		}
		return m["new_mint"].(map[string]any)
	}

	one := decode(lines[0])
	byClass := one["by_class"].(map[string]any)
	if byClass["evidence-claim-not-documented"] != float64(2) || byClass["scope-closure-missing"] != float64(1) {
		t.Errorf("round 1 by_class = %v, want the two classes counted 2 and 1", byClass)
	}
	// Round 1 has no predecessor, so a repeat rate would be a fabricated number.
	if one["class_repeat_rate"] != nil {
		t.Errorf("round 1 class_repeat_rate = %v, want null — there is no prior round to repeat", one["class_repeat_rate"])
	}

	two := decode(lines[1])
	if got := two["class_repeat_rate"]; got != 0.5 {
		t.Errorf("round 2 class_repeat_rate = %v, want 0.5 (1 of 2 mints repeats a round-1 class)", got)
	}
}

// A mint with no class contributes to the COUNT but not to the distribution — a "" bucket
// would read as a real class and quietly inflate the repeat rate against itself.
func TestTelemetryClasslessMintDoesNotBecomeAClass(t *testing.T) {
	runDir := t.TempDir()
	seat := "red-merge-r1"
	writeShard(t, runDir, seat, "aaaaaaaa", []record.Event{
		ev(seat, "aaaaaaaa", 0, 1, "mint", seat+":mint:R1-1", record.NewPayload().
			Set("gap_id", "R1-1").Set("problem", "p").
			Set("severity", "low").Set("likelihood", "low").Set("impact", "low")),
	})
	raw, err := TelemetryJSONL(runDir)
	if err != nil {
		t.Fatal(err)
	}
	nm := func() map[string]any {
		var m map[string]any
		if err := json.Unmarshal([]byte(strings.TrimRight(string(raw), "\n")), &m); err != nil {
			t.Fatal(err)
		}
		return m["new_mint"].(map[string]any)
	}()
	if nm["count"] != float64(1) {
		t.Errorf("count = %v, want 1", nm["count"])
	}
	if bc := nm["by_class"].(map[string]any); len(bc) != 0 {
		t.Errorf("by_class = %v, want empty — an empty class must not become a bucket", bc)
	}
}

func TestTelemetryUndefinedSeverityKey(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	writeShard(t, runDir, seatID, "aaaaaaaa", []record.Event{
		ev(seatID, "aaaaaaaa", 0, 1, "mint", seatID+":mint:R1-1", record.NewPayload().Set("gap_id", "R1-1").Set("problem", "p")),
	})
	raw, err := TelemetryJSONL(runDir)
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, `"by_severity":{"undefined":1}`) {
		t.Errorf("an ungraded mint must key by_severity under \"undefined\":\n%s", s)
	}
	if !strings.Contains(s, `"max_severity":null`) {
		t.Errorf("max_severity for an ungraded board must be null:\n%s", s)
	}
	if !strings.Contains(s, `"ratio":null`) {
		t.Errorf("ratio with no closures must be null:\n%s", s)
	}
}

func TestMarkdownDebateChangelogInquiryAndCitations(t *testing.T) {
	runDir := t.TempDir()
	merge := "red-merge-r1"
	blue := "blue-lane-1"
	judge := "judge-r1"
	lens := "red-lens-r1-L1"
	writeShard(t, runDir, merge, "aaaaaaaa", []record.Event{
		ev(merge, "aaaaaaaa", 0, 1, "position", merge+":position", record.NewPayload().Set("text", "red says so")),
		ev(merge, "aaaaaaaa", 1, 1, "closing", merge+":closing:R1-1", record.NewPayload().Set("gap_id", "R1-1").Set("text", "red closes")),
	})
	writeShard(t, runDir, blue, "bbbbbbbb", []record.Event{
		ev(blue, "bbbbbbbb", 0, 1, "position", blue+":position", record.NewPayload().Set("text", "blue says otherwise")),
		ev(blue, "bbbbbbbb", 1, 1, "revision", blue+":revision", record.NewPayload().Set("text", "blue revised")),
		ev(blue, "bbbbbbbb", 2, 1, "avenue", blue+":avenue:a1", record.NewPayload().
			Set("avenue_id", "A1").Set("status", "abandoned").Set("line", "try the archive").
			Set("method", "full-text search").Set("reason", "the archive is offline")),
		ev(blue, "bbbbbbbb", 3, 1, "avenue", blue+":avenue:a2", record.NewPayload().
			Set("avenue_id", "A2").Set("status", "pursued").Set("line", "read the source")),
	})
	writeShard(t, runDir, judge, "cccccccc", []record.Event{
		ev(judge, "cccccccc", 0, 1, "opinion", judge+":opinion:R1-1", record.NewPayload().
			Set("gap_id", "R1-1").Set("disposition", "upheld").Set("principle", "correctness first").
			Set("tension", "economy").Set("review_flag", "none").Set("rationale", "because")),
	})
	writeShard(t, runDir, lens, "dddddddd", []record.Event{
		ev(lens, "dddddddd", 0, 1, "cite", lens+":cite:https://x", record.NewPayload().
			Set("claim", "the claim").Set("reference", "https://x").
			Set("confidence", "high").Set("access_date", "2026-07-18")),
		ev(lens, "dddddddd", 1, 1, "cite", lens+":cite:https://y", record.NewPayload().Set("reference", "https://y")),
	})

	debate := md(t, runDir, "debate")
	for _, want := range []string{"## Round 1", "### RED\nred says so", "### BLUE\nblue says otherwise",
		"### RED CLOSING (round 1) — R1-1", "### LEAD", "upheld — principle: correctness first"} {
		if !strings.Contains(debate, want) {
			t.Errorf("debate is missing %q:\n%s", want, debate)
		}
	}

	changelog := md(t, runDir, "changelog")
	if !strings.Contains(changelog, "## Round 1\nblue revised") {
		t.Errorf("changelog is wrong:\n%s", changelog)
	}

	inquiry := md(t, runDir, "lines-of-inquiry")
	pursuedAt := strings.Index(inquiry, "## pursued (1)")
	abandonedAt := strings.Index(inquiry, "## abandoned (1)")
	if pursuedAt < 0 || abandonedAt < 0 || pursuedAt > abandonedAt {
		t.Errorf("lines-of-inquiry sections are wrong or out of order:\n%s", inquiry)
	}
	if !strings.Contains(inquiry, "- **A1 try the archive** _(full-text search)_ — the archive is offline (blue-lane-1)") {
		t.Errorf("an abandoned avenue row is wrong:\n%s", inquiry)
	}
	if strings.Contains(inquiry, "## declined") {
		t.Errorf("an empty status produced a heading:\n%s", inquiry)
	}

	cites := md(t, runDir, "citation-ledger")
	if !strings.Contains(cites, "the claim | https://x | high | r1 | 2026-07-18") {
		t.Errorf("citation row is wrong:\n%s", cites)
	}
	if !strings.Contains(cites, "undefined | https://y | undefined | r1 | undefined") {
		t.Errorf("a sparse citation must render its gaps as \"undefined\":\n%s", cites)
	}
}

func TestMarkdownDebateSkipsEmptyRounds(t *testing.T) {
	runDir := t.TempDir()
	lens := "red-lens-r3-L1"
	writeShard(t, runDir, lens, "aaaaaaaa", []record.Event{
		ev(lens, "aaaaaaaa", 0, 3, "friction", lens+":friction:#1", record.NewPayload().Set("text", "not debate content")),
	})
	debate := md(t, runDir, "debate")
	if strings.Contains(debate, "## Round 3") {
		t.Errorf("a round with no positions, closings or opinions got a heading:\n%s", debate)
	}
}

// Views are FULL STATE derived from the record: re-rendering the same log is byte-identical.
func TestMarkdownIsDeterministic(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	writeShard(t, runDir, seatID, "aaaaaaaa", []record.Event{
		ev(seatID, "aaaaaaaa", 0, 1, "mint", seatID+":mint:R1-1", record.NewPayload().
			Set("gap_id", "R1-1").Set("problem", "p <with> & entities").
			Set("severity", "high").Set("likelihood", "medium").Set("impact", "high")),
		ev(seatID, "aaaaaaaa", 1, 1, "mint", seatID+":mint:R1-2", record.NewPayload().
			Set("gap_id", "R1-2").Set("problem", "q").Set("severity", "low")),
	})
	names := []string{"ledger", "archive", "debate", "changelog", "lines-of-inquiry", "citation-ledger"}
	first := map[string]string{}
	for _, n := range names {
		first[n] = md(t, runDir, n)
	}
	for i := 0; i < 3; i++ {
		for _, n := range names {
			if got := md(t, runDir, n); got != first[n] {
				t.Errorf("%s changed on re-render %d — the projection is not full state", n, i+1)
			}
		}
	}
	// Prose entities reach the projection unescaped: a markdown file is not HTML.
	if !strings.Contains(first["ledger"], "p <with> & entities") {
		t.Errorf("prose was escaped in the ledger:\n%s", first["ledger"])
	}
}
