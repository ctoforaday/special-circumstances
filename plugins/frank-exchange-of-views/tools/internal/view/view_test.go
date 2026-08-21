package view

import (
	"encoding/json"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
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
	// THE REAL SET, not a copy of it. This loop used to carry its own list, which is how two
	// renderers kept being exercised after their last caller went away.
	for _, name := range MarkdownViews() {
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
		recordtest.At(t, seatID, "aaaaaaaa", 0, 1, seatID+":mint:R1-1", &recordpb.Mint{Problem: proto.String("an open problem"), Location: proto.String("§2"), RequiredFix: proto.String("do the thing"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH)}),
		recordtest.At(t, seatID, "aaaaaaaa", 1, 1, seatID+":mint:R1-2", &recordpb.Mint{Problem: proto.String("a closed problem"), Location: proto.String("§3")}),
		recordtest.At(t, seatID, "aaaaaaaa", 2, 1, seatID+":close:R1-2", &recordpb.Close{ClosureClass: recordtest.P(recordpb.ClosureClass_CLOSURE_CLASS_CLOSED_WITH_REGRESSION), AnchorTool: proto.String("git show"), AnchorTarget: proto.String("7bc501e:f")}),
		recordtest.At(t, seatID, "aaaaaaaa", 3, 1, seatID+":mint:R1-3", &recordpb.Mint{Problem: proto.String("an unclassed problem"), Location: proto.String("§4")}),
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
		recordtest.At(t, seatID, "aaaaaaaa", 0, 2, seatID+":mint:R2-1", &recordpb.Mint{GapId: proto.String("R2-1"), Problem: proto.String("p")}),
		recordtest.At(t, seatID, "aaaaaaaa", 1, 2, seatID+":close:R2-1", &recordpb.Close{CarriedFrom: proto.String("1")}),
	})
	archive := md(t, runDir, "archive")
	if !strings.Contains(archive, "verification anchor: CARRIED from round 1") {
		t.Errorf("a carried closure must say so:\n%s", archive)
	}
	if strings.Contains(archive, "undefined | undefined") {
		t.Errorf("a carried closure printed an empty anchor triple:\n%s", archive)
	}
}

func TestMarkdownSurfacesAnomaliesAndUncreditedFindings(t *testing.T) {
	runDir := t.TempDir()
	lens := "red-lens-r1-L1"
	writeShard(t, runDir, lens, "aaaaaaaa", []record.Event{
		recordtest.At(t, lens, "aaaaaaaa", 0, 1, lens+":finding:F1", &recordpb.Finding{Text: proto.String(strings.Repeat("long prose ", 40))}),
	})
	writeShard(t, runDir, lens, "bbbbbbbb", []record.Event{
		recordtest.At(t, lens, "bbbbbbbb", 0, 1, lens+":finding:F2", &recordpb.Finding{Label: proto.String("F2"), Text: proto.String("short")}),
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
	writeShard(t, runDir, r1, "aaaaaaaa", []record.Event{
		recordtest.At(t, r1, "aaaaaaaa", 0, 1, r1+":mint:R1-1", &recordpb.Mint{Problem: proto.String("p1"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH)}),
		recordtest.At(t, r1, "aaaaaaaa", 1, 1, r1+":mint:R1-2", &recordpb.Mint{Problem: proto.String("p2"), Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Impact: recordtest.P(recordpb.Grade_GRADE_LOW)}),
	})
	writeShard(t, runDir, r2, "bbbbbbbb", []record.Event{
		recordtest.At(t, r2, "bbbbbbbb", 0, 2, r2+":mint:R2-1", &recordpb.Mint{Problem: proto.String("p3"), Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Impact: recordtest.P(recordpb.Grade_GRADE_LOW)}),
		recordtest.At(t, r2, "bbbbbbbb", 1, 2, r2+":close:R1-1", &recordpb.Close{GapId: proto.String("R1-1")}),
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
	mint := func(seat, nonce, id, class string, i, round int) record.Event {
		return recordtest.At(t, seat, nonce, i, round, seat+":mint:"+id, &recordpb.Mint{Problem: proto.String("p"), Class: proto.String(class), Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Impact: recordtest.P(recordpb.Grade_GRADE_LOW)})
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
		recordtest.At(t, seat, "aaaaaaaa", 0, 1, seat+":mint:R1-1", &recordpb.Mint{Problem: proto.String("p"), Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Impact: recordtest.P(recordpb.Grade_GRADE_LOW)}),
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
		recordtest.At(t, seatID, "aaaaaaaa", 0, 1, seatID+":mint:R1-1", &recordpb.Mint{GapId: proto.String("R1-1"), Problem: proto.String("p")}),
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

func TestMarkdownDebateAndInquiry(t *testing.T) {
	runDir := t.TempDir()
	merge := "red-merge-r1"
	blue := "blue-lane-1"
	judge := "judge-r1"
	lens := "red-lens-r1-L1"
	writeShard(t, runDir, merge, "aaaaaaaa", []record.Event{
		recordtest.At(t, merge, "aaaaaaaa", 0, 1, merge+":position", &recordpb.Position{Text: proto.String("red says so")}),
		recordtest.At(t, merge, "aaaaaaaa", 1, 1, merge+":closing:R1-1", &recordpb.Closing{GapId: proto.String("R1-1"), Text: proto.String("red closes")}),
	})
	writeShard(t, runDir, blue, "bbbbbbbb", []record.Event{
		recordtest.At(t, blue, "bbbbbbbb", 0, 1, blue+":position", &recordpb.Position{Text: proto.String("blue says otherwise")}),
		recordtest.At(t, blue, "bbbbbbbb", 1, 1, blue+":revision", &recordpb.Revision{Text: proto.String("blue revised")}),
		ev(blue, "bbbbbbbb", 2, 1, "line-of-inquiry", blue+":line-of-inquiry:q1", record.NewPayload().
			Set("inquiry_id", "Q1").Set("status", "abandoned").Set("line", "try the archive").
			Set("method", "full-text search").Set("reason", "the archive is offline")),
		ev(blue, "bbbbbbbb", 3, 1, "line-of-inquiry", blue+":line of inquiry:a2", record.NewPayload().
			Set("inquiry_id", "Q2").Set("status", "pursued").Set("line", "read the source")),
	})
	writeShard(t, runDir, judge, "cccccccc", []record.Event{
		recordtest.At(t, judge, "cccccccc", 0, 1, judge+":opinion:R1-1", &recordpb.Opinion{Disposition: proto.String("upheld"), Principle: proto.String("correctness first"), ReviewFlag: proto.String("none"), Rationale: proto.String("because")}),
	})
	writeShard(t, runDir, lens, "dddddddd", []record.Event{
		recordtest.At(t, lens, "dddddddd", 0, 1, lens+":verify:https://x", &recordpb.Verify{Anchor: proto.String("c-abc"), Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_MEDIUM), AccessDate: proto.String("2026-07-18")}),
		recordtest.At(t, lens, "dddddddd", 1, 1, lens+":verify:https://y", &recordpb.Verify{}),
	})

	debate := md(t, runDir, "debate")
	for _, want := range []string{"## Round 1", "### RED\nred says so", "### BLUE\nblue says otherwise",
		"### RED CLOSING (round 1) — R1-1", "### LEAD", "upheld — principle: correctness first"} {
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
	lens := "red-lens-r3-L1"
	writeShard(t, runDir, lens, "aaaaaaaa", []record.Event{
		recordtest.At(t, lens, "aaaaaaaa", 0, 3, lens+":friction:#1", &recordpb.Friction{Text: proto.String("not debate content")}),
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
		recordtest.At(t, seatID, "aaaaaaaa", 0, 1, seatID+":mint:R1-1", &recordpb.Mint{Problem: proto.String("p <with> & entities"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH)}),
		recordtest.At(t, seatID, "aaaaaaaa", 1, 1, seatID+":mint:R1-2", &recordpb.Mint{Problem: proto.String("q"), Severity: recordtest.P(recordpb.Grade_GRADE_LOW)}),
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
