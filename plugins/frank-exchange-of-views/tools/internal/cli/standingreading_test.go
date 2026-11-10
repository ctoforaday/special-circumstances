package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A GAP IS THE ONE PASS-HOLDER FOR A DEFECT (#1727). `mint` of claim_contradicts_source and
// citation_figure_misattribution stands on the minting seat's OWN STANDING READING of the mint's
// subject — the citation for a cited source, the sentence for a source the seat found — and is refused
// until that reading stands. The one-gap rule gains an anchor arm, for every class, and a cross-arm: a
// gap about a citation and a gap quoting the sentence that holds it are one subject per class.
//
// IN-ACT SCAFFOLD (phase 1a of plans/gap-is-the-one-holder.md). `verify` still mints a gap in the same
// act as a reading that finds against, so the readings here carry an in-act class the tests never
// mint about the citation — within_source_condition_misattribution — and a found-source refutes
// reading's in-act claim_contradicts_source on the sentence is named in --supersedes by the mint that
// stands on the reading. Phase 1b deletes the in-act mint and with it both arrangements: the readings
// then take no class and the mints name no --supersedes.
const inActClass = "within_source_condition_misattribution"

// read is seat's reading of citation c with outcome as, the in-act class supplied where the reading
// mints (refutes, absent).
func (b readingBoard) read(seat, c, as string) readingJSON {
	b.t.Helper()
	class := ""
	if as == "refutes" || as == "absent" {
		class = inActClass
	}
	return b.mustRead(seat, citeReading(c, as, class)...)
}

// readFound is seat's reading of url, a source it found, against quote with outcome as; the in-act
// class is claim_contradicts_source for refutes and unsourced_claim for a backing outcome.
func (b readingBoard) readFound(seat, url, quote, as string) readingJSON {
	b.t.Helper()
	class := ""
	switch as {
	case "refutes":
		class = "claim_contradicts_source"
	case "supports", "weak":
		class = "unsourced_claim"
	}
	return b.mustRead(seat, urlReading(url, quote, as, class)...)
}

// supersedingInAct is the --supersedes flag naming r's in-act gap, where r minted one of class; nothing
// otherwise.
func supersedingInAct(r readingJSON, class string) []string {
	if r.Gap != nil && r.Gap.Minted && r.Gap.Class == class {
		return []string{"--supersedes", r.Gap.ID}
	}
	return nil
}

// mintSource is seat's mint of class with the subject args given, a `source` gap as a reading's gap is.
func (b readingBoard) mintSource(seat, class string, subject ...string) (string, string, error) {
	b.t.Helper()
	out, err := run(b.t, append([]string{"mint", "--run", b.runDir, "--seat-id", seat, "--class", class,
		"--problem", "the source says otherwise than the sentence", "--fix", "the fix", "--check-kind", "source",
		"--check", "the sentence states what the source says", "--severity", "high", "--likelihood", "high", "--impact", "medium"}, subject...)...)
	return gapID(out), out, err
}

// about is the subject args of a mint about citation c.
func about(c string) []string { return []string{"--about-kind", "anchor", "--about", c} }

// readingBehind is the events.id of the reading behind gap, 0 where none stands.
func readingBehind(t *testing.T, runDir, gap string) int64 {
	t.Helper()
	_, id, has, err := record.ReadingBehind(runtest.Open(t, runDir), gap)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		return 0
	}
	return id
}

// mintedAbout asserts the stored Mint of gap is about citation c with no quote, and returns it.
func mintedAbout(t *testing.T, runDir, gap, c string) *recordpb.Mint {
	t.Helper()
	m := mintOf(t, runDir, gap)
	if m == nil {
		t.Fatalf("no Mint of %s is on the record", gap)
	}
	if m.GetAboutKind() != recordpb.AboutKind_ABOUT_KIND_ANCHOR || m.GetAboutRef() != c || m.GetLocation() != "" {
		t.Errorf("gap %s is about %v %q with location %q, want about anchor %s with none", gap, m.GetAboutKind(), m.GetAboutRef(), m.GetLocation(), c)
	}
	return m
}

// rewordCited is blue's edit of the cited sentence old to new, every token the report places before
// its stop carried — the citation's and any gap's.
func (b readingBoard) rewordCited(old, new string) {
	b.t.Helper()
	rep := readReport(b.t, b.runDir)
	body, stop := strings.TrimSuffix(old, "."), "."
	i := strings.Index(rep, body)
	if i < 0 {
		b.t.Fatalf("the report does not hold %q:\n%s", body, rep)
	}
	j := i + len(body)
	for strings.HasPrefix(rep[j:], "<!--") {
		j += strings.Index(rep[j:], "-->") + len("-->")
	}
	tokens := rep[i+len(body) : j]
	if _, err := run(b.t, "edit", "--run", b.runDir, "--seat-id", blueSeat, "--quote", body+tokens+stop,
		"--new", strings.TrimSuffix(new, ".")+tokens+stop, "--reason", "reworded"); err != nil {
		b.t.Fatalf("edit %q: %v", old, err)
	}
}

// standingRefusal is what the standing-reading check says, whatever the class and subject.
const standingRefusal = "no standing reading of yours"

// T3. A claim_contradicts_source gap about a citation stands on the minter's own refuting reading of it.
func TestAContradictionGapAboutACitationStandsOnTheMintersRefutingReading(t *testing.T) {
	b := readingRun(t)
	_, out, err := b.mintSource(lensL, "claim_contradicts_source", about(b.c1)...)
	mustRefuse(t, err, out, standingRefusal, "citation "+b.c1, "verify --anchor "+b.c1+" --as refutes", "(none)")
	if c := b.counts(); c != [3]int{} {
		t.Errorf("the refusal wrote verify/mint/anchor %v", c)
	}
	b.read(lensL, b.c1, "supports")
	_, out, err = b.mintSource(lensL, "claim_contradicts_source", about(b.c1)...)
	mustRefuse(t, err, out, standingRefusal, "is `supports`")

	b.sit(lensL)
	r := b.read(lensL, b.c1, "refutes")
	_, at := b.lastVerify()
	if r.Gap == nil || r.Gap.Class != inActClass {
		t.Fatalf("the scaffold's in-act gap is %+v, want one of class %s", r.Gap, inActClass)
	}
	gap, out, err := b.mintSource(lensL, "claim_contradicts_source", about(b.c1)...)
	if err != nil {
		t.Fatalf("the mint on the standing refuting reading was refused: %v\n%s", err, out)
	}
	mintedAbout(t, b.runDir, gap, b.c1)
	if got := readingBehind(t, b.runDir, gap); got != at {
		t.Errorf("the reading behind %s is event %d, want the refuting reading at %d", gap, got, at)
	}
}

// T4. Standing is the minter's LATEST reading of the citation since its sentence last changed.
func TestTheReadingBehindAMintIsTheMintersLatestSinceTheEdit(t *testing.T) {
	t.Run("blue's edit of the sentence unseats the reading", func(t *testing.T) {
		b := readingRun(t)
		b.read(lensL, b.c1, "refutes")
		b.rewordCited(rdCited, "Costs rose gently in the first quarter.")
		_, out, err := b.mintSource(lensL, "claim_contradicts_source", about(b.c1)...)
		mustRefuse(t, err, out, standingRefusal, "is of a sentence blue has since edited")
		b.sit(lensL)
		b.read(lensL, b.c1, "refutes")
		_, at := b.lastVerify()
		gap, out, err := b.mintSource(lensL, "claim_contradicts_source", about(b.c1)...)
		if err != nil {
			t.Fatalf("the mint on the re-read was refused: %v\n%s", err, out)
		}
		if got := readingBehind(t, b.runDir, gap); got != at {
			t.Errorf("the reading behind %s is event %d, want the re-read at %d", gap, got, at)
		}
	})
	t.Run("the seat's later reading replaces its earlier one", func(t *testing.T) {
		b := readingRun(t)
		b.read(lensL, b.c1, "refutes")
		b.sit(lensL)
		b.read(lensL, b.c1, "supports")
		_, out, err := b.mintSource(lensL, "claim_contradicts_source", about(b.c1)...)
		mustRefuse(t, err, out, standingRefusal, "is `supports`")
	})
}

// T5. Another seat's reading admits nothing: the key is the minting seat.
func TestAnotherSeatsReadingDoesNotAdmitTheMint(t *testing.T) {
	b := readingRun(t)
	b.read(lensM, b.c1, "refutes")
	b.read(lensM, b.c2, "absent")
	for _, c := range []struct{ class, anchor string }{
		{"claim_contradicts_source", b.c1}, {"citation_figure_misattribution", b.c2},
	} {
		_, out, err := b.mintSource(lensL, c.class, about(c.anchor)...)
		mustRefuse(t, err, out, standingRefusal, "citation "+c.anchor, "(none)")
	}
	if n := countType(t, b.runDir, recordpb.EventType_EVENT_TYPE_MINT) - b.base[1]; n != 2 {
		t.Errorf("the record holds %d mints beyond the fixture's, want the two in-act scaffold gaps alone", n)
	}
}

// T6. A citation_figure_misattribution gap about a citation stands on the minter's `absent` reading of
// the work, OR on its reading of a copy that is not the work (two predicates, F3 (a)).
func TestAMisattributionGapAboutACitationStandsOnAbsentOrANotTheWorkCopy(t *testing.T) {
	t.Run("absent", func(t *testing.T) {
		b := readingRun(t)
		b.read(lensL, b.c1, "absent")
		_, at := b.lastVerify()
		gap, out, err := b.mintSource(lensL, "citation_figure_misattribution", about(b.c1)...)
		if err != nil {
			t.Fatalf("the mint on the absent reading was refused: %v\n%s", err, out)
		}
		mintedAbout(t, b.runDir, gap, b.c1)
		if got := readingBehind(t, b.runDir, gap); got != at {
			t.Errorf("the reading behind %s is event %d, want the absent reading at %d", gap, got, at)
		}
	})
	t.Run("a copy that is not the work", func(t *testing.T) {
		b := readingRun(t)
		b.cacheAs(rdU1, fetchcache.CompletenessNotTheWork)
		r := b.mustRead(lensL, citeReading(b.c1, "refutes", "")...)
		if r.Gap != nil {
			t.Fatalf("a not-the-work reading minted %+v", r.Gap)
		}
		_, at := b.lastVerify()
		gap, out, err := b.mintSource(lensL, "citation_figure_misattribution", about(b.c1)...)
		if err != nil {
			t.Fatalf("the mint on the not-the-work reading was refused: %v\n%s", err, out)
		}
		mintedAbout(t, b.runDir, gap, b.c1)
		if got := readingBehind(t, b.runDir, gap); got != at {
			t.Errorf("the reading behind %s is event %d, want the not-the-work reading at %d", gap, got, at)
		}
		// The gap answers the unread check the reading left, takes no mint, and the ledger keeps the
		// citation unread.
		fs := failedOf(t, b.runDir, b.c1)
		if len(fs) != 1 || fs[0].CheckState() != record.CheckUnread || fs[0].Owner != lensL || fs[0].AnsweredBy() != gap {
			t.Errorf("C's failed checks = %+v, want one unread owned by %s answered by %s", fs, lensL, gap)
		}
		if st := b.standing(lensL); st.Open != 0 || st.Answering != 1 {
			t.Errorf("%s's standing is open %d answering %d, want 0 and 1: the gap takes no mint", lensL, st.Open, st.Answering)
		}
	})
	t.Run("refutes admits no misattribution", func(t *testing.T) {
		b := readingRun(t)
		b.read(lensL, b.c1, "refutes")
		_, out, err := b.mintSource(lensL, "citation_figure_misattribution", about(b.c1)...)
		mustRefuse(t, err, out, standingRefusal, "is `refutes`", "--as absent")
	})
}

// T7. A found source's claim_contradicts_source gap quotes the sentence and stands on the minter's
// reading of a source it found for THAT sentence — any url; the key is the sentence.
func TestAFoundSourceContradictionGapStandsOnTheMintersReadingOfTheSentence(t *testing.T) {
	const class = "claim_contradicts_source"
	t.Run("the forms", func(t *testing.T) {
		b := readingRun(t)
		_, out, err := b.mintSource(lensL, class, "--quote", rdFree1)
		mustRefuse(t, err, out, standingRefusal, "a source you found", "verify --url", "(none)")
		r := b.readFound(lensL, rdU3, rdFree1, "refutes")
		_, at := b.lastVerify()
		gap, out, err := b.mintSource(lensL, class, append([]string{"--quote", rdFree1}, supersedingInAct(r, class)...)...)
		if err != nil {
			t.Fatalf("the mint on the found-source reading was refused: %v\n%s", err, out)
		}
		if got := readingBehind(t, b.runDir, gap); got != at {
			t.Errorf("the reading behind %s is event %d, want the reading at %d", gap, got, at)
		}
		_, out, err = b.mintSource(lensL, class, "--quote", rdFree2)
		mustRefuse(t, err, out, standingRefusal, "(none)")
		_, out, err = b.mintSource(lensL, "citation_figure_misattribution", "--quote", rdFree1)
		mustRefuse(t, err, out, "a cited source's defect", "--about-kind anchor --about C-")
	})
	t.Run("a quote the report holds twice is the placement's refusal", func(t *testing.T) {
		b := readingRunOf(t, rdFillerReport, 0, rdCited, rdBacked)
		_, out, err := b.mintSource(lensL, class, "--quote", "stands here.")
		if err == nil {
			t.Fatalf("a quote the report holds eight times placed a gap: %s", out)
		}
		if strings.Contains(err.Error(), standingRefusal) {
			t.Errorf("the refusal is the standing-reading check's, want the placement's own:\n%v", err)
		}
	})
	t.Run("the key is the sentence", func(t *testing.T) {
		b := readingRun(t)
		r1 := b.readFound(lensL, rdU3, rdFree1, "refutes")
		_, at1 := b.lastVerify()
		r2 := b.readFound(lensL, rdU3, rdFree2, "refutes")
		_, at2 := b.lastVerify()
		g1, out, err := b.mintSource(lensL, class, append([]string{"--quote", rdFree1}, supersedingInAct(r1, class)...)...)
		if err != nil {
			t.Fatalf("the first sentence's mint: %v\n%s", err, out)
		}
		g2, out, err := b.mintSource(lensL, class, append([]string{"--quote", rdFree2}, supersedingInAct(r2, class)...)...)
		if err != nil {
			t.Fatalf("the second sentence's mint: %v\n%s", err, out)
		}
		// A reading of the same url against a third sentence leaves the first two standing.
		b.readFound(lensL, rdU3, rdFree3, "refutes")
		if got := readingBehind(t, b.runDir, g1); got != at1 {
			t.Errorf("the reading behind %s is event %d, want %d", g1, got, at1)
		}
		if got := readingBehind(t, b.runDir, g2); got != at2 {
			t.Errorf("the reading behind %s is event %d, want %d", g2, got, at2)
		}
		// A later reading of ANOTHER url for the first sentence replaces it: the standing reading of
		// the sentence is the backing one, and a contradiction gap on it is refused for its ground —
		// before the one-gap rule, which would also hold it.
		b.readFound(lensL, rdU4, rdFree1, "supports")
		_, out, err = b.mintSource(lensL, class, "--quote", rdFree1)
		mustRefuse(t, err, out, standingRefusal, rdU4, "is `supports`")
		if got := readingBehind(t, b.runDir, g1); got != 0 {
			t.Errorf("the reading behind %s is event %d after a backing reading of the sentence, want none", g1, got)
		}
	})
}

// T8. One open gap per class about one citation, for every class; a gap about a citation and a gap
// quoting the sentence that holds it are one subject (F9); the gap a seat's own unread check obliges
// is the one exception.
func TestOneGapPerClassAboutOneCitationAndAcrossItsSentence(t *testing.T) {
	const (
		contradicts = "claim_contradicts_source"
		judgement   = "claim_contradicts_own_record"
		plain       = "spec_underspecification"
		oneGap      = "Two defects of one class on one subject are one gap"
	)
	t.Run("the anchor arm", func(t *testing.T) {
		b := readingRun(t)
		b.read(lensL, b.c1, "refutes")
		gx, out, err := b.mintSource(lensL, contradicts, about(b.c1)...)
		if err != nil {
			t.Fatalf("X's gap about C: %v\n%s", err, out)
		}
		b.read(lensM, b.c1, "refutes")
		_, out, err = b.mintSource(lensM, contradicts, about(b.c1)...)
		mustRefuse(t, err, out, gx, "citation "+b.c1, oneGap)
		// Every class: a plain judgement about the citation is held to it too.
		gs, out, err := b.mintSource(lensM, plain, about(b.c1)...)
		if err != nil {
			t.Fatalf("Y's plain gap about C: %v\n%s", err, out)
		}
		_, out, err = b.mintSource(lensN, plain, about(b.c1)...)
		mustRefuse(t, err, out, gs, "citation "+b.c1, oneGap)
		// The exception: the gap Y's own unread check obliges lands over X's open gap of its class.
		if _, out, err := b.mintSource(lensL, plain, about(b.c2)...); err != nil {
			t.Fatalf("X's plain gap about C2: %v\n%s", err, out)
		}
		b.read(lensM, b.c2, "unreachable")
		if _, out, err := b.mintSource(lensM, plain, about(b.c2)...); err != nil {
			t.Fatalf("the gap Y's unread check obliges was refused over X's open gap: %v\n%s", err, out)
		}
	})
	t.Run("the cross-arm, quote over about", func(t *testing.T) {
		b := readingRun(t)
		b.read(lensL, b.c1, "refutes")
		gx, out, err := b.mintSource(lensL, contradicts, about(b.c1)...)
		if err != nil {
			t.Fatalf("X's gap about C: %v\n%s", err, out)
		}
		r := b.readFound(lensM, rdU5, rdCited, "refutes")
		_, out, err = b.mintSource(lensM, contradicts, append([]string{"--quote", rdCited}, supersedingInAct(r, contradicts)...)...)
		mustRefuse(t, err, out, gx, "citation "+b.c1, oneGap)
		// A judgement class, the same direction.
		gj, out, err := b.mintSource(lensL, judgement, about(b.c1)...)
		if err != nil {
			t.Fatalf("X's judgement about C: %v\n%s", err, out)
		}
		_, out, err = b.mintSource(lensM, judgement, "--quote", rdCited)
		mustRefuse(t, err, out, gj, "citation "+b.c1, oneGap)
	})
	t.Run("the cross-arm, about over quote", func(t *testing.T) {
		b := readingRun(t)
		r := b.readFound(lensL, rdU6, rdBacked, "refutes")
		gq := r.Gap.ID // the in-act quote gap of class claim_contradicts_source on C2's sentence
		b.read(lensM, b.c2, "refutes")
		_, out, err := b.mintSource(lensM, contradicts, about(b.c2)...)
		mustRefuse(t, err, out, gq, "the sentence", "holds citation "+b.c2, oneGap)
		// A judgement class, the same direction.
		gj, out, err := b.mintSource(lensL, judgement, "--quote", rdBacked)
		if err != nil {
			t.Fatalf("X's judgement quoting C2's sentence: %v\n%s", err, out)
		}
		_, out, err = b.mintSource(lensM, judgement, about(b.c2)...)
		mustRefuse(t, err, out, gj, "the sentence", oneGap)
	})
	t.Run("a sentence holding two citations", func(t *testing.T) {
		const (
			first  = "Costs rose sharply"
			second = "margins held steady"
			joined = first + " and " + second + " through the winter."
		)
		b := readingRunOf(t, "# Findings\n\n"+joined+" "+rdFree1+"\n", 0, first, second)
		if _, out, err := b.mintSource(lensL, judgement, about(b.c2)...); err != nil {
			t.Fatalf("X's judgement about C2: %v\n%s", err, out)
		}
		_, out, err := b.mintSource(lensM, judgement, "--quote", joined)
		mustRefuse(t, err, out, "citation "+b.c2, oneGap)
		if _, out, err := b.mintSource(lensN, judgement, about(b.c1)...); err != nil {
			t.Fatalf("a gap about C1 was refused by the open gap about C2 on the same sentence: %v\n%s", err, out)
		}
	})
	t.Run("the obliged gap lands over an open quote gap of its class", func(t *testing.T) {
		b := readingRun(t)
		if _, out, err := b.mintSource(lensL, plain, "--quote", rdCited); err != nil {
			t.Fatalf("X's plain gap on C's sentence: %v\n%s", err, out)
		}
		b.read(lensM, b.c1, "unreachable")
		if _, out, err := b.mintSource(lensM, plain, about(b.c1)...); err != nil {
			t.Fatalf("the gap Y's unread check obliges was refused over X's open quote gap: %v\n%s", err, out)
		}
	})
}

// T16. Every class whose mint stands on a reading is driven: a row per class names the reading that
// admits it, a wrong outcome, and whether a found source's reading admits its `--quote` form. A class
// with no row fails naming it.
func TestEveryReadingMintedClassIsDriven(t *testing.T) {
	rows := map[string]struct {
		admits, wrong string
		found         bool
	}{
		"claim_contradicts_source":       {admits: "refutes", wrong: "absent", found: true},
		"citation_figure_misattribution": {admits: "absent", wrong: "refutes", found: false},
	}
	for _, c := range record.ReadingMintedClasses() {
		class := recordpb.Word(c)
		row, ok := rows[class]
		if !ok {
			t.Errorf("class %s stands on a reading and has no fixture row — add one, so its refusal and its admitted mint are driven", class)
			continue
		}
		t.Run(class, func(t *testing.T) {
			b := readingRun(t)
			_, out, err := b.mintSource(lensL, class, about(b.c1)...)
			mustRefuse(t, err, out, standingRefusal, "(none)")
			b.read(lensL, b.c1, row.wrong)
			_, out, err = b.mintSource(lensL, class, about(b.c1)...)
			mustRefuse(t, err, out, standingRefusal, fmt.Sprintf("is `%s`", row.wrong))
			b.sit(lensL)
			b.read(lensL, b.c1, row.admits)
			_, at := b.lastVerify()
			gap, out, err := b.mintSource(lensL, class, about(b.c1)...)
			if err != nil {
				t.Fatalf("the mint on the %s reading was refused: %v\n%s", row.admits, err, out)
			}
			if got := readingBehind(t, b.runDir, gap); got != at {
				t.Errorf("the reading behind %s is event %d, want %d", gap, got, at)
			}
			if !row.found {
				_, out, err := b.mintSource(lensL, class, "--quote", rdFree1)
				mustRefuse(t, err, out, "a cited source's defect")
				return
			}
			_, out, err = b.mintSource(lensL, class, "--quote", rdFree1)
			mustRefuse(t, err, out, standingRefusal, "a source you found")
			r := b.readFound(lensL, rdU3, rdFree1, row.admits)
			_, at = b.lastVerify()
			gap, out, err = b.mintSource(lensL, class, append([]string{"--quote", rdFree1}, supersedingInAct(r, class)...)...)
			if err != nil {
				t.Fatalf("the quote mint on the found-source reading was refused: %v\n%s", err, out)
			}
			if got := readingBehind(t, b.runDir, gap); got != at {
				t.Errorf("the reading behind %s is event %d, want %d", gap, got, at)
			}
		})
	}
}
