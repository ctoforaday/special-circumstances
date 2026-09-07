# Research Topic: Automating the NextToken AI Code Audit Rubric as a hook-driven Claude Code plugin

> **Status:** Research topic / prompt — to be investigated. This is an input
> specification (see the `_input` convention used by other research prompts in
> this directory, e.g. `anti_spinning_research_input.md`). The deliverable is a
> research report produced from the prompt below.
>
> **Source document:** The full NextToken AI Code Audit Rubric (PASS 0–6,
> including References/footnotes) referenced by this prompt is captured in
> [`nexttoken_audit_rubric_source.md`](nexttoken_audit_rubric_source.md).
> Original (may require access):
> <https://www.perplexity.ai/computer/a/97317781-4c97-4de0-88e9-e7c28186c99a>.
> Step 1 of the task below operates on that document.
>
> **Note:** the rubric defines **PASS 0 through PASS 6** (seven passes). PASS 6
> (Iterative Regression Audit) is the "feedback loop security degradation"
> content that step 5 below targets with the iteration-degradation hook.

## Context

We have an audit rubric ("NextToken AI Code Audit Rubric") that defines an
ordered sequence of audit passes (PASS 0 through PASS 6) for detecting
architectural flaws, async/state-management defects, and security
vulnerabilities characteristic of AI-generated code. We want to automate this
rubric and incorporate it into a development workflow. **Hook automation is the
most important dimension** — the design should be optimized around what can be
enforced automatically via hooks.

## Task

**Turn the NextToken AI Code Audit Rubric into an automated, hook-driven Claude
Code plugin.**

The source report is the full NextToken AI Code Audit Rubric (PASS 0 through
PASS 6, including any References/footnotes).

Primary objective: determine how best to automate this audit and incorporate it
into a development workflow. Hook automation is the most important dimension —
optimize the entire design around what can be enforced automatically via hooks.

Do the following:

1. **Decompose the rubric into atomic checks.** Go through every PASS and every
   numbered Step. For each check, extract: what it inspects, the exact defect
   signature it hunts for, and the pass/fail (or flag) condition. Produce a
   normalized inventory (one row per atomic check) — don't summarize the passes,
   enumerate their contents.

2. **Score each check for automatability.** Classify every check into:
   (a) fully deterministic — decidable by regex, AST/static analysis, dependency
   scan, git inspection, or secret-scanning, with no judgment; (b) LLM-assisted
   — needs a model/subagent to judge (e.g. "does this abstraction relocate vs.
   encapsulate complexity", orphan-state tracing, architectural drift);
   (c) manual-only — genuinely needs a human. Justify each classification and
   note the tool/technique that would implement it.

3. **Design the hook automation layer (core of the deliverable).** Research the
   current Claude Code / Agent SDK hook system — every hook event (PreToolUse,
   PostToolUse, UserPromptSubmit, Stop, SubagentStop, SessionStart, PreCompact,
   Notification, etc.), their input/output/exit-code contract, and how each can
   allow, block, warn, or modify. Then map the atomic checks onto hooks:
   - Which checks fire on **PostToolUse** after an Edit/Write (catch defects the
     moment code is generated)?
   - Which belong on **PreToolUse** as guardrails that block insecure actions
     before they happen (e.g. writing a hardcoded secret, string-concatenated
     SQL)?
   - Which run on **Stop / SubagentStop** as a full audit sweep before a session
     or subagent finishes?
   - Which run on **SessionStart** (PASS 0 inventory / iteration-depth estimate)
     or **UserPromptSubmit** (detect the four degrading prompting styles —
     EF/FF/SF/AI)?

   For each proposed hook give: event, matcher/trigger condition, what it
   inspects, the action (allow/block/warn/annotate), exit-code and JSON-output
   behavior, and a config + script sketch (shell/AST/subagent call).

4. **Handle the ordered-PASS nature.** The rubric is a strict sequence (PASS 0
   first, etc.). Explain how to preserve ordering under an event-driven hook
   model — e.g. gating passes, state carried in a session file, or a dispatcher
   hook that runs the next eligible pass.

5. **Address the report's own core finding.** It documents "feedback loop
   security degradation" (vulnerabilities rising with each refinement cycle).
   Design a hook that tracks iteration count / re-audits changed surfaces on each
   cycle so degradation is caught continuously rather than at the end.

6. **Non-hook fallbacks.** For checks that can't be hooks, specify the right
   plugin primitive instead (slash command for an on-demand full audit, subagent
   for deep architectural review, MCP tool for external scanners) — but keep
   these secondary to the hook layer.

7. **Deliverable:**
   - Atomic-check inventory table (check -> PASS/Step -> automatability class ->
     mechanism).
   - Hook design spec (the table + per-hook detail from step 3).
   - Plugin file/directory layout (manifest, hooks/, scripts, agents, commands).
   - Ordering/state strategy and the iteration-degradation hook.
   - A minimal working plugin skeleton (manifest + at least two representative
     hook scripts: one deterministic PreToolUse guardrail, one LLM-assisted
     Stop-sweep) as a proof of concept.
   - Open questions, false-positive risks, and version-dependent caveats.

Verify all hook capabilities against current official documentation and flag
anything uncertain or version-specific.
