package fuzz

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/debatejs"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// The party whose barrier convened a hearing: the engine hears after the chair, after its lenses
// and after blue, and nowhere else.
const (
	barrierChair = "chair"
	barrierLens  = "lens"
	barrierBlue  = "blue"
)

// THE TWO BARRIERS A DRAW MIGHT MISS ARE FORCED, ONE RUN EACH: each seed's party files a petition in
// its first sitting, so the sweep's count of hearings at that barrier is fed by construction.
const (
	lensPetitionSeed = 13
	bluePetitionSeed = 14
)

// petitionSeatOf is the party a seed's run forces a petition from; "" for a seed that forces none.
func petitionSeatOf(seed int64) string {
	switch seed {
	case lensPetitionSeed:
		return barrierLens
	case bluePetitionSeed:
		return barrierBlue
	}
	return ""
}

// partyOfBarrier is the barrier seatID's return reaches; "" for a seat with none: a lane, the
// synthesis and the bench convene nothing.
func partyOfBarrier(seatID string) string {
	switch {
	case seatID == "red-chair":
		return barrierChair
	case strings.HasPrefix(seatID, "red-lens"):
		return barrierLens
	case seatID == "blue-respond":
		return barrierBlue
	}
	return ""
}

// partyReturned notes the party whose barrier a hearing that opens next is held at.
func (r *runner) partyReturned(seatID string) {
	if party := partyOfBarrier(seatID); party != "" {
		r.mu.Lock()
		r.lastParty = party
		r.mu.Unlock()
	}
}

// hearingOpens counts a hearing by the barrier it opened at, and holds it to the rule that convenes
// one: nothing is running.
func (r *runner) hearingOpens(seatID, prompt string) {
	if occ := debatejs.OccasionOf(prompt); occ != "petition" && occ != "position" {
		return
	}
	r.mu.Lock()
	var running []string
	for seat, n := range r.unreturned {
		if n > 0 {
			running = append(running, seat)
		}
	}
	if r.hearingsAfter == nil {
		r.hearingsAfter = map[string]int{}
	}
	r.hearingsAfter[r.lastParty]++
	r.mu.Unlock()
	if len(running) > 0 {
		slices.Sort(running)
		r.noteActErr(fmt.Sprintf("%s opened a hearing while %v had not returned: the bench hears only when nothing is running", seatID, running))
	}
}

// writeRecordRead is the PreToolUse hook's write on an envelope call, over the same fold.
func (r *runner) writeRecordRead(seatID string, env map[string]any) {
	read, err := recordReadAsEnvelope(r)
	if err != nil {
		r.noteActErr(seatID + "'s envelope could not carry the record's read: " + err.Error())
		return
	}
	env["record_read"] = read
}

func recordReadAsEnvelope(r *runner) (map[string]any, error) {
	read, err := record.ReadRecord(r.run())
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(read)
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	return doc, json.Unmarshal(b, &doc)
}

// A PETITION IS HEARD AT THE BARRIER AFTER THE SEAT THAT FILED IT, off the read the record gave, and
// a halt recorded there stops whoever has not started. Each case is debate.js over a real record: a
// seat files through the verb, the read is the Go fold, and the bench rules through the verb. What
// the chair relays is stubbed, so that one plan names a lens and blue by construction. Each read the
// engine acted on is then recomputed at the event it names, which is what capture's audit holds a run to.
func TestAFilingIsHeardAtTheBarrierAfterItsSeat(t *testing.T) {
	script, err := repotree.DebateJS()
	if err != nil {
		t.Fatal(err)
	}
	const lens = "red-lens-evidence"
	for _, tc := range []struct {
		name, filer string
		halts       bool
		want        []string
		verdict     string
	}{
		{"the chair's", "red-chair", false, []string{"chair", "petition", "lens", "blue", "chair"}, "CEILING"},
		{"a lens's", lens, false, []string{"chair", "lens", "petition", "blue", "chair"}, "CEILING"},
		{"blue's", "blue-respond", false, []string{"chair", "lens", "blue", "petition", "chair"}, "CEILING"},
		{"the chair's, halting", "red-chair", true, []string{"chair", "petition"}, "HALTED"},
		{"a lens's, halting: blue never starts", lens, true, []string{"chair", "lens", "petition"}, "HALTED"},
		{"blue's, halting: no chair follows", "blue-respond", true, []string{"chair", "lens", "blue", "petition"}, "HALTED"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := stagedRunner(t)
			r.forceVerified = true
			r.forceHalt = tc.halts
			sittings := 0
			var sat []string
			var reads []record.RecordRead
			sits := func(role, seatID string, occasion ...string) {
				sat = append(sat, role)
				r.sit(role, seatID, occasion...)
				if seatID == tc.filer && r.petitionsFiledN() == 0 {
					r.filePetition(seatID)
				}
			}
			backend := func(seatID, label, prompt string) debatejs.Envelope {
				e := debatejs.Envelope{}
				switch {
				case seatID == "red-chair":
					sittings++
					sits(barrierChair, seatID)
					if sittings > 1 {
						e.Relay(plan([]any{}, false, true, nil, "epoch limit reached"))
						return e
					}
					e.Relay(plan([]any{party(lens), party("blue-respond", "G-00000001")}, false, false, nil, "both sit"))
				case seatID == lens:
					sits(barrierLens, seatID)
				case seatID == "blue-respond":
					sits(barrierBlue, seatID)
				case debatejs.OccasionOf(prompt) == "petition":
					sat = append(sat, "petition")
					r.sit("bench", "judge", "petition")
					r.rulePetitions("judge")
				}
				return e
			}
			read := func(d debatejs.Dispatch) any {
				if d.Occasion == "assemble" || d.Occasion == "terminal" {
					return debatejs.StandInRead(d)
				}
				doc, err := recordReadAsEnvelope(r)
				if err != nil {
					t.Errorf("reading the record at %s's close: %v", d.Label, err)
					return debatejs.StandInRead(d)
				}
				b, _ := json.Marshal(doc)
				var got record.RecordRead
				if err := json.Unmarshal(b, &got); err != nil {
					t.Fatal(err)
				}
				reads = append(reads, got)
				return doc
			}
			_, out, err := debatejs.CaptureRun(script, debatejs.Config{
				Topic: "barrier", RunDir: t.TempDir(), BinDir: t.TempDir(), Lanes: 1,
				Model: "haiku", JudgmentModel: "haiku", Backend: backend, Read: read, Timeout: 2 * time.Minute,
			})
			if err != nil {
				t.Fatal(err)
			}
			if r.actErr != "" {
				t.Fatalf("an act the drive issued: %s", r.actErr)
			}
			if n := r.petitionsFiledN(); n != 1 {
				t.Fatalf("%d petitions filed, want the filer's one: the case exercises nothing", n)
			}
			if !reflect.DeepEqual(sat, tc.want) {
				t.Errorf("sat in the order %v, want %v", sat, tc.want)
			}
			if out.Verdict != tc.verdict || out.Halted != tc.halts {
				t.Errorf("the run ended %s (halted %v), want %s (halted %v)", out.Verdict, out.Halted, tc.verdict, tc.halts)
			}
			fam, err := record.FamilyOf(r.run())
			if err != nil {
				t.Fatal(err)
			}
			for _, got := range reads {
				want, err := record.RecordReadAt(fam.Events, fam.At, got.Event)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("the read the engine acted on = %+v, and the record at event %d says %+v", got, got.Event, want)
				}
			}
		})
	}
}
