package checkpointseal

import "embed"

// THE SEAL RECORD'S READERS LIVE HERE, and until this package existed they lived in a plan.
//
// `.claude/checkpoints/seals.jsonl` is written by three binaries and read by nothing that
// compiles. Its readers were the `jq` programs in §V of `plans/checkpoint-freshness.md` — a
// design document — and a test reached five directories up to parse that markdown so it could
// check the field names still resolved. That made a PLAN load-bearing for a build: editing prose
// could fail CI, and moving the file would fail it for a reason no reader would guess.
//
// The queries are the artifact, so they are files. The plan keeps the prose that argues what
// each one measures and points here for the program itself; there is one copy, and it is this
// one.
//
// WHY THEY STILL MATTER, which is the part worth not losing: criteria 4 and 6 of the freshness
// work are DECIDED by these programs. Rename a JSON tag on sealRow and `select(.handles_measured)`
// matches nothing, `map(.emission_bytes_max) | max` is null, and the gate reports zero emissions —
// which is the PASSING answer. The broken query and the clean board produce the same output, which
// is [[facts-are-fields]] clause 3: ask what a no-match returns.
//
// They are not yet RUN by anything in CI — a human runs them against a collected corpus. That is
// the remaining half, and naming it is the point of putting them somewhere a runner could reach.
//
//go:embed queries/*.jq
var queryFS embed.FS

// Queries returns each embedded jq program by file name.
func Queries() map[string]string {
	entries, err := queryFS.ReadDir("queries")
	if err != nil {
		panic("checkpointseal: queries/ is embedded at compile time and cannot fail to read: " + err.Error())
	}
	out := make(map[string]string, len(entries))
	for _, e := range entries {
		b, err := queryFS.ReadFile("queries/" + e.Name())
		if err != nil {
			panic("checkpointseal: " + err.Error())
		}
		out[e.Name()] = string(b)
	}
	return out
}
