---
name: complete-the-concept
description: Always-on completion discipline — a change is ONE conceptual action, finished when the concept is, not when the first commit merges. Follow the thread to every carrier and consumer.
---

# complete-the-concept

One conceptual change, followed to its end. A merge is not a finish line.

- BEFORE declaring a change complete, YOU MUST follow the concept to **every carrier and consumer** and confirm each speaks the new model: the code AND its call sites; tests, fuzzers, and goldens; the agent-facing surfaces — prompts, agent definitions, constitutions, help text; docs and diagrams; the version surface. A change that lands its primary edit while a carrier still speaks the OLD model is a **half-state that reads as done** — and it reads as done to you, which is why it survives. A system ships green with its prompts rewired and its agent constitutions still instructing the old behaviour, and a fuzzer that never exercises either new command — every gate passing, the concept half-landed.
- A conceptual change MAY span multiple pull requests, but it is ONE action: **a PR boundary is not a completion boundary.** YOU MUST NOT let "the PR merged" stand in for "the concept is finished" — carry the remaining thread explicitly (an issue, the plan's §III, [[context-checkpointing]]) so the unfinished half is *tracked*, never merely remembered.
- BEFORE cutting scope to limit blast radius, YOU MUST confirm the smaller scope is a **complete concept, not a truncated one**, and enumerate what the full concept would additionally touch. Scoping down is legitimate; scoping down *silently* is how the class survives. The instinct to keep a change small is the one this rule most often has to overrule — see [[refactoring-safety]] (fix the class, not the instance) and [[think-around-problem]].
- During planning, the concept's full carrier/consumer set MUST be enumerated in the spec — [[spec-driven-development]]'s **Completeness** dimension fails a plan that does not, and the auditor re-runs each census itself rather than trusting the list.
