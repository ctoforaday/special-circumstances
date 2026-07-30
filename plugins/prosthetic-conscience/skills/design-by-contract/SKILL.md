---
name: design-by-contract
description: Use when writing rules, defining agent roles, or authoring multi-step operational instructions — anchor every instruction to verifiable execution states.
---

# design-by-contract

All complex instructions MUST be state-based logical contracts.

- Format every rule with the contract keywords: `BEFORE [action], YOU MUST [behaviour]`; `During [state], YOU MUST NOT [behaviour]`; `AFTER [event], YOU MUST [behaviour]`.
- YOU MUST NOT use vague "Always / Never" imperatives for multi-step logic.
- **Semantic Precision**: YOU MUST NOT sacrifice intent or critical context for brevity; every contract MUST be unambiguous.
- **Literal Language**: YOU MUST write contracts in operational language a reader can execute; YOU MUST NOT use metaphor, koans, or figurative framing that could be taken literally.
- **A script must do what a script can do**: BEFORE writing a multi-step instruction for an agent to execute, YOU MUST check whether the step has a computable correct answer from observable state — if it does, it belongs in a tested script/binary the instruction *invokes*, never in prose the agent *interprets*. Prose is for decisions (weighing tradeoffs); prose-executed mechanics are an unenforced good-faith contract — the documented failure class behind staged-input loss, skipped setup steps, and silent capture omissions. For promises about FUTURE behavior (freezes, "never do X while Y"), YOU MUST record the commitment as observable state (a marker file) and let a deterministic guard consult it — memory is not an enforcement mechanism.

- BEFORE shipping a change to a RULE — a skill, a seat prompt, a constitution, a
  template, or law — YOU MUST name the defect's CLASS and sweep its siblings.
  Fixing the instance in front of you leaves the class alive at the adjacent
  seat, where it passes the very tests your patch added: measured four times out
  of four (lens labels patched, blue-lane footnote namespaces bit the next run;
  blue-reads-transcript patched, the lossy gap summary bit the next run; grade
  enums widened, the mass mapping distorted the next run; friction-to-file
  shipped, lens seats still had no write path). Classes and the minting
  discipline live in `feov-memory/protocol-class-registry.md`; `Rule-Class:` and
  `Sibling-Sweep:` commit trailers record it and `scripts/rulesweep`
  enforces it, because a sweep requirement nothing checks is itself the
  policy-without-mechanism class it would be enforcing. [[refactoring-safety]]
  already states this duty for CODE faults ("prevent the class, don't patch the
  instance"); the rulebook broke it four times running because nothing scoped it
  to rules and nothing checked it — which is why this clause names the surface
  and points at a mechanism rather than restating the principle.
