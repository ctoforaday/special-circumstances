package capture

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// LensTurnAudit names every dispatch that engaged a lens on a gap before its turn had come
// (record.EarlyLensDispatches): blue had recorded no act on the gap since the lens's own last act
// there, so the lens, sitting before blue in its epoch, audited an answer that did not exist yet. The
// plan never makes one (ruled by gblock 2026-10-30), so a row FAILS: the dispatch broke its rule, or
// the run was dispatched by a binary that did not hold it.
func LensTurnAudit(run record.Run) Audit {
	early, err := record.EarlyLensDispatches(run)
	if err != nil {
		return Audit{Check: "lens-turn", Verdict: "SKIP", Detail: "the record could not be read: " + err.Error()}
	}
	if len(early) == 0 {
		return Audit{Check: "lens-turn", Verdict: "PASS", Detail: "no dispatch engaged a lens on a gap before blue had acted on it since the lens's last act there"}
	}
	named := make([]string, len(early))
	for i, e := range early {
		named[i] = fmt.Sprintf("%s on %s (dispatch %d)", e.SeatID, e.GapID, e.At)
	}
	return Audit{Check: "lens-turn", Verdict: "FAIL", Detail: fmt.Sprintf(
		"%d lens engagement(s) on a gap blue had not acted on since the lens's last act there: %s. Each lens sat before blue and audited an answer that did not exist yet",
		len(early), strings.Join(named, "; "))}
}

// DocketOrderAudit names every docket motion the dispatch filed before the gap's minting lens had its
// re-audit (record.EarlyDockets). The plan readies the lens alone there and files after its sitting
// (ruled by gblock 2026-10-31), so a row FAILS: the dispatch broke its rule, or the run was dispatched
// by a binary that did not hold it.
func DocketOrderAudit(run record.Run) Audit {
	early, err := record.EarlyDockets(run)
	if err != nil {
		return Audit{Check: "docket-order", Verdict: "SKIP", Detail: "the record could not be read: " + err.Error()}
	}
	if len(early) == 0 {
		return Audit{Check: "docket-order", Verdict: "PASS", Detail: "no docket motion of the dispatch's came before the gap's minting lens had its re-audit"}
	}
	named := make([]string, len(early))
	for i, e := range early {
		named[i] = fmt.Sprintf("%s on %s (event %d)", e.Motion, e.GapID, e.At)
	}
	return Audit{Check: "docket-order", Verdict: "FAIL", Detail: fmt.Sprintf(
		"%d docket motion(s) filed with the gap's minting lens's audit still to come: %s. The bench was asked before the lens had read blue's answer",
		len(early), strings.Join(named, "; "))}
}
