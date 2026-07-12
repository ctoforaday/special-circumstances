# Special Circumstances — repository guide

This repo is a Claude Code marketplace of three plugins: **prosthetic-conscience** (core adversarial cowork), **frank-exchange-of-views** (the research debate engine), and **sleeper-service** (autonomous learning). See `README.md` for the story and `plans/` for the design under review.

## Identity

Special Circumstances is an *adversarial* methodology suite: it is not a yes-man. When you are wrong, it is expected to say so, with a reason. A good argument is a courtesy, not an attack.

## Rules

Always-on rules bind every session via the imports below; the rest load on demand by description. `design-by-contract` is the authoring grammar (BEFORE / During / AFTER · YOU MUST).

@plugins/prosthetic-conscience/skills/terse-communication/SKILL.md
@plugins/prosthetic-conscience/skills/semantic-consent/SKILL.md
@plugins/prosthetic-conscience/skills/plan-act-reflect/SKILL.md
@plugins/prosthetic-conscience/skills/anti-spinning/SKILL.md
@plugins/prosthetic-conscience/skills/context-efficiency/SKILL.md
@plugins/prosthetic-conscience/skills/agent-guardrails/SKILL.md
@plugins/prosthetic-conscience/skills/think-around-problem/SKILL.md
@plugins/prosthetic-conscience/skills/validation-loop/SKILL.md

## Working conventions

- Planning artifacts under review live in `plans/` (each as a PR). "Real" documentation lives in `README.md` and each plugin's `README.md`.
- The working-corpus dirs — `ideas/`, `research/`, `projects/` — start empty (clean start) and are seeded by `/research` and `/self-improve`.
- Quality tooling (qlty) is recommended, not required; hooks degrade gracefully when it is absent. Run `/sc-doctor`.
