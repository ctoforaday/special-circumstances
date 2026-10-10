package difftest

import "time"

// A SCRIPT NAMES AN ID BY THE PLACEHOLDER ITS GOLDEN SHOWS. Ids are random, so `--id GAP001` is the
// first gap id the scenario's output printed, `MOTION002` the second motion id, and the harness
// puts the real id back before the command runs (nonceMapper.resolve). Numbering is by first
// APPEARANCE, per kind.
//
// noSuchGap is the gap id that names nothing: the shape of a gap id, which no mint produces in
// practice. It prints as written and takes no ordinal, and a placeholder no output has printed yet
// reaches the tool as the such id of its kind.
const noSuchGap = "G-00000000"

// The scenarios replay the oracle suite's behaviours as CLI command lists — the
// "29 oracle tests' scenarios exported as replayable command lists" the R2g plan
// calls for. Each names the oracle test it stands in for.

const registry = `{"classes":[{"slug":"propagation-incomplete","material_default":"by_grade"},{"slug":"citation-drift","material_default":"by_grade"},{"slug":"scope-creep","material_default":"by_grade"}]}`

// hostile prose: the quoting recurrence class, plus the characters that expose
// Go's default HTML escaping (<, >, &) — a divergence that would otherwise show
// up only as a byte difference in ordinary seat prose.
const hostile = "quotes \" ' `backtick` $(subshell) ${var}\nnewline\ttab\nangle <brackets> & ampersand\nunicode — em-dash · ✓ 日本語\nbackslash \\ and \\n literal\n"

func scenarios() []scenario {
	base := func(verb string, args ...string) cmd { return cmd{verb: verb, args: args} }

	return []scenario{
		{
			name: "register_pointer_and_seq", // oracle: roundOf/register; per-shard monotonic seq
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("register", "--run", "{RUN}", "--seat-id", "blue-respond"),
				base("position", "--run", "{RUN}", "--seat-id", "blue-respond", "--reason", "first pass"),
				base("log", "--run", "{RUN}", "--seat-id", "blue-respond", "--reason", "no PDF extraction"),
				// implicit register: a seat that never registered still records
				base("finding", "--run", "{RUN}", "--seat-id", "red-lens-logic", "--key", "F1",
					"--severity", "medium", "--likelihood", "high", "--impact", "medium", "--quote", "A claim sits under S2.", "--reason", "unfounded leap"),
			},
		},
		{
			name: "mint_validation", // oracle: acceptance_check required; dangling supersedes; unknown grade
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "scope-creep", "--problem", "no check given"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--check-kind", "document", "--check", "run it", "--problem", "no class given"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "scope-creep", "--check-kind", "document", "--check", "run it",
					"--severity", "catastrophic", "--problem", "bad grade"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "scope-creep", "--check-kind", "document", "--check", "run it",
					"--severity", "low", "--likelihood", "low", "--impact", "low", "--supersedes", noSuchGap, "--problem", "dangling lineage"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "scope-creep", "--check-kind", "document", "--check", "grep the sites",
					"--severity", "high", "--likelihood", "high", "--impact", "medium", "--complexity", "low", "--problem", "a real one"),
			},
		},
		{
			name: "class_registry", // oracle: unknown class refused with hint; --new extends
			seed: map[string]string{"records/class-registry.json": registry},
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "invented-class", "--check-kind", "document", "--check", "x", "--problem", "p"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "attestation-inflation", "--check-kind", "document", "--check", "x", "--problem", "p"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "attestation-inflation", "--check-kind", "document", "--check", "x", "--problem", "p"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "attestation-inflation", "--check-kind", "document", "--check", "compare anchors", "--severity", "medium", "--likelihood", "medium", "--impact", "high", "--problem", "inflation"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "attestation-inflation",
					"--check-kind", "document", "--check", "same class again", "--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "extension accepted"),
			},
		},
		{
			name: "close_validation_and_archive", // oracle: anchor OR carried-from; regression demands successor
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "citation-drift", "--check-kind", "document", "--check", "refetch",
					"--severity", "high", "--likelihood", "high", "--impact", "high", "--complexity", "medium", "--problem", "source moved"),
				base("close", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "GAP001"),
				base("close", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", noSuchGap, "--verified-by", "L1",
					"--verified-with", "git show", "--verified-against", "7bc501e:x", "--reason", "checked"),
				base("close", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "GAP001", "--as", "repaired_with_regression",
					"--verified-by", "L1", "--verified-with", "git show", "--verified-against", "7bc501e:x"),
				base("close", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "GAP001", "--as", "repaired",
					"--verified-by", "L1", "--verified-with", "WebFetch", "--verified-against", "https://example.invalid/spec#s3",
					"--reason", "refetched; the source now resolves and supports the claim"),
			},
		},
		{
			name: "carried_from_renders_as_carried", // oracle: E0.5a inflation becomes unphraseable
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "scope-creep", "--check-kind", "document", "--check", "reread",
					"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "carried case"),
				base("carry", "--run", "{RUN}", "--seat-id", "red-chair", "--id", "GAP001", "--carried-from", "1"),
			},
		},
		{
			name: "multi_nonce_terminal_event_wins", // oracle: the 8/50 duplicate-dispatch anomaly
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "scope-creep", "--check-kind", "document", "--check", "a",
					"--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "from the stale dispatch"),
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"), // re-dispatch rotates the nonce
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "scope-creep", "--check-kind", "document", "--check", "b",
					"--severity", "high", "--likelihood", "high", "--problem", "from the live dispatch",
					"--impact", "high", "--distinct-from", "GAP001"), // a second gap, and the mint's screen is told so
				base("verdict", "--run", "{RUN}", "--seat-id", "red-chair", "--as", "FAIL"),
			},
		},
		{
			name: "multi_nonce_mtime_fallback", // oracle: no terminal event -> latest mtime, explicitly
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("finding", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--key", "F1",
					"--severity", "low", "--likelihood", "low", "--impact", "low", "--quote", "A claim sits under S2.", "--reason", "older shard"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				{
					// `finding`, not `lens` + "finding": the role-prefixed spelling exited 2 with
					// `no command named "lens"`, and the golden recorded that refusal as this
					// scenario's second finding — the newer shard this case is named for was never written.
					verb: "finding",
					args: []string{"--run", "{RUN}", "--seat-id", "red-lens-evidence", "--key", "F2",
						"--severity", "low", "--likelihood", "low", "--impact", "low", "--quote", "A claim sits under S2.", "--reason", "newer shard"},
					mtimes: map[string]time.Time{
						"events-red-lens-evidence-NONCE001.jsonl": time.Unix(1_700_000_000, 0),
						"events-red-lens-evidence-NONCE002.jsonl": time.Unix(1_700_000_600, 0),
					},
				},
			},
		},
		{
			// A finding's id is the tool's, and a seat's --key is what makes a retry safe: a second key
			// is a second finding with its own id, and the first key again answers with the first id and
			// writes nothing.
			name: "finding_key_retry_returns_the_same_finding",
			cmds: []cmd{
				base("finding", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--key", "F1",
					"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--quote", "A claim sits under S2.", "--reason", "round one"),
				base("finding", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--key", "F2", // the seat's SECOND finding, in a later sitting: its own key, its own id
					"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--quote", "A claim sits under S2.", "--reason", "round two"),
				base("finding", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--key", "F1", // the retry: FINDING001 again, and no third event
					"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--quote", "A claim sits under S2.", "--reason", "round one"),
			},
		},
		{
			name: "regrade_history_is_recoverable", // oracle: E0.5b unauditability case
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "scope-creep", "--check-kind", "document", "--check", "a",
					"--severity", "high", "--likelihood", "high", "--impact", "high", "--complexity", "high", "--problem", "graded high at mint"),
				base("regrade", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "GAP001", "--severity", "medium"),
				base("regrade", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "GAP001", "--severity", "medium",
					"--likelihood", "low", "--reason", "blue narrowed the scope; consequence shrank"),
				base("regrade", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "GAP001", "--impact", "low",
					"--reason", "second movement, same id"),
			},
		},
		{
			// oracle: an act stands as filed, and a repeat the record refuses names the act that
			// answers the first — each followed here by that act, at the time its row states. The
			// named cases: a manifest row (the next sitting's), a ruling (the ruler's docket
			// motion), an appeal (a new docket motion), a closure (a docket motion on the closed
			// gap) and a position (the next sitting's).
			name: "a_repeated_act_names_its_answer",
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("register", "--run", "{RUN}", "--seat-id", "blue-respond"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "scope-creep", "--check-kind", "document", "--check", "a",
					"--severity", "high", "--likelihood", "high", "--impact", "high", "--problem", "the first gap"),
				base("manifest-row", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "GAP001", "--reason", "G1 is reproducible via "),
				base("manifest-row", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "GAP001", "--reason", "G1 is reproducible via the recorded proof"),
				base("position", "--run", "{RUN}", "--seat-id", "blue-respond", "--reason", "the board is  going in"),
				base("position", "--run", "{RUN}", "--seat-id", "blue-respond", "--reason", "the board is clean going in"),
				base("motion", "grade", "file", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "GAP001", "--dimension", "severity",
					"--proposed", "low", "--reason", "the consequence is bounded"),
				base("motion", "grade", "rule", "--run", "{RUN}", "--seat-id", "red-chair", "--id", "MOTION001", "--as", "rejected", "--reason", "the evidence does not  it"),
				base("motion", "grade", "rule", "--run", "{RUN}", "--seat-id", "red-chair", "--id", "MOTION001", "--as", "rejected", "--reason", "the evidence does not reach it"),
				base("motion", "docket", "file", "--run", "{RUN}", "--seat-id", "red-chair", "--id", "GAP001", "--reason", "my ruling lost a word: the evidence does not reach it"),
				base("motion", "grade", "appeal", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "MOTION001", "--reason", "pressing it on  grounds"),
				base("motion", "grade", "appeal", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "MOTION001", "--reason", "pressing it on new grounds"),
				base("motion", "docket", "file", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "GAP001", "--reason", "my appeal lost a word: I press it on new grounds"),
				base("close", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "GAP001", "--verified-by", "L1", "--verified-with", "Read",
					"--verified-against", "report.md#S2", "--reason", "verified at the  leaf"),
				base("close", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "GAP001", "--verified-by", "L1", "--verified-with", "Read",
					"--verified-against", "report.md#S2", "--reason", "verified at the leaf"),
				base("motion", "docket", "file", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "GAP001", "--reason", "my closure lost a word: verified at the leaf"),
				base("register", "--run", "{RUN}", "--seat-id", "blue-respond"),
				base("manifest-row", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "GAP001", "--reason", "G1 is reproducible via the recorded proof"),
				base("position", "--run", "{RUN}", "--seat-id", "blue-respond", "--reason", "the board is clean going in"),
			},
		},
		{
			name: "mint_idempotency_on_crash_retry", // oracle: --key returns the EXISTING id
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--key", "L5-F3", "--class", "scope-creep",
					"--check-kind", "document", "--check", "x", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "minted once"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--key", "L5-F3", "--class", "scope-creep",
					"--check-kind", "document", "--check", "x", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "minted once"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--key", "L6-F1", "--class", "scope-creep",
					"--check-kind", "document", "--check", "y", "--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "a different key mints"),
			},
		},
		{
			// oracle: the quoting recurrence class. The name is historical — the prose went through
			// --reason-file when this was written, and the golden is keyed on the name. It now goes
			// through --reason, the one prose spelling, as argv (no shell between harness and tool).
			name: "hostile_prose_via_file",
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("position", "--run", "{RUN}", "--seat-id", "red-chair", "--reason", hostile),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "scope-creep", "--check-kind", "document", "--check", "x",
					"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--reason", hostile),
				base("finding", "--run", "{RUN}", "--seat-id", "red-lens-logic", "--key", "F1",
					"--severity", "low", "--likelihood", "low", "--impact", "low", "--quote", "A claim sits under S2.", "--reason", hostile),
			},
		},
		{
			name: "projections_debate_changelog_citations", // oracle: R2 projections
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("position", "--run", "{RUN}", "--seat-id", "red-chair", "--reason", "red's round position"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "scope-creep", "--check-kind", "document", "--check", "x",
					"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "docketed"),
				base("closing", "--run", "{RUN}", "--seat-id", "red-chair", "--id", "GAP001", "--reason", "red's closing"),
				base("position", "--run", "{RUN}", "--seat-id", "blue-respond", "--reason", "blue's round position"),
				base("closing", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "GAP001", "--reason", "blue's closing"),
				base("manifest-row", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "GAP001", "--reason", "figures recomputed; check run: pass"),
				base("motion", "grade", "file", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "GAP001",
					"--dimension", "likelihood", "--proposed", "low", "--reason", "the harm needs two failures"),
				base("verify", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--quote", "throughput doubled",
					"--title", "https://example.invalid/paper", "--trust", "high", "--access-date", "2026-07-18"),
				base("register", "--run", "{RUN}", "--seat-id", "judge", "--occasion", "docket"),
				// THE BENCH'S DISPOSITION IS TWO COMMANDS: red docksets the gap it cannot settle,
				// the bench rules on that filing. The differential drives both because the
				// projection under test reads the JOIN — the gap is on the filing and the
				// disposition on the ruling.
				base("motion", "docket", "file", "--run", "{RUN}", "--seat-id", "red-chair", "--id", "GAP001",
					"--reason", "contested, and not red's to close"),
				base("motion", "docket", "rule", "--run", "{RUN}", "--seat-id", "judge", "--id", "MOTION002", "--as", "remanded",
					"--principle", "correctness over economy", "--tension", "thoroughness vs cost",
					"--review-flag", "the figure was never recomputed", "--settled", "the proposition this ruling bars", "--reopens-on", "the figure recomputed from its source", "--reason", "the rationale body"),
				base("certify", "--run", "{RUN}", "--seat-id", "judge", "--reason", "what a human should re-examine"),
			},
		},
		{
			name: "bench_petitions_and_halt", // oracle: W2c verbs, filed and ruled as motions
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("motion", "petition", "file", "--run", "{RUN}", "--seat-id", "red-chair", "--class", "safety",
					"--reason", "the design erodes a consent gate", "--relief", "halt and escalate"),
				base("register", "--run", "{RUN}", "--seat-id", "judge", "--occasion", "docket"),
				base("motion", "petition", "rule", "--run", "{RUN}", "--seat-id", "judge", "--id", "MOTION001",
					"--as", "granted", "--binds", "both", "--reason", "the relief binds the coming seats"),
				base("halt", "--run", "{RUN}", "--seat-id", "judge", "--reason", "continuing would compromise the consent gate"),
			},
		},
		{
			name: "docket_ruling_requires_each_unconditional_field", // oracle: reasons, not just fates
			// A REAL GAP AND A REAL FILING FIRST, or the refusal measured is the dangling
			// reference rather than the missing field. `bench opinion` needed only the gap; the
			// bench's ruling answers a MOTION, so both halves have to be on the record before the
			// field rules are what answers.
			//
			// `--reason` is supplied on both attempts for the same reason: cobra refuses the
			// reason flag-group at PARSE, before the record sees the body at all, so an
			// invocation without it pins the parser's message and never reaches the contract this
			// scenario is named for. The four DocketRuling fields are refused at the RECORD write,
			// with the schema's reason (#1234): the first attempt omits --review-flag and
			// --settled, the second --tension, the third --principle, each pinning the refusal
			// for the first field it lacks; the fourth supplies all four, so the reopens-on/final
			// refusal answers. The fifth states --tension as "", which the write refuses as a
			// question skipped, and the sixth remands with --final, which the write refuses
			// because a remand owes the direction its --reopens-on states.
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "scope-creep",
					"--check-kind", "document", "--check", "x", "--severity", "medium",
					"--likelihood", "medium", "--impact", "medium", "--problem", "docketed"),
				base("motion", "docket", "file", "--run", "{RUN}", "--seat-id", "red-chair", "--id", "GAP001",
					"--reason", "contested, and not red's to close"),
				base("register", "--run", "{RUN}", "--seat-id", "judge", "--occasion", "docket"),
				base("motion", "docket", "rule", "--run", "{RUN}", "--seat-id", "judge", "--id", "MOTION001",
					"--as", "repaired", "--reason", "r", "--principle", "p", "--tension", "t"),
				base("motion", "docket", "rule", "--run", "{RUN}", "--seat-id", "judge", "--id", "MOTION001",
					"--as", "repaired", "--reason", "r", "--principle", "p", "--review-flag", "none", "--settled", "s"),
				base("motion", "docket", "rule", "--run", "{RUN}", "--seat-id", "judge", "--id", "MOTION001",
					"--as", "repaired", "--reason", "r", "--tension", "t", "--review-flag", "none", "--settled", "s"),
				base("motion", "docket", "rule", "--run", "{RUN}", "--seat-id", "judge", "--id", "MOTION001",
					"--as", "repaired", "--reason", "r", "--principle", "p", "--tension", "t", "--review-flag", "none", "--settled", "s"),
				base("motion", "docket", "rule", "--run", "{RUN}", "--seat-id", "judge", "--id", "MOTION001",
					"--as", "repaired", "--reason", "r", "--principle", "p", "--tension", "", "--review-flag", "none", "--settled", "s", "--final"),
				base("motion", "docket", "rule", "--run", "{RUN}", "--seat-id", "judge", "--id", "MOTION001",
					"--as", "remanded", "--reason", "r", "--principle", "p", "--tension", "t", "--review-flag", "none", "--settled", "s", "--final"),
			},
		},
		{
			name: "role_boundaries_and_help_contracts", // oracle: the drift test
			cmds: []cmd{
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "x"),
				base("mint", "--run", "{RUN}", "--seat-id", "blue-respond", "--class", "x"),
				base("close", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", noSuchGap),
				base("mint", "--run", "{RUN}", "--seat-id", "judge", "--class", "x"),
				// Each seat's own tree, which is what a boundary IS now. These rows read
				// `lens help` … `bench help` and pinned four copies of the `--seat-id IS REQUIRED`
				// refusal: the role groups are gone, so no row here showed any seat's verbs.
				base("help", "--seat-id", "red-lens-evidence"),
				base("help", "--seat-id", "red-chair"),
				base("help", "--seat-id", "blue-respond"),
				base("help", "--seat-id", "judge"),
			},
		},
		{
			// THE SELF-EXEC PATH, through the built binary: every page below is this binary run again
			// with `<path> --help`. internal/cli's tests cannot reach it — their os.Executable() is the
			// test binary — so this golden is where the real path is pinned. The bench's surface and
			// blue's: the smallest seat tree, and the one the prompts send a seat to read first.
			// RENDERED FROM THE OPERATOR SURFACE, because `manual` is not on a seat's: the seat is
			// handed what it prints. --for names whose surface to walk, which is how agentgen writes
			// a seat's constitution, so this golden pins the path the generator actually takes.
			name: "manual_runs_every_help_page_live",
			cmds: []cmd{
				base("manual", "--seat-id", "operator", "--for", "judge"),
				base("manual", "--seat-id", "operator", "--for", "blue-respond"),
				// AND THE OPERATOR'S OWN, which nothing pinned: the two rows above walk OTHER seats'
				// surfaces THROUGH this one, so every operator page — the diagnostics, the migration,
				// the dashboards — was help text no golden read. The rule that a surface change is
				// reviewed line by line has nothing to review for the surface the operator uses.
				base("manual", "--seat-id", "operator"),
			},
		},
		{
			name: "missing_required_flags", // oracle: --run and --seat-id are refused, not defaulted
			cmds: []cmd{
				base("mint", "--seat-id", "red-lens-evidence"),
				base("mint", "--run", "{RUN}"),
				// An unknown verb on a real seat's tree. This was `merge not-a-verb`, which refused
				// the dead role word `merge` and never reached the verb it names.
				base("not-a-verb", "--run", "{RUN}", "--seat-id", "red-chair"),
			},
		},
		{
			name: "ids_stay_distinct_across_rounds", // a gap minted after the chair sits again is a third id, and the chair's spot-check and blue's motion name the earlier ones
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "a", "--check-kind", "document", "--check", "x",
					"--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "r1 first"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "a", "--check-kind", "document", "--check", "x",
					"--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "r1 second"),
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "a", "--check-kind", "document", "--check", "x",
					"--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "r2 first"),
				base("spot-check", "--run", "{RUN}", "--seat-id", "red-chair", "--ids", "GAP001, GAP002", "--reason", "both re-read"),
				base("register", "--run", "{RUN}", "--seat-id", "blue-respond"),
				base("motion", "grade", "file", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "GAP001",
					"--dimension", "likelihood", "--proposed", "high", "--reason", "the second failure is not required"),
				base("motion", "grade", "rule", "--run", "{RUN}", "--seat-id", "red-chair", "--id", "MOTION001", "--as", "rejected",
					"--reason", "the consequence stands"),
			},
		},
		{
			name: "full_round_integration", // oracle: the integration test
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("verify", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--quote", "claim one",
					"--title", "https://example.invalid/a", "--trust", "high", "--access-date", "2026-07-18"),
				base("finding", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--key", "F1",
					"--severity", "medium", "--likelihood", "medium", "--impact", "high", "--quote", "A claim sits under S2.", "--reason", "citation does not support"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-logic"),
				base("finding", "--run", "{RUN}", "--seat-id", "red-lens-logic", "--key", "F1",
					"--severity", "high", "--likelihood", "high", "--impact", "high", "--quote", "A claim sits under S4.", "--reason", "a leap of faith"),
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "citation-drift", "--check-kind", "document", "--check", "refetch and diff",
					"--severity", "high", "--likelihood", "high", "--impact", "high", "--complexity", "medium",
					"--quote", "A claim sits under S2.", "--found-by", "FINDING001,FINDING002", "--problem", "the cited source does not say this"),
				base("position", "--run", "{RUN}", "--seat-id", "red-chair", "--reason", "round one: FAIL"),
				base("verdict", "--run", "{RUN}", "--seat-id", "red-chair", "--as", "FAIL"),
			},
		},
	}
}
