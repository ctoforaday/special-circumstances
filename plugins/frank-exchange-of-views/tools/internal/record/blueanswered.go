package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// DID BLUE ANSWER THIS GAP, AND CAN THE RECORD SAY? (#1122)
//
// THE ANSWERING PREDICATE IS THE ONE EVERY OTHER READER USES, and picking a different one was the
// near miss here. There are two joins between an edit and a gap and they are not the same fact:
//
//   - `gap_edit` (WorkGapState.Edits) is the LOCATION-DRIFT history — edits that moved the sentence
//     a gap points at, whether or not they claimed to answer it. It exists so red can see what
//     changed underneath its pointer.
//   - `BlueEdit.answers` is blue's CLAIM to have answered, validated at the write (requireGap
//     refuses an unknown gap), and with `old != new` it is what `exchangesOf` counts as movement
//     and what ManifestOwed counts as answered.
//
// Reading the first would have invented a third definition of "answered", disagreeing with the
// impasse fold and the manifest audit, and would additionally have produced a false accusation on
// any gap anchored by --about: those have no location, so they have no drift history, and "nobody
// edited the sentence" is not "nobody answered".
//
// THE NOT-MEASURED ARM IS THE POINT OF THE WHOLE FUNCTION. "Blue has not answered" and "blue is
// still sitting and may yet answer" are the same silence on the record, and a lens handed the first
// where the second is true files a closing argument about a dispute that is still in progress.
// BlueSitting.Unresolved is that fact, and it is read here rather than re-derived.
func blueAnswers(evs []*Event, gaps []WorkGapState, seatID string) []string {
	minted := mintedBy(evs)
	var mine []string
	for _, g := range gaps {
		if g.Open && minted[g.ID] == seatID {
			mine = append(mine, g.ID)
		}
	}
	if len(mine) == 0 {
		return nil
	}
	// WhileRunning, because this list is read by a seat that is sitting: a later reading is not the
	// question, and the latest blue sitting is exactly the one that may still be answering.
	answered, engaged, pending := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, s := range BlueSittings(evs, WhileRunning) {
		for _, g := range s.Engaged {
			engaged[g] = true
			if s.Unresolved {
				pending[g] = true
			}
		}
		for _, e := range s.Acts {
			be, ok := recordpb.BodyAs[*recordpb.BlueEdit](e)
			if !ok {
				continue
			}
			// old != new is the same substantive-movement test exchangesOf applies: an edit that
			// replaced a span with itself is an act and not an answer.
			if g := be.GetAnswers(); g != "" && be.GetOld() != be.GetNew() {
				answered[g] = true
			}
		}
	}
	var out []string
	for _, g := range mine {
		switch {
		case answered[g]:
			out = append(out, fmt.Sprintf("gap %s has been ANSWERED by an edit — re-audit it against the acceptance check you set at mint, and either close it or state what your own check still fails", g))
		case pending[g]:
			out = append(out, fmt.Sprintf("gap %s: whether blue has answered it is NOT MEASURED — a blue sitting engaged on it is open on the record, so nothing here is a report that no answer came", g))
		case engaged[g]:
			out = append(out, fmt.Sprintf("gap %s is open and NO edit answers it — blue has been engaged on it and its sitting closed without one, which is a fact to state rather than a gap to re-argue", g))
		}
		// A GAP BLUE HAS NEVER BEEN ENGAGED ON GETS NOTHING, and the silence is correct: it is
		// waiting for a dispatch, which is the chair's business and not news about blue.
	}
	return out
}
