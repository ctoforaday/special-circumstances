# Git hooks

Run `scripts/setup-git-hooks.sh` on any box that commits to this repository.

```sh
./scripts/setup-git-hooks.sh                 # install, verify
CHECK_ONLY=1 ./scripts/setup-git-hooks.sh    # audit an existing box, change nothing
```

It points `core.hooksPath` at the tracked `.githooks/` directory, so the hooks are
reviewable, update with a pull, and are the same for everyone. `.git/hooks/` is not
tracked; a hook committed here runs for nobody until each clone opts in, which is why
this is a script and not just a file.

## `pre-commit` — refuses unformatted Go

Unformatted Go has reached CI three times (#351, #352, and the 2026-08-22 vocabulary
rename). Two checks fail on it — `feov-record` and `qlty` — so one miss lights two red
boxes and reads like two problems, and it costs a full round-trip to learn something
`gofmt -l` answers in under a second.

**The cause is always the same, and it is not CRLF.** `.gitattributes` already carries
`*.go text eol=lf`, `autocrlf=input`, and the committed blobs were verified LF (#352).
It is an edit that changes the WIDTH of a key in an aligned map literal or struct-tag
block: gofmt re-aligns the whole block, and a scripted rename never re-runs it.

It **refuses rather than fixing**. A hook that silently rewrites staged content changes
what you are about to commit without showing you, and the diff you reviewed is no longer
the diff you ship. The fix is one command and the refusal prints it.

It checks the **staged content**, not the worktree file — those differ under `git add -p`,
and the worktree copy is not what is being committed. A dirty tree is the ordinary state
of development and is not grounds for refusing anything.

Bypass once with `git commit --no-verify`.
