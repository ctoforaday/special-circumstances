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
