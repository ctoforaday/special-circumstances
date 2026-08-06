---
name: facts-are-fields
description: Always-on structural discipline — a fact other parties depend on belongs in a field a writer can REFUSE, never in a filename, a heading, a name segment, or a prose substring recovered by pattern. Both shapes fail by returning a plausible zero.
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

- BEFORE writing a fact another party will act on, YOU MUST put it in a **field on a
  record a writer can refuse** — validated at the write, where a wrong value is an error
  the author sees. YOU MUST NOT compose it into a name, a path, a heading, or prose and
  plan to recover it later; recovery is a hope about string shape, and the hope fails
  silently. Measured: red's accumulated gap memory is 60 markdown files whose class is
  parsed by five regexes over frontmatter — **1 of 60 carries the field**, so the auditor
  begins each run substantially memoryless while its memory directory looks full.
- BEFORE reading a fact back, YOU MUST ask what a **no-match** returns. If the miss is
  indistinguishable from the honest zero, the read is not a read — it is a coin flip
  reported as a measurement. Where the shape cannot be made refusable, YOU MUST make the
  miss LOUD (error, or a stated "not measured") rather than let it fold into the zero.
  Measured: an estoppel counter keyed on the substring `"estoppel —"` reports 0 both when
  the guard never fires and when someone rewords the message.
- BEFORE removing a string-encoded fact, YOU MUST find **every other reader of that
  string** — the encoding may be load-bearing for a reason nobody wrote down. Measured:
  a lens index recovered from a seat name by regex turned out to be the **concurrency**
  namespace, making a lock-free counter safe under parallel dispatch; "simplifying" it
  would have recreated a collision that made 39 of 60 disposals ambiguous. This clause is
  the brake on the other two — see [[refactoring-safety]] and [[think-around-problem]].
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
