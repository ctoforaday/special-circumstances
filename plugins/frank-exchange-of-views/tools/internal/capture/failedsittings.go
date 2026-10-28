package capture

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// FailedSittingsAudit names every sitting that closed with no admitted envelope (record.FailedSittings):
// a seat whose work arrives as items returned without the call that ends its sitting being admitted —
// its envelope refused until the harness gave up, or its agent died.
//
// A failed sitting WARNS rather than fails. The engine dispatches the seat again while its items are
// owed, so the run goes on; what the reader needs is which sittings did not finish, because what each
// owed was carried to a later sitting or never done.
func FailedSittingsAudit(run record.Run) Audit {
	failed, err := record.FailedSittings(run)
	if err != nil {
		return Audit{Check: "failed-sittings", Verdict: "SKIP", Detail: "the record could not be read: " + err.Error()}
	}
	if len(failed) == 0 {
		return Audit{Check: "failed-sittings", Verdict: "PASS", Detail: "every closed sitting of a seat whose work arrives as items ended on an admitted envelope"}
	}
	named := make([]string, len(failed))
	for i, f := range failed {
		named[i] = fmt.Sprintf("%s sitting %d (epoch %d, agent %s)", f.SeatID, f.Sitting, f.Epoch, f.AgentID)
	}
	return Audit{Check: "failed-sittings", Verdict: "WARN", Detail: fmt.Sprintf(
		"%d sitting(s) closed with no admitted envelope: %s. Each seat returned without ending its sitting; read what it owed as carried to its next sitting, or never done",
		len(failed), strings.Join(named, "; "))}
}
