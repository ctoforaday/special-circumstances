package record

import (
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// The role boundary is the engine's premise, so it gets a test per crossing
// rather than one happy-path assertion. Before this binding existed, every
// "refused" case below SUCCEEDED.
func TestSeatRoleBinding(t *testing.T) {
	allowed := map[string][]string{
		"lens":  {"red-lens-evidence", "red-lens-dark-side"},
		"chair": {"red-chair", "red-chair"},
		"blue":  {"blue-lane-1", "blue-respond", "blue-synthesize", "frontier"},
		"bench": {"judge"},
	}
	for role, seats := range allowed {
		for _, s := range seats {
			if err := CheckSeatRole(role, s); err != nil {
				t.Errorf("%s should accept its own seat %q: %v", role, s, err)
			}
		}
	}
	// Every cross-role pairing is refused, not just the one that motivated this.
	for role := range allowed {
		for other, seats := range allowed {
			if other == role {
				continue
			}
			for _, s := range seats {
				err := CheckSeatRole(role, s)
				if err == nil {
					t.Errorf("%s accepted %s seat %q — the role boundary is not enforced", role, other, s)
					continue
				}
				if got := err.Error(); !contains(got, other) {
					t.Errorf("refusal for %q should name its real role %q: %s", s, other, got)
				}
			}
		}
	}
	if err := CheckSeatRole("blue", "totally-made-up"); err == nil {
		t.Error("a seat id no dispatch created was accepted")
	}
}

func contains(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}

func TestRegisterSeatRejectsMalformedSeatIDs(t *testing.T) {
	cases := []struct {
		name string
		id   string
	}{
		{"empty", ""},
		{"leading digit", "1lens"},
		{"leading hyphen", "-lens"},
		{"path traversal", "../escape"},
		{"path separator", "red/lens"},
		{"windows separator", `red\lens`},
		{"underscore is not in the alphabet", "red_lens"},
		{"space", "red lens"},
		{"dot", "red.lens"},
		{"nul byte", "red\x00lens"},
		{"newline", "red\nlens"},
		{"unicode", "red-日本語"},
		{"glob", "red-*"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runDir := newRun(t)
			_, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: tc.id}, "", "")
			if err == nil {
				t.Fatalf("RegisterSeat accepted %q — the id becomes a FILENAME", tc.id)
			}
			if !strings.Contains(err.Error(), "invalid --seat-id") {
				t.Errorf("refusal does not name the flag: %v", err)
			}
			// Nothing was written under a rejected id.
			entries, rerr := os.ReadDir(recordsDirT(runDir))
			if rerr == nil {
				for _, e := range entries {
					if strings.HasPrefix(e.Name(), "events-") || strings.HasPrefix(e.Name(), ".active-") {
						t.Errorf("a rejected seat id still produced %s", e.Name())
					}
				}
			}
		})
	}
}

func TestRegisterSeatAcceptsTheEngineAssignedShapes(t *testing.T) {
	// The bench is the one seat that owes an OCCASION — its id is the same for four different
	// sittings — so its row carries one and every other row is empty. A blank here for `judge`
	// would be refused, which is the point of the field rather than a wrinkle in this test.
	for _, c := range []struct{ id, occasion string }{
		{"red-lens-evidence", ""}, {"red-chair", ""}, {"blue-lane-3", ""}, {"blue-respond", ""},
		{"blue-synthesize", ""}, {"frontier", ""}, {"judge", "docket"}, {"operator", ""},
	} {
		// A LANE NEEDS THE CAST AND THE OPERATOR MUST NOT HAVE ONE. A lane is dispatchable because
		// the run's cast names it, not because its id looks like one — so `blue-lane-3` gets a run
		// whose cast seats three lanes. The operator is not a cast member by design (CastFor never
		// writes it), and on a run that HAS a cast the membership gate refuses it, which is a
		// different gate from the shape one this loop is about.
		runDir := newRun(t)
		if c.id == "blue-lane-3" {
			runDir = newRunWithLanes(t, 3)
		}
		if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: c.id}, "", c.occasion); err != nil {
			t.Errorf("RegisterSeat(%q, occasion %q) = %v, want accepted", c.id, c.occasion, err)
		}
	}
}

// AND IT REFUSES WHAT NO DISPATCH PRODUCES, which is the half that makes the list above a roster
// rather than a sample.
//
// `a` USED TO BE IN THE ACCEPTED LIST. That is what the guard was: any string of safe characters,
// so `red-lens-banana` and a bare `a` both registered and bound. Register is the one call that
// takes a seat's word for who it is — after it, the wrong id is not a claim, it is the record.
func TestRegisterSeatRefusesAnIdNoDispatchProduces(t *testing.T) {
	for _, id := range []string{
		"a",                             // the old contract: any safe string
		"red-lens-banana",               // the prefix guard's blind spot
		"red-lens-r1",                   // a lens with no lens index
		"red-lens-evidence-oops",        // a real id with something appended
		"blue-r1",                       // invented, and it was live in three fixtures
		"judge-petition",                // the bare pre-#394 form: one shard for every sitting
		"judge-petition-judge-petition", // there is no sitting about a sitting
		"Red-Merge-R1",                  // the right shape in the wrong case
	} {
		runDir := newRun(t)
		if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: id}, "", ""); err == nil {
			t.Errorf("RegisterSeat(%q) was accepted; it binds for the whole run and no dispatch created it", id)
		}
	}
}

// THE ROSTER AND THE ROLE TABLE MUST NOT FORK. roleSeats matches PREFIXES for role lookup and
// the roster holds whole ids for legitimacy — two statements of one vocabulary, which is exactly
// the shape that drifts. Every roster id must land in its own role, and every role must have at
// least one id, so a role added to one table without the other fails here.
func TestTheRosterAndTheRoleTableAgree(t *testing.T) {
	covered := map[string]bool{}
	for id, s := range seatclass.Seats {
		if got := roleOfSeat(id); got != s.Role {
			t.Errorf("roster id %q is role %q by the roster and %q by roleSeats", id, s.Role, got)
		}
		covered[s.Role] = true
	}
	// The lane shape is the roster's one remaining pattern (#1153b retires it), so its sample
	// carries it here rather than being the one id nothing reconciles.
	if lane := SampleSeatOf("blue"); roleOfSeat(lane) != "blue" {
		t.Errorf("lane sample %q is role %q by roleSeats, not blue", lane, roleOfSeat(lane))
	}
	for role := range roleSeats {
		if !covered[role] {
			t.Errorf("role %q has a prefix in roleSeats and no shape in the roster, so every id under it registers unchecked", role)
		}
	}
}
