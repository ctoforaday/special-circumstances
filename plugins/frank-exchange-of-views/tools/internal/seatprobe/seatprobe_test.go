package seatprobe

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// The ANALYSIS is deterministic and is tested here. The PROBE — dispatching a real seat at a real
// board — is not, and is an operator report rather than a gate (#363). Keeping the two apart is
// what lets this file be a normal test: it feeds a synthetic event stream and checks that the
// reading of it is right.

func writeRun(t *testing.T, events []struct {
	seat, typ string
	payload   *record.Payload
}) string {
	t.Helper()
	runDir := t.TempDir()
	seen := map[string]bool{}
	for _, e := range events {
		if !seen[e.seat] {
			if _, _, err := record.RegisterSeat(record.Identity{RunDir: runDir, SeatID: e.seat, Round: record.RoundOf(e.seat)}); err != nil {
				t.Fatal(err)
			}
			seen[e.seat] = true
		}
		if e.typ == "" {
			continue
		}
		if _, err := record.Append(record.Identity{RunDir: runDir, SeatID: e.seat, Round: record.RoundOf(e.seat)}, e.typ, e.payload); err != nil {
			t.Fatalf("append %s: %v", e.typ, err)
		}
	}
	return runDir
}

// A VERB THE ROLE DOES NOT OFFER IS NOT A VERB THE SEAT DECLINED.
//
// The distinction this whole package turns on: `Unused` must list only what the seat COULD have
// reached for. Conflating "never had it" with "did not take it" would fill every report with
// noise, and a report nobody reads catches nothing.
func TestUnusedListsOnlyWhatTheRoleOffers(t *testing.T) {
	runDir := writeRun(t, []struct {
		seat, typ string
		payload   *record.Payload
	}{
		{seat: "blue-respond-r1", typ: "position", payload: record.NewPayload().Set("text", "the round's narrative")},
	})

	c, err := Read(surface(), runDir, "blue-respond-r1")
	if err != nil {
		t.Fatal(err)
	}
	if c.Role != "blue" {
		t.Fatalf("role = %q, want blue", c.Role)
	}
	if c.Used["position"] != 1 {
		t.Errorf("position not counted: %v", c.Used)
	}
	for _, verb := range c.Unused {
		if verb == "mint" || verb == "verdict" || verb == "opinion" {
			t.Errorf("Unused names %q, which blue never had — reporting a verb a role does not offer as one the seat declined is noise that trains a reader to skim", verb)
		}
	}
	var sawProve bool
	for _, verb := range c.Unused {
		if verb == "prove" {
			sawProve = true
		}
	}
	if !sawProve {
		t.Error("prove is blue's and was not used, so it must be reported unused — this is the exact miss the package exists to surface")
	}
}

// THE EVENT TYPE IS NOT THE VERB, and reading it as one would report every seat as having skipped
// the verbs it actually used. `blue edit` writes `blue_edit`; `blue prove` writes `proof`.
func TestTheVerbIsRecoveredFromTheEventType(t *testing.T) {
	runDir := writeRun(t, []struct {
		seat, typ string
		payload   *record.Payload
	}{
		{seat: "blue-respond-r1", typ: "blue_edit", payload: record.NewPayload().
			Set("old", "a").Set("new", "b").Set("prose", "why")},
	})
	c, err := Read(surface(), runDir, "blue-respond-r1")
	if err != nil {
		t.Fatal(err)
	}
	if c.Used["edit"] != 1 {
		t.Errorf("`blue_edit` did not count as the `edit` verb: %v — a report that cannot map an event back to its verb reports a seat as having skipped what it did", c.Used)
	}
	for _, v := range c.Unused {
		if v == "edit" {
			t.Error("edit reported unused after an edit event")
		}
	}
}

// AN UNMET EXPECTATION NAMES WHAT THE SEAT DID INSTEAD.
//
// "blue did not use prove" is half a finding. The useful half is that it used `edit` four times —
// because that says the seat had a theory of the task and the theory was wrong, which is a
// constitution problem, not a discoverability one.
func TestAnUnmetExpectationNamesTheSubstitute(t *testing.T) {
	var evs []struct {
		seat, typ string
		payload   *record.Payload
	}
	for i := 0; i < 4; i++ {
		evs = append(evs, struct {
			seat, typ string
			payload   *record.Payload
		}{"blue-respond-r1", "blue_edit", record.NewPayload().
			Set("old", "a").Set("new", "b").Set("prose", "why")})
	}
	runDir := writeRun(t, evs)

	got, err := Check(surface(), runDir, []Expectation{{Seat: "blue-respond-r1", Verb: "prove", Because: "the board is arithmetic"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Met {
		t.Fatalf("expectation should be unmet: %+v", got)
	}
	if !strings.Contains(got[0].Substitute, "edit") {
		t.Errorf("substitute = %q, want it to name the edit verb — the finding is what the seat did INSTEAD, not merely what it omitted", got[0].Substitute)
	}
}

// AN EMPTY FRICTION LOG IS REPORTED AS AMBIGUOUS, NEVER AS CLEAN.
//
// Zero friction entries is equally consistent with a seat that met no obstacle and one that met
// several and never used the channel. Printing it as a clean board is the plausible zero this
// project keeps finding, and the report must refuse to produce it.
func TestNoFrictionIsNotReportedAsACleanBoard(t *testing.T) {
	runDir := writeRun(t, []struct {
		seat, typ string
		payload   *record.Payload
	}{
		{seat: "blue-respond-r1", typ: "position", payload: record.NewPayload().Set("text", "n")},
	})
	out, err := Report(surface(), runDir, []string{"blue-respond-r1"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "does not mean none was met") {
		t.Errorf("an empty friction log must be reported as ambiguous, got:\n%s", out)
	}
}

// EVERY EXPECTATION STATES ITS ARGUMENT.
//
// `Because` is what a constitution would have to say to a seat for it to take the verb. An
// expectation without one is a rule with no teachable reason, and the fix it implies ("tell the
// seat to use prove") is the kind that produces compliance rather than judgement.
func TestEveryBoardExpectationArguesItsCase(t *testing.T) {
	for name, b := range Boards() {
		if len(b.Expect) == 0 {
			t.Errorf("board %q expects nothing — a board that demands no verb measures nothing", name)
		}
		for _, e := range b.Expect {
			if strings.TrimSpace(e.Because) == "" {
				t.Errorf("board %q expects %s to use %q with no argument for why", name, e.Seat, e.Verb)
			}
			if e.Seat == "" || e.Verb == "" {
				t.Errorf("board %q has an incomplete expectation: %+v", name, e)
			}
		}
		for _, g := range b.Gaps {
			if g.Baits == "" || g.Why == "" {
				t.Errorf("board %q gap %q does not say what it baits or why — then it is scenery", name, g.Key)
			}
			// A computation-kind gap that baits anything but `prove` is a contradiction: the
			// write path refuses to close it on prose, so no other verb can discharge it.
			if g.CheckKind == "computation" && g.Baits != "prove" {
				t.Errorf("board %q gap %q is check-kind computation but baits %q — a computation check closes ONLY on a proof", name, g.Key, g.Baits)
			}
		}
	}
}
