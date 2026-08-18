---
description: Report acts this session did more than once — the same file written again, the same command re-run — and acts repeated 3+ times BACK-TO-BACK, the [[anti-spinning]] limit adjudicated against the trajectory rather than the hook counter. `--json` for rows; pass a transcript path to inspect a different session.
---

Report repetition from the trajectory. Model [[terse-communication]]: relay the binary's rows, add no interpretation of your own.

The grouping logic lives in the tested binary, not in this prompt. Your job is to run it and report what it says.

1. Run `${CLAUDE_PLUGIN_ROOT}/bin/gray-area` (`.exe` on Windows) as `gray-area stalls` for the [[anti-spinning]] question, or `gray-area rework` for every repeated act. Pass `--json` straight through if the caller gave it. With no transcript argument the binary resolves this session's from gray-area's own manifest and prints which row it used.
2. Relay the rows verbatim, including the coverage line at the end.
3. Both exit 0. **Repetition is not a failure**, and neither verb is a gate.

**What a row says, and what it does not.**

- A row reports that an act happened again, and cites every occurrence with file, line and uuid. **It does not say the repetition was waste.** A file written five times may be five careful increments; a suite run five times may be five different fixes each verified.
- `stalls` counts **consecutive** repetition, because [[anti-spinning]]'s rule is about a repair loop and not about ordinary iteration. The same check run three times across a session is normal work; three times in a row is the shape the rule is about.
- **The trajectory records what was RUN, never what the run SAID.** Result bodies are conversation content and this plugin does not copy them. So whether three identical invocations were three failures or three deliberate re-runs is **NOT MEASURED**, and the tool prints that on every listing. YOU MUST NOT report a `stalls` row as evidence the agent was spinning — the record cannot support it.

**Read the coverage line.** It states how many citable tool uses were searched, how many could not be keyed, and how many acts happened exactly once. An empty listing with that line is a clean session; an empty listing without it would be indistinguishable from a tool that read nothing. If most events could not be keyed, say so rather than reporting a quiet board.

**Why this exists alongside the strike counter.** prosthetic-conscience's `PostToolUseFailure` hook counts strikes deterministically and is documented as blind to most of the class: it counts TOOL failures only, so a command that exits 0 while the work still failed, an edit that applies cleanly and does not fix the bug, and a test runner reporting failures in its own output all pass it silently. This reads the trajectory instead, where a repetition is a repetition whatever the exit code was.
