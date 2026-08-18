---
name: facts-are-fields
description: Always-on structural discipline — where a record already holds a fact, it belongs in a field a writer can REFUSE, never in a filename, a heading, a name segment, or a prose substring recovered by pattern. Both shapes fail by returning a plausible zero. Scoped to bypassing an existing record; where none exists, prefer generating the derived carrier over guarding two hand-written ones.
---

# facts-are-fields

A fact nothing can refuse is a fact nothing can trust.

Two shapes, one root. A **document standing in for a record** — a markdown file holding
state that other parties act on. A **pattern standing in for a schema** — a fact encoded
into text at one end and recovered by regex, `Contains`, or `Sprintf` at the other. Both
are unmediated: nothing sits between the writer and the reader that can say *no*.

**They fail identically, and that is the point: not loudly, but by returning a plausible
zero.** A regex that matches nothing returns zero findings, which reads exactly like a
clean board. A file whose classification lives in prose delivers nothing, which reads
exactly like an empty memory. The absent case and the healthy case are the same bytes.

**SCOPE, and it is narrow.** This is about **bypassing a record you already have**: a schema
exists, and the fact goes into a markdown line or a filename beside it instead. The rule does
NOT say "any two things sharing a string need a schema between them" — read that way it
obliges inventing a record wherever a value appears twice, and the invention has a price the
rule does not pay. Read as a universal law it justifies thousands of lines of drift-guards
over CI gate lists, version constants and asset names — none of them a record — in a
repository whose real defect is that nothing is GENERATED at build time. That is
`invariant-at-wrong-level`: the enforcement real, the altitude wrong.

Where no record exists, creating one is a design decision with a cost, not an obligation.
**Prefer GENERATING the derived carrier and gating staleness** over guarding two hand-written
copies of it — a guard is what you build when generation is impossible, and it MUST say why
it was. A guard whose own allowlist is hand-kept has reproduced the defect one level up.

- BEFORE writing a fact another party will act on **through a record that already holds it**,
  YOU MUST put it in a **field on a record a writer can refuse** — validated at the write,
  where a wrong value is an error the author sees. YOU MUST NOT compose it into a name, a
  path, a heading, or prose and plan to recover it later; recovery is a hope about string
  shape, and the hope fails silently. An audit that reads a seat's claims out of a markdown
  line — fields the record already holds, joined into prose and split apart again — reports
  "nothing to reconcile" on every run once that file stops being written, and says so in the
  same words it would use for a clean board.
- BEFORE blaming the format, YOU MUST find what actually produced the number. **This clause is
  the one most often cited and least often obeyed, including by the agent citing the rule.** A
  prose key is a real defect and is still, often, not the reason the run went wrong: a memory
  source can deliver nothing because a flag meant to ADD to the list REPLACED it, while the
  frontmatter it is tempting to blame was fine. The regex is the visible smell; the
  composition is the bug, and fixing the smell leaves the run just as blind.
- BEFORE reading a fact back, YOU MUST ask what a **no-match** returns. If the miss is
  indistinguishable from the honest zero, the read is not a read — it is a coin flip
  reported as a measurement. Where the shape cannot be made refusable, YOU MUST make the
  miss LOUD (error, or a stated "not measured") rather than let it fold into the zero. A
  counter keyed on a substring reports 0 both when the guard never fires and when someone
  rewords the message.
- BEFORE removing a string-encoded fact, YOU MUST find **every other reader of that
  string** — the encoding may be load-bearing for a reason nobody wrote down. An index
  recovered from a name by regex can turn out to be the **concurrency** namespace, making a
  lock-free counter safe under parallel dispatch; "simplifying" it recreates a collision that
  makes most of the work ambiguous. This clause is the brake on the other two — see
  [[refactoring-safety]] and [[think-around-problem]].
- Prose for a HUMAN audience is not the violation, and YOU MUST NOT strip it. The
  instructive case is a report: prose is the medium, and every load-bearing point in it
  carries a tool-placed anchor. Write for the reader; put the part other machinery depends
  on in a field. Where an identifier is needed, use one that is unique by construction
  rather than assembled from parts — and never re-derive from the assembled form what the
  record could simply carry.

Both shapes are authored by the **developer**, never by the agent downstream of them, and
both survive review because a half-mediated system reads as done — see
[[complete-the-concept]] for the sweep that catches the carrier still speaking the old
model.
