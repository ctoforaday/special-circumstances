package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/bluedoc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/consistency"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// cut is blue's edit of quote to nothing: a pure cut when quote covers whole sentences.
func cut(t *testing.T, runDir, quote string, extra ...string) (string, error) {
	t.Helper()
	return run(t, append([]string{"edit", "--run", runDir, "--seat-id", blueSeat, "--quote", quote, "--new", "", "--reason", "the claim cannot stand"}, extra...)...)
}

// A PURE CUT TAKES ITS ANCHORS (P12b, ruling 2): an edit whose replacement holds nothing of the whole
// sentences it locates removes their anchors, of every kind, and records them on the edit,
// tool-stamped, as removed_anchors. The sentence and its token leave together; the edit's --reason is
// why. No anchor stands bare after it, and none is listed as reopened. A gap whose marker the cut took
// stays open, reading gone: the cut is its raiser's turn, never a close.
func TestAPureCutTakesItsAnchors(t *testing.T) {
	const sentence = "The grass is green."
	for _, kind := range anchor.Kinds() {
		t.Run(kind, func(t *testing.T) {
			runDir := citedRun(t, "https://water/1")
			registerChairOnce(t, runDir)
			registerLensOnce(t, runDir)
			id, _ := markerOfKind(t, runDir, kind, sentence)
			if !strings.Contains(readReport(t, runDir), "The grass is green"+anchor.Token(id)) {
				t.Fatalf("no %s anchor on the sentence:\n%s", kind, readReport(t, runDir))
			}
			out, err := cut(t, runDir, sentence)
			if err != nil {
				t.Fatalf("the pure cut of the %s's sentence was refused: %v", kind, err)
			}
			// The seat is told what its cut took, by id: the act names what it removed.
			if !strings.Contains(out, id) || !strings.Contains(out, "cut") {
				t.Errorf("the cut's result does not name the %s anchor %s it took:\n%s", kind, id, out)
			}
			ev := lastBody(t, runDir, &recordpb.BlueEdit{})
			if !slices.Equal(ev.GetRemovedAnchors(), []string{id}) || slices.Contains(ev.GetReopened(), id) {
				t.Errorf("the cut recorded removed_anchors %v reopened %v, want [%s] and not reopened", ev.GetRemovedAnchors(), ev.GetReopened(), id)
			}
			want := "# Findings\n\nThe sky is blue.\n\n- Water is wet.\n- Fire is hot.\n\nClosing paragraph stays.\n"
			rep := readReport(t, runDir)
			if rep != want {
				t.Errorf("the report after the cut:\n got %q\nwant %q", rep, want)
			}
			if bare := claimcount.BareAnchorIDs(rep); len(bare) != 0 {
				t.Errorf("the cut left %v bare", bare)
			}
			r, found, err := record.AnchorRemovedAt(runtest.Open(t, runDir), id)
			if err != nil || !found || r.By != record.RemovedByEdit || r.Reason != "the claim cannot stand" {
				t.Errorf("anchor_removal for %s = %+v found=%v (%v), want the edit with its reason", id, r, found, err)
			}
			if kind == "gap" {
				b := boardOf(t, runDir)
				if len(b.Open) == 0 || b.Open[len(b.Open)-1].ID != id || b.Open[len(b.Open)-1].LocationState != record.LocationGone {
					t.Errorf("gap %s is not open and gone after its marker's cut: %+v", id, b.Open)
				}
			}
		})
	}
}

// A REWRITE CARRIES EVERY ANCHOR. A replacement with prose drops no token: one it neither carries nor
// keeps the sentence of word for word is refused (ErrAnchorDropped), naming the two ways — carry it, or
// cut its sentence whole — and nothing is recorded. Carried, the rewrite lands and removes nothing.
func TestAProseRewriteCarriesEveryAnchor(t *testing.T) {
	runDir := citedRun(t, "https://grass/1")
	c := citeSentence(t, runDir, "The grass is green.", "https://grass/1")
	tok := anchor.Token(c)
	before, edits := readReport(t, runDir), countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT)
	_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--quote", "The grass is green"+tok+".", "--new", "The grass is lush.", "--reason", "reword")
	if err == nil {
		t.Fatal("a rewrite dropping the citation landed")
	}
	for _, w := range []string{c, "cut its sentence whole"} {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("the refusal does not say %q:\n%v", w, err)
		}
	}
	if readReport(t, runDir) != before || countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT) != edits {
		t.Fatal("the refused rewrite changed the report or the record")
	}
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--quote", "The grass is green"+tok+".", "--new", "The grass is lush"+tok+".", "--reason", "reword"); err != nil {
		t.Fatalf("a rewrite carrying the citation was refused: %v", err)
	}
	if ev := lastBody(t, runDir, &recordpb.BlueEdit{}); len(ev.GetRemovedAnchors()) != 0 {
		t.Errorf("a rewrite recorded removed_anchors %v", ev.GetRemovedAnchors())
	}
	if !strings.Contains(readReport(t, runDir), "The grass is lush"+tok+".") {
		t.Errorf("the rewrite did not carry the citation:\n%s", readReport(t, runDir))
	}
}

// A MIXED EDIT IS A REWRITE: a non-empty replacement of a two-sentence span (S1+S2 → S1′) must carry S2's
// tokens, or it is refused as dropping them — it is never a cut of S2.
func TestAMixedEditIsARewrite(t *testing.T) {
	runDir := citedRun(t, "https://grass/1")
	c := citeSentence(t, runDir, "The grass is green.", "https://grass/1")
	before := readReport(t, runDir)
	_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--quote", "The sky is blue. The grass is green"+anchor.Token(c)+".",
		"--new", "The sky is azure.", "--reason", "fold the two")
	if err == nil {
		t.Fatal("a rewrite of S1+S2 to S1′ dropping S2's citation landed as a cut")
	}
	if !strings.Contains(err.Error(), c) {
		t.Errorf("the refusal does not name %s:\n%v", c, err)
	}
	if readReport(t, runDir) != before || countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT) != 0 {
		t.Error("the refused mixed edit changed the report or the record")
	}
}

// NO WRITER LEAVES A BARE ANCHOR. A rewrite that leaves a token of any kind on a sentence with no words is
// refused, carried or not — a carried gap marker included — naming the anchor and the two
// routes; a lane's part holding any anchor token is refused at ingest; a prescription that would leave a
// token on no words is refused at the mint.
func TestNoWriterLeavesABareAnchor(t *testing.T) {
	const sentence = "Water is wet."
	for _, kind := range anchor.Kinds() {
		t.Run("edit/"+kind, func(t *testing.T) {
			runDir := citedRun(t, "https://water/1")
			registerChairOnce(t, runDir)
			registerLensOnce(t, runDir)
			id, _ := markerOfKind(t, runDir, kind, sentence)
			tok := anchor.Token(id)
			before := readReport(t, runDir)
			_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--quote", "Water is wet"+tok+".", "--new", tok, "--reason", "gut it")
			if err == nil {
				t.Fatalf("an edit leaving the %s anchor %s bare landed:\n%s", kind, id, readReport(t, runDir))
			}
			for _, w := range []string{tok, "keep words on", "cut it whole"} {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("the refusal does not say %q:\n%v", w, err)
				}
			}
			if readReport(t, runDir) != before || countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT) != 0 {
				t.Error("the refused edit changed the report or the record")
			}
		})
	}
	t.Run("ingest", func(t *testing.T) {
		runDir := laneRun(t, 1)
		writePart(t, runDir, "blue-lane-1", "## The factors\n\n91 is 7 x 13"+anchor.Token("G-0000abcd")+".\n")
		if _, err := ingestPart(t, runDir, "blue-lane-1"); err == nil || !strings.Contains(err.Error(), anchor.Token("G-0000abcd")) {
			t.Errorf("a part holding an anchor token was not refused naming it: %v", err)
		}
		if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REPORT_PART); n != 0 {
			t.Errorf("the refused ingest recorded %d part(s)", n)
		}
	})
	t.Run("mint", func(t *testing.T) {
		runDir := fixNewReport(t)
		if _, err := mintWithNew(t, runDir, "V1", "."); err == nil || !strings.Contains(err.Error(), "NOTHING IS RECORDED") {
			t.Errorf("a prescription leaving the gap's marker on no words was not refused: %v", err)
		}
		if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); n != 0 {
			t.Errorf("the refused mint recorded %d mint(s)", n)
		}
	})
}

// REPLAY REMOVES WHAT THE WRITE REMOVED, ONCE (F17 (a)). The splice never removes a token; the cut's list
// does, after it. One subtest per token position: inside the span, quoted or not, and abutting it,
// quoted or not. Each: the list names the token once, report_op holds one remove for it, the edit's rows
// carry ord 0 for the splice and 1 for the remove, and the replayed report is the write's plan.
func TestReplayRemovesWhatTheWriteRemoved(t *testing.T) {
	for _, tc := range []struct {
		name, on   string // the gap's quote
		quote      func(tok string) string
		quotedForm string
	}{
		{"inside the span, unquoted", "The grass", func(string) string { return "The grass is green." }, ""},
		{"inside the span, quoted", "The grass", func(tok string) string { return "The grass" + tok + " is green." }, ""},
		{"abutting, quoted", "The grass is green.", func(tok string) string { return "The grass is green" + tok + "." }, ""},
		{"abutting, unquoted", "The grass is green.", func(string) string { return "The grass is green." }, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runDir := citedRun(t, "https://grass/1")
			registerLensOnce(t, runDir)
			f := markClaim(t, runDir, "G1", tc.on, "spec_underspecification")
			tok := anchor.Token(f)
			before := readReport(t, runDir)
			q := tc.quote(tok)
			start, end, err := bluedoc.LocateUniqueReplacing(before, q)
			if err != nil {
				t.Fatalf("the quote %q does not locate:\n%s\n%v", q, before, err)
			}
			if _, err := cut(t, runDir, q); err != nil {
				t.Fatalf("the cut %q was refused: %v", q, err)
			}
			ev := lastBody(t, runDir, &recordpb.BlueEdit{})
			if !slices.Equal(ev.GetRemovedAnchors(), []string{f}) {
				t.Errorf("removed_anchors = %v, want [%s] once", ev.GetRemovedAnchors(), f)
			}
			db := openRunDB(t, runDir)
			var editID int64
			if err := db.QueryRow(`SELECT max("id") FROM "events" WHERE "type" = ?`, recordpb.Word(recordpb.EventType_EVENT_TYPE_BLUE_EDIT)).Scan(&editID); err != nil {
				t.Fatal(err)
			}
			rows, err := db.Query(`SELECT "kind", COALESCE("a", ''), "ord" FROM "report_op" WHERE "event_id" = ? ORDER BY "ord"`, editID)
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for rows.Next() {
				var kind, a string
				var ord int
				if err := rows.Scan(&kind, &a, &ord); err != nil {
					t.Fatal(err)
				}
				if kind == "edit" {
					a = ""
				}
				got = append(got, kind+":"+a+":"+string(rune('0'+ord)))
			}
			rows.Close()
			if want := []string{"edit::0", "remove:" + f + ":1"}; !slices.Equal(got, want) {
				t.Errorf("the edit's report_op rows = %v, want %v", got, want)
			}
			var removes int
			if err := db.QueryRow(`SELECT count(*) FROM "report_op" WHERE "kind" = 'remove' AND "a" = ?`, f).Scan(&removes); err != nil || removes != 1 {
				t.Errorf("report_op holds %d remove(s) of %s (%v), want 1", removes, f, err)
			}
			plan, err := reportproj.RemoveListed(reportproj.ApplySplice(before, start, end, ""), []string{f})
			if err != nil {
				t.Fatalf("the write's plan: %v", err)
			}
			if rep := readReport(t, runDir); rep != plan || strings.Contains(rep, tok) || !strings.Contains(rep, "The sky is blue.") {
				t.Errorf("replay:\n got %q\nwant %q (the write's plan, without %s)", rep, plan, tok)
			}
		})
	}
}

// L6 HOLDS THE CUT (ruling 9, F19). A pure cut whose removed anchors include a citation a computation
// still in the report applies is refused, naming the computation, and nothing changes; blue re-points or
// removes the computation first, by a retire or a cut, or takes it in the same cut.
func TestACutIsHeldToL6(t *testing.T) {
	const cSentence, kSentence = "The grass is green.", "The sky is blue."
	setup := func(t *testing.T, kOn string) (runDir, c, k string) {
		t.Helper()
		runDir = citedRun(t, "https://grass/1", "https://water/1")
		c = citeSentence(t, runDir, cSentence, "https://grass/1")
		s := script(t, runDir, "k.js", "console.log('k');")
		if _, err := run(t, "compute", "--run", runDir, "--seat-id", blueSeat, "--quote", kOn, "--dir", s,
			"--reason", "the method applied", "--cites", c, "--", "node", "k.js"); err != nil {
			t.Fatalf("compute --cites %s: %v", c, err)
		}
		return runDir, c, lastBody(t, runDir, &recordpb.Computation{}).GetComputationId()
	}
	refused := func(t *testing.T, runDir, k string) {
		t.Helper()
		before, edits := readReport(t, runDir), countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT)
		_, err := cut(t, runDir, cSentence)
		if err == nil || !strings.Contains(err.Error(), k) {
			t.Fatalf("the cut of a citation %s applies was not refused naming it: %v", k, err)
		}
		if readReport(t, runDir) != before || countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT) != edits {
			t.Fatal("the refused cut changed the report or the record")
		}
	}
	lands := func(t *testing.T, runDir, quote string, want ...string) {
		t.Helper()
		if _, err := cut(t, runDir, quote); err != nil {
			t.Fatalf("the cut of %q was refused: %v", quote, err)
		}
		if got := lastBody(t, runDir, &recordpb.BlueEdit{}).GetRemovedAnchors(); !slices.Equal(got, want) {
			t.Errorf("removed_anchors = %v, want %v", got, want)
		}
		consistentL6(t, runDir)
	}
	t.Run("refused", func(t *testing.T) {
		runDir, _, k := setup(t, kSentence)
		refused(t, runDir, k)
	})
	t.Run("re-point, then cut", func(t *testing.T) {
		runDir, c, k := setup(t, kSentence)
		refused(t, runDir, k)
		c2 := citeSentence(t, runDir, "Water is wet.", "https://water/1")
		s := script(t, runDir, "k2.js", "console.log('k2');")
		if _, err := run(t, "compute", "--run", runDir, "--seat-id", blueSeat, "--quote", kSentence, "--dir", s,
			"--reason", "the method applied", "--cites", c2, "--", "node", "k2.js"); err != nil {
			t.Fatalf("compute --cites %s: %v", c2, err)
		}
		if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", k, "--reason", "re-pointed"); err != nil {
			t.Fatalf("retire %s: %v", k, err)
		}
		lands(t, runDir, cSentence, c)
	})
	t.Run("retire, then cut", func(t *testing.T) {
		runDir, c, k := setup(t, kSentence)
		if _, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", k, "--reason", "unsound"); err != nil {
			t.Fatalf("retire %s: %v", k, err)
		}
		lands(t, runDir, cSentence, c)
	})
	t.Run("same act", func(t *testing.T) {
		runDir, c, k := setup(t, cSentence)
		_, err := cut(t, runDir, cSentence)
		if err != nil {
			t.Fatalf("the cut taking %s with %s was refused: %v", c, k, err)
		}
		if got := lastBody(t, runDir, &recordpb.BlueEdit{}).GetRemovedAnchors(); len(got) != 2 || !slices.Contains(got, c) || !slices.Contains(got, k) {
			t.Errorf("removed_anchors = %v, want %s and %s", got, c, k)
		}
		consistentL6(t, runDir)
	})
	t.Run("the computation cut first", func(t *testing.T) {
		runDir, c, k := setup(t, kSentence)
		lands(t, runDir, kSentence, k)
		lands(t, runDir, cSentence, c)
	})
}

// consistentL6 holds the record to consistency.Check's L6 fold: a landed cut leaves no computation in
// the report naming a citation in no document.
func consistentL6(t *testing.T, runDir string) {
	t.Helper()
	res, err := consistency.Check(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range res.Violations {
		if v.Rule == "l6" {
			t.Errorf("the landed cut left an L6 violation: %s", v)
		}
	}
}
