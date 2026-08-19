# Issue labels

Four namespaces. Every open issue carries exactly one `area:` and one `priority:`; `state:` and the
kind labels (`bug`, `enhancement`, `documentation`, `question`) are optional.

**No gate reads any of this.** Checked when the scheme was introduced: nothing in `scripts/` or
`.github/` parses a label. These are for a human or an agent deciding what to pick up, which means
a wrong label costs attention rather than a build — and also means nothing will catch one, so the
rubric below is the only thing keeping them consistent.

## `priority:` — what earns each level

| | |
|---|---|
| `high` | a shipped artifact asserts something false, a claim we rely on is unverified in a risky direction, or it is small AND unblocks other work |
| `medium` | real and worth doing, nobody is blocked. **The default** |
| `low` | speculative, long-horizon, or waiting on evidence that has not arrived |

**`low` is mostly "waiting", not "unimportant".** #431 and #469 both need a second observation
before anything can be done; #107 is explicitly gated on #62; a spike whose outcome may be
"won't do" sits here too. Do not read the bucket as a judgement about value.

**`medium` is doing the most work and defends itself the least.** When the scheme was applied,
33 of 57 landed there — enough to mean "not distinguished" as much as "equivalent". The ordering
WITHIN medium is not defensible; the boundaries against high and low are.

## `area:` — which thing it touches

Four plugins, plus two buckets that are not plugins and should not pretend to be:

| | |
|---|---|
| `area:frank-exchange-of-views` · `area:prosthetic-conscience` · `area:gray-area` · `area:sleeper-service` | the plugin whose shipped surface moves |
| `area:dev` | this repository's own tooling — `scripts/`, CI, the gate set, the dev loop, `law/` |
| `area:cross` | the seam BETWEEN plugins: an undeclared contract, a duplicated concept, a capability in the wrong plugin |

`area:cross` is not a synonym for "big". It is for issues whose whole point is that they belong to
no single plugin — gray-area parsing prosthetic-conscience's note schema is the shape. Filing one of
those under a plugin puts it in one owner's queue when the defect is that it has no owner.

`area:dev` covers work that never reaches a consumer. CLAUDE.md draws the same line in its first
paragraph, and it is the line that decides whether something is product or scaffolding.

## `state:`

| | |
|---|---|
| `state:triage` | filed, not yet assessed |
| `state:needs-verify` | **believed fixed or believed true, and not checked.** |

`state:needs-verify` is the one worth using deliberately. It carries two different things and both
are honest: an issue that looks resolved by later work but was never confirmed (#84, #85, #86 —
all three look fixed by #474, and a grep is not a verification), and a finding that is waiting on a
second instance before it can be acted on (#431, #469). Closing the first kind on a partial reading
is the failure this label exists to prevent.
