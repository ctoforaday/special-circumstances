package fuzz

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/capture"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	reportdoc "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// lifecycleSeed is the seed whose run is FORCED through the anchor lifecycle's acts and refusals
// (plans/compute-and-show.md III.1′.7), at blue-respond's first sitting (forcedLifecycle). It is none
// of the other forced seeds.
//
// Left to the draw, each arm is a chain of coins no sweep reaches reliably: a computation is retired only
// where an edit would leave it bare, the drawn owed gaps are minted `--check-kind document`, so no
// demand reaches them, and a lens re-checks an anchor only in a later sitting, after blue has remedied
// it. The drive runs every arm in one sitting, opening each lens a sitting of its own wherever the
// record keeps one check of an anchor per sitting.

// lifecycleDriven is lifecycleStage once forcedLifecycle has driven every step.
const lifecycleDriven = "driven"

// otherLens is the lens whose check of another lens's anchor meets nothing of that lens's demand.
const otherLens = "red-lens-logic"

// lifecycleClaims are the sentences the forced drive adds after sourceDemandOpen, one per arm: no
// other drive quotes them, so no gap the drive mints meets another gap of its class on its sentence.
// The last two stand together, for the claim retire that cuts the first onto the second.
var lifecycleClaims = [...]string{
	"The first lifecycle claim holds",
	"The second lifecycle claim holds",
	"The third lifecycle claim holds",
	"The fourth lifecycle claim holds",
	"The fifth lifecycle claim holds",
	"The sixth lifecycle claim holds",
	"The seventh lifecycle claim holds",
	"The eighth lifecycle claim holds",
	"The ninth lifecycle claim holds",
	"The tenth lifecycle claim holds",
	"The eleventh lifecycle claim holds",
	"The twelfth lifecycle claim holds",
}

// lifecycleCounts are the anchor lifecycle's counts (III.1′.7), each a tally gate over the sweep.
type lifecycleCounts struct {
	// computationRetires: `retire --anchor K-` admitted.
	computationRetires int
	// sourceAnswers: `cite --answers G` admitted.
	sourceAnswers int
	// closedAnswersRetired: a closed gap the board reads closed, with `answers_retired_event` set, after
	// the retire of its last answer; retiredItemsCleared: its lens's item gone once the lens minted a gap
	// superseding it.
	closedAnswersRetired, retiredItemsCleared int
	// retiredAboutCloses: the lens's close admitted of a `source` or `computation` gap about an anchor
	// blue retired naming no gap (L1).
	retiredAboutCloses int
	// citedRetiresRefused: a retire taking a citation a live computation applies, refused naming it (L6).
	citedRetiresRefused int
	// readingDemandMetByRetire and readingDemandMetByReread: the close of a `source` gap minted about a
	// citation on the lens's refuting reading, admitted after blue's retire of the citation met its
	// demand through the gap's about arm, and after the minter's own backing re-read met it.
	readingDemandMetByRetire, readingDemandMetByReread int
	// unreadDemandReread and unreadDemandRefused: the close of a `source` gap owed about an unread
	// citation, admitted after the reader's own backing reading, and refused after another lens's.
	unreadDemandReread, unreadDemandRefused int
	// unsoundDemandRejudged and unsoundDemandRefused: the close of a `computation` gap owed about an
	// unsound computation, admitted after the lens's own `sound` judgement, and refused after another
	// lens's.
	unsoundDemandRejudged, unsoundDemandRefused int
	// retiredAboutMintsRefused: a mint about a retired evidence anchor, refused naming the retire.
	retiredAboutMintsRefused int
	// reachingRetireTurns: a plan engaging a gap's minting lens whose turn a reaching retire gave it.
	reachingRetireTurns int
	// screenedUnsound: an assembly screen FAIL naming a live computation a standing judgement calls
	// unsound, at the end of a run that ends with one.
	screenedUnsound int
}

// lifecycleTally is lifecycleCounts as a run's tally file carries it.
type lifecycleTally struct {
	ComputationRetires       int `json:"computation_retires"`
	SourceAnswers            int `json:"source_answers"`
	ClosedAnswersRetired     int `json:"closed_answers_retired"`
	RetiredItemsCleared      int `json:"retired_items_cleared"`
	RetiredAboutCloses       int `json:"retired_about_closes"`
	CitedRetiresRefused      int `json:"cited_retires_refused"`
	ReadingDemandMetByRetire int `json:"reading_demand_met_by_retire"`
	ReadingDemandMetByReread int `json:"reading_demand_met_by_reread"`
	UnreadDemandReread       int `json:"unread_demand_reread"`
	UnreadDemandRefused      int `json:"unread_demand_refused"`
	UnsoundDemandRejudged    int `json:"unsound_demand_rejudged"`
	UnsoundDemandRefused     int `json:"unsound_demand_refused"`
	RetiredAboutMintsRefused int `json:"retired_about_mints_refused"`
	ReachingRetireTurns      int `json:"reaching_retire_turns"`
	ScreenedUnsound          int `json:"screened_unsound"`
}

func (c lifecycleCounts) tally() lifecycleTally {
	return lifecycleTally{
		ComputationRetires: c.computationRetires, SourceAnswers: c.sourceAnswers,
		ClosedAnswersRetired: c.closedAnswersRetired, RetiredItemsCleared: c.retiredItemsCleared,
		RetiredAboutCloses: c.retiredAboutCloses, CitedRetiresRefused: c.citedRetiresRefused,
		ReadingDemandMetByRetire: c.readingDemandMetByRetire, ReadingDemandMetByReread: c.readingDemandMetByReread,
		UnreadDemandReread: c.unreadDemandReread, UnreadDemandRefused: c.unreadDemandRefused,
		UnsoundDemandRejudged: c.unsoundDemandRejudged, UnsoundDemandRefused: c.unsoundDemandRefused,
		RetiredAboutMintsRefused: c.retiredAboutMintsRefused, ReachingRetireTurns: c.reachingRetireTurns,
		ScreenedUnsound: c.screenedUnsound,
	}
}

// add folds another run's tally into t.
func (t *lifecycleTally) add(o lifecycleTally) {
	t.ComputationRetires += o.ComputationRetires
	t.SourceAnswers += o.SourceAnswers
	t.ClosedAnswersRetired += o.ClosedAnswersRetired
	t.RetiredItemsCleared += o.RetiredItemsCleared
	t.RetiredAboutCloses += o.RetiredAboutCloses
	t.CitedRetiresRefused += o.CitedRetiresRefused
	t.ReadingDemandMetByRetire += o.ReadingDemandMetByRetire
	t.ReadingDemandMetByReread += o.ReadingDemandMetByReread
	t.UnreadDemandReread += o.UnreadDemandReread
	t.UnreadDemandRefused += o.UnreadDemandRefused
	t.UnsoundDemandRejudged += o.UnsoundDemandRejudged
	t.UnsoundDemandRefused += o.UnsoundDemandRefused
	t.RetiredAboutMintsRefused += o.RetiredAboutMintsRefused
	t.ReachingRetireTurns += o.ReachingRetireTurns
	t.ScreenedUnsound += o.ScreenedUnsound
}

// lifecycleGates is the sweep's gate on each count, with what a zero means (III.1′.7's tally column,
// sourceClosesRefused gated beside the reading counts).
func (t lifecycleTally) gates() []struct {
	n    int
	what string
} {
	return []struct {
		n    int
		what string
	}{
		{t.ComputationRetires, "retires of a computation (`retire --anchor K-`) admitted"},
		{t.SourceAnswers, "citations placed as a source gap's answer (`cite --answers`) admitted"},
		{t.ClosedAnswersRetired, "closed gaps whose answers were all retired read closed with answers_retired_event set"},
		{t.RetiredItemsCleared, "answers-retired items gone once the lens minted the gap superseding the closed one"},
		{t.RetiredAboutCloses, "closes admitted of a source or computation gap about an anchor blue retired naming no gap (L1)"},
		{t.CitedRetiresRefused, "retires of a citation a live computation applies, refused naming the computation (L6)"},
		{t.ReadingDemandMetByRetire, "closes admitted of a source gap minted about a citation on a refuting reading, its demand met by blue's retire of the citation"},
		{t.ReadingDemandMetByReread, "closes admitted of a source gap minted about a citation on a refuting reading, its demand met by the minter's own backing re-read"},
		{t.UnreadDemandReread, "closes admitted of a source gap owed about an unread citation, after the reader's own backing reading"},
		{t.UnreadDemandRefused, "closes of a source gap owed about an unread citation refused by its demand after another lens's backing reading"},
		{t.UnsoundDemandRejudged, "closes admitted of a computation gap owed about an unsound computation, after the lens's own sound judgement"},
		{t.UnsoundDemandRefused, "closes of a computation gap owed about an unsound computation refused by its demand after another lens's sound judgement"},
		{t.RetiredAboutMintsRefused, "mints about a retired evidence anchor refused naming the retire"},
		{t.ReachingRetireTurns, "plans engaging a gap's minting lens whose turn a reaching retire gave it"},
		{t.ScreenedUnsound, "assembly screens FAILING on a live unsound computation at a run's end"},
	}
}

// forcedLifecycle is lifecycleSeed's arm, at blue-respond's first sitting, before any act of its own
// (lifecycleSeed). It draws nothing from the run's RNG.
func (r *runner) forcedLifecycle(seatID string) {
	r.mu.Lock()
	due := r.forceLifecycle && seatID == "blue-respond" && r.lifecycleStage == ""
	if due {
		r.lifecycleStage = "begun"
	}
	r.mu.Unlock()
	if !due {
		return
	}
	stage := r.driveLifecycle(seatID)
	r.mu.Lock()
	r.lifecycleStage = stage
	r.mu.Unlock()
}

// lifecycleForcedErr is why lifecycleSeed's run failed its arm, or "" where it drove it or is not
// forced. Called with r.mu held.
func (r *runner) lifecycleForcedErr() string {
	switch {
	case !r.forceLifecycle:
		return ""
	case r.lifecycleStage == "":
		return "forced anchor lifecycle: blue-respond never sat"
	case r.lifecycleStage != lifecycleDriven:
		return "forced anchor lifecycle: " + r.lifecycleStage
	case r.lifecycle.reachingRetireTurns == 0:
		return "forced anchor lifecycle: no plan engaged " + citationOwner + " on gap " + r.lifecycleLeftOpen + ", whose turn blue's reaching retire gave it"
	}
	return ""
}

// openSitting opens seat a sitting of its own, which no dispatch carries: the record keeps one check of
// an anchor per seat per sitting, and the drive's re-checks are the seat's own.
func (r *runner) openSitting(seat string) bool {
	_, ok := r.admit(registerArgs(seat, "")...)
	return ok
}

// lifecycleMint is seat's mint of a gap under check kind on claim, or about anchor where about is set,
// graded grade on every axis, with extra flags; it returns the gap id, or "" and the refusal.
func (r *runner) lifecycleMint(seat, kind, claim, about, grade string, extra ...string) (string, string) {
	args := []string{"--json", "mint", "--seat-id", seat, "--class", mintPoolClasses[0],
		"--problem", "fuzz: the lifecycle arm's defect", "--check-kind", kind, "--check", "the evidence settles it",
		"--fix", dirCounter, "--severity", grade, "--likelihood", grade, "--impact", grade, "--reason", "fuzz: the lifecycle arm"}
	if about != "" {
		args = append(args, "--about-kind", "anchor", "--about", about)
	} else {
		args = append(args, "--quote", claim)
	}
	out, err := r.exec(append(args, extra...)...)
	var minted struct {
		Result struct {
			GapID string `json:"gap_id"`
		} `json:"result"`
	}
	if err != nil || json.Unmarshal([]byte(strings.TrimSpace(out)), &minted) != nil || minted.Result.GapID == "" {
		return "", firstLine(out)
	}
	r.noteMint(seat, minted.Result.GapID)
	return minted.Result.GapID, ""
}

// lifecycleClose is seat's close of gap id as repaired; err is the refusal, nil where admitted.
func (r *runner) lifecycleClose(seat, id string) (feov.RefusalEnvelope, bool) {
	out, err := r.exec("close", "--json", "--seat-id", seat, "--id", id, "--as", "repaired", "--reason", "fuzz: the lifecycle arm's close",
		"--verified-by", seat, "--verified-with", "fuzz", "--verified-against", "rec")
	if err == nil {
		return feov.RefusalEnvelope{}, true
	}
	env, ok := refusalEnvelopeIn(out)
	if !ok {
		env = feov.RefusalEnvelope{Error: firstLine(out)}
	}
	return env, false
}

// demandRefusal is whether env is the live-answer demand's refusal of gap id, minted under kind.
func demandRefusal(env feov.RefusalEnvelope, id, kind string) bool {
	return env.Code == string(feov.Conflict) && strings.Contains(env.Error, "gap "+id+" was minted --check-kind "+kind) &&
		strings.Contains(env.Error, "nothing meets its live-answer demand")
}

// lifecycleCite is blue's citation of claim, of a source of its own, answering gap where gap is set; it
// returns the citation's label, or "".
func (r *runner) lifecycleCite(seat, claim, path, gap string) string {
	c := r.do("cite", seat).set("--quote", claim).set("--url", sourceURL(path)).set("--title", sourceTitle).
		set("--source-span", sourceBody(path)).set("--reason", "fuzz: the source the claim rests on").bare("--json")
	if gap != "" {
		c.set("--answers", gap)
	}
	out, ok := c.admit()
	var label struct {
		Result struct {
			Label string `json:"label"`
		} `json:"result"`
	}
	if !ok || json.Unmarshal([]byte(strings.TrimSpace(out)), &label) != nil {
		return ""
	}
	if gap != "" {
		r.mu.Lock()
		r.lifecycle.sourceAnswers++
		r.mu.Unlock()
	}
	return label.Result.Label
}

// lifecycleCompute is blue's computation of claim, of a bundle of its own, answering gap and applying
// the method of citation where each is set; it returns the computation's id, or "".
func (r *runner) lifecycleCompute(seat, claim, name, gap, cites string) string {
	dir, argv, err := r.writeBundle(name, "console.log('fuzz: "+name+"');")
	if err != nil {
		return ""
	}
	c := r.do("compute", seat).set("--quote", claim).set("--dir", dir).set("--reason", "fuzz: the computation the claim rests on").runs(argv)
	if gap != "" {
		c.set("--answers", gap)
	}
	if cites != "" {
		c.set("--cites", cites)
	}
	if _, ok := c.admit(); !ok {
		return ""
	}
	return lastComputationOn(r.run(), claim)
}

// lifecycleRetire is blue's retire of evidence anchor id alone, answering gap where it is set.
//
// A CITATION'S RETIRE ASKS THE RECORD FIRST which live computations apply it (retireCitingFirst):
// blue's drawn `compute --cites` takes any live citation off the record (someCitation), a seed's own
// among them, and which one it took depends on the order the run's sittings drew the RNG in. The
// record refuses the citation's retire while one stands (L6), so the drive retires those first. CI run
// 42952536058 (seed 25) retired the reading-removal seed's citation a drawn computation applied.
func (r *runner) lifecycleRetire(seat, id, gap string) bool {
	if strings.HasPrefix(id, anchorPrefix["citation"]) && !r.retireCitingFirst(seat, id) {
		return false
	}
	c := r.do("retire", seat).set("--anchor", id).set("--reason", "fuzz: the evidence leaves its sentence")
	if gap != "" {
		c.set("--answers", gap)
	}
	if _, ok := c.admit(); !ok {
		return false
	}
	if strings.HasPrefix(id, anchorPrefix["computation"]) {
		r.mu.Lock()
		r.lifecycle.computationRetires++
		r.mu.Unlock()
	}
	return true
}

// lifecycleRead is seat's reading of citation id as outcome; after a refuting reading the seat mints
// its claim_contradicts_source `source` gap about the citation on that reading, whose id it returns.
func (r *runner) lifecycleRead(seat, id, outcome string) (gap string, ok bool) {
	r.firstThisSitting(seat, id)
	c := r.do("verify", seat).set("--anchor", id).set("--as", outcome).set("--confidence", "high").
		set("--reason", "fuzz: the lifecycle arm's reading").bare("--json")
	out, err := c.result()
	if err != nil {
		return "", false
	}
	v, at := r.checkReading(seat, out, c.args)
	if v == nil {
		return "", false
	}
	if outcome != "refutes" {
		return "", true
	}
	args := append(readingMintArgs(seat, classContradicts), aboutArgs(id)...)
	setArg(args, "--check-kind", "source")
	setArg(args, "--check", "a source states what the sentence states")
	setArg(args, "--severity", "low")
	setArg(args, "--likelihood", "low")
	setArg(args, "--impact", "low")
	mout, merr := r.exec(args...)
	if merr != nil {
		r.noteActErr("the claim_contradicts_source mint about " + id + " on " + seat + "'s refuting reading was refused: " + firstLine(mout))
		return "", false
	}
	gap = r.noteReadingMintLanded(seat, mout, classContradicts, "about")
	if gap == "" {
		return "", false
	}
	if _, got, has, err := record.ReadingBehind(r.run(), gap); err != nil || !has || got != at {
		r.noteActErr(fmt.Sprintf("the reading behind %s is event %d (found %v, %v), and the mint stood on the reading at event %d", gap, got, has, err, at))
	}
	return gap, r.debtDischargedBy(seat, id, "mint", "the lens's mint about the citation on its reading")
}

// lifecycleJudge is seat's judgement of computation id as `sound` or `unsound`.
func (r *runner) lifecycleJudge(seat, id, as string) bool {
	_, ok := r.do("reproduce", seat).set("--anchor", id).set("--as", as).set("--reason", "fuzz: the lifecycle arm's judgement").admit()
	return ok
}

func (r *runner) countLifecycle(f func(c *lifecycleCounts)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	f(&r.lifecycle)
}

// driveLifecycle drives forcedLifecycle's arms and returns how far it reached: lifecycleDriven, or the
// step that failed.
func (r *runner) driveLifecycle(blue string) string {
	const lens, comp = citationOwner, computationOwner
	c := lifecycleClaims
	if _, ok := r.do("edit", blue).set("--quote", sourceDemandOpen).set("--new", sourceDemandOpen+". "+strings.Join(c[:], ". ")).
		set("--reason", "fuzz: the claims the lifecycle arms rest on").admit(); !ok {
		return "blue's edit adding the claims was refused"
	}
	for _, seat := range []string{lens, comp, otherLens} {
		if !r.openSitting(seat) {
			return "the register of " + seat + " was refused"
		}
	}
	for _, arm := range []func() string{
		func() string { return r.lifecycleClosedAnswerRetired(blue, lens, c[0]) },
		func() string { return r.lifecycleComputationRetired(blue, comp, c[1]) },
		func() string { return r.lifecycleUnsoundGap(blue, comp, c[2]) },
		func() string { return r.lifecycleUnreadGap(blue, lens, c[3]) },
		func() string { return r.lifecycleReadingGap(blue, lens, c[4], c[5], c[6]) },
		func() string { return r.lifecycleRetiredAbout(blue, lens, comp, c[7], c[8]) },
		func() string { return r.lifecycleReachingRetire(blue, lens, c[9]) },
		func() string { return r.lifecycleCitedRetire(blue, c[10], c[11]) },
	} {
		if s := arm(); s != "" {
			return s
		}
	}
	return lifecycleDriven
}

// lifecycleClosedAnswerRetired: a source gap closed on blue's citation; blue retires the citation; the
// board reads the gap closed with answers_retired_event set, and the lens's list holds the item, blocking
// nothing; the lens mints the gap superseding it, and the item is gone.
func (r *runner) lifecycleClosedAnswerRetired(blue, lens, claim string) string {
	g, why := r.lifecycleMint(lens, "source", claim, "", "low")
	if g == "" {
		return "the source gap's mint was refused: " + why
	}
	if env, ok := r.lifecycleClose(lens, g); ok || !demandRefusal(env, g, "source") {
		return "the close of source gap " + g + " with nothing answering it was not refused by its demand: " + env.Error
	}
	r.countLifecycle(func(*lifecycleCounts) { r.sourceClosesRefused++ })
	cite := r.lifecycleCite(blue, claim, "/lifecycle-1", g)
	if cite == "" {
		return "blue's cite --answers " + g + " was refused"
	}
	if env, ok := r.lifecycleClose(lens, g); !ok {
		return "the close of source gap " + g + " on blue's citation was refused: " + env.Error
	}
	if !r.lifecycleRetire(blue, cite, "") {
		return "blue's retire of " + cite + ", the closed gap's answer, was refused"
	}
	bj, err := record.BoardJSONOfRun(r.run())
	if err != nil {
		return "the board could not be read: " + err.Error()
	}
	closed := slices.IndexFunc(bj.Closed, func(x record.GapJSON) bool { return x.ID == g })
	if closed < 0 || bj.Closed[closed].AnswersRetiredEvent == nil {
		return "gap " + g + " does not read closed with answers_retired_event set once its answer " + cite + " was retired"
	}
	r.countLifecycle(func(c *lifecycleCounts) { c.closedAnswersRetired++ })
	if it, ok := r.answersRetiredItem(lens, g); !ok || it.Blocks {
		return fmt.Sprintf("%s's list holds no unblocking answers-retired item for %s (found %v, blocks %v)", lens, g, ok, it.Blocks)
	}
	succ, why := r.lifecycleMint(lens, "document", claim, "", "low", "--supersedes", g)
	if succ == "" {
		return "the mint superseding closed gap " + g + " was refused: " + why
	}
	if _, ok := r.answersRetiredItem(lens, g); ok {
		return "the answers-retired item for " + g + " stands after " + lens + " minted " + succ + " superseding it"
	}
	r.countLifecycle(func(c *lifecycleCounts) { c.retiredItemsCleared++ })
	if env, ok := r.lifecycleClose(lens, succ); !ok {
		return "the close of " + succ + " was refused: " + env.Error
	}
	return ""
}

// answersRetiredItem is seat's work-list item for closed gap g whose answers are all retired.
func (r *runner) answersRetiredItem(seat, g string) (record.Item, bool) {
	w, err := record.WorkOfSeat(r.run(), "lens", seat)
	if err != nil {
		return record.Item{}, false
	}
	for _, it := range w.Sitting.Open {
		if it.Kind == record.ItemNews && itemAbout(it, g, "") {
			return it, true
		}
	}
	return record.Item{}, false
}

// lifecycleComputationRetired: the lens judges blue's computation unsound and mints its gap under
// `computation`; blue retires the computation answering the gap, then computes the kept sentence again
// as its answer, and the lens's close is admitted.
func (r *runner) lifecycleComputationRetired(blue, comp, claim string) string {
	k := r.lifecycleCompute(blue, claim, "fuzz-lifecycle-retired.js", "", "")
	if k == "" || !r.lifecycleJudge(comp, k, "unsound") {
		return "blue's computation of the second claim, or its unsound judgement, was refused"
	}
	g, why := r.lifecycleMint(comp, "computation", "", k, "low")
	if g == "" {
		return "the computation gap about " + k + " was refused: " + why
	}
	if !r.lifecycleRetire(blue, k, g) {
		return "blue's retire of " + k + " answering " + g + " was refused"
	}
	if r.lifecycleCompute(blue, claim, "fuzz-lifecycle-recomputed.js", g, "") == "" {
		return "blue's computation again of the kept sentence, answering " + g + ", was refused"
	}
	if env, ok := r.lifecycleClose(comp, g); !ok {
		return "the close of " + g + " on blue's new computation was refused: " + env.Error
	}
	return ""
}

// lifecycleUnsoundGap: the lens judges blue's computation unsound and mints its gap under
// `computation`; another lens judges it sound, and the close is refused by the demand; the lens judges
// it sound itself, and the close is admitted.
func (r *runner) lifecycleUnsoundGap(blue, comp, claim string) string {
	k := r.lifecycleCompute(blue, claim, "fuzz-lifecycle-unsound.js", "", "")
	if k == "" || !r.lifecycleJudge(comp, k, "unsound") {
		return "blue's computation of the third claim, or its unsound judgement, was refused"
	}
	g, why := r.lifecycleMint(comp, "computation", "", k, "low")
	if g == "" {
		return "the computation gap about " + k + " was refused: " + why
	}
	if !r.lifecycleJudge(otherLens, k, "sound") {
		return otherLens + "'s sound judgement of " + k + " was refused"
	}
	if env, ok := r.lifecycleClose(comp, g); ok || !demandRefusal(env, g, "computation") {
		return "the close of " + g + " after another lens's sound judgement was not refused by its demand: " + env.Error
	}
	r.countLifecycle(func(c *lifecycleCounts) { c.unsoundDemandRefused++ })
	if !r.openSitting(comp) || !r.lifecycleJudge(comp, k, "sound") {
		return comp + "'s own sound judgement of " + k + " was refused"
	}
	if env, ok := r.lifecycleClose(comp, g); !ok {
		return "the close of " + g + " after " + comp + "'s own sound judgement was refused: " + env.Error
	}
	r.countLifecycle(func(c *lifecycleCounts) { c.unsoundDemandRejudged++ })
	return ""
}

// lifecycleUnreadGap: the lens reads blue's citation unreachable and mints its gap about it under
// `source`; another lens's backing reading meets nothing, and the close is refused by the demand; the
// lens's own backing reading meets it, and the close is admitted.
func (r *runner) lifecycleUnreadGap(blue, lens, claim string) string {
	cite := r.lifecycleCite(blue, claim, "/lifecycle-4", "")
	if cite == "" {
		return "blue's citation of the fourth claim was refused"
	}
	if _, ok := r.lifecycleRead(lens, cite, "unreachable"); !ok {
		return lens + "'s unreachable reading of " + cite + " was refused"
	}
	g, why := r.lifecycleMint(lens, "source", "", cite, "low")
	if g == "" {
		return "the source gap about unread " + cite + " was refused: " + why
	}
	if _, ok := r.lifecycleRead(otherLens, cite, "supports"); !ok {
		return otherLens + "'s backing reading of " + cite + " was refused"
	}
	if env, ok := r.lifecycleClose(lens, g); ok || !demandRefusal(env, g, "source") {
		return "the close of " + g + " after another lens's backing reading was not refused by its demand: " + env.Error
	}
	r.countLifecycle(func(c *lifecycleCounts) {
		c.unreadDemandRefused++
		r.sourceClosesRefused++
	})
	if !r.openSitting(lens) {
		return "the register of " + lens + " was refused"
	}
	if _, ok := r.lifecycleRead(lens, cite, "supports"); !ok {
		return lens + "'s own backing reading of " + cite + " was refused"
	}
	if env, ok := r.lifecycleClose(lens, g); !ok {
		return "the close of " + g + " after " + lens + "'s own backing reading was refused: " + env.Error
	}
	r.countLifecycle(func(c *lifecycleCounts) { c.unreadDemandReread++ })
	return ""
}

// lifecycleReadingGap: three citations, each read refuting by the lens, which mints its `source` gap
// about each on the reading. The first: blue retires the citation, whose reaching retire meets the
// demand through the gap's about arm, and the close is admitted. The second: blue edits the sentence,
// the close is refused by the demand, the lens re-reads it backing, and the close is admitted. The
// third: the lens re-reads it backing with no edit, and the close is admitted.
func (r *runner) lifecycleReadingGap(blue, lens, retired, edited, reread string) string {
	gaps := map[string]string{}
	cites := map[string]string{}
	for i, claim := range []string{retired, edited, reread} {
		cite := r.lifecycleCite(blue, claim, fmt.Sprintf("/lifecycle-reading-%d", i), "")
		if cite == "" {
			return "blue's citation of " + claim + " was refused"
		}
		g, ok := r.lifecycleRead(lens, cite, "refutes")
		if !ok || g == "" {
			return lens + "'s refuting reading of " + cite + " was refused or minted no gap"
		}
		gaps[claim], cites[claim] = g, cite
	}
	if !r.lifecycleRetire(blue, cites[retired], "") {
		return "blue's retire of " + cites[retired] + " was refused"
	}
	if env, ok := r.lifecycleClose(lens, gaps[retired]); !ok {
		return "the close of reading gap " + gaps[retired] + " after blue retired its citation was refused: " + env.Error
	}
	r.countLifecycle(func(c *lifecycleCounts) { c.readingDemandMetByRetire++ })
	if _, ok := r.do("edit", blue).set("--quote", edited).set("--new", strings.Replace(edited, "holds", "stands", 1)).
		set("--reason", "fuzz: the sentence now says what the source says").admit(); !ok {
		return "blue's edit of the sentence " + cites[edited] + " backs was refused"
	}
	g := gaps[edited]
	if env, ok := r.lifecycleClose(lens, g); ok || !demandRefusal(env, g, "source") {
		return "the close of reading gap " + g + " after blue's edit was not refused by its demand: " + env.Error
	}
	r.countLifecycle(func(*lifecycleCounts) { r.sourceClosesRefused++ })
	if !r.openSitting(lens) {
		return "the register of " + lens + " was refused"
	}
	for _, claim := range []string{edited, reread} {
		if _, ok := r.lifecycleRead(lens, cites[claim], "supports"); !ok {
			return lens + "'s backing re-read of " + cites[claim] + " was refused"
		}
		if env, ok := r.lifecycleClose(lens, gaps[claim]); !ok {
			return "the close of reading gap " + gaps[claim] + " after " + lens + "'s backing re-read was refused: " + env.Error
		}
		r.countLifecycle(func(c *lifecycleCounts) { c.readingDemandMetByReread++ })
	}
	return ""
}

// lifecycleRetiredAbout (L1, ruling Retired mint): a `source` gap about an unread citation and a
// `computation` gap about an unsound computation; blue retires each anchor naming no gap; blue's list
// holds no item for either, and each lens's close is admitted. A mint about either retired anchor, and a
// check of it, are refused naming the retire.
func (r *runner) lifecycleRetiredAbout(blue, lens, comp, cited, computed string) string {
	cite := r.lifecycleCite(blue, cited, "/lifecycle-8", "")
	k := r.lifecycleCompute(blue, computed, "fuzz-lifecycle-about.js", "", "")
	if cite == "" || k == "" {
		return "blue's citation of the eighth claim or computation of the ninth was refused"
	}
	if _, ok := r.lifecycleRead(lens, cite, "unreachable"); !ok || !r.lifecycleJudge(comp, k, "unsound") {
		return "the failed checks of " + cite + " and " + k + " were refused"
	}
	for _, a := range []struct{ seat, kind, anchor, check string }{
		{lens, "source", cite, "verify"},
		{comp, "computation", k, "reproduce"},
	} {
		g, why := r.lifecycleMint(a.seat, a.kind, "", a.anchor, "low")
		if g == "" {
			return "the " + a.kind + " gap about " + a.anchor + " was refused: " + why
		}
		if !r.lifecycleRetire(blue, a.anchor, "") {
			return "blue's retire of " + a.anchor + ", naming no gap, was refused"
		}
		if w, err := record.WorkOfSeat(r.run(), "blue", blue); err != nil || slices.ContainsFunc(w.Sitting.Open, func(it record.Item) bool { return itemAbout(it, g, a.anchor) }) {
			return "blue's list holds an item for " + g + " once its anchor " + a.anchor + " was retired"
		}
		if env, ok := r.lifecycleClose(a.seat, g); !ok {
			return "the close of " + g + " after its anchor's retire was refused: " + env.Error
		}
		r.countLifecycle(func(c *lifecycleCounts) { c.retiredAboutCloses++ })
		if !r.openSitting(a.seat) {
			return "the register of " + a.seat + " was refused"
		}
		if _, why := r.lifecycleMint(a.seat, a.kind, "", a.anchor, "low"); why == "" || !strings.Contains(why, "retire") {
			return "a mint about retired " + a.anchor + " was not refused naming the retire: " + why
		}
		r.countLifecycle(func(c *lifecycleCounts) { c.retiredAboutMintsRefused++ })
		as := map[string]string{"verify": "supports", "reproduce": "sound"}[a.check]
		args := withJSON([]string{a.check, "--seat-id", a.seat, "--anchor", a.anchor, "--as", as, "--reason", "fuzz: a check of a retired anchor"})
		if a.check == "verify" {
			args = append(args, "--confidence", "high")
		}
		if out, err := r.exec(args...); err == nil || !strings.Contains(out, "retire") {
			return "a " + a.check + " of retired " + a.anchor + " was not refused naming the retire: " + firstLine(out)
		}
	}
	return ""
}

// lifecycleReachingRetire: a `source` gap about an unread citation, graded material so a plan engages
// on it, whose citation blue retires naming no gap; the gap is left open, and the next plan must engage
// its lens on it (tallyLensTurns), whose close is then admitted (closeLifecycleLeftOpen).
func (r *runner) lifecycleReachingRetire(blue, lens, claim string) string {
	cite := r.lifecycleCite(blue, claim, "/lifecycle-10", "")
	if cite == "" {
		return "blue's citation of the tenth claim was refused"
	}
	if !r.openSitting(lens) {
		return "the register of " + lens + " was refused"
	}
	if _, ok := r.lifecycleRead(lens, cite, "unreachable"); !ok {
		return lens + "'s unreachable reading of " + cite + " was refused"
	}
	g, why := r.lifecycleMint(lens, "source", "", cite, "medium")
	if g == "" {
		return "the source gap about unread " + cite + " was refused: " + why
	}
	if !r.lifecycleRetire(blue, cite, "") {
		return "blue's retire of " + cite + ", naming no gap, was refused"
	}
	r.mu.Lock()
	r.lifecycleLeftOpen = g
	r.mu.Unlock()
	return ""
}

// closeLifecycleLeftOpen is the lens's close of the gap lifecycleReachingRetire left open, at a sitting
// whose plan engages it there: the reaching retire met its demand, and the close is admitted (L1). It
// returns the gap it closed, which the sitting's drawn acts then leave alone, or "".
func (r *runner) closeLifecycleLeftOpen(seatID string) string {
	r.mu.Lock()
	g := r.lifecycleLeftOpen
	r.mu.Unlock()
	if g == "" || seatID != citationOwner || !slices.Contains(r.engaged(seatID), g) || !slices.Contains(r.openGaps(), g) {
		return ""
	}
	if env, ok := r.lifecycleClose(seatID, g); !ok {
		r.noteActErr("the close of " + g + ", whose anchor blue retired naming no gap, was refused: " + env.Error)
		return ""
	}
	r.countLifecycle(func(c *lifecycleCounts) { c.retiredAboutCloses++ })
	return g
}

// lifecycleCitedRetire (L6, ruling Same-act L6): blue's computation applies the method of blue's
// citation; the citation's retire is refused naming the computation. Blue's edit cuts the sentence onto
// the next, carrying both anchors; a claim retire taking the citation alone is refused naming the
// computation, and one taking both is admitted.
func (r *runner) lifecycleCitedRetire(blue, claim, next string) string {
	cite := r.lifecycleCite(blue, claim, "/lifecycle-11", "")
	if cite == "" {
		return "blue's citation of the eleventh claim was refused"
	}
	k := r.lifecycleCompute(blue, claim, "fuzz-lifecycle-method.js", "", cite)
	if k == "" {
		return "blue's computation applying " + cite + " was refused"
	}
	refusedNaming := func(args ...string) bool {
		out, err := r.exec(withJSON(args)...)
		env, ok := refusalEnvelopeIn(out)
		return err != nil && ok && strings.Contains(env.Error, k)
	}
	if !refusedNaming("retire", "--seat-id", blue, "--anchor", cite, "--reason", "fuzz: the citation leaves while a computation applies it") {
		return "the retire of " + cite + " while " + k + " applies it was not refused naming it"
	}
	r.countLifecycle(func(c *lifecycleCounts) { c.citedRetiresRefused++ })
	md, err := record.RenderRun(r.run())
	if err != nil {
		return "the report could not be rendered: " + err.Error()
	}
	run := abuttingRun(md, claim)
	if !strings.Contains(run, cite) || !strings.Contains(run, k) {
		return "the citation and the computation do not stand at the eleventh claim's end: " + run
	}
	if _, ok := r.do("edit", blue).set("--quote", claim+run+". "+next).set("--new", next+run).
		set("--reason", "fuzz: the claim cannot stand").admit(); !ok {
		return "blue's cut of the eleventh claim carrying its anchors was refused"
	}
	if !refusedNaming("retire", "--seat-id", blue, "--quote", claim+".", "--anchor", cite, "--reason", "fuzz: the claim leaves") {
		return "a claim retire taking " + cite + " while " + k + " applies it was not refused naming it"
	}
	r.countLifecycle(func(c *lifecycleCounts) { c.citedRetiresRefused++ })
	if _, ok := r.do("retire", blue).set("--quote", claim+".").set("--anchor", cite+","+k).set("--reason", "fuzz: the claim leaves with its evidence").admit(); !ok {
		return "a claim retire taking " + cite + " with " + k + " was refused"
	}
	return ""
}

// screenUnsoundComputationName is the bundle of the computation limitSeed's run ends on, judged unsound
// (screenAtLimit).
const screenUnsoundComputationName = "fuzz-screen-unsound.js"

// screenAtLimit is limitSeed's half of the assembly screen arm, driven at the chair's sitting whose plan
// reaches the epoch limit, which dispatches nobody: blue-respond, in a sitting of its own, computes the
// fuzz sentence, and the computation owner, in its own, judges it unsound. No seat sits after, so the run
// ends with it live, and the screen at the run's end must FAIL naming it (screenUnsound). The limit plan
// is the one point the forced run is sure to reach; blue-respond is not sure to be dispatched before it,
// since which seats the epochs ready turns on the run's draws and the seats' interleaving. Draws nothing.
func (r *runner) screenAtLimit(plan map[string]any) {
	r.mu.Lock()
	due := r.forceScreenUnsound && r.screenUnsoundComputation == ""
	r.mu.Unlock()
	if limit, _ := plan["epoch_limit_reached"].(bool); !limit || !due {
		return
	}
	const blue = "blue-respond"
	k := ""
	if r.openSitting(blue) {
		k = r.lifecycleCompute(blue, "§ fuzz sentence", screenUnsoundComputationName, "", "")
	}
	r.mu.Lock()
	r.screenUnsoundComputation = cmp.Or(k, "refused")
	r.mu.Unlock()
	if k != "" && r.openSitting(computationOwner) {
		r.lifecycleJudge(computationOwner, k, "unsound")
	}
	// AND A READING DEBT STANDS AT THE RUN'S END: blue cites a sentence of the report, and the citation
	// owner, in a sitting of its own, reads the citation `refutes` and mints nothing — the run ends
	// UNVERIFIED on the debt, its reason naming it (tallyDebtsAtEnd), and the assembly screen fails
	// on it beside the unsound computation.
	// THE CITATION IS RESERVED BEFORE IT IS READ (reservedDebtSubject): a sitting of the owner's or
	// blue's still running from the last plan would otherwise mint on the debt, re-read the citation
	// as backing, or retire it, and the run would end on no debt.
	// THE SENTENCE CITED HOLDS NO OPEN claim_contradicts_source GAP (limitDebtSentence): a gap of the
	// class on the sentence, quoting it or about a citation it holds, discharges the owner's debt the
	// moment the reading lands (F9, `open now`), and the run would end on none — the lenses' drawn
	// sittings mint that gap quoting lane 1's sentence (oneURLChain, a found reading's mint).
	quote, sentence := r.limitDebtSentence()
	if quote == "" {
		r.mu.Lock()
		r.limitDebtCite = "refused"
		r.mu.Unlock()
		r.noteActErr("the limit plan found no sentence to cite for the reading debt: every candidate holds an open claim_contradicts_source gap or is not held once")
		return
	}
	cite := r.lifecycleCite(blue, quote, "/limit-debt", "")
	r.mu.Lock()
	r.limitDebtCite, r.limitDebtSentenceCited = cmp.Or(cite, "refused"), sentence
	r.mu.Unlock()
	if cite == "" || !r.openSitting(citationOwner) {
		return
	}
	at := r.readCited(citationOwner, cite, "refutes")
	r.mu.Lock()
	r.limitDebtRead = at
	r.mu.Unlock()
	if at == 0 {
		return
	}
	// THE DEBT IS READ BACK AT ONCE: a reading that leaves none names, from the record, what discharged
	// it, so the run fails here with the cause rather than at its end with the absence.
	if _, stands := r.debtOf(citationOwner, cite); !stands {
		r.noteActErr(fmt.Sprintf("the owner's refuting reading of %s at the limit plan (event %d) left no reading debt: %s", cite, at, r.limitDebtDischarge()))
	}
}

// limitDebtCandidates are the sentences the limit plan's citation may sit in, with the span blue
// cites, in the order tried: lane 1's anchor sentence, lane 3's, then the lanes' closing sentences.
var limitDebtCandidates = []struct{ quote, sentence string }{
	{"§ fuzz sentence", "A § fuzz sentence to anchor findings."},
	{strings.TrimSuffix(foundSourceSentences[0], "."), foundSourceSentences[0]},
	{strings.TrimSuffix(laneCloseSentences[0], "."), laneCloseSentences[0]},
	{strings.TrimSuffix(laneCloseSentences[1], "."), laneCloseSentences[1]},
}

// limitDebtSentence is the first of limitDebtCandidates the report holds once with no open
// claim_contradicts_source gap on it — quoting it, or about a citation it holds (sameClassAndSubject,
// both arms) — and the span blue cites in it; "" where none qualifies.
func (r *runner) limitDebtSentence() (quote, sentence string) {
	cur, err := record.RenderRun(r.run())
	if err != nil {
		return "", ""
	}
	for _, c := range limitDebtCandidates {
		if _, once := sentenceMarked(cur, c.sentence); !once {
			continue
		}
		if open, _ := r.sameClassAndSubject(classContradicts, quoteArgs(c.sentence)); len(open) > 0 {
			continue
		}
		return c.quote, c.sentence
	}
	return "", ""
}

// limitDebtDischarge is what the record holds against the limit plan's reading debt (limitDebtCite,
// read at limitDebtRead): every act after the reading that names the citation — the owner's later
// reading of it, blue's retire of it, blue's edit reopening it, a claim_contradicts_source mint
// about it or quoting the sentence it sits in — and every open gap of that class on its sentence.
// Named so a run ending on no debt says which discharge the record holds, never a guess.
func (r *runner) limitDebtDischarge() string {
	r.mu.Lock()
	cite, at, sentence := r.limitDebtCite, r.limitDebtRead, r.limitDebtSentenceCited
	r.mu.Unlock()
	b, ok := r.board()
	if !ok {
		return "the board did not read"
	}
	var found []string
	for _, e := range b.Events {
		id := b.At.Of(e).ID
		if id <= at {
			continue
		}
		switch body := e.GetBody().(type) {
		case *recordpb.Event_Verify:
			if body.Verify.GetAnchor() == cite && e.GetSeatId() == citationOwner {
				found = append(found, fmt.Sprintf("the owner read %s again at event %d (%s)", cite, id, recordpb.Word(body.Verify.GetOutcome())))
			}
		case *recordpb.Event_Retire:
			if slices.Contains(body.Retire.GetAnchors(), cite) {
				found = append(found, fmt.Sprintf("%s retired %s at event %d", e.GetSeatId(), cite, id))
			}
		case *recordpb.Event_BlueEdit:
			if slices.Contains(body.BlueEdit.GetReopened(), cite) {
				found = append(found, fmt.Sprintf("%s's edit at event %d reopened %s", e.GetSeatId(), id, cite))
			}
		case *recordpb.Event_Mint:
			m := body.Mint
			if recordpb.Word(m.GetClass()) != classContradicts {
				continue
			}
			switch {
			case m.GetAboutKind() == recordpb.AboutKind_ABOUT_KIND_ANCHOR && m.GetAboutRef() == cite:
				found = append(found, fmt.Sprintf("%s minted %s about %s at event %d", e.GetSeatId(), m.GetGapId(), cite, id))
			case strings.TrimSpace(m.GetLocation()) != "" && sameDebtSubject(sentence, m.GetLocation()):
				found = append(found, fmt.Sprintf("%s minted %s quoting the citation's sentence at event %d", e.GetSeatId(), m.GetGapId(), id))
			}
		}
	}
	if open, _ := r.sameClassAndSubject(classContradicts, quoteArgs(sentence)); len(open) > 0 {
		for _, g := range open {
			found = append(found, fmt.Sprintf("open %s gap %s (minted by %s) is on the sentence: about %s %s, at %q", classContradicts, g.ID, g.MintedBy, g.AboutKind, g.AboutRef, g.Location))
		}
	}
	if len(found) == 0 {
		return fmt.Sprintf("no act after event %d names %s and no open %s gap is on its sentence %q", at, cite, classContradicts, sentence)
	}
	return strings.Join(found, "; ")
}

// screenUnsound holds a run that ends with a live computation a standing judgement calls unsound to the
// assembly screen's FAIL naming each such computation, and counts it; msg is the run's failure.
func (r *runner) screenUnsound() (screened int, msg string) {
	var unsound []string
	failures, err := record.AssemblyFailures(r.run())
	if err != nil {
		return 0, "assembly screen: the record could not be read: " + err.Error()
	}
	for _, f := range failures {
		if f.Kind == string(record.CheckUnsound) {
			unsound = append(unsound, f.Subject)
		}
	}
	if len(unsound) == 0 {
		return 0, ""
	}
	got := capture.AssemblyScreen(r.run())
	if got.Verdict != "FAIL" {
		return 0, fmt.Sprintf("assembly screen: the run ends with unsound computation(s) %v live, and the screen reads %s: %s", unsound, got.Verdict, got.Detail)
	}
	for _, k := range unsound {
		if !strings.Contains(got.Detail, "computation "+k) {
			return 0, "assembly screen: the FAIL does not name unsound computation " + k + ": " + got.Detail
		}
	}
	return 1, ""
}

// stagedLifecycleRunner is stagedSynthesizedRunner with blue-respond sitting.
func stagedLifecycleRunner(t *testing.T) *runner {
	t.Helper()
	r := stagedSynthesizedRunner(t)
	r.register("blue", "blue-respond")
	return r
}

// stagedSynthesizedRunner is a staged run whose report holds the lanes' parts, framed by the synthesis.
func stagedSynthesizedRunner(t *testing.T) *runner {
	t.Helper()
	r := stagedRunner(t)
	if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2") + "\n" + lanePartText("blue-lane-3")); err != nil {
		t.Fatalf("record the lanes' parts: %v", err)
	}
	r.synthesize("blue-synthesize")
	if r.actErr != "" {
		t.Fatalf("the synthesis: %s", r.actErr)
	}
	return r
}

// THE FORCED LIFECYCLE RUN DRIVES EVERY ARM at blue-respond's first sitting: each count it gates is
// above zero once it has run, but the reaching retire's turn, which the next plan gives. With the run not
// forced, none of it is driven and the run owes nothing.
func TestTheLifecycleSeedDrivesEveryArm(t *testing.T) {
	t.Run("forced", func(t *testing.T) {
		r := stagedLifecycleRunner(t)
		r.forceLifecycle = true
		r.mu.Lock()
		if msg := r.lifecycleForcedErr(); !strings.Contains(msg, "never sat") {
			t.Fatalf("a forced run nothing has driven reads %q, want the sitting it never reached", msg)
		}
		r.mu.Unlock()
		r.envelopeFor("blue-respond", "")
		r.mu.Lock()
		stage, counts, left, msg := r.lifecycleStage, r.lifecycle, r.lifecycleLeftOpen, r.lifecycleForcedErr()
		r.mu.Unlock()
		if stage != lifecycleDriven || r.actErr != "" {
			t.Fatalf("the forced arm reached %q (%s)", stage, r.actErr)
		}
		// EVERY GATE BUT THE TWO THE ARM CANNOT REACH IN ITS SITTING: the reaching retire's turn, which
		// the next plan gives (TestAReachingRetireGivesTheLensItsTurn), and the screen at a run's end
		// (TestTheLimitSeedEndsOnAScreenedUnsoundComputation).
		later := lifecycleTally{ReachingRetireTurns: 1, ScreenedUnsound: 1}.gates()
		for i, g := range counts.tally().gates() {
			if g.n == 0 && later[i].n == 0 {
				t.Errorf("the forced arm drove ZERO %s", g.what)
			}
		}
		// TWO COMPUTATIONS RETIRE: the second claim's, answering its gap before blue computes it again, and
		// the ninth's, naming no gap (L1). Either arm gone leaves one.
		if counts.computationRetires != 2 {
			t.Errorf("the forced arm retired %d computation(s), want 2", counts.computationRetires)
		}
		if !strings.Contains(msg, "no plan engaged") || !slices.Contains(r.openGaps(), left) {
			t.Errorf("the forced arm left gap %q open (open: %v) and its run reads %q, want the reaching retire's turn owed", left, r.openGaps(), msg)
		}
	})
	t.Run("not forced", func(t *testing.T) {
		r := stagedLifecycleRunner(t)
		r.forcedLifecycle("blue-respond")
		r.mu.Lock()
		defer r.mu.Unlock()
		if r.lifecycleStage != "" || r.lifecycle != (lifecycleCounts{}) || r.lifecycleForcedErr() != "" {
			t.Errorf("a run not forced drove the arm (stage %q, counts %+v) or owes it", r.lifecycleStage, r.lifecycle)
		}
	})
}

// A REACHING RETIRE GIVES THE LENS ITS TURN: the plan after the forced arm engages the citation owner on
// the gap its reaching retire left open, the lens-turn oracle counts the turn, and the lens's close of it
// is admitted at that sitting. A plan engaging only blue there fails the run.
func TestAReachingRetireGivesTheLensItsTurn(t *testing.T) {
	plan := func(seat, gap string) map[string]any {
		return map[string]any{"parties": []any{map[string]any{"seat_id": seat, "gap_ids": []any{gap}}}}
	}
	drive := func(t *testing.T) (*runner, string) {
		r := stagedLifecycleRunner(t)
		r.forceLifecycle = true
		r.forcedLifecycle("blue-respond")
		if r.lifecycleStage != lifecycleDriven || r.lifecycleLeftOpen == "" {
			t.Fatalf("the forced arm reached %q (%s)", r.lifecycleStage, r.actErr)
		}
		return r, r.lifecycleLeftOpen
	}
	t.Run("the lens engaged", func(t *testing.T) {
		r, g := drive(t)
		closes := r.lifecycle.retiredAboutCloses
		r.planThisSitting = plan(citationOwner, g)
		r.tallyLensTurns(r.planThisSitting)
		if r.actErr != "" || r.lifecycle.reachingRetireTurns != 1 {
			t.Fatalf("a plan engaging %s on %s counted %d reaching-retire turn(s) (%s), want 1", citationOwner, g, r.lifecycle.reachingRetireTurns, r.actErr)
		}
		r.mu.Lock()
		msg := r.lifecycleForcedErr()
		r.mu.Unlock()
		if msg != "" {
			t.Errorf("with the turn counted the forced run still reads %q", msg)
		}
		if got := r.closeLifecycleLeftOpen(citationOwner); got != g || r.actErr != "" || r.lifecycle.retiredAboutCloses != closes+1 || slices.Contains(r.openGaps(), g) {
			t.Errorf("the lens's close of %s at the sitting that engages it: closed %q (%s), %d close(s) counted, want it closed and counted", g, got, r.actErr, r.lifecycle.retiredAboutCloses-closes)
		}
	})
	t.Run("blue alone engaged", func(t *testing.T) {
		r, g := drive(t)
		r.tallyLensTurns(plan("blue-respond", g))
		if r.lifecycle.reachingRetireTurns != 0 || !strings.Contains(r.actErr, "not its minting lens") {
			t.Errorf("a plan engaging blue alone on %s counted %d turn(s) and failed the run with %q, want it failed naming the lens", g, r.lifecycle.reachingRetireTurns, r.actErr)
		}
	})
}

// THE LIMIT SEED ENDS ON A SCREENED UNSOUND COMPUTATION: at the plan that reaches the epoch limit, blue
// computes the fuzz sentence and the computation owner judges it unsound, and the assembly screen at the
// run's end FAILS naming it. Blue-respond has never sat: the drive owes nothing to a dispatch of blue
// before the limit, which the forced run is not sure to reach. A plan short of the limit drives nothing,
// and a run with no live unsound computation is screened for none.
func TestTheLimitSeedEndsOnAScreenedUnsoundComputation(t *testing.T) {
	stage := func(t *testing.T) *runner {
		r := stagedSynthesizedRunner(t)
		r.forceScreenUnsound = true
		return r
	}
	screen := func(t *testing.T, r *runner) (int, string) {
		if _, err := reportdoc.Assemble(r.run()); err != nil {
			t.Fatalf("assemble: %v", err)
		}
		return r.screenUnsound()
	}
	t.Run("at the limit", func(t *testing.T) {
		r := stage(t)
		r.screenAtLimit(map[string]any{"epoch_limit_reached": true})
		if !strings.HasPrefix(r.screenUnsoundComputation, anchorPrefix["computation"]) || r.actErr != "" {
			t.Fatalf("blue's computation of the fuzz sentence at the limit: %q (%s)", r.screenUnsoundComputation, r.actErr)
		}
		if n, msg := screen(t, r); n != 1 || msg != "" {
			t.Errorf("the screen of a run ending on unsound %s counted %d (%s), want 1", r.screenUnsoundComputation, n, msg)
		}
	})
	t.Run("short of the limit", func(t *testing.T) {
		r := stage(t)
		r.screenAtLimit(map[string]any{"epoch_limit_reached": false})
		if r.screenUnsoundComputation != "" {
			t.Errorf("a plan short of the limit drove the arm: %q", r.screenUnsoundComputation)
		}
		if n, msg := screen(t, r); n != 0 || msg != "" {
			t.Errorf("the screen of a run with no unsound computation counted %d (%s), want 0", n, msg)
		}
	})
	t.Run("not forced", func(t *testing.T) {
		r := stagedSynthesizedRunner(t)
		r.screenAtLimit(map[string]any{"epoch_limit_reached": true})
		if r.screenUnsoundComputation != "" {
			t.Errorf("a run not forced drove the arm: %q", r.screenUnsoundComputation)
		}
	})
}

// itemAbout reports whether a work-list item is about gap g, or about anchor a where a is not "" —
// read off the item's subject, the field its sentence is built from.
func itemAbout(it record.Item, g, a string) bool {
	if it.Subject == nil {
		return false
	}
	switch it.Subject.Kind {
	case record.SubjectGap:
		return it.Subject.ID == g
	case record.SubjectAnchor:
		return a != "" && it.Subject.ID == a
	}
	return false
}

// THE LIMIT PLAN CITES A SENTENCE FREE OF THE CLASS'S GAP: an open claim_contradicts_source gap on
// lane 1's sentence — a lens's quote mint on its own found reading — would discharge the owner's
// debt the moment it is read (F9), so the limit debt sits on the next candidate, and stands; and
// a run ending on no debt names, from the record, what discharged it.
func TestTheLimitDebtSitsOnASentenceFreeOfTheClassGap(t *testing.T) {
	r := stagedSynthesizedRunner(t)
	r.forceScreenUnsound = true
	const other = "red-lens-dark-side"
	r.register("lens", other)
	if r.readFound(other, foundSourceURL(other), "A § fuzz sentence to anchor findings.", "refutes") == 0 {
		t.Fatalf("the other lens's reading: %s", r.actErr)
	}
	g := r.mintOnReading(other, classContradicts, quoteArgs("A § fuzz sentence to anchor findings."), 0, "quote").gap
	if g == "" || r.actErr != "" {
		t.Fatalf("the other lens's mint quoting lane 1's sentence: %q (%s)", g, r.actErr)
	}
	r.screenAtLimit(map[string]any{"epoch_limit_reached": true})
	if r.actErr != "" || !strings.HasPrefix(r.limitDebtCite, anchorPrefix["citation"]) || r.limitDebtRead == 0 {
		t.Fatalf("the limit plan's cite %q read at %d (%s)", r.limitDebtCite, r.limitDebtRead, r.actErr)
	}
	if r.limitDebtSentenceCited != foundSourceSentences[0] {
		t.Errorf("the limit debt sits on %q, want lane 3's sentence with lane 1's holding %s", r.limitDebtSentenceCited, g)
	}
	if _, stands := r.debtOf(citationOwner, r.limitDebtCite); !stands {
		t.Fatalf("the owner's debt on %s does not stand: %s", r.limitDebtCite, r.limitDebtDischarge())
	}
	if got := r.limitDebtDischarge(); !strings.HasPrefix(got, "no act after event") {
		t.Errorf("the record holds a discharge while the debt stands: %s", got)
	}
	// THE OWNER'S LATER BACKING READ, named by the oracle as the discharge.
	r.register("lens", citationOwner)
	if r.readCited(citationOwner, r.limitDebtCite, "supports") == 0 {
		t.Fatalf("the owner's re-read: %s", r.actErr)
	}
	if _, stands := r.debtOf(citationOwner, r.limitDebtCite); stands {
		t.Fatal("the owner's backing re-read left the debt")
	}
	if got := r.limitDebtDischarge(); !strings.Contains(got, "the owner read "+r.limitDebtCite+" again") || !strings.Contains(got, "(supports)") {
		t.Errorf("the oracle names %q, want the owner's re-read", got)
	}
}
