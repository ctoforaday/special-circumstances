# Session-Telemetry Commit Annotations

**Status:** Proposal — under review
**Depends on:** prosthetic-conscience Go hook toolchain (`plugins/prosthetic-conscience/tools/`), capability-gating (`requirements.json`), `/sc-doctor` install flow
**Relates to:** memory-architecture.md (trajectory data), the planned `sc-secrets-gate` (redaction patterns)

---

## 1. Problem

Every agent commit in this repo already carries two trailers:

```
Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01ThHwXsvBrDkDoYDEAbbbhz
```

That tells us *who* and *which session* — but nothing about *what kind of session*. Was this commit the product of three focused turns or a 400-turn slog? Was the agent at 8% of its context window or 94%? Had it survived two compactions (and therefore lost fidelity on the original instructions)? How many `go test` runs preceded it?

This proposal specifies **session-telemetry trailers**: machine-readable annotations on every agent commit recording turn counts, tool-use steps, token usage, compaction survival, a per-tool command histogram, and skill loads. The payoff is a git history that doubles as an agent-behavior dataset — greppable, diffable, and correlatable with code quality after the fact ("do commits made above 80% context utilization get reverted more often?").

### Goals

- Every commit produced by a Claude Code session in this repo carries a telemetry annotation, generated automatically.
- Telemetry **never blocks or breaks a commit**. Absence of telemetry degrades to today's behavior.
- The parser and hook are Go binaries in the existing tested toolchain: table-driven tests, CI matrix, capability-gated.
- Manual inspection first (`/session-stats`), automation second.

### Non-goals

- Cost accounting in dollars (pricing tables drift; token counts are the durable primitive).
- Telemetry for human commits (annotate them as `human` at most — no session to describe).
- Cross-session analytics tooling (a later consumer of this data, not part of this design).

---

## 2. Grounding: what is actually available

Everything below was verified against the live Claude Code hooks documentation (code.claude.com/docs/en/hooks, fetched 2026-07-11) and by parsing a real ~2,600-line session transcript on the development machine. Uncertain items are flagged inline; §9 collects them.

### 2.1 The hook contract

Every hook receives JSON on stdin with, at minimum:

```json
{
  "session_id": "05c428b1-...",
  "transcript_path": "~/.claude/projects/<sanitized-cwd>/<session-id>.jsonl",
  "cwd": "...",
  "hook_event_name": "PreToolUse",
  "permission_mode": "default"
}
```

`PreToolUse`/`PostToolUse` additionally carry `tool_name` and `tool_input` (for Bash, `tool_input.command` is the full command string). A `PreToolUse` hook **can modify tool input**: exit 0 with

```json
{"hookSpecificOutput": {"hookEventName": "PreToolUse",
                        "permissionDecision": "allow",
                        "updatedInput": {"command": "..."}}}
```

— confirmed in the current docs. So the reviewer's "pre tool invocation hook" idea is *technically feasible*. §3 explains why we should not use it for message mutation anyway.

Hook processes get `CLAUDE_PROJECT_DIR`, `CLAUDE_PLUGIN_ROOT`, and `CLAUDE_PLUGIN_DATA` in their environment. **There is no documented environment variable carrying the session id into the Bash tool's subprocesses** (i.e., into `git commit` itself) — which is exactly why a native git hook needs a handshake file (§3.4).

Relevant hook events beyond PreToolUse: `SessionStart`, `PostToolUse`, `PreCompact`/`PostCompact`, `Stop`. We use `SessionStart` and `PreToolUse` only.

### 2.2 The session transcript

`~/.claude/projects/<sanitized-cwd>/<session-id>.jsonl` is an append-only JSONL log. Observed line types (counts from the reference session): `user` (512), `assistant` (811), `system` (80), `attachment` (169), plus harness bookkeeping (`queue-operation`, `file-history-snapshot`, `mode`, ...). The shapes that matter:

- **Assistant lines** carry `message.model` (e.g. `claude-opus-4-8`), `requestId`, and `message.usage`:

  ```json
  {"input_tokens": 2, "cache_creation_input_tokens": 1033,
   "cache_read_input_tokens": 386824, "output_tokens": 380, ...}
  ```

  The sum `input + cache_read + cache_creation` on the *latest* assistant line is the prompt size at that point — i.e., context in use. One API turn may span several assistant lines (one per content block) sharing a `requestId`; usage must be de-duplicated per request.

- **User lines** carry `message.content` as either a string / text blocks (a real human turn) or a list of `tool_result` blocks (a tool round-trip). `isMeta: true` marks harness-injected user lines; `isCompactSummary: true` marks the post-compaction summary. In the reference session: 152 human-content lines (of which 28 meta) vs 360 tool-result lines — so naive "count user lines" overstates human turns by ~3×. The discriminator is content shape + `isMeta`.

- **Tool use** appears as `tool_use` content blocks on assistant lines, with `name` and `input`. Reference session: `{"Bash": 113, "Write": 76, "Read": 69, "Edit": 66, "Agent": 12, ...}`. `Skill` tool invocations carry `input.skill` — skill loads are countable by name. `Agent` invocations give subagent-spawn counts.

- **Compactions** appear as `type: "system", subtype: "compact_boundary"` lines with rich metadata:

  ```json
  {"subtype": "compact_boundary", "content": "Conversation compacted",
   "compactMetadata": {"trigger": "manual", "preTokens": 251767,
                       "postTokens": 12551, "cumulativeDroppedTokens": 239216, ...}}
  ```

  This is better than hoped: not just a count, but per-compaction trigger (`manual`/`auto`) and token deltas.

**Caveat:** the transcript format is an internal, undocumented harness format. It has been stable in shape across recent versions (each line carries a `version` field we can log), but the parser must treat unknown line types and missing fields as skippable, never fatal.

### 2.3 Context-window denominators

Percent-of-window needs the window size per model. Current values: 1M tokens for Fable 5 / Mythos 5 / Opus 4.6–4.8 / Sonnet 4.6–5; 200K for Haiku 4.5. These live in a small static table in the binary (overridable via config), with the model id taken from the transcript's assistant lines. The Models API (`GET /v1/models/{id}` → `max_input_tokens`) exists for live lookup but requires credentials and a network call inside a commit path — not worth it; a stale denominator makes the percentage mildly wrong, not the commit broken. The figure is additionally approximate because the harness reserves headroom (autocompact buffer, system prompt overhead) that we cannot see. **Label the field `ctx_pct≈` and document it as an estimate.**

---

## 3. Interception point

Three candidate mechanisms, one recommendation.

### 3.1 Option A — PreToolUse rewrites the `git commit` command

Match `Bash` tool calls whose command contains `git commit`, compute telemetry, and use `updatedInput` to splice trailers into the commit message.

*Feasible* (§2.1), *rejected*. The hook would have to parse and rewrite an arbitrary shell string: `-m` with single/double quotes, multiple `-m` flags, heredocs (the dominant pattern in this repo's commit convention), `-F file`, `--amend`, `-c`/`--fixup`, editor-based commits with no inline message at all. Getting shell-string surgery wrong corrupts the *command*, which is categorically worse than missing telemetry. It also annotates only commits made through the Bash tool in this harness — nothing from a human, another agent, or a different harness. Message mutation is git's own extension point; fighting for it at the shell-string layer is the wrong altitude.

### 3.2 Option B — native git hook (`prepare-commit-msg`) + handshake file — **recommended**

Git's `prepare-commit-msg` hook runs for every commit in the repo, receives the path to the message file, and may edit it before the editor opens (or before the commit finalizes for `-m`/`-F`). Appending trailers there via `git interpret-trailers --in-place` is the mechanism git itself provides for exactly this job; it is immune to how the commit was invoked.

The one gap: a git hook has no session context. Git runs as a subprocess of the Bash tool, and no session-identifying environment variable is documented to reach it. So we split responsibilities:

- **Producer (Claude Code side):** a `PreToolUse` hook matching `Bash` calls whose command matches `\bgit\b.*\bcommit\b` runs `sc-telemetry snapshot`. It reads the hook stdin (giving `session_id` and `transcript_path`), summarizes the transcript, and writes `$GIT_DIR/claude/session-telemetry.json` — a complete, redacted telemetry snapshot plus a timestamp. It always exits 0 and never emits `updatedInput`; it is a pure data producer. (A `SessionStart` hook additionally writes a lightweight identity stub so the git hook can distinguish "agent session, snapshot stale" from "no agent involved".)
- **Mutator (git side):** `prepare-commit-msg` runs `sc-telemetry trailers`, which reads the snapshot, checks freshness (skip if older than a few minutes — a stale snapshot means this commit didn't come through the producer path), and appends trailers with `git interpret-trailers`. It skips `merge`/`squash`/`commit -c` sources by inspecting the hook's second argument, and skips entirely when no snapshot exists — so human commits are untouched.

This division puts each half at its natural altitude: the harness hook knows the session; the git hook owns the message. Freshness plus PreToolUse timing (the snapshot is written *immediately before* the commit command runs) makes the pairing tight without any brittle correlation logic.

`commit-msg` would work too, but `prepare-commit-msg` is preferred: trailers become visible in the editor for interactive commits, and `commit-msg` is conventionally reserved for validating gates (we may later want an `sc-quality-gate`-style validator there; don't squat on it).

### 3.3 Option C — PostToolUse + `git commit --amend`

Rejected outright: it mutates the SHA *after* the commit exists (racing any immediate `git push`, breaking `commit → tag` sequences, re-running hooks, and invalidating signatures), and an amend from a hook while the agent believes the commit is final is exactly the kind of surprising state change prosthetic-conscience exists to prevent.

### 3.4 Hook installation

Native git hooks are per-clone and not versioned. Installation goes through the existing consent flow: `/sc-doctor --fix` (already the plugin's install surface) offers to either symlink/copy the `prepare-commit-msg` shim into `.git/hooks/` or set `core.hooksPath` to a versioned `hooks.d/` directory. The shim is three lines of sh calling the Go binary; if the binary is missing it exits 0. Uninstall is deleting the shim. If a `prepare-commit-msg` hook already exists, `/sc-doctor` chains rather than clobbers, and asks first.

---

## 4. Feasibility matrix

| Field | Source | Verdict |
|---|---|---|
| Session id | hook stdin `session_id` | **Possible** — exact |
| Model id(s) | transcript `message.model` (may be plural per session) | **Possible** — exact |
| User turns | `user` lines with string/text content, `isMeta` excluded | **Derivable** — exact given format holds |
| Assistant turns | `assistant` lines de-duplicated by `requestId` | **Derivable** — exact |
| Steps (tool-use) | count of `tool_use` content blocks | **Derivable** — exact |
| Tool counts + command map | `tool_use` name histogram; Bash commands normalized to verb + subcommand (`"git commit"`, `"go test"`) | **Derivable** — normalization is heuristic but the histogram is exact |
| Output tokens (cumulative) | Σ `usage.output_tokens` per `requestId` | **Derivable** — exact for the main thread |
| Context in use at commit | last assistant `usage`: `input + cache_read + cache_creation` | **Derivable** — exact at snapshot time |
| % of context window | above ÷ static per-model window table | **Approximate** — denominator excludes harness headroom; labeled `≈` |
| Compactions survived | `system`/`compact_boundary` lines incl. `trigger`, `preTokens`/`postTokens` | **Possible** — exact, richer than requested |
| Skills loaded (+counts) | `Skill` tool_use `input.skill` | **Derivable** for explicit Skill-tool loads; **partial** for implicit loads (slash-command expansions and frontmatter preloads surface differently — counted best-effort, flagged in output as `skills_explicit`) |
| Subagent count | `Agent` tool_use count | **Derivable** — spawns only; subagent-internal tokens/steps live in separate transcripts and are **not available** from the main transcript |
| Wall-clock duration | first/last line timestamps | **Derivable** — exact |
| Dollar cost | token counts × pricing | **Not included** — computable downstream; pricing drifts |
| Anything from *other* concurrent sessions in the same repo | — | **Not available** — the snapshot describes the committing session only |

---

## 5. Where the data goes

### Options

**(a) One trailer, one JSON blob** (`Claude-Telemetry: {...}`). Compact, but the headline numbers (turns, tokens, compactions) are buried inside JSON — not greppable with plain `git log --grep`, ugly in `git log` and the GitHub commit UI.

**(b) Scalar trailers + one compact JSON trailer for the map-shaped data.** Trailers are RFC-822-style `Key: value` lines; scalars fit them natively, and a single-line minified JSON object is a legal trailer value for the one genuinely map-shaped field. Greppable (`git log --grep='Claude-Compactions: [1-9]'`), readable, visible in the GitHub UI, and — because trailers travel *inside* the message — survive rebase, cherry-pick, and `--amend`. (GitHub squash-merges concatenate messages, preserving the data but mangling per-commit attribution; acceptable, noted as a limitation.)

**(c) Git notes** (`refs/notes/session-telemetry`). Attractive for the full blob — unbounded size, no message noise — but notes don't push or fetch by default, don't render on GitHub, detach on rebase unless `notes.rewriteRef` is configured on every clone, and add a second sync surface to get wrong. Wrong default, fine extension.

**(d) A committed session-log file.** Pollutes the tree and the diff, causes merge conflicts, conflates telemetry with content. Rejected.

### Recommendation: (b), with (c) as an opt-in Phase C extension

```
docs: add session-telemetry trailers design

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01ThHwXsvBrDkDoYDEAbbbhz
Claude-Model: claude-opus-4-8
Claude-Turns: 14u/41a
Claude-Steps: 63
Claude-Tokens: in=386824 out=21930 ctx_pct≈39
Claude-Compactions: 1 (auto=0 manual=1)
Claude-Skills: critical-stance=1 verify=2
Claude-Tools: {"Bash":{"git commit":3,"go test":2,"other":9},"Edit":14,"Read":9}
```

Rules: `Claude-Tools` is minified JSON on one line, capped at 512 bytes (top-N commands per tool, remainder folded into `"other"`); every other trailer is a scalar. Keys are fixed and documented so downstream tooling can parse with `git interpret-trailers --parse`. If a future need outgrows the size cap, the full uncapped JSON goes into a git note and `Claude-Tools` gains a `"note":true` marker — that is the entire (c) extension.

---

## 6. Component map

One new Go binary in the existing toolchain, three thin wiring points.

```
plugins/prosthetic-conscience/
  tools/cmd/sc-telemetry/          # new binary, same layout as sc-quality-gate
    main.go                        # subcommand dispatch
    main_test.go                   # table-driven, fixture JSONL transcripts
  tools/internal/transcript/       # transcript parser (line-tolerant, versioned)
  tools/internal/toolchain/        # existing: capability gating, hook IO
  hooks/hooks.json                 # + SessionStart, + PreToolUse(Bash∩git commit)
  hooks/prepare-commit-msg         # 3-line sh shim → sc-telemetry trailers
  commands/session-stats.md        # /session-stats → sc-telemetry summarize --pretty
```

**`sc-telemetry` subcommands:**

- `summarize --transcript <path> [--json|--pretty]` — pure function: JSONL in, telemetry summary out. This is where all parsing logic and all tests live.
- `snapshot` — PreToolUse mode: hook JSON on stdin → `summarize` → redact → write `$GIT_DIR/claude/session-telemetry.json` atomically (temp + rename). Uses an **incremental cursor** (§7). Always exits 0.
- `trailers [--msg-file <path> --source <src>]` — prepare-commit-msg mode: read snapshot, enforce freshness window, emit/append trailers via `git interpret-trailers --in-place`. Exits 0 on any failure.
- `init-session` — SessionStart mode: write the identity stub.

The transcript parser is deliberately its own internal package: `/session-stats`, the memory-architecture consolidation loop, and any future trajectory analysis all want "parse a session transcript into a summary struct" — this is the one place that knowledge lives.

---

## 7. Trust, cost, privacy

**Never block a commit.** Both hook entry points hard-exit 0 on every error path (missing transcript, malformed line, snapshot write failure) with a one-line stderr note. Explicit sub-timeout (5s) in `hooks.json` well under the default. Capability-gated via `requirements.json` like the rest of the plugin: no Go binary → hooks are inert, commits are plain.

**Parsing cost.** Transcripts reach tens of MB in long sessions; parsing from byte 0 on every commit is wasteful. `snapshot` keeps a cursor file next to the snapshot (`session-telemetry.cursor.json`: session id, byte offset, running aggregates) and resumes from the offset — safe because the transcript is append-only (compaction *appends* a boundary line; it does not rewrite history). Cursor mismatch (different session id, offset beyond EOF, version change) → full reparse. Cold parse of a 25MB JSONL in Go is well under a second; warm incremental parse is milliseconds.

**Privacy.** Two layers. First, structural: the command map records only *normalized command heads* (first token, plus subcommand for known multiword CLIs: `git`, `go`, `npm`, `gh`, `docker`) — never arguments, so `curl -H "Authorization: Bearer sk-..."` contributes exactly `"curl"`. Skill names and tool names are non-sensitive by construction. Second, belt-and-braces: the snapshot passes through the same deny-pattern scrub planned for `sc-secrets-gate` before hitting disk, so pattern-shaped secrets can't leak even if normalization regresses. The snapshot lives under `$GIT_DIR/claude/` (never tracked); only the trailer block enters history.

**Multi-session repos.** Two concurrent sessions in one clone race on the snapshot file. The freshness window plus PreToolUse timing means the *committing* session wrote the snapshot moments before its own `git commit` runs; a lost race mislabels a commit only if two sessions commit within seconds of each other. Acceptable for Phase B; a per-session snapshot keyed by session id with the git hook picking the freshest is the fix if it ever matters.

---

## 8. Phased build plan

**Phase A — parser + `/session-stats` (no commit mutation).**
`sc-telemetry summarize`, the `internal/transcript` package, fixture transcripts (including a fixture with a `compact_boundary`, a `Skill` call, and multi-line `requestId` groups), table-driven tests, CI matrix entry. Ship `/session-stats` printing the pretty summary for the current session. This validates every number against reality — humans eyeball the stats for a week or two before any of them touch git history. Exit criterion: `/session-stats` numbers match manual transcript inspection on three real sessions.

**Phase B — automatic commit annotation.**
`snapshot` + `trailers` + `init-session`, hooks.json wiring, the `prepare-commit-msg` shim, `/sc-doctor --fix` install/uninstall with consent, cursor caching, redaction pass. Exit criterion: a normal working session produces annotated commits; killing the binary mid-write, deleting the snapshot, and committing as a human all produce clean unannotated commits.

**Phase C (optional, demand-driven) — git notes overflow + aggregation.**
Full uncapped telemetry JSON into `refs/notes/session-telemetry`; an `sc-telemetry report` over `git log --format` for cross-commit queries. Only if Phase B data proves useful enough that someone wants the long tail.

---

## 9. Risks and open questions

| Risk | Severity | Mitigation |
|---|---|---|
| Transcript format is internal and undocumented; a harness update changes line shapes | Med | Tolerant parser (skip unknown, never fatal); log `version` field; fixtures pinned per observed version; failure mode is missing trailers, not broken commits |
| Skill-load counting under-reports implicit loads (slash-command expansion, frontmatter preload) | Low | Field documented as explicit-loads; revisit when the attachment `skill_listing` shape is better understood |
| `ctx_pct` denominator ignores harness headroom / autocompact reserve | Low | Labeled `≈`; static table documented; direction of error is consistent (understates fullness) |
| Squash merges concatenate trailer blocks from multiple commits | Low | Documented; per-commit data survives, attribution is per-PR |
| Existing `prepare-commit-msg` hook in a clone | Low | `/sc-doctor` detects, chains, asks |
| Snapshot race between concurrent sessions in one clone | Low | Freshness window; per-session snapshots if it recurs |
| Windows: `.git/hooks` shims need sh; this repo already assumes Git Bash | Low | Same constraint as existing toolchain; covered by CI matrix |

Open question for review: should human commits get a minimal `Claude-Session: none` marker (making "unannotated" distinguishable from "hook broken"), or stay untouched? Current draft: untouched — absence of the identity stub already encodes it, and unsolicited trailers on human commits feel like overreach.

---

## 10. Alternatives considered (summary)

- **PreToolUse `updatedInput` rewrite of the commit command** — feasible per the hook contract, rejected: shell-string surgery on heredoc commit messages is the highest-risk possible failure point, and it misses non-Bash-tool commits (§3.1).
- **PostToolUse `--amend`** — rejected: SHA mutation after the fact races pushes and breaks signing (§3.3).
- **Single JSON-blob trailer** — rejected as primary format: kills greppability of headline scalars (§5a).
- **Git notes as primary store** — rejected: no default sync, invisible on GitHub, rebase-fragile; retained as capped-overflow extension (§5c).
- **Committed telemetry file** — rejected: tree pollution, merge conflicts (§5d).
- **Live Models API lookup for window sizes** — rejected in the commit path: network + credentials for a cosmetic denominator; static table with config override instead (§2.3).
