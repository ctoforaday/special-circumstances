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

## Minting a new class

Same discipline as the gap registry: a new class needs a slug, a one-line
definition, its nearest NEIGHBOUR, and the DISTINGUISHING question that tells the
two apart. A class that cannot state its neighbour is usually an instance
wearing a class's name.
