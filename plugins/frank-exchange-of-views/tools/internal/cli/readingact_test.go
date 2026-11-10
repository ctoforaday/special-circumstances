package cli

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE READING ACT: a reading that finds its source against the sentence mints the gap in the same act,
// with the judgment flags the seat gives it, and a reading that backs an uncited sentence from a source
// the lens found mints the gap blue answers with a citation.

// mintedBy asserts r minted its gap of class: the stored reading carries it with carrier_minted, the
// Mint is the reader's of that class and always material, and one Anchor of the gap places it as
// place says.
func mintedBy(t *testing.T, b readingBoard, seat string, r readingJSON, class recordpb.GapClass, place func(*recordpb.Anchor) string) string {
	t.Helper()
	if r.Gap == nil || !r.Gap.Minted || r.Gap.ID == "" {
		t.Fatalf("the reading minted no gap: %+v", r)
	}
	g := r.Gap.ID
	v, _ := b.lastVerify()
	if v.GetCarriedBy() != g || !v.GetCarrierMinted() {
		t.Errorf("the stored reading carries %q (carrier_minted %v), want %s minted in its act", v.GetCarriedBy(), v.GetCarrierMinted(), g)
	}
	evs := storedOf[*recordpb.Mint](t, b.runDir)
	var m *recordpb.Mint
	for _, e := range evs {
		if x, _ := recordpb.BodyAs[*recordpb.Mint](e.Ev); x.GetGapId() == g {
			m = x
			if e.Seat != seat {
				t.Errorf("gap %s was minted by %s, want the reader %s", g, e.Seat, seat)
			}
		}
	}
	if m == nil {
		t.Fatalf("no Mint of %s is on the record", g)
	}
	if m.GetClass() != class {
		t.Errorf("gap %s has class %v, want %v", g, m.GetClass(), class)
	}
	if m.GetClassMaterial() != recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS && class != recordpb.GapClass_GAP_CLASS_UNSOURCED_CLAIM {
		t.Errorf("gap %s has class_material %v, want always", g, m.GetClassMaterial())
	}
	as := anchorsOfGap(t, b.runDir, g)
	if len(as) != 1 {
		t.Fatalf("%d Anchors place %s, want one", len(as), g)
	}
	if why := place(as[0]); why != "" {
		t.Errorf("gap %s's Anchor %v: %s", g, as[0], why)
	}
	return g
}

func bySentenceOf(c string) func(*recordpb.Anchor) string {
	return func(a *recordpb.Anchor) string {
		if a.GetSentenceOf() != c || a.Location != nil {
			return "want sentence_of " + c + " and no location"
		}
		return ""
	}
}

func byLocation(s string) func(*recordpb.Anchor) string {
	return func(a *recordpb.Anchor) string {
		if a.GetLocation() != s || a.SentenceOf != nil {
			return fmt.Sprintf("want location %q and no sentence_of", s)
		}
		return ""
	}
}

// 15. A reading that finds against its sentence mints the gap in its own act, in both forms.
func TestARefutingReadingMintsItsGapInTheSameAct(t *testing.T) {
	b := readingRun(t)
	cases := []struct {
		name  string
		seat  string
		args  []string
		class recordpb.GapClass
		place func(*recordpb.Anchor) string
	}{
		{"verify --anchor --as refutes", lensL, citeReading(b.c1, "refutes", "claim_contradicts_source"),
			recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE, bySentenceOf(b.c1)},
		{"verify --anchor --as absent", lensM, citeReading(b.c2, "absent", "citation_figure_misattribution"),
			recordpb.GapClass_GAP_CLASS_CITATION_FIGURE_MISATTRIBUTION, bySentenceOf(b.c2)},
		{"verify --url --as refutes", lensN, urlReading(rdU3, rdFree1, "refutes", "claim_contradicts_source"),
			recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE, byLocation(rdFree1)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			before := b.counts()
			r := b.mustRead(c.seat, c.args...)
			g := mintedBy(t, b, c.seat, r, c.class, c.place)
			if after := b.counts(); after != [3]int{before[0] + 1, before[1] + 1, before[2] + 1} {
				t.Errorf("the act wrote verify/mint/anchor %v over %v, want one of each", after, before)
			}
			if !strings.Contains(readReport(t, b.runDir), anchor.Token(g)) {
				t.Errorf("the report does not carry %s's token", g)
			}
		})
	}
}

// 16. A second reading of a sentence a standing gap of its class holds names that gap and mints none;
// a different admitted class mints.
func TestASecondRefutingReadingNamesTheStandingGap(t *testing.T) {
	b := readingRun(t)
	g := b.mustRead(lensL, citeReading(b.c1, "refutes", "claim_contradicts_source")...).Gap.ID
	for _, c := range []struct {
		name string
		seat string
		args []string
		held string
	}{
		{"another lens's citation reading", lensM, citeReading(b.c1, "refutes", "claim_contradicts_source"), g},
	} {
		before := b.counts()
		r := b.mustRead(c.seat, c.args...)
		if r.Gap == nil || r.Gap.ID != c.held || r.Gap.Minted {
			t.Fatalf("%s: gap = %+v, want %s named and not minted", c.name, r.Gap, c.held)
		}
		v, _ := b.lastVerify()
		if v.GetCarriedBy() != c.held || v.GetCarrierMinted() {
			t.Errorf("%s: stored carried_by %q carrier_minted %v, want %s and false", c.name, v.GetCarriedBy(), v.GetCarrierMinted(), c.held)
		}
		if after := b.counts(); after != [3]int{before[0] + 1, before[1], before[2]} {
			t.Errorf("%s: wrote verify/mint/anchor %v over %v, want the reading alone", c.name, after, before)
		}
	}
	// A DIFFERENT ADMITTED CLASS is a different gap.
	r := b.mustRead(lensN, citeReading(b.c1, "refutes", "within_source_condition_misattribution")...)
	if r.Gap == nil || !r.Gap.Minted || r.Gap.ID == g {
		t.Fatalf("a reading of another admitted class = %+v, want a gap of its own minted", r.Gap)
	}
	// THE FOUND-SOURCE FORM: the second source read against one sentence names the first's gap.
	g2 := b.mustRead(lensP, urlReading(rdU3, rdFree1, "refutes", "claim_contradicts_source")...).Gap.ID
	r = b.mustRead(lensK, urlReading(rdU4, rdFree1, "refutes", "claim_contradicts_source")...)
	if r.Gap == nil || r.Gap.ID != g2 || r.Gap.Minted {
		t.Fatalf("a second found-source reading of one sentence = %+v, want %s named", r.Gap, g2)
	}
}

// 17. A reading whose classes are non-empty requires the seven gap flags; one that mints nothing
// takes none.
func TestANegativeReadingWithoutGapFlagsIsRefused(t *testing.T) {
	b := readingRun(t)
	full := gapFlagArgs("claim_contradicts_source")
	for _, form := range []struct {
		name string
		base []string
	}{
		{"--anchor", citeReading(b.c1, "refutes", "")},
		{"--url", urlReading(rdU3, rdFree1, "refutes", "")},
	} {
		// Every flag absent, then each of the seven alone.
		_, err := b.verify(lensL, form.base...)
		mustRefuse(t, err, "", "--class is REQUIRED", "Nothing is recorded")
		for i := 0; i < len(full); i += 2 {
			args := append(append([]string{}, form.base...), full[:i]...)
			args = append(args, full[i+2:]...)
			_, err := b.verify(lensL, args...)
			mustRefuse(t, err, "", full[i]+" is REQUIRED")
		}
	}
	_, err := b.verify(lensL, urlReading(rdU3, rdFree1, "supports", "")...)
	mustRefuse(t, err, "", "--class is REQUIRED")
	_, err = b.verify(lensL, citeReading(b.c1, "absent", "")...)
	mustRefuse(t, err, "", "--class is REQUIRED")
	if c := b.counts(); c != [3]int{} {
		t.Errorf("refused readings wrote verify/mint/anchor %v", c)
	}
}

func TestABackingReadingWithGapFlagsIsRefused(t *testing.T) {
	b := readingRun(t)
	for _, c := range []struct {
		name string
		args []string
		says string
	}{
		{"--anchor supports", citeReading(b.c1, "supports", "claim_contradicts_source"), "a reading that backs the sentence mints nothing"},
		{"--anchor weak", citeReading(b.c1, "weak", "claim_contradicts_source"), "a reading that backs the sentence mints nothing"},
		{"--anchor unreachable", citeReading(b.c1, "unreachable", "claim_contradicts_source"), "an unreachable source was not read and mints no gap"},
		{"--url absent", urlReading(rdU3, rdFree1, "absent", "claim_contradicts_source"), "is `absent` records the reading and mints no gap"},
	} {
		_, err := b.verify(lensL, c.args...)
		mustRefuse(t, err, c.name, "--class", c.says)
	}
	if c := b.counts(); c != [3]int{} {
		t.Errorf("refused readings wrote verify/mint/anchor %v", c)
	}
}

// 18. A found-source `absent` records the reading and mints nothing (F5).
func TestAFoundSourceAbsentMintsNothing(t *testing.T) {
	b := readingRun(t)
	r := b.mustRead(lensL, urlReading(rdU3, rdFree1, "absent", "")...)
	if r.Gap != nil {
		t.Errorf("a found-source absent printed a gap: %+v", r.Gap)
	}
	v, _ := b.lastVerify()
	if v.CarriedBy != nil || v.CarrierMinted != nil {
		t.Errorf("a found-source absent carries %q / %v", v.GetCarriedBy(), v.CarrierMinted)
	}
	if c := b.counts(); c != [3]int{1, 0, 0} {
		t.Errorf("a found-source absent wrote verify/mint/anchor %v, want the reading alone", c)
	}
}

// 20. A citation reading's gap takes no mint: at the cap it lands, and the count does not move.
func TestAReadingsGapTakesNoMint(t *testing.T) {
	b := readingRunOf(t, rdFillerReport, 1, rdCited, rdBacked)
	b.fillTo(lensL, 0)
	before := b.standing(lensL)
	r := b.mustRead(lensL, citeReading(b.c1, "refutes", "claim_contradicts_source")...)
	g := mintedBy(t, b, lensL, r, recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE, bySentenceOf(b.c1))
	after := b.standing(lensL)
	if after.Open != before.Open || after.Answering != before.Answering+1 {
		t.Errorf("standing %+v → %+v; want open unchanged and %s answering outside it", before, after, g)
	}
	_, err := b.plainMint(lensL, rdFree2, "spec_underspecification")
	mustRefuse(t, err, "", "mint budget admits no more")
}

// 21. A plain mint a later reading names keeps its mint; a successor of a reading's gap is a plain
// mint and takes one.
func TestAPlainMintALaterReadingNamesKeepsItsMint(t *testing.T) {
	// A class a refuting reading admits that no reading has to precede: a plain
	// claim_contradicts_source mint stands on the lens's reading (standingreading_test.go).
	const wsc = "within_source_condition_misattribution"
	b := readingRunOf(t, rdFillerReport, 1, rdCited, rdBacked)
	g, err := b.plainMint(lensL, rdCited, wsc)
	if err != nil || g == "" {
		t.Fatalf("plain mint: %v", err)
	}
	b.fillTo(lensL, 0)
	before := b.standing(lensL)
	b.sit(lensL)
	r := b.mustRead(lensL, citeReading(b.c1, "refutes", wsc)...)
	if r.Gap == nil || r.Gap.ID != g || r.Gap.Minted {
		t.Fatalf("the reading's gap = %+v, want the plain mint %s named", r.Gap, g)
	}
	if v, _ := b.lastVerify(); v.GetCarriedBy() != g || v.GetCarrierMinted() {
		t.Errorf("stored carried_by %q carrier_minted %v, want %s and false", v.GetCarriedBy(), v.GetCarrierMinted(), g)
	}
	if after := b.standing(lensL); after.Open != before.Open || after.Answering != before.Answering {
		t.Errorf("standing %+v → %+v; the plain mint %s still takes its mint", before, after, g)
	}
	_, err = b.plainMint(lensL, rdFree2, "spec_underspecification")
	mustRefuse(t, err, "", "mint budget admits no more")

	t.Run("successor", func(t *testing.T) {
		b := readingRunOf(t, rdFillerReport, 1, rdCited, rdBacked)
		g1 := b.mustRead(lensL, citeReading(b.c1, "refutes", wsc)...).Gap.ID
		b.fillTo(lensL, 0)
		// The reading's gap awaits its `source` answer, so its successor carries that check kind.
		_, err := b.plainMint(lensL, rdCited, wsc, "--supersedes", g1, "--check-kind", "source")
		mustRefuse(t, err, "", "mint budget admits no more")
	})
}

// 22. The held gap is decided inside the reading's write: another lens's reading that lands between
// the verb's read and its transaction is named, not refused.
func TestTheReadingsHeldGapIsDecidedInsideItsWrite(t *testing.T) {
	b := readingRun(t)
	const g = "G-0b0b0b0b"
	cc := recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE
	acts := []*recordpb.Event{
		recordtest.At(t, lensM, lensM+":mint:"+g, &recordpb.Mint{GapId: proto.String(g), Class: cc.Enum(), Location: proto.String(rdCited),
			ClassMaterial: recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS.Enum(),
			Problem:       proto.String("the source says otherwise"), AcceptanceCheck: proto.String("the sentence states what the source says"),
			CheckKind: recordpb.CheckKind_CHECK_KIND_SOURCE.Enum(), Severity: recordpb.Grade_GRADE_HIGH.Enum(),
			Likelihood: recordpb.Grade_GRADE_HIGH.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum()}),
		recordtest.At(t, lensM, lensM+":anchor:"+g, &recordpb.Anchor{Id: proto.String(g), SentenceOf: proto.String(b.c1)}),
		recordtest.At(t, lensM, lensM+":verify:"+b.c1, &recordpb.Verify{Anchor: proto.String(b.c1), Claim: proto.String(rdCited),
			Outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(),
			Text: proto.String("read at the leaf"), CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum(),
			SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_ASKED.Enum(),
			CarriedBy:          proto.String(g), CarrierMinted: proto.Bool(true)}),
	}
	err := runtest.Overtaken(t, otherSeat(t, b.runDir), acts, func() error {
		_, err := b.verify(lensL, citeReading(b.c1, "refutes", "claim_contradicts_source")...)
		return err
	})
	if err != nil {
		t.Fatalf("A's reading over B's act that landed at its write was refused: %v", err)
	}
	var mine *recordpb.Verify
	for _, e := range storedOf[*recordpb.Verify](t, b.runDir) {
		if e.Seat == lensL {
			mine, _ = recordpb.BodyAs[*recordpb.Verify](e.Ev)
		}
	}
	if mine == nil || mine.GetCarriedBy() != g || mine.GetCarrierMinted() {
		t.Fatalf("A's reading = %v, want it carried by %s, not minted", mine, g)
	}
	n := 0
	for _, m := range mintsOf(t, b.runDir) {
		if m.GetClass() == cc {
			n++
		}
	}
	if n != 1 {
		t.Errorf("the record holds %d Mints of %v, want B's alone", n, cc)
	}
}

// 24. After every reading of a citation the verb words each standing failed check of it by its state.
func TestAReadingOfAFoundAgainstCitationIsWordedByItsState(t *testing.T) {
	banned := []string{"stays unread", "you owe a gap", "takes no mint of your budget", "read nothing first", "cites a source a lens can read"}
	check := func(t *testing.T, b readingBoard, seat string, args []string, owner, answered string, want ...string) {
		t.Helper()
		r := b.mustRead(seat, args...)
		var found bool
		for _, f := range r.FailedChecks {
			if f.State == string(record.CheckFoundAgainst) && f.Owner == owner && f.AnsweredBy == answered {
				found = true
			}
		}
		if !found {
			t.Errorf("failed_checks = %+v, want C's found_against check owned by %s answered by %q", r.FailedChecks, owner, answered)
		}
		b.sit(seat)
		out, err := b.verify(seat, args...)
		if err != nil {
			t.Fatalf("the human run: %v", err)
		}
		for _, w := range append([]string{"stays found against", "edits the sentence or retires or replaces the citation"}, want...) {
			if !strings.Contains(out, w) {
				t.Errorf("the text does not say %q:\n%s", w, out)
			}
		}
		for _, w := range banned {
			if strings.Contains(out, w) {
				t.Errorf("the text says %q:\n%s", w, out)
			}
		}
	}
	unreachable := func(b readingBoard) []string { return citeReading(b.c1, "unreachable", "") }

	b := readingRun(t)
	g := b.mustRead(lensL, citeReading(b.c1, "refutes", "claim_contradicts_source")...).Gap.ID
	t.Run("a: the reader's own unreachable", func(t *testing.T) {
		b.sit(lensL)
		check(t, b, lensL, unreachable(b), lensL, g)
	})
	t.Run("b: another lens's unreachable", func(t *testing.T) {
		check(t, b, lensM, unreachable(b), lensL, g)
	})
	t.Run("c: another lens's backing reading", func(t *testing.T) {
		check(t, b, lensN, citeReading(b.c1, "supports", ""), lensL, g, "clears nothing")
	})
	t.Run("d: carrier-less", func(t *testing.T) {
		b := readingRun(t)
		// AN ARCHIVED READING, AS THE REPLAY WRITES IT: from before the carrier fields, naming no gap.
		recordtest.Seed(t, b.runDir, recordtest.At(t, lensL, lensL+":seeded:archived-reading", &recordpb.Verify{Anchor: proto.String(b.c1), Claim: proto.String(rdCited),
			Outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(),
			Text: proto.String("read at the leaf"), CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum(),
			SourceCompleteness: recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_ASKED.Enum()}))
		b.sit(lensL)
		check(t, b, lensL, unreachable(b), lensL, "", "no gap names it")
		check(t, b, lensM, unreachable(b), lensL, "", "no gap names it")
	})
	t.Run("unread keeps its three sentences", func(t *testing.T) {
		b := readingRun(t)
		r := b.mustRead(lensL, unreachable(b)...)
		if len(r.FailedChecks) != 1 || r.FailedChecks[0].State != string(record.CheckUnread) {
			t.Fatalf("failed_checks = %+v, want one unread", r.FailedChecks)
		}
		b.sit(lensL)
		out, err := b.verify(lensL, unreachable(b)...)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out, "stays unread") || !strings.Contains(out, "you owe a gap about it") {
			t.Errorf("the unread words moved:\n%s", out)
		}
	})
}

// 35a. A found-source reading mints the gap its outcome admits, placed by its location.
func TestAFoundSourceReadingMintsItsGap(t *testing.T) {
	b := readingRunOf(t, rdFillerReport, 0, rdCited, rdBacked)
	before := citeTokens(readReport(t, b.runDir))
	us := recordpb.GapClass_GAP_CLASS_UNSOURCED_CLAIM
	for i, c := range []struct {
		as    string
		class recordpb.GapClass
		extra []string
	}{
		{"supports", us, nil},
		{"supports_with_bridge", us, []string{"--bridge", "summary"}},
		{"weak", us, nil},
		{"refutes", recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE, nil},
	} {
		t.Run(c.as, func(t *testing.T) {
			s := rdFillers[i]
			class := strings.TrimPrefix(strings.ToLower(c.class.String()), "gap_class_")
			r := b.mustRead(lensL, urlReading(rdU3, s, c.as, class, c.extra...)...)
			mintedBy(t, b, lensL, r, c.class, byLocation(s))
		})
	}
	if after := citeTokens(readReport(t, b.runDir)); len(after) != len(before) {
		t.Errorf("found-source readings placed citations: %v → %v", before, after)
	}
}

// 35b. One source backs many sentences in one sitting; one sentence read twice from it is refused.
func TestOneSourceBacksManySentences(t *testing.T) {
	b := readingRunOf(t, rdFillerReport, 0, rdCited, rdBacked)
	for _, s := range []string{rdFree1, rdFree2, rdFree3} {
		b.mustRead(lensL, urlReading(rdU3, s, "supports", "unsourced_claim")...)
	}
	if c := b.counts(); c != [3]int{3, 3, 3} {
		t.Errorf("three sentences backed by one source wrote verify/mint/anchor %v, want three of each", c)
	}
	b.mustRead(lensL, urlReading(rdU3, rdFillers[0], "absent", "")...)
	_, first := b.lastVerify()
	b.mustRead(lensL, urlReading(rdU3, rdFillers[1], "absent", "")...)
	held := b.counts()
	_, err := b.verify(lensL, urlReading(rdU3, rdFillers[0], "absent", "")...)
	mustRefuse(t, err, "", "you recorded a reading of "+rdU3+" for this sentence", fmt.Sprintf("at event %d", first))
	if c := b.counts(); c != held {
		t.Errorf("the refused repeat wrote %v over %v", c, held)
	}
}

// 35c. A backing found-source reading of a sentence a citation marks is refused by the guard, with or
// without the gap flags; a refuting one is admitted.
func TestABackingFoundSourceReadingOfAMarkedSentenceIsRefused(t *testing.T) {
	b := readingRun(t)
	for _, class := range []string{"unsourced_claim", ""} {
		_, err := b.verify(lensL, urlReading(rdU3, rdBacked, "supports", class)...)
		mustRefuse(t, err, "", "this sentence is marked by "+anchor.Token(b.c2), "--anchor "+b.c2)
		if err != nil && strings.Contains(err.Error(), "REQUIRED") {
			t.Errorf("the guard did not meet the reading first: %v", err)
		}
	}
	if c := b.counts(); c != [3]int{} {
		t.Errorf("refused readings wrote verify/mint/anchor %v", c)
	}
	r := b.mustRead(lensL, urlReading(rdU3, rdBacked, "refutes", "claim_contradicts_source")...)
	mintedBy(t, b, lensL, r, recordpb.GapClass_GAP_CLASS_CLAIM_CONTRADICTS_SOURCE, byLocation(rdBacked))

	// UNDER THE WRITE LOCK: blue's cite of the sentence lands after the verb read the report and
	// before the reading's transaction, and the guard inside the write refuses the reading. The cite
	// is the one blue's verb wrote on a twin board, whose report is this one's.
	twin := readingRun(t)
	if _, err := run(t, "cite", "--run", twin.runDir, "--seat-id", blueSeat, "--quote", rdFree1, "--url", rdU1, "--title", "Source C"); err != nil {
		t.Fatalf("the twin's cite: %v", err)
	}
	var acts []*recordpb.Event
	for _, e := range storedOf[*recordpb.Cite](t, twin.runDir) {
		if c, _ := recordpb.BodyAs[*recordpb.Cite](e.Ev); c.GetLocation() == rdFree1 || strings.Contains(rdFree1, c.GetLocation()) {
			acts = append(acts, recordtest.At(t, blueSeat, blueSeat+":cite:race", c))
			for _, a := range storedOf[*recordpb.Anchor](t, twin.runDir) {
				if x, _ := recordpb.BodyAs[*recordpb.Anchor](a.Ev); x.GetId() == c.GetLabel() {
					acts = append(acts, recordtest.At(t, blueSeat, blueSeat+":anchor:race", x))
				}
			}
		}
	}
	if len(acts) != 2 {
		t.Fatalf("the twin's cite of %q and its Anchor: %d events", rdFree1, len(acts))
	}
	held := b.counts()
	err := runtest.Overtaken(t, otherSeat(t, b.runDir), acts, func() error {
		_, err := b.verify(lensM, urlReading(rdU3, rdFree1, "supports", "unsourced_claim")...)
		return err
	})
	mustRefuse(t, err, "", "this sentence is marked by")
	if c := b.counts(); c[0] != held[0] || c[1] != held[1] {
		t.Errorf("the refused reading wrote verify/mint %v over %v", c, held)
	}
}

// 35d. verify takes exactly one subject, through cobra's flag groups: each refusal's diagnosis — its
// first line, before the usage of the flags involved — names the group the arguments violate.
func TestVerifyTakesExactlyOneSubject(t *testing.T) {
	b := readingRun(t)
	tail := []string{"--as", "supports", "--confidence", "high", "--reason", "r"}
	for _, c := range []struct {
		name string
		args []string
		says string
	}{
		{"neither", tail, "at least one of the flags in the group [anchor url] is required"},
		{"both", append([]string{"--anchor", b.c1, "--url", rdU3, "--title", "T", "--quote", rdFree1}, tail...), "[anchor url] were all set"},
		{"--anchor with --quote", append([]string{"--anchor", b.c1, "--quote", rdCited}, tail...), "group [url title quote] are set they must all be set; missing [title url]"},
		{"--anchor with --title", append([]string{"--anchor", b.c1, "--title", "T"}, tail...), "group [url title quote] are set they must all be set; missing [quote url]"},
		{"--url with --page", append([]string{"--url", rdU3, "--title", "T", "--quote", rdFree1, "--page", "2"}, tail...), "[page url] were all set"},
		{"--url without --title", append([]string{"--url", rdU3, "--quote", rdFree1}, tail...), "group [url title quote] are set they must all be set; missing [title]"},
		{"--url without --quote", append([]string{"--url", rdU3, "--title", "T"}, tail...), "group [url title quote] are set they must all be set; missing [quote]"},
	} {
		_, err := b.verify(lensL, c.args...)
		if err == nil {
			t.Errorf("%s: admitted; want a refusal saying %q", c.name, c.says)
			continue
		}
		if diag, _, _ := strings.Cut(err.Error(), "\n"); !strings.Contains(diag, c.says) {
			t.Errorf("%s: the refusal's diagnosis %q does not say %q", c.name, diag, c.says)
		}
	}
	if c := b.counts(); c != [3]int{} {
		t.Errorf("refused readings wrote verify/mint/anchor %v", c)
	}
	_, err := run(t, "corroborate", "--run", b.runDir, "--seat-id", lensL, "--url", rdU3)
	mustRefuse(t, err, "", `no command named "corroborate" exists`)
}

// 35e. One sentence read against many sources: the first mints, the rest name its gap, and a repeat is
// refused naming the stored reading (b7's shape).
func TestOneSentenceReadAgainstManySources(t *testing.T) {
	b := readingRun(t)
	var g string
	var second int64
	for i, u := range []string{rdU3, rdU4, rdU5, rdU6} {
		r := b.mustRead(lensL, urlReading(u, rdFree1, "supports", "unsourced_claim")...)
		if r.Gap == nil {
			t.Fatalf("reading %d carries no gap", i)
		}
		v, id := b.lastVerify()
		if i == 0 {
			g = r.Gap.ID
			if !r.Gap.Minted || !v.GetCarrierMinted() {
				t.Fatalf("the first reading did not mint: %+v", r.Gap)
			}
			continue
		}
		if i == 1 {
			second = id
		}
		if r.Gap.ID != g || r.Gap.Minted || v.GetCarriedBy() != g || v.GetCarrierMinted() {
			t.Errorf("reading %d: gap %+v, stored %q/%v; want %s named", i, r.Gap, v.GetCarriedBy(), v.GetCarrierMinted(), g)
		}
	}
	if c := b.counts(); c != [3]int{4, 1, 1} {
		t.Errorf("four sources against one sentence wrote verify/mint/anchor %v, want 4/1/1", c)
	}
	_, err := b.verify(lensL, urlReading(rdU4, rdFree1, "supports", "unsourced_claim")...)
	mustRefuse(t, err, "", "you recorded a reading of "+rdU4+" for this sentence", fmt.Sprintf("at event %d", second))
	if c := b.counts(); c != [3]int{4, 1, 1} {
		t.Errorf("the refused repeat wrote %v", c)
	}
}

// 35f. A found-source reading's gap takes a mint; a citation reading's does not.
func TestAFoundSourceGapTakesAMint(t *testing.T) {
	b := readingRunOf(t, rdFillerReport, 1, rdCited, rdBacked)
	b.fillTo(lensL, 0)
	held := b.counts()
	for _, args := range [][]string{
		urlReading(rdU3, rdFree1, "supports", "unsourced_claim"),
		urlReading(rdU3, rdFree2, "refutes", "claim_contradicts_source"),
	} {
		_, err := b.verify(lensL, args...)
		mustRefuse(t, err, "", "mint budget admits no more")
	}
	if c := b.counts(); c != held {
		t.Errorf("the refused readings wrote %v over %v", c, held)
	}
	st := b.standing(lensL)
	b.mustRead(lensL, citeReading(b.c1, "refutes", "claim_contradicts_source")...)
	if after := b.standing(lensL); after.Open != st.Open {
		t.Errorf("a citation reading's gap at the cap moved the open count %d → %d", st.Open, after.Open)
	}

	t.Run("below the cap", func(t *testing.T) {
		b := readingRunOf(t, rdFillerReport, 1, rdCited, rdBacked)
		b.fillTo(lensL, 1)
		st := b.standing(lensL)
		g := b.mustRead(lensL, urlReading(rdU3, rdFree1, "supports", "unsourced_claim")...).Gap.ID
		if after := b.standing(lensL); after.Open != st.Open+1 {
			t.Errorf("a found-source reading's gap moved the open count %d → %d, want +1", st.Open, after.Open)
		}
		st = b.standing(lensL)
		b.mustRead(lensL, citeReading(b.c1, "refutes", "claim_contradicts_source")...)
		if after := b.standing(lensL); after.Open != st.Open {
			t.Errorf("a citation reading's gap moved the open count %d → %d", st.Open, after.Open)
		}
		// AT THE CAP, a reading of a sentence a standing gap of its class carries names it.
		r := b.mustRead(lensL, urlReading(rdU4, rdFree1, "supports", "unsourced_claim")...)
		if r.Gap == nil || r.Gap.ID != g || r.Gap.Minted {
			t.Errorf("at the cap, a reading of a held sentence = %+v, want %s named", r.Gap, g)
		}
	})
}
