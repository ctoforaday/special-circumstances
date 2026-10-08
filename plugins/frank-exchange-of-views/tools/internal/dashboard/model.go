package dashboard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cost"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// Config carries the run's launch parameters the dashboard displays. File values (setup's
// inputs/run-config.json) are overridden by any non-empty CLI flag — the JS merge order.
type Config struct {
	Topic         string `json:"topic"`
	Model         string `json:"model"`
	JudgmentModel string `json:"judgmentModel"`
	Lanes         string `json:"lanes"`
}

// LogTile is the log tile's data: how many log entries the record holds, of every type, and the
// latest as "seat: text". Count<0 == unavailable (the JS null, rendered "unavailable"), since a
// Go int has no null.
type LogTile struct {
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
	// Epoch and Open are POINTERS because the schema makes them optional: a row that carries no
	// epoch is a different thing from epoch 0, and the dashboard prints "—" for the first and a
	// number for the second. They were `any` holding a json.Number or a nil map entry, which
	// collapsed those two cases into whatever anyStr made of them.
	Epoch     *int32
	Opened    int
	Closed    int
	Open      *int32
	CloseRate int
}

type Judiciary struct {
	// Measured is false when the record could not be read: every figure below is then not
	// measured, never zero, and the page says so.
	Measured      bool
	JudgeSittings int
	Rulings       map[string]int
	Disputes      struct{ Raised, Accepted, Rejected int }
	ChainSpans    map[int]int
	Chains        int
	MigDown       int
	MigUp         int
	MigFlat       int
	LatestVerdict string
	// VerdictEpoch is the epoch whose recorded gate delivered LatestVerdict.
	VerdictEpoch int
}

// CostRow is one seat-epoch-tier cost bucket for the dashboard's per-seat-epoch breakdown.
type CostRow struct {
	Epoch  int
	Seat   string
	Tier   string
	Agents int
	Cost   float64
}

// Model is what renderHtml consumes — the Go analogue of buildModel's returned object.
type Model struct {
	// Run, not a path: the model is built from a resolved run and both its readers want that
	// same run, so carrying the string would mean one of them re-deriving what the other
	// already holds. No JSON tag — this struct is consumed by renderHtml alone.
	Run             record.Run
	Telemetry       []*recordpb.TelemetryLine
	Latest          *recordpb.TelemetryLine
	Seats           []Seat
	Cost            float64
	CostRows        []CostRow
	APIRounds       int
	Agents          int
	Log             LogTile
	Shards          Shards
	BlueClaims      *int
	Steps           []Step
	Rates           []Rate
	Judiciary       Judiciary
	Eta             Eta
	Config          Config
	TerminalVerdict string
	Terminal        bool
	// Live is the run's PULL-BASED liveness — see record.Liveness. The board used to answer
	// "is this running" from the presence of a marker that a killed workflow can never lift,
	// so a dead run rendered as live forever, ETA and all.
	Live      record.Liveness
	Generated string
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

// BuildModel gathers the run's instruments — the Go port of buildModel. It reads the record
// IN-PROCESS (BoardState → the JSON views) rather than spawning the tool; nowMs is injected
// (the CLI defaults it to the real clock) so tests are deterministic.
func BuildModel(run record.Run, transcriptDir string, cfg Config, nowMs float64) Model {
	// Config: file (setup) then non-empty CLI overrides.
	var fileCfg Config
	if b, err := os.ReadFile(filepath.Join(run.Dir(), "inputs", "run-config.json")); err == nil {
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
	if cfg.Lanes != "" {
		merged.Lanes = cfg.Lanes
	}

	// Telemetry: computed on read from the record via the shared view library — no
	// materialized board-telemetry.jsonl, no markdown/jsonl intermediate.
	telemetry, _ := view.Telemetry(run)
	journal := jsonl(filepath.Join(transcriptDir, "journal.jsonl"))

	// THE RECORD, READ ONCE, ahead of the seats and the cost that both join to it. Unavailable ⟺
	// no record yet: no records/ dir means nothing to read, which must render "unavailable", NOT a
	// misleading all-zero board. Resolution failure lands in the same arm, and here that is RIGHT:
	// both mean the tiles have nothing truthful to show. It is a dashboard — the loud version of
	// the diagnosis belongs to the tool.
	var fam record.Family
	haveRecord := false
	if _, statErr := os.Stat(run.Records()); statErr == nil {
		if f, err := record.FamilyOf(run); err == nil {
			fam, haveRecord = f, true
		}
	}
	// agentId → (seat, epoch, sitting), off the register events. The label used to be the seat
	// class plus an epoch PARSED OUT OF THE TRANSCRIPT HEAD; the record already holds the binding
	// as a field on `register` and counts both windows off the stream, so the head no longer says
	// which sitting this is — only which CLASS of seat, which is seatclass's job and stays so.
	bindings := cost.SeatBindingsOf(fam.Events, fam.At)

	// Lifecycle by agentId; class by transcript-head classification; identity (which seat, which
	// sitting) from the record; times from the file.
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
		// `seat #sitting` where the record bound the agent; the bare class where it did not — an
		// unbound agent has no sitting to number, and a number invented for it would read as one.
		label := c.Seat
		b, bound := bindings[id]
		if bound {
			label = b.SeatID + " #" + itoa(b.Sitting)
			// THE OCCASION GOES IN THE LABEL, because that is the whole reason it is recorded: a
			// bench sitting used to read as `judge #4` and a human could not tell which of the four
			// questions it answered. It is empty for every seat whose id already says.
			if b.Occasion != "" {
				label += " · " + b.Occasion
			}
		}
		seats = append(seats, Seat{AgentID: id, Done: s.done, Result: s.result, Label: label, Seat: c.Seat, Occasion: b.Occasion, Epoch: b.Epoch, Sitting: b.Sitting, StartedMs: startedMs, EndedMs: endedMs})
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
		// ONE PASS, AND ONE DEFINITION OF THE COST.
		//
		// This read and fully JSON-parsed every transcript TWICE per render: once here for the
		// headline total, and again below for the per-seat-epoch breakdown. The files are
		// append-only and grow through the run, so both halves got more expensive together
		// (#684 F15).
		//
		// Collapsing them also removes a divergence nobody was checking. The total priced each
		// MESSAGE at its own `model` field; cost.ScanTranscript picks one model per FILE and
		// prices that file's totals at that tier. For a single-model transcript — every ordinary
		// seat — the two agree exactly. For a transcript carrying more than one model, which is
		// what `setup --allow-substitution` exists to permit, they do not, and the page would
		// show a total that did not equal the breakdown printed under it. The breakdown's
		// definition wins because it is the one cost.md already uses; a dashboard and a cost
		// report disagreeing about one run is the two-readers defect this codebase keeps finding.
		var crows []cost.Row
		for _, f := range files {
			agents++
			b, err := os.ReadFile(filepath.Join(transcriptDir, f))
			if err != nil {
				continue
			}
			row := cost.ScanTranscript(string(b))
			// Both windows come from the record's binding — the epoch it always did, and the
			// occasion, which keeps the assembly's spend out of a docket ruling's row.
			bind := bindings[cost.AgentIDOfTranscript(f)]
			row.Epoch, row.Occasion = bind.Epoch, bind.Occasion
			crows = append(crows, row)
			apiRounds += row.Turns
			costTotal += row.Cost
		}
		agg := cost.Aggregate(crows)
		keys := make([]string, 0, len(agg))
		for k := range agg {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			parts := strings.SplitN(k, "|", 3)
			epoch, _ := strconv.Atoi(parts[0])
			costRows = append(costRows, CostRow{Epoch: epoch, Seat: parts[1], Tier: parts[2], Agents: agg[k].N, Cost: agg[k].Cost})
		}
	}

	// Record views in-process, off the family read above. The JS gated the board tiles on
	// config.bin (the operator passing the tool path); the Go binary IS the tool, so it gates on
	// the record's EXISTENCE.
	logTile := LogTile{Count: -1}
	shards := Shards{OpenBySeverity: map[string]int{}, Findings: -1, Citations: -1}
	if haveRecord {
		bj, bjErr := record.BoardJSONOfRun(run)
		if bjErr != nil {
			bj = record.BoardJSON{Open: []record.GapJSON{}, Closed: []record.GapJSON{}, Anomalies: []string{}}
		}
		fj := record.FindingsJSONOf(fam.Events, fam.At)
		frj := record.LogJSONOf(fam.Events, fam.At)
		logTile.Count = frj.Counts.Total
		if n := len(frj.Log); n > 0 {
			last := frj.Log[n-1]
			logTile.Last = last.SeatID + ": " + last.Text
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

	var jud Judiciary
	if haveRecord {
		jud = buildJudiciary(fam)
	}
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

	steps := buildSteps(seats)
	rates := buildRates(telemetry)

	var latest *recordpb.TelemetryLine
	if len(telemetry) > 0 {
		latest = telemetry[len(telemetry)-1]
	}
	eta := projectCompletion(seats, nowMs)

	// A record the dashboard cannot read renders as a run with no terminal verdict: the next
	// render asks again, and the Model has no surface for a read error — the other instruments
	// on this page read the same record and show its state. Folded here, at the render, for that.
	terminalVerdict, _ := record.TerminalVerdict(run)
	return Model{
		Run: run, Telemetry: telemetry, Latest: latest, Seats: seats,
		Cost: costTotal, CostRows: costRows, APIRounds: apiRounds, Agents: agents, Log: logTile,
		Shards: shards, BlueClaims: blueClaims, Steps: steps, Rates: rates,
		Judiciary: jud, Eta: eta, Config: merged,
		// ONE READ, TWO USES, so the pair cannot disagree — and Terminal now answers from the
		// record like its neighbour instead of from a filename.
		//
		// It was a stat of run.Dir()/report.md. setup's skeleton then stubbed report.md (while
		// documenting it as `bench assemble`'s output), so
		// Terminal was true from the moment setup ran, before a seat was dispatched, for the
		// entire life of every run. Measured 2026-08-22: the dashboard rendered "run complete —
		// the assembler wrote the report" while blue-lane-1 was visibly live in the very next
		// section, and went on saying it for 55 minutes.
		//
		// The essay on record.TerminalVerdict is about precisely this shape — a fact the record
		// holds, recovered from the prose or the filename it was rendered into. It was written one
		// line above the field that did it.
		TerminalVerdict: terminalVerdict, Terminal: terminalVerdict != "",
		// ASSESSED AT THE INJECTED CLOCK, not time.Now(), so a test can put the record in the
		// past and watch this flip.
		Live:      record.Assess(run, time.UnixMilli(int64(nowMs)).UTC(), terminalVerdict != ""),
		Generated: nowISO(nowMs),
	}
}

// buildJudiciary reads the bench's traffic off the record: judge sittings, rulings by disposition,
// grade-motion traffic, the latest recorded verdict, and how long arguments live over supersedes
// chains (union-find), with grade migration.
//
// THE RECORD, NOT THE JOURNAL. The record holds every input as a field a writer can refuse —
// docket and grade motions with their rulings, the chair's gate, each gap's mint and close epoch,
// its grades then and now, and what it supersedes. A journal envelope carries only what its
// seat's schema declares, and a key no schema declares reads as an empty list on every run: the
// page for a bench that never sat.
func buildJudiciary(fam record.Family) Judiciary {
	j := Judiciary{Measured: true, Rulings: map[string]int{}, ChainSpans: map[int]int{}}
	// A judge sitting is a stored sitting whose register names an occasion: the record refuses an
	// occasion on any seat but the bench and requires one on the bench, so every occasion — docket,
	// petition, terminal, assemble — is the bench sitting. Counted by the sitting the record holds
	// each register in, so a register that joined its hook's bracket, a second register in one
	// sitting, or a repair's register adds none.
	judgeSat := map[int64]bool{}
	for _, e := range fam.Live() {
		reg, ok := recordpb.BodyAs[*recordpb.Register](e)
		if !ok || reg.GetOccasion() == recordpb.Occasion_OCCASION_UNSPECIFIED {
			continue
		}
		if id := fam.At.Of(e).SittingID; id != 0 {
			judgeSat[id] = true
		}
	}
	j.JudgeSittings = len(judgeSat)
	// MotionsOf reads the acts that stand, so a ruling struck in its sitting is not counted.
	for _, m := range record.MotionsOf(fam.Events, fam.At) {
		switch m.Subject {
		case "docket":
			if m.Ruled() {
				j.Rulings[m.Ruling]++
			}
		case "grade":
			j.Disputes.Raised++
			switch m.Ruling {
			case "accepted":
				j.Disputes.Accepted++
			case "rejected":
				j.Disputes.Rejected++
			}
		}
	}
	// The epoch an open gap has lived to is the record's current one: the last with work in it.
	// A chair that has just sat opens an epoch nothing has happened in yet, and a gap has not
	// lived through it.
	current := fam.At.CurrentEpoch(fam.Events)
	for _, ep := range record.DebateJSONOfEvents(fam.Events, fam.At).Epochs {
		if ep.Verdict != "" {
			j.LatestVerdict, j.VerdictEpoch = strings.ToUpper(ep.Verdict), ep.Epoch
		}
	}

	// A gap lives from its mint epoch to its close epoch, or to the current epoch while open. Its
	// first mass is the grades it was minted at; its last is the grades it holds now.
	type life struct {
		first, last         int
		firstMass, lastMass float64
	}
	lives := map[string]life{}
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
	var gapOrder []string
	for _, g := range fam.Gaps {
		mint := g.Mint
		if mint == nil {
			continue
		}
		last := current
		if g.HasClosed {
			last = g.ClosedEpoch
		}
		lives[g.ID] = life{
			first: g.Epoch, last: max(last, g.Epoch),
			firstMass: record.GapMass(recordpb.Word(mint.GetLikelihood()), recordpb.Word(mint.GetImpact())),
			lastMass:  record.GapMass(recordpb.Word(g.Likelihood), recordpb.Word(g.Impact)),
		}
		gapOrder = append(gapOrder, g.ID)
		// UNION THE ROOTS: a gap superseding several ancestors joins all their chains into one.
		for _, anc := range mint.GetSupersedes() {
			if ra, rg := find(anc), find(g.ID); ra != rg {
				parent[ra] = rg
			}
		}
	}
	type chain struct {
		first, last         int
		firstMass, lastMass float64
	}
	chains := map[string]*chain{}
	for _, id := range gapOrder {
		e := lives[id]
		root := find(id)
		c := chains[root]
		if c == nil {
			c = &chain{first: e.first, last: e.last, firstMass: e.firstMass, lastMass: e.lastMass}
			chains[root] = c
		}
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

// buildSteps segments the progress bar by the EPOCHS THE RECORD HAS SEEN — one step per chair
// sitting so far, plus the one in progress. The run ends when nobody is ready or at its epoch
// limit, a term setup records (plans/roundless.md §III.B.2); the bar counts what the record has
// seen, not the limit.
func buildSteps(seats []Seat) []Step {
	maxEpoch := 0
	for _, s := range seats {
		if s.Epoch > maxEpoch {
			maxEpoch = s.Epoch
		}
	}
	if maxEpoch == 0 {
		maxEpoch = 1
	}
	// A step is one EPOCH — one dispatch cycle of the chair — so a seat belongs to step r when the
	// record put its register in epoch r. Under the dispatch loop a seat's r-th sitting and epoch r
	// were the same number; under dispatch a lens may sit in epoch 5 for the third time, and the
	// segment it lights up is the fifth. epoch 0 means "any", for the bookends outside the cycle.
	seen := func(seat string, epoch int) bool {
		for _, s := range seats {
			if s.Seat == seat && (epoch == 0 || s.Epoch == epoch) {
				return true
			}
		}
		return false
	}
	doneSeat := func(seat string, epoch int) bool {
		for _, s := range seats {
			if s.Seat == seat && (epoch == 0 || s.Epoch == epoch) && s.Done {
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
	for r := 1; r <= maxEpoch; r++ {
		epochDone := doneSeat("blue-respond", r)
		anySeen := seen("red-lens", r) || seen("red-chair", r) || seen("blue-respond", r) || seen("judge", r)
		steps = append(steps, Step{"epoch " + itoa(r), state(epochDone, anySeen)})
	}
	steps = append(steps, Step{"assembly", state(doneSeat("assemble", 0), seen("assemble", 0))})
	return steps
}

func buildRates(telemetry []*recordpb.TelemetryLine) []Rate {
	rates := make([]Rate, 0, len(telemetry))
	for i, t := range telemetry {
		prevOpen := 0.0
		if i > 0 {
			prevOpen = float64(telemetry[i-1].GetOpenCount())
		}
		minted := float64(t.GetNewMint().GetCount())
		open := float64(t.GetOpenCount())
		closed := prevOpen + minted - open
		closeRate := 0
		if minted+prevOpen > 0 {
			closeRate = int(round(100 * closed / (prevOpen + minted)))
		}
		rates = append(rates, Rate{Epoch: t.Epoch, Opened: int(minted), Closed: int(closed), Open: t.OpenCount, CloseRate: closeRate})
	}
	return rates
}
