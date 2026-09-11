package difftest

import "time"

// The scenarios replay the oracle suite's behaviours as CLI command lists — the
// "29 oracle tests' scenarios exported as replayable command lists" the R2g plan
// calls for. Each names the oracle test it stands in for.

const registry = `{"classes":[{"slug":"propagation-incomplete"},{"slug":"citation-drift"},{"slug":"scope-creep"}]}`

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
				base("revision", "--run", "{RUN}", "--seat-id", "blue-respond", "--reason", "first pass"),
				base("log", "--run", "{RUN}", "--seat-id", "blue-respond", "--reason", "no PDF extraction"),
				// implicit register: a seat that never registered still records
				base("finding", "--run", "{RUN}", "--seat-id", "red-lens-logic", "--key", "F1",
					"--severity", "medium", "--likelihood", "high", "--impact", "medium", "--quote", "## S2", "--reason", "unfounded leap"),
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
					"--supersedes", "G1", "--problem", "dangling lineage"),
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
				base("close", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "G1"),
				base("close", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "G2", "--verified-by", "L1",
					"--verified-with", "git show", "--verified-against", "7bc501e:x"),
				base("close", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "G1", "--as", "repaired_with_regression",
					"--verified-by", "L1", "--verified-with", "git show", "--verified-against", "7bc501e:x"),
				base("close", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "G1", "--as", "repaired",
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
				base("carry", "--run", "{RUN}", "--seat-id", "red-chair", "--id", "G1", "--carried-from", "1"),
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
					"--severity", "high", "--likelihood", "high", "--impact", "high", "--problem", "from the live dispatch"),
				base("verdict", "--run", "{RUN}", "--seat-id", "red-chair", "--as", "FAIL"),
			},
		},
		{
			name: "multi_nonce_mtime_fallback", // oracle: no terminal event -> latest mtime, explicitly
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("finding", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--key", "F1",
					"--severity", "low", "--likelihood", "low", "--impact", "low", "--quote", "## S2", "--reason", "older shard"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				{
					// `finding`, not `lens` + "finding": the role-prefixed spelling exited 2 with
					// `no command named "lens"`, and the golden recorded that refusal as this
					// scenario's second finding — the newer shard this case is named for was never written.
					verb: "finding",
					args: []string{"--run", "{RUN}", "--seat-id", "red-lens-evidence", "--key", "F2",
						"--severity", "low", "--likelihood", "low", "--impact", "low", "--quote", "## S2", "--reason", "newer shard"},
					mtimes: map[string]time.Time{
						"events-red-lens-evidence-NONCE001.jsonl": time.Unix(1_700_000_000, 0),
						"events-red-lens-evidence-NONCE002.jsonl": time.Unix(1_700_000_600, 0),
					},
				},
			},
		},
		{
			name: "finding_labels_run_unique_per_role_across_rounds", // oracle: the tool assigns <area>-F{N}, the sequence spanning rounds — a lens cannot reuse a label, so round two gets evidence-F2, not another evidence-F1
			cmds: []cmd{
				base("finding", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--key", "F1",
					"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--quote", "## S2", "--reason", "round one"),
				base("finding", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--key", "F2", // the seat's SECOND finding, in a later sitting: its own key, the next label
					"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--quote", "## S2", "--reason", "round two"),
			},
		},
		{
			name: "regrade_history_is_recoverable", // oracle: E0.5b unauditability case
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "scope-creep", "--check-kind", "document", "--check", "a",
					"--severity", "high", "--likelihood", "high", "--impact", "high", "--complexity", "high", "--problem", "graded high at mint"),
				base("regrade", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "G1", "--severity", "medium"),
				base("regrade", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "G1", "--severity", "medium",
					"--likelihood", "low", "--reason", "blue narrowed the scope; consequence shrank"),
				base("regrade", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "G1", "--impact", "low",
					"--reason", "second movement, same id"),
			},
		},
		{
			// oracle: same-sitting correction, accepted, through each command's own handler
			// (plans/same-sitting-correction.md III.C.4). Every act is followed at once by its
			// correction, before any other seat acts; the keys are the ones each success line prints.
			// The named cases: a retry that writes nothing, a proposal corrected after its move
			// (keeps Q1), a ruling and its appeal, a closure of a gap its own target closed, a
			// petition ruling, and an outcome.
			name: "correction_accepted_per_command",
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("register", "--run", "{RUN}", "--seat-id", "blue-respond"),
				base("register", "--run", "{RUN}", "--seat-id", "judge"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "scope-creep", "--check-kind", "document", "--check", "a",
					"--severity", "high", "--likelihood", "high", "--impact", "high", "--problem", "the first gap"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "scope-creep", "--check-kind", "document", "--check", "b",
					"--severity", "high", "--likelihood", "high", "--impact", "high", "--problem", "the second gap"),
				base("manifest-row", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "G1", "--reason", "G1 is reproducible via "),
				base("manifest-row", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "G1", "--reason", "G1 is reproducible via the recorded proof",
					"--corrects", "blue-respond:manifest_row:#1:G1", "--correction-why", "the row lost its method"),
				base("manifest-row", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "G1", "--reason", "G1 is reproducible via the recorded proof",
					"--corrects", "blue-respond:manifest_row:#1:G1", "--correction-why", "the row lost its method"),
				base("line-of-inquiry", "propose", "--run", "{RUN}", "--seat-id", "blue-respond", "--reason", "try the  method"),
				base("line-of-inquiry", "move", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "Q1", "--as", "pursued", "--reason", "the method held"),
				base("line-of-inquiry", "propose", "--run", "{RUN}", "--seat-id", "blue-respond", "--reason", "try the recorded method",
					"--corrects", "blue-respond:avenue:#1", "--correction-why", "a word was lost"),
				base("motion", "grade", "file", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "G2", "--dimension", "severity",
					"--proposed", "low", "--reason", "the consequence is bounded"),
				base("motion", "grade", "rule", "--run", "{RUN}", "--seat-id", "red-chair", "--id", "M1", "--as", "rejected", "--reason", "the evidence does not  it"),
				base("motion", "grade", "rule", "--run", "{RUN}", "--seat-id", "red-chair", "--id", "M1", "--as", "rejected", "--reason", "the evidence does not reach it",
					"--corrects", "red-chair:motion_rule:#1", "--correction-why", "a word was lost"),
				base("motion", "grade", "appeal", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "M1", "--reason", "pressing it on  grounds"),
				base("motion", "grade", "appeal", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "M1", "--reason", "pressing it on new grounds",
					"--corrects", "blue-respond:motion_appeal:#1", "--correction-why", "a word was lost"),
				base("close", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "G1", "--verified-by", "L1", "--verified-with", "Read",
					"--verified-against", "report.md#S2", "--reason", "verified at the  leaf"),
				base("close", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--id", "G1", "--verified-by", "L1", "--verified-with", "Read",
					"--verified-against", "report.md#S2", "--reason", "verified at the leaf",
					"--corrects", "red-lens-evidence:close:#1:G1", "--correction-why", "a word was lost"),
				base("motion", "petition", "file", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "safety",
					"--relief", "halt before the next round", "--reason", "a consent gate is missing"),
				base("motion", "petition", "rule", "--run", "{RUN}", "--seat-id", "judge", "--id", "M2", "--as", "denied", "--reason", "the gate  exists"),
				base("motion", "petition", "rule", "--run", "{RUN}", "--seat-id", "judge", "--id", "M2", "--as", "denied", "--reason", "the gate already exists",
					"--corrects", "judge:motion_rule:#1", "--correction-why", "a word was lost"),
				base("outcome", "--run", "{RUN}", "--seat-id", "judge", "--as", "UNVERIFIED", "--reason", "it stopped because  refused"),
				base("outcome", "--run", "{RUN}", "--seat-id", "judge", "--as", "UNVERIFIED", "--reason", "it stopped because the gate refused",
					"--corrects", "judge:outcome:#1", "--correction-why", "a word was lost"),
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
					"--severity", "low", "--likelihood", "low", "--impact", "low", "--quote", "## S2", "--reason", hostile),
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
				base("closing", "--run", "{RUN}", "--seat-id", "red-chair", "--id", "G1", "--reason", "red's closing"),
				base("position", "--run", "{RUN}", "--seat-id", "blue-respond", "--reason", "blue's round position"),
				base("closing", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "G1", "--reason", "blue's closing"),
				base("revision", "--run", "{RUN}", "--seat-id", "blue-respond", "--reason", "repairs landed"),
				base("manifest-row", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "G1", "--reason", "figures recomputed; check run: pass"),
				base("motion", "grade", "file", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "G1",
					"--dimension", "likelihood", "--proposed", "low", "--reason", "the harm needs two failures"),
				base("verify", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--quote", "throughput doubled",
					"--title", "https://example.invalid/paper", "--trust", "high", "--access-date", "2026-07-18"),
				base("register", "--run", "{RUN}", "--seat-id", "judge"),
				// THE BENCH'S DISPOSITION IS TWO COMMANDS: red docksets the gap it cannot settle,
				// the bench rules on that filing. The differential drives both because the
				// projection under test reads the JOIN — the gap is on the filing and the
				// disposition on the ruling.
				base("motion", "docket", "file", "--run", "{RUN}", "--seat-id", "red-chair", "--id", "G1",
					"--reason", "contested, and not red's to close"),
				base("motion", "docket", "rule", "--run", "{RUN}", "--seat-id", "judge", "--id", "M2", "--as", "remanded",
					"--principle", "correctness over economy", "--tension", "thoroughness vs cost",
					"--review-flag", "the figure was never recomputed", "--settled", "the proposition this ruling bars", "--final", "--reason", "the rationale body"),
				base("certify", "--run", "{RUN}", "--seat-id", "judge", "--reason", "what a human should re-examine"),
			},
		},
		{
			name: "bench_petitions_and_halt", // oracle: W2c verbs, filed and ruled as motions
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("motion", "petition", "file", "--run", "{RUN}", "--seat-id", "red-chair", "--class", "safety",
					"--reason", "the design erodes a consent gate", "--relief", "halt and escalate"),
				base("register", "--run", "{RUN}", "--seat-id", "judge-petition-red-chair"),
				base("motion", "petition", "rule", "--run", "{RUN}", "--seat-id", "judge-petition-red-chair", "--id", "M1",
					"--as", "granted", "--binds", "both", "--reason", "the relief binds the coming seats"),
				base("halt", "--run", "{RUN}", "--seat-id", "judge-petition-red-chair", "--reason", "continuing would compromise the consent gate"),
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
			// scenario is named for.
			cmds: []cmd{
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-evidence"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "scope-creep",
					"--check-kind", "document", "--check", "x", "--severity", "medium",
					"--likelihood", "medium", "--impact", "medium", "--problem", "docketed"),
				base("motion", "docket", "file", "--run", "{RUN}", "--seat-id", "red-chair", "--id", "G1",
					"--reason", "contested, and not red's to close"),
				base("register", "--run", "{RUN}", "--seat-id", "judge"),
				base("motion", "docket", "rule", "--run", "{RUN}", "--seat-id", "judge", "--id", "M1",
					"--as", "repaired", "--reason", "r"),
				base("motion", "docket", "rule", "--run", "{RUN}", "--seat-id", "judge", "--id", "M1",
					"--as", "repaired", "--reason", "r", "--principle", "p", "--tension", "t"),
			},
		},
		{
			name: "role_boundaries_and_help_contracts", // oracle: the drift test
			cmds: []cmd{
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "x"),
				base("mint", "--run", "{RUN}", "--seat-id", "blue-respond", "--class", "x"),
				base("close", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "G1"),
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
			name: "manual_runs_every_help_page_live",
			cmds: []cmd{
				base("manual", "--seat-id", "judge"),
				base("manual", "--seat-id", "blue-respond"),
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
			name: "sequential_ids_across_rounds", // oracle: gap ids are run-global, and so are motion ids
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
				base("spot-check", "--run", "{RUN}", "--seat-id", "red-chair", "--ids", "G1, G2", "--reason", "both re-read"),
				base("register", "--run", "{RUN}", "--seat-id", "blue-respond"),
				base("motion", "grade", "file", "--run", "{RUN}", "--seat-id", "blue-respond", "--id", "G1",
					"--dimension", "likelihood", "--proposed", "high", "--reason", "the second failure is not required"),
				base("motion", "grade", "rule", "--run", "{RUN}", "--seat-id", "red-chair", "--id", "M1", "--as", "rejected",
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
					"--severity", "medium", "--likelihood", "medium", "--impact", "high", "--quote", "## S2", "--reason", "citation does not support"),
				base("register", "--run", "{RUN}", "--seat-id", "red-lens-logic"),
				base("finding", "--run", "{RUN}", "--seat-id", "red-lens-logic", "--key", "F1",
					"--severity", "high", "--likelihood", "high", "--impact", "high", "--quote", "## S4", "--reason", "a leap of faith"),
				base("register", "--run", "{RUN}", "--seat-id", "red-chair"),
				base("mint", "--run", "{RUN}", "--seat-id", "red-lens-evidence", "--class", "citation-drift", "--check-kind", "document", "--check", "refetch and diff",
					"--severity", "high", "--likelihood", "high", "--impact", "high", "--complexity", "medium",
					"--quote", "## S2", "--found-by", "evidence-F1,logic-F1", "--problem", "the cited source does not say this"),
				base("position", "--run", "{RUN}", "--seat-id", "red-chair", "--reason", "round one: FAIL"),
				base("verdict", "--run", "{RUN}", "--seat-id", "red-chair", "--as", "FAIL"),
			},
		},
	}
}
