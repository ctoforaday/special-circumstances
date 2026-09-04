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
@plugins/prosthetic-conscience/skills/complete-the-concept/SKILL.md
@plugins/prosthetic-conscience/skills/facts-are-fields/SKILL.md

## Repository structure

| Path | Role |
|---|---|
| `plugins/<name>/` | The product: everything a consumer installs (skills, agents, commands, hooks, tools) |
| `.claude-plugin/marketplace.json` | Marketplace manifest listing the four plugins — keep in step with `PLUGINS` in `scripts/bootstrap-plugins.sh` |
| `plans/` | Design artifacts under review — each arrives as a PR; graduates into the plugins |
| `README.md`, `plugins/*/README.md` | The shipped documentation |
| `ideas/` `research/` `projects/` | Working corpus — starts empty, seeded by `/research` (and by `/self-improve` once sleeper-service ships) |
| `run-archive/` | The raw record of each captured run, gzipped. `research/` is gitignored, so a run directory survives a resume but NOT the container; this is the only part of a run that outlives it, and every audit re-reads it. Records and proofs only — the fetched-source cache is re-fetchable and 100× larger. |

## Developing this repo

- **Versions move at a RELEASE BOUNDARY, not per PR.** An ordinary PR changes plugin content and leaves `version` alone. A release is a human call — made when the binary/text contract has actually moved — and it is ONE act: bump the plugin's `version` in `plugin.json` and tag that commit `<plugin>--v<version>`. The release job refuses a tag whose manifest disagrees (`versionguard -tag`), so the two cannot drift.
  - A bump per PR makes the version a commit counter that tells a consumer nothing while the tags never move, so no plugin ends up with a tag matching its own version and `sc-doctor -fix` pins every download to a release that does not exist.
  - The cost, stated plainly: between releases `/plugin update` ships nothing, because it is version-gated. That is what a release model means. AFTER a release, run `/plugin update` + `/reload-plugins` to pull it.
  - `versionguard` fails a version that goes BACKWARDS, and sweeps for stale per-binary version constants.
- BEFORE trusting a coverage number, YOU MUST remember what it cannot see: `internal/secrets` reported **100.0% of statements** while two of its eight secret patterns could be deleted outright with the suite green. `(cd scripts && go run ./mutate)` asks the question coverage cannot — would the tests NOTICE? It is an on-demand audit, not a CI gate (minutes to run, and the residue is judgement: equivalent and platform-conditional mutants cannot be killed by any test). Survivors are a list to explain, not a number to drive to zero.
