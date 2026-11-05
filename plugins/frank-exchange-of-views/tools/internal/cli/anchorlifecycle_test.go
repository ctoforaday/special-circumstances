package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// ONE ANCHOR LIFECYCLE, EVERY EVIDENCE KIND (V1a′.1, the verbs' arms). The same script runs for a
// citation and a computation through the CLI: placed as a gap's answer; an edit leaving it bare is
// refused, naming `retire --anchor`; a gap about it is admitted while it stands; `retire --anchor`
// alone takes it out and keeps its sentence word for word, the result naming it by its kind; its
// check and a gap about it are then refused, each naming the retire, and nothing is recorded.
func TestAnEvidenceAnchorLivesOneLifecycle(t *testing.T) {
	const sentence, words = "The sky is blue.", "The sky is blue"
	for _, kind := range []string{"citation", "computation"} {
		t.Run(kind, func(t *testing.T) {
			runDir := citedRun(t, "https://sky/1")
			gap := mintGap(t, runDir, "G", "spec_underspecification")
			var id string
			switch kind {
			case "citation":
				if _, err := run(t, "cite", "--run", runDir, "--seat-id", blueSeat, "--quote", sentence, "--url", "https://sky/1",
					"--title", "Source", "--answers", gap); err != nil {
					t.Fatalf("cite --answers: %v", err)
				}
				c := lastBody(t, runDir, &recordpb.Cite{})
				id = c.GetLabel()
				if c.GetAnswers() != gap {
					t.Errorf("the citation answers %q, want %s", c.GetAnswers(), gap)
				}
			case "computation":
				s := script(t, runDir, "sky.js", "console.log('blue');")
				if _, err := run(t, "compute", "--run", runDir, "--seat-id", blueSeat, "--quote", sentence, "--dir", s,
					"--reason", "the computation settles it", "--answers", gap, "--", "node", "sky.js"); err != nil {
					t.Fatalf("compute --answers: %v", err)
				}
				c := lastBody(t, runDir, &recordpb.Computation{})
				id = c.GetComputationId()
				if c.GetAnswers() != gap {
					t.Errorf("the computation answers %q, want %s", c.GetAnswers(), gap)
				}
			}
			tok := anchor.Token(id)
			// shows is what evidence.md prints of the anchor: a computation's entry carries its id, a
			// citation's its source.
			shows := id
			if kind == "citation" {
				shows = "https://sky/1"
			}

			// AN EDIT LEAVING IT BARE IS REFUSED, carried or not, naming the anchor by its kind and the retire.
			before := readReport(t, runDir)
			for _, cut := range [][2]string{{words, ""}, {words + tok + ".", tok}} {
				_, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--quote", cut[0], "--new", cut[1], "--reason", "the claim cannot stand", "--answers", gap)
				if err == nil {
					t.Fatalf("an edit leaving %s %s bare was recorded (--quote %q)", kind, id, cut[0])
				}
				for _, w := range []string{kind + " " + id + " standing bare", "`retire --anchor <its id> --answers " + gap + "`", "Nothing is recorded"} {
					if !strings.Contains(err.Error(), w) {
						t.Errorf("the refusal does not say %q:\n%v", w, err)
					}
				}
			}
			if after := readReport(t, runDir); after != before {
				t.Fatalf("a refused edit changed the report:\n%s", after)
			}

			// A GAP ABOUT IT IS ADMITTED WHILE IT STANDS.
			if _, err := mintWith(t, runDir, "--key", "about-live", "--about-kind", "anchor", "--about", id); err != nil {
				t.Fatalf("a gap about a live %s was refused: %v", kind, err)
			}

			// EVERY DOCUMENT AND THE EVIDENCE TABLE CARRY IT WHILE IT STANDS: evidence.md's entry names
			// the sentence it backs.
			set := assembledSet(t, runDir)
			if ev := set[report.FileEvidence]; !strings.Contains(ev, shows) || !strings.Contains(ev, "- **backs**: "+sentence) {
				t.Errorf("evidence.md does not carry the live %s %s with the sentence it backs:\n%s", kind, id, ev)
			}
			if rows, _ := evidenceRows(t, runDir); !rows[id] {
				t.Errorf("show evidence's lists do not hold the live %s %s", kind, id)
			}

			// RETIRED ALONE, ITS SENTENCE KEPT.
			out, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", id, "--answers", gap, "--reason", "it does not hold")
			if err != nil {
				t.Fatalf("retire --anchor %s alone: %v", id, err)
			}
			ev := lastBody(t, runDir, &recordpb.Retire{})
			if !ev.GetClaimKept() || !slices.Equal(ev.GetAnchors(), []string{id}) || ev.GetClaim() != sentence || ev.GetAnswers() != gap {
				t.Errorf("the retire recorded claim %q kept=%v anchors %v answers %q, want %q kept, [%s], %s", ev.GetClaim(), ev.GetClaimKept(), ev.GetAnchors(), ev.GetAnswers(), sentence, id, gap)
			}
			if !strings.Contains(out, "retired "+kind+" "+id) {
				t.Errorf("the retire's result does not name the %s by its kind:\n%s", kind, out)
			}
			if rep := readReport(t, runDir); strings.Contains(rep, tok) || !strings.Contains(rep, sentence) {
				t.Errorf("after the retire the report should hold the sentence and not the anchor:\n%s", rep)
			}

			// A RETIRED ANCHOR IS IN NO DOCUMENT, and in none of show evidence's lists or counts; read at
			// its id it says it is retired, and every surface naming the retire names it by its kind.
			set = assembledSet(t, runDir)
			for _, file := range []string{report.FileReport, report.FileEvidence} {
				if strings.Contains(set[file], id) || strings.Contains(set[file], shows) {
					t.Errorf("%s still holds the retired %s %s:\n%s", file, kind, id, set[file])
				}
			}
			if cl := set[report.FileChangelog]; !strings.Contains(cl, "a "+kind+" it stood on was retired from it, answering **"+gap+"**") {
				t.Errorf("CHANGELOG does not name the retire by its kind:\n%s", cl)
			}
			if rows, counts := evidenceRows(t, runDir); rows[id] || counts != 0 {
				t.Errorf("show evidence still lists or counts the retired %s %s (%d anchor(s) counted)", kind, id, counts)
			}
			at, err := run(t, "show", "evidence", "--run", runDir, "--seat-id", lensSeat, "--anchor", id)
			if err != nil || !strings.Contains(at, `"retired": {`) || !strings.Contains(at, `"reason": "it does not hold"`) {
				t.Errorf("show evidence --anchor %s = %v, want its row with the retire:\n%s", id, err, at)
			}
			if ch, err := run(t, "show", "changes", "--run", runDir, "--seat-id", lensSeat); err != nil || !strings.Contains(ch, "retired "+kind+" "+id+", its sentence kept") {
				t.Errorf("show changes does not name the retire by its kind (%v):\n%s", err, ch)
			}

			// ITS CHECK IS REFUSED, NAMING THE RETIRE.
			check := []string{"verify", "--run", runDir, "--seat-id", lensSeat, "--anchor", id, "--as", "supports", "--confidence", "high", "--reason", "read it"}
			if kind == "computation" {
				check = []string{"reproduce", "--run", runDir, "--seat-id", "red-lens-computation", "--anchor", id, "--as", "sound", "--reason", "re-ran it"}
			}
			if _, err := run(t, check...); err == nil || !strings.Contains(err.Error(), kind+" "+id+" is retired") || !strings.Contains(err.Error(), `reason: "it does not hold"`) {
				t.Errorf("a check of the retired %s = %v, want refused naming its retire", kind, err)
			}

			// A GAP ABOUT IT IS REFUSED, NAMING THE RETIRE, and nothing is recorded.
			mints := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_MINT)
			_, err = mintWith(t, runDir, "--key", "about-retired", "--about-kind", "anchor", "--about", id)
			if err == nil {
				t.Fatalf("a gap about a retired %s was recorded", kind)
			}
			for _, w := range []string{kind + " " + id + ", which is retired", "blue retired it at event", `reason: "it does not hold"`, "its sentence \"" + sentence + "\" kept", "no gap about it", "file it with --quote"} {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("the refusal does not say %q:\n%v", w, err)
				}
			}
			if n := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); n != mints {
				t.Errorf("the refused mint recorded %d mint(s)", n-mints)
			}

			// A FINDING ABOUT IT IS REFUSED THE SAME WAY, and nothing is recorded.
			findings := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_FINDING)
			_, err = run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", "F-about-retired", "--about-kind", "anchor", "--about", id,
				"--reason", "the anchor is wrong", "--severity", "low", "--likelihood", "low", "--impact", "low")
			if err == nil {
				t.Fatalf("a finding about a retired %s was recorded", kind)
			}
			for _, w := range []string{kind + " " + id + ", which is retired", "blue retired it at event", "no finding about it", "file it with --quote"} {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("the finding's refusal does not say %q:\n%v", w, err)
				}
			}
			if n := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_FINDING); n != findings {
				t.Errorf("the refused finding recorded %d finding(s)", n-findings)
			}

			// THE REPORT READ AT ITS ID NAMES IT BY ITS KIND AND ITS RETIRE.
			_, err = run(t, "show", "report", "--run", runDir, "--seat-id", lensSeat, "--anchor", id)
			if err == nil {
				t.Fatalf("show report --anchor %s read a retired %s", id, kind)
			}
			for _, w := range []string{"a retired " + kind, "blue retired it at event", "its sentence \"" + sentence + "\" kept", "read that sentence with --quote"} {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("show report --anchor's refusal does not say %q:\n%v", w, err)
				}
			}
		})
	}
}

// assembledSet is the run's assembled report set, each document's text by its file name.
func assembledSet(t *testing.T, runDir string) map[string]string {
	t.Helper()
	a, err := report.Assemble(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, file := range []string{report.FileReport, report.FileEvidence, report.FileChangelog} {
		if b, err := os.ReadFile(filepath.Join(filepath.Dir(a.Report), file)); err == nil {
			out[file] = string(b)
		}
	}
	return out
}

// evidenceRows is the anchors `show evidence --json` lists, and how many sources and computations it counts.
func evidenceRows(t *testing.T, runDir string) (map[string]bool, int) {
	t.Helper()
	out, err := run(t, "show", "evidence", "--run", runDir, "--seat-id", lensSeat, "--json")
	if err != nil {
		t.Fatalf("show evidence --json: %v", err)
	}
	var ev record.EvidenceJSON
	if err := json.Unmarshal([]byte(out), &ev); err != nil {
		t.Fatalf("show evidence --json: %v\n%s", err, out)
	}
	rows := map[string]bool{}
	for _, s := range ev.Sources {
		rows[s.Anchor] = true
	}
	for _, c := range ev.Computations {
		rows[c.Anchor] = true
	}
	return rows, ev.Counts.Sources + ev.Counts.Computations
}

// RETRY, CROSS-KIND REFERENCE AND THE CITATION'S ANSWER (V1a′.4, the verbs' arms). A keyed compute
// retry returns its computation's id; a keyed cite retry naming another gap is refused; a Cite naming no
// gap is refused by the record past the flag; `--cites` refuses a retired citation; and a citation two
// live computations apply cannot be retired until both are (ruling L6).
func TestACitationAComputationAppliesStaysWhileTheComputationDoes(t *testing.T) {
	runDir := citedRun(t, "https://sky/1", "https://grass/1")
	g1 := mintGap(t, runDir, "G1", "spec_underspecification")
	g2 := mintGap(t, runDir, "G2", "invalid_inference")

	// THE KEYED CITE RETRY RESTATES THE GAP IT ANSWERS.
	cite := func(answers string) (string, error) {
		return run(t, "cite", "--run", runDir, "--seat-id", blueSeat, "--quote", "The sky is blue.", "--url", "https://sky/1",
			"--title", "Source", "--key", "C1", "--answers", answers)
	}
	if _, err := cite(g1); err != nil {
		t.Fatalf("cite --answers: %v", err)
	}
	c := lastBody(t, runDir, &recordpb.Cite{}).GetLabel()
	if _, err := cite(g2); err == nil || !strings.Contains(err.Error(), "answers") {
		t.Errorf("a keyed cite retry naming another gap = %v, want refused naming answers", err)
	}

	// THE RECORD REFUSES A CITATION ANSWERING NO GAP, past the flag.
	stray := proto.Clone(lastBody(t, runDir, &recordpb.Cite{})).(*recordpb.Cite)
	stray.Label, stray.CiteKey, stray.Answers = proto.String(record.NewID("citation")), proto.String("stray"), proto.String("G-0000dead")
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: blueSeat}, stray); err == nil || !strings.Contains(err.Error(), "no mint event created") {
		t.Errorf("a Cite whose answers names no gap = %v, want refused by the record", err)
	}

	// THE KEYED COMPUTE RETRY RETURNS ITS ID.
	compute := func(sentence, name, key string) (string, error) {
		s := script(t, runDir, name, "console.log('it holds');")
		return run(t, "compute", "--run", runDir, "--seat-id", blueSeat, "--quote", sentence, "--dir", s, "--key", key,
			"--cites", c, "--reason", "the computation settles it", "--", "node", name)
	}
	if _, err := compute("Water is wet.", "wet.js", "P1"); err != nil {
		t.Fatalf("compute --cites: %v", err)
	}
	k1 := lastBody(t, runDir, &recordpb.Computation{}).GetComputationId()
	if out, err := compute("Water is wet.", "wet.js", "P1"); err != nil || !strings.Contains(out, k1) {
		t.Errorf("the keyed compute retry = %q (%v), want it to name %s", out, err, k1)
	}
	if _, err := compute("Fire is hot.", "hot.js", "P2"); err != nil {
		t.Fatalf("compute --cites: %v", err)
	}
	k2 := lastBody(t, runDir, &recordpb.Computation{}).GetComputationId()

	// L6: REFUSED WHILE EITHER STANDS, NAMING EACH THAT DOES; ADMITTED ONCE BOTH ARE RETIRED.
	retire := func(id string) (string, error) {
		return run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--anchor", id, "--reason", "it does not hold")
	}
	retires := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE)
	if _, err := retire(c); err == nil || !strings.Contains(err.Error(), k1) || !strings.Contains(err.Error(), k2) {
		t.Errorf("retiring %s while %s and %s apply it = %v, want refused naming both", c, k1, k2, err)
	}
	if _, err := retire(k1); err != nil {
		t.Fatalf("retire --anchor %s: %v", k1, err)
	}
	if _, err := retire(c); err == nil || strings.Contains(err.Error(), k1) || !strings.Contains(err.Error(), k2) {
		t.Errorf("retiring %s while %s alone applies it = %v, want refused naming it alone", c, k2, err)
	}
	if n := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_RETIRE); n != retires+1 {
		t.Errorf("%d retire(s) recorded, want the computation's alone", n-retires)
	}
	if _, err := retire(k2); err != nil {
		t.Fatalf("retire --anchor %s: %v", k2, err)
	}
	if _, err := retire(c); err != nil {
		t.Fatalf("retiring %s once no live computation applies it: %v", c, err)
	}
	if ev := lastBody(t, runDir, &recordpb.Retire{}); !slices.Equal(ev.GetAnchors(), []string{c}) {
		t.Errorf("the citation's retire took %v, want [%s] alone", ev.GetAnchors(), c)
	}

	// --CITES A RETIRED CITATION IS REFUSED.
	if _, err := compute("The grass is green.", "green.js", "P3"); err == nil || !strings.Contains(err.Error(), "retired citation") {
		t.Errorf("compute --cites a retired citation = %v, want refused", err)
	}
}

// A CLAIM RETIRE IS HELD TO L6, AND ONE ACT MAY TAKE BOTH (ruling Same-act L6). Blue's edit cuts the
// sentence a citation and the computation applying it stood on, carrying both onto the sentence that
// stays; a claim retire taking the citation alone is refused, naming the computation, and one taking
// both is admitted.
func TestAClaimRetireTakesACitationWithTheComputationsApplyingIt(t *testing.T) {
	runDir := citedRun(t, "https://sky/1")
	c := citeSentence(t, runDir, "The sky is blue.", "https://sky/1")
	s := script(t, runDir, "sky.js", "console.log('blue');")
	if _, err := run(t, "compute", "--run", runDir, "--seat-id", blueSeat, "--quote", "The sky is blue", "--dir", s, "--cites", c,
		"--reason", "the computation settles it", "--", "node", "sky.js"); err != nil {
		t.Fatalf("compute --cites: %v", err)
	}
	k := lastBody(t, runDir, &recordpb.Computation{}).GetComputationId()
	run1 := abutting(t, readReport(t, runDir), "The sky is blue")
	if !strings.Contains(run1, anchor.Token(c)) || !strings.Contains(run1, anchor.Token(k)) {
		t.Fatalf("the citation and the computation do not stand at the sentence's end: %q\n%s", run1, readReport(t, runDir))
	}
	if _, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--quote", "The sky is blue"+run1+". The grass is green",
		"--new", "The grass is green"+run1, "--reason", "the sky claim cannot stand"); err != nil {
		t.Fatalf("the cut carrying both anchors onto the sentence that stays: %v", err)
	}
	retire := func(anchors string) error {
		_, err := run(t, "retire", "--run", runDir, "--seat-id", blueSeat, "--quote", "The sky is blue.", "--anchor", anchors, "--reason", "refuted")
		return err
	}
	if err := retire(c); err == nil || !strings.Contains(err.Error(), k) {
		t.Errorf("a claim retire taking %s while %s applies it = %v, want refused naming it", c, k, err)
	}
	if err := retire(c + "," + k); err != nil {
		t.Fatalf("a claim retire taking the citation and the computation applying it: %v", err)
	}
	if ev := lastBody(t, runDir, &recordpb.Retire{}); !slices.Contains(ev.GetAnchors(), c) || !slices.Contains(ev.GetAnchors(), k) {
		t.Errorf("the retire took %v, want both %s and %s", ev.GetAnchors(), c, k)
	}
}

// abutting is the run of anchor tokens standing right after words in md.
func abutting(t *testing.T, md, words string) string {
	t.Helper()
	at := strings.Index(md, words)
	if at < 0 {
		t.Fatalf("the report does not hold %q:\n%s", words, md)
	}
	rest := md[at+len(words):]
	return rest[:anchor.SkipRun(rest, 0)]
}

// THE RETIRED-ANCHOR REFUSALS ARE ASKED IN THE WRITE. A retire of the citation commits on another
// seat's connection at the instant a `compute --cites` of it, or a gap about it, opens its write: the
// flag checks have passed, and the write still refuses each.
func TestARetireRacingAnActOnItsAnchorRefusesTheAct(t *testing.T) {
	for _, act := range []string{"compute --cites", "mint --about-kind anchor"} {
		t.Run(act, func(t *testing.T) {
			runDir := citedRun(t, "https://sky/1")
			registerChairOnce(t, runDir)
			registerLensOnce(t, runDir)
			c := citeSentence(t, runDir, "The sky is blue.", "https://sky/1")
			retired := recordtest.At(t, "blue-lane-1", "race:retire", &recordpb.Retire{Claim: proto.String("The sky is blue."), Reason: proto.String("unreadable"),
				Anchors: []string{c}, ClaimKept: proto.Bool(true), RemovalBasis: proto.String(record.RemovalAsserted)})
			s := script(t, runDir, "wet.js", "console.log('wet');")
			err := runtest.Overtaken(t, otherSeat(t, runDir), []*recordpb.Event{retired}, func() error {
				var err error
				if act == "compute --cites" {
					_, err = run(t, "compute", "--run", runDir, "--seat-id", blueSeat, "--quote", "Water is wet.", "--dir", s, "--cites", c,
						"--reason", "the computation settles it", "--", "node", "wet.js")
				} else {
					_, err = mintWith(t, runDir, "--key", "race", "--about-kind", "anchor", "--about", c)
				}
				return err
			})
			if err == nil || !strings.Contains(err.Error(), "retired") {
				t.Errorf("%s racing the retire of %s = %v, want refused naming the retire", act, c, err)
			}
			if n := countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_COMPUTATION) + countEvents(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); act == "compute --cites" && n != 0 {
				t.Errorf("the refused compute recorded %d event(s)", n)
			}
		})
	}
}
