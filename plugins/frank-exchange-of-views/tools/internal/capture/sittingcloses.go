package capture

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// SittingClosesAudit reads the SubagentStop hook's closes against the sittings that owed one
// (record.AgentSittings): every sitting whose opening carries an agent closes when that agent returns.
//
// NOTHING ELSE NOTICES A HOOK THAT STOPPED CLOSING. A sitting with no close has not failed
// (record.FailedSittingsOf reads closed sittings only), so a run whose every close vanished reports
// zero failed sittings; no sitting reads sat through, so no lens pin moves and no lens retires, and
// every gate stays green.
//
// It FAILS a run whose sittings ended — their seat sat again — with not one close on the record: that
// is the hook, not an agent. It WARNS naming each ended sitting with no close among closes that did
// land: its agent died before SubagentStop fired, or the hook lost that close, and either way the
// sitting moved no pin and counted nothing toward retirement. A seat's latest sitting with no close is
// not a finding, because it may be in flight: capture mid-run is the operator's call, and the record
// does not say when an agent died.
func SittingClosesAudit(run record.Run) Audit {
	const check = "sitting-closes"
	sittings, err := record.AgentSittings(run)
	if err != nil {
		return Audit{Check: check, Verdict: "SKIP", Detail: "the record could not be read: " + err.Error()}
	}
	if len(sittings) == 0 {
		return Audit{Check: check, Verdict: "SKIP", Detail: "NOT MEASURED — no sitting's opening carries an agent, so no close was owed to compare: no hook reached this run's seats"}
	}
	closed, latest := 0, 0
	var neverClosed []string
	for _, s := range sittings {
		switch {
		case s.Closed:
			closed++
		case s.Over:
			neverClosed = append(neverClosed, fmt.Sprintf("%s sitting %d (epoch %d, agent %s)", s.SeatID, s.Sitting, s.Epoch, s.AgentID))
		default:
			latest++
		}
	}
	if closed == 0 && len(neverClosed) > 0 {
		return Audit{Check: check, Verdict: "FAIL", Detail: fmt.Sprintf(
			"%d sitting(s) opened under an agent, %d of them ended by their seat sitting again, and not one closed: the SubagentStop hook wrote no close this run, so no sitting reads sat through — no lens pin moved and no lens retired",
			len(sittings), len(neverClosed))}
	}
	if len(neverClosed) > 0 {
		return Audit{Check: check, Verdict: "WARN", Detail: fmt.Sprintf(
			"%d of %d sitting(s) opened under an agent ended with no close: %s. Each agent died before SubagentStop fired, or the hook lost its close; read each as a sitting its seat did not sit through",
			len(neverClosed), len(sittings), strings.Join(neverClosed, "; "))}
	}
	return Audit{Check: check, Verdict: "PASS", Detail: fmt.Sprintf(
		"%d sitting(s) opened under an agent: %d closed, %d not closed yet (each its seat's latest)", len(sittings), closed, latest)}
}
