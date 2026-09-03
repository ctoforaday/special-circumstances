// Package surface holds the STATIC agreement gates: every test here reads shipped text —
// docs, prompts, constitutions, source literals, debate.js enum bindings — and checks it
// against the tool surface it describes, without building or driving the binary.
//
// They lived in integration/fuzz, whose sweep drives the real binary for ~20 minutes — so
// the common case, a docs or prompt change, either paid the full fuzz price to verify a
// text agreement or (measured, 2026-09-02) the package was skipped for being slow and the
// gate that would have caught a README defect was skipped with it. A gate nobody runs
// locally fires only in CI, at its most expensive moment.
//
// The split is by DEPENDENCY, not by speed: everything here needs only the repository tree
// (via internal/repotree, which refuses to hand back an empty set) and the importable
// internal packages. A gate that needs the built binary or a driven run stays in
// integration/fuzz regardless of how fast it is — runhandlerace stays there for a second
// reason: check's narrowRace gate names that package path.
package surface
