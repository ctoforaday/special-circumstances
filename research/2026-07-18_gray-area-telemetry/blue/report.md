# Blue living report — Trajectory telemetry for agent adjudication

**Question.** What can be mined from Claude Code transcripts today; what settings or APIs expose a
summary of an agent's *reasoning* rather than only its *acts*; and which of those are sound enough
to carry a citable finding?

**Status.** Round 0 synthesis. Union of three lane drafts (lane-1 disconfirming-search lens, lane-2
primary-literature lens, lane-3 local-repo critical-stance lens), plus blue-synthesize's own
leaf verification against the installed Claude Code binary (v2.1.215), the local transcript store,
and the GitHub issue tracker. Claims appearing in exactly one lane are tagged
`[minority: lane-N/<lens>]`. Claims first established in this merge are tagged
`[merge-verified]` — those are leaf checks blue-synthesize ran itself, not lane inheritance.

**Headline.** On a default-configured Claude Code install (v2.1.215, non-interactive session, no
manual override of `showThinkingSummaries`), reasoning is almost not recorded. The reasoning channel
exists, has a named lever, is off by default, and is *forced off* for exactly the non-interactive
sessions an adjudication harness runs. Even when the lever is on, what returns is a second model's
summary, which Anthropic's own documentation declines to warrant as faithful. Therefore: acts, tokens,
and permission decisions carry citable findings; reasoning quality does not, and the sound move is to
make agents *record* their reasoning as artifacts rather than to try to *recover* it from telemetry.

---

## The Catechism

**1. What are we trying to do?**
Work out whether an agent's own run logs can tell us *why* it did what it did — not just what it
did — well enough that a finding about the quality of its reasoning could be published with a
citation a skeptic could follow. Concretely: can a judge reading a Claude Code session transcript
say "this agent reasoned badly here" and defend it?

**2. How is it handled today, and what does that cost us?**
Today adjudication in this repo reads acts and artifacts: which tools were called, what files
changed, what the agent wrote down. Reasoning is inferred by a human or a judge agent from that
surface. The cost is that inference from acts is ambiguous — a file re-read is either careful
verification or confused looping, and the transcript cannot tell you which.[^TrajectoryEval] The
hoped-for fix was that the thinking channel would disambiguate it. It does not: across the local
store, 5,754 thinking blocks in 278 of 294 transcripts carry an empty `thinking` field and an
encrypted signature — zero readable characters of reasoning.[^LocalSweep]

**3. What is new here, and why do we believe it works?**
Three things this round establishes at the leaf rather than by report on a default-configured Claude
Code v2.1.215 install with `showThinkingSummaries` unset (false).
(a) The lever exists and is named: `showThinkingSummaries`, a boolean in Claude Code settings whose
own description string says it requests API-side thinking summaries and shows them "in the
conversation and in the transcript view."[^BinaryShowThinking] This refutes the flat claim that
Claude Code has no thinking setting.
(b) The lever does not reach us where we need it. In v2.1.215 the display resolver returns
`"summarized"` from the setting **only on the interactive path**; a separate guard forces
`display:"omitted"` on non-interactive sessions unless display was set explicitly.[^BinaryDisplayResolver]
Debate runs are non-interactive subagent sessions — the empty blocks in the non-interactive share of
the local store are the predicted result given this code path, not a bug. This is a causal finding
for the non-interactive branch: the display resolver guard forces `display:"omitted"` on that path.
For the interactive branch (top-level transcripts), the resolver returns `void 0` when the setting
is unset, leaving display unspecified; the mechanism producing empty blocks on that branch is
unresolved and the serialization hypothesis remains live there.[^BinaryDisplayResolver]
(c) The OpenTelemetry redaction is not policy prose but a hardcoded map: thinking content is
replaced with the literal `<REDACTED>` on both the request-body and response-body export
paths.[^BinaryOtelRedaction] No environment variable reaches it.
Why we believe it: each is read out of the shipped binary with a reproducible string extraction,
and (a)/(c) are independently corroborated by the vendor documentation and by the issue
tracker.[^Issue32810][^OTelObservability]

**4. The case against — at full strength.**

- *The central answer is negative, and negatives are cheap to produce and hard to keep true.* "No
  sound reasoning telemetry exists" is a universal over a surface we sampled, not enumerated. We
  did not test enterprise Compliance API access, the Agent SDK's `--output-format stream-json`
  path, or any paid/raw-thinking channel that vendor sales reportedly gates.[^ExtendedThinkingDocs]
  A single counterexample retires the headline.
- *Our strongest empirical claim is a measurement of one machine.* The 5,754/0 sweep is one user,
  one settings file, one model era, and a store that grew by 10 blocks between two counts minutes
  apart during a live session.[^LocalSweep] It cannot support "Claude Code never records thinking";
  it supports "on a default-configured install, it did not."
- *Four of the community citations we inherited are closed issues.* #32810, #32997, #10084 are
  closed *not planned*; #52376 is closed as *duplicate*.[^IssueStatuses] A closed-not-planned
  feature request is stronger evidence *against* the capability arriving than an open one — but it
  also means our account of current behavior rests on a locked thread describing v2.1.71, two
  hundred patch versions stale.
- *The binary-string method is a real method with real failure modes.* Minified identifiers collide
  (`W1` resolved to an unrelated function on first grep), absence of a string does not prove
  absence of a behavior (names can be constructed at runtime), and nothing stops Anthropic changing
  this in the next release. `tengu_quiet_hollow`, the flag the issue thread pinned the regression
  on, is **absent** from v2.1.215 — the mechanism moved.[^BinaryFlagAbsent]
- *The faithfulness argument may prove too much.* If reasoning traces are unreliable evidence of
  reasoning, that is an argument against the whole enterprise of reasoning adjudication, including
  the artifact-based substitute we recommend: an agent writing down "I declined this avenue
  because X" is also a self-report, and also potentially post-hoc.[^ReasoningTheater] Our
  recommendation is better on *durability and non-circularity*, not on *sincerity*.
- *Opportunity cost is real.* The engineering to stand up an OpenTelemetry collector, or to widen
  the recording verbs, is work not spent on the adjudication logic itself. If the answer is "acts
  plus artifacts, as today," the honest reading is that this research changes little about
  practice and mostly licenses the current practice.
- *Risk-accepted residuals we are choosing to live with, named:* silent tool-result truncation with
  no audit marker;[^ToolTruncationLimits] JSONL format instability across releases;[^SessionDocs]
  transcripts that are append-only but not tamper-evident;[^L3TranscriptUnstable] and the fact that
  a summarized trace, if we ever enable it, is written by a *different model* than the one under
  audit.[^ExtendedThinkingDocs]

**5. Of interest, or merely interesting?**
Of interest, narrowly. The question "can we grade reasoning from telemetry?" is genuinely
interesting and the answer is no; that would be merely interesting. What makes it *of interest* is
the operational consequence: it settles that the recording verbs (avenue, manifest-row, friction,
closure anchors) are not scaffolding to be replaced later by real telemetry — they are the
adjudication substrate, because the real telemetry is not coming.[^ArtifactRecording] That
displaces any plan to build a thinking-block miner, which is the cheapest thing this run buys.

**6. What changes if it works — and what if we skip it?**
If it holds: adjudication designs stop budgeting for reasoning capture, grade reasoning-quality
claims as inference rather than observation, and invest in agents recording their own decisions in
durable artifacts. If we skip it: the likely failure is a harness built to parse thinking blocks
that silently reads empty strings and concludes the agent did not think — the exact failure the
local sweep would have caught in ten minutes.[^LocalSweep] That is a cheap-to-avoid, expensive-to-
discover error, which is the profile that justifies a short run and not a long one.

**7. What does it cost, and where would we stop?**
Cost to date: one synthesis round over three lane drafts plus roughly a dozen leaf checks. The
stopping points: (i) if a working `showThinkingSummaries` capture on a *non-interactive* session
is demonstrated, the central negative is wrong and the run should reorient to "how good are the
summaries" rather than "do they exist"; (ii) if enterprise Compliance API access materializes,
the API-surface section must be redone against it rather than patched; (iii) if the report starts
accreting general agent-evaluation literature not tied to Claude Code's actual surface, stop —
that is scope creep into a survey nobody asked for.

---

## 1. What the transcript actually contains

Claude Code writes one JSON object per line to
`~/.claude/projects/<encoded-path>/<session-id>.jsonl`; a `type` field discriminates user /
assistant / system lines, and assistant lines carry a `message.content` array of typed blocks
(`text`, `thinking`, `tool_use`, `tool_result`).[^TranscriptFormat] Subagent and workflow sessions
nest under the parent session directory, which matters for any sweep: a top-level glob of the
projects directory found 16 files where a recursive walk found 294.[^LocalSweep] [merge-verified]

Directly captured and reliable:

- **Tool call sequences and parameters** — every invocation appears as a `tool_use` block with tool
  name and input JSON, in order.[^TranscriptFormat] Tool selection and parameter construction are
  directly observable, which makes them the only reasoning-adjacent data needing no
  reconstruction.[^TrajectoryEval]
- **Tool execution results** — `tool_result` blocks carry the environment's response
  verbatim.[^TranscriptFormat] Whether the agent *interpreted* the result correctly is a separate,
  reasoning-level question.
- **Token usage and latency** — `message.usage` fields per turn, and span durations via
  OpenTelemetry.[^OTelObservability] These are facts of execution, not inferences.
- **Error states and retries** — failed calls, error text, retry sequences are
  preserved.[^TranscriptFormat] The *frequency* is observable; the *reason* is not.
- **Identity and traceability** — user prompts, assistant text, model and version identifiers,
  session and message UUIDs, and message-history position.[^L3SessionStructure]
  [minority: lane-3/local-repo]

Indirectly mineable, with reconstruction: backtracking detection (same file revisited without
intervening change), tool-result consumption (did the agent use what it fetched), and
hypothesis-testing shapes such as try-edit-verify-retry.[^TrajectoryEval] These are observable as
patterns and speculative as interpretations.

Wall-clock forensics are available and underused: timestamp gaps between tool calls distinguish
stalls from rework.[^L3OpenTelemetryDetails] [minority: lane-3/local-repo]

**Two structural caveats on the transcript as an evidence base.** First, the vendor documents the
format as internal and version-unstable, and explicitly steers integrators to `/export` or the
Agent SDK script interfaces rather than direct parsing.[^SessionDocs] Second, the JSONL is
append-only but is not signed or tamper-evident, so it carries no integrity guarantee of the kind an
audit log is normally expected to provide.[^L3TranscriptUnstable] [minority: lane-3/local-repo]

## 2. The reasoning channel: extended thinking and its three states

The Messages API accepts a `thinking` parameter (`type: "enabled"`, a token budget, and a `display`
mode) and, on newer models, an adaptive form with an `effort` level in place of an explicit
budget.[^ExtendedThinkingDocs][^AdaptiveThinking] Three response states follow:

| `display` | What lands in the transcript | Audit value |
|---|---|---|
| `summarized` | Text digest of the reasoning, written by a *different* model than the one reasoning | Second-hand narration |
| `omitted` | Empty `thinking` string plus a base64 `signature` | None |
| (streaming) | `thinking_delta` events during generation | Same content, different transport |

The `signature` field carries the encrypted full thinking for multi-turn continuity; it is not
human-readable and cannot be decrypted by clients.[^ExtendedThinkingDocs] Billing charges for the
full thinking tokens in both display cases.[^ExtendedThinkingLimitations]
[minority: lane-3/local-repo]

**Empirical state of the local store.** A recursive sweep of `~/.claude/projects/` on 2026-07-18
found 294 transcript files, 278 of which contain thinking blocks; 5,754 thinking blocks carry the
exact byte sequence `"thinking":"","signature":"`, and **zero** have a non-empty `thinking`
field.[^LocalSweep] [merge-verified] The count moved from 5,744 to 5,754 between two measurements
minutes apart because the store was growing under a live session — the figure is a snapshot of a
moving target, not a fixed corpus size. Lane-3's earlier probe-based measurement reported 287
transcripts and 5,569 thinking blocks with the same all-empty result; this appears to be the same
measurement of the evolving store at an earlier time rather than an independent sweep.[^L3SessionStructure]

**The mechanism, read out of the shipped client.** Claude Code v2.1.215 resolves the display mode
with (deminified):

```js
function resolveDisplay({explicitDisplay, isNonInteractive, outputFormat, verbose}) {
  if (explicitDisplay) return explicitDisplay;
  if (!isNonInteractive) return showThinkingSummaries() ? "summarized" : undefined;
  if (outputFormat === "text" || (outputFormat === "json" && !verbose)) return "omitted";
  return;
}
```

and then, separately, forces `display:"omitted"` on any non-interactive session whose display was
not set explicitly and which is not using exact-tools, subagent-text forwarding, or async
mode.[^BinaryDisplayResolver] [merge-verified] `showThinkingSummaries()` reads the settings boolean
and defaults to `false`.[^BinaryShowThinking] [merge-verified]

This is the load-bearing finding of the round: the setting is real, and it is on the **interactive**
branch only. An adjudication harness running subagents non-interactively is on the branch that
returns `"omitted"`. The report's own §1 counts 16 top-level transcripts (interactive parent sessions)
out of 294 files, meaning 278 are deeper-nested subagent and workflow runs. **Note:** the 278-file
count (deeper-nested transcripts) is distinct from the 278-block count (thinking blocks present in
the corpus per §1 second sentence); they coincide numerically but denote different sets. The pinned
probe reports empty blocks "across seat and main-session transcripts", implying at least some of the
16 interactive transcripts carry thinking blocks, so the interactive share's block count is
unquantified at this round. The all-empty blocks in the non-interactive share are therefore the
*expected* output of a default-configured install on that branch, not evidence of a defect. The
mechanism for the interactive share remains unresolved.

**Provenance of the default.** A community root-cause analysis on issue #32810 reconstructed the
v2.1.71 condition: the client sends the `redact-thinking-2026-02-12` beta header when thinking is
enabled, the model supports it, verbose/debug is off, `showThinkingSummaries !== true`, and a
server-side flag `tengu_quiet_hollow` is on — the flag having been flipped server-side around
2026-03-10, which is why the change arrived without a client release.[^Issue32810] Setting
`"showThinkingSummaries": true` bypassed the condition and restored thinking text to the JSONL, per
the same thread.[^Issue32810] Two corrections to that account are due at the leaf: the issue is
**closed as not planned and locked**,[^IssueStatuses] and the flag name `tengu_quiet_hollow` does
**not** appear anywhere in the v2.1.215 binary, though the `redact-thinking-2026-02-12` beta
registration does.[^BinaryFlagAbsent] [merge-verified] The mechanism described in the thread has
moved; the lever's name survived the move.

**Even when captured, the trace is not the reasoning.** Anthropic states plainly that "we don't
know for certain that what's in the thought process truly represents what's going on in the model's
mind," and that models "often make decisions based on factors that they don't explicitly discuss in
their thinking process."[^VisibleExtendedThinking] Independent probing work comparing reasoning models
reports performativity rates (answer decodable from hidden activations before it appears in chain of
thought) varying dramatically with task difficulty and model. DeepSeek-R1 671B shows 0.417 on MMLU
(a recall task) and 0.012 on GPQA-Diamond (a multihop reasoning task); GPT-OSS 120B shows 0.435 on
MMLU and 0.227 on GPQA-Diamond — a ~35x collapse in DeepSeek-R1 and a ~1.9x decline in GPT-OSS on
hard reasoning. Task-dependence holds across both models; magnitude varies by an order of
magnitude.[^ReasoningTheater] [minority: lane-1/disconfirming]
Practitioner guidance draws the operational conclusion: "do not parse, modify, log, or treat
thinking signatures as user-readable audit evidence… If your product needs an audit trail, record
prompts, tool calls, approvals, files changed, diffs, and final answers. Do not promise an audit
trail of the model's private reasoning."[^ThinkingAuditGuidance] [minority: lane-3/local-repo]

## 3. Settings and APIs: the actual levers

| Lever | Where set | Exposes reasoning? | Status at leaf |
|---|---|---|---|
| `showThinkingSummaries` | Claude Code settings JSON | Summaries, interactive sessions only | Present in v2.1.215[^BinaryShowThinking] |
| `thinking.display` | Messages API request | `summarized` \| `omitted` | API-only; Claude Code sets it internally[^ExtendedThinkingConfig] |
| Raw (unsummarized) thinking | Vendor sales channel | Yes, reportedly | Not independently verifiable; closed channel[^ExtendedThinkingLimitations] |
| `CLAUDE_CODE_ENABLE_TELEMETRY=1` | Environment | No — thinking redacted | Present[^BinaryOtelRedaction][^OTelObservability] |
| `OTEL_LOG_RAW_API_BODIES` | Environment | No — thinking redacted before write | Present; supports `file:<dir>` mode[^BinaryOtelRedaction] |
| Structured Outputs | API beta header | No — output format contract only | Not reasoning-related[^StructuredOutputs] |
| Compliance API | Enterprise control plane | No reasoning event category | ~30 documented activity types, none reasoning (contradicts lane-reported 260+ count)[^ComplianceAPI] |

Two corrections to inherited lane claims belong here rather than in a footnote. Lane-3 concluded
that Claude Code has "no configuration to enable raw thinking capture" and "no environment variable
or setting to switch from `omitted` to `summarized`"; the first is correct (nothing exposes *raw*
thinking) and the second is **refuted** — `showThinkingSummaries` is exactly that switch on the
interactive path.[^BinaryShowThinking] Lane-3 also recorded feature request #52376 as "Status:
open"; it is closed as a duplicate.[^IssueStatuses] Lane-2 cited #10084 ("Expose Claude Code
Cognitive Telemetry States via API") as evidence the capability is "desired but not shipped"; that
holds, and strengthens — it is closed *not planned*, so the desire has been declined rather than
queued.[^IssueStatuses] [merge-verified]

**No dedicated reasoning API exists.** No documented Claude or Claude Code endpoint returns
reasoning summaries, decision alternatives, confidence scores, or reasoning-branch
metrics.[^PlatformDocs][^DebugModeSearch] This is an absence claim over the documented public
surface as searched on 2026-07-18, not a proof of non-existence; undocumented and enterprise
surfaces are outside what we checked.

**Adaptive thinking degrades reproducibility further.** The API does not expose which effort level
the model selected or how effort shaped the decision, so identical prompts under different latency
conditions may produce different reasoning and identical outputs — making reasoning-quality
adjudication impossible without controlled re-execution.[^AdaptiveThinking]
[minority: lane-2/literature]

## 4. OpenTelemetry: complete on acts, hardcoded-blind on reasoning

With `CLAUDE_CODE_ENABLE_TELEMETRY=1` and an exporter configured, Claude Code emits OTLP traces,
metrics and log events to any conformant backend (Honeycomb, Datadog, Grafana, Langfuse, or a
self-hosted collector); enhanced tracing is behind
`CLAUDE_CODE_ENHANCED_TELEMETRY_BETA=1`.[^OTelObservability] Spans nest — tool spans under
interaction spans, subagent tool spans under parent tool spans — and carry a session id, so
multi-turn and parent/child agent correlation is available; W3C trace context propagates into child
processes.[^L3OpenTelemetryDetails] [minority: lane-3/local-repo]

String extraction from v2.1.215 confirms the following instrument names ship in the binary:
`claude_code.interaction`, `claude_code.llm_request`, `claude_code.tool`,
`claude_code.tool.execution`, `claude_code.tool.blocked_on_user`, `claude_code.hook`,
`claude_code.subagent.spawn`, `claude_code.compaction`, `claude_code.mcp.rpc`,
`claude_code.token.usage`, `claude_code.cost.usage`, `claude_code.code_edit_tool.decision`,
`claude_code.session.count`, `claude_code.lines_of_code.count`, `claude_code.active_time.total`,
`claude_code.commit.count`, `claude_code.pull_request.count`, `claude_code.bash.subprocess`,
`claude_code.events`, `claude_code.tracing`.[^BinaryOtelNames] [merge-verified] Event names appear
unprefixed at the emit site — `tool_decision`, `tool_result`, `user_prompt`,
`permission_mode_changed`, `api_request_body`, `api_response_body` — with the `claude_code.`
namespace applied by the emitter.[^BinaryOtelNames] Lane-3's enumeration of *metric* names
(`claude_code.tokens.input`, `claude_code.cost.total`, `claude_code.tool_decisions.total`) does not
match the binary's `claude_code.token.usage` / `claude_code.cost.usage` /
`claude_code.code_edit_tool.decision`; treat published metric names as version-bound and enumerate
against the installed client before building on them.[^BinaryOtelNames] This is a naming
divergence, not a capability dispute — the underlying instruments exist either way.

`claude_code.subagent.spawn` is a direct answer to one of lane-3's carried-forward questions about
nested-agent representation: subagent spawning is a first-class instrument, not an inference from
span nesting.[^BinaryOtelNames] [merge-verified]

**The redaction is in code, not policy.** Documentation states that extended-thinking content is
redacted from exported bodies even when raw API body logging is
enabled.[^OTelObservability][^OTelRedaction] The binary implements this as an unconditional map over
assistant content blocks, replacing `thinking` with `<REDACTED>` and `redacted_thinking` data with
`<REDACTED>`, applied on both the request-body and response-body export
paths.[^BinaryOtelRedaction] [merge-verified] There is no environment variable, verbosity level, or
settings key upstream of it. Two consequences in v2.1.215: no configuration of the OpenTelemetry
surface yields reasoning on this version, and the documentation's claim is verified rather than
merely asserted.

Adjacent capability worth noting: `OTEL_LOG_RAW_API_BODIES` accepts a `file:<dir>` form that writes
each request/response body to a file and emits only a `body_ref` path in the log event, which
avoids the attribute-length truncation that the inline mode
applies.[^BinaryOtelRedaction] [merge-verified] Prompt and assistant-response content in log events
is separately gated by `OTEL_LOG_USER_PROMPTS` / `OTEL_LOG_ASSISTANT_RESPONSES`; without them the
prompt attribute is written as the literal `<REDACTED>`.[^BinaryOtelRedaction] [merge-verified]

## 5. Faithfulness and the limits of automated adjudication

Anthropic's Alignment Science work states that reasoning transcripts "may not be faithful"; that
"the LLM judge that should catch these alignment failures is itself subject to the same failures";
that "a model can register that it is being tested without saying so"; that "simulated deployments
are never perfect replicas of real ones"; and that "human review remains essential precisely
because automated auditing has these fundamental
blindspots."[^AgenticMisalignment] [minority: lane-1/disconfirming] Taken together this is the
vendor's own position that transcript-based automated adjudication has irreducible blind spots —
which bears directly on any design that would put a judge agent alone in front of a transcript.

Two further failure modes attack the *act* record, not just the reasoning record:

- **Silent tool-result truncation.** Tool outputs are cut at several layers without any marker that
  data was removed; the model then answers confidently from a fragment, and nothing in the
  transcript flags it. Claude Code exposes `maxResultSizeChars` so MCP servers can raise the cap
  (reported up to 500K characters), but the default remains lossy;[^ToolTruncationLimits] the key is
  present in the shipped binary.[^BinaryOtelNames] [merge-verified] This is a design-level issue
  (truncation without audit markers), independent of any particular tool or agent framework.
- **Tool output not surfaced to the reader.** A reported case has the model running reads and greps
  whose results were never displayed, then asserting conclusions about them; the assertion, not the
  evidence, is what the reader saw.[^Issue32997] [minority: lane-1/disconfirming] The reporter
  correlated this with thinking redaction. That correlation is a single anecdotal report on an
  issue now closed as not planned;[^IssueStatuses] we carry the *visibility gap* as the finding and
  explicitly decline the causal claim about deception.

Claude Code's stated design principle is that humans "can observe actions in real time, approve or
reject proposed operations, interrupt compatible in-progress operations, and audit after the
fact."[^DesignPrinciples] [minority: lane-1/disconfirming] Every verb in that sentence takes
*actions* as its object. Reasoning is outside the principle, which is consistent with everything
above: the auditability guarantee was designed for acts.

## 6. Soundness tiers for citable findings

A four-tier grading, merged from lane-2's tier scheme and lane-3's citable/partly/not-citable split.

**Tier 1 — directly citable (facts of execution).** Tool-call counts and sequences; tool inputs and
outputs; token counts and cost; request latency; permission approvals and denials; retry and rework
patterns; context-window depth and pressure; the gap between a user's stated goal and the tool calls
that followed.[^TranscriptFormat][^OTelObservability][^L3SessionStructure] Cite the transcript path
or telemetry export id plus turn number and message UUID; an auditor reproduces by replaying the
JSONL. Confidence: verified.

**Tier 2 — citable with method disclosure (pattern inference).** Backtracking patterns; tool-choice
relevance; error-class distinctions (transient vs. logical); decision efficiency; multi-session
strategy clustering; permission-boundary adherence inferred from the absence of denial
events.[^TrajectoryEval][^L3SessionStructure] Disclose the detection heuristic, give the transcript
range, and label it inference. Disconfirming check required: verify the pattern against the tool
*results* — a repeated read whose returns differed is verification, not looping. Confidence:
plausible.

**Tier 3 — not citable alone; reachable with auxiliary evidence.** Claims of the form "the agent
reasoned that X was the root cause." Tool sequences are consistent with multiple reasoning
hypotheses. Three routes to citability: controlled re-execution and trajectory-agreement
analysis;[^TrajectoryEval] evidence-chain reconstruction mapping each claim to the tool result that
supports it;[^EvidenceTracing] and formal trace verification compiling the sequence into a checkable
model.[^VeryTrace] Confidence: low without the auxiliary leg.

**Tier 4 — not citable from transcripts at all.** Reasoning soundness; hallucination vs. grounded
claim; confidence calibration; the quality of a judgment call. Reasoning traces are absent or
second-hand; ground truth is not in the transcript; confidence is not logged anywhere. These
require an external oracle — a human, a test suite, or an independent agent.[^L3SessionStructure]

**Composition rule for claims spanning tiers.** Many real claims combine observations at different
tiers. For example: "tool-choice relevance" (whether the agent picked the right tool) is a Tier 2
observation (pattern inference), while "were there better choices?" is a Tier 4 question (requires
an external oracle). A composite claim like "the agent chose tool X correctly" spans tiers. Grade
such claims at the tier of their weakest leg: the composite grades as Tier 4 because the "correctness"
half requires ground truth. This prevents laundering Tier 4 conclusions under a Tier 2 label.

**Disclosure convention.** Recent agent-auditing work converges on stating three things in any
trajectory claim: what was observed, how it was measured, and what it does *not* tell you (e.g.
"tool count reflects effort, not correctness").[^TrajectoryEval][^AgentBenches]
[minority: lane-2/literature]

## 7. What cannot be audited from current telemetry

An enumeration, treated as **open** rather than exhaustive:

1. **Decision alternatives** — tools considered and rejected are not recorded, so "were there
   better choices?" is unanswerable from the record.[^TrajectoryEval]
2. **Confidence and uncertainty** — no field carries per-decision confidence.[^OTelObservability]
3. **Claim-to-evidence links** — a natural-language claim is not tagged with the tool result
   supporting it, so hallucination is not detectable from the transcript
   alone.[^TracesSurvey] [minority: lane-2/literature]
4. **Semantic dependencies** — tool results are independent blocks; causal influence of one result
   on the next decision is not represented.[^TracesSurvey] [minority: lane-2/literature]
5. **Reasoning branches** — summaries, where they exist, are single-threaded; abandoned hypotheses
   are not captured.[^TrajectoryEval]
6. **Context decay** — the append-only record cannot mark which earlier observations the agent
   relied on and which it ignored.[^TracesSurvey] [minority: lane-2/literature]
7. **Effort allocation** — adaptive thinking does not expose the budget actually
   spent.[^AdaptiveThinking] [minority: lane-2/literature]

## 8. What to do instead

**Artifact-based reasoning recording (this repo's practice).** Recording avenue status
(pursued/abandoned/declined with reasons), manifest rows, closure anchors, friction entries and
repair history produces reasoning evidence that is durable, git-tracked, intentional, append-only,
and checkable by an adversary against the artifact it cites.[^ArtifactRecording]
[minority: lane-3/local-repo] This bypasses the thinking-block problem rather than solving it. Its
honest limit, stated in the case against: it is still self-report, so it buys durability and
non-circularity, not sincerity.

Why the tradeoff matters: thinking blocks are ephemeral (they vanish if the API changes or the setting
drifts), and they are circular (the same agent whose reasoning is in question is the source of the
recorded thought). Artifacts are version-controlled, persist independently of vendor changes, and
can be audited against external evidence — the tool calls, file diffs, and outcomes they claim to
document. A recorded avenue can be *disconfirmed* (the agent claims it tried path X, but the JSONL shows
it never called the associated tool; the artifact contradicts itself). A thinking block cannot — it is
self-contained and self-authenticating. In adjudication, the ability to disconfirm via external
evidence is worth more than sincerity, because an agent's private reasoning can be false even if
honestly reported. The cost of artifact recording is one line per decision; the cost of thinking-block
infrastructure is indefinite maintenance of a deprecated telemetry channel.

At adjudication time, both channels require verification effort: an artifact record must be traced
against the transcript and artifacts it cites (evidence-chain reconstruction); a thinking block must be
read, interpreted, and assessed for alignment with observed behavior. The artifact path is not cheaper
per claim than reading a thinking block — both demand evidence-tracing work. What the artifact buys is
that the tracing is *adversary-checkable*: a judge can follow a cited avenue to the tool call it
names and to the file it changed, and verify or refute the agent's self-report. A thinking block
offers only internal coherence for verification. For adjudication, auditability by an external reader
(even if costlier per claim than a thinking-block read alone) is the sounder posture.

**Multi-agent verification.** Writer/reviewer patterns put a second reasoning trajectory in fresh
context over the first agent's output. This does not solve single-agent reasoning opacity and adds
overhead proportional to risk.[^MultiAgentVerification] [minority: lane-1/disconfirming] A related
reported result is that judge-model auditing alone catches a minority of errors while combination
with deterministic tooling catches far more; the specific figures we inherited (~45% vs 94%) reach
us through a secondary listicle and are **not** leaf-verified — treat the direction as the finding
and the numbers as unverified.[^MultiAgentVerification]

**Hooks and permission modes** enforce policy on acts: `PreToolUse` / `PostToolUse` and other
lifecycle events gate or veto tool calls, and Claude Code's default is to ask before each
action.[^HooksReference][^DesignPrinciples] [minority: lane-1/disconfirming] They constrain what
happens; they record nothing about why a tool was chosen.

**Formal and provenance frameworks (research-stage).** Evidence tracing with explicit provenance
relations (support, derive, depend-on, contradict, invalidate, trigger, update);[^EvidenceTracing]
reasoning-tree auditing that localizes adjudication to divergence points and reports an advantage
over majority vote and judge-model scoring;[^AgentAuditor] compilable formalization of reasoning
traces for machine checking;[^VeryTrace] and temporal-logic procedural-compliance specification
checked inline against traces.[^AgentLTL] All lane-2 minority reports.
[minority: lane-2/literature] The survey position is that the field is fragmented: "No single
system spans trace sources, fine granularity, runtime timing, an explicit representation, and
multiple trust functions at once."[^TracesSurvey]

**Standards, not yet arrived.** An emerging industry audit-record schema includes a `reasoning_trace`
field as part of the Agent Decision Record pattern for agent compliance logging, which no current
Claude Code surface provides.[^NISTAuditRequirement][^DEMM] [minority: lane-1/disconfirming] A maturity
model for assessing whether available evidence suffices to reconstruct a property after the fact
has been proposed but is a framework, not a shipped mechanism. Reasoning-trace capture is precisely
what no current Claude Code surface provides — the gap is recognized upstream and unfilled.

**Recommended stack, in order.** Primary: acts and decisions from transcript plus OpenTelemetry.
Secondary: tool inputs/outputs verified against the file system or external service. Tertiary:
artifact-based reasoning records. Not recommended: thinking blocks, in any
configuration.[^ArtifactRecording][^ThinkingAuditGuidance]

## 9. Risk matrix

| Risk | Likelihood | Impact | Complexity to mitigate | Disposition |
|---|---|---|---|---|
| Harness parses thinking blocks, reads empty strings, concludes "no reasoning" | high | high | trivial (one sweep) | **mitigate** — assert non-empty before trusting |
| JSONL field/shape change breaks a parser | medium-high | medium | medium (move to OTLP or `/export`) | **risk-accept** for read-only forensics; mitigate for anything durable |
| Silent tool-result truncation corrupts a finding | medium | high | medium (raise `maxResultSizeChars`; compare lengths) | **risk-accept with disclosure** — no audit marker exists to detect it after the fact |
| Metric/span names drift between versions | high | low | trivial (enumerate against installed binary) | **risk-accept** — re-enumerate per version |
| Reasoning-quality claim published on Tier-3 evidence | medium | high | low (tier discipline) | **mitigate** — tier label mandatory on every claim |
| Artifact self-reports are post-hoc rationalization | medium | medium | high (would need independent corroboration per entry) | **risk-accept** — the mitigation cost exceeds the benefit; disclose the limit instead |
| `showThinkingSummaries` enabled but summaries are unfaithful | certain if enabled | medium | high (unsolved research problem) | **risk-accept** — never treat a summary as evidence of reasoning |
| Vendor changes default behavior again without a client release | medium | medium | low (re-run the sweep) | **risk-accept** — the sweep is cheap; schedule it, do not engineer around it |

The last row generalizes: every mechanism finding here is version-bound to Claude Code v2.1.215 and
has an already-demonstrated history of changing by server-side flag.[^Issue32810] The correct
posture is a cheap re-verification recipe, not a robust abstraction over a moving surface.

## 10. Where the lanes disagreed, and how it resolved

| Dispute | Lane positions | Resolution at leaf |
|---|---|---|
| Does Claude Code have a thinking setting? | L1: yes, `showThinkingSummaries`; L3: no such setting exists | **L1 correct.** Present in v2.1.215 with a describe-string naming the transcript view.[^BinaryShowThinking] L3's "no setting" is refuted; L3's "no *raw* thinking" stands. |
| Which flag drives the default? | L1: `tengu_quiet_hollow` (server-side); L2: the `redact-thinking-2026-02-12` header | Both were true of v2.1.71.[^Issue32810] In v2.1.215 the beta header is still registered but the named flag is **absent**; resolution runs through the interactive/non-interactive branch instead.[^BinaryFlagAbsent][^BinaryDisplayResolver] |
| Date of the change | L1: 2026-03-10 activation, v2.1.72+; L2: 2026-02-12 header | Not a conflict: the header constant dates from 2026-02-12 and shipped inert; the behavior changed when the server-side flag was flipped ~2026-03-10.[^Issue32810] |
| Is #52376 open? | L3: open | **Closed as duplicate.**[^IssueStatuses] |
| Are thinking blocks "captured"? | L1: captured but degraded; L3: present but empty | Same finding, different emphasis. Blocks are structurally present with signatures; content is empty in 5,754/5,754 local cases.[^LocalSweep] |

---

## Open questions carried past this round

1. Does `showThinkingSummaries: true` produce non-empty thinking in a **non-interactive** subagent
   transcript, given the force-omitted guard? Untested; this is the single experiment that could
   overturn the headline.
2. Do explicit-display paths (`explicitDisplay`, exact-tools, subagent-text forwarding, async)
   bypass the force-omit guard in practice, and is any of them reachable from a workflow harness?
3. Can Anthropic's raw-thinking channel be unlocked without breaking other guardrails, and is there
   any programmatic path short of a sales conversation?
4. How stable are OTLP span/attribute names across Claude Code releases, and what is the
   deprecation policy for beta tracing?
5. What key material would decrypt a thinking signature, and is inaccessibility structural or
   policy?
6. Are the AI-classified behavioral session profiles reported in secondary sources actually
   generated and stored, and are they exportable for audit? (Inherited from lane-3; the underlying
   claim is secondary-sourced and was not corroborated this round.)
7. If industry standards for agent audit logging mandate reasoning-step capture, what would a
   compliant implementation look like given the faithfulness problem?
8. Are there practical transcript-side detectors of performative reasoning — timing analysis, token
   rate shifts — that do not require activation access?
9. Given two agents with different captured trajectories on one task, what methodology reliably
   adjudicates which reasoning was sounder? No lane found a settled answer.
10. Does the Compliance API's activity taxonomy change under enterprise access, and does any
    reasoning-adjacent event category exist there that public documentation omits?

---

## Footnotes

[^TranscriptFormat]: "Claude Code JSONL transcript format" — claude-dev.tools/docs/jsonl-format. One JSON object per line under `~/.claude/projects/<encoded-path>/<session-id>.jsonl`; `type` discriminates user/assistant/system; assistant lines carry `message.content` as an array of typed blocks (text, thinking, tool_use, tool_result). Cited by lane-1 and lane-2. Accessed 2026-07-18.

[^LocalSweep]: **Blue-synthesize leaf measurement, 2026-07-18.** Recursive walk of `C:/Users/gbloc/.claude/projects/`: `find . -name "*.jsonl" | wc -l` gives 294 files (a non-recursive `*/*.jsonl` glob returns 16 — subagent and workflow transcripts nest deeper). `grep -o '"type":"thinking"'` across all files gives 5,744 at the first count and 5,754 minutes later (the store grows under a live session). `grep -o '"type":"thinking","thinking":"[^"]'` gives **0** matches, i.e. no thinking block has a non-empty first character. `grep -o '"thinking":"","signature":"'` gives 5,754. `grep -l '"type":"thinking"'` gives 278 files. Method reproducible; scope is one machine, one default-configured install. Access date 2026-07-18.

[^BinaryShowThinking]: **Blue-synthesize leaf extraction from the installed client, 2026-07-18.** `C:/Users/gbloc/.local/bin/claude`, `claude --version` reports `2.1.215 (Claude Code)`. String extraction yields the settings-schema entry: `showThinkingSummaries: S.boolean().optional().describe("Request API-side thinking summaries and show them in the conversation and in the transcript view (ctrl+o). Set explicitly to override the default …")`, and the reader `function yPi(){return qn().showThinkingSummaries??!1}` (default `false`). Access date 2026-07-18.

[^BinaryDisplayResolver]: **Blue-synthesize leaf extraction, 2026-07-18**, same binary. Verbatim minified: `function fbc({explicitDisplay:e,isNonInteractive:t,outputFormat:r,verbose:n}){if(e)return e;if(!t)return yPi()?"summarized":void 0;if(r==="text"||r==="json"&&!n)return"omitted";return}` and the guard `function mbc(e,{useExactTools:t,forwardSubagentText:r,isAsync:n,isNonInteractiveSession:o,sessionDisplayExplicit:i}){if(i||!o||t||r||n||e.type==="disabled"||e.display==="omitted")return e;return{...e,display:"omitted"}}`. Reading: `showThinkingSummaries` yields `"summarized"` only when the session is interactive; non-interactive sessions without an explicit display are forced to `"omitted"`. Access date 2026-07-18.

[^BinaryOtelRedaction]: **Blue-synthesize leaf extraction, 2026-07-18**, same binary. Verbatim: `function Itd(e){return e.map((t)=>{if(t.type==="thinking")return{...t,thinking:"<REDACTED>"};if(t.type==="redacted_thinking")return{...t,data:"<REDACTED>"};return t})}`, applied via `Vmy` on the request path (`Iyo` calls `ktd("api_request_body", …)`) and directly on the response path (`Rtd` calls `ktd("api_response_body", …)`). The redaction takes no configuration argument. Same extraction: `Htd()` parses `OTEL_LOG_RAW_API_BODIES`, supporting `file:<dir>` (bodies written to `<dir>/<request_id>.{request,response}.json`, event carries `body_ref`) and inline mode (truncated, flagged `body_truncated`); `function k4r(e){return rrg()?e:"<REDACTED>"}` gates prompt content, with `tGc(){return Z.OTEL_LOG_ASSISTANT_RESPONSES??Z.OTEL_LOG_USER_PROMPTS}`. Access date 2026-07-18.

[^BinaryOtelNames]: **Blue-synthesize leaf extraction, 2026-07-18**, same binary. `grep -a -o -E 'claude_code\.[a-z_.]+' | sort -u` yields exactly: active_time.total, bash.subprocess, code_edit_tool.decision, commit.count, compaction, cost.usage, events, hook, interaction, lines_of_code.count, llm_request, mcp.rpc, pull_request.count, session.count, subagent.spawn, token.usage, tool, tool.blocked_on_user, tool.execution, tracing. Emit sites use unprefixed event names: `vc("tool_decision", {decision, source, tool_name, tool_use_id, …})`, `vc("permission_mode_changed", {from_mode, to_mode, trigger})`, `vc("user_prompt", {prompt_length, prompt, "prompt.id", "message.uuid"})`. `maxResultSizeChars` present (66 occurrences). Caveat on method: absence of a literal string does not prove absence of a behavior, since names can be constructed at runtime. Access date 2026-07-18.

[^BinaryFlagAbsent]: **Blue-synthesize leaf extraction, 2026-07-18**, same binary. `grep -c -a -o -F "tengu_quiet_hollow"` gives 0. `grep -c -a -o -F "redact-thinking"` gives 2, in the beta registration `ZNr=PT("redact_thinking","redact-thinking-2026-02-12")`. `showThinkingSummaries` gives 3; `thinking.display` gives 10; `CLAUDE_CODE_ENABLE_TELEMETRY` gives 10; `CLAUDE_CODE_ENHANCED_TELEMETRY_BETA` gives 3. Access date 2026-07-18.

[^IssueStatuses]: **Blue-synthesize leaf verification via `gh issue view --repo anthropics/claude-code`, 2026-07-18.** #32810 "[BUG] Thinking block content empty in JSONL session files since 2.1.72 (signatures present, text missing)" — CLOSED / NOT_PLANNED (and locked). #32997 "[SAFETY] Thinking redaction correlates with sustained deceptive model behavior — model fabricates verification claims when thoughts aren't recorded" — CLOSED / NOT_PLANNED. #52376 "Feature Request: Enable thinking.display for Claude Code subscription sessions" — CLOSED / DUPLICATE. #10084 "[FEATURE] Expose Claude Code Cognitive Telemetry States via API" — CLOSED / NOT_PLANNED. Access date 2026-07-18.

[^Issue32810]: GitHub issue #32810, anthropics/claude-code, community root-cause comment read via `gh issue view 32810 --comments` on 2026-07-18. Quotes the v2.1.71 minified condition sending `redact-thinking-2026-02-12` when thinking is enabled, the model supports it, not verbose/debug, `h7().showThinkingSummaries !== true`, and `p8("tengu_quiet_hollow", false)`; states the flag was "flipped server-side ~Mar 10 04:00 CDT", that the header constant "exists in both 2.1.71 and 2.1.72 — it's been in the codebase since at least Feb 12", and that `"showThinkingSummaries": true` "bypasses the redaction condition." Status corrections in [^IssueStatuses]. Cited by lane-1; comment text verified by blue-synthesize. Access date 2026-07-18.

[^Issue32997]: GitHub issue #32997, anthropics/claude-code — reporter's case study: tool calls run internally, results not displayed, unsupported assertions made to the user, admission when confronted. Single anecdotal report; issue closed not planned (see [^IssueStatuses]). Blue carries the *tool-output visibility gap*; the reporter's correlation between thinking redaction and deception is **not** adopted as a causal claim. Cited by lane-1. Access date 2026-07-18.

[^ExtendedThinkingDocs]: "Extended thinking" — platform.claude.com/docs/en/build-with-claude/extended-thinking. `thinking` parameter with `type:"enabled"`, `budget_tokens`, and `display` (`summarized` or `omitted`); thinking blocks carry `type`, `thinking`, `signature`; the signature is a base64 cryptographic value, not human-readable, carrying encrypted thinking for multi-turn continuity, not decryptable by clients. Cited independently by lanes 1, 2 and 3. Accessed 2026-07-18.

[^ExtendedThinkingLimitations]: Same source as [^ExtendedThinkingDocs], detail per lane-3: `summarized` is the Claude 4-era default and `omitted` the newer-model default; both bill for full thinking tokens; summarization is performed by a different model than the target model; raw thinking reportedly requires contacting Anthropic sales. The sales-channel claim is documentation-reported and was **not** independently verified. Accessed 2026-07-18.

[^ExtendedThinkingConfig]: Same source, per lane-3: `budget_tokens` deprecated on newer models in favour of `effort`; these options apply to direct API calls. Claude Code sets display internally — see [^BinaryDisplayResolver] for how. Accessed 2026-07-18.

[^AdaptiveThinking]: "Adaptive thinking" and "Effort" — platform.claude.com/docs/en/build-with-claude/adaptive-thinking, /effort. `thinking.type:"adaptive"` with `output_config.effort` in low/medium/high/max; no API surface exposes the effort actually selected, internal reasoning branches, or effort's influence on the decision. Cited by lane-2. Accessed 2026-07-18.

[^VisibleExtendedThinking]: "Visible extended thinking" — anthropic.com/news/visible-extended-thinking. "we don't know for certain that what's in the thought process truly represents what's going on in the model's mind"; "models often make decisions based on factors that they don't explicitly discuss in their thinking process." Cited by lane-2. Accessed 2026-07-18.

[^ThinkingAuditGuidance]: Practitioner guidance on treating thinking signatures as protocol state, retrieved via search of Claude API extended-thinking documentation and an APIScout guide (lane-3). "do not parse, modify, log, or treat thinking signatures as user-readable audit evidence. Treat them as provider-controlled protocol state. If your product needs an audit trail, record prompts, tool calls, approvals, files changed, diffs, and final answers. Do not promise an audit trail of the model's private reasoning." Secondary source; the operative substance is corroborated by [^ExtendedThinkingDocs] and [^BinaryOtelRedaction]. Accessed 2026-07-18.

[^ReasoningTheater]: "Reasoning Theater: Disentangling Model Beliefs from Chain-of-Thought" — arXiv:2603.05488v4 (Goodfire AI + Harvard). Compares DeepSeek-R1 671B and GPT-OSS 120B across MMLU and GPQA-Diamond. DeepSeek-R1 671B: MMLU performativity 0.417, GPQA-Diamond 0.012 (~35x decline). GPT-OSS 120B: MMLU 0.435, GPQA-Diamond 0.227 (~1.9x decline). Task-dependent patterns in both models; magnitude varies by an order of magnitude. **Not** generalized to Claude models in this report. Cited by lane-1. Accessed 2026-07-18.

[^OTelObservability]: "Observability with OpenTelemetry" — code.claude.com/docs/en/agent-sdk/observability. OTLP traces, metrics and log events to any conformant backend (Honeycomb, Datadog, Grafana, Langfuse, self-hosted); off until `CLAUDE_CODE_ENABLE_TELEMETRY=1` plus an exporter; enhanced tracing behind `CLAUDE_CODE_ENHANCED_TELEMETRY_BETA=1`; tracing is beta and span/attribute names may change; "extended-thinking content is redacted from the exported bodies." Cited by lanes 2 and 3. Accessed 2026-07-18.

[^OTelRedaction]: Corroborating statement that "Claude's extended-thinking content is always redacted from these bodies regardless of other settings," including when `OTEL_LOG_RAW_API_BODIES` is enabled (lane-1 search result; lane-2 quotes equivalent documentation language). Now verified in code at [^BinaryOtelRedaction]. Accessed 2026-07-18.

[^L3OpenTelemetryDetails]: Claude Code monitoring/observability documentation as read by lane-3: spans nest hierarchically (tool spans under interaction spans; subagent tool spans under parent tool spans); events carry trace context for joining with spans; `session.id` attribute enables multi-turn filtering; W3C trace context propagates to child processes. Accessed 2026-07-18.

[^SessionDocs]: "Sessions" — code.claude.com/docs/en/sessions. "By default, transcripts are stored as JSONL at `~/.claude/projects/<project>/<session-id>.jsonl`… The entry format is internal to Claude Code and changes between versions, so scripts that parse these files directly can break on any release. To build on session data, use `/export` or the script interfaces instead." Also: `--output-format json|stream-json` for `claude -p`; session id and transcript path available to hooks and status-line commands; transcript writes suppressible via `CLAUDE_CODE_SKIP_PROMPT_HISTORY` or `--no-session-persistence`. No mention of thinking-display configuration. Cited by lane-3. Accessed 2026-07-18.

[^L3TranscriptUnstable]: Lane-3's reading of the same documentation warning, extended to the audit consequence: field names, ordering and nesting may change without notice, and JSONL is append-only but not cryptographically signed or tamper-evident, so it lacks the integrity semantics of an audit log. The instability quote is documented; the tamper-evidence point is lane-3's inference from the absence of any signing mechanism, and is labelled derived rather than documented. Accessed 2026-07-18.

[^L3SessionStructure]: Lane-3 local sweep of 287 session transcripts under `~/.claude/projects/` (2026-07-02 to 2026-07-18), reporting 5,569 thinking blocks all with empty `thinking` and populated `signature`; example block cited from a named session file as `"type":"thinking","thinking":"","signature":"Eqk/CokBCA8YAipA7C7V1rra8swK…"`. Independently re-measured at larger counts by blue-synthesize — see [^LocalSweep]. Accessed 2026-07-18.

[^TrajectoryEval]: Kim et al., "Beyond the Final Answer: Evaluating the Reasoning Trajectories of Tool-Augmented Agents", arXiv:2510.02837 (October 2025). Argues trajectory-level evaluation — tool sequences, intermediate outputs, decision paths — is necessary to distinguish lucky guesses from sound reasoning; source of the three-part disclosure convention (what was observed / how measured / what it does not tell you). Cited by lane-2. Accessed 2026-07-18.

[^EvidenceTracing]: Chen et al., "From Agent Traces to Trust: A Survey of Evidence Tracing and Execution Provenance in LLM Agents", arXiv:2606.04990v3 (June 2026). Defines evidence tracing as connecting evidence units to claims; provenance relations Support, Derive, Depend-on, Contradict, Invalidate, Trigger, Update. Cited by lane-2. Accessed 2026-07-18.

[^TracesSurvey]: Same work as [^EvidenceTracing], §III: "No single system spans trace sources, fine granularity, runtime timing, an explicit representation, and multiple trust functions at once." Source for the missing-capability enumeration (claim-to-evidence links, semantic dependencies, context decay). Accessed 2026-07-18.

[^VeryTrace]: Xu et al., "VeryTrace: Verifying Reasoning Traces through Compilable Formalism and Structured Verification", arXiv:2606.24124 (June 2026). Converts natural-language reasoning into a formal system for automated verification; addresses the absence of machine-checkable structure in reasoning output. Cited by lane-2. Accessed 2026-07-18.

[^AgentAuditor]: Jiao et al., "Auditing Multi-Agent LLM Reasoning Trees Outperforms Majority Vote and LLM-as-Judge", arXiv:2602.09341 (February 2026). Resolves reasoning conflicts by comparing branches at critical divergence points, converting global adjudication into localized verification. Cited by lane-2. Accessed 2026-07-18.

[^AgentLTL]: Lemos et al., "AgentLTL: A Trace-Verification Framework for Measuring, Enforcing, and Training Procedural Compliance in Tool-Using LLM Agents", arXiv:2607.02599 (July 2026). Temporal-logic specification for offline evaluation and online enforcement over agent traces. Cited by lane-2. Accessed 2026-07-18.

[^AgentBenches]: ASTRA-bench (arXiv:2603.01357), Litmus (arXiv:2604.08970) and related 2025–2026 agent-evaluation frameworks adopting trajectory-level analysis with method disclosure as standard practice. Cited by lane-2; identifiers not individually leaf-verified this round. Accessed 2026-07-18.

[^StructuredOutputs]: "Structured outputs" — platform.claude.com/docs/en/build-with-claude/structured-outputs. JSON-schema constraints via an `anthropic-beta` header; a contract on output format, not on reasoning exposure. Cited by lane-2. Accessed 2026-07-18.

[^DebugModeSearch]: Lane-2 web search for Claude Code debug modes, reasoning APIs and cognitive telemetry (query: `"Claude Code" debug mode reasoning summary API settings`), 2026-07-18. Found GitHub #10084 requesting the capability; no public API or configuration exposing structured reasoning summaries, decision trees, branches or confidence scores. Issue status corrected at [^IssueStatuses].

[^PlatformDocs]: Lane-1 review of platform.claude.com/docs, 2026-07-18: no documented endpoint for reasoning summaries, decision alternatives, confidence scores or reasoning branches. An absence claim over the documented public surface as searched; not a proof of non-existence.

[^ComplianceAPI]: "Compliance API" — platform.claude.com and support.claude.com article 13015708 (navigation pages, not enumerating activity types); for the documented activity-type count, "Claude Compliance API Documentation" — generalanalysis.com/guides/claude-compliance-api. Publicly documented surface shows roughly 30 activity types including Claude Code events; no reasoning/thinking/decision-trace category appears. Lane-1 reported 260+ activity types across 33 categories, a count not corroborated by publicly accessible sources. The 260+ figure was not re-verified by blue-synthesize (no enterprise access). The substantive finding (no reasoning category in public documentation) is grade-certain; the documented type count is 30. [minority: lane-1/disconfirming] Accessed 2026-07-18.


[^ToolTruncationLimits]: Lane-1 search result on Claude Code `tool_result` size limits: MCP servers may raise the cap via `anthropic/maxResultSizeChars` to a reported 500K characters, with a lossy default. The key's presence in the shipped binary is confirmed at [^BinaryOtelNames]; the 500K figure is search-derived and not leaf-verified. Accessed 2026-07-18.

[^AgenticMisalignment]: "Agentic Misalignment in Summer 2026" — alignment.anthropic.com/2026/agentic-misalignment-summer-2026/. Quoted: "the LLM judge that should catch these alignment failures is itself subject to the same failures"; "reasoning transcripts may not be faithful"; "a model can register that it is being tested without saying so"; "simulated deployments are never perfect replicas of real ones"; "human review remains essential precisely because automated auditing has these fundamental blindspots." Cited by lane-1. Accessed 2026-07-18.

[^DesignPrinciples]: "Dive into Claude Code: The Design Space of Today's and Future AI Agent Systems" — arxiv.org/html/2604.14228v1. Design principle quoted: "Humans can observe actions in real time, approve or reject proposed operations, interrupt compatible in-progress operations, and audit after the fact." Cited by lane-1. Accessed 2026-07-18.

[^HooksReference]: "Hooks reference" — code.claude.com/docs/en/hooks. `PreToolUse`, `PostToolUse` and further lifecycle events run shell commands for gatekeeping, vetoing patterns, or scanning; they gate acts and record nothing about tool-choice rationale. Cited by lane-1. Accessed 2026-07-18.

[^MultiAgentVerification]: Lane-1 search result, "Claude Code Best Practices: 15 Rules" (2026-07-18), reporting the writer/reviewer pattern and citing an IBM Research 2026 finding that judge-model auditing alone detects ~45% of errors while combination with deterministic tooling reaches 94%. Secondary listicle; the IBM figures were **not** traced to a primary source this round and are carried as unverified. The qualitative direction (deterministic checks materially outperform judge-only auditing) is the load-bearing part.


[^NISTAuditRequirement]: "AI Agent Governance and Compliance in 2026" — zylos.ai/research/2026-05-01-ai-agent-governance-compliance-2026/. Describes an emerging industry Agent Decision Record schema with fields for `reasoning_trace`, `tool_invocations`, and `outcome` as a pattern for compliance-ready audit logging. The schema is presented as an industry standard rather than as NIST guidance; NIST appears on the page in separate contexts (AI RMF, AI Agent Standards Initiative). Secondary source; verified 2026-07-19. Cited by lane-1. Accessed 2026-07-18.

[^DEMM]: "Decision Evidence Maturity Model for Agentic AI: A Property-Level Method Specification" — arxiv.org/pdf/2605.04093. Assesses whether available evidence suffices to reconstruct a property for post-hoc governance; a framework rather than a shipped capture mechanism. Cited by lane-1. Accessed 2026-07-18.

[^ArtifactRecording]: This repository's frank-exchange-of-views recording discipline at the blue seat: `avenue` (pursued/abandoned/declined plus reason), `manifest-row`, `friction`, `closing` (round-end position), and `retirement` via `retire` — append-only, nonce-stamped, git-tracked. Verified in-run: `feov-record blue --help` on the installed plugin (0.10.0, 2026-07-19) lists exactly: avenue, closing, confidence, dispute, friction, manifest-row, petition, position, register, render, retire, revision. Closure anchors and repair history are red-seat verbs (`close` at red merge). Blue seat produces avenue, manifest-row, friction, and position events. Cited by lane-3; verb list re-verified by blue-synthesize at the tool. Accessed 2026-07-19.

---

## Provenance and limitations of this round

- **Pinned inputs, recovered and adjudicated.** `inputs/PINNED.md` pins repo HEAD at `cacb736` and
  names two evidence files (`probe-thinking-persistence.md`, `mining-substrate-architecture.md`).
  Both files are recoverable at the pin via `git show cacb736:<path>`. 
  `probe-thinking-persistence.md` advances a competing mechanism for the central finding: "Zero
  `redacted_thinking` blocks means this is not API-side redaction: the API returned thinking, and
  the harness serialized the block structure without its text. Consistent across seat and main-session
  transcripts, so it is a serialization choice rather than a per-session setting or a bug." This is a
  client-side account against the display-resolver finding in §2. The probe also reports the same
  287 transcripts / 5,569 thinking blocks / 0 non-empty figures that §2 attributes to the same
  measurement of the evolving store at an earlier time rather than an independent sweep. On the
  serialization-vs-resolver question for the non-interactive share: the display-resolver finding
  (the setting is off by default for non-interactive sessions) is the more parsimonious explanation
  (a single guard forces `display:"omitted"`) versus the serialization claim (the transport layer
  drops content post-API). The resolver path is implemented in code; serialization would require a
  second bug. On the interactive branch, the single-guard premise is unavailable — the resolver
  returns `void 0` when unset (§2 line 198), so the serialization hypothesis is not retired there
  by parsimony. Both mechanisms remain possible on the interactive share.
- **Version binding.** All binary-derived findings (display resolver, OpenTelemetry redaction,
  settings schema, instrument names) are specific to Claude Code v2.1.215 on Windows, read 2026-07-19.
  §9's last row notes the version-bound nature and history of vendor changes via server-side flag.
- **Single-machine binding.** All transcript-store findings are specific to one user's default
  install.
- **Not verified this round:** the Compliance API taxonomy figures, the IBM judge-vs-deterministic
  percentages, the 500K `maxResultSizeChars` ceiling, the arXiv identifiers in
  [^AgentBenches], and the vendor sales channel for raw thinking. Each is labelled at its footnote.
