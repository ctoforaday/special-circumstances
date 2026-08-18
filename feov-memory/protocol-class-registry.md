# Protocol-rule defect classes — the registry for defects in OUR OWN rules

The gap-class registry (`class-registry-seed.md`) classifies defects the debate
finds in a REPORT. This one classifies defects in the RULEBOOK: the skills, seat
prompts, constitutions, templates and law that govern how the engine runs.

## Why this exists

Every protocol patch that shipped recurred in adjacent form:

| patched | recurred as | when |
|---|---|---|
| lens-scoped finding labels (#15/#16) | blue-lane footnote namespace collision | run 4 |
| blue reads the transcript (#15/#16) | gap-JSON lossy summary, nine commands short | run 5 |
| grade enums widened (#15/#16) | mass mapping conflates existence with harm | run 5 |
| friction-to-file (#15/#16) | lens seats still had no write path | run 5 |

Four for four. Each fix was correct and each fix was an INSTANCE fix: the class
survived and re-emerged at the adjacent seat, where it passed every test the
patch had added because those tests were written about the instance too.

The rule this registry enforces: **a rule patch names its class and sweeps the
siblings before it ships.** `scripts/rulesweep` (Go) checks it mechanically,
because a sweep requirement with nothing checking it would itself be an instance
of `policy-without-mechanism` — a class in this very registry.

## The classes

### adjacent-seat-omission
A duty, clause or capability is fixed at one seat class while sibling seats
carrying the same duty are left untouched.
- **Instances**: lens-scoped labels shipped while blue lanes kept colliding
  footnote namespaces; friction-to-file shipped while lens seats still had no
  write path; the record-invocation regex matched the mjs shape while the engine
  emitted a quoted binary path.
- **Sweep question**: which OTHER seats carry this duty, and does the fix reach
  each of them?
- **Neighbour**: `partial-control-coverage` (gap registry) — that one is about a
  control covering part of its surface; this is about a fix covering part of its
  seats.

### lossy-channel-substitution
A derived, summarized or projected channel is handed to a consumer that needs the
source, and the consumer cannot tell the difference.
- **Instances**: blue repairing from the gap JSON, whose carve-out enumeration
  was nine commands short of the source; the CHANGELOG used as a round-state
  signal when a revision shipped without one.
- **Sweep question**: does any consumer of this channel need the source, and does
  it know the channel is lossy?
- **Neighbour**: `reader-channel-mismatch` (gap registry).

### enum-vocabulary-undercoverage
An enum is widened for the traffic already observed rather than the traffic it
will meet, so the next unmodelled case is silently rounded into a wrong value.
- **Instances**: compound grades widened, then the mass mapping conflated
  existence with harm; the judge resolution enum lacking
  `routed_to_infrastructure` and a moot class, forcing `risk_accepted` as
  "least-wrong"; no verdict class for a ceiling-terminated run.
- **Sweep question**: what traffic exists that this enum still cannot express,
  and what does a seat do when it meets it?

### unobservable-duty
A MUST with no recorded artifact: nothing distinguishes a discharged duty from a
skipped one, so the clause is unenforceable and eventually unpractised.
- **Instances**: the MUST-try citation clause with no attempt line (a false
  "paywalled" on an open-access paper survived to round 1); "confidence
  self-graded" mandated and practised five times in 1,892 lines.
- **Sweep question**: what artifact would a reviewer look at to tell whether this
  was done?
- **Neighbour**: `self-attestation` (gap registry) — that is a claim standing on
  its own word; this is a duty with no place to leave a trace at all.

### staged-not-delivered
Content is staged where a seat COULD read it, and the rule assumes reading is
binding. Measured false: run 5's lanes read the gap patterns and committed both
warned patterns anyway.
- **Instances**: red's gap-pattern corpus staged into `inputs/`; scorecards
  computed into files no future seat read.
- **Sweep question**: does this content reach the seat AT THE DECISION POINT, or
  only at seat start where it competes with everything else for salience?

### policy-without-mechanism
A rule with no enforcer — sometimes because the enforcer was withdrawn and the
invariant was assumed to hold on its own.
- **Instances**: the round-record requirement before W1.7 gated it; the sweep
  requirement this registry describes, had it shipped as prose; the checkpoint
  note's validation-loop numbering (#192/#193) — the schema *showed* `1.`, `2.`,
  `3.` in an example and never stated it as a rule, so two parsers in two plugins
  each assumed it independently. A note numbered from `0.` then made re-arm state
  key `2` mean the note's `1.`, and lettered sub-entries opened no check in either
  parser, so a loop a reader counted as eleven adjudicated as nine. **An example
  is not a mechanism, and it is not even a policy** — nothing could break, because
  nothing was stated to break.
- **Sweep question**: what fails, loudly, when this rule is broken?
- **Found by the first sweep this registry demanded**: `refactoring-safety`
  already required fixing the class rather than the instance, for code faults.
  The rulebook broke it four times running because nothing scoped that duty to
  RULES and nothing checked it. The duty existed; the mechanism did not — which
  is the class, exactly.

### archaeology-in-live-surface
A surface an agent reads carries prose about a thing that NO LONGER EXISTS — a
retired verb, a deleted file, a mode that was removed — so every seat spends
context learning what it cannot use. The obituary outlives the thing and reads
as current, because nothing distinguishes "this is how it works" from "this is
how it used to work" except tense.
- **Instances**: `blue confidence` carried a RETIRED notice in two constitutions
  and a dead `calibrationClause` that returned `''` and was never called; the
  legacy prompt-set mode described at length in a refusal for a mode that cannot
  be selected; a `--bin` flag documented as belonging to a capture that is gone.
- **Sweep question**: does the subject of this prose still exist? If it does not,
  the git history holds it and the live surface should not.
- **Neighbour**: `prose-carried-harness-trivia` — there the prose describes a
  REAL condition a hook could answer in place; here it describes nothing that
  exists at all.
- **Distinguisher**: could a seat ACT on this sentence today? If the thing it
  names is gone, no action exists and it is archaeology.

### prose-carried-harness-trivia
Harness guidance carried in every prompt that a hook could answer at the moment
it bites. The prose costs every seat every round; the condition it describes
fires rarely, and a paragraph read at seat start is not present when the tool
call actually fails.
- **Instances**: the Glob/Grep working-directory limit and its sanctioned
  fallback; the Grep-counts-LINES footgun; the heredoc-eats-backslashes footgun.
- **Sweep question**: could a hook detect this condition and answer it in place?

### format-selects-audit-surface
A formatting or notation convention silently decides what gets audited.
- **Instance**: `claim_count` defined as FOOTNOTED claims and the citation ledger
  keyed on footnote labels, so unfootnoted prose receives zero adversarial
  contact (E0.5h).
- **Sweep question**: what material does this convention EXCLUDE from the audit
  surface, and is that exclusion intended?

### invariant-at-wrong-level
An invariant is enforced at a surface COARSER than the thing it protects, so the
rule costs more than it needs to and proves less than it claims. The enforcement
is real; the altitude is wrong.
- **Instance**: "never subtract substance" protected CLAIMS by forbidding edits
  to PROSE. It stopped run 3's real failure (blue quietly dropping content under
  repair pressure) at the price of a report that could only grow — 1178 to 1668
  lines in one run — with every audit seat re-reading all of it every round. And
  it still could not prove what it claimed: an edit could remove a claim
  silently, because prose has no arithmetic.
- **Sweep question**: what is the SMALLEST thing this invariant actually protects,
  and can that thing carry the check itself?
- **Neighbour**: `policy-without-mechanism` — there the rule has no enforcer at
  all; here it has one, aimed at the wrong surface.
- **Distinguisher**: is anything enforcing this today? If nothing is, it is
  policy-without-mechanism. If something is, but it guards a coarser surface than
  the invariant needs, it is this.

### co-resident-rules-disagree
Two rules that are ALWAYS loaded together fire at the same moment and give
opposite instructions, so an agent obeying both literally must violate one. The
cost is not the wasted words: it teaches that the rulebook is approximate, which
discounts every other rule in it.
- **Instance**: `semantic-consent` said "BEFORE acting on a vague instruction,
  YOU MUST ask for the missing intent" while its own next clause said to split by
  reversibility and DECIDE the reversible ones — the second written specifically
  to stop the stalling the first caused. In the same set, `semantic-consent`
  required stating the goal of every read-only search while `terse-communication`
  forbade narrating process; every `Grep` fired both.
- **Sweep question**: which OTHER always-on rules fire at the same moment as this
  clause, and do they agree? A clause is not finished until that is checked
  against the rules it will be loaded beside.
- **Neighbour**: `invariant-at-wrong-level` — there ONE rule guards a surface
  coarser than the thing it protects; here TWO rules guard the same moment and
  disagree about it.
- **Distinguisher**: is there a second rule firing at the same moment? If the
  defect is one rule aimed at the wrong altitude, it is invariant-at-wrong-level.
  If two rules collide, it is this.

### port-retarget
A caller still invokes the runtime a port has replaced — a spawn/invocation site
names the old script while the new implementation is the contract — so the pointer
and the implementation drift, and the sibling call sites that share the old runtime
are the ones that quietly keep pointing at it.
- **Instance**: `/research` step 2 spawned `node setup-research-run.mjs` after
  setup was ported to the `feov-record setup` subcommand; the sibling spawns
  (`capture-research-run.mjs` at step 5, `render-run-dashboard.mjs` at monitoring)
  still call `node`, ported in later #121 slices.
- **Sweep question**: which OTHER invocation sites still name the ported-away
  runtime, and are they retargeted or explicitly deferred with a reason?
- **Neighbour**: `staged-not-delivered` — there an artifact exists but never
  reaches the consumer; here the invocation reaches a consumer that has MOVED.
- **Distinguisher**: is the call's TARGET gone/relocated (port-retarget), or
  present-but-never-consumed (staged-not-delivered)?

### misattributed-enforcement
A surface tells a reader — a seat, or a human auditing coverage — that X enforces
this duty, while the enforcement actually lives at Y, or nowhere. Nothing breaks.
The rule holds, the gate refuses, the run completes. What is wrong is the reader's
model of WHERE the refusal comes from, and that is load-bearing: a seat calibrates
how much to trust each surface, and an audit counts checks.
- **The harm is double-counting**, which is why it survives review. The write and
  the named reader look like two independent checks. There is one. Coverage reads
  as 2/2 while the second is dead, unrun, or answering a different question — and
  removing the live one then looks safe.
- **Instances**, all measured 2026-08-15 in `debate.js` and its prompts (#417):
  - `found_by`: the red-merge prompt said *"verify checks each names a recorded
    finding"*. `requireFindings` refuses it AT THE WRITE, in the seat's face.
    Enforcer real, credited to the wrong component.
  - `supersedes`: the engine declined to fail a `closed_with_regression` closure
    with no successor, stating *"`verify`'s supersedes-resolve runs over the board
    at capture"*. Wrong twice — nothing runs it at capture, and
    supersedes-resolve asks whether a NAMED ancestor exists, not whether anything
    names this closure. Even where it runs it does not answer this question.
  - the archive spot-check floor: the prompt said verify *"fails the run"*.
    `record.SpotCheckAudit` reaches the REPORT, which renders the debt
    deliberately — `assemble.go` says so in its own words — and nothing fails.
- **Sweep question**: for each surface that names a checker, does that checker
  RUN on this path, and does it check THIS? Both halves, separately: a checker
  that runs and answers a neighbouring question is still this class.
- **Neighbour**: `policy-without-mechanism` — there nothing enforces the rule at
  all; here something does, and the text credits something else.
- **Neighbour**: `port-retarget` — there a live INVOCATION names a target that
  moved, and it breaks or misfires. Here nothing is invoked: the text is a
  CITATION, so there is no failure to observe, only a wrong belief.
- **Distinguisher**: is anything enforcing this today? Nothing →
  `policy-without-mechanism`. Something, and the text is a call to it →
  `port-retarget`. Something, and the text merely NAMES a different component →
  this.
- **The pattern the sweep found, worth carrying**: across `debate.js`, the three
  agent constitutions and the protocol skill, every claim naming THE WRITE was
  true and every claim naming an after-the-fact reader was false. Where a prompt
  names a checker, that checker was sound exactly when it was `record.Append`.
  Start the sweep there.
- **Caution, from getting this wrong while filing it.** The audit for this class
  is itself prone to under-counting: #415 was published claiming `verify` was
  invoked nowhere, because the search covered importers of the package and
  invocations in `.md`/`.js`/`.mjs`/`.yml`. `internal/fuzz` invokes it as its
  primary oracle by shelling out to the built binary from a `_test.go` — a
  mechanism the search did not enumerate. Before concluding a checker is unrun,
  enumerate the INVOCATION MECHANISMS (import, subprocess from a test, shell, CI
  step, prompt text), not the ones that came to mind.

## Minting a new class

Same discipline as the gap registry: a new class needs a slug, a one-line
definition, its nearest NEIGHBOUR, and the DISTINGUISHING question that tells the
two apart. A class that cannot state its neighbour is usually an instance
wearing a class's name.
