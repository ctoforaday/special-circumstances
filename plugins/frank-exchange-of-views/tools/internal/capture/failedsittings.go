package capture

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// FailedSittingsAudit names every sitting that closed with no admitted envelope (record.FailedSittingsOf):
// a seat whose work arrives as items returned without the call that ends its sitting being admitted —
// its envelope refused until the harness gave up, its agent died, or a delivery that could not advance
// stopped it. Beside each it lists the sitting's tool defects (record.ToolDefectsIn), or says it has
// none.
//
// A failed sitting WARNS rather than fails. The engine dispatches the seat again while its items are
// owed, so the run goes on; what the reader needs is which sittings did not finish, because what each
// owed was carried to a later sitting or never done.
//
// A TOOL DEFECT IS LISTED, NEVER NAMED THE FAILURE'S REASON: cite, compute and reproduce log defects
// that fail no sitting, so a defect beside a failure is not its cause.
func FailedSittingsAudit(run record.Run) Audit {
	m, err := record.MergedEvents(run)
	if err != nil {
		return Audit{Check: "failed-sittings", Verdict: "SKIP", Detail: "the record could not be read: " + err.Error()}
	}
	failed := record.FailedSittingsOf(m.At)
	if len(failed) == 0 {
		return Audit{Check: "failed-sittings", Verdict: "PASS", Detail: "every closed sitting of a seat whose work arrives as items ended on an admitted envelope"}
	}
	named := make([]string, len(failed))
	for i, f := range failed {
		defects := "no tool defect in it"
		if d := record.ToolDefectsIn(m.Events, m.At, f.SittingID); len(d) > 0 {
			defects = fmt.Sprintf("tool defects (%d) in it: %s", len(d), strings.Join(quoted(d), ", "))
		}
		named[i] = fmt.Sprintf("%s sitting %d (epoch %d, agent %s; %s)", f.SeatID, f.Sitting, f.Epoch, f.AgentID, defects)
	}
	return Audit{Check: "failed-sittings", Verdict: "WARN", Detail: fmt.Sprintf(
		"%d sitting(s) closed with no admitted envelope: %s. Each seat returned without ending its sitting; read what it owed as carried to its next sitting, or never done",
		len(failed), strings.Join(named, "; "))}
}

// quoted is each text Go-quoted, so a defect holding a separator stays one item.
func quoted(texts []string) []string {
	out := make([]string, len(texts))
	for i, t := range texts {
		out[i] = strconv.Quote(t)
	}
	return out
}
