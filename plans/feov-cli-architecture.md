# feov-record CLI — decollapsing `seat` (the #59 architecture)

Agreed 2026-07-19 in design dialogue. This is an experiment, private, and the bar is
architectural pride, not production-shippability. #58 lands global `--json` with typed
per-verb results; this document is the NEXT cut (#59), which reshapes what `seat` is.

## The one sentence

`seat` should stop owning things that already live somewhere structural:

- **behavior belongs in the command** — a verb is a plain `*cobra.Command` that owns its
  `RunE`, not a thin `Handler` a central factory wraps;
- **meaning belongs in the error** — an error carries its own code, the edge only transports;
- **role belongs in the tree** — the role is the verb's parent node, not a `string`
  threaded through every constructor.

What survives is small and honest: verbs own their RunE; a thin `Emit`/`Of` reads position
and format; the role commands are tree nodes wiring their verb sets; a `feov.Error` type
carries the code. No generic "seat" parameterized by role, no god-factory.

> **AMENDED 2026-08-09 by `plans/command-groups.md`.** Cut 1's argument — that a `role
> string` threaded through every constructor is a command being told its own position —
> stands and is kept. What changes is WHERE the position is read: the tree becomes ENTITY
> groups (`board mint`, `evidence cite`), and role is read once from the inherited
> `--seat-id`, a fact about the run rather than about the CLI's shape. No constructor takes
> a role either way.
>
> The genuine reversal, stated so it is not discovered later: after that change an
> out-of-role verb is **no longer structurally impossible**. `lens dispute` cannot be typed
> today; it becomes a permission lookup. See §IV of the successor plan — the mitigation
> (one table that both gates the write and generates per-seat help) is load-bearing.

## Cut 1 — role is structure, not data

`feov-record merge mint`: `merge` is a node; `mint`'s parent IS the role. Every place that
threads a `role string` is passing a command its own position.

- `Of(cmd, role)` → **`Of(cmd)`**: RunDir/SeatID from inherited persistent flags, `Role` from
  `cmd.Parent().Name()`.
- `New(role, ...)` → drop the `role` param. A verb can't know its parent at construction
  (it is mounted later) — the tell that role is a runtime, positional fact.
- `CheckSeatRole`, the error prefix, the shared verbs (`Register(role,…)` → `Register(…)`)
  all read `cmd.Parent().Name()` when they need the role.

This is what dissolves the "generic seat parameterized by role" idea: there is no generic
seat, only four concrete role commands and role-agnostic helpers that read position.

## Cut 2 — decollapse `seat.New`; settle render-on-mutation FIRST

`seat.New` takes over `RunE` to apply preconditions + comment + render-on-mutation + output.
That single move is why the whole tool reads as un-idiomatic Cobra, and it HIDES behavior.

**Open question to decide before the mechanical work:** `New` fires a `PostRunE` that
re-renders ALL projections from the full event log after every mutation — O(events) per
call, invisible at the call site, and a measured chunk of the 3.2s/call the timing analysis
flagged. Should a run really re-render on every write, or lazily on read / once per
seat-batch? Deciding this may delete the heaviest reason the factory exists.

Then:
- Preconditions → `PersistentPreRunE` on the ROLE command (Cobra-native; runs for every
  verb because verbs carry only `RunE`). Sidesteps the Pre/PostRunE non-chaining gotcha by
  keeping verbs free of their own Pre/PostRunE.
- Output → a `seat.Emit(cmd, result)` each verb calls EXPLICITLY (gh's shape), with a test
  asserting every verb calls it — the `--json`-can't-be-forgotten invariant kept without the
  RunE takeover.
- `register`'s render exemption becomes explicit rather than an `if name != "register"` in
  the factory.
- Each verb becomes a plain `*cobra.Command` with its own `RunE`.

## Cut 3 — `feov.Error`: one type, an enum, a mint helper

Not a family of bespoke types (that was over-reach) — polymorphism by ENUM VALUE within one
type. Structure in the error, FORMAT at the edge (never a `Render(w)` on the error — that
recouples errors to output).

```go
package feov // a LEAF package: record (raises) and seat (reads) import it, no cycle

type Code string
const ( NotFound Code = "not_found"; MissingField Code = "missing_field"
        RoleViolation Code = "role_violation"; Validation Code = "validation" )

type Error struct { Code Code; Msg string; err error }
func (e *Error) Error() string { return e.Msg }
func (e *Error) Unwrap() error { return e.err }
func Errorf(code Code, format string, a ...any) *Error { return &Error{Code: code, Msg: fmt.Sprintf(format, a...)} }
```

- Raise sites keep their rich human guidance, just tagged:
  `return feov.Errorf(feov.NotFound, "no finding … `show --view board` …", id)`.
- Edge does ONE lookup, no type switch: `var fe *feov.Error; if errors.As(err, &fe) { code = string(fe.Code) }`.
- Not everything is minted — plain `fmt.Errorf` falls through to the default `"error"`; only
  the domain faults a seat would branch on get a code. The enum IS the taxonomy.
- The error envelope carries `{verb, role, ok:false, code, error}` — role/verb structured,
  not string-mashed into the message.

## Build order (each step compiles; goldens stay byte-identical)

1. `Of(cmd)` + role-from-tree; delete the `role` params.
2. Settle render-on-mutation; move preconditions to role `PersistentPreRunE`; verbs own
   `RunE`; `Emit` becomes explicit.
3. `feov.Error` + enum + mint; convert the domain errors; edge reads `Code` structurally.

Off a clean `main` (after #57 + #58 merge), as #59 — "this is how feov-record is actually shaped."

## OUTCOME (#59, 2026-07-19) — and where I changed the plan

Cut 1 and cut 3 landed as written. **Cut 2 changed shape while building it**, and the change is
worth recording because it contradicts the plan above:

- **Cut 1 — role from the tree.** Done. `Of(cmd)`; `roleOf` = `cmd.Parent().Name()`; the `role`
  param is gone from `New`, `Begin` and every shared verb. `Role(role, …)` keeps it — it is the
  node's `Use`. Behavior-preserving (goldens unchanged).
- **Cut 2 — render-on-mutation.** Dropped, per operator call. It was O(events) per write and
  hidden; `show` renders on read, the prompt renders explicitly, verdict and capture render at
  the end. The difftest now renders once at end-of-run (mirroring capture).
- **Cut 2, the OTHER half — verbs own RunE — was NOT done, on purpose.** Working it through, the
  `--json` result contract is a genuine thing Cobra doesn't provide (its `RunE` returns only
  `error`), and the cohesive spot for it is "handler returns `(Result, error)`, one wrapper
  renders both success and error." Fully decollapsing into per-verb `RunE`s either recreates a
  `New`-shaped wrapper or copies the cobra scaffolding into all ~28 verbs (replication). So the
  prouder move was to make `New` **hook-free** — preconditions in `Begin(cmd)` called at the top
  of the one `RunE`, no `PreRunE`/`PostRunE` — rather than to dissolve it. `New` is now a thin,
  linear, no-chaining command builder, not the god-factory we were rightly suspicious of.
- **Cut 3 — `feov.Error`.** Done. One `Error` type + `Code` enum, minted inline; `feov.CodeOf`
  (errors.As over `%w`) at the edge, no type switch. `Emit` is the single success/error renderer
  and the single code-reader. Envelope carries `role` (both outcomes) and `code` (errors).
  Two domain faults coded as examples (role_violation, not_found); the rest stay plain.

Versions: cli.Version + recordToolVersion → 0.5.0; plugin → 0.28.0.
