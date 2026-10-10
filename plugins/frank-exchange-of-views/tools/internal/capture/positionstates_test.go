package capture

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// THE POSITION STATES ARE READ IN ONE PLACE. Two surfaces say whether a sitting filed its position:
// capture's record-parity audit and blue's work list. Each is held here to record.PositionSittings
// on one record at one reading — so a sitting cannot be short of a position on one and complete on
// the other.
//
//   - missing: the audit FAILS naming the sitting; blue-respond's list blocks on it and the chair's
//     holds no item.
//   - unresolved: no FAIL — NOT MEASURED, and a SKIP; blue-respond's list blocks on it (the seat can
//     still file).
//   - not owed: no finding and no item.
//   - filed: nothing anywhere but the position.
//   - a lane, the frontier and the synthesizer: no row, no finding, no item.
//
// THE TRANSCRIPT PRINTS WHAT WAS FILED AND IS SILENT ON THE REST. A sitting that holds no position
// has NO section — never a heading over an empty body, and no line saying one is missing, owed or
// pending: the audit above is the reader that reports that. So both transcript forms hold exactly
// one section per position on the record, whatever the states are.
func TestThePositionStatesAreReadInOnePlace(t *testing.T) {
	position := &recordpb.Position{Text: proto.String("what the bench is asked to weigh")}
	type act struct {
		seat string
		body proto.Message
	}
	chairSat := []act{{"red-chair", &recordpb.Register{AgentId: proto.String("chair-agent")}}, {"red-chair", position},
		{record.HarnessSeat, &recordpb.SittingClose{AgentId: proto.String("chair-agent"), AgentType: proto.String("frank-exchange-of-views:red-chair")}}}
	mint := act{"red-lens-logic", &recordpb.Mint{GapId: proto.String("G1"), Problem: proto.String("p"),
		RequiredFix: proto.String("f"), AcceptanceCheck: proto.String("the check runs"), Class: proto.String("self-attestation"),
		CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}}
	dispatch := act{"red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}}}
	closeG1 := act{"red-lens-logic", &recordpb.Close{GapId: proto.String("G1"),
		ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), Prose: proto.String("verified at the leaf")}}
	blueSits := act{"blue-respond", &recordpb.Register{AgentId: proto.String("blue-agent")}}
	blueStops := act{record.HarnessSeat, &recordpb.SittingClose{AgentId: proto.String("blue-agent"), AgentType: proto.String("frank-exchange-of-views:blue-researcher")}}
	// The chair sits, files and returns; a lens mints G1; the chair's dispatch engages blue on it.
	// The dispatch row is the returned chair's, written before its stop in a real run; its place
	// after it here changes nothing a sitting's bounds read.
	engaged := append(append([]act{}, chairSat...), mint, dispatch)

	type want struct {
		states  []record.PositionState // every row, in stream order
		verdict string
		detail  []string // each a substring of the audit's detail
		item    bool     // blue-respond's work list blocks on its position
	}
	cases := map[string]struct {
		acts           []act
		running, after want
	}{
		"blue-respond filed": {
			acts:    append(append([]act{}, engaged...), blueSits, act{"blue-respond", position}, blueStops),
			running: want{states: []record.PositionState{record.PositionFiled, record.PositionFiled}, verdict: "PASS"},
			after:   want{states: []record.PositionState{record.PositionFiled, record.PositionFiled}, verdict: "PASS"},
		},
		"blue-respond closed its sitting with none": {
			acts: append(append([]act{}, engaged...), blueSits, blueStops),
			running: want{states: []record.PositionState{record.PositionFiled, record.PositionMissing}, verdict: "FAIL",
				detail: []string{"blue sitting 1, engaged on G1 still open when it sat, filed no position"}, item: true},
			after: want{states: []record.PositionState{record.PositionFiled, record.PositionMissing}, verdict: "FAIL",
				detail: []string{"blue sitting 1, engaged on G1 still open when it sat, filed no position"}, item: true},
		},
		"blue-respond in flight with none": {
			acts: append(append([]act{}, engaged...), blueSits),
			running: want{states: []record.PositionState{record.PositionFiled, record.PositionUnresolved}, verdict: "SKIP",
				detail: []string{"blue sitting 1", "shows no position so far", "NOT MEASURED"}, item: true},
			after: want{states: []record.PositionState{record.PositionFiled, record.PositionMissing}, verdict: "FAIL",
				detail: []string{"blue sitting 1, engaged on G1 still open when it sat, filed no position"}, item: true},
		},
		"blue-respond found its gap closed": {
			acts:    append(append([]act{}, engaged...), closeG1, blueSits, blueStops),
			running: want{states: []record.PositionState{record.PositionFiled, record.PositionNotOwed}, verdict: "PASS"},
			after:   want{states: []record.PositionState{record.PositionFiled, record.PositionNotOwed}, verdict: "PASS"},
		},
		"the chair closed a sitting with none": {
			acts: append(append([]act{}, chairSat...), act{"red-chair", &recordpb.Register{AgentId: proto.String("chair-2")}},
				act{record.HarnessSeat, &recordpb.SittingClose{AgentId: proto.String("chair-2"), AgentType: proto.String("frank-exchange-of-views:red-chair")}}),
			running: want{states: []record.PositionState{record.PositionFiled, record.PositionMissing}, verdict: "FAIL",
				detail: []string{"red-chair sitting 2 filed no position"}},
			after: want{states: []record.PositionState{record.PositionFiled, record.PositionMissing}, verdict: "FAIL",
				detail: []string{"red-chair sitting 2 filed no position"}},
		},
		"the chair in flight with none": {
			acts: append(append([]act{}, chairSat...), act{"red-chair", &recordpb.Register{AgentId: proto.String("chair-2")}}),
			running: want{states: []record.PositionState{record.PositionFiled, record.PositionUnresolved}, verdict: "SKIP",
				detail: []string{"red-chair sitting 2 shows no position so far", "NOT MEASURED"}},
			after: want{states: []record.PositionState{record.PositionFiled, record.PositionMissing}, verdict: "FAIL",
				detail: []string{"red-chair sitting 2 filed no position"}},
		},
		"a lane, the frontier and the synthesizer sat and filed none": {
			acts: append(append([]act{}, chairSat...),
				act{"blue-lane-1", &recordpb.Register{AgentId: proto.String("lane")}}, act{record.HarnessSeat, &recordpb.SittingClose{AgentId: proto.String("lane"), AgentType: proto.String("frank-exchange-of-views:blue-researcher")}},
				act{"frontier", &recordpb.Register{AgentId: proto.String("frontier")}}, act{record.HarnessSeat, &recordpb.SittingClose{AgentId: proto.String("frontier"), AgentType: proto.String("frank-exchange-of-views:blue-researcher")}},
				act{"blue-synthesize", &recordpb.Register{AgentId: proto.String("synth")}}, act{record.HarnessSeat, &recordpb.SittingClose{AgentId: proto.String("synth"), AgentType: proto.String("frank-exchange-of-views:blue-synthesizer")}}),
			running: want{states: []record.PositionState{record.PositionFiled}, verdict: "PASS"},
			after:   want{states: []record.PositionState{record.PositionFiled}, verdict: "PASS"},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			var evs []*record.Event
			for i, a := range c.acts {
				evs = append(evs, recordtest.At(t, a.seat, fmt.Sprintf("%s:%d", a.seat, i+1), a.body))
			}
			recordtest.Seed(t, dir, evs...)
			run := runtest.Open(t, dir)
			fam, err := record.FamilyOf(run)
			if err != nil {
				t.Fatal(err)
			}
			for _, read := range []struct {
				when record.ReadWhen
				want want
			}{{record.WhileRunning, c.running}, {record.AfterTheRun, c.after}} {
				rows := record.PositionSittings(fam.Events, fam.At, read.when)
				var states []record.PositionState
				for _, p := range rows {
					states = append(states, p.State)
					if p.Seat != "red-chair" && p.Seat != "blue-respond" {
						t.Errorf("a row for %s: only a seat that owes a position has one", p.Seat)
					}
				}
				if fmt.Sprint(states) != fmt.Sprint(read.want.states) {
					t.Fatalf("read %v: states = %v, want %v", read.when, states, read.want.states)
				}

				// CAPTURE fails a missing sitting, never an unresolved one, and says nothing of the rest.
				a := recordParityAt(fam.Events, fam.At, read.when)
				if a.Verdict != read.want.verdict {
					t.Errorf("read %v: record-parity is %s (%s), want %s", read.when, a.Verdict, a.Detail, read.want.verdict)
				}
				for _, d := range read.want.detail {
					if !strings.Contains(a.Detail, d) {
						t.Errorf("read %v: record-parity's detail does not hold %q:\n%s", read.when, d, a.Detail)
					}
				}
				if a.Verdict != "FAIL" && strings.Contains(a.Detail, "filed no position") {
					t.Errorf("read %v: a %s audit says a sitting filed no position:\n%s", read.when, a.Verdict, a.Detail)
				}

			}

			// THE TRANSCRIPT: one section per filed position, in both forms, and nothing for a sitting
			// that holds none.
			filed := map[string]int{}
			for _, e := range fam.Events {
				if e.GetType() == recordpb.EventType_EVENT_TYPE_POSITION {
					filed[record.PartyOf(e)]++
				}
			}
			var served record.DebateJSON
			b, err := record.DebateJSONBytes(run)
			if err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(b, &served); err != nil {
				t.Fatal(err)
			}
			red, blue := 0, 0
			for _, ep := range served.Epochs {
				red, blue = red+len(ep.Red), blue+len(ep.Blue)
			}
			if red != filed["chair"] || blue != filed["blue"] {
				t.Errorf("the structured transcript holds %d red and %d blue section(s); the record holds %d and %d position(s)", red, blue, filed["chair"], filed["blue"])
			}
			md, err := view.Markdown(run, "debate", "")
			if err != nil {
				t.Fatal(err)
			}
			headings := regexp.MustCompile(`(?m)^### (RED|BLUE)(.*)\n(.*)$`).FindAllStringSubmatch(string(md), -1)
			got := map[string]int{}
			for _, h := range headings {
				got[h[1]]++
				if strings.TrimSpace(h[2]) != "" && !strings.Contains(h[2], "CLOSING") {
					t.Errorf("the transcript heads a section %q: a position's heading says the party and nothing of a position not filed", h[0])
				}
				if strings.TrimSpace(h[3]) == "" {
					t.Errorf("the transcript holds a heading with an empty body — a seat with no position has no section:\n%s", md)
				}
			}
			if got["RED"] != filed["chair"] || got["BLUE"] != filed["blue"] {
				t.Errorf("the markdown transcript holds %d RED and %d BLUE section(s); the record holds %d and %d position(s):\n%s", got["RED"], got["BLUE"], filed["chair"], filed["blue"], md)
			}
			w, err := record.WorkOfSeat(run, "blue", "blue-respond")
			if err != nil {
				t.Fatal(err)
			}
			item := false
			for _, it := range w.Sitting.Open {
				if strings.Contains(it.What, "this sitting's position is missing") {
					item = it.Blocks
				}
			}
			if item != c.running.item {
				t.Errorf("blue-respond's work list blocks on its position: %v, want %v\n%+v", item, c.running.item, w.Sitting.Open)
			}
			cw, err := record.WorkOfSeat(run, "chair", "red-chair")
			if err != nil {
				t.Fatal(err)
			}
			for _, it := range cw.Sitting.Open {
				if strings.Contains(it.What, "position is missing") {
					t.Errorf("the chair's work list holds a position item: %q", it.What)
				}
			}
		})
	}
}
