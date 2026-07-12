# debate.md format — the literal three-party transcript

Appended every round; never rewritten. `blue/CHANGELOG.md` carries the mechanical edit log so this file stays argument-focused.

```markdown
## Round <N>

### RED (audit)
VERDICT: FAIL — <k> gaps (<new>/<carried>).
| id | location | problem | required fix | severity | likelihood | impact | complexity |
|----|----------|---------|--------------|----------|------------|--------|------------|
Corroboration flags: <statement↔reference pairs with low/medium confidence>.

### BLUE (response)
Gap R<N>-1: accepted — §<x> expanded, <what was added>.
Gap R<N>-2: REBUTTAL — <evidence-backed counter>.

### LEAD (resolution — only on deadlock check or final)
R<N>-1 closed. R<N>-2 rebuttal_sustained (<rationale>). R<N>-3 risk_accepted (<tradeoff rationale>). R<N>-4 carried.
```
