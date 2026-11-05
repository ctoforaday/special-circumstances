package cli

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE LIVE-ANSWER DEMAND, ASKED AT THE CLOSE (V1a′.2). A `source` or `computation` gap closes only
// when its demand is met (recordpb.DemandClause); the close verb asks it, under every disposition,
// before anything the record asks.

// demandGap mints gap key on the Findings section with check kind kind, from lensSeat.
func demandGap(t *testing.T, runDir, key, kind string, extra ...string) string {
	t.Helper()
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	args := append([]string{"--run", runDir, "--seat-id", lensSeat, "--class", "spec_underspecification", "--problem", "p", "--fix", "f",
		"--check", "c", "--severity", "low", "--likelihood", "low", "--impact", "low", "--key", key, "--check-kind", kind}, extra...)
	if !containsFlag(extra, "--about-kind") {
		args = append(args, "--about-kind", "section", "--about", "Findings")
	}
	out, err := runMint(t, args...)
	if err != nil {
		t.Fatalf("mint %s --check-kind %s: %v", key, kind, err)
	}
	return gapID(out)
}

func containsFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

// lensClose is lensSeat's close of gap as disposition, with the triple, and extra flags.
func lensClose(t *testing.T, runDir, seat, gap, as string, extra ...string) error {
	t.Helper()
	_, err := run(t, append([]string{"close", "--run", runDir, "--seat-id", seat, "--id", gap, "--as", as,
		"--verified-by", seat, "--verified-with", "show evidence", "--verified-against", "blue/report.md", "--reason", "checked it"}, extra...)...)
	return err
}

// placeAnswer places an anchor of kind on sentence as gap's answer (gap "" places none) and returns its id.
func placeAnswer(t *testing.T, runDir, kind, sentence, url, gap string) string {
	t.Helper()
	switch kind {
	case "citation":
		args := []string{"cite", "--run", runDir, "--seat-id", blueSeat, "--quote", sentence, "--url", url, "--title", "Source " + url}
		if gap != "" {
			args = append(args, "--answers", gap)
		}
		if _, err := run(t, args...); err != nil {
			t.Fatalf("cite %q --answers %s: %v", sentence, gap, err)
		}
		return lastBody(t, runDir, &recordpb.Cite{}).GetLabel()
	default:
		name := "answer-" + strings.Trim(strings.Fields(sentence)[1], ".") + ".js"
		s := script(t, runDir, name, "console.log('it holds');")
		args := []string{"compute", "--run", runDir, "--seat-id", blueSeat, "--quote", sentence, "--dir", s, "--reason", "the computation settles it"}
		if gap != "" {
			args = append(args, "--answers", gap)
		}
		if _, err := run(t, append(args, "--", "node", name)...); err != nil {
			t.Fatalf("compute %q --answers %s: %v", sentence, gap, err)
		}
		return lastBody(t, runDir, &recordpb.Computation{}).GetComputationId()
	}
}

func retireAnchor(t *testing.T, runDir, id string) {
	t.Helper()
	if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", id, "--reason", "it does not hold"); err != nil {
		t.Fatalf("retire --anchor %s: %v", id, err)
	}
}

// wantDemandRefusal holds err to the demand's refusal of gap, minted under kind, naming each of says.
func wantDemandRefusal(t *testing.T, err error, gap, kind string, says ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("the close of %s was admitted with its demand unmet", gap)
	}
	for _, w := range append([]string{"gap " + gap + " was minted --check-kind " + kind, "nothing meets its live-answer demand", "Nothing is recorded",
		recordpb.DemandClause, "`motion docket file`"}, says...) {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("the refusal does not say %q:\n%v", w, err)
		}
	}
}

func gapRow(t *testing.T, runDir, gap string) (open, answered, awaiting bool) {
	t.Helper()
	if err := openRunDB(t, runDir).QueryRow(`SELECT "open", "evidence_answered", "awaiting_evidence" FROM "gap" WHERE "gap_id" = ?`, gap).
		Scan(&open, &answered, &awaiting); err != nil {
		t.Fatal(err)
	}
	return open, answered, awaiting
}

// closingDispositions is every disposition whose facet says it closes, off the generated enum.
func closingDispositions(t *testing.T) []string {
	t.Helper()
	var out []string
	ds := recordpb.Disposition(0).Descriptor().Values()
	for i := 0; i < ds.Len(); i++ {
		if closes, declared, err := recordpb.Facet(ds.Get(i), "closes"); err == nil && declared && closes {
			out = append(out, recordpb.Word(recordpb.Disposition(ds.Get(i).Number())))
		}
	}
	if len(out) == 0 {
		t.Fatal("no disposition closes, so this reads nothing")
	}
	return out
}

// Per check kind: refused under every closing disposition, naming the kind's verb; a regrade is
// admitted; an answer of the other kind, and a retired answer of the demanded kind, meet nothing; a
// live answer of the demanded kind admits the close.
func TestTheCloseOfAnUnansweredEvidenceGapIsRefusedUnderEveryDisposition(t *testing.T) {
	for _, c := range []struct{ check, kind, other, verb string }{
		{"source", "citation", "computation", "cite"},
		{"computation", "computation", "citation", "compute"},
	} {
		t.Run(c.check, func(t *testing.T) {
			runDir := citedRun(t, "https://sky/1", "https://grass/1", "https://water/1")
			g := demandGap(t, runDir, "G", c.check)
			succ := demandGap(t, runDir, "D", "document")
			closes := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_CLOSE)
			for _, as := range closingDispositions(t) {
				var extra []string
				switch as {
				case "repaired_with_regression":
					extra = []string{"--superseded-by", succ}
				}
				wantDemandRefusal(t, lensClose(t, runDir, lensSeat, g, as, extra...), g, c.check, "blue's `"+c.verb+" --answers "+g+"`")
			}
			if n := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_CLOSE); n != closes {
				t.Fatalf("the refused closes recorded %d close(s)", n-closes)
			}
			if open, answered, awaiting := gapRow(t, runDir, g); !open || answered || !awaiting {
				t.Errorf("%s reads open=%v evidence_answered=%v awaiting_evidence=%v, want open, unanswered and awaiting", g, open, answered, awaiting)
			}
			if _, err := run(t, "regrade", "--run", runDir, "--seat-id", lensSeat, "--id", g, "--severity", "high", "--reason", "it still fails"); err != nil {
				t.Errorf("a regrade of the unanswered gap was refused: %v", err)
			}

			placeAnswer(t, runDir, c.other, "The sky is blue.", "https://sky/1", g)
			wantDemandRefusal(t, lensClose(t, runDir, lensSeat, g, "repaired"), g, c.check)
			retired := placeAnswer(t, runDir, c.kind, "The grass is green.", "https://grass/1", g)
			if _, answered, _ := gapRow(t, runDir, g); !answered {
				t.Fatalf("%s's live answer %s does not meet its demand", g, retired)
			}
			retireAnchor(t, runDir, retired)
			wantDemandRefusal(t, lensClose(t, runDir, lensSeat, g, "repaired"), g, c.check)
			placeAnswer(t, runDir, c.kind, "Water is wet.", "https://water/1", g)
			if err := lensClose(t, runDir, lensSeat, g, "repaired"); err != nil {
				t.Fatalf("the close of %s with a live %s answering it was refused: %v", g, c.kind, err)
			}
		})
	}
}

// L1, WIDENED (ruling L1): a gap about an evidence anchor is met by the anchor's retire, named or not,
// and the lens still closes it. Before the retire, the refusal names the retire and the lens's own
// check again of that anchor.
func TestAGapAboutARetiredEvidenceAnchorCloses(t *testing.T) {
	for _, c := range []struct{ check, kind, again string }{
		{"source", "citation", "a reading of %s by " + lensSeat},
		{"computation", "computation", "a `sound` judgement of %s by " + lensSeat},
	} {
		t.Run(c.check, func(t *testing.T) {
			runDir := citedRun(t, "https://sky/1")
			registerChairOnce(t, runDir)
			registerLensOnce(t, runDir)
			id := placeAnswer(t, runDir, c.kind, "The sky is blue.", "https://sky/1", "")
			g := demandGap(t, runDir, "about", c.check, "--about-kind", "anchor", "--about", id)
			wantDemandRefusal(t, lensClose(t, runDir, lensSeat, g, "repaired"), g, c.check,
				"blue's `retire --anchor "+id+"`, the "+c.kind+" it is about", strings.Replace(c.again, "%s", id, 1))
			retireAnchor(t, runDir, id)
			if open, answered, awaiting := gapRow(t, runDir, g); !open || !answered || awaiting {
				t.Errorf("after the retire %s reads open=%v evidence_answered=%v awaiting_evidence=%v, want open and met — nothing auto-closes", g, open, answered, awaiting)
			}
			if err := lensClose(t, runDir, lensSeat, g, "repaired"); err != nil {
				t.Fatalf("the close of %s after its anchor's retire was refused: %v", g, err)
			}
		})
	}
}

// CARRY ASKS NO DEMAND, AND CLOSES NO OPEN GAP. A carry of an open demand gap names no closure and
// is refused, stating the demand its lens's close asks, and the gap stays open and awaiting; a carry
// of a closed gap whose answer has since been retired restates the closure, and the gap stays
// closed.
func TestACarryRestatesAClosureAndClosesNoOpenDemandGap(t *testing.T) {
	runDir := citedRun(t, "https://sky/1")
	open := demandGap(t, runDir, "open", "source")
	g := demandGap(t, runDir, "closed", "source")
	c := placeAnswer(t, runDir, "citation", "The sky is blue.", "https://sky/1", g)
	if err := lensClose(t, runDir, lensSeat, g, "repaired"); err != nil {
		t.Fatalf("close %s: %v", g, err)
	}
	var epoch int
	if err := openRunDB(t, runDir).QueryRow(`SELECT e."epoch" FROM "close" c JOIN "events" e ON e."id" = c."event_id" WHERE c."gap_id" = ?`, g).Scan(&epoch); err != nil {
		t.Fatal(err)
	}
	retireAnchor(t, runDir, c)
	must(t, runDir, "register", "--seat-id", "red-chair")

	_, err := run(t, "carry", "--run", runDir, "--seat-id", "red-chair", "--id", open, "--carried-from", "1", "--as", "repaired")
	if err == nil || !strings.Contains(err.Error(), "no closure of it exists") || !strings.Contains(err.Error(), recordpb.DemandClause) {
		t.Fatalf("a carry of open demand gap %s = %v, want refused naming no closure and stating the demand", open, err)
	}
	if o, _, awaiting := gapRow(t, runDir, open); !o || !awaiting {
		t.Errorf("after the refused carry %s reads open=%v awaiting_evidence=%v, want open and awaiting", open, o, awaiting)
	}
	if _, err := run(t, "carry", "--run", runDir, "--seat-id", "red-chair", "--id", g, "--carried-from", itoa(epoch), "--as", "repaired"); err != nil {
		t.Fatalf("a carry of closed gap %s, its answer since retired, was refused: %v", g, err)
	}
	if o, _, _ := gapRow(t, runDir, g); o {
		t.Errorf("%s reopened", g)
	}
}

// THE PLACING VERB EACH ROW NAMES IS BLUE'S: a refusal that names `<verb> --answers G` names a
// command a seat can type.
func TestEveryEvidenceRowNamesABlueVerbThatPlacesIt(t *testing.T) {
	for _, row := range recordsql.EvidenceTables {
		out, err := run(t, row.Verb, "--seat-id", blueSeat, "--help")
		if err != nil || !strings.Contains(out, "--answers") {
			t.Errorf("row %s names verb %q, whose page = %v, want a blue verb taking --answers", row.Kind, row.Verb, err)
		}
	}
}

// THE DEMAND ON A READING'S GAP (V1a′.10 (2), #1647's shape): lens L's `refutes` reading of C1 mints
// `source` gap G, carried by the reading. Each row of the source reading-gap table, on its own board.
func TestTheDemandOnAReadingsSourceGap(t *testing.T) {
	const edited = "Costs rose steeply in the first quarter."
	reading := func(t *testing.T) (readingBoard, string) {
		b := readingRun(t)
		r := b.mustRead(lensL, citeReading(b.c1, "refutes", "claim_contradicts_source")...)
		if r.Gap == nil || r.Gap.ID == "" {
			t.Fatalf("the refuting reading minted no gap: %+v", r)
		}
		return b, r.Gap.ID
	}
	edit := func(t *testing.T, b readingBoard) {
		if _, err := run(t, "edit", "--run", b.runDir, "--seat-id", blueSeat, "--quote", "Costs rose sharply", "--new", "Costs rose steeply",
			"--reason", "the sentence now says what the source says"); err != nil {
			t.Fatalf("edit: %v", err)
		}
	}
	reread := func(t *testing.T, b readingBoard, seat string) {
		b.sit(seat)
		b.mustRead(seat, citeReading(b.c1, "supports", "")...)
	}
	for _, c := range []struct {
		name    string
		act     func(t *testing.T, b readingBoard, g string)
		refused []string
	}{
		{"blue retires the citation: met, arm 3", func(t *testing.T, b readingBoard, g string) { retireAnchor(t, b.runDir, b.c1) }, nil},
		{"blue edits the sentence: refused by the demand", func(t *testing.T, b readingBoard, _ string) { edit(t, b) },
			[]string{"blue's `cite --answers", "a reading of %c1 by " + lensL}},
		{"blue edits, then the reader's backing re-read: met, arm 4", func(t *testing.T, b readingBoard, _ string) {
			edit(t, b)
			reread(t, b, lensL)
		}, nil},
		{"blue edits, then cites on the edited sentence as its answer: met, arm 1", func(t *testing.T, b readingBoard, g string) {
			edit(t, b)
			if _, err := run(t, "cite", "--run", b.runDir, "--seat-id", blueSeat, "--quote", edited, "--url", rdU7, "--title", "Survey", "--answers", g); err != nil {
				t.Fatalf("cite --answers %s: %v", g, err)
			}
		}, nil},
		{"the reader's backing re-read, no edit: met, arm 4", func(t *testing.T, b readingBoard, _ string) { reread(t, b, lensL) }, nil},
		{"another lens's backing reading: refused by the demand", func(t *testing.T, b readingBoard, _ string) { reread(t, b, lensM) },
			[]string{"blue's `retire --anchor %c1`, the citation whose reading by " + lensL + " it carries"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			b, g := reading(t)
			c.act(t, b, g)
			err := lensClose(t, b.runDir, lensL, g, "repaired")
			if c.refused == nil {
				if err != nil {
					t.Fatalf("the close of %s was refused: %v", g, err)
				}
				return
			}
			var says []string
			for _, s := range c.refused {
				says = append(says, strings.ReplaceAll(s, "%c1", b.c1))
			}
			wantDemandRefusal(t, err, g, "source", says...)
		})
	}
}

// THE DEMAND ON A GAP OWED ABOUT AN ANCHOR (V1a′.10 (5), (6)): a check of the gap's own seat that
// holds, made since the mint, meets it; another lens's does not, and its refusal is the demand's.
func TestTheDemandOnAGapOwedAboutAnAnchor(t *testing.T) {
	t.Run("unread citation, source", func(t *testing.T) {
		for _, own := range []bool{true, false} {
			b := readingRun(t)
			b.mustRead(lensL, citeReading(b.c1, "unreachable", "")...)
			g := demandGap(t, b.runDir, "unread", "source", "--about-kind", "anchor", "--about", b.c1)
			seat := lensM
			if own {
				seat = lensL
			}
			b.sit(seat)
			b.mustRead(seat, citeReading(b.c1, "supports", "")...)
			err := lensClose(t, b.runDir, lensL, g, "repaired")
			if own && err != nil {
				t.Errorf("after the reader's own backing reading the close of %s was refused: %v", g, err)
			}
			if !own {
				wantDemandRefusal(t, err, g, "source", "a reading of "+b.c1+" by "+lensL)
			}
		}
	})
	t.Run("unsound computation, computation", func(t *testing.T) {
		for _, own := range []bool{true, false} {
			runDir := citedRun(t)
			registerChairOnce(t, runDir)
			registerLensOnce(t, runDir)
			k := placeAnswer(t, runDir, "computation", "The sky is blue.", "", "")
			if _, err := run(t, "reproduce", "--run", runDir, "--seat-id", lensSeat, "--anchor", k, "--as", "unsound", "--reason", "it computes another claim"); err != nil {
				t.Fatalf("reproduce unsound: %v", err)
			}
			g := demandGap(t, runDir, "unsound", "computation", "--about-kind", "anchor", "--about", k)
			seat := "red-lens-logic"
			if own {
				seat = lensSeat
			}
			registerSeatOnce(t, runDir, seat)
			if _, err := run(t, "reproduce", "--run", runDir, "--seat-id", seat, "--anchor", k, "--as", "sound", "--reason", "it computes the claim"); err != nil {
				t.Fatalf("reproduce sound by %s: %v", seat, err)
			}
			err := lensClose(t, runDir, lensSeat, g, "repaired")
			if own && err != nil {
				t.Errorf("after the lens's own sound judgement the close of %s was refused: %v", g, err)
			}
			if !own {
				wantDemandRefusal(t, err, g, "computation", "a `sound` judgement of "+k+" by "+lensSeat)
			}
		}
	})
}

// THE DEMAND IS THE RECORD'S FIELD, READ AS THE VERB READS IT: an about anchor of a kind that backs
// nothing is no anchor the demand reaches.
func TestTheDemandReachesOnlyEvidenceAnchors(t *testing.T) {
	runDir := citedRun(t)
	g := demandGap(t, runDir, "G", "source")
	d, err := record.EvidenceDemandOf(runtest.Open(t, runDir), g)
	if err != nil {
		t.Fatal(err)
	}
	if d.AnswerKind != "citation" || d.AnswerVerb != "cite" || d.About != "" || len(d.Readings) != 0 || d.MintedBy != lensSeat {
		t.Errorf("the demand of section gap %s = %+v, want a citation owed, minted by %s, reaching no anchor", g, d, lensSeat)
	}
}

// THE DEMAND IS ASKED IN THE WRITE. A retire of the gap's only answer commits on another seat's
// connection at the instant the close opens its write: the verb's read saw the demand met, and the
// write still refuses the close, because evidence_answered reads false on the record it lands on.
func TestARetireOfTheOnlyAnswerRacingTheCloseRefusesIt(t *testing.T) {
	runDir := citedRun(t, "https://sky/1")
	g := demandGap(t, runDir, "G", "source")
	c := placeAnswer(t, runDir, "citation", "The sky is blue.", "https://sky/1", g)
	if _, answered, _ := gapRow(t, runDir, g); !answered {
		t.Fatalf("the citation placed as %s's answer does not answer it", g)
	}
	retired := recordtest.At(t, "blue-lane-1", "race:retire", &recordpb.Retire{Claim: proto.String("The sky is blue."), Reason: proto.String("unreadable"),
		Anchors: []string{c}, ClaimKept: proto.Bool(true), RemovalBasis: proto.String(record.RemovalAsserted)})
	err := runtest.Overtaken(t, otherSeat(t, runDir), []*recordpb.Event{retired}, func() error {
		return lensClose(t, runDir, lensSeat, g, "repaired")
	})
	wantDemandRefusal(t, err, g, "source")
	if open, _, _ := gapRow(t, runDir, g); !open {
		t.Errorf("the refused close closed %s", g)
	}
}
