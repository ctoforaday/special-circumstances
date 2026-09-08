# Reasoning telemetry — what is observable, and how to expose thinking summaries

> STATUS 2026-09-02: shipped — historical record (research complete; its consumer, Phase 3 of
> `gray-area.md` §7, is not built, and #126's verification debt is still open).

> Design research for **Gray Area** (the fourth plugin). Companion to [`gray-area.md`](gray-area.md).
> Status: research complete, no code. Supersedes one load-bearing finding in PR #3.

**Client under test:** Claude Code **2.1.220**, native binary (ELF, 275 MB),
`BUILD_TIME 2026-07-24T22:17:45Z`, `GIT_SHA 4073f59596e272f39393db4f96abc5f4b10eff21`.
Every claim below is read out of that binary or measured against a live run on it. Where a claim
is inherited rather than re-verified, it says so.

---

## 1. The finding that changes the design

PR #3's feasibility probe concluded that reasoning is unavailable:

> *"The reasoning was sent to the API and **not retained on disk** — only an opaque signature
> survives… So 'review the entire thought process' is not available from persisted trajectories.
> Any design resting on it is built on sand."*

and the 2026-07-18 debate run hardened it with a corpus sweep — 294 transcripts, 5,754 thinking
blocks, **zero** non-empty. The operator left one avenue open and untested: *"whether thinking
retention is controllable at agent spawn time."*

*(The shipped run's two now-refuted claims are filed as **#125**. The sweep itself is accurate and
is not among them — it measured a default, and the correction is to what was concluded from it.)*

**It is controllable, and the corpus sweep was measuring a default rather than a ceiling.**

Claude Code accepts a documented command-line flag:

```text
--thinking-display <display>    How thinking content appears in the response
                                (summarized | omitted)
```

Set it, and thinking summaries are written to the session transcript as non-empty `thinking`
blocks — **including in headless (`--print`) runs, and including in subagent transcripts.**

### The controlled measurement

Same prompt, same forced thinking budget, same machine, same client, one variable:

| Run | Command | `thinking` blocks | non-empty |
|---|---|---|---|
| control | `MAX_THINKING_TOKENS=4000 claude -p "<prompt>"` | 1 | **0** |
| treatment | `MAX_THINKING_TOKENS=4000 claude -p --thinking-display summarized "<prompt>"` | 1 | **1** |

Treatment sample, read back out of the transcript JSONL:

> *"I see this is a riddle about animals. When it says 'all but 9 run away,' that means 9 stay
> behind. Then someone buys twice as many as remain, which is 18 more animals, bringing the total
> to 27. After selling 5, there are 22 left."*

**Subagent propagation, measured separately.** A headless run carrying the flag that spawned a
`general-purpose` subagent produced non-empty thinking in **both** transcripts — the parent, and
`<session>/subagents/agent-<id>.jsonl` (`attributionAgent: general-purpose`). This is the case that
matters: in this suite every debate seat is a subagent, and the sweep's all-empty result was
dominated by exactly those files.

### Why the default looked like a ceiling

The resolver, deminified from 2.1.220 (identical in structure to the 2.1.215 reading in the
2026-07-18 run, which stands):

```js
function showThinkingSummaries() { return settings().showThinkingSummaries ?? false }

function resolveDisplay({explicitDisplay, isNonInteractive, outputFormat, verbose}) {
  if (explicitDisplay) return explicitDisplay;                    // ← the flag wins outright
  if (!isNonInteractive) return showThinkingSummaries() ? "summarized" : undefined;
  if (outputFormat === "text" || (outputFormat === "json" && !verbose)) return "omitted";
  return;
}

function subagentThinkingConfig(cfg, {useExactTools, forwardSubagentText, isAsync,
                                      isNonInteractiveSession, sessionDisplayExplicit}) {
  if (sessionDisplayExplicit || !isNonInteractiveSession || useExactTools ||
      forwardSubagentText || isAsync || cfg.type === "disabled" || cfg.display === "omitted")
    return cfg;
  return {...cfg, display: "omitted"};                            // ← forced omit for subagents
}
```

Two gates, and the flag opens both. Passing `--thinking-display` sets a session-wide
`thinkingDisplayExplicit` latch; `sessionDisplayExplicit` is the **first** disjunct in the subagent
path, so the forced-omit is skipped and the child inherits `summarized`. Without the flag, a
non-interactive session lands on the `"omitted"` branch and every subagent is forced there too —
which is precisely the population the sweep counted.

The interactive path has its own switch, `showThinkingSummaries`, a settings boolean described in
the binary as: *"Request API-side thinking summaries and show them in the conversation and in the
transcript view (ctrl+o). Set explicitly to override the default for your install."* Present and
defaulting to `false`. The `tengu_quiet_hollow` server flag named in issue #32810 remains **absent**
from the binary at 2.1.220, as it was at 2.1.215.

**What PR #3 got right and should keep:** the corpus was empty, the signature is opaque and
undecryptable, and no configuration anywhere yields *raw* thinking. What changes is only this: the
emptiness was a default this suite never overrode, not a property of the platform.

---

## 2. The three channels, graded

| Channel | Carries acts | Carries reasoning | Verdict |
|---|---|---|---|
| Session transcript JSONL | Complete | **Summaries, if the flag is set** | The substrate |
| OpenTelemetry export | Complete | **No content observed** — see the bounds below | Metrics, not evidence; **carries operator identity** (#834) |
| Repo artifacts (`feov-record` events) | Declared acts | Declared reasoning | Citable; still self-report |

### Transcript

`~/.claude/projects/<encoded-cwd>/<session-id>.jsonl`, one JSON object per line; subagents nest
under `<session-id>/subagents/`. Directly citable: every `tool_use` with full input, every
`tool_result` linked by `tool_use_id`, the `parentUuid` causal chain, per-event ISO timestamps,
`message.usage` token counts, model identifier, `effort`, `attributionAgent` for seats, and all user
text. This is the Tier 1 surface and nothing here has changed.

Unchanged caveats, both still live: the format is vendor-internal and version-unstable, and the file
is append-only but **not signed or tamper-evident** — it is an evidence base, not an audit log.

### OpenTelemetry

Twenty instruments ship in 2.1.220, byte-identical to the 2.1.215 enumeration:

```text
claude_code.interaction        claude_code.llm_request      claude_code.tool
claude_code.tool.execution     claude_code.tool.blocked_on_user
claude_code.hook               claude_code.subagent.spawn   claude_code.compaction
claude_code.mcp.rpc            claude_code.token.usage      claude_code.cost.usage
claude_code.code_edit_tool.decision                         claude_code.session.count
claude_code.lines_of_code.count               claude_code.active_time.total
claude_code.commit.count       claude_code.pull_request.count
claude_code.bash.subprocess    claude_code.events           claude_code.tracing
```

**No conversation content of any kind was observed in the export.** The mechanism is a map over
assistant content blocks replacing `thinking` and `redacted_thinking` with `<REDACTED>`, applied on
both request- and response-body export paths, with no environment variable or verbosity level
upstream of it. `--thinking-display summarized` does **not** open this channel — it changes what the
transcript keeps, not what the exporter emits.

**What is observation and what is still inference (2026-09-08, client 2.1.260).** The sentence above
originally rested entirely on string extraction from the binary; #126 was filed against exactly that.
Part of it has since been observed and part has not, and the two must not be read as one:

- **Observed — the `<REDACTED>` substitution firing.** Captured API bodies via
  `OTEL_LOG_RAW_API_BODIES=file:<dir>` across four runs, three models and both display modes: every
  response body carried `content[].thinking = "<REDACTED>"`, the client's own sentinel.
- **Observed — the export carries no conversation content.** A local OTLP collector on loopback
  received metrics and log payloads (15–19 KB) containing `claude_code.events`,
  `hook_execution_start`/`_complete`, `hook_registered`, `plugin_loaded` and their attributes.
  Neither the user prompt nor the assistant's answer appeared, **with `OTEL_LOG_USER_PROMPTS=1`
  set**. No content blocks of any kind.
- **NOT observed — the positive-control test #126 asks for.** A run where thinking text is known
  present *and* the export is captured has not been achieved: headless runs and a pty-driven
  interactive run all produced empty `thinking` blocks, so none of them could serve as the control.
  Non-empty thinking demonstrably exists in other sessions on this machine (79 blocks in one, 35 of
  them at 2.1.260), so the control is constructible; the configuration that produces it has not been
  reproduced on demand. **Until it is, "never, in any configuration" remains stronger than the
  evidence** — what is established is that nothing was seen on the paths tested.

**And reasoning was never the only disclosure surface.** The export carries `user.email` in
plaintext, `organization.id`, `user.account_uuid`, `user.account_id`, `session.id`, OS and host
attributes, by default. That is observed, and it is filed as **#834**. Any judgement about whether
this feed is safe to send off-box has to answer for identity regardless of how the reasoning
question resolves — a clean result on `thinking` would not settle it.

Consequence for Gray Area: OpenTelemetry is the right feed for cost, wall-clock, spawn topology and
compaction counts, and the wrong feed for anything reasoning-shaped. Do not build a reasoning
capability on it.

### Artifacts

The suite's existing answer — `manifest-row`, `avenue`, closure anchors, `friction` — remains the
only channel where reasoning is *citable*, because an artifact can be disconfirmed against the
trajectory that is supposed to support it and a thinking block cannot. Summaries do not replace it.

---

## 3. What a summary is, and what it is not

Stated plainly so no downstream document has to rediscover it:

1. **It is written by a different model than the one reasoning.** It is second-hand narration of
   the reasoning, not the reasoning.
2. **It is self-report.** Anthropic's own position is that "we don't know for certain that what's in
   the thought process truly represents what's going on in the model's mind," and that models "often
   make decisions based on factors that they don't explicitly discuss in their thinking process."
3. **Therefore it does not promote a finding.** Under the soundness tiers already agreed in the
   2026-07-18 report, a claim of the form *"the seat reasoned that X"* remains **Tier 3** — not
   citable alone — whether or not a summary is present. A summary is an additional self-report
   channel, not an oracle. A seat that intends to mislead can produce a misleading summary as easily
   as misleading prose.
4. **It is not free of the adjudication trap.** Feeding summaries to a judge agent reintroduces
   exactly the defect the attestation audit was built to remove.

**So what is it good for?** Three things, all *exploration*, none *adjudication* — the distinction
the operator already drew on PR #3, applied here:

- **Hypothesis generation.** "Which seats look like they went in circles?" is cheap to ask of a
  summary corpus and expensive to ask of raw acts.
- **Act-versus-claim triangulation.** A summary that names an intention the trajectory never
  executed is a *lead*, and the act record adjudicates it. The summary points; the tool calls rule.
- **Friction mining.** The 2026-07-18 run found a seat logging a friction it had already recovered
  from. Summaries around that moment are the cheapest route to the pattern, with the trajectory
  still deciding the finding.

**Rule to carry into the plugin, unchanged from the operator's framing:** *exploration may
summarize; adjudication must cite.*

**And the line is held where the adjudication happens, which is FEOV — not in Gray Area's read
path** (gblock, 2026-09-08). An earlier draft of this section put the guard in the miner, and that
was the wrong altitude twice over. Gray Area cannot see what a caller intends to do with a row, so
a refusal there is a convention wearing a structure's clothes — the exact thing T2 says not to
build. FEOV's citation ledger already refuses an unanchored closure, and that is a real structure
at the point where a ruling is actually made. Two enforcement points for one rule is how they drift
apart.

So Gray Area **mines summaries freely** — recovering an agent's intent is an explicit goal of the
plugin, not a tolerated side effect — and the evidence chain is FEOV's to defend.

---

## 4. Recommended capture posture

**NOT DONE (2026-09-02):** no surface adopts either switch — `--thinking-display` appears nowhere
under `plugins/` or `scripts/`.

| Surface | Setting | Rationale |
|---|---|---|
| Debate runs (`/research`) | Launch with `--thinking-display summarized` | Seats are subagents; this is the only way their reasoning is retained at all |
| Interactive development | `"showThinkingSummaries": true` in settings | Same content, interactive path; also enables the ctrl+o transcript view |
| OpenTelemetry | Enable for cost/latency/topology. **On-box or trusted collector only** | No conversation content was observed in the export, but it publishes operator identity — email, org and account ids (#834). Sending it to a third-party backend is a disclosure decision, not a metrics decision |
| Adjudication input (FEOV) | Acts and artifacts only | Summaries are excluded from the evidence chain **by FEOV's citation ledger**, at the point of ruling — not by withholding them from the miner. Gray Area serves summaries; a summary simply anchors nothing |

**Cost.** Thinking tokens are billed in full whether or not the summary is displayed, so the flag
does not change what the reasoning costs. It adds the summary tokens themselves, which are a small
fraction of the thinking budget. Not measured here — measure it on the first instrumented run before
quoting a number.

**Storage.** Summaries make transcripts materially larger. The 2026-07-18 run already produced 6 MB
trajectory files with reasoning *absent*. Size the retention policy before turning this on across a
long run, and keep the aggregate-only probing discipline — never pull a corpus into a context.

---

## 5. Risks

| # | Risk | Mitigation |
|---|---|---|
| T1 | **The flag is version-bound.** It is a documented CLI option, but the resolver around it has moved once already (2.1.71 → 2.1.215). | Probe it at run start, record the client version in the run record, and degrade to acts-only when absent. Never assume a summary exists because a previous run had one. |
| T2 | **Summaries get treated as evidence** once they are present and readable. This is the real risk; availability creates temptation. | Enforce structurally, not by convention — **in FEOV, at the point of ruling**: the citation ledger refuses an unanchored closure, and a summary anchors nothing. NOT in the mining tool (gblock, 2026-09-08): Gray Area cannot know what a caller will do with a row, so a refusal there would be the convention this row exists to reject. |
| T3 | **Adaptive thinking may skip thinking entirely** on easy turns — the first measurement here produced zero blocks until the budget was forced. Absence of a summary is not evidence of absence of reasoning. | Record whether thinking was configured and at what budget, so "no summary" is distinguishable from "no thinking". |
| T4 | **Transcript growth** inflates both storage and any naive read path. | Aggregate-only probes; size checks before retention decisions. |
| T5 | **Interactive and headless paths diverge** — two different switches, easy to set one and assume both. | `/doctor` reports both, and the run record states which path was used. |

---

## 6. Validation loop

The commands that prove this document, in order. Re-runnable on any box with the client installed.

1. **Flag exists on the installed client**
   `claude --help | grep -- --thinking-display` → prints the option and its two values.
2. **Control** — in a scratch directory:
   `MAX_THINKING_TOKENS=4000 claude -p "<multi-step reasoning prompt>"`
   then count `thinking` blocks in the session JSONL → **≥1 block, 0 non-empty**.
3. **Treatment** — fresh scratch directory:
   `MAX_THINKING_TOKENS=4000 claude -p --thinking-display summarized "<same prompt>"`
   → **≥1 block, ≥1 non-empty**.
4. **Subagent propagation** — fresh scratch directory, prompt that forces an Agent/Task call, flag
   set → non-empty thinking in `<session>/subagents/agent-*.jsonl`.
5. **OpenTelemetry stays closed** — stand up a collector, capture the export, and confirm no
   conversation content reaches it. **Re-arms on:** any client upgrade, and any change to the
   `OTEL_LOG_*` family. Reading the binary does not satisfy this step; that is what #126 is about.

Steps 1–4 were run on 2.1.220 and passed as recorded in §1.

**Step 5, partially closed 2026-09-08 at client 2.1.260.** The `<REDACTED>` substitution was
observed firing on the response-body path, and a loopback OTLP collector received payloads carrying
hook lifecycle events and identity attributes and **no conversation content** — no prompt, no
answer, no thinking — with `OTEL_LOG_USER_PROMPTS=1` set. What is still open is the positive
control: a captured export from a session whose thinking text is known to be non-empty. Every run
attempted (headless across three models and both display modes, plus a pty-driven interactive
session) produced empty `thinking` blocks, so none of them could distinguish "the export redacted
it" from "there was nothing there". **An empty export is only evidence when something was there to
find**, which is the whole reason #126 specifies a control.

Tracked in **#126**, which carries the closing procedure. The identity surface the same measurement
turned up is **#834**, and it is independent: #126 closing green would not resolve it.

---

## 7. Open questions

- **Do summaries survive compaction?** A `PostCompact` hook receives the summary (see
  [`gray-area.md`](gray-area.md) §3) — for *reading* only; it cannot inject
  ([`hook-surface-spike.md`](hook-surface-spike.md) §3a). Whether pre-compaction thinking blocks
  remain readable in the transcript after the boundary is still unmeasured, and is the question
  that matters here: mining reads the transcript, not the live context, so the answer is
  independent of what any hook can inject.
- **Is there a per-agent override?** The latch is session-wide. Whether an individual subagent can
  be spawned with a different display than its parent was not found in the binary and is assumed
  not possible.
- **What does the summarizer cost?** Unmeasured (§4).
- **Raw thinking** remains unavailable through every public surface checked, and the signature
  remains undecryptable by clients. No change; not expected to change.
