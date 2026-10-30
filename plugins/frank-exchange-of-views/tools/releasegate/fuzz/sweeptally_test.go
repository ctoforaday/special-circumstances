package fuzz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE SWEEP RUNS IN PARTS, AND ITS GATES READ ONE TALLY (#1616).
//
// The pull request's release-fuzz job splits TestFuzzDebate's seeds across parallel jobs: part k of
// K runs the seeds where (seed-1)%K == k, numbered as the whole sweep numbers them, so every forced
// seed stays the seed its constant names. A part runs its seeds, fails on its own failed runs, and
// writes what it measured to FUZZ_TALLY; it judges no sweep gate, because a part under the quorum
// would report NOT MEASURED and pass. TestFuzzSweepTally reads every part's tally from
// FUZZ_TALLIES, refuses a set that is not one whole sweep, and judges the gates over the merge.
//
// ONE TYPE, BOTH PATHS. The single-process sweep folds the same runTally values and judges the same
// processTally, so the gates read one shape whether the sweep ran in one process or in K.
//
// THE ONCE-PER-SWEEP DRIVES ARM IN PART 0 ONLY (gblock, 2026-10-28): `close --as moot`, the engage
// refusal and the deliberate class mints each belong to the sweep, and arming them in every part
// would drive each K times.

// sweepPart is which part of the sweep this process runs: K of Of.
type sweepPart struct{ K, Of int }

// wholeSweep is the sweep run in one process.
var wholeSweep = sweepPart{K: 0, Of: 1}

// parseSweepPart reads FUZZ_PART: "k/K", two unsigned decimal integers with 0 <= k < K. Empty is
// the whole sweep. Anything else is refused rather than read as some part.
func parseSweepPart(s string) (sweepPart, error) {
	if s == "" {
		return wholeSweep, nil
	}
	ks, ofs, ok := strings.Cut(s, "/")
	if !ok || !isDecimal(ks) || !isDecimal(ofs) {
		return sweepPart{}, fmt.Errorf("FUZZ_PART=%q is not k/K, two unsigned decimal integers", s)
	}
	k, kerr := strconv.Atoi(ks)
	of, oferr := strconv.Atoi(ofs)
	if kerr != nil || oferr != nil {
		return sweepPart{}, fmt.Errorf("FUZZ_PART=%q is out of range", s)
	}
	if of < 1 || k >= of {
		return sweepPart{}, fmt.Errorf("FUZZ_PART=%q names part %d of %d; a part is 0 to K-1 of K >= 1", s, k, of)
	}
	return sweepPart{K: k, Of: of}, nil
}

// isDecimal is a non-empty run of ASCII digits with no leading zero, so each part has one spelling.
func isDecimal(s string) bool {
	if s == "" || (len(s) > 1 && s[0] == '0') {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// owns reports whether seed is this part's.
func (p sweepPart) owns(seed int64) bool { return (seed-1)%int64(p.Of) == int64(p.K) }

// split reports whether this process runs one part of several.
func (p sweepPart) split() bool { return p.Of > 1 }

// armsSweepDrives reports whether this process carries the once-per-sweep drives.
func (p sweepPart) armsSweepDrives() bool { return p.K == 0 }

// runTally is what the sweep's fold reads from one run.
type runTally struct {
	Seed                       int64                     `json:"seed"`
	Err                        string                    `json:"err"`
	Verdict                    string                    `json:"verdict"`
	Why                        string                    `json:"why"`
	Epochs                     int                       `json:"epochs"`
	EpochLimit                 bool                      `json:"epoch_limit"`
	Dialectic                  map[string]int            `json:"dialectic"`
	BlueSittings               int                       `json:"blue_sittings"`
	BlueAgain                  int                       `json:"blue_again"`
	CiteAnchors                int                       `json:"cite_anchors"`
	CacheFiles                 int                       `json:"cache_files"`
	EditAnswers                int                       `json:"edit_answers"`
	VerifiedBasis              int                       `json:"verified_basis"`
	FoundByMints               int                       `json:"found_by_mints"`
	VerbatimApplied            int                       `json:"verbatim_applied"`
	TypedVerbatim              int                       `json:"typed_verbatim"`
	EstoppelMisses             map[string]int            `json:"estoppel_misses"`
	ApplyMisses                map[string]int            `json:"apply_misses"`
	Estoppels                  int                       `json:"estoppels"`
	InquiryRecords             int                       `json:"inquiry_records"`
	ClosingsFiled              int                       `json:"closings_filed"`
	ClosingsRefused            int                       `json:"closings_refused"`
	ClosingWaits               int                       `json:"closing_waits"`
	PositionsWithheld          int                       `json:"positions_withheld"`
	OrdersRefused              int                       `json:"orders_refused"`
	SpansRefused               int                       `json:"spans_refused"`
	CopiesRead                 map[string]map[string]int `json:"copies_read"`
	FragmentsDriven            int                       `json:"fragments_driven"`
	TruncationsDriven          int                       `json:"truncations_driven"`
	QuotesRefused              int                       `json:"quotes_refused"`
	BridgesRefused             int                       `json:"bridges_refused"`
	BridgesNamed               int                       `json:"bridges_named"`
	ReopenedReadied            int                       `json:"reopened_readied"`
	ReopenedPlans              int                       `json:"reopened_plans"`
	UnsoundReadied             int                       `json:"unsound_readied"`
	UnsoundPlans               int                       `json:"unsound_plans"`
	UnsoundAnswers             int                       `json:"unsound_answers"`
	UnsoundRepeats             int                       `json:"unsound_repeats"`
	UnsoundOverOpen            int                       `json:"unsound_over_open"`
	ComputationReopenedReadied map[string]int            `json:"computation_reopened_readied"`
	OwedKept                   int                       `json:"owed_kept"`
	ComputationOwedKept        int                       `json:"computation_owed_kept"`
	OwedKeptBy                 map[string]int            `json:"owed_kept_by"`
	OwnerWhys                  []string                  `json:"owner_whys,omitempty"`
	OwnerOffered               int                       `json:"owner_offered"`
	CitationsOffered           int                       `json:"citations_offered"`
	UnreadAnswers              int                       `json:"unread_answers"`
	UnreadRepeats              int                       `json:"unread_repeats"`
	UnreadRetires              int                       `json:"unread_retires"`
	UnreadRetiresJoined        int                       `json:"unread_retires_joined"`
	UnreadClosesRefused        int                       `json:"unread_closes_refused"`
	UnsoundClosesRefused       int                       `json:"unsound_closes_refused"`
	RepeatsRefused             int                       `json:"repeats_refused"`
	BudgetRefused              int                       `json:"budget_refused"`
	OneGapRefused              int                       `json:"one_gap_refused"`
	OneGapFindings             int                       `json:"one_gap_findings"`
	OwnAnswers                 int                       `json:"own_answers"`
	UnclassifiedMinted         int                       `json:"unclassified_minted"`
	UnclassifiedRefused        int                       `json:"unclassified_refused"`
	RulerFilingsRefused        int                       `json:"ruler_filings_refused"`
	ClosedRemandsRefused       int                       `json:"closed_remands_refused"`
	TrifleRemandsRouted        int                       `json:"trifle_remands_routed"`
	BlueRemandActs             int                       `json:"blue_remand_acts"`
	LensRemandActs             int                       `json:"lens_remand_acts"`
	PastLimitRemandsRefused    int                       `json:"past_limit_remands_refused"`
	LensFirstHolds             int                       `json:"lens_first_holds"`
	BenchRegrades              int                       `json:"bench_regrades"`
	LockRefused                int                       `json:"lock_refused"`
	LockLifted                 int                       `json:"lock_lifted"`
	GroundsRefused             int                       `json:"grounds_refused"`
	GroundsAdmitted            int                       `json:"grounds_admitted"`
	PartRefused                map[string]int            `json:"part_refused"`
	PartsRecorded              int                       `json:"parts_recorded"`
}

// tallyOf is the run's tally.
func tallyOf(o outcome) runTally {
	return runTally{
		Seed: o.seed, Err: o.err, Verdict: o.verdict, Why: o.why, Epochs: o.epochs, EpochLimit: o.epochLimit,
		Dialectic: o.dialectic, BlueSittings: o.blueSittings, BlueAgain: o.blueAgain,
		CiteAnchors: o.citeAnchors, CacheFiles: o.cacheFiles, EditAnswers: o.editAnswers,
		VerifiedBasis: o.verifiedBasis, FoundByMints: o.foundByMints, VerbatimApplied: o.verbatimApplied,
		TypedVerbatim: o.typedVerbatim, EstoppelMisses: o.estoppelMisses, ApplyMisses: o.applyMisses,
		Estoppels: o.estoppels, InquiryRecords: o.inquiryRecords, ClosingsFiled: o.closingsFiled,
		ClosingsRefused: o.closingsRefused, ClosingWaits: o.closingWaits, PositionsWithheld: o.positionsWithheld,
		OrdersRefused: o.ordersRefused, SpansRefused: o.spansRefused, CopiesRead: o.copiesRead,
		FragmentsDriven: o.fragmentsDriven, TruncationsDriven: o.truncationsDriven, QuotesRefused: o.quotesRefused,
		BridgesRefused: o.bridgesRefused, BridgesNamed: o.bridgesNamed, ReopenedReadied: o.reopenedReadied,
		ReopenedPlans: o.reopenedPlans, UnsoundReadied: o.unsoundReadied, UnsoundPlans: o.unsoundPlans,
		UnsoundAnswers: o.unsoundAnswers, UnsoundRepeats: o.unsoundRepeats, UnsoundOverOpen: o.unsoundOverOpen,
		ComputationReopenedReadied: o.computationReopenedReadied, OwedKept: o.owedKept, ComputationOwedKept: o.computationOwedKept,
		OwedKeptBy: o.owedKeptBy, OwnerWhys: o.ownerWhys, OwnerOffered: o.ownerOffered, CitationsOffered: o.citationsOffered,
		UnreadAnswers: o.unreadAnswers, UnreadRepeats: o.unreadRepeats, UnreadRetires: o.unreadRetires, UnreadRetiresJoined: o.unreadRetiresJoined,
		UnreadClosesRefused: o.unreadClosesRefused, UnsoundClosesRefused: o.unsoundClosesRefused, RepeatsRefused: o.repeatsRefused, BudgetRefused: o.budgetRefused, OwnAnswers: o.ownAnswers,
		OneGapRefused: o.oneGapRefused, OneGapFindings: o.oneGapFindings,
		UnclassifiedMinted: o.unclassifiedMinted, UnclassifiedRefused: o.unclassifiedRefused,
		RulerFilingsRefused: o.rulerFilingsRefused, ClosedRemandsRefused: o.closedRemandsRefused,
		TrifleRemandsRouted: o.trifleRemandsRouted, BlueRemandActs: o.blueRemandActs, LensRemandActs: o.lensRemandActs,
		PastLimitRemandsRefused: o.pastLimitRemandsRefused, LensFirstHolds: o.lensFirstHolds,
		BenchRegrades: o.benchRegrades, LockRefused: o.lockRefused, LockLifted: o.lockLifted, GroundsRefused: o.groundsRefused, GroundsAdmitted: o.groundsAdmitted,
		PartRefused: o.partRefused, PartsRecorded: o.partsRecorded,
	}
}

// processTally is the package-level state the sweep's gates read: what every run in this process
// left outside its own outcome. Counts add across parts, sets union, MootSpent ORs, and a
// first-refusal row keeps the earlier part's.
type processTally struct {
	ExecRuns            map[string]int             `json:"exec_runs"`
	ExecFail            map[string]int             `json:"exec_fail"`
	ExecFailWhy         map[string]string          `json:"exec_fail_why"`
	ExecInProc          map[string]int             `json:"exec_in_proc"`
	ExecEdges           map[string]map[string]int  `json:"exec_edges"`
	ExecFlags           map[string]map[string]bool `json:"exec_flags"`
	ExecValues          map[string]map[string]bool `json:"exec_values"`
	NamingWalked        int                        `json:"naming_walked"`
	NamingWrong         map[string]string          `json:"naming_wrong"`
	DrivenView          map[string]bool            `json:"driven_view"`
	MootSpent           bool                       `json:"moot_spent"`
	LookedFirst         int64                      `json:"looked_first"`
	RegisteredTwice     int64                      `json:"registered_twice"`
	SittingsClosed      int64                      `json:"sittings_closed"`
	SittingsLeftOpen    int64                      `json:"sittings_left_open"`
	AdmittedNeverClosed int64                      `json:"admitted_never_closed"`
	FailedSittings      int64                      `json:"failed_sittings"`
	ClassLanded         map[string]bool            `json:"class_landed"`
	Either              map[string][2]int          `json:"either"`
	Envelope            map[string]int             `json:"envelope"`
	Fail                map[string]int             `json:"fail"`
}

// add folds q into p.
func (p *processTally) add(q processTally) {
	p.ExecRuns = addCounts(p.ExecRuns, q.ExecRuns)
	p.ExecFail = addCounts(p.ExecFail, q.ExecFail)
	p.ExecFailWhy = addFirst(p.ExecFailWhy, q.ExecFailWhy)
	p.ExecInProc = addCounts(p.ExecInProc, q.ExecInProc)
	if p.ExecEdges == nil {
		p.ExecEdges = map[string]map[string]int{}
	}
	for k, v := range q.ExecEdges {
		p.ExecEdges[k] = addCounts(p.ExecEdges[k], v)
	}
	p.ExecFlags = addNestedSet(p.ExecFlags, q.ExecFlags)
	p.ExecValues = addNestedSet(p.ExecValues, q.ExecValues)
	p.NamingWalked += q.NamingWalked
	p.NamingWrong = addFirst(p.NamingWrong, q.NamingWrong)
	p.DrivenView = addSet(p.DrivenView, q.DrivenView)
	p.MootSpent = p.MootSpent || q.MootSpent
	p.LookedFirst += q.LookedFirst
	p.RegisteredTwice += q.RegisteredTwice
	p.SittingsClosed += q.SittingsClosed
	p.SittingsLeftOpen += q.SittingsLeftOpen
	p.AdmittedNeverClosed += q.AdmittedNeverClosed
	p.FailedSittings += q.FailedSittings
	p.ClassLanded = addSet(p.ClassLanded, q.ClassLanded)
	if p.Either == nil {
		p.Either = map[string][2]int{}
	}
	for k, v := range q.Either {
		row := p.Either[k]
		p.Either[k] = [2]int{row[0] + v[0], row[1] + v[1]}
	}
	p.Envelope = addCounts(p.Envelope, q.Envelope)
	p.Fail = addCounts(p.Fail, q.Fail)
}

func addCounts(into, from map[string]int) map[string]int {
	if into == nil {
		into = map[string]int{}
	}
	for k, v := range from {
		into[k] += v
	}
	return into
}

func addSet(into, from map[string]bool) map[string]bool {
	if into == nil {
		into = map[string]bool{}
	}
	for k, v := range from {
		if v {
			into[k] = true
		}
	}
	return into
}

func addNestedSet(into, from map[string]map[string]bool) map[string]map[string]bool {
	if into == nil {
		into = map[string]map[string]bool{}
	}
	for k, v := range from {
		into[k] = addSet(into[k], v)
	}
	return into
}

// addFirst keeps into's row where both hold one.
func addFirst(into, from map[string]string) map[string]string {
	if into == nil {
		into = map[string]string{}
	}
	for k, v := range from {
		if _, held := into[k]; !held {
			into[k] = v
		}
	}
	return into
}

// snapshotProcess copies this process's package-level state, each piece under its own lock.
func snapshotProcess() processTally {
	var p processTally
	p.add(processTally{})
	execMu.Lock()
	p.add(processTally{ExecRuns: execRuns, ExecFail: execFail, ExecFailWhy: execFailWhy, ExecInProc: execInProc,
		ExecEdges: execEdges, ExecFlags: execFlags, ExecValues: execValues})
	execMu.Unlock()
	namingMu.Lock()
	p.add(processTally{NamingWalked: namingWalked, NamingWrong: namingWrong})
	namingMu.Unlock()
	drivenViewMu.Lock()
	p.add(processTally{DrivenView: drivenView})
	drivenViewMu.Unlock()
	classDrive.mu.Lock()
	p.add(processTally{ClassLanded: classDrive.landed})
	classDrive.mu.Unlock()
	eitherMu.Lock()
	for k, row := range eitherRow {
		p.add(processTally{Either: map[string][2]int{k: *row}})
	}
	eitherMu.Unlock()
	envelopeMu.Lock()
	p.add(processTally{Envelope: envelopeRow})
	envelopeMu.Unlock()
	failMu.Lock()
	p.add(processTally{Fail: failRow})
	failMu.Unlock()
	p.MootSpent = mootSpent.Load()
	p.LookedFirst = lookedFirst.Load()
	p.RegisteredTwice = registeredTwice.Load()
	p.SittingsClosed = sittingsClosed.Load()
	p.SittingsLeftOpen = sittingsLeftOpen.Load()
	p.AdmittedNeverClosed = admittedNeverClosed.Load()
	p.FailedSittings = failedSittingsClosed.Load()
	return p
}

// sweepTally is one part's file: which part it is, of how many, of a sweep of how many seeds, and
// what it measured.
type sweepTally struct {
	Part    int          `json:"part"`
	Parts   int          `json:"parts"`
	N       int          `json:"n"`
	Runs    []runTally   `json:"runs"`
	Process processTally `json:"process"`
}

// writeTally writes st to path.
func writeTally(path string, st sweepTally) error {
	b, err := json.Marshal(st)
	if err != nil {
		return fmt.Errorf("encode the part's tally: %w", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write the part's tally: %w", err)
	}
	return nil
}

// decodeTally reads one part's file, refusing a field the type does not hold and anything after
// the one value.
func decodeTally(b []byte) (sweepTally, error) {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var st sweepTally
	if err := dec.Decode(&st); err != nil {
		return sweepTally{}, err
	}
	if dec.More() {
		return sweepTally{}, fmt.Errorf("more than one value")
	}
	return st, nil
}

// loadTallies reads every part's tally in dir and merges them into one sweep. It refuses — before
// any gate reads the merge — a set that is not exactly one whole sweep: no file, a file that is not
// a tally, parts that disagree on how many parts or seeds there are, a part twice or not at all, a
// run in a part that does not own its seed, and seeds other than 1..N once each. A missing part is
// the refusal that matters most: its runs would leave the merge under the quorum, and the gates
// would log NOT MEASURED and pass.
func loadTallies(dir string) (sweepTally, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return sweepTally{}, fmt.Errorf("read the tallies: %w", err)
	}
	var parts []sweepTally
	var names []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			return sweepTally{}, fmt.Errorf("%s is not a part's tally (a .json file); the directory holds tallies only", e.Name())
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return sweepTally{}, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		st, err := decodeTally(b)
		if err != nil {
			return sweepTally{}, fmt.Errorf("%s is not a part's tally: %w", e.Name(), err)
		}
		parts = append(parts, st)
		names = append(names, e.Name())
	}
	if len(parts) == 0 {
		return sweepTally{}, fmt.Errorf("%s holds no tally", dir)
	}
	of, n := parts[0].Parts, parts[0].N
	if of < 1 || n < 1 {
		return sweepTally{}, fmt.Errorf("%s states %d part(s) of a %d-seed sweep", names[0], of, n)
	}
	byPart := make([]int, of)
	for i := range byPart {
		byPart[i] = -1
	}
	for i, st := range parts {
		switch {
		case st.Parts != of:
			return sweepTally{}, fmt.Errorf("%s states %d part(s) and %s states %d", names[0], of, names[i], st.Parts)
		case st.N != n:
			return sweepTally{}, fmt.Errorf("%s states a %d-seed sweep and %s states %d", names[0], n, names[i], st.N)
		case st.Part < 0 || st.Part >= of:
			return sweepTally{}, fmt.Errorf("%s states part %d of %d", names[i], st.Part, of)
		case byPart[st.Part] >= 0:
			return sweepTally{}, fmt.Errorf("part %d appears twice: %s and %s", st.Part, names[byPart[st.Part]], names[i])
		}
		byPart[st.Part] = i
	}
	var missing []string
	for k, i := range byPart {
		if i < 0 {
			missing = append(missing, strconv.Itoa(k))
		}
	}
	if len(missing) > 0 {
		return sweepTally{}, fmt.Errorf("part(s) %s of %d hold no tally; a sweep without them is not the sweep", strings.Join(missing, ", "), of)
	}
	merged := sweepTally{Parts: of, N: n}
	merged.Process.add(processTally{})
	seen := map[int64]string{}
	for k, i := range byPart {
		st, part := parts[i], sweepPart{K: k, Of: of}
		for _, r := range st.Runs {
			if r.Seed < 1 || r.Seed > int64(n) {
				return sweepTally{}, fmt.Errorf("%s holds seed %d, outside 1..%d", names[i], r.Seed, n)
			}
			if !part.owns(r.Seed) {
				return sweepTally{}, fmt.Errorf("%s holds seed %d, which part %d of %d does not run", names[i], r.Seed, k, of)
			}
			if prev, dup := seen[r.Seed]; dup {
				return sweepTally{}, fmt.Errorf("seed %d appears twice: %s and %s", r.Seed, prev, names[i])
			}
			seen[r.Seed] = names[i]
			merged.Runs = append(merged.Runs, r)
		}
		merged.Process.add(st.Process)
	}
	if len(seen) != n {
		var absent []string
		for s := int64(1); s <= int64(n); s++ {
			if _, ok := seen[s]; !ok {
				absent = append(absent, strconv.FormatInt(s, 10))
			}
		}
		return sweepTally{}, fmt.Errorf("%d of %d seeds hold no run: %s", len(absent), n, strings.Join(absent, " "))
	}
	sort.Slice(merged.Runs, func(a, b int) bool { return merged.Runs[a].Seed < merged.Runs[b].Seed })
	return merged, nil
}

// TestFuzzSweepTally judges the sweep's gates over the parts' tallies in FUZZ_TALLIES. It runs
// in the release-fuzz-tally CI job, after every part of the release-fuzz job.
func TestFuzzSweepTally(t *testing.T) {
	dir := os.Getenv("FUZZ_TALLIES")
	if dir == "" {
		t.Skip("judges a split sweep's parts: CI's release-fuzz-tally job runs it as FUZZ_TALLIES=<dir> go test -v -count=1 -run '^TestFuzzSweepTally$' ./releasegate/fuzz/")
	}
	st, err := loadTallies(dir)
	if err != nil {
		t.Fatalf("the parts' tallies are not one whole sweep, so no gate is judged: %v", err)
	}
	t.Logf("judging %d run(s) of a %d-seed sweep from %d part(s)", len(st.Runs), st.N, st.Parts)
	judgeSweep(t, foldRuns(st.Runs), st.Process)
	var failed []string
	for _, r := range st.Runs {
		if r.Err != "" {
			failed = append(failed, strconv.FormatInt(r.Seed, 10))
		}
	}
	if len(failed) > 0 {
		t.Errorf("%d/%d fuzz runs failed in their parts (seeds %s) — each part's log names its failures; reproduce with that seed",
			len(failed), st.N, strings.Join(failed, " "))
	}
}

// sweepFold is the sweep's runs folded for its gates.
type sweepFold struct {
	completed, failed                               int
	verdicts                                        map[string]int
	blueSits, blueAgains                            int
	epochHist                                       map[int]int
	limitExits                                      int
	limitForced                                     bool
	whyHist, dcov                                   map[string]int
	citeAnchors, cacheFiles, editAnswers            int
	forcedVerdict, forcedWhy                        string
	verifiedVerdict, verifiedWhy                    string
	verifiedBasis, foundByMints                     int
	verbatimApplied, typedVerbatim, estoppels       int
	inquiryRecords, quotesRefused                   int
	bridgesRefused, bridgesNamed                    int
	ordersRefused, spansRefused, forcedSpansRefused int
	copiesRead                                      map[string]map[string]int
	closingsFiled, closingsRefused, closingWaits    int
	positionsWithheld                               int
	fragmentsDriven, truncationsDriven              int
	reopenedReadied, reopenedPlans                  int
	computationReopenedReadied, owedKeptBy          map[string]int
	owedKept, computationOwedKept                   int
	ownerOfferedRuns, citationOfferedRuns           int
	unreadAnswers, unreadRepeats                    int
	unreadRetires                                   int
	unreadRetiresJoined                             int
	unreadClosesRefused, unsoundClosesRefused       int
	repeatsRefused, budgetRefused, ownAnswers       int
	oneGapRefused, oneGapFindings                   int
	unsoundReadied, unsoundPlans, unsoundAnswers    int
	forcedUnsoundReadied, forcedComputationOwedKept int
	forcedOwnerWhys                                 []string
	unsoundRepeats, unsoundOverOpen                 int
	unclassifiedMinted, unclassifiedRefused         int
	rulerFilingsRefused                             int
	closedRemandsRefused, trifleRemandsRouted       int
	blueRemandActs, lensRemandActs                  int
	pastLimitRemandsRefused, lensFirstHolds         int
	benchRegrades, lockRefused, lockLifted          int
	groundsRefused, groundsAdmitted                 int
	applyMisses, estoppelMisses                     map[string]int
	partRefused                                     map[string]int
	partsRecorded, multiPartRuns                    int
}

// foldRuns folds the sweep's runs, in any order.
func foldRuns(runs []runTally) sweepFold {
	f := sweepFold{
		verdicts: map[string]int{}, epochHist: map[int]int{}, whyHist: map[string]int{}, dcov: map[string]int{},
		copiesRead: map[string]map[string]int{}, computationReopenedReadied: map[string]int{}, owedKeptBy: map[string]int{},
		applyMisses: map[string]int{}, estoppelMisses: map[string]int{}, partRefused: map[string]int{},
	}
	for _, r := range runs {
		if r.Err != "" {
			f.failed++
		}
		f.completed++
		f.verdicts[r.Verdict]++
		f.blueSits += r.BlueSittings
		f.blueAgains += r.BlueAgain
		// THE FORCED RUN'S OWN VERDICT, kept rather than folded into the histogram. A
		// declining drive has to be reported as itself, and "the forced run ended CEILING"
		// is the whole diagnosis; a count says only that nothing reached the word.
		if r.Seed == unverifiedSeed {
			f.forcedVerdict, f.forcedWhy = r.Verdict, r.Why
		}
		if r.Seed == verifiedSeed {
			f.verifiedVerdict, f.verifiedWhy = r.Verdict, r.Why
		}
		f.epochHist[r.Epochs]++
		if r.EpochLimit {
			f.limitExits++
			f.limitForced = f.limitForced || r.Seed == limitSeed
		}
		if r.Why != "" {
			f.whyHist[r.Verdict+" ← "+gapIDPat.ReplaceAllString(r.Why, "G-<id>")]++
		}
		for k, v := range r.Dialectic {
			f.dcov[k] += v
		}
		f.citeAnchors += r.CiteAnchors
		f.cacheFiles += r.CacheFiles
		f.editAnswers += r.EditAnswers
		f.verifiedBasis += r.VerifiedBasis
		f.foundByMints += r.FoundByMints
		f.verbatimApplied += r.VerbatimApplied
		f.typedVerbatim += r.TypedVerbatim
		f.estoppels += r.Estoppels
		f.inquiryRecords += r.InquiryRecords
		f.closingsFiled, f.closingsRefused, f.closingWaits = f.closingsFiled+r.ClosingsFiled, f.closingsRefused+r.ClosingsRefused, f.closingWaits+r.ClosingWaits
		f.positionsWithheld += r.PositionsWithheld
		f.ordersRefused += r.OrdersRefused
		f.spansRefused += r.SpansRefused
		if r.Seed == absentSpanSeed {
			f.forcedSpansRefused += r.SpansRefused
		}
		for verb, words := range r.CopiesRead {
			if f.copiesRead[verb] == nil {
				f.copiesRead[verb] = map[string]int{}
			}
			for w, n := range words {
				f.copiesRead[verb][w] += n
			}
		}
		f.fragmentsDriven += r.FragmentsDriven
		f.truncationsDriven += r.TruncationsDriven
		f.quotesRefused += r.QuotesRefused
		f.bridgesRefused += r.BridgesRefused
		f.bridgesNamed += r.BridgesNamed
		f.reopenedReadied += r.ReopenedReadied
		f.unsoundReadied += r.UnsoundReadied
		if r.Seed == unsoundComputationSeed {
			f.forcedUnsoundReadied += r.UnsoundReadied
		}
		f.unsoundPlans += r.UnsoundPlans
		f.unsoundAnswers += r.UnsoundAnswers
		f.unsoundRepeats += r.UnsoundRepeats
		f.unsoundOverOpen += r.UnsoundOverOpen
		f.reopenedPlans += r.ReopenedPlans
		for k, v := range r.ComputationReopenedReadied {
			f.computationReopenedReadied[k] += v
		}
		f.owedKept += r.OwedKept
		f.computationOwedKept += r.ComputationOwedKept
		if r.Seed == owedComputationSeed {
			f.forcedComputationOwedKept += r.ComputationOwedKept
			f.forcedOwnerWhys = r.OwnerWhys
		}
		for k, v := range r.OwedKeptBy {
			f.owedKeptBy[k] += v
		}
		if r.CitationsOffered > 0 {
			f.citationOfferedRuns++
		}
		if r.OwnerOffered > 0 {
			f.ownerOfferedRuns++
		}
		f.unreadAnswers += r.UnreadAnswers
		f.unreadRepeats += r.UnreadRepeats
		f.unreadRetires += r.UnreadRetires
		f.unreadRetiresJoined += r.UnreadRetiresJoined
		f.unreadClosesRefused += r.UnreadClosesRefused
		f.unsoundClosesRefused += r.UnsoundClosesRefused
		f.repeatsRefused += r.RepeatsRefused
		f.budgetRefused += r.BudgetRefused
		f.oneGapRefused += r.OneGapRefused
		f.oneGapFindings += r.OneGapFindings
		f.ownAnswers += r.OwnAnswers
		f.unclassifiedMinted += r.UnclassifiedMinted
		f.unclassifiedRefused += r.UnclassifiedRefused
		f.rulerFilingsRefused += r.RulerFilingsRefused
		f.closedRemandsRefused += r.ClosedRemandsRefused
		f.trifleRemandsRouted += r.TrifleRemandsRouted
		f.blueRemandActs += r.BlueRemandActs
		f.lensRemandActs += r.LensRemandActs
		f.pastLimitRemandsRefused += r.PastLimitRemandsRefused
		f.lensFirstHolds += r.LensFirstHolds
		f.benchRegrades += r.BenchRegrades
		f.lockRefused += r.LockRefused
		f.lockLifted += r.LockLifted
		f.groundsRefused += r.GroundsRefused
		f.groundsAdmitted += r.GroundsAdmitted
		for kind, n := range r.PartRefused {
			f.partRefused[kind] += n
		}
		f.partsRecorded += r.PartsRecorded
		if r.PartsRecorded >= 2 {
			f.multiPartRuns++
		}
		for why, n := range r.EstoppelMisses {
			f.estoppelMisses[why] += n
		}
		for why, n := range r.ApplyMisses {
			f.applyMisses[why] += n
		}
	}
	return f
}

// judgeSweep holds the sweep to its gates over the folded runs and the process state.
func judgeSweep(t *testing.T, f sweepFold, p processTally) {
	t.Helper()
	// AND WHY IT DID NOT, WHENEVER IT DID NOT. `%d edits applied a proposal verbatim` read 0
	// across 60 runs and named no cause: the branch could have gone untaken, found no proposal
	// on the gap, found the span already edited away, or run and been refused. Four causes, one
	// zero, in the instrument whose job is to detect exactly that shape.
	misses := ""
	if len(f.applyMisses) > 0 {
		causes := make([]string, 0, len(f.applyMisses))
		for why, n := range f.applyMisses {
			causes = append(causes, fmt.Sprintf("%d× %s", n, why))
		}
		sort.Strings(causes)
		misses = "\n  verbatim-apply declined: " + strings.Join(causes, "\n                          ")
	}
	t.Logf("BLUESTAT blue-respond sittings=%d (on an engaged successor %d) over %d runs", f.blueSits, f.blueAgains, f.completed)
	t.Logf("fuzzed %d debate runs · %d failed · verdicts=%v · epochs=%v · exits=%v\n  dialectic events emitted: %v\n  citation axis: %d anchors spliced · %d sources cached\n  provenance: %d of %d blue_edit ops carried --answers · %d of %d gaps earned fix_basis=verified · %d edits applied a proposal verbatim (%d typed) · %d estoppel refusals%s",
		f.completed, f.failed, f.verdicts, f.epochHist, f.whyHist, f.dcov, f.citeAnchors, f.cacheFiles, f.editAnswers, f.dcov["blue_edit"], f.verifiedBasis, f.dcov["mint"], f.verbatimApplied, f.typedVerbatim, f.estoppels, misses)
	t.Logf("epoch limit: %d of %d runs ended at their epoch limit (%d; seed %d is forced to it at %d)", f.limitExits, f.completed, record.DefaultParams.MaxEpochs, limitSeed, forcedEpochLimit)
	t.Logf("inquiry lifecycle: %d fates recorded on an id read from `inquiry propose` (of %d inquiry events)", f.inquiryRecords, f.dcov["inquiry"])
	t.Logf("computation owner: %d run(s) offered an unjudged computation to its owner, at a sitting or the run's end", f.ownerOfferedRuns)
	t.Logf("citation owner: %d run(s) offered an unverified citation to the evidence lens, at a sitting or the run's end", f.citationOfferedRuns)
	t.Logf("unsound computations: %d plan(s) readied the judge or blue for every unsound computation they held (of %d holding one), %d of them seed %d's forced run; %d gap(s) minted carrying an unsound judgement", f.unsoundReadied, f.unsoundPlans, f.forcedUnsoundReadied, unsoundComputationSeed, f.unsoundAnswers)
	t.Logf("source spans: %d cite(s) refused for a span the source does not hold, %d of them seed %d's forced run", f.spansRefused, f.forcedSpansRefused, absentSpanSeed)
	if f.completed > 0 && (f.unsoundReadied == 0 || f.unsoundAnswers == 0) {
		t.Errorf("fuzz readied for an unsound computation in %d plan(s) (of %d holding one) and minted %d gap(s) carrying one, across %d runs — the unsound-computation blocker's readiness or its answering gap is unexercised (false green)", f.unsoundReadied, f.unsoundPlans, f.unsoundAnswers, f.completed)
	}
	// A REPEAT UNSOUND OVER THE JUDGE'S OPEN GAP IS REFUSED, and a gap minted over an open one of the
	// judge's own about the same computation is what the refusal removes: a fresh judgement after blue's
	// edit, or one after the judge's own sound, is the only way to it.
	t.Logf("unsound repeats: %d repeat unsound judgement(s) refused over the judge's open gap; %d gap(s) minted about a computation while the judge's own gap about the sentence as it stands was open", f.unsoundRepeats, f.unsoundOverOpen)
	if f.completed > 0 && f.unsoundRepeats == 0 {
		t.Errorf("fuzz drew no repeat unsound judgement over an open gap across %d runs — the refusal is unexercised (false green)", f.completed)
	}
	t.Logf("reopened readied: %d plan(s) readied the owner of every reopened blocker (of %d holding one); reopened computations whose owner a plan readied: %v", f.reopenedReadied, f.reopenedPlans, f.computationReopenedReadied)
	t.Logf("owed kept: %d lens fold(s) kept from retiring by a debt (three or more barren sittings), %d of them by an unjudged computation (%d in seed %d's forced run); by kind %v", f.owedKept, f.computationOwedKept, f.forcedComputationOwedKept, owedComputationSeed, f.owedKeptBy)
	for i, w := range f.forcedOwnerWhys {
		t.Logf("  seed %d's plan %d: %s", owedComputationSeed, i+1, w)
	}
	t.Logf("unread citations: %d gap(s) minted about a citation the lens's reading could not read, %d repeat unread reading(s) refused over that gap, %d citation(s) blue retired off a sentence for it (%d of them answering that gap and listed in its change log), %d close(s) of that gap refused while the citation stood", f.unreadAnswers, f.unreadRepeats, f.unreadRetires, f.unreadRetiresJoined, f.unreadClosesRefused)
	t.Logf("unsound closes: %d close(s) of a judge's gap refused while the unsound judgement it carries stood", f.unsoundClosesRefused)
	t.Logf("catch-all: %d unclassified mint(s) landed, %d refused for want of a reason", f.unclassifiedMinted, f.unclassifiedRefused)
	t.Logf("classes: %d by_grade in the mint pool; %d always/never minted once each, %d landed", len(mintPoolClasses), len(mintDrivenClasses), len(mintDrivenClasses)-len(unlandedClassDrive(p.ClassLanded)))
	// FULL-SURFACE COVERAGE GATE. A green fuzz that never drove a verb is a false green (the lens
	// stub emitted neither cite nor finding for the whole life of PR-1, unexercised end to end).
	// Assert EVERY event-emitting seat verb fired at least once across the run set — so a
	// regression that silently stops emitting one (a dropped envelope branch, a broken dispatch)
	// fails here, not silently. `halt` is exempt (terminal; covered by TestFuzzHaltPath). Each
	// gated verb fires with per-run probability well above ~20%, so P(missed across the default 60
	// runs) is negligible; if a verb ever flakes here it is a real coverage regression, not noise.
	// The full surface needs enough runs for every verb (incl. the ~10%/run ones like regrade) to
	// fire; below the quorum a low-frequency verb could flake to zero, so the -short smoke keeps
	// only the cite/finding floor (the original false-green guard) and the default+ size asserts
	// all.
	//
	// A SWEEP UNDER QUORUM NOW SAYS SO. These gates used to fall silent below 40 and print the
	// same green as a sweep that ran them — so `-short` (N=15) has been running this package with
	// its three coverage gates inactive and nothing on the record saying which. That is #428's
	// class (a gate that did not fire rendered as a gate that held) inside the sweep built to
	// detect exactly it.
	measured := f.completed >= surfaceQuorum
	if !measured {
		t.Logf("NOT MEASURED: %d runs is under the surface quorum of %d, so these gates did NOT run: "+
			"the per-verb event gate, the citation/provenance floors, the inquiry-record floor, the full-surface command gate, "+
			"and the flag/enum coverage sweeps. This is not a pass over them — only the cite/finding "+
			"floor below was checked. Run the default sweep to assert the surface.", f.completed, surfaceQuorum)
	}
	// EVERY VIEW REACHED SOMEBODY. The per-role sweep skips a view whose group a role does not
	// carry, and this is the witness for that skip: a view no role drove has been swept by nobody
	// while the sweep reported success, which is the exact shape this oracle's own comment warns
	// about. Under quorum it is not asserted, for the same reason the verb gates are not.
	if measured {
		for _, v := range viewNamesForFuzz {
			if !p.DrivenView[v] {
				t.Errorf("no role drove the %q projection across %d runs — it lives under group %q, "+
					"which no role carries, so the sweep covered it with nobody", v, f.completed, seat.GroupOf(v))
			}
		}
	}
	if measured {
		for _, k := range verbsWithEvents {
			if coverExempt[k] {
				continue
			}
			if f.dcov[k] == 0 {
				t.Errorf("fuzz drove ZERO %s events across %d runs — that seat verb is unexercised (false green); a generator branch dropped it", k, f.completed)
			}
		}
	}
	// #256 CITATION-CHAIN GATE. `cite` in the verb gate above is satisfied by red's `lens cite`,
	// which neither fetches nor anchors — so a blue-cite regression would sail through it. These
	// two assert the chain that actually matters ran end to end through the real binary: a source
	// was FETCHED into <run>/cache (loopback server, see sourceURL) and an INVISIBLE anchor was
	// spliced into blue/report.md. Zero of either across the whole sweep is a false green.
	if measured {
		if f.citeAnchors == 0 {
			t.Errorf("fuzz spliced ZERO <!--cite:--> anchors across %d runs — `blue cite` never anchored (false green); the citation axis is unexercised", f.completed)
		}
		if f.cacheFiles == 0 {
			t.Errorf("fuzz cached ZERO sources across %d runs — `fetch`/`blue cite` never populated <run>/cache (false green); the fetch path is unexercised", f.completed)
		}
		// THE EPOCH-LIMIT GATE. Every run setup creates carries an epoch limit, and the exit it
		// makes — nobody dispatched at the last sitting, the terminal bench ruling what stands,
		// CEILING derived from the term — is one no other gate here notices going undriven: the
		// CEILING word is written by the carried ceiling too.
		if !f.limitForced {
			t.Errorf("the forced epoch-limit run on seed %d did not end at its limit of %d (%d run(s) of %d ended at a limit) — the exit a real run takes when its term runs out is unexercised (false green); epochs=%v",
				limitSeed, forcedEpochLimit, f.limitExits, f.completed, f.epochHist)
		}
		// THE INQUIRY-RECORD GATE. A proposed inquiry's fate is recorded by the id its propose result
		// printed, and the reader of that result is a pattern: one that matches nothing leaves every
		// proposal with no fate recorded and every inquiry `--about` reference unpassed, with the `inquiry` verb gate
		// above still green on the proposals themselves.
		if f.inquiryRecords == 0 {
			t.Errorf("fuzz recorded ZERO fates by an id read from `inquiry propose` across %d runs (with %d inquiry events) — "+
				"the propose result named no id the harness reads, or the tool refused every record; the inquiry lifecycle past `proposed` is unexercised by the proposing seat (false green)", f.completed, f.dcov["inquiry"])
		}
		// THE REOPENED GATE. A citation or computation whose sentence changed after its check is owed by
		// its owner, which is then not retired, and the edit moved the head: the plan readies it. A
		// sweep that readied the computation lens for a reopened computation zero times leaves the computation
		// arm of anchorOwnerOf unexercised.
		if f.reopenedReadied == 0 || f.computationReopenedReadied[computationOwner] == 0 {
			t.Errorf("fuzz readied the owner of every reopened blocker in %d plan(s) (of %d holding one), and the owner of a reopened computation %v times by owner, across %d runs — a reopened computation's routing to the computation lens is unexercised (false green)", f.reopenedReadied, f.reopenedPlans, f.computationReopenedReadied, f.completed)
		}
		// THE DEBT GATE. A lens that owes is not retired; a sweep where no fold kept a lens with three
		// barren sittings from retiring — or none for an unjudged computation — leaves the rule unexercised.
		//
		// THE UNJUDGED-COMPUTATION DEBT IS FED BY CONSTRUCTION, not left to the draw (gblock, 2026-10-25): the
		// draw rarely holds it, since the owner checks every computation it owes in the sitting it is offered
		// one, and a sweep of 40 drew it zero times (#1604). owedComputationSeed's run proves a fresh sentence
		// in every blue sitting while the owner, which mints nothing in that run (mintsFresh), is barren, so a fold holds the owner for a computation it has
		// not yet sat to judge. TestALensThatOwesIsNotRetired (internal/record) proves the rule itself.
		if f.owedKept == 0 || f.computationOwedKept == 0 {
			t.Errorf("fuzz kept %d lens fold(s) from retiring by a debt, %d of them by an unjudged computation, across %d runs — the rule that a lens that owes does not retire is unexercised (false green)", f.owedKept, f.computationOwedKept, f.completed)
		}
		// THE UNREAD-CITATION GATE. A reading that reads nothing leaves the citation on a failed check:
		// the lens owes the gap about it, and blue's remedy is to retire the citation. A sweep where no
		// lens minted that gap, or blue never retired such a citation, leaves the obligation and the
		// citation-only retire unexercised.
		if f.unreadRetiresJoined == 0 {
			t.Errorf("no citation retire answering the gap about an unread citation was listed in that gap's change log, across %d runs — the retire's join to the gap it answers is unexercised (false green)", f.completed)
		}
		if f.unreadAnswers == 0 || f.unreadRetires == 0 {
			t.Errorf("fuzz minted %d gap(s) about an unread citation and retired %d such citation(s) off its sentence, across %d runs — the failed-check path for a citation is unexercised (false green)", f.unreadAnswers, f.unreadRetires, f.completed)
		}
		// THE STANDING-FAILED-CHECK CLOSE GATE. A gap that carries a failed check does not close while
		// the check stands (record.requireFailedCheckAnswered); answerFailed drives the close of the gap
		// it has just minted, once a run for each kind, so a sweep that refused none of either kind has
		// lost the drive or the refusal.
		if f.unreadClosesRefused == 0 || f.unsoundClosesRefused == 0 {
			t.Errorf("fuzz had %d close(s) refused over an unread citation and %d over an unsound computation, across %d runs — the refusal of a close over a failed check that still stands is unexercised (false green)", f.unreadClosesRefused, f.unsoundClosesRefused, f.completed)
		}
		// THE ANCHOR-OWNER GATE. checkAnchorOffers reads each lens's list at each of its sittings, and
		// anchorOffersAgree every lens's at each run's end; a sweep in which no owner was offered an
		// anchor of a kind asks neither anything of that kind, and leaves the routing unexercised.
		if f.citationOfferedRuns == 0 {
			t.Errorf("fuzz offered no unverified citation to the evidence lens in any of %d runs — the citation routing is unexercised (false green)", f.completed)
		}
		if f.ownerOfferedRuns == 0 {
			t.Errorf("fuzz offered no unjudged computation to its owner in any of %d runs — the computation routing is unexercised (false green)", f.completed)
		}
		// THE CLOSING GATE. The `closing` verb gate above is satisfied by any one act, so a plan that
		// stopped holding the bench for an awaited closing, or a drive that stopped leaving one unfiled,
		// would leave the duty's other half unreached behind a green sweep.
		if f.closingsFiled == 0 || f.closingsRefused == 0 || f.closingWaits == 0 {
			t.Errorf("fuzz filed %d closing(s), was refused %d envelope(s) while one stood and saw %d plan(s) keep the bench waiting on one, across %d runs — each must be nonzero, or the closing duty is unexercised (false green)", f.closingsFiled, f.closingsRefused, f.closingWaits, f.completed)
		}
		// NOT A GATE: a plan permitting a PASS is reached by the forced-verified run, which must record
		// it, and rarely by any other, so the withheld PASS is a draw the sweep may not make. The duty is
		// held by record's TestTheChairOwesAPositionWhereItWithheldAPermittedPass and capture's
		// record-parity tests; the count is reported so a sweep that does draw it says so.
		t.Logf("chair sittings that withheld a permitted PASS and argued why: %d across %d runs", f.positionsWithheld, f.completed)
		// THE SECTION-ORDER GATE. The generators here write the report's text in order, so the
		// refusal that holds it is reached only by the drive that writes a section out of place.
		if f.ordersRefused == 0 {
			t.Errorf("fuzz drove ZERO edits refused for a section out of order across %d runs — no drive wrote one, so the refusal that holds the order the report ships in is unexercised (false green)", f.completed)
		}
		// THE SOURCE-SPAN GATE. Every other cite quotes words its source's text holds, so the
		// refusal of a span the source does not hold is reached only by the drive that types one.
		if f.spansRefused == 0 {
			t.Errorf("fuzz drove ZERO cites refused for a source span the source does not hold across %d runs — the refusal that keeps blue's own wording off the record as the source's is unexercised (false green)", f.completed)
		}
		// THE COPY GATE. A reading records the copy the cache held for its source; corroborate is
		// driven over every seeded copy, and verify reads the copies those corroborations placed.
		for _, s := range seededCopies {
			if f.copiesRead["corroborate"][s.word] == 0 {
				t.Errorf("fuzz drove ZERO corroborations of a %s copy across %d runs — the copy a reading records is unexercised for it (false green): %v", s.word, f.completed, f.copiesRead)
			}
		}
		substituted := 0
		for w, n := range f.copiesRead["verify"] {
			if w != "none" {
				substituted += n
			}
		}
		if substituted == 0 {
			t.Errorf("fuzz drove ZERO verifications of a copy other than the url's own across %d runs — the copy a verification records is unexercised (false green): %v", f.completed, f.copiesRead)
		}
		// THE SPAN-BOUNDARY GATE. Every other held span is a whole sentence, so the refusal of one
		// cut from inside a clause is reached only by the drives that type sourceFragment and
		// sourceTruncation.
		if f.fragmentsDriven == 0 {
			t.Errorf("fuzz drove ZERO cites of a span starting mid-clause across %d runs — the refusal that keeps a fragment reading the other way off the record is unexercised (false green)", f.completed)
		}
		if f.truncationsDriven == 0 {
			t.Errorf("fuzz drove ZERO cites of a span ending mid-clause across %d runs — the refusal that keeps a span stopped before its clause ends off the record is unexercised (false green)", f.completed)
		}
		// THE REPORT'S-SENTENCE GATE. The lens drive quotes a real span of the report under every
		// outcome, so the refusal of a quote the report does not hold is reached only by the drive
		// that quotes the source's sentence.
		if f.quotesRefused == 0 {
			t.Errorf("fuzz drove ZERO corroborations refused for a quote the report does not hold across %d runs — no lens quoted its source's sentence, so the refusal that keeps a contradiction's claim the report's own sentence is unexercised (false green)", f.completed)
		}
		// THE BRIDGE GATE. supports_with_bridge names what it bridged; a reading naming none is
		// refused. Both halves must run, or the required value is unexercised behind a green sweep.
		if f.bridgesRefused == 0 || f.bridgesNamed == 0 {
			t.Errorf("fuzz drove %d supports_with_bridge readings refused for no --bridge and %d recorded with one across %d runs — both must be non-zero, or the required bridge kind is unexercised (false green)", f.bridgesRefused, f.bridgesNamed, f.completed)
		}
		// THE REPEATED-ACT GATE. A seat that got an act wrong meets the
		// refusal of its second, and that refusal names the act that answers the first. A drive
		// that stopped repeating an act would leave those refusals unreached behind a green sweep.
		if f.repeatsRefused == 0 {
			t.Errorf("fuzz drove ZERO refused repeats across %d runs — no seat filed a second position or closing in a sitting, so the refusal that names the answering act is unexercised (false green)", f.completed)
		}
		// THE MINT-BUDGET GATE. The budget caps the gaps a lens holds open, and the sweep's floor of
		// one keeps most lenses at it: a mint from a lens already holding its budget is refused. The
		// `mint` verb gate is satisfied by the mints that land, so a budget that stopped refusing — or
		// terms too generous for any lens to reach it — would leave the refusal unreached behind a
		// green sweep.
		if f.budgetRefused == 0 {
			t.Errorf("fuzz drove ZERO mints refused for budget across %d runs — no lens held its whole mint budget open and minted again, so the cap on open gaps is unexercised (false green)", f.completed)
		}
		// THE ONE-GAP GATE. A mint of a class on a sentence or inquiry where an open gap of that class
		// stands is refused, naming it, and the lens files its point as a finding about that gap. The
		// `mint` verb gate is satisfied by the mints that land, so a rule that stopped refusing — or a
		// drive that stopped reaching it — would hide behind a green sweep.
		t.Logf("one gap per class and subject: %d mint(s) refused, %d finding(s) filed about the gap that holds the subject", f.oneGapRefused, f.oneGapFindings)
		if f.oneGapRefused == 0 || f.oneGapFindings == 0 {
			t.Errorf("fuzz drove %d mints refused over an open gap of their class and subject and %d findings about that gap across %d runs — both must be non-zero, or the one-gap rule is unexercised (false green)", f.oneGapRefused, f.oneGapFindings, f.completed)
		}
		// THE ANSWERING-MINT GATE. A gap minted from a lens's own raising finding takes no mint, and
		// a run fails on one the budget refused (runner.answerBudgetErr); that oracle is silent where
		// no such mint was driven, so the sweep fails unless some landed.
		t.Logf("answering mints: %d landed outside the minting lens's budget, %d mint(s) refused for budget", f.ownAnswers, f.budgetRefused)
		if f.ownAnswers == 0 {
			t.Errorf("fuzz drove ZERO mints answering the minting lens's own raise across %d runs — the gap outside the budget is unexercised (false green)", f.completed)
		}
		// THE CATCH-ALL GATE. A gap no class fits is minted `unclassified` with a reason, and the
		// record refuses one without. The `mint` verb gate is satisfied by every other class, so a
		// catch-all that stopped landing — or a reason check that stopped refusing — would hide
		// behind a green sweep.
		if f.unclassifiedMinted == 0 {
			t.Errorf("fuzz minted ZERO gaps under unclassified across %d runs — the catch-all is unexercised (false green)", f.completed)
		}
		if f.unclassifiedRefused == 0 {
			t.Errorf("fuzz drove ZERO unclassified mints refused for want of a reason across %d runs — the record's reason check is unexercised (false green)", f.completed)
		}
		// THE RULER-FILES-NOTHING GATE. The chair's grade motion and the bench's docket motion are
		// refused at the surface; a drive that stopped reaching for them would leave the refusal
		// unreached behind a green sweep.
		if f.rulerFilingsRefused == 0 {
			t.Errorf("fuzz drove ZERO refused ruler filings across %d runs — neither the chair filed a grade motion nor the bench a docket motion, so the refusal of a filing by its subject's ruler is unexercised (false green)", f.completed)
		}
		// THE REMAND GATES (gblock, 2026-10-24). A remand of a closed gap is refused naming the close,
		// and a remand of an open gap that is not material is readied like any remand; the run oracles
		// are silent where neither was driven.
		t.Logf("remands: %d of a closed gap refused naming the close, %d of a gap not material readied for their exchange", f.closedRemandsRefused, f.trifleRemandsRouted)
		if f.closedRemandsRefused == 0 {
			t.Errorf("fuzz drove ZERO remands of a closed gap across %d runs — the refusal that names the close is unexercised (false green)", f.completed)
		}
		if f.trifleRemandsRouted == 0 {
			t.Errorf("fuzz readied ZERO remands of a gap that is not material across %d runs — the route by which the bench's remand outranks materiality is unexercised (false green)", f.completed)
		}
		// THE REMAND'S ITEM BLOCKS, AND ONE REMAND AT IMPASSE IS THE LIMIT (gblock, 2026-10-25): each seat a
		// remand readies is held until it acts on the gap, and a second remand at impasse is refused.
		t.Logf("remands: %d act(s) by blue and %d by the minting lens on a remanded gap its item held them to; %d remand(s) past the limit refused", f.blueRemandActs, f.lensRemandActs, f.pastLimitRemandsRefused)
		if f.blueRemandActs == 0 || f.lensRemandActs == 0 {
			t.Errorf("fuzz drove %d blue and %d lens act(s) on a remanded gap whose item blocked the seat, across %d runs — the remand's blocking item is unexercised for a seat it holds (false green)", f.blueRemandActs, f.lensRemandActs, f.completed)
		}
		if f.pastLimitRemandsRefused == 0 {
			t.Errorf("fuzz drove ZERO remands past the limit across %d runs — the refusal that names the dispositions that close the gap is unexercised (false green)", f.completed)
		}
		// THE LENS SITS FIRST AND BEGINS NOTHING (ruled by gblock 2026-10-27): a remanded gap the
		// minting lens acted on before blue still holds blue, read before blue registers.
		t.Logf("remands: %d item(s) held blue before its register after the minting lens acted first", f.lensFirstHolds)
		if f.lensFirstHolds == 0 {
			t.Errorf("fuzz drove ZERO remands in which the minting lens acted before blue and blue's item was read before its register, across %d runs — lens-first ordering under a remand is unexercised (false green)", f.completed)
		}
		// THE BENCH'S ORDERS ON THE RECORD (gblock, 2026-10-30): the bench sets a grade while it hears
		// the gap, the minting lens may not move it, and a ruled gap comes back on new grounds only.
		t.Logf("bench orders: %d bench regrade(s), %d lens move(s) of a locked axis refused, %d of a released axis admitted, %d reconsideration(s) refused for wanting new grounds, %d admitted on them", f.benchRegrades, f.lockRefused, f.lockLifted, f.groundsRefused, f.groundsAdmitted)
		if f.benchRegrades == 0 {
			t.Errorf("fuzz drove ZERO bench regrades across %d runs — the bench setting the grade its ruling rules on is unexercised (false green)", f.completed)
		}
		if f.lockRefused == 0 {
			t.Errorf("fuzz drove ZERO moves of a bench-set axis refused across %d runs — the lock on the bench's grade is unexercised (false green)", f.completed)
		}
		if f.lockLifted == 0 {
			t.Errorf("fuzz drove ZERO lens moves of an axis the bench released across %d runs — the bench's second move returning the axis to the lens is unexercised (false green)", f.completed)
		}
		if f.groundsRefused == 0 || f.groundsAdmitted == 0 {
			t.Errorf("fuzz drove %d reconsiderations refused for wanting new grounds and %d admitted on them across %d runs — both must be non-zero, or the new-grounds check is unexercised (false green)", f.groundsRefused, f.groundsAdmitted, f.completed)
		}
		// THE REPORT IS THE LANES' PARTS (parts_test.go): every lane records one, the report joins two or
		// more of them, and each refusal of a part is driven, partsSeed's run driving every one.
		t.Logf("parts: %d recorded across %d runs, %d run(s) joining two or more; refused %v", f.partsRecorded, f.completed, f.multiPartRuns, f.partRefused)
		if f.partsRecorded == 0 || f.multiPartRuns == 0 {
			t.Errorf("fuzz recorded %d part(s) and joined two or more in %d run(s) across %d runs — both must be non-zero, or the lanes' parts are unexercised (false green)", f.partsRecorded, f.multiPartRuns, f.completed)
		}
		for _, kind := range partRefusals {
			if f.partRefused[kind] == 0 {
				t.Errorf("fuzz drove ZERO part refusals for %s across %d runs — the refusal is unexercised (false green)", kind, f.completed)
			}
		}
		// #267 PROVENANCE GATE. The verb gate above accepts any blue_edit, so an edit drive
		// that stopped sending --answers would satisfy it while the join key every #267
		// measurement reads went silently missing.
		if f.editAnswers == 0 {
			t.Errorf("fuzz recorded ZERO blue_edit events carrying --answers across %d runs — the provenance key is unexercised (false green)", f.completed)
		}
		if f.verifiedBasis == 0 {
			t.Errorf("fuzz minted ZERO gaps with fix_basis=verified across %d runs — no concrete proposal ever validated, so the whole stage-3 path is unexercised (false green)", f.completed)
		}
		if f.foundByMints == 0 {
			t.Errorf("fuzz minted ZERO gaps naming a finding in found_by across %d runs — no accepted mint carried --found-by, so the id a found_by names a finding by is unexercised (false green)", f.completed)
		}
		if f.verbatimApplied == 0 {
			t.Errorf("fuzz recorded ZERO verbatim applications across %d runs — nothing ever estopped red, so the stage-4 guard is unexercised (false green)", f.completed)
		}
		if f.typedVerbatim == 0 {
			t.Errorf("fuzz recorded ZERO typed verbatim applications across %d runs (%d by --accept) — blue typing the board's pair never matched the pair --accept sends, so the typed arm is unexercised (false green)", f.completed, f.verbatimApplied)
		}
		// AND THE GUARD ITSELF FIRED, which the line above only establishes the PRECONDITION for.
		// Verbatim applications were counted for the whole life of this gate while the estoppel
		// path was never once reached: red never minted against the text blue had applied, so the
		// refusal that writes `log --type estoppel` had no way to happen and the enum value was
		// reported as undriven — for however long, on a gate that only runs at a tag.
		if f.estoppels == 0 {
			// THE ZERO NAMES ITS REASON. `verbatimApplied` proves the precondition and this proves
			// the refusal; with nothing between them a red gate could not say whether the guard
			// had gone or the drive had simply declined, and those want opposite fixes.
			why := "the drive never declined either — it was not reached"
			if len(f.estoppelMisses) > 0 {
				causes := make([]string, 0, len(f.estoppelMisses))
				for c, n := range f.estoppelMisses {
					causes = append(causes, fmt.Sprintf("%d× %s", n, c))
				}
				sort.Strings(causes)
				why = "the drive declined: " + strings.Join(causes, "; ")
			}
			t.Errorf("fuzz recorded ZERO tool estoppel refusals across %d runs (with %d verbatim applications) — "+
				"red never minted against text it had prescribed and blue had applied, so lens/mint.go's estoppel guard "+
				"is unexercised and `log --type estoppel` is unreachable (false green).\n          %s", f.completed, f.verbatimApplied, why)
		}
	}
	if !measured {
		for _, k := range []string{"cite", "finding"} {
			if f.dcov[k] == 0 {
				t.Errorf("fuzz drove ZERO %s events across %d runs — the lens record channel is unexercised (false green)", k, f.completed)
			}
		}
	}
	// FULL-SURFACE GATE, DERIVED FROM THE COMMAND TREE.
	//
	// The harness tallies EVERY invocation it makes (noteExec), so "what did the fuzz drive"
	// is observed rather than declared. cli.CommandPaths() walks the real cobra tree for
	// "what exists". Nothing hand-written states the surface, which is the point: the gap
	// this closes was created by a hand list going stale — 18 of 44 seat verbs and 7 of 9
	// root commands undriven, every undriven seat verb one that records nothing and so was
	// invisible to the event gate above.
	//
	// The only hand-written list left is exemptSurfaces, and each entry states its reason.
	if measured {
		if un := unreachedSurfaces(p); len(un) > 0 {
			t.Errorf("%d command path(s) exist in the tool and were NEVER invoked across %d runs (false green — add a drive, or an exemption with its reason):\n  %s",
				len(un), f.completed, strings.Join(un, "\n  "))
		}
		// DRIVING A VERB IS NOT DRIVING ITS SURFACE. The gate above asks whether every verb
		// ran; these ask whether the flags those verbs take were ever passed, and whether
		// every value of every closed set was ever recorded. An unpassed flag is code no run
		// has executed, and an unreached enum member is a word nothing has written — both
		// sitting behind a sweep that reports full coverage.
		if un := unreachedFlags(p); len(un) > 0 {
			t.Errorf("%d flag(s) on verbs the sweep DOES drive were never passed:\n  %s", len(un), strings.Join(un, "\n  "))
		}
		if un := unreachedEnumValues(p); len(un) > 0 {
			t.Errorf("%d enum value(s) were never driven:\n  %s", len(un), strings.Join(un, "\n  "))
		}
		// AND THE DRIVE, NAMED, not only the word it was for. `close --as moot` is reached by one
		// deliberate drive per sweep rather than by the pool (see closeGap); when it declines, the
		// census above reports a missing enum value and says nothing about why, which reads as the
		// vocabulary having outgrown the fuzz rather than as a driver that did not fire.
		// AND THE FORCED-UNVERIFIED DRIVE, for the same reason. `outcome --as UNVERIFIED` is
		// reachable about once in sixty by the draw, so the sweep pins it to unverifiedSeed and
		// the census above is the only thing that notices when that run does not deliver — as a
		// line naming the WORD, which reads as the vocabulary having outgrown the fuzz and sends
		// the next reader to the schema. It is a DRIVER that declined.
		//
		// MEASURED (#870, #847): with `moot` in closeGap's random pool the forced run ended
		// CEILING or VERIFIED on 2 of 2 sweeps; with the pool restored it ended UNVERIFIED on 5
		// of 5, always exactly once, which is the forced seed and not a draw. The drive is not
		// chronically flaky — it breaks when something else perturbs the run, and every enum
		// value added to a pool is such a perturbation. This says so in one line when it happens.
		//
		// It asks the VERDICT rather than the census because the two answer different questions:
		// the census says the word was never written by anything, this says the run that exists
		// to write it went somewhere else, and names where.
		if f.forcedVerdict != "UNVERIFIED" {
			t.Errorf("the forced-UNVERIFIED run on seed %d ended %q, not UNVERIFIED — the drive that makes "+
				"`outcome --as UNVERIFIED` a fact rather than a 1-in-60 draw did not deliver, and the enum census "+
				"reports that as a missing WORD rather than as this declining DRIVER.%s",
				unverifiedSeed, f.forcedVerdict, forcedWhyClause(f.forcedWhy))
		}
		// AND THE FORCED-VERIFIED DRIVE, for the same reason: the run that exists to write the word
		// went somewhere else, and this names where.
		if f.verifiedVerdict != "VERIFIED" {
			t.Errorf("the forced-VERIFIED run on seed %d ended %q, not VERIFIED — the drive that makes "+
				"`outcome --as VERIFIED` a fact rather than a draw did not deliver.%s",
				verifiedSeed, f.verifiedVerdict, forcedWhyClause(f.verifiedWhy))
		}
		if !p.MootSpent {
			t.Error("the sweep's single `close --as moot` drive never fired: no run reached closeGap with it pending, or every attempt was refused. The word's whole coverage is that one drive")
		}
		if p.LookedFirst == 0 {
			t.Error("no lens sitting opened with a look before its register: the drive that holds a board read to registering no one never completed")
		}
		if p.RegisteredTwice == 0 {
			t.Error("no sitting opened by registerTwice: the drive that holds a repeat register under one agent id to a refusal that writes nothing never completed")
		}
		// AND BOTH ENDS OF AN AGENT-BEARING SITTING: closed at its return, as SubagentStop closes it,
		// and left open by the deliberate draw, as a subagent that dies before SubagentStop leaves it.
		if p.SittingsClosed == 0 {
			t.Error("no agent-bearing sitting was closed at its return: every such sitting the sweep opened reads never closed, a state no live run with working hooks produces")
		}
		if p.SittingsLeftOpen == 0 {
			t.Error("no agent-bearing sitting was left open by the never-closed draw: the record's never-closed sitting went unexercised")
		}
		// AND THE TWO ENDS THAT SPLIT ADMISSION FROM CLOSE: a sitting admitted and never closed, as a
		// subagent that dies after its envelope call leaves it, and one closed with no admission,
		// which the record reads as failed and runOne holds the record's failed-sitting count to.
		if p.AdmittedNeverClosed == 0 {
			t.Error("no admitted sitting was left open by the died-after-admission draw: an admission with no close went unexercised")
		}
		if p.FailedSittings == 0 {
			t.Error("no agent-bearing sitting of a seat whose work arrives as items was closed with no admission: runOne's failed-sitting check compared zero with zero on every run")
		}
		// AND THE DELIBERATE CLASS MINTS, for the same reason: each `always`/`never` class is one
		// drive per sweep (r.mint), and the census above counts a refused attempt as driven.
		if un := unlandedClassDrive(p.ClassLanded); len(un) > 0 {
			t.Errorf("the sweep's deliberate class mints landed no gap under %d of %d `always`/`never` class(es): %s — no run minted with them owed, or every attempt was refused. Each word's whole coverage is that one drive",
				len(un), len(mintDrivenClasses), strings.Join(un, ", "))
		}
		// AND THE FLAGS INHERITED ONTO EVERY COMMAND, which this gate could not see at all until
		// PersistentFlagNames existed: CommandFlags reports local flags only, so `--json`,
		// `--run`, `--schema` and `--seat-id` were never in the denominator above.
		if un := unreachedPersistentFlags(p); len(un) > 0 {
			t.Errorf("%d persistent flag(s) — inherited onto every command — were never passed and "+
				"carry no exemption:\n  %s", len(un), strings.Join(un, "\n  "))
		}
	}
	// AT EVERY SIZE, both of these. The tally is the sweep's own trajectory and costs nothing to
	// print; withholding it under quorum hid the evidence for the gate below on exactly the runs
	// most likely to need it. And a path that never succeeded is a coverage hole at ANY N — it
	// needs no quorum, because one refusal per invocation is not a sampling accident.
	t.Logf("agent-bearing sittings: %d closed at their return (%d of them failed: closed with no admission), %d left open by the never-closed draw, %d left open after their admission; %d opened by registerTwice, %d by lookThenRegister",
		p.SittingsClosed, p.FailedSittings, p.SittingsLeftOpen, p.AdmittedNeverClosed, p.RegisteredTwice, p.LookedFirst)
	t.Log(execReport(p))
	t.Log(eitherReport(p))
	t.Log(failReport(p))
	t.Log(envelopeReport(p))
	// THE SAME SWEEP, READ AS A GRAPH (#535 step 1). Emitted here rather than from a test of its
	// own for the reason #635 was filed about: a second test reading this package-level tally
	// would run at whatever point its declaration order put it, and a gate that reads a tally
	// before its writers have run reports a plausible zero. Here the writers are done and the
	// quorum is already known.
	t.Log(graphReport(p))
	t.Log(permittedReport())
	// THE PERMITTED SIDE HAS ITS OWN FLOOR, and it is not the observed side's. The reference edge
	// is derived by type-asserting flag values, so it fails in exactly one direction: a change to
	// the flag types stops the assertion matching, the map comes back EMPTY, and the report says
	// "0 checked flag(s)" — which reads like a surface where nothing is checked rather than a
	// walker that stopped looking. Same shape as #654, one edge over.
	// The walker's own floor and its join with CommandPaths live in internal/cli, beside the
	// other walkers' (TestTheSurfaceWalkersAllSeeTheSameTree) — one home for "is this walker
	// broken", rather than a second copy here that could disagree with it. What belongs HERE is
	// the vocabulary check, because the kinds are what this report renders.
	for path, byFlag := range cli.CommandReferences() {
		for flag, kind := range byFlag {
			if referenceKinds[kind] == "" {
				t.Errorf("%s --%s is checked against entity kind %q, which is not in referenceKinds: add it with what it points at, so a new entity class arrives rather than joining silently", path, flag, kind)
			}
		}
	}
	if measured {
		// A GRAPH WITH NO EDGES RENDERS AS "0 seats · 0 edges" AND READS LIKE A QUIET RUN. If
		// seatOfArgs stops recovering the seat — a flag rename, a harness that stops passing
		// --seat-id — every edge silently vanishes and the report still prints, still green,
		// naming no seat and therefore accusing no verb of being unreachable. The floor is
		// deliberately loose: this asserts the instrument ran, not how much it found.
		// AND THE OTHER SIDE OF THE JOIN, which this gate missed on its first run: it guarded the
		// edges and not the map they are diffed against, so `cli.CommandRecords()` came back
		// EMPTY and the report rendered "0 of 0 recording verbs reached" — a sentence that reads
		// like full coverage and means the opposite. Pinned permanently in internal/cli by
		// TestTheSurfaceWalkersAllSeeTheSameTree; repeated here because this report is the thing
		// a reader believes.
		if len(cli.CommandRecords()) == 0 {
			t.Error("cli.CommandRecords() is EMPTY — the graph's verb->event side measured nothing and rendered it as \"0 of 0 recording verbs reached\" (#654)")
		}
		if seats, edges, _, _ := graphEdges(p); seats < 2 || edges == 0 {
			t.Errorf("the observed graph recorded %d seat(s) and %d edge(s) across %d runs — the sweep drove the surface, so this is the EMITTER failing to attribute it, not a quiet run",
				seats, edges, f.completed)
		}
	}
	walked, wrong := namedOtherThanOnce(p)
	t.Logf("%d refusals walked for how many times each names its command; %d other than once", walked, len(wrong))
	if len(wrong) > 0 {
		t.Errorf("%d refusal(s) the sweep met name their command other than exactly once — the renderer prints the path from the command that resolved, and the diagnosis starts at what is wrong:\n  %s",
			len(wrong), strings.Join(wrong, "\n  "))
	}
	if un := neverSucceeded(p); len(un) > 0 {
		t.Errorf("%d command path(s) were driven and NEVER SUCCEEDED — the surface gate counts invocations, so these read as covered while their success path has never run:\n  %s",
			len(un), strings.Join(un, "\n  "))
	}
}
