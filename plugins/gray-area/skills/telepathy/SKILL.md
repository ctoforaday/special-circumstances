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

`find` answers in one query: one row per session and agent, **newest first**, with the hit count,
the time of the most recent hit, the channel it landed in, and the text around it. `--in assistant`
keeps only what agents *said*; `--paths` adds the transcript path for citation.

- AFTER a `find`, YOU MUST read the **IN** column before quoting anything. `find` searches the
  whole transcript, so a hit may be a tool result, a task notification or a seat prompt — on
  `bench rul` across this box, 9 of the first 11 rows were prompt boilerplate. Quoting a `result`
  row back to a session as its own words is the mistake this column exists to prevent.
- `find` takes a **literal** by default; pass `--regex` for a pattern. Use it for word boundaries —
  searching `roving` rather than `\broving\b` returns every occurrence of "p*roving*", which is how
  this rule was earned. A substring match is the default failure mode of every search here,
  including `touched` and `sql`'s `LIKE`.

Every verb takes `--help`, and each one's help states what that verb cannot tell you. `sql` takes
`--limit` (default 200) and prints a line when it truncates, so a capped result is never mistaken
for a complete one.

`sql` is read-only and the views are the published contract: **`v_session`, `v_action`, `v_word`,
`v_thought`, `v_skip`**. Write your own query rather than asking for a verb — the questions worth
asking are not knowable in advance, which is why the surface is SQL and not a menu. `sql --help`
prints the columns of each view, generated from the schema rather than retyped beside it.

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
- **Reasoning is mostly WITHHELD, and `v_skip` is where that fact lives.** The client emits
  thinking blocks carrying a signature and no text. Measured on this box 2026-09-09: **15,503
  such blocks across 99 sessions, against 248 thoughts actually stored**, and none with text at
  all since 2026-09-08. Those blocks are now recorded as `thinking-empty` skips rather than
  dropped, so the two states are finally distinguishable. YOU MUST check `v_skip` before reading
  an empty `v_thought` as evidence an agent did not reason — for most sessions on this box it is
  evidence of nothing but the client's settings.
- **A session that predates capture contributes nothing** until `backfill` has read it. Absence of
  rows is not absence of work.
- **`touched` sees a path only where the record NAMES it.** For `Read`/`Edit`/`Write` that is the
  `file_path`, and the answer is exact. For `Bash` it is the command string, truncated at 200
  characters — so an edit made by a long shell command, or a heredoc that names the file past that
  point, does not appear. AFTER a `touched` that returns nothing on a file you expect contention
  on, YOU MUST fall back to `telepathy find <basename>`, which reads the transcripts themselves.
- **`unresolved` is not `error`.** A tool call whose result never arrived — an interrupted session
  — is recorded as `unresolved`. YOU MUST NOT fold it into a failure count; doing so overstates
  every error rate on the box, and the sessions it overstates are exactly the ones that were cut
  short rather than the ones that went wrong.

## The rule this exists to serve

Exploration may summarize; **adjudication must cite**. The catalogue is a citing surface: every
row names the session, the agent and the act it came from. Use it to find leads and to settle
questions of fact about what ran — and where a claim needs evidence, cite the act, not the
summary of it.
