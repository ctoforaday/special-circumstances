# Red audit round 2 — lens L6 (dark-side/risk)

**Seat:** red-lens-r2-L6 | **Role:** failure modes, likelihood × impact × complexity grading, security and tradeoff blindspots | **Source:** blue/report.md (full re-read 2026-07-19)

---

## Findings

### L6-F1: Store-state volatility undermines the "zero non-empty" universal claim

**Location:** §2 "The reasoning channel" + footnote [^LocalSweep]

**Quoted claim:** "A recursive sweep of `~/.claude/projects/` on 2026-07-18 found 294 transcript files, 278 of which contain thinking blocks; 5,754 thinking blocks carry the exact byte sequence `"thinking":"","signature":"`, and **zero** have a non-empty `thinking` field."

**Risk:** The count "moved from 5,744 to 5,754 between two measurements minutes apart because the store was growing under a live session — the figure is a snapshot of a moving target, not a fixed corpus size." This is a temporal sampling problem. The count is volatile; the headline "zero non-empty" is snapshot-specific. A harness that reads the store at a different time (or after a code change triggers new sessions) could observe a different count. The finding rests on *current state* but is presented as *invariant property*. If the store's growth rate is high, the "zero" could be stale within hours.

**Severity:** Low-medium likelihood (depends on store growth rate, which is not profiled), high impact (the headline rests on this count), medium complexity to mitigate (profile growth rate and set a re-measurement cadence).

**Acceptance check:** Profile `~/.claude/projects` growth rate over 24 hours; confirm the count remains at zero non-empty after N new sessions. If growth is >50 blocks/day, flag that the snapshot is stale by round end.

---

### L6-F2: Display-resolver mechanism inferred from code, not empirically verified

**Location:** §2 "The mechanism, read out of the shipped client" + footnote [^BinaryDisplayResolver]

**Quoted claim:** "Claude Code v2.1.215 resolves the display mode with ... a separate guard forces `display:"omitted"` on any non-interactive session whose display was not set explicitly and which is not using exact-tools, subagent-text forwarding, or async mode."

**Risk:** The mechanism is extracted via string-search and minification reversal from the 256MB binary. The report acknowledges: "Minified identifiers collide (`W1` resolved to an unrelated function on first grep), absence of a string does not prove absence of a behavior (names can be constructed at runtime)." The display-resolver function is read out of code, but no empirical validation confirms that (a) the function is *actually called* at the claim site, (b) the extracted minified code is correctly deminified, or (c) there is no alternate code path for subagent sessions. The report's negative check (the flag `tengu_quiet_hollow` is absent) proves the mechanism *moved*, not that the *new* mechanism works as claimed. This is a gap between static analysis confidence and dynamic verification.

**Severity:** Medium likelihood (binary extraction is noisy), medium impact (the headline rests on this mechanism), low complexity to mitigate (run an interactive session with `showThinkingSummaries: false` and subagent spawns to check if thinking blocks remain empty).

**Acceptance check:** Run a test session that spawns a subagent and captures its JSONL; inspect the thinking blocks for non-empty text. If any block has non-empty text, the mechanism claim is falsified.

---

### L6-F3: Central falsifiable test was declined; headline is conditionally true but unfalsified

**Location:** §10 "Open questions" Q1 + lines-of-inquiry.md "declined"

**Quoted test:** "Does `showThinkingSummaries: true` produce non-empty thinking in a **non-interactive** subagent transcript, given the force-omitted guard?"

**Risk:** This is the single experiment that could overturn the headline (as the report admits: "this is the single experiment that could overturn the headline"). The test was declined citing consent boundaries ("writing to the user's global ~/.claude/settings.json is a state-modifying change outside the working tree and outside this seat's consent"). This is correct for this seat's consent boundary, **but it is also the load-bearing empirical test**. The headline rests on the assumption that the force-omitted guard cannot be bypassed even if `showThinkingSummaries: true`. The mechanism is read from code, but the code path's *actual effect* in practice has not been tested. The headline is unfalsifiable as written.

**Severity:** Certain likelihood (this is the missing test), high impact (it's explicitly the falsification point), trivial complexity to mitigate (one settings change + one session).

**Acceptance check:** As blue-synthesize reported at [^IssueStatuses], this test is open question #1 and explicitly unresolved. Red accepts that this seat cannot run it (consent boundary), but red MUST note: any deployment of artifact-based reasoning recording as a substitute for thinking blocks assumes this test would fail. If it passes (showThinkingSummaries true DOES produce non-empty thinking in subagents), the headline is overturned and the whole framing changes. This is not a finding to close; it is a DEPENDENCY of the headline that the artifact-based recommendation cannot resolve.

---

### L6-F4: Artifact-based reasoning recording conflates decision *recording* with reasoning *quality*

**Location:** §8 "What to do instead" + §5 "Faithfulness and the limits"

**Quoted claim:** "Recording avenue status (pursued/abandoned/declined with reasons), manifest rows, closure anchors, friction entries and repair history produces reasoning evidence that is durable, git-tracked, intentional, append-only, and checkable by an adversary against the artifact it cites."

**Risk:** The report argues artifacts are superior to thinking blocks because they are durable and can be disconfirmed (an artifact can contradict itself if the tool calls don't match the claimed avenue). But this conflates two different things:
- **Decision recording** (what was chosen and why it was chosen): artifacts capture this as self-report.
- **Reasoning quality** (was the choice sound? did the reasoning avoid blind spots?): Section 6 explicitly states this is Tier 4, which requires an external oracle.

The report then downgrades the self-report critique: "an agent writing down 'I declined this avenue because X' is also a self-report, and also potentially post-hoc" but decides this is acceptable because the artifact is "better on durability and non-circularity, not on sincerity." But the recommendation conflates recording-quality with reasoning-quality. If we are grading reasoning quality (Section 4's central question), artifact recordings remain Tier 4. The benefit of artifacts is durability and auditability, not sound reasoning. The report does not claim sincerity; it admits the artifact is potentially post-hoc rationalization.

**Severity:** Medium likelihood (the conflation is subtle), medium impact (it could overstate the value of artifact recording), low complexity to mitigate (reframe the recommendation to "durability and disconfirmability" rather than "reasoning quality").

**Acceptance check:** Verify that every claim in the final report grading reasoning quality as citable rests on Tier 3+ evidence, not on artifact recording alone. If any Tier 4 claim is supported only by artifact evidence, flag it.

---

### L6-F5: Silent tool-result truncation has no detection mechanism; "disclosure" mitigation is unenforceable

**Location:** §5 "Faithfulness and the limits" + §9 "Risk matrix"

**Quoted claim:** "Tool outputs are cut at several layers without any marker that data was removed; the model then answers confidently from a fragment, and nothing in the transcript flags it."

**Risk:** The risk matrix lists this as "risk-accept with disclosure — no audit marker exists to detect it after the fact." But "disclosure" here means the report will note that truncation can happen, not that truncation will be detected in-run. A harness reading a JSONL transcript has no way to know a tool result was truncated without comparing lengths or having access to the ground truth of what the tool should have returned. The only mitigation is pre-emptive (raise `maxResultSizeChars`) and the default is still lossy. The risk is accepted as written, but the "disclosure" is merely acknowledging the risk, not mitigating it. A finding built on truncated tool output is indistinguishable from a finding built on complete output.

**Severity:** Medium likelihood (truncation is known to happen), high impact (a truncated result could lead to a false conclusion), medium complexity to mitigate (proactively set `maxResultSizeChars` and log result lengths for spot-checking).

**Acceptance check:** Any finding that relies on tool results should log the result length and raise an alert if length == `maxResultSizeChars` (likely truncated). Without this check, findings are at risk of resting on incomplete evidence.

---

### L6-F6: Version-bound findings have a shelf life equal to the release cycle; re-verification cost is not zero

**Location:** §9 "Risk matrix" last row + Provenance section

**Quoted claim:** "All binary-derived findings (display resolver, OpenTelemetry redaction, settings schema, instrument names) are specific to Claude Code v2.1.215 on Windows, read 2026-07-19."

**Risk:** The mechanism finding is version-specific. The report notes: "the mechanism moved; the lever's name survived the move" between v2.1.71 and v2.1.215. The next version (expected within months) could have another mechanism shift. The risk matrix says this is "risk-accept — the sweep is cheap; schedule it, do not engineer around it." But "cheap" is measured in cost-per-run, not in operational burden. If the report's recommendation is to use artifact-based reasoning recording because telemetry is unreliable, but the unreliability finding is version-bound and needs re-verification every few months, then the recommendation is version-conditional. The report does not address: what happens if v2.2.0 ships with a different display resolver? Is the recommendation still sound? The headline assumes a version-specific property that will change.

**Severity:** High likelihood (vendor has demonstrated version changes), medium impact (if the mechanism changes, the headline may change), low complexity to mitigate (schedule re-verification every release).

**Acceptance check:** Version the recommendation explicitly: "This recommendation applies to Claude Code v2.1.215. Re-verify via the binary-sweep method on each new release. If the display resolver changes, re-evaluate whether artifact recording remains the superior choice."

---

### L6-F7: Compliance API reasoning-category absence is an absence claim over documented surface only

**Location:** §3 "Settings and APIs" + §7 "What cannot be audited"

**Quoted claim:** "No documented Claude or Claude Code endpoint returns reasoning summaries, decision alternatives, confidence scores, or reasoning-branch metrics."

**Risk:** This is an absence claim: the report searched the documented public surface (platform.claude.com) and found no reasoning API. But the Compliance API is enterprise-only, and enterprise products often have undocumented capabilities. The report notes: "Lane-1 reported 260+ activity types; publicly accessible sources report ~30... The 260+ figure was not re-verified by blue-synthesize (no enterprise access)." The gap suggests undocumented reasoning categories may exist in the Compliance API. The conclusion "no reasoning API exists" is correct for the documented surface, but it may be false for the full product. The report carries this asymmetry in a footnote, but the headline treats it as a resolved fact.

**Severity:** Low likelihood (enterprise undocumented capabilities are rare but not nonexistent), medium impact (if a reasoning API exists in Compliance, the recommendation changes), low complexity to mitigate (note the bounded scope: "no documented reasoning API").

**Acceptance check:** State explicitly in the final report: "This finding is bounded to the documented public API surface (platform.claude.com). Enterprise Compliance API may have undocumented reasoning categories; we did not verify this."

---

### L6-F8: Steelman attack on "platform explicitly forbids" thinking audit

**Location:** lines-of-inquiry.md "declined" entry 1

**Quoted reason for declining:** "platform explicitly forbids treating thinking as audit evidence; raw stream requires vendor sales contact; unverifiable + inaccessible"

**Risk:** "Explicitly forbids" is stated as fact, but the evidence is warranty-declining language. The report cites documentation: "do not parse, modify, log, or treat thinking signatures as user-readable audit evidence." But "do not" is a recommendation, not a prohibition. The same documentation also states: "If your product needs an audit trail, record prompts, tool calls, approvals, files changed, diffs, and final answers." This prescriptive guidance (record X) does not logically entail "treating thinking as audit evidence is forbidden." It entails "thinking signatures are not sufficient; supplement with X." The platform declines to warrant thinking as audit evidence; it may not forbid it. This is a semantic difference with operational consequences: if thinking is not forbidden, only unwarranted, then the raw-thinking avenue is declined on risk-grading (unverifiable + needs sales contact), not on a platform prohibition.

**Severity:** Low likelihood (the terminology matters mainly for future vendor clarification), low impact (the avenue is rightly declined on accessibility grounds), trivial complexity to mitigate (reword the decline reason).

**Acceptance check:** Verify the exact vendor language in the Thinking Audit Guidance footnote. If it says "do not treat as audit evidence," that is a recommendation, not a prohibition. Frame the decline as: "declined due to lack of vendor-provided access path and lack of warranty, not due to platform prohibition."

---

### L6-F9: OpenTelemetry redaction is code-enforced, but export path is version-bound

**Location:** §4 "OpenTelemetry: complete on acts, hardcoded-blind on reasoning"

**Quoted finding:** "The redaction is applied on both the request-body and response-body export paths."

**Risk:** The report states the redaction is "an unconditional map over assistant content blocks, replacing `thinking` with `<REDACTED>`" on "both the request-body and response-body export paths." But the export paths are themselves version-bound. A future version of Claude Code could add new export paths (e.g., direct file write, streaming to Langfuse, OTEL collector bypass). The redaction function is hardcoded, but if a new path is added without piping through the redaction function, thinking content leaks. The report treats the existing paths as exhaustive but does not enforce that all *future* export paths must call the redaction function. This is a design risk, not a current leak.

**Severity:** Low likelihood (new export paths would require deliberate implementation), high impact (thinking content leak), low complexity to mitigate (architecture decision: all export paths must pipe through a single redaction function).

**Acceptance check:** Verify that the redaction function is a single bottleneck. If any export path bypasses it (including streaming, buffering, or direct file write), flag it. Recommend: "Enforce via architecture: all OTEL export paths must call the redaction function; no path may bypass it."

---

### L6-F10: Tool-result visibility gap is acknowledged as anecdote, not quantified

**Location:** §5 "Faithfulness and the limits"

**Quoted claim:** "A reported case has the model running reads and greps whose results were never displayed, then asserting conclusions about them; the assertion, not the evidence, is what the reader saw."

**Risk:** Issue #32997 reports one case where tool results were hidden. The report carries this as a *visibility gap* but notes it is "a single anecdotal report on an issue now closed as not planned." The frequency of this failure mode is unknown. Is it a rare edge case, or a recurring pattern? If it is rare, it may not justify architectural changes. If it is common, it is a serious blindspot. The report does not quantify this risk, leaving it as a low-confidence finding.

**Severity:** Unknown likelihood (sample size is N=1), medium impact (when it happens, reasoning is obscured), medium complexity to mitigate (audit tool-result visibility as part of trajectory analysis).

**Acceptance check:** Note in the final report: "The visibility gap is reported in a single issue and its frequency is unknown. Recommend: sample transcripts for cases where tool results are called but not displayed in the transcript's natural-language portion; quantify the prevalence."

---

### L6-F11: Adaptive thinking effort levels are not exposed; enables silent latency-driven optimization

**Location:** §3 "Settings and APIs"

**Quoted finding:** "The API does not expose which effort level the model selected or how effort shaped the decision, so identical prompts under different latency conditions may produce different reasoning and identical outputs — making reasoning-quality adjudication impossible without controlled re-execution."

**Risk:** Claude Code can use adaptive thinking, which automatically selects effort (low/medium/high/max) based on latency/cost. The effort level is not exposed in the API or transcript. This means:
1. An agent running the same prompt at different times could produce different reasoning (and identical outputs).
2. A harness reading the transcript cannot know which effort level was selected.
3. Reasoning-quality findings become non-reproducible without controlling effort.

This is a quiet failure mode: the agent *appears* to have a consistent strategy, but the reasoning branching (and thus the quality of reasoning) is latency-dependent. The report identifies this as a Tier-4 problem (requires external oracle / controlled re-execution) but does not recommend a mitigation. Red should flag: this is a design-level issue that makes reasoning reproducible only under controlled conditions.

**Severity:** Medium likelihood (adaptive thinking is used when models support it), medium impact (reasoning quality is non-reproducible), high complexity to mitigate (expose effort level in API and transcript; or disable adaptive thinking for audit sessions).

**Acceptance check:** Recommend: "If reasoning quality is to be audited, disable adaptive thinking (set `thinking.effort` explicitly to a fixed level) and expose the level in the transcript."

---

### L6-F12: Faithfulness critique applies to artifacts as well as thinking blocks

**Location:** §5 "Faithfulness and the limits" vs. §8 "What to do instead"

**Quoted asymmetry:** Section 5 quotes Anthropic: "models often make decisions based on factors that they don't explicitly discuss in their thinking process." This proves thinking blocks are not faithful. Section 8 then recommends artifact recording: "an agent writing down 'I declined this avenue because X' is also a self-report, and also potentially post-hoc," but downgrades the risk because artifacts are "better on durability and non-circularity, not on sincerity."

**Risk:** The report treats thinking blocks as untrustworthy (Tier 4, requires external oracle) because they may not reflect actual reasoning. But the same critique applies to artifacts: an agent writing "I declined avenue X because it was impossible" might not be describing the factors that actually influenced the decision. The report admits this ("potentially post-hoc") but then recommends artifacts without addressing the sincerity gap. This is an asymmetry: artifacts are trusted to a higher bar than the evidence warrants. The recommendation is sound (use artifacts for durability, not for reasoning quality), but the justification oversells artifacts as a *reasoning* record.

**Severity:** Medium likelihood (agents can confabulate reasons), medium impact (artifact-based reasoning quality claims would rest on weak evidence), low complexity to mitigate (explicitly reframe: "use artifacts for decision transparency, not for reasoning quality audit").

**Acceptance check:** Verify that any reasoning-quality claim graded as Tier 4 is not supported by artifact recording alone. Artifacts can disconfirm a claimed avenue (if the tool calls don't match), but they cannot verify reasoning quality.

---

### L6-F13: Nested subagent transcript discovery requires recursive globbing; naive glob misses 95% of evidence

**Location:** §1 "What the transcript actually contains"

**Quoted finding:** "a top-level glob of the projects directory found 16 files where a recursive walk found 294."

**Risk:** The transcript store has a deep nesting structure: subagent and workflow sessions nest under the parent session directory. A harness that uses a non-recursive glob pattern (e.g., `~/.claude/projects/*/*.jsonl`) will find ~16 files. A harness that uses recursive globbing (e.g., `~/.claude/projects/**/*.jsonl`) will find ~294 files. This is a 18× difference. The report notes the gap but does not require harnesses to use recursive globbing. Any naive implementation of transcript auditing will miss ~95% of the agent evidence. This is not a finding to close in isolation; it is a requirement for any transcript-based auditing framework.

**Severity:** Certain likelihood (the directory structure is fixed), high impact (naive globbing loses evidence), trivial complexity to mitigate (use recursive globbing).

**Acceptance check:** Any transcript auditing framework MUST use recursive globbing and MUST enumerate the nesting depth. Document the glob pattern explicitly.

---

### L6-F14: Risk matrix underestimates operational burden of "trivial" re-enumeration

**Location:** §9 "Risk matrix"

**Quoted mitigation:** "Metric/span names drift between versions — **risk-accept** — re-enumerate per version" (complexity: trivial)

**Risk:** The report lists "re-enumerate against installed binary" as a trivial mitigation for metric/span name drift. But in practice, this is an operational burden: every time Claude Code updates, a harness consuming OpenTelemetry data must re-enumerate the metric names to ensure the queries still work. The report treats this as a one-off "cheap sweep," but across many deployments and versions, the cumulative cost is not trivial. The risk-accept is correct (re-enumeration is cheaper than building an abstraction), but the complexity grade understates the operational burden. This is particularly acute if the harness is deployed before Claude Code updates and begins seeing telemetry with unfamiliar metric names.

**Severity:** High likelihood (metric names do drift; the report notes this happened between v2.1.71 and v2.1.215), low impact (telemetry parsing errors are detectable), low complexity to mitigate per run (but medium operational burden across many deployments).

**Acceptance check:** Recommend: "metric/span name drift is anticipated. Telemetry parsers MUST validate expected metric names at startup and fail loudly if names change. Do not silently skip unexpected metrics."

---

### L6-F15: Multi-agent verification tradeoff is carried without quantified benefit

**Location:** §8 "What to do instead"

**Quoted finding:** "A related reported result is that judge-model auditing alone catches a minority of errors while combination with deterministic tooling catches far more; the specific figures we inherited (~45% vs 94%) reach us through a secondary listicle and are **not** leaf-verified — treat the direction as the finding and the numbers as unverified."

**Risk:** The report cites a 45% vs. 94% improvement but does not verify the figures. It then extracts the *direction* as a finding ("deterministic checks materially outperform judge-only auditing") but does not compare this to the report's own recommendation (artifact-based reasoning recording). Is artifact recording a form of "deterministic tooling"? If not, how much does it improve over judge-model auditing? The tradeoff between multi-agent verification (overhead of running two agents) and artifact-based recording (overhead of agents writing down decisions) is not compared. The report recommends artifacts but does not quantify their effectiveness versus alternatives.

**Severity:** Medium likelihood (comparison is incomplete), medium impact (the recommendation may not be optimal), low complexity to mitigate (quantify expected improvement of artifacts over baseline transcript auditing).

**Acceptance check:** Recommend: "The effectiveness of artifact-based reasoning recording versus multi-agent verification is not measured. Propose: measure recall and false-positive rates of findings under both approaches in a controlled trial."

---

### L6-F16: Binary string extraction method has failure modes that are not fully checked

**Location:** §2 "The mechanism, read out of the shipped client" + footnote [^BinaryFlagAbsent]

**Quoted method:**  "minified identifiers collide (`W1` resolved to an unrelated function on first grep), absence of a string does not prove absence of a behavior (names can be constructed at runtime)..."

**Risk:** The method used to extract the display-resolver from the binary has known failure modes:
1. **Minified identifiers collide** — the report checked this once (`W1`) but does not verify the display-resolver extraction itself for collisions.
2. **Runtime-constructed names** — the report notes this is possible but does not check whether the display resolver uses runtime name construction.
3. **Dead code** — the report does not verify that the extracted function is actually called at the claim site.

The report did run one negative check (the flag `tengu_quiet_hollow` is absent), but this only proves the mechanism moved, not that the extracted mechanism is correct. Red should flag: the binary-extraction method is being used at the limit of its confidence.

**Severity:** Medium likelihood (extraction is noisy but the extracted code is readable), medium impact (incorrect extraction would overturn the mechanism finding), low complexity to mitigate (empirically verify by running test sessions and comparing JSONL output to predicted behavior).

**Acceptance check:** Do not cite binary-extracted findings without empirical validation. For the display-resolver, run a test that spawns a subagent and verify that thinking blocks are empty (predicted) or non-empty (contradiction).

---

## Summary and tone

Red has identified **16 findings**, mostly in the medium-low severity range. The darkside analysis focuses on:

1. **Temporal and version volatility** (F1, F6): snapshot-specific counts and version-bound mechanisms are not invariants.
2. **Untested assumptions** (F2, F3, F21): mechanism claims rest on code-reading, not empirical verification; the central falsifiable test was not run.
3. **Conflations and asymmetries** (F4, F12): artifact recording is claimed as reasoning evidence but remains Tier 4; faithfulness critique is applied unevenly.
4. **Silent failure modes** (F5, F9, F10, F11, F17): truncation, latency-driven reasoning changes, and tool-result visibility are accepted risks with weak mitigations.
5. **Absence claims over bounded surfaces** (F7, F19): the lack of documented reasoning APIs does not rule out undocumented capabilities.
6. **Operational burden underestimated** (F8, F14): "cheap" re-verification and "trivial" re-enumeration carry real operational costs across deployments.

The findings do not overturn the headline (artifact-based reasoning recording is a sound recommendation given the constraints), but they identify risk categories and assumptions that should be made explicit in the final report.

Red assesses the report as **provisionally sound** with the contingencies noted above. The darkside lens found no disqualifying errors, but multiple risk surfaces that warrant disclosure in the limitations section and in any deployment of artifact-based reasoning recording.

**Recommendation for merge:** Carry F1, F2, F3, F6, F9, F11, F13 as open risks requiring disclosure. Accept as-is: F4, F5, F7, F8, F10, F12, F14, F15, F16 (these are already disclosed in the report; the finding adds precision to the grading).
