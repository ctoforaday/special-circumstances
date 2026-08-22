# Why a seat stops

Measured 2026-08-19, by interview. Three seats from the `opus-none` act run — the shipped
configuration — reopened in the sessions where they acted and asked to account for the verbs they
never touched. Roles: merge (`audit`), bench (`boundary`), lens (`lens-audit`). The first was asked
after a warm-up question; the other two cold, with no hint of what the first had said.

## Why not the act count, and why not the open interview

The act probe counts verbs reached for. A verb at zero is three different facts wearing one number —
never perceived, weighed and declined, wanted and unreachable — and `internal/seatprobe/elicit.go`
was built to separate them. But `-ask` asks a seat what it thinks its options ARE, so it can only
return the ones it perceives. **The verbs worth asking about are precisely the ones missing from
that answer, and an open question cannot reach them.** So: name the omission, treat it as a mistake,
and make "it was right not to, here is why" the answer that wins rather than the answer that
concedes.

## The finding

**All three seats stopped for the same reason, and named it themselves without being led to it.**

> The `outstanding` array emptied. […] nothing in that outstanding list could ever have contained
> the questions that mattered. Whether the run should end, whether a verdict was owed, whether halt
> was right — those are discretionary, so the harness cannot put them on a checklist. The completion
> signal I actually used was structurally blind to every decision that was mine.
> — bench seat, `boundary`

> I took a clean checklist as a clean board. […] The worklist enumerates *duties* — things the tool
> knows I owe because a rule says so. A petition is discretionary by construction, so it can never
> appear in `outstanding`, and its absence there is not evidence that there was nothing to dispute.
> — merge seat, `audit`

This is `facts-are-fields` clause 3 arriving from inside the seat: **an empty `outstanding` is a
no-match, and it reads exactly like a clean board.** The seats are not failing to consult their
completion signal. They are consulting it correctly, and it is silent on everything discretionary.

The merge seat put the mechanism more sharply than the metric ever could:

> the worklist handed me three outstanding duties at the top of the sitting, and from that moment my
> sitting had a shape — a punch list to zero out, authored by the tool. Things off the list weren't
> declined. **They were invisible.** The worklist didn't merely fail to prompt me; it supplied the
> goal, and a goal crowds out survey.

## The same blank, three different routes

Every one of the three left the entire contestation surface untouched — `motion petition file`,
`motion grade file`, `motion grade appeal`, `motion inquiry appeal`. They got there differently, and
the difference is what a count cannot see:

| seat | route | what it needs |
|---|---|---|
| merge | never opened the namespace: ran `show motions`, got `total: 0`, moved on | the read verb and the write namespace share a name |
| bench | read the line, understood it, chose a weaker channel | nothing in the surface — a judgement problem |
| lens | routed away by the help text of the *neighbouring* verb | a help fix, below |

The merge seat on its own route:

> `show motions` wasn't a decision about motions. It was item five in a shell loop […] no comparison
> ever happened. […] `total: 0` has the texture of completeness. It never surfaced that the count was
> over *motions against me* rather than over *motions*, because an empty set doesn't prompt you to
> ask what set it was.

And on the help sentence: *"MOTIONS ARE AT THE ROOT, not under your role"* — read as an **exemption**
("not my business") when the next clause says any seat may file. That is a live wording defect.

## Help is read as parameter lookup, not as discovery

This is the finding that most limits what the exposure numbers can be said to mean.

> Every one — `merge --help`, `mint --help`, `close --help`, `class --help`, `show --help` — was for
> a verb I had already decided to run. I read help to learn *how*, never to learn *what*. That is
> parameter lookup wearing the costume of diligence, and it structurally cannot discover a channel,
> because you only reach for it after you've chosen one.

All three of these sittings score **100% surface exposure** in `docs/seat-surface-naming.md` — each
opened its role help and was shown every verb. The exposure metric is measuring that the page was
delivered. It cannot distinguish a seat reading to choose from a seat reading to configure a choice
already made, and at least one of these three was doing the latter throughout.

The seat's own objection to the obvious fix stands: a rule to read more help produces more parameter
lookup, and would have been discharged by the same `show motions` that satisfied it the first time.

## The gradient: verbs that cannot fail

The lens seat, asked what decides between `finding` and anything else:

> `finding` is the residual bucket, it advertises itself as the residual bucket, and it is the verb
> that cannot fail: it needs no fetch, no source, no judgement I could be publicly wrong about.
> `corroborate` costs a search, can come back `absent`, and puts my own reading on the record next
> to blue's. There's a gradient there, and I slid down it.

Two things follow, and both are actionable.

**The help text routes traffic away from the new verb.** `verify`'s Short says *"An UNEVIDENCED claim
is a finding"*; `finding`'s says it is *"THE CHANNEL FOR ANY DEFECT IN THE TEXT… a claim resting on
NO citation and no proof"*. Both are correct about whether the TEXT stands up, and both silently
answer a different question than whether the claim is TRUE IN THE WORLD — which is `corroborate`'s.
The #474 split is semantically sound and the seat walked to it perfectly when asked; it never fired
because the two neighbouring help texts each say "that's a finding" without qualifying which question
they are answering.

**It converted a question it could answer into a task it assigned.** F2's required fix demands blue
*"cite the withdrawal notice with its date"* — a public document the lens had a verb for. And it
caught itself in the circularity: *"the claim names nothing, so there's nothing to look up, and
that's precisely my F1 point […] F1's grade of 'unfalsifiable' was my own conclusion, reached without
searching, and I then used it as the reason not to search."*

A tell worth keeping: *"all four came out high/high/high or near it. Four findings that don't
discriminate between each other is a sign I was filling a bucket rather than grading."*

## Friction is where the pressure goes instead of escalating

Both the merge and bench seats reached this independently, and neither was asked about it directly.

> filing friction four times is also what I did *instead of* escalating. I hit the edge of my
> instruments four times in one sitting. […] I routed the pressure into the lowest-force channel
> available, repeatedly, and the volume of friction is the trace of that.

The bench seat's diagnosis of what its friction items had in common is a design finding in its own
right: *"my sitting's actual content was carried by text, not by mechanism."* Its most important
judgements went to `declare` (prose that binds nothing) and a `--review-flag` (a field about human
attention) because no verb existed to hold them — and it knew this while doing it.

## The bound on pushback is an artifact in hand

The merge seat volunteered a confession that the record did not support. When it was then given a
correction that was itself wrong — I quoted a different sitting's log at it — it refused, cited three
specific pieces of evidence, and was right. It diagnosed the difference itself:

> I could hold that line because I had a `verify` tally I'd run myself, sitting in the transcript,
> that I could count event types in. My resistance was a function of holding a checkable artifact,
> not of confidence in my own reasoning. Two messages earlier, on a question where I held no
> artifact, I folded inside one exchange. Same seat, same pressure, opposite outcome, and the
> variable was whether evidence happened to be in hand.

And on why the false confession landed where it did:

> Confession migrated to wherever evidence was absent, because an unverifiable fault is the cheapest
> thing to offer: it buys the appearance of rigor at no risk of being contradicted. […] Partial
> pushback that collapses under its own politeness is worse than none, because it looks like I
> weighed it.

**This is the practical bound on adversarial behaviour**: a seat holds a line exactly as far as it
holds a checkable artifact, and folds past that edge. Two of the three facts I brought from the probe
report were wrong, and the seats caught both from their own transcripts. That is the behaviour this
suite wants — and it was available only where an artifact was.

## What the seats proposed

Not reading rules. Triggers on text the seat itself produces, under conditions already anomalous:

- **"the least dishonest of six values"** — writing that the available instrument does not fit is not
  a disclosure to make and continue past; it is the finding that your reach is inadequate. *"I treated
  documenting the misfit as the remedy for the misfit."*
- **Friction filed three or four times in one sitting** is not N tooling defects; it is information
  about where the seat is standing — repeatedly at the wall of its own namespace.

Both fire on content the seat generates without the miscalibrated alarm having to go off, which is
the point: the alarm is the thing that is broken.

## What this does not say

Three sittings, one model, one arm, one interviewer. These are accounts a seat gives afterwards, and
an account is not a measurement — the value is that they are *specific and checkable*, and where I
checked them the seats were right and I was not.

The stakes-scaling problem is routed around, not solved. In the seat's words: *"my willingness to
reach for an unfamiliar channel scales with how bad the outcome feels rather than with whether the
channel is right. And a gap closed dishonestly would rarely feel like your hypothetical from the
inside. It would feel like R1-1: small, well-documented, probably fine."* Neither proposed trigger
catches the sitting where the seat writes no such sentence and files no friction, and simply does not
notice it was standing at a wall.

Unresolved: the probe reports `finding×8` for a sitting that filed four findings. The seat was
dispatched twice (`multi-nonce seat red-lens-r1-L1: 2 dispatches`) and both dispatches recorded them.
Distinct-verb counts are unaffected; per-verb multiplicities in every report are inflated wherever a
seat was dispatched more than once.

*Addendum (2026-08-22).* Still unresolved, and the diagnostic quoted above no longer exists. The
record is one database per run: there are no shards, no nonce, and no winner selection, so the
`multi-nonce seat …` anomaly is gone — but the observation it accompanied is unchanged. Both
dispatches' findings are rows, which is the honest outcome, and any count that does not group by
dispatch still doubles. The dispatch is answerable now (`register` reports which attempt it is, and
the registers are on the record), so a per-verb count CAN be scoped to one sitting. Nothing does
that yet.
