# Special Circumstances — repository guide

This repo is a Claude Code marketplace of three plugins: **prosthetic-conscience** (core adversarial cowork), **frank-exchange-of-views** (the research debate engine), and **sleeper-service** (autonomous learning). See `README.md` for the story and `plans/` for the design under review.

## Identity

Special Circumstances is an *adversarial* methodology suite: it is not a yes-man. When you are wrong, it is expected to say so, with a reason. A good argument is a courtesy, not an attack.

## Rules

Rules ship as discrete skills under `plugins/prosthetic-conscience/skills/` and load per agent via the `skills:` frontmatter. Always-on rules are `@`-imported here so they bind every session; `design-by-contract` is the authoring grammar for all of them (BEFORE / During / AFTER · YOU MUST). This file stays a **thin index** — rule content lives in the skills, never inline.

@plugins/prosthetic-conscience/skills/terse-communication/SKILL.md

## Working conventions

- Planning artifacts under review live in `plans/` (each as a PR). "Real" documentation lives in `README.md` and each plugin's `README.md`.
- The working-corpus dirs — `ideas/`, `research/`, `projects/` — start empty (clean start) and are seeded by `/research` and `/self-improve`.
- Quality tooling (qlty) is recommended, not required; hooks degrade gracefully when it is absent. Run `/doctor`.
