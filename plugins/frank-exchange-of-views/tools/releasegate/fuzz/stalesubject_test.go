package fuzz

import (
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// staleOnSubjectTries bounds the gaps on a gap the drive mints in a run.
const staleOnSubjectTries = 3

// staleOnASubject drives a Stale event once a run, deterministically, through the one path parallel
// lenses share a subject on (plans/findings-fold.md §III.B7; the stale gate, lead call 2026-11-27):
// this lens raises a gap G2 on another lens's open gap G, G's minter amends G while this lens's sitting
// is open, and this lens's regrade of G2 — read off the G its sitting opened on — is refused stale
// naming the amend, then recorded again (admitFresh). G2 is graded low on every axis, so it holds no
// PASS and asks nothing more of the run. An amend the record refuses (G docketed, or already amended in
// its minter's sitting) leaves G2 standing and the drive tries again, at most staleOnSubjectTries tries.
func (r *runner) staleOnASubject(seatID string) {
	r.mu.Lock()
	over := r.staleOnSubject || r.staleOnSubjectMints >= staleOnSubjectTries
	r.mu.Unlock()
	if over {
		return
	}
	b, ok := r.board()
	if !ok {
		return
	}
	open := r.openGaps()
	var g, minter string
	for _, e := range b.Events {
		m, ok := recordpb.BodyAs[*recordpb.Mint](e)
		if !ok || m.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_GAP || !strings.HasPrefix(e.GetSeatId(), "red-lens") ||
			e.GetSeatId() == seatID || !slices.Contains(open, m.GetGapId()) {
			continue
		}
		g, minter = m.GetGapId(), e.GetSeatId()
		break
	}
	if g == "" {
		return
	}
	// A TRY IS TAKEN UNDER THE LOCK BEFORE THE MINT: lenses sit concurrently, so a count read and
	// raised apart lets each sitting lens past the bound.
	r.mu.Lock()
	done := r.staleOnSubject || r.staleOnSubjectMints >= staleOnSubjectTries
	if !done {
		r.staleOnSubjectMints++
	}
	r.mu.Unlock()
	if done {
		return
	}
	out, err := r.exec(withJSON([]string{"mint", "--seat-id", seatID, "--key", "on-" + g, "--about-kind", "gap", "--about", g,
		"--class", "spec_underspecification", "--problem", "The gap's acceptance check names no observable outcome.",
		"--fix", "State a check that can fail.", "--check-kind", "document", "--check", "the check names an observable outcome",
		"--reason", "fuzz: the subject's check, raised on " + g,
		"--severity", "low", "--likelihood", "low", "--impact", "low", "--complexity", "low"})...)
	g2 := gapIDOf(out)
	if err != nil || g2 == "" {
		return
	}
	if _, err := r.exec("amend", "--seat-id", minter, "--id", g, "--acceptance-check", "fuzz: the check for "+g+", restated beside "+g2,
		"--reason", "amend-basis-for-"+g); err != nil {
		return
	}
	r.mu.Lock()
	r.staleOnSubject = true
	r.mu.Unlock()
	r.admitFresh(g2, "regrade", "--seat-id", seatID, "--id", g2, "--complexity", "medium", "--reason", "regrade-basis-for-"+g2)
}
