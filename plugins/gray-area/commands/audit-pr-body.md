---
description: Adjudicate a pull request body's claims against what the session actually ran. Reports CITED / NO-EVIDENCE / UNCHECKABLE with provenance on both sides, and names the parts of each claim the record cannot check. Exit is 0 even with findings.
---

Put a pull request body against the trajectory. Model [[terse-communication]]: relay the binary's rows, add no interpretation of your own.

The adjudication lives in the tested binary, not in this prompt. Your job is to run it and report what it says.

1. Get the body into a file — from the pull request, or from the draft you are about to post. If there is no body, say so in one line and stop.
2. Run `${CLAUDE_PLUGIN_ROOT}/bin/gray-area` (`.exe` on Windows) as `gray-area pr <body.md>`, passing `--json` straight through if the caller gave it. With no transcript argument the binary resolves this session's from gray-area's own manifest and prints which row it used.
3. Relay the rows verbatim, including the closing line.

**THE BOUNDARY, AND YOU MUST NOT CROSS IT.** The trajectory records what was **RUN**. It does not record what the run **SAID** — result bodies are conversation content and this plugin does not copy them. So a claim like ``` `go run ./check` → 26 passed ``` splits into an act this record can check and an outcome it cannot see at all. Every row whose claim asserted an outcome prints `NOT MEASURED:` naming it.

- YOU MUST NOT report a `CITED` row as confirming the numbers in the body. It confirms the command was invoked. Nothing here checked what it printed.
- YOU MUST NOT report a `NO-EVIDENCE` row as evidence the body is false. It is an ABSENCE: the command may have been spelled in a form the tokens miss, or run in a different session. The row prints what was searched and how much, so a reader can judge.
- **Exit is 0 even with findings**, unlike `checkpoint`. A body is adjudicated after the fact, for a human. Failing on a `NO-EVIDENCE` row would turn "the tokens did not match" into "the pull request is wrong".

**What counts as a claim.** A backticked command, and nothing else. A body is prose written for a human, and [[facts-are-fields]] is explicit that prose for a human audience is not the violation — the load-bearing part goes in the code span. Fenced blocks are skipped, because they hold output as often as commands. So a clean run means *the claims that could be checked, checked out* — never *the body is accurate*.

**A body with no adjudicable claim says so.** That is not a pass. It means nothing in it was checkable, and the tool reports it in those words rather than printing an empty list.

**The failure this exists to avoid** is a false positive on an accurate body, which puts the human between two of the agent's outputs with no way to tell which is lying. When in doubt, relay the row and let them convict.
