---
name: pattern-doctrine-vs-implementation
description: A code comment/doc states a design principle ("never cheapen the adversary") that the literal routing/logic below it contradicts — check every declared invariant against its own implementation, not just external behavior
metadata:
  type: feedback
---

Gap pattern: **doctrine-vs-implementation contradiction**. A source file states its own design
principle in a comment or doc string, then the concrete logic immediately below doesn't follow it
— caught by literally reading the routing table against the doctrine sentence, not by any
external test.

**Why:** Caught in the FEOV-retrospective audit (round 1, lens 5). `debate.js`'s own comment:
"cheapen redundancy and mechanics, never judgment or the adversary" — followed immediately by a
routing table that assigns the cheap/bulk model tier to red-lens passes (the actual adversarial
audit work: leaf-node citation checks, gap-finding) and reserves the judgment tier only for
red-merge (mechanical consolidation). Whichever reading is "intended," the document's own stated
invariant and its own next ten lines disagree — and the report under audit, which spent a whole
item grading this exact routing choice, never noticed the tension.

**How to apply:**
- When a file states a doctrine/principle in prose (comments, docstrings, design-doc sentences),
  treat it as a testable claim about the code immediately following it — read the two side by
  side, don't take the prose at face value.
- Two valid resolutions exist when a mismatch is found: (a) the code is wrong relative to its
  stated doctrine — fix the routing/logic; (b) the doctrine's terms are ambiguous and the code is
  a defensible reading — tighten the prose to remove the ambiguity. Present both; don't assume
  which was "intended."
- This is cheap to catch (grep the doctrine sentence, then read the next N lines) and easy to miss
  under time pressure because both halves *individually* read as reasonable — the contradiction
  only shows up on direct juxtaposition.

**Extension — doctrine contradicted by ARTIFACT PLACEMENT, not adjacent code (sleeper-service
R4, L5-F1):** the prose said "the command markdown is the wrapper's phase-1 payload, not a
standalone entry point" — while the file ships under `commands/`, the harness's entry-point
directory, making it user- AND model-invocable regardless of the prose. Sharper still when the
design applies the known mechanical guard (`disable-model-invocation: true`) to a SIBLING
artifact but not this one: the asymmetry proves the mechanism was known and the omission is a
real hole, not ignorance. Check where a "not an entry point" file physically lives and what
the harness does with that location; and diff sibling artifacts for guard flags one carries
and the other lacks.

**Extension — the report's own HONESTY/REMEDY paragraph contradicted by a sibling program in
the same section (quadratic-formula R4, L2-F3 -> R4-1):** a paragraph titled "one defect this
report cannot fix, disclosed so the footnote ships explained not unexplained" narrated ONE
stranded title anchor ("A citation anchor placed in round 2..."), singular throughout — while
two paragraphs above, the section's own shipped doc-check program was written against "the two
known stranded ones", and the title line actually rendered two cite anchors. So the prose
under-counts and mis-attributes the very defect it exists to disclose, and a sibling program in
the same section already carries the true count. QUESTION to carry: when a report has a
self-disclosure / self-audit / "here is our own defect" paragraph, does its count and
attribution match the section's own program output and the actual rendered artifact (title
line, anchor set, token count)? The honesty-showcase site is where an inaccuracy is least
expected and most damaging. Also: a bench "defect_owed_elsewhere" ruling on the ANCHOR debt
(tooling) does NOT estop a successor about whether the PROSE disclosing that debt is accurate —
different site, different question, name the ruled gap in supersedes.
