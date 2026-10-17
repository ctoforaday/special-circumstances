package cli

import (
	"sort"
)

// A REFUSAL NAMES ITS COMMAND ONCE, AND THE RENDERER IS WHAT NAMES IT.
//
// seat.RenderRefusal prints `feov-record: <path>: <diagnosis>` and carries the role and the verb as
// fields of the JSON envelope, both read from the command that resolved. A diagnosis that opens
// with its own role or its own path therefore says it twice — `feov-record: finding: lens finding
// requires --reason` — and a seat reading for what was wrong reads past two names to reach it.
//
// Every refusal any test drives through `run` is read against this (cli_test.go), the binary's
// own output is in TestEveryRefusalKindHasOneEnvelope and in every difftest scenario, and the
// messages no test reaches are read at the source (refusalshape_test.go) — so a refusal written
// with a hand-typed head fails in whichever of them first meets it. The reading itself is
// internal/refusalcheck's, shared by all of them.

// roleWords are the role names the trees are built for. It runs on every refused invocation, so
// it reads the names AllRoots builds from rather than building the trees: five trees per refusal
// was 46% of this package's live heap.
func roleWords() []string {
	out := SurfaceRoles()
	sort.Strings(out)
	return out
}
