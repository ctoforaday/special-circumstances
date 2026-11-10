package cli

import (
	"encoding/json"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// foundSourceRun is a run with a registered lens and a seeded report, for a reading of a source the
// lens found.
func foundSourceRun(t *testing.T) string {
	t.Helper()
	runDir := recordtest.TmpRun(t)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", lensSeat); err != nil {
		t.Fatal(err)
	}
	seedBlueReport(t, runDir)
	return runDir
}

// foundSentence is a sentence of the seeded report no citation or computation marks.
const foundSentence = "§2 the finding prose lands in a quoted sentence."

// assembled runs the composer and returns the report's TEXT. Assemble writes the file and
// returns what it wrote, and its report path reads like markdown to a Contains check and always fails to match.
func assembled(t *testing.T, runDir string) string {
	t.Helper()
	a, err := report.Assemble(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	md, err := os.ReadFile(a.Report)
	if err != nil {
		t.Fatal(err)
	}
	return string(md)
}

// THE READING BOARD: a report of five sentences in two sections, two of them cited by blue — rdCited
// (C1, from rdU1) and rdBacked (C2, from rdU2) — and three no citation or computation marks. Blue,
// the chair and five lenses are seated, each served the report.
const (
	rdCited  = "Costs rose sharply in the first quarter."
	rdFree1  = "Volume recovered later in the year."
	rdBacked = "Margins held steady through the winter."
	rdFree2  = "The rain fell on the plain all spring."
	rdFree3  = "Prices settled once the harvest came in."
	rdReport = "# Findings\n\n" + rdCited + " " + rdFree1 + " " + rdBacked + "\n\n# Notes\n\n" + rdFree2 + " " + rdFree3 + "\n"
	rdU1     = "https://src.example/costs"
	rdU2     = "https://src.example/margins"
	rdU3     = "https://found.example/one"
	rdU4     = "https://found.example/two"
	rdU5     = "https://found.example/three"
	rdU6     = "https://found.example/four"
	rdU7     = "https://src.example/survey"
)

// The lenses a reading board seats: L is lensSeat; M, N, P and K are other lenses.
const (
	lensL = lensSeat
	lensM = "red-lens-logic"
	lensN = "red-lens-adversary"
	lensP = "red-lens-dark-side"
	lensK = "red-lens-voice"
)

type readingBoard struct {
	t              *testing.T
	runDir, c1, c2 string
	// cs is every citation the fixture placed, in the order cited.
	cs []string
	// base is the fixture's own Verify, Mint and Anchor count (blue's cites place anchors), which
	// counts() subtracts.
	base [3]int
}

// readingRunOf builds the reading board over report, citing cited[i] from rdU1, rdU2 and rdU7 in turn.
func readingRunOf(t *testing.T, report string, cited ...string) readingBoard {
	t.Helper()
	return readingBoardOn(t, newRun(t), report, cited...)
}

// readingRunCast is readingRunOf on a run whose cast names every seat the board seats, as setup
// writes it: the chair's dispatch plans against it.
func readingRunCast(t *testing.T, report string, cited ...string) readingBoard {
	t.Helper()
	runDir := newRun(t)
	cast := recordtest.At(t, "harness", "harness:cast", &recordpb.Cast{SeatIds: []string{
		"red-chair", "blue-synthesize", blueSeat, "judge", lensL, lensM, lensN, lensP, lensK}})
	// Stamped as a write path stamps it, so the run migrates as a written one does.
	cast.Ts = proto.String(record.Stamp())
	recordtest.Seed(t, runDir, cast)
	return readingBoardOn(t, runDir, report, cited...)
}

func readingBoardOn(t *testing.T, runDir, report string, cited ...string) readingBoard {
	t.Helper()
	writeReport(t, runDir, report)
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{
		rdU1: []byte("<html>a source on costs</html>"),
		rdU2: []byte("<html>a source on margins</html>"),
		rdU7: []byte("<html>a source on the survey</html>"),
	}})
	urls := []string{rdU1, rdU2, rdU7}
	b := readingBoard{t: t, runDir: runDir}
	for i, q := range cited {
		before := citeTokens(readReport(t, runDir))
		if _, err := run(t, "cite", "--run", runDir, "--seat-id", blueSeat, "--quote", q, "--url", urls[i], "--title", "Source "+string(rune('A'+i))); err != nil {
			t.Fatalf("cite %q: %v", q, err)
		}
		after := citeTokens(readReport(t, runDir))
		var id string
		for _, c := range after {
			if !slices.Contains(before, c) {
				id = c
			}
		}
		b.cs = append(b.cs, id)
	}
	b.c1 = b.cs[0]
	if len(b.cs) > 1 {
		b.c2 = b.cs[1]
	}
	registerChairOnce(t, runDir)
	for _, s := range []string{lensL, lensM, lensN, lensP, lensK} {
		b.sit(s)
	}
	b.base = b.counts()
	return b
}

func readingRun(t *testing.T) readingBoard {
	t.Helper()
	return readingRunOf(t, rdReport, rdCited, rdBacked)
}

func citeTokens(report string) []string {
	var out []string
	for _, m := range regexp.MustCompile(`<!--cite:(C-[0-9a-f]{8})-->`).FindAllStringSubmatch(report, -1) {
		out = append(out, m[1])
	}
	return out
}

// sit opens seat's next sitting and serves it the report, as a lens's audit does.
func (b readingBoard) sit(seat string) {
	b.t.Helper()
	if _, err := run(b.t, "register", "--run", b.runDir, "--seat-id", seat); err != nil {
		b.t.Fatalf("register %s: %v", seat, err)
	}
	if _, err := run(b.t, "show", "report", "--run", b.runDir, "--seat-id", seat); err != nil {
		b.t.Fatalf("serve %s the report: %v", seat, err)
	}
}

// citeReading is a reading of citation anchor with outcome as; class != "" gives the gap flags.
func citeReading(anchor, as, class string, extra ...string) []string {
	a := []string{"--anchor", anchor, "--as", as, "--confidence", "high", "--reason", "read at the leaf"}
	if class != "" {
		a = append(a, gapFlagArgs(class)...)
	}
	return append(a, extra...)
}

// urlReading is a reading of url, a source the lens found, against quote; class != "" gives the gap flags.
func urlReading(url, quote, as, class string, extra ...string) []string {
	a := []string{"--url", url, "--title", "Found " + url[len(url)-3:], "--quote", quote, "--as", as, "--confidence", "high", "--reason", "read at the leaf"}
	if class != "" {
		a = append(a, gapFlagArgs(class)...)
	}
	return append(a, extra...)
}

// readingJSON is verify's --json result, as a test reads it.
type readingJSON struct {
	Anchor             string `json:"anchor"`
	URL                string `json:"url"`
	Outcome            string `json:"outcome"`
	Claim              string `json:"claim"`
	SourceCompleteness string `json:"source_completeness"`
	Gap                *struct {
		ID         string   `json:"id"`
		Class      string   `json:"class"`
		Minted     bool     `json:"minted"`
		Check      string   `json:"acceptance_check"`
		VoiceTells []string `json:"voice_tells"`
	} `json:"gap"`
	FailedChecks []struct {
		State      string `json:"state"`
		Owner      string `json:"owner"`
		Since      int64  `json:"since"`
		AnsweredBy string `json:"answered_by"`
	} `json:"failed_checks"`
}

// verify runs seat's verify with args and returns its human text.
func (b readingBoard) verify(seat string, args ...string) (string, error) {
	b.t.Helper()
	return run(b.t, append([]string{"verify", "--run", b.runDir, "--seat-id", seat}, args...)...)
}

// verifyJSON runs seat's verify --json with args and returns its result.
func (b readingBoard) verifyJSON(seat string, args ...string) (readingJSON, error) {
	b.t.Helper()
	out, err := run(b.t, append([]string{"--json", "verify", "--run", b.runDir, "--seat-id", seat}, args...)...)
	if err != nil {
		return readingJSON{}, err
	}
	var env struct {
		Result readingJSON `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		b.t.Fatalf("verify --json: %v\n%s", err, out)
	}
	return env.Result, nil
}

// mustRead is verifyJSON that fails the test on a refusal.
func (b readingBoard) mustRead(seat string, args ...string) readingJSON {
	b.t.Helper()
	r, err := b.verifyJSON(seat, args...)
	if err != nil {
		b.t.Fatalf("%s verify %q: %v", seat, args, err)
	}
	return r
}

// readingEvent is one stored event of a reading act, with its events.id.
type readingEvent struct {
	ID   int64
	Seat string
	Ev   *recordpb.Event
}

// stored is every event of the board's record whose body is a T, in record order.
func storedOf[T proto.Message](t *testing.T, runDir string) []readingEvent {
	t.Helper()
	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	ids := m.At.IDs(m.Events)
	var out []readingEvent
	for i, e := range m.Events {
		if _, ok := recordpb.BodyAs[T](e); ok {
			out = append(out, readingEvent{ID: ids[i], Seat: e.GetSeatId(), Ev: e})
		}
	}
	return out
}

func verifysOf(t *testing.T, runDir string) []*recordpb.Verify {
	var out []*recordpb.Verify
	for _, e := range storedOf[*recordpb.Verify](t, runDir) {
		v, _ := recordpb.BodyAs[*recordpb.Verify](e.Ev)
		out = append(out, v)
	}
	return out
}

func mintsOf(t *testing.T, runDir string) []*recordpb.Mint {
	var out []*recordpb.Mint
	for _, e := range storedOf[*recordpb.Mint](t, runDir) {
		m, _ := recordpb.BodyAs[*recordpb.Mint](e.Ev)
		out = append(out, m)
	}
	return out
}

// anchorsOfGap is every stored Anchor whose id is gap.
func anchorsOfGap(t *testing.T, runDir, gap string) []*recordpb.Anchor {
	var out []*recordpb.Anchor
	for _, e := range storedOf[*recordpb.Anchor](t, runDir) {
		if a, _ := recordpb.BodyAs[*recordpb.Anchor](e.Ev); a.GetId() == gap {
			out = append(out, a)
		}
	}
	return out
}

// mintOf is the stored Mint of gap, or nil.
func mintOf(t *testing.T, runDir, gap string) *recordpb.Mint {
	for _, m := range mintsOf(t, runDir) {
		if m.GetGapId() == gap {
			return m
		}
	}
	return nil
}

// openOf is how many gaps seat minted that its work list shows open.
func (b readingBoard) openOf(seat string) int {
	b.t.Helper()
	w, err := record.WorkOfSeat(runtest.Open(b.t, b.runDir), "lens", seat)
	if err != nil {
		b.t.Fatal(err)
	}
	n := 0
	for _, g := range w.Open {
		if g.MintedBy == seat {
			n++
		}
	}
	return n
}

// rdFillers are sentences of a third section a test mints plain gaps on.
var rdFillers = []string{
	"Filler sentence one stands here.", "Filler sentence two stands here.", "Filler sentence three stands here.",
	"Filler sentence four stands here.", "Filler sentence five stands here.", "Filler sentence six stands here.",
	"Filler sentence seven stands here.", "Filler sentence eight stands here.",
}

// rdFillerReport is rdReport with rdFillers as a third section.
var rdFillerReport = rdReport + "\n# Filler\n\n" + strings.Join(rdFillers, " ") + "\n"

// plainMint is seat's `mint --quote quote` of class, its other flags given.
func (b readingBoard) plainMint(seat, quote, class string, extra ...string) (string, error) {
	b.t.Helper()
	// A `document` gap unless extra names its check kind.
	if !slices.Contains(extra, "--check-kind") {
		extra = append(extra, "--check-kind", "document")
	}
	out, err := run(b.t, append([]string{"mint", "--run", b.runDir, "--seat-id", seat, "--quote", quote, "--class", class,
		"--problem", "the defect", "--fix", "the fix", "--check", "the acceptance check",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium"}, extra...)...)
	if err != nil {
		return "", err
	}
	return gapID(out), nil
}

// fill plain-mints a gap on every one of rdFillers as seat.
func (b readingBoard) fill(seat string) {
	b.t.Helper()
	for _, s := range rdFillers {
		if _, err := b.plainMint(seat, s, "spec_underspecification"); err != nil {
			b.t.Fatalf("%s's mint on %q: %v", seat, s, err)
		}
	}
}

// counts is how many Verify, Mint and Anchor events the record holds beyond the fixture's.
func (b readingBoard) counts() [3]int {
	return [3]int{
		countType(b.t, b.runDir, recordpb.EventType_EVENT_TYPE_VERIFY) - b.base[0],
		countType(b.t, b.runDir, recordpb.EventType_EVENT_TYPE_MINT) - b.base[1],
		countType(b.t, b.runDir, recordpb.EventType_EVENT_TYPE_ANCHOR) - b.base[2],
	}
}

// lastVerify is the last stored Verify and its events.id.
func (b readingBoard) lastVerify() (*recordpb.Verify, int64) {
	b.t.Helper()
	all := storedOf[*recordpb.Verify](b.t, b.runDir)
	if len(all) == 0 {
		b.t.Fatal("the record holds no reading")
	}
	e := all[len(all)-1]
	v, _ := recordpb.BodyAs[*recordpb.Verify](e.Ev)
	return v, e.ID
}
