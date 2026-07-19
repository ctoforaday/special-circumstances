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
	"time"
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
	mtime  time.Time
}

// Merged is the deterministic replay: winner selection per seat, global ordering,
// key-level dedup, and the anomalies that must never be silently normalized.
type Merged struct {
	Events    []Event
	Anomalies []string
}

var shardRe = regexp.MustCompile(`^events-(.+)-([0-9a-f]{8})\.jsonl$`)

func MergedEvents(runDir string) (Merged, error) {
	dir := recordsDir(runDir)
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
		st, err := os.Stat(full)
		if err != nil {
			return Merged{}, err
		}
		if _, seen := bySeat[seatID]; !seen {
			seatOrder = append(seatOrder, seatID)
		}
		bySeat[seatID] = append(bySeat[seatID], shardInfo{nonce: nonce, file: full, events: evs, mtime: st.ModTime()})
	}

	var anomalies []string
	var winners []Event
	for _, seatID := range seatOrder {
		shards := bySeat[seatID]
		if len(shards) == 1 {
			winners = append(winners, shards[0].events...)
			continue
		}
		// Multi-nonce: the winner is the nonce whose shard carries the seat's
		// TERMINAL event (verdict or revision — the last verb of a seat contract);
		// with neither terminal, fall EXPLICITLY to latest mtime.
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
		winner := pool[0]
		for _, s := range pool[1:] {
			if s.mtime.After(winner.mtime) { // reduce((a,b) => b.mtime > a.mtime ? b : a)
				winner = s
			}
		}
		nonces := make([]string, len(shards))
		for i, s := range shards {
			nonces[i] = s.nonce
		}
		by := "mtime fallback"
		if len(terminal) > 0 {
			by = "terminal event"
		}
		anomalies = append(anomalies, fmt.Sprintf("multi-nonce seat %s: %d dispatches (%s); winner %s by %s",
			seatID, len(shards), strings.Join(nonces, ", "), winner.nonce, by))
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
	return Merged{Events: events, Anomalies: anomalies}, nil
}

// Gap is the replayed state of one board gap.
//
// The four grades are `any` rather than string because ABSENCE is renderable:
// the oracle interpolates them straight into a template literal, so a gap minted
// without --cx renders the literal text "undefined". Collapsing that to "" here
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
	switch disposition {
	case "closed", "risk_accepted", "rebuttal_sustained", "routed_to_infrastructure":
		return true
	default:
		return false
	}
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

// Observation is a lens finding or note plus the merge's disposition of it.
type Observation struct {
	SeatID      string
	Key         string
	Kind        string
	Payload     *Payload
	Disposition *Payload
}

// Board is the replayed board: gaps in mint order, observations in event order.
type Board struct {
	Events       []Event
	Anomalies    []string
	GapOrder     []string
	Gaps         map[string]*Gap
	Observations []*Observation
}

func BoardState(runDir string) (*Board, error) {
	m, err := MergedEvents(runDir)
	if err != nil {
		return nil, err
	}
	b := &Board{Events: m.Events, Anomalies: m.Anomalies, Gaps: map[string]*Gap{}}

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
		case "finding", "observe":
			b.Observations = append(b.Observations, &Observation{
				SeatID: e.SeatID, Key: e.Key, Kind: e.Type, Payload: e.Payload,
			})
		case "dispose":
			target := e.Payload.Str("observation")
			for _, o := range b.Observations { // find() — FIRST match wins
				name := o.Payload.Str("label")
				if name == "" {
					name = o.Key
				}
				if name == target {
					o.Disposition = e.Payload
					break
				}
			}
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
	if key == "" {
		return "", nil
	}
	m, err := MergedEvents(runDir)
	if err != nil {
		return "", err
	}
	for _, e := range m.Events {
		if e.Type == "mint" && e.SeatID == seatID && e.Payload.Str("mint_key") == key {
			return e.Payload.Str("gap_id"), nil
		}
	}
	return "", nil
}

// ---- class registry ----

type registryClass struct {
	Slug string `json:"slug"`
}

type classRegistry struct {
	Classes []registryClass `json:"classes"`
}

func loadRegistry(runDir string) *classRegistry {
	p := filepath.Join(recordsDir(runDir), "class-registry.json")
	b, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	var reg classRegistry
	if err := json.Unmarshal(b, &reg); err != nil {
		return nil // unparseable registry degrades to advisory mode, as in the oracle
	}
	return &reg
}

func validateClass(runDir string, p *Payload) error {
	reg := loadRegistry(runDir)
	m, err := MergedEvents(runDir)
	if err != nil {
		return err
	}
	var extensions []string
	for _, e := range m.Events {
		if e.Type == "class-new" {
			extensions = append(extensions, e.Payload.Str("slug"))
		}
	}
	classNew, _ := p.Get("class_new")
	if b, ok := classNew.(bool); ok && b {
		if p.Str("definition") == "" || p.Str("neighbor") == "" || p.Str("distinguisher") == "" {
			return fmt.Errorf("record: --class-new requires --definition, --neighbor <existing-slug>, and --distinguisher (the tie-break question)")
		}
		if reg != nil {
			known := map[string]bool{}
			for _, c := range reg.Classes {
				known[c.Slug] = true
			}
			for _, s := range extensions {
				known[s] = true
			}
			if !known[p.Str("neighbor")] {
				return fmt.Errorf("record: --neighbor %s is not a known class", p.Str("neighbor"))
			}
		}
		return nil
	}
	if reg == nil {
		return nil // no registry staged — advisory mode (R1 tolerance; R4 makes it strict)
	}
	known := map[string]bool{}
	var slugs []string
	for _, c := range reg.Classes {
		known[c.Slug] = true
		slugs = append(slugs, c.Slug)
	}
	for _, s := range extensions {
		known[s] = true
	}
	if !known[p.Str("class")] {
		hintN := 6
		if len(slugs) < hintN {
			hintN = len(slugs)
		}
		hint := strings.Join(slugs[:hintN], ", ")
		return fmt.Errorf("record: unknown class %s — use a registry slug (e.g. %s, ...) or mint the class with --class-new (definition + neighbor + distinguisher required)", p.Str("class"), hint)
	}
	return nil
}
