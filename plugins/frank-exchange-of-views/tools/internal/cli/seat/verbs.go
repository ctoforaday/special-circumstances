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
// Each role still supplies its OWN help text, because the contract is shared
// while the guidance is not: a lens is told what its friction entries are for,
// the bench is told the same verb from the other side.

func Register(help string) *cobra.Command {
	return New("register", help, func(s Context, _ *cobra.Command) (Result, error) {
		nonce, _, err := record.RegisterSeat(s.Identity())
		if err != nil {
			return nil, err
		}
		return registerResult{SeatID: s.SeatID, Nonce: nonce}, nil
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
func Friction(help string) *cobra.Command {
	c := Prose(New("friction", help, func(s Context, cmd *cobra.Command) (Result, error) {
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

func Position(help string) *cobra.Command {
	return Prose(New("position", help, func(s Context, cmd *cobra.Command) (Result, error) {
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

func Closing(help string) *cobra.Command {
	c := Prose(New("closing", help, func(s Context, cmd *cobra.Command) (Result, error) {
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
	name, desc, defaultFor string
}{
	// EVERY DESCRIPTION NAMES THE VERB THAT FILLS THE VIEW, and that is a contract
	// (viewnaming_test.go), not a convention. A seat navigates by what the tool PRINTS: measured
	// on a probe, one read `--view lines-of-inquiry` and then typed `blue line-of-inquiry`, a verb
	// that does not exist, because nothing in the projection it had just read said `line of inquiry`. It
	// found the right verb by failing twice. The next seat may instead conclude the capability is
	// missing and write prose, which loses it for the whole run and is reported nowhere.
	// THE ARTIFACT THE WHOLE DEBATE IS ABOUT, and the last thing a seat still opened by hand.
	// The event record was moved out of reach so `show` became the only way to the board;
	// report.md stayed behind as the one file a seat had to know the layout to find.
	// --anchor IS NAMED HERE, IN THE GROUP LISTING, and that placement is the measurement rather
	// than a preference. Across 18 elicitation sittings (2026-08-17, two naming arms), seats
	// invoked `show board` 26 times and `show report` 7, and ran `show report --help` ZERO times:
	// the bare form succeeds, so nothing ever sends a seat one level deeper. A flag on a
	// subcommand whose bare form works is not a discoverable capability. `changes` already names
	// its --id the same way, for the same reason.
	{"report", "THE ARTIFACT UNDER AUDIT — blue's living report, read THROUGH the tool instead of off disk; add --anchor <id> to read just the passage AT one anchor (with its section and line numbers) rather than the whole document. Anchors are shown AS THEY ARE: `blue edit` refuses an edit that drops one, so a token inside the span you are replacing is yours to carry into --new. TO LOOK ONE UP rather than carry it: `show findings` resolves `<!--fx:f-…-->`, `show evidence` resolves `<!--cite:c-…-->` and `<!--proof:p-…-->`. Written by the round-0 synthesis and every `blue edit`", ""},
	{"board", "THE BOARD — open and closed gaps with grades, closures, anchors, observations and their fates, counts, and any replay anomalies. STRUCTURED JSON by default (the form a seat acts on); `--format markdown` gives the human-verification rendering, open gaps then the closure archive with its prose. Written by `mint`, `close`, `regrade` and `retire`", ""},
	{"findings", "STRUCTURED JSON: every lens finding on the record (label, seat, round, role, grades, location, text) — the merge coalesces these into gaps; replaces the red/candidates/*.md files", ""},
	{"worklist", "STRUCTURED JSON: YOUR PENDING WORK and whether this sitting is finished (`sitting.complete`, with every outstanding duty and the verb that discharges it), plus the shrinking working set — OPEN gaps only (grades, class, location, a problem synopsis, found_by) plus a prose-free closed_index (id, location, class); the once-per-turn read the merge acts on. `merge show` defaults here. Written by `mint` and `close`", "*"},
	{"motions", "STRUCTURED JSON: every motion and its answer — id, subject, filer, the BASIS (the ask in the filer's words), and the ruling if it has one. An unruled motion blocks `merge verdict --as PASS`, and this is the only way to read what it asks. Written by `motion <subject> file`, `rule` and `appeal`", ""},
	{"debate", "the round-by-round transcript, every seat's sections in order (add --json for the STRUCTURED form: rounds with red/blue/lead sections as data, for the audits). Written by `position`, `closing` and `opinion`", ""},
	{"changes", "every recorded edit to blue/report.md (the blue_edit diff stack), in round order; add --id <gap> to put red's required_fix and the edits answering it SIDE BY SIDE — the comparison that replaces inferring whether a gap was fixed. Written by `edit`", ""},
	{"evidence", "STRUCTURED JSON: WHAT BACKS THE REPORT, AND WHAT HAS BEEN CHECKED OF IT — every source keyed by the `<!--cite:c-…-->` anchor in the text (url, title, sha256, the sentence it backs), every computation keyed by its `<!--proof:p-…-->` anchor WITH the sha256 `reproduce --id` wants and red's re-run (or null, meaning nobody re-ran it), and red's verified claims with their trust grades. THIS IS HOW YOU RESOLVE AN ANCHOR you are reading in the report. Written by `cite`, `prove`, `verify` and `reproduce`", ""},
	{"lines-of-inquiry", "the exploration space: lines taken, deferred, declined and abandoned, and the ones still undecided. Written by `line-of-inquiry` (propose and move) and `motion inquiry rule` (red's ruling)", ""},
	{"telemetry", "STRUCTURED JSONL, one line per round: open count, max severity, mass under the pinned mapping, new mints BY SEVERITY AND BY CLASS with the class repeat rate, repair-regression ratio, and edge deltas — the trend the STOPPING judgment reads. The bench's signal for whether the findings are still changing character or merely recurring", ""},
}

// ViewNames is the projection vocabulary — the single source behind the help text, the
// unknown-view error, and (exported for this reason) the gate that asserts every `--view`
// an agent-facing surface NAMES actually exists. See cli.ViewNames.
// ViewMenu is the view list WITH ITS SEMANTICS, one per line.
//
// THE DESCRIPTIONS WERE DEAD TEXT UNTIL THIS EXISTED. The `views` table carried a written line for
// every projection and the `desc` field was read NOWHERE — `--help` printed
// `board|findings|worklist|…`, a bare list of nouns, and every seat that ever asked what a view
// was got names with no meanings.
//
// Measured on a probe: a haiku seat read `--view lines-of-inquiry`, had no way to learn which verb
// writes into it, and invented `blue line-of-inquiry` — a verb that does not exist. It found
// `line of inquiry` by failing twice. The next seat may instead conclude the capability is missing and
// write prose, which loses it for the run and is reported nowhere.
//
// A field declared and never read is the shape this suite keeps finding: it reads as documented
// while documenting nothing.
func ViewMenu() string {
	var b strings.Builder
	for _, v := range views {
		fmt.Fprintf(&b, "  %-16s %s\n", v.name, v.desc)
	}
	return b.String()
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
		Use:          "show",
		Short:        "read a projection on STDOUT (the tool is the read path; the .md files are for human verification). With no projection named, you get YOUR PENDING WORK: " + strings.Join(ViewNames(), " | "),
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		// Its bare form is the seat's pending work — a capability, not a refusal.
		Annotations: map[string]string{"bare-is-a-capability": "yes"},
	}
	// An unknown projection gets the SURFACE, not a parse error. Cobra's default answers
	// `unknown command "x" for "feov-record blue show"` and stops — at the one moment a seat is
	// definitively looking for what exists, which is the same argument that made the root and
	// the role groups teach.
	c.Args = cobra.ArbitraryArgs
	c.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return RefuseAndTeach(cmd, fmt.Sprintf(
				"no projection named %q. The ones below are the whole set, and each names the verb that fills it.", args[0]))
		}
		// The seat's own default — its pending work and whether the sitting is finished.
		return renderView(cmd, "")
	}
	for _, v := range views {
		v := v
		sub := &cobra.Command{
			Use:          v.name,
			Short:        v.desc,
			Args:         cobra.NoArgs,
			SilenceUsage: true,
			RunE:         func(cmd *cobra.Command, _ []string) error { return renderView(cmd, v.name) },
		}
		if v.name == "changes" {
			sub.Flags().String(flags.ID, "", "scope to one gap — put red's required_fix and the edits answering it side by side")
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
		case "board", "findings", "friction", "motions", "worklist", "telemetry", "evidence":
			return fmt.Errorf("%s show: show %s is already JSON by name — drop --json (it is the single way to that projection's JSON)", role, want)
		case "":
			return fmt.Errorf("%s show: name a projection. They are:\n\n%s\nEach names the verb that fills it", role, ViewMenu())
		default:
			return fmt.Errorf("%s show: show %s has no --json form (only 'debate' does; board/findings/friction/motions/worklist are JSON by name)", role, want)
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
	if want == "friction" {
		b, err := record.FrictionJSONBytes(runDir)
		if err != nil {
			return err
		}
		cmd.OutOrStdout().Write(b)
		return nil
	}
	// worklist is JSON by name too — the merge ACTS on it (scans the open set, screens
	// candidates), so it reads structured fields, not prose. It is the shrinking
	// once-per-turn read that the full board JSON is not.
	if want == "worklist" {
		b, err := record.WorklistJSONBytes(runDir, role, Of(cmd).SeatID)
		if err != nil {
			return err
		}
		cmd.OutOrStdout().Write(b)
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
		return fmt.Errorf("%s show: name a projection. They are:\n\n%s\nEach names the verb that fills it", role, ViewMenu())
	}
	known := false
	for _, v := range views {
		if v.name == want {
			known = true
		}
	}
	if !known {
		// THE MENU, NOT THE NAMES. A seat that mistypes a view is a seat that does not
		// know the view space, and a list of twelve nouns tells it which words are legal
		// while leaving it to guess which one holds what it wants.
		return fmt.Errorf("%s show: unknown view %q. The projections are:\n\n%s\nEach names the verb that fills it", role, want, ViewMenu())
	}

	// --id SCOPES a view that supports scoping, and is an ERROR on one that does not
	// ([[one-way-no-aliases]]: a wrong guess fails loudly rather than being ignored).
	// Silently dropping it is the worse failure here — a seat that asked for one gap's
	// edits and received every edit would read the answer as the answer to its question.
	scope := Str(cmd, flags.ID)
	if scope != "" && want != "changes" {
		return fmt.Errorf("%s show: --id scopes `show changes` and nothing else; `show %s` has no scoped form, and answering it unscoped would hand you a different question's answer", role, want)
	}

	b, err := view.Markdown(runDir, want, scope)
	if err != nil {
		return err
	}
	cmd.OutOrStdout().Write(b)
	return nil
}

func Role(role, short string, verbs ...*cobra.Command) *cobra.Command {
	c := &cobra.Command{
		Use:   role,
		Short: short,
		Long:  short + "\n" + MotionFooter + "\n" + FrictionFooter,
		// ArbitraryArgs with flag parsing off at the ROLE level so an unrecognised
		// verb reaches RunE instead of failing on a flag the role does not own:
		// `lens mint --run x` must answer "the lens has no mint verb", not
		// "unknown flag: --run". Subcommand resolution happens before flag
		// parsing, so each verb still parses and validates its own flags strictly.
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		SilenceUsage:       true,
	}
	names := make([]string, 0, len(verbs)+1)
	for _, v := range verbs {
		// Applied HERE rather than in each verb: a verb that had to remember to mark its
		// own required flags is a verb that can forget, and the forgetting is silent —
		// the help simply looks like everything is optional. Every role passes through
		// this loop, so every verb is annotated by construction.
		markRequired(v, v.Name())
		names = append(names, v.Name())
		c.AddCommand(v)
	}
	c.AddCommand(Show())
	names = append(names, "show")

	c.RunE = func(cmd *cobra.Command, args []string) error {
		for _, a := range args {
			if a == "--help" || a == "-h" || a == "help" {
				return cmd.Help()
			}
		}
		// A FLAG IS NOT A VERB. DisableFlagParsing means `blue --run x` arrives here with
		// "--run" as args[0], and the first draft answered `verb "--run" is outside this seat's
		// role` — which is false twice over: it is not a verb, and it is not out of role. A seat
		// told that looks for a permissions problem that does not exist.
		if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
			return RefuseUnknownVerb(cmd, role, args[0])
		}
		if len(args) > 0 {
			return RefuseAndTeach(cmd, fmt.Sprintf(
				"%s: %q is a flag, not a verb — you named a role and its flags with no verb between them, so nothing was recorded. A verb is required; pick one below and pass the flags to it.", role, args[0]))
		}
		// A role invoked with no verb is a usage error, not a no-op: silently
		// succeeding would let a mis-scripted seat believe it recorded something.
		return RequireVerb(cmd, role)
	}
	return c
}

func join(names []string) string {
	return strings.Join(names, ", ")
}

// registerResult and closingResult are the shared-verb results: these verbs are built by
// seat for every role, so their result types live beside them.
type registerResult struct {
	SeatID string `json:"seat_id"`
	Nonce  string `json:"nonce"`
}

func (r registerResult) Human() string {
	return "registered " + r.SeatID + " (shard nonce " + r.Nonce + ")"
}

type closingResult struct {
	ID string `json:"id"`
}

func (r closingResult) Human() string { return "closing filed for " + r.ID }
