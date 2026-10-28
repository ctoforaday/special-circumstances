package fuzz

// Fuzz the ACTUAL debate.js orchestrator against the REAL feov-record binary. goja runs the
// unmodified script (harness faked); each agent() call is a seat that makes coherent, random,
// VALID tool calls into a real run directory and returns a randomised envelope to drive
// debate.js's own branches (verdict, gap counts, epochs, petitions). A failing run is a real
// finding (in debate.js, the tool, or verify), reproducible from its seed.
//
// COVERAGE CONTRACT. envelopeFor drives every eligible seat to exercise its whole verb surface,
// not a happy path: lens (cite/finding/inquiry/log/closing), merge (position/
// mint/close incl. repaired_with_regression/regrade any axis/
// dispute-respond/spot-check/verdict/petition), blue (closing/dispute
// across all four dimensions/manifest-row/inquiry/retire/petition), bench
// (opinion/outcome/certify/assemble/petition-rule). The
// disputes docket is driven through the ENVELOPE (raiseDisputes/answerDisputes) and the
// petition->petition-rule docket through the RECORD and the plan that relays it
// (maybePetition/rulePetitions), so debate.js's routing runs too.
//
// ORACLES per run: (1) verify passes — whatever path the debate took, the record satisfies every
// invariant; (2) the JSON views (findings/log/debate) exit 0 and parse; (2c) the six markdown
// views (ledger/archive/debate/changelog/citation-ledger/inquiries) render in-memory via
// view.Markdown and exit 0; (3) every dialectic prose renders in the report (the A1-A3
// write-here/read-there class); (4) #111 tier — every seat dispatched on the configured tier.
// COVERAGE GATE: across the run set, every event-emitting verb
// in verbsWithEvents must fire at least once (a regression that silently drops one fails loudly).
// `halt` terminates the run, so it is covered by the dedicated TestFuzzHaltPath, not the sweep.
//
// Run: go test ./releasegate/fuzz -run TestFuzzDebate -count=1   (respects -short by shrinking N).
// Confidence sweep: FUZZ_N=1000 go test ./releasegate/fuzz -run TestFuzzDebate -timeout 1200s.
// FUZZ_C overrides concurrency (runs are subprocess-bound, so the default oversubscribes cores).

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/consistency"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
	"maps"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/debatejs"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	reportdoc "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
)

// Capture only seat-id characters — NOT the trailing "." in "SEAT_ID: red-chair." — or the
// malformed id makes every tool call fail silently and the fuzz degrades to trivial PASS runs.
var seatRe = regexp.MustCompile(`SEAT_ID:\s*([A-Za-z0-9-]+)`)

// THE CITATION AXIS NEEDS A SOURCE TO FETCH (#256). `feov-record fetch` performs a REAL http GET —
// that IS the code under test, caps and all — so this harness serves it from 127.0.0.1 rather than
// stubbing it out. Loopback only: the same discipline internal/fetchcache's own tests use, so CI
// touches no external network while the whole fetch → cache → cite → anchor path runs end to end
// through the real binary. One server per test binary; the path varies so distinct URLs exercise
// cache misses as well as hits.
var (
	sourceSrvOnce sync.Once
	sourceSrvURL  string
)

func sourceURL(path string) string {
	sourceSrvOnce.Do(func() {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// ONE REAL SCAN, so the one-page render has a success path to run: a PDF with no text
			// layer, cached with --ocr=false and never read.
			if req.URL.Path == "/scan.pdf" {
				w.Header().Set("Content-Type", "application/pdf")
				_, _ = w.Write(seatprobe.ScannedFixturePDF())
				return
			}
			fmt.Fprintln(w, sourceBody(req.URL.Path))
		}))
		sourceSrvURL = srv.URL // deliberately never closed: it lives for the test binary
	})
	return sourceSrvURL + path
}

// sourceBody is the text the loopback source serves at path, which a cite of it quotes: one
// sentence, so a span of all of it starts and ends where the source starts and ends a sentence.
func sourceBody(path string) string { return "Fuzz source body for " + path + "." }

// sourceFragment is a span sourceBody holds once that starts inside its clause, which the tool
// refuses (fetchcache.SpanBoundary).
func sourceFragment(path string) string { return "source body for " + path }

// sourceTruncation is a span sourceBody holds once that starts its sentence and stops inside its
// clause, which the tool refuses (fetchcache.SpanBoundary).
func sourceTruncation(string) string { return "Fuzz source body" }

// mootPending and mootSpent carry the sweep's single `close --as moot` drive. Package-level and
// atomic because the sweep runs its seeds concurrently and each seed runs its seats concurrently:
// the drive belongs to the SWEEP, so its namespace has to be the sweep's. Reset by TestFuzzDebate
// so -count=2 measures twice rather than reporting the first run's fact for the second.
var (
	mootPending atomic.Bool
	mootSpent   atomic.Bool
)

// mintPoolClasses are the GapClass words a random mint draws from: every value whose
// material_default is `by_grade`, bar the catch-all, which has its own drive (r.mint). Read off the
// enum, so a value added to the closed set joins the pool or the deliberate drive by its facet.
var mintPoolClasses, mintDrivenClasses = gapClassesByMateriality()

// gapClassesByMateriality splits the GapClass vocabulary by its material_default facet: the
// `by_grade` words (the catch-all excepted) and the `always`/`never` words.
func gapClassesByMateriality() (byGrade, fixed []string) {
	vals := recordpb.GapClass(0).Descriptor().Values()
	for i := 0; i < vals.Len(); i++ {
		c := recordpb.GapClass(vals.Get(i).Number())
		if c == recordpb.GapClass_GAP_CLASS_UNSPECIFIED || c == recordpb.GapClass_GAP_CLASS_UNCLASSIFIED {
			continue
		}
		switch recordpb.MaterialDefault(c) {
		case recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE:
			byGrade = append(byGrade, recordpb.Word(c))
		case recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS, recordpb.ClassMaterial_CLASS_MATERIAL_NEVER:
			fixed = append(fixed, recordpb.Word(c))
		}
	}
	return byGrade, fixed
}

// classDrive carries the sweep's deliberate mints: each `always`/`never` class is minted ONCE per
// sweep, as closeGap closes as `moot`. Sweep-wide for the same reason, and reset by TestFuzzDebate.
// pending is what is still owed; landed is what a mint landed under.
var classDrive struct {
	mu      sync.Mutex
	pending []string
	landed  map[string]bool
}

// resetClassDrive arms the sweep's deliberate class mints.
func resetClassDrive() {
	classDrive.mu.Lock()
	defer classDrive.mu.Unlock()
	classDrive.pending = slices.Clone(mintDrivenClasses)
	classDrive.landed = map[string]bool{}
}

// takeClassDrive hands out one owed class, or false when none is owed.
func takeClassDrive() (string, bool) {
	classDrive.mu.Lock()
	defer classDrive.mu.Unlock()
	if len(classDrive.pending) == 0 {
		return "", false
	}
	c := classDrive.pending[0]
	classDrive.pending = classDrive.pending[1:]
	return c, true
}

// settleClassDrive records a deliberate class's mint: landed spends it; a refusal puts it back,
// because the word's whole coverage is this one drive.
func settleClassDrive(c string, landed bool) {
	classDrive.mu.Lock()
	defer classDrive.mu.Unlock()
	if landed {
		classDrive.landed[c] = true
		return
	}
	classDrive.pending = append(classDrive.pending, c)
}

// unlandedClassDrive is the deliberate classes no mint landed under, sorted.
func unlandedClassDrive() []string {
	classDrive.mu.Lock()
	defer classDrive.mu.Unlock()
	var out []string
	for _, c := range mintDrivenClasses {
		if !classDrive.landed[c] {
			out = append(out, c)
		}
	}
	sort.Strings(out)
	return out
}

// lockedRand is the seed's generator, safe for the concurrent seats phase 3 introduces.
//
// A WRAPPER RATHER THAN A MUTEX AT EVERY CALL SITE: there are 23 draws across this file, all
// Intn, and a lock the caller has to remember is a lock somebody eventually does not. The type
// makes the guarantee structural — a draw cannot be taken without it.
//
// It does NOT preserve per-seed determinism, and that is stated rather than hidden: the SEQUENCE
// is deterministic, but which seat takes which draw depends on the interleaving, so a seed no
// longer reproduces its own run. That cost is #630's, argued there; the run directory is kept on
// failure and the record carries the true ordering, which is what a failing seed is read from.
type lockedRand struct {
	mu sync.Mutex
	r  *rand.Rand
}

// registerArgs is a seat's `register` argv. THE BENCH NAMES WHAT ITS SITTING IS FOR and no other
// seat may, so this asks the write path's own question rather than keeping a list of which ids are
// the bench. `docket` is the sitting the chair dispatches, which is what the fuzz drives.
//
// A bench register without one is REFUSED, and these call sites discard the error — so the miss
// arrived many steps later as a TypeError out of the engine, on a run whose bench had simply never
// sat. That discard is its own defect and is filed separately; this is the fix for the argv.
func registerArgs(seatID, occasion string) []string {
	args := []string{"register", "--seat-id", seatID}
	if record.SeatOwesOccasion(seatID) {
		// THE SITTING NAMES ITSELF. Hardcoding `docket` would register all four bench sittings as
		// docket rulings, which is the fact this field exists to stop being guessed — and it would
		// leave three of the four words undriven while the surface census counted the flag covered.
		if occasion == "" {
			occasion = "docket"
		}
		args = append(args, "--occasion", occasion)
	}
	return args
}

func newLockedRand(seed int64) *lockedRand {
	return &lockedRand{r: rand.New(rand.NewSource(seed))}
}

func (l *lockedRand) Intn(n int) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.r.Intn(n)
}

// run is the runner's resolved handle. NewRun rather than OpenRun: a runner is constructed
// before the directory necessarily exists, and a fuzz iteration that cannot resolve should
// fail on the assertion it was testing, not here.
//
// RESOLVED BY THE CONSTRUCTOR, NOT ON FIRST USE. It was written lazily — resolve once, cache,
// return the cached handle — on a base where the seats were serial. Phase 3 made them
// concurrent, and `envelopeFor` runs on a goroutine per seat, so every board read below reaches
// this from several goroutines at once: the lazy write is a data race that no reordering of the
// draws would make safe. A cached resolution is a snapshot, and a snapshot taken lazily is
// taken at a moment the caller does not control. newRunner does it once, up front, where both
// call sites already hold a runDir that resolves.
func (r *runner) run() record.Run { return r.runHandle }

// newRunner resolves the run once and hands back a runner whose handle is immutable.
func newRunner(bin, runDir string, rng *lockedRand) *runner {
	run, _ := record.NewRun(runDir)
	return &runner{bin: bin, runDir: runDir, runHandle: run, rng: rng, registered: map[string]bool{}}
}

// gapSet is a set of gap ids that concurrent seats mark and read. Each carries its OWN lock rather
// than r.mu: the seats mark these from envelopeFor, which runs on a goroutine per seat on purpose
// (sibling lanes contend on one run directory — the contention this gate exists to create), and a
// lock scoped to one set is held for a map access and nothing else, so it can neither span a
// subprocess nor nest inside a caller that already holds r.mu.
type gapSet struct {
	mu sync.Mutex
	m  map[string]bool
}

func (g *gapSet) add(ids ...string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.m == nil {
		g.m = map[string]bool{}
	}
	for _, id := range ids {
		g.m[id] = true
	}
}

func (g *gapSet) has(id string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.m[id]
}

// runner shells the real binary for one run's seats. runDir + bin are fixed per fuzz iteration.
type runner struct {
	// mu guards every mutable field below that has no guard of its own (each gapSet's lock is
	// named where it sits). The seats run concurrently as of #630 phase 3, and
	// these were single-threaded by accident rather than by design. It is held around the state
	// itself and NEVER across r.exec — a lock spanning a subprocess would put the serialization
	// back one layer down and buy nothing but a slower version of the old behaviour.
	mu          sync.Mutex
	bin, runDir string
	// runHandle is runDir RESOLVED, filled by the constructor and never written again. It is
	// the one field mu does not guard, and deliberately so: it is immutable after construction,
	// which is what makes it safe to read from the concurrent seats above.
	runHandle record.Run
	rng       *lockedRand
	// anchorOfferErr is the first lens sitting whose work list offered a citation or proof it does
	// not own.
	anchorOfferErr string
	// owedRecheck is, by seat, the latest reason dischargeOwed left an owed anchor standing.
	owedRecheck map[string]string
	// offeredToOwner counts, by kind, the owner sittings whose work list offered an anchor of that
	// kind for its first check (checkAnchorOffers).
	offeredToOwner map[string]int

	// progress is when a command of this run last finished (UnixNano) — what driveDebate's stall
	// tripwire reads. Atomic: every seat's goroutine writes it.
	progress atomic.Int64
	// maxEpochs is the epoch limit this run's config states; zero is setup's default, which every
	// run but the forced-limit seed's carries.
	maxEpochs int

	registered map[string]bool
	// lensSittings is each lens's sittings this run, by seat: the cost the retirement rule bounds.
	lensSittings map[string]int
	// blueSittings is blue-respond's sittings this run, and blueAgain those it sat on successors it
	// engaged (gblock's ruling R2): the cost the second sitting adds, and how often it is driven.
	blueSittings, blueAgain int
	// againOn is the successors blue engaged its next sitting on, which the engine sits it on next:
	// the plan's party names the chair's gaps, so the second sitting reads its own here.
	againOn []string
	// chairRegisters is how many times red-chair has registered — with no hook bracket in this
	// driver each register opens a chair sitting, so it is the epoch the record derives for the
	// next write — kept here so the driver can name a prior epoch to --carried-from without reading
	// it back out of a seat id, which carries none.
	chairRegisters int
	// declared is how many `declare` calls the record accepted — the count every role's work list
	// must carry, read back by the declarations oracle.
	declared int
	// minter: gap id -> the LENS seat whose mint created it. THE ORIGINATOR CLOSES (roundless
	// §III.B.3): the record refuses a `close` (without --carried-from) or a `regrade` on a gap
	// from any seat but the one whose mint created it, so every close and regrade this driver
	// issues has to be issued as that lens. The chair mints nothing now; its acts on a gap are
	// `carry`, `spot-check`, `position`, `verdict` and `dispatch`.
	minter map[string]string
	// planThisSitting is the plan `dispatch next` printed at the start of the chair sitting in progress.
	planThisSitting map[string]any
	// lastPassRefusal is the gate's refusal of a PASS the plan permitted — the chair's duty undone.
	lastPassRefusal string
	// passOverAnotherSeat names a blocker another seat must clear, standing when the plan permitted
	// the PASS the gate then refused. pass_permitted means only the chair's own items remain, so this
	// is the plan and the gate disagreeing (#1202), never the chair's lapse.
	passOverAnotherSeat string
	// passBlockersErr is the oracle's own failure to read the blocker list when the gate refused a
	// PASS the plan permitted. It fails the run: read as "no other seat's blocker" it would pass the
	// oracle it disabled.
	passBlockersErr string
	// inquiryRulingErr is the drive's first failure to read an inquiry's ruling. It fails the run:
	// read as "not ruled" it would have red rule a settled inquiry again, or blue leave a ruled one
	// unanswered, and the inquiry oracle would then find nothing to check.
	inquiryRulingErr string
	// inquiryIDErr is the first `inquiry propose` that SUCCEEDED and whose result carried no inquiry
	// id the harness could read. It fails the run: read as "nothing to record" it would leave the
	// record of a fate and the inquiry `--about` reference undriven behind a green sweep.
	inquiryIDErr string
	// inquiryRecords counts the fates the tool ACCEPTED on an id read from a propose result — the
	// drive inquiryIDOf feeds. The sweep gates it at zero.
	inquiryRecords int
	// closingsFiled counts the closings a party filed on a docketed gap its plan said it owed, and
	// closingsRefused the envelopes refused while one stood unfiled, so the refusal is driven too.
	// closingWaits counts the plans that held a closing owed on a gap and convened no bench on
	// it — the bench waiting; the sweep gates each at zero. positionsWithheld counts the chair
	// sittings that withheld a PASS their plan permitted and argued why, reported and not gated.
	closingsFiled, closingsRefused, closingWaits, positionsWithheld int
	// ordersRefused counts the edits the tool REFUSED for a section out of the order the report
	// ships in (orderRefused). The sweep gates it at zero. orderAdmittedErr is the first such edit
	// the tool recorded, or refused without naming the order; it fails the run.
	ordersRefused    int
	orderAdmittedErr string
	// spansRefused counts the cites the tool REFUSED for a source span the source's text does not
	// hold (citeSpan). The sweep gates it at zero. spanAdmittedErr is the first such cite the tool
	// recorded, or refused without saying the span is not in the text; it fails the run.
	spansRefused    int
	spanAdmittedErr string
	// copiesRead counts, per reading verb, the copy words its results named (noteCopy). The sweep
	// gates that corroborate named every seeded copy and verify a substituted one.
	copiesRead map[string]map[string]int
	// fragmentsDriven counts the cites driven to be refused for a span that starts mid-clause
	// (sourceFragment), and truncationsDriven for one that ends mid-clause (sourceTruncation);
	// refuse holds each to the refusal. bogusSpans rotates the three refused spans with no draw of
	// its own. The sweep gates both counts at zero.
	fragmentsDriven, truncationsDriven, bogusSpans int
	// reopenedReadied counts the plans that readied the owner of every reopened blocker they hold
	// (readiedForReopened), and reopenedPlans the plans holding one, whoever they readied.
	// proofReopenedReadied counts, by owner, the reopened PROOFS whose owner a plan readied; the
	// sweep gates both owners at zero. owedKept counts the lenses a plan's fold kept from retiring
	// because they owed work (three barren sittings or more, and a debt), and proofOwedKept those
	// whose debt held an unjudged proof; the sweep gates each at zero (THE DEBT GATE says why the
	// second is fed). reopenedOwnerErr is the first
	// reopened proof a plan gave an owner other than the run's proof owner; it fails the run.
	// owedKeptBy counts those folds by each kind of debt they held (owedKind).
	reopenedReadied, reopenedPlans int
	// unsoundReadied counts the plans holding an unsound-proof blocker that readied its judge or blue,
	// unsoundPlans the plans holding one, and unsoundAnswers the gaps a judge minted about a proof it
	// judged unsound. unsoundReadyErr is the first plan that held one and readied neither, with no
	// reason saying either had sat since. unsoundRepeats counts the repeat unsound judgements the tool
	// refused over the judge's open gap, and unsoundOverOpen the gaps a judge minted about a proof
	// while a gap of its own about that proof was open.
	unsoundReadied, unsoundPlans, unsoundAnswers int
	unsoundRepeats, unsoundOverOpen              int
	unsoundReadyErr                              string
	// unreadAnswers counts the gaps a lens minted about a citation whose source its check could not
	// read, unreadRepeats the repeat unreadable readings the tool refused over that gap, unreadRetires
	// blue's retires of such a citation, and failedCloseRefused the closes the tool refused of a gap
	// carrying a failed check that still stood.
	unreadAnswers, unreadRepeats, unreadRetires, failedCloseRefused int
	proofReopenedReadied                                            map[string]int
	owedKept, proofOwedKept                                         int
	owedKeptBy                                                      map[string]int
	reopenedOwnerErr                                                string
	// quotesRefused counts the corroborations the tool REFUSED for a quote the report does not
	// hold (sourceQuoteRefused). The sweep gates it at zero. quoteAdmittedErr is
	// the first such act the tool recorded, or refused without saying whose sentence the quote
	// is; it fails the run.
	quotesRefused    int
	quoteAdmittedErr string
	// bridgesRefused counts the supports_with_bridge readings refused for naming no --bridge;
	// bridgesNamed the ones recorded with one; bridgeAdmittedErr is the first kindless one the tool
	// recorded. The sweep gates both counts at zero and the error at set.
	bridgesRefused, bridgesNamed int
	bridgeAdmittedErr            string
	// quoteDrives counts the lens drives that reached sourceQuoteRefused, which fires on every seventh.
	quoteDrives int
	// repeatsRefused counts the second acts the tool refused in a sitting that holds one — a
	// position, a closing on a gap — each naming the act that answers the first. The sweep gates
	// it at zero.
	repeatsRefused int
	// budgetRefused counts the mints the tool refused a lens holding its whole mint budget open.
	// The sweep gates it at zero.
	budgetRefused int
	// ownAnswers counts the mints that landed answering a contradiction the minting lens's own
	// finding raised — the gaps outside its budget; the sweep gates it at zero. answerBudgetErr is
	// the first such mint the budget refused, which fails the run: a lens can always answer what it
	// raised.
	ownAnswers      int
	answerBudgetErr string
	// unclassifiedMinted counts the mints under the catch-all `unclassified` that landed, and
	// unclassifiedRefused the ones the record refused for carrying no --reason. The sweep gates
	// both at zero.
	unclassifiedMinted, unclassifiedRefused int
	// repeatErr is the first repeated act the tool admitted, or refused without naming the act
	// that answers the first. It fails the run: an act stands as filed, and the refusal of a
	// second is where a seat learns what answers a wrong one.
	repeatErr string
	// rulerFilingsRefused counts the filings the tool refused the seat that rules their subject —
	// the chair's grade motion, the bench's docket motion. The sweep gates it at zero.
	rulerFilingsRefused int
	// rulerFilingErr is the first such filing the tool admitted, or refused for another reason
	// than the gavel. It fails the run: a motion its own ruler filed is one no other seat can rule.
	rulerFilingErr string
	// closedRemandsRefused counts the bench's remands of a closed gap the tool refused naming the
	// close, trifleRemanded the remanded gaps that are not material whose exchange no plan has yet
	// readied, and trifleRemandsRouted those a plan readied (remandroute_test.go). The sweep gates
	// both counts at zero.
	closedRemandsRefused, trifleRemandsRouted int
	trifleRemanded                            map[string]bool
	// remandActs counts, by role, the acts on a remanded gap the remand's blocking item held a seat to
	// (answerRemands), and pastLimitRemandsRefused the bench's remands past the limit the tool refused
	// naming the dispositions that close the gap (remandroute_test.go). The sweep gates each at zero.
	remandActs              map[string]int
	pastLimitRemandsRefused int
	// actErr is the run's first act the tool answered otherwise than the drive expects
	// (acts_test.go): an admitted act refused, or a refused act admitted or refused for another
	// reason. driveDebate fails the run on it ahead of every oracle downstream of the act.
	actErr string
	// judged is, by seat, the anchors the seat has verified or reproduced in its sitting in
	// progress: the record keeps one such act per anchor per sitting, so a drive picking an anchor
	// for one asks here first (firstThisSitting). sit() clears a seat's set.
	judged map[string]map[string]bool
	// closedNow is the gaps each seat filed its closing on in its sitting in progress: a docket
	// motion it files on one of them is refused (record.requireNoClosingThisSitting).
	closedNow map[string]map[string]bool
	// lastRuleRefusal is the last grade-motion ruling the record refused, for the exit tally.
	lastRuleRefusal string
	// lensMints: lens seat -> mints that LANDED. The gaps a lens holds open are bounded by a budget
	// whose floor is the run's mintBudget (the report's size can only raise it), so the driver
	// spreads its mints over the lenses debate.js dispatches rather than raising the floor.
	lensMints map[string]int
	// verbatimGaps: gap id -> the fix_new text blue applied verbatim from red's own mint. The
	// estoppel drive reads it to build a --quote that red's OWN prescription must refuse.
	verbatimGaps map[string]string
	// inquiryIDs: the tool-assigned inquiry ids this run has proposed. `--about-kind inquiry`
	// CHECKS its reference against the record, so a drive needs a real one — a composed id
	// would drive the refusal and never the success.
	inquiryIDs []string
	// #277: the gap ids minted with --check-kind computation. Such a gap CANNOT be closed
	// until a proof answers it, so closeGap satisfies it first — otherwise the fuzzer
	// accumulates unclosable gaps and every open-gap-scaled drive grows with them (measured:
	// a uniform draw took the 60-run sweep from 47s to a 900s timeout).
	computationGaps gapSet
	// #62 Stage 2: disputes blue RAISED (event emitted + envelope ref), awaiting red's answer
	// next round — mirrors debate.js's pendingDisputes so the fuzz drives the docket machinery
	// through the ENVELOPE, not just the events. Each: {gap_id, dimension, proposed}.
	raised []map[string]any
	// presented records the gaps a responder was actually shown: a gap minted in the terminal
	// round never reaches blue, so its scenario was never dispatched and cannot be asserted.
	presented gapSet
	// evaluated records the gaps red actually SAT ON after blue had responded — the only ones
	// whose terminal fate red is answerable for. Blue repairing in the final round leaves a
	// satisfied gap legitimately open, because red never sits again.
	evaluated gapSet
	// reproduced records the PROVE gaps whose proof a lens re-ran and confirmed. Red closes on
	// THIS, not on its own assertion that a computation happened.
	reproduced gapSet
	// applyMisses counts, by CAUSE, every time the verbatim-apply branch declined to apply.
	// Reported beside the verbatim tally so a zero there names its reason instead of implying one.
	applyMisses map[string]int
	// estoppelMisses is the same accounting one layer on, for the drive that turns a verbatim
	// apply INTO an estoppel refusal. `verbatimApplied` proved the precondition and the gate then
	// reported `estoppels=0` with nothing between the two numbers to say which step lost it —
	// which is how a red gate could name the guard as gone when the drive had simply declined.
	estoppelMisses map[string]int
	// schemas holds each dispatch's envelope schema compiled, and schemaRefusals every envelope
	// one of them refused, by seat (envelopeschema_test.go).
	schemas        envelopeSchemas
	schemaRefusals []string
	// disputeRefusals is why a grade filing was REFUSED, per gap. The scenario oracle reports a
	// missing dispute event; without this it cannot distinguish a drive that never ran from one
	// that ran and was told no, and those want opposite fixes.
	disputeRefusals map[string]string
	// #111: every model an agent() call carried, one per dispatch ("unset" if absent). The tier
	// oracle asserts all equal the configured tier — map-free, needs no bulk-seat list here.
	models []string
	// petition docket: petitions a party seat FILED on the record, awaiting the petition sitting the
	// chair's next plan convenes. Each: {who, class, motion}. Mirrors the disputes machinery
	// (raised) so the fuzz drives the petition/petition-rule path end to end.
	petitioned []map[string]any
	// petitionsFiled is every petition filed, with the chair sitting it was filed in; petitionRuledAt
	// is the bench occasion each was ruled at. The petition oracle reads the two.
	petitionsFiled  []filedPetition
	petitionRuledAt map[string]string
	// forceUnverified drives the run to the UNVERIFIED terminal verdict, which the random sweep
	// reaches about once in sixty — debate.js computes UNVERIFIED only when the bench declares
	// deadlock AND gaps remain open, and a bench that disposes its whole docket clears the board
	// and earns red a further sitting instead. A 1-in-60 outcome cannot carry an enum gate: the
	// sweep failed on `outcome --as UNVERIFIED` at N=40 having passed at N=60 on luck, which is a
	// flake either way. So it is driven the way HALTED is — deterministically, by a dedicated
	// test — rather than exempted. coverage_test.go says why an exemption is the wrong answer
	// here, and it is right.
	forceUnverified bool
	// forceEstoppel drives the ORDERED PAIR the estoppel guard needs — red prescribes concrete
	// text, blue applies it verbatim, red then mints against that same text — on one dedicated
	// seed, instead of hoping the sweep draws it.
	//
	// Left to the draw it is a chain of coins: the gap must be minted dirApply (2 in 10), must
	// carry a concrete proposal (40%), blue must reach the apply branch with the span still on
	// the page, and the estoppel mint must then win a 35% coin. Measured across 13 uncached
	// sweeps on clean main, that chain failed 6 times — a 46% red on the gate that guards a tag,
	// which runs it ONCE. More verbatim applications do not help: reds were observed with 6, 8
	// and 9 of them, so what is missing is not applications but the PAIR.
	forceEstoppel bool
	// forceVerified drives the VERIFIED terminal verdict on one seed instead of leaving it to the
	// draw. VERIFIED needs a board with no gap left at its limit, and every gap the sweep mints has
	// about a one-in-four chance of ending there (see directives), so the chance a run reaches
	// VERIFIED falls with every gap a run mints: seven lenses mint about seven.
	forceVerified bool
	// forceReopenedProof drives one proof through judgement and reopening on a reopenedProofSeed
	// run: blue proves the sentence the edit drive rewrites, the owner judges the proof when its
	// work list offers it, blue rewrites the sentence, and the next plan readies the owner for the
	// reopened proof. Each step reads the record, never a count of sittings, so the chain holds on
	// every interleaving. reopenStage is the last step it reached, which the run's error names
	// when no plan readied the owner.
	forceReopenedProof bool
	reopenStage        string
	// forceUnsoundProof drives one proof to a standing unsound judgement on unsoundProofSeed's run:
	// blue proves the seeded sentence, the owner judges it unsound when its work list offers it, the
	// owner's gap carries it, and the next plan readies the judge or blue. unsoundStage is the last
	// step it reached, which the run's error names when no plan was readied for it.
	forceUnsoundProof bool
	unsoundStage      string
	// forceOwedProof makes every blue-respond sitting of owedProofSeed's run prove the seeded sentence
	// afresh, so each plan after one holds a proof nobody has checked, owed by the computation lens;
	// owedProofs counts them, naming each script.
	forceOwedProof bool
	owedProofs     int
	// forceAbsentSpan drives one cite of a span its source does not hold on absentSpanSeed's run,
	// and absentSpanDriven is whether it has fired.
	forceAbsentSpan, absentSpanDriven bool
	// provedExpectedError fires the --expect-error drive once per run. A proof whose FAILURE is
	// the result — proving a path is absent, a command missing — is a distinct contract from the
	// answering proof beside it, and the sweep had never passed the flag that says so.
	provedExpectedError bool
	// registering closes when a seat's register HAS LANDED. `registered` above answers whether
	// one was started; this answers whether it finished, which is what every append needs.
	registering map[string]chan struct{}
	// forceHalt makes the judge-petition sitting rule HALT (the dedicated halt-path test), instead
	// of the granted/denied the random path uses. A halt ends the run HALTED; kept out of the
	// random oracle on purpose (it reshapes every downstream expectation).
	forceHalt bool
}

func (r *runner) coin(pct int) bool { return r.rng.Intn(100) < pct }

// maybe runs fn with pct% probability — the coin-gated-action idiom, without the `if` nesting.
func (r *runner) maybe(pct int, fn func()) {
	if r.coin(pct) {
		fn()
	}
}

// applyTurn alternates the two verbatim-apply arms ACROSS THE WHOLE SWEEP. Package-level and
// atomic because runs execute in parallel goroutines and each reaches the apply branch only
// about once — a per-run counter never reaches its second turn. See its use for the measurement.
var applyTurn atomic.Int64

// inquiryIDOf pulls the tool-assigned inquiry id out of a propose result, or "" when the result
// names none.
func inquiryIDOf(out string) string { return inquiryIDPat.FindString(out) }

// THE SHAPE IS THE TABLE's. The record mints the id, it comes back on stdout and is never
// recomposed here; the pattern that finds it there is anchor.IDPattern, the one id matcher, so a
// change to the id shape moves this reader with it. A pattern kept by hand matches nothing after
// such a change, and every drive fed from it — the record of a proposed inquiry's fate, the inquiry
// `--about` reference — silently stops: see runner.inquiryIDErr and the inquiry-record gate.
var inquiryIDPat = regexp.MustCompile(`\b` + anchor.IDPattern("inquiry") + `\b`)

// gapIDPat folds a gap id out of a stated reason, so the reasons tally by what they say.
var gapIDPat = regexp.MustCompile(`\b` + anchor.IDPattern("gap") + `\b`)

// unmintedFinding and unmintedGap are ids of the RIGHT SHAPE that no run mints (a mint is four
// random bytes; these are one value in 2^32). The probes that pass them test the record's
// refusal of a reference it does not hold — a value of the wrong shape is refused by the flag
// before any view is asked, and would pass the probe without reaching what it probes.
var (
	unmintedFinding = anchor.ID("finding", [4]byte{0xff, 0xff, 0xff, 0xff})
	unmintedGap     = anchor.ID("gap", [4]byte{0xff, 0xff, 0xff, 0xff})
)

// refusedByShape reports whether out is a flag's SHAPE refusal of id (flags.ShapedValue.Set)
// rather than the record's refusal of a reference it does not hold.
func refusedByShape(out []byte, id string) bool {
	return strings.Contains(string(out), strconv.Quote(id)+" is not ")
}

// cmd is a small fluent builder for a seat verb — `<role> <verb> --seat-id <seatID> …` (exec
// appends --run). It collapses the conditional-flag arg-slice boilerplate: set() always adds a
// flag, on() adds it with pct% probability, bare() adds a boolean flag. run() shells the binary.
type cmd struct {
	r    *runner
	args []string
}

// do takes the verb as it is TYPED, so a verb under a subgroup is passed as the words a seat
// would write ("inquiry propose") rather than needing a second parameter for the group.
// do builds one invocation. THE ROLE IS GONE FROM THE ARGS: the tree is scoped to the seat, so the
// seat id is the only thing that says which verbs exist.
func (r *runner) do(verb, seatID string) *cmd {
	return &cmd{r: r, args: append(strings.Fields(verb), "--seat-id", seatID)}
}
func (c *cmd) set(flag, val string) *cmd { c.args = append(c.args, flag, val); return c }

// keyed adds, pct times in a hundred, a --key drawn from everything the act states so far: a key
// names one act, so two acts under one key state the same thing (contentKey).
func (c *cmd) keyed(pct int, prefix string) *cmd {
	if c.r.coin(pct) {
		c.args = append(c.args, "--key", contentKey(prefix, c.args))
	}
	return c
}
func (c *cmd) bare(flag string) *cmd { c.args = append(c.args, flag); return c }
func (c *cmd) on(pct int, flag, val string) *cmd {
	if c.r.coin(pct) {
		c.args = append(c.args, flag, val)
	}
	return c
}

// noteInquiry remembers a tool-assigned inquiry id so the --about drive can name a real one.
func (r *runner) noteInquiry(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inquiryIDs = append(r.inquiryIDs, id)
}

// noteInquiryIDErr keeps the first propose result no inquiry id could be read from.
func (r *runner) noteInquiryIDErr(msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.inquiryIDErr == "" {
		r.inquiryIDErr = msg
	}
}

// seedReport is the report every run freezes: text before the first heading, which is one of the
// report's own sections, and the section that closes the text (reportdoc.TextOrder).
const seedReport = "# § fuzz\n\nA § fuzz sentence to anchor findings.\n\nThe cost is rising over time across the whole sampled corpus.\n\n## Open questions\n\nWhether the trend outlasts the sample.\n"

// orderRefused drives an edit that would put a section out of the order the report ships in — one
// written after the section that closes the text, or a TL;DR after text of the report's own — and
// holds the tool to refusing it with the order named and the report as it stood.
func (r *runner) orderRefused(seatID string) {
	before, err := reportproj.RenderFromRecord(r.run())
	if err != nil || !strings.Contains(before, "Whether the trend outlasts the sample") {
		return
	}
	text := "Whether the trend outlasts the sample.\n\n## Appendix\n\nA section after the last one"
	if r.coin(50) {
		text = "Whether the trend outlasts the sample.\n\n## TL;DR\n\nA summary at the end"
	}
	out, err := r.exec("edit", "--seat-id", seatID, "--quote", "Whether the trend outlasts the sample", "--new", text, "--reason", "fuzz: a section out of order")
	after, _ := reportproj.RenderFromRecord(r.run())
	r.mu.Lock()
	defer r.mu.Unlock()
	switch {
	case err == nil || after != before:
		if r.orderAdmittedErr == "" {
			r.orderAdmittedErr = seatID + " put a section out of order and the edit was recorded: " + firstLine(out)
		}
	case !strings.Contains(out, reportdoc.TextOrder):
		if r.orderAdmittedErr == "" {
			r.orderAdmittedErr = seatID + "'s edit was refused without naming the order the text stands in (" + reportdoc.TextOrder + "): " + firstLine(out)
		}
	default:
		r.ordersRefused++
	}
}

// bogusSpan is which span a cite typed to be refused types.
type bogusSpan int

const (
	spanAbsent    bogusSpan = iota // words the source does not hold
	spanFragment                   // starts inside its clause (sourceFragment)
	spanTruncated                  // ends inside its clause (sourceTruncation)
)

// nextBogusSpan rotates the cites that type a span to be refused through a mid-clause start, a
// mid-clause end and words the source does not hold, drawing nothing from the run's RNG, and
// counts the fragments and truncations.
func (r *runner) nextBogusSpan() bogusSpan {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.bogusSpans++
	switch r.bogusSpans % 3 {
	case 1:
		r.fragmentsDriven++
		return spanFragment
	case 2:
		r.truncationsDriven++
		return spanTruncated
	}
	return spanAbsent
}

// citeSpanRefused holds the tool to refusing a cite whose source span the source's text does not
// hold: an admitted one would put blue's own words on the record as the source's. A keyed retry
// that differs from its first cite is refused for the key before the span is read, and counts as
// neither.
func (r *runner) citeSpanRefused(seatID, out string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch {
	case err == nil:
		if r.spanAdmittedErr == "" {
			r.spanAdmittedErr = seatID + " cited a span its source does not hold and the cite was recorded: " + firstLine(out)
		}
	case strings.Contains(out, "already names citation"):
	case !strings.Contains(out, "not in the text at"):
		if r.spanAdmittedErr == "" {
			r.spanAdmittedErr = seatID + "'s cite of a span its source does not hold was refused without saying so: " + firstLine(out)
		}
	default:
		r.spansRefused++
	}
}

// sourceQuoteRefused drives a lens corroborating with the SOURCE's sentence as the claim, under any
// outcome, and holds the tool to refusing it: the quote is the report's sentence whatever the source
// did for it. Under `refutes` or `absent` a quote the report does not hold is the subject of a PASS
// blocker no finding can quote, so an admitted one ends the run UNVERIFIED with nobody to ask. A
// verification takes no quote — its claim is the sentence its anchor marks — so it has no such case.
//
// IT DRAWS NOTHING FROM THE RUN'S RNG. Every seventh lens drive fires, and the outcome turns with
// that count: a draw here would move every later draw of the run, and the forced
// UNVERIFIED seed ends elsewhere when one does.
func (r *runner) sourceQuoteRefused(seatID string) {
	r.mu.Lock()
	n := r.quoteDrives
	r.quoteDrives++
	r.mu.Unlock()
	if n%7 != 0 {
		return
	}
	n /= 7
	outcome := verifyOutcomes[n%len(verifyOutcomes)]
	out, err := r.exec(append([]string{"corroborate", "--seat-id", seatID, "--url", "https://fuzz.invalid/quoted/" + seatID, "--title", "fuzz source quoted for " + seatID,
		"--quote", "The fuzz source states this in its own words, and the report does not."},
		append([]string{"--as", outcome, "--confidence", "high", "--reason", "fuzz: the source's sentence, quoted as the claim"}, bridgeFor(outcome, n)...)...)...)
	r.mu.Lock()
	defer r.mu.Unlock()
	switch {
	case err == nil:
		if r.quoteAdmittedErr == "" {
			r.quoteAdmittedErr = seatID + " quoted the source's sentence under --as " + outcome + " and `corroborate` recorded it: " + firstLine(out)
		}
	case !strings.Contains(out, "REPORT's sentence"):
		if r.quoteAdmittedErr == "" {
			r.quoteAdmittedErr = seatID + "'s `corroborate --as " + outcome + "` was refused without saying the quote is the report's sentence: " + firstLine(out)
		}
	default:
		r.quotesRefused++
	}
}

// closingsOwedBy is the gaps this epoch's plan says seatID owes a closing on: the plan's
// closing_owed blockers it owns (record/closingowed.go) — read off the plan, as the workflow reads
// them, never off the docket the chair's sitting filed.
func (r *runner) closingsOwedBy(seatID string) []string {
	var out []string
	if r.planThisSitting == nil {
		return out
	}
	bs, _ := r.planThisSitting["blockers"].([]any)
	for _, b := range bs {
		m, _ := b.(map[string]any)
		if record.BlockerKind(fmt.Sprint(m["kind"])) != record.BlockerClosingOwed || fmt.Sprint(m["owner"]) != seatID {
			continue
		}
		if g, ok := m["subject"].(string); ok {
			out = append(out, g)
		}
	}
	return out
}

// fileClosings files the party's closing on each docketed gap its plan says it owes — the write path
// must admit each — and one time in ten first tries to end with it unfiled, which the record refuses.
// The forced-verified run files every one. Unique prose per act, so the report
// oracle can prove each rendered.
func (r *runner) fileClosings(seatID string) {
	for _, id := range r.closingsOwedBy(seatID) {
		// ONE IN TEN, THE SEAT TRIES TO END WITH THE CLOSING UNFILED, and the record refuses its
		// envelope naming it: a party's sitting cannot end while its closing is owed.
		if !r.forceVerified && r.coin(10) && r.closingOwed(seatID, id) {
			r.envelopeRefusedEarly(seatID, "the closing argument of "+seatID+" on it")
			r.mu.Lock()
			r.closingsRefused++
			r.mu.Unlock()
		}
		if _, ok := r.admitFresh(id, "closing", "--seat-id", seatID, "--id", id, "--reason", "closing-for-"+id+"-by-"+seatID); !ok {
			continue
		}
		r.noteClosed(seatID, id)
		r.mu.Lock()
		r.closingsFiled++
		r.mu.Unlock()
		// A SEAT THAT GOT ITS CLOSING WRONG TYPES IT AGAIN. The sitting holds one per gap, so the
		// second is refused and the refusal names the act that answers the first.
		r.maybe(10, func() {
			r.repeatRefused(seatID, "closing", "--id", id, "--reason", "fuzz: the closing again, restated")
		})
	}
}

// closingOwed reports whether the record holds seatID's closing on gapID as owed (a closing_owed
// blocker), read when the drive is about to end with it standing.
func (r *runner) closingOwed(seatID, gapID string) bool {
	bs, err := record.PassBlockers(r.runHandle)
	if err != nil {
		return false
	}
	for _, b := range bs {
		if b.Kind == record.BlockerClosingOwed && b.Owner == seatID && b.Subject == gapID {
			return true
		}
	}
	return false
}

// noteClosed marks gapID as closed on by seatID in its sitting in progress.
func (r *runner) noteClosed(seatID, gapID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closedNow == nil {
		r.closedNow = map[string]map[string]bool{}
	}
	if r.closedNow[seatID] == nil {
		r.closedNow[seatID] = map[string]bool{}
	}
	r.closedNow[seatID][gapID] = true
}

// closedThisSitting reports whether seatID filed its closing on gapID in its sitting in progress.
func (r *runner) closedThisSitting(seatID, gapID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.closedNow[seatID][gapID]
}

// dischargeOwed clears what the seat's sitting still owes before its envelope, by the acts that
// clear each (record.EnvelopeRefusal refuses an envelope while one stands): its closing on each
// docketed gap, and for blue-respond a proof for each computation gap its sitting is engaged on and
// nothing answers, and the remedy of each failed check (fixFailed); for the anchor's owner each
// citation or proof the record holds unchecked; and each contradiction the lens read, and the gap
// about each check of its that failed (answerContradictions and answerFailed, in the lens branch).
func (r *runner) dischargeOwed(seatID string) {
	if bs, err := record.PassBlockers(r.runHandle); err == nil {
		for _, b := range bs {
			if b.Kind == record.BlockerClosingOwed && b.Owner == seatID {
				if _, ok := r.admitFresh(b.Subject, "closing", "--seat-id", seatID, "--id", b.Subject, "--reason", "closing-for-"+b.Subject+"-by-"+seatID); ok {
					r.noteClosed(seatID, b.Subject)
				}
			}
		}
	}
	// A REMAND'S ITEM HOLDS BLUE AND THE MINTING LENS UNTIL EACH ACTS ON THE GAP (answerRemands).
	if seatID == "blue-respond" || strings.HasPrefix(seatID, "red-lens-") {
		r.answerRemands(seatID)
	}
	if seatID == "blue-respond" {
		r.fixFailed(seatID)
		w, err := record.WorkOfSeat(r.run(), "blue", seatID)
		if err != nil || w.Engaged == nil {
			return
		}
		for _, g := range w.Open {
			if !g.AwaitingProof || !slices.Contains(w.Engaged.GapIDs, g.ID) {
				continue
			}
			name := "fuzz-owed-" + g.ID + ".js"
			if err := r.writeScript(name, "console.log('answers "+g.ID+"');"); err == nil {
				r.admit("prove", "--seat-id", seatID, "--quote", "§ fuzz sentence", "--script", name, "--answers", g.ID,
					"--reason", "fuzz: the computation its sitting owes")
			}
		}
		return
	}
	if seatID != "red-lens-evidence" && seatID != proofOwner {
		return
	}
	// AN ANCHOR THE DRIVE ALREADY TOUCHED THIS SITTING IS TRIED AGAIN, its refusal not expected either
	// way: the touch may have been a verify the drive meant to be refused (a --page on a pageless
	// source), which recorded nothing and left the item standing; one that was recorded makes this
	// second act a refused repeat, and the item cleared or not this sitting's to clear. The envelope
	// check after the drive is what says the item is gone, and what the re-check got back is kept for
	// it (owedRecheck), so an envelope refused over the item says why the drive did not clear it.
	if _, ok := r.board(); !ok {
		r.noteOwedRecheck(seatID, "the board did not read, so no unchecked anchor was re-checked")
	}
	for _, it := range r.unchecked() {
		if !(it.Sha != "" && seatID == proofOwner || it.Sha == "" && seatID == "red-lens-evidence") {
			continue
		}
		var a *cmd
		if it.Sha != "" {
			a = r.do("reproduce", seatID).set("--anchor", it.Anchor).set("--as", "sound").
				set("--reason", "fuzz: read the script against the sentence as it now stands")
		} else {
			a = r.do("verify", seatID).set("--anchor", it.Anchor).set("--as", "supports").set("--confidence", pick(r.rng, verifyConfidence)).
				set("--reason", "fuzz: read the sentence as it now stands against its source")
		}
		if r.firstThisSitting(seatID, it.Anchor) {
			a.admit()
		} else {
			out, err := a.result()
			noteEither(a.args, "a reopened anchor the drive touched earlier in the sitting", err == nil)
			if err != nil {
				r.noteOwedRecheck(seatID, it.Anchor+", touched earlier this sitting, re-checked and refused: "+firstLine(out+" "+err.Error()))
			}
		}
	}
}

// noteOwedRecheck keeps the latest word on why dischargeOwed left an owed anchor standing for seatID,
// which envelopeEnds names beside a refused envelope.
func (r *runner) noteOwedRecheck(seatID, why string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.owedRecheck == nil {
		r.owedRecheck = map[string]string{}
	}
	r.owedRecheck[seatID] = why
}

// fixFailed is blue's half of a failed check (record.BlockerFailedCheck): for every proof a lens's
// standing judgement calls unsound, it rewrites the sentence the proof backs, prefacing its words and
// keeping the anchor — the fix the work list names, which leaves the proof unchecked for a fresh
// judgement; for every citation whose source no check could read, it retires the citation and keeps
// its sentence (`retire --anchor`), the remedy the ruling names.
func (r *runner) fixFailed(seatID string) {
	// ONE PROOF AT A TIME, FROM THE BLOCKERS AS THEY STAND AFTER THE LAST FIX: an edit of a sentence
	// reopens every proof on it, so a proof read before it may stand no longer.
	fixed := map[string]bool{}
	for {
		bs, err := record.PassBlockers(r.runHandle)
		if err != nil {
			r.noteActErr("the blockers could not be read to say which failed checks " + seatID + " owes a remedy: " + err.Error())
			return
		}
		var b record.Blocker
		for _, x := range bs {
			if x.Kind == record.BlockerFailedCheck && !fixed[x.Subject] {
				b = x
				break
			}
		}
		if b.Subject == "" {
			return
		}
		fixed[b.Subject] = true
		if b.Detail == "" {
			// ONE TIME IN TEN, BLUE TRIES TO END WITH THE CITATION STANDING, and the record refuses its
			// envelope naming it.
			if r.coin(10) {
				r.envelopeRefusedEarly(seatID, "check of citation "+b.Subject)
			}
			r.do("retire", seatID).set("--anchor", b.Subject).
				set("--reason", "fuzz: no lens could read the source of "+b.Subject+", so the sentence stands without it").admit()
			r.mu.Lock()
			r.unreadRetires++
			r.mu.Unlock()
			continue
		}
		cur, err := reportproj.RenderFromRecord(r.run())
		if err != nil {
			r.noteActErr("the report could not be rendered to fix proof " + b.Subject + ": " + err.Error())
			return
		}
		// ONE TIME IN TEN, BLUE TRIES TO END WITH THE PROOF UNFIXED, and the record refuses its envelope
		// naming it.
		if r.coin(10) {
			r.envelopeRefusedEarly(seatID, "judged the sentence proof "+b.Subject+" backs unsound")
		}
		words := wordsBefore(cur, b.Subject)
		if words == "" {
			r.noteActErr("proof " + b.Subject + " is judged unsound and no prose standing once in the report precedes its anchor, so the drive has no span to fix")
			continue
		}
		// PREPENDED, so every span the drive quotes from this prose still stands in it: the seeded
		// anchor sentence carries the drive's corroborations and findings.
		r.do("edit", seatID).set("--quote", words).set("--new", "Recomputed, "+words).
			set("--reason", "fuzz: the sentence now says what the script behind "+b.Subject+" shows").admit()
	}
}

// wordsBefore is the run of prose standing right before the anchor run that holds id's token in cur
// — back to the previous anchor token or line break — when it is non-blank and occurs once in cur;
// "" otherwise.
func wordsBefore(cur, id string) string {
	at := strings.Index(cur, anchor.Token(id))
	if at < 0 {
		return ""
	}
	// THE RUN THE TOKEN STANDS IN: anchors abut (anchor.SkipRun), so the prose ends where the first
	// token of the run begins.
	ends := map[int]int{}
	var last []int
	anchor.Each(cur[:at], func(start, end int, _ string) {
		ends[end] = start
		last = append(last, end)
	})
	for {
		start, abuts := ends[at]
		if !abuts {
			break
		}
		at = start
	}
	start := 0
	for _, e := range last {
		if e <= at {
			start = e
		}
	}
	if nl := strings.LastIndexByte(cur[:at], '\n'); nl+1 > start {
		start = nl + 1
	}
	words := strings.TrimSpace(cur[start:at])
	if words == "" || strings.Count(cur, words) != 1 {
		return ""
	}
	return words
}

// chairPosition files the chair's position: owed where its sitting withheld a PASS the plan
// permitted, and at any other sitting the chair's to file — so it is filed there on a coin.
func (r *runner) chairPosition(seatID string, owed bool) {
	if !owed && !r.coin(30) {
		return
	}
	if _, err := r.exec("position", "--seat-id", seatID, "--reason", "narrative from "+seatID); err != nil {
		return
	}
	if owed {
		r.mu.Lock()
		r.positionsWithheld++
		r.mu.Unlock()
	}
	// A SEAT THAT GOT ITS POSITION WRONG TYPES IT AGAIN. The sitting holds one, so the second is
	// refused and the refusal names the act that answers the first.
	r.maybe(20, func() {
		r.repeatRefused(seatID, "position", "--reason", "fuzz: the position again, restated")
	})
}

// noteInquiryRecord counts a fate the tool accepted on an id read from a propose result.
func (r *runner) noteInquiryRecord() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inquiryRecords++
}

// aboutAnchor picks one --about-kind/--about pair, or ("","") when nothing on the record can be
// named yet. It is the anchor for a defect that is NOT report text (#787, #742).
//
// EVERY REFERENCE IS REAL, and for two of the three kinds it has to be: ResolveAbout checks an
// inquiry ref against the inquiries this run proposed and a gap ref against the board, so a composed
// id would drive the refusal forever and the success path would read as covered. `section` is
// deliberately unchecked by the tool — blue may rename a heading mid-run, and refusing on a stale
// one would refuse the finding rather than the staleness — so the seeded heading is honest there.
func (r *runner) aboutAnchor() (kind, ref string) {
	r.mu.Lock()
	inq := append([]string{}, r.inquiryIDs...)
	r.mu.Unlock()
	kinds := []string{"section"}
	if len(inq) > 0 {
		kinds = append(kinds, "inquiry")
	}
	gaps := r.openGaps()
	if len(gaps) > 0 {
		kinds = append(kinds, "gap")
	}
	// A CITATION OR PROOF BY ITS ANCHOR: the subject a failed check's answering gap names
	// (answerFailed), and one any finding or gap may name, which answers no failed check.
	var anchors []string
	if b, ok := r.board(); ok {
		ev := record.EvidenceJSONOf(b.Events, b.At)
		for _, c := range ev.Sources {
			anchors = append(anchors, c.Anchor)
		}
		for _, p := range ev.Proofs {
			anchors = append(anchors, p.Anchor)
		}
	}
	if len(anchors) > 0 {
		kinds = append(kinds, "anchor")
	}
	switch pick(r.rng, kinds) {
	case "anchor":
		return "anchor", anchors[r.rng.Intn(len(anchors))]
	case "inquiry":
		return "inquiry", inq[r.rng.Intn(len(inq))]
	case "gap":
		return "gap", gaps[r.rng.Intn(len(gaps))]
	default:
		// The seeded report's own heading. Unchecked by the tool, so it stays honest even after
		// blue's edits move the prose underneath it.
		return "section", "§ fuzz"
	}
}

// noteInquiry and aboutAnchor sit beside noteVerbatimGap

// noteVerbatimGap remembers a gap whose prescribed text blue has just applied VERBATIM, and the
// text itself. That pair is the estoppel precondition: red minting a fresh gap whose --quote
// overlaps this text is red attacking its own prescription, and the tool refuses it.
//
// It is carried here rather than re-derived because the fuzz's red drive has to choose the quote
// BEFORE it can know whether one applies, and EstoppelConflict is not exported to this package.
func (r *runner) noteVerbatimGap(gapID, fixNew string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.verbatimGaps == nil {
		r.verbatimGaps = map[string]string{}
	}
	r.verbatimGaps[gapID] = fixNew
}

// aVerbatimGap returns one gap/text pair blue applied verbatim, or "" if none has landed yet.
func (r *runner) aVerbatimGap() (string, string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, fn := range r.verbatimGaps {
		return id, fn // map order is fine: any applied pair estops equally
	}
	return "", ""
}

// noteApplyMiss records WHY the verbatim-apply branch declined, so the estoppel precondition's
// count of zero arrives with a cause rather than inviting the reader to guess one.
func (r *runner) noteApplyMiss(why string) {
	if r.applyMisses == nil {
		r.applyMisses = map[string]int{}
	}
	r.applyMisses[why]++
}

// rulerFilingRefused files a motion as the seat that rules its subject, and holds the tool to
// refusing it for that reason (record's rulerfiles.go, through the surface's refusal).
func (r *runner) rulerFilingRefused(seatID string, args ...string) {
	out, err := r.exec(append([]string{args[0], args[1], args[2], "--seat-id", seatID}, args[3:]...)...)
	r.mu.Lock()
	defer r.mu.Unlock()
	what := seatID + "'s `" + strings.Join(args[:3], " ") + "`"
	switch {
	case err == nil:
		if r.rulerFilingErr == "" {
			r.rulerFilingErr = what + " was admitted, and that seat rules the subject: " + firstLine(out)
		}
	case !strings.Contains(out, "so `file` is not on your surface") || !strings.Contains(out, "is one no other seat can rule"):
		if r.rulerFilingErr == "" {
			r.rulerFilingErr = what + " was refused without saying that its ruler files nothing on the subject: " + firstLine(out)
		}
	default:
		r.rulerFilingsRefused++
	}
}

// noteEstoppelMiss records WHY the estoppel drive declined, so `estoppels=0` names its reason.
func (r *runner) noteEstoppelMiss(why string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.estoppelMisses == nil {
		r.estoppelMisses = map[string]int{}
	}
	r.estoppelMisses[why]++
}

// firstLine keeps a refusal's diagnosis and drops the help cobra staples to it — the tally is a
// histogram, and a full help page per key makes it unreadable.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// repeatRefused files an act a second time in a sitting that holds one of it, and holds the tool to
// refusing it with the act that answers the first (record's superseders, through its refusal). It
// is driven only after the first was admitted: a second act after a refused first is a first.
func (r *runner) repeatRefused(seatID string, args ...string) {
	out, err := r.exec(append([]string{args[0], "--seat-id", seatID}, args[1:]...)...)
	r.mu.Lock()
	defer r.mu.Unlock()
	what := seatID + "'s second `" + args[0] + "` in one sitting"
	switch {
	case err == nil:
		if r.repeatErr == "" {
			r.repeatErr = what + " was admitted: " + firstLine(out)
		}
	case !strings.Contains(out, "The first stands. If it was wrong, say so in "):
		if r.repeatErr == "" {
			r.repeatErr = what + " was refused without naming the act that answers the first: " + firstLine(out)
		}
	default:
		r.repeatsRefused++
	}
}

// THE SECOND DISPUTE DRIVER IS GONE. `raiseDisputes` sat here, fully written and NEVER CALLED —
// the live driver is in blueRespondTo's dirDisputeWon/dirDisputeLost arm. Migrating it to the
// motion verbs would have produced a second correct-looking copy of a path nothing runs, and the
// command tally would have gone on reporting the verb as driven either way.
//
// answerDisputes is red's Stage-2 answer to blue's pending disputes: EMIT a dispute-respond event
// (rationale on the record) and return the ROUTING REF ({gap_id, dimension, response}) for the
// envelope's dispute_responses. Clears the pending set. A random accept/reject exercises both the
// accepted-delta and rejected-held docket branches.
func (r *runner) answerDisputes(seatID string) []map[string]any {
	var refs []map[string]any
	// FROM THE RECORD, NOT FROM MEMORY. r.raised held the disputes blue filed in its last sitting
	// and was cleared whether or not the ruling landed, so a refused ruling — or a motion filed
	// in a sitting this chair never saw — stood unruled forever, and PASS was refused on it at
	// the exit ("4 motion(s) filed and never ruled"). The chair's prompt says the same: rule the
	// grade motions blue filed, read from the motions projection, never counted by hand.
	pending := r.raised
	// THE GROUP IS ASKED, NOT SPELLED. `motions` moved to `inquest`, and a literal "show" here read
	// the projection through a group that no longer holds it. This is the CHAIR's act (see the call
	// site) and the chair carries inquest, so no membership check is needed.
	//
	// THE err == nil GUARD STAYS, and what makes it safe is elsewhere: the per-role sweep drives
	// `inquest motions` for the chair STRICTLY, so a broken read fails the run there rather than
	// falling back to `r.raised` unseen. Tolerance here with no strict driver anywhere would be how
	// the from-memory defect above came back.
	if out, err := r.exec(seat.GroupOf("motions"), "motions", "--seat-id", seatID); err == nil {
		var page struct {
			Motions []struct {
				ID      string            `json:"id"`
				Subject string            `json:"subject"`
				Ruled   bool              `json:"ruled"`
				GapID   string            `json:"gap_id"`
				Fields  map[string]string `json:"fields"`
			} `json:"motions"`
		}
		if json.Unmarshal([]byte(out), &page) == nil {
			pending = nil
			for _, m := range page.Motions {
				if m.Subject == "grade" && !m.Ruled {
					pending = append(pending, map[string]any{"motion_id": m.ID, "gap_id": m.GapID, "dimension": m.Fields["dimension"]})
				}
			}
		}
	}
	for _, d := range pending {
		id, _ := d["gap_id"].(string)
		dim, _ := d["dimension"].(string)
		// THE ANSWER FOLLOWS THE SCENARIO, not a coin. The gap was minted WON or LOST, and
		// both branches must run: an accepted regrade moves the board's mass (the
		// accepted-delta path), a rejected one is held a round and then dockets. Coining it
		// meant red's answer bore no relation to the dispute it was answering.
		resp := "rejected"
		if r.scenarioOf(id) == dirDisputeWon {
			resp = "accepted"
		}
		mid, _ := d["motion_id"].(string)
		if mid == "" {
			continue
		}
		// THE RULER FILES NO MOTION ON ITS OWN SUBJECT: the chair reaching for a grade motion on a
		// gap whose motion it is about to rule is refused, and the motion it rules is blue's.
		if r.coin(25) {
			r.rulerFilingRefused(seatID, "motion", "grade", "file", "--id", id, "--dimension", pick(r.rng, disputeDims),
				"--proposed", "high", "--reason", "fuzz: the chair contesting a grade it rules on")
		}
		if _, err := r.exec("motion", "grade", "rule", "--seat-id", seatID, "--id", mid,
			"--as", resp, "--reason", "respond-rationale-for-"+id+"-by-"+seatID); err != nil {
			// KEPT, because the exit tally names a PASS refused over an unruled motion and could
			// not say why the ruling never landed.
			r.mu.Lock()
			r.lastRuleRefusal = firstLine(strings.TrimSpace(strings.SplitN(err.Error(), "\n", 2)[len(strings.SplitN(err.Error(), "\n", 2))-1]))
			r.mu.Unlock()
			continue
		}
		refs = append(refs, map[string]any{"gap_id": id, "dimension": dim, "response": resp})
	}
	r.raised = nil
	return refs
}

// filedPetition is one petition the fuzz filed: its motion id, its filer, and the chair sitting it
// was filed in — a petition filed before the last chair plan is one that plan must convene the
// bench for.
type filedPetition struct {
	id, filer    string
	chairSitting int
}

// maybePetition sometimes files a petition (W2c) ON THE RECORD — the only channel a petition has —
// tracking {who,class,motion} so the petition sitting the chair's next plan convenes can rule on it.
// Every party seat is eligible, the lenses included: a lens's envelope carries nothing, and the
// record is how its petition reaches the bench. The random path never returns a HALT ruling, so the run
// continues. It files nothing most of the time — a petition detours the run through a bench
// sitting, so it stays occasional.
func (r *runner) maybePetition(seatID string) {
	if !r.forceHalt && !r.coin(20) { // forceHalt guarantees a petition so the halt sitting fires
		return
	}
	r.filePetition(seatID)
}

// filePetition files one petition as seatID, and tracks it for the bench.
func (r *runner) filePetition(seatID string) {
	class := pick(r.rng, petitionClasses)
	basis := "fuzz petition basis from " + seatID
	out, err := r.exec("--json", "motion", "petition", "file", "--seat-id", seatID,
		"--class", class, "--relief", "fuzz relief", "--reason", basis)
	if err != nil {
		return
	}
	id := motionIDOf(out)
	if id == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.petitioned = append(r.petitioned, map[string]any{"who": seatID, "class": class, "motion": id})
	r.petitionsFiled = append(r.petitionsFiled, filedPetition{id: id, filer: seatID, chairSitting: r.chairRegisters})
}

// petitionsFiledN is how many petitions the run has filed, read under the lock the concurrent
// lenses append under.
func (r *runner) petitionsFiledN() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.petitionsFiled)
}

// rulePetitions is the petition sitting: it rules on every pending petition and returns the
// envelope rulings, clearing the docket. petition-rule takes only granted|denied (a halt is its
// OWN verb), so a forceHalt run emits `bench halt` and returns a halt ruling — the dedicated
// halt-path test; the random path rules granted/denied and the run continues.
func (r *runner) rulePetitions(seatID string) map[string]any {
	rulings := r.rulePending(seatID, "petition")
	env := map[string]any{"rulings": rulings}
	if r.forceHalt {
		// `bench halt` writes the record; the envelope's halt object is only what stops the
		// engine. The fake already drove the verb correctly before #329 — it was the PROMPT
		// that told a real judge to record a halt through petition-rule, where the enum refuses
		// it. The fuzz stayed green over a production path that could not work.
		haltOpinion := "fuzz judicial halt — safety boundary"
		r.admit("halt", "--seat-id", seatID, "--reason", haltOpinion)
		env["halt"] = map[string]any{"opinion": haltOpinion}
	}
	return env
}

// rulePending rules every pending petition at the bench sitting of the given occasion — the
// petition sitting a plan convened, the terminal sitting, or the assembly that holds the gavel for
// one filed after the last plan — and returns the envelope rulings.
func (r *runner) rulePending(seatID, occasion string) []any {
	rulings := arr()
	if r.petitionRuledAt == nil {
		r.petitionRuledAt = map[string]string{}
	}
	// Taken under the lock the filing lenses append under; one filed while these are ruled waits
	// for the next sitting.
	r.mu.Lock()
	pending := r.petitioned
	r.petitioned = nil
	r.mu.Unlock()
	for _, p := range pending {
		who, _ := p["who"].(string)
		class, _ := p["class"].(string)
		opinion := "fuzz ruling opinion for " + who
		ruling := "granted"
		if r.coin(50) {
			ruling = "denied"
		}
		// THE PETITION IS ALWAYS RULED ON ITS MERITS. A halt is a separate decision about the
		// RUN, on its own channel (#329) — both can be true, and the petition does not stop
		// being answered because the bench also ended the run.
		id, _ := p["motion"].(string)
		// A GRANT NAMES WHO IT BINDS. Relief with no addressee is what #360 measured — the engine
		// threaded it into one hardcoded prompt, so relief for any other seat reached nothing.
		binds := pick(r.rng, []string{"blue", "red", "both"})
		args := []string{"motion", "petition", "rule", "--seat-id", seatID, "--id", id,
			"--as", ruling, "--reason", opinion}
		if ruling == "granted" {
			args = append(args, "--binds", binds)
		}
		if _, err := r.exec(args...); err == nil {
			r.petitionRuledAt[id] = occasion
		}
		rulings = append(rulings, map[string]any{"petitioner": who, "class": class, "ruling": ruling, "opinion": opinion, "relief": opinion, "binds": binds})
	}
	return rulings
}

// petitionOracle holds every petition the run filed to its route: each is ruled, and one filed
// before the last chair plan is ruled at the sitting a plan convened for it — a petition sitting,
// or the terminal sitting at the exit — never left to the assembly's backstop. A halted run is
// exempt: the bench stopped it, and the outcome refusal exempts a halt too. "" when it holds.
func (r *runner) petitionOracle(halted bool) string {
	if halted {
		return ""
	}
	for _, p := range r.petitionsFiled {
		at, ruled := r.petitionRuledAt[p.id]
		switch {
		case !ruled:
			return fmt.Sprintf("petition oracle: %s, filed by %s in chair sitting %d, was never ruled", p.id, p.filer, p.chairSitting)
		case at != "petition" && at != "terminal" && p.chairSitting < r.chairRegisters:
			return fmt.Sprintf("petition oracle: %s, filed by %s in chair sitting %d of %d, was ruled at the %s sitting — a later plan should have convened the bench for it", p.id, p.filer, p.chairSitting, r.chairRegisters, at)
		}
	}
	return ""
}

// exec runs one seat command. THE ERROR CARRIES WHAT THE TOOL SAID.
//
// It returned exec.ExitError verbatim — `exit status 2` — so every caller that kept its error
// still learned nothing. The dispute oracle was instrumented to print the refusal and printed
// "exit status 2"; the refusal itself was in the output the error did not mention. Six defects
// this session were a discarded refusal, and the cost was always the distance between the
// refusal and the symptom. Attaching it here shortens that distance for every call site at once.
// dispatchNext is the chair's first act (plans/roundless.md §III.B.1): the record computes who is
// ready and RECORDS it; the driver relays the verb's JSON into the chair's envelope exactly as a
// live chair does. A refusal is the run's — reported, never papered over with a plan of the
// driver's own.
func (r *runner) dispatchNext(seatID string) map[string]any {
	out, err := r.exec("--json", "dispatch", "next", "--seat-id", seatID)
	if err != nil {
		r.noteEstoppelMiss("dispatch next refused: " + err.Error())
		return debatejs.WithKey(map[string]any{"head": 0, "parties": []any{}, "docket": []any{}, "remand_owed": []any{}, "exchanges": []any{}, "pass_permitted": false, "ceiling": false, "why": []any{"dispatch refused: " + err.Error()}, "max_epochs": 0, "epoch": 0, "final_sitting": false, "epoch_limit_reached": false, "stale_areas": []any{}, "blockers": []any{}})
	}
	// THE MACHINE FORM IS THE PLAN ALONE, copied whole into the envelope as a live chair copies it.
	var plan map[string]any
	if err := json.Unmarshal([]byte(out), &plan); err != nil || plan == nil {
		return debatejs.WithKey(map[string]any{"head": 0, "parties": []any{}, "docket": []any{}, "remand_owed": []any{}, "exchanges": []any{}, "pass_permitted": false, "ceiling": false, "why": []any{"dispatch output unreadable"}, "max_epochs": 0, "epoch": 0, "final_sitting": false, "epoch_limit_reached": false, "stale_areas": []any{}, "blockers": []any{}})
	}
	// RELAYED EXACTLY AS PRINTED. The verb emits every array as an array, never null, and the
	// engine refuses a plan missing one — so a patch here would hide the defect the relay check
	// exists to catch.
	return plan
}

// planReadies is whether the current plan names seatID among its parties.
func (r *runner) planReadies(seatID string) bool {
	parties, _ := r.planThisSitting["parties"].([]any)
	for _, p := range parties {
		if m, _ := p.(map[string]any); m != nil && m["seat_id"] == seatID {
			return true
		}
	}
	return false
}

// engaged is the gap ids the current plan dispatched seatID on — the party's gap_ids, or none
// when the plan does not name the seat (a fresh lens sitting, or a seat outside this epoch).
func (r *runner) engaged(seatID string) []string {
	var out []string
	if r.planThisSitting == nil {
		return out
	}
	parties, _ := r.planThisSitting["parties"].([]any)
	for _, p := range parties {
		m, _ := p.(map[string]any)
		if m == nil || m["seat_id"] != seatID {
			continue
		}
		ids, _ := m["gap_ids"].([]any)
		for _, id := range ids {
			if sid, ok := id.(string); ok {
				out = append(out, sid)
			}
		}
	}
	return out
}

// staleAreas is the current plan's stale areas — each retired lens whose pin the head moved
// past, by seat — which the chair's spot-check names before a PASS.
func (r *runner) staleAreas() []string {
	var out []string
	if r.planThisSitting == nil {
		return out
	}
	areas, _ := r.planThisSitting["stale_areas"].([]any)
	for _, a := range areas {
		m, _ := a.(map[string]any)
		if sid, ok := m["seat_id"].(string); ok {
			out = append(out, sid)
		}
	}
	return out
}

// chairEnvelope is what the chair returns: the plan it relayed — whose blockers are what the
// terminal bench sitting fires on — and the verdict it recorded this sitting, if any. Gaps,
// closures, dispute responses and petitions are the record's.
func (r *runner) chairEnvelope(seatID, verdict, printed string, responses []map[string]any) map[string]any {
	_ = responses // grade motions are ruled on the record; the envelope no longer restates them
	r.mu.Lock()
	filed := len(r.petitionsFiled)
	r.mu.Unlock()
	r.maybePetition(seatID)
	r.mu.Lock()
	petitioned := len(r.petitionsFiled) > filed
	r.mu.Unlock()
	// THE PLAN RELAYED IS THE ONE PRINTED LAST. A verdict prints the plan carrying it, and the
	// chair relays that; a petition filed after it, or a sitting that recorded no verdict, asks the
	// dispatch again — the workflow reads what still stands at the end of the sitting off the
	// relayed plan, and capture holds every relayed plan to the record there.
	if verdict == "" || petitioned {
		r.planThisSitting = r.dispatchNext(seatID)
	} else {
		var plan map[string]any
		if err := json.Unmarshal([]byte(printed), &plan); err != nil || plan == nil {
			r.noteActErr(fmt.Sprintf("the %s verdict's machine form is not the plan: %q", verdict, firstLine(printed)))
			plan = map[string]any{}
		}
		r.planThisSitting = plan
	}
	// THE ENVELOPE IS THE PLAN, AND THE VERDICT IS ITS FIELD: the engine ends the debate on it. A
	// relayed plan that does not carry the verdict this sitting recorded is the record losing it on
	// the way to the engine.
	if got, _ := r.planThisSitting["verdict"].(string); got != verdict {
		r.noteActErr(fmt.Sprintf("the plan relayed after a %q verdict carries verdict %q — the plan carries the verdict its sitting recorded", verdict, got))
	}
	return maps.Clone(r.planThisSitting)
}

func (r *runner) exec(args ...string) (string, error) { return r.execAs("", args...) }

// execAs is exec under a harness agent id, as the PreToolUse hook exports one into a dispatched
// seat's shell; "" is exec's own environment, which carries none.
func (r *runner) execAs(agent string, args ...string) (string, error) {
	cmd := exec.Command(r.bin, append(args, "--run", r.runDir)...)
	if agent != "" {
		cmd.Env = append(os.Environ(), seatenv.AgentVar+"="+agent)
	}
	out, err := cmd.CombinedOutput()
	r.progress.Store(time.Now().UnixNano())
	noteExec(args, err, out)
	if err != nil {
		err = fmt.Errorf("%s: %w\n  %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), err
}

// sit is the register a DISPATCH makes: every time debate.js hands a seat a prompt, that seat sits
// again, and with no hook bracket in this driver its register opens that sitting
// (plans/roundless.md §III.A.0 — the sitting ordinal and the chair's epoch rank those stored
// sittings). `register` below is once-per-run, and the seat id is stable: a chair that registered
// once would sit forty times at sitting #1, every singleton act after the first would be refused as
// a duplicate, the epoch would never advance, `dispatch next` would never see a lens sit against a
// moved head, and the loop would run until the ten-minute budget killed it.
// sit registers a seat for a sitting. occasion is the BENCH's — what this sitting was convened to
// do — and is variadic because only the bench has one; every other caller passes nothing.
func (r *runner) sit(role, seatID string, occasion ...string) {
	occ := ""
	if len(occasion) > 0 {
		occ = occasion[0]
	}
	r.mu.Lock()
	if done, seen := r.registering[seatID]; seen {
		select {
		case <-done:
		default:
			// A register of this seat is still in flight from another goroutine — a lens that
			// closeGap or minterOf registered mid-sitting. Wait for it outside the lock, then sit.
			r.mu.Unlock()
			<-done
			r.mu.Lock()
		}
	}
	if r.registering == nil {
		r.registering = map[string]chan struct{}{}
	}
	done := make(chan struct{})
	r.registering[seatID] = done
	r.registered[seatID] = true
	delete(r.judged, seatID)
	delete(r.closedNow, seatID)
	if strings.HasPrefix(seatID, "red-lens-") {
		if r.lensSittings == nil {
			r.lensSittings = map[string]int{}
		}
		r.lensSittings[seatID]++
	}
	if seatID == "blue-respond" {
		r.blueSittings++
	}
	if seatID == "red-chair" {
		r.chairRegisters++
	}
	r.mu.Unlock()
	switch {
	case r.coin(25):
		r.registerTwice(seatID, occ)
	case strings.HasPrefix(seatID, "red-lens-") && r.coin(33):
		r.lookThenRegister(seatID)
	default:
		r.admit(registerArgs(seatID, occ)...)
	}
	close(done)
}

// dispatchAgents numbers the agent ids registerTwice gives the dispatches it drives.
var dispatchAgents atomic.Int64

// registerTwice is a dispatch whose seat registers twice under one agent id — its first act
// registered it, then it ran `register` itself, as four m31 sittings did. The repeat is refused as a
// conflict, under the same occasion or another, and writes no second register.
func (r *runner) registerTwice(seatID, occ string) {
	agent := fmt.Sprintf("agent_fuzz_%d", dispatchAgents.Add(1))
	if _, ok := r.admitAs(agent, registerArgs(seatID, occ)...); !ok {
		return
	}
	again := occ
	if occ != "" && r.coin(50) {
		again = "petition"
		if occ == again {
			again = "docket"
		}
	}
	r.refuseAs(agent, feov.Conflict, registerArgs(seatID, again)...)
	if fam, ok := r.board(); ok {
		n := 0
		for _, e := range fam.Events {
			if reg, ok := recordpb.BodyAs[*recordpb.Register](e); ok && reg.GetAgentId() == agent {
				n++
			}
		}
		if n != 1 {
			r.noteActErr(fmt.Sprintf("%s registered under one agent id holds %d registers; a repeat writes none", seatID, n))
		}
	}
}

// lookedFirst counts the lens sittings lookThenRegister drove to the end, so the sweep can say
// when the drive stopped reaching it.
var lookedFirst atomic.Int64

// lookThenRegister is a lens dispatch whose first act is a look: under a fresh agent id it reads the
// board, which records nothing and so registers no one, and then runs `register` itself, which is
// admitted. A repeat is still refused. universe-m33's voice lens had
// its own register refused because the look had registered it.
func (r *runner) lookThenRegister(seatID string) {
	agent := fmt.Sprintf("agent_fuzz_%d", dispatchAgents.Add(1))
	if _, ok := r.admitAs(agent, "show", "board", "--seat-id", seatID); !ok {
		return
	}
	if _, ok := r.admitAs(agent, registerArgs(seatID, "")...); !ok {
		return
	}
	r.refuseAs(agent, feov.Conflict, registerArgs(seatID, "")...)
	lookedFirst.Add(1)
}

// register is the ONCE-PER-RUN register a helper makes for a seat it needs on the record before
// acting as it — blue-respond proving a computation from inside closeGap, the minter a lookup
// falls back to. A dispatched seat sits through sit(); this is for the seats a drive reaches for
// between dispatches, and it is idempotent so the reach never double-registers a seat mid-sitting.
func (r *runner) register(role, seatID string) {
	// A SEAT'S OWN REGISTER MUST LAND BEFORE ITS OWN APPENDS, and the flag alone did not say so.
	//
	// This set `registered[seatID] = true`, released the lock, and THEN ran the register. A second
	// goroutine reaching the same seat in that window saw the flag, returned immediately, and
	// appended — while the register was still in flight. `verify` then refused the run with
	// `register-before-append: a seat emitted an event before registering`, on about one seed in
	// forty, attributed to whatever drive happened to be second.
	//
	// The flag answered "has someone started registering this seat", and every caller read it as
	// "is this seat registered". Two questions, one bool.
	//
	// The waiter keeps the contention the comment below was protecting: seats still race each
	// other, three lanes still hit a run directory whose database may not exist. What is
	// serialized is one seat against ITSELF, which was never concurrency — it was a missing
	// happens-before.
	r.mu.Lock()
	if done, seen := r.registering[seatID]; seen {
		r.mu.Unlock()
		<-done // another goroutine is registering this seat; its register must land first
		return
	}
	done := make(chan struct{})
	if r.registering == nil {
		r.registering = map[string]chan struct{}{}
	}
	r.registering[seatID] = done
	r.registered[seatID] = true
	if seatID == "red-chair" {
		r.chairRegisters++
	}
	r.mu.Unlock()
	// OUTSIDE THE LOCK, deliberately: this is the invocation three concurrent lanes make against
	// a run directory whose database may not exist yet, and holding the lock across it would
	// serialize exactly the contention phase 3 exists to produce.
	//
	// No occasion: this path registers the concurrent LANES, which are blue and whose seat ids
	// already say what their sittings are for.
	r.admit(registerArgs(seatID, "")...)
	close(done)
}

// fuzzLensSeats are the lens seats this driver mints from — the default cast's lenses, every area,
// read off record.CastFor so the driver and setup cannot disagree about who sits. Each lens's
// budget is at least the run's mintBudget and grows with the report, so the ceiling on gaps open
// at once is at least seven times the floor; a lens holding its budget open sees `mint` refused
// with the budget text, which noteExec tallies, until one of its gaps closes.
var fuzzLensSeats = func() []string {
	var out []string
	seats, _ := record.CastFor(1)
	for _, s := range seats {
		if strings.HasPrefix(s, "red-lens-") {
			out = append(out, s)
		}
	}
	return out
}()

// proofOwner is the lens that owns every proof, and citationOwner the lens that owns every citation
// (anchorOwnerOf).
const (
	proofOwner    = "red-lens-computation"
	citationOwner = "red-lens-evidence"
)

// minterOf is the seat that may close or regrade gapID: the lens whose mint created it. Every gap
// this driver puts on the board goes through r.mint, which records the minter, so a miss here is
// a gap the driver did not mint — it falls back to the first lens and the record's originator
// refusal says so in the tally rather than the driver guessing silently.
func (r *runner) minterOf(gapID string) string {
	r.mu.Lock()
	m, ok := r.minter[gapID]
	r.mu.Unlock()
	if !ok {
		m = fuzzLensSeats[0]
	}
	r.register("lens", m)
	return m
}

// fuzzLensStock is the defects one lens finds in a run: past it, its sittings draw no fresh mint.
const fuzzLensStock = 2

// findsFresh reports whether a lens sitting now may put a fresh gap on the board: always in a drawn
// run, and only at the first chair sitting in a run forced to a verdict.
//
// A FORCED VERDICT IS REACHED INSIDE THE EPOCH LIMIT BY CONSTRUCTION, NOT BY THE DRAW. Both forced
// verdicts need a board nobody is ready for, and a fresh material gap readies its exchange, blue's
// edit moves the head, the moved head readies the lenses, and a lens sitting may mint again — each
// late mint buys the run another march to the docket. When those mints land is not a function of
// the seed: the seats of an epoch run on their own goroutines and draw from one shared stream in
// the order the scheduler runs them (lockedRand says why). MEASURED on one binary and one seed, 48 forced-UNVERIFIED
// runs: 1 to 7 gaps landed at the first sitting and the run took 2 to 12 chair sittings. The
// length follows the LAST mint: the seven whose mints all landed at the first sitting ended in 2
// to 5, and the four that ended CEILING at the limit of twelve last minted at their sixth or
// eighth. Load widens the spread; it does not cause it.
//
// PER-SEAT STREAMS WOULD NOT REMOVE IT. A stream per (seed, seat, sitting) fixes what each seat
// draws, not what its draw meets: which gaps openGaps lists,
// and whether a lens's stock has landed are the order the writes land in, and the seats contend
// on one run directory on purpose. A forced run needs a BOUND, and the bound is structural.
//
// Minted at the first sitting only — where every lens sits, never having sat — the forced run's
// gaps all start their march together, and the run ends one march after its first sitting on
// every interleaving. A drawn run mints at every sitting.
func (r *runner) findsFresh() bool {
	if !r.forceUnverified && !r.forceVerified {
		return true
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.chairRegisters <= 1
}

// landedMints is the mints of seatID that landed in this run.
func (r *runner) landedMints(seatID string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.lensMints[seatID]
}

// noteMint records that seatID's mint of gapID landed — the originator, and the count its stock
// of defects is read against.
func (r *runner) noteMint(seatID, gapID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.minter == nil {
		r.minter = map[string]string{}
		r.lensMints = map[string]int{}
	}
	r.minter[gapID] = seatID
	r.lensMints[seatID]++
}

// UNDERSCORES, because that is what the grade enum spells. This was `low-medium` and
// `medium-high`: two of five values refused at every site that generates a grade, which is
// `--proposed` on a grade dispute and both axes of a regrade. Measured — `merge regrade=8(7
// refused)`, `motion grade file=7(2 refused)`, and a scenario oracle three layers away reporting
// `G1 has no dispute event` for a contest whose drive was fine.

var grades = []string{"low", "low_medium", "medium", "medium_high", "high"}

func (r *runner) g() string { return grades[r.rng.Intn(len(grades))] }

// currentGrade reads the gap's grade at one axis, in the words a seat types. Empty when the board
// cannot be read or the axis is unknown — the caller treats that as "no constraint".
func (r *runner) currentGrade(gapID, dim string) string {
	b, err := record.FamilyOf(r.run())
	if err != nil {
		return ""
	}
	g := b.Gap(gapID)
	if g == nil {
		return ""
	}
	d, ok := record.GradeDimensionOf(dim)
	if !ok {
		return ""
	}
	cur, known := g.GradeAt(d)
	if !known {
		return ""
	}
	return recordpb.Word(cur)
}

// gradeOtherThan picks a grade that is not the one given, so a dispute always asks for a change.
func (r *runner) gradeOtherThan(cur string) string {
	for i := 0; i < len(grades)*4; i++ {
		if g := r.g(); g != cur {
			return g
		}
	}
	return grades[0] // unreachable while len(grades) > 1; a stated fallback beats a silent loop
}

// verifyOutcomes is the value space of `lens verify --as` — what a source DID for the claim.
//
// It was `--trust high|medium|low`, all three of which mean the source SUPPORTS the claim. The
// negative half is what this list exists to drive: `refutes` and `absent` are the outcomes that
// make the assembly screen fire, and a fuzz that only ever generated supporting verdicts would
// leave that whole path unexercised while the coverage gate read green.
var verifyOutcomes = []string{"supports", "supports_with_bridge", "weak", "refutes", "absent", "unreachable"}

// bridgeKinds is the value space of `--bridge`: what a supports_with_bridge reading bridged.
var bridgeKinds = []string{"summary", "secondary_citation", "near_restatement"}

// bridgeFor is the --bridge a supports_with_bridge drive names, drawn from n so the drive takes
// nothing from the run's RNG; every other outcome names none.
func bridgeFor(outcome string, n int) []string {
	if outcome != "supports_with_bridge" {
		return nil
	}
	return []string{"--bridge", bridgeKinds[n%len(bridgeKinds)]}
}

// noteBridge records what the write did with a supports_with_bridge reading: a kindless one must be
// refused for naming no bridge, and one naming a kind that the record took counts as named.
func (r *runner) noteBridge(seatID, outcome string, kindless bool, out string, err error) {
	if outcome != "supports_with_bridge" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	switch {
	case kindless && err == nil:
		if r.bridgeAdmittedErr == "" {
			r.bridgeAdmittedErr = seatID + " recorded supports_with_bridge with no --bridge: " + firstLine(out)
		}
	case kindless && strings.Contains(out, "requires --bridge"):
		r.bridgesRefused++
	case !kindless && err == nil:
		r.bridgesNamed++
	}
}

// verifyConfidence is the ORTHOGONAL axis: how sure red is of the outcome it just recorded.
// Driven independently of the outcome, because the pairs that matter most — `refutes` at low
// confidence, `supports` at low confidence — only exist if the two are drawn apart.
var verifyConfidence = []string{"high", "medium", "low"}

// someCitation returns a citation anchor on the record, or "" if blue has cited nothing yet —
// a REAL tool-assigned citation id, the same discipline someFinding uses. `lens verify --anchor`
// refuses an id that names no citation, so a fabricated one would drive only the reject path.
func (r *runner) someCitation() string {
	// A SEAT ID, because `show` only exists inside a seat's tree. Any seat reads the same
	// projection; the lens is the one that acts on citations.
	b, ok := r.board()
	if !ok {
		return ""
	}
	// A RETIRED CITATION IS NOT IN THE REPORT, and verify refuses it by design (it holds no sentence
	// to read the source against); the reading drive expects its verify admitted, so it skips those.
	var live []string
	for _, src := range record.EvidenceJSONOf(b.Events, b.At).Sources {
		if _, _, _, retired, err := record.AnchorRetiredAt(r.run(), src.Anchor); err == nil && !retired {
			live = append(live, src.Anchor)
		}
	}
	if len(live) == 0 {
		return ""
	}
	return live[r.rng.Intn(len(live))]
}

// citationToVerify is the citation seatID's reading drive verifies, or "" where it verifies none: the
// citation owner alone verifies (anchorOwnerOf), and every other lens corroborates a source instead.
// On a coin the owner takes a citation its work list offers; otherwise any source on the record,
// a corroboration's footnote among them, so a reading of a substituted copy stays driven.
func (r *runner) citationToVerify(seatID string) string {
	if seatID != citationOwner {
		return ""
	}
	if offered, err := anchorItemsOn(r.run(), seatID, "citation"); err == nil && len(offered) > 0 && r.coin(50) {
		return pick(r.rng, offered)
	}
	return r.someCitation()
}

// answerContradictions raises a finding, by its id, for every contradiction this lens owes: the
// rows of show evidence's unanswered_contradictions whose reader it is.
func (r *runner) answerContradictions(seatID string) {
	b, ok := r.board()
	if !ok {
		return
	}
	for _, c := range record.EvidenceJSONOf(b.Events, b.At).UnansweredContradictions {
		if c.Reader != seatID {
			continue
		}
		if c.RaisedBy != "" {
			r.mintFrom(seatID, c.RaisedBy)
			continue
		}
		r.raiseAndMint(seatID, "--about-kind", "contradiction", "--about", c.ID)
	}
}

// raisePrinted raises the contradiction a reading printed, by that id, then answers every other
// contradiction this lens owes.
func (r *runner) raisePrinted(seatID, printed string) {
	if printed != "" {
		r.raiseAndMint(seatID, "--about-kind", "contradiction", "--about", printed)
	}
	r.answerContradictions(seatID)
}

// ownUnansweredRaise is whether finding is seatID's own and raised a contradiction `show evidence`
// still lists as unanswered: the mint naming it in found_by answers it outside seatID's budget.
func (r *runner) ownUnansweredRaise(seatID, finding string) bool {
	b, ok := r.board()
	if !ok {
		return false
	}
	raised := false
	for _, c := range record.EvidenceJSONOf(b.Events, b.At).UnansweredContradictions {
		raised = raised || c.RaisedBy == finding
	}
	if !raised {
		return false
	}
	for _, e := range b.Events {
		if f, ok := recordpb.BodyAs[*recordpb.Finding](e); ok && f.GetId() == finding {
			return e.GetSeatId() == seatID
		}
	}
	return false
}

// raiseAndMint raises a contradiction with a finding anchored by anchorArgs and answers it with the
// gap minted from that finding: a finding alone answers nothing, because a source the report
// misstates is material by its class. A mint the budget refuses leaves the contradiction raised and
// held, for the lens's next sitting (answerContradictions).
func (r *runner) raiseAndMint(seatID string, anchorArgs ...string) {
	args := append([]string{"--json", "finding", "--seat-id", seatID,
		"--severity", r.g(), "--likelihood", r.g(), "--impact", r.g(), "--reason", "fuzz: the source says otherwise"}, anchorArgs...)
	out, err := r.exec(args...)
	// THE LENSES SIT CONCURRENTLY, and any lens's finding naming a contradiction raises it, so
	// another lens can raise one between this lens's board read and its write.
	// A refusal of a raise BY ID is admissible only when the record now holds that contradiction
	// raised or answered; every other refusal fails the run.
	if id := contradictionAbout(anchorArgs); id != "" {
		if err != nil && r.contradictionUnraised(id) {
			r.noteActErr("refused a raise of a contradiction the record holds unraised: " + err.Error())
		}
		noteEither([]string{"finding"}, "raising a contradiction a concurrent lens may answer first", err == nil)
	} else if err != nil {
		r.noteActErr("refused, and the drive expects it admitted: " + err.Error())
	}
	if err != nil {
		return
	}
	var env struct {
		Result struct {
			FindingID string `json:"finding_id"`
		} `json:"result"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.FindingID == "" {
		r.noteActErr("an admitted finding's --json result names no finding_id: " + firstLine(out))
		return
	}
	r.mintFrom(seatID, env.Result.FindingID)
}

// contradictionAbout is the contradiction id a finding's anchor arguments name, "" for none.
func contradictionAbout(anchorArgs []string) string {
	kind, ref := "", ""
	for i := 0; i+1 < len(anchorArgs); i++ {
		switch anchorArgs[i] {
		case "--about-kind":
			kind = anchorArgs[i+1]
		case "--about":
			ref = anchorArgs[i+1]
		}
	}
	if kind != "contradiction" {
		return ""
	}
	return ref
}

// quoteRaisesNothing files a finding quoting claim and fails the run if any contradiction reads
// raised by it: only a finding naming a contradiction's id raises it.
func (r *runner) quoteRaisesNothing(seatID, claim string) {
	out, err := r.exec("--json", "finding", "--seat-id", seatID, "--quote", claim,
		"--severity", r.g(), "--likelihood", r.g(), "--impact", r.g(), "--reason", "fuzz: the source says otherwise")
	if err != nil {
		r.noteActErr("refused, and the drive expects it admitted: " + err.Error())
		return
	}
	var env struct {
		Result struct {
			FindingID string `json:"finding_id"`
		} `json:"result"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.FindingID == "" {
		r.noteActErr("an admitted finding's --json result names no finding_id: " + firstLine(out))
		return
	}
	b, ok := r.board()
	if !ok {
		return
	}
	for _, c := range record.EvidenceJSONOf(b.Events, b.At).UnansweredContradictions {
		if c.RaisedBy == env.Result.FindingID {
			r.noteActErr("finding " + env.Result.FindingID + " quoting the claim raised contradiction " + c.ID + ", which it does not name")
		}
	}
}

// contradictionUnraised reports whether the record lists contradiction id as unanswered with no
// finding raising it.
func (r *runner) contradictionUnraised(id string) bool {
	b, ok := r.board()
	if !ok {
		return true
	}
	for _, c := range record.EvidenceJSONOf(b.Events, b.At).UnansweredContradictions {
		if c.ID == id {
			return c.RaisedBy == ""
		}
	}
	return false
}

// forcedProofs is the anchors of the proofs on the record whose script is reopenedProofScript.
func (r *runner) forcedProofs() []string { return r.proofsOfScript(reopenedProofScript) }

// proofsOfScript is the anchors of the proofs on the record whose script is script, read by its
// sha256 on the record (EvidenceProofJSON.Sha256), never recovered from report text.
func (r *runner) proofsOfScript(script string) []string {
	b, ok := r.board()
	if !ok {
		return nil
	}
	sum := sha256.Sum256([]byte(script))
	sha := hex.EncodeToString(sum[:])
	var out []string
	for _, p := range record.EvidenceJSONOf(b.Events, b.At).Proofs {
		if p.Sha256 == sha {
			out = append(out, p.Anchor)
		}
	}
	return out
}

// reopenReadied is whether a plan of this run has readied the proof owner for a reopened proof,
// which ends the forced chain.
func (r *runner) reopenReadied() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.proofReopenedReadied[proofOwner] > 0
}

func (r *runner) reachedStage(s string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.reopenStage = s
}

// proveToReopen is blue's half of the forced chain: it proves the sentence the edit drive rewrites
// while no forced proof stands, and rewrites that sentence once the owner has judged one — a
// forced proof its owner is no longer offered and that is not already reopened.
func (r *runner) proveToReopen(seatID string) {
	if !r.forceReopenedProof || r.reopenReadied() {
		return
	}
	anchors := r.forcedProofs()
	if len(anchors) == 0 {
		name := "fuzz-reopened-proof.js"
		if err := os.WriteFile(filepath.Join(r.runDir, name), []byte(reopenedProofScript), 0o644); err != nil {
			r.reachedStage("could not write the script: " + err.Error())
			return
		}
		if _, err := r.exec("prove", "--seat-id", seatID, "--quote", "across the whole sampled corpus",
			"--script", name, "--reason", "fuzz: the computation a later edit reopens"); err != nil {
			r.reachedStage("prove refused: " + firstLine(err.Error()))
			return
		}
		r.reachedStage("proved")
		return
	}
	offered, err := anchorItemsOn(r.run(), proofOwner, "proof")
	if err != nil {
		return
	}
	reopened := map[string]bool{}
	for _, it := range r.unchecked() {
		reopened[it.Anchor] = true
	}
	for _, a := range anchors {
		if slices.Contains(offered, a) || reopened[a] {
			continue
		}
		cur, err := reportproj.RenderFromRecord(r.run())
		if err != nil {
			return
		}
		oldSpan, newSpan := "rising over time", "climbing sharply"
		if !strings.Contains(cur, oldSpan) {
			oldSpan, newSpan = newSpan, oldSpan
		}
		if !strings.Contains(cur, oldSpan) {
			r.reachedStage("judged, and neither phrasing of the edit target is on the page")
			return
		}
		run := r.edgeRun(cur, oldSpan)
		if _, err := r.do("edit", seatID).set("--quote", oldSpan+run).set("--new", newSpan+run).
			set("--reason", "fuzz: rewording the sentence a judged proof backs").result(); err != nil {
			r.reachedStage("edit refused: " + firstLine(err.Error()))
			return
		}
		r.reachedStage("judged and edited")
		return
	}
}

// forcedChainProof is whether anchor is the proof a forced chain judges (judgeToReopen, judgeUnsound).
func (r *runner) forcedChainProof(anchor string) bool {
	return r.forceReopenedProof && slices.Contains(r.forcedProofs(), anchor) ||
		r.forceUnsoundProof && slices.Contains(r.proofsOfScript(unsoundProofScript), anchor)
}

// judgeToReopen is the owner's half of the forced chain: it judges every forced proof its work list
// offers.
func (r *runner) judgeToReopen(seatID string) {
	if !r.forceReopenedProof || seatID != proofOwner || r.reopenReadied() {
		return
	}
	offered, err := anchorItemsOn(r.run(), seatID, "proof")
	if err != nil {
		return
	}
	for _, a := range r.forcedProofs() {
		if !slices.Contains(offered, a) {
			continue
		}
		if !r.firstThisSitting(seatID, a) {
			continue
		}
		if _, err := r.do("reproduce", seatID).set("--anchor", a).set("--as", "sound").
			set("--reason", "fuzz: read the script against the sentence it backs").result(); err != nil {
			r.reachedStage("reproduce refused: " + firstLine(err.Error()))
			return
		}
		r.reachedStage("judged")
	}
}

// proveOwed is blue-respond's act in owedProofSeed's run: a fresh proof of the seeded sentence, which
// the computation lens owes a check of until it sits. It draws nothing.
func (r *runner) proveOwed(seatID string) {
	if !r.forceOwedProof || seatID != "blue-respond" {
		return
	}
	r.mu.Lock()
	r.owedProofs++
	n := r.owedProofs
	r.mu.Unlock()
	name := fmt.Sprintf("fuzz-owed-proof-%d.js", n)
	if err := r.writeScript(name, fmt.Sprintf("console.log('fuzz: owed proof %d');", n)); err == nil {
		r.admit("prove", "--seat-id", seatID, "--quote", "§ fuzz sentence", "--script", name,
			"--reason", "fuzz: a proof its owner has not yet checked")
	}
}

// unsoundFed is whether a plan of this run has readied the judge or blue for every unsound proof it
// held, which ends the forced unsound chain.
func (r *runner) unsoundFed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.unsoundReadied > 0
}

func (r *runner) reachedUnsoundStage(s string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.unsoundStage = s
}

// proveToJudgeUnsound is blue's half of the forced unsound chain: while no proof of
// unsoundProofScript stands, it proves the seeded sentence no edit rewrites, so the judgement the
// owner records stands until blue's fix (fixUnsound).
func (r *runner) proveToJudgeUnsound(seatID string) {
	if !r.forceUnsoundProof || r.unsoundFed() || len(r.proofsOfScript(unsoundProofScript)) > 0 {
		return
	}
	name := "fuzz-unsound-proof.js"
	if err := os.WriteFile(filepath.Join(r.runDir, name), []byte(unsoundProofScript), 0o644); err != nil {
		r.reachedUnsoundStage("could not write the script: " + err.Error())
		return
	}
	if _, err := r.exec("prove", "--seat-id", seatID, "--quote", "§ fuzz sentence",
		"--script", name, "--reason", "fuzz: the computation its owner judges unsound"); err != nil {
		r.reachedUnsoundStage("prove refused: " + firstLine(err.Error()))
		return
	}
	r.reachedUnsoundStage("proved")
}

// judgeUnsound is the owner's half of the forced unsound chain: in a sitting whose plan does not ready
// blue-respond, it judges every forced proof its work list offers unsound. answerUnsound, after the
// sitting's last judgement, mints the gap carrying it, and the next plan holds the blocker and
// readies the judge or blue (tallyUnsound).
//
// ONLY WHERE BLUE IS NOT READIED, because blue's sitting fixes every unsound proof before its envelope
// (fixUnsound): a judgement in an epoch blue-respond sits in is fixed before the next plan is read,
// and no plan holds it. The first plan readies the lenses and not blue, and blue-synthesize's
// sitting has proved the forced proof before it.
func (r *runner) judgeUnsound(seatID string) {
	if !r.forceUnsoundProof || seatID != proofOwner || r.unsoundFed() {
		return
	}
	forced := r.proofsOfScript(unsoundProofScript)
	if len(forced) == 0 {
		return
	}
	if r.planReadies("blue-respond") {
		r.reachedUnsoundStage("proved, and every owner sitting since was in a plan that readies blue-respond")
		return
	}
	offered, err := anchorItemsOn(r.run(), seatID, "proof")
	if err != nil {
		return
	}
	for _, a := range forced {
		if !slices.Contains(offered, a) || !r.firstThisSitting(seatID, a) {
			continue
		}
		if _, err := r.do("reproduce", seatID).set("--anchor", a).set("--as", "unsound").
			set("--reason", "fuzz: the script does not establish the sentence it backs").result(); err != nil {
			r.reachedUnsoundStage("reproduce refused: " + firstLine(err.Error()))
			return
		}
		r.reachedUnsoundStage("judged unsound")
	}
}

// citeAbsentSpan is the forced drive of absentSpanSeed's run: blue's first sitting cites words its
// source does not hold, and the tool must refuse it (citeSpanRefused). It draws nothing from the
// run's RNG and leaves the bogus-span rotation where it was, so the drawn cites rotate as before.
func (r *runner) citeAbsentSpan(seatID string) {
	r.mu.Lock()
	due := r.forceAbsentSpan && !r.absentSpanDriven
	r.absentSpanDriven = true
	r.mu.Unlock()
	if !due {
		return
	}
	out, err := r.do("cite", seatID).set("--quote", "§ fuzz sentence").set("--url", sourceURL("/blue-shared")).
		set("--title", sourceTitle).set("--source-span", "fuzz words the source does not hold").result()
	r.citeSpanRefused(seatID, out, err)
}

// anchorRef is a citation or proof the record holds unchecked: its anchor, and a proof's script sha
// ("" for a citation).
type anchorRef struct{ Anchor, Sha string }

// unchecked is every citation and proof the record holds unchecked — nobody checked it, or an edit
// changed its sentence since — off `show evidence`'s `unchecked` rows.
func (r *runner) unchecked() []anchorRef {
	b, ok := r.board()
	if !ok {
		return nil
	}
	ev := record.EvidenceJSONOf(b.Events, b.At)
	sha := map[string]string{}
	for _, p := range ev.Proofs {
		sha[p.Anchor] = p.Sha256
	}
	var out []anchorRef
	for _, it := range ev.Unchecked {
		out = append(out, anchorRef{Anchor: it.Anchor, Sha: sha[it.Anchor]})
	}
	return out
}

// readiedForReopened reports whether plan holds a reopened blocker, and whether it readies the
// owner of every one it holds. Read off the plan's own fields.
func readiedForReopened(plan map[string]any) (held, readied bool) {
	party := map[any]bool{}
	parties, _ := plan["parties"].([]any)
	for _, p := range parties {
		m, _ := p.(map[string]any)
		party[m["seat_id"]] = true
	}
	readied = true
	blockers, _ := plan["blockers"].([]any)
	for _, b := range blockers {
		m, _ := b.(map[string]any)
		if record.BlockerKind(fmt.Sprint(m["kind"])) == record.BlockerUnchecked {
			held = true
			readied = readied && party[m["owner"]]
		}
	}
	return held, held && readied
}

// tallyUnsound counts the plan's failed-check blockers and whether it readied their lens or blue for
// each (record dispatch Source 4): blue is readied for every one, unless the plan says blue has sat
// since its cause, or the plan is at the epoch limit, which dispatches nobody. Called with r.mu held.
func (r *runner) tallyUnsound(plan map[string]any, party map[any]bool) {
	whys, _ := plan["why"].([]any)
	satSince := func(seat, anchor string) bool {
		for _, w := range whys {
			if s := fmt.Sprint(w); strings.HasPrefix(s, seat+" has sat since") && strings.Contains(s, " "+anchor+"'s check failed") {
				return true
			}
		}
		return false
	}
	limit, _ := plan["epoch_limit_reached"].(bool)
	held, readied := false, true
	bs, _ := plan["blockers"].([]any)
	for _, b := range bs {
		m, _ := b.(map[string]any)
		if record.BlockerKind(fmt.Sprint(m["kind"])) != record.BlockerFailedCheck {
			continue
		}
		held = true
		owner, anchor := fmt.Sprint(m["owner"]), fmt.Sprint(m["subject"])
		ok := party[owner] || party["blue-respond"]
		readied = readied && ok
		if !ok && !limit && !satSince("blue-respond", anchor) && r.unsoundReadyErr == "" {
			r.unsoundReadyErr = fmt.Sprintf("a plan holds %s on a failed check by %s and readies neither it nor blue, with no reason saying either sat since: why %v", anchor, owner, whys)
		}
	}
	if held {
		r.unsoundPlans++
		if readied {
			r.unsoundReadied++
		}
	}
}

// judgement is a lens's verdict on a proof's sentence: drawn, except in a run forced to a verdict,
// whose drive is its verdict — a standing unsound holds the PASS (record.BlockerFailedCheck) and
// readies its judge and blue, which no forced run's march to its verdict accounts for.
func (r *runner) judgement() string {
	if r.forceVerified || r.forceUnverified {
		return "sound"
	}
	return pick(r.rng, []string{"sound", "unsound"})
}

// answerFailed is the half of a failed check that is the lens's whose check failed: for every proof
// it judged unsound, and every citation whose source its check could not read, that no gap of its
// carries, it mints its gap about the anchor — outside its budget, so a refusal for budget fails the
// run — and, on a coin, checks again an anchor whose gap it minted in an earlier sitting: a proof's
// judgement, which is what clears the blocker where blue has not retired or reopened it, or a
// citation read unreachable again. A drawn `unsound`, or the `unreachable`, while that gap is open,
// is driven to the record's refusal and tallied.
//
// THE GAPS ARE MINTED AFTER THE RE-JUDGEMENTS, from the blockers as they then stand: a proof judged
// unsound again in this sitting is owed a gap from this sitting on (the earlier one carried the
// earlier judgement), and the drive's last act on it is the one its envelope is checked against. The
// lens drive calls this again after its last judging act for the same reason.
func (r *runner) answerFailed(seatID string) {
	bs, err := record.PassBlockers(r.run())
	if err != nil {
		r.noteActErr("the blockers could not be read to say which unsound judgements " + seatID + " owes a gap: " + err.Error())
		return
	}
	open := map[string]bool{}
	for _, g := range r.openGaps() {
		open[g] = true
	}
	for _, b := range bs {
		if b.Kind == record.BlockerFailedCheck && b.Detail == "" && b.Owner == seatID && b.AnsweredBy() != "" && open[b.AnsweredBy()] &&
			r.coin(50) && r.firstThisSitting(seatID, b.Subject) {
			// A READING THAT READS NOTHING AGAIN, while the gap carrying the first is open, is refused.
			r.do("verify", seatID).set("--anchor", b.Subject).set("--as", "unreachable").set("--confidence", pick(r.rng, verifyConfidence)).
				set("--reason", "fuzz: the source is still unreadable").refuse(feov.Conflict)
			r.mu.Lock()
			r.unreadRepeats++
			r.mu.Unlock()
			continue
		}
		if b.Kind == record.BlockerFailedCheck && b.Detail != "" && b.Owner == seatID && b.AnsweredBy() != "" &&
			r.coin(50) && r.firstThisSitting(seatID, b.Subject) {
			as := r.judgement()
			a := r.do("reproduce", seatID).set("--anchor", b.Subject).set("--as", as).
				set("--reason", "fuzz: judged the sentence its gap carried again")
			// AN UNSOUND JUDGEMENT OF THE SAME SENTENCE, WHILE ITS GAP IS OPEN, IS REFUSED: the gap
			// already carries it, and no second gap is owed.
			if as == "unsound" && open[b.AnsweredBy()] {
				a.refuse(feov.Conflict)
				r.mu.Lock()
				r.unsoundRepeats++
				r.mu.Unlock()
				continue
			}
			a.admit()
		}
	}
	if bs, err = record.PassBlockers(r.run()); err != nil {
		r.noteActErr("the blockers could not be read to say which unsound judgements " + seatID + " owes a gap: " + err.Error())
		return
	}
	for _, b := range bs {
		if b.Kind != record.BlockerFailedCheck || b.Owner != seatID || b.AnsweredBy() != "" {
			continue
		}
		citation := b.Detail == ""
		// ONE TIME IN TEN, THE LENS TRIES TO END WITH ITS GAP UNMINTED, and the record refuses its
		// envelope naming the failed check.
		if r.coin(10) {
			if citation {
				r.envelopeRefusedEarly(seatID, "check of citation "+b.Subject)
			} else {
				r.envelopeRefusedEarly(seatID, "judged the sentence proof "+b.Subject+" backs unsound")
			}
		}
		problem, check := "fuzz: the script behind "+b.Subject+" does not establish the sentence it backs", "the sentence says what the script shows"
		if citation {
			problem, check = "fuzz: no lens could read the source "+b.Subject+" cites, and the report still cites it", "the citation is retired or cites a source a lens can read"
		}
		args := []string{"--json", "mint", "--seat-id", seatID,
			"--problem", problem,
			"--check-kind", "document", "--check", check, "--fix", pick(r.rng, []string{dirApply, dirCounter}),
			"--severity", r.g(), "--likelihood", r.g(), "--impact", r.g(), "--class", pick(r.rng, mintPoolClasses),
			"--about-kind", "anchor", "--about", b.Subject, "--reason", "fuzz: carrying a failed check to blue"}
		overOpen := r.ownOpenGapAbout(seatID, b.Subject)
		out, err := r.exec(args...)
		if err != nil {
			if strings.Contains(out, "its mint budget admits no more") {
				r.noteActErr(seatID + "'s gap about " + b.Subject + ", whose check of its failed, was refused for budget: " + firstLine(out))
			}
			continue
		}
		var env struct {
			Result struct {
				GapID string `json:"gap_id"`
			} `json:"result"`
		}
		if json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil || env.Result.GapID == "" {
			r.noteActErr("an admitted mint's --json result names no gap_id: " + firstLine(out))
			continue
		}
		// THE MINTER IS RECORDED, so the drive's closes and regrades of the gap go to its originator;
		// the lens's stock of fresh defects is not spent — this gap carries a judgement, not a find.
		r.mu.Lock()
		if r.minter == nil {
			r.minter = map[string]string{}
			r.lensMints = map[string]int{}
		}
		r.minter[env.Result.GapID] = seatID
		if citation {
			r.unreadAnswers++
		} else {
			r.unsoundAnswers++
		}
		if overOpen {
			r.unsoundOverOpen++
		}
		r.mu.Unlock()
	}
}

// ownOpenGapAbout reports whether the board holds an open gap seatID minted about proof anchor
// since the sentence last changed (an edit reopening the anchor) and since seatID last judged it
// sound: a gap carrying an unsound judgement of the sentence as it stands. A gap about the sentence
// before an edit, or before the judge's own sound, carried a judgement since answered.
func (r *runner) ownOpenGapAbout(seatID, anchor string) bool {
	fam, err := record.FamilyOf(r.runHandle)
	if err != nil {
		return false
	}
	open := map[string]bool{}
	for _, g := range r.openGaps() {
		open[g] = true
	}
	ids := fam.At.IDs(fam.Events)
	var since int64
	minted := map[string]int64{}
	for i, e := range fam.Events {
		body, _ := recordpb.Body(e)
		switch b := body.(type) {
		case *recordpb.BlueEdit:
			if slices.Contains(b.GetReopened(), anchor) {
				since = ids[i]
			}
		case *recordpb.Reproduce:
			if e.GetSeatId() == seatID && b.GetProofId() == anchor && b.GetSoundness() == recordpb.Soundness_SOUNDNESS_SOUND {
				since = ids[i]
			}
		case *recordpb.Mint:
			if e.GetSeatId() == seatID && b.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_ANCHOR && b.GetAboutRef() == anchor {
				minted[b.GetGapId()] = ids[i]
			}
		}
	}
	for g, at := range minted {
		if open[g] && at > since {
			return true
		}
	}
	return false
}

// keptByDebt reads one reason of a plan's lens fold: a lens readied with retiresAfter or more barren
// sittings because it owes, and what it owes. The reason's wording is the fold's
// (record.lensStates); a fold that stops saying it leaves owedKept at zero, which the sweep gates.
var keptByDebt = regexp.MustCompile(`^red-lens-[a-z-]+: (?:the whole report was last served to it at its pin \d+|no sitting it sat through was served the whole report), behind head \d+ — audits the report \((\d+) sitting\(s\) in a row with no fresh material mint; it does not retire while it owes (.*)\)$`)

// tallyOwed counts what this chair sitting's plan readies its lenses for: the owner of each
// reopened blocker, by owner for a reopened proof, and each lens kept from retiring by a debt. A
// reopened proof's owner is the run's proof owner (anchorOwnerOf), or the run fails.
func (r *runner) tallyOwed(plan map[string]any) {
	held, readied := readiedForReopened(plan)
	proofs := map[string]bool{}
	if b, ok := r.board(); ok {
		for _, p := range record.EvidenceJSONOf(b.Events, b.At).Proofs {
			proofs[p.Anchor] = true
		}
	}
	party := map[any]bool{}
	parties, _ := plan["parties"].([]any)
	for _, p := range parties {
		m, _ := p.(map[string]any)
		party[m["seat_id"]] = true
	}
	// THE BENCH WAITS: a closing owed on a gap, and no bench convened on it this plan.
	benched := map[string]bool{}
	for _, p := range parties {
		m, _ := p.(map[string]any)
		if m["seat_id"] == "judge" {
			gs, _ := m["gap_ids"].([]any)
			for _, g := range gs {
				benched[fmt.Sprint(g)] = true
			}
		}
	}
	waits := 0
	if bs, _ := plan["blockers"].([]any); len(bs) > 0 {
		for _, b := range bs {
			m, _ := b.(map[string]any)
			if record.BlockerKind(fmt.Sprint(m["kind"])) == record.BlockerClosingOwed && !benched[fmt.Sprint(m["subject"])] {
				waits = 1
			}
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closingWaits += waits
	r.tallyUnsound(plan, party)
	if held {
		r.reopenedPlans++
		if readied {
			r.reopenedReadied++
		}
	}
	blockers, _ := plan["blockers"].([]any)
	for _, b := range blockers {
		m, _ := b.(map[string]any)
		if record.BlockerKind(fmt.Sprint(m["kind"])) != record.BlockerUnchecked || !proofs[fmt.Sprint(m["subject"])] {
			continue
		}
		owner := fmt.Sprint(m["owner"])
		if owner != proofOwner && r.reopenedOwnerErr == "" {
			r.reopenedOwnerErr = fmt.Sprintf("unchecked proof %v is owned by %s, and the proofs are %s's", m["subject"], owner, proofOwner)
		}
		if party[owner] {
			if r.proofReopenedReadied == nil {
				r.proofReopenedReadied = map[string]int{}
			}
			r.proofReopenedReadied[owner]++
		}
	}
	whys, _ := plan["why"].([]any)
	for _, w := range whys {
		g := keptByDebt.FindStringSubmatch(fmt.Sprint(w))
		if g == nil {
			continue
		}
		if n, _ := strconv.Atoi(g[1]); n >= 3 {
			r.owedKept++
			if strings.Contains(g[2], string(record.BlockerUnchecked)+" P-") {
				r.proofOwedKept++
			}
			if r.owedKeptBy == nil {
				r.owedKeptBy = map[string]int{}
			}
			for k := range owedKinds(g[2]) {
				r.owedKeptBy[k]++
			}
		}
	}
}

// owedKinds is the kinds of debt a fold's "it owes" clause names, each once: the fold names each item
// by its kind's words (record.lensOwesOf). An item no kind matches is "unnamed", so a reworded debt
// shows up in the tally rather than vanishing from it. A blocker the lens owns is named by its kind
// and subject (record.owedBlocker), so its rows are every kind of the blocker table
// (record.BlockerKinds): a new kind is named without a row here.
func owedKinds(clause string) map[string]bool {
	words := []struct{ word, kind string }{
		{"blue's answer to ", "blue's answer"}, {"the accepted grade motion on ", "grade motion"},
		{"the bench's remand of ", "remand"}, {"'s finding ", "finding"},
	}
	out := map[string]bool{}
	for _, item := range strings.Split(clause, ", ") {
		kind := "unnamed"
		for _, k := range words {
			if strings.Contains(item, k.word) {
				kind = k.kind
				break
			}
		}
		for _, k := range record.BlockerKinds() {
			if kind == "unnamed" && strings.HasPrefix(item, string(k)+" ") {
				kind = string(k)
			}
		}
		out[kind] = true
	}
	return out
}

// anchorPrefix is the id prefix of each anchor kind a work list offers for its first check.
var anchorPrefix = map[string]string{"proof": "P-", "citation": "C-"}

// anchorItemsOn is the anchors of one kind ("proof" or "citation") seatID's work list offers for
// their first check ("proof P-…", "citation C-…"). A reopened anchor is not among them: it is its
// owner's reopened blocker (anchorOwnerOf).
func anchorItemsOn(run record.Run, seatID, kind string) ([]string, error) {
	w, err := record.WorkOfSeat(run, "lens", seatID)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, it := range w.Sitting.Open {
		what, offered := strings.CutPrefix(it.What, kind+" ")
		if !offered || !strings.HasPrefix(what, anchorPrefix[kind]) {
			continue
		}
		if i := strings.IndexAny(what, "' "); i > 0 {
			out = append(out, what[:i])
		}
	}
	return out, nil
}

// ownerOf is the lens that owns this kind of anchor.
func (r *runner) ownerOf(kind string) string {
	if kind == "proof" {
		return proofOwner
	}
	return citationOwner
}

// checkAnchorOffers records, once, a lens sitting whose work list offers a citation or proof it does
// not own, and counts each owner sitting offered one for its first check.
//
// THE OWNER'S OFFER IS COUNTED AT ITS SITTING, NOT ONLY AT THE RUN'S END. An unchecked anchor's item
// blocks its owner's envelope, so the owner checks every one it is offered before its sitting ends,
// and a run ends with one offered only when the run stops inside that sitting: read at the end alone,
// the routing would almost never be exercised.
func (r *runner) checkAnchorOffers(seatID string) {
	for _, kind := range []string{"proof", "citation"} {
		owner := r.ownerOf(kind)
		got, err := anchorItemsOn(r.run(), seatID, kind)
		if err != nil || len(got) == 0 {
			continue
		}
		if seatID == owner {
			r.mu.Lock()
			if r.offeredToOwner == nil {
				r.offeredToOwner = map[string]int{}
			}
			r.offeredToOwner[kind]++
			r.mu.Unlock()
			continue
		}
		r.mu.Lock()
		if r.anchorOfferErr == "" {
			r.anchorOfferErr = fmt.Sprintf("anchor owner: %s's work list offers %s %q, and they are %q's", seatID, kind, got, owner)
		}
		r.mu.Unlock()
	}
}

// anchorOffersAgree is "" when, at the run's end, the record's own answer and the work lists agree:
// each kind's owner's list offers for a first check exactly the anchors of that kind the record holds
// unchecked with no edit since they were placed (`show evidence`'s `unchecked` rows with edit 0), and
// every other cast lens's list offers none; and every citation on a failed check is on blue's list,
// with the remedy it owes. Otherwise what disagreed. proofs and citations are how many the owners
// were offered.
func (r *runner) anchorOffersAgree(run record.Run) (msg string, proofs, citations int) {
	fam, err := record.FamilyOf(run)
	if err != nil {
		return "anchor owner: the record does not read: " + err.Error(), 0, 0
	}
	ev := record.EvidenceJSONOf(fam.Events, fam.At)
	isProof := map[string]bool{}
	for _, p := range ev.Proofs {
		isProof[p.Anchor] = true
	}
	want := map[string][]string{}
	for _, u := range ev.Unchecked {
		if u.Edit != 0 {
			continue
		}
		kind := "citation"
		if isProof[u.Anchor] {
			kind = "proof"
		}
		want[kind] = append(want[kind], u.Anchor)
	}
	cast, err := record.CastOf(run)
	if err != nil {
		return "anchor owner: the cast does not read: " + err.Error(), 0, 0
	}
	for _, kind := range []string{"proof", "citation"} {
		owner := r.ownerOf(kind)
		for _, seat := range cast {
			if !strings.HasPrefix(seat, "red-lens-") {
				continue
			}
			got, err := anchorItemsOn(run, seat, kind)
			if err != nil {
				return "anchor owner: " + seat + "'s work list does not read: " + err.Error(), 0, 0
			}
			if seat != owner {
				if len(got) > 0 {
					return fmt.Sprintf("anchor owner: %s's work list offers %s %q at the run's end, and they are %q's", seat, kind, got, owner), 0, 0
				}
				continue
			}
			if !slices.Equal(got, want[kind]) {
				return fmt.Sprintf("anchor owner: %s's work list offers %s %q at the run's end; the record holds %q unchecked since it was placed", seat, kind, got, want[kind]), 0, 0
			}
		}
	}
	bs, err := record.PassBlockers(run)
	if err != nil {
		return "anchor owner: the blockers do not read: " + err.Error(), 0, 0
	}
	blue, err := record.WorkOfSeat(run, "blue", "blue-respond")
	if err != nil {
		return "anchor owner: blue's work list does not read: " + err.Error(), 0, 0
	}
	for _, b := range bs {
		if b.Kind != record.BlockerFailedCheck || b.Detail != "" {
			continue
		}
		if !slices.ContainsFunc(blue.Sitting.Open, func(it record.Item) bool {
			return strings.Contains(it.What, "check of citation "+b.Subject+" ") && strings.Contains(it.What, "Retire the citation")
		}) {
			return fmt.Sprintf("anchor owner: citation %s is on a failed check at the run's end and blue's work list does not name its remedy", b.Subject), 0, 0
		}
	}
	return "", len(want["proof"]), len(want["citation"])
}

// mint records a gap as LENS seatID and returns the tool-assigned id. The first mint of a
// run introduces the class; the rest reuse it. seatID is the lens that is sitting: the chair
// has no mint verb (roundless §III.B.3), and the record remembers the minter as the one seat that
// may later close or regrade the gap, which is why every landed mint is noted on r.minter.
func (r *runner) mint(seatID string) string { return r.mintFrom(seatID, "") }

// mintFrom mints as mint does, naming foundBy in --found-by when it is set: the gap that answers
// a contradiction the finding foundBy raised.
func (r *runner) mintFrom(seatID, foundBy string) string {
	// #277: the KIND is randomised, not pinned. Pinning it to `document` would leave the
	// computation branch — a gap that CANNOT be closed without a proof answering it — driven by
	// nothing, which is the dead-drive class #276 measured (183/183 refusals on a verb that did
	// not exist, behind a green sweep). Both the refusal and the satisfied path run across a
	// sweep, because the prove drive below answers a computation gap when one is open.
	// THE SCENARIO IS DRAWN HERE, ONCE, and travels in required_fix.
	//
	// The fake used to decide everything downstream by coin flip, and nothing read
	// acceptance_check or required_fix — `grep -c` returned 0 — so the one question that
	// matters, DID THE RESPONSE SATISFY WHAT WAS ASKED, was never posed, because nothing knew
	// what had been asked. A directive fixes that without modelling a seat's judgement: red
	// draws the outcome at mint and every later seat READS IT BACK FROM THE BOARD. We model
	// COMMAND CHAINS, not minds — the seat boundary is only where the JS happens to cut. And
	// because the outcome is chosen rather than emergent, it is ASSERTABLE.
	directive := directives[r.rng.Intn(len(directives))]
	// A FORCED-UNVERIFIED RUN MINTS ONLY CONTESTED GAPS, and this is the condition the first
	// three attempts at this test each missed in turn. debate.js sits the bench only when the
	// docket is contested (`if (contested.length > 0)`), and the bench is what declares the
	// deadlock that makes the exit UNVERIFIED rather than CEILING. Left to the draw, a run whose
	// gaps all happened to be repairable never contested anything, the judge never sat, and the
	// run reached the round ceiling — which is a different verdict and a passing sweep saying so.
	// DISPUTE-LOST specifically: blue contests and red REJECTS, so the docket is contested AND
	// the gap stays open, which are the two things the deadlock arm reads.
	if r.forceUnverified {
		directive = dirDisputeLost
	}
	// The estoppel run mints only gaps blue will APPLY, because applied_verbatim is the only
	// state the guard keys on and dirApply is the only branch that sets it.
	if r.forceEstoppel {
		directive = dirApply
	}
	// The verified run mints only gaps blue repairs and the lens closes: an applied or a countered
	// edit is movement, so neither reaches impasse before the lens's closing sitting.
	if r.forceVerified {
		directive = pick(r.rng, []string{dirApply, dirCounter})
	}
	kind := checkKinds[r.rng.Intn(len(checkKinds))]
	if directive == dirProve || directive == dirProveDrifts {
		// The demand and the answer must agree: a gap settled by computing is minted as a
		// COMPUTATION check, so the tool's own close guard is in force alongside the scenario.
		kind = "computation"
	}
	args := []string{"--json", "mint", "--seat-id", seatID, "--problem", "fuzz problem", "--check-kind", kind,
		"--check", "acc", "--fix", directive, "--severity", r.g(), "--likelihood", r.g(), "--impact", r.g(),
		// --quote is the anchor ESTOPPEL keys on and the sweep never passed it; --key is
		// mint's crash-retry idempotency
		// handle; --reason is the prose channel. Four fields on the most consequential verb in
		// the tool, none of them ever exercised by a run.
		// A QUOTE FROM THE SEEDED REPORT, not a composed label. Since 0.63.0 a mint's --location
		// is matched against blue/report.md — the same rule `lens finding` has always been held
		// to — so `"§ fuzz — " + directive[:12]` is refused. The sweep went to ZERO gaps the
		// moment the check landed, and every gap-dependent verb (close, regrade, opinion,
		// closing, manifest-row, reproduce) followed it to zero. The coverage gate caught that;
		// the mint refusal alone would have looked like a quieter run.
		//
		// It quotes the ANCHOR sentence, not the cost sentence: the blue-edit drive swaps the
		// cost sentence between two phrasings, so a mint quoting it fails once an edit has run —
		// which is correct behaviour (red must quote text that is actually there) and useless as
		// a fixture. The anchor sentence is stable, and the invisible anchor layer spliced into
		// it is ignored by the match.
		"--reason", "fuzz: the argument for raising this"}
	// ONE SUBJECT: --quote OR --about, and `lens mint` refuses both together. The about arm is
	// how a gap names a defect that is not report text at all, which is the anchor #742 added
	// and #787 shipped — and which no drive reached, so the whole pair read as covered because
	// the mint verb itself (then `merge mint`) ran 115 times.
	//
	// THE CONCRETE-PROPOSAL ARM BELOW APPENDS ITS OWN --quote, so the choice has to be made ONCE,
	// here, for the whole call. It was made twice: this block took the about arm on a 25% coin
	// and the proposal block then appended --quote on a 40% coin, and the mint refused the
	// pair exactly as it is supposed to — silently, because r.mint discards the error and returns
	// "". About one mint in ten never reached the board, which thins every gap-dependent drive
	// downstream of it and is invisible in a passing sweep.
	// A gap with no concrete proposal has nothing for blue to apply verbatim, so the estoppel run
	// always prescribes one.
	proposing := r.forceEstoppel || r.coin(40)
	if k, ref := r.aboutAnchor(); k != "" && !proposing && r.coin(25) {
		args = append(args, "--about-kind", k, "--about", ref)
	} else if !proposing {
		args = append(args, "--quote", "A § fuzz sentence to anchor findings.")
	}
	// THE CLASS IS A WORD OF THE CLOSED SET, and the pool is `by_grade` on purpose: an `always` or
	// `never` class in the pool would move a share of every run's gaps out of grade, and with them
	// the boards, the epochs and the terminal verdicts the sweep's other gates count on — a
	// coverage drive that displaces the sweep's other coverage is a trade (see closeGap's `moot`).
	// So each `always`/`never` class is minted ONCE per sweep, deliberately, outside the forced
	// runs, whose verdicts are their drive. ONE MINT IN TEN takes the catch-all `unclassified` (also
	// `by_grade`), and half of those first try it without --reason, which the record refuses.
	unclassified := r.coin(10)
	driven := ""
	if !unclassified && !r.forceUnverified && !r.forceEstoppel && !r.forceVerified {
		driven, _ = takeClassDrive()
	}
	switch {
	case unclassified:
		args = append(args, "--class", "unclassified")
	case driven != "":
		args = append(args, "--class", driven)
	default:
		args = append(args, "--class", pick(r.rng, mintPoolClasses))
	}
	landed := false
	if driven != "" {
		defer func() { settleClassDrive(driven, landed) }()
	}
	// --key is mint's crash-retry idempotency handle, keyed on the gap it names (contentKey) once
	// every flag is drawn: a key is one gap's name, and a repeat stating another gap is refused.
	keyed := r.coin(40)
	// optional grade/lineage flags — exercised so mint's full flag surface runs, not the minimum.
	if r.coin(50) {
		args = append(args, "--severity", r.g())
	}
	if r.coin(50) {
		args = append(args, "--complexity", r.g())
	}
	if fl := r.someFinding(); foundBy != "" {
		args = append(args, "--found-by", foundBy)
	} else if fl != "" && r.coin(50) {
		args = append(args, "--found-by", fl) // the lens finding that surfaced it (real TOOL-assigned id)
	}
	if open := r.openGaps(); len(open) > 0 && r.coin(30) {
		args = append(args, "--supersedes", open[r.rng.Intn(len(open))]) // lineage: this gap replaces an ancestor
	}
	// #267 stage 3: a CONCRETE proposed fix, which the tool validates against the live report
	// and which DERIVES fix_basis: verified. It swaps the seeded edit-target sentence, exactly
	// as blue's own edit drive does, so a legal pair exists whichever way the last edit left
	// the file — and both branches (proposal present / prose only) run across the sweep.
	if proposing {
		anchored := false
		if cur, err := reportproj.RenderFromRecord(r.run()); err == nil {
			fixOld, fixNew := "rising over time", "climbing sharply"
			if !strings.Contains(string(cur), fixOld) {
				fixOld, fixNew = fixNew, fixOld
			}
			if strings.Contains(string(cur), fixOld) {
				run := r.edgeRun(cur, fixOld)
				args = append(args, "--quote", fixOld+run, "--new", fixNew+run)
				anchored = true
			}
		}
		// The proposal's span was not on the page, so this mint still needs an anchor of its own.
		if !anchored {
			args = append(args, "--quote", "A § fuzz sentence to anchor findings.")
		}
	}
	if keyed {
		args = append(args, "--key", contentKey("K", args))
	}
	own := foundBy != "" && r.ownUnansweredRaise(seatID, foundBy)
	// A GAP NO CLASS FITS SAYS WHY. Tried without --reason, the catch-all is refused by the record's
	// own check, naming the duty; the refusal is counted only when it says so.
	if unclassified && r.coin(50) {
		var bare []string
		for k := 0; k < len(args); k++ {
			if args[k] == "--reason" {
				k++
				continue
			}
			bare = append(bare, args[k])
		}
		if out, err := r.exec(bare...); err != nil && strings.Contains(out, "a gap no class fits says why") {
			r.mu.Lock()
			r.unclassifiedRefused++
			r.mu.Unlock()
		} else if err == nil {
			r.noteApplyMiss("an unclassified mint without --reason landed")
		}
	}
	out, err := r.exec(args...)
	if err != nil {
		if strings.Contains(out, "its mint budget admits no more") {
			// READ AGAIN AFTER THE REFUSAL: a raise answered in between is a mint that counts, and
			// an answer is never undone, so a raise still unanswered now was unanswered at the write.
			violated := own && r.ownUnansweredRaise(seatID, foundBy)
			r.mu.Lock()
			r.budgetRefused++
			if violated && r.answerBudgetErr == "" {
				r.answerBudgetErr = seatID + "'s mint from its own raising finding " + foundBy + " was refused for budget: " + firstLine(out)
			}
			r.mu.Unlock()
		}
		return ""
	}
	if own {
		r.mu.Lock()
		r.ownAnswers++
		r.mu.Unlock()
	}
	var env struct {
		Result struct {
			GapID string `json:"gap_id"`
		} `json:"result"`
	}
	if json.Unmarshal([]byte(strings.TrimSpace(out)), &env) != nil {
		return ""
	}
	if env.Result.GapID != "" && keyed && r.coin(30) {
		r.retryKeyed(args, "--check", "acc, restated")
	}
	if env.Result.GapID != "" {
		landed = true
		r.noteMint(seatID, env.Result.GapID)
		if unclassified {
			r.mu.Lock()
			r.unclassifiedMinted++
			r.mu.Unlock()
		}
	}
	if kind == "computation" && env.Result.GapID != "" {
		r.computationGaps.add(env.Result.GapID)
	}
	return env.Result.GapID
}

// collapseWS collapses runs of whitespace: the guard measures its threshold on
// whitespace-collapsed text, so a length taken any other way would pick a different sentence.
func collapseWS(s string) string { return strings.Join(strings.Fields(s), " ") }

// mintEstopped mints, as LENS seatID, a gap whose --quote is text blue applied VERBATIM from red's
// own --fix-new, with no --supersedes naming the gap that prescribed it. That is exactly the shape
// lens/mint.go refuses, and the refusal is what writes the tool's `log --type estoppel` entry. The
// guard keys on the TEXT, not the seat (record.EstoppelConflict), so any lens's mint against the
// prescription is refused, not only the prescribing lens's.
//
// IT EXPECTS TO FAIL, and the failure is the drive. `exec` records the refusal through noteExec
// either way, so a run where the guard STOPPED firing shows up as estoppels=0 with the applies
// still counted — the two numbers together say whether the guard is quiet or gone. A drive that
// merely called `log --type estoppel` would have satisfied the gate while proving neither.
func (r *runner) mintEstopped(seatID string) {
	gapID, fixNew := r.aVerbatimGap()
	if gapID == "" {
		// The precondition genuinely does not hold: nothing has been applied verbatim YET. Said
		// out loud, because "the guard is gone" and "the drive never had its input" are the two
		// readings of estoppels=0 and they want opposite fixes.
		r.noteEstoppelMiss("nothing applied verbatim yet when the drive fired")
		return
	}
	// THE QUOTE IS THE WHOLE SENTENCE, AND IT HAS TO BE — the guard's skip is an AND:
	//
	//   if len(prescribed) < minEstoppelOverlap && len(quote) < minEstoppelOverlap { continue }
	//
	// minEstoppelOverlap is 40, and the prescription this fuzz swaps ("climbing sharply") is 16,
	// so a mint quoting the prescription ALONE leaves both sides short and the guard declines by
	// design — a short span can legitimately occur in text red never wrote. Quoting the sentence
	// that CONTAINS it puts one side over the threshold, which is the real shape: red raising a
	// finding against a passage whose substance is its own prescription.
	//
	// This is why the seeded report sentence is longer than it needs to be for the swap. At its
	// old length ("The cost is rising over time.", 29 chars) NO span of the fixture reached 40,
	// so the estoppel guard was unreachable from this sweep no matter what was minted — and
	// `verbatimApplied` counted the precondition happily for the whole time.
	cur, rerr := reportproj.RenderFromRecord(r.run())
	if rerr != nil {
		r.noteEstoppelMiss("could not render the report to find the applied sentence")
		return
	}
	sentence := ""
	for _, line := range strings.Split(string(cur), "\n") {
		// VISIBLE TEXT, BOTH SIDES, as EstoppelConflict compares: the applied text carries the gap's
		// anchor and any anchor the edit put back.
		if strings.Contains(anchortext.Visible(line), anchortext.Visible(fixNew)) && len(collapseWS(line)) >= 40 {
			sentence = strings.TrimSpace(line)
			break
		}
	}
	if sentence == "" {
		r.noteEstoppelMiss("the applied text is no longer on the page, or no line around it reaches " +
			"the 40-char overlap the guard requires")
		return
	}
	r.refuse(feov.Conflict, "mint", "--seat-id", seatID, "--class", "spec_underspecification",
		"--problem", "fuzz: raising a defect against text I prescribed", "--check-kind", "document",
		"--check", "acc", "--fix", "reword it", "--severity", r.g(), "--likelihood", r.g(), "--impact", r.g(),
		"--quote", sentence,
		"--reason", "fuzz: estoppel probe — this quotes red's own prescription on "+gapID)
	// REFUSED, AND AS A CONFLICT WITH THE RECORD: an admitted mint here is the guard gone.
	// THE REFUSAL IS NOT COUNTED HERE. `estoppels` is read off the RECORD in the walk below,
	// where the tool's own `log --type estoppel` entry either exists or does not — a driver that
	// counted its own error would report the guard as firing on any refusal at all, including one
	// for a quote that simply was not on the page.
}

// someFinding returns the id of a random lens finding on the record, or "" if none — feeds mint's
// --found-by with a real TOOL-assigned id rather than a fabricated one.
func (r *runner) someFinding() string {
	// THE FINDING FAMILY, NOT THE WHOLE FOLD. `show findings` is FindingsJSONBytes now —
	// EventsOf(run, FINDING) into FindingsJSONOf — so this reproduces the CLI's own path rather
	// than the board it used to ask for (#719). Byte-identity is the property that keeps this
	// safe against the RNG, and it is identity with what the VERB does, not with what it did.
	evs, win, err := record.EventsOf(r.runHandle, recordpb.EventType_EVENT_TYPE_FINDING)
	if err != nil {
		return ""
	}
	findings := record.FindingsJSONOf(evs, win).Findings
	if len(findings) == 0 {
		return ""
	}
	return findings[r.rng.Intn(len(findings))].ID
}

// board reads the record IN PROCESS, through the same builders the binary renders from.
//
// THE HARNESS ASKING ITSELF A QUESTION IS NOT THE HARNESS TESTING THE BINARY, and conflating the
// two is what made this sweep expensive. `openGaps`, `someCitation` and `someFinding` exist so the
// FUZZ can decide what to do next — which gap to act on, which citation to name — and each shelled
// `feov-record` to find out. Measured at 4 runs: `chair show board` 127 invocations, `lens show
// evidence` 45, `chair show findings` 19, out of 1039 total. That is ~48 subprocesses per run, at
// ~12ms each, buying nothing the sweep asserts.
//
// The COVERAGE those verbs owe is unaffected and is not weakened here: the projection sweep in
// runOne drives every view on every role explicitly, which is what the surface gate reads (it goes
// through `drive` now — the first invocation of each path per sweep spawns the binary, the rest
// dispatch the same tree in process). These three call sites were never what proved `show board`
// works — they were the harness reading its own board through a subprocess.
//
// BYTE-IDENTICAL BY CONSTRUCTION, which is why this is safe against the RNG. record.FamilyOf,
// EvidenceJSONOf and FindingsJSONOf are the exact functions the CLI marshals, so the lists below
// hold the same elements in the same order the parsed stdout did — and these feed rng.Intn, where
// a reordering would silently move every downstream draw.
func (r *runner) board() (record.Family, bool) {
	b, err := record.FamilyOf(r.runHandle)
	return b, err == nil
}

func (r *runner) openGaps() []string {
	bj, err := record.BoardJSONOfRun(r.runHandle)
	if err != nil {
		return nil
	}
	var ids []string
	for _, g := range bj.Open {
		ids = append(ids, g.ID)
	}
	return ids
}

// closeGap closes gap id from the sitting lens. The close itself, the regression close and the
// amends_prior close are issued as the gap's ORIGINATOR — the lens whose mint created it — because
// the record refuses a close from anyone else (roundless §III.B.3, requireOriginator); the chair
// dispatching that lens is what this driver models by looking the minter up. A successor minted
// here is the closer's own, against its own budget. The chair's `carry` lives in carryPrior.
func (r *runner) closeGap(chairID, id string, allowReg bool) {
	_ = chairID // the sitting seat; the closer is the minter, looked up below
	closer := r.minterOf(id)
	// A COMPUTATION CHECK IS SATISFIED BEFORE IT IS CLOSED, deterministically.
	//
	// The guard refuses a closure with no proof answering the gap, so leaving it to chance
	// meant unclosable gaps piling up run after run — and every drive that iterates the open
	// set grew with them. Answering it here drives the SATISFIED path on every computation
	// gap, which is the branch a probabilistic prove would mostly skip, and the refusal is
	// still covered by the integration tests that assert it.
	if r.computationGaps.has(id) {
		// A SEAT REGISTERS BEFORE IT APPENDS, AND THIS ONE DID NOT (#664).
		//
		// closeGap is driven from the CHAIR's branch, and it proves as `blue-respond` —
		// `blue prove` RECORDS a `prove` event, so that is an append. In round 1 red-chair
		// runs BEFORE blue-respond ever registers, so a computation gap that becomes
		// closable in round 1 appends as a seat the record has never seen, and `verify` refuses
		// the run with `register-before-append`.
		//
		// It is rare rather than constant because it needs a computation gap closable that
		// early — measured at roughly 1 run in 200, which is exactly why it read as a flake and
		// survived the #656 waiter: that ordered a seat against ITSELF for callers going through
		// register(), and this path never called register() at all.
		//
		// register() is idempotent and waits for the register to land, so this is a no-op on
		// every later round and the ordering it needs on the first.
		r.register("blue", "blue-respond")
		name := "fuzz-answer-" + id + ".js"
		if err := r.writeScript(name, "console.log('answers "+id+"');"); err == nil {
			r.admit("prove", "--seat-id", "blue-respond", "--quote", "§ fuzz sentence",
				"--script", name, "--answers", id, "--reason", "fuzz: settling the computation check")
		}
		// A PROOF WHOSE FAILURE IS THE RESULT, once per run. `prove` refuses a script that exits
		// non-zero — correctly, because a broken proof usually means a broken script — UNLESS the
		// seat says the failure is the finding, which is what --expect-error asserts. That is a
		// second contract on the same verb and nothing had ever driven it; the flag gate could
		// not say so while cli.CommandFlags() came back empty (#654). It deliberately does NOT
		// pass --answers: a proof that something is absent settles no computation check.
		r.mu.Lock()
		first := !r.provedExpectedError
		r.provedExpectedError = true
		r.mu.Unlock()
		if first {
			absent := "fuzz-absent-" + id + ".js"
			// The output carries an environment error the tool recognises: --expect-error on a run
			// that failed nothing is refused.
			body := "console.error('ls: cannot access /fuzz-absent: No such file or directory'); process.exit(3);"
			if err := r.writeScript(absent, body); err == nil {
				r.admit("prove", "--seat-id", "blue-respond", "--quote", "§ fuzz sentence",
					"--script", absent, "--expect-error",
					"--reason", "fuzz: the absence IS the result, recorded as such")
			}
			clean := "fuzz-clean-" + id + ".js"
			if err := r.writeScript(clean, "console.log(1)"); err == nil {
				r.refuse(feov.Validation, "prove", "--seat-id", "blue-respond", "--quote", "§ fuzz sentence",
					"--script", clean, "--expect-error", "--reason", "fuzz: claims a failure the run does not have")
			}
		}
	}
	// A regression close carries lineage forward: it mints a successor and closes WITH it
	// (record.go requires --successor for repaired_with_regression). Only allowed on the first close
	// pass, so the successors it spawns are plain-closed on a later pass and the loop terminates.
	if allowReg && r.coin(35) {
		// THE CLOSER MINTS ITS OWN SUCCESSOR: it is the seat that is sitting, and a mint past its
		// budget is refused rather than laundered through a lens that is not. The gap being closed
		// is still open when its successor is minted, so under the sweep's floor of one open gap
		// per lens that refusal is the usual case, and the remainder is
		// forwarded to a gap the board already holds open — which is what a successor is: a real,
		// still-open gap, and the record refuses a dead-end forwarding address either way.
		// A forced run past its first sitting forwards only: a successor is a fresh gap (see findsFresh).
		succ := ""
		if r.findsFresh() {
			succ = r.mint(closer)
		}
		if succ == "" {
			for _, other := range r.openGaps() {
				if other != id {
					succ = other
					break
				}
			}
		}
		if succ != "" {
			if _, err := r.exec("close", "--seat-id", closer, "--id", id, "--as", "repaired_with_regression",
				"--superseded-by", succ, "--reason", "fuzz regression close", "--verified-by", closer, "--verified-with", "fuzz", "--verified-against", "rec"); err == nil {
				return
			}
		}
	}
	// THE WHOLE CLOSURE VOCABULARY, not just `closed`. #342 closed the set, so the
	// enum-coverage sweep now demands every value be reached — and three of them
	// (not_a_defect, defect_accepted, defect_owed_elsewhere) had never been driven by
	// anything, on either closing verb, in the tool's life.
	//
	// `amends_prior` takes --supersedes: it names a defect found BETWEEN two repairs that each
	// closed clean earlier, so it needs a prior closure to amend.
	// 20% ON TOP OF "a prior closure exists" left `close --as amends_prior` never driven across
	// 60 runs. The precondition is already the rare part.
	if prior := r.closedGapIDs(); len(prior) > 0 {
		if _, err := r.exec("close", "--seat-id", closer, "--id", id, "--as", "amends_prior",
			"--supersedes", prior[r.rng.Intn(len(prior))], "--reason", "fuzz: found between two clean repairs",
			"--verified-by", closer, "--verified-with", "fuzz", "--verified-against", "rec"); err == nil {
			return
		}
	}
	// `moot` IS DRIVEN ONCE PER SWEEP, NOT MIXED INTO THE POOL. The vocabulary is shared between
	// the closing verbs (#342), so a word added for the bench's docket ruling reaches this verb
	// too, and the enum-coverage gate names it the moment it exists without a drive.
	//
	// Putting it in the pool below is the obvious answer and it costs the sweep its other gates:
	// a fifth word takes a fifth of every early closure from the four that were there, and it
	// closes a gap while leaving the merits UNREACHED, so the boards, the epochs and the terminal
	// verdicts all move. MEASURED — with `moot` in the pool the sweep went red twice on
	// `outcome --as UNVERIFIED` and once on `lens regrade --severity`, three rare draws that have
	// nothing to do with closing a gap, while the same sweep was green three for three without it.
	// A coverage drive that displaces the sweep's other coverage is not a drive; it is a trade.
	//
	// SWEEP-WIDE AND NOT PER-SEED, which the seed version taught: pinning it to one seed asks that
	// run to reach closeGap, and whether it does is decided by the interleaving rather than the
	// seed — the same seed fired at N=3 and did not at N=40. Across the whole sweep every run
	// closes gaps, so ONE closure anywhere is a fact. mootSpent is what the sweep asserts, and it
	// names this drive rather than leaving the enum census to name the word.
	if mootPending.CompareAndSwap(true, false) {
		if _, err := r.exec("close", "--seat-id", closer, "--id", id, "--as", "moot",
			"--reason", "fuzz: the text this gap attached to is no longer in the report",
			"--verified-by", closer, "--verified-with", "fuzz", "--verified-against", "rec"); err == nil {
			mootSpent.Store(true)
			return
		}
		// A refusal puts it back: the word's whole coverage is this one drive, and a single
		// refused attempt must not spend it.
		mootPending.Store(true)
	}
	as := pick(r.rng, []string{"repaired", "not_a_defect", "defect_accepted", "defect_owed_elsewhere"})
	// A GAP THAT CARRIES A FAILED CHECK STILL STANDING DOES NOT CLOSE (record.requireFailedCheckAnswered):
	// the record still holds its citation unread or its proof unsound, and the close is refused
	// naming it. It closes once blue's remedy or a check that holds answers it.
	if r.carriesStandingFailedCheck(id) {
		r.refuse(feov.Conflict, "close", "--seat-id", closer, "--id", id, "--as", as, "--reason", "fuzz close as "+as,
			"--verified-by", closer, "--verified-with", "fuzz", "--verified-against", "rec")
		r.mu.Lock()
		r.failedCloseRefused++
		r.mu.Unlock()
		return
	}
	args := []string{"close", "--seat-id", closer, "--id", id, "--as", as, "--reason", "fuzz close as " + as,
		"--verified-by", closer, "--verified-with", "fuzz", "--verified-against", "rec"}
	// A CLOSURE NAMES WHAT IT WAS VERIFIED WITH: another gap the record holds is admitted. Two anchors
	// are refused at the write — one no act placed, and the gap being closed, which is not evidence of
	// its own repair — and the closer reads the refusal and closes again in the same sitting without
	// it, as a seat does. A refused close left for a later sitting stays open when this sitting is the
	// lens's last, which the terminal-fate oracle reads as work done and never finished — and whether a
	// later sitting comes is the interleaving's call, so the sweep goes red on a draw, not a defect.
	if r.coin(40) {
		var held []string
		for _, g := range append(r.openGaps(), r.closedGapIDs()...) {
			if g != id {
				held = append(held, g)
			}
		}
		// THE TWO REFUSED ANCHORS ARE PINNED BY DIFFERENCE. The record codes neither refusal (each
		// renders as "error"), so the code alone would pass a refusal for any reason; the close that
		// follows is the same act without the anchor, and it is ADMITTED — which leaves the anchor
		// as the one thing the record refused.
		switch {
		case r.coin(15):
			r.refuse(feov.Internal, append(args, "--anchor", "C-ffffffff")...)
		case r.coin(15):
			r.refuse(feov.Internal, append(args, "--anchor", id)...)
		case len(held) > 0:
			if _, ok := r.admit(append(args, "--anchor", held[r.rng.Intn(len(held))])...); ok {
				return
			}
		}
	}
	r.admit(args...)
}

// carriesStandingFailedCheck reports whether gap id is the gap a standing failed check's seat minted
// about its anchor (record.BlockerFailedCheck's AnsweredBy).
func (r *runner) carriesStandingFailedCheck(id string) bool {
	bs, err := record.PassBlockers(r.runHandle)
	if err != nil {
		return false
	}
	for _, b := range bs {
		if b.Kind == record.BlockerFailedCheck && b.AnsweredBy() == id {
			return true
		}
	}
	return false
}

// carryPrior is the CHAIR's one closure act (roundless §III.B.3): restating a closure the archive
// already holds, from an earlier epoch, with --carried-from naming that epoch. It used to run inside
// closeGap — a lens act — and so ran in the lens's sitting under the chair's seat id; the chair sits
// first each epoch now, so it is driven from the chair branch, where the seat performing it is the
// seat that is sitting. The successor a carry may name is an OPEN gap the board already holds: the
// chair mints nothing, so it forwards to a gap a lens minted.
func (r *runner) carryPrior(seatID string) {
	// A CARRY restates last round's verification rather than asserting a fresh act, so it is its
	// own verb and takes --carried-from instead of the verification triple. IT IS THE CHAIR'S —
	// the one closure verb left in the chair tree, and exempt from the originator check because
	// it restates a closure the archive already holds rather than making one.
	//
	// AND IT CARRIES A GAP THE RECORD ALREADY CLOSED. Driving it against `id` — the gap this call
	// is closing for the FIRST time — was refused every time, because a carry of a gap with no
	// prior closure is a laundering path the record exists to refuse. 20 of 20 refused, and the
	// coverage line read as a driven verb.
	// UNCONDITIONAL WHEN THERE IS SOMETHING TO CARRY. Behind a 30% coin on a precondition that
	// needs a PRIOR round's closure, this fired ONCE across 60 runs — and that one was refused,
	// so the verb's real coverage was zero while the tally read 1. What it was hiding: `merge
	// carry` could not run without --reason at all, because `Close.prose` was annotated
	// unconditionally required and refused before validate's carry exemption could execute. A
	// documented invocation was broken for as long as the drive was too thin to say so.
	// A PRIOR ROUND'S CLOSURE, which is the only thing a carry can restate. `--carried-from`
	// names the round, and a close keys on gap_id, so carrying a gap closed in THIS sitting is a
	// duplicate the record refuses.
	r.mu.Lock()
	round := r.chairRegisters
	r.mu.Unlock()
	if prior := r.closedInARoundBefore(round); len(prior) > 0 {
		carried := prior[r.rng.Intn(len(prior))]
		epoch, ok := r.closedEpoch(carried)
		if !ok {
			return
		}
		// THE EPOCH IS THE CLOSURE'S OWN, read off the board: a carry naming another sitting is
		// refused, and a drive in ten states one to see it refused.
		if r.coin(10) {
			r.refuse(feov.Validation, "carry", "--seat-id", seatID, "--id", carried,
				"--carried-from", strconv.Itoa(epoch+1000), "--as", "repaired")
		}
		carry := []string{"carry", "--seat-id", seatID, "--id", carried,
			"--carried-from", strconv.Itoa(epoch), "--as", "repaired"}
		// THE EXEMPTION ITSELF, half the time. A carry restates a closure an earlier round already
		// argued, so it owes no fresh argument — and nothing drove the no-reason form, which is
		// how the annotation deleted it silently. Both shapes are legal and both are driven.
		if r.coin(50) {
			carry = append(carry, "--reason", "fuzz: carried from the prior round")
		}
		// A carry names where the remainder went as often as a close does; the flag is on both
		// verbs and was driven on neither once the two split. The successor is MINTED here for
		// the same reason the regression close mints one: it must be a real, still-open gap, and
		// the record refuses a dead-end forwarding address.
		if open := r.openGaps(); len(open) > 0 && r.coin(50) {
			carry = append(carry, "--superseded-by", open[r.rng.Intn(len(open))])
		}
		r.admit(carry...)
	}
}

// checkKinds draws uniformly: closeGap satisfies a computation gap before closing it, so the
// kind no longer decides whether a run accumulates dead gaps.
// directives are the scenario vocabulary a mint draws from. The string IS the expected
// behaviour of every seat downstream, and the oracle asserts the record shows it happened.
const (
	dirApply       = "FUZZ-APPLY: blue applies the proposed pair verbatim; red closes it"
	dirCounter     = "FUZZ-COUNTER: blue edits, but not the proposed text; red closes it"
	dirDisputeWon  = "FUZZ-DISPUTE-WON: blue contests the grade and red accepts the regrade"
	dirDisputeLost = "FUZZ-DISPUTE-LOST: blue contests the grade and red rejects it"
	dirIgnore      = "FUZZ-IGNORE: blue does nothing; the gap stays open"
	// dirProve is the COMPUTATION CHAIN, end to end. It is the only axis where the tool
	// re-executes evidence rather than re-reading it, and it was the least represented: `blue
	// prove` fired from a 35% probe, `lens reproduce` picked ANY proof on the record at 35%
	// with no connection to the gap being audited, and red closed the gap BEFORE any of that.
	// The order was inverted — reproduction decorated a decision already taken, when it is the
	// thing that should EARN it. Nothing would have noticed if reproduce returned garbage.
	dirProve = "FUZZ-PROVE: blue settles it by computing; red re-runs the proof and closes only if it holds"
	// dirProveDrifts is the case that makes the re-run MEAN something. A deterministic script
	// always reproduces, so with only dirProve the scenario check and the tool's own
	// computation guard are indistinguishable — probing by deleting the check caught NOTHING.
	// A drifting script (an unseeded random, a live sample) is graded `observed` rather than
	// `reproducible`, and red must NOT close on it: that is the whole difference between a
	// proof and a measurement, and it is the branch a fake with only tidy scripts never shows.
	dirProveDrifts = "FUZZ-PROVE-DRIFTS: blue computes, but the output moves between runs; red cannot close on it"
)

// WEIGHTED so a run can still reach PASS. Every fate is now DERIVED from the directive, so a
// board whose gaps are all IGNORE correctly never closes and the run correctly ends CEILING.
// Under the sweep's terms (k 1, kMax 2) only APPLY and COUNTER are movement before the lens's
// closing sitting; a proof, a grade motion or silence is a stalled exchange, the gap is at impasse
// after one, and the bench remands every docket but a lost dispute: the remand's one more exchange —
// blue and the lens each held to an act on the gap — stalls the same way, the gap is docketed again,
// and the bench, refused a second remand, rules a disposition that closes it.
// MEASURED over 80 runs at seven lenses: 153 of 558 minted gaps ended remanded (PROVE 70,
// PROVE-DRIFTS 33, DISPUTE-WON 26, IGNORE 24), every one of the 71 CEILING runs held at least one,
// and 8 runs reached VERIFIED. At four lenses 63 of 250 gaps ended remanded and 29 of 80 runs
// reached VERIFIED: the share of gaps left at their limit did not move, the gaps per run did. The
// draw keeps both terminal states in the sweep's distribution; verifiedSeed makes VERIFIED a fact
// of every sweep.
var directives = []string{dirApply, dirApply, dirCounter, dirCounter, dirDisputeWon, dirDisputeLost, dirIgnore, dirProve, dirProve, dirProveDrifts}

// satisfied reports whether a directive means the gap is repaired and red should close it.
func satisfied(directive string) bool { return directive == dirApply || directive == dirCounter }

// outcomeRe reads the verdict debate.js states in the assembler's prompt.
var outcomeRe = regexp.MustCompile(`Debate outcome: ([A-Z]+)`)

var checkKinds = []string{"document", "computation", "source"}

// `deferred` was added as a fate in #246 and never driven — a value the tool accepts, the
// registry declares, and no run has ever recorded.
var inquiryStatus = []string{"proposed", "underway", "confirmed", "refuted", "abandoned", "declined", "deferred"}

// nextInquiryStatus round-robins the undirected inquiry's fate so every value is driven BY
// CONSTRUCTION rather than by luck.
//
// It was a uniform random pick behind a 30% branch — a ~6% draw per inquiry for any one fate —
// so the enum-coverage gate passed on most seeds and failed the moment an unrelated change
// shifted the RNG stream. A gate whose verdict depends on the draw is sampling coverage, not
// measuring it, and the failure it produces looks like the change's fault rather than its own.
//
// The counter is PACKAGE-LEVEL and atomic, not a field on runner. Per-run it would be worse than
// random: the sweep makes about 1.6 undirected inquiries per run, so a counter resetting each time
// would only ever reach index 0 and 1 and the last three fates would never be driven at all.
// Atomic because the sweep's runs are concurrent.
var inquiryTick atomic.Int64

func nextInquiryStatus() string {
	return inquiryStatus[int(inquiryTick.Add(1)-1)%len(inquiryStatus)]
}

var obsKind = []string{"reason", "checked-held"}

// disputeDims is the full grade-dimension domain — the fuzz must contest each, not only impact.
var disputeDims = []string{"severity", "likelihood", "impact", "complexity"}

// petitionClasses is the full petition-class domain (debate.js PETITIONS enum).
var petitionClasses = []string{"ethical", "safety", "integrity", "constitutional"}

// regradeDims are the grade axes a regrade draws one of, each spelled as its flag; --complexity is
// passed on every regrade beside it, so every axis is driven.
var regradeDims = []string{"severity", "likelihood", "impact"}

func pick[T any](rng *lockedRand, xs []T) T { return xs[rng.Intn(len(xs))] }

// extras fires a RANDOM subset of a role's REMAINING verb surface, so no two fuzz paths
// look alike and every seat exercises far more than its happy path. Only verbs that keep
// the oracles intact (verify passes, the run terminates, the report renders) are here;
// terminal/verdict-shaping acts (halt, certify, outcome, verdict) stay in the structured
// cases above. Reference-taking verbs are gated on a referent existing.
func (r *runner) extras(role, seatID string, open []string) {
	// BOTH ARMS OF THE CHANNEL. --none is the explicit negative a sitting closes with when
	// nothing blocked it, and it writes a DIFFERENT event type — so a sweep that only ever
	// drove the complaint arm would leave the attestation path unexercised, which is the
	// arm the whole change exists to make observable.
	//
	// EXCLUSIVE, AND ONE OF THEM ALWAYS FIRES. These were independent coins at 50% and 30%, so a
	// sitting could close with NEITHER — which the verb's own help forbids ("CLOSE THIS CHANNEL
	// EVERY SITTING"), and a drive that contradicts the contract it exercises is testing a system
	// nobody ships. It also left `bench friction --none` never passed across 60 runs: the bench
	// sits rarely (5 frictions in the whole sweep), and 30% of rarely is a path the coverage gate
	// reports as missing while the drive is right there.
	// EVERY TYPE A SEAT MAY FILE. The drive must carry all three or the coverage gate reports an
	// undriven enum value on a surface that only runs at a tag. There is no clean-sitting type to
	// drive: clean is derived from a bracketed sitting that filed nothing, so the fuzz exercises it
	// by having seats sit without logging rather than by filing a word.
	//
	// `estoppel` is NOT here and its absence is a rule, not a gap: the enum's own text says it is
	// recorded by the TOOL, not filed by a seat. Driving it from a seat would manufacture a
	// refusal that never happened and teach the sweep an act the design forbids. It is driven
	// where it is actually produced — see the estoppel scenario in blueRespondTo/mint.
	switch pick(r.rng, []string{"defect", "defect", "defect", "request", "request", "friction", "friction"}) {
	case "defect":
		r.do("log", seatID).set("--type", "defect").set("--reason", "fuzz defect from "+seatID).admit()
	case "request":
		r.do("log", seatID).set("--type", "request").set("--reason", "fuzz: "+seatID+" wanted an act that is on no surface").admit()
	case "friction":
		r.do("log", seatID).set("--type", "friction").set("--reason", "fuzz: something cost "+seatID+" a call, noted without a claim that it is actionable").admit()
	default:
		r.do("log", seatID).set("--type", "friction").set("--reason", "fuzz: something cost "+seatID+" a call, noted without a claim that it is actionable").admit()
	}
	// A recorded fate carries an optional --method; feed it sometimes so that flag is exercised too.
	// An inquiry has an id and a LIFECYCLE: propose it, then record its fate.
	inquiry := func(role string) {
		// A DECLINED OR ABANDONED inquiry requires --reason (record.go: an unexplained fate is the
		// decoration this verb exists to refuse), so every fate this drive records carries one.
		// A FATE-CARRYING INQUIRY STAYS `proposed` until red rules it, so the ruling cycle can run
		// on it: red rules it, blue answers. Recording it straight to `underway` or `declined`
		// would skip the cycle, and the oracle would report endorsed inquiries that ended declined
		// without ever being through a ruling.
		//
		// An UNDIRECTED inquiry takes the round-robin fate: one already declined is a real shape
		// (blue weighed it and did not start), and it carries no fate for red to rule from, so the
		// ruling cycle and the oracle both skip it by construction.
		line, st := pick(r.rng, inquiryFates)+" ("+seatID+")", "proposed"
		if r.coin(30) {
			line, st = "fuzz undirected inquiry "+seatID, nextInquiryStatus()
		}
		// PROPOSE AND RECORD ARE TWO VERBS. A proposal is born `proposed` — the tool supplies it,
		// so there is no status to pass — and any other fate is reached by RECORDING it on the
		// inquiry proposed, which is the cycle red rules on and the oracle reads.
		out, ok := r.do("inquiry propose", seatID).set("--question", line).
			set("--hypothesis", "fuzz: what would be true if "+seatID+" paid off").admit()
		// RECORDING `proposed` AGAIN IS A REAL STATE — the inquiry is still open and the seat is
		// saying so rather than settling it — so the fate is drawn from the whole set, not from
		// the set minus its default.
		if ok {
			// A PROPOSAL THAT LANDED AND NAMED NO ID IS A FAILURE, NOT A SKIP. The tool minted an
			// id and printed it; a reader that finds none there has stopped reading the tool.
			id := inquiryIDOf(out)
			if id == "" {
				r.noteInquiryIDErr("`inquiry propose` succeeded and its result names no inquiry id matching " +
					inquiryIDPat.String() + ":\n" + truncate(out))
				return
			}
			r.noteInquiry(id)
			if _, merr := r.do("inquiry record", seatID).set("--id", id).set("--as", st).
				set("--reason", "fuzz: what changed this inquiry's fate").
				on(50, "--method", "fuzz-method").result(); merr == nil {
				r.noteInquiryRecord()
			}
		}
	}
	switch role {
	case "lens":
		// NO inquiry DRIVE HERE: the lens role has no inquiry verb (register/finding/
		// cite/log/show), and a drive of one would be refused every time while the verb gate
		// stayed green on blue's inquiry events — a dead drive that reads as coverage.
	case "chair":
		// THE PER-ROUND REVIEW OF THE REPORT'S ACCOUNT OF ITS OWN RESEARCH. The verb is
		// `inquiry review` and the event it writes is `inquiry_review` — the names differ. It
		// names no inquiry and casts no verdict: a shortfall in the body is an ordinary gap, so
		// `--reason` is its whole payload.
		// UNCONDITIONAL, BECAUSE IT IS A PER-ROUND DUTY AND A HARD PRECONDITION FOR PASS.
		//
		// This sat behind a 40% coin, so three rounds in five could not pass: the tool refuses a
		// PASS with no review for the round, and debate.js refuses a FAIL that names no gaps, so
		// the seat had NO legal verdict and the run wedged. Before the drive honoured the refusal
		// that was invisible — the error was discarded and the harness was told PASS anyway.
		//
		// A coin is the wrong shape for a duty. What varies between runs is what the review FINDS,
		// not whether the read happened.
		r.do("inquiry review", seatID).
			set("--reason", "fuzz: read the report against the lines on the record; the treatment matches").admit()
		// ONE MOTION DRIVER, AND THERE WERE BRIEFLY TWO.
		//
		// The additive stage added a blue-files/merge-rules pair here beside the `blue dispute`
		// scenario driver, because the old verb was still live and both had to be exercised. When
		// the old verb went, the scenario driver became a motion driver too — and the two minted
		// ids independently while tracking them in separate lists. They crossed: M2 was ruled
		// twice, once by each, and the second ruling OVERWROTE the first at replay. The report
		// showed M2 answered with the reasoning written for another gap.
		//
		// The rulings live in answerDisputes now, where the ask is, so the id never leaves the
		// bookkeeping that minted it.
		// THE SPOT-CHECK IS OWED BEFORE A PASS AND AT NO OTHER TIME, and the driver models both
		// halves. A retired lens does not sit again, so the plan lists each whose pin the
		// head has moved past, and the PASS gate refuses a verdict until a spot-check this sitting
		// names every one: with stale areas on the plan the drive always files it, or every run
		// whose lenses retired ends UNVERIFIED on the refusal and VERIFIED goes undriven.
		// With none, the chair samples the archive in about half its sittings and files nothing in
		// the rest — a sitting without a spot-check is a sitting the record holds as sound, and a
		// drive that filed one every sitting would never put that in front of the oracle.
		//
		// A spot-check names no closure only when the board holds none.
		if stale := r.staleAreas(); len(stale) > 0 || r.coin(50) {
			sc := r.do("spot-check", seatID)
			if len(stale) > 0 {
				sc.set("--areas", strings.Join(stale, ","))
			}
			if closed := r.closedGapIDs(); len(closed) > 0 {
				sc.set("--ids", closed[r.rng.Intn(len(closed))]).
					set("--reason", "fuzz: re-read the closure record; the anchor still resolves").admit()
			} else {
				sc.set("--reason", "fuzz: the archive was empty at round start").admit()
			}
		}
		// RED RULES EVERY UNRULED INQUIRY, and the ruling follows the question it was proposed
		// as: the inquiry's own --question carries its intended fate, exactly as a gap's --fix
		// carries its scenario, so a ruling has a relation to what was proposed and a
		// consequence the oracle can check.
		r.ruleOpenInquiries(seatID)
	case "blue":
		// Before any draw, so a run that is not forced draws exactly what it drew without it.
		r.proveToReopen(seatID)
		r.proveToJudgeUnsound(seatID)
		r.proveOwed(seatID)
		r.citeAbsentSpan(seatID)
		r.maybe(45, func() { inquiry("blue") })
		// NO RANDOM FATE HERE. answerInquiryRulings owns the fates: it answers red's ruling,
		// comply or contest, one decision per inquiry. A second writer recording fates at random
		// would fight it, and the oracle would report endorsed inquiries that ended declined and
		// contests the record never recorded because the fate landed before the ruling did.
		// THE CITATION AXIS (#256), driven end to end through the real binary: `blue cite` fetches
		// the source through the run cache and splices an INVISIBLE, IMMORTAL <!--cite:c-…--> anchor
		// at the quoted sentence. --quote must appear VERBATIM in blue/report.md, so it quotes the
		// seeded "§ fuzz" text the same way lens finding does. A --key from a small space exercises
		// the retry short-circuit; a shared URL across seats exercises the cache HIT path (fetch-once),
		// while the per-seat path exercises the MISS path.
		r.maybe(55, func() {
			path := "/blue-shared"
			if r.coin(50) {
				path = "/" + seatID
			}
			url := sourceURL(path)
			cite := r.do("cite", seatID).
				set("--quote", "§ fuzz sentence").
				set("--url", url).
				set("--title", sourceTitle).
				// SOME CITATIONS BACK THE SENTENCE THE EDIT DRIVE REWRITES. Anchored only in the
				// seeded sentence no edit touches, no checked citation is ever reopened, and the
				// plan's reopened arm and the gate behind it are unreached. One cite in four: universe-m19
				// reopened a checked anchor on 9 of its 35 edits, and the sweep's share is held near that.
				on(25, "--quote", "across the whole sampled corpus").
				on(50, "--reason", "fuzz: why this source backs the claim")
			// EVERY CITE OF A SOURCE WITH TEXT QUOTES WORDS THE TOOL FINDS IN IT. The sweep's sources
			// are synthetic text, each holding sourceBody once; 15 cites in 100 type a span the tool
			// must refuse instead — in turn a fragment that starts inside its clause (sourceFragment),
			// one that stops inside it (sourceTruncation), and words the source does not hold
			// (citeSpanRefused). One draw, where the drive drew one before, so every later draw of
			// the run is the one it was.
			span, held := sourceBody(path), true
			bogus := spanAbsent
			if r.coin(15) {
				span, held, bogus = "fuzz words the source does not hold", false, r.nextBogusSpan()
			}
			switch {
			case bogus == spanFragment:
				// A SPAN CUT FROM INSIDE A CLAUSE, which the source holds once: refused, and the
				// refusal shows the sentence rather than placing the fragment.
				cite.set("--source-span", sourceFragment(path)).refuse(feov.Validation)
			case bogus == spanTruncated:
				// A SPAN STOPPED INSIDE ITS CLAUSE: refused, and the tool does not widen it to the
				// sentence it could see.
				cite.set("--source-span", sourceTruncation(path)).refuse(feov.Validation)
			case !held:
				out, err := cite.set("--source-span", span).result()
				r.citeSpanRefused(seatID, out, err)
			default:
				// THE KEY NAMES THE CITATION, everything it states: a retry of the same act under it is
				// idempotent, and the record refuses a key naming another.
				cite.set("--source-span", span).keyed(40, "C").admit()
			}
		})
		// An UNREACHABLE source is an unusable citation: the cite must be REJECTED and the failure
		// logged by the tool as a defect. Driving it here proves the reject path never wedges the run.
		r.maybe(15, func() {
			r.do("cite", seatID).
				set("--quote", "§ fuzz sentence").
				set("--url", "http://127.0.0.1:1/unreachable").
				set("--title", sourceTitle).
				refuse(feov.Internal)
		})
		// A SECTION OUT OF ORDER IS REFUSED: some sittings write one, and the tool refuses each.
		r.maybe(20, func() { r.orderRefused(seatID) })
		// THE ONLY WRITE PATH to blue/report.md, driven end to end through the real binary.
		// It swaps the seeded edit-target sentence between two phrasings, so a valid unique
		// --quote exists whichever way the previous edit left the file — no state to thread.
		//
		// --answers carries the PROVENANCE (#267): the gap this edit responds to. Sent when
		// the board has an open gap, omitted otherwise, so BOTH validation branches run —
		// the reference check and the legal no-gap edit. The --reason deliberately names no
		// gap: prose naming a real gap with --answers empty is REFUSED, and this drive must
		// exercise the path that is allowed to succeed.
		r.maybe(45, func() {
			cur, err := reportproj.RenderFromRecord(r.run())
			if err != nil {
				return
			}
			oldSpan, newSpan := "rising over time", "climbing sharply"
			if !strings.Contains(string(cur), oldSpan) {
				oldSpan, newSpan = newSpan, oldSpan
			}
			if !strings.Contains(string(cur), oldSpan) {
				return // an anchor landed mid-span; skip rather than force a mis-quote
			}
			// The anchors standing at the span's end are quoted and carried, or left out of both.
			run := r.edgeRun(cur, oldSpan)
			// THIS DRIVE NO LONGER CLAIMS TO ANSWER A GAP.
			//
			// It used to pick a random open gap for --answers, and the scenario oracle caught it
			// on its first run: 5 of 60 seeds attached provenance to a gap whose directive said
			// blue does nothing. Two writers disagreed about who had repaired what — which is
			// exactly the defect --answers exists to make impossible, reproduced inside the fake.
			//
			// Answering a gap is blueRespondTo's job, one decision per gap, from the directive.
			// What remains is an UNATTRIBUTED edit: blue sharpening its own prose, which is real
			// and must stay legal (--answers' own help says to omit it when no gap is answered).
			edit := r.do("edit", seatID).
				set("--quote", oldSpan+run).
				set("--new", newSpan+run).
				set("--reason", "fuzz edit: sharper phrasing, answering no gap")
			keyed := r.coin(40)
			if keyed {
				edit.set("--key", contentKey("E", []string{oldSpan + run, newSpan + run}))
			}
			if _, ok := edit.admit(); ok && keyed && r.coin(30) {
				r.retryKeyed(edit.args, "--new", newSpan+" again"+run)
			}
		})
		// #277: settle a claim by COMPUTING it. The script is written into the run dir and the
		// tool runs it twice — so this drives the real interpreter, the cache, the anchor
		// splice and the reproducible/observed grading, not a stub.
		r.maybe(35, func() {
			name := "fuzz-proof-" + seatID + ".js"
			body := "console.log('fuzz proof for " + seatID + "');"
			if r.coin(25) {
				body = "console.log(Math.random());" // exercises the OBSERVED grade
			}
			if err := r.writeScript(name, body); err != nil {
				return
			}
			// SOME PROOFS BACK THE SENTENCE THE EDIT DRIVE REWRITES, as some citations do: anchored
			// only in the sentence no edit touches, no judged proof is ever reopened, and the
			// routing of a reopened proof to its owner (anchorOwnerOf) is unreached.
			pv := r.do("prove", seatID).
				set("--quote", "§ fuzz sentence").
				on(25, "--quote", "across the whole sampled corpus").
				set("--script", name).
				set("--reason", "fuzz: computing rather than arguing")
			// ANSWER a real gap most of the time: a computation check closes only on a proof
			// that names it, so without this the guard would only ever be seen refusing.
			if len(open) > 0 && r.coin(70) {
				pv = pv.set("--answers", pick(r.rng, open))
			}
			// THE KEY NAMES THE PROOF, everything it states with the program's body: a retry of the
			// same act under it is idempotent, and the record refuses a key naming another.
			if r.coin(40) {
				pv.set("--key", contentKey("P", append(slices.Clone(pv.args), body)))
			}
			pv.admit()
		})
		r.maybe(30, func() {
			// RETIRE WHAT WAS ACTUALLY REMOVED. This used to retire "fuzz claim <seat>" — a
			// string that was never in the report — 45 times a sweep. A phantom retire is not
			// merely uninformative: the changelog lists every retirement as a claim that left the
			// report, so retiring something that was never there tells the reader of a removal
			// the report never had.
			//
			// A real retirement names text a recorded edit took out, which is what makes the
			// removal something the record can SHOW rather than something a seat says.
			if claim := r.recentlyEditedOut(); claim != "" {
				rt := r.do("retire", seatID).set("--quote", claim).set("--reason", "fuzz: the claim went with the edit").
					on(50, "--new", "fuzz replacement claim")
				// A MARKER THE CUT LEFT INSIDE A SENTENCE is named, so --anchor runs through its
				// validation — accepted when the edit left that marker alone, refused otherwise.
				id, left := r.anchorsLeftFor(claim)
				switch {
				case id == "" || !r.coin(60):
					rt.admit()
				case left:
					rt.set("--anchor", id).admit()
				default:
					// A MARKER DRAWN FROM THE WHOLE REPORT is admitted where it stands bare where the
					// claim was and refused elsewhere, which is the draw's call; a refused one is
					// followed by the same retire without it, which is admitted.
					bare := append([]string(nil), rt.args...)
					if _, err := rt.set("--anchor", id).result(); err != nil {
						r.admit(bare...)
						noteEither([]string{"retire"}, "a marker drawn from the whole report", false)
					} else {
						noteEither([]string{"retire"}, "a marker drawn from the whole report", true)
					}
				}
			}
		})
		// THE CORRECTNESS MANIFEST: one row per repaired gap, on the record (#318).
		//
		// It used to fire at 40% for ONE random gap, which modelled a seat that mostly skipped
		// its own self-audit — and nothing noticed, because the row had no reader and the
		// coverage metric was counting the ENVELOPE array instead. The verb is now what the
		// prompt asks for and what the scorecard counts, so the fake discharges it the way a
		// compliant blue would: every gap it is repairing this round.
		for _, id := range open {
			r.do("manifest-row", seatID).set("--id", id).
				set("--reason", "fuzz: figures recomputed, universals enumerated, acceptance check run — held").
				on(50, "--reason", "fuzz: the receipt's argument").admit()
		}
	case "bench":
		// run-end certification statement — the bench's, additive (a recorded opinion).
		r.maybe(45, func() {
			r.do("certify", seatID).set("--reason", "fuzz certification statement from "+seatID).admit()
		})
		// A DECLARATION MOVES NO GAP, which is why the coverage walk would otherwise never reach it:
		// every other bench act names something. It is ABOUT an open gap most of the time, declared
		// with --ids — the gap's id bare, or as the report's anchor token in the holding's prose,
		// which is never read for one — so every later work list states that gap's board status
		// beside it as the gap moves, and the declarations oracle in runOne holds each list to the
		// board. Now and then it names a gap nobody minted, which the write refuses.
		r.maybe(35, func() {
			holding := "fuzz holding: a term is construed this way for the rest of the run"
			c := r.do("declare", seatID)
			if len(open) > 0 && r.coin(70) {
				id := pick(r.rng, open)
				named := id
				if r.coin(50) {
					named = anchor.Token(id)
				}
				holding = named + " remains open; " + holding
				about := id
				if r.coin(10) {
					about = "G-ffffffff"
				}
				c = c.set("--ids", about)
			}
			c = c.set("--reason", holding)
			if _, err := c.result(); err == nil {
				r.mu.Lock()
				r.declared++
				r.mu.Unlock()
			}
		})
	}
}

// arr / obj are goja-friendly envelope builders.
func arr(v ...any) []any { return append([]any{}, v...) }

// envelopeFor performs the seat's tool acts and returns the envelope debate.js expects. The
// randomisation here is what drives debate.js's control flow: red's verdict and gap count
// decide whether the loop continues, deadlocks, or terminates.
func (r *runner) envelopeFor(seatID, prompt string) map[string]any {
	switch {
	case strings.HasPrefix(seatID, "blue-synthesize"):
		r.sit("blue", seatID)
		r.extras("blue", seatID, nil)
		r.maybePetition(seatID)
		// THE ENVELOPE RELAYS NOTHING THE RECORD HOLDS: no claim count and no attestation of what the
		// sitting put on the record (TestNoEnvelopeRestatesARecordFact).
		return map[string]any{}

	case strings.HasPrefix(seatID, "red-chair"):
		r.sit("chair", seatID)
		r.planThisSitting = r.dispatchNext(seatID)
		plan := r.planThisSitting
		r.tallyOwed(plan)
		r.tallyTrifleRoutes(plan)
		// THE CHAIR RUNS THE DEBATE AND MINTS NOTHING (plans/roundless.md §III.B.3). Every act
		// this branch used to perform on the board — evaluating repairs, closing, minting fresh
		// gaps through a lens it picked — is the LENS's now and lives in the lens branch below,
		// where the seat that performs it is the seat that is sitting. What is left here is what
		// only the chair does: rule blue's grade motions, sample the archive, keep the inquiries
		// reviewed, record the verdict the board permits, and argue a PASS it withholds.
		//
		// THE FORCED-UNVERIFIED SHAPE leaves blue's grade motions UNRULED. PASS is refused while a
		// motion stands unruled, and once the bench has disposed of the disputed gap nobody is
		// ready — so the plan comes back empty with neither pass_permitted nor ceiling, which is
		// the UNVERIFIED exit. The terminal bench sitting then rules what the chair left.
		var responses []map[string]any
		if !r.forceUnverified {
			responses = r.answerDisputes(seatID)
		}
		r.extras("chair", seatID, r.openGaps())
		r.carryPrior(seatID)
		permitted, _ := plan["pass_permitted"].(bool)
		// A CHAIR MAY WITHHOLD A PASS THE BOARD PERMITS, and its position then says why — the one
		// sitting the position is owed at (record.PositionSittings). Drawn rarely, never in the
		// forced-verified run: the debate ends there UNVERIFIED, nobody being ready.
		if permitted && !r.forceVerified && r.coin(5) {
			r.chairPosition(seatID, true)
			return r.chairEnvelope(seatID, "", "", responses)
		}
		if permitted {
			if out, err := r.exec("--json", "verdict", "--seat-id", seatID, "--as", "PASS"); err == nil {
				r.chairPosition(seatID, false)
				return r.chairEnvelope(seatID, "PASS", out, responses)
			} else {
				// A PASS THE BOARD PERMITS AND THE GATE REFUSES is the chair's own duty undone —
				// an unruled motion, a missing inquiry review, a contradiction never raised — and
				// the run then ends UNVERIFIED with the plan saying nothing about why. Kept on
				// the outcome so the sweep's exit tally names it.
				// exec's error is "<args>: exit status N\n  <the tool's message>" — the message
				// is the second line.
				msg := err.Error()
				if i := strings.IndexByte(msg, '\n'); i >= 0 {
					msg = firstLine(strings.TrimSpace(msg[i+1:]))
				}
				over := ""
				bl, berr := record.PassBlockers(r.runHandle)
				for _, b := range bl {
					if !b.ChairOwned() {
						over = fmt.Sprintf("%s %q, owned by %q", b.Kind, b.Subject, b.Owner)
						break
					}
				}
				r.mu.Lock()
				r.lastPassRefusal = msg
				if over != "" && r.passOverAnotherSeat == "" {
					r.passOverAnotherSeat = over
				}
				if berr != nil && r.passBlockersErr == "" {
					r.passBlockersErr = berr.Error()
				}
				r.mu.Unlock()
			}
		}
		// A FAIL names what stops the chair, and a FAIL over a converged board is refused — so
		// the chair records one only while something material is open, and otherwise records no
		// verdict this sitting, which is what the prompt tells it to do. The FAIL half of the
		// verdict enum is still driven here, as it was: a FAIL is the ordinary mid-run sitting.
		verdict, printed := "", ""
		if len(r.openGaps()) > 0 {
			verdict, printed = r.chairFail(seatID)
		}
		r.chairPosition(seatID, permitted)
		return r.chairEnvelope(seatID, verdict, printed, responses)

	case strings.HasPrefix(seatID, "blue-respond"):
		r.sit("blue", seatID)
		// BLUE ANSWERS WHAT THE PLAN ENGAGED IT ON, not every open gap: the record dispatched blue
		// for exactly these, and its manifest must name them (W2b). Closings are owed on the
		// docketed gaps, which the plan also names.
		// AND THE BOARD IS AUTHORITATIVE, as blue's prompt says: the plan engaged blue on a gap
		// as the epoch opened, and the lens sat first — so a gap the plan names may have been
		// closed by its originator before blue sits, and an act on it (a grade motion, an edit
		// answering it) is refused as an act on a closed gap. Blue works what is still open; the
		// manifest still names what it was engaged on.
		engaged := r.engaged(seatID)
		r.mu.Lock()
		again := r.againOn
		r.againOn = nil
		r.mu.Unlock()
		if again != nil {
			engaged = again // the engine sits blue again on the successors it engaged, not the plan's gaps
			r.mu.Lock()
			r.blueAgain++
			r.mu.Unlock()
		}
		stillOpen := map[string]bool{}
		for _, id := range r.openGaps() {
			stillOpen[id] = true
		}
		var open []string
		for _, id := range engaged {
			if stillOpen[id] {
				open = append(open, id)
			}
		}
		r.fileClosings(seatID) // its closing on each docketed gap it owes one on
		// THE ESCALATION ROUTE IS A PARTY'S. The dispatch dockets a gap at impasse; a party may
		// file the docket motion itself, earlier, and the bench is readied for it. Driven here
		// so `motion docket file` — a verb the dispatch's own docketing left unreached — is a
		// path a seat actually takes.
		// A GAP ITS CLOSING WAS FILED ON THIS SITTING IS REFUSED: docketed again, it would owe a second
		// closing this sitting cannot file.
		if len(open) > 0 && r.coin(15) {
			id := open[r.rng.Intn(len(open))]
			args := []string{"motion", "docket", "file", "--seat-id", seatID, "--id", id,
				"--reason", "fuzz: " + id + " is not converging in the exchange and is the bench's to settle"}
			if r.closedThisSitting(seatID, id) {
				r.refuse(feov.Conflict, args...)
			} else {
				r.admitFresh(id, args...)
			}
		}
		r.escalateTrifle(seatID)
		r.extras("blue", seatID, open)
		// ONE DECISION PER GAP, TAKEN FROM THE GAP. Each open gap carries the scenario it was
		// minted with; blue reads it back off the board and does what it says. This replaces a
		// blanket 40%-per-gap dispute roll plus a 50% verbatim-apply coin, neither of which
		// looked at what red had actually asked for.
		r.answerInquiryRulings(seatID)
		r.presented.add(open...)
		r.blueRespondTo(seatID, open)
		// AN ARRAY, NEVER null, as the envelope schema demands: a Go nil slice marshals as null,
		// and the harness refuses the envelope for it.
		// debate.js rejects an EMPTY manifest on a round with open gaps — a repair must show its
		// receipt. #318: the envelope is a ROUTING REF now, gap ids only; the rows themselves are
		// on the record, written by the manifest-row calls above against this same `open` set.
		manifest := []any{}
		for _, id := range engaged {
			manifest = append(manifest, id)
		}
		r.maybePetition(seatID)
		r.dischargeOwed(seatID)
		env := map[string]any{"manifest": manifest}
		if again == nil {
			r.engageSuccessors(seatID, env)
		}
		return env

	// THE BENCH IS ONE SEAT ASKED FOUR QUESTIONS, so `seatID` no longer discriminates its sittings:
	// every bench prompt renders `SEAT_ID: judge`. Routing on the id sent the petition sitting and
	// the assembly down the adjudication arm, which returns `{dispositions, log}` — `hearPetitions`
	// then iterated `sitting.rulings` and the engine died with "Cannot convert undefined or null to
	// object". 40 of 40 fuzzed runs failed, and a plain `go test` SKIPS this suite, so it read green
	// in PR CI and was red on a tag.
	//
	// THE QUESTION IS IN THE PROMPT, WHERE THE SEAT READS IT. It was recovered from the prompt's
	// opening WORDS through seatclass.ClassifySeat — which answered both "who sat" and "which
	// sitting" from one table. It no longer answers the second: the bench is one seat id and the
	// occasion is a field on its register (#1100), so ClassifySeat returns `judge` for all four
	// sittings by design. A backend that kept asking it got one envelope shape for every bench
	// sitting and reproduced this exact failure.
	//
	// This reads the OCCASION recordClause writes into the prompt — the same word a real bench seat
	// is told and types at `register` — so the fuzz stands in for the seat instead of guessing at it.
	case debatejs.OccasionOf(prompt) == "petition":
		r.sit("bench", seatID, debatejs.OccasionOf(prompt))
		// THE FORCED-UNVERIFIED SHAPE TRIES TO END RULING NOTHING, and the record refuses its envelope
		// naming the petitions it was convened for: a petition sitting cannot end with one standing,
		// so it then rules them like any other. The run's UNVERIFIED exit is the chair's unruled
		// grade motions, which hold a PASS and not the chair's sitting.
		if r.forceUnverified {
			r.envelopeRefusedEarly(seatID, "motion ")
		}
		return r.rulePetitions(seatID) // rule every pending petition (petition-rule events + envelope rulings)

	case strings.HasPrefix(seatID, "judge") && debatejs.OccasionOf(prompt) != "assemble": // adjudication + terminal
		r.sit("bench", seatID, debatejs.OccasionOf(prompt))
		// THE BOARD'S OPEN GAPS, for a declaration to name: the bench is engaged on no gap of its
		// own here, so this list is the only one that gives a fuzz holding a gap to write about.
		r.extras("bench", seatID, r.openGaps())
		// THE TERMINAL SITTING RULES WHAT STILL STANDS whose gavel is the bench's: the petitions,
		// and the docket motions — a gap docketed at the sitting that reached the epoch limit was
		// dispatched to nobody, and `bench outcome` refuses while its motion stands.
		if debatejs.OccasionOf(prompt) == "terminal" {
			r.rulePending(seatID, "terminal")
			r.ruleStandingDockets(seatID)
		}
		// THE BENCH RULES ON WHAT HAPPENED, not on a coin. A gap reaches the docket because
		// its scenario left it open, and the scenario says why: a LOST dispute is a contest
		// red refused and the bench settles it; an IGNORE is blue owing work, which is what
		// `remanded` means — "the material needs another exchange", stated as a decision.
		// THE BENCH RULES THE DOCKET, which the plan names: a gap reaches it at impasse, once per
		// docketing. A remanded gap goes back for one more exchange and, still at impasse, is
		// docketed again; a second remand is refused, and the bench closes it; a disposed one leaves
		// the board. There is no run-level deadlock for the
		// bench to declare any more (plans/roundless.md §III.B.2): impasse is per gap and the
		// record counts it, so the envelope carries only the resolutions.
		// THE BENCH RULES WHAT IT WAS ENGAGED ON — the plan's fresh docket AND the gaps a party
		// escalated by hand, which stand docketed before impasse; iterating the fresh docket alone
		// left the hand-filed motions unruled, and a bench that sat and ruled nothing is not
		// re-readied for them.
		// THE ENVELOPE IS THE BENCH'S ACCOUNT OF WHAT IT RULED, in JUDGE_ENVELOPE's shape: debate.js
		// builds the ruling every later seat is handed from it. A remand row states no reopens_on —
		// a remand closes nothing — and its direction reaches blue and the lens from the record,
		// through the plan's remand_owed.
		res := []any{} // an array, never null — the schema refuses null
		stillOpen := map[string]bool{}
		for _, id := range r.openGaps() {
			stillOpen[id] = true
		}
		for _, id := range r.engaged(seatID) {
			const direction = "blue pursuing the stated direction and reporting what it found"
			row := map[string]any{"gap_id": id, "disposition": "remanded", "settled": settledByFuzzBench,
				"rationale": "fuzz: carried with a stated direction for " + id}
			// A GAP CLOSED SINCE ITS DOCKETING IS NOT REMANDED: the bench reaches for the remand its
			// scenario asks for, the tool refuses it naming the close, and the bench rules the motion
			// with a disposition that closes, as for a lost dispute.
			closed := !stillOpen[id]
			if closed && r.scenarioOf(id) != dirDisputeLost {
				r.remandOfClosedRefused(seatID, id)
			}
			// A GAP AT ITS REMAND LIMIT IS NOT REMANDED EITHER: the bench's work list says a remand of it
			// is refused, the bench reaches for one and is refused naming the dispositions that close
			// the gap, and it rules one of them.
			atLimit := !closed && r.scenarioOf(id) != dirDisputeLost && r.remandPastLimit(seatID, id)
			if atLimit {
				r.remandPastLimitRefused(seatID, id)
			}
			if !closed && !atLimit && r.scenarioOf(id) != dirDisputeLost {
				r.benchDisposes(seatID, id, "remanded", "--direction", direction, "--reason", row["rationale"].(string))
				r.remandRouted(id)
			}
			if closed || atLimit || r.scenarioOf(id) == dirDisputeLost {
				disp := pick(r.rng, []string{"repaired", "not_a_defect", "defect_accepted", "defect_owed_elsewhere", "amends_prior"})
				// A RULING THAT CLOSES THE GAP ANSWERS WHAT WOULD REOPEN IT ONE OF TWO WAYS, and the
				// drive takes both: nothing would (--final), or a stated condition (--reopens-on).
				row = map[string]any{"gap_id": id, "disposition": disp, "settled": settledByFuzzBench,
					"rationale": "docket-rationale-for-" + id, "final": true}
				answer := []string{"--final"}
				if r.coin(50) {
					const reopens = "a source the ruling did not weigh"
					row["final"], row["reopens_on"] = false, reopens
					answer = []string{"--reopens-on", reopens}
				}
				r.benchDisposes(seatID, id, disp, append(answer, "--reason", row["rationale"].(string))...)
			}
			res = append(res, row)
		}
		return map[string]any{"dispositions": res}

	case debatejs.OccasionOf(prompt) == "assemble":
		r.sit("bench", seatID, debatejs.OccasionOf(prompt))
		// THE VERDICT IS READ, NOT INVENTED. debate.js computes the terminal outcome and TELLS
		// the assembler ("Debate outcome: <verdict> after N round(s)") — exactly as a real seat
		// is told. The fake used to ignore that and draw from a hat, so `bench outcome --as`
		// could record a verdict contradicting the engine's own computation, and no oracle had
		// ever checked that a board with open gaps yields UNVERIFIED or a closed one VERIFIED.
		// The single most consequential value the engine emits was uncorrelated with everything
		// upstream of it.
		verd := "UNVERIFIED"
		if m := outcomeRe.FindStringSubmatch(prompt); m != nil {
			verd = m[1]
		}
		if r.forceHalt {
			verd = "HALTED" // a halted run's terminal outcome is HALTED (debate.js computes this)
		}
		// --reason is REQUIRED on outcome (#375): it is the run's terminal act, and every claim
		// or judgment act on this record carries its reasoning. Driving it without one leaves
		// `bench outcome` refused on every seed, which this sweep reports as a false green.
		oargs := []string{"outcome", "--seat-id", seatID, "--as", verd,
			"--reason", "fuzz: the run reached " + verd + " and the bench recorded how it ended"}
		// HOW THE RUN ENDED IS THE VERDICT'S TO SAY. The `--ended` modifier that used to trail it
		// (two 40% coins once, then read off this prompt) is retired with the roundless record:
		// CEILING is derived from the board, deadlock is per gap, and the verdict read from the
		// prompt three lines up is the whole terminal fact.
		//
		// A PETITION FILED AFTER THE LAST PLAN has had no sitting, and `bench outcome` refuses while
		// it stands; the assembling bench holds the gavel, so it rules it first, as the refusal says.
		if !r.forceHalt {
			r.rulePending(seatID, "assemble")
		}
		r.admit(oargs...)
		// THE ASSEMBLY IS THE BENCH. `--seat-id assemble` was the seat that ran this verb; the bench
		// is one seat now and `assemble` is refused at the surface, so the call recorded nothing and
		// every fuzzed run failed on the missing assembly.
		r.admit("assemble", "--seat-id", "judge")
		open := len(r.openGaps())
		return map[string]any{"open_gaps": open}

	case strings.HasPrefix(seatID, "red-lens"):
		// A LENS FINDS AND MINTS; THE ORIGINATOR CLOSES (plans/roundless.md §III.B.3). The plan
		// dispatched this lens either ENGAGED — on gaps it minted that are open and below their
		// limits, which it now re-evaluates against blue's repair — or because it has never sat or
		// the head has moved since it sat. Either way it audits (the shared lens acts below the
		// fallthrough) and may put new gaps on the board against its own budget.
		r.sit("lens", seatID)
		// THE AUDIT IS SERVED THE WHOLE REPORT, and that read is the lens's pin: a lens that is not
		// served it at the head is ready at the head again, until it is or retires. Most sittings
		// read it, in either whole form; the rest read only a selection, which serves part of it and
		// moves nothing, so the sweep reaches a lens readied again at a head it sat at.
		switch n := r.rng.Intn(10); {
		case n < 5:
			r.readOnly("show", seatID, "report")
		case n < 9:
			r.readOnly("show", seatID, "report", "--json")
		default:
			r.readOnly("show", seatID, "report", "--json", "--match", "§")
		}
		r.checkAnchorOffers(seatID)
		r.closeDocketed(seatID)
		engaged := r.engaged(seatID)
		for _, id := range engaged {
			d := r.scenarioOf(id)
			if r.presented.has(id) {
				r.evaluated.add(id)
			}
			// BLUE SITS AFTER THE LENSES IN AN EPOCH, so a gap the plan engages both on has not
			// been answered yet the first time the lens sees it: closing it then is red closing
			// on nothing — which is exactly what left `--answers`, verbatim application and the
			// estoppel guard undriven across forty runs. A lens closes a gap blue has SAT on.
			switch {
			case satisfied(d) && r.presented.has(id):
				// ONE COIN, NOT TWO — see closeGap: the regression close is sampled in exactly
				// one place, so it is driven rather than multiplied down to noise.
				r.closeGap(seatID, id, true)
			case d == dirProve && r.reproduced.has(id):
				r.closeGap(seatID, id, false)
			case d == dirDisputeWon && r.presented.has(id):
				r.closeGap(seatID, id, false)
			default:
				// A GAP NEITHER MOVED NOR CLOSED IS A NULL TURN, which counts toward impasse; a
				// regrade is movement, so sample one — by the minter, which this seat is. EVERY
				// AXIS, not three of four: --complexity is spelled by its flag here.
				r.maybe(30, func() {
					dim := pick(r.rng, regradeDims)
					r.admitFresh(id, "regrade", "--seat-id", seatID, "--id", id, "--reason", "regrade-basis-for-"+id,
						"--"+dim, r.g(), "--complexity", r.g())
				})
			}
		}
		fresh := r.rng.Intn(3)
		if len(engaged) > 0 {
			fresh = r.rng.Intn(2) // an engaged sitting mostly works its own gaps
		}
		for range fresh {
			// A LENS RUNS OUT OF DEFECTS, AND THE DRIVER HAS TO SAY SO. The budget caps what a lens
			// holds open, not what it mints in a run, so a lens that draws fresh mints at every
			// sitting mints again after every close and the run never converges: measured on this
			// sweep, 35 of 40 runs ended CEILING, 27 at the epoch limit, and the forced-VERIFIED
			// run with them. A stock of two per lens over the floor of one still drives both sides
			// of the budget: the second mint is refused while the first is open, and lands once it
			// has closed.
			if r.landedMints(seatID) >= fuzzLensStock || !r.findsFresh() {
				break
			}
			r.mint(seatID) // refused past the budget, which the sweep counts as a driven refusal
		}
		if r.forceEstoppel {
			r.mintEstopped(seatID)
		} else if r.findsFresh() {
			r.maybe(35, func() { r.mintEstopped(seatID) })
		}
		fallthrough
	default: // frontier, blue lanes, red lenses — register, and lenses record onto the channel
		if seatID != "" {
			// THE DEFAULT IS lens, SO EVERY MAPPING MISS IS A SILENTLY UNREGISTERED SEAT.
			// `frontier` is a BLUE seat (roles.go: blue owns blue-*, frontier) and was mapped
			// to lens here, so its register was REFUSED once per run, in every run, and that
			// seat's whole path went unexercised behind a green sweep. Found by the execution
			// tally (lens register: exactly 60 refusals across 60 runs — one per run is a
			// pattern, not noise).
			role := "lens"
			if strings.HasPrefix(seatID, "blue") || strings.HasPrefix(seatID, "frontier") {
				role = "blue"
			}
			r.register(role, seatID)
			// The non-prose lens channel is the RECORD now, not red/candidates files: a
			// citation lens records cite events, and every red lens records finding events
			// (label TOOL-assigned). This is the ONLY harness that drives that path end to
			// end through the real debate.js + binary, so it must actually exercise it.
			// RED RE-RUNS a recorded proof (#277) — the one audit that does not end in believing
			// bytes someone else chose.
			// RED RE-RUNS THE PROOF FOR THE GAP IT IS AUDITING, and the result is what red
			// closes on. It used to pick ANY proof on the record at 35% with no connection to
			// the gap, AFTER red had already closed it — reproduction decorating a decision
			// instead of earning it. Nothing would have noticed if reproduce returned garbage.
			r.reproveOpenProofs(seatID)
			if strings.HasPrefix(seatID, "red-lens") {
				// AN UNCHECKED CITATION OR PROOF IS CHECKED, EVERY ONE THE OWNER'S SITTING OWES: its
				// item blocks the owner's envelope. A citation is verified at its anchor — read, or
				// found `unreachable`, which leaves it on a failed check whose gap answerFailed
				// mints; a proof is judged, sound or unsound, and re-run by
				// the sha256 its row carries. The anchor's owner alone drives it, as the seat whose
				// work list holds the item — the evidence lens a citation, the run's proof owner a
				// proof: seven lenses each re-running a proof per sitting put the longest runs at
				// the sweep's ten-minute limit.
				//
				// THE FORCED-VERIFIED RUN CHECKS EVERY ONE, AND READS EACH. The owner is readied once
				// per head move, so an item its sitting leaves holds the PASS until the report moves
				// again, and the run that exists to write VERIFIED ends UNVERIFIED on a coin.
				// A CONTRADICTION THIS LENS LEFT OPEN IN AN EARLIER SITTING IS ANSWERED NOW, by its id.
				r.answerContradictions(seatID)
				// A PROOF IT JUDGED UNSOUND IS CARRIED TO BLUE BY ITS GAP, and judged again once carried.
				r.answerFailed(seatID)
				// ITS CLOSING ON EACH GAP OF ITS OWN THAT IS DOCKETED: the lens is one of the gap's two parties.
				r.fileClosings(seatID)
				if seatID == "red-lens-evidence" || seatID == proofOwner {
					var owed []anchorRef
					for _, it := range r.unchecked() {
						if it.Sha != "" && seatID == proofOwner || it.Sha == "" && seatID == "red-lens-evidence" {
							owed = append(owed, it)
						}
					}
					// EVERY ONE, because each is owed: an unchecked anchor's item blocks its owner's
					// envelope, so an item this sitting leaves refuses it. What is drawn is the result.
					//
					// A FAILED CHECK ONE TIME IN TEN. Each starts a chain — the reader's gap, blue's
					// remedy, a fresh check — and drawn at even odds over every owed anchor the chains
					// outran the sweep's time limit; the archive holds two unreachable readings in 31
					// runs. The forced-VERIFIED run reads every one, and neither forced run judges a
					// proof unsound (judgement).
					for _, it := range owed {
						// A FORCED CHAIN'S PROOF IS ITS OWN DRIVE'S: judgeToReopen and judgeUnsound judge
						// it below, and dischargeOwed clears any they leave before the envelope.
						if it.Sha != "" && r.forcedChainProof(it.Anchor) {
							continue
						}
						if !r.firstThisSitting(seatID, it.Anchor) {
							continue
						}
						if it.Sha != "" {
							as := "sound"
							if r.judgement() == "unsound" && r.coin(20) {
								as = "unsound"
							}
							r.do("reproduce", seatID).set("--anchor", it.Anchor).set("--as", as).
								set("--reason", "fuzz: read the script against the sentence as it now stands").admit()
						} else {
							as := "supports"
							if !r.forceVerified && r.coin(10) {
								as = "unreachable"
							}
							r.do("verify", seatID).set("--anchor", it.Anchor).
								set("--as", as).set("--confidence", pick(r.rng, verifyConfidence)).
								set("--reason", "fuzz: read the sentence as it now stands against its source").admit()
						}
					}
				}
				// THE PROOF OWNER JUDGES ONE PROOF IT IS OFFERED, ON A COIN — so a judged proof can be
				// reopened by a later edit of its sentence, and so a run still ends with proofs it was
				// offered and left, which hold it from retiring while they stand. A reopenedProofSeed
				// run judges its forced proof first, and draws nothing doing it; so does the
				// unsoundProofSeed run, judging its forced proof unsound.
				r.judgeToReopen(seatID)
				r.judgeUnsound(seatID)
				if seatID == proofOwner && r.coin(35) {
					if offered, err := anchorItemsOn(r.run(), seatID, "proof"); err == nil && len(offered) > 0 {
						if a := pick(r.rng, offered); r.firstThisSitting(seatID, a) {
							r.do("reproduce", seatID).set("--anchor", a).set("--as", r.judgement()).
								set("--reason", "fuzz: read the script against the sentence it backs").admit()
						}
					}
				}
				// BOTH CASES, and the anchored one uses a REAL citation off the record: red
				// adjudicating a citation blue authored, and red corroborating a source it found
				// itself.
				// TWO VERBS, because they are two acts: `verify` adjudicates a citation BLUE
				// authored and requires its anchor, `corroborate` records a source RED found
				// and requires the url and title that are the only thing identifying it.
				// Falling back to corroborate when blue has cited nothing keeps the axis driven
				// from round 0 rather than only after the first cite lands. Only the citation's
				// owner verifies (citationToVerify); every other lens corroborates.
				// A REAL SPAN OF THE LIVE REPORT. This quoted `"fuzz claim "+seatID`, which is in
				// no document — and a SUPPORTING corroboration splices a citation anchor at the
				// claim now, so every one of them would be refused and the drive discards its
				// error. The seeded anchor sentence is the right span: the edit target below it
				// is swapped back and forth by blue's edits, this one is left alone precisely so
				// findings and cites have something stable to attach to.
				// ONE PAGE OF A CACHED PDF, for a citation with pages: the lens caches the served scan
				// unread and draws page 1, or asks for a page the one-page document does not have and
				// is refused. It records nothing either way.
				r.maybe(10, func() {
					out, ok := r.admit("fetch", "--seat-id", seatID, "--url", sourceURL("/scan.pdf"), "--ocr=false", "--json")
					var sum struct {
						Sha256 string `json:"sha256"`
					}
					if !ok {
						return
					}
					if json.Unmarshal([]byte(out), &sum) != nil || sum.Sha256 == "" {
						r.noteActErr("fetch --json of the scan names no sha256: " + firstLine(out))
						return
					}
					// PAGE 2 OF A ONE-PAGE SCAN IS REFUSED, uncoded, so the refusal is pinned by
					// difference: the same render of page 1 follows it and is admitted.
					if r.coin(25) {
						r.do("render-page", seatID).set("--sha", sum.Sha256).set("--page", "2").refuse(feov.Internal)
					}
					r.do("render-page", seatID).set("--sha", sum.Sha256).set("--page", "1").admit()
				})
				r.sourceQuoteRefused(seatID)
				claim := "A § fuzz sentence to anchor findings."
				outcome := verifyOutcomes[r.rng.Intn(len(verifyOutcomes))]
				// A CONTRADICTION OWES A GAP, AND A GAP IS A MINT. A forced run mints at its first
				// sitting only (findsFresh), which is its bound; a contradicting reading after it would
				// owe a mint the bound forbids, so it reads the source as supporting instead.
				if (outcome == "refutes" || outcome == "absent") && !r.findsFresh() {
					outcome = "supports"
				}
				// A BRIDGE IS NAMED, AND ONE READING IN TEN NAMES NONE: the refusal, driven.
				kindless := outcome == "supports_with_bridge" && r.coin(10)
				bridge := bridgeKinds[r.rng.Intn(len(bridgeKinds))]
				axes := func(c *cmd) *cmd {
					if outcome == "supports_with_bridge" && !kindless {
						c = c.set("--bridge", bridge)
					}
					return c.set("--as", outcome).
						set("--confidence", verifyConfidence[r.rng.Intn(len(verifyConfidence))]).
						set("--reason", "fuzz: what the source actually says").
						on(60, "--access-date", "2026-07-24")
				}
				read := claim // the sentence the reading recorded, which the finding below quotes
				// THE CONTRADICTION THE READING NAMES, from its own output: the raise below names it
				// by that id, so a reading that stopped printing it fails the run.
				printed := ""
				names := func(out string, err error) {
					if err != nil {
						return
					}
					id, ok := contradictionOf(out)
					switch contradicting := outcome == "refutes" || outcome == "absent"; {
					case contradicting && !ok:
						r.noteActErr("a " + outcome + " reading printed no contradiction: " + firstLine(out))
					case !contradicting && ok:
						r.noteActErr("a " + outcome + " reading printed contradiction " + id + ": " + firstLine(out))
					}
					printed = id
				}
				if anchor := r.citationToVerify(seatID); anchor != "" && r.coin(70) && r.firstThisSitting(seatID, anchor) {
					// --page names a page image, and no citation this sweep writes has pages, so
					// a verify carrying one is refused: the flag and its refusal, driven.
					// The claim is the one verify prints back, as a lens reads it.
					paged := r.coin(10)
					v := axes(r.do("verify", seatID)).set("--anchor", anchor).bare("--json")
					if paged {
						v = v.set("--page", "1")
					}
					// THE COPY THE READING NAMES IS THE ONE ITS CITATION'S SOURCE ROW NAMES: the sweep
					// never clears the cache and the first entry for a url is the one every read
					// returns, so a reading and the act that placed the citation resolve to one copy.
					wantCopy, wantCopyURL, cited := r.citationCopy(anchor)
					out, err := v.result()
					if err == nil {
						read = claimOf(out)
						if cited {
							r.noteCopy("verify", out, wantCopy, wantCopyURL)
						}
					}
					names(out, err)
					r.noteBridge(seatID, outcome, kindless, out, err)
					// A READING THE DRIVE DOES NOT DRIVE TO BE REFUSED IS ADMITTED: the kindless
					// bridge is noteBridge's to hold, and a --page is refused as a validation — no
					// citation this sweep writes has pages, and a corroboration holds none.
					switch {
					case paged && err == nil:
						r.noteActErr("verify --page on a source with no pages was admitted: " + firstLine(out))
					case paged:
						if env, ok := refusalEnvelopeIn(out); !ok || env.Code != string(feov.Validation) {
							r.noteActErr("verify --page refused without a validation envelope: " + firstLine(out))
						}
					case !kindless && err != nil:
						r.noteActErr("refused, and the drive expects it admitted: " + err.Error())
					}
				} else {
					// A FRESH SOURCE EACH TIME: a corroboration is one per source and claim, so the
					// drive's own retry below is the only repeat it makes.
					// AND A COPY OF IT IN THE CACHE, turned through every route by which the cache
					// answers a url with another's bytes, and sometimes none: the reading records it.
					url := fmt.Sprintf("https://fuzz.invalid/%s/%d", seatID, corroborations.Add(1))
					wantCopy, wantCopyURL := r.seedCopy(url)
					c := axes(r.do("corroborate", seatID)).set("--quote", claim).
						set("--url", url).
						set("--title", sourceTitle).bare("--json")
					out, err := c.result()
					if err == nil {
						r.noteCopy("corroborate", out, wantCopy, wantCopyURL)
					}
					names(out, err)
					r.noteBridge(seatID, outcome, kindless, out, err)
					if !kindless && err != nil {
						r.noteActErr("refused, and the drive expects it admitted: " + err.Error())
					}
					if err == nil && backsTheClaim(outcome) {
						r.retryCorroboration(c.args, out)
					}
				}
				// AND RED RAISES WHAT IT CONTRADICTED. A `refutes` or `absent` reading is red
				// saying the report asserts what its source does not, and a PASS is refused until
				// a finding naming the contradiction by its id raises it and the gap minted from
				// that finding answers it. The tool will not write the finding itself, because that
				// would mean inventing its three grades. Driving the remedy here is what makes both
				// the gate and the act that clears it real in the sweep; without it every run with a
				// negative reading would wedge at the verdict.
				//
				// A QUOTE RAISES NOTHING. On a coin the lens first files a finding quoting the claim
				// — for a verification, the sentence its anchor marks, which for a quarter of the
				// citations is the edited one — and the contradiction must still read unraised.
				//
				// NOW OR AT THE END OF THE SITTING. Raised now, it is raised by the id the reading
				// printed. Left on a coin, it is raised by answerContradictions before the envelope:
				// the sitting cannot end while it stands.
				if outcome == "refutes" || outcome == "absent" {
					switch {
					case r.coin(40):
						r.quoteRaisesNothing(seatID, read)
						r.raisePrinted(seatID, printed)
					case r.forceVerified || r.coin(50):
						r.raisePrinted(seatID, printed)
					}
				}
				// --key from a small space so a repeated dispatch exercises retry idempotency.
				//
				// EXACTLY ONE ANCHOR, AND SOMETIMES IT IS THE OTHER ONE. `lens finding` refuses a
				// finding carrying both --quote and --about ("claiming two subjects"), so this
				// SWAPS rather than adds. The about arm is the anchor an ABSENCE gets: a section
				// missing something, an inquiry whose stated reason is being argued against, a
				// gap already on the docket — none of which has a sentence to quote, and all of
				// which used to borrow an innocent one (#742).
				severity := r.g()
				findArgs := []string{"finding", "--seat-id", seatID,
					"--severity", severity, "--likelihood", r.g(), "--impact", r.g(), "--reason", "fuzz finding"}
				aboutGap := ""
				if k, ref := r.aboutAnchor(); k != "" && r.coin(30) {
					findArgs = append(findArgs, "--about-kind", k, "--about", ref)
					if k == "gap" {
						aboutGap = ref
					}
				} else {
					findArgs = append(findArgs, "--quote", "§ fuzz sentence")
				}
				// keyed on what the finding states, so a repeat is a retry of it (contentKey).
				findArgs = append(findArgs, "--key", contentKey("F", findArgs[3:]))
				var landed bool
				if aboutGap != "" {
					_, landed = r.admitFresh(aboutGap, findArgs...)
				} else {
					_, landed = r.admit(findArgs...)
				}
				if landed && r.coin(20) {
					r.retryKeyed(findArgs, "--severity", otherGrade(severity))
				}
				// Red verifies a cited source by reading the CACHED bytes (#256): the same
				// `fetch` any seat uses. Driving it here is what makes the cache path — miss,
				// store, hit — real in the fuzz rather than unit-tested only.
				//
				// --ocr=false ALWAYS, AND THAT IS NOT A COVERAGE DODGE. The reading is local now
				// (plans/local-ocr.md) — no model, no network — but this fuzz binary is a default
				// build whose OCR engine is the stub, and its synthetic sources are not scanned
				// PDFs, so "on" would exercise nothing while making the sweep's output depend on
				// which build variant ran it. Passing the flag drives the stated-refusal path and
				// satisfies the gate that every flag be exercised; the default (on) belongs to a
				// seat reading one real document, not to a fuzzer.
				// --via live IS DRIVEN; THE OTHER FIVE ARE NOT, AND THAT IS THE HERMETIC LINE.
				//
				// `live` takes the ordinary Resolve path against this sweep's own test server, so
				// naming it explicitly is a real drive of a real branch (the near-miss refusal
				// above runs, and `live` is checked against Vias() before the equality test picks
				// the local path). Every other backend — archive, oa, metadata, arxiv, auto —
				// reaches an EXTERNAL service through Recover, and a sweep that called them would
				// make 40 runs' worth of network requests to third parties and gain a green that
				// depended on their uptime. The flag is exercised; the backends are not this
				// gate's to prove.
				fetchArgs := []string{"fetch", "--seat-id", seatID, "--url", sourceURL("/" + seatID), "--ocr=false"}
				if r.coin(40) {
					fetchArgs = append(fetchArgs, "--via", "live")
				}
				r.admit(fetchArgs...)
				// --at DRIVES ITS REFUSAL, which is the only honest way to drive it here: it is
				// read on the non-live path alone, so a hermetic sweep cannot reach its success
				// case at all. It used to be a SILENT NO-OP on a live fetch — the date dropped,
				// today's page returned, success reported — and that is now a refusal, so this
				// exercises a path a seat will actually hit rather than standing in for one.
				r.maybe(20, func() {
					r.refuse(feov.Validation, "fetch", "--seat-id", seatID, "--url", sourceURL("/"+seatID),
						"--ocr=false", "--at", "20190101")
				})
				r.extras("lens", seatID, nil)
			}
		}
		// EVERY CONTRADICTION THIS LENS OWES IS RAISED BEFORE ITS ENVELOPE: its sitting cannot end
		// while one stands (record.EnvelopeRefusal), so one the drive left for later is raised here.
		r.answerContradictions(seatID)
		r.dischargeOwed(seatID)
		// A PROOF JUDGED UNSOUND AFTER answerFailed RAN — a reopened proof re-run, one the owner was
		// offered — is carried by its gap here, after the drive's last judgement of the sitting.
		r.answerFailed(seatID)
		// AN ENVELOPE THE RECORD REFUSES IS ANSWERED AND OFFERED AGAIN, as the hook makes a live seat do:
		// another seat's act can land after the pass above — blue's cite or proof while the anchor's
		// owner sits is that sitting's to check — and the seat clears what it now owes, at most four
		// times, as a parallel blue can place more than once.
		for range 4 {
			if refusal, err := record.EnvelopeRefusal(r.run(), seatID); err != nil || refusal == "" {
				break
			}
			r.answerContradictions(seatID)
			r.dischargeOwed(seatID)
			r.answerFailed(seatID)
		}
		// A LENS PETITIONS ON THE RECORD — its envelope carries nothing, so the record is its only channel.
		// The forced-UNVERIFIED run guarantees one, which its petition sitting tries to end on
		// unruled and is refused (envelopeRefusedEarly).
		if r.forceUnverified && r.petitionsFiledN() == 0 {
			r.filePetition(seatID)
		} else {
			r.maybePetition(seatID)
		}
		// THE LENS'S ENVELOPE IS EMPTY (record.LensEnvelope); the frontier's and a lane's return
		// is read by nobody.
		return map[string]any{}
	}
}

const preamble = `
globalThis.phase = function(){}; globalThis.log = function(){};
globalThis.budget = { total:null, spent:function(){return 0;}, remaining:function(){return Infinity;} };
globalThis.parallel = async function(t){ return Promise.all(t.map(function(x){return x();})); };
globalThis.pipeline = async function(items){ var st=Array.prototype.slice.call(arguments,1);
  return Promise.all(items.map(async function(it,i){ var v=it; for(var s=0;s<st.length;s++){ try{v=await st[s](v,it,i);}catch(e){return null;} } return v; })); };
`

func debateWrapped(t *testing.T) string {
	t.Helper()
	p, err := repotree.DebateJS()
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read debate.js: %v", err)
	}
	src := strings.Replace(string(b), "export const meta", "const meta", 1)
	return "globalThis.__result = (async () => {\n" + src + "\n})()"
}

type outcome struct {
	seed      int64
	err       string // non-empty = a finding
	runDir    string
	verdict   string         // the terminal verdict this run reached (coverage signal)
	epochs    int            // how many chair sittings it ran (coverage signal)
	why       string         // the plan's first stated reason at exit — what UNVERIFIED and CEILING are made of
	dialectic map[string]int // dialectic events this run left on the record (coverage signal)

	// epochLimit is the debate ending at the run's epoch limit — the last plan's
	// epoch_limit_reached, as debate.js returns it — which is CEILING by the run's term.
	epochLimit bool
	// lensSittings is the runner's.
	lensSittings map[string]int
	// blueSittings and blueAgain are the runner's.
	blueSittings, blueAgain int

	// #256 citation axis: the fetch -> cache -> cite -> anchor chain leaves TWO artifacts a
	// `cite` event alone does not prove — a cached source file and an INVISIBLE anchor spliced
	// into blue/report.md. `cite` is already in verbsWithEvents but is satisfied by red's
	// `lens cite`, which touches neither, so a blue-cite regression would pass the verb gate
	// silently. These two counters are what actually pin the new path.
	citeAnchors int
	cacheFiles  int
	// #267 provenance: blue_edit events carrying --answers. The verb gate below is satisfied
	// by ANY blue_edit, so an edit drive that never sent the flag would pass it silently —
	// exactly the false green the citation counters exist to prevent for cite.
	editAnswers int
	// #267 stage 3: gaps whose fix_basis was EARNED by a validated concrete proposal. The
	// mint verb gate is satisfied by any mint, so a proposal path that stopped validating
	// would pass it while the axis quietly went all-prose.
	verifiedBasis int
	// Mints the record accepted WITH a found_by. The unreached-flags gate counts --found-by when the
	// mint carrying it is refused, so only the event shows a finding was ever named in found_by.
	foundByMints int
	// #267 stage 4: edits that applied red's proposal EXACTLY. Without a counter, the fuzz
	// could counter-edit every time and the estoppel path would never be reached at all.
	verbatimApplied int
	// typedVerbatim is the verbatim applications blue TYPED rather than accepted: the pair the board
	// serves, compared with the one --accept would send (R-24). Counted apart, so a typed arm that
	// stops recording fails even while --accept records.
	typedVerbatim int
	// estoppelMisses is why the estoppel drive declined, carried out for the same reason.
	estoppelMisses map[string]int
	// applyMisses is why it did not, by cause — carried out of the run so a ZERO above arrives
	// with its explanation instead of inviting one.
	applyMisses map[string]int
	// estoppels is the TOOL's own `log --type estoppel` entries: a mint refused against text
	// blue had applied verbatim from red's prescription. Read off the record, never from the
	// driver, so it distinguishes "the guard never fired" from "the guard is gone".
	estoppels int
	// inquiryRecords is the fates the tool accepted on an id the harness read from a propose
	// result (runner.inquiryRecords). The `inquiry` verb gate is satisfied by a proposal alone, and
	// blue's answer to a ruling records a fate by an id read off the record, so neither notices the
	// propose-result reader going blind.
	inquiryRecords int
	// closingsFiled, closingsRefused, closingWaits and positionsWithheld are the runner's.
	closingsFiled, closingsRefused, closingWaits, positionsWithheld int
	// ordersRefused is the edits the tool refused for a section out of order (runner.ordersRefused).
	ordersRefused int
	// spansRefused is the cites the tool refused for a span the source does not hold
	// (runner.spansRefused).
	spansRefused int
	// copiesRead is the copy words each reading verb named (runner.copiesRead).
	copiesRead map[string]map[string]int
	// fragmentsDriven and truncationsDriven are the cites driven to be refused for a span starting
	// and ending mid-clause (runner.fragmentsDriven, runner.truncationsDriven).
	fragmentsDriven, truncationsDriven int
	// quotesRefused is the corroborations the tool refused for a quote the report does not hold
	// (runner.quotesRefused).
	quotesRefused int
	// bridgesRefused and bridgesNamed are runner's.
	bridgesRefused, bridgesNamed int
	// reopenedReadied is the plans that readied the owner of every reopened blocker they held
	// (runner.reopenedReadied), and reopenedPlans the plans that held one at all.
	// proofReopenedReadied, owedKept, proofOwedKept and owedKeptBy are the runner's.
	reopenedReadied, reopenedPlans int
	// unsoundReadied, unsoundPlans and unsoundAnswers are the runner's (runner.tallyUnsound,
	// runner.answerUnsound), and so are unsoundRepeats and unsoundOverOpen.
	unsoundReadied, unsoundPlans, unsoundAnswers int
	unsoundRepeats, unsoundOverOpen              int
	proofReopenedReadied                         map[string]int
	owedKept, proofOwedKept                      int
	owedKeptBy                                   map[string]int
	// unreadAnswers, unreadRepeats, unreadRetires and failedCloseRefused are the runner's.
	unreadAnswers, unreadRepeats, unreadRetires, failedCloseRefused int
	// reproduces is the reproduce events on the record, and reruns those that ran the script.
	reproduces, reruns int
	// ownerOffered is the proofs nobody judged that the run ended with, each offered on its owner's
	// work list and on no other lens's (anchorOffersAgree), plus the owner sittings offered one for its
	// first check (checkAnchorOffers). citationsOffered is the same count of the citations nobody
	// verified, offered to the evidence lens.
	ownerOffered     int
	citationsOffered int
	// repeatsRefused is the second acts the tool refused in a sitting that holds one, each naming
	// the act that answers the first (runner.repeatsRefused).
	repeatsRefused int
	// budgetRefused is the mints the tool refused a lens holding its whole budget open
	// (runner.budgetRefused).
	budgetRefused int
	// ownAnswers is the mints that answered the minting lens's own raise (runner.ownAnswers).
	ownAnswers int
	// unclassifiedMinted and unclassifiedRefused are the catch-all's mints that landed and those
	// refused for want of a reason (runner.unclassifiedMinted, runner.unclassifiedRefused).
	unclassifiedMinted, unclassifiedRefused int
	// rulerFilingsRefused is the filings the tool refused the seat that rules their subject
	// (runner.rulerFilingsRefused).
	rulerFilingsRefused int
	// closedRemandsRefused and trifleRemandsRouted are runner's.
	closedRemandsRefused, trifleRemandsRouted int
	// blueRemandActs, lensRemandActs and pastLimitRemandsRefused are runner's remandActs and
	// pastLimitRemandsRefused.
	blueRemandActs, lensRemandActs, pastLimitRemandsRefused int
}

// installAgent wires r as the seat backend on vm: it parses the seat id from each agent() prompt
// (falling back to the label), records the model tier the dispatch carried (#111 oracle), and
// resolves the promise with r's envelope for that seat.
func (r *runner) installAgent(loop *eventloop.EventLoop, vm *goja.Runtime) {
	vm.Set("agent", func(call goja.FunctionCall) goja.Value {
		prompt := ""
		if len(call.Arguments) > 0 {
			prompt = call.Argument(0).String()
		}
		seatID := ""
		if m := seatRe.FindStringSubmatch(prompt); m != nil {
			seatID = m[1]
		}
		var opts *goja.Object
		if len(call.Arguments) > 1 {
			opts = call.Argument(1).ToObject(vm)
		}
		if seatID == "" && opts != nil { // fall back to label
			if v := opts.Get("label"); v != nil {
				if fields := strings.Fields(v.String()); len(fields) > 0 {
					seatID = fields[0]
				}
			}
		}
		// #111 tier capture: record the model this dispatch carried ("unset" if absent).
		if opts != nil {
			mdl := "unset"
			if v := opts.Get("model"); v != nil && !goja.IsUndefined(v) && !goja.IsNull(v) {
				mdl = v.String()
			}
			r.models = append(r.models, mdl)
		}
		// THE SCHEMA THE HARNESS HOLDS THIS SEAT'S ENVELOPE TO, read on the loop like every other
		// look at the VM; nil for a seat whose return is prose.
		var schema any
		if opts != nil {
			if v := opts.Get("schema"); v != nil && !goja.IsUndefined(v) && !goja.IsNull(v) {
				schema = v.Export()
			}
		}
		// A SEAT WHOSE WORK ARRIVES AS ITEMS IS DISPATCHED WITH A SCHEMA, or it ends its sitting with
		// text and the record is never asked whether it refuses the envelope (record.EnvelopeRefusal,
		// which the PreToolUse hook asks on the call that returns one). A sweep that puts to the record
		// an envelope the engine never asks for reports green a gate no live sitting reaches.
		if schema == nil && record.WorkArrivesAsItems(seatclass.RoleOfSeat(seatID), seatID) {
			r.noteActErr(seatID + " was dispatched with no envelope schema, so its sitting ends with text and the record's refusal of its envelope is never asked")
		}
		// THE SEAT'S WORK LEAVES THE LOOP (#630 phase 3).
		//
		// This built the envelope INLINE and resolved an already-settled promise, so the
		// preamble's `parallel` — `Promise.all(t.map(x => x()))` — ran each thunk to completion
		// before starting the next. debate.js's lane and lens fan-outs were therefore serial in
		// the fuzz however wide they were configured, and the ~270 binary invocations of a run
		// went one at a time. Nothing ever contended for the record's locks, which is why #557's
		// window sat unreachable in the one harness that reaches a seat before the database
		// exists.
		//
		// Now the promise is returned PENDING and the work runs on a goroutine, so a thunk
		// returns immediately and its siblings start. Three lanes hit `register` on a fresh run
		// directory at once, which is the cross-process contention this exists for.
		//
		// goja is single-threaded and the Runtime must not be touched from here: the envelope is
		// built off-loop with no VM access, and vm.ToValue happens back ON the loop.
		p, resolve, reject := vm.NewPromise()
		go func() {
			env := r.envelopeFor(seatID, prompt)
			// ONLY A DISPATCH CARRYING A SCHEMA ENDS THROUGH THE ENVELOPE CALL, so only its envelope
			// is put to the record — as the hook puts it in a live run.
			if env != nil && schema != nil {
				r.envelopeEnds(seatID)
				if why := r.schemas.refusal(schema, env); why != "" {
					r.noteSchemaRefusal(seatID, why)
				}
			}
			loop.RunOnLoop(func(vm *goja.Runtime) {
				if env == nil {
					// A seat that produced nothing is a rejection, not a silent fulfilment with
					// an empty envelope — debate.js reads fields off it and would fail further
					// from the cause.
					_ = reject(vm.ToValue("seat " + seatID + " produced no envelope"))
					return
				}
				_ = resolve(vm.ToValue(env))
			})
		}()
		return vm.ToValue(p)
	})
}

// stallAfter is how long a run may go without one of its commands finishing before driveDebate
// calls it stuck. A command is one invocation of the binary — well under a second on a quiet box —
// and the seats of a run finish several a second between them, so three minutes of none is a
// stuck run at any load this box has been measured at.
const stallAfter = 3 * time.Minute

// driveDebate runs the wrapped debate.js once under goja with r as the seat backend. It returns the
// fulfilled result map (nil unless the debate settled fulfilled) and a non-empty settledErr for any
// preamble error, rejection, hang, or panic. Shared by the sweep (runOne) and the halt test — the
// one place the goja event loop and __result promise are driven.
func driveDebate(r *runner, wrapped string) (result map[string]any, settledErr string) {
	// THE ACT FIRST: an act answered otherwise than expected is the cause of whatever the run did
	// next, so it leads the failure and what followed from it trails.
	defer func() {
		r.mu.Lock()
		act := r.actErr
		r.mu.Unlock()
		if act == "" {
			return
		}
		if settledErr != "" {
			act += "\n  and then: " + settledErr
		}
		settledErr = "an act the drive issued: " + act
	}()
	defer func() {
		if p := recover(); p != nil {
			settledErr = fmt.Sprintf("panic: %v", p)
		}
	}()
	// The fuzz builds its run directly rather than through `setup`, so it must lay down the
	// same inputs/run-config.json setup writes — the run's TERMS (plans/roundless.md §III.B.2:
	// k, kMax, mintBudget, convergenceFraction) are what `dispatch next` reads impasse and the
	// mint budget from, and what the CEILING verdict is DERIVED against (#308).
	//
	// TERMS TIGHTER THAN A SMOKE'S, NOT THE DEFAULTS. This sweep's job is to drive every branch of the engine
	// forty times inside a release gate, not to run forty production-length debates: under the
	// terms of production size (kMax 6, mintBudget 5, every lens area) a run is up to twenty gaps of six exchanges
	// each, and 29 of 40 blew the ten-minute per-run budget while every verb the sweep counts had
	// already been reached. One open gap per lens and two exchanges per gap reach impasse, the
	// docket, the ceiling and the pass in a few epochs. `/research --smoke` runs a floor of three;
	// one keeps a lens at its budget, and the budget's refusal driven, in most sittings.
	//
	// THE EPOCH LIMIT IS SETUP'S DEFAULT, READ FROM WHERE SETUP READS IT. Every run setup creates
	// carries one, and a config that states none is held to none (record.RunParams): without it
	// the computation lens — whose mint budget grows with blue's proofs — carried some seeds to
	// 21–28 chair sittings, past the per-run budget on a loaded box, and no run of the sweep ever
	// took the epoch-limit exit a real run can.
	maxEpochs := record.DefaultParams.MaxEpochs
	if r.maxEpochs > 0 {
		maxEpochs = r.maxEpochs
	}
	if err := writeSetup(filepath.Join(r.runDir, "inputs", "run-config.json"),
		fmt.Appendf(nil, `{"topic":"fuzz","model":"haiku","judgmentModel":"haiku","lanes":3,"k":1,"kMax":2,"mintBudget":1,"convergenceFraction":0.25,"maxEpochs":%d}`,
			maxEpochs)); err != nil {
		return nil, err.Error()
	}

	// STARTED, NOT Run(). `loop.Run` returns once the JS job queue drains, which was correct
	// while agent() resolved inline — the whole debate settled inside one call. Phase 3 (#630)
	// resolves a seat's promise from a goroutine, so the queue can be momentarily empty with
	// work still outstanding, and Run would return mid-debate leaving __result pending. The
	// fuzz would then report "never settled (hang)" for every run: a harness artefact wearing
	// the exact words of a real defect.
	//
	// So the loop runs until the debate SAYS it is done, through __settle below.
	loop := eventloop.NewEventLoop()
	loop.Start()
	defer loop.Stop()

	// Buffered: __settle may fire before the receive, and an unbuffered send on the loop
	// goroutine would deadlock against a loop.Stop() this function has not reached yet.
	done := make(chan string, 1)
	loop.RunOnLoop(func(vm *goja.Runtime) {
		// THE SHAPE PRODUCTION ACTUALLY RUNS (#630). This drove `lanes: 1` under a
		// laneFloorOverride — the --smoke shape, which debate.js:82 REFUSES by default and
		// which the floor exists to stop ("run 2 silently ran under-provisioned at lanes=2").
		// At width one the blue lane fan-out (debate.js:725), red's lens-pass fan-out (:824)
		// and the per-lane method diversity (:560) were all exercised at a width no keeper run
		// uses. Every committed run's inputs/run-config.json says lanes 3; so does this. There is
		// no round count to pass: the run ends when the record says nobody is ready, and
		// debate.js REFUSES a maxRounds argument. The override's own branch stays covered by
		// debate.test.mjs, which is where a JS-level guard belongs.
		vm.Set("args", map[string]any{
			"topic": "fuzz", "runDir": r.runDir, "binDir": binDir(r.bin),
			"lanes": 3,
			// #111: both tiers are REQUIRED — nil refuses dispatch. Both haiku so the tier oracle
			// expects every dispatched seat to carry exactly "haiku".
			"model": "haiku", "judgmentModel": "haiku",
		})
		if _, err := vm.RunString(preamble); err != nil {
			settledErr = "preamble: " + err.Error()
			return
		}
		r.installAgent(loop, vm)
		// __settle is how the debate tells Go it is over, from inside JS. Polling the promise
		// state from outside cannot do it: with async seats there is no moment at which "the
		// queue is empty" means "the debate finished".
		vm.Set("__settle", func(call goja.FunctionCall) goja.Value {
			msg := ""
			if len(call.Arguments) > 0 {
				msg = call.Argument(0).String()
			}
			select {
			case done <- msg:
			default: // already settled; the first word wins
			}
			return goja.Undefined()
		})
		if _, err := vm.RunString(wrapped); err != nil {
			done <- "run: " + err.Error()
			return
		}
		// ATTACHED IN JS, because a rejection has to reach the same channel a fulfilment does.
		// "" is the fulfilled signal; every other string is the failure, so an empty rejection
		// reason cannot read as success.
		if _, err := vm.RunString(`globalThis.__result.then(
			function(){ __settle(""); },
			function(e){ __settle("debate rejected: " + String(e)); })`); err != nil {
			done <- "attaching the settle handler: " + err.Error()
		}
	})

	// THE HANG IS STILL DETECTED, and now it is the only thing this wait can mean. It is a
	// stuck-run tripwire, not a performance assertion, so it measures PROGRESS and not the clock:
	// a run is stuck when no command of its own has finished for stallAfter. A ten-minute budget
	// per run read a slow run as a stuck one — a 22-epoch run took 607s at a load of 25 on four
	// cores — and a run's wall clock is the box's load as much as the run's own work. A run that
	// keeps finishing commands without settling is bounded by its epoch limit, and past that by
	// the package's -timeout.
	r.progress.Store(time.Now().UnixNano())
	stalled := time.NewTicker(5 * time.Second)
	defer stalled.Stop()
	for settled := false; !settled; {
		select {
		case msg := <-done:
			if msg != "" {
				return nil, msg
			}
			settled = true
		case <-stalled.C:
			if idle := time.Since(time.Unix(0, r.progress.Load())); idle > stallAfter {
				return nil, fmt.Sprintf("debate never settled (hang): no command of the run finished for %s", idle.Round(time.Second))
			}
		}
	}

	// READ THE RESULT BACK ON THE LOOP. Touching the Runtime from this goroutine is the one
	// thing goja does not allow, and it is not diagnosable when it goes wrong.
	read := make(chan struct{})
	loop.RunOnLoop(func(vm *goja.Runtime) {
		defer close(read)
		v := vm.Get("__result")
		pr, ok := v.Export().(*goja.Promise)
		if v == nil || !ok {
			settledErr = "debate produced no __result promise"
			return
		}
		switch pr.State() {
		case goja.PromiseStateRejected:
			settledErr = "debate rejected: " + truncate(pr.Result().String())
		case goja.PromiseStatePending:
			settledErr = "debate never settled (hang)"
		case goja.PromiseStateFulfilled:
			m, ok := pr.Result().Export().(map[string]any)
			if !ok {
				// A FULFILLED PROMISE THAT IS NOT AN OBJECT IS NOT A SETTLED DEBATE. The old
				// form left `result` nil and `settledErr` empty, which every caller reads as
				// "the run succeeded" — and the verdict then goes missing downstream instead of
				// here, where the actual value is still in hand to name.
				settledErr = fmt.Sprintf("debate fulfilled with %T, not an object: %s",
					pr.Result().Export(), truncate(pr.Result().String()))
				return
			}
			result = m
		}
	})
	<-read
	return result, settledErr
}

// readVerdict pulls the terminal verdict out of debate.js's settled result, and REFUSES a
// result that does not carry one.
//
// THE ABSENCE OF A VERDICT IS A FAILURE, NEVER A CATEGORY (#637). debate.js computes
//
//	const verdict = halted ? 'HALTED' : redEnv && redEnv.verdict === 'PASS' ? 'VERIFIED'
//	              : ceilingUnaudited ? 'CEILING' : 'UNVERIFIED'
//
// and returns it on every path it can reach, so the value is always one of four non-empty
// strings. A missing or non-string verdict therefore means THIS side failed to read the
// result — it never means the debate reached a nameless outcome.
//
// Tolerated, that miss reached the sweep's tally as `verdicts[""]++`, which counts exactly
// like a real verdict and is not one: the run that opened #637 recorded
// `verdicts=map[:29 CEILING:8 VERIFIED:3]`, 29 of 40 runs settling into a key no reader can
// distinguish from an outcome. That is the plausible zero [[facts-are-fields]] clause 3
// forbids — and worse than a bare zero, because the sweep's own coverage gates then pass over
// a distribution whose largest bucket means "we did not look".
//
// The error names the value AND the keys that were present, because "no verdict" alone does
// not separate a result that arrived empty from one that arrived with a different shape.
func readVerdict(result map[string]any) (string, error) {
	raw, present := result["verdict"]
	if !present {
		return "", fmt.Errorf("debate settled with no verdict: result has no %q key (keys present: %v)",
			"verdict", sortedKeys(result))
	}
	s, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("debate settled with a non-string verdict: %#v (%T)", raw, raw)
	}
	if s == "" {
		return "", fmt.Errorf("debate settled with an EMPTY verdict string (keys present: %v)", sortedKeys(result))
	}
	return s, nil
}

// sortedKeys renders a result's keys in a stable order, so two failures of the same shape
// produce the same message and can be told apart from two different shapes.
func sortedKeys(m map[string]any) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// unverifiedSeed is the one seed in the sweep whose run is FORCED to the UNVERIFIED terminal
// verdict rather than left to the draw.
//
// THE SWEEP WAS STILL HOPING. #630 built the machinery to drive this deterministically —
// `runner.forceUnverified`, and TestFuzzUnverifiedPath which sets it — and the comment at
// forceUnverified says exactly why it is needed: "Left to the draw, a run whose gaps all
// happened to be repairable never contested anything, the judge never sat, and the run reached
// the round ceiling." But `execValues` is in-memory package state and TestFuzzUnverifiedPath
// sits ~600 lines BELOW TestFuzzDebate's coverage assertion, with no t.Parallel() in the file.
// Go runs them in source order, so the deterministic driver could never contribute to the gate
// that needs it, and the gate went on asking the 40-run draw for a value that occurs about once
// in sixty.
//
// MEASURED on origin/main at 65b27a9d with no other diff: red 3 of 4 runs, failing on
// `outcome --as UNVERIFIED`. One forced run costs one seed's worth of natural variation and
// makes the value a fact rather than a coin flip (#637).
const unverifiedSeed = 1

// estoppelSeed is the one seed whose run is FORCED to produce the estoppel conjunction rather
// than left to the draw. It is not unverifiedSeed: that run mints only dirDisputeLost gaps so its
// docket stays contested, and dirApply is the only scenario that applies anything verbatim, so
// one run cannot serve both.
//
// MEASURED on clean main across 13 uncached sweeps, in two worktrees at different times: 6 red,
// all on `ZERO tool estoppel refusals`, two of them also on `blue edit --accept` never passed.
// ~46% on a gate that a tag runs ONCE — and a green told you only that the draw went your way.
const estoppelSeed = 2

// verifiedSeed is the one seed whose run is FORCED to VERIFIED: it mints only APPLY and COUNTER
// gaps, which the lens closes once blue has sat. It is neither of the seeds above: a
// dispute-lost docket holds the PASS, and the estoppel run exists for its refused mint.
//
// MEASURED at seven lenses (plans/feov-lens-bar.md §V #9): sweeps reached VERIFIED on 2 of 40, 8 of
// 80 and 8 of 80 runs, the estoppel seed's all-APPLY run among them each time, carrying the word by
// accident of its drive. At that rate 38 drawn runs find none about once in fifty-five, and
// `outcome --as VERIFIED` then fails the enum census as a missing word.
const verifiedSeed = 3

// limitSeed is the one seed whose run is FORCED to the epoch-limit exit: its config states a limit
// of forcedEpochLimit where every other run states setup's default. At the chair's second sitting
// every lens has sat once and none has retired, so the plan readies them, the limit dispatches
// nobody, and the run ends CEILING by its term — on every sweep, not on a draw.
//
// MEASURED under setup's default of 12: 5, 9, 3, 6, 1 and 2 runs in 40 ended at the limit across
// six sweeps, so a floor on the drawn runs alone finds none about once in fifty. Those drawn exits
// are the deep ones — a gap docketed at the last sitting, a petition standing — and the sweep
// prints their count; the gate holds the forced one.
//
// IT IS PAST THE SIX SEEDS THE RACE VARIANT RUNS (FUZZ_N=6). A run cut to two sittings drives a
// sixth less of everything else, and at six runs `blue retire` succeeds as few as once.
const limitSeed = 7

// forcedEpochLimit is the epoch limit limitSeed's run is set up with.
const forcedEpochLimit = 2

// reopenedProofSeed is the seed whose run is FORCED through a reopened proof; the sweep gates a plan
// readying the proof owner for one.
//
// Left to the draw it is a chain of coins: blue proves the sentence the edit drive rewrites (35%,
// then 25%), the owner judges an offered proof (35%, picking one of several), blue rewrites that
// sentence (45%), and a plan follows. MEASURED on main 61107ba4 across five interleaved sweeps, an
// owner resting on eight runs was readied for a reopened proof 9, 0, 5, 6 and 11 times — red once —
// and seed 9 alone carried it in four. A branch that drew one or two fewer numbers per cite read
// 0, 0, 1, 0, 1, red three times; the same branch drawing those numbers and discarding them read 5,
// 9, 2, 2, 5. The path was intact throughout; the draw moved.
const reopenedProofSeed = 5

// reopenedProofScript is the forced proof's script. Its sha256 is what identifies the forced proof
// on the record (EvidenceProofJSON.Sha256), so the chain never recovers it from report text.
const reopenedProofScript = "console.log('fuzz: the proof a later edit reopens');"

// unsoundProofSeed is the seed whose run is FORCED to a standing unsound judgement that a plan
// readies the judge or blue for (forceUnsoundProof). It is within the six seeds the race variant
// runs, whose sweep gates the arm as well, and casts every lens, so its proofs fall to the
// computation lens.
//
// Left to the draw it is a race the judgement almost always loses. A plan holds an unsound blocker
// only if it still stands when the next chair sitting plans, and blue-respond's sitting fixes every
// unsound proof before its envelope (fixUnsound) — so a judgement made in an epoch whose plan readies
// blue is fixed before any plan reads it. Blue is readied in nearly every plan after the first, and
// the first plan's lens sittings can judge only a proof that stood before them. MEASURED on main
// b434cda0: one sweep minted 505 gaps carrying an unsound judgement and 2 plans held one; an
// eight-run sweep logging the blockers at every sitting minted 34 such gaps, saw blue-respond sit
// with up to 3 unsound proofs standing and the plan after it with none, and 1 of its 59 plans held
// one.
const unsoundProofSeed = 6

// unsoundProofScript is the forced unsound proof's script; its sha256 identifies the proof on the
// record (proofsOfScript).
const unsoundProofScript = "console.log('fuzz: the proof its owner judges unsound');"

// owedProofSeed is the seed whose run is FORCED to keep a proof nobody has checked before every plan
// that follows a blue-respond sitting (forceOwedProof), so the computation lens, once it has sat
// three sittings in a row with no fresh material mint — which the run's mint budget of one makes
// the common case — is a lens a plan keeps from retiring by an unchecked proof (proofOwedKept). It is
// within the six seeds the race variant runs.
//
// Left to the draw, the arm needs blue's proof to land after the owner's last sitting of an epoch and
// before a plan, while the owner is barren: the owner checks every unchecked proof before its
// envelope (dischargeOwed), so only a proof blue records after it is still unchecked when the chair
// plans. One Relocations sweep counted it zero.
const owedProofSeed = 4

// absentSpanSeed is the seed whose run is FORCED to cite a span its source does not hold
// (forceAbsentSpan). Left to the draw, the absent span is the third arm of a rotation counted per
// run (nextBogusSpan): blue's cite drive fires 55 times in 100 and types a bogus span 15 times in
// 100 of those, and only every third bogus span of a run is the absent one. About eight blue sittings
// in a hundred draw a bogus span, so a run of fifteen draws three about one time in eight, and a
// sweep where no run does reads ZERO.
const absentSpanSeed = 8

func runOne(t *testing.T, wrapped, bin string, seed int64, forceUnverified, forceEstoppel, forceVerified bool) (res outcome) {
	runDir, err := os.MkdirTemp("", "fuzz-run-")
	if err != nil {
		// A FULL TMPDIR IS THE LIKELY CAUSE, AND IT HAS TO SAY SO. Discarded, this left runDir
		// empty — and every filepath.Join below then resolves RELATIVE TO THE PACKAGE
		// DIRECTORY, so the sweep seeds a blue/report.md inside the checkout and drives sixty
		// debates against it, failing with a message that names anything except the disk.
		// TMPDIR is a 2 GB tmpfs both in CI and in the record-run validation loop, so running
		// it out is the expected failure here rather than an exotic one.
		return outcome{seed: seed, err: "temp dir for the run: " + err.Error()}
	}
	// A lens finding anchors into blue/report.md and is rejected unless its --location
	// quote is present (slice 1b). Seed a report carrying the fuzzer's finding quote
	// ("§ fuzz") so findings are accepted and the coverage gate sees them.
	// The trailing sentence is the EDIT TARGET: `blue edit` swaps it between two fixed
	// phrasings, so every round has a valid unique span to replace whichever way the last
	// edit left it, while the anchor quote above stays untouched for findings and cites.
	if err := writeSetup(filepath.Join(runDir, "blue", "report.md"), []byte(seedReport)); err != nil {
		return outcome{seed: seed, runDir: runDir, err: err.Error()}
	}
	// THE RUN DECLARES ITS CLASS VOCABULARY, because a real run does. The fuzz used to be exempt
	// by accident: no registry staged meant `validateClass` accepted every slug, so the fuzz drove
	// `mint` for its whole life without ever exercising the check the flag's help describes. When
	// the exemption went, the coverage gate here reported it in one line — ZERO mint events across
	// 60 runs — which is what that gate is for.
	stageRun, srErr := record.NewRun(runDir)
	if srErr != nil {
		return outcome{seed: seed, runDir: runDir, err: "resolve the run: " + srErr.Error()}
	}
	// THE CAST, as setup writes it (plans/roundless.md §III.B.1): every area — the lenses this
	// driver mints through — the chair, the lanes, the bookends. Every register and every
	// `dispatch next` is checked against it.
	// THREE LANES, because the run dispatches three (args.lanes above): a cast of one lane refused
	// blue-lane-2 and blue-lane-3 at register on every run — 80 refusals across 40 — and every act
	// of two of the three lanes ran unregistered behind a green sweep.
	fuzzCast, fuzzLanes := record.CastFor(3)
	if _, err := record.Append(record.Identity{Run: stageRun, SeatID: record.HarnessSeat}, &recordpb.Cast{SeatIds: fuzzCast, LaneSeatIds: fuzzLanes}); err != nil {
		return outcome{seed: seed, runDir: runDir, err: "write the cast: " + err.Error()}
	}
	r := newRunner(bin, runDir, newLockedRand(seed))
	r.forceUnverified = forceUnverified
	r.forceEstoppel = forceEstoppel
	r.forceVerified = forceVerified
	r.forceReopenedProof = seed == reopenedProofSeed
	r.forceUnsoundProof = seed == unsoundProofSeed
	r.forceOwedProof = seed == owedProofSeed
	r.forceAbsentSpan = seed == absentSpanSeed
	if seed == limitSeed {
		r.maxEpochs = forcedEpochLimit
	}

	// INGEST THE ROUND-0 REPORT (#709). The report is the record projection now: blue-synthesize
	// freezes the seeded report into the record and the file is deleted, exactly as the engine does
	// after synthesis and before any sitting. Every verb from here reads and mutates the report
	// through the record — without this, mint/finding/cite/edit all refuse with "no base has been
	// ingested". Driven at synthesis, so base_ingest is no longer exempt from the coverage gate.
	r.register("blue", "blue-synthesize")
	if _, err := r.exec("ingest", "--seat-id", "blue-synthesize"); err != nil {
		return outcome{seed: seed, runDir: runDir, err: "ingest the round-0 report: " + err.Error()}
	}

	res = outcome{seed: seed, runDir: runDir}
	// NAMED RETURN + defer, because this function returns from a dozen places — an oracle that
	// fires early would otherwise drop the run's apply-miss causes on the floor, which is the
	// same silent loss the causes exist to explain.
	defer func() { res.applyMisses, res.estoppelMisses = r.applyMisses, r.estoppelMisses }()
	result, settledErr := driveDebate(r, wrapped)
	if len(r.schemaRefusals) > 0 {
		res.err = fmt.Sprintf("the harness refuses %d envelope(s) under the schema their dispatch carried:\n  %s", len(r.schemaRefusals), strings.Join(r.distinctSchemaRefusals(), "\n  "))
		return res
	}
	if settledErr != "" {
		res.err = settledErr
		return res
	}
	verdict, verr := readVerdict(result)
	if verr != nil {
		res.err = verr.Error()
		return res
	}
	res.verdict = verdict
	// A SETTLED DEBATE HAS A CHAIR-SITTING COUNT (#649). This read used to be permissive — a
	// missing or differently-typed `epochs` left the count at 0 and went into the histogram as a
	// debate that ran no epochs — one field over from the verdict, which #639 made a refusal.
	// Same rule now: the result names the keys it did carry, so a renamed field reads as a
	// renamed field and not as a run that never sat.
	n, ok := result["epochs"].(int64)
	if !ok {
		res.err = fmt.Sprintf("the settled result carries no integer epochs (got %T) — a settled debate has a chair-sitting count; keys carried: %v", result["epochs"], sortedKeys(result))
		return res
	}
	res.epochs = int(n)
	if term, ok := result["termination"].(map[string]any); ok {
		if why, ok := term["why"].([]any); ok && len(why) > 0 {
			res.why = fmt.Sprint(why[0])
		}
		res.epochLimit, _ = term["epoch_limit_reached"].(bool)
	}
	if r.passBlockersErr != "" {
		res.err = "pass_permitted oracle: the gate refused a PASS the plan permitted, and the blocker list could not be read to say whose blocker held it: " + r.passBlockersErr
		return res
	}
	if r.inquiryRulingErr != "" {
		res.err = "inquiry drive: an inquiry's ruling could not be read, so the drive could not tell a ruled line from an unruled one: " + r.inquiryRulingErr
		return res
	}
	if r.inquiryIDErr != "" {
		res.err = "inquiry drive: " + r.inquiryIDErr
		return res
	}
	if r.quoteAdmittedErr != "" {
		res.err = "quote the report does not hold: " + r.quoteAdmittedErr
		return res
	}
	if r.bridgeAdmittedErr != "" {
		res.err = "bridge: " + r.bridgeAdmittedErr
		return res
	}
	if r.orderAdmittedErr != "" {
		res.err = "section order: " + r.orderAdmittedErr
		return res
	}
	if r.spanAdmittedErr != "" {
		res.err = "source span: " + r.spanAdmittedErr
		return res
	}
	if r.repeatErr != "" {
		res.err = "a repeated act: " + r.repeatErr
		return res
	}
	if r.rulerFilingErr != "" {
		res.err = "a ruler's filing: " + r.rulerFilingErr
		return res
	}
	if r.answerBudgetErr != "" {
		res.err = "the answering mint: " + r.answerBudgetErr
		return res
	}
	if r.passOverAnotherSeat != "" {
		res.err = "pass_permitted oracle: the plan permitted a PASS while " + r.passOverAnotherSeat + " stood, and the gate refused it: " + r.lastPassRefusal
		return res
	}
	halted, _ := result["halted"].(bool)
	if o := r.petitionOracle(halted); o != "" {
		res.err = o
		return res
	}
	if r.lastPassRefusal != "" && res.verdict != "VERIFIED" {
		res.why = "PASS refused: " + r.lastPassRefusal
		if r.lastRuleRefusal != "" {
			res.why += " ‖ last ruling refused: " + r.lastRuleRefusal
		}
	}
	// Oracle #111 (map-free): both tiers were configured haiku, so every dispatched seat must have
	// carried model "haiku" — none unset, none on another tier. This needs no bulk-seat list; the
	// SEAT_CLASS<->dispatch binding is proved separately by debate-dispatch.test.mjs.
	if len(r.models) == 0 {
		res.err = "no agent() call carried a model — the resolver regressed to unset"
		return res
	}
	for _, mdl := range r.models {
		if mdl != "haiku" {
			res.err = "a seat dispatched on tier " + mdl + ", not the configured haiku"
			return res
		}
	}
	// Oracle 1: the record the run left must pass verify.
	if out, err := tracked(bin, "verify", "--seat-id", "operator", "--run", runDir); err != nil {
		res.err = "verify FAILED:\n" + truncate(string(out))
		return res
	}
	// Oracle 1b: the JSON views the operator side reads in json-mode must exit 0 and parse.
	// `board` is already exercised above; `findings` and the operator's log read are JSON by name; `debate --json`
	// is the structured debate the capture audits count sections from. A broken view is what
	// would silently blank a dashboard tile or make an audit read an empty transcript.
	ids := mintedGapIDs(stageRun)
	// EVERY PROJECTION, ON EVERY ROLE. `show` became a GROUP (0.56.0), so each projection is its
	// own command path — and the coverage gate immediately reported 34 paths never invoked,
	// because the sweep had only ever driven the handful it asserted on. A view that only the
	// merge reads is a view nobody checks from the seat that actually reads it.
	// EVERY SEAT READS EVERY PROJECTION, and the seat id is what says which tree `show` is in.
	//
	// THROUGH `drive`, WHICH SPAWNS THE BINARY ONCE PER PATH AND THEN STOPS. What varies between
	// runs here is the RECORD each projection meets, never the argv — measured at N=4, every one
	// of these 48 paths fired at exactly 1-4 invocations per run, the same set on every seed. So
	// the shape variation these oracles are built on is kept in full and the 76 process spawns a
	// run paid for it are not; `drive`'s own comment carries the accounting.
	for role, sid := range map[string]string{
		"blue": "blue-respond", "lens": "red-lens-evidence", "chair": "red-chair", "bench": "judge",
	} {
		// EACH VIEW UNDER ITS OWN GROUP, FOR THE ROLES THAT CARRY IT. This drove everything as
		// `show` for all four roles, and when the bench's raw-record projections moved to `inquest`
		// every role started failing on them — 40 of 40 runs, with the per-verb coverage gates
		// reporting ZERO events downstream because no run survived to be tallied. The group and the
		// membership are both ASKED rather than held: GroupOf reads the view table, and roleHasGroup
		// reads the real command tree, so neither can go stale the way a literal "show" did.
		for _, v := range viewNamesForFuzz {
			group := seat.GroupOf(v)
			if !roleHasGroup(role, group) {
				continue // this role does not carry the group this view lives under
			}
			markViewDriven(v)
			args := []string{group, v, "--run", runDir, "--seat-id", sid}
			if v == "changes" && len(ids) > 0 {
				args = append(args, "--id", ids[0])
			}
			if _, err := drive(bin, args...); err != nil {
				res.err = role + " " + group + " " + v + " failed: " + err.Error()
				return res
			}
			// AND BOTH OF ITS SELECTORS. Thirty-four flags reached the surface with no end-to-end
			// drive at all, which this sweep's own flag gate named: a flag no run ever passes is code
			// no run has executed. They are seat-facing, so an exemption would have been a claim that
			// nobody checks them.
			//
			// THE ASSERTION IS AN INVARIANT, NOT A FIXTURE, because the sweep reaches an arbitrary run
			// shape and there is no expected entry list to compare against. A regex of one dot matches
			// every character, so it must keep everything the unselected read had; a quote nothing can
			// contain must keep nothing. The counts come from the tool's own `selection` block rather
			// than from counting the output here, so this reads the answer the SEAT reads.
			//
			// ON THE --json FORM, WHICH IS WHAT THE PAIR SELECTS. A markdown form refuses it, and on a
			// view that is JSON by name the flag is the same bytes; so every view that carries the pair
			// answers with a selection block, and one that does not is a failure rather than a skip. A
			// view whose page carries no pair is not driven with one.
			if !viewSelects(role, group, v) {
				continue
			}
			for _, sel := range [][]string{{"--match", "."}, {"--quote", impossibleQuote}} {
				sargs := append([]string{group, v, "--run", runDir, "--seat-id", sid, "--json"}, sel...)
				out, err := drive(bin, sargs...)
				if err != nil {
					res.err = role + " " + group + " " + v + " " + strings.Join(sel, " ") + " failed: " + err.Error()
					return res
				}
				var doc struct {
					Selection *struct {
						Matched int `json:"matched"`
						Of      int `json:"of"`
					} `json:"selection"`
				}
				if json.Unmarshal([]byte(strings.TrimSpace(string(out))), &doc) != nil || doc.Selection == nil {
					res.err = fmt.Sprintf("%s %s %s --json %s answered with no selection block — the pair it carries did not select:\n%s",
						role, group, v, strings.Join(sel, " "), out)
					return res
				}
				// THE REPORT'S `of` IS ITS LINE COUNT, blank lines included, and a selection never keeps a
				// blank line — so there everything means every line with text on it.
				want := doc.Selection.Of
				if v == "report" {
					n, err := reportTextLines(bin, group, runDir, sid)
					if err != nil {
						res.err = role + " " + group + " report --json failed: " + err.Error()
						return res
					}
					want = n
				}
				if sel[0] == "--match" && doc.Selection.Matched != want {
					res.err = fmt.Sprintf("%s %s %s --match . kept %d of %d — a pattern matching every character dropped entries",
						role, group, v, doc.Selection.Matched, want)
					return res
				}
				if sel[0] == "--quote" && doc.Selection.Matched != 0 {
					res.err = fmt.Sprintf("%s %s %s --quote %q kept %d of %d — a quote nothing contains matched something",
						role, group, v, impossibleQuote, doc.Selection.Matched, doc.Selection.Of)
					return res
				}
			}
		}
		if _, err := drive(bin, "show", "--run", runDir, "--seat-id", sid); err != nil {
			res.err = role + " show (bare, the seat's pending work) failed: " + err.Error()
			return res
		}
	}
	// A DECLARED GAP READS AS THE BOARD STATES IT, on every role's list. The fuzz bench writes
	// "remains open" whatever becomes of the gap, so a status read from the prose, or off a read the
	// board does not make, disagrees here on whatever shape the sweep reached.
	if msg := declarationsAgreeWithBoard(stageRun, r.declared); msg != "" {
		res.err = msg
		return res
	}
	// READING THE REPORT AT AN ANCHOR, over the anchors this run actually minted.
	//
	// The unit tests drive a fixture; this drives whatever shape the sweep reached — a report
	// whose anchors sit at the top, at the bottom, adjacent to each other, or inside a section
	// that a later edit rewrote around them. The window's line arithmetic is where an
	// off-by-one lives, and a fixture with the anchor comfortably in the middle never asks.
	if a := someReportAnchor(stageRun); a != "" {
		// EVERY ROLE, BOTH FLAGS. `show report` is defined once in internal/cli/seat, so a
		// window that worked on the chair and not on blue would mean the projection had grown
		// a per-role surface — and --window 0 is the degenerate size that must still resolve to
		// the anchored line rather than to nothing.
		// A SEAT ID PER ROLE, because the tree is scoped to the dispatched identity: without one
		// the binary builds the OPERATOR surface, where `show` does not exist at all.
		for _, sid := range map[string]string{
			"blue": "blue-respond", "lens": "red-lens-evidence", "chair": "red-chair", "bench": "judge",
		} {
			for _, extra := range [][]string{nil, {"--window", "0"}} {
				args := append([]string{"show", "report", "--anchor", a, "--run", runDir, "--seat-id", sid}, extra...)
				out, err := drive(bin, args...)
				if err != nil {
					res.err = strings.Join(args, " ") + " failed:\n" + truncate(string(out))
					return res
				}
				// THE WINDOW MUST CONTAIN ITS OWN ANCHOR. An exit-0 window that does not is
				// the plausible zero: a read that reports success and shows the wrong place.
				if !strings.Contains(string(out), a) {
					res.err = strings.Join(args, " ") + " returned a window WITHOUT the anchor it was addressed by:\n" + truncate(string(out))
					return res
				}
			}
		}
	}
	// THE ASSEMBLED LAYOUT, for every role and in both forms, over whatever record the sweep reached.
	// The layout assembles the set and slices it by the table's rows, so each claim it makes is
	// checked against the other: a section it says ships carries text, one it says does not carries
	// none, and the report's own text holds no section the tool composes — the write refuses one,
	// and a generator here that wrote one would have its ingest or edit refused.
	for _, sid := range []string{"blue-respond", "red-lens-evidence", "red-chair", "judge"} {
		if out, err := drive(bin, "show", "report", "--assembled", "--run", runDir, "--seat-id", sid); err != nil || !strings.Contains(string(out), "## Risk matrix") {
			res.err = sid + " show report --assembled failed, or lists no risk matrix:\n" + truncate(string(out))
			return res
		}
		out, err := drive(bin, "show", "report", "--assembled", "--json", "--run", runDir, "--seat-id", sid)
		var layout reportdoc.LayoutJSON
		if err != nil || json.Unmarshal(out, &layout) != nil || len(layout.Sections) != len(reportdoc.Sections()) {
			res.err = sid + " show report --assembled --json is not the layout, a row for each section of the table:\n" + truncate(string(out))
			return res
		}
		for _, sec := range layout.Sections {
			if sec.Ships != (sec.Text != "") || sec.Path == "" || sec.From == "" {
				res.err = fmt.Sprintf("%s show report --assembled --json: section %q ships=%v with %d byte(s) of text, from %q, path %q",
					sid, sec.Heading, sec.Ships, len(sec.Text), sec.From, sec.Path)
				return res
			}
		}
	}
	if text, err := drive(bin, "show", "report", "--run", runDir, "--seat-id", "blue-respond"); err == nil {
		if line, _, found := reportdoc.ComposedIn("", string(text)); found {
			res.err = "the report's text carries " + line + ", a section the tool composes, past the refusal at the write"
			return res
		}
		if label, _, found := reportdoc.WovenNoteIn("", string(text)); found {
			res.err = "the report's text carries the footnote label " + label + ", one the weave writes, past the refusal at the write"
			return res
		}
		if m, found := reportdoc.OutOfOrderIn("", string(text)); found {
			res.err = "the report's text carries " + m.Line + " out of the order it ships in, past the refusal at the write"
			return res
		}
	}
	// A PROOF READ AT ITS ANCHOR CARRIES THE STORED SCRIPT — the read a lens makes before it
	// judges one. Every role, for the reason above: the projection is defined once. The script is
	// compared to the proof store's own bytes, because an exit-0 read with an empty `script` is a
	// lens judging a computation it was not shown.
	if anchor, sha := someProof(stageRun); anchor != "" {
		for _, sid := range []string{"blue-respond", "red-lens-evidence", "red-chair", "judge"} {
			args := []string{"show", "evidence", "--anchor", anchor, "--run", runDir, "--seat-id", sid}
			out, err := drive(bin, args...)
			if err != nil {
				res.err = strings.Join(args, " ") + " failed:\n" + truncate(string(out))
				return res
			}
			var at record.EvidenceAtJSON
			if jerr := json.Unmarshal(out, &at); jerr != nil || at.Proof == nil {
				res.err = strings.Join(args, " ") + " did not return the proof at its anchor:\n" + truncate(string(out))
				return res
			}
			stored, rerr := os.ReadFile(filepath.Join(runDir, "proofs", sha, at.Proof.ScriptFile))
			if rerr != nil || at.Proof.ScriptFile == "" || at.Proof.Script != string(stored) {
				res.err = fmt.Sprintf("%s returned a script that is not the proof store's (%s, read error %v):\n%s",
					strings.Join(args, " "), at.Proof.ScriptFile, rerr, truncate(string(out)))
				return res
			}
		}
		if out, err := drive(bin, "show", "evidence", "--anchor", "P-ffffffff", "--run", runDir, "--seat-id", "red-lens-evidence"); err == nil {
			res.err = "show evidence --anchor P-ffffffff SUCCEEDED on an anchor nobody placed:\n" + truncate(string(out))
			return res
		}
	}
	// AN ANCHOR NOBODY MINTED IS REFUSED, NOT READ EMPTY — the read-side twin of `show changes
	// --id G1`. An empty window says "the report has nothing here", which is a different
	// fact from "that anchor is not in this report".
	if out, err := drive(bin, "show", "report", "--anchor", unmintedFinding, "--run", runDir, "--seat-id", "blue-respond"); err == nil {
		res.err = "show report --anchor " + unmintedFinding + " SUCCEEDED on an anchor nobody minted — a window over nothing:\n" + truncate(string(out))
		return res
	} else if refusedByShape(out, unmintedFinding) {
		res.err = "show report --anchor " + unmintedFinding + " was refused for its SHAPE, so the refusal of an anchor nobody minted was never reached:\n" + truncate(string(out))
		return res
	}
	// The OPERATOR's read of the log — seats write it, the human reads it back.
	//
	// UNDER `ops`, NOT `show`. The operator's reads moved into a group of their own, and this call
	// kept the old word — the same stale-caller shape as the inquest views above, from the same
	// reorganisation. `log` is not a seat projection, so GroupOf does not answer for it: the group
	// is spelled once, here, and the `operator log read` failure names it if it moves again.
	if _, err := tracked(bin, "ops", "log", "--run", runDir, "--seat-id", "operator"); err != nil {
		res.err = "operator log read failed: " + err.Error()
		return res
	}
	// THE WRONG-ADDRESS DRIVE IS GONE, because the address collision is. A seat's `friction
	// --reason` used to land on the OPERATOR's read, which takes no --reason, so it died in the
	// parser before any message could teach — and the channel it could not reach is the one for
	// reporting exactly that. The two frictions are on different trees now: a seat's carries its
	// write verb and not the read, the operator's the read and no seat verbs. There is nothing
	// left to land on by accident, so there is nothing here to drive.
	// `friction` left the SEAT menu (0.57.0) — it is the operator's read. The verb stays on
	// every role; only the view moved.
	for _, v := range []string{"findings"} {
		out, err := drive(bin, "show", v, "--run", runDir, "--seat-id", "red-chair")
		var parsed any
		if err != nil || json.Unmarshal([]byte(strings.TrimSpace(string(out))), &parsed) != nil {
			res.err = "show " + v + " did not return valid JSON:\n" + truncate(string(out))
			return res
		}
	}
	{
		out, err := drive(bin, seat.GroupOf("debate"), "debate", "--json", "--run", runDir, "--seat-id", "red-chair")
		var parsed any
		if err != nil || json.Unmarshal([]byte(strings.TrimSpace(string(out))), &parsed) != nil {
			res.err = "show debate --json did not return valid JSON:\n" + truncate(string(out))
			return res
		}
	}
	// Oracle 1c: the markdown views now render in-memory (view.Markdown) with no render-shadow
	// round-trip. Each must exit 0 on whatever state the run reached — this restores the randomized
	// coverage the `render` verb + the difftest RENDERS section used to give the projection renderer,
	// which the render-shadow removal (#203) took away (view_test.go pins the bytes on fixed
	// fixtures; only the fuzz drives them across arbitrary run shapes).
	//
	// DRIVEN FROM cli.ViewNames(), not a list maintained here. A hand-kept roster of what
	// exists is how the coverage gap this oracle closes was created in the first place: a
	// view ships, nobody adds it to the list, and the sweep reports full coverage of a
	// surface it never drove. The JSON-by-name views take their own dispatch branches above.
	// EVERY ROLE'S show, not merge's alone. Each role has a different DEFAULT view and its own
	// role gate, so driving only `chair show --view` left the other three reachable but never
	// reached — and --id (the scoped form) never passed at all.
	for _, role := range []string{"blue", "lens", "bench", "chair"} {
		group := seat.GroupOf("debate")
		if !roleHasGroup(role, group) {
			continue
		}
		if out, err := drive(bin, group, "debate", "--run", runDir, "--seat-id", seatOfRole(role)); err != nil {
			res.err = role + " " + group + " debate failed:\n" + truncate(string(out))
			return res
		}
	}
	if ids := mintedGapIDs(stageRun); len(ids) > 0 {
		for _, role := range []string{"blue", "lens", "bench"} {
			if out, err := drive(bin, "show", "changes", "--id", ids[0], "--run", runDir, "--seat-id", seatOfRole(role)); err != nil {
				res.err = role + " show changes --id " + ids[0] + " failed:\n" + truncate(string(out))
				return res
			}
		}
	}
	// THE SET IS DERIVED, because this exclusion was four names written here by hand and two of
	// them were wrong in opposite directions: `friction` stopped being a view entirely (it is
	// the OPERATOR's read of the channel, never a projection), so the case excluded nothing;
	// and `motions`, `telemetry` and `evidence` are JSON by name and were NOT excluded, so this
	// loop drove three projections down "the markdown path" while the comment said their own
	// oracles had them. An exclusion list that names a view which does not exist, and misses
	// half the ones it is about, reads as a considered boundary and is neither.
	jsonByName := map[string]bool{}
	for _, v := range cli.JSONByNameViews() {
		jsonByName[v] = true
	}
	for _, v := range cli.ViewNames() {
		if jsonByName[v] {
			continue // JSON by name — driven by their own oracles, not the markdown path
		}
		if out, err := drive(bin, seat.GroupOf(v), v, "--run", runDir, "--seat-id", "red-chair"); err != nil {
			res.err = seat.GroupOf(v) + " " + v + " (projection) failed:\n" + truncate(string(out))
			return res
		}
	}
	// And the SCOPED form, over a gap the run actually minted: `--view changes --id <gap>`
	// takes a different path (it resolves the gap and renders the comparison), so the unscoped
	// render above proves nothing about it. A gap the board does not know must be REFUSED, not
	// rendered empty — the read-side twin of requireGap.
	if ids := mintedGapIDs(stageRun); len(ids) > 0 {
		if out, err := drive(bin, "show", "changes", "--id", ids[0], "--run", runDir, "--seat-id", "red-chair"); err != nil {
			res.err = "show changes --id " + ids[0] + " failed:\n" + truncate(string(out))
			return res
		}
		if out, err := drive(bin, "show", "changes", "--id", unmintedGap, "--run", runDir, "--seat-id", "red-chair"); err == nil {
			res.err = "show changes --id " + unmintedGap + " SUCCEEDED on a gap nobody minted — a view that invents a comparison:\n" + truncate(string(out))
			return res
		} else if refusedByShape(out, unmintedGap) {
			res.err = "show changes --id " + unmintedGap + " was refused for its SHAPE, so the refusal of a gap nobody minted was never reached:\n" + truncate(string(out))
			return res
		}
	}
	// Oracle 1d: THE READ-ONLY CENSUS. Every record-nothing surface must survive whatever
	// shape this run reached. These emit no events, so the verb-coverage gate below is blind
	// to them by construction — this is the only thing that exercises them at all.
	if msg := sweepReadOnly(bin, runDir); msg != "" {
		res.err = msg
		return res
	}
	// SCENARIO ORACLE: the record must SHOW the decision the scenario specified.
	//
	// This is what the directives buy. Every other oracle here asks "did anything break" —
	// verify passes, the views parse, prose renders. None can tell a chain that CARRIED a
	// decision from one that dropped it, because until now no decision was written down before
	// it was taken. Each gap is now minted with the outcome it should reach, so the terminal
	// record is checkable rather than merely well-formed.
	//
	// IGNORE is the sharpest: if --answers stopped recording provenance, an ignored gap and a
	// repaired one would look identical — and that join is what #267's whole measurement axis
	// is built on.
	// ONE READ OF THE RECORD SERVES EVERY ORACLE BELOW, AND A READ THAT FAILS FAILS THE RUN. Nothing
	// from here on writes, so the board each oracle judges is the same board. An oracle that runs
	// only when its own read succeeds reports a record it cannot read as a record with nothing
	// wrong in it.
	oracleRun := runtest.Open(t, runDir)
	// THE DOCUMENTS ARE RE-ASSEMBLED FROM THE RECORD AT THE CLOSE, as capture does. The probes above
	// read the run after its outcome, and a read the tool refuses then is logged — the record takes
	// friction after the outcome (record.requireOpenToSeats) — so the bench's documents predate
	// entries the documents that ship carry. The oracles judge those. Like capture, it re-assembles
	// only a run the bench assembled: the bench's act is what makes a run's documents exist.
	if written, werr := reportdoc.AssemblyWritten(oracleRun); werr != nil {
		res.err = "documents: whether the bench assembled them could not be read: " + werr.Error()
		return res
	} else if written {
		if _, aerr := reportdoc.Assemble(oracleRun); aerr != nil {
			res.err = "documents: re-assembling the run at its close failed: " + aerr.Error()
			return res
		}
	}
	board, berr := record.FamilyOf(oracleRun)
	if berr != nil {
		res.err = "board: the oracles could not read the record: " + berr.Error()
		return res
	}
	{
		answered, disputed, proved := map[string]bool{}, map[string]bool{}, map[string]bool{}
		// THE SWITCH IS ON THE BODY, not on a type string beside it. Each arm reaches straight for
		// a field, so binding the message and the type in one step removes the pair that could
		// disagree — and a grade motion's gap id now comes off the GradeMotion arm rather than
		// from a payload key that petitions and directions also had.
		for _, e := range board.Events {
			body, ok := recordpb.Body(e)
			if !ok {
				continue
			}
			switch t := body.(type) {
			case *recordpb.Proof:
				if t.GetAnswers() != "" {
					proved[t.GetAnswers()] = true
				}
			case *recordpb.BlueEdit:
				if t.GetAnswers() != "" {
					answered[t.GetAnswers()] = true
				}
			case *recordpb.Motion:
				if g := t.GetGrade(); g != nil && g.GetGapId() != "" {
					disputed[g.GetGapId()] = true
				}
			}
		}
		for _, g := range board.Gaps {
			id := g.ID
			// Only gaps a responder was SHOWN can be judged: one minted in the terminal round
			// never reaches blue, so its scenario was never dispatched.
			if g == nil || g.Mint == nil || !r.presented.has(id) {
				continue
			}
			switch g.Mint.GetRequiredFix() {
			case dirIgnore:
				if answered[id] {
					res.err = "scenario IGNORE: " + id + " was answered by a blue_edit, but the scenario said blue does nothing — either a writer ignored the directive or --answers is attaching provenance nobody claimed"
					return res
				}
			case dirDisputeWon, dirDisputeLost:
				if !disputed[id] {
					res.err = "scenario DISPUTE: " + id + " has no dispute event — the contest the scenario specified is absent from the record"
					if why := r.disputeRefusals[id]; why != "" {
						res.err += "\nthe filing was REFUSED, and this is what it said:\n  " + why
					} else {
						res.err += "\nno refusal was recorded for it either, so the drive never reached `motion grade file` at all"
					}
					return res
				}
			case dirProve, dirProveDrifts:
				// The chain must be VISIBLE ON THE RECORD, not merely to have happened: a proof
				// answering this gap, and the gap closed only behind a reproduction that held.
				if !proved[id] {
					res.err = "scenario PROVE: " + id + " has no proof answering it — the computation the scenario specified left no artifact red could re-run"
					return res
				}
				// A gap ONLY the bench closed is not red accepting anything: the terminal bench rules
				// a docket standing at the exit, whatever its scenario had still to show. The test is
				// red's own closing body, not ClosedByBench — that flag follows the last closing
				// event, so a ruling on a gap red had already closed would hide red's closure here.
				if r.evaluated.has(id) && !g.Open && g.Closure != nil && !r.reproduced.has(id) {
					res.err = "scenario PROVE: " + id + " was CLOSED without its proof reproducing — red accepted a computation on somebody's word, which is the one thing re-running exists to prevent"
					return res
				}
			case dirApply:
				if !answered[id] {
					res.err = "scenario APPLY: " + id + " has no blue_edit answering it — a repair the scenario specified left no provenance"
					return res
				}
			}
			// THE TERMINAL FATE, which is the assertion the first pass was missing. Found by
			// probe: making red never close a repaired gap produced ZERO oracle failures — the
			// run simply ended CEILING with work sitting done-but-open, and nothing said so.
			// Scoped to gaps red actually sat on after blue answered: a repair landed in the
			// final round leaves the gap legitimately open, because red never sits again.
			if satisfied(g.Mint.GetRequiredFix()) && r.evaluated.has(id) && g.Open {
				res.err = "scenario " + g.Mint.GetRequiredFix() + ": " + id + " is still OPEN after red sat on the repaired board — work was done and the board never recorded it as finished"
				return res
			}
		}
	}

	// A RULING HAS A CONSEQUENCE, AND A CONTEST IS VISIBLE AS ONE.
	//
	// Red's ruling and the inquiry's fate are both on the record, and without a join blue taking up
	// an inquiry red called out-of-scope would look exactly like taking up one red endorsed. A
	// ruling is an ARGUMENT, not a command: it binds no fate, and what blue does about one it
	// disagrees with is the fate it records. The oracle joins the two — the ruling the view states
	// and the status the fold reads — against the fate each inquiry was proposed to take.
	if msg := inquiryRulingOracle(oracleRun, board); msg != "" {
		res.err = msg
		return res
	}

	// EVERY RETIREMENT IS EVIDENCED. The fake retires only text a recorded edit removed, so an
	// `asserted` retire means either the fake regressed to phantom retirements or the
	// edit-tracking that evidences them broke. A phantom retire cancels real claim loss in the
	// scorecard's additive-integrity detector, so it must not pass unnoticed here either. A
	// citation retired from a sentence that stays (claim_kept) removes no claim, so it carries no
	// removal basis, and one that names a basis says a claim left that did not.
	{
		for _, e := range board.Events {
			r, ok := recordpb.BodyAs[*recordpb.Retire](e)
			if ok && r.GetClaimKept() {
				if r.GetRemovalBasis() != "" {
					res.err = "a citation retired from a sentence that stays recorded removal_basis=" + r.GetRemovalBasis() + " — its claim did not leave the report"
					return res
				}
				continue
			}
			if ok && r.GetRemovalBasis() != record.RemovalVerified {
				res.err = "a retire recorded removal_basis=" + r.GetRemovalBasis() +
					" — nothing on the record shows that claim was ever in the report, and an unevidenced retirement cancels real claim loss in the additive-integrity detector"
				return res
			}
		}
	}

	// AN ASSERTED VERDICT IS EITHER UNVERIFIED OR A BROKEN DERIVATION, and this is the tripwire
	// that keeps the second from hiding behind the first.
	//
	// The tool refuses an --as that contradicts the record (#308), so a recorded verdict is
	// either DERIVED or came from the one case the record cannot decide: a run that stopped
	// before the record reached a terminal state, which is what UNVERIFIED names. The tool now
	// refuses any other word over such a record, so an asserted basis on any verdict but
	// UNVERIFIED means the derivation stopped working — or the refusal did.
	//
	// THIS USED TO TURN ON `ended: deadlock`, the bench's run-level judgement; deadlock is per
	// gap under the roundless record and the field is retired, so the sanctioned case is the
	// verdict word itself.
	{
		for _, e := range board.Events {
			o, ok := recordpb.BodyAs[*recordpb.Outcome](e)
			if !ok {
				continue
			}
			switch o.GetVerdictBasis() {
			case record.VerdictDerived:
			case "":
				res.err = "the outcome event carries no verdict_basis — the field that says whether the verdict was computed or claimed has gone missing"
				return res
			default:
				if o.GetVerdict() == recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED {
					continue
				}
				res.err = "the run recorded an ASSERTED verdict (" + recordpb.Word(o.GetVerdict()) + ") that is not UNVERIFIED — the only word the bench may assert is the one the record cannot derive a verdict for, and without it an asserted verdict means the derivation stopped working"
				return res
			}
		}
	}

	// VERDICT ORACLE: the recorded outcome must agree with the board it describes.
	//
	// The terminal verdict is the single most consequential value the engine emits, and until
	// now the fake DREW IT FROM A HAT — so `bench outcome --as` could record VERIFIED over a
	// board with open gaps and nothing anywhere would notice. The assembler is now told the
	// verdict the way a real seat is (debate.js states it in the prompt), which makes the
	// agreement between outcome and board checkable for the first time.
	{
		openCount := 0
		for _, gOrd := range board.Gaps {
			id := gOrd.ID
			if g := board.Gap(id); g != nil && g.Open {
				openCount++
			}
		}
		// A PASS HOLDS OVER WORK THAT IS NOT MATERIAL: a gap that is not material — by its class, or
		// graded below medium — does not hold the gate, so "any open gap" is the wrong count — the record's own dispatch
		// plan says whether the board permitted the PASS, and that is the oracle: a VERIFIED the
		// board does not permit means the refusal in `verdict` stopped firing.
		plan, perr := record.PlanDispatch(oracleRun)
		if perr != nil {
			res.err = "verdict oracle: the dispatch plan could not be read, so whether the board permits a PASS is unknown: " + perr.Error()
			return res
		}
		permitted := plan.PassPermitted
		recorded := ""
		for _, e := range board.Events {
			if o, ok := recordpb.BodyAs[*recordpb.Outcome](e); ok {
				// UPPERCASE, because that is the word this oracle compares against and the
				// schema's spelling is lowercase. The fold is stated here rather than left to
				// a reader to notice.
				recorded = strings.ToUpper(recordpb.Word(o.GetVerdict()))
			}
		}
		// EVERY RUN ENDS WITH AN OUTCOME ON THE RECORD, AND IT IS THE ENGINE'S WORD. The assembly's
		// `bench outcome` is refused when the record derives another verdict or a bench-owned
		// motion stands, and this driver discards that refusal: a run that ended with no outcome
		// event was reported only when a later oracle happened to trip over the unstamped report.
		if recorded != verdict {
			res.err = fmt.Sprintf("outcome oracle: debate.js ended the run %s and the record's outcome is %q — the assembly's `bench outcome` was refused, or recorded another word", verdict, recorded)
			return res
		}
		if recorded == "VERIFIED" && !permitted {
			res.err = fmt.Sprintf("verdict oracle: the run recorded VERIFIED over a board that does not permit a PASS (%d gap(s) open, material or not) — a pass over unfinished work", openCount)
			return res
		}
		// NO PASS OVER AN ANCHOR NOBODY CHECKED, OR ONE WHOSE CHECK FAILED — the record's own answer,
		// read at the run's end, not a hand copy of the rule.
		if recorded == "VERIFIED" {
			bs, berr := record.PassBlockers(oracleRun)
			if berr != nil {
				res.err = "anchor oracle: the PASS blockers could not be read at the run's end: " + berr.Error()
				return res
			}
			for _, b := range bs {
				if b.Kind == record.BlockerUnchecked || b.Kind == record.BlockerFailedCheck {
					res.err = fmt.Sprintf("anchor oracle: the run recorded VERIFIED while %s stands %s — a pass over a citation or proof no check has read", b.Subject, b.Kind)
					return res
				}
			}
		}
		// THE CONVERSE NEEDS A CARVE-OUT, and finding out why is the point of running it.
		//
		// It fired 2 of 60 on "CEILING with every gap closed", which looked like a verdict
		// contradicting its board — and is not. The BENCH closes gaps at the terminal sitting,
		// AFTER the ceiling has already been determined: the verdict describes the debate's
		// exit, not the post-adjudication board. That is why Gap.ClosedByBench exists.
		//
		// So the check is only sound over gaps RED closed. A non-passing verdict whose board
		// was cleaned entirely by the bench is correct; one whose board red itself emptied is
		// the contradiction worth catching — and only where no blocker another seat owns stands
		// (unfinishedOverRedsWork).
		benchClosed := 0
		for _, gOrd := range board.Gaps {
			if g := board.Gap(gOrd.ID); g != nil && g.ClosedByBench {
				benchClosed++
			}
		}
		if unfinishedOverRedsWork(recorded, len(board.Gaps), openCount, benchClosed, plan.Blockers) {
			res.err = "verdict oracle: the run recorded " + recorded + " with every gap closed BY RED — an unfinished verdict over work red itself finished"
			return res
		}
	}

	// Oracle 2: every dialectic event's prose must actually RENDER in the report — the A1-A3
	// class (prose written under one key, read under another) is invisible to verify but caught
	// here, on every run.
	if m := proseRenders(t, board, runDir); m != "" {
		res.err = m
	}
	if res.err == "" {
		if m := basisRenders(t, board, runDir); m != "" {
			res.err = m
		}
	}
	res.dialectic = tallyDialectic(board)
	// #256: count the citation axis's two real artifacts (see outcome). A cite that fetched and
	// anchored leaves BOTH; a cite event alone leaves neither. The report is the record projection
	// now (#709), so the anchors are counted in the render, not a file.
	md, mErr := reportproj.RenderFromRecord(oracleRun)
	if mErr != nil && res.err == "" {
		res.err = "citation axis: the report could not be rendered, so its cite anchors are uncounted: " + mErr.Error()
	}
	res.citeAnchors = strings.Count(md, "<!--cite:")
	if fam, err := record.FamilyOf(oracleRun); err == nil {
		for _, e := range fam.Events {
			if rp, ok := recordpb.BodyAs[*recordpb.Reproduce](e); ok {
				res.reproduces++
				if rp.GetRerun() != recordpb.Rerun_RERUN_NOT_RUN {
					res.reruns++
				}
			}
		}
	}
	// A CITATION OR PROOF IS OFFERED TO ITS OWNER ALONE, read at the run's end when no seat is writing.
	if msg, proofs, citations := r.anchorOffersAgree(oracleRun); msg != "" && res.err == "" {
		res.err = msg
	} else {
		res.ownerOffered, res.citationsOffered = proofs, citations
	}
	r.mu.Lock()
	res.ownerOffered += r.offeredToOwner["proof"]
	res.citationsOffered += r.offeredToOwner["citation"]
	if r.anchorOfferErr != "" && res.err == "" {
		res.err = r.anchorOfferErr
	}
	res.inquiryRecords = r.inquiryRecords
	res.closingsFiled, res.closingsRefused, res.closingWaits, res.positionsWithheld = r.closingsFiled, r.closingsRefused, r.closingWaits, r.positionsWithheld
	res.ordersRefused = r.ordersRefused
	res.spansRefused = r.spansRefused
	res.copiesRead = r.copiesRead
	res.fragmentsDriven, res.truncationsDriven = r.fragmentsDriven, r.truncationsDriven
	res.quotesRefused = r.quotesRefused
	res.bridgesRefused, res.bridgesNamed = r.bridgesRefused, r.bridgesNamed
	res.reopenedReadied, res.reopenedPlans = r.reopenedReadied, r.reopenedPlans
	res.proofReopenedReadied, res.owedKept, res.proofOwedKept, res.owedKeptBy = r.proofReopenedReadied, r.owedKept, r.proofOwedKept, r.owedKeptBy
	res.unreadAnswers, res.unreadRepeats, res.unreadRetires, res.failedCloseRefused = r.unreadAnswers, r.unreadRepeats, r.unreadRetires, r.failedCloseRefused
	if r.unsoundReadyErr != "" && res.err == "" {
		res.err = r.unsoundReadyErr
	}
	if r.reopenedOwnerErr != "" && res.err == "" {
		res.err = r.reopenedOwnerErr
	}
	if r.forceReopenedProof && r.proofReopenedReadied[proofOwner] == 0 && res.err == "" {
		stage := r.reopenStage
		if stage == "" {
			stage = "nothing: no blue sitting reached the prove"
		}
		res.err = fmt.Sprintf("forced reopened proof: no plan readied %s for a reopened proof; the chain reached %s", proofOwner, stage)
	}
	if r.forceUnsoundProof && r.unsoundReadied == 0 && res.err == "" {
		stage := r.unsoundStage
		if stage == "" {
			stage = "nothing: no blue sitting reached the prove"
		}
		res.err = fmt.Sprintf("forced unsound proof: no plan readied %s or blue for an unsound proof; the chain reached %s", proofOwner, stage)
	}
	if r.forceAbsentSpan && r.spansRefused == 0 && res.err == "" {
		res.err = "forced absent span: no cite of a span its source does not hold was refused"
		if !r.absentSpanDriven {
			res.err += "; no blue sitting reached the cite"
		}
	}
	res.repeatsRefused = r.repeatsRefused
	res.lensSittings = r.lensSittings
	res.blueSittings, res.blueAgain = r.blueSittings, r.blueAgain
	res.budgetRefused = r.budgetRefused
	res.ownAnswers = r.ownAnswers
	res.unsoundReadied, res.unsoundPlans, res.unsoundAnswers = r.unsoundReadied, r.unsoundPlans, r.unsoundAnswers
	res.unsoundRepeats, res.unsoundOverOpen = r.unsoundRepeats, r.unsoundOverOpen
	res.unclassifiedMinted, res.unclassifiedRefused = r.unclassifiedMinted, r.unclassifiedRefused
	res.rulerFilingsRefused = r.rulerFilingsRefused
	res.closedRemandsRefused, res.trifleRemandsRouted = r.closedRemandsRefused, r.trifleRemandsRouted
	res.blueRemandActs, res.lensRemandActs, res.pastLimitRemandsRefused = r.remandActs["blue"], r.remandActs["lens"], r.pastLimitRemandsRefused
	r.mu.Unlock()
	// NO POSITION ON THE RECORD IS FROM A SEAT BUT THE CHAIR, AND NO CLOSING FROM A SEAT THAT IS NO
	// PARTY — read off the record, so a refusal that logged and wrote the act anyway fails here.
	minter := map[string]string{}
	for _, e := range board.Events {
		if m := e.GetMint(); m != nil {
			minter[m.GetGapId()] = e.GetSeatId()
		}
	}
	for _, e := range board.Events {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_POSITION && e.GetSeatId() != "red-chair" && res.err == "" {
			res.err = "position duty: the record holds a position by " + e.GetSeatId() + ", and the position is the chair's"
		}
		if c := e.GetClosing(); c != nil && !record.IsClosingParty(e.GetSeatId(), minter[c.GetGapId()]) && res.err == "" {
			res.err = "closing duty: the record holds a closing on " + c.GetGapId() + " by " + e.GetSeatId() + ", party to no dispute on it"
		}
	}
	for _, e := range board.Events {
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		switch t := body.(type) {
		case *recordpb.BlueEdit:
			if t.GetAnswers() != "" {
				res.editAnswers++
			}
			if t.GetAppliedVerbatim() {
				res.verbatimApplied++
				if !t.GetAccepted() {
					res.typedVerbatim++
				}
			}
		case *recordpb.Mint:
			if t.GetFixBasis() == "verified" {
				res.verifiedBasis++
			}
			if len(t.GetFoundBy()) > 0 {
				res.foundByMints++
			}
		case *recordpb.Log:
			// SOURCE IS CHECKED, NOT JUST TYPE. An estoppel entry a SEAT filed is a seat claiming
			// the tool refused it; only the tool's own is evidence the guard ran.
			if t.GetType() == recordpb.LogType_LOG_TYPE_ESTOPPEL &&
				t.GetSource() == recordpb.LogSource_LOG_SOURCE_TOOL {
				res.estoppels++
			}
		}
	}
	// A run that fetched nothing has no cache directory, and that is the honest zero. Any other
	// failure to list it is not.
	ents, cerr := os.ReadDir(filepath.Join(runDir, "cache"))
	if cerr != nil && !os.IsNotExist(cerr) && res.err == "" {
		res.err = "citation axis: the source cache could not be listed, so its files are uncounted: " + cerr.Error()
	}
	for _, e := range ents {
		if !e.IsDir() && e.Name() != "index" {
			res.cacheFiles++
		}
	}
	return res
}

// inquiryRulingOracle holds every ruled inquiry to its consequence, and returns the first breach
// or "". It joins two independent reads: the Go fold's inquiries (status, question) and the record's
// own inquiry_state view (the ruling), through record.InquiryRuling.
//
// A RULING IT CANNOT READ IS A BREACH, NOT AN UNRULED INQUIRY. "" from InquiryRuling means red has
// not ruled, and that inquiry has nothing to check; an error means the check did not happen, and
// passing over it would report an unaudited inquiry as a clean one.
func inquiryRulingOracle(run record.Run, board record.Family) string {
	for _, a := range record.InquiriesOf(board.Events, board.At) {
		ruling, err := record.InquiryRuling(run, a.ID)
		if err != nil {
			return "inquiry " + a.ID + ": its ruling could not be read, so whether the ruling had its consequence is unchecked: " + err.Error()
		}
		if ruling == "" || a.Status == "proposed" {
			continue // never ruled, or blue has not answered yet
		}
		switch {
		case strings.HasPrefix(a.Question, avContest):
			if a.Status != "underway" {
				return "inquiry " + a.ID + " was proposed as CONTESTED but ended " + a.Status + " — blue was to pursue it against the ruling"
			}
		case ruling == "endorsed":
			if a.Status != "underway" {
				return "inquiry " + a.ID + " was ENDORSED and ended " + a.Status + " — a ruling with no consequence"
			}
		default:
			if a.Status == "underway" {
				return "inquiry " + a.ID + " was ruled " + ruling + " and underway anyway, on a line the drive declines"
			}
		}
	}
	return ""
}

// #111: debate.js must REFUSE to dispatch when either model tier is unset — the engine never
// guesses or inherits a tier. The guard throws at the top of the async IIFE, so __result rejects
// with the flag-named message. No binary needed: the throw precedes any tool act.
func TestDispatchRefusesUnsetModel(t *testing.T) {
	wrapped := debateWrapped(t)
	cases := []struct {
		name, want string
		args       map[string]any
	}{
		{"model unset", "refusing dispatch — model unset", map[string]any{"topic": "x", "runDir": "research/2026-01-01_t", "lanes": 1, "laneFloorOverride": "t", "judgmentModel": "haiku"}},
		{"judgmentModel unset", "refusing dispatch — judgmentModel unset", map[string]any{"topic": "x", "runDir": "research/2026-01-01_t", "lanes": 1, "laneFloorOverride": "t", "model": "haiku"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			loop := eventloop.NewEventLoop()
			var got string
			loop.Run(func(vm *goja.Runtime) {
				vm.Set("args", c.args)
				vm.Set("agent", func(goja.FunctionCall) goja.Value { return goja.Undefined() })
				if _, err := vm.RunString(preamble); err != nil {
					t.Fatalf("preamble: %v", err)
				}
				if _, err := vm.RunString(wrapped); err != nil {
					got = "run: " + err.Error()
				}
			})
			loop.Run(func(vm *goja.Runtime) {
				if v := vm.Get("__result"); v != nil {
					if pr, ok := v.Export().(*goja.Promise); ok && pr.State() == goja.PromiseStateRejected {
						got = pr.Result().String()
					}
				}
			})
			if !strings.Contains(got, c.want) {
				t.Errorf("expected rejection containing %q, got %q", c.want, got)
			}
		})
	}
}

// verbsWithEvents are the record-event-emitting verbs the fuzz is meant to drive — the whole seat
// surface bar the read-only ones (register/show emit no board event). The coverage gate
// asserts each fired at least once across the run set, so a regression that silently STOPS
// emitting one fails loudly (the old gate guarded only cite/finding). `halt` terminates the run
// and is covered by TestFuzzHaltPath, not the random sweep — the gate skips it (see coverExempt).
var verbsWithEvents = []string{
	"closing", "position", "regrade", "mint", "close",
	"cite", "verify", "finding", "inquiry", "reproduce", "log", "retire",
	"manifest_row", "verdict", "spot_check", "certify", "declare", "halt",
	// friction-none is the EXPLICIT NEGATIVE arm of the friction verb — a distinct event type,
	// so a gate listing only "friction" would report the channel covered while the arm that
	// makes an empty log meaningful went undriven.
	"log",
	// The motion collapse (#344): filed by any seat, ruled by one.
	"motion", "motion_rule",
	// Added 2026-08-04 by a census of every type record.Append can write: these two were
	// APPENDABLE BUT UNGATED, so a regression that stopped emitting either would have left the
	// sweep green. `anchor` is the finding-marker's own record (the immortal-marker detector's
	// EXPECTED set is exactly these), `outcome` is the bench's.
	"blue_edit", "anchor", "outcome", "proof",
	// base_ingest is the frozen round-0 report (#709), written by `blue ingest` at the start of
	// every run (runOne) and so REQUIRED by the coverage gate like any other event-emitting verb.
	"base_ingest",
	// The sitting span (#265), and the first types here that NO VERB WRITES: they are appended by
	// the SubagentStart/SubagentStop hook binaries, about a seat, and the goja harness execs the
	// binary directly without a harness to fire hooks. Listed so the census is complete and
	// exempted below from the random sweep, which cannot reach them.
	"sitting_open", "sitting_close",
	// A sitting stopped at the tool-call limit: written by feov-sitting-write when the PreToolUse
	// hook first refuses a call past the limit. No verb writes it either.
	"sitting_limit",
	// The remaining schema types, named so the census below has a complete list to check against.
	"closing", "inquiry_review", "register",
	// The cast and the dispatch (plans/roundless.md §III.B.1). `cast` is written by setup, before
	// any seat sits; `dispatch` by the chair's `dispatch next`. Exempted below until the workflow
	// the fuzz drives runs the dispatch loop (roundless B-iii), when `dispatch` joins the sweep. `plan`
	// is the plan `dispatch next` stores beside its rows.
	"cast", "dispatch", "plan",
	// The tool's refusal of an act on a gap that moved after the seat's sitting opened, recorded as
	// the seat's read point. Lenses sit concurrently in the sweep, so a seed reaches it.
	"stale",
	// The whole report served to a lens: its pin. A lens sitting reads the report whole most of the
	// time (the red-lens arm of the dispatch), so every seed reaches it.
	"report_read",
}

// coverExempt names verbs tallied but NOT required in the random-sweep coverage gate.
var coverExempt = map[string]bool{
	"halt": true, // terminal — covered by TestFuzzHaltPath
	// THE SITTING SPAN IS WRITTEN BY HOOKS, NOT BY VERBS. These two have a real writer the FUZZ
	// cannot invoke — the harness execs the binary with no client to fire SubagentStart or
	// SubagentStop, so no seed could ever produce them however long it ran.
	//
	// They are driven, just not here: internal/hookcmd/sitting_test.go asserts both ends land,
	// that a main-agent turn end does not, and that neither hook writes to stdout. Named rather
	// than dropped, so "the sweep does not cover this" stays a line somebody reads.
	"sitting_open":  true,
	"sitting_close": true,
	// Written by the hook-spawned writer when a sitting passes the tool-call limit, which no seed
	// reaches. Driven in internal/hookcmd/limit_test.go (the handoff) and
	// internal/sittingwrite/write_test.go (the event).
	"sitting_limit": true,
	// THE CAST IS SETUP'S, NOT A VERB'S: written once before the first seat sits, so no seed of
	// the random sweep can produce it. `dispatch` is driven: the chair's first act every sitting.
	"cast": true,
}

// TestFuzzHaltPath drives the JUDICIAL HALT terminal path — kept OUT of the random sweep because a
// halt ends the run and reshapes every downstream oracle. blue-synthesize files a petition
// (forceHalt guarantees it); the judge-petition sitting rules HALT via the `halt` verb; debate.js
// sets halted, breaks the loop, and the run ends HALTED. Asserts: the run settles as HALTED, a
// halt event is on the record, and the halted run STILL passes verify (a safety exit is a valid
// record, not a broken one). This is the dedicated coverage for `halt`, which the sweep exempts.
func TestFuzzHaltPath(t *testing.T) {
	releaseGate(t)
	bin := buildBinary(t)
	wrapped := debateWrapped(t)
	runDir, err := os.MkdirTemp("", "fuzz-halt-")
	if err != nil {
		// Not swallowed: an empty runDir resolves every path below against the package
		// directory, which turns a full TMPDIR into a failure that names anything but the disk.
		t.Fatalf("temp dir for the run: %v", err)
	}
	// RELEASE, THEN REMOVE — in that order, and in one defer so the order is the order it reads
	// in. Two defers express it backwards (LIFO), which is the same "the two lines matter in the
	// opposite order to how they read" trap that has now been walked into eleven times. The sweep
	// in runOne releases; these dedicated-path tests copied its MkdirTemp/RemoveAll shape without
	// it, and on Linux the removal succeeds anyway so nothing said a word (#666).
	defer func() {
		_ = recordsql.CloseUnder(runDir)
		_ = os.RemoveAll(runDir)
	}()
	// THE PETITION REACHES THE BENCH THROUGH THE CHAIR'S PLAN, so this run needs what `dispatch
	// next` reads: the cast setup writes, and the report blue freezes. Without them the chair's
	// dispatch is refused ("the record holds no cast"), the plan is empty, and the run ends
	// UNVERIFIED with the petition never heard.
	if err := writeSetup(filepath.Join(runDir, "blue", "report.md"), []byte(seedReport)); err != nil {
		t.Fatal(err)
	}
	r := newRunner(bin, runDir, newLockedRand(1))
	haltCast, haltLanes := record.CastFor(3)
	if _, err := record.Append(record.Identity{Run: r.run(), SeatID: record.HarnessSeat}, &recordpb.Cast{SeatIds: haltCast, LaneSeatIds: haltLanes}); err != nil {
		t.Fatalf("write the cast: %v", err)
	}
	r.register("blue", "blue-synthesize")
	if _, err := r.exec("ingest", "--seat-id", "blue-synthesize"); err != nil {
		t.Fatalf("ingest the round-0 report: %v", err)
	}
	r.forceHalt = true

	result, settledErr := driveDebate(r, wrapped)
	if settledErr != "" {
		t.Fatalf("halt run did not settle cleanly: %s", settledErr)
	}
	if v, _ := result["verdict"].(string); v != "HALTED" {
		t.Fatalf("expected verdict HALTED, got %q", v)
	}
	if h, _ := result["halted"].(bool); !h {
		t.Fatal("forceHalt run did not end halted — the judicial-halt terminal path is unexercised")
	}
	board, err := record.FamilyOf(runtest.Open(t, runDir))
	if err != nil {
		t.Fatalf("board: %v", err)
	}
	halts := 0
	for _, e := range board.Events {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_HALT {
			halts++
		}
	}
	if halts == 0 {
		t.Fatal("no halt event on the record — the halt verb never ran")
	}
	if out, err := tracked(bin, "verify", "--seat-id", "operator", "--run", runDir); err != nil {
		t.Fatalf("verify FAILED on the halted run (a safety exit must still be a valid record):\n%s", truncate(string(out)))
	}
}

// tallyDialectic counts the events that prove the fuzz exercised the paths it claims to.
func tallyDialectic(board record.Family) map[string]int {
	want := map[string]bool{}
	for _, v := range verbsWithEvents {
		want[v] = true
	}
	m := map[string]int{}
	for _, e := range board.Events {
		// The type is an enum now; the tally is keyed on the WORD, which is what verbsWithEvents
		// carries and what the failure message names.
		w := recordpb.Word(e.GetType())
		if want[w] {
			m[w]++
		}
	}
	return m
}

// THIS MAP IS NOT AN ALLOWLIST. A type absent from an allowlist is not checked, and the absence
// is silent: every event type added later defaults to "the report need not render this".
//
// It is EXHAUSTIVE over the types a run actually produces: every event type seen on the
// record must appear here or in reportExemptions with a stated reason, and an unclassified type
// FAILS. A new verb cannot be added without deciding, in writing, what its reader gets.
var dialecticProseKey = map[string]string{
	// The debate proper.
	// THE FIELD NAMES, NOT THE FLAG NAMES. Every entry here said `reason` because that is what a
	// seat TYPES; the schema spells the prose field per verb — a closing stores `text`, an opinion
	// `rationale`, a regrade `basis`. Read against a payload map the wrong name simply never
	// matched, so this gate was inert for most of the types it claims to cover, reporting "the
	// report renders every recorded prose" over rules that could not fire. fieldStr makes the miss
	// loud, which is how these were found.
	"position": "text", "closing": "text",
	// Motions: the ASK and the ANSWER, on one id. `petition-rule` alone used to be the half-dialog — the
	// reader got the bench's ruling with no question attached.
	// The board.
	"mint": "problem", "finding": "text", "regrade": "basis",
	// Inquiries: the question, red's ruling on it, and blue's reason for its fate.
	"inquiry": "question",
	// Red's per-round verdict that the REPORT still carries the inquiry. It renders ON THE
	// INQUIRY'S OWN ROW in the three research areas rather than in a section of its own, because the claim it
	// answers ("we pursued X") lives there and a verdict a reader has to go and find is a verdict
	// most readers do not find. `reason` is what red read at that inquiry — the quoted text, not the
	// grade, which is the half a reader can check.

	// The lens's below-the-bar work and the fate the chair gave it.
	// Substance leaving the report, on the record, with its reason.
	"retire": "claim",
	// Run-level voices.
	"halt": "opinion", "certify": "statement", "declare": "holding",
	// The log: what a seat told the operator. It renders in run.md, beside the seats that sat
	// and filed nothing.
	"log": "text",
	// Blue's self-audit receipt, one per repaired gap. `row` is what blue checked and what
	// checking it showed — the receipt reached no reader for a year, because the coverage metric
	// counted the ENVELOPE array and the verb was named in no prompt at all (#318).
	"manifest_row": "row",
	// The motion group. `basis` is the ASK, `opinion` the answer — both render in the report's
	// one Motions section, joined on the motion id (#344).
	"motion":      "basis",
	"motion_rule": "opinion",
	// Red re-reading its own closure archive. The prose is what the sample FOUND — the whole
	// point of sampling — and it reached no reader at all until the floor was enforced (#317).
	"spot_check": "reason",
}

// reportExemptions are event types whose prose is deliberately NOT expected in the report, each
// with its reason. Stated rather than omitted: an absence with no reason is indistinguishable
// from an oversight, which is precisely how this gate decayed.
var reportExemptions = map[string]string{
	"register":    "a seat announcing itself to the run — attribution machinery, and the attribution reaches the reader on every act that seat records, never as an entry of its own",
	"anchor":      "an estoppel key spliced INTO blue/report.md — it is machinery for the edit path, and the text it anchors is the report's own",
	"blue_edit":   "mutates blue/report.md, which assembly ships as written; the edit's effect IS in the report, and rendering the old/new spans again would duplicate the document",
	"base_ingest": "the synthesized report frozen into the record (#709) — its text IS blue's report, the base every projection renders from, which assembly ships as written; rendering it as an entry of its own would duplicate the whole document",
	// THE DISPATCH MACHINERY (plans/roundless.md §III.B.1). A cast names who may sit; a dispatch
	// records who the chair relayed for one epoch. Both are attribution and routing, and both
	// reach the reader through what the dispatched seats then record — never as entries of
	// their own; capture's dispatch-parity audit is their reader.
	"cast":        "the run's roster of seats — routing, read by the register refusal and the dispatch plan, never rendered",
	"dispatch":    "who the chair relayed for one epoch — routing, audited by capture against who sat, never rendered",
	"plan":        "the plan the chair's dispatch printed, stored whole — routing, the original capture compares each relay with, never rendered",
	"stale":       "the read point a refusal moved a seat to on a gap that changed under its sitting — the write path's own gate, carrying no prose; the act re-filed carries the seat's",
	"report_read": "the whole report the tool served a lens, at a head — the lens's pin, carrying no prose; the transcript renders it as a served line, and the reader gets the report itself",
	// Red's independent re-run. The NOTE is its judgement; whether it reproduced is computed
	// by the tool and rendered beside the proof either way (#343).
	"reproduce": "reason",
	"verify":    "red adjudicating ONE citation. It reaches the reader through the evidence view rather than the report body: the report carries BLUE's citations (woven into the bibliography), and red's verdict on them is audit provenance rather than a claim the reader acts on. It IS surfaced — attached to the source it names in `show evidence`, in the board's citations count, and, for `refutes` and `absent`, as an assembly-screen FAILURE if the report still cites what red found against",
	"cite":      "resolved rather than rendered — the anchor becomes a visible [^N] with its note, and the source a ## Bibliography line (weaveCitations)",
	"proof":     "resolved rather than rendered — weaveProofs splices the computation at its anchor",
	"close":     "the closure's prose is red's acceptance argument and reaches the reader only as an index row today; rendering it in full is tracked, not silently accepted",
	// The review of the report against the inquiries on the record. Its READER IS A GATE, not the
	// document: record.InquiryReviewDueOf asks whether a review is recorded since the report or an
	// inquiry last changed, and the PASS is refused until it is. Rendering it in the report would put a process check
	// where the debate goes — the review says "I read the report against the record", and what
	// a reader wants is the outcome of that reading, which arrives as an ordinary GAP when the
	// treatment falls short. Named here rather than left silent, because the gate is right that
	// an unclassified event type is how a report loses a whole exchange.
	"inquiry_review": "read by record.InquiryReviewDueOf as a PASS gate owed when the report or an inquiry changes, not rendered: a shortfall the review finds is minted as an ordinary gap, which is what reaches the reader",
	"outcome":        "composed into the outcome stamp by outcomeStamp, from the payload's verdict/deadlocked/exhausted fields rather than a prose field",
	"verdict":        "red's per-round PASS/FAIL, consumed by DeriveVerdict into the terminal outcome; the round-by-round spine is not yet a transcript section",
}

// basisFields are the DERIVED-NOT-ASSERTED fields, mapped to the event that carries each and a
// phrase the report must contain when that value is on the record.
//
// Every one of these exists because a seat asked to self-report reports the flattering value —
// so each is computed by the tool at the write. Each then gated something (estoppel, the outcome
// cross-check, the claim-loss detector) and reached the reader as NOTHING, which collapsed the
// distinction at the only point where a human could act on it: a verdict the record itself
// decided printed the same word as one the bench merely asserted.
//
// prose keys cannot cover this — a basis is a field on an event whose OTHER prose renders fine,
// so the prose gate is green while the qualifier is gone.
var basisFields = []struct{ evType, key, value, want string }{
	{"outcome", "verdict_basis", "derived", "derived from the record"},
	{"outcome", "verdict_basis", "asserted", "asserted by the bench"},
	{"mint", "fix_basis", "verified", "with the text in front of it"},
	{"mint", "fix_basis", "proposed", "nothing checked this demand"},
	{"retire", "removal_basis", "verified", "the record shows it leaving"},
	{"retire", "removal_basis", "asserted", "nothing on the record shows it was ever present"},
	// proof_basis was already rendered before this gate existed (weaveProofs prints it with an
	// explanation of reproducible vs observed). It is here so the coverage is stated rather
	// than assumed — the one that was fine is the easiest to break unnoticed.
	{"proof", "proof_basis", "reproducible", "reproducible"},
	{"proof", "proof_basis", "observed", "observed"},
}

func basisRenders(t *testing.T, board record.Family, runDir string) string {
	rpt, err := assembledSet(runDir)
	if err != nil {
		return err.Error()
	}
	var missing []string
	for _, bf := range basisFields {
		present := false
		for _, e := range board.Events {
			// A CLOSED gap's mint is still on the record but its fix_basis is no longer shown —
			// the demand was met, and the report's closure index is a one-line index by design.
			// Only OPEN gaps are checked, which is where the qualifier changes what a reader does.
			if recordpb.Word(e.GetType()) != bf.evType {
				continue
			}
			if fieldStr(t, e, bf.key) == bf.value {
				if bf.evType == "mint" {
					m, _ := recordpb.BodyAs[*recordpb.Mint](e)
					if !gapIsOpen(board, m.GetGapId()) {
						continue
					}
				}
				present = true
				break
			}
		}
		if present && !strings.Contains(rpt, bf.want) {
			missing = append(missing, fmt.Sprintf("%s.%s=%s is on the record and the report never says so (looked for %q)", bf.evType, bf.key, bf.value, bf.want))
		}
	}
	if len(missing) > 0 {
		return "derived-basis absent from report — the qualifier the field exists to preserve:\n" + strings.Join(missing, "\n")
	}
	return ""
}

func gapIsOpen(board record.Family, id string) bool {
	g := board.Gap(id)
	return g != nil && g.Open
}

// assembledSet is EVERY document the assembler wrote, concatenated.
//
// The deliverable is a set, and "did this prose reach the reader" is a question about the set:
// a log entry lands in run.md, a motion in judgments.md, a withdrawn claim in CHANGELOG.md.
// An oracle that reads report.md alone measures one document of the artifact and reports the
// others as clean — which is the exact failure shape these gates exist to catch, one level up.
func assembledSet(runDir string) (string, error) {
	var b strings.Builder
	for _, name := range reportdoc.Files() {
		body, err := os.ReadFile(filepath.Join(runDir, name))
		if err != nil {
			continue
		}
		b.Write(body)
		b.WriteString("\n")
	}
	if b.Len() == 0 {
		return "", fmt.Errorf("no assembled document in %s", runDir)
	}
	return b.String(), nil
}

func proseRenders(t *testing.T, board record.Family, runDir string) string {
	rpt, err := assembledSet(runDir)
	if err != nil {
		return err.Error()
	}
	var missing, unclassified []string
	seen := map[string]bool{}
	for _, e := range board.Events {
		w := recordpb.Word(e.GetType())
		key, ok := dialecticProseKey[w]
		if !ok {
			// An event type in neither table is a DECISION NOBODY MADE. Registering it here
			// costs one line; leaving it unclassified is how four exchanges reached the record
			// and never the reader.
			if reportExemptions[w] == "" && !seen[w] {
				seen[w] = true
				unclassified = append(unclassified, w)
			}
			continue
		}
		// The reader gets the prose as assembly ships it: the tokens of every kind it strips are
		// gone — a gap anchor a seat quoted into its retire or a prescription carried is not prose.
		prose := strings.TrimSpace(anchor.StripAssembled(fieldStr(t, e, key)))
		if prose == "" || strings.Contains(rpt, prose) {
			continue
		}
		missing = append(missing, fmt.Sprintf("%s/%s prose absent from report: %q", e.GetSeatId(), w, prose))
		if len(missing) >= 5 {
			break
		}
	}
	if len(unclassified) > 0 {
		sort.Strings(unclassified)
		return "event type(s) on the record with NO entry in dialecticProseKey and NO stated exemption: " + strings.Join(unclassified, ", ") +
			"\nDecide what the reader gets from each — render it (add the prose key) or say why not (add the exemption). Silence here is how the report loses a whole exchange."
	}
	if len(missing) > 0 {
		return "prose-not-rendered (A1-A3 class):\n" + strings.Join(missing, "\n")
	}
	return ""
}

// motionIDOf pulls the tool-assigned motion id out of a filing's JSON envelope.
func motionIDOf(out string) string {
	var env struct {
		Result struct {
			MotionID string `json:"motion_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		return ""
	}
	return env.Result.MotionID
}

// claimOf is the claim a --json verify result names, or "".
// contradictionOf is the id of the contradiction a reading's --json result names, and whether it
// names one.
func contradictionOf(out string) (string, bool) {
	var env struct {
		Result struct {
			Contradiction *struct {
				ID string `json:"id"`
			} `json:"contradiction"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil || env.Result.Contradiction == nil {
		return "", false
	}
	return env.Result.Contradiction.ID, true
}

func claimOf(out string) string {
	var env struct {
		Result struct {
			Claim string `json:"claim"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		return ""
	}
	return env.Result.Claim
}

func binDir(bin string) string { return filepath.ToSlash(filepath.Dir(bin)) }

func truncate(s string) string {
	if len(s) > 600 {
		return s[:600] + "…"
	}
	return s
}

// viewNamesForFuzz is the projection set, taken from the command tree so a view added or removed
// is swept without anyone remembering to edit a list here.
var viewNamesForFuzz = seat.ViewNames()

// roleHasGroup asks the REAL command tree whether this role carries this command group.
//
// IT IS A QUERY AND NOT A COPY OF THE RULE. Which roles hold `inquest` is decided in
// seat.RoleVerbs — the seats that adjudicate, which is the chair as well as the bench — and a
// literal `role == "bench" || role == "chair"` here would be that rule written twice, going stale
// the next time it moves. The rule has already moved once, and a hardcoded "show" at the call sites
// is exactly what failed when it did.
// reportTextLines is how many of the report's lines carry text, read off its unselected --json form.
func reportTextLines(bin, group, runDir, sid string) (int, error) {
	out, err := drive(bin, group, "report", "--run", runDir, "--seat-id", sid, "--json")
	if err != nil {
		return 0, err
	}
	var doc struct {
		Lines []struct {
			Text string `json:"text"`
		} `json:"lines"`
	}
	if err := json.Unmarshal(out, &doc); err != nil {
		return 0, err
	}
	n := 0
	for _, ln := range doc.Lines {
		if strings.TrimSpace(ln.Text) != "" {
			n++
		}
	}
	return n, nil
}

// viewSelects reports whether the view's page, in this role's tree, carries the selector pair.
func viewSelects(role, group, view string) bool {
	for _, c := range seat.RoleVerbs(role) {
		if c.Name() != group {
			continue
		}
		for _, sub := range c.Commands() {
			if sub.Name() == view {
				return sub.Flags().Lookup("match") != nil
			}
		}
	}
	return false
}

func roleHasGroup(role, group string) bool {
	for _, c := range seat.RoleVerbs(role) {
		if c.Name() == group {
			return true
		}
	}
	return false
}

// drivenView records which projections the per-role sweep actually drove.
//
// THE SKIP NEEDS A WITNESS. The sweep now skips a view whose group a role does not carry, and a
// skip is how this oracle's own comment says coverage gaps are born: "a view ships, nobody adds it
// to the list, and the sweep reports full coverage of a surface it never drove." So a view that NO
// role drove fails the sweep rather than passing quietly.
//
// IT IS WRITTEN FROM EVERY WORKER, so it is guarded. The sweep runs FUZZ_C runs at once and each
// one drives every view its roles carry, which made this the one piece of cross-run state on a hot
// path — and an unguarded map written by twelve goroutines is not a flake, it is a crash the
// runtime takes when it happens to notice: `fatal error: concurrent map writes`, killing the gate
// mid-sweep after 400 seconds with no failing assertion to read. Reproduce with
// `FEOV_RELEASE_GATE=1 FUZZ_N=6 FUZZ_C=6 go test -race -run TestFuzzDebate ./releasegate/fuzz/`.
//
// The READ is after wg.Wait, so only the writes ever contended.
var (
	drivenViewMu sync.Mutex
	drivenView   = map[string]bool{}
)

// markViewDriven records that some role drove this projection.
func markViewDriven(v string) {
	drivenViewMu.Lock()
	drivenView[v] = true
	drivenViewMu.Unlock()
}

// surfaceQuorum is the run count at or above which the coverage gates can hold the sweep to the
// FULL surface. Below it a low-frequency path can flake to zero and fail an honest run.
//
// MEASURED, not guessed: across a default 60-run sweep the scarcest path fired
// 9 times (0.15/run), which gives ~10% odds of never firing at N=15 against
// ~0.25% at N=40. The number is shape-dependent — lane-driven verbs scale with `lanes`, dispute-
// driven ones with `maxRounds` — so it is a floor for THIS shape, and a shape change re-tunes it.
// That is why the gates below assert the tally rather than trusting this constant to stand in
// for coverage (#630).
const surfaceQuorum = 40

func TestFuzzDebate(t *testing.T) {
	releaseGate(t)
	bin := buildBinary(t)
	wrapped := debateWrapped(t)

	// The DEFAULT is a modest smoke that proves the harness and catches gross regressions in
	// ~15s on Linux — which is what the Linux legs, main pushes, tag builds and the nightly
	// all run. Windows PULL REQUESTS run -short (hooks.yml): each run shells the real binary
	// ~50-70 times, and Windows pays process-spawn cost on every one, which made this
	// package that leg's critical path (573s, measured 2026-08-25) after the RAM disk took
	// I/O out of the bill. What a run proves about the PLATFORM — the .exe build, the
	// spawns, the temp plumbing — run 15 proves as well as run 60; depth past that is
	// statistics on debate semantics, which are platform-independent and keep full depth on
	// Linux. The full 1000-run confidence sweep is on demand:
	// FUZZ_N=1000 go test ./releasegate/fuzz -run TestFuzzDebate -timeout 1200s.
	// THE DEFAULT IS THE QUORUM, so the default sweep is always one that can assert the surface.
	// 60 was a number the gates did not depend on; at production shape a run costs ~2.2x what the
	// smoke shape cost, and the sweep needs exactly enough runs for the coverage gates to hold.
	mootPending.Store(true)
	engageRefusalPending.Store(true)
	mootSpent.Store(false)
	resetClassDrive()

	n := surfaceQuorum
	if testing.Short() {
		n = 15
	}
	if v := os.Getenv("FUZZ_N"); v != "" {
		if k, err := strconv.Atoi(v); err == nil && k > 0 {
			n = k
		}
	}
	// Each run is process-spawn-bound (it shells the real binary ~50-70 times), not CPU-bound,
	// so the goroutines mostly WAIT on subprocesses — oversubscribe past core count to keep every
	// core busy spawning. FUZZ_C overrides.
	concurrency := runtime.NumCPU() * 3
	if v := os.Getenv("FUZZ_C"); v != "" {
		if k, err := strconv.Atoi(v); err == nil && k > 0 {
			concurrency = k
		}
	}
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var failures []outcome
	var completed int
	verdicts := map[string]int{}
	blueSits, blueAgains := 0, 0 // blue-respond sittings across the sweep, and those on an engaged successor (R2)
	epochHist := map[int]int{}
	// limitExits is the runs the epoch limit ended: CEILING by the run's term. limitForced is
	// whether limitSeed's run was one of them.
	limitExits, limitForced := 0, false
	whyHist := map[string]int{} // verdict → first stated reason, with the gap id folded
	dcov := map[string]int{}
	citeAnchors, cacheFiles := 0, 0                         // dialectic-event coverage across all runs (proves the fuzz emits them)
	editAnswers := 0                                        // #267: blue_edit events that carried the provenance key
	forcedVerdict, forcedWhy := "", ""                      // what the FORCED-UNVERIFIED seed actually ended as
	verifiedVerdict, verifiedWhy := "", ""                  // and the FORCED-VERIFIED seed
	verifiedBasis := 0                                      // #267 stage 3: gaps whose fix_basis was EARNED by a validated pair
	foundByMints := 0                                       // mints accepted with a found_by naming a recorded finding
	verbatimApplied := 0                                    // #267 stage 4: edits that applied red's proposal exactly (the estoppel precondition)
	typedVerbatim := 0                                      // those blue typed rather than accepted
	estoppels := 0                                          // the TOOL's own refusals of a mint against text blue applied verbatim
	inquiryRecords := 0                                     // fates accepted on an inquiry id read from a propose result
	quotesRefused := 0                                      // corroborations refused for a quote the report does not hold
	bridgesRefused, bridgesNamed := 0, 0                    // supports_with_bridge refused for no --bridge / recorded with one
	ordersRefused := 0                                      // edits refused for a section out of the order the report ships in
	spansRefused := 0                                       // cites refused for a source span the source's text does not hold
	forcedSpansRefused := 0                                 // of them, absentSpanSeed's run's
	copiesRead := map[string]map[string]int{}               // copy words each reading verb named
	closingsFiled, closingsRefused := 0, 0                  // closings filed on a docketed gap, and envelopes refused while one stood
	closingWaits := 0                                       // plans that kept the bench waiting on a closing
	positionsWithheld := 0                                  // chair sittings that withheld a permitted PASS and argued why
	fragmentsDriven := 0                                    // cites driven to be refused for a span starting mid-clause
	truncationsDriven := 0                                  // cites driven to be refused for a span ending mid-clause
	reopenedReadied, reopenedPlans := 0, 0                  // plans that readied every reopened blocker's owner, of those holding one
	proofReopenedReadied := map[string]int{}                // reopened proofs whose owner a plan readied, by owner
	owedKeptBy := map[string]int{}                          // those folds, by each kind of debt they held
	owedKept, proofOwedKept := 0, 0                         // lenses a plan kept from retiring by a debt; of them, by an unjudged proof
	unreadAnswers, unreadRepeats := 0, 0                    // gaps minted about a citation a lens's reading could not read; repeat unread readings refused
	unreadRetires, failedCloseRefused := 0, 0               // blue's retires of such a citation; closes refused over a standing failed check
	ownerOfferedRuns := 0                                   // runs whose proof owner was offered an unjudged proof, at a sitting or the run's end
	citationOfferedRuns := 0                                // runs whose evidence lens was offered an unverified citation, at a sitting or the run's end
	repeatsRefused := 0                                     // second acts refused in a sitting that holds one, naming the answering act
	budgetRefused := 0                                      // mints refused a lens holding its whole mint budget open
	ownAnswers := 0                                         // mints that answered the minting lens's own raise
	unsoundReadied, unsoundPlans, unsoundAnswers := 0, 0, 0 // plans readying for an unsound proof, of those holding one; gaps carrying one
	forcedUnsoundReadied := 0                               // of them, unsoundProofSeed's run's
	unsoundRepeats, unsoundOverOpen := 0, 0                 // repeat unsound judgements refused; gaps minted over an open own gap about the proof
	unclassifiedMinted, unclassifiedRefused := 0, 0         // catch-all mints landed; refused for want of a reason
	rulerFilingsRefused := 0                                // filings refused the seat that rules their subject
	closedRemandsRefused, trifleRemandsRouted := 0, 0       // remands of a closed gap refused; remands of a gap not material routed
	blueRemandActs, lensRemandActs := 0, 0                  // acts on a remanded gap its blocking item held blue and the lens to
	pastLimitRemandsRefused := 0                            // remands past the limit refused naming the closing dispositions
	applyMisses := map[string]int{}                         // and why it did not, by cause — a bare 0 above named none of them
	estoppelMisses := map[string]int{}                      // and why the estoppel drive declined, for the same reason

	for i := 0; i < n; i++ {
		seed := int64(i) + 1
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			began := time.Now()
			o := runOne(t, wrapped, bin, seed, seed == unverifiedSeed, seed == estoppelSeed, seed == verifiedSeed)
			// ONE LINE PER RUN, so a slow or stuck seed is named with its length beside it.
			t.Logf("RUNSTAT seed=%d secs=%d epochs=%d verdict=%q limit=%t failed=%t reproduces=%d reruns=%d owedKept=%d lensSittings=%d maxLensSittings=%d bySeat=%v blueSittings=%d blueAgain=%d unsoundReadied=%d/%d spansRefused=%d closedRemandsRefused=%d trifleRemandsRouted=%d err=%q", seed, int(time.Since(began).Seconds()), o.epochs, o.verdict, o.epochLimit, o.err != "", o.reproduces, o.reruns, o.owedKept, sumSittings(o.lensSittings), maxSittings(o.lensSittings), o.lensSittings, o.blueSittings, o.blueAgain, o.unsoundReadied, o.unsoundPlans, o.spansRefused, o.closedRemandsRefused, o.trifleRemandsRouted, firstLine(o.err))
			// THE ORACLE RUNS ON EVERY RECORD THE SWEEP PRODUCES. The drives above assert that
			// each command SUCCEEDED; the oracle asserts that the record those commands built is
			// one every projection agrees about — the cross-reader class the unit suites cannot
			// see, because each exercises one reader against its own fixture. A violation keeps
			// the run directory like any other failure, so the disagreement can be inspected.
			if o.err == "" {
				if violations, cerr := consistency.Check(runtest.Open(t, o.runDir)); cerr != nil {
					o.err = "consistency oracle: " + cerr.Error()
				} else if len(violations) > 0 {
					// THE EVIDENCE TRAVELS WITH THE VIOLATION, because the run directory does
					// not. "A violation keeps the run directory so the disagreement can be
					// inspected" is true on a developer's machine and false everywhere this
					// actually fires: #645 was seen once in ~160 sweeps, inside a CI job, and by
					// the time anyone read the log the directory was gone with the container.
					// A rare violation whose evidence is unreachable can only ever be reported,
					// never diagnosed — so the anchor layer's two sides go into the failure text
					// itself, where the log keeps them.
					o.err = "consistency violations:\n  " + strings.Join(violations, "\n  ") + anchorEvidence(o.runDir)
				}
			}
			// The oracle opened this run's cached handle in-process; release it or the
			// RemoveAll below fails on Windows and the sweep leaks one handle per seed.
			_ = recordsql.CloseUnder(o.runDir)
			mu.Lock()
			completed++
			verdicts[o.verdict]++
			blueSits += o.blueSittings
			blueAgains += o.blueAgain
			// THE FORCED RUN'S OWN VERDICT, kept rather than folded into the histogram. A
			// declining drive has to be reported as itself, and "the forced run ended CEILING"
			// is the whole diagnosis; a count says only that nothing reached the word.
			if o.seed == unverifiedSeed {
				forcedVerdict, forcedWhy = o.verdict, o.why
			}
			if o.seed == verifiedSeed {
				verifiedVerdict, verifiedWhy = o.verdict, o.why
			}
			epochHist[o.epochs]++
			if o.epochLimit {
				limitExits++
				limitForced = limitForced || o.seed == limitSeed
			}
			if o.why != "" {
				whyHist[o.verdict+" ← "+gapIDPat.ReplaceAllString(o.why, "G-<id>")]++
			}
			for k, v := range o.dialectic {
				dcov[k] += v
			}
			citeAnchors += o.citeAnchors
			cacheFiles += o.cacheFiles
			editAnswers += o.editAnswers
			verifiedBasis += o.verifiedBasis
			foundByMints += o.foundByMints
			verbatimApplied += o.verbatimApplied
			typedVerbatim += o.typedVerbatim
			estoppels += o.estoppels
			inquiryRecords += o.inquiryRecords
			closingsFiled, closingsRefused, closingWaits = closingsFiled+o.closingsFiled, closingsRefused+o.closingsRefused, closingWaits+o.closingWaits
			positionsWithheld += o.positionsWithheld
			ordersRefused += o.ordersRefused
			spansRefused += o.spansRefused
			if o.seed == absentSpanSeed {
				forcedSpansRefused += o.spansRefused
			}
			for verb, words := range o.copiesRead {
				if copiesRead[verb] == nil {
					copiesRead[verb] = map[string]int{}
				}
				for w, n := range words {
					copiesRead[verb][w] += n
				}
			}
			fragmentsDriven += o.fragmentsDriven
			truncationsDriven += o.truncationsDriven
			quotesRefused += o.quotesRefused
			bridgesRefused += o.bridgesRefused
			bridgesNamed += o.bridgesNamed
			reopenedReadied += o.reopenedReadied
			unsoundReadied += o.unsoundReadied
			if o.seed == unsoundProofSeed {
				forcedUnsoundReadied += o.unsoundReadied
			}
			unsoundPlans += o.unsoundPlans
			unsoundAnswers += o.unsoundAnswers
			unsoundRepeats += o.unsoundRepeats
			unsoundOverOpen += o.unsoundOverOpen
			reopenedPlans += o.reopenedPlans
			for k, v := range o.proofReopenedReadied {
				proofReopenedReadied[k] += v
			}
			owedKept += o.owedKept
			proofOwedKept += o.proofOwedKept
			unreadAnswers += o.unreadAnswers
			unreadRepeats += o.unreadRepeats
			unreadRetires += o.unreadRetires
			failedCloseRefused += o.failedCloseRefused
			for k, v := range o.owedKeptBy {
				owedKeptBy[k] += v
			}
			if o.citationsOffered > 0 {
				citationOfferedRuns++
			}
			if o.ownerOffered > 0 {
				ownerOfferedRuns++
			}
			repeatsRefused += o.repeatsRefused
			budgetRefused += o.budgetRefused
			ownAnswers += o.ownAnswers
			unclassifiedMinted += o.unclassifiedMinted
			unclassifiedRefused += o.unclassifiedRefused
			rulerFilingsRefused += o.rulerFilingsRefused
			closedRemandsRefused += o.closedRemandsRefused
			trifleRemandsRouted += o.trifleRemandsRouted
			blueRemandActs += o.blueRemandActs
			lensRemandActs += o.lensRemandActs
			pastLimitRemandsRefused += o.pastLimitRemandsRefused
			for why, n := range o.estoppelMisses {
				estoppelMisses[why] += n
			}
			for why, n := range o.applyMisses {
				applyMisses[why] += n
			}
			if o.err != "" {
				failures = append(failures, o)
			} else {
				os.RemoveAll(o.runDir) // keep only failing runs, for inspection
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	// AND WHY IT DID NOT, WHENEVER IT DID NOT. `%d edits applied a proposal verbatim` read 0
	// across 60 runs and named no cause: the branch could have gone untaken, found no proposal
	// on the gap, found the span already edited away, or run and been refused. Four causes, one
	// zero, in the instrument whose job is to detect exactly that shape.
	misses := ""
	if len(applyMisses) > 0 {
		causes := make([]string, 0, len(applyMisses))
		for why, n := range applyMisses {
			causes = append(causes, fmt.Sprintf("%d× %s", n, why))
		}
		sort.Strings(causes)
		misses = "\n  verbatim-apply declined: " + strings.Join(causes, "\n                          ")
	}
	t.Logf("BLUESTAT blue-respond sittings=%d (on an engaged successor %d) over %d runs", blueSits, blueAgains, completed)
	t.Logf("fuzzed %d debate runs · %d failed · verdicts=%v · epochs=%v · exits=%v\n  dialectic events emitted: %v\n  citation axis: %d anchors spliced · %d sources cached\n  provenance: %d of %d blue_edit ops carried --answers · %d of %d gaps earned fix_basis=verified · %d edits applied a proposal verbatim (%d typed) · %d estoppel refusals%s",
		completed, len(failures), verdicts, epochHist, whyHist, dcov, citeAnchors, cacheFiles, editAnswers, dcov["blue_edit"], verifiedBasis, dcov["mint"], verbatimApplied, typedVerbatim, estoppels, misses)
	t.Logf("epoch limit: %d of %d runs ended at their epoch limit (%d; seed %d is forced to it at %d)", limitExits, completed, record.DefaultParams.MaxEpochs, limitSeed, forcedEpochLimit)
	t.Logf("inquiry lifecycle: %d fates recorded on an id read from `inquiry propose` (of %d inquiry events)", inquiryRecords, dcov["inquiry"])
	t.Logf("proof owner: %d run(s) offered an unjudged proof to its owner, at a sitting or the run's end", ownerOfferedRuns)
	t.Logf("citation owner: %d run(s) offered an unverified citation to the evidence lens, at a sitting or the run's end", citationOfferedRuns)
	t.Logf("unsound proofs: %d plan(s) readied the judge or blue for every unsound proof they held (of %d holding one), %d of them seed %d's forced run; %d gap(s) minted carrying an unsound judgement", unsoundReadied, unsoundPlans, forcedUnsoundReadied, unsoundProofSeed, unsoundAnswers)
	t.Logf("source spans: %d cite(s) refused for a span the source does not hold, %d of them seed %d's forced run", spansRefused, forcedSpansRefused, absentSpanSeed)
	if completed > 0 && (unsoundReadied == 0 || unsoundAnswers == 0) {
		t.Errorf("fuzz readied for an unsound proof in %d plan(s) (of %d holding one) and minted %d gap(s) carrying one, across %d runs — the unsound-proof blocker's readiness or its answering gap is unexercised (false green)", unsoundReadied, unsoundPlans, unsoundAnswers, completed)
	}
	// A REPEAT UNSOUND OVER THE JUDGE'S OPEN GAP IS REFUSED, and a gap minted over an open one of the
	// judge's own about the same proof is what the refusal removes: a fresh judgement after blue's
	// edit, or one after the judge's own sound, is the only way to it.
	t.Logf("unsound repeats: %d repeat unsound judgement(s) refused over the judge's open gap; %d gap(s) minted about a proof while the judge's own gap about the sentence as it stands was open", unsoundRepeats, unsoundOverOpen)
	if completed > 0 && unsoundRepeats == 0 {
		t.Errorf("fuzz drew no repeat unsound judgement over an open gap across %d runs — the refusal is unexercised (false green)", completed)
	}
	t.Logf("reopened readied: %d plan(s) readied the owner of every reopened blocker (of %d holding one); reopened proofs whose owner a plan readied: %v", reopenedReadied, reopenedPlans, proofReopenedReadied)
	t.Logf("owed kept: %d lens fold(s) kept from retiring by a debt (three or more barren sittings), %d of them by an unjudged proof; by kind %v", owedKept, proofOwedKept, owedKeptBy)
	t.Logf("unread citations: %d gap(s) minted about a citation the lens's reading could not read, %d repeat unread reading(s) refused over that gap, %d citation(s) blue retired off a sentence for it, %d close(s) refused over a standing failed check", unreadAnswers, unreadRepeats, unreadRetires, failedCloseRefused)
	t.Logf("catch-all: %d unclassified mint(s) landed, %d refused for want of a reason", unclassifiedMinted, unclassifiedRefused)
	t.Logf("classes: %d by_grade in the mint pool; %d always/never minted once each, %d landed", len(mintPoolClasses), len(mintDrivenClasses), len(mintDrivenClasses)-len(unlandedClassDrive()))
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
	measured := completed >= surfaceQuorum
	if !measured {
		t.Logf("NOT MEASURED: %d runs is under the surface quorum of %d, so these gates did NOT run: "+
			"the per-verb event gate, the citation/provenance floors, the inquiry-record floor, the full-surface command gate, "+
			"and the flag/enum coverage sweeps. This is not a pass over them — only the cite/finding "+
			"floor below was checked. Run the default sweep to assert the surface.", completed, surfaceQuorum)
	}
	// EVERY VIEW REACHED SOMEBODY. The per-role sweep skips a view whose group a role does not
	// carry, and this is the witness for that skip: a view no role drove has been swept by nobody
	// while the sweep reported success, which is the exact shape this oracle's own comment warns
	// about. Under quorum it is not asserted, for the same reason the verb gates are not.
	if measured {
		for _, v := range viewNamesForFuzz {
			if !drivenView[v] {
				t.Errorf("no role drove the %q projection across %d runs — it lives under group %q, "+
					"which no role carries, so the sweep covered it with nobody", v, completed, seat.GroupOf(v))
			}
		}
	}
	if measured {
		for _, k := range verbsWithEvents {
			if coverExempt[k] {
				continue
			}
			if dcov[k] == 0 {
				t.Errorf("fuzz drove ZERO %s events across %d runs — that seat verb is unexercised (false green); a generator branch dropped it", k, completed)
			}
		}
	}
	// #256 CITATION-CHAIN GATE. `cite` in the verb gate above is satisfied by red's `lens cite`,
	// which neither fetches nor anchors — so a blue-cite regression would sail through it. These
	// two assert the chain that actually matters ran end to end through the real binary: a source
	// was FETCHED into <run>/cache (loopback server, see sourceURL) and an INVISIBLE anchor was
	// spliced into blue/report.md. Zero of either across the whole sweep is a false green.
	if measured {
		if citeAnchors == 0 {
			t.Errorf("fuzz spliced ZERO <!--cite:--> anchors across %d runs — `blue cite` never anchored (false green); the citation axis is unexercised", completed)
		}
		if cacheFiles == 0 {
			t.Errorf("fuzz cached ZERO sources across %d runs — `fetch`/`blue cite` never populated <run>/cache (false green); the fetch path is unexercised", completed)
		}
		// THE EPOCH-LIMIT GATE. Every run setup creates carries an epoch limit, and the exit it
		// makes — nobody dispatched at the last sitting, the terminal bench ruling what stands,
		// CEILING derived from the term — is one no other gate here notices going undriven: the
		// CEILING word is written by the carried ceiling too.
		if !limitForced {
			t.Errorf("the forced epoch-limit run on seed %d did not end at its limit of %d (%d run(s) of %d ended at a limit) — the exit a real run takes when its term runs out is unexercised (false green); epochs=%v",
				limitSeed, forcedEpochLimit, limitExits, completed, epochHist)
		}
		// THE INQUIRY-RECORD GATE. A proposed inquiry's fate is recorded by the id its propose result
		// printed, and the reader of that result is a pattern: one that matches nothing leaves every
		// proposal with no fate recorded and every inquiry `--about` reference unpassed, with the `inquiry` verb gate
		// above still green on the proposals themselves.
		if inquiryRecords == 0 {
			t.Errorf("fuzz recorded ZERO fates by an id read from `inquiry propose` across %d runs (with %d inquiry events) — "+
				"the propose result named no id the harness reads, or the tool refused every record; the inquiry lifecycle past `proposed` is unexercised by the proposing seat (false green)", completed, dcov["inquiry"])
		}
		// THE REOPENED GATE. A citation or proof whose sentence changed after its check is owed by
		// its owner, which is then not retired, and the edit moved the head: the plan readies it. A
		// sweep that readied the computation lens for a reopened proof zero times leaves the proof
		// arm of anchorOwnerOf unexercised.
		if reopenedReadied == 0 || proofReopenedReadied[proofOwner] == 0 {
			t.Errorf("fuzz readied the owner of every reopened blocker in %d plan(s) (of %d holding one), and the owner of a reopened proof %v times by owner, across %d runs — a reopened proof's routing to the computation lens is unexercised (false green)", reopenedReadied, reopenedPlans, proofReopenedReadied, completed)
		}
		// THE DEBT GATE. A lens that owes is not retired; a sweep where no fold kept a lens with three
		// barren sittings from retiring — or none for an unjudged proof — leaves the rule unexercised.
		//
		// THE UNJUDGED-PROOF DEBT IS FED BY CONSTRUCTION, not left to the draw (gblock, 2026-10-25): the
		// draw rarely holds it, since the owner checks every proof it owes in the sitting it is offered
		// one, and a sweep of 40 drew it zero times (#1604). owedProofSeed's run proves a fresh sentence
		// in every blue sitting while the owner is barren, so a fold holds the owner for a proof it has
		// not yet sat to judge. TestALensThatOwesIsNotRetired (internal/record) proves the rule itself.
		if owedKept == 0 || proofOwedKept == 0 {
			t.Errorf("fuzz kept %d lens fold(s) from retiring by a debt, %d of them by an unjudged proof, across %d runs — the rule that a lens that owes does not retire is unexercised (false green)", owedKept, proofOwedKept, completed)
		}
		// THE UNREAD-CITATION GATE. A reading that reads nothing leaves the citation on a failed check:
		// the lens owes the gap about it, and blue's remedy is to retire the citation. A sweep where no
		// lens minted that gap, or blue never retired such a citation, leaves the obligation and the
		// citation-only retire unexercised.
		if unreadAnswers == 0 || unreadRetires == 0 {
			t.Errorf("fuzz minted %d gap(s) about an unread citation and retired %d such citation(s) off its sentence, across %d runs — the failed-check path for a citation is unexercised (false green)", unreadAnswers, unreadRetires, completed)
		}
		// THE ANCHOR-OWNER GATE. checkAnchorOffers reads each lens's list at each of its sittings, and
		// anchorOffersAgree every lens's at each run's end; a sweep in which no owner was offered an
		// anchor of a kind asks neither anything of that kind, and leaves the routing unexercised.
		if citationOfferedRuns == 0 {
			t.Errorf("fuzz offered no unverified citation to the evidence lens in any of %d runs — the citation routing is unexercised (false green)", completed)
		}
		if ownerOfferedRuns == 0 {
			t.Errorf("fuzz offered no unjudged proof to its owner in any of %d runs — the proof routing is unexercised (false green)", completed)
		}
		// THE CLOSING GATE. The `closing` verb gate above is satisfied by any one act, so a plan that
		// stopped holding the bench for an awaited closing, or a drive that stopped leaving one unfiled,
		// would leave the duty's other half unreached behind a green sweep.
		if closingsFiled == 0 || closingsRefused == 0 || closingWaits == 0 {
			t.Errorf("fuzz filed %d closing(s), was refused %d envelope(s) while one stood and saw %d plan(s) keep the bench waiting on one, across %d runs — each must be nonzero, or the closing duty is unexercised (false green)", closingsFiled, closingsRefused, closingWaits, completed)
		}
		// NOT A GATE: a plan permitting a PASS is reached by the forced-verified run, which must record
		// it, and rarely by any other, so the withheld PASS is a draw the sweep may not make. The duty is
		// held by record's TestTheChairOwesAPositionWhereItWithheldAPermittedPass and capture's
		// record-parity tests; the count is reported so a sweep that does draw it says so.
		t.Logf("chair sittings that withheld a permitted PASS and argued why: %d across %d runs", positionsWithheld, completed)
		// THE SECTION-ORDER GATE. The generators here write the report's text in order, so the
		// refusal that holds it is reached only by the drive that writes a section out of place.
		if ordersRefused == 0 {
			t.Errorf("fuzz drove ZERO edits refused for a section out of order across %d runs — no drive wrote one, so the refusal that holds the order the report ships in is unexercised (false green)", completed)
		}
		// THE SOURCE-SPAN GATE. Every other cite quotes words its source's text holds, so the
		// refusal of a span the source does not hold is reached only by the drive that types one.
		if spansRefused == 0 {
			t.Errorf("fuzz drove ZERO cites refused for a source span the source does not hold across %d runs — the refusal that keeps blue's own wording off the record as the source's is unexercised (false green)", completed)
		}
		// THE COPY GATE. A reading records the copy the cache held for its source; corroborate is
		// driven over every seeded copy, and verify reads the copies those corroborations placed.
		for _, s := range seededCopies {
			if copiesRead["corroborate"][s.word] == 0 {
				t.Errorf("fuzz drove ZERO corroborations of a %s copy across %d runs — the copy a reading records is unexercised for it (false green): %v", s.word, completed, copiesRead)
			}
		}
		substituted := 0
		for w, n := range copiesRead["verify"] {
			if w != "none" {
				substituted += n
			}
		}
		if substituted == 0 {
			t.Errorf("fuzz drove ZERO verifications of a copy other than the url's own across %d runs — the copy a verification records is unexercised (false green): %v", completed, copiesRead)
		}
		// THE SPAN-BOUNDARY GATE. Every other held span is a whole sentence, so the refusal of one
		// cut from inside a clause is reached only by the drives that type sourceFragment and
		// sourceTruncation.
		if fragmentsDriven == 0 {
			t.Errorf("fuzz drove ZERO cites of a span starting mid-clause across %d runs — the refusal that keeps a fragment reading the other way off the record is unexercised (false green)", completed)
		}
		if truncationsDriven == 0 {
			t.Errorf("fuzz drove ZERO cites of a span ending mid-clause across %d runs — the refusal that keeps a span stopped before its clause ends off the record is unexercised (false green)", completed)
		}
		// THE REPORT'S-SENTENCE GATE. The lens drive quotes a real span of the report under every
		// outcome, so the refusal of a quote the report does not hold is reached only by the drive
		// that quotes the source's sentence.
		if quotesRefused == 0 {
			t.Errorf("fuzz drove ZERO corroborations refused for a quote the report does not hold across %d runs — no lens quoted its source's sentence, so the refusal that keeps a contradiction's claim the report's own sentence is unexercised (false green)", completed)
		}
		// THE BRIDGE GATE. supports_with_bridge names what it bridged; a reading naming none is
		// refused. Both halves must run, or the required value is unexercised behind a green sweep.
		if bridgesRefused == 0 || bridgesNamed == 0 {
			t.Errorf("fuzz drove %d supports_with_bridge readings refused for no --bridge and %d recorded with one across %d runs — both must be non-zero, or the required bridge kind is unexercised (false green)", bridgesRefused, bridgesNamed, completed)
		}
		// THE REPEATED-ACT GATE. A seat that got an act wrong meets the
		// refusal of its second, and that refusal names the act that answers the first. A drive
		// that stopped repeating an act would leave those refusals unreached behind a green sweep.
		if repeatsRefused == 0 {
			t.Errorf("fuzz drove ZERO refused repeats across %d runs — no seat filed a second position or closing in a sitting, so the refusal that names the answering act is unexercised (false green)", completed)
		}
		// THE MINT-BUDGET GATE. The budget caps the gaps a lens holds open, and the sweep's floor of
		// one keeps most lenses at it: a mint from a lens already holding its budget is refused. The
		// `mint` verb gate is satisfied by the mints that land, so a budget that stopped refusing — or
		// terms too generous for any lens to reach it — would leave the refusal unreached behind a
		// green sweep.
		if budgetRefused == 0 {
			t.Errorf("fuzz drove ZERO mints refused for budget across %d runs — no lens held its whole mint budget open and minted again, so the cap on open gaps is unexercised (false green)", completed)
		}
		// THE ANSWERING-MINT GATE. A gap minted from a lens's own raising finding takes no mint, and
		// a run fails on one the budget refused (runner.answerBudgetErr); that oracle is silent where
		// no such mint was driven, so the sweep fails unless some landed.
		t.Logf("answering mints: %d landed outside the minting lens's budget, %d mint(s) refused for budget", ownAnswers, budgetRefused)
		if ownAnswers == 0 {
			t.Errorf("fuzz drove ZERO mints answering the minting lens's own raise across %d runs — the gap outside the budget is unexercised (false green)", completed)
		}
		// THE CATCH-ALL GATE. A gap no class fits is minted `unclassified` with a reason, and the
		// record refuses one without. The `mint` verb gate is satisfied by every other class, so a
		// catch-all that stopped landing — or a reason check that stopped refusing — would hide
		// behind a green sweep.
		if unclassifiedMinted == 0 {
			t.Errorf("fuzz minted ZERO gaps under unclassified across %d runs — the catch-all is unexercised (false green)", completed)
		}
		if unclassifiedRefused == 0 {
			t.Errorf("fuzz drove ZERO unclassified mints refused for want of a reason across %d runs — the record's reason check is unexercised (false green)", completed)
		}
		// THE RULER-FILES-NOTHING GATE. The chair's grade motion and the bench's docket motion are
		// refused at the surface; a drive that stopped reaching for them would leave the refusal
		// unreached behind a green sweep.
		if rulerFilingsRefused == 0 {
			t.Errorf("fuzz drove ZERO refused ruler filings across %d runs — neither the chair filed a grade motion nor the bench a docket motion, so the refusal of a filing by its subject's ruler is unexercised (false green)", completed)
		}
		// THE REMAND GATES (gblock, 2026-10-24). A remand of a closed gap is refused naming the close,
		// and a remand of an open gap that is not material is readied like any remand; the run oracles
		// are silent where neither was driven.
		t.Logf("remands: %d of a closed gap refused naming the close, %d of a gap not material readied for their exchange", closedRemandsRefused, trifleRemandsRouted)
		if closedRemandsRefused == 0 {
			t.Errorf("fuzz drove ZERO remands of a closed gap across %d runs — the refusal that names the close is unexercised (false green)", completed)
		}
		if trifleRemandsRouted == 0 {
			t.Errorf("fuzz readied ZERO remands of a gap that is not material across %d runs — the route by which the bench's remand outranks materiality is unexercised (false green)", completed)
		}
		// THE REMAND'S ITEM BLOCKS, AND ONE REMAND AT IMPASSE IS THE LIMIT (gblock, 2026-10-25): each seat a
		// remand readies is held until it acts on the gap, and a second remand at impasse is refused.
		t.Logf("remands: %d act(s) by blue and %d by the minting lens on a remanded gap its item held them to; %d remand(s) past the limit refused", blueRemandActs, lensRemandActs, pastLimitRemandsRefused)
		if blueRemandActs == 0 || lensRemandActs == 0 {
			t.Errorf("fuzz drove %d blue and %d lens act(s) on a remanded gap whose item blocked the seat, across %d runs — the remand's blocking item is unexercised for a seat it holds (false green)", blueRemandActs, lensRemandActs, completed)
		}
		if pastLimitRemandsRefused == 0 {
			t.Errorf("fuzz drove ZERO remands past the limit across %d runs — the refusal that names the dispositions that close the gap is unexercised (false green)", completed)
		}
		// #267 PROVENANCE GATE. The verb gate above accepts any blue_edit, so an edit drive
		// that stopped sending --answers would satisfy it while the join key every #267
		// measurement reads went silently missing.
		if editAnswers == 0 {
			t.Errorf("fuzz recorded ZERO blue_edit events carrying --answers across %d runs — the provenance key is unexercised (false green)", completed)
		}
		if verifiedBasis == 0 {
			t.Errorf("fuzz minted ZERO gaps with fix_basis=verified across %d runs — no concrete proposal ever validated, so the whole stage-3 path is unexercised (false green)", completed)
		}
		if foundByMints == 0 {
			t.Errorf("fuzz minted ZERO gaps naming a finding in found_by across %d runs — no accepted mint carried --found-by, so the id a found_by names a finding by is unexercised (false green)", completed)
		}
		if verbatimApplied == 0 {
			t.Errorf("fuzz recorded ZERO verbatim applications across %d runs — nothing ever estopped red, so the stage-4 guard is unexercised (false green)", completed)
		}
		if typedVerbatim == 0 {
			t.Errorf("fuzz recorded ZERO typed verbatim applications across %d runs (%d by --accept) — blue typing the board's pair never matched the pair --accept sends, so the typed arm is unexercised (false green)", completed, verbatimApplied)
		}
		// AND THE GUARD ITSELF FIRED, which the line above only establishes the PRECONDITION for.
		// Verbatim applications were counted for the whole life of this gate while the estoppel
		// path was never once reached: red never minted against the text blue had applied, so the
		// refusal that writes `log --type estoppel` had no way to happen and the enum value was
		// reported as undriven — for however long, on a gate that only runs at a tag.
		if estoppels == 0 {
			// THE ZERO NAMES ITS REASON. `verbatimApplied` proves the precondition and this proves
			// the refusal; with nothing between them a red gate could not say whether the guard
			// had gone or the drive had simply declined, and those want opposite fixes.
			why := "the drive never declined either — it was not reached"
			if len(estoppelMisses) > 0 {
				causes := make([]string, 0, len(estoppelMisses))
				for c, n := range estoppelMisses {
					causes = append(causes, fmt.Sprintf("%d× %s", n, c))
				}
				sort.Strings(causes)
				why = "the drive declined: " + strings.Join(causes, "; ")
			}
			t.Errorf("fuzz recorded ZERO tool estoppel refusals across %d runs (with %d verbatim applications) — "+
				"red never minted against text it had prescribed and blue had applied, so lens/mint.go's estoppel guard "+
				"is unexercised and `log --type estoppel` is unreachable (false green).\n          %s", completed, verbatimApplied, why)
		}
	}
	if !measured {
		for _, k := range []string{"cite", "finding"} {
			if dcov[k] == 0 {
				t.Errorf("fuzz drove ZERO %s events across %d runs — the lens record channel is unexercised (false green)", k, completed)
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
		if un := unreachedSurfaces(); len(un) > 0 {
			t.Errorf("%d command path(s) exist in the tool and were NEVER invoked across %d runs (false green — add a drive, or an exemption with its reason):\n  %s",
				len(un), completed, strings.Join(un, "\n  "))
		}
		// DRIVING A VERB IS NOT DRIVING ITS SURFACE. The gate above asks whether every verb
		// ran; these ask whether the flags those verbs take were ever passed, and whether
		// every value of every closed set was ever recorded. An unpassed flag is code no run
		// has executed, and an unreached enum member is a word nothing has written — both
		// sitting behind a sweep that reports full coverage.
		if un := unreachedFlags(); len(un) > 0 {
			t.Errorf("%d flag(s) on verbs the sweep DOES drive were never passed:\n  %s", len(un), strings.Join(un, "\n  "))
		}
		if un := unreachedEnumValues(); len(un) > 0 {
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
		if forcedVerdict != "UNVERIFIED" {
			t.Errorf("the forced-UNVERIFIED run on seed %d ended %q, not UNVERIFIED — the drive that makes "+
				"`outcome --as UNVERIFIED` a fact rather than a 1-in-60 draw did not deliver, and the enum census "+
				"reports that as a missing WORD rather than as this declining DRIVER.%s",
				unverifiedSeed, forcedVerdict, forcedWhyClause(forcedWhy))
		}
		// AND THE FORCED-VERIFIED DRIVE, for the same reason: the run that exists to write the word
		// went somewhere else, and this names where.
		if verifiedVerdict != "VERIFIED" {
			t.Errorf("the forced-VERIFIED run on seed %d ended %q, not VERIFIED — the drive that makes "+
				"`outcome --as VERIFIED` a fact rather than a draw did not deliver.%s",
				verifiedSeed, verifiedVerdict, forcedWhyClause(verifiedWhy))
		}
		if !mootSpent.Load() {
			t.Error("the sweep's single `close --as moot` drive never fired: no run reached closeGap with it pending, or every attempt was refused. The word's whole coverage is that one drive")
		}
		if lookedFirst.Load() == 0 {
			t.Error("no lens sitting opened with a look before its register: the drive that holds a board read to registering no one never completed")
		}
		// AND THE DELIBERATE CLASS MINTS, for the same reason: each `always`/`never` class is one
		// drive per sweep (r.mint), and the census above counts a refused attempt as driven.
		if un := unlandedClassDrive(); len(un) > 0 {
			t.Errorf("the sweep's deliberate class mints landed no gap under %d of %d `always`/`never` class(es): %s — no run minted with them owed, or every attempt was refused. Each word's whole coverage is that one drive",
				len(un), len(mintDrivenClasses), strings.Join(un, ", "))
		}
		// AND THE FLAGS INHERITED ONTO EVERY COMMAND, which this gate could not see at all until
		// PersistentFlagNames existed: CommandFlags reports local flags only, so `--json`,
		// `--run`, `--schema` and `--seat-id` were never in the denominator above.
		if un := unreachedPersistentFlags(); len(un) > 0 {
			t.Errorf("%d persistent flag(s) — inherited onto every command — were never passed and "+
				"carry no exemption:\n  %s", len(un), strings.Join(un, "\n  "))
		}
	}
	// AT EVERY SIZE, both of these. The tally is the sweep's own trajectory and costs nothing to
	// print; withholding it under quorum hid the evidence for the gate below on exactly the runs
	// most likely to need it. And a path that never succeeded is a coverage hole at ANY N — it
	// needs no quorum, because one refusal per invocation is not a sampling accident.
	t.Log(execReport())
	t.Log(eitherReport())
	t.Log(failReport())
	t.Log(envelopeReport())
	// THE SAME SWEEP, READ AS A GRAPH (#535 step 1). Emitted here rather than from a test of its
	// own for the reason #635 was filed about: a second test reading this package-level tally
	// would run at whatever point its declaration order put it, and a gate that reads a tally
	// before its writers have run reports a plausible zero. Here the writers are done and the
	// quorum is already known.
	t.Log(graphReport())
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
		if seats, edges, _, _ := graphEdges(); seats < 2 || edges == 0 {
			t.Errorf("the observed graph recorded %d seat(s) and %d edge(s) across %d runs — the sweep drove the surface, so this is the EMITTER failing to attribute it, not a quiet run",
				seats, edges, completed)
		}
	}
	walked, wrong := namedOtherThanOnce()
	t.Logf("%d refusals walked for how many times each names its command; %d other than once", walked, len(wrong))
	if len(wrong) > 0 {
		t.Errorf("%d refusal(s) the sweep met name their command other than exactly once — the renderer prints the path from the command that resolved, and the diagnosis starts at what is wrong:\n  %s",
			len(wrong), strings.Join(wrong, "\n  "))
	}
	if un := neverSucceeded(); len(un) > 0 {
		t.Errorf("%d command path(s) were driven and NEVER SUCCEEDED — the surface gate counts invocations, so these read as covered while their success path has never run:\n  %s",
			len(un), strings.Join(un, "\n  "))
	}
	if len(failures) > 0 {
		// EIGHT, AND IT SAYS SO WHEN IT TRUNCATES. Sixty seeds failing the same way is a wall of
		// identical text, so the cap is right — but "17/60 failed, see seeds above" over eight
		// printed seeds invites a reader to take the eight as the whole distribution and count
		// classes from them. A silent cap reads as full coverage; this one names what it dropped.
		show := failures
		if len(show) > 8 {
			show = show[:8]
		}
		for _, f := range show {
			t.Errorf("seed %d FAILED (runDir %s):\n%s", f.seed, f.runDir, f.err)
		}
		more := ""
		if len(failures) > len(show) {
			var rest []string
			for _, f := range failures[len(show):] {
				rest = append(rest, fmt.Sprintf("%d", f.seed))
			}
			more = fmt.Sprintf(" — %d of them NOT printed above (seeds %s); re-run one of those directly to see its failure",
				len(rest), strings.Join(rest, " "))
		}
		t.Fatalf("%d/%d fuzz runs failed — see seeds above (reproduce with that seed)%s", len(failures), n, more)
	}
}

// buildBinary is testbuild.Binary: built once per test binary instead of once per call, and
// named by the platform convention rather than an unconditional ".exe".
func buildBinary(t *testing.T) string {
	t.Helper()
	return testbuild.Binary(t, "feov-record")
}

// mintedGapIDs lists the gaps a run actually created, so a scoped-view oracle asks about
// something real rather than about a shape the fuzz happened not to produce this seed.
// someReportAnchor returns one anchor id this run actually put in the report, or "" before any
// finding has been recorded. Read from the RECORD (the `finding` event's label) rather than by
// scanning report.md for a token, so the oracle is not testing the reader with the reader.
// It reads the ANCHOR events, not the findings. Sourcing it from findings assumed every finding
// splices a marker, which stopped being true when a finding could anchor to a section, an inquiry
// or a gap (#742/#787): those name something that is not report text, so there is nothing to mark
// and `show report --anchor` on one correctly finds no window. Asking the anchor events instead
// asks the question this function's name already claimed to ask.
// An anchor `retire` took out with its claim is not in the report, and `show report --anchor`
// refuses it by design; the oracle reads a window, so it skips those.
func someReportAnchor(run record.Run) string {
	b, err := record.FamilyOf(run)
	if err != nil {
		return ""
	}
	for _, e := range b.Events {
		if a, ok := recordpb.BodyAs[*recordpb.Anchor](e); ok && a.GetId() != "" {
			if _, _, _, retired, err := record.AnchorRetiredAt(run, a.GetId()); err != nil || retired {
				continue
			}
			return a.GetId()
		}
	}
	return ""
}

// someProof is one recorded proof's anchor and sha256, or "" when the run recorded none.
func someProof(run record.Run) (anchor, sha string) {
	b, err := record.FamilyOf(run)
	if err != nil {
		return "", ""
	}
	for _, e := range b.Events {
		if p, ok := recordpb.BodyAs[*recordpb.Proof](e); ok && p.GetProofId() != "" && p.GetProofSha() != "" {
			return p.GetProofId(), p.GetProofSha()
		}
	}
	return "", ""
}

func mintedGapIDs(run record.Run) []string {
	b, err := record.FamilyOf(run)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range b.Events {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok {
			out = append(out, m.GetGapId())
		}
	}
	return out
}

// scenarioOf reads a gap's directive BACK FROM THE BOARD, not from Go memory. The round trip
// is part of what is tested: if required_fix stopped surviving a mint, every seat would
// silently revert to coin-flip behaviour and the oracle would catch it.

// recentlyEditedOut returns text a recorded edit removed and which is absent from the report
// now — a claim whose retirement the record can evidence.
func (r *runner) recentlyEditedOut() string {
	b, err := record.FamilyOf(r.run())
	if err != nil {
		return ""
	}
	cur, err := reportproj.RenderFromRecord(r.run())
	if err != nil {
		return ""
	}
	for i := len(b.Events) - 1; i >= 0; i-- {
		e := b.Events[i]
		be, ok := recordpb.BodyAs[*recordpb.BlueEdit](e)
		if !ok {
			continue
		}
		old := be.GetOld()
		if old != "" && !strings.Contains(string(cur), old) {
			return old
		}
	}
	return ""
}

// anchorsLeftFor names a marker for a retire's --anchor: one the edit that took claim out left in
// its place, when it still stands; else any marker the report holds, so the flag is driven through
// its validation on every retire the sweep makes — accepted or refused, never unpassed.
//
// left says which: true for a marker the edit left, false for one drawn from the whole report.
func (r *runner) anchorsLeftFor(claim string) (id string, left bool) {
	b, err := record.FamilyOf(r.run())
	if err != nil {
		return "", false
	}
	cur, err := reportproj.RenderFromRecord(r.run())
	if err != nil {
		return "", false
	}
	for i := len(b.Events) - 1; i >= 0; i-- {
		be, ok := recordpb.BodyAs[*recordpb.BlueEdit](b.Events[i])
		if !ok || be.GetOld() != claim {
			continue
		}
		for _, id := range anchor.IDs(be.GetNew()) {
			if strings.Contains(string(cur), anchor.Token(id)) {
				return id, true
			}
		}
	}
	if ids := anchor.IDs(string(cur)); len(ids) > 0 {
		return ids[r.rng.Intn(len(ids))], false
	}
	return "", false
}

// inquiryFates are the scenarios an inquiry is proposed under. The question itself carries it, the
// way a gap's required_fix does — red rules from it, and the fate blue records next answers the ruling.
const (
	avEndorse = "FUZZ-INQUIRY-ENDORSE: a line worth the run's time"
	avScope   = "FUZZ-INQUIRY-OUT-OF-SCOPE: a real question, but not this run's"
	avThin    = "FUZZ-INQUIRY-TOO-THIN: in scope, but the hypothesis does not carry its budget"
	// avContest is the contest: red rules the inquiry out, and blue PURSUES IT ANYWAY with an
	// argument. A ruling is an argument, never a command, and disputes are gap-shaped, so the
	// fate blue records is itself the contest, and it is recorded as one.
	avContest = "FUZZ-INQUIRY-CONTESTED: red rules it out and blue pursues it anyway, with reasons"
)

var inquiryFates = []string{avEndorse, avEndorse, avScope, avThin, avContest}

// rulingFor maps a proposed question to the ruling red should give it.
//
// UNDERSCORES: these are the schema's words, the ones `motion inquiry rule --as` accepts and
// record.InquiryRuling answers with. A hyphenated word is refused at the write, the drive discards
// the refusal, and everything below a ruling starves with it — blue answers a ruling only once
// the inquiry is ruled.
func rulingFor(line string) string {
	switch {
	case strings.HasPrefix(line, avEndorse):
		return "endorsed"
	case strings.HasPrefix(line, avScope):
		return "out_of_scope"
	case strings.HasPrefix(line, avThin):
		return "too_thin"
	case strings.HasPrefix(line, avContest):
		return "out_of_scope"
	}
	return ""
}

// ruleOpenInquiries has red rule every inquiry that has no ruling yet, from the question it carries.
func (r *runner) ruleOpenInquiries(seatID string) {
	b, err := record.FamilyOf(r.run())
	if err != nil {
		return
	}
	for _, a := range record.InquiriesOf(b.Events, b.At) {
		if rulingFor(a.Question) == "" {
			continue
		}
		if ruling, ok := r.inquiryRuling(a.ID); !ok || ruling != "" {
			continue
		}
		r.admit("motion", "inquiry", "rule", "--seat-id", seatID, "--id", a.ID,
			"--as", rulingFor(a.Question), "--reason", "fuzz: ruling as the line was proposed")
	}
}

// inquiryRuling is the drive's read of red's ruling on an inquiry: record.InquiryRuling, with a read
// that fails kept on the runner so runOne fails the run on it. ok is false for that read, and the
// caller leaves the inquiry alone — it knows neither that the inquiry is ruled nor that it is not.
func (r *runner) inquiryRuling(inquiryID string) (ruling string, ok bool) {
	ruling, err := record.InquiryRuling(r.run(), inquiryID)
	if err != nil {
		r.mu.Lock()
		if r.inquiryRulingErr == "" {
			r.inquiryRulingErr = err.Error()
		}
		r.mu.Unlock()
		return "", false
	}
	return ruling, true
}

// answerInquiryRulings is blue's answer after red has ruled: comply, or CONTEST by pursuing
// anyway with an argument. Which one is decided by the question, not by a coin.
func (r *runner) answerInquiryRulings(seatID string) {
	b, err := record.FamilyOf(r.run())
	if err != nil {
		return
	}
	for _, a := range record.InquiriesOf(b.Events, b.At) {
		ruling, ok := r.inquiryRuling(a.ID)
		if !ok || ruling == "" || a.Status != "proposed" {
			continue
		}
		switch {
		case strings.HasPrefix(a.Question, avContest):
			// THE CONTEST IS THE FATE RECORDED: the ruling binds none, so blue takes the inquiry up against it.
			r.do("inquiry record", seatID).set("--id", a.ID).set("--as", "underway").
				set("--reason", "fuzz: the scope call is wrong, this bears on the core claim").admit()
		case ruling == "endorsed":
			r.do("inquiry record", seatID).set("--id", a.ID).set("--as", "underway").
				set("--reason", "fuzz: endorsed, taking it up").admit()
		default:
			r.do("inquiry record", seatID).set("--id", a.ID).set("--as", "declined").
				set("--reason", "fuzz: accepting the ruling").admit()
		}
	}
}

// reproveOpenProofs audits the proof answering each open PROVE gap, and the gap's minting lens
// closes on the proof's standing re-run rather than on its own say-so. The proof's OWNER re-runs it
// and judges its sentence — the one audit that does not end in believing bytes somebody else chose.
// Every other lens reads the standing re-run off `show evidence` and re-runs nothing: a proof's
// sentence is content to it.
func (r *runner) reproveOpenProofs(seatID string) {
	b, err := record.FamilyOf(r.run())
	if err != nil {
		return
	}
	proofFor := map[string]string{} // gap -> the anchor of the proof that answers it
	for _, e := range b.Events {
		if p, ok := recordpb.BodyAs[*recordpb.Proof](e); ok && p.GetAnswers() != "" && p.GetProofId() != "" {
			proofFor[p.GetAnswers()] = p.GetProofId()
		}
	}
	standing := map[string]bool{} // anchor -> its script's standing re-run reproduced
	for _, p := range record.EvidenceJSONOf(b.Events, b.At).Proofs {
		standing[p.Anchor] = p.Rerun == recordpb.Word(recordpb.Rerun_RERUN_REPRODUCED)
	}
	for _, id := range r.openGaps() {
		if d := r.scenarioOf(id); d != dirProve && d != dirProveDrifts {
			continue
		}
		at := proofFor[id]
		if at == "" {
			continue // blue has not answered it yet; red has nothing to re-run
		}
		if seatID != proofOwner {
			if standing[at] {
				r.reproduced.add(id)
			}
			continue
		}
		out, err := r.exec("--json", "reproduce", "--seat-id", seatID, "--anchor", at, "--as", r.judgement(), "--reason", "fuzz: read the script and re-ran it")
		if err != nil {
			continue
		}
		var env struct {
			Result struct {
				Reproduced bool `json:"reproduced"`
			} `json:"result"`
		}
		if json.Unmarshal([]byte(strings.TrimSpace(out)), &env) == nil && env.Result.Reproduced {
			r.reproduced.add(id)
		}
	}
}

func (r *runner) scenarioOf(gapID string) string {
	b, err := record.FamilyOf(r.run())
	if err != nil {
		return ""
	}
	g := b.Gap(gapID)
	if g == nil || g.Mint == nil {
		return ""
	}
	return g.Mint.GetRequiredFix()
}

// blueRespondTo carries out the scenario each open gap was minted with — one decision per gap,
// taken from the gap rather than from a coin, which is what makes the terminal state assertable.
func (r *runner) blueRespondTo(seatID string, open []string) {
	for _, id := range open {
		switch r.scenarioOf(id) {
		case dirApply:
			// The only branch that sets applied_verbatim, and so the only one that estops red.
			//
			// EVERY WAY IT CAN FAIL IS COUNTED, because the fallthrough below is silent and the
			// gate downstream reported "0 verbatim applications" across 60 runs without being
			// able to say whether the branch was never taken, took and found no proposal, took
			// and found the span already edited away, or ran and was REFUSED. Four causes, one
			// zero — the plausible-zero shape, in the instrument meant to detect it.
			cur, err := reportproj.RenderFromRecord(r.run())
			gid, fo, fn := r.proposalFor(id, cur)
			switch {
			case gid == "":
				r.noteApplyMiss("no proposal on the gap (red minted prose-only)")
			default:
				switch {
				case err != nil:
					r.noteApplyMiss("report unreadable: " + err.Error())
				case !strings.Contains(string(cur), fo):
					r.noteApplyMiss("the proposed span is no longer in the report (an earlier edit moved it)")
				default:
					// TWO WAYS TO APPLY THE SAME PROPOSAL, and the sweep only ever drove one.
					//
					// `--accept` is the whole point of the flag: the tool reads --quote and --new
					// off the gap's own mint, so blue accepts red's text WITHOUT retyping it.
					// Retyping is the path that can silently diverge — a seat that mistypes one
					// character has authored an edit while claiming to have accepted one — which
					// is why the flag exists and why leaving it undriven mattered.
					//
					// Both arms stay. They are different write paths: `--accept` is refused when
					// red prescribed no concrete text, and the explicit arm is what a seat uses
					// when it is applying something red did NOT prescribe verbatim.
					// ALTERNATING, NOT A COIN — and the difference is a gate that flakes.
					//
					// This was `r.coin(50)` on a branch that is itself rare: the apply path needs
					// a dirApply scenario, a gap carrying a concrete proposal, and the proposed
					// span still in the report. Measured across sweeps it reaches here 8-13 times
					// in 40 runs, so a 50% coin on top has a real chance of never choosing one
					// arm — and it did, reporting `blue edit --accept` as never passed on an
					// honest sweep. That is the surfaceQuorum problem one layer down: the gate
					// asserts full coverage, so a low-frequency drive must not be probabilistic.
					//
					// ACROSS THE SWEEP, NOT WITHIN A RUN — and the first attempt got that wrong.
					// A per-runner counter alternates nothing when the branch fires about once
					// per run: every run took turn 1 and chose the same arm, so `--accept` was
					// still never passed. The coverage gate is a property of the SWEEP, so the
					// thing that guarantees both arms has to be too.
					useAccept := applyTurn.Add(1)%2 == 0
					edit := r.do("edit", seatID).set("--answers", id)
					if useAccept {
						edit = edit.bare("--accept").
							set("--reason", "fuzz: accepting red's prescribed fix on "+id+" exactly as recorded")
					} else {
						edit = edit.set("--quote", fo).set("--new", fn).
							set("--reason", "fuzz: applying red's proposed text verbatim")
					}
					if _, err := edit.result(); err != nil {
						r.noteApplyMiss("blue edit REFUSED: " + firstLine(err.Error()))
						break
					}
					r.noteVerbatimGap(id, fn)
					continue
				}
			}
			// No applicable pair — degrade to a counter-edit rather than fake one.
			fallthrough
		case dirCounter:
			r.counterEdit(seatID, id)
		case dirDisputeWon, dirDisputeLost:
			dim := pick(r.rng, disputeDims)
			// A DIFFERENT GRADE FROM THE ONE ON THE BOARD. A motion proposing the grade already
			// there is refused now (it asks for no change), and a drive that hits that 1 time in
			// 5 would report as a missing dispute event three layers away — which is exactly the
			// shape the hyphenated grade words produced.
			proposed := r.gradeOtherThan(r.currentGrade(id, dim))
			out, err := r.exec("--json", "motion", "grade", "file", "--seat-id", seatID, "--id", id,
				"--dimension", dim, "--proposed", proposed, "--reason", "fuzz: contesting the grade on "+id)
			// THE REFUSAL IS KEPT, because the oracle downstream reports its ABSENCE and cannot
			// say why. `scenario DISPUTE: G1 has no dispute event` is what a discarded error
			// looks like from the far end — a coverage report about a drive that ran fine.
			if err != nil {
				if r.disputeRefusals == nil {
					r.disputeRefusals = map[string]string{}
				}
				r.disputeRefusals[id] = err.Error()
			}
			if err == nil {
				if mid := motionIDOf(out); mid != "" {
					// The envelope ref keeps (gap_id, dimension) because debate.js routes the
					// DOCKET on them; the motion id rides alongside so the ruling can name the
					// ask it answers. Not redundant: the pair says which grade, the id says
					// which ASK about it, and #312 is the case where one gap and one dimension
					// had two asks and the pair could not tell them apart.
					ref := map[string]any{"gap_id": id, "dimension": dim, "proposed": proposed, "motion_id": mid}
					r.raised = append(r.raised, ref)
				}
			}
		case dirProve, dirProveDrifts:
			// SETTLE IT BY COMPUTING. The script is written into the run dir and the tool runs
			// it TWICE, so this drives the real interpreter, the cache, the anchor splice and
			// the reproducible/observed grading — and --answers ties the execution to the gap
			// it settles, which is what makes red's re-run targetable rather than arbitrary.
			name := "fuzz-answer-" + id + ".js"
			body := "console.log('settles " + id + "');"
			if r.scenarioOf(id) == dirProveDrifts {
				body = "console.log(Math.random());" // graded `observed`, and it will not re-run the same
			}
			if err := r.writeScript(name, body); err == nil {
				prove := r.do("prove", seatID).
					set("--quote", "§ fuzz sentence").
					set("--script", name).
					set("--answers", id).
					set("--reason", "fuzz: computing rather than arguing")
				// The METHOD is cited and the instance computed — that pairing is the whole
				// claim of the proof axis. It passed the FABRICATED label "c-fuzz", which named
				// no citation on the record: a proof claiming a provenance it did not have, and
				// the report rendered the link. `requireCitation` refuses it now, so this takes
				// a real anchor off the record — the same discipline someFinding and someCitation
				// already impose everywhere else a label is passed.
				if anchor := r.someCitation(); anchor != "" && r.coin(50) {
					prove.set("--cites", anchor)
				}
				prove.admit()
			}
		case dirIgnore:
			// Deliberately nothing. The oracle asserts NO edit answers this gap, which is what
			// proves --answers records provenance rather than decorating it.
		}
	}
}

// counterEdit makes a real edit that is NOT red's proposed text.
func (r *runner) counterEdit(seatID, gapID string) {
	cur, err := reportproj.RenderFromRecord(r.run())
	if err != nil {
		return
	}
	oldSpan, newSpan := "rising over time", "climbing sharply"
	if !strings.Contains(string(cur), oldSpan) {
		oldSpan, newSpan = newSpan, oldSpan
	}
	if !strings.Contains(string(cur), oldSpan) {
		return
	}
	run := r.edgeRun(cur, oldSpan)
	r.do("edit", seatID).set("--quote", oldSpan+run).set("--new", newSpan+run).
		set("--answers", gapID).set("--reason", "fuzz: counter-edit, not red's text").admit()
}

// proposalFor returns the concrete pair for ONE gap over the report cur, when it carries one — the
// pair the board serves and `--accept` sends (record.Proposal), so typing it is applying it verbatim.
func (r *runner) proposalFor(gapID, cur string) (string, string, string) {
	b, err := record.FamilyOf(r.run())
	if err != nil {
		return "", "", ""
	}
	g := b.Gap(gapID)
	if g == nil || g.Mint == nil || g.Mint.GetFixBasis() != "verified" {
		return "", "", ""
	}
	old, fixNew, found, err := record.Proposal(r.run(), gapID, cur)
	if err != nil || !found {
		return "", "", ""
	}
	return gapID, old, fixNew
}

// edgeRun is what an edit of span writes after it in both --quote and --new: half the time the
// anchor run standing at its end, quoted as `show report` prints it and carried, and half the time
// nothing — the quote stops before the run, which stays where it stands. Both are legal, so neither
// may add a refusal.
func (r *runner) edgeRun(cur, span string) string {
	if r.coin(50) {
		return ""
	}
	return abuttingRun(cur, span)
}

// abuttingRun is the anchor run standing right after span's first occurrence in cur.
func abuttingRun(cur, span string) string {
	i := strings.Index(cur, span)
	if i < 0 {
		return ""
	}
	after := cur[i+len(span):]
	return after[:anchor.SkipRun(after, 0)]
}

// ---- read-only INVOCATION: the surfaces the event gate is blind to ----
//
// These verbs RECORD NOTHING, so the event-type gate cannot see them: of 44 seat verbs, every
// one with zero fuzz coverage was of this kind, plus 7 of 9 root commands. What follows only
// CALLS them — the counting is done by the execution tally (noteExec), which observes every
// invocation the harness makes rather than keeping a second ledger of its own.

// readOnly invokes a record-nothing seat verb. r.exec tallies it; a non-zero exit shows up in
// the report's refusal count, which is where the dead `lens inquiry` drive and the misrouted
// `frontier` registration were both found.
func (r *runner) readOnly(verb, seatID string, extra ...string) {
	args := append([]string{verb, "--seat-id", seatID}, extra...)
	r.admit(args...)
}

// readOnlySurfaces is the CENSUS, not a sample: every record-nothing surface a seat or the
// operator can reach, with the argv that exercises it. Adding a read-only verb without adding
// it here leaves it unfuzzed, and the gate below is what says so out loud.
//
// `setup` and `capture` are deliberately absent and that is stated rather than silently
// omitted: setup CREATES a run (the fuzz builds its run dir directly, so driving setup would
// mean fuzzing a different thing), and capture needs a transcript directory the harness has no
// analogue for. Both carry their own package tests. `hook` reads a JSON payload on stdin rather
// than argv and is covered by internal/cli's hook tests.
var readOnlySurfaces = [][]string{
	// Every SEAT's `show` with no view, which resolves that seat's DEFAULT. Only the chair's was
	// ever driven, so a regression in any other default was invisible. The seat id is what selects
	// the tree now, so it is what distinguishes these four rather than a role word in front.
	{"show", "--seat-id", "blue-respond"},
	{"show", "--seat-id", "red-lens-evidence"},
	{"show", "--seat-id", "red-chair"},
	{"show", "--seat-id", "judge"},
	// The operator renders, over whatever shape the run actually reached.
	{"graph", "--format", "mermaid"},
	{"graph", "--format", "dot"},
	{"count-claims"},
	{"scorecard"},
	{"scorecard", "--card", "red"},
	// Every command's help on a seat's surface, which it produces by running this binary again
	// once per page. One seat is enough: it is one contract on every surface, keyed once.
	//
	// THE OPERATOR RENDERS IT, `--for` NAMES THE SURFACE. `manual` is a generator's command and its
	// own help says a seat "does not run it and cannot" — it left every seat's surface, and this call
	// kept the old form of passing the seat's own id. The read it is asserting is unchanged: blue's
	// whole surface, every page's help produced live.
	{"manual", "--seat-id", "operator", "--for", "blue-respond"},
}

// dashboardArgv is separate because `dashboard` is POSITIONAL — `dashboard <runDir>
// <transcript-dir>` — not `--run`. Discovered by this very census: the first version of it
// passed --run and dashboard failed on all 60 runs with a usage error, which is the census
// working (the surface had never been driven, so nothing here was known).
func dashboardArgv(runDir string) []string {
	return []string{"dashboard", runDir, runDir, "--seat-id", "operator"}
}

func containsFlag(argv []string, flag string) bool {
	for _, a := range argv {
		if a == flag || strings.HasPrefix(a, flag+"=") {
			return true
		}
	}
	return false
}

// sweepReadOnly runs the census against a finished run and reports the first surface that
// fails. Run AFTER the debate so the state is arbitrary rather than empty — an empty run
// proves only that nothing crashed on nothing.
func sweepReadOnly(bin, runDir string) string {
	// The transcript dir is the run dir here: dashboard tolerates finding no agent-*.jsonl,
	// and what is under test is that it RENDERS whatever board shape the run reached.
	if out, err := drive(bin, dashboardArgv(runDir)...); err != nil {
		return "read-only surface `dashboard` failed on a real run shape:\n" + truncate(string(out))
	}
	for _, argv := range readOnlySurfaces {
		args := append(append([]string{}, argv...), "--run", runDir)
		// THE OPERATOR SAYS SO TOO. graph, count-claims, scorecard and the rest are the operator's,
		// and a surface scoped to whoever is asking has nothing to show a caller who has not said.
		if !containsFlag(argv, "--seat-id") {
			args = append(args, "--seat-id", "operator")
		}
		if out, err := drive(bin, args...); err != nil {
			return "read-only surface `" + strings.Join(argv, " ") + "` failed on a real run shape:\n" + truncate(string(out))
		}
	}
	return ""
}

// NO PER-INQUIRY VOTE: an inquiry is on the worklist because the record says so, not because a
// seat voted it there, so PRESENCE is not a question. What is genuinely open — whether
// blue's body delivered the research a line claims — is an ORDINARY GAP now, with the id, the
// grade, the blue duty and the PASS gate every gap already has.
//
// A fuzz action driving a retired verb is not coverage. It would spend every run issuing a command
// the tool refuses, and the refusals would read as the tool working.

// fieldStr reads a named string field off an event's body, by DESCRIPTOR.
//
// The basisFields table is keyed on field names, which is the right shape for it — the same rule
// stated once per (event, field, value) triple. Reading them off typed bodies with a switch would
// be seven near-identical arms, and reading them out of a map is what the migration removed.
//
// THE MISS IS LOUD. A name the schema does not carry FAILS THE TEST rather than returning "", and
// that is the whole reason this is not a bare lookup: a stale key silently never matches, so the
// gate would report "the report never says so" for a field that no longer exists — or, worse,
// report nothing at all and read as covered. record.go's keyFields carries the same warning about
// the same shape.
func fieldStr(t *testing.T, e *record.Event, name string) string {
	t.Helper()
	body, ok := recordpb.Body(e)
	if !ok {
		return ""
	}
	m := body.ProtoReflect()
	fd := m.Descriptor().Fields().ByName(protoreflect.Name(name))
	if fd == nil {
		t.Fatalf("fuzz: %s has no field %q — the basisFields table names a field the schema does not "+
			"carry, so its rule can never match and the gate would read as covered",
			m.Descriptor().FullName(), name)
	}
	if !m.Has(fd) {
		return ""
	}
	return m.Get(fd).String()
}

// seatOfRole is a dispatched seat of the given role, for drives that iterate roles. The role is no
// longer part of an invocation; the seat id is what selects the tree it runs in.
func seatOfRole(role string) string {
	return map[string]string{
		"blue": "blue-respond", "lens": "red-lens-evidence",
		"chair": "red-chair", "bench": "judge",
	}[role]
}

// TestFuzzUnverifiedPath drives the UNVERIFIED terminal verdict deterministically.
//
// IT USED TO BE LUCK. debate.js computes UNVERIFIED only for a run that is not halted, whose red
// never passed, and which did NOT stop at the ceiling — the judged-deadlock exit with gaps still
// open. In the random sweep that happens about once in sixty runs, because a bench that disposes
// its whole docket clears the board and the engine grants red a further sitting instead, which
// then passes. So `outcome --as UNVERIFIED` was an enum value the coverage gate demanded and the
// sweep supplied by chance: it passed at N=60 with 2 occurrences and 1 at N=40, then failed at
// N=40 with none. Roughly a one-in-three flake, and it had been one all along.
//
// coverage_test.go already refused the other answer, in writing: `outcome --as UNVERIFIED` was
// once exempted on a claim that the engine had no path to it, a real run falsified that, and the
// note left behind says an unreachable value should make the sweep SAY so rather than earn an
// exemption explaining why it never was. So this drives it, the way TestFuzzHaltPath drives
// HALTED — a terminal shape too disruptive for the random sweep gets a dedicated run, and its
// invocations land in the same package-level tally the gate reads.
func TestFuzzUnverifiedPath(t *testing.T) {
	releaseGate(t)
	bin := buildBinary(t)
	wrapped := debateWrapped(t)
	runDir, err := os.MkdirTemp("", "fuzz-unverified-")
	if err != nil {
		// Not swallowed: an empty runDir resolves every path below against the package
		// directory, which turns a full TMPDIR into a failure that names anything but the disk.
		t.Fatalf("temp dir for the run: %v", err)
	}
	// RELEASE, THEN REMOVE — in that order, and in one defer so the order is the order it reads
	// in. Two defers express it backwards (LIFO), which is the same "the two lines matter in the
	// opposite order to how they read" trap that has now been walked into eleven times. The sweep
	// in runOne releases; these dedicated-path tests copied its MkdirTemp/RemoveAll shape without
	// it, and on Linux the removal succeeds anyway so nothing said a word (#666).
	defer func() {
		_ = recordsql.CloseUnder(runDir)
		_ = os.RemoveAll(runDir)
	}()
	// THE SAME RUN THE SWEEP BUILDS, because driveDebate is only half of one. Without the seeded
	// report a lens finding has no anchor quote to attach to and is refused, so red mints nothing
	// and round 3 is a FAIL with an empty gaps array — the engine's degenerate-merge refusal,
	// which is correct and is not this path. And without INGESTING the report (#709) every verb that
	// reads it is refused with "no base has been ingested". runOne does both before it drives; so
	// does this.
	if err := writeSetup(filepath.Join(runDir, "blue", "report.md"), []byte(seedReport)); err != nil {
		t.Fatal(err)
	}
	r := newRunner(bin, runDir, newLockedRand(1))
	// THE CAST SETUP WRITES, as the halt path and the sweep write it. Without it every `dispatch
	// next` is refused ("the record holds no cast"), the driver's fallback plan dispatches nobody,
	// and the run reached UNVERIFIED without one seat dispatched — through the refusal, not the path.
	unverifiedCast, unverifiedLanes := record.CastFor(3)
	if _, err := record.Append(record.Identity{Run: r.run(), SeatID: record.HarnessSeat}, &recordpb.Cast{SeatIds: unverifiedCast, LaneSeatIds: unverifiedLanes}); err != nil {
		t.Fatalf("write the cast: %v", err)
	}
	r.register("blue", "blue-synthesize")
	if _, err := r.exec("ingest", "--seat-id", "blue-synthesize"); err != nil {
		t.Fatalf("ingest the round-0 report: %v", err)
	}
	r.forceUnverified = true

	result, settledErr := driveDebate(r, wrapped)
	if settledErr != "" {
		t.Fatalf("unverified run did not settle cleanly: %s", settledErr)
	}
	if v, _ := result["verdict"].(string); v != "UNVERIFIED" {
		t.Fatalf("expected verdict UNVERIFIED, got %q — the terminal path this test exists to drive did not run", v)
	}
	// AND THE RECORD SAYS SO, not just the returned map. This is the assertion the enum gate
	// actually depends on: `outcome --as UNVERIFIED` has to be a value the tool WROTE.
	//
	// It asserted `len(r.openGaps()) > 0` first, on the reasoning that UNVERIFIED-with-nothing-
	// open is a contradiction. That reasoning is debate.js's and it holds where debate.js
	// applies it — at the decision — but not here: the bench disposes the docket in its terminal
	// sitting, so an empty board AFTER the run is ordinary. Measured, it failed 12 of 12 while
	// the verdict itself was right every time, which makes it an assertion about the wrong
	// moment rather than a defect it caught.
	board, err := record.FamilyOf(runtest.Open(t, runDir))
	if err != nil {
		t.Fatalf("board: %v", err)
	}
	recorded, basis := "", ""
	for _, e := range board.Events {
		if o, ok := recordpb.BodyAs[*recordpb.Outcome](e); ok {
			recorded = strings.ToUpper(recordpb.Word(o.GetVerdict()))
			basis = o.GetVerdictBasis()
		}
	}
	if recorded != "UNVERIFIED" {
		t.Fatalf("the record carries outcome %q, not UNVERIFIED — the enum value this test exists to write was not written", recorded)
	}
	// THE BASIS, not only the stamp. UNVERIFIED is the one verdict the record cannot derive, so
	// the outcome that carries it must say it was ASSERTED — a derived UNVERIFIED would mean the
	// derivation invented a terminal state the record does not hold.
	if basis != record.VerdictAsserted {
		t.Fatalf("outcome verdict_basis %q, expected %q — UNVERIFIED was reached, but not as the asserted word this path drives", basis, record.VerdictAsserted)
	}
}

// THE GUARD THAT WOULD HAVE NAMED #637's 29 RUNS. Every case below reached the sweep's
// distribution as `verdicts[""]++` before this change — a bucket that counts like a verdict,
// sorts like a verdict, and satisfies a coverage gate like a verdict, while meaning "this side
// could not read the result".
//
// The gate is on readVerdict rather than on a driven debate deliberately: the miss is a
// PROPERTY OF THE READ, not of any particular run, and a test that had to reproduce the
// upstream conditions would be exactly as unreliable as the symptom it chases.
func TestReadVerdictRefusesAResultThatCarriesNoVerdict(t *testing.T) {
	for _, tc := range []struct {
		name    string
		result  map[string]any
		wantErr string // a substring the message must carry
	}{
		{
			// The measured shape: driveDebate returned (nil, "") because the promise fulfilled
			// with something that was not an object, and every caller read that as success.
			name:    "a nil result is not a settled debate",
			result:  nil,
			wantErr: "no verdict",
		},
		{
			name:    "a result carrying other keys but no verdict",
			result:  map[string]any{"epochs": int64(3), "runDir": "/tmp/x"},
			wantErr: "no verdict",
		},
		{
			name:    "a verdict that is not a string",
			result:  map[string]any{"verdict": 7},
			wantErr: "non-string verdict",
		},
		{
			name:    "an explicitly empty verdict",
			result:  map[string]any{"verdict": ""},
			wantErr: "EMPTY verdict",
		},
	} {
		got, err := readVerdict(tc.result)
		if err == nil {
			t.Errorf("%s: readVerdict returned %q and NO error — the miss is still indistinguishable "+
				"from a terminal verdict", tc.name, got)
			continue
		}
		if !strings.Contains(err.Error(), tc.wantErr) {
			t.Errorf("%s: error = %q, want it to mention %q", tc.name, err, tc.wantErr)
		}
		if got != "" {
			t.Errorf("%s: readVerdict returned %q alongside an error; a refused read yields nothing", tc.name, got)
		}
	}
}

// AND THE FOUR REAL VERDICTS STILL PASS. debate.js computes exactly these; a guard that also
// rejected one of them would turn a working sweep red and be indistinguishable, from the
// outside, from the bug it replaced.
func TestReadVerdictAcceptsEveryVerdictDebateJSComputes(t *testing.T) {
	for _, want := range []string{"HALTED", "VERIFIED", "CEILING", "UNVERIFIED"} {
		got, err := readVerdict(map[string]any{"verdict": want, "epochs": int64(1)})
		if err != nil {
			t.Errorf("readVerdict rejected %q, which debate.js line 1166 can produce: %v", want, err)
		}
		if got != want {
			t.Errorf("readVerdict(%q) = %q", want, got)
		}
	}
}

// The keys are named in the failure, and named STABLY — two failures of one shape must produce
// one message, or a reader cannot tell a recurring defect from a family of them.
func TestSortedKeysIsStableAcrossMapIterationOrder(t *testing.T) {
	m := map[string]any{"runDir": "", "verdict": "", "epochs": 0, "lanes": 0}
	first := fmt.Sprint(sortedKeys(m))
	for i := 0; i < 50; i++ {
		if got := fmt.Sprint(sortedKeys(m)); got != first {
			t.Fatalf("sortedKeys is order-dependent: %s vs %s", got, first)
		}
	}
	if first != "[epochs lanes runDir verdict]" {
		t.Errorf("sortedKeys = %s, want sorted order", first)
	}
}

// unruledDocketMotions reads the motions projection for every docket motion on gapID that has no
// ruling yet, in the projection's order — the one the dispatch filed, and any a party filed beside it.
func (r *runner) unruledDocketMotions(gapID string) []string {
	// `show` renders json by default (--format); the global --json envelope is not its shape.
	out, err := r.exec(seat.GroupOf("motions"), "motions", "--seat-id", "judge")
	if err != nil {
		return nil
	}
	// `show --json` writes the projection's bytes directly — record.MotionsJSON, not a seat
	// envelope — so the motions sit at the top level.
	var page struct {
		Motions []struct {
			ID      string `json:"id"`
			Subject string `json:"subject"`
			Ruled   bool   `json:"ruled"`
			GapID   string `json:"gap_id"`
		} `json:"motions"`
	}
	if json.Unmarshal([]byte(out), &page) != nil {
		return nil
	}
	var ids []string
	for _, m := range page.Motions {
		if m.Subject == "docket" && !m.Ruled && m.GapID == gapID {
			ids = append(ids, m.ID)
		}
	}
	return ids
}

// docketFiler is the seat that files the fallback docket motion the bench then rules: the chair,
// which sits before the bench does. Never the bench — it rules docket motions and files none, so a
// filing in its name is refused and the drive would rule nothing.
const docketFiler = "red-chair"

// settledByFuzzBench is the proposition every fuzz ruling bars, on the record and on the envelope.
const settledByFuzzBench = "the proposition this ruling bars"

// benchDisposes puts a gap before the bench and rules it, which is TWO acts now.
//
// IT WAS ONE, and the difference is the point of #681: `bench opinion --id <gap>` needed only a
// gap, so the disposition joined to nothing and nothing could ask whether a gap that reached the
// bench was ever answered. A ruling names the MOTION; the gap rides the filing.
//
// THE FILING'S RESULT IS READ RATHER THAN DISCARDED. Both halves used to be `r.admit(...)`.
// noteExec records a refusal either way, so the run was not blind — but the id is needed here, and
// a filing that failed would otherwise be followed by a ruling naming an empty motion, which
// refuses for a second reason and buries the first. Returning early leaves ONE refusal in
// execFailWhy: the one that actually happened.
func (r *runner) benchDisposes(seatID, gapID, disposition string, extra ...string) {
	// THE DISPATCH DOCKETS (plans/roundless.md §III.B.1): `dispatch next` files the docket motion
	// for a gap at impasse, and the bench rules THAT motion — the record re-readies nothing for a
	// docketing whose motion stands unruled, so a bench that files a fresh motion and rules it
	// has, on the record, sat and ruled nothing. Five runs in forty ended UNVERIFIED on exactly
	// that before this read the board. The fallback filing stays for a gap no dispatch docketed,
	// and it files as the CHAIR seat, never the bench (the gavel problem in miniature).
	//
	// EVERY DOCKET MOTION STANDING ON THE GAP IS RULED, not the first. A party may file one beside the
	// dispatch's (blue's `motion docket file` drive does not look first), each is a blocker on the
	// bench's work list, and the plan keys the bench's sitting on the newest filing: a bench that
	// rules one of two has, on the record, sat for the other and ruled nothing, and is never readied
	// again — the forced-VERIFIED run ended UNVERIFIED on exactly that, the sibling motion standing
	// with nobody ready until the terminal sitting ruled it. Two remands in one sitting are one
	// remand (remandSittingsOf), so ruling each sibling as the scenario says moves no limit.
	motionIDs := r.unruledDocketMotions(gapID)
	// THE BENCH FILES NO DOCKET MOTION: reaching for one on the gap it is about to rule is refused.
	if r.coin(25) {
		r.rulerFilingRefused(seatID, "motion", "docket", "file", "--id", gapID, "--reason", "fuzz: the bench docketing "+gapID+" to itself")
	}
	if len(motionIDs) == 0 {
		out, err := r.exec("motion", "docket", "file", "--json", "--seat-id", docketFiler, "--id", gapID,
			"--reason", "fuzz: "+gapID+" is contested and is the bench's to settle")
		if err != nil {
			return // recorded by noteExec under `motion docket file`
		}
		// THE RESULT IS NESTED IN THE ENVELOPE ({"ok":true,"result":{…}}); reading `motion_id` at
		// the top unmarshals cleanly into a zero, and the ruling below is silently skipped.
		var filed struct {
			Result struct {
				MotionID string `json:"motion_id"`
			} `json:"result"`
		}
		if json.Unmarshal([]byte(out), &filed) != nil || filed.Result.MotionID == "" {
			return
		}
		motionIDs = []string{filed.Result.MotionID}
	}
	for i, motionID := range motionIDs {
		// THE BENCH TRIES TO END WITH A TWIN STANDING, and the record refuses its envelope naming
		// the twin (record.BenchEnvelopeRefusal — the PreToolUse hook denies the envelope call with
		// it). Driven before the last of two or more motions on the gap is ruled.
		if i == len(motionIDs)-1 && i > 0 && r.coin(50) {
			r.envelopeRefusedEarly(seatID, "motion "+motionID+" ")
		}
		args := append([]string{"motion", "docket", "rule", "--seat-id", seatID, "--id", motionID,
			"--as", disposition, "--principle", "correctness", "--tension", "cost",
			"--review-flag", "none: the scenario the gap came from settles it", "--settled", settledByFuzzBench}, extra...)
		r.admit(args...)
	}
}

// envelopeRefusedEarly drives a seat's envelope while an item it owes stands: the record must refuse
// it (record.EnvelopeRefusal — the PreToolUse hook denies the envelope call with it), naming the item.
func (r *runner) envelopeRefusedEarly(seatID, standing string) {
	refusal, err := record.EnvelopeRefusal(r.run(), seatID)
	switch {
	case err != nil:
		r.noteActErr(seatID + "'s envelope could not be checked against the record: " + err.Error())
	case !strings.Contains(refusal, standing):
		r.noteActErr(fmt.Sprintf("%s's envelope with %q standing was not refused naming it: %q", seatID, standing, refusal))
	default:
		noteEnvelope(seatID, "refused, driven before its last item was cleared")
	}
}

// envelopeEnds is every seat's envelope at the end of its drive: a seat's sitting ends only once
// nothing it owes stands (record.EnvelopeRefusal), so a drive that returns its envelope with an item
// standing is a drive that leaves undone what the record would refuse the seat for.
func (r *runner) envelopeEnds(seatID string) {
	refusal, err := record.EnvelopeRefusal(r.run(), seatID)
	switch {
	case err != nil:
		r.noteActErr(seatID + "'s envelope could not be checked against the record: " + err.Error())
	case refusal != "":
		noteEnvelope(seatID, "refused at the end of its drive")
		r.mu.Lock()
		why := r.owedRecheck[seatID]
		r.mu.Unlock()
		if why == "" {
			why = "no re-check of an owed anchor was refused"
		}
		r.noteActErr(seatID + " ended its drive with an item it owes standing, and its envelope was refused: " + refusal + " [drive: " + why + "]")
	default:
		noteEnvelope(seatID, "admitted")
	}
}

// ruleStandingDockets is the terminal bench ruling every docket motion that stands unruled at the
// exit, each with a disposition that ends it: the terminal prompt names them and says nothing can
// be remanded — there is no next sitting to remand it to.
func (r *runner) ruleStandingDockets(seatID string) {
	out, err := r.exec(seat.GroupOf("motions"), "motions", "--seat-id", seatID)
	if err != nil {
		return // recorded by noteExec; the outcome oracle names the run that ends over it
	}
	var page struct {
		Motions []struct {
			ID      string `json:"id"`
			Subject string `json:"subject"`
			Ruled   bool   `json:"ruled"`
		} `json:"motions"`
	}
	if json.Unmarshal([]byte(out), &page) != nil {
		return
	}
	for _, m := range page.Motions {
		if m.Subject != "docket" || m.Ruled {
			continue
		}
		disp := pick(r.rng, []string{"repaired", "not_a_defect", "defect_accepted", "defect_owed_elsewhere", "amends_prior"})
		r.admit("motion", "docket", "rule", "--seat-id", seatID, "--id", m.ID,
			"--as", disp, "--final", "--principle", "correctness", "--tension", "cost",
			"--review-flag", "none: the run's term ended with this motion standing", "--settled", settledByFuzzBench,
			"--reason", "fuzz: ruled at the exit, where nothing can be remanded")
	}
}

// forcedWhyClause renders the forced run's exit reason when it has one. Split out because the
// reason is the diagnosis — "PASS refused: … 3 motion(s) filed and never ruled" is what the shape
// is made of — and an empty one must read as absent rather than as an empty pair of quotes.
func forcedWhyClause(why string) string {
	if why == "" {
		return " The run stated no exit reason."
	}
	return " It exited on: " + why
}

// declarationsAgreeWithBoard is "" when every role's work list carries the run's declared count of
// declarations and each gap one names reads as the board states that gap; otherwise what disagreed.
func declarationsAgreeWithBoard(run record.Run, declared int) string {
	bj, err := record.BoardJSONOfRun(run)
	if err != nil {
		return "declarations oracle: the board does not read: " + err.Error()
	}
	board := map[string]record.GapJSON{}
	for _, g := range append(append([]record.GapJSON{}, bj.Open...), bj.Closed...) {
		board[g.ID] = g
	}
	for _, rs := range [][2]string{{"blue", "blue-respond"}, {"lens", "red-lens-evidence"}, {"chair", "red-chair"}, {"bench", "judge"}} {
		w, err := record.WorkOfSeat(run, rs[0], rs[1])
		if err != nil {
			return "declarations oracle: " + rs[1] + "'s work list does not read: " + err.Error()
		}
		if len(w.Declarations) != declared {
			return fmt.Sprintf("declarations oracle: %s's work list carries %d declarations; the record accepted %d", rs[1], len(w.Declarations), declared)
		}
		for _, d := range w.Declarations {
			for _, dg := range d.Gaps {
				g, onBoard := board[dg.ID]
				if !onBoard {
					return fmt.Sprintf("declarations oracle: %s's list names gap %s, which the board does not hold — the write admits only a held gap", rs[1], dg.ID)
				}
				want := record.DeclaredGapJSON{ID: dg.ID, Status: "open"}
				if !g.Open {
					want = record.DeclaredGapJSON{ID: dg.ID, Status: "closed", Fate: dg.Fate, ClosedBy: "red", ClosedEpoch: g.ClosedEpoch}
					if g.ClosedByBench {
						want.ClosedBy = "bench"
					}
					if dg.Fate == "" {
						return fmt.Sprintf("declarations oracle: %s's list names closed gap %s with no fate", rs[1], dg.ID)
					}
				}
				if dg != want {
					return fmt.Sprintf("declarations oracle: %s's list reads declared gap %+v; the board says %+v", rs[1], dg, want)
				}
			}
		}
	}
	return ""
}

// sumSittings and maxSittings are the RUNSTAT line's lens-sitting cost: the run's total, and the
// most any one lens sat.
func sumSittings(m map[string]int) (n int) {
	for _, v := range m {
		n += v
	}
	return n
}

func maxSittings(m map[string]int) (n int) {
	for _, v := range m {
		n = max(n, v)
	}
	return n
}

// unfinishedOverRedsWork is the verdict oracle's converse check: a run that ended neither VERIFIED nor
// HALTED over a board red itself emptied — every gap closed, none by the bench — while nothing on the
// record's plan holds the PASS that another seat must clear.
//
// ONLY WHERE NO BLOCKER ANOTHER SEAT OWNS STANDS. A blocker the chair cannot clear is work red has not
// finished, whatever the gaps say: a reopened citation or proof its owner sat for and left, and a lens
// still ready at the epoch limit — readied by a head move it has not sat at, owing an unjudged proof
// or below its barren limit. The final sitting dispatches nobody and the run ends CEILING over that
// ready lens, which is the limit's meaning. Seed 26 (fuzz on #1446's pre-merge head) ended CEILING at
// epoch 12 of 12 with every gap closed by red and one blocker, "lens red-lens-computation is ready":
// it held eleven proofs nobody had judged and the last edit had moved the head past its pin.
func unfinishedOverRedsWork(recorded string, gaps, open, benchClosed int, blockers []record.PlanBlocker) bool {
	if recorded == "" || recorded == "VERIFIED" || recorded == "HALTED" || gaps == 0 || open > 0 || benchClosed > 0 {
		return false
	}
	return !slices.ContainsFunc(blockers, func(b record.PlanBlocker) bool { return b.Owner != "red-chair" })
}

// THE CONVERSE VERDICT ORACLE FIRES ON A BOARD RED EMPTIED AND ON NO OTHER. Seed 26's board — CEILING
// at the epoch limit, every gap closed by red, a lens ready — is work red had not finished, and so is a
// reopened anchor its owner left; a board whose only blockers are the chair's own is the contradiction.
func TestTheConverseVerdictOracleFiresOnlyOverWorkRedFinished(t *testing.T) {
	ready := record.PlanBlocker{Kind: record.BlockerLensReady, Subject: "red-lens-computation", Owner: "red-lens-computation"}
	reopened := record.PlanBlocker{Kind: record.BlockerUnchecked, Subject: "P-1", Owner: "red-lens-computation"}
	stale := record.PlanBlocker{Kind: record.BlockerStaleArea, Subject: "red-lens-voice", Owner: "red-chair"}
	for _, c := range []struct {
		name                    string
		recorded                string
		gaps, open, benchClosed int
		blockers                []record.PlanBlocker
		fires                   bool
	}{
		{"seed 26: CEILING at the limit with a lens ready", "CEILING", 9, 0, 0, []record.PlanBlocker{ready}, false},
		{"a reopened anchor its owner left", "UNVERIFIED", 3, 0, 0, []record.PlanBlocker{reopened}, false},
		{"only the chair's own blockers", "CEILING", 3, 0, 0, []record.PlanBlocker{stale}, true},
		{"no blockers at all", "UNVERIFIED", 3, 0, 0, nil, true},
		{"a gap still open", "CEILING", 3, 1, 0, nil, false},
		{"a gap the bench closed", "CEILING", 3, 0, 1, nil, false},
		{"no gaps", "CEILING", 0, 0, 0, nil, false},
		{"VERIFIED", "VERIFIED", 3, 0, 0, nil, false},
		{"HALTED", "HALTED", 3, 0, 0, nil, false},
	} {
		if got := unfinishedOverRedsWork(c.recorded, c.gaps, c.open, c.benchClosed, c.blockers); got != c.fires {
			t.Errorf("%s: fires = %v, want %v", c.name, got, c.fires)
		}
	}
}

// engageSuccessors is blue's R2 act (gblock, 2026-10-23): where its work list names an engaged gap it
// found closed with an open successor, it records `engage` on exactly those and relays them in its
// envelope, and the engine sits it once more this epoch. Before the owed set, once per sweep, it
// asks with a gap the record names as no successor — the refusal the ruling requires.
func (r *runner) engageSuccessors(seatID string, env map[string]any) {
	w, err := record.WorkOfSeat(r.run(), "blue", seatID)
	if err != nil || w.Engaged == nil {
		return
	}
	var closed, lineage []string
	for _, f := range w.Engaged.FoundClosed {
		closed = append(closed, f.ID)
		lineage = append(lineage, f.Successors...)
	}
	succ := w.Engaged.SuccessorsOwed
	if len(closed) == 0 {
		return
	}
	found := []any{}
	for _, id := range closed {
		found = append(found, id)
	}
	env["found_closed"] = found
	if len(succ) == 0 {
		return
	}
	if engageRefusalPending.CompareAndSwap(true, false) {
		named := map[string]bool{}
		for _, g := range lineage {
			named[g] = true
		}
		for _, other := range r.openGaps() {
			if !named[other] {
				r.refuse(feov.Validation, "engage", "--seat-id", seatID, "--ids", other)
				break
			}
		}
	}
	if _, err := r.exec("engage", "--seat-id", seatID, "--ids", strings.Join(succ, ",")); err != nil {
		return
	}
	out := []any{}
	for _, g := range succ {
		out = append(out, g)
	}
	env["successors"] = out
	r.mu.Lock()
	r.againOn = succ
	r.mu.Unlock()
}

// engageRefusalPending drives the engage refusal once per sweep.
var engageRefusalPending atomic.Bool
