---
name: telepathy
description: Before asking a peer agent what happened, or assuming nobody else is working on a file, use `telepathy` to query the host-wide trajectory catalogue — it answers from what agents DID rather than from what they remember.
---

# telepathy

`SendMessage` asks an agent what it **believes**, and needs it alive and willing. The catalogue
asks the record what **happened**.

- BEFORE messaging a peer to ask what it did, what it found, or whether something is done, YOU
  MUST query the catalogue first. A peer's answer is a self-report composed from a context window
  that may have been compacted; the catalogue is the acts themselves.
- BEFORE editing a file you did not create, YOU MUST run `telepathy touched <path>`. Two agents
  editing one file is the cheapest coordination failure to prevent and the most expensive to
  discover at merge.
- BEFORE relaying a claim you did not verify — "the peer says X merged", "that was fixed in #N" —
  YOU MUST check it. On 2026-09-08 three such relays between sessions on this box were wrong, and
  each was refutable from the record in one query.
- AFTER reading a zero from any verb, YOU MUST read WHICH zero it is. Every absence here is
  worded: *no store* names the path, *no rows for a path* is a statement about the catalogue and
  not the file, *no sessions advertised* means nothing is running. They are different facts and
  the tool says which.

## The verbs

```
telepathy agents                 who is running, in which worktree, doing what
telepathy session <id>           one session's shape: calls and errors by tool
telepathy touched <path>         which sessions acted on a path, and when
telepathy find <term>            ripgrep across every local transcript, joined to who
telepathy sql '<SELECT …>'       anything else
telepathy backfill               read the existing corpus into the store (explicit, safe to repeat)
```

`sql` is read-only and the views are the published contract: **`v_session`, `v_action`, `v_word`,
`v_thought`, `v_skip`**. Write your own query rather than asking for a verb — the questions worth
asking are not knowable in advance, which is why the surface is SQL and not a menu.

```sh
# which agents failed most, across the whole box
telepathy sql "SELECT session_id, count(*) n FROM v_action WHERE outcome='error' GROUP BY 1 ORDER BY 2 DESC"

# what did that agent actually do in the last hour
telepathy sql "SELECT tool, target FROM v_action WHERE session_id LIKE '5627%' AND ts > strftime('%s','now','-1 hour')"
```

## What it cannot tell you

- **`unknown` liveness is not `ended`.** Liveness is exact on Linux only, and a session from
  another pid namespace is not ours to judge. YOU MUST NOT read `unknown` as "gone".
- **A session's final turn may be missing** when its transcript lagged the last hook and no
  closure pass read it. Measured residue, deliberately not engineered around.
- **Reasoning is only there if it was captured.** `showThinkingSummaries` defaults OFF; a session
  that ran without it has acts and words and no thoughts. `session` says so rather than letting
  zero read as "it did not think".
- **A session that predates capture contributes nothing** until `backfill` has read it. Absence of
  rows is not absence of work.

## The rule this exists to serve

Exploration may summarize; **adjudication must cite**. The catalogue is a citing surface: every
row names the session, the agent and the act it came from. Use it to find leads and to settle
questions of fact about what ran — and where a claim needs evidence, cite the act, not the
summary of it.
