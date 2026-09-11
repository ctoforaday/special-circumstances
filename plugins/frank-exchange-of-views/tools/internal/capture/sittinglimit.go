package capture

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingcap"
)

// SittingLimitAudit names every sitting the PreToolUse hook stopped at the run's per-sitting
// tool-call limit (sitting_limit events).
//
// A stopped sitting WARNS rather than fails. The run is sound; what the reader needs is which
// seat's work in which sitting ended at the limit rather than at the seat's own judgement, because
// that sitting's position, revision or findings may stop short of what the seat would have filed.
func SittingLimitAudit(run record.Run) Audit {
	m, err := record.MergedEvents(run)
	if err != nil {
		return Audit{Check: "sitting-limit", Verdict: "SKIP", Detail: "the record could not be read: " + err.Error()}
	}
	var stopped []string
	for _, e := range m.Events {
		if e.GetType() != recordpb.EventType_EVENT_TYPE_SITTING_LIMIT {
			continue
		}
		l := e.GetSittingLimit()
		stopped = append(stopped, fmt.Sprintf("%s sitting %d at %d calls", l.GetSeatId(), l.GetSitting(), l.GetLimit()))
	}
	if len(stopped) == 0 {
		limit, lerr := sittingcap.Limit(run.Dir())
		if lerr != nil {
			return Audit{Check: "sitting-limit", Verdict: "PASS", Detail: "no sitting reached the per-sitting tool-call limit; the limit itself does not read: " + lerr.Error()}
		}
		return Audit{Check: "sitting-limit", Verdict: "PASS", Detail: fmt.Sprintf("no sitting reached the per-sitting tool-call limit of %d", limit)}
	}
	return Audit{Check: "sitting-limit", Verdict: "WARN", Detail: fmt.Sprintf(
		"%d sitting(s) stopped at the per-sitting tool-call limit: %s. Each seat was refused every further call and returned; read those sittings' acts as cut short",
		len(stopped), strings.Join(stopped, "; "))}
}
