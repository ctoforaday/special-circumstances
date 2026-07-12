# Special Circumstances — repository guide

**Scope: this file governs work *on* this repo only.** It is not part of any plugin: behaviour delivered to consuming projects lives entirely under `plugins/`, and nothing in this file reaches an installing project. Dev concerns belong here; consumer behaviour belongs in the plugins.

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

## Repository structure

| Path | Role |
|---|---|
| `plugins/<name>/` | The product: everything a consumer installs (skills, agents, commands, hooks, tools) |
| `.claude-plugin/marketplace.json` | Marketplace manifest listing the three plugins |
| `plans/` | Design artifacts under review — each arrives as a PR; graduates into the plugins |
| `README.md`, `plugins/*/README.md` | The shipped documentation |
| `ideas/` `research/` `projects/` | Working corpus — starts empty, seeded by `/research` and `/self-improve` |

## Developing this repo

- Every PR that changes a plugin's content MUST bump that plugin's `version` in its `plugin.json` — `/plugin update` is version-gated and ships nothing without it. AFTER merging, run `/plugin update` + `/reload-plugins` to pull the change.
