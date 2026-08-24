// Liveness answers "is this run still doing anything" from the record, and it is PULL-BASED on
// purpose.
//
// # The failure this exists for
//
// Measured 2026-08-22. A run's workflow was killed mid-round — the environment SIGTERMs Claude
// Code when the session idles, and a workflow is in-process state, so it went with the process.
// Forty-one minutes later every instrument still described a healthy run in flight:
//
//	.claude/run-live.json   present
//	dashboard header        "Seats live now"
//	progress bar            "round 1: live"
//	ETA                     "projected 3-3 min remaining"
//	cost                    unchanged, which reads as a quiet moment
//
// Nothing anywhere said it had stopped. The marker is lifted by `capture`, and `capture` only
// runs when the workflow RETURNS — so a workflow that dies never lifts its own marker, and the
// run stays "live" forever.
//
// # Why the fix is not a signal handler, and not a heartbeat
//
// A killed process cannot tell you it was killed. That is a law, not an implementation gap:
// every push-based liveness scheme asks the thing that died to be the thing that reports the
// death. (Measured too: the workflow sandbox has no `process`, no `require` and no
// `addEventListener`, so no handler can be installed — and the grace period is five seconds,
// which could not finish an in-flight seat anyway.) A separate heartbeat daemon just moves the
// same defect one level down: it is another process that can die, invisibly.
//
// What a live run DOES leave is evidence, in the record, written by the work itself. Its
// ABSENCE is observable by anyone who looks, with no cooperation from the dead party. So
// liveness is derived from the newest event's own timestamp, and the threshold that separates
// "quiet" from "stopped" is derived from the same record rather than hard-coded — a constant
// would be a new magic number, wrong for a slow lens and unfalsifiable.
package record

import (
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"sort"
	"time"
)

// Activity is the record's own answer to "when did anything last happen here".
type Activity struct {
	At     time.Time
	Seat   string
	Type   string
	Events int
}

// LiveState is what a reader should SAY about the run. Four states, because the two the system
// had were being asked to cover four situations, and "stale" had nowhere to land — so it
// rendered as live, permanently.
type LiveState string

const (
	// StateEnded: the run finished and was captured. Nothing is expected to move again.
	StateEnded LiveState = "ENDED"
	// StateLive: the record moved recently enough to be consistent with work in progress.
	StateLive LiveState = "LIVE"
	// StateStale: the marker says a run is open and the record has stopped moving. The run is
	// PRESUMED DEAD and a reader is told how to resume it.
	StateStale LiveState = "STALE"
	// StateUnmeasured: the honest zero. Not enough of a record to say either way — reported as
	// itself, never folded into LIVE, which is the mistake this whole file is about.
	StateUnmeasured LiveState = "NOT MEASURED"
)

// Liveness is the verdict plus everything needed to argue with it.
type Liveness struct {
	State     LiveState
	Age       time.Duration // since the newest event
	Threshold time.Duration // what counted as stopped
	Basis     string        // how Threshold was derived, or why it could not be
	Last      Activity
}

// minStaleFloor keeps a run with a very fast cadence from being called dead during one slow
// seat. A frontier seat routinely runs minutes; the floor is deliberately generous because a
// false STALE costs an operator a pointless resume, while a false LIVE costs the whole run.
const minStaleFloor = 6 * time.Minute

// LastActivity reports the newest event on the record, by the event's OWN ts field.
//
// NOT FILE MTIME. mtime is a fact about the filesystem standing in for a fact the record
// already holds — the same substitution this package exists to refuse elsewhere. Events carry
// `ts` (see Event.TS, which replay already orders by), so the record can answer for itself.
//
// An empty record is an ERROR, not a zero time: a run that has recorded nothing and a run whose
// record cannot be read must not both arrive as "very old".
func LastActivity(runDir string) (Activity, error) {
	merged, err := MergedEvents(runDir)
	if err != nil {
		return Activity{}, err
	}
	if len(merged.Events) == 0 {
		return Activity{}, fmt.Errorf("record: no events in %s — a run that has written nothing "+
			"is not a run that has gone quiet, and reporting it as an age would make the two the same", runDir)
	}
	var newest Activity
	newest.Events = len(merged.Events)
	for _, e := range merged.Events {
		t, perr := time.Parse(time.RFC3339Nano, e.GetTs())
		if perr != nil {
			continue
		}
		if t.After(newest.At) {
			newest.At, newest.Seat, newest.Type = t, e.GetSeatId(), recordpb.Word(e.GetType())
		}
	}
	if newest.At.IsZero() {
		return Activity{}, fmt.Errorf("record: %d event(s) in %s and not one carried a parseable ts — "+
			"the ordering key replay depends on is absent, so nothing here can be aged", len(merged.Events), runDir)
	}
	return newest, nil
}

// staleThreshold derives "how long is too long" from THIS run's own cadence.
//
// A hard-coded constant is a new magic number: wrong for a slow lens, unfalsifiable, and it
// would need a second constant the day a seat class got slower. The record already shows how
// often this run writes — the gaps between consecutive events ARE the cadence — so the threshold
// is 3x the p90 gap, floored.
//
// Too few gaps to characterise returns ok=false, and the caller must say NOT MEASURED rather
// than pick a default. A threshold nobody can justify is the same defect as a fact nobody can
// refuse.
func staleThreshold(events []*Event) (time.Duration, string, bool) {
	var ts []time.Time
	for _, e := range events {
		if t, err := time.Parse(time.RFC3339Nano, e.GetTs()); err == nil {
			ts = append(ts, t)
		}
	}
	if len(ts) < 6 {
		return 0, fmt.Sprintf("only %d timestamped event(s) — too few to characterise this run's cadence", len(ts)), false
	}
	sort.Slice(ts, func(i, j int) bool { return ts[i].Before(ts[j]) })
	gaps := make([]time.Duration, 0, len(ts)-1)
	for i := 1; i < len(ts); i++ {
		gaps = append(gaps, ts[i].Sub(ts[i-1]))
	}
	sort.Slice(gaps, func(i, j int) bool { return gaps[i] < gaps[j] })
	p90 := gaps[(len(gaps)*90)/100]
	d := 3 * p90
	basis := fmt.Sprintf("3x the p90 gap between %d events (p90 %s)", len(ts), p90.Round(time.Second))
	if d < minStaleFloor {
		return minStaleFloor, basis + fmt.Sprintf(", floored at %s", minStaleFloor), true
	}
	return d, basis, true
}

// Assess is the whole verdict. `ended` is the caller's answer to "has capture run" — the one
// fact the record cannot supply, because it is about the run's disposition rather than its
// contents.
func Assess(runDir string, now time.Time, ended bool) Liveness {
	if ended {
		return Liveness{State: StateEnded, Basis: "the run was captured"}
	}
	last, err := LastActivity(runDir)
	if err != nil {
		return Liveness{State: StateUnmeasured, Basis: err.Error()}
	}
	merged, err := MergedEvents(runDir)
	if err != nil {
		return Liveness{State: StateUnmeasured, Basis: err.Error(), Last: last}
	}
	age := now.Sub(last.At)
	thr, basis, ok := staleThreshold(merged.Events)
	if !ok {
		return Liveness{State: StateUnmeasured, Age: age, Basis: basis, Last: last}
	}
	st := StateLive
	if age > thr {
		st = StateStale
	}
	return Liveness{State: st, Age: age, Threshold: thr, Basis: basis, Last: last}
}

// Says renders the verdict as one line a human can act on. The STALE arm names the age AND the
// threshold, because "stopped" is a judgement and the reader is owed the basis for it.
func (l Liveness) Says() string {
	switch l.State {
	case StateEnded:
		return "run complete — captured"
	case StateLive:
		return fmt.Sprintf("live — last event %s ago (%s), stale after %s",
			l.Age.Round(time.Second), l.Last.Seat, l.Threshold.Round(time.Second))
	case StateStale:
		return fmt.Sprintf("STALE — nothing on the record for %s (last: %s, %s), past the %s "+
			"this run's own cadence implies. The workflow is presumed dead: an idle SIGTERM kills "+
			"it and it cannot lift its own marker. Resume it rather than reading this board as live.",
			l.Age.Round(time.Second), l.Last.Seat, l.Last.Type, l.Threshold.Round(time.Second))
	case StateUnmeasured:
		return "liveness NOT MEASURED — " + l.Basis
	default:
		// THE ZERO VALUE IS ITS OWN ANSWER, and it must not read like a measurement. A caller
		// that never called Assess leaves State empty; saying "NOT MEASURED" here would claim
		// the record was consulted and found wanting, when nobody looked at all. A forgotten
		// assessment has to be visible, or the page renders clean for the one reason it should not.
		return "liveness not assessed by this caller"
	}
}

// MOVED HERE FROM internal/dashboard, because it reads only the record and two other
// packages needed it. capture importing a presentation layer to ask what the bench decided
// would have been the wrong direction; one derivation, three consumers.
// TerminalVerdict answers from the RECORD, and from nothing else.
//
// IT WAS A REGEX OVER report.md, and then briefly a regex kept "as a last resort", which is not a
// justification. The verdict is a FIELD on the `outcome` event — the bench's terminal act writes
// it — so parsing it back out of the prose it was rendered into is the shape [[facts-are-fields]]
// names, and keeping the parse as a fallback keeps the shape.
//
// WHAT THE FALLBACK ACTUALLY SERVED, measured across the 9 assembled runs in research/ rather than
// assumed:
//
//	2  carry an `outcome` event — the record answers
//	1  carries a round `verdict` event and no terminal act
//	5  carry NO terminal act at all, and their reports say "UNVERIFIED"
//	1  has no verdict in the report either
//
// Those five are pre-#289 artifacts, assembled before the terminal act was an event. Their
// "UNVERIFIED" is backed by no record anywhere — so the fallback's job was to read a word out of
// prose and hand it to an operator as the run's verdict, which is exactly the unbacked assertion
// `basisNote` exists to keep out of the report. A fact that no record holds is not recovered by
// finding it written down.
//
// The honest answer when the record cannot say is that the record cannot say, and the renderer
// already says it well: an empty terminal verdict falls through to the round verdict off the
// record and is RELABELLED from "final verdict" to "latest verdict (rN)". The operator sees a
// different claim rather than the same claim from a worse source.
// It is exported for every caller that needs to know whether the bench recorded an outcome —
// the dashboard SERVER, whose lifetime was keyed only to the run-live marker (a file whose only
// remover is `capture`, so a killed run left it watching forever, #270), and capture's own
// liveness audit, which cannot tell a finished run from a killed one without it. The record
// knows what the marker cannot.
func TerminalVerdict(runDir string) string {
	b, err := BoardState(runDir)
	if err != nil {
		return ""
	}
	// The bench's own terminal act, latest wins.
	for i := len(b.Events) - 1; i >= 0; i-- {
		if b.Events[i].GetType() != recordpb.EventType_EVENT_TYPE_OUTCOME {
			continue
		}
		if v := recordpb.Word(b.Events[i].GetOutcome().GetVerdict()); v != "" {
			return v
		}
	}
	// Or what the record decides for itself. ok is false only where the record genuinely
	// cannot — a judged deadlock — and that is a real answer, not a gap to paper over.
	if v, _, ok := DeriveVerdict(runDir); ok {
		return v
	}
	return ""
}
