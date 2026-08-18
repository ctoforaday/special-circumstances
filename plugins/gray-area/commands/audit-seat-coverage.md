---
description: Does the manifest name every seat transcript that exists? Reconciles the recorded seat rows against the session's own seat-transcript directory. Exits 1 when it could NOT measure, 0 when it did — whatever it found.
---

Reconcile recorded seats against seats on disk. Model [[terse-communication]]: relay the binary's rows, add no interpretation of your own.

1. Run `${CLAUDE_PLUGIN_ROOT}/bin/gray-area` (`.exe` on Windows) as `gray-area coverage`. It resolves this session's transcript from gray-area's own manifest and derives the seat-transcript directory from it.
2. Relay the output verbatim.

**Read the exit code as written.** `1` means it could **not** measure — not that it found problems. `0` means it measured, whatever it found. That split is the point: unnamed transcripts are a finding for a human to act on, while an unmeasurable board is a broken instrument, and an instrument reporting a clean board when it cannot see is the failure this plugin exists to prevent.

**Two directions, and only one was ever established.**

- **No phantom seats** — every seat row names a file that exists. Confirmed at 19/19 (#189).
- **No missed seats** — every transcript on disk is named by a row. **Not established**: measured against transcripts rather than rows, the hook saw 19 of 20 (#469).

**An `UNNAMED` row is a seat nothing in the manifest can lead a reader to.** YOU MUST NOT report it as a bug in the harness or as a lost subagent. Two benign explanations exist and neither is measured: the subagent may have been killed before `SubagentStop` fired, or it may not have been spawned through the `Agent` tool at all.

**YOU MUST NOT state seat coverage as a number while any `UNNAMED` row stands.** "No rework found across this run's seats" reads identically whether the seats were clean or whether one of them was never in the manifest to inspect. That is the plausible zero, aimed at this plugin's own output.

**A `MISSING` row** — a seat row naming a file that is not there — is the alarming direction and was measured at zero. It is reported so that stays a measurement rather than an assumption the tool cannot contradict.
