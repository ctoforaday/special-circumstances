package fuzz

import (
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// A MINT SUPERSEDING AN OPEN GAP CLOSES IT (ruled by gblock 2026-11-04). The drive supersedes only
// what the record admits — a gap the minting lens holds open and may close — and holds the record to
// the closing: the ancestor leaves the open board, closed, naming the successor.

// holdClosing marks gap id as one whose close the drive is driving (on), or clears it: a mint drawn
// in between does not supersede it, since the supersession would close it before that close.
func (r *runner) holdClosing(id string, on bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closingNow == nil {
		r.closingNow = map[string]bool{}
	}
	if on {
		r.closingNow[id] = true
	} else {
		delete(r.closingNow, id)
	}
}

// supersedable is each open gap seatID may supersede with a successor minted --check-kind kind, in
// board order: one its work list says is its to close — it minted it, and no docket motion stands on
// it — no close of which is being driven, and, where it awaits a live answer, whose check kind is kind:
// the demand passes to the successor (ruled by gblock 2026-11-05). owing is the gaps left out for that
// last reason alone, the ones refuseDroppedDemand drives the refusal on; carries is each gap in out
// that awaits a live answer, whose successor then does.
func (r *runner) supersedable(seatID, kind string) (out, owing []string, carries map[string]bool) {
	w, err := record.WorkOfSeat(r.run(), "lens", seatID)
	if err != nil {
		r.noteActErr(seatID + "'s work list could not be read to draw a gap to supersede: " + err.Error())
		return nil, nil, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	carries = map[string]bool{}
	for _, g := range w.Open {
		switch {
		case !g.YoursToClose || r.closingNow[g.ID]:
		case g.AwaitingEvidence && g.CheckKind != kind:
			owing = append(owing, g.ID)
		default:
			out = append(out, g.ID)
			carries[g.ID] = g.AwaitingEvidence
		}
	}
	return out, owing, carries
}

// refuseDroppedDemand drives seatID's mint of a successor at check kind kind over ancestor, an open
// gap of its own that awaits a live answer at another check kind: the record refuses it naming the
// ancestor's demand, and the refusal is counted only when it says so.
func (r *runner) refuseDroppedDemand(seatID, ancestor, kind string) {
	out, err := r.exec("--json", "mint", "--seat-id", seatID, "--quote", "A § fuzz sentence to anchor findings.",
		"--class", mintPoolClasses[0], "--problem", "fuzz: a successor that drops its ancestor's demand", "--check-kind", kind,
		"--check", "acc", "--fix", dirIgnore, "--severity", "low", "--likelihood", "low", "--impact", "low",
		"--reason", "fuzz: the ancestor awaits a live answer of another kind", "--supersedes", ancestor)
	env, ok := refusalEnvelopeIn(out)
	switch {
	case err == nil:
		r.noteActErr(seatID + "'s --check-kind " + kind + " successor of " + ancestor + ", which awaits a live answer of another kind, was admitted: " + firstLine(out))
	case !ok || env.Code != string(feov.Conflict) || !strings.Contains(env.Error, "--supersedes names "+ancestor+", an open gap minted --check-kind "):
		r.noteActErr(seatID + "'s --check-kind " + kind + " successor of " + ancestor + " was refused, and not by the demand it carries: " + firstLine(out))
	default:
		r.mu.Lock()
		r.supersedeDemandRefused++
		r.mu.Unlock()
	}
}

// supersessionClosed holds the record to a landed mint of successor that superseded ancestor, open
// when it was drawn: the board holds the ancestor closed, and names successor as the gap whose mint
// closed it. Where the ancestor awaited a live answer (carries), the successor carries its check kind,
// and with it the demand.
func (r *runner) supersessionClosed(ancestor, successor string, carries bool) {
	if ancestor == "" {
		return
	}
	bj, err := record.BoardJSONOfRun(r.run())
	if err != nil {
		r.noteActErr("the board could not be read after a mint superseding " + ancestor + ": " + err.Error())
		return
	}
	for _, g := range bj.Closed {
		if g.ID == ancestor {
			if g.SupersededBy != successor {
				r.noteActErr("gap " + ancestor + " closed, and the board names " + g.SupersededBy + " as the successor whose mint closed it, not " + successor + ", which superseded it open")
			}
			if carries {
				r.demandCarried(bj, ancestor, g.CheckKind, successor)
			}
			return
		}
	}
	r.noteActErr("gap " + ancestor + " is still open after " + successor + " was minted superseding it: a mint superseding an open gap closes it")
}

// demandCarried holds the board to successor carrying ancestor's check kind, which awaited a live
// answer when the mint closed it, and counts the demand carried.
func (r *runner) demandCarried(bj record.BoardJSON, ancestor, kind, successor string) {
	for _, gs := range [][]record.GapJSON{bj.Open, bj.Closed} {
		for _, g := range gs {
			if g.ID != successor {
				continue
			}
			if g.CheckKind != kind {
				r.noteActErr("gap " + successor + " superseded " + ancestor + ", which awaited a live answer under --check-kind " + kind + ", and carries --check-kind " + g.CheckKind)
				return
			}
			r.mu.Lock()
			r.supersedeDemandCarried++
			r.mu.Unlock()
			return
		}
	}
	r.noteActErr("gap " + successor + ", the successor of " + ancestor + ", is not on the board")
}
