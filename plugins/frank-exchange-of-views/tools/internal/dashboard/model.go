package dashboard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cost"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// Config carries the run's launch parameters the dashboard displays. File values (setup's
// inputs/run-config.json) are overridden by any non-empty CLI flag — the JS merge order.
type Config struct {
	Topic         string `json:"topic"`
	Model         string `json:"model"`
	JudgmentModel string `json:"judgmentModel"`
	MaxRounds     string `json:"maxRounds"`
	Lanes         string `json:"lanes"`
}

// Friction is the friction tile's data (count + latest "seat: text"); Count<0 == unavailable
// (the JS null, rendered "unavailable"), since a Go int has no null.
type Friction struct {
	Count int // -1 == unavailable
	Last  string
}

// Shards is the Red's-board view data. Findings/Citations are -1 when unavailable (JS null).
type Shards struct {
	LedgerExists     bool
	OpenRows         int
	OpenBySeverity   map[string]int
	Findings         int // -1 == unavailable
	Citations        int // -1 == unavailable
	ClosureIndexRows int
	ArchiveRecords   int
}

type Step struct{ Name, State string }

type Rate struct {
	Round     any
	Opened    int
	Closed    int
	Open      any
	CloseRate int
}

type Judiciary struct {
	JudgeSittings int
	Rulings       map[string]int
	Disputes      struct{ Raised, Accepted, Rejected int }
	ChainSpans    map[int]int
	Chains        int
	MigDown       int
	MigUp         int
	MigFlat       int
	LatestVerdict string
	VerdictRound  int
}

// CostRow is one seat-round-tier cost bucket for the dashboard's per-seat-round breakdown.
type CostRow struct {
	Round  int
	Seat   string
	Tier   string
	Agents int
	Cost   float64
}

// Model is what renderHtml consumes — the Go analogue of buildModel's returned object.
type Model struct {
	RunDir          string
	Telemetry       []map[string]any
	Latest          map[string]any
	Seats           []Seat
	Cost            float64
	CostRows        []CostRow
	APIRounds       int
	Agents          int
	Friction        Friction
	Shards          Shards
	BlueClaims      *int
	Steps           []Step
	Rates           []Rate
	Judiciary       Judiciary
	Eta             Eta
	Config          Config
	TerminalVerdict string
	Terminal        bool
	Generated       string
}

func jsonl(path string) []map[string]any {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var out []map[string]any
	for _, line := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var m map[string]any
		if json.Unmarshal([]byte(line), &m) == nil {
			out = append(out, m)
		}
	}
	return out
}

func numOr(v any, d float64) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return d
}

// BuildModel gathers the run's instruments — the Go port of buildModel. It reads the record
// IN-PROCESS (BoardState → the JSON views) rather than spawning the tool; nowMs is injected
// (the CLI defaults it to the real clock) so tests are deterministic.
func BuildModel(runDir, transcriptDir string, cfg Config, nowMs float64) Model {
	// Config: file (setup) then non-empty CLI overrides.
	var fileCfg Config
	if b, err := os.ReadFile(filepath.Join(runDir, "inputs", "run-config.json")); err == nil {
		_ = json.Unmarshal(b, &fileCfg)
	}
	merged := fileCfg
	if cfg.Topic != "" {
		merged.Topic = cfg.Topic
	}
	if cfg.Model != "" {
		merged.Model = cfg.Model
	}
	if cfg.JudgmentModel != "" {
		merged.JudgmentModel = cfg.JudgmentModel
	}
	if cfg.MaxRounds != "" {
		merged.MaxRounds = cfg.MaxRounds
	}
	if cfg.Lanes != "" {
		merged.Lanes = cfg.Lanes
	}

	// Telemetry: computed on read from the record via the shared view library — no
	// materialized board-telemetry.jsonl, no markdown/jsonl intermediate.
	telemetry, _ := view.Telemetry(runDir)
	journal := jsonl(filepath.Join(transcriptDir, "journal.jsonl"))

	// Lifecycle by agentId; identity by transcript-head classification; times from the file.
	type raw struct {
		done   bool
		result any
	}
	byID := map[string]*raw{}
	var idOrder []string
	for _, j := range journal {
		id, _ := j["agentId"].(string)
		if id == "" {
			continue
		}
		s := byID[id]
		if s == nil {
			s = &raw{}
			byID[id] = s
			idOrder = append(idOrder, id)
		}
		if r, ok := j["result"]; ok {
			s.done = true
			s.result = r
		}
	}
	var seats []Seat
	for _, id := range idOrder {
		s := byID[id]
		tp := filepath.Join(transcriptDir, "agent-"+id+".jsonl")
		head := ""
		if b, err := os.ReadFile(tp); err == nil {
			if len(b) > 3000 {
				b = b[:3000]
			}
			head = string(b)
		}
		var startedMs, endedMs *float64
		if fi, err := os.Stat(tp); err == nil {
			st := seatStart(tp, fi)
			startedMs = &st
			if s.done {
				m := float64(fi.ModTime().UnixNano()) / 1e6
				endedMs = &m
			}
		}
		c := seatclass.ClassifySeat(head)
		label := c.Seat
		if c.Round != 0 {
			label = c.Seat + "-r" + itoa(c.Round)
		}
		seats = append(seats, Seat{AgentID: id, Done: s.done, Result: s.result, Label: label, Seat: c.Seat, Round: c.Round, StartedMs: startedMs, EndedMs: endedMs})
	}

	// Cost from transcripts.
	var costTotal float64
	var costRows []CostRow
	apiRounds, agents := 0, 0
	if entries, err := os.ReadDir(transcriptDir); err == nil {
		var files []string
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "agent-") && strings.HasSuffix(e.Name(), ".jsonl") {
				files = append(files, e.Name())
			}
		}
		sort.Strings(files)
		for _, f := range files {
			agents++
			for _, j := range jsonl(filepath.Join(transcriptDir, f)) {
				msg, _ := j["message"].(map[string]any)
				if msg == nil {
					continue
				}
				u, _ := msg["usage"].(map[string]any)
				if u == nil {
					continue
				}
				apiRounds++
				model, _ := msg["model"].(string)
				p := cost.Prices[cost.Tier(model)]
				costTotal += (numOr(u["input_tokens"], 0)*p[0] + numOr(u["output_tokens"], 0)*p[1] +
					numOr(u["cache_read_input_tokens"], 0)*p[2] + numOr(u["cache_creation_input_tokens"], 0)*p[3]) / 1e6
			}
		}
		// Per-seat-round breakdown, single-sourced from cost.ScanTranscript/Aggregate (the pricing
		// cost.md uses), so the dashboard shows WHERE the spend went, not only the total.
		var crows []cost.Row
		for _, f := range files {
			if b, err := os.ReadFile(filepath.Join(transcriptDir, f)); err == nil {
				crows = append(crows, cost.ScanTranscript(string(b)))
			}
		}
		agg := cost.Aggregate(crows)
		keys := make([]string, 0, len(agg))
		for k := range agg {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			parts := strings.SplitN(k, "|", 3)
			rnd, _ := strconv.Atoi(parts[0])
			costRows = append(costRows, CostRow{Round: rnd, Seat: parts[1], Tier: parts[2], Agents: agg[k].N, Cost: agg[k].Cost})
		}
	}

	// Record views in-process (BoardState); unavailable ⟺ no record yet. The JS gated the
	// board tiles on config.bin (the operator passing the tool path); the Go binary IS the
	// tool, so it gates on the record's EXISTENCE — no records/ dir ⟺ nothing to read, which
	// must render "unavailable", NOT a misleading all-zero board (BoardState returns an empty
	// board, not an error, when the dir is absent).
	friction := Friction{Count: -1}
	shards := Shards{OpenBySeverity: map[string]int{}, Findings: -1, Citations: -1}
	// Resolution failure lands in the same arm as "no record yet", and here that is RIGHT: both
	// mean the tiles have nothing truthful to show, and "unavailable" is what this branch already
	// exists to render. It is a dashboard — the loud version of the diagnosis belongs to the tool.
	recDir, recErr := record.RecordsDir(runDir)
	if _, statErr := os.Stat(recDir); recErr == nil && statErr == nil {
		if board, err := record.BoardState(runDir); err == nil {
			bj := record.BoardJSONOf(board)
			fj := record.FindingsJSONOf(board)
			frj := record.FrictionJSONOf(board)
			friction.Count = frj.Counts.Total
			if n := len(frj.Friction); n > 0 {
				last := frj.Friction[n-1]
				friction.Last = last.SeatID + ": " + last.Text
			}
			obs := map[string]int{}
			for _, g := range bj.Open {
				s := strings.TrimSpace(strings.ToLower(anyStr(g.Severity)))
				if s != "" {
					obs[s]++
				}
			}
			shards = Shards{
				LedgerExists: true, OpenRows: bj.Counts.Open, OpenBySeverity: obs,
				Findings: fj.Counts.Total, Citations: bj.Counts.Citations,
				ClosureIndexRows: bj.Counts.Closed, ArchiveRecords: len(bj.Closed),
			}
		}
	}

	// Blue claims: last journal result carrying a numeric claim_count.
	var blueClaims *int
	for _, j := range journal {
		r, _ := j["result"].(map[string]any)
		if r == nil {
			continue
		}
		if c, ok := r["claim_count"].(float64); ok {
			n := int(c)
			blueClaims = &n
		}
	}

	jud := buildJudiciary(journal)
	// CHRONOLOGICAL, NOT JOURNAL ORDER. idOrder is first-appearance in the workflow journal, which
	// is DISPATCH order — and dispatch order is arbitrary for a parallel() batch (the round-1
	// lenses landed L6, L5, L1) and is reshuffled again by a resume, where cached agents replay
	// instead of re-dispatching. A reader scanning the seat list is asking "what happened, in what
	// order", so order by StartedMs — already computed above and, until now, never used to order.
	//
	// HONEST CAVEAT, because the field does not mean the same thing everywhere: StartedMs is a true
	// start only where the platform records a birth time — seat_windows.go reads CreationTime,
	// seat_linux.go statx BTIME, and seat_fallback.go has neither and returns ModTime, the LAST
	// write. On the fallback this therefore sorts by COMPLETION, not start. Still more truthful
	// than dispatch order, but the field wants fixing at its source: the agent transcript's first
	// JSONL line carries a real "timestamp". Tracked separately — the same defect makes EndedMs
	// equal StartedMs there, so every completed seat renders a ZERO duration and the ETA built on
	// it is unreliable. Seats with no transcript yet sort last; STABLE, so a tie keeps journal order.
	sort.SliceStable(seats, func(i, j int) bool {
		a, b := seats[i].StartedMs, seats[j].StartedMs
		if a == nil || b == nil {
			return a != nil // a started, b did not → a first
		}
		return *a < *b
	})

	steps := buildSteps(seats, merged.MaxRounds)
	rates := buildRates(telemetry)

	var latest map[string]any
	if len(telemetry) > 0 {
		latest = telemetry[len(telemetry)-1]
	}
	eta := projectCompletion(seats, nowMs)

	return Model{
		RunDir: runDir, Telemetry: telemetry, Latest: latest, Seats: seats,
		Cost: costTotal, CostRows: costRows, APIRounds: apiRounds, Agents: agents, Friction: friction,
		Shards: shards, BlueClaims: blueClaims, Steps: steps, Rates: rates,
		Judiciary: jud, Eta: eta, Config: merged,
		TerminalVerdict: readTerminalVerdict(runDir), Terminal: fileExists(filepath.Join(runDir, "report.md")),
		Generated: nowISO(nowMs),
	}
}

func fileExists(p string) bool { _, err := os.Stat(p); return err == nil }

// readTerminalVerdict answers from the RECORD, and falls back to the rendered report only when
// the record cannot say.
//
// IT WAS THE REGEX ALONE, over report.md. The verdict is a field on the `outcome` event — the
// tool's own terminal act writes it — and this recovered it from the prose that field had already
// been rendered into. That is the shape [[facts-are-fields]] names: a fact bypassing the record
// that already holds it, recovered by a hope about string shape.
//
// The hope was load-bearing in a specific way. `assemble.basisNote` appends " — **derived from the
// record**…" straight after the verdict word, so the pattern must stop at an em-dash — which
// couples this reader to a sentence assemble.go is free to rewrite. Measured against the 9 real
// assembled reports in research/: 8 matched. The ninth has no verdict at all, and produced exactly
// the same "" as a drifted pattern would.
//
// The record is asked first and answers in two ways that this deliberately keeps apart:
// the outcome event's own field (the bench recorded a terminal act), then DeriveVerdict (the
// record decides for itself). The regex survives only for a run assembled before the event
// existed, and its miss no longer hides behind the same "" as an honest absence — an unreadable
// record and a missing verdict now say different things to the caller.
func readTerminalVerdict(runDir string) string {
	if b, err := record.BoardState(runDir); err == nil {
		for i := len(b.Events) - 1; i >= 0; i-- {
			if b.Events[i].Type == "outcome" {
				if v := b.Events[i].Payload.Str("verdict"); v != "" {
					return v
				}
			}
		}
	}
	if v, _, ok := record.DeriveVerdict(runDir); ok {
		return v
	}
	// LAST RESORT, and named as one: a run assembled before the terminal act was an event.
	b, err := os.ReadFile(filepath.Join(runDir, "report.md"))
	if err != nil {
		return ""
	}
	if m := verdictRe.FindStringSubmatch(string(b)); m != nil {
		return m[1]
	}
	return ""
}

// buildJudiciary ports the journal-envelope analytics: rulings by type, dispute traffic, and
// argument longevity over supersedes chains (union-find), with grade migration.
func buildJudiciary(journal []map[string]any) Judiciary {
	j := Judiciary{Rulings: map[string]int{}, ChainSpans: map[int]int{}}
	type ge struct {
		first, last         int
		firstMass, lastMass float64
	}
	gapRounds := map[string]*ge{}
	var gapOrder []string
	parent := map[string]string{}
	var find func(string) string
	find = func(x string) string {
		for {
			p, ok := parent[x]
			if !ok || p == x {
				return x
			}
			x = p
		}
	}
	redSeen := 0
	for _, env := range journal {
		r, _ := env["result"].(map[string]any)
		if r == nil {
			continue
		}
		if res, ok := r["resolutions"].([]any); ok {
			j.JudgeSittings++
			for _, x := range res {
				m, _ := x.(map[string]any)
				if s, ok := m["resolution"].(string); ok {
					j.Rulings[s]++
				}
			}
		}
		if gd, ok := r["grade_disputes"].([]any); ok {
			j.Disputes.Raised += len(gd)
		}
		if dr, ok := r["dispute_responses"].([]any); ok {
			for _, d := range dr {
				m, _ := d.(map[string]any)
				switch m["response"] {
				case "accepted":
					j.Disputes.Accepted++
				case "rejected":
					j.Disputes.Rejected++
				}
			}
		}
		gaps, hasGaps := r["gaps"].([]any)
		if v, ok := r["verdict"].(string); ok && hasGaps {
			redSeen++
			j.LatestVerdict = v
			j.VerdictRound = redSeen
			for _, gx := range gaps {
				g, _ := gx.(map[string]any)
				id, _ := g["id"].(string)
				gm := record.GapMass(anyStr(g["likelihood"]), anyStr(g["impact"]))
				e := gapRounds[id]
				if e == nil {
					e = &ge{first: redSeen, firstMass: gm}
					gapRounds[id] = e
					gapOrder = append(gapOrder, id)
				}
				e.last = redSeen
				e.lastMass = gm
				if sup, ok := g["supersedes"].([]any); ok {
					for _, a := range sup {
						if anc, ok := a.(string); ok {
							parent[id] = find(anc)
						}
					}
				}
			}
		}
	}
	type chain struct {
		first, last         int
		firstMass, lastMass float64
		ids                 int
	}
	chains := map[string]*chain{}
	for _, id := range gapOrder {
		e := gapRounds[id]
		root := find(id)
		c := chains[root]
		if c == nil {
			c = &chain{first: e.first, last: e.last, firstMass: e.firstMass, lastMass: e.lastMass}
			chains[root] = c
		}
		c.ids++
		if e.first <= c.first {
			c.first = e.first
			c.firstMass = e.firstMass
		}
		if e.last >= c.last {
			c.last = e.last
			c.lastMass = e.lastMass
		}
	}
	for _, c := range chains {
		span := c.last - c.first + 1
		j.ChainSpans[span]++
		if span > 1 {
			d := c.lastMass - c.firstMass
			switch {
			case d < 0:
				j.MigDown++
			case d > 0:
				j.MigUp++
			default:
				j.MigFlat++
			}
		}
	}
	j.Chains = len(chains)
	return j
}

func buildSteps(seats []Seat, maxRoundsStr string) []Step {
	maxRounds := 8
	if maxRoundsStr != "" {
		if n := atoiOr(maxRoundsStr, 8); n > 0 {
			maxRounds = n
		}
	}
	seen := func(seat string, round int) bool {
		for _, s := range seats {
			if s.Seat == seat && (round == 0 || s.Round == round) {
				return true
			}
		}
		return false
	}
	doneSeat := func(seat string, round int) bool {
		for _, s := range seats {
			if s.Seat == seat && (round == 0 || s.Round == round) && s.Done {
				return true
			}
		}
		return false
	}
	allDone := func(seat string) bool {
		n, all := 0, true
		for _, s := range seats {
			if s.Seat == seat {
				n++
				if !s.Done {
					all = false
				}
			}
		}
		return n > 0 && all
	}
	state := func(done, live bool) string {
		if done {
			return "done"
		}
		if live {
			return "live"
		}
		return "todo"
	}
	steps := []Step{
		{"frontier", state(doneSeat("frontier", 0), seen("frontier", 0))},
		{"blue lanes", state(allDone("blue-lane"), seen("blue-lane", 0))},
		{"synthesis", state(doneSeat("blue-synthesize", 0), seen("blue-synthesize", 0))},
	}
	for r := 1; r <= maxRounds; r++ {
		roundDone := doneSeat("blue-respond", r)
		anySeen := seen("red-lens", r) || seen("red-merge", r) || seen("blue-respond", r) || seen("judge", r)
		steps = append(steps, Step{"round " + itoa(r), state(roundDone, anySeen)})
	}
	steps = append(steps, Step{"assembly", state(doneSeat("assemble", 0), seen("assemble", 0))})
	return steps
}

func buildRates(telemetry []map[string]any) []Rate {
	rates := make([]Rate, 0, len(telemetry))
	for i, t := range telemetry {
		prevOpen := 0.0
		if i > 0 {
			prevOpen = numOr(telemetry[i-1]["open_count"], 0)
		}
		minted := 0.0
		if nm, ok := t["new_mint"].(map[string]any); ok {
			minted = numOr(nm["count"], 0)
		}
		open := numOr(t["open_count"], 0)
		closed := prevOpen + minted - open
		closeRate := 0
		if minted+prevOpen > 0 {
			closeRate = int(round(100 * closed / (prevOpen + minted)))
		}
		rates = append(rates, Rate{Round: t["round"], Opened: int(minted), Closed: int(closed), Open: t["open_count"], CloseRate: closeRate})
	}
	return rates
}
