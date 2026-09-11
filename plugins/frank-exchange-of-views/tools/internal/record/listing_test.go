package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A LISTING SHOWS EVERY ACT, THE STRUCK ONES MARKED, IN THE PLACE THE SEAT MADE THEM: a chain of two
// corrections lists the first wording, the second, then the one that stands — before a later act
// the seat made in between, because the correction took the first act's place.
func TestListingMarksAStruckChainInItsPlace(t *testing.T) {
	corr := func(k, r string) *Event {
		return recordtest.At(t, "s", "s:correction:"+k, &recordpb.Correction{
			Corrects: proto.String(k), Replacement: proto.String(r), Why: proto.String("a word was lost")})
	}
	evs := []*Event{
		recordtest.At(t, "s", "s:log:#1", seatLog("one")),
		recordtest.At(t, "s", "s:log:#2", seatLog("later act")),
		recordtest.At(t, "s", "s:log:#1~1", seatLog("one, corrected")),
		corr("s:log:#1", "s:log:#1~1"),
		recordtest.At(t, "s", "s:log:#1~2", seatLog("one, corrected again")),
		corr("s:log:#1~1", "s:log:#1~2"),
	}
	var got []string
	for _, l := range Listing(evs) {
		if lg, ok := recordpb.BodyAs[*recordpb.Log](l.Event); ok {
			got = append(got, l.Markdown(lg.GetText()))
		}
	}
	want := []string{
		"~~one~~ (struck by s: a word was lost)",
		"~~one, corrected~~ (struck by s: a word was lost)",
		"one, corrected again",
		"later act",
	}
	if strings.Join(got, " | ") != strings.Join(want, " | ") {
		t.Errorf("Listing = %q\nwant      %q", got, want)
	}
	// With no correction on the stream, a listing is the stream.
	if n := len(Listing(evs[:2])); n != 2 {
		t.Errorf("a stream with no correction listed %d acts, want 2", n)
	}
}

// A NARROWED READ KEEPS ITS CORRECTIONS, or Live and Listing have nothing to strike with.
func TestEventsOfACorrectableTypeCarriesItsCorrections(t *testing.T) {
	run := corrRun(t)
	blue := sit(t, run, "blue-respond")
	k := mustAppend(t, blue, &recordpb.Position{Text: proto.String("the board is  going in")}).GetKey()
	if _, err := Append(correcting(blue, recordpb.EventType_EVENT_TYPE_POSITION, k, "a word was lost"),
		&recordpb.Position{Text: proto.String("the board is clean going in")}); err != nil {
		t.Fatal(err)
	}
	evs, err := EventsOf(run, recordpb.EventType_EVENT_TYPE_POSITION)
	if err != nil {
		t.Fatal(err)
	}
	live := Live(evs)
	var texts []string
	for _, e := range live {
		if p, ok := recordpb.BodyAs[*recordpb.Position](e); ok {
			texts = append(texts, p.GetText())
		}
	}
	if len(texts) != 1 || texts[0] != "the board is clean going in" {
		t.Errorf("the positions standing on a narrowed read = %q, want only the corrected one", texts)
	}
}

func correctionOf(t *testing.T, seat, key string) *Event {
	return recordtest.At(t, seat, seat+":correction:"+key, &recordpb.Correction{
		Corrects: proto.String(key), Replacement: proto.String(key + "~1"), Why: proto.String("a word was lost")})
}

// THE JSON LISTINGS CARRY A STRUCK ACT AS A FIELD, not as struck-through prose a reader would have
// to parse back out: the log entry is listed with its `struck`, and not counted; the debate's
// arrays hold what stands and the epoch's `struck` holds what was corrected.
func TestJSONListingsCarryTheStruckActAsAField(t *testing.T) {
	pos := func(s string) *recordpb.Position { return &recordpb.Position{Text: proto.String(s)} }
	evs := []*Event{
		recordtest.At(t, "blue-respond", "blue-respond:log:#1", seatLog("the tool  refused")),
		recordtest.At(t, "blue-respond", "blue-respond:position:#1", pos("the report is  now")),
		recordtest.At(t, "blue-respond", "blue-respond:log:#1~1", seatLog("the tool refused the cite")),
		correctionOf(t, "blue-respond", "blue-respond:log:#1"),
		recordtest.At(t, "blue-respond", "blue-respond:position:#1~1", pos("the report is sound now")),
		correctionOf(t, "blue-respond", "blue-respond:position:#1"),
	}
	lj := LogJSONOf(evs)
	if len(lj.Log) != 2 || lj.Log[0].Struck == nil || lj.Log[0].Text != "the tool  refused" ||
		lj.Log[0].Struck.Replacement != "blue-respond:log:#1~1" || lj.Log[1].Struck != nil {
		t.Errorf("log listing = %+v, want the struck entry marked with its replacement, then the one that stands", lj.Log)
	}
	if lj.Counts.Total != 1 {
		t.Errorf("log total = %d, want 1 — a corrected entry is one entry", lj.Counts.Total)
	}
	dj := DebateJSONOfEvents(evs)
	if len(dj.Epochs) != 1 {
		t.Fatalf("%d epochs", len(dj.Epochs))
	}
	ep := dj.Epochs[0]
	if strings.Join(ep.Blue, "|") != "the report is sound now" {
		t.Errorf("blue = %q, want only the position that stands", ep.Blue)
	}
	if len(ep.Struck) != 1 || ep.Struck[0].Text != "the report is  now" || ep.Struck[0].By != "blue-respond" || ep.Struck[0].Type != "position" {
		t.Errorf("struck = %+v, want the corrected position, with who struck it", ep.Struck)
	}
}

// THE FIRST-WINS ANSWER IS THE CORRECTED ONE. A ruling its ruler corrected in the sitting is read as
// its replacement, in its place — not as the struck ruling first-wins would otherwise quote, and not
// as a second ruling.
func TestMotionsReadACorrectedRulingAsTheOneThatStands(t *testing.T) {
	ruling := func(opinion string) *recordpb.MotionRule {
		return &recordpb.MotionRule{MotionId: proto.String("M1"), Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
			Opinion: proto.String(opinion),
			Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
				Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_CARRIED), Principle: proto.String("p"),
				Tension: proto.String("t"), ReviewFlag: proto.String("none"), Settled: proto.String("s"), ReopensOn: proto.String("r")}}}
	}
	k := "judge:motion_rule:#1"
	evs := []*Event{
		recordtest.At(t, "red-chair", "red-chair:motion:#1", &recordpb.Motion{MotionId: proto.String("M1"),
			Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET), Basis: proto.String("red cannot settle G1"),
			Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}}}),
		recordtest.At(t, "judge", k, ruling("because  refuses")),
		recordtest.At(t, "judge", k+"~1", ruling("because the gate refuses")),
		correctionOf(t, "judge", k),
	}
	ms := MotionsOf(evs)
	if len(ms) != 1 || ms[0].Opinion != "because the gate refuses" {
		var got []string
		for _, m := range ms {
			got = append(got, m.ID+": "+m.Opinion)
		}
		t.Errorf("motions = %q, want M1 answered by the corrected ruling", got)
	}
}

// THE STRUCK TEXT IS TRIMMED: a closing ~~ after whitespace does not close in CommonMark, so
// `~~text ~~` would render literal tildes and no strike at all.
func TestStruckTextIsTrimmedSoTheStrikeRenders(t *testing.T) {
	s := &Struck{By: "judge", Why: "a word was lost"}
	if got := StruckMarkdown("G2 is reproducible via ", s); got != "~~G2 is reproducible via~~ (struck by judge: a word was lost)" {
		t.Errorf("StruckMarkdown = %q", got)
	}
	if got := (Listed{Struck: s}).Strike("  a header  "); got != "~~a header~~" {
		t.Errorf("Strike = %q", got)
	}
	if got := StruckMarkdown(" as written ", nil); got != " as written " {
		t.Errorf("an act that stands must render as written, untouched: %q", got)
	}
}

// THE LINES OF INQUIRY AND THE EVIDENCE LIST A STRUCK ACT, never drop it.
func TestInquiryAndEvidenceListingsKeepTheStruckAct(t *testing.T) {
	prop := func(line string) *recordpb.Avenue {
		return &recordpb.Avenue{AvenueId: proto.String("Q1"), Line: proto.String(line), Status: recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED.Enum()}
	}
	rerun := func(note string) *recordpb.Reproduce {
		return &recordpb.Reproduce{ProofSha: proto.String("abc"), Reproduced: proto.Bool(true),
			Soundness: recordpb.Soundness_SOUNDNESS_SOUND.Enum(), Note: proto.String(note)}
	}
	evs := []*Event{
		recordtest.At(t, "blue-respond", "blue-respond:avenue:#1", prop("try the  method")),
		recordtest.At(t, "blue-respond", "blue-respond:avenue:#1~1", prop("try the recorded method")),
		correctionOf(t, "blue-respond", "blue-respond:avenue:#1"),
		recordtest.At(t, "blue-respond", "blue-respond:proof:#1", &recordpb.Proof{ProofId: proto.String("p-1"), ProofSha: proto.String("abc")}),
		recordtest.At(t, "red-lens-evidence", "red-lens-evidence:reproduce:#1", rerun("it  computes")),
		recordtest.At(t, "red-lens-evidence", "red-lens-evidence:reproduce:#1~1", rerun("it computes primality")),
		correctionOf(t, "red-lens-evidence", "red-lens-evidence:reproduce:#1"),
	}
	if got := StruckInquiryTexts(evs)["Q1"]; len(got) != 1 || got[0].Text != "try the  method" || got[0].By != "blue-respond" {
		t.Errorf("the struck wording of Q1 = %+v, want the corrected proposal's", got)
	}
	ej := EvidenceJSONOf(evs)
	if len(ej.Proofs) != 1 {
		t.Fatalf("%d proofs", len(ej.Proofs))
	}
	p := ej.Proofs[0]
	if p.Verified == nil || p.Verified.Note != "it computes primality" {
		t.Errorf("verified = %+v, want the re-run that stands", p.Verified)
	}
	if len(p.StruckReruns) != 1 || p.StruckReruns[0].Note != "it  computes" || p.StruckReruns[0].Struck == nil {
		t.Errorf("struck re-runs = %+v, want the corrected one, marked", p.StruckReruns)
	}
}

// ON A REAL RECORD: the gap's grade history and the board list a corrected regrade struck; the
// gap's standing regrades and its current grade are the replacement's; the proofs read back list
// a corrected re-run struck.
func TestRegradeHistoryAndProofsListTheStruckAct(t *testing.T) {
	runDir := newRun(t)
	run := mustRun(t, runDir)
	red := sit(t, run, "red-chair")
	mustAppend(t, red, corrMint("G1"))
	regrade := func(sev recordpb.Grade, basis string) *recordpb.Regrade {
		return &recordpb.Regrade{GapId: proto.String("G1"), Severity: sev.Enum(), Basis: proto.String(basis)}
	}
	k := mustAppend(t, red, regrade(recordpb.Grade_GRADE_LOW, "the consequence is  bounded")).GetKey()
	if _, err := Append(correcting(red, recordpb.EventType_EVENT_TYPE_REGRADE, k, "a word was lost"),
		regrade(recordpb.Grade_GRADE_HIGH, "the consequence reaches every caller")); err != nil {
		t.Fatal(err)
	}
	gaps, err := GapStates(run)
	if err != nil || len(gaps) != 1 {
		t.Fatalf("gap states: %v %d", err, len(gaps))
	}
	g := gaps[0]
	if len(g.Regrades) != 1 || g.Severity != recordpb.Grade_GRADE_HIGH {
		t.Errorf("standing regrades %d, severity %v — want the replacement alone, and its grade", len(g.Regrades), g.Severity)
	}
	h := g.RegradeListing()
	if len(h) != 2 || h[0].Struck == nil || h[0].Regrade.GetBasis() != "the consequence is  bounded" || h[1].Struck != nil {
		t.Errorf("regrade history = %+v, want the struck regrade, marked, then the one that stands", h)
	}
	bj, err := BoardJSONOfRun(run)
	if err != nil {
		t.Fatal(err)
	}
	var regrades []map[string]any
	for _, gj := range append(bj.Open, bj.Closed...) {
		regrades = append(regrades, gj.Regrades...)
	}
	if len(regrades) != 2 || regrades[0]["struck"] == nil || regrades[1]["struck"] != nil {
		t.Errorf("board regrades = %v, want two, the first marked struck", regrades)
	}

	rerun := func(note string) *recordpb.Reproduce {
		return &recordpb.Reproduce{ProofSha: proto.String("abc"), Reproduced: proto.Bool(true),
			Soundness: recordpb.Soundness_SOUNDNESS_SOUND.Enum(), Note: proto.String(note)}
	}
	recordtest.Seed(t, runDir,
		recordtest.At(t, "blue-respond", "blue-respond:proof:#1", &recordpb.Proof{ProofId: proto.String("p-1"), ProofSha: proto.String("abc"), ProofBasis: proto.String("reproducible")}),
		recordtest.At(t, "red-lens-evidence", "red-lens-evidence:reproduce:#1", rerun("it  computes")),
		recordtest.At(t, "red-lens-evidence", "red-lens-evidence:reproduce:#1~1", rerun("it computes primality")),
		correctionOf(t, "red-lens-evidence", "red-lens-evidence:reproduce:#1"),
	)
	proofs, err := RecordedProofs(run)
	if err != nil || len(proofs) != 1 {
		t.Fatalf("proofs: %v %d", err, len(proofs))
	}
	if proofs[0].Verified == nil || proofs[0].Verified.Note != "it computes primality" ||
		len(proofs[0].StruckReruns) != 1 || proofs[0].StruckReruns[0].Struck == nil {
		t.Errorf("proof = verified %+v struck %+v — want the standing re-run, and the struck one listed", proofs[0].Verified, proofs[0].StruckReruns)
	}
}
