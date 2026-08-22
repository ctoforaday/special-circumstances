package record

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// marshalEvent serializes exactly as JSON.stringify does.
//
// TRAP: Go's encoding/json escapes <, > and & to <>& by default;
// JSON.stringify does not. Seat prose routinely contains angle brackets (quoted
// markdown, "<seat scratchpad>", diff fragments), so the default would diverge
// from the oracle on ordinary input and only at the byte level — the hardest
// class of port bug to see. SetEscapeHTML(false) is mandatory here.
func marshalEvent(ev Event) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(ev); err != nil {
		return nil, err
	}
	return bytes.TrimRight(b.Bytes(), "\n"), nil
}

// MarshalEvent is the exported shard-line encoder, symmetric to ReadShard, for
// tests in other packages (internal/view) that build a run's shards on disk.
func MarshalEvent(ev Event) ([]byte, error) { return marshalEvent(ev) }

// marshalCompact is marshalEvent's rule for any other value written to disk.
func marshalCompact(v any) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(b.Bytes(), "\n"), nil
}

// MarshalCompact is the exported form for the view package, which needs the
// exact byte-identical encoding when it computes the telemetry JSONL on read.
func MarshalCompact(v any) ([]byte, error) { return marshalCompact(v) }

// ReadShard parses a shard, silently dropping unparseable lines — a torn fragment
// from a crashed append stays visible in the file and inert in the replay.
func ReadShard(path string) ([]Event, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Event
	for _, line := range strings.Split(string(b), "\n") {
		if line == "" {
			continue
		}
		var ev Event
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}
		out = append(out, ev)
	}
	return out, nil
}

type shardInfo struct {
	nonce  string
	file   string
	events []Event
}

// Merged is the deterministic replay: winner selection per seat, global ordering,
// key-level dedup, and the anomalies that must never be silently normalized.
type Merged struct {
	Events    []Event
	Anomalies []string
	// Discarded names every seat where losing shards carried events the winner does not.
	//
	// Multi-nonce is NORMAL: a register rotates the nonce so a crash re-dispatch is a fresh
	// shard, measured at 8/50 in run 5. On a re-dispatch the retry rewrites the SAME keys, so
	// the loser contributes nothing the winner lacks and this stays empty.
	//
	// It is NOT empty when one seat id was used for two different sittings, because then the
	// losing shard holds work that happened and no longer exists anywhere downstream. The
	// anomaly string said "multi-nonce seat X: 7 dispatches" for both cases — the healthy one
	// and the lossy one, in the same words, and nothing gated on either (#394). This is the
	// field that tells them apart.
	Discarded []DiscardedShard
}

// DiscardedShard is one seat's lost work: how many event keys existed in a losing shard and
// survive nowhere. A COUNT, not a sentence — a caller decides with it instead of matching prose
// (facts-are-fields).
type DiscardedShard struct {
	SeatID     string
	Dispatches int
	Winner     string
	// Keys are the discarded event keys, in shard order. Bounded in practice by one sitting's
	// output, and worth carrying whole: "which rulings vanished" is the actionable half.
	Keys []string
}

var shardRe = regexp.MustCompile(`^events-(.+)-([0-9a-f]{8})\.jsonl$`)

func MergedEvents(runDir string) (Merged, error) {
	// A RESOLUTION FAILURE IS NOT AN EMPTY RUN. The IsNotExist arm below is the honest zero —
	// a run that exists and has recorded nothing yet — and it is exactly what an unreachable
	// separated record would otherwise be flattened into. RecordsDir refuses instead, so the
	// two stay distinguishable here, where every board projection begins.
	dir, err := RecordsDir(runDir)
	if err != nil {
		return Merged{}, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return Merged{}, nil
		}
		return Merged{}, err
	}
	var names []string
	for _, e := range entries {
		n := e.Name()
		if strings.HasPrefix(n, "events-") && strings.HasSuffix(n, ".jsonl") {
			names = append(names, n)
		}
	}
	sort.Strings(names) // mirrors readdirSync(...).sort()

	// Insertion-ordered grouping: Go map iteration is randomized, and the oracle
	// walks a JS Map in insertion order. Anomaly ORDER is part of the render output.
	var seatOrder []string
	bySeat := map[string][]shardInfo{}
	for _, n := range names {
		m := shardRe.FindStringSubmatch(n)
		if m == nil {
			continue
		}
		seatID, nonce := m[1], m[2]
		full := filepath.Join(dir, n)
		evs, err := ReadShard(full)
		if err != nil {
			return Merged{}, err
		}
		if _, seen := bySeat[seatID]; !seen {
			seatOrder = append(seatOrder, seatID)
		}
		bySeat[seatID] = append(bySeat[seatID], shardInfo{nonce: nonce, file: full, events: evs})
	}

	var anomalies []string
	var discarded []DiscardedShard
	var winners []Event
	for _, seatID := range seatOrder {
		shards := bySeat[seatID]
		if len(shards) == 1 {
			winners = append(winners, shards[0].events...)
			continue
		}
		// Multi-nonce: the winner is the nonce whose shard carries the seat's
		// TERMINAL event (verdict or revision — the last verb of a seat contract);
		// with neither terminal, fall EXPLICITLY to the latest RECORDED event stamp.
		var terminal []shardInfo
		for _, s := range shards {
			for _, e := range s.events {
				if e.Type == "verdict" || e.Type == "revision" {
					terminal = append(terminal, s)
					break
				}
			}
		}
		pool := terminal
		if len(pool) == 0 {
			pool = shards
		}
		// THE RECORD'S OWN CLOCK DECIDES, NOT THE FILESYSTEM'S.
		//
		// This was `s.mtime.After(winner.mtime)` — the shard FILE's modification time — and it
		// was wrong twice over.
		//
		// MEASURED 2026-08-16: TestDiscardedEventsAudit passed on ubuntu-latest and FAILED on
		// windows-latest at the same commit, while main's own run was green. Two shards written
		// back to back land on distinct mtimes under Linux and on IDENTICAL ones under Windows'
		// coarser timestamp resolution, so nothing was ever `After` anything and pool[0] stood.
		// Which sitting's work survived was decided by a clock the record does not own.
		//
		// AND NO BETTER CLOCK WOULD HAVE FIXED IT, because mtime is metadata about the CARRIER
		// rather than a fact on the record. A `git checkout`, an rsync, a container copy or a
		// restore from the recovery mirror all rewrite it, and every one of those would silently
		// change which sitting won — on a filesystem with all the resolution in the world.
		//
		// The record already issues exactly the clock this needs. `nextStamp` takes a lock, reads
		// `.clock`, and returns `prev + 1ns` whenever the wall clock has not advanced, so every
		// event's TS is STRICTLY INCREASING across seats and processes within a run — built that
		// way precisely because "wall clock from many short-lived processes is not monotonic and
		// can go backwards" (see the global ordering below). stampLayout is fixed-width, zero-
		// padded and UTC, so string order is time order.
		//
		// The nonce tie-break remains underneath, and is ARBITRARY AND STABLE rather than a claim
		// about which sitting came later: nonces are random. nextStamp makes a tie unreachable
		// within one run, so it is the answer to a corrupt or hand-written record, not to a race.
		latest := func(s shardInfo) string {
			out := ""
			for _, e := range s.events {
				if e.TS > out {
					out = e.TS
				}
			}
			return out
		}
		winner := pool[0]
		for _, s := range pool[1:] {
			switch a, b := latest(s), latest(winner); {
			case a > b:
				winner = s
			case a == b && s.nonce > winner.nonce:
				winner = s
			}
		}
		tied := false
		for _, s := range pool {
			if s.nonce != winner.nonce && latest(s) == latest(winner) {
				tied = true
				break
			}
		}

		nonces := make([]string, len(shards))
		for i, s := range shards {
			nonces[i] = s.nonce
		}
		by := "latest recorded event"
		if len(terminal) > 0 {
			by = "terminal event"
		}

		// WHAT DID THE LOSERS HOLD THAT THE WINNER DOES NOT? A crash re-dispatch rewrites the
		// same idempotency keys, so the answer is normally nothing and this is silent. A seat id
		// reused for a second SITTING has no overlap at all, and every key here is work that
		// happened and survives nowhere.
		//
		// `register` IS EXCLUDED, and the check is worthless without that: its key is
		// `<seat>:register:<nonce>` (record.go:227), so the loser's register can never match the
		// winner's. Counting it would report one discarded event for every healthy crash
		// re-dispatch — the signal firing on precisely the case it exists to permit.
		kept := make(map[string]bool, len(winner.events))
		for _, e := range winner.events {
			kept[e.Key] = true
		}
		var lost []string
		for _, s := range shards {
			if s.nonce == winner.nonce {
				continue
			}
			for _, e := range s.events {
				if e.Type == "register" {
					continue
				}
				if !kept[e.Key] {
					lost = append(lost, e.Key)
				}
			}
		}

		if tied && len(terminal) == 0 {
			by = "recorded stamps TIED, broken by nonce — no event separates these sittings"
		}
		note := fmt.Sprintf("multi-nonce seat %s: %d dispatches (%s); winner %s by %s",
			seatID, len(shards), strings.Join(nonces, ", "), winner.nonce, by)
		if len(lost) > 0 {
			note += fmt.Sprintf(" — %d event(s) DISCARDED, surviving nowhere: %s", len(lost), strings.Join(lost, ", "))
			discarded = append(discarded, DiscardedShard{
				SeatID: seatID, Dispatches: len(shards), Winner: winner.nonce, Keys: lost,
			})
		}
		anomalies = append(anomalies, note)
		winners = append(winners, winner.events...)
	}

	// Deterministic global order. sort.SliceStable mirrors Array.prototype.sort's
	// stability, which matters where round+seatId+seq collide across shards.
	sort.SliceStable(winners, func(i, j int) bool {
		a, b := winners[i], winners[j]
		if a.Round != b.Round {
			return a.Round < b.Round
		}
		if a.SeatID != b.SeatID {
			return a.SeatID < b.SeatID
		}
		return a.Seq < b.Seq
	})

	seen := map[string]bool{}
	events := make([]Event, 0, len(winners))
	for _, e := range winners {
		if e.Type != "register" && seen[e.Key] {
			anomalies = append(anomalies, fmt.Sprintf("duplicate key dedup'd: %s (nonce %s)", e.Key, e.Nonce))
			continue
		}
		seen[e.Key] = true
		events = append(events, e)
	}
	return Merged{Events: events, Anomalies: anomalies, Discarded: discarded}, nil
}

// Gap is the replayed state of one board gap.
//
// The four grades are `any` rather than string because ABSENCE is renderable:
// the oracle interpolates them straight into a template literal, so a gap minted
// without --complexity renders the literal text "undefined". Collapsing that to "" here
// would be a silent, byte-level divergence in every ledger — exactly the class the
// differential gate exists to catch, and it did catch it.
type Gap struct {
	ID             string
	Round          int
	Open           bool
	ClosedRound    int
	HasClosed      bool
	Mint           *Payload
	Closure        *Payload
	Regrades       []*Payload
	Severity       any
	Likelihood     any
	Impact         any
	ComplexityCost any
	// ClosedByBench distinguishes a JUDICIAL closure from one red made. Red cannot
	// close a bench-closed gap itself without double-counting closure history and
	// corrupting the repair_regression denominator, so the projection has to record WHO
	// closed it, not merely that it is closed.
	ClosedByBench bool
}

// benchClosesGap says which dispositions end a gap's life on the board.
//
// `carried` is the one that does NOT: it defers the question to a later round, and
// treating it as a closure would silently retire gaps the bench deliberately kept alive
// — the opposite of the defect this function exists to fix, and worse.
//
// The rest all end the dispute, by different routes: the defect is gone, the risk is
// accepted as it stands, blue's rebuttal was sustained, or the work moved to the lead's
// infrastructure debt. Each leaves the board with nothing further to adjudicate.
func benchClosesGap(disposition string) bool {
	// ONE VOCABULARY (#342). Every ClosureClass ends a gap; `carried` is the sole
	// disposition that does not, because it defers the question to a later round.
	// Enumerating the closing words here was a THIRD list of the same concept, beside
	// `close`'s classes and the envelope's — and the three disagreed.
	return disposition != "" && disposition != DispositionCarried
}

// missingGap describes a mutation that referenced a gap the replay has never seen.
//
// This used to be a bare `continue`, and that silence is what let the bench's closures
// vanish for an entire run: the events were recorded correctly, the replay dropped them,
// and every projection downstream reported a board that had never existed. Anomalies are
// rendered into the projection and never silently normalized, so the same failure now
// announces itself in the artifact a human reads.
func missingGap(verb string, e Event) string {
	return fmt.Sprintf("%s by %s referenced unknown gap %s (event %s) — the mutation was DROPPED, not applied",
		verb, e.SeatID, e.Payload.Str("gap_id"), e.Key)
}

// payloadVal returns the raw value, or nil when the key is absent — nil is the
// port's spelling of JavaScript `undefined`.
func payloadVal(p *Payload, key string) any {
	v, ok := p.Get(key)
	if !ok {
		return nil
	}
	return v
}

// GradeStr reads a grade for mass lookup: a non-string (absent, or a bare
// boolean flag) contributes nothing, matching `MASS[g] ?? 0`.
func GradeStr(v any) string {
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

// Observation is a lens FINDING as replayed. The name is historical: it once covered both
// findings and `observe` notes, and carried the merge's `dispose` fate for each.
//
// Both retired (#327). A finding's fate is now COALESCENCE and only coalescence — it is
// addressed by being credited in some gap's found_by — so there is no Disposition to carry and
// no second kind to distinguish. A finding credited nowhere is the report's "Lens findings not
// raised to a gap", which is where that work reaches the reader.
type Observation struct {
	SeatID  string
	Key     string
	Kind    string
	Payload *Payload
}

// Board is the replayed board: gaps in mint order, observations in event order.
type Board struct {
	Events    []Event
	Anomalies []string
	// Discarded is MergedEvents' loss report, carried onto the board so every consumer of a
	// board can gate on it without re-running the replay.
	Discarded    []DiscardedShard
	GapOrder     []string
	Gaps         map[string]*Gap
	Observations []*Observation
}

func BoardState(runDir string) (*Board, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return nil, err
	}
	// ORDERED BY WHEN IT HAPPENED, not by what the shard file is called.
	//
	// Events merge per shard, so before this an entire seat replayed before the next
	// seat began, and every cross-seat reference was ordered by how seat names sort.
	// The bench closing a gap red minted was dropped SILENTLY because the gap did not
	// exist yet; the merge seat disposing a lens observation worked only because
	// red-lens sorts before red-merge — one rename from the same failure.
	//
	// (TS, SeatID, Seq) is the full key. Wall clock from many short-lived processes is
	// not monotonic and can go backwards, so time alone is not enough; the tail makes
	// ties and skew deterministic instead of arbitrary.
	ordered := make([]Event, len(m.Events))
	copy(ordered, m.Events)
	sort.SliceStable(ordered, func(i, j int) bool {
		a, b := ordered[i], ordered[j]
		if a.TS != b.TS {
			return a.TS < b.TS
		}
		if a.SeatID != b.SeatID {
			return a.SeatID < b.SeatID
		}
		return a.Seq < b.Seq
	})

	// AND THE BOARD PUBLISHES THE ORDER IT REDUCED IN.
	//
	// The sort above landed; the publish did not. `Events` was assigned m.Events BEFORE this
	// block ran, so the reduction below walked the corrected order while every CONSUMER of
	// Board.Events kept reading (Round, SeatID, Seq) — the seat-name order this whole comment
	// exists to replace. A fix applied to a local copy is a fix for one reader.
	//
	// FOUND 2026-08-22 BY A BLUE SEAT, mid-run, with a reproduction: seat `zulu` proposes at
	// 10:00 and seat `alpha` abandons at 11:00, and because `alpha` < `zulu` the projection
	// reports the line as `proposed`. It caught the defect biting its own sitting — six
	// line-of-inquiry moves replayed before the frontier proposals they answered, so
	// `show lines-of-inquiry` reported every line unmoved.
	//
	// The blast radius is every consumer that walks Board.Events rather than re-reducing:
	// record.Inquiries (behind `show lines-of-inquiry`), report assembly, the scorecards and
	// the graph renderer. They were reading a chronology assembled from filenames.
	b := &Board{Events: ordered, Anomalies: m.Anomalies, Discarded: m.Discarded, Gaps: map[string]*Gap{}}

	for _, e := range ordered {
		switch e.Type {
		case "mint":
			id := e.Payload.Str("gap_id")
			if _, exists := b.Gaps[id]; !exists {
				b.GapOrder = append(b.GapOrder, id)
			}
			b.Gaps[id] = &Gap{
				ID: id, Round: e.Round, Open: true, Mint: e.Payload,
				Severity:   payloadVal(e.Payload, "severity"),
				Likelihood: payloadVal(e.Payload, "likelihood"),
				Impact:     payloadVal(e.Payload, "impact"),
				// mirrors `{...payload}`: complexity_cost travels under its payload name
				ComplexityCost: payloadVal(e.Payload, "complexity_cost"),
			}
		case "regrade":
			g := b.Gaps[e.Payload.Str("gap_id")]
			if g == nil {
				b.Anomalies = append(b.Anomalies, missingGap("regrade", e))
				continue
			}
			// Object.assign(g, pickGrades(payload)) — only the grade keys present move.
			if e.Payload.Has("severity") {
				g.Severity = payloadVal(e.Payload, "severity")
			}
			if e.Payload.Has("likelihood") {
				g.Likelihood = payloadVal(e.Payload, "likelihood")
			}
			if e.Payload.Has("impact") {
				g.Impact = payloadVal(e.Payload, "impact")
			}
			if e.Payload.Has("complexity_cost") {
				g.ComplexityCost = payloadVal(e.Payload, "complexity_cost")
			}
			g.Regrades = append(g.Regrades, e.Payload)
		case "close":
			g := b.Gaps[e.Payload.Str("gap_id")]
			if g == nil {
				b.Anomalies = append(b.Anomalies, missingGap("close", e))
				continue
			}
			g.Open = false
			g.Closure = e.Payload
			g.ClosedRound = e.Round
			g.HasClosed = true
		case "opinion":
			// THE BENCH'S RULINGS REACH RED'S BOARD.
			//
			// Bench dispositions lived only in the judge's event stream, so the
			// projection over-reported open gaps by the number of bench closures after
			// every sitting and diverged further each round. The 2026-07-18 run's
			// red-merge-r3 measured it: the render said 9 open / 9 closed against a
			// hand-written board of 3 open / 15 closed, the difference being exactly the
			// six gaps judge-r2 had closed. Nothing carried them across, and red could
			// not close them itself without corrupting its own closure history.
			g := b.Gaps[e.Payload.Str("gap_id")]
			if g == nil {
				b.Anomalies = append(b.Anomalies, missingGap("opinion", e))
				continue
			}
			if !benchClosesGap(e.Payload.Str("disposition")) {
				continue
			}
			g.Open = false
			g.Closure = e.Payload
			g.ClosedRound = e.Round
			g.HasClosed = true
			g.ClosedByBench = true
		case "finding":
			b.Observations = append(b.Observations, &Observation{
				SeatID: e.SeatID, Key: e.Key, Kind: e.Type, Payload: e.Payload,
			})
		}
	}
	return b, nil
}

func allGapIDs(runDir string) (map[string]bool, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, e := range m.Events {
		if e.Type == "mint" {
			out[e.Payload.Str("gap_id")] = true
		}
	}
	return out, nil
}

// priorClosureRounds returns the rounds in which this gap was already closed.
//
// A `--carried-from` closure claims to restate an earlier one, and a claim about the
// record is checked against the record — the same rule mint applies to `supersedes`,
// which refuses an ancestor no mint event created.
func priorClosureRounds(runDir, gapID string) ([]int, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return nil, err
	}
	var out []int
	for _, e := range m.Events {
		if e.Type == "close" && e.Payload.Str("gap_id") == gapID {
			out = append(out, e.Round)
		}
	}
	return out, nil
}

// MintGapID assigns ids tool-side, sequentially per round — the collision class
// that made four different "R5-1"s in one round simply cannot occur.
func MintGapID(runDir string, round int) (string, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return "", err
	}
	n := 0
	for _, e := range m.Events {
		if e.Type == "mint" && e.Round == round {
			n++
		}
	}
	return fmt.Sprintf("R%d-%d", round, n+1), nil
}

// ExistingMintByKey gives crash-retry idempotency: a seat whose message died
// after a successful mint retries the SAME command, and --key (its stable local
// label) returns the EXISTING id instead of double-minting.
func ExistingMintByKey(runDir, seatID, key string) (string, error) {
	result, _, err := ExistingByKey(runDir, seatID, "mint", key)
	return result, err
}

// ---- class registry ----

type registryClass struct {
	Slug string `json:"slug"`
}

type classRegistry struct {
	Classes []registryClass `json:"classes"`
}

// loadRegistry reads the gap-class registry, and DISTINGUISHES "there is none" from "there is one
// and it is broken".
//
// It used to return nil on any failure, and `validateClass` reads nil as advisory mode — so an
// unparseable registry accepted every class slug ever passed, silently. That is this codebase's
// recurring shape: the miss and the honest zero produce the same output, and a corrupt file turns
// off the gate that keeps the class vocabulary honest without anything saying so.
//
// AN ABSENT REGISTRY IS NO LONGER ADVISORY. The tolerance read: "a run set up before the registry
// existed, or one deliberately run without one, has nothing to validate against". Two halves, and
// they are different in kind. The first is backwards compatibility for runs that no longer exist.
// The second is real — the fuzz, the seat probe and the tests build runs by hand — but the answer
// to it is to STAGE a registry, not to exempt the run from the check: StageForRun does that in one
// call, and a harness that mints against a vocabulary it never declared is not a run anybody can
// interpret afterwards.
//
// A registry that IS there and cannot be read stays its own error — somebody staged it, so somebody
// meant it to bind.
//
// An unreachable record root is not this function's error to report: MergedEvents fails loudly on
// it, and every caller reaches that first, so erroring twice only makes the diagnosis harder.
func loadRegistry(runDir string) (*classRegistry, error) {
	recDir, err := RecordsDir(runDir)
	if err != nil {
		return nil, nil
	}
	p := filepath.Join(recDir, "class-registry.json")
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, nil // absent — knownClasses turns this into a refusal, see above
	}
	var reg classRegistry
	if err := json.Unmarshal(b, &reg); err != nil {
		return nil, fmt.Errorf("record: the class registry at %s is staged but unreadable (%v) — every --class would be accepted while it stays that way, so this is refused rather than waved through. Fix the file — removing it does NOT loosen the check any more, it refuses every mint instead", p, err)
	}
	return &reg, nil
}

// knownClasses is the registry as it stands for this run: the staged corpus plus every slug the
// run has coined. An absent registry is an ERROR, not a licence — see below.
func knownClasses(runDir string) (map[string]bool, []string, error) {
	reg, err := loadRegistry(runDir)
	if err != nil {
		return nil, nil, err
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return nil, nil, err
	}
	coined := map[string]bool{}
	for _, e := range m.Events {
		if e.Type == "class-new" {
			coined[e.Payload.Str("slug")] = true
		}
	}
	// NO ADVISORY MODE. This returned nil,nil,nil when no registry was staged — "R1 tolerance; R4
	// makes it strict" — and R4 never came. A check that silently no-ops when its input is absent
	// is the plausible zero in its purest form: `mint --class anything-at-all` succeeded and the
	// gap reached the board carrying a slug nothing recognised, while the help said the registry
	// constrained it. The tolerance was the only thing standing between the two.
	//
	// An absent registry is a MISCONFIGURED RUN, not permission. run-setup stages it on every real
	// run (setup.StageClassRegistry); a run without one has lost the file, and accepting every
	// slug is the worst of the three available answers.
	if reg == nil {
		return nil, nil, fmt.Errorf("record: no gap-class registry is staged for this run — every --class would be accepted, so a mint is refused rather than waved through. run-setup stages it from feov-memory/class-registry.json; if this run was built by hand, stage one or coin the classes it needs with `class new`")
	}
	known, slugs := map[string]bool{}, []string(nil)
	for _, c := range reg.Classes {
		known[c.Slug] = true
		slugs = append(slugs, c.Slug)
	}
	for s := range coined {
		known[s] = true
	}
	return known, slugs, nil
}

// StageForRun writes a gap-class registry holding exactly these slugs, for a run built by hand.
//
// The fuzz, the seat probe and the tests construct runs without going through run-setup, and a mint
// names a class. Before this they were exempted from the class check by an absent-registry
// tolerance, which is how `--class anything-at-all` reached the board with nothing objecting. They
// declare their vocabulary now, which is both the honest fixture and the thing that makes the check
// mean something in the runs that matter.
func StageForRun(runDir string, slugs ...string) error {
	recDir, err := RecordsDir(runDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(recDir, 0o755); err != nil {
		return err
	}
	var reg classRegistry
	for _, s := range slugs {
		reg.Classes = append(reg.Classes, registryClass{Slug: s})
	}
	b, err := json.Marshal(reg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(recDir, "class-registry.json"), b, 0o644)
}

// ClassCoinedInRun reports whether this run coined the slug, rather than inheriting it from the
// staged registry.
//
// IT WAS A FLAG A SEAT SET. `--class-new` asserted "this class is new", which the registry
// already knows — so the assertion could be wrong in both directions, and the boolean's real
// meaning was "I also passed --definition, --neighbor and --distinguisher". Coining is
// `merge class new` now; this derives the fact from what that verb wrote.
func ClassCoinedInRun(runDir, slug string) bool {
	m, err := MergedEvents(runDir)
	if err != nil {
		return false
	}
	for _, e := range m.Events {
		if e.Type == "class-new" && e.Payload.Str("slug") == slug {
			return true
		}
	}
	return false
}

// validateClass holds a MINT to a class the registry recognises.
func validateClass(runDir string, p *Payload) error {
	known, slugs, err := knownClasses(runDir)
	if err != nil {
		return err
	}
	if known[p.Str("class")] {
		return nil
	}
	hintN := 6
	if len(slugs) < hintN {
		hintN = len(slugs)
	}
	return fmt.Errorf("record: unknown class %s — use a registry slug (e.g. %s, ...) or coin this one first with `merge class new --class %s --definition ... --neighbor ... --distinguisher ...`",
		p.Str("class"), strings.Join(slugs[:hintN], ", "), p.Str("class"))
}

// validateClassNew holds a COINING to the three facts that make a class discriminate: what it is,
// what it is nearest to, and how to tell them apart. The neighbor must itself be known, or the
// registry grows a tree with no root.
func validateClassNew(runDir string, p *Payload) error {
	if p.Str("slug") == "" {
		return fmt.Errorf("record: class new requires --class (the slug being coined)")
	}
	for _, f := range []string{"definition", "neighbor", "distinguisher"} {
		if p.Str(f) == "" {
			return fmt.Errorf("record: class new requires --%s — a class coined without one of definition, neighbor and distinguisher is a synonym, and the registry stops discriminating", f)
		}
	}
	known, _, err := knownClasses(runDir)
	if err != nil {
		return err
	}
	if known == nil {
		return nil
	}
	if known[p.Str("slug")] {
		return fmt.Errorf("record: class %s already exists — mint against it with `merge mint --class %s` rather than coining it twice", p.Str("slug"), p.Str("slug"))
	}
	if !known[p.Str("neighbor")] {
		return fmt.Errorf("record: --neighbor %s is not a known class", p.Str("neighbor"))
	}
	return nil
}
