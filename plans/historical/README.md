# Historical plans

Plans whose design **shipped**, kept as the record of what was decided and why.

**They describe the tree as it was, not as it is.** A historical plan naming a
heading, a verb or a file that has since been renamed is not stale — it is
accurate about its own moment, and editing it to match today would destroy the
only record of the change. That is what moving it here says, and it is why the
sweep that renames something does not rewrite these.

A plan belongs here when its design is delivered AND nothing live still cites it
as the design of record. The second half is the one that catches people:
`claude-port-plan.md` is marked "shipped — historical record" in `../README.md`
and stays in `plans/` regardless, because four plugin READMEs, the repository
README, `MEMORY.md` and a Go source comment all cite it by path — and its
Phase 4 (sleeper-service) is still unbuilt.

Before moving one, run the census:

```
$ grep -rn "<plan-filename>" . 
```

Zero hits outside the file itself is the bar. Anything else is a link to fix or
a reason not to move it.
