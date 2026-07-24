# FEOV handoff / memento — checkpoint 2026-07-24

Resuming **frank-exchange-of-views** (the research-debate engine). This is the pickup point.
Read it, then `gh issue list` for the live queue. **Verify before trusting** — several claims this
run were wrong on first telling and caught only by checking (a subagent that did the wrong task; a
git-author assumption never checked at the leaf; #70 that looked bounded and wasn't).

## Identity (do this FIRST, before any gh)
- **gh runs as the USER by default — switch it.** `gh auth switch --user gblock-agent`, verify with
  `gh api user --jq .login` (must print `gblock-agent`). Both accounts are in the keyring. PRs/issues
  must be the bot's, not the user's ([[gh-bot-account]]).
- Commit author is ALREADY the bot: `Ethics Gradient <302690890+gblock-agent@users.noreply.github.com>`
  (repo-local git config). Do NOT "fix" it to gblock@ — both accounts share the display name
  "Ethics Gradient"; the EMAIL is the discriminator.

## State right now
- **`main`** = `2b00595` (PR #104 merged). Plugin **0.41.0**, recordToolVersion/cli.Version **0.14.0**
  (tool + plugin bump together only when the TOOL changes; a docs-only PR bumps plugin alone — #104
  moved plugin 0.40→0.41 with the tool held at 0.14).
- Env READY (qlty/jq on PATH; gcc still off PATH — [[tools-installed-but-off-path]]).
- Local run recipe unchanged — see [[feov-projection-retirement-queued]].

## What SHIPPED this session (all merged)
- **#100** — `confidence` wired end-to-end, NON-AUTHORITATIVE: blue emits per-claim calibration
  (both blue seats), the debate view + report render it, it NEVER reaches the risk matrix (guarded by
  construction + test). Was a total dead-letter before.
- **#103** — **#62 Stage 2**: grade disputes onto the record. blue emits `dispute` events, red emits
  `dispute-respond`; the envelope keeps only routing refs (proposed/response); render.go's debate view
  now renders the dispute thread (else the bench, reading `--view debate`, goes blind). Fuzz drives the
  docket through the envelope (folds in #101's dispute coverage).
- **#104** — **#73**: repointed the three seat constitutions off retired file paths (debate.md,
  red/ledger.md, red/archive.md) to the tool (`show --view` / the emitting verbs). Stage-3 prep.
- **PR #3 (Gray Area memento plan)** — updated with live-run learnings (I1 note-is-not-the-queue,
  I2 record-each-check's-trigger-surface, I3 durable-pointer). Not merged yet — its own review.
- Filed **#101** (fuzz deferred branches: docket/deadlock/petitions/lineage — pre-run gate with #80)
  and **#102** (wire-or-drop the remaining write-only events); slimmed **#95** to fan-out/collisions.

## The #62 push — plans/record-only-channel.md
- **verify** ✓ · **Stage 1** (positions/closings) ✓ · **Stage 2** (disputes) ✓ (#103)
- **Stage 2.5** — the ORCHESTRATOR on the record: debate.js's mechanics (docket, round, deadlock,
  verdict) are ephemeral; record them via a lead-mechanics channel emitted by an AGENT PROXY (the
  script is sandboxed, can't call the tool). **#87 (friction verb bypassed) folds in here** — same
  "sandboxed orchestrator can't persist to the record" problem. This is the next #62 step.
- **Stage 3** — retire debate.md as write AND read target; update the parity-audit.

## Queue order (build before the ONE final run)
**Stage 2.5 (folds #87) → #70 → #71 → #68 → #72 → #63/#64 → decide-E → {#101 + #80} → the run.**
- **#70** is SCOPED + de-risked on the issue (see the 2026-07-24 comment): the deterministic claim
  counter is BUILT + 13-case tested (code in the comment). The blocker found: the `revision` verb is
  vestigial (blue hand-counts into the envelope, not via revision), so #70 = shared count pkg + a
  read-only `count-claims` root command (sibling to verify/graph) + prompt rewire + drop
  revision's --claim-count. Medium PR, not a quick win — anti-spinning aborted the rushed attempt.
- **decide-E** = the contract/ergonomics issues (#84/#85/#86/#87/#88) are DECISIONS (wire-or-drop),
  not blind builds — batch them like the confidence call.

## Fixed-pending-a-run — the ONE run closes these (never close a bug until a RUN confirms, [[bug-state-tracking]])
#83, #94, #66, #67, #78, #65, #74, #75, #76, #79. Re-verify with a `/research` run + `verify`.

## Earned pitfalls (this session)
- **rule-sweep gates more than SKILL.md** — also `agents/*.md`, `scripts/debate.js`, `references/*.md`,
  `commands/*.md`, `law/`. Any of those needs Rule-Class + Sibling-Sweep trailers. Pre-check:
  `node scripts/rule-sweep.mjs --base origin/main`. It bit #100 (I assumed SKILL.md-only) ([[ci-checks-and-local-loop]]).
- **Verify config at the leaf, not from context** — I told the user commit author was the user's email
  because the session context lists "user email"; the repo config was already the bot noreply. Check
  `git config`, don't infer.
- **A FORK inherits your context and follows ITS gravity** — a fork spun for an orthogonal task (memento
  analysis) did the SPINE work instead. For orthogonal work use a FRESH general-purpose agent, walled
  off from the current focus.
- **Anti-spinning is real** — #70's 3 approach-changes were the signal to STOP and reassess, not force.
  Preserve the built artifact (recorded on the issue), revert to a clean tree, checkpoint.
- Bump cli.Version BEFORE regenerating stamp goldens, `-count=1` (the #57 lesson).
