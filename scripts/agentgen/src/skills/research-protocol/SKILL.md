---
name: research-protocol
description: Use when performing or auditing deep research — the tool-mediated exchange, both sides' bibliography, how the report is read, and the report set's structure.
disable-slash-command: true
user-invocable: false
---

# research-protocol

Research that survives an adversary.

## The bibliography

- **The bibliography is BOTH sides'.** A red CORROBORATION — a source red went and found for a
  claim blue made — mints an anchor and joins the footnotes the same way when it SUPPORTS the
  claim. A reader cares that the text has appropriate references, not which seat inserted them.
  A `refutes` or `absent` reading is not a reference backing the sentence and is never spliced: it
  is red finding the text unsupported, which is a defect, and it goes to the board as a finding —
  a PASS is refused until one is raised for it.

## The exchange is TOOL-MEDIATED

Everything the two sides exchange — findings, closures, citations, proofs, edits,
inquiries, disputes, log entries, opinions — is an **event on the record**, written through a
verb that can refuse it, and read back through a projection. This is the governing clause
of the protocol, not a storage preference: a hand-written file is an exchange nothing
validated, and a fact recovered from a filename or a prose substring is one only pretending
to be mediated. Both fail the same way — **by returning a plausible zero**, which reads
exactly like a clean board. See [[facts-are-fields]].

The report set is the instructive non-exception: prose, because its audience is human. But
every point of *argument* in it carries a tool-placed anchor — `cite:` where a source backs
a claim, `fx:` where red challenged, `proof:` where a computation settles it, `gap:` where red
holds a gap open — and dropping one is a hard refusal. Write for the reader; put what the machinery depends on in a field.

## Reading the report

There is no search index.

**Full read for the report** — red reads blue's report whole, in
context, every sitting THE REPORT HAS MOVED. A snippet NEVER substitutes: a decontextualized
quote is how audits go blind. This clause outranks any token saving, and the exception is not
one: where the report stands at the head red already read, the whole-document read it is owed
is the one already performed, and performing it again returns the same bytes to the same reader.
The Read tool caps ~25k tokens — a full-document read over that cap is consecutive whole windows,
which satisfies the full-re-read MUST without a confidence discount.

To find text inside the run's own artifacts, search for it lexically — with the search tool the
session offers, or the shell's `grep` where it offers none: the terms you want are the terms you
already have, and a lexical match over a known file beats a ranked guess over a corpus. A search's
count mode — the search tool's or `grep -c` — counts LINES, not occurrences (anchor patterns when counting).

## Report structure

What a run hands over is a SET, not a file (see `references/report_template.md`), and every document
in it opens with a link bar to the others.

`report.md` is the research: verdict stamp (the word alone — its argument opens *Read this
first*) → **the TL;DR** (the answer, the confidence, the sharpest caveat) → **the Catechism**
(`references/catechism_template.md` — the worth-our-time decision,
adapted from Heilmeier) → **the report's own sections** (the argument, in the order and under the
headings its author chose — each names the reader's question it answers; no heading is required,
and red audits whether the structure carries the argument, never whether a section of a given name
exists) → risk matrix (graded likelihood × impact × complexity, including risk-accepted items
with rationale) → **the three
inquiry sections, Research areas, Future research directions and Alternatives considered** (composed by the tool from the inquiries on the record, one row per inquiry under the heading its fate names; blue records each inquiry and what became of it, and writes none of the three into the report) → **open questions, left open by this run** (authored by blue into the report's `## Open questions`, audited by red every sitting it has moved) →
footnotes (with access dates).

**The report's text ships in the order it is written.** Blue writes the TL;DR first, the Catechism
second, its own sections in the order it chooses and the open questions last; freezing the report
and every edit refuse a text that holds one of those out of its place, so the order red audits is
the order a reader gets. The tool's sections stand between blue's last section of its own and the
open questions, and move no line of blue's text.

The debate's own documents are beside it, one per audience: `docket.md` (the board in
full), `debate.md` (the transcript), `judgments.md` (motions and rulings), `inquiries.md`
(the inquiries and the path each took), `evidence.md` (the computations, and each citation with every reading of it), `run.md` (the log, record verification, cost), `CHANGELOG.md` (the claims blue
retired, the asks a later certification superseded, and every edit with blue's reason). Nothing is summarized away by the split — the union is the
directory, indexed by `README.md`.

THE CITATION AND PROOF LAYERS ARE WOVEN PER DOCUMENT. A footnote definition cannot cross a file
boundary, so each document numbers and defines the references it actually carries; proof numbers
are run-wide, so `P3` is the same computation wherever it is cited.
