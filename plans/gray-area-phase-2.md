# Gray Area Phase 2 — scope proposal

> **Status: RULED ON 2026-08-17. The operator overruled the recommendation.** See §0.
>
> Written because the operator asked for a scoped proposal to accept, cut down, or redirect,
> rather than for direction to be invented. They redirected it.
> Parent: [`gray-area.md`](gray-area.md) §7.
>
> The recommendation was to build roughly half of what Phase 2 was specified as and defer the rest.
> **The operator chose the whole of it, gated on #189 instead.** The original argument is kept below
> rather than rewritten, because a proposal that quietly becomes the decision it lost teaches nothing.

---

## 0. The ruling, and what it overturns

**Option B: fix the substrate (#189) first, then build Phase 2 whole.** Not Option C, which this
document recommended and marked ✅.

The recommendation traded coverage for availability: build the half whose substrate is complete,
accept that seat-level behaviour goes unmeasured, and wait for #189 to resolve on its own. The
ruling refuses that trade, and the reason it is the operator's to make rather than mine is that it
turns on what the plugin is FOR. Seat-level behaviour — what each debating role actually did, versus
what it claimed — is the product. A Phase 2 that inspects only the lead agent is a working tool
aimed at the less interesting half, and shipping it first would relieve exactly the pressure that
gets #189 solved.

**So #189 is no longer a blocker to route around. It is the work.** Its risk is stated plainly and
was already in §Option B: three hypotheses have failed (environment, session, seat-as-cause — the
third descriptive rather than causal), and nobody can estimate it. That risk is accepted, not
dissolved.

**Second ruling: act-vs-claim DOES point at PR bodies, with full post-hoc adjudication.** The
narrower options — self-check before posting, or validation-sections-only — were declined. Findings
are recorded against merged pull requests and readable by the operator.

That settles the question §7 left open, and it settles it in the direction that costs the agent
most: a PR body stops being testimony and becomes a record under inspection. The case for it is
measured rather than theoretical — in the session that wrote this document I told the operator
"nothing was published" when three releases had already gone out, and quoted a plugin version that
had moved two hours earlier. Both are exactly what adjudication against the trajectory catches.

The failure mode that comes with it is real and is now the builder's problem to hold: a false
positive on an accurate body puts the operator between two of the agent's outputs. §6's check 4
covers the absence case; a PR-body reader will need its own answer for the false-positive case, and
`NO-EVIDENCE`'s asymmetry (state the search, let the reader convict) is the shape to start from.

## 1. What Phase 2 was specified as

From `gray-area.md` §7:

> **2. The miner** — Go CLI over the manifest: act-vs-claim, rework, stalls. Provenance on every
> answer.

Three inspections bundled under one phase. That bundling is the thing this proposal argues with.

## 2. What has changed since it was written

Two things, and both move the boundary.

### 2.1 One third of Phase 2 already shipped, under Phase 5

Phase 5 (`internal/claims`) adjudicates **checkpoint claims against the trajectory**. That *is*
act-vs-claim, for one claim source. It shipped first because continuity shipped first, and it
established the substrate the rest of Phase 2 would reuse:

| shipped | what it gives Phase 2 |
|---|---|
| `trajectory.Parse` → citable `Event` with `File`/`Line`/`UUID` | the reader, with `Cited()` as a hard gate |
| `Verdict` = `CITED` / `STALE` / `NO-EVIDENCE` / `UNCHECKABLE` | a four-state vocabulary that already refuses to call absence a fact |
| `Finding` carrying `Searched` and `EventsSeen` | the G4 discipline — the tool states its search and the reader convicts |
| `gray-area checkpoint <note>` | a working command shape, exercised on real data |

**So Phase 2 does not start from zero, and should not re-invent the verdict model.** The
`NO-EVIDENCE` asymmetry in particular was argued once and should not be re-argued per inspection.

### 2.2 The substrate is blind to 72% of seats (#189)

Measured 2026-08-15, `gray-area.md` §11.10, across all 69 `kind: "seat"` rows of one manifest:

| | `agent_type` populated | `agent_type` empty |
|---|---|---|
| `resolved: true` | **19** | 0 |
| `resolved: false` | 0 | **50** |

Zero exceptions, both populations interleaved seconds apart in one session. Checked against disk
rather than against the flag: 0 resolved-but-missing. **What produces the typeless population is
undetermined** — six probe agents across four spawn shapes all carried a type and all resolved, and
the one attempt to identify the others was contaminated and discarded.

This is the fact that decides the scope, and it is easy to read past because it does not look like
a blocker — capture succeeds, the manifest is written, rows appear. They just describe seats whose
transcripts do not exist.

---

## 3. The split that falls out of it

Phase 2's three inspections do **not** share a substrate. They divide by *which transcript they
need*, and that division is clean:

| inspection | needs | substrate today |
|---|---|---|
| **act-vs-claim** (main session) | the parent session transcript | **complete** — Phase 5 cites it today |
| **rework** — same file edited N times, same test re-run after a claimed fix | parent transcript for the lead's own work; per-seat for a seat's | complete for the lead, **28%** for seats |
| **stalls** — time between acts, retry loops, abandoned lines | mostly per-seat | **28%** |

The parent transcript is not a special case: it is the file Phase 5 already reads, and the one the
lead agent's entire tool history lives in. Everything the *lead* did is fully observable now.

---

## 4. Three options

### Option A — build Phase 2 as specified, over 19/69 seats

Honest only if every seat-scoped output carries its coverage. The risk is not inaccuracy, it is
**a plausible zero**: "no rework found in this run" reads identically whether the run was clean or
whether three of four seats had no transcript to read. That is `facts-are-fields` clause 3 aimed at
the plugin's own output, and it is the failure mode this plugin exists to prevent in others.

Buildable, but it puts the burden on a caveat rather than a mechanism.

### Option B — fix the substrate first (#189), then build Phase 2 whole ✅ CHOSEN

The honest sequencing if per-seat inspection is the point. The cost is that #189 is **undetermined
after three failed hypotheses** (§11.8 environment, §11.9 session, §11.10 seat — the first two
wrong, the third descriptive rather than causal). Nobody can estimate this, and gating Phase 2 on it
means Phase 2 may not happen.

### Option C — build the LEAD-SCOPED half now; hold the seat-scoped half behind #189 — RECOMMENDED, NOT CHOSEN

**Recommended at the time, and overruled — see §0.** Build the inspections whose substrate is complete, and do not pretend the others
are one flag away.

Concretely in scope:

- **act-vs-claim beyond checkpoints.** Phase 5 adjudicates one claim source. The same engine
  pointed at other declared artifacts — a PR body's validation section, a plan's §V loop, a commit
  message asserting a check ran — is the same shape with a different reader in front of it.
- **rework, lead-scoped.** Same file written N times in a session; a test command re-run after a
  claimed fix; an edit reverted and re-applied. All of it is in the parent transcript.
- **stalls, lead-scoped.** Repeated identical tool invocations, and the [[anti-spinning]] 3-strike
  rule adjudicated against what actually happened rather than against the hook counter — which
  counts only TOOL failures and is documented as blind to a command that exits 0 while the work
  failed.

Explicitly OUT, and tracked rather than forgotten:

- every per-seat inspection, until #189 resolves or a coverage mechanism lands;
- Phase 3 (instrumented reasoning) and Phase 4 (bench symmetry), untouched by this proposal.

**The completeness argument.** Option C is a complete concept rather than a truncated one because
lead-scoped inspection answers a whole question — *did the agent driving this session do what it
said* — without a per-seat answer. It is not half of an inspection; it is one of two populations,
fully covered.

---

## 5. What Option C would have to touch, enumerated

[[complete-the-concept]] requires the carrier census in the spec, and requires the auditor to re-run
it rather than trust it.

| carrier | change |
|---|---|
| `internal/claims` | generalise `Claim` past checkpoint notes, or add a sibling package; the `Verdict`/`Finding` model is reused unchanged |
| `internal/trajectory` | rework and stalls need *ordering and repetition*, which `ToolUses()` supports but nothing yet groups |
| `cmd/gray-area` | new verbs beside `checkpoint`; the help contract and its goldens move |
| `plugins/gray-area/README.md` | the shipped documentation |
| `commands/audit-checkpoint.md` | a command doc, and the model for any new one |
| `.claude-plugin/plugin.json` | **no version bump** — versions move at a release boundary (#444) |
| `scripts/check/gates.go` | only if a new CI gate appears |
| `plans/gray-area.md` §7 | the phase table's Phase 2 row, which this proposal would rewrite rather than leave contradicting the build |

## 6. Validation loop, written before implementation

Per [[validation-loop]], with what re-arms each:

1. `go test -C plugins/gray-area/tools ./...` — clean · re-armed by `plugins/gray-area/tools/`
2. `plugins/gray-area/bin/gray-area checkpoint .claude/checkpoints/CHECKPOINT.md` — still exits 0
   with every claim adjudicated · re-armed by `internal/claims/` — **the existing Phase 5 check must
   not regress while its engine is generalised**
3. Every new inspection refuses to emit a row whose `Cited()` is false — asserted per inspection,
   not once · re-armed by `internal/trajectory/`
4. An inspection run against a session with **zero** matching events reports its search and its
   `EventsSeen`, and is distinguishable in output from one that found nothing wrong · re-armed by
   `internal/claims/` — this is the plausible-zero gate and the one most likely to be dropped
5. `go run ./golden -review` — the help contract moves deliberately · re-armed by the goldens

Check 4 is the one worth stating loudly: it is the criterion Option A could not meet by
construction, and it is what makes Option C's narrower scope honest rather than merely smaller.

## 7. What this proposal does not settle

- **Whether `friction.md` stays seat-authored with the miner adjudicating it, or becomes
  miner-derived.** `gray-area.md` §9 left this open and this proposal does not close it.
- **The acceptance test for the plugin as a whole**, also open from §9.
- ~~**Whether act-vs-claim over PR bodies is wanted at all.**~~ **SETTLED 2026-08-17: yes, with full
  post-hoc adjudication.** See §0. What remains open is how a PR-body reader avoids convicting an
  accurate body — the false-positive direction, which no existing verdict covers.
- **Phase 2's original acceptance test remains unavailable.** §7 already records that the historical
  hand analyses cannot be re-derived because those transcripts were never committed. Option C does
  not recover it; the first captured run remains the real fixture.
