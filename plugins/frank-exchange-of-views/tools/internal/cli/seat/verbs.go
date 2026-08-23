package seat

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/scorecard"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
)

// The verbs every role shares, defined once.
//
// register, friction, position and closing are the SAME contract
// wherever they appear — a friction entry from a lens and one from the bench are
// the same event with the same payload. Restating them per role would be four
// copies to drift apart, and the drift would be silent because each copy would
// still pass its own tests.
//
// The help lives in help/<key>.md. `register` and `friction` mean the same thing in every chair and
// are ONE document each; `position` and `closing` genuinely differ — a RED section is not a BLUE one
// — and are keyed by role.

func Register() *cobra.Command {
	return New("register", func(s Context, _ *cobra.Command) (Result, error) {
		nonce, _, err := record.RegisterSeat(s.Identity())
		if err != nil {
			return nil, err
		}
		r := registerResult{SeatID: s.SeatID, Nonce: nonce}
		// A SECOND SITTING IS TOLD WHAT ITS FIRST ONE LOST, HERE, BECAUSE THIS IS THE ONLY
		// MOMENT IT IS LISTENING FOR WHO IT IS.
		//
		// Read AFTER the register, deliberately: the rotation this call just performed is what
		// makes the earlier shard a loser, so the discard does not exist until now.
		//
		// A read failure is NOT reported as "nothing was discarded". It leaves the field nil and
		// the seat is told the check could not run — the alternative is a silence that means
		// either "your predecessor lost nothing" or "nobody looked", which is the shape this
		// whole warning exists to break.
		if d, found, derr := record.DiscardedForSeat(s.RunDir, s.SeatID); derr != nil {
			r.PriorSittingUnknown = derr.Error()
		} else if found && len(d.Keys) > 0 {
			r.PriorSitting = &priorSitting{Dispatches: d.Dispatches, DiscardedKeys: d.Keys}
		}
		// THE IDENTITY DID NOT ARRIVE, AND THE SEAT IS THE ONLY PARTY THAT CAN STILL ACT ON IT.
		//
		// register is where the agent->seat binding is written, so it is also the first and only
		// place that can see the binding was never possible. The record already handles this
		// correctly at the write — absent is recorded as absent — and capture already reports it
		// hours later. What nothing did was tell the SEAT, while the sitting it affects is still
		// running.
		//
		// MEASURED, and it is the whole of #512. In research/2026-08-22_is-7-prime all FOURTEEN
		// registers carry no agent_id: the hook never fired for that run at all, so every later
		// call was refused "this agent has not registered" whatever shape it took. Across the
		// session that message was returned 92 times and was FALSE every single time — zero of
		// them were a seat that had genuinely not registered.
		//
		// A WARNING, NOT A REFUSAL. Refusing here would wedge a run whose only fault is a hook
		// that did not fire, and the seat can work perfectly well by passing --seat-id. What it
		// cannot do is guess that, or know that re-registering — the obvious reading of "you have
		// not registered" — rotates its shard and orphans its work.
		if seatenv.AgentID() == "" && s.SeatID != record.OperatorRole {
			r.IdentityAbsent = true
		}
		return r, nil
	})
}

// Friction records a capability or protocol complaint — or, with --none, the explicit statement
// that nothing blocked this sitting.
//
// WHY THE EXPLICIT NEGATIVE EXISTS. Across eighteen probed seat dispatches, not one friction
// event was ever recorded, and "no friction on the record" is equally consistent with a clean
// sitting and with a seat that hit walls and never used the channel. Those are the same bytes,
// which is this project's recurring defect and the reason nobody could tell which had happened.
//
// It is not that seats were unwilling. Most refusals they met were their OWN errors, correctly
// read as such — that is not friction and should not be filed. But one seat, blocked by a motion
// it could not read, reasoned in its own words that no read verb existed, searched ten-plus
// calls, and then GUESSED rather than reporting it: filing costs a turn and does not unblock,
// while guessing might. A seat under that pressure needs the report to be a duty it discharges,
// not an invitation it declines.
//
// The shape is `spot-check --none --reason`, already in this tree for the same reason: a duty
// whose empty case must be ASSERTED rather than inferred from silence.
func Friction() *cobra.Command {
	c := Prose(New("friction", func(s Context, cmd *cobra.Command) (Result, error) {
		none, _ := cmd.Flags().GetBool(flags.None)
		text, err := Reason(cmd)
		if err != nil {
			return nil, err
		}
		if none {
			// An empty discharge that does not say what you looked for is indistinguishable
			// from a skipped duty — the same rule spot-check applies to its own --none.
			if _, err := record.Append(s.Identity(), "friction-none", record.NewPayload().Set("reason", text)); err != nil {
				return nil, err
			}
			return Msg{Message: "recorded: nothing blocked this sitting"}, nil
		}
		if _, err := record.Append(s.Identity(), "friction", record.NewPayload().Set("reason", text)); err != nil {
			return nil, err
		}
		return Msg{Message: "friction recorded"}, nil
	}))
	c.Flags().Bool(flags.None, false,
		"nothing blocked this sitting — the EXPLICIT negative, with --reason saying what you reached for and found. "+
			"Silence cannot say this: an empty friction log reads the same whether the sitting was clean or the channel went unused")
	return c
}

func Position(key string) *cobra.Command {
	return Prose(NewKeyed("position", key, func(s Context, cmd *cobra.Command) (Result, error) {
		text, err := Reason(cmd)
		if err != nil {
			return nil, err
		}
		if _, err := record.Append(s.Identity(), "position", record.NewPayload().Set("reason", text)); err != nil {
			return nil, err
		}
		return Msg{Message: "position recorded"}, nil
	}))
}

func Closing(key string) *cobra.Command {
	c := Prose(NewKeyed("closing", key, func(s Context, cmd *cobra.Command) (Result, error) {
		text, err := Reason(cmd)
		if err != nil {
			return nil, err
		}
		p := Set(cmd, record.NewPayload(), "gap_id", flags.ID)
		p.Set("reason", text)
		if _, err := record.Append(s.Identity(), "closing", p); err != nil {
			return nil, err
		}
		return closingResult{ID: Str(cmd, flags.ID)}, nil
	}))
	c.Flags().Var(flags.GapID().WithCheck(record.GapExists), flags.ID, "the gap id this closing argues")
	return c
}

// views are the projections a seat may read. `defaultFor` is the role whose default this view
// is; "*" means every role.
//
// EVERY SEAT DEFAULTS TO ITS PENDING WORK. It did not: blue's bare `show` returned `changelog`
// — a record of what blue had ALREADY done, handed to it before it had done anything — the lens
// got `citation-ledger` and the bench got `debate`. Asked what would tell them a sitting was
// finished, only the merge could name a mechanism; blue and the bench answered with another
// seat's future act ("red agrees it's sound"), which is not observable at the moment they have
// to decide to stop.
//
// THREE VIEWS ALSO CLAIMED "merge" AND THE LAST ONE SILENTLY WON, because the resolution loop
// keeps overwriting. A default decided by slice order is a default nobody chose.
var views = []struct {
	name, short, long, defaultFor string
}{
	{"report", "THE DOCUMENT UNDER AUDIT, as it stands now — read it through the tool, not off disk. `changes` says how it got that way. Written by the round-0 synthesis and every `edit`, with anchors from `cite`, `finding` and `prove`", "THE ARTIFACT UNDER AUDIT — blue's living report, read THROUGH the tool instead of off disk; add --anchor <id> to read just the passage AT one anchor (with its section and line numbers) rather than the whole document. Anchors are shown AS THEY ARE: `blue edit` refuses an edit that drops one, so a token inside the span you are replacing is yours to carry into --new. TO LOOK ONE UP rather than carry it: `show findings` resolves `<!--fx:f-…-->`, `show evidence` resolves `<!--cite:c-…-->` and `<!--proof:p-…-->`. Written by the round-0 synthesis and every `blue edit`", ""},
	{"board", "EVERY GAP THE RUN HAS, yours or not — open and closed, with grades, fates and closure prose. `work` narrows this to what is yours and blocking. Written by `mint`, `close`, `regrade` and `retire`", "THE BOARD — open and closed gaps with grades, closures, anchors, observations and their fates, counts, and any replay anomalies. STRUCTURED JSON by default (the form a seat acts on); `--format markdown` gives the human-verification rendering, open gaps then the closure archive with its prose. Written by `mint`, `close`, `regrade` and `retire`", ""},
	{"findings", "THE RAW LENS FINDINGS, BEFORE the merge coalesces them into gaps — several findings can become one gap, and this is where you see which. Written by `finding`", "STRUCTURED JSON: every lens finding on the record (label, seat, round, role, grades, location, text) — the merge coalesces these into gaps; replaces the red/candidates/*.md files", ""},
	{"work", "WHAT IS OPEN TO YOU, AND WHETHER YOU MAY STOP — your pending work, not the whole board. Run it first and again before you finish. Written by `mint`, `close` and the bench's `opinion`", "**RUN THIS FIRST AND RUN IT AGAIN BEFORE YOU STOP.** STRUCTURED JSON: EVERYTHING OPEN TO YOU, in one list. `sitting.open` is every work item — each with `blocks`, saying whether it is what is stopping you closing — and `sitting.complete` is true exactly when nothing blocking is left. An item with `blocks: false` is work this board affords you that nobody will refuse you for skipping: a citation nobody verified, a source blue never cited, a proof nobody re-ran, a line of inquiry never revisited, a grade you could move, a motion you could file. IT IS STILL YOUR WORK. `complete: true` with items open means the gates are satisfied, NOT that there is nothing to do. Carries the shrinking working set too — OPEN gaps only (grades, class, location, a problem synopsis, found_by). AND `closed_index`, WHICH IS NOT DEBRIS: it is the ESTOPPEL REGISTER, and this description used to call it a shrinking working set, which reads as `done, ignore it` and is the opposite of what it is for. Each entry carries id, location, class, the `fate` that ended it, and `closed_by` — `bench` or `red`. THAT DISTINCTION IS LOAD-BEARING: red may reopen its OWN closure on new evidence, but a bench ruling is ESTOPPED and re-raising it is relitigation, not diligence. If you hold new evidence against a bench-ruled gap, it is a lineage successor — mint it under a new id naming the ruled gap in `supersedes`, and say what the ruling did not account for. A `fate` of defect_owed_elsewhere means still broken and NOT yours to fix; repaired_with_regression means a live successor exists. The reasoning behind any fate is on the record, not here: `show debate` carries the bench's opinions, `show board --format markdown` the closure archive with its prose. Read the one you are about to rely on or work around. Bare `show` defaults here for every role. Written by `mint`, `close` and the bench's `opinion`", "*"},
	{"motions", "WHAT HAS BEEN CONTESTED AND HOW IT WAS RULED — the ask in the filer's words, and the ruling if it has one. `debate` is what each side ARGUED; this is what was formally disputed. Written by `motion`, `rule` and `appeal`", "STRUCTURED JSON: every motion and its answer — id, subject, filer, the BASIS (the ask in the filer's words), and the ruling if it has one. An unruled motion blocks `merge verdict --as PASS`, and this is the only way to read what it asks. Written by `motion <subject> file`, `rule` and `appeal`", ""},
	{"debate", "WHAT EACH SIDE ARGUED, round by round — the transcript, in order. Written by `position`, `closing` and `opinion`", "the round-by-round transcript, every seat's sections in order (add --json for the STRUCTURED form: rounds with red/blue/lead sections as data, for the audits). Written by `position`, `closing` and `opinion`", ""},
	{"changes", "HOW THE REPORT GOT THAT WAY — every edit in round order, and with `--id <gap>` the fix red asked for beside the edits answering it. Written by `edit`", "every recorded edit to blue/report.md (the blue_edit diff stack), in round order; add --id <gap> to put red's required_fix and the edits answering it SIDE BY SIDE — the comparison that replaces inferring whether a gap was fixed. Written by `edit`", ""},
	{"evidence", "WHAT BACKS A CLAIM, AND WHAT RED MADE OF IT — the lookup table for an anchor you are holding while reading. Written by `cite`, `prove`, `verify` and `reproduce`", "STRUCTURED JSON: WHAT BACKS THE REPORT, AND WHAT HAS BEEN CHECKED OF IT — every source keyed by the `<!--cite:c-…-->` anchor in the text (url, title, sha256, the sentence it backs), every computation keyed by its `<!--proof:p-…-->` anchor WITH the sha256 `reproduce --id` wants and red's re-run (or null, meaning nobody re-ran it), and red's verified claims with their trust grades. THIS IS HOW YOU RESOLVE AN ANCHOR you are reading in the report. Written by `cite`, `prove`, `verify` and `reproduce`", ""},
	{"lines-of-inquiry", "WHICH DIRECTIONS WERE TAKEN AND WHICH WERE NOT — pursued, deferred, declined, abandoned, and the ones still undecided. Written by `line-of-inquiry` (propose and move) and `motion inquiry rule`", "the exploration space: lines taken, deferred, declined and abandoned, and the ones still undecided. Written by `line-of-inquiry` (propose and move) and `motion inquiry rule` (red's ruling)", ""},
	{"telemetry", "HOW THE NUMBERS MOVED ACROSS ROUNDS — a trend, not a snapshot: one line per round, and the signal the STOPPING judgment reads. Computed from the record, so no verb fills it", "STRUCTURED JSONL, one line per round: open count, max severity, mass under the pinned mapping, new mints BY SEVERITY AND BY CLASS with the class repeat rate, repair-regression ratio, and edge deltas — the trend the STOPPING judgment reads. The bench's signal for whether the findings are still changing character or merely recurring", ""},
	{"scorecard", "HOW YOUR CHAIR IS DOING ON THIS QUESTION — your own side's performance, this run only. No selector: your chair is the seat you registered as. Computed from the record, so no verb fills it", "YOUR CHAIR'S IN-RUN SCORECARD — how the side you sit on is doing on THIS question, computed live from this run's record. Your own performance, not another party's: there is no selector, because your chair is the seat you registered as. A number reading badly means RECOGNISE the failure and adapt — never perform the metric at the expense of the duty it measures, because a diagnostic gamed is itself a defect and a detector firing is a finding. Rows reading \"not computed\" are HONEST, not gaps to fill: the envelope-derived rows fill in at capture. Computed from the record, so no verb fills it", ""},
}

// ViewNames is the projection vocabulary — the single source behind the help text, the
// unknown-view error, and (exported for this reason) the gate that asserts every `--view`
// an agent-facing surface NAMES actually exists. See cli.ViewNames.
// showGroup resolves the `show` GROUP from wherever a refusal is raised, so the listing a seat
// gets is the one cobra renders for that group — not a second copy with its own column widths.
//
// A refusal may be raised from the group itself (bare `show`) or from one of its view
// subcommands, and only the group's help carries the projection list.
func showGroup(c *cobra.Command) *cobra.Command {
	if c == nil {
		return c
	}
	if c.Name() != "show" && c.Parent() != nil && c.Parent().Name() == "show" {
		return c.Parent()
	}
	return c
}

func ViewNames() []string {
	out := make([]string, 0, len(views))
	for _, v := range views {
		out = append(out, v.name)
	}
	return out
}

// Show prints a projection to STDOUT.
//
// READS BELONG IN THE TOOL. Every seat could already WRITE through the tool while still
// having to READ the run by opening markdown files at paths it learned from a prompt —
// which is how a seat comes to trust a hand-written artifact over the event log, and how
// the two came to disagree. A seat that asks the tool gets the answer the events support.
//
// It does NOT re-derive the markdown. It renders through the exact path `render` uses and
// then prints the resulting file, so the bytes on stdout and the bytes on disk cannot
// differ. A second renderer would be a second reader of one artifact, which is the defect
// class this whole tool exists to remove — writing one to serve reads would reintroduce it
// at the read surface.
// Show is a GROUP, not a verb with a flag.
//
// It was `show --view <name>`, and cobra models commands and flags — never a flag's VALUE
// space. That is the same undiscoverability the motion collapse fixed one layer up: a value has
// no --help of its own, no completion, and no place to say what it is for, so every projection's
// description had to be crammed into one flag's usage string. Making `--view` optional (a bare
// `show` now answers with the seat's pending work) made it worse rather than better: the flag
// became something a seat had no reason to discover at all.
//
// As subcommands each projection is a first-class thing: `show board`, `show motions`, its own
// --help, its own completion, and an unknown one gets the refusal that lists the whole surface
// rather than a flag-parse error naming the value.
func Show() *cobra.Command {
	c := &cobra.Command{
		Use: "show",
		// THE NAME LIST CAME OFF THIS LINE when the root page started marking which entries hold
		// commands. Ten projection names in a parent's listing is a slice of the surface handed
		// out where the seat cannot yet see what any of them is FOR — and a seat that reads a
		// name here and never opens the page has learned a word, which is how one of them typed a
		// projection name as a verb. The group heading now says this page hides commands; each
		// projection says what it is for on its own page, which is where the choosing happens.
		Short:        "read a projection of the record — the tool is the read path, and the .md files are for human verification. Bare, it answers with YOUR PENDING WORK",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		// Its bare form is the seat's pending work — a capability, not a refusal.
		Annotations: map[string]string{"bare-is-a-capability": "yes"},
	}
	// An unknown projection gets the SURFACE, not a parse error. Cobra's default answers
	// `unknown command "x" for "feov-record blue show"` and stops — at the one moment a seat is
	// definitively looking for what exists, which is the same argument that makes the root teach.
	c.Args = cobra.ArbitraryArgs
	c.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return RefuseAndTeach(cmd, fmt.Sprintf(
				"no projection named %q. The ones below are the whole set, and each names the verb that fills it.", args[0]))
		}
		// The seat's own default — its pending work and whether the sitting is finished.
		return renderView(cmd, "")
	}
	// --id LIVES ON THE GROUP so the refusal below it can fire.
	//
	// It was registered on `changes` alone, which meant `show telemetry --id R1-1` never reached
	// renderView: cobra rejected an unknown flag first and answered `unknown flag: --id`. The
	// refusal written for that case — naming `changes` as the one view --id scopes — was
	// unreachable, and the message a seat actually got did not say where --id does work. A
	// carefully argued error nobody can reach is the same as no error at all.
	c.PersistentFlags().String(flags.ID, "",
		"scope `show changes` to one gap — red's required_fix beside the edits answering it. No other projection has a scoped form")
	for _, v := range views {
		v := v
		sub := &cobra.Command{
			Use: v.name,
			// SHORT IS THE MENU, LONG IS THE INSTRUCTION, and they are separate fields because a
			// seat reads them at two different moments.
			//
			// They were ONE string. Cobra prints Short in the parent's listing and, with no Long
			// set, prints it AGAIN as the whole body of the leaf page — so `show --help` carried
			// all eleven essays (6,615 bytes, the largest page in the system) and every
			// `show <view> --help` beneath it repeated one of them verbatim. A seat that read the
			// group page had already read every leaf page, which is why `pagesNeverUsed` came back
			// at 11 of 13 for a seat that opened its whole tree: the pages beneath held nothing new.
			//
			// SHORT IS WRITTEN TO BE READ AS A SET, not alone. Eleven of them arrive on one page
			// and the seat's question there is which ONE to open, so each says what question it
			// answers and how it differs from its neighbours — `work` narrows `board`, `changes`
			// explains `report`, `motions` is what was disputed where `debate` is what was argued.
			// Each still names the verb that fills it: viewnaming_test asks that of `show --help`,
			// which renders Short, and a projection whose writer is unnamed teaches a name that
			// does not work.
			Short:        v.short,
			Long:         v.long,
			Args:         cobra.NoArgs,
			SilenceUsage: true,
			RunE:         func(cmd *cobra.Command, _ []string) error { return renderView(cmd, v.name) },
		}

		// ONE COMMAND, BOTH FORMS. `ledger` and `archive` were markdown-only views of data the
		// board JSON already carries whole — three names for one projection, which is the alias
		// problem this vocabulary bans everywhere else. --format is the flag `graph` already
		// uses for the same question.
		if v.name == "board" {
			sub.Flags().String(flags.Format, "json", "json (the form a seat acts on) | markdown (the human-verification rendering: open gaps, then the closure archive with its prose)")
		}
		// READING AT AN ANCHOR, rather than pulling the whole document to check one sentence.
		// The window is addressed by anchor because a line number is a fact about a rendering —
		// see internal/anchor/window.go for why that distinction is load-bearing here.
		if v.name == "report" {
			sub.Flags().String(flags.Anchor, "",
				"read the report AT one anchor id (`f-…`, `c-…`, `p-…`) rather than whole — you get the LIVE text there, its section heading, and line numbers to quote back. "+
					"`show findings` resolves finding anchors; `show evidence` resolves citation and proof anchors")
			sub.Flags().Int(flags.Window, anchor.DefaultWindow,
				"with --anchor: how many paragraphs of content either side of it (blank lines are carried, not counted)")
		}
		c.AddCommand(sub)
	}
	return c
}

// renderView writes one projection. want == "" means this role's default.
func renderView(cmd *cobra.Command, want string) error {
	role := roleOf(cmd)
	// RESOLVED, NOT READ OFF THE FLAG. The engine injects FEOV_RUN and every WRITE verb
	// honours it through Begin/Of — reads did not, so a seat that correctly omitted --run
	// could record all round and then be told its board did not exist. Measured with the
	// identity injected: register, friction and revision all succeeded; `show` and
	// `claim-index` demanded the flag.
	runDir := Of(cmd).RunDir
	if runDir == "" {
		return fmt.Errorf("%s: --run <runDir> is required", role)
	}
	// --id SCOPES a view that supports scoping, and is an ERROR on one that does not
	// ([[one-way-no-aliases]]: a wrong guess fails loudly rather than being ignored). Silently
	// dropping it is the worse failure — a seat that asked for one gap's edits and received
	// every edit would read the answer as the answer to its question.
	//
	// CHECKED FIRST, BEFORE ANY VIEW DISPATCH, and that placement is the whole fix. It used to
	// sit at the BOTTOM of this function, after every early-returning handler — so `show
	// telemetry --id R1-1` reached the telemetry writer, printed the unscoped series and exited
	// 0, which is precisely the failure the paragraph above describes. A guard is only a guard
	// at the point the request arrives.
	if scope := Str(cmd, flags.ID); scope != "" && want != "changes" && want != "" {
		return fmt.Errorf("%s show: --id scopes `show changes` and nothing else; `show %s` has no scoped form, and answering it unscoped would hand you a different question's answer", role, want)
	}
	if want == "" {
		// FIRST MATCH WINS, not last. The loop used to keep overwriting, so three views
		// claiming "merge" resolved by slice order — a default nobody chose.
		for _, v := range views {
			if v.defaultFor == role || v.defaultFor == "*" {
				want = v.name
				break
			}
		}
	}

	// --json on a read opts into the STRUCTURED form of a view whose native form is
	// markdown. It is the same inherited flag the mutating verbs use for their JSON
	// envelopes; on reads it had been ignored (reads were view-selected only). It now
	// selects the structured debate — the sole view with both a markdown transcript (for
	// the human) and a JSON form (for the audits that used to regex the sections).
	//
	// One canonical way to each form ([[one-way-no-aliases]]): --json is an ERROR on the
	// views already JSON by name (board/findings/friction — `--view board` is the single
	// way to board JSON, no alias) and on markdown views with no JSON form. It is checked
	// BEFORE the board/findings/friction branches so `--view board --json` refuses rather
	// than falling through to the flagless JSON. A wrong guess fails loudly.
	if asJSON, _ := cmd.Flags().GetBool(flags.JSON); asJSON {
		switch want {
		case "debate":
			b, err := record.DebateJSONBytes(runDir)
			if err != nil {
				return err
			}
			cmd.OutOrStdout().Write(b)
			return nil
		// `friction` WAS NAMED HERE AND IS NOT A PROJECTION. `show friction` is refused —
		// the friction reader is the OPERATOR command (cli/friction.go), never a seat view — so
		// this case could not fire and the default's message below advertised it to seats by
		// name. A seat that read it learned a projection it cannot open.
		case "board", "findings", "motions", "work", "telemetry", "evidence":
			return fmt.Errorf("%s show: show %s is already JSON by name — drop --json (it is the single way to that projection's JSON)", role, want)
		case "":
			return RefuseAndTeach(showGroup(cmd), fmt.Sprintf("%s show: name a projection. Each below names the verb that fills it", role))
		default:
			return fmt.Errorf("%s show: show %s has no --json form (only 'debate' does; board/findings/motions/work/telemetry/evidence are JSON by name)", role, want)
		}
	}

	// The BOARD is served as structured JSON, because a seat ACTS on it.
	//
	// Every other view is prose a seat reads; the board is state a seat has to make
	// decisions against — which gap is open, what its grades are, whether a closure
	// carried its anchor. Parsing that back out of markdown is precisely where the
	// scorecard defects came from (anchored_closures_pct read 0 against an 89
	// baseline because it parsed sentences while the anchors sat in structured
	// fields). Markdown for the same state stays available behind --view ledger.
	//
	// Resolved AFTER the role default, not before: `merge show` with no flags must
	// reach this branch, and checking the raw flag first sent the merge seat's own
	// default view looking for a board.md that no renderer writes.
	if want == "board" {
		// The markdown arm is what `ledger` and `archive` rendered: the open board for a human
		// verification pass, then the closure archive with its prose. Both were separate views
		// of data this JSON already carries whole.
		// AN UNKNOWN --format IS REFUSED, NOT ROUNDED TO JSON.
		//
		// This read `if f == "markdown" || f == "md"` with no else, so every other value —
		// including a typo, and including `dot`, which `graph --format` accepts — fell through to
		// the JSON arm and exited 0. A seat that asked for a rendering it did not get had no way
		// to find out: `--format banana` and `--format json` produced identical bytes.
		//
		// Found 2026-08-16 by widening setInHelp in the enum-help gate. The usage line spells
		// `json (…) | markdown (…)`, which is a closed-set promise, and nothing was keeping it.
		switch f, _ := cmd.Flags().GetString(flags.Format); f {
		case "", "json":
			// The default arm, below.
		case "markdown", "md":
			led, err := view.Markdown(runDir, "ledger", "")
			if err != nil {
				return err
			}
			arc, err := view.Markdown(runDir, "archive", "")
			if err != nil {
				return err
			}
			cmd.OutOrStdout().Write(led)
			cmd.OutOrStdout().Write([]byte("\n"))
			cmd.OutOrStdout().Write(arc)
			return nil
		default:
			return feov.Errorf(feov.Validation, "show board: unknown --format %q (json | markdown) — "+
				"an unrecognised format used to render JSON and exit 0, so a seat could not tell a typo from the default", f)
		}
		// The role and seat are passed so the board CAN carry the sitting; whether it does is
		// the duty arm's decision, and unset means the board is exactly what it always was.
		b, err := record.BoardJSONBytesFor(runDir, role, Of(cmd).SeatID)
		if err != nil {
			return err
		}
		cmd.OutOrStdout().Write(b)
		return nil
	}
	// findings is served as JSON too, and for the same reason: the merge ACTS on it
	// (coalesces findings into gaps), so it reads structured fields, not prose it must
	// parse. This is the channel that replaced red/candidates/*.md.
	if want == "findings" {
		b, err := record.FindingsJSONBytes(runDir)
		if err != nil {
			return err
		}
		cmd.OutOrStdout().Write(b)
		return nil
	}
	// motions is JSON by name: a seat reads it to ANSWER a motion, so it needs the filer's
	// basis as a field rather than prose it must find in a transcript.
	// The artifact under audit, through the tool rather than off disk. Anchors intact: they are
	// what `blue edit` holds a seat responsible for carrying across an edit.
	if want == "report" {
		b, err := report.BlueReportForReading(runDir)
		if err != nil {
			return err
		}
		if a, _ := cmd.Flags().GetString(flags.Anchor); a != "" {
			n, _ := cmd.Flags().GetInt(flags.Window)
			w, err := anchor.ReadAround(string(b), a, n)
			if err != nil {
				return err
			}
			cmd.OutOrStdout().Write([]byte(w.Render()))
			return nil
		}
		// --window WITHOUT --anchor IS REFUSED, not quietly ignored. A seat that asked to
		// narrow its read and got the whole document back cannot tell that from a report which
		// is simply that long — the same shape as `--format banana` rendering JSON and exiting
		// 0, one branch up.
		if cmd.Flags().Changed(flags.Window) {
			return feov.Errorf(feov.Validation, "show report: --window sizes a window and there is no window without --anchor <id>. "+
				"Either name the anchor you are reading at, or drop --window and take the whole report")
		}
		cmd.OutOrStdout().Write(b)
		return nil
	}
	// evidence is JSON by name: it is a LOOKUP TABLE keyed by the anchor token a seat is holding,
	// and a markdown rendering of it would be a table to parse rather than a field to read.
	if want == "evidence" {
		b, err := record.EvidenceJSONBytes(runDir)
		if err != nil {
			return err
		}
		cmd.OutOrStdout().Write(b)
		return nil
	}
	if want == "motions" {
		b, err := record.MotionsJSONBytes(runDir)
		if err != nil {
			return err
		}
		cmd.OutOrStdout().Write(b)
		return nil
	}
	// work is JSON by name too — the seat ACTS on it (scans the open set, screens candidates),
	// so it reads structured fields, not prose. It is the shrinking once-per-turn read that the
	// full board JSON is not, and it is the ONE command a seat is told to run: everything open
	// to it, whether it may close, and which items are what stop it closing.
	if want == "work" {
		b, err := record.WorkJSONBytes(runDir, role, Of(cmd).SeatID)
		if err != nil {
			return err
		}
		cmd.OutOrStdout().Write(b)
		return nil
	}
	// The scorecard resolves its chair from the SEAT, so there is nothing to pass and nothing to
	// get wrong. A role with no chair is refused by name rather than handed an empty card —
	// operator is not a party to the debate and reads chairs explicitly.
	if want == "scorecard" {
		chair, ok := record.ChairOf(role)
		if !ok {
			return fmt.Errorf("%s show: no chair sits for role %q, so there is no scorecard that is yours — "+
				"a scorecard grades a side of the debate, and this role is not one", role, role)
		}
		var board *record.Board
		if b, err := record.BoardState(runDir); err == nil {
			board = b
		}
		rows := scorecard.Compute(runDir, scorecard.ReadResults(runDir), board)[chair]
		fmt.Fprint(cmd.OutOrStdout(), scorecard.RenderChair(chair, rows, "this run")+"\n")
		return nil
	}
	// telemetry is JSONL by name — one line per round, the wire shape the stopping
	// judgment reads. It is a SERIES, not a snapshot, and the series is the whole
	// point: a single round's numbers cannot show a trend changing character.
	if want == "telemetry" {
		b, err := view.TelemetryJSONL(runDir)
		if err != nil {
			return err
		}
		cmd.OutOrStdout().Write(b)
		return nil
	}
	if want == "" {
		return RefuseAndTeach(showGroup(cmd), fmt.Sprintf("%s show: name a projection. Each below names the verb that fills it", role))
	}
	known := false
	for _, v := range views {
		if v.name == want {
			known = true
		}
	}
	if !known {
		// THE MENU, NOT THE NAMES. A seat that mistypes a view is a seat that does not know
		// the view space, and a list of twelve nouns tells it which words are legal while
		// leaving it to guess which one holds what it wants. The menu is COBRA'S — every view
		// is a real subcommand with its own Short, so the framework already renders this list
		// and renders it the same way everywhere.
		return RefuseAndTeach(showGroup(cmd), fmt.Sprintf("%s show: unknown view %q. Each below names the verb that fills it", role, want))
	}

	scope := Str(cmd, flags.ID)

	b, err := view.Markdown(runDir, want, scope)
	if err != nil {
		return err
	}
	cmd.OutOrStdout().Write(b)
	return nil
}

// RoleVerbs is the seat's verb set, ready to mount at the ROOT of that seat's tree.
//
// It used to build a `<role>` command GROUP and the root mounted all four. The role level is gone:
// the seat is bound at `register`, the tool derives the role from that binding, and the root IS the
// seat's surface. A merge seat runs `mint`; a lens seat running `mint` gets an unknown command,
// which is the same boundary the role group drew and is now drawn by the only fact that decides it.
//
// The seat therefore no longer types a word the tool can already look up. ResolveSeat has always
// refused a --seat-id that disagrees with the identity — while the tree required the role DERIVED
// from that identity to be typed correctly, with CheckSeatRole reconciling the two copies. There is
// one copy now, and it is on the record.
func RoleVerbs(role string, verbs ...*cobra.Command) []*cobra.Command {
	out := make([]*cobra.Command, 0, len(verbs)+1)
	for _, v := range verbs {
		// Applied HERE rather than in each verb: a verb that had to remember to mark its own
		// required flags is a verb that can forget, and the forgetting is silent — the help
		// simply looks like everything is optional.
		markTree(v)
		out = append(out, v)
	}
	return append(out, Show())
}

func join(names []string) string {
	return strings.Join(names, ", ")
}

// registerResult and closingResult are the shared-verb results: these verbs are built by
// seat for every role, so their result types live beside them.
type registerResult struct {
	SeatID string `json:"seat_id"`
	Nonce  string `json:"nonce"`
	// PriorSitting is present ONLY when an earlier dispatch of this same seat left work that
	// replay has dropped. Absent is the ordinary case.
	PriorSitting *priorSitting `json:"prior_sitting,omitempty"`
	// PriorSittingUnknown carries the reason the check could not run, so "not measured" never
	// renders as "nothing lost".
	PriorSittingUnknown string `json:"prior_sitting_unknown,omitempty"`
	// IdentityAbsent is true when a DISPATCHED seat registered with no agent id — the hook did
	// not reach this call, so nothing on the record can bind this agent to this seat.
	IdentityAbsent bool `json:"identity_absent,omitempty"`
}

// priorSitting is the discarded work of an earlier dispatch of this seat — the keys, not a
// count, because a `cite` key carries the anchor id and that is the half a seat can act on.
type priorSitting struct {
	Dispatches    int      `json:"dispatches"`
	DiscardedKeys []string `json:"discarded_keys"`
}

func (r registerResult) Human() string {
	out := "registered " + r.SeatID + " (shard nonce " + r.Nonce + ")"
	if r.IdentityAbsent {
		out += "\n\nYOUR IDENTITY DID NOT REACH THE TOOL, so nothing on the record binds this agent " +
			"to this seat. You are registered and your events are recorded; what is missing is the " +
			"binding that lets a LATER call resolve who you are without being told.\n\n" +
			"TWO CONSEQUENCES, AND THE SECOND IS THE EXPENSIVE ONE:\n" +
			"  PASS --seat-id ON EVERY CALL. Without it the tool cannot tell which seat is asking " +
			"and will refuse you.\n" +
			"  DO NOT REGISTER AGAIN. If a later call says \"this agent has not registered\", it is " +
			"THIS, not a lost registration — re-registering rotates your shard nonce and replay " +
			"then discards everything the first sitting recorded."
	}
	if r.PriorSittingUnknown != "" {
		return out + "\n\nWHETHER AN EARLIER SITTING OF THIS SEAT LOST WORK COULD NOT BE CHECKED: " +
			r.PriorSittingUnknown + ". Treat this as NOT MEASURED, not as a clean board."
	}
	if r.PriorSitting == nil {
		return out
	}
	byType := map[string]int{}
	var order []string
	for _, k := range r.PriorSitting.DiscardedKeys {
		t := k
		if parts := strings.Split(k, ":"); len(parts) >= 2 {
			t = parts[1]
		}
		if byType[t] == 0 {
			order = append(order, t)
		}
		byType[t]++
	}
	var kinds []string
	for _, t := range order {
		kinds = append(kinds, fmt.Sprintf("%d %s", byType[t], t))
	}
	// THE GUIDANCE NAMES ONLY THE KINDS ACTUALLY LOST. The first draft listed `cite` and
	// `blue_edit` unconditionally, and the golden corpus caught it immediately: a merge seat that
	// lost one `mint` was being told what a dropped citation means. Advice about work a seat did
	// not lose is noise, and noise in a warning is how the real line stops landing.
	var notes []string
	for _, t := range order {
		if n, ok := discardNotes[t]; ok {
			notes = append(notes, "  "+t+" — "+n)
		}
	}
	guidance := "Re-do what still applies rather than assuming your predecessor's output survived it."
	if len(notes) > 0 {
		guidance = strings.Join(notes, "\n") + "\n\n" + guidance
	}
	return out + fmt.Sprintf(
		"\n\nYOU ARE SITTING %d OF THIS SEAT, AND AN EARLIER SITTING'S WORK IS GONE. %d event(s) it "+
			"recorded (%s) lost replay's winner selection and survive nowhere — not in the board, not "+
			"in the report's evidence, not in any projection you are about to read. This is not a "+
			"warning about the record being wrong; the record is correct and that work is simply "+
			"absent.\n\n%s\n\n%s",
		r.PriorSitting.Dispatches, len(r.PriorSitting.DiscardedKeys),
		strings.Join(kinds, ", "), guidance,
		strings.Join(r.PriorSitting.DiscardedKeys, "\n"))
}

// discardNotes says what losing one event of each kind COSTS, for the kinds a lost sitting can
// plausibly carry. A kind with no note still appears in the count and the key list — the absence
// of advice is not the absence of loss, and inventing a sentence per kind to look complete would
// be worse than saying nothing about the ones whose consequence is simply "it is not there".
var discardNotes = map[string]string{
	"cite":            "a source was fetched and is no longer cited, so the report may make a claim nothing backs",
	"prove":           "a computation was run and its proof anchor resolves to nothing",
	"blue_edit":       "a change to the report has no event behind it, so `show changes` cannot explain how the text got that way",
	"mint":            "a gap was raised and is not on the board — nobody is being asked to fix it",
	"close":           "a closure was recorded and the gap it closed reads as still open",
	"finding":         "a lens finding is gone, so the merge has nothing to coalesce from it",
	"line-of-inquiry": "a direction was opened or disposed and reads as never asked",
	"verify":          "a citation was judged and reads as unchecked",
	"reproduce":       "a proof was re-run and reads as never re-run",
}

type closingResult struct {
	ID string `json:"id"`
}

func (r closingResult) Human() string { return "closing filed for " + r.ID }
