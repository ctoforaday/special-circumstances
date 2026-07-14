# Red pass — Round 1, Lens 1 (leaf-node citation verification), slice 1 of 3

**Slice:** Verdict + §1 (H1 substrate) + §2 (H2 consolidation/dedup) + §3 (competitive landscape).
**Method:** followed each load-bearing citation to its source (web + on-machine); graded
corroboration confidence per statement↔reference pair (high / medium / low). Verified the
on-machine local claims directly; fetched Claude Code docs, both cited GitHub issues, and four
arXiv/blog sources.

**Verdict for this slice: PASS-with-gaps.** The structural spine of §1–§3 is well-sourced —
the strongest recommendation (append-only expansion, §2.3b) rests on a paper that fully
corroborates it, the native-surface mechanics (§1.2) are confirmed by the docs almost verbatim,
and the transcript-substrate claim (§1.4) verifies on this machine. But three citation defects
are real and stay raised until fixed: one striking quantitative claim is **miscited to the wrong
source**, one statistic pair is **uncorroborated after two fetches**, and two GitHub issues the
report calls "open" are in fact **closed as not planned** — one of which carries a blocking-change
plan-dependency that can therefore never be satisfied as written.

---

## Corroborated at leaf node (HIGH confidence) — recorded so grading is graded, not just negative

- **§1.4 transcript substrate.** "transcripts are per-session JSONL at
  `~/.claude/projects/<project-slug>/<session-uuid>.jsonl`" and the schema
  (`uuid`/`parentUuid`, `sessionId`, `cwd`, `gitBranch`, `version`, `isSidechain`, typed records).
  Verified on this machine: 8 JSONL files present; a `user` record carries exactly those keys plus
  a `version` field. **HIGH.** (One nit below on the specific version string.)
- **§1.2 native surfaces**, `[^MemoryDocs]`. Fetched https://code.claude.com/docs/en/memory. All
  of: @-import 4-hop max; code-span/fenced-block skip; imports load at launch and consume context;
  first-external-import one-time approval dialog with silent-disable-on-decline (docs Warning box);
  MEMORY.md 200-line/25KB load; auto memory on by default; `autoMemoryDirectory`;
  `autoMemoryEnabled` / `CLAUDE_CODE_DISABLE_AUTO_MEMORY`; `.claude/rules/` with `paths:`
  frontmatter, symlink support, and "User-level rules are loaded before project rules";
  CLAUDE.md "delivered as a user message after the system prompt." **HIGH — near-verbatim.**
- **§2.3b append-only expansion**, `[^FaultyMemories]` (arXiv 2605.12978). Fetched. Title,
  authors, and findings match: inverted-U utility (rises then declines), corruption via
  interference / meaning-drift / specificity-loss, intensifying with update frequency. The report's
  most important structural fix is soundly grounded. **HIGH.**
- **§2.2 / §6 four consolidation levers + "decay is the lever most systems skip"**,
  `[^ConsolidationProblem]` (Hindsight). Fetched: "Every consolidation policy operates on four
  levers: (1) Importance, (2) Merge, (3) Decay, (4) Eviction"; exponential decay preferred;
  eviction-by-unretrievability; "Decay is the lever most agent memory systems skip. It is also the
  one that matters most." **HIGH** (for these specific claims — see gap G1 for what this footnote
  does NOT support).
- **§1.2 / §1.3 both GitHub issues exist and report what is claimed.** #57507 reports the
  `memory:` field non-functional when a `tools:` allowlist is present (v2.1.137); #56540 reports
  parallel Task fan-out hanging under a non-TTY parent. Substance **HIGH** (status/scope defects
  in G3/G4 below).

---

## Graded gaps (cumulative; each anchored to heading + quoted sentence)

### G1 — MISCITED figure: the §2.1 headline number is not in its cited source [MEDIUM]

- **Location:** §2.1 "The failure is documented, not hypothetical" —
  > "One study storing 2,000 facts and compressing 36.7× found **60% of the knowledge base
  > irretrievably lost**"
  cited to `[^ConsolidationProblem]` (Hindsight blog).
- **Leaf-node result:** The Hindsight page does **not** contain this figure (no "2,000 facts",
  no "36.7×", no "60%"). The number is *real* but originates in a different paper — **"Facts as
  First Class Objects: Knowledge Objects for Persistent LLM Memory," arXiv 2603.17781** ("60% loss
  after 36.7× compression"; same paper reports 54% goal-preservation loss after three cascading
  compactions). A skeptic following the footnote lands on a page without the claim.
- **Also under-corroborated by this footnote:** "summarization drift" as a *named* failure mode
  (the Hindsight page says "Summaries lose entity-level detail," not that phrase), and the
  "OpenClaw 'details unavailable' pattern" attribution.
- **Grade:** likelihood the claim is false = low (figure is genuine); impact = medium (it is the
  lead quantitative evidence for "documented, not hypothetical," and the misattribution is exactly
  the "laundered into fact" failure the protocol names); complexity-to-fix = low.
- **Ask:** re-attribute to arXiv 2603.17781; keep `[^ConsolidationProblem]` for the four-levers /
  decay claims only. Confidence in statement-as-cited: **LOW** until re-attributed.

### G2 — UNCORROBORATED statistics: §2.4's 61.4% / 71.6% not found in the cited paper [MEDIUM]

- **Location:** §2.4 "Review-by-git-diff is a weak sole guard" —
  > "**61.4% of agent-authored pull requests received no recorded review activity at all**, and
  > 71.6% of review comments on them were authored by other agents."
  cited to `[^UnreviewedPRs]` (arXiv 2604.24450).
- **Leaf-node result:** The paper exists with the exact title claimed ("On the Footprints of
  Reviewer Bots' Feedback on Agentic Pull Requests in OSS GitHub Repositories") and is on-topic
  (7,416 reviewer-bot comments on 4,532 agentic PRs). But **two independent fetches** (HTML v1)
  failed to surface "61.38", "71.58", "61.4%", or "71.6%", or any no-review-activity /
  bot-authorship share. Fetch reported only category distributions (e.g., "Bugfix 14.0%").
- **Caveat (fairness to blue):** small-model HTML fetches routinely miss numbers embedded in
  tables/figures, so this is *"unable to corroborate at the leaf node,"* not *"contradicted."* But
  the burden is on the citation to be followable.
- **Grade:** likelihood-miscited = medium; impact = medium (these are the two most concrete numbers
  behind demoting git-diff from preventive to forensic — though §2.4 also stands on
  `[^BotReviewFatigue]` ~54% Dependabot and `[^AIApprovingPRs]`, so the *conclusion* survives even
  if the figures move); complexity-to-fix = low.
- **Ask:** blue re-verify the two figures against the paper PDF and quote the sentence, or relabel
  as approximate / move to a source that carries them. Confidence: **LOW** as cited.

### G3 — MISCHARACTERIZED status: issue #57507 is CLOSED (not planned), not "open" [MEDIUM]

- **Location:** §1.2 agent-`memory:` row —
  > "there is an open bug where the `memory:` field is **non-functional when a tools allowlist is
  > present** (issue #57507) — the row is load-bearing on a currently-flaky feature."
  and §8 change #2 / §9 risk row: "contingent on issue #57507 resolution."
- **Leaf-node result:** #57507 is **Closed as not planned.** Two consequences: (a) "open bug" is
  factually wrong; (b) more materially, a *blocking* change is made "contingent on issue #57507
  resolution" — but a not-planned issue will not be resolved by Anthropic, so the plan dependency
  is unsatisfiable as written. The correct framing is the opposite of blue's: the flakiness is
  *permanent / won't-fix*, with a known workaround (add `Write, Edit` explicitly to `tools:`), and
  the design must own that rather than wait on an upstream fix. This *strengthens* blue's
  substantive caution but the characterization and the plan hook are both wrong.
- **Note:** the issue also documents Subpattern B (memory not written even with full tool access,
  5+ invocations) — a reliability concern broader than "allowlist present"; §1.2 narrows it.
- **Grade:** likelihood = certain (status verified); impact = medium (correctness of a blocking
  change's dependency); complexity-to-fix = low (re-word to "closed won't-fix; apply the explicit-
  tools workaround; do not gate the phase on upstream resolution").

### G4 — SCOPE OVERREACH: issue #56540 is CLOSED and macOS-launchd-specific; operator is on Windows [LOW-MEDIUM]

- **Location:** §1.3 —
  > "there is an open issue where **parallel Task fan-out hangs under non-TTY parents**
  > (cron/scheduled contexts) — precisely the dream loop's runtime."
  and §8 change #9 / §9 risk "Headless hooks/fan-out failures … High in cron context."
- **Leaf-node result:** #56540 is **Closed as not planned**, and its repro is **macOS 25.3.0**
  under `launchctl asuser` / launchd, CLI 2.1.128–2.1.129. The report generalizes to
  "cron/scheduled contexts" and "non-TTY parents" without noting the evidence is macOS-launchd-
  specific. The operator's box is **Windows 11** (Task Scheduler; different IPC/pipe semantics) —
  the cited evidence does not directly bear on the target platform.
- **Fairness:** the mitigation (sequential subagents in the scheduled pass) is cheap, correct, and
  platform-agnostic, so the *design impact* of this gap is low — I am not asking blue to absorb
  complexity here, only to state the evidence's platform scope and stop calling a closed issue
  "open."
- **Grade:** likelihood = certain (status/scope verified); impact = low (mitigation unaffected);
  complexity-to-fix = low. Confidence in claim-as-generalized: **MEDIUM** (phenomenon plausible
  cross-platform, but unverified on Windows).

### G5 — UNATTRIBUTED version number: "v2.1.59" for auto memory [LOW]

- **Location:** §1.2 —
  > "**`MEMORY.md` auto-memory** (native, on by default since v2.1.59)"
  and §3 "Auto memory is native and on by default (v2.1.59+)", cited to `[^MemoryDocs]`.
- **Leaf-node result:** The docs page confirms auto memory is native and on by default but gives
  **no version number**; "v2.1.59" is not in the cited source. On this machine the transcript
  version field reads 2.1.198 (a later build), so the feature's presence is consistent, but the
  *specific* "since v2.1.59" is uncorroborated by the footnote it hangs on.
- **Grade:** likelihood-wrong = low; impact = low (nice-to-have precision, not load-bearing);
  complexity-to-fix = low. Confidence: **LOW** for the exact version; drop it or cite a changelog.

---

## Appropriately hedged (no gap — noted for the record)

- **§3 native "Auto Dream."** `[^AutoDream]`/`[^DreamSkill]` support it as concept + community
  replication; the report explicitly labels it "verified as concept, unverified as a dependable
  API (server-side flag)" and lists it in §10 Unverified. This is correct labeling, not laundering
  — the "two-writer collision" (§3 consequence 2) is a conditional risk ("High if flag lands" in
  §9), which is the honest framing. No gap.
- **§2.1 ARC-AGI 54% figure.** Already labeled secondary/unverified in-text and in §10 via
  `[^AgentsDumber]`. Note only: a distinct "54%" (goal-preservation loss) appears in the
  arXiv 2603.17781 paper surfaced under G1 — watch for number-coincidence confusion if blue
  re-cites, but no action required.

---

## Not covered by this slice (handed to instances 2/3)

§4 (memory poisoning / CVE-2026-21852 / SpAIware / 80–99% attack success), §5 (cadence: Letta
sleep-time, Stanford generative agents, RecMem), §6 (context-rot Chroma, instruction budget,
belief-memory ALFWorld, the two local "does-not-exist" verifications), §7 (alternatives),
§8–§10. Leaf-node verification of those footnotes belongs to the other two lens-1 slices.

## Friction

- Aggregation deferred by design: this invocation scoped me to the candidate file only, so I did
  **not** write to the shared `red/findings.md` or `debate.md` (parallel instances 2/3 would race).
  The lead must union the three lens-1 candidates into `findings.md` and the `### RED` debate entry.
- HTML-arXiv leaf-node verification is lossy: the small-model WebFetch cannot reliably read numbers
  in tables/figures, which is why G2 is graded "uncorroborated" rather than "false." A tool that
  extracts a paper's tables (or full-PDF text search) would let me discharge G2 definitively rather
  than leaving it low-confidence. I flagged rather than silently degraded.
