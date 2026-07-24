package record

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"
)

// The render is what a human reads. Its two hard parts are JavaScript's string
// and number semantics, both of which produced real port bugs: byte slicing that
// cut multibyte prose in the wrong place, and `${undefined}` interpolating five
// literal characters where Go would produce none.

// truncate counts UTF-16 code units, as JS String.prototype.slice does — not
// bytes, and not runes. All three disagree on real seat prose.
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
		// An em-dash is 3 BYTES but 1 code unit: byte slicing would keep fewer
		// characters and could split the rune into mojibake.
		{"em-dash costs one unit, not three bytes", "a—b—c", 3, "a—b"},
		{"cjk costs one unit each", "日本語です", 3, "日本語"},
		{"cjk under the limit", "日本語", 10, "日本語"},
		// An emoji outside the BMP is a surrogate PAIR: 2 units, 4 bytes, 1 rune.
		{"an astral emoji costs two units", "😀😀😀", 4, "😀😀"},
		{"cutting mid-surrogate keeps the pair count honest", "a😀b", 2, "a�"},
		// a(1) é(1) 日(1) 😀(2) = 5 units, so a cut at 4 lands INSIDE the surrogate
		// pair and leaves a lone high surrogate — which is what JS slice(0,4) does
		// too. Decoding it back yields the replacement character.
		{"mixed scripts, cutting into the surrogate pair", "aé日😀", 4, "aé日�"},
		{"mixed scripts, cutting before the pair", "aé日😀", 3, "aé日"},
		{"combining marks are counted separately, as JS does", "éx", 2, "é"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := truncate(tc.in, tc.n)
			if got != tc.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tc.in, tc.n, got, tc.want)
			}
			// The invariant that makes this correct regardless of the case table:
			// the result never exceeds n code units.
			if u := len(utf16.Encode([]rune(got))); u > tc.n {
				t.Errorf("result is %d code units, over the limit of %d", u, tc.n)
			}
		})
	}
}

// The property that byte slicing violates: for prose of any script, truncating
// at n never returns more than n UTF-16 units and never returns more than the
// input.
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

// `${undefined}` is the five-character string "undefined", not empty. Every
// ledger for a gap minted without --cx literally reads "cx undefined".
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
	p := NewPayload().Set("present", "v").Set("empty", "").Set("null", nil)
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

// JS prints integral floats without a decimal point and uses the shortest
// round-tripping representation otherwise; NaN and Infinity stringify to null in
// JSON.
func TestJsNum(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{46, "46"},
		{0, "0"},
		{-1, "-1"},
		{0.5, "0.5"},
		{2.25, "2.25"},
		{-0.75, "-0.75"},
		{12.25, "12.25"},
		{math.NaN(), "null"},
		{math.Inf(1), "null"},
		{math.Inf(-1), "null"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			if got := jsNum(tc.in); got != tc.want {
				t.Errorf("jsNum(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// round2 mirrors `+x.toFixed(2)`: fix to two decimals, then re-read as a number
// so trailing zeros vanish.
func TestRound2(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{0.6, 0.6},
		{1.0, 1},
		{0.666666, 0.67},
		{0.664, 0.66},
		{1.005, 1.0}, // binary representation puts it just below the midpoint
		{2.0 / 3.0, 0.67},
		{-0.666, -0.67},
		{0, 0},
		{100, 100},
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
	gaps := []*Gap{
		{Likelihood: "medium", Impact: "high"}, // 2 * 3
		{Likelihood: "low", Impact: "low"},     // 1 * 1
		{Likelihood: nil, Impact: "high"},      // absent -> 0
		{Likelihood: true, Impact: "high"},     // non-string -> 0
		{Likelihood: "realized", Impact: "high"},
	}
	if got, want := massSum(gaps), 7.0; got != want {
		t.Errorf("massSum = %v, want %v", got, want)
	}
	if got := massSum(nil); got != 0 {
		t.Errorf("massSum(nil) = %v, want 0", got)
	}
}

// Render on an empty run must still produce every projection: a reader opening
// ledger.md before the first mint should find a file, not a missing path.
func TestRenderOnAnEmptyRunWritesEveryProjection(t *testing.T) {
	runDir := t.TempDir()
	res, err := Render(runDir, "")
	if err != nil {
		t.Fatal(err)
	}
	if res.Open != 0 || res.Closed != 0 || res.Anomalies != 0 {
		t.Errorf("empty run reported %+v", res)
	}
	for _, name := range []string{"ledger.md", "archive.md", "board-telemetry.jsonl", "debate.md", "CHANGELOG.md", "lines-of-inquiry.md", "citation-ledger.md"} {
		if _, err := os.Stat(filepath.Join(res.Out, name)); err != nil {
			t.Errorf("projection %s was not written: %v", name, err)
		}
	}
	// Telemetry with no rounds is EMPTY, not a blank line: a stray newline would
	// parse as a malformed record downstream.
	b, err := os.ReadFile(filepath.Join(res.Out, "board-telemetry.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 0 {
		t.Errorf("empty telemetry = %q, want zero bytes", b)
	}
}

// Render honours an explicit output directory; the shadow default is opt-out.
func TestRenderWritesToAnExplicitOutDir(t *testing.T) {
	runDir := t.TempDir()
	out := filepath.Join(t.TempDir(), "nested", "target")
	res, err := Render(runDir, out)
	if err != nil {
		t.Fatal(err)
	}
	if res.Out != out {
		t.Errorf("Out = %q, want %q", res.Out, out)
	}
	if _, err := os.Stat(filepath.Join(out, "ledger.md")); err != nil {
		t.Errorf("nothing was written to the explicit out dir: %v", err)
	}
	// The shadow dir must NOT be created when an explicit target was given.
	if _, err := os.Stat(filepath.Join(recordsDir(runDir), "render-shadow", "ledger.md")); err == nil {
		t.Error("the shadow projection was written as well as the explicit one")
	}
}

func TestRenderLedgerAndArchive(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	writeShard(t, runDir, seatID, "aaaaaaaa", []Event{
		ev(seatID, "aaaaaaaa", 0, 1, "mint", seatID+":mint:R1-1", NewPayload().
			Set("gap_id", "R1-1").Set("problem", "an open problem").Set("location", "§2").
			Set("class", "scope-creep").Set("required_fix", "do the thing").
			Set("acceptance_check", "run the check").
			Set("severity", "high").Set("likelihood", "medium").Set("impact", "high")),
		// A gap minted WITHOUT --cx and without a class: both render as "undefined".
		ev(seatID, "aaaaaaaa", 1, 1, "mint", seatID+":mint:R1-2", NewPayload().
			Set("gap_id", "R1-2").Set("problem", "a closed problem").Set("location", "§3")),
		ev(seatID, "aaaaaaaa", 2, 1, "close", seatID+":close:R1-2", NewPayload().
			Set("gap_id", "R1-2").Set("closure_class", "closed_with_regression").
			Set("successor", "R2-1").
			Set("anchor_seat", "L1").Set("anchor_tool", "git show").Set("anchor_target", "7bc501e:f")),
		// An OPEN gap with neither --class nor --check, so the "undefined"
		// interpolation is observable in the section that renders it.
		ev(seatID, "aaaaaaaa", 3, 1, "mint", seatID+":mint:R1-3", NewPayload().
			Set("gap_id", "R1-3").Set("problem", "an unclassed problem").Set("location", "§4")),
	})
	res, err := Render(runDir, "")
	if err != nil {
		t.Fatal(err)
	}
	if res.Open != 2 || res.Closed != 1 {
		t.Errorf("counts = %d open / %d closed, want 2/1", res.Open, res.Closed)
	}
	ledger := readFile(t, filepath.Join(res.Out, "ledger.md"))
	if !strings.Contains(ledger, "## OPEN GAPS (2)") {
		t.Errorf("open-gap heading missing:\n%s", ledger)
	}
	if !strings.Contains(ledger, "### R1-1 — an open problem") {
		t.Errorf("the open gap is not in the ledger:\n%s", ledger)
	}
	if strings.Contains(ledger, "### R1-2") {
		t.Error("a CLOSED gap appears in the open section")
	}
	// The `${undefined}` contract, verbatim.
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
	// The closure index carries the class and the successor.
	if !strings.Contains(ledger, "R1-2 | closed_with_regression | a closed problem | R2-1") {
		t.Errorf("closure index row is wrong:\n%s", ledger)
	}

	archive := readFile(t, filepath.Join(res.Out, "archive.md"))
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

// A carried closure says so instead of naming an anchor it does not have.
func TestRenderArchiveShowsCarriedClosures(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r2"
	writeShard(t, runDir, seatID, "aaaaaaaa", []Event{
		ev(seatID, "aaaaaaaa", 0, 2, "mint", seatID+":mint:R2-1", NewPayload().Set("gap_id", "R2-1").Set("problem", "p")),
		ev(seatID, "aaaaaaaa", 1, 2, "close", seatID+":close:R2-1", NewPayload().
			Set("gap_id", "R2-1").Set("carried_from", "1")),
	})
	res, err := Render(runDir, "")
	if err != nil {
		t.Fatal(err)
	}
	archive := readFile(t, filepath.Join(res.Out, "archive.md"))
	if !strings.Contains(archive, "verification anchor: CARRIED from round 1") {
		t.Errorf("a carried closure must say so:\n%s", archive)
	}
	if strings.Contains(archive, "undefined | undefined") {
		t.Errorf("a carried closure printed an empty anchor triple:\n%s", archive)
	}
}

// Anomalies are NEVER silently normalized: they get their own render footer.
func TestRenderSurfacesAnomaliesAndUndisposedObservations(t *testing.T) {
	runDir := t.TempDir()
	lens := "red-lens-r1-L1"
	// Two shards for one seat: a double dispatch, which is an anomaly.
	writeShard(t, runDir, lens, "aaaaaaaa", []Event{
		ev(lens, "aaaaaaaa", 0, 1, "finding", lens+":finding:F1", NewPayload().
			Set("label", "F1").Set("text", strings.Repeat("long prose ", 40))),
	})
	writeShard(t, runDir, lens, "bbbbbbbb", []Event{
		ev(lens, "bbbbbbbb", 0, 1, "finding", lens+":finding:F2", NewPayload().Set("label", "F2").Set("text", "short")),
	})
	res, err := Render(runDir, "")
	if err != nil {
		t.Fatal(err)
	}
	if res.Anomalies == 0 {
		t.Fatal("a double dispatch produced no anomaly count")
	}
	ledger := readFile(t, filepath.Join(res.Out, "ledger.md"))
	if !strings.Contains(ledger, "## render anomalies (never silently normalized)") {
		t.Errorf("the anomaly footer is missing:\n%s", ledger)
	}
	if !strings.Contains(ledger, "multi-nonce seat "+lens) {
		t.Errorf("the anomaly does not name the seat:\n%s", ledger)
	}
	// Every observation demands a disposition, so undisposed ones are listed.
	if !strings.Contains(ledger, "## undisposed lens observations") {
		t.Errorf("the undisposed footer is missing:\n%s", ledger)
	}
	// The listed prose is truncated to 120 units — the footer is an index, not a dump.
	for _, line := range strings.Split(ledger, "\n") {
		if strings.HasPrefix(line, "- "+lens+" F") {
			prose := line[strings.Index(line, ": ")+2:]
			if len(utf16.Encode([]rune(prose))) > 120 {
				t.Errorf("undisposed prose was not truncated: %d units", len(utf16.Encode([]rune(prose))))
			}
		}
	}
}

// Telemetry is COMPUTED, never self-reported, and its key order is insertion
// order so the line is byte-stable.
func TestRenderTelemetryIsComputed(t *testing.T) {
	runDir := t.TempDir()
	r1 := "red-merge-r1"
	r2 := "red-merge-r2"
	writeShard(t, runDir, r1, "aaaaaaaa", []Event{
		ev(r1, "aaaaaaaa", 0, 1, "mint", r1+":mint:R1-1", NewPayload().
			Set("gap_id", "R1-1").Set("problem", "p1").
			Set("severity", "high").Set("likelihood", "medium").Set("impact", "high")),
		ev(r1, "aaaaaaaa", 1, 1, "mint", r1+":mint:R1-2", NewPayload().
			Set("gap_id", "R1-2").Set("problem", "p2").
			Set("severity", "low").Set("likelihood", "low").Set("impact", "low")),
	})
	writeShard(t, runDir, r2, "bbbbbbbb", []Event{
		// A round-2 mint that supersedes a round-1 gap: lineage, and a mass DROP.
		ev(r2, "bbbbbbbb", 0, 2, "mint", r2+":mint:R2-1", NewPayload().
			Set("gap_id", "R2-1").Set("problem", "p3").
			Set("severity", "low").Set("likelihood", "low").Set("impact", "low").
			Set("supersedes", []string{"R1-1"})),
		ev(r2, "bbbbbbbb", 1, 2, "close", r2+":close:R1-1", NewPayload().Set("gap_id", "R1-1")),
	})
	res, err := Render(runDir, "")
	if err != nil {
		t.Fatal(err)
	}
	raw := readFile(t, filepath.Join(res.Out, "board-telemetry.jsonl"))
	lines := strings.Split(strings.TrimRight(raw, "\n"), "\n")
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
	// max_severity is the highest-mass grade among open gaps.
	if got := r1line["max_severity"]; got != "high" {
		t.Errorf("round 1 max_severity = %v, want high", got)
	}
	// mass is computed from likelihood x impact: (2*3) + (1*1).
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
	// R1-1 was likelihood medium x impact high (6); R2-1 is low x low (1). The
	// lineage therefore records 5 units of DOWN mass and none up.
	ed := r2line["edge_deltas"].(map[string]any)
	if ed["down_mass"] != float64(5) || ed["up_mass"] != float64(0) {
		t.Errorf("round 2 edge_deltas = %v, want down 5 / up 0", ed)
	}
}

// A gap minted without a severity keys by_severity under the literal string
// "undefined", as `bySev[g.severity]` does in JS.
func TestRenderTelemetryUndefinedSeverityKey(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	writeShard(t, runDir, seatID, "aaaaaaaa", []Event{
		ev(seatID, "aaaaaaaa", 0, 1, "mint", seatID+":mint:R1-1", NewPayload().Set("gap_id", "R1-1").Set("problem", "p")),
	})
	res, err := Render(runDir, "")
	if err != nil {
		t.Fatal(err)
	}
	raw := readFile(t, filepath.Join(res.Out, "board-telemetry.jsonl"))
	if !strings.Contains(raw, `"by_severity":{"undefined":1}`) {
		t.Errorf("an ungraded mint must key by_severity under \"undefined\":\n%s", raw)
	}
	// max_severity is `[...].sort()[0] || null`: a falsy top grade becomes null,
	// never the string "undefined".
	if !strings.Contains(raw, `"max_severity":null`) {
		t.Errorf("max_severity for an ungraded board must be null:\n%s", raw)
	}
	// ratio is null when nothing closed: a division by zero would be Infinity.
	if !strings.Contains(raw, `"ratio":null`) {
		t.Errorf("ratio with no closures must be null:\n%s", raw)
	}
}

func TestRenderDebateChangelogInquiryAndCitations(t *testing.T) {
	runDir := t.TempDir()
	merge := "red-merge-r1"
	blue := "blue-lane-1"
	judge := "judge-r1"
	lens := "red-lens-r1-L1"
	writeShard(t, runDir, merge, "aaaaaaaa", []Event{
		ev(merge, "aaaaaaaa", 0, 1, "position", merge+":position", NewPayload().Set("text", "red says so")),
		ev(merge, "aaaaaaaa", 1, 1, "closing", merge+":closing:R1-1", NewPayload().Set("gap_id", "R1-1").Set("text", "red closes")),
	})
	writeShard(t, runDir, blue, "bbbbbbbb", []Event{
		ev(blue, "bbbbbbbb", 0, 1, "position", blue+":position", NewPayload().Set("text", "blue says otherwise")),
		ev(blue, "bbbbbbbb", 1, 1, "revision", blue+":revision", NewPayload().Set("text", "blue revised")),
		ev(blue, "bbbbbbbb", 2, 1, "avenue", blue+":avenue:a1", NewPayload().
			Set("label", "a1").Set("status", "abandoned").Set("line", "try the archive").
			Set("method", "full-text search").Set("reason", "the archive is offline")),
		ev(blue, "bbbbbbbb", 3, 1, "avenue", blue+":avenue:a2", NewPayload().
			Set("label", "a2").Set("status", "pursued").Set("line", "read the source")),
	})
	writeShard(t, runDir, judge, "cccccccc", []Event{
		ev(judge, "cccccccc", 0, 1, "opinion", judge+":opinion:R1-1", NewPayload().
			Set("gap_id", "R1-1").Set("disposition", "upheld").Set("principle", "correctness first").
			Set("tension", "economy").Set("review_flag", "none").Set("rationale", "because")),
	})
	writeShard(t, runDir, lens, "dddddddd", []Event{
		ev(lens, "dddddddd", 0, 1, "cite", lens+":cite:https://x", NewPayload().
			Set("claim", "the claim").Set("reference", "https://x").
			Set("confidence", "high").Set("access_date", "2026-07-18")),
		// A citation missing its optional columns renders them as "undefined".
		ev(lens, "dddddddd", 1, 1, "cite", lens+":cite:https://y", NewPayload().Set("reference", "https://y")),
	})

	res, err := Render(runDir, "")
	if err != nil {
		t.Fatal(err)
	}

	debate := readFile(t, filepath.Join(res.Out, "debate.md"))
	for _, want := range []string{"## Round 1", "### RED\nred says so", "### BLUE\nblue says otherwise",
		"### RED CLOSING (round 1) — R1-1", "### LEAD", "upheld — principle: correctness first"} {
		if !strings.Contains(debate, want) {
			t.Errorf("debate.md is missing %q:\n%s", want, debate)
		}
	}

	// The CHANGELOG projection renders the revision's prose; claim_count no longer
	// rides on a revision event (#70 moved it to the count-claims command).
	changelog := readFile(t, filepath.Join(res.Out, "CHANGELOG.md"))
	if !strings.Contains(changelog, "## Round 1\nblue revised") {
		t.Errorf("CHANGELOG.md is wrong:\n%s", changelog)
	}

	inquiry := readFile(t, filepath.Join(res.Out, "lines-of-inquiry.md"))
	// Grouped by FATE, and pursued comes before abandoned.
	pursuedAt := strings.Index(inquiry, "## pursued (1)")
	abandonedAt := strings.Index(inquiry, "## abandoned (1)")
	if pursuedAt < 0 || abandonedAt < 0 || pursuedAt > abandonedAt {
		t.Errorf("lines-of-inquiry sections are wrong or out of order:\n%s", inquiry)
	}
	if !strings.Contains(inquiry, "- **try the archive** _(full-text search)_ — the archive is offline (blue-lane-1)") {
		t.Errorf("an abandoned avenue row is wrong:\n%s", inquiry)
	}
	// A status with no entries gets no empty heading.
	if strings.Contains(inquiry, "## declined") {
		t.Errorf("an empty status produced a heading:\n%s", inquiry)
	}

	cites := readFile(t, filepath.Join(res.Out, "citation-ledger.md"))
	if !strings.Contains(cites, "the claim | https://x | high | r1 | 2026-07-18") {
		t.Errorf("citation row is wrong:\n%s", cites)
	}
	if !strings.Contains(cites, "undefined | https://y | undefined | r1 | undefined") {
		t.Errorf("a sparse citation must render its gaps as \"undefined\":\n%s", cites)
	}
}

// A round with no debate content produces no heading: an empty "## Round 3"
// would claim a round happened that nobody argued.
func TestRenderDebateSkipsEmptyRounds(t *testing.T) {
	runDir := t.TempDir()
	lens := "red-lens-r3-L1"
	writeShard(t, runDir, lens, "aaaaaaaa", []Event{
		ev(lens, "aaaaaaaa", 0, 3, "friction", lens+":friction:#1", NewPayload().Set("text", "not debate content")),
	})
	res, err := Render(runDir, "")
	if err != nil {
		t.Fatal(err)
	}
	debate := readFile(t, filepath.Join(res.Out, "debate.md"))
	if strings.Contains(debate, "## Round 3") {
		t.Errorf("a round with no positions, closings or opinions got a heading:\n%s", debate)
	}
}

// Renders are FULL STATE, so re-rendering the same log is byte-identical: the
// projection never accumulates.
func TestRenderIsDeterministicAndIdempotent(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	writeShard(t, runDir, seatID, "aaaaaaaa", []Event{
		ev(seatID, "aaaaaaaa", 0, 1, "mint", seatID+":mint:R1-1", NewPayload().
			Set("gap_id", "R1-1").Set("problem", "p <with> & entities").
			Set("severity", "high").Set("likelihood", "medium").Set("impact", "high")),
		ev(seatID, "aaaaaaaa", 1, 1, "mint", seatID+":mint:R1-2", NewPayload().
			Set("gap_id", "R1-2").Set("problem", "q").Set("severity", "low")),
	})
	names := []string{"ledger.md", "archive.md", "board-telemetry.jsonl", "debate.md", "CHANGELOG.md", "lines-of-inquiry.md", "citation-ledger.md"}

	first := map[string]string{}
	res, err := Render(runDir, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range names {
		first[n] = readFile(t, filepath.Join(res.Out, n))
	}
	for i := 0; i < 3; i++ {
		if _, err := Render(runDir, ""); err != nil {
			t.Fatal(err)
		}
		for _, n := range names {
			if got := readFile(t, filepath.Join(res.Out, n)); got != first[n] {
				t.Errorf("%s changed on re-render %d — the projection is not full state", n, i+1)
			}
		}
	}
	// Prose entities reach the projection unescaped: a markdown file is not HTML.
	if !strings.Contains(first["ledger.md"], "p <with> & entities") {
		t.Errorf("prose was escaped in the ledger:\n%s", first["ledger.md"])
	}
}

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(string(b), "\r\n", "\n")
}
