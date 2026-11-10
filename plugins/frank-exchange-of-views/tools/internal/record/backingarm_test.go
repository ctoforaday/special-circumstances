package record

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// ARM 4 OF THE LIVE-ANSWER DEMAND IS THE LEDGER'S HOLDS, OVER ORDERED SEQUENCES. The gap view's own
// re-check arm is SQL over rows; the anchor ledger is the Go fold every other reader of a check's
// standing asks. For every ordered sequence of up to three checks of the gap's anchor by its minter,
// made after the mint — beside one check by that seat before the mint
// and one by another seat after it — the two agree. The Go mirror is the ledger itself: the seat's
// check that stands (anchorStanding.stands, the same-seat replacement rule; no edit is in any
// sequence), at or after the mint, and holding.
//
// The shapes: under `source`, a gap minted about a citation (its minter's check); under `computation`,
// a gap minted about a computation.
// Sequences of one and two readings draw from every SourceOutcome × SourceCompleteness pair; those
// of three draw each reading from one outcome of each way a reading can stand (`supports` and `weak`
// back, `refutes` and `absent` read against, `unreachable` reads nothing), each of a copy whose
// completeness is absent or `not_the_work` — the only completeness either side reads. Each insert
// asks the seat's latest sitting of the whole record, so the sweep runs in chunks of one record each.
func TestTheBackingArmIsTheLedgersHolds(t *testing.T) {
	var pairs, third []check
	outcomes := recordpb.SourceOutcome(0).Descriptor().Values()
	completeness := recordpb.SourceCompleteness(0).Descriptor().Values()
	for i := 0; i < outcomes.Len(); i++ {
		o := recordpb.SourceOutcome(outcomes.Get(i).Number())
		if o == recordpb.SourceOutcome_SOURCE_OUTCOME_UNSPECIFIED {
			continue
		}
		for j := 0; j < completeness.Len(); j++ {
			pairs = append(pairs, check{outcome: o, compl: recordpb.SourceCompleteness(completeness.Get(j).Number())})
		}
	}
	for _, o := range []recordpb.SourceOutcome{recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS, recordpb.SourceOutcome_SOURCE_OUTCOME_WEAK,
		recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES, recordpb.SourceOutcome_SOURCE_OUTCOME_ABSENT, recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE} {
		for _, c := range []recordpb.SourceCompleteness{recordpb.SourceCompleteness_SOURCE_COMPLETENESS_UNSPECIFIED, recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_THE_WORK} {
			third = append(third, check{outcome: o, compl: c})
		}
	}
	judgements := []check{{sound: recordpb.Soundness_SOUNDNESS_SOUND}, {sound: recordpb.Soundness_SOUNDNESS_UNSOUND}}

	var scenarios []armScenario
	for _, shape := range []string{"source about"} {
		for _, seq := range sequencesOf(pairs, third) {
			scenarios = append(scenarios, armScenario{shape: shape, seq: seq})
		}
	}
	for _, seq := range sequencesOf(judgements, judgements) {
		scenarios = append(scenarios, armScenario{shape: "computation about", seq: seq})
	}
	const chunk = 150
	var met, all int
	var mu sync.Mutex
	t.Run("chunks", func(t *testing.T) {
		for lo := 0; lo < len(scenarios); lo += chunk {
			part := scenarios[lo:min(lo+chunk, len(scenarios))]
			t.Run(fmt.Sprint(lo), func(t *testing.T) {
				t.Parallel()
				m := runArmScenarios(t, part)
				mu.Lock()
				met, all = met+m, all+len(part)
				mu.Unlock()
			})
		}
	})
	if met == 0 || met == all {
		t.Fatalf("the ledger met %d of %d sequences: the sweep never tells the two answers apart", met, all)
	}
}

// check is one check a sequence makes: a reading's outcome and completeness, or a judgement.
type check struct {
	outcome recordpb.SourceOutcome
	compl   recordpb.SourceCompleteness
	sound   recordpb.Soundness
}

// armScenario is one gap of the sweep: its shape and the checks its seat makes after the mint.
type armScenario struct {
	shape string
	seq   []check
}

// sequencesOf is every ordered sequence of length 0 to 3: lengths 1 and 2 over all, 3 over third.
func sequencesOf(all, third []check) [][]check {
	out := [][]check{nil}
	for _, a := range all {
		out = append(out, []check{a})
		for _, b := range all {
			out = append(out, []check{a, b})
		}
	}
	for _, a := range third {
		for _, b := range third {
			for _, c := range third {
				out = append(out, []check{a, b, c})
			}
		}
	}
	return out
}

// runArmScenarios writes each scenario into one fresh record, asks arm 4 of every gap and the ledger
// of every anchor, fails each disagreement, and returns how many the ledger met.
func runArmScenarios(t *testing.T, scenarios []armScenario) int {
	t.Helper()
	db := scratchStore(t)
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	add := func(seat string, body proto.Message) {
		ev := &recordpb.Event{SeatId: proto.String(seat), Ts: proto.String("2026-01-01T00:00:00Z"), Key: proto.String(fmt.Sprintf("act:%d", n))}
		n++
		typ, err := recordpb.SetBody(ev, body)
		if err != nil {
			t.Fatal(err)
		}
		ev.Type = &typ
		if _, err := recordsql.InsertTx(tx, ev); err != nil {
			t.Fatal(err)
		}
	}
	const owner, other = "red-lens-evidence", "red-lens-logic"
	reading := func(seat, anchor string, c check) {
		v := &recordpb.Verify{Claim: proto.String("c"), Anchor: proto.String(anchor), Outcome: c.outcome.Enum(),
			Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(), Text: proto.String("t")}
		if c.compl != recordpb.SourceCompleteness_SOURCE_COMPLETENESS_UNSPECIFIED {
			v.SourceCompleteness = c.compl.Enum()
		}
		add(seat, v)
	}
	judgement := func(seat, anchor string, c check) {
		add(seat, &recordpb.Reproduce{ComputationId: proto.String(anchor), Soundness: c.sound.Enum(), Ran: proto.Bool(true)})
	}
	mint := func(gap string, kind recordpb.CheckKind, about string) {
		m := &recordpb.Mint{ClassMaterial: recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE.Enum(), Severity: recordpb.Grade_GRADE_MEDIUM.Enum(),
			GapId: proto.String(gap), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(),
			Problem: proto.String("p"), AcceptanceCheck: proto.String("a"), CheckKind: kind.Enum(),
			Likelihood: recordpb.Grade_GRADE_MEDIUM.Enum(), Impact: recordpb.Grade_GRADE_MEDIUM.Enum(),
			AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings")}
		if about != "" {
			m.AboutKind, m.AboutRef = recordpb.AboutKind_ABOUT_KIND_ANCHOR.Enum(), proto.String(about)
		}
		add(owner, m)
	}
	backing := check{outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS}
	sound := check{sound: recordpb.Soundness_SOUNDNESS_SOUND}
	gaps, anchors := make([]string, len(scenarios)), make([]string, len(scenarios))
	for i, sc := range scenarios {
		gaps[i] = fmt.Sprintf("G-%08x", i+1)
		switch sc.shape {
		case "computation about":
			anchors[i] = fmt.Sprintf("K-%08x", i+1)
			add("blue-respond", &recordpb.Computation{ComputationId: proto.String(anchors[i]), Basis: recordpb.Basis_BASIS_REPRODUCIBLE.Enum(), Argv: []string{"true"}})
			judgement(owner, anchors[i], sound) // before the mint: never meets the arm
			mint(gaps[i], recordpb.CheckKind_CHECK_KIND_COMPUTATION, anchors[i])
			judgement(other, anchors[i], sound) // another seat's: never meets the arm
			for _, c := range sc.seq {
				judgement(owner, anchors[i], c)
			}
		default:
			anchors[i] = fmt.Sprintf("C-%08x", i+1)
			add("blue-respond", &recordpb.Cite{Label: proto.String(anchors[i])})
			reading(owner, anchors[i], backing) // before the mint: never meets the arm
			mint(gaps[i], recordpb.CheckKind_CHECK_KIND_SOURCE, anchors[i])
			reading(other, anchors[i], backing) // another seat's: never meets the arm
			for _, c := range sc.seq {
				reading(owner, anchors[i], c)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	answered := map[string]bool{}
	rows, err := db.Query(`SELECT "gap_id", "evidence_answered" FROM "gap"`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var g string
		var a bool
		if err := rows.Scan(&g, &a); err != nil {
			t.Fatal(err)
		}
		answered[g] = a
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	evs, win, err := eventsAt(db)
	if err != nil {
		t.Fatal(err)
	}
	minted := map[string]int64{}
	for _, e := range evs {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
			minted[m.GetGapId()] = win.Of(e).ID
		}
	}
	ledger := anchorChecksOf(evs, win)
	met := 0
	for i, sc := range scenarios {
		s := ledger.at(anchors[i])
		want := false
		for _, c := range s.checks {
			if c.seat == owner && s.stands(c) && c.id >= minted[gaps[i]] && c.holds {
				want = true
			}
		}
		if want {
			met++
		}
		if got := answered[gaps[i]]; got != want {
			t.Errorf("%s gap after %+v: arm 4 says %v, the ledger %v", sc.shape, sc.seq, got, want)
		}
	}
	return met
}

// EVERY EVIDENCE ANCHOR ON EVERY ARCHIVE IS THE LEDGER'S. The evidence_anchor view and the anchor
// ledger are two statements of which citations and computations a record holds and which are
// retired; on every archived run, migrated, they name the same anchors and agree on each one's
// retirement.
func TestEvidenceAnchorIsTheLedgersAnchorSet(t *testing.T) {
	if archiveMigrate == nil {
		t.Fatal("no migration installed: archivemigrate_test.go installs migrate before any test runs")
	}
	for _, name := range archiveNames(t) {
		t.Run(strings.TrimSuffix(name, ".tar.gz"), func(t *testing.T) {
			t.Parallel()
			migrated := filepath.Join(t.TempDir(), "run")
			if err := archiveMigrate(recordtest.RequireArchive(t, name), migrated); err != nil {
				t.Fatalf("migrating %s: %v", name, err)
			}
			if err := recordsql.CloseUnder(migrated); err != nil {
				t.Fatal(err)
			}
			db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(filepath.Join(migrated, "records", "record.db"))+"?mode=ro")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			evs, win, err := eventsAt(db)
			if err != nil {
				t.Fatal(err)
			}
			want := map[string]bool{}
			for _, s := range anchorChecksOf(evs, win).order {
				want[s.anchor] = s.retired
			}
			got := map[string]bool{}
			rows, err := db.Query(`SELECT "anchor_id", "retired_event" IS NOT NULL FROM "evidence_anchor"`)
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			for rows.Next() {
				var id string
				var retired bool
				if err := rows.Scan(&id, &retired); err != nil {
					t.Fatal(err)
				}
				if _, dup := got[id]; dup {
					t.Errorf("%s: evidence_anchor holds %s twice", name, id)
				}
				got[id] = retired
			}
			for id, r := range want {
				g, ok := got[id]
				switch {
				case !ok:
					t.Errorf("%s: the ledger holds %s and evidence_anchor does not", name, id)
				case g != r:
					t.Errorf("%s: %s is retired=%v in the ledger and %v in evidence_anchor", name, id, r, g)
				}
			}
			for id := range got {
				if _, ok := want[id]; !ok {
					t.Errorf("%s: evidence_anchor holds %s and the ledger does not", name, id)
				}
			}
		})
	}
}

// archiveNames is every archived run's tarball name under the repository's run-archive.
func archiveNames(t *testing.T) []string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for d := wd; ; d = filepath.Dir(d) {
		got, _ := filepath.Glob(filepath.Join(d, "run-archive", "*.tar.gz"))
		if len(got) > 0 {
			for i := range got {
				got[i] = filepath.Base(got[i])
			}
			return got
		}
		if filepath.Dir(d) == d {
			t.Fatal("no run-archive above this checkout")
		}
	}
}
