package fuzz

import (
	"strings"
	"sync/atomic"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE QUESTION'S DRIVE (record/questionpart.go). Blue splits the question the run was asked into
// question parts and records each part's coverage; the logic lens checks each coverage; a PASS over a
// part not covered derives QUALIFIED. A question with no part, a part with no coverage and a part with
// no standing check each hold the PASS, so a drive that skipped any of them would end every run
// UNVERIFIED and leave VERIFIED undriven.

// partCoverages is the round-robin the drive's coverage records draw from: covered twice in four, so a
// run ends VERIFIED as well as QUALIFIED, and each value not covered once. PACKAGE-LEVEL AND ATOMIC for
// the reason inquiryTick is: a run records a part or two, and a per-run counter would never reach the
// values past its first.
var partCoverages = []string{"covered", "not_obtained", "covered", "excluded"}

var partCoverageTick atomic.Int64

// nextPartCoverage is the coverage a drawn record states: the round-robin's, and covered in a run forced
// to a verdict, whose outcome is the one it is forced to.
func (r *runner) nextPartCoverage() string {
	if r.forcedVerdict() {
		return "covered"
	}
	return partCoverages[int(partCoverageTick.Add(1)-1)%len(partCoverages)]
}

// questionParts is the drive's read of the question's parts; a read that fails is kept on the runner,
// which fails the run on it, and ok is false so the caller leaves the question alone.
func (r *runner) questionParts() ([]record.QuestionPartRow, bool) {
	parts, err := record.QuestionParts(r.run())
	if err != nil {
		r.mu.Lock()
		if r.questionReadErr == "" {
			r.questionReadErr = err.Error()
		}
		r.mu.Unlock()
		return nil, false
	}
	return parts, true
}

// splitQuestion is blue's split of a question that has no part: one part, and a second one run in two
// not forced to a verdict. A seat whose work list owes the split (the frontier, blue-respond) drives it.
func (r *runner) splitQuestion(seatID string) {
	parts, ok := r.questionParts()
	if !ok || len(parts) > 0 {
		return
	}
	texts := []string{"whether the trend holds across the whole sampled corpus"}
	if !r.forcedVerdict() && r.coin(50) {
		texts = append(texts, "whether the trend outlasts the sample's period")
	}
	for _, text := range texts {
		r.do("question split", seatID).set("--text", text).admit()
	}
}

// coverQuestionParts is blue's coverage of every part with none, one record each this sitting: each is
// a blocking item on the synthesizer's and blue-respond's work lists. A sound part is sometimes covered
// again, which leaves its check owed again, the same as a new fate on an inquiry.
func (r *runner) coverQuestionParts(seatID string) {
	parts, ok := r.questionParts()
	if !ok {
		return
	}
	for _, p := range parts {
		switch {
		case p.State == record.PartOpen:
			r.do("question cover", seatID).set("--id", p.ID).set("--as", r.nextPartCoverage()).
				set("--reason", "fuzz: what in the report answers the part, or why it does not").admit()
		case p.State == record.PartSound && !r.forcedVerdict() && r.coin(10):
			r.do("question cover", seatID).set("--id", p.ID).set("--as", r.nextPartCoverage()).
				set("--reason", "fuzz: an edit moved the report's answer to the part").admit()
		}
	}
}

// The refusal classes of `question check` the drive holds the tool to, once a run each, with the code
// each carries.
const (
	partCheckNonOwner = "non_owner"
	partCheckOpen     = "open"
	partCheckStands   = "already_stands"
)

var partCheckCodes = map[string]feov.Code{
	partCheckNonOwner: feov.RoleViolation,
	partCheckOpen:     feov.Conflict,
	partCheckStands:   feov.Conflict,
}

// checkQuestionParts is a red lens's part in the question's judgement. The OWNER (record.QuestionPartOwner)
// checks every part with coverage and no standing check, the blocking item its work list carries for
// each, and drives each refusal class once a run; any other lens drives the owner's refusal once a run.
// A part whose coverage, or a gap about it, landed after the owner's sitting opened is left for the
// sitting it is readied for next: the tool refuses that check.
func (r *runner) checkQuestionParts(seatID string) {
	parts, ok := r.questionParts()
	if !ok || len(parts) == 0 {
		return
	}
	if seatID != record.QuestionPartOwner() {
		r.partCheckRefused(partCheckNonOwner, seatID, parts[r.rng.Intn(len(parts))].ID)
		return
	}
	b, ok := r.board()
	if !ok {
		return
	}
	sitting := b.At.LatestSittingOf(seatID)
	for _, p := range parts {
		switch p.State {
		case record.PartOpen:
			r.partCheckRefused(partCheckOpen, seatID, p.ID)
		case record.PartSound:
			r.partCheckRefused(partCheckStands, seatID, p.ID)
		case record.PartUnchecked:
			if !p.CheckableFrom(sitting) {
				continue
			}
			args := []string{"question", "check", "--seat-id", seatID, "--id", p.ID}
			out, err := r.exec(args...)
			if err == nil {
				continue
			}
			// A COVERAGE OR A MINT ABOUT IT LANDED BESIDE THE CHECK, from a seat sitting alongside: the
			// refusal is the record's answer to a judgement of what no longer stands.
			if now, ok := r.questionPart(p.ID); ok && (now.State != record.PartUnchecked || now.Coverage != p.Coverage || !now.CheckableFrom(sitting)) {
				noteEither(args, "the question part moved under the owner's check", false)
				continue
			}
			r.noteActErr("refused, and the drive expects it admitted: " + firstLine(out) + ": " + err.Error())
		}
	}
}

// questionPart is one part as the record now reads it.
func (r *runner) questionPart(id string) (record.QuestionPartRow, bool) {
	parts, ok := r.questionParts()
	if !ok {
		return record.QuestionPartRow{}, false
	}
	for _, p := range parts {
		if p.ID == id {
			return p, true
		}
	}
	return record.QuestionPartRow{}, false
}

// partCheckRefused drives one refusal class of `question check` on part id, once a run, and holds the
// tool to refusing it with the class's code. AN ADMISSION IS A DEFECT unless the part moved out of the
// class's state between the read and the check: blue covering an open one.
func (r *runner) partCheckRefused(class, seatID, id string) {
	r.mu.Lock()
	driven := r.questionCheckRefused[class]
	r.mu.Unlock()
	if driven {
		return
	}
	args := withJSON([]string{"question", "check", "--seat-id", seatID, "--id", id})
	out, err := r.exec(args...)
	if err == nil {
		if now, ok := r.questionPart(id); ok && class == partCheckOpen && now.State != record.PartOpen {
			noteEither(args, "the question part left the open state under the check", true)
			return
		}
		r.noteActErr(strings.Join(args, " ") + ": admitted, and the drive drives it to be refused (" + class + "): " + firstLine(out))
		return
	}
	env, ok := refusalEnvelopeIn(out)
	if !ok || env.Code != string(partCheckCodes[class]) {
		r.noteActErr(strings.Join(args, " ") + ": refused (" + class + ") without the code " + string(partCheckCodes[class]) + ": " + firstLine(out))
		return
	}
	r.mu.Lock()
	if r.questionCheckRefused == nil {
		r.questionCheckRefused = map[string]bool{}
	}
	r.questionCheckRefused[class] = true
	r.mu.Unlock()
}

// someQuestionPart is one part of the question as the record holds it, "" where the question has none.
func (r *runner) someQuestionPart() string {
	parts, ok := r.questionParts()
	if !ok || len(parts) == 0 {
		return ""
	}
	return parts[r.rng.Intn(len(parts))].ID
}

// coverageMisstatedClass is the gap class of a question part's coverage the report does not bear out:
// always material, and minted only about the question.
var coverageMisstatedClass = recordpb.Word(recordpb.GapClass_GAP_CLASS_COVERAGE_MISSTATED)

// stampWord is the outcome a PASS derives, which the assembly stamps: its prompt names both and says
// the record derives which, so the drive reads it off the record as the seat reads the parts. A read
// that fails stamps VERIFIED, and the outcome's own refusal says the record derives otherwise.
func (r *runner) stampWord() string {
	got, _, ok, err := record.DeriveVerdict(r.run())
	if err != nil || !ok {
		r.noteActErr("the assembly could not read the outcome a PASS derives: " + deriveFailure(err))
		return "VERIFIED"
	}
	return got
}

func deriveFailure(err error) string {
	if err != nil {
		return err.Error()
	}
	return "the record derives no outcome"
}
