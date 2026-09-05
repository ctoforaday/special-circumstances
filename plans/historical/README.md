# Historical plans

Plans that are **done being planned from**, kept as the record of what was
decided and why. Two ways to qualify, and the second was found by the first
sweep rather than anticipated:

- **Shipped** — the design was built, and nothing live still cites the plan.
- **Superseded** — the design was answered a different way and never built, so
  nothing live can be describing it. `seat-identity.md` is the standing example:
  the opaque-id invariant it proposed lost to the roster gate that shipped
  instead (`record/roster.go`), and it carries its own retirement audit.

- **The archaeology half of a split** — the plan's live design stayed in
  `plans/`, and what changed is here beside it. These come in pairs: each half
  opens with a pointer to the other, and the historical half **keeps the original
  section numbers** so a citation of `§17` still names §17. Only the path moves.

An earlier version of this file said "shipped" alone. That was false about its
own second inhabitant within a day, and silent about the third shape entirely —
which the first sweep produced six of.

**They describe the tree as it was, not as it is.** A historical plan naming a
heading, a verb or a file that has since been renamed is not stale — it is
accurate about its own moment, and editing it to match today would destroy the
only record of the change. That is what moving it here says, and it is why the
sweep that renames something does not rewrite these.

A plan belongs here when its design is delivered AND nothing live still cites it
as the design of record. The second half is the one that catches people:
`claude-port-plan.md` is marked "shipped — historical record" in `../README.md`
and stays in `plans/` regardless. Its census, re-run during the first sweep and
larger than the estimate it replaced: **three** plugin READMEs (not four —
gray-area's does not cite it), the repository `README.md`, `MEMORY.md`, a Go
source comment, `requirements.json`, `.qlty/qlty.toml`, and three sibling plans.
Its Phase 4 (sleeper-service) is also still unbuilt.

**Citations are not only by filename.** Live code and sibling plans cite these
documents by SECTION (`§3a′`), by wave id (`W1.13`), by item number ("item 8")
and by risk id (`T2`). A move must keep every one of those findable at the path
that names it, so the census below is a floor, not the whole question.

Before moving one, run the census:

```
$ grep -rn "<plan-filename>" . 
```

Zero hits outside the file itself is the bar. Anything else is a link to fix or
a reason not to move it.
