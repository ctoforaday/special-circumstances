package view

import (
	"encoding/json"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"google.golang.org/protobuf/proto"
	"math"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// ---- fixtures ----

// writeShard seeds a run's record. The name is kept because 17 call sites read it as "put these
// events in this run", which is still exactly what it does.
//
// It WROTE A SHARD FILE — `events-<seat>-<nonce>.jsonl`, hand-marshalled through
// record.MarshalEvent. There are no shard files and no nonces: a fixture that still wrote one
// would leave the run's record EMPTY, and every assertion below would keep passing against a board
// that was never built. The seat and nonce parameters are gone for the same reason the fields are.
func writeShard(t *testing.T, runDir string, evs []*record.Event) {
	t.Helper()
	recordtest.Seed(t, runDir, evs...)
}

// md is view.Markdown or a fatal.
func md(t *testing.T, runDir, name string) string {
	t.Helper()
	b, err := Markdown(runtest.Open(t, runDir), name, "")
	if err != nil {
		t.Fatalf("Markdown(runtest.Open(t, %q)): %v", name, err)
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

// ABSENT AND EMPTY ARE DIFFERENT ANSWERS, and the distinction moved from a payload KEY to field
// PRESENCE — which is the whole reason every field on the schema is `optional`. A seat that said
// nothing and a seat that said "" are not the same fact, and a projection that renders both blank
// erases the difference in the artifact a reader trusts.
func TestUndefStrDistinguishesAbsentFromEmpty(t *testing.T) {
	if got := undefStr(nil); got != "undefined" {
		t.Errorf("an absent value renders %q, want \"undefined\"", got)
	}
	if got := undefStr(proto.String("")); got != "" {
		t.Errorf("a present-but-empty value renders %q, want empty", got)
	}
	// The explicit-null case is gone with the payload: a proto field is set or it is not, and
	// "set to null" is not a third state the record can hold. Absence covers it, which the nil
	// case above already asserts.
	if got := undefStr(proto.String("v")); got != "v" {
		t.Errorf("a present value renders %q", got)
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
	// UNGRADED IS THE ENUM'S ZERO NOW, not a nil or a stray non-string. The old fixture needed
	// three shapes to express "no grade" — nil, and a bool that had wandered in — because a
	// payload could hold anything. The schema admits exactly one: UNSPECIFIED, which is what an
	// absent grade decodes to, and mass must skip it rather than score it as the lowest grade.
	gaps := []*record.Gap{
		{Likelihood: recordpb.Grade_GRADE_MEDIUM, Impact: recordpb.Grade_GRADE_HIGH},
		{Likelihood: recordpb.Grade_GRADE_LOW, Impact: recordpb.Grade_GRADE_LOW},
		{Likelihood: recordpb.Grade_GRADE_UNSPECIFIED, Impact: recordpb.Grade_GRADE_HIGH},
		// `realized` scores ZERO BY DESIGN: mass forecasts what is still to come, and a realized
		// defect is measured by its damage instead. It is a grade, not a missing entry.
		{Likelihood: recordpb.Grade_GRADE_REALIZED, Impact: recordpb.Grade_GRADE_HIGH},
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
	open, closed, err := Counts(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if open != 0 || closed != 0 {
		t.Errorf("empty run reported open=%d closed=%d", open, closed)
	}
	// THE REAL SET, not a copy of it. This loop used to carry its own list, which is how two
	// renderers kept being exercised after their last caller went away.
	for _, name := range MarkdownViews() {
		if _, err := Markdown(runtest.Open(t, runDir), name, ""); err != nil {
			t.Errorf("projection %s errored on an empty run: %v", name, err)
		}
	}
	// Telemetry with no rounds is EMPTY, not a blank line.
	b, err := TelemetryJSONL(runtest.Open(t, runDir))
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
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, seatID, 1, seatID+":mint:R1-1", &recordpb.Mint{GapId: proto.String("R1-1"), Class: proto.String("overclaim"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Problem: proto.String("an open problem"), Location: proto.String("§2"), RequiredFix: proto.String("do the thing"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH)}),
		recordtest.At(t, seatID, 1, seatID+":mint:R1-2", &recordpb.Mint{GapId: proto.String("R1-2"), Class: proto.String("overclaim"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("a closed problem"), Location: proto.String("§3")}),
		recordtest.At(t, seatID, 1, seatID+":mint:R1-3", &recordpb.Mint{GapId: proto.String("R1-3"), Class: proto.String("overclaim"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("an unclassed problem"), Location: proto.String("§4")}),
		recordtest.At(t, seatID, 1, seatID+":close:R1-2", &recordpb.Close{GapId: proto.String("R1-2"), Prose: proto.String("verified at the leaf"), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED_WITH_REGRESSION), Successor: proto.String("R1-3"), AnchorSeat: proto.String("L1"), AnchorTool: proto.String("git show"), AnchorTarget: proto.String("7bc501e:f")}),
	})
	open, closed, err := Counts(runtest.Open(t, runDir))
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
	// `class` and `acceptance_check` USED TO BE ASSERTED HERE as "undefined", for a gap minted
	// without them. Both are required now — the write path refuses a mint that omits either and
	// the columns are NOT NULL — so a gap in that state cannot reach a projection. What is still
	// worth asserting is that a class the seat DID give is rendered, and `cx` above covers the
	// undefined arm with a field that is genuinely optional.
	if !strings.Contains(ledger, "class overclaim") {
		t.Errorf("a gap's class must be rendered:\n%s", ledger)
	}
	if !strings.Contains(ledger, "R1-2 | repaired_with_regression | a closed problem | R1-3") {
		t.Errorf("closure index row is wrong:\n%s", ledger)
	}

	archive := md(t, runDir, "archive")
	if !strings.Contains(archive, "verification anchor: L1 | git show | 7bc501e:f") {
		t.Errorf("the anchor triple is not in the archive:\n%s", archive)
	}
	if !strings.Contains(archive, "successor: R1-3") {
		t.Errorf("the successor is not in the archive:\n%s", archive)
	}
	if strings.Contains(archive, "R1-1") {
		t.Error("an OPEN gap appears in the archive")
	}
}

func TestMarkdownArchiveShowsCarriedClosures(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r2"
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, seatID, 2, seatID+":mint:R2-1", &recordpb.Mint{Class: proto.String("overclaim"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String("R2-1"), Problem: proto.String("p")}),
		recordtest.At(t, seatID, 2, seatID+":close:R2-1", &recordpb.Close{GapId: proto.String("R2-1"), Prose: proto.String("verified at the leaf"), CarriedFrom: proto.String("1")}),
	})
	archive := md(t, runDir, "archive")
	if !strings.Contains(archive, "verification anchor: CARRIED from round 1") {
		t.Errorf("a carried closure must say so:\n%s", archive)
	}
	if strings.Contains(archive, "undefined | undefined") {
		t.Errorf("a carried closure printed an empty anchor triple:\n%s", archive)
	}
}

// THE ANOMALY HALF OF THIS TEST IS GONE; the uncredited-findings half is what it was really for.
//
// It asserted a "multi-nonce seat" anomaly and a `## render anomalies` footer, produced when one
// seat had two shard files. There are no shards and no nonce: both dispatches are rows, nothing
// selects a winner, and nothing is dropped — so the anomaly it looked for cannot occur and the
// footer that printed it is deleted rather than left permanently empty.
//
// What survives is the question that still has an answer: does the ledger name lens findings that
// no gap credits? That footer is a live work list, and it is the half that was ever load-bearing.
func TestMarkdownSurfacesUncreditedFindings(t *testing.T) {
	runDir := t.TempDir()
	lens := "red-lens-r1-evidence"
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, lens, 1, lens+":finding:F1", &recordpb.Finding{Text: proto.String(strings.Repeat("long prose ", 40))}),
		recordtest.At(t, lens, 1, lens+":finding:F2", &recordpb.Finding{Label: proto.String("F2"), Text: proto.String("short")}),
	})
	ledger := md(t, runDir, "ledger")
	if !strings.Contains(ledger, "## lens findings credited by no gap") {
		t.Errorf("the undisposed footer is missing:\n%s", ledger)
	}
	for _, line := range strings.Split(ledger, "\n") {
		if strings.HasPrefix(line, "- "+lens+" F") {
			prose := line[strings.Index(line, ": ")+2:]
			if len(utf16.Encode([]rune(prose))) > 120 {
				t.Errorf("uncredited-finding prose was not truncated: %d units", len(utf16.Encode([]rune(prose))))
			}
		}
	}
}

func TestTelemetryIsComputed(t *testing.T) {
	runDir := t.TempDir()
	r1 := "red-merge-r1"
	r2 := "red-merge-r2"
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, r1, 1, r1+":mint:R1-1", &recordpb.Mint{GapId: proto.String("R1-1"), Class: proto.String("overclaim"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Problem: proto.String("p1"), Severity: recordtest.P(recordpb.Grade_GRADE_HIGH), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH)}),
		recordtest.At(t, r1, 1, r1+":mint:R1-2", &recordpb.Mint{GapId: proto.String("R1-2"), Class: proto.String("overclaim"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Problem: proto.String("p2"), Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Impact: recordtest.P(recordpb.Grade_GRADE_LOW)}),
	})
	writeShard(t, runDir, []*record.Event{
		// THE LINEAGE MINT. `supersedes` is what makes a gap a lineage mint — not the closure's
		// successor, which is a different edge — and it is what repair_regression counts. The
		// edge_deltas assertion reads the mass drop across that edge: R1-1 is medium x high (6),
		// this one is low x low (1), so the repair moved 5 points of mass down.
		recordtest.At(t, r2, 2, r2+":mint:R2-1", &recordpb.Mint{
			GapId:           proto.String("R2-1"),
			Class:           proto.String("overclaim"),
			Problem:         proto.String("p3"),
			AcceptanceCheck: proto.String("the check runs"),
			CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Supersedes:      []string{"R1-1"},
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_LOW),
			Impact:          recordtest.P(recordpb.Grade_GRADE_LOW),
		}),
		// A LINEAGE closure: it reports a regression and names the round-2 gap carrying it forward,
		// which is what makes repair_regression's ratio 1 rather than 0. The record refuses this
		// closure class without a successor, so the fixture cannot express the half-state.
		recordtest.At(t, r2, 2, r2+":close:R1-1", &recordpb.Close{
			GapId:        proto.String("R1-1"),
			ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED_WITH_REGRESSION),
			Successor:    proto.String("R2-1"),
			Prose:        proto.String("verified at the leaf"),
		}),
	})
	raw, err := TelemetryJSONL(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("%d telemetry lines, want one per round:\n%s", len(lines), raw)
	}
	// Key order is the contract, not merely the key set.
	if !strings.HasPrefix(lines[0], `{"round":1,"mapping_version":"v2","open_count":`) {
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
	mint := func(seat, id, class string, round int) *record.Event {
		return recordtest.At(t, seat, round, seat+":mint:"+id, &recordpb.Mint{
			GapId:           proto.String(id),
			Class:           proto.String(class),
			Problem:         proto.String("p"),
			AcceptanceCheck: proto.String("the check runs"),
			CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_LOW),
			Impact:          recordtest.P(recordpb.Grade_GRADE_LOW),
		})
	}
	r1, r2 := "red-merge-r1", "red-merge-r2"
	writeShard(t, runDir, []*record.Event{
		mint(r1, "R1-1", "evidence-claim-not-documented", 1),
		mint(r1, "R1-2", "evidence-claim-not-documented", 1),
		mint(r1, "R1-3", "scope-closure-missing", 1),
	})
	// Round 2: one class REPEATS from round 1, one is fresh. Repeat rate = 1/2.
	writeShard(t, runDir, []*record.Event{
		mint(r2, "R2-1", "scope-closure-missing", 2),
		mint(r2, "R2-2", "co-resident-rules-disagree", 2),
	})

	raw, err := TelemetryJSONL(runtest.Open(t, runDir))
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

// THE CLASSLESS-MINT TEST IS GONE, because a classless mint is no longer writable.
//
// It asserted that a mint with an empty `class` did not become a bucket in by_class — an empty
// string keying a distribution is a category called "", which reads as a real class to anyone
// looking at the chart. That was a live hazard while class was an optional payload key.
//
// `Mint.class` is REQUIRED now: the write path refuses it and the column is NOT NULL, so the state
// this guarded cannot be reached. Keeping the test would mean seeding a mint the record rejects,
// which is not a test of the projection at all.

func TestTelemetryUndefinedSeverityKey(t *testing.T) {
	runDir := t.TempDir()
	seatID := "red-merge-r1"
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, seatID, 1, seatID+":mint:R1-1", &recordpb.Mint{Class: proto.String("overclaim"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), GapId: proto.String("R1-1"), Problem: proto.String("p")}),
	})
	raw, err := TelemetryJSONL(runtest.Open(t, runDir))
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

func TestMarkdownDebateAndInquiry(t *testing.T) {
	runDir := t.TempDir()
	merge := "red-merge-r1"
	blue := "blue-lane-1"
	judge := "judge-r1"
	lens := "red-lens-r1-evidence"
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, merge, 1, merge+":position", &recordpb.Position{Text: proto.String("red says so")}),
		// The gap has to EXIST before anything speaks about it — the closing statement and the
		// bench opinion below both reference it, and both are foreign keys onto the mint.
		recordtest.At(t, merge, 1, merge+":mint:R1-1", &recordpb.Mint{
			GapId:           proto.String("R1-1"),
			Class:           proto.String("overclaim"),
			Problem:         proto.String("the claim outruns its evidence"),
			AcceptanceCheck: proto.String("the check runs"),
			CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
			Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		}),
		recordtest.At(t, merge, 1, merge+":closing:R1-1", &recordpb.Closing{GapId: proto.String("R1-1"), Text: proto.String("red closes")}),
	})
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, blue, 1, blue+":position", &recordpb.Position{Text: proto.String("blue says otherwise")}),
		recordtest.At(t, blue, 1, blue+":revision", &recordpb.Revision{Text: proto.String("blue revised")}),
		recordtest.At(t, blue, 1, blue+":avenue:Q1", &recordpb.Avenue{
			AvenueId: proto.String("Q1"),
			Line:     proto.String("try the archive"),
			Method:   proto.String("full-text search"),
			Status:   recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_ABANDONED),
			Reason:   proto.String("the archive is offline"),
		}),
		recordtest.At(t, blue, 1, blue+":avenue:Q2", &recordpb.Avenue{
			AvenueId: proto.String("Q2"),
			Line:     proto.String("read the source"),
			Status:   recordtest.P(recordpb.AvenueStatus_AVENUE_STATUS_PURSUED),
		}),
	})
	// THE BENCH'S DISPOSITION IS A DOCKET MOTION'S RULING, and the LEAD row needs both halves:
	// the gap comes from the FILING through record.Motions, the disposition from the RULING.
	// Seeded alone, the ruling renders a LEAD line naming the empty string for its gap.
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, merge, 1, merge+":motion:M1", &recordpb.Motion{
			MotionId: proto.String("M1"), Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Basis:  proto.String("red cannot settle R1-1"),
			Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("R1-1")}},
		}),
		recordtest.At(t, judge, 1, judge+":motion-rule:M1", &recordpb.MotionRule{
			MotionId: proto.String("M1"), Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Opinion: proto.String("because"),
			Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
				Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
				Principle:   proto.String("correctness first"),
				Tension:     proto.String("speed against certainty"),
				ReviewFlag:  proto.String("none"),
				Settled:     proto.String("the claim as it stood may not be re-asserted"),
				Final:       proto.Bool(true),
			}},
		}),
	})
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, lens, 1, lens+":verify:https://x", &recordpb.Verify{Claim: proto.String("the source says so"), Anchor: proto.String("c-abc"), Outcome: recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS), Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_MEDIUM), Text: proto.String("read at the leaf"), AccessDate: proto.String("2026-07-18")}),
		recordtest.At(t, lens, 1, lens+":verify:https://y", &recordpb.Verify{Claim: proto.String("a second claim"), Outcome: recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_WEAK), Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_LOW), Text: proto.String("thin support")}),
	})

	debate := md(t, runDir, "debate")
	for _, want := range []string{"## Round 1", "### RED\nred says so", "### BLUE\nblue says otherwise",
		"### RED CLOSING (round 1) — R1-1", "### LEAD", "repaired — principle: correctness first"} {
		if !strings.Contains(debate, want) {
			t.Errorf("debate is missing %q:\n%s", want, debate)
		}
	}

	inquiry := md(t, runDir, "lines-of-inquiry")
	pursuedAt := strings.Index(inquiry, "## pursued (1)")
	abandonedAt := strings.Index(inquiry, "## abandoned (1)")
	if pursuedAt < 0 || abandonedAt < 0 || pursuedAt > abandonedAt {
		t.Errorf("lines-of-inquiry sections are wrong or out of order:\n%s", inquiry)
	}
	if !strings.Contains(inquiry, "- **Q1 try the archive** _(full-text search)_ — the archive is offline (blue-lane-1)") {
		t.Errorf("an abandoned line of inquiry row is wrong:\n%s", inquiry)
	}
	if strings.Contains(inquiry, "## declined") {
		t.Errorf("an empty status produced a heading:\n%s", inquiry)
	}

	// THE CITATION-LEDGER ASSERTIONS ARE GONE WITH THEIR RENDERER. `citation-ledger` and
	// `changelog` lost their last caller and kept passing here, which is what kept them looking
	// alive — the only place either projection still existed was this test. The evidence layer is
	// asserted where it is now read: record.EvidenceJSONOf, and `show evidence` end to end.
}

func TestMarkdownDebateSkipsEmptyRounds(t *testing.T) {
	runDir := t.TempDir()
	lens := "red-lens-r3-evidence"
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, lens, 3, lens+":friction:#1", &recordpb.Log{Text: proto.String("not debate content"), Type: recordpb.LogType_LOG_TYPE_DEFECT.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}),
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
	writeShard(t, runDir, []*record.Event{
		recordtest.At(t, seatID, 1, seatID+":mint:R1-1", &recordpb.Mint{GapId: proto.String("R1-1"), Class: proto.String("overclaim"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Problem: proto.String("p <with> & entities"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH)}),
		recordtest.At(t, seatID, 1, seatID+":mint:R1-2", &recordpb.Mint{GapId: proto.String("R1-2"), Class: proto.String("overclaim"), AcceptanceCheck: proto.String("the check runs"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("q"), Severity: recordtest.P(recordpb.Grade_GRADE_LOW)}),
	})
	names := MarkdownViews()
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
