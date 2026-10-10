package record

import (
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A RULER DOES NOT FILE THE SUBJECT IT RULES.
//
// A gavel rests with one seat (gavelSeatOf), so a motion that seat filed has no ruler but its
// filer: its ruling is the filer answering itself, and unruled it holds the PASS gate with no
// other seat able to clear it. Refusing the RULING instead leaves exactly that motion on the
// record — one nobody may rule — so the refusal is at the filing, where nothing is yet owed.
//
// ONE PREDICATE, READ FROM THE GAVEL. HoldsGavel is the schema's `ruled_by` and nothing else. The
// command tree mounts a subject's `rule` where it is true and its `file` where it is false
// (cli/motion subject), and the write refuses a filing where it is true (requireFilerHoldsNoGavel)
// — so a subject added to the enum is ruled by one role and filed by the others without a line
// here, and no surface can hand one role both verbs.

// HoldsGavel reports whether a seat of this role rules motions on the subject. A subject that
// declares no ruler is held by nobody: TestEveryMotionSubjectNamesItsRuler refuses that at the
// descriptor, and the command tree panics on it at construction.
func HoldsGavel(role string, subj recordpb.MotionSubject) bool {
	ruler, err := recordpb.SubjectRuler(subj)
	return err == nil && role != "" && role == ruler
}

// requireFilerHoldsNoGavel refuses a filing on a subject by the seat that rules it. proposal marks
// the one subject whose filing is not a motion body: an avenue's is its proposal.
//
// Migrate replays what an archive holds, and an archive holds such filings: the docket pair
// migrate synthesises for a record that predates docket motions is filed and ruled by the bench
// (migrate's opinionEntry). Refusing them would drop real acts and call the loss a translation.
func requireFilerHoldsNoGavel(seatID string, subj recordpb.MotionSubject, proposal bool) error {
	if Migrating || !HoldsGavel(roleOfSeat(seatID), subj) {
		return nil
	}
	return RefuseRulerFiling(subj, seatID, proposal)
}

// RefuseRulerFiling is the refusal a seat meets when it files on a subject it rules — one text for the
// surface that does not carry the verb and for the write that refuses the body. It names both
// parties, as the refusal of a `rule` that is not the seat's does: who holds the gavel, and who
// asked.
func RefuseRulerFiling(subj recordpb.MotionSubject, seatID string, proposal bool) error {
	word := recordpb.Word(subj)
	ruler, _ := recordpb.SubjectRuler(subj)
	filing, who := "`file`", "The "+strings.Join(filingRoles(subj), ", ")+" seats file one."
	if i := strings.LastIndex(who, ", "); i >= 0 {
		who = who[:i] + " and" + who[i+1:]
	}
	if proposal {
		filing, who = "the proposal that files one", "An avenue is blue's to propose."
	}
	msg := Article(word) + " " + word + " motion is ruled by the " + ruler + " seat, so " + filing + " is not on your surface; you are " +
		seatID + " (" + roleOfSeat(seatID) + "). The seat that rules a subject files nothing on it, because a motion its own ruler " +
		"filed is one no other seat can rule. " + who
	if to := rulerTurnsTo[subj]; to != "" {
		msg += " " + to
	}
	return feov.Errorf(feov.RoleViolation, "%s", msg)
}

// rulerTurnsTo is what a subject's ruler does about the thing a filing would have raised, where it
// holds an act for it. No entry is a statement too: the bench originates nothing on a docket, and
// rules the gaps another seat puts before it.
var rulerTurnsTo = map[recordpb.MotionSubject]string{
	recordpb.MotionSubject_MOTION_SUBJECT_GRADE:    "To contest a gap's grade yourself, file `motion docket file` on the gap, which the bench rules.",
	recordpb.MotionSubject_MOTION_SUBJECT_PETITION: "Where the run itself must stop, `halt` is yours.",
}

// filingRoles is the roles that file a motion on the subject: every role a seat sits in, less the
// one that rules it. The operator sits in no seat and files nothing.
func filingRoles(subj recordpb.MotionSubject) []string {
	var out []string
	for role := range roleSeats {
		if role != OperatorRole && !HoldsGavel(role, subj) {
			out = append(out, role)
		}
	}
	sort.Strings(out)
	return out
}

// Article is the indefinite article for a subject word — "an avenue", "a grade".
func Article(word string) string {
	if word != "" && strings.ContainsRune("aeiou", rune(word[0])) {
		return "an"
	}
	return "a"
}
