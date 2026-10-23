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

// storedPlan is the plan a chair's `dispatch next` stores, permitting a PASS or not.
func storedPlan(permitted bool) *recordpb.Plan {
	p := record.Plan{Head: 1, Parties: []record.Party{}, Docket: []string{}, RemandOwed: []record.RemandOwed{}, Exchanges: []record.ExchangeCount{},
		Why: []string{}, StaleAreas: []record.StaleArea{}, Blockers: []record.PlanBlocker{}, PassPermitted: permitted}
	if !permitted {
		p.Blockers = []record.PlanBlocker{{Kind: record.BlockerMaterialGap, Subject: "G1", Owner: "red-lens-logic"}}
	}
	return &recordpb.Plan{Printed: proto.String(p.WithKey().Printed())}
}

// THE POSITION STATES ARE READ IN ONE PLACE. capture's record-parity audit says whether a chair
// sitting that withheld a PASS its plan permitted filed its position, and it is held here to
// record.PositionSittings on one record at one reading.
//
//   - missing: the audit FAILS naming the sitting; the chair's list holds no item.
//   - unresolved: no FAIL — NOT MEASURED, and a SKIP.
//   - not owed (the PASS recorded, or none permitted): no finding.
//   - filed: nothing anywhere but the position.
//
// THE TRANSCRIPT PRINTS WHAT WAS FILED AND IS SILENT ON THE REST. A sitting that holds no position
// has NO section — never a heading over an empty body, and no line saying one is missing, owed or
// pending: the audit above is the reader that reports that. So both transcript forms hold exactly
// one RED section per position on the record, and no BLUE section: blue files none.
func TestThePositionStatesAreReadInOnePlace(t *testing.T) {
	position := &recordpb.Position{Text: proto.String("why the PASS is withheld")}
	type act struct {
		seat string
		body proto.Message
	}
	sits := func(agent string, acts ...proto.Message) []act {
		out := []act{{"red-chair", &recordpb.Register{AgentId: proto.String(agent)}}}
		for _, a := range acts {
			out = append(out, act{"red-chair", a})
		}
		return append(out, act{record.HarnessSeat, &recordpb.SittingClose{AgentId: proto.String(agent)}})
	}
	withheld := sits("chair-a", storedPlan(true), position)
	type want struct {
		states  []record.PositionState
		verdict string
		detail  []string
	}
	cases := map[string]struct {
		acts           []act
		running, after want
	}{
		"withheld and filed": {
			acts:    withheld,
			running: want{states: []record.PositionState{record.PositionFiled}, verdict: "PASS"},
			after:   want{states: []record.PositionState{record.PositionFiled}, verdict: "PASS"},
		},
		"withheld, closed with none": {
			acts: append(append([]act{}, withheld...), sits("chair-b", storedPlan(true))...),
			running: want{states: []record.PositionState{record.PositionFiled, record.PositionMissing}, verdict: "FAIL",
				detail: []string{"red-chair sitting 2, which withheld the PASS its plan permitted, filed no position"}},
			after: want{states: []record.PositionState{record.PositionFiled, record.PositionMissing}, verdict: "FAIL",
				detail: []string{"red-chair sitting 2, which withheld the PASS its plan permitted, filed no position"}},
		},
		"withheld, in flight with none": {
			acts: append(append([]act{}, withheld...), act{"red-chair", &recordpb.Register{AgentId: proto.String("chair-b")}}, act{"red-chair", storedPlan(true)}),
			running: want{states: []record.PositionState{record.PositionFiled, record.PositionUnresolved}, verdict: "SKIP",
				detail: []string{"red-chair sitting 2", "shows no position so far", "NOT MEASURED"}},
			after: want{states: []record.PositionState{record.PositionFiled, record.PositionMissing}, verdict: "FAIL",
				detail: []string{"red-chair sitting 2, which withheld the PASS its plan permitted, filed no position"}},
		},
		"the PASS recorded": {
			acts:    sits("chair-a", storedPlan(true), &recordpb.Gate{Verdict: recordpb.Verdict_VERDICT_PASS.Enum()}),
			running: want{states: []record.PositionState{record.PositionNotOwed}, verdict: "PASS"},
			after:   want{states: []record.PositionState{record.PositionNotOwed}, verdict: "PASS"},
		},
		"no PASS permitted": {
			acts:    sits("chair-a", storedPlan(false)),
			running: want{states: []record.PositionState{record.PositionNotOwed}, verdict: "PASS"},
			after:   want{states: []record.PositionState{record.PositionNotOwed}, verdict: "PASS"},
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
				var states []record.PositionState
				for _, p := range record.PositionSittings(fam.Events, fam.At, read.when) {
					states = append(states, p.State)
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

			// THE TRANSCRIPT: one RED section per filed position, in both forms, and nothing for a
			// sitting that holds none.
			filed := 0
			for _, e := range fam.Events {
				if e.GetType() == recordpb.EventType_EVENT_TYPE_POSITION {
					filed++
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
			red := 0
			for _, ep := range served.Epochs {
				red += len(ep.Red)
			}
			if red != filed {
				t.Errorf("the structured transcript holds %d red section(s); the record holds %d position(s)", red, filed)
			}
			md, err := view.Markdown(run, "debate", "")
			if err != nil {
				t.Fatal(err)
			}
			headings := regexp.MustCompile(`(?m)^### (RED|BLUE)(.*)\n(.*)$`).FindAllStringSubmatch(string(md), -1)
			got := map[string]int{}
			for _, h := range headings {
				got[h[1]]++
				if strings.TrimSpace(h[2]) != "" {
					t.Errorf("the transcript heads a section %q: a position's heading says the party and nothing of a position not filed", h[0])
				}
				if strings.TrimSpace(h[3]) == "" {
					t.Errorf("the transcript holds a heading with an empty body — a seat with no position has no section:\n%s", md)
				}
			}
			if got["RED"] != filed || got["BLUE"] != 0 {
				t.Errorf("the markdown transcript holds %d RED and %d BLUE section(s); the record holds %d position(s), all the chair's:\n%s", got["RED"], got["BLUE"], filed, md)
			}
			cw, err := record.WorkOfSeat(run, "chair", "red-chair")
			if err != nil {
				t.Fatal(err)
			}
			for _, it := range cw.Sitting.Open {
				if strings.Contains(it.What, "position") {
					t.Errorf("the chair's work list holds a position item: %q", it.What)
				}
			}
		})
	}
}
