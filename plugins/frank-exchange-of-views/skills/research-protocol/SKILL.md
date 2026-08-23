---
name: research-protocol
description: Use when performing or auditing deep research — the protocol (frontier hypotheses, saturation, disconfirming budget, semantic footnotes), the run-directory layout, and the debate envelopes.
---

# research-protocol

Research that survives an adversary.

## Protocol

- BEFORE searching, YOU MUST formulate 3–5 frontier hypotheses — what would be true if each candidate answer were right — and record each one as a LINE OF INQUIRY on the record — the approach, and what would be true if it paid off; searches then test hypotheses instead of wandering. On the record rather than in a file, because a hypothesis red cannot rule `too-thin` or `out-of-scope` is one nobody can contest — and the round-0 hypotheses are the ones that shape the entire run.
- During research, YOU MUST search to **saturation**: stop only when new searches return already-seen sources (typically 20–30 searches for a deep topic).
- During research, YOU MUST spend at least one search in five hunting **disconfirming** evidence against your current position. This is a drafting floor, not the verification: it keeps confirmation bias out of the draft; systematic disconfirmation is red's entire job.
- During writing, YOU MUST add every citation with the TOOL, against the exact sentence it backs — never by hand. The tool fetches the source once into the run cache, then splices an INVISIBLE, IMMORTAL `<!--cite:c-…-->` anchor at that sentence; assembly weaves the anchors into the visible `[^N]` footnotes and composes the `## Bibliography`. A hand-typed `[^label]` is not a citation: nothing backs it, the claim counter does not see it, and the unbacked-citations detector flags it. An unreachable source is unusable — the cite is rejected and logged as friction.
- **The bibliography is BOTH teams'.** A red CORROBORATION — a source red went and found for a
  claim blue made — mints an anchor and joins the footnotes the same way when it SUPPORTS the
  claim. A reader cares that the text has appropriate references, not which seat inserted them.
  A `refutes` or `absent` reading is not a reference backing the sentence and is never spliced: it
  is red finding the text unsupported, which is a defect, and it goes to the board as a finding —
  a PASS is refused until one is raised for it.
- **An anchor is part of the text you edit.** The report as the tool renders it prints every
  `<!--fx:…-->`, `<!--cite:…-->` and `<!--proof:…-->` as it is; quote the span as printed and copy each token
  into the replacement like any other character. A quote that stops just short of the anchor on
  the sentence it rewrites is refused, and the refusal names the token to carry. The anchor is
  never lost, and an edit that moves the words under one REOPENS it — the reference stands, its
  referent moved, so a verification of it is stale rather than refuted.
- AFTER drafting, every claim MUST trace to a source a skeptic can follow; unverifiable claims are labeled as such, not laundered into fact.
- For PDF-only sources, YOU MUST try the document-extraction MCP tools before grading down on a lossy fetch: `arxiv-latex` (exact LaTeX for arXiv figures/tables) and `pdf-reader` (page/table extraction with provenance) — discoverable via ToolSearch when the project's `.mcp.json` servers are approved. Two runs of friction ranked lossy PDF fetches the #1 capability gap; a claim capped at "unable to corroborate" without trying these is an incomplete audit.

## The exchange is TOOL-MEDIATED

Everything the two sides exchange — findings, closures, citations, proofs, revisions,
lines of inquiry, disputes, friction, opinions — is an **event on the record**, written through a
verb that can refuse it, and read back through a projection. This is the governing clause
of the protocol, not a storage preference: a hand-written file is an exchange nothing
validated, and a fact recovered from a filename or a prose substring is one only pretending
to be mediated. Both fail the same way — **by returning a plausible zero**, which reads
exactly like a clean board. See [[facts-are-fields]].

`report.md` is the instructive non-exception: prose, because its audience is human. But
every point of *argument* in it carries a tool-placed anchor — `cite:` where a source backs
a claim, `fx:` where red challenged, `proof:` where a computation settles it — and dropping
one is a hard refusal. Write for the reader; put what the machinery depends on in a field.

## The run directory (the blackboard)

**The tool is the read path.** Where a line below says RECORD, that artifact has no
authoritative file — read it with `show <name>` and never from disk — `show` is a GROUP, so `show --help` lists every projection.

```
research/<date>_<slug>/
├── records/           # THE EVENT LOG — the source of truth; one append-only shard per seat
│                      # (MAY live outside the run entirely; a `.records-elsewhere` note appears
│                      #  here instead. Nothing changes for a seat, because a seat reads the
│                      #  record with `show <name>` and never from disk — which is the
│                      #  point: a run can be configured so that is the ONLY way, and then a
│                      #  missing verb has to surface as friction instead of a workaround)
├── report.md          # final deliverable — assembled LAST, by union (authored)
├── inputs/PINNED.md   # the evidence base, pinned: repo HEAD at launch + cited corpora's commit/round
├── blue/
│                      # (the opening hypotheses are LINES OF INQUIRY on the record, not a file — read
│                      #  read them as the `lines-of-inquiry` projection. A hypothesis in a file is one red
│                      #  cannot rule too-thin or out-of-scope, and the round-0 ones shape the
│                      #  whole run)
│   ├── report.md      # blue's LIVING report — grows every round, never summarized away.
│   │                  #   Authored prose, but every EDIT after round 0 goes through the `edit` verb
│   └── candidates/    # best-of-N method-lens lane drafts, preserved (authored)
└── cost.md            # measured tokens + dollars per seat-round (feov-record cost)

RECORD — no file at all; read through the tool. Every projection, what each is for, and the
verb that WRITES each one are in your role's `--help`, which is generated from the command tree
and cannot disagree with it. A catalogue here would be a second copy that can.

  run this first, and again before you stop: your work — everything open to you, each item
  saying whether it is what blocks you closing, plus whether the sitting may close at all.
  `complete: true` with items still open means the gates are satisfied, NOT that you are done.

trajectories/       journal.jsonl (the HARNESS's lifecycle record, tracked)
                    + agent-transcripts.tar.gz (gitignored)
```

`setup` lays down exactly two stubs — `report.md` and `blue/report.md` — because those are the
two a later seat actually fills. **A stub is not an artifact**, and a stub nobody fills is worse
than an absent file: it reads as an empty artifact rather than a missing one. Measured in the
2026-08-05 run, stubs for the transcript and the citation ledger finished at 36 and 46 bytes
while the record held 122 events; both are projections now, with no writer and no stub. Anything
under RECORD above has no file at all — read it with `show <name>`.

**Termination is judged, and the standing practice is stop-and-resume**: `maxRounds` is a cost
ceiling, never the terminator of record. Red owns PASS/FAIL — *is it defensible*. **The bench
owns the stopping judgment** — *is it close enough*, the one call that weighs remaining defect
against remaining cost, and the only terminal value (economy) that otherwise has no organ. It
reads the telemetry projection — the series, never a snapshot — and files a reasoned,
cost-stated opinion; the operator acts on it,
stopping a run past its value and resuming with a reduced `maxRounds` for the honest UNVERIFIED
assembly — cache replay makes the stop ~$0 (measured). **Stopping is not passing**: the verdict
stays UNVERIFIED with the open count stated. Automatic severity-floor termination was evaluated
and REJECTED (run-4 report §1): it automates the one call that belongs to judgment. NEVER
change models on the resume.

All artifacts are git-tracked; nothing is summarized away. The payload is the file; the envelope is the handle — no large content travels through agent return values.

## Reading the corpus — two access modes, never confused

There is no search index, and there are two access modes:

1. **Full read for the document under audit** — red reads blue's living report whole, in
   context, every round. A snippet NEVER substitutes: a decontextualized quote is how audits
   go blind. This clause outranks any token saving.
2. **Leaf-node fetch for verification** — a citation is checked against its source, never against a
   summary. For a source BLUE CITED, read the exact bytes blue read from the run cache
   (a cache hit, so you audit the same artifact, not a page that may have drifted since). For a source you discover yourself, pull it verbatim (Bash `curl`, PDF MCPs).
   WebFetch is not used: it returns a summary, not the source.

To find text inside the run's own artifacts, use `Grep` — the terms you want are the terms you
already have, and a lexical match over a known file beats a ranked guess over a corpus.

## Harness contract (one referenceable paragraph — three seats re-derived this at token cost)

The Workflow script's `log()` is operator-console-EPHEMERAL: it persists nowhere. The
transcript directory's `journal.jsonl` is the HARNESS's lifecycle record — `started`/`result`
events only, never script logs. Per-agent API transcripts are `agent-*.jsonl` (the cost
audit's input). Durable in-run state lives ONLY on the blackboard (git-tracked run files) or
in envelopes; anything else evaporates with the session. Tool footguns with live recurrences:
Grep's count mode counts LINES, not occurrences (anchor patterns when counting); quoted
heredocs can eat backslashes (prefer the Write tool for scripts); the Read tool caps ~25k
tokens — a full-document read over that cap is consecutive whole windows, which satisfies the
full-re-read MUST without a confidence discount.

## Report structure

The final `report.md` (see `references/report_template.md`): verdict stamp (VERIFIED/UNVERIFIED + rounds) → **the Catechism** (`references/catechism_template.md` — the worth-our-time decision, adapted from Heilmeier) → analytical core (foundations / analysis / risk matrix graded likelihood × impact × complexity, including risk-accepted items with rationale) → **blue's report in full** → **red's findings in full** → debate record → **open questions carried past this run** (blue's final envelope, verbatim) → footnotes (with access dates; volatility noted for living sources).

## Friction (the complaint channel)

A subagent's only voice is its return value — so capability complaints travel in the envelope.

- AFTER any task where a missing tool, denied permission, or capability gap impeded you, YOU MUST report it in the envelope's `friction` field: name the capability and what you would have done with it.
- AFTER any task where the material did not fit the shape you were given — a template section that made no sense for the topic, a protocol step that fought the work, an envelope field you had nothing honest to put in, content with no home — YOU MUST report the misfit as friction: name the template/step/field and what shape the work actually wanted.
- YOU MUST NOT silently work around a capability gap — the workaround destroys the signal that would get you retooled.
- Friction is recorded on the RECORD via the friction verb — every seat WRITES it, and the read is the OPERATOR's, on the operator's own surface and not on yours, because a capability gap is a report to the human who can retool the seat, not material for the debate; capture reconciles every envelope's `friction` field against it, and the self-improvement loop consumes it. Complaints are how the system learns what its agents actually need.
