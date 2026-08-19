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
	base := func(role string, args ...string) cmd { return cmd{role: role, args: args} }

	return []scenario{
		{
			name: "register_pointer_and_seq", // oracle: roundOf/register; per-shard monotonic seq
			cmds: []cmd{
				base("merge", "register", "--run", "{RUN}", "--seat-id", "red-merge-r1"),
				base("blue", "register", "--run", "{RUN}", "--seat-id", "blue-respond-r2"),
				base("blue", "revision", "--run", "{RUN}", "--seat-id", "blue-respond-r2", "--reason", "first pass"),
				base("blue", "friction", "--run", "{RUN}", "--seat-id", "blue-respond-r2", "--reason", "no PDF extraction"),
				// implicit register: a seat that never registered still records
				base("lens", "finding", "--run", "{RUN}", "--seat-id", "red-lens-r1-L5", "--key", "F1",
					"--severity", "medium", "--likelihood", "high", "--impact", "medium", "--quote", "## S2", "--reason", "unfounded leap"),
			},
		},
		{
			name: "mint_validation", // oracle: acceptance_check required; dangling supersedes; unknown grade
			cmds: []cmd{
				base("merge", "register", "--run", "{RUN}", "--seat-id", "red-merge-r1"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "scope-creep", "--problem", "no check given"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--check-kind", "document", "--check", "run it", "--problem", "no class given"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "scope-creep", "--check-kind", "document", "--check", "run it",
					"--severity", "catastrophic", "--problem", "bad grade"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "scope-creep", "--check-kind", "document", "--check", "run it",
					"--supersedes", "R9-9", "--problem", "dangling lineage"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "scope-creep", "--check-kind", "document", "--check", "grep the sites",
					"--severity", "high", "--likelihood", "high", "--impact", "medium", "--complexity", "low", "--problem", "a real one"),
			},
		},
		{
			name: "class_registry", // oracle: unknown class refused with hint; --new extends
			seed: map[string]string{"records/class-registry.json": registry},
			cmds: []cmd{
				base("merge", "register", "--run", "{RUN}", "--seat-id", "red-merge-r1"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "invented-class", "--check-kind", "document", "--check", "x", "--problem", "p"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "attestation-inflation", "--check-kind", "document", "--check", "x", "--problem", "p"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "attestation-inflation", "--check-kind", "document", "--check", "x", "--problem", "p"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "attestation-inflation", "--check-kind", "document", "--check", "compare anchors", "--severity", "medium", "--likelihood", "medium", "--impact", "high", "--problem", "inflation"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "attestation-inflation",
					"--check-kind", "document", "--check", "same class again", "--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "extension accepted"),
			},
		},
		{
			name: "close_validation_and_archive", // oracle: anchor OR carried-from; regression demands successor
			cmds: []cmd{
				base("merge", "register", "--run", "{RUN}", "--seat-id", "red-merge-r1"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "citation-drift", "--check-kind", "document", "--check", "refetch",
					"--severity", "high", "--likelihood", "high", "--impact", "high", "--complexity", "medium", "--problem", "source moved"),
				base("merge", "close", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--id", "R1-1"),
				base("merge", "close", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--id", "R9-9", "--verified-by", "L1",
					"--verified-with", "git show", "--verified-against", "7bc501e:x"),
				base("merge", "close", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--id", "R1-1", "--as", "closed_with_regression",
					"--verified-by", "L1", "--verified-with", "git show", "--verified-against", "7bc501e:x"),
				base("merge", "close", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--id", "R1-1", "--as", "closed",
					"--verified-by", "L1", "--verified-with", "WebFetch", "--verified-against", "https://example.invalid/spec#s3",
					"--reason", "refetched; the source now resolves and supports the claim"),
			},
		},
		{
			name: "carried_from_renders_as_carried", // oracle: E0.5a inflation becomes unphraseable
			cmds: []cmd{
				base("merge", "register", "--run", "{RUN}", "--seat-id", "red-merge-r2"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r2", "--class", "scope-creep", "--check-kind", "document", "--check", "reread",
					"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "carried case"),
				base("merge", "carry", "--run", "{RUN}", "--seat-id", "red-merge-r2", "--id", "R2-1", "--carried-from", "1"),
			},
		},
		{
			name: "multi_nonce_terminal_event_wins", // oracle: the 8/50 duplicate-dispatch anomaly
			cmds: []cmd{
				base("merge", "register", "--run", "{RUN}", "--seat-id", "red-merge-r1"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "scope-creep", "--check-kind", "document", "--check", "a",
					"--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "from the stale dispatch"),
				base("merge", "register", "--run", "{RUN}", "--seat-id", "red-merge-r1"), // re-dispatch rotates the nonce
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "scope-creep", "--check-kind", "document", "--check", "b",
					"--severity", "high", "--likelihood", "high", "--impact", "high", "--problem", "from the live dispatch"),
				base("merge", "verdict", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--as", "FAIL"),
			},
		},
		{
			name: "multi_nonce_mtime_fallback", // oracle: no terminal event -> latest mtime, explicitly
			cmds: []cmd{
				base("lens", "register", "--run", "{RUN}", "--seat-id", "red-lens-r1-L1"),
				base("lens", "finding", "--run", "{RUN}", "--seat-id", "red-lens-r1-L1", "--key", "F1",
					"--severity", "low", "--likelihood", "low", "--impact", "low", "--quote", "## S2", "--reason", "older shard"),
				base("lens", "register", "--run", "{RUN}", "--seat-id", "red-lens-r1-L1"),
				{
					role: "lens",
					args: []string{"finding", "--run", "{RUN}", "--seat-id", "red-lens-r1-L1", "--key", "F2",
						"--severity", "low", "--likelihood", "low", "--impact", "low", "--quote", "## S2", "--reason", "newer shard"},
					mtimes: map[string]time.Time{
						"events-red-lens-r1-L1-NONCE001.jsonl": time.Unix(1_700_000_000, 0),
						"events-red-lens-r1-L1-NONCE002.jsonl": time.Unix(1_700_000_600, 0),
					},
				},
			},
		},
		{
			name: "finding_labels_run_unique_per_role_across_rounds", // oracle: the tool assigns L{role}-F{N}, the sequence spanning rounds — a lens cannot reuse a label, so round two gets L1-F2, not another L1-F1
			cmds: []cmd{
				base("lens", "finding", "--run", "{RUN}", "--seat-id", "red-lens-r1-L1", "--key", "F1",
					"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--quote", "## S2", "--reason", "round one"),
				base("lens", "finding", "--run", "{RUN}", "--seat-id", "red-lens-r2-L1", "--key", "F1",
					"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--quote", "## S2", "--reason", "round two"),
			},
		},
		{
			name: "regrade_history_is_recoverable", // oracle: E0.5b unauditability case
			cmds: []cmd{
				base("merge", "register", "--run", "{RUN}", "--seat-id", "red-merge-r1"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "scope-creep", "--check-kind", "document", "--check", "a",
					"--severity", "high", "--likelihood", "high", "--impact", "high", "--complexity", "high", "--problem", "graded high at mint"),
				base("merge", "regrade", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--id", "R1-1", "--severity", "medium"),
				base("merge", "regrade", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--id", "R1-1", "--severity", "medium",
					"--likelihood", "low", "--reason", "blue narrowed the scope; consequence shrank"),
				base("merge", "regrade", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--id", "R1-1", "--impact", "low",
					"--reason", "second movement, same id"),
			},
		},
		{
			name: "mint_idempotency_on_crash_retry", // oracle: --key returns the EXISTING id
			cmds: []cmd{
				base("merge", "register", "--run", "{RUN}", "--seat-id", "red-merge-r3"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r3", "--key", "L5-F3", "--class", "scope-creep",
					"--check-kind", "document", "--check", "x", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "minted once"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r3", "--key", "L5-F3", "--class", "scope-creep",
					"--check-kind", "document", "--check", "x", "--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "minted once"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r3", "--key", "L6-F1", "--class", "scope-creep",
					"--check-kind", "document", "--check", "y", "--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "a different key mints"),
			},
		},
		{
			name: "hostile_prose_via_file", // oracle: the quoting recurrence class
			seed: map[string]string{"prose.md": hostile},
			cmds: []cmd{
				base("merge", "register", "--run", "{RUN}", "--seat-id", "red-merge-r1"),
				base("merge", "position", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--reason-file", "{RUN}/prose.md"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "scope-creep", "--check-kind", "document", "--check", "x",
					"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--reason-file", "{RUN}/prose.md"),
				base("lens", "finding", "--run", "{RUN}", "--seat-id", "red-lens-r1-L5", "--key", "F1",
					"--severity", "low", "--likelihood", "low", "--impact", "low", "--quote", "## S2", "--reason-file", "{RUN}/prose.md"),
			},
		},
		{
			name: "projections_debate_changelog_citations", // oracle: R2 projections
			seed: map[string]string{"red.md": "red's round position\n", "blue.md": "blue's round position\n"},
			cmds: []cmd{
				base("merge", "register", "--run", "{RUN}", "--seat-id", "red-merge-r1"),
				base("merge", "position", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--reason-file", "{RUN}/red.md"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "scope-creep", "--check-kind", "document", "--check", "x",
					"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--problem", "docketed"),
				base("merge", "closing", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--id", "R1-1", "--reason", "red's closing"),
				base("blue", "position", "--run", "{RUN}", "--seat-id", "blue-respond-r1", "--reason-file", "{RUN}/blue.md"),
				base("blue", "closing", "--run", "{RUN}", "--seat-id", "blue-respond-r1", "--id", "R1-1", "--reason", "blue's closing"),
				base("blue", "revision", "--run", "{RUN}", "--seat-id", "blue-respond-r1", "--reason", "repairs landed"),
				base("blue", "manifest-row", "--run", "{RUN}", "--seat-id", "blue-respond-r1", "--id", "R1-1", "--reason", "figures recomputed; check run: pass"),
				base("blue", "dispute", "--run", "{RUN}", "--seat-id", "blue-respond-r1", "--id", "R1-1",
					"--dimension", "likelihood", "--proposed", "low", "--reason", "the harm needs two failures"),
				base("blue", "confidence", "--run", "{RUN}", "--seat-id", "blue-respond-r1", "--quote", "C7", "--confidence", "medium"),
				base("lens", "verify", "--run", "{RUN}", "--seat-id", "red-lens-r1-L1", "--quote", "throughput doubled",
					"--title", "https://example.invalid/paper", "--trust", "high", "--access-date", "2026-07-18"),
				base("bench", "register", "--run", "{RUN}", "--seat-id", "judge-r1"),
				base("bench", "opinion", "--run", "{RUN}", "--seat-id", "judge-r1", "--id", "R1-1", "--as", "carried",
					"--principle", "correctness over economy", "--tension", "thoroughness vs cost",
					"--review-flag", "the figure was never recomputed", "--reason", "the rationale body"),
				base("bench", "certify", "--run", "{RUN}", "--seat-id", "judge-r1", "--reason", "what a human should re-examine"),
			},
		},
		{
			name: "bench_petitions_and_halt", // oracle: W2c verbs
			cmds: []cmd{
				base("merge", "petition", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "safety",
					"--reason", "the design erodes a consent gate", "--relief", "halt and escalate"),
				base("bench", "register", "--run", "{RUN}", "--seat-id", "judge-petition"),
				base("bench", "petition-rule", "--run", "{RUN}", "--seat-id", "judge-petition", "--petitioner", "red-merge-r1",
					"--class", "safety", "--as", "granted", "--reason", "the relief binds the coming seats"),
				base("bench", "halt", "--run", "{RUN}", "--seat-id", "judge-petition", "--reason", "continuing would compromise the consent gate"),
			},
		},
		{
			name: "opinion_requires_all_five_fields", // oracle: opinions, not dispositions
			cmds: []cmd{
				base("bench", "register", "--run", "{RUN}", "--seat-id", "judge-r2"),
				base("bench", "opinion", "--run", "{RUN}", "--seat-id", "judge-r2", "--id", "R2-1", "--as", "closed"),
				base("bench", "opinion", "--run", "{RUN}", "--seat-id", "judge-r2", "--id", "R2-1", "--as", "closed",
					"--principle", "p", "--tension", "t"),
			},
		},
		{
			name: "role_boundaries_and_help_contracts", // oracle: the drift test
			cmds: []cmd{
				base("lens", "mint", "--run", "{RUN}", "--seat-id", "red-lens-r1-L1", "--class", "x"),
				base("blue", "mint", "--run", "{RUN}", "--seat-id", "blue-respond-r1", "--class", "x"),
				base("blue", "close", "--run", "{RUN}", "--seat-id", "blue-respond-r1", "--id", "R1-1"),
				base("bench", "mint", "--run", "{RUN}", "--seat-id", "judge-r1", "--class", "x"),
				base("lens", "help"),
				base("merge", "help"),
				base("blue", "help"),
				base("bench", "help"),
			},
		},
		{
			name: "missing_required_flags", // oracle: --run and --seat-id are refused, not defaulted
			cmds: []cmd{
				base("merge", "mint", "--seat-id", "red-merge-r1"),
				base("merge", "mint", "--run", "{RUN}"),
				base("merge", "not-a-verb", "--run", "{RUN}", "--seat-id", "red-merge-r1"),
			},
		},
		{
			name: "sequential_ids_across_rounds", // oracle: mintGapId is sequential per round
			cmds: []cmd{
				base("merge", "register", "--run", "{RUN}", "--seat-id", "red-merge-r1"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "a", "--check-kind", "document", "--check", "x",
					"--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "r1 first"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "a", "--check-kind", "document", "--check", "x",
					"--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "r1 second"),
				base("merge", "register", "--run", "{RUN}", "--seat-id", "red-merge-r2"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r2", "--class", "a", "--check-kind", "document", "--check", "x",
					"--severity", "low", "--likelihood", "low", "--impact", "low", "--problem", "r2 first"),
				base("merge", "spot-check", "--run", "{RUN}", "--seat-id", "red-merge-r2", "--ids", "R1-1, R1-2", "--reason", "both re-read"),
				base("merge", "dispute-respond", "--run", "{RUN}", "--seat-id", "red-merge-r2", "--id", "R1-1", "--as", "rejected",
					"--reason", "the consequence stands"),
			},
		},
		{
			name: "full_round_integration", // oracle: the integration test
			cmds: []cmd{
				base("lens", "register", "--run", "{RUN}", "--seat-id", "red-lens-r1-L1"),
				base("lens", "verify", "--run", "{RUN}", "--seat-id", "red-lens-r1-L1", "--quote", "claim one",
					"--title", "https://example.invalid/a", "--trust", "high", "--access-date", "2026-07-18"),
				base("lens", "finding", "--run", "{RUN}", "--seat-id", "red-lens-r1-L1", "--key", "F1",
					"--severity", "medium", "--likelihood", "medium", "--impact", "high", "--quote", "## S2", "--reason", "citation does not support"),
				base("lens", "register", "--run", "{RUN}", "--seat-id", "red-lens-r1-L5"),
				base("lens", "finding", "--run", "{RUN}", "--seat-id", "red-lens-r1-L5", "--key", "F1",
					"--severity", "high", "--likelihood", "high", "--impact", "high", "--quote", "## S4", "--reason", "a leap of faith"),
				base("merge", "register", "--run", "{RUN}", "--seat-id", "red-merge-r1"),
				base("merge", "mint", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--class", "citation-drift", "--check-kind", "document", "--check", "refetch and diff",
					"--severity", "high", "--likelihood", "high", "--impact", "high", "--complexity", "medium",
					"--quote", "## S2", "--found-by", "L1-F1,L5-F1", "--problem", "the cited source does not say this"),
				base("merge", "position", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--reason", "round one: FAIL"),
				base("merge", "verdict", "--run", "{RUN}", "--seat-id", "red-merge-r1", "--as", "FAIL"),
			},
		},
	}
}
