---
name: restart-recovery
description: After a host restart, crash or shutdown cut off running Claude sessions, use this to find them with `telepathy agents --lost` and bring each back in the mode it had — through a running Remote Control server first, then by reattach or local resume, one session at a time and only the ones the human chooses.
---

# restart-recovery

A restart leaves every transcript on disk and nothing that says which sessions were running.
`telepathy agents --lost` answers from the catalogue: each session whose last sign of life precedes
the boot and that nothing shows ending, with where to resume it from, its cloud id when capture
recorded one, its last recorded permission mode, and the commands that bring it back.

The list is **candidates for a human**, never a verdict: a session that exited cleanly after the
last closure sweep reads exactly like one the restart killed. **The resume attempt is the check** —
`claude --resume` from the wrong directory finds no conversation, and a reattach to an expired
environment says so. Both fail loudly; this skill falls back and reports.

## 1. Read the board

- BEFORE listing, YOU MUST run `telepathy backfill`, then `telepathy agents --lost`. The backfill is
  safe to repeat: it reads sessions whose hooks never wrote a row, and clears the rebuild warning an
  upgrade leaves.
- The first line is the boot and its source. The second is the **newest pre-boot activity** and how
  long ago it was, already judged against the window below. A second line reading `none in the
  store` means the store has not read the corpus: run `telepathy backfill` and ask again.
- IF a session the human expects is missing, YOU MUST widen `--window` (default `48h`) before
  concluding it is gone: an idle session that went quiet long before the others falls outside it.

## 2. First, let a running server do it

A restarted Remote Control server brings its own sessions back for about four hours after it
STOPPED, and brings a server-spawned session back when that session is OPENED in the app. The
server stopped at or after the newest pre-boot activity, so the second line's "ago" is an UPPER
bound on the time since it stopped: `inside the ~4 h window` means a running server will bring its
sessions back when they are opened; `past` means it may not, and opening is still worth one try.

- BEFORE resuming anything, YOU MUST check whether a server is running again: a systemd user unit
  such as `claude-rc@<name>` for which `systemctl --user is-active claude-rc@<name>` prints
  `active`, or a `claude rc` / `claude remote-control` process whose working directory is the
  repository (`pgrep -af 'claude (rc|remote-control)'`, then `readlink /proc/<pid>/cwd`).
- IF one is running, YOU MUST tell the human to open, in the app, every `lost` row whose
  `RESUME FROM` is that repository or one of its `.claude/worktrees/bridge-cse_*` worktrees — the two
  places a server-spawned session runs — and then re-run `telepathy agents --lost`. A session that
  came back reads `live` and drops off the list; only what is still `lost` goes on to step 3.

## 3. The human chooses

- BEFORE resuming, YOU MUST get the human's choice of sessions. The list has false positives, and
  resuming re-activates an agent.
- **A recovered session resumes in exactly the mode it began in.** `MODE` is its last recorded
  permission mode, read from its own transcript. BEFORE the human chooses, YOU MUST name every
  session whose `MODE` is `auto` or `bypassPermissions` and say that, once resumed, it acts
  without asking. The human's choice stands.
- A row reading `MODE unknown — the resume starts in whatever mode the CLI chooses` carries no mode
  flag: say so in the choice rather than naming a mode it may not get.
- A `RESUME FROM` marked `(unverified)` is a hint — the session moved folders, or its path was
  long. Say so; the resume attempt will confirm or refute it.

## 4. Resume, one session at a time

Per chosen session, in its own detached tmux SESSION — never a window inside another:

```sh
tmux new-session -d -s rr-<first 8 of its id> -c '<RESUME FROM>' '<command>'
```

- `<RESUME FROM>` is the PATH alone: drop a trailing ` (unverified)` before passing it to `-c`.
  IF the block reads `unknown (transcript folder …)`, YOU MUST ask the human for the directory
  rather than guess one — a resume from the wrong directory finds no conversation.

- IF the block shows a `REATTACH` command, YOU MUST run it first. It carries no mode flag and needs
  none: the session resumes on the server side in the mode it had.
- IF the reattach reports that *its environment has expired*, YOU MUST stop it at once —
  `tmux kill-window -t =rr-<id8>:` — because it has created an empty cloud session; then start the
  `LOCAL RESUME` command the same way. The local resume passes `--permission-mode <MODE>` (a
  recorded `default` is written `manual`), which overrides the CLI's own default.
- IF the resume reports *no conversation found*, the directory was wrong: YOU MUST report it and
  ask the human where the session ran, never guess another directory.
- BEFORE starting the next session, YOU MUST wait until `telepathy agents` shows the resumed one
  `live` (its session file has appeared). Two resumes starting at once can close each other.
- To see what a resumed session shows: `tmux capture-pane -p -t =rr-<id8>:`.

## 5. Report

AFTER each session, YOU MUST report:

- the outcome: **reattached**, **resumed locally** (same session id, new cloud session), or
  **failed** with the error it printed;
- the mode it resumed in;
- where to find it — the Claude app, or https://claude.ai/code;
- every empty cloud session an expired reattach created, for the human to archive;
- that display names are server-side and must be set again by hand;
- that each reattach holds its own single-session "at capacity" entry in the app until that session
  ends.

## Traps (measured)

- `claude rc --help` does not exit. Do not run it unattended.
- The trust prompt for a directory the client has not seen defaults to *No, exit*: send Down, then
  Enter.
- tmux TARGETS are `=<name>:` — exact session, with the colon — for `send-keys`, `capture-pane` and
  `kill-window`. `<name>:` PREFIX-MATCHES another session (`rr-1:` hits `rr-12`), and `=<name>`
  without the colon fails. `new-session -s` takes the PLAIN name: `-s '=rr-x:'` creates a session
  literally named `=rr-x_` that no target matches.
- A fresh interactive `claude` may start in `auto`; that is why the local resume carries the
  recorded mode rather than trusting the default.
- Transcript folders under `~/.claude/projects/` begin with `-`: pass them after `--`, or a tool
  reads the name as a flag.

## What it cannot tell you

- **Whether a candidate really was cut off.** Clean exits after the last closure sweep, sessions the
  sweep's cap deferred, and sessions that never finished a turn all read as `lost`.
- **Sessions idle past `--window`.** Widen it to reach them.
- **After an upgrade that rebuilt the store,** every session closed before the next boot can read as
  `lost` once, and a cloud id returns only when a hook has re-recorded it — the local resume works
  without one.
- **Display names**, which live on the server only.
