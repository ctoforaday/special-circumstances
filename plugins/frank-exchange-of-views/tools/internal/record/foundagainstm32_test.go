package record

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// m32Cited is the citation m32's four lenses read against its sentence, and sqliteCited the one
// sqlite-schema-source's record leaves found against at its end.
const (
	m32Cited      = "C-a36dce3f"
	sqliteArchive = "2026-08-22_sqlite-schema-source.tar.gz"
	sqliteCited   = "C-65ca0a9e"
)

// m32Readers is the four lenses whose readings of m32Cited stood until blue's edit.
var m32Readers = []string{"red-lens-adversary", "red-lens-architecture", "red-lens-computation", "red-lens-evidence"}

// m32BeforeTheEdit is the position of blue's edit reopening m32Cited in the migrated stream: the
// record before it holds the four readings standing.
func m32BeforeTheEdit(t *testing.T, whole []*Event) int {
	t.Helper()
	for i, e := range whole {
		if slices.Contains(e.GetBlueEdit().GetReopened(), m32Cited) {
			return i
		}
	}
	t.Fatalf("no migrated m32 edit reopens %s", m32Cited)
	return 0
}

// foundAgainstChecks is the run's found-against checks of anchor, by owner.
func foundAgainstChecks(t *testing.T, run Run, anchor string) map[string]Blocker {
	t.Helper()
	out := map[string]Blocker{}
	for _, b := range failedOf(t, run) {
		if b.Subject == anchor && b.CheckState() == CheckFoundAgainst {
			out[b.Owner] = b
		}
	}
	return out
}

func ownersOf(m map[string]Blocker) []string {
	var out []string
	for o := range m {
		out = append(out, o)
	}
	sort.Strings(out)
	return out
}

// THE MIGRATED M32 HOLDS A FAILED CHECK PER STANDING READING (§III.9 item 4). Before blue's edit, four
// lenses' archived `refutes` readings of C-a36dce3f stand, each a found-against check no gap carries;
// the edit and the `supports` re-reads clear them all. The two stored plans that held the archived
// contradictions hold failed_check blockers on the citation instead.
func TestTheMigratedM32HoldsAFailedCheckPerStandingReading(t *testing.T) {
	whole := archivedEvents(t, m32Archive)
	cut := m32BeforeTheEdit(t, whole)
	checks := foundAgainstChecks(t, archiveAt(t, m32Archive, whole, cut), m32Cited)
	if got := ownersOf(checks); !slices.Equal(got, m32Readers) {
		t.Errorf("before the edit, %s stands found against by %v, want %v", m32Cited, got, m32Readers)
	}
	for owner, b := range checks {
		if b.AnsweredBy() != "" {
			t.Errorf("%s's check is carried by %q; an archived reading names no gap", owner, b.AnsweredBy())
		}
	}
	if got := foundAgainstChecks(t, archiveAt(t, m32Archive, whole, len(whole)), m32Cited); len(got) != 0 {
		t.Errorf("at the end of m32, %s stands found against by %v, want none", m32Cited, ownersOf(got))
	}
	stored, err := StoredPlans(whole)
	if err != nil {
		t.Fatal(err)
	}
	holding := 0
	for _, sp := range stored {
		var owners []string
		for _, b := range sp.Plan.Blockers {
			if b.Kind == "contradiction" {
				t.Errorf("the plan at %d holds a blocker of kind contradiction: %+v", sp.At, b)
			}
			if b.Kind == BlockerFailedCheck && b.Subject == m32Cited {
				owners = append(owners, b.Owner)
			}
		}
		if len(owners) == 0 {
			continue
		}
		holding++
		sort.Strings(owners)
		if want := []string{"red-lens-adversary", "red-lens-architecture", "red-lens-computation"}; !slices.Equal(owners, want) {
			t.Errorf("the plan at %d holds failed checks of %s owned by %v, want %v", sp.At, m32Cited, owners, want)
		}
	}
	if holding != 2 {
		t.Errorf("%d stored plan(s) hold a failed check of %s; the two that held the archived contradictions do", holding, m32Cited)
	}
}

// archiveWith is the named archive after its event n, with acts appended — a copy a test may act on.
func archiveWith(t *testing.T, archive string, whole []*Event, n int, acts ...*Event) Run {
	t.Helper()
	dir := recordtest.RequireArchive(t, archive)
	for _, f := range []string{"record.db", "record.db-wal", "record.db-shm"} {
		if err := os.Remove(filepath.Join(dir, "records", f)); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	recordtest.Seed(t, dir, append(slices.Clone(whole[:n]), acts...)...)
	return mustRun(t, dir)
}

// A MIGRATED FOUND-AGAINST CHECK READIES BLUE ALONE (§III.3). A carrier-less check is blue's to
// remedy: the plan readies blue for it and no reader; each reader's list offers it and does not
// demand it, its envelope is not refused over it, it owes nothing for it, and its mint about the
// anchor is counted; every sentence that states it says the reading found the source against the
// sentence and names blue's remedy. Asserted on m32 before blue's edit, and on the whole of
// sqlite-schema-source, which ends with its citation found against (read-only: that run is closed).
func TestAMigratedFoundAgainstCheckReadiesBlueAlone(t *testing.T) {
	m32 := archivedEvents(t, m32Archive)
	m32Cut := m32BeforeTheEdit(t, m32)
	sqlite := archivedEvents(t, sqliteArchive)
	for _, c := range []struct {
		name, archive, anchor string
		whole                 []*Event
		n                     int
	}{
		{"m32 before the edit", m32Archive, m32Cited, m32, m32Cut},
		{"sqlite-schema-source at its end", sqliteArchive, sqliteCited, sqlite, len(sqlite)},
	} {
		t.Run(c.name, func(t *testing.T) {
			run := archiveAt(t, c.archive, c.whole, c.n)
			checks := foundAgainstChecks(t, run, c.anchor)
			if len(checks) == 0 {
				t.Fatalf("%s does not stand found against", c.anchor)
			}
			plan, err := PlanDispatch(run)
			if err != nil {
				t.Fatal(err)
			}
			passRefusal := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate)
			if passRefusal == nil {
				t.Fatalf("the PASS is admitted over %s, found against", c.anchor)
			}
			why := strings.Join(plan.Why, "\n")
			what := "citation " + c.anchor
			if !strings.Contains(why, "blue-respond: "+what+" stands found against") && !strings.Contains(why, "blue-respond has sat since "+what+"'s check failed") {
				t.Errorf("the plan's why does not name blue-respond for %s: %q", c.anchor, why)
			}
			for reader, b := range checks {
				// 1. dispatch: no reader is readied for it.
				for _, s := range []string{reader + "'s check of " + what, reader + " has sat since its check of " + what} {
					if strings.Contains(why, s) {
						t.Errorf("the plan's why readies %s for its check of %s: %q", reader, c.anchor, why)
					}
				}
				// 2. work list: offered, never demanded; the envelope does not name it.
				if blocking, _ := itemsOn(t, run, "lens", reader, c.anchor); len(blocking) != 0 {
					t.Errorf("%s's list demands %s: %+v", reader, c.anchor, blocking)
				}
				if refusal, err := EnvelopeRefusal(run, reader); err != nil || strings.Contains(refusal, c.anchor) {
					t.Errorf("%s's envelope refusal names %s: %q, %v", reader, c.anchor, refusal, err)
				}
				// 3. retirement: it owes nothing for it.
				if owes := foldOf(t, run, reader).owes; slices.ContainsFunc(owes, func(o string) bool { return strings.Contains(o, c.anchor) }) {
					t.Errorf("%s owes %q for a check blue answers", reader, owes)
				}
				// 4. budget: its mint about the anchor is counted.
				if got, err := answersOwnFailedCheckAt(recordOf(t, run), reader, aboutAnchor("G-00000009", c.anchor)); err != nil || got != "" {
					t.Errorf("%s's mint about %s answers its found-against check (%q, %v), so it would take no mint", reader, c.anchor, got, err)
				}
				// 5. the printed sentence.
				stated := map[string]string{
					"the work item":     b.WorkItem(),
					"the plan's reason": kindOf(b.Kind).why(b),
					"blue's item":       blueFailedItem(b),
					"the PASS refusal":  passRefusal.Error(),
				}
				for name, s := range stated {
					if !strings.Contains(s, "the source against the sentence") || !strings.Contains(s, "edits the sentence, or retires or replaces the citation") {
						t.Errorf("%s of %s's check does not say its reading found the source against the sentence and name blue's remedy: %q", name, reader, s)
					}
					for _, never := range []string{"owes its gap about the anchor", "takes no mint of its budget", "has minted no gap about it yet", "read nothing of its source"} {
						if strings.Contains(s, never) {
							t.Errorf("%s of %s's check says %q: %q", name, reader, never, s)
						}
					}
				}
			}
		})
	}

	// 6. What clears it, on copies of m32 before the edit: the reader's own backing reading a sitting
	// later, an edit naming the anchor, a retire of the citation. Another lens's backing reading is a
	// disagreement and clears nothing.
	reading := func(seat string) []*Event {
		return []*Event{
			recordtest.At(t, seat, seat+":register:m32-clear", &recordpb.Register{}),
			recordtest.At(t, seat, seat+":verify:m32-clear", &recordpb.Verify{Anchor: proto.String(m32Cited),
				Claim: proto.String("the sentence as it stands"), Text: proto.String("the source backs it"),
				Outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(),
				CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum()}),
		}
	}
	const reader = "red-lens-architecture"
	for _, c := range []struct {
		name string
		acts []*Event
		want []string
	}{
		{"the reader's own backing reading a sitting later", reading(reader), []string{"red-lens-adversary", "red-lens-computation", "red-lens-evidence"}},
		{"an edit naming the anchor", []*Event{recordtest.At(t, "blue-respond", "blue-respond:edit:m32-clear",
			&recordpb.BlueEdit{Old: proto.String("before"), New: proto.String("after"), Text: proto.String("why"), Reopened: []string{m32Cited}})}, nil},
		{"a retire of the citation", []*Event{recordtest.At(t, "blue-respond", "blue-respond:retire:m32-clear",
			&recordpb.Retire{Claim: proto.String("the sentence"), Reason: proto.String("the source says otherwise"), Anchors: []string{m32Cited}, ClaimKept: proto.Bool(true)})}, nil},
		{"another lens's backing reading", reading("red-lens-logic"), m32Readers},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := ownersOf(foundAgainstChecks(t, archiveWith(t, m32Archive, m32, m32Cut, c.acts...), m32Cited))
			if !slices.Equal(got, c.want) {
				t.Errorf("after %s, %s stands found against by %v, want %v", c.name, m32Cited, got, fmt.Sprint(c.want))
			}
		})
	}
}
