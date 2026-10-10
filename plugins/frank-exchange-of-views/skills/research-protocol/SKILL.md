---
name: research-protocol
description: Use when performing or auditing deep research — the protocol (frontier hypotheses, saturation, disconfirming budget, semantic footnotes) and the debate envelopes.
disable-slash-command: true
user-invocable: false
---

# research-protocol

Research that survives an adversary.

## Protocol

- BEFORE searching, YOU MUST formulate 3–5 frontier hypotheses — what would be true if each candidate answer were right — and record each one as a AVENUE on the record — the approach, and what would be true if it paid off; searches then test hypotheses instead of wandering. On the record rather than in a file, because a hypothesis red cannot rule `too-thin` or `out-of-scope` is one nobody can contest — and the opening hypotheses are the ones that shape the entire run.
- During research, YOU MUST search to **saturation**: stop only when new searches return already-seen sources (typically 20–30 searches for a deep topic).
- During research, YOU MUST spend at least one search in five hunting **disconfirming** evidence against your current position. This is a drafting floor, not the verification: it keeps confirmation bias out of the draft; systematic disconfirmation is red's entire job.
- During writing, YOU MUST add every citation with the TOOL, against the exact sentence it backs — never by hand. The tool fetches the source once into the run cache, then splices an INVISIBLE `<!--cite:C-…-->` anchor at that sentence; assembly weaves the anchors into the visible `[^N]` footnotes and composes the `## Bibliography`. A hand-typed `[^label]` is not a citation: nothing backs it, the claim counter does not see it, and the unbacked-citations detector flags it. An unreachable source is unusable — the cite is rejected, and the log is where you report it.
- **The bibliography is BOTH sides'.** A red CORROBORATION — a source red went and found for a
  claim blue made — mints an anchor and joins the footnotes the same way when it SUPPORTS the
  claim. A reader cares that the text has appropriate references, not which seat inserted them.
  A `refutes` or `absent` reading is not a reference backing the sentence and is never spliced: it
  is red finding the text unsupported, which is a defect, and it goes to the board as a finding —
  a PASS is refused until one is raised for it.
- **An anchor is part of the text you edit.** The report as the tool renders it prints every
  anchor, of every kind — `<!--fx:…-->`, `<!--cite:…-->`, `<!--proof:…-->` and `<!--gap:…-->` — as
  it is; quote the span as printed and copy each token into the replacement like any other
  character. Every kind lives one way: the tool places an anchor at the end of its quote, an edit
  carries it, the tool puts back one whose sentence the edit kept word for word, an edit that drops
  or types one is refused, and retiring its claim is the one way it leaves. A quote that stops just
  short of the anchor on the sentence it rewrites is refused, and the refusal names the token to
  carry. An edit that moves the words under an anchor REOPENS it — the reference stands, its
  referent moved, so a verification of it is stale rather than refuted.
- AFTER drafting, every claim MUST trace to a source a skeptic can follow; unverifiable claims are labeled as such, not laundered into fact.
- For PDF-only sources, a scanned PDF is read by the run's cached source read itself, and that reading is what a citation can locate a page in; YOU MUST read it there before grading down on a lossy fetch. For arXiv figures and tables, `arxiv-latex` gives the exact LaTeX. A claim capped at "unable to corroborate" without trying these is an incomplete audit.

## The exchange is TOOL-MEDIATED

Everything the two sides exchange — findings, closures, citations, proofs, revisions,
avenues, disputes, log entries, opinions — is an **event on the record**, written through a
verb that can refuse it, and read back through a projection. This is the governing clause
of the protocol, not a storage preference: a hand-written file is an exchange nothing
validated, and a fact recovered from a filename or a prose substring is one only pretending
to be mediated. Both fail the same way — **by returning a plausible zero**, which reads
exactly like a clean board. See [[facts-are-fields]].

The report set is the instructive non-exception: prose, because its audience is human. But
every point of *argument* in it carries a tool-placed anchor — `cite:` where a source backs
a claim, `fx:` where red challenged, `proof:` where a computation settles it, `gap:` where red
holds a gap open — and dropping one is a hard refusal. Write for the reader; put what the machinery depends on in a field.

## Reading the corpus — two access modes, never confused

There is no search index, and there are two access modes:

1. **Full read for the report** — red reads blue's report whole, in
   context, every sitting THE REPORT HAS MOVED. A snippet NEVER substitutes: a decontextualized
   quote is how audits go blind. This clause outranks any token saving, and the exception is not
   one: where the report stands at the head red already read, the whole-document read it is owed
   is the one already performed, and performing it again returns the same bytes to the same reader.
2. **Leaf-node fetch for verification** — a citation is checked against its source, never against a
   summary. For a source BLUE CITED, read the exact bytes blue read from the run cache
   (a cache hit, so you audit the same artifact, not a page that may have drifted since). For a source you discover yourself, pull it verbatim through the run's cached source read, so every seat that reads it after you reads the same bytes.
   WebFetch is not used: it returns a summary, not the source.

To find text inside the run's own artifacts, search for it lexically — with the search tool the
session offers, or the shell's `grep` where it offers none: the terms you want are the terms you
already have, and a lexical match over a known file beats a ranked guess over a corpus.

## Harness contract

The Workflow script's `log()` is operator-console-EPHEMERAL: it persists nowhere. The
transcript directory's `journal.jsonl` is the HARNESS's lifecycle record — `started`/`result`
events only, never script logs. Per-agent API transcripts are `agent-*.jsonl` (the cost
audit's input). Durable in-run state lives ONLY in the run directory (git-tracked run files) or
in envelopes; anything else evaporates with the session. Tool footguns with live recurrences:
a search's count mode — the search tool's or `grep -c` — counts LINES, not occurrences (anchor patterns when counting); the Read tool caps ~25k
tokens — a full-document read over that cap is consecutive whole windows, which satisfies the
full-re-read MUST without a confidence discount.

## Report structure

What a run hands over is a SET, not a file (see `references/report_template.md`), and every document
in it opens with a link bar to the others.

`report.md` is the research: verdict stamp (the word alone — its argument opens *Read this
first*) → **the Catechism** (`references/catechism_template.md` — the worth-our-time decision,
adapted from Heilmeier) → analytical core (foundations / analysis / risk matrix graded
likelihood × impact × complexity, including risk-accepted items with rationale) → the three
avenue sections → **open questions, left open by this run** (authored by blue into the report's `## Open questions`, audited by red every sitting it has moved, lifted verbatim) →
footnotes (with access dates; volatility noted for living sources).

The debate's own documents are beside it, one per audience: `docket.md` (the board in
full), `debate.md` (the transcript), `judgments.md` (motions and rulings), `avenues.md`
(the avenues and the path each took), `evidence.md` (the computations), `run.md` (the log, record verification, cost), `CHANGELOG.md` (the report's own
revisions and withdrawn claims). Nothing is summarized away by the split — the union is the
directory, indexed by `README.md`.

THE CITATION AND PROOF LAYERS ARE WOVEN PER DOCUMENT. A footnote definition cannot cross a file
boundary, so each document numbers and defines the references it actually carries; proof numbers
are run-wide, so `P3` is the same computation wherever it is cited.
