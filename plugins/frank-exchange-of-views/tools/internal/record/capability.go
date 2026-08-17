package record

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// WHAT A STALE BINARY CANNOT DO, as a field rather than as prose (#407).
//
// `setup` refuses a run when the binary's version disagrees with the plugin manifest, and until
// now the refusal could say only "yours is X, the plugin expects Y — a skewed binary writes
// events under a different contract". True, and useless: it does not tell the operator whether
// to care.
//
// The answer already existed. 48 of root.go's 70 changelog entries ended with exactly the
// sentence a refusing preflight wants — "A stale binary has no `declare`, takes no `--binds`,
// and renders neither" — and it sat in a Go comment where no program could reach it. The one
// audience that needed it, an operator whose run had just been refused, was the one audience
// that could not see it. That is the shape this suite keeps finding: the fact composed into
// prose, recoverable by nobody.
//
// The rest of that changelog is gone, and `git log` is the canonical half. Verified rather than
// assumed before deleting: each entry's headline resolves to a real commit, and the commit
// BODIES carry the same reasoning in more detail — 0.60.0's carries a reproduction transcript
// the changelog entry does not. Nothing was lost that git does not hold.
//
// NEW ENTRIES ARE REQUIRED, NOT OPTIONAL: TestEveryReleasedVersionDeclaresItsDelta fails when
// `Version` moves without one. A version with no capability change says so in words ("no
// capability change") rather than by absence — an absent entry and an unremarkable release
// must not be the same bytes.
var capabilityDeltas = map[string]string{
	"0.10.0": "A stale binary silently produces the duplicated, contradicted report the preflight must now refuse.",
	"0.12.0": "A stale binary produces the graph-less, prose-dropping report the record now expects.",
	"0.14.0": "A stale binary renders a debate view missing the dispute argument the bench now rules on from the record.",
	"0.13.0": "A stale binary drops blue's calibration the record now carries.",
	"0.16.0": "A stale binary lacks the citations tally the merge prompt now reads.",
	"0.42.0": "A stale binary records unaddressed answers.",
	"0.41.0": "A stale binary records the pursuit and loses the argument.",
	"0.40.0": "A stale binary records unevidenced retirements.",
	"0.39.0": "A stale binary records an unchecked verdict.",
	"0.37.0": "A stale binary writes friction without the field, so its rejections read 0 — the honest answer for events that never recorded the fact.",
	"0.36.0": "A stale binary ignores FEOV_RUN and silently obeys the typo.",
	"0.35.0": "A stale binary answers \"unknown view\" to the read the bench's constitution now tells it to make.",
	"0.34.0": "A stale binary lacks both verbs.",
	"0.33.0": "A stale binary lacks rule-avenue and the move form.",
	"0.32.0": "A stale binary neither records the fact nor enforces the guard.",
	"0.31.0": "A stale binary rejects --fix-old, failing the mint red's prompt now instructs.",
	"0.30.0": "A stale binary rejects `--view changes`, failing the read both prompts now instruct.",
	"0.29.0": "A stale binary rejects --answers outright, which fails every edit the seat prompt now instructs — the hard break this version gate exists for.",
	"0.28.0": "A stale binary lacks `fetch`/`blue cite`, so citations are unmanaged and the weave/detector silently do not engage.",
	"0.27.0": "A stale binary lacks `blue edit` and the hook backends, so the lockdown silently does not engage.",
	"0.26.0": "A stale binary would record findings without anchoring them and miss the tampering detector.",
	"0.43.0": "A stale binary still ACCEPTS both verbs and writes events this replay drops, which is why the contract version moves: setup must refuse it.",
	"0.44.0": "A stale binary still accepts `lens cite` and writes the ambiguous event, so setup must refuse it.",
	"0.45.0": "A stale binary accepts `rebuttal_accepted` and `risk_argued`, words no reader now understands.",
	"0.46.0": "A stale binary treats reproduce as a read and records nothing.",
	"0.47.0": "A stale binary writes events with no `role` field; PartyOf falls back for those, so old records still read.",
	"0.50.0": "A stale binary ignores FEOV_RECORD_ROOT and writes into <runDir>/records.",
	"0.51.0": "A stale binary calls itself feov-record whatever it is named, and answers a mistyped command by complaining about its flags.",
	"0.52.0": "A stale binary hides check_kind, has no motions view, and rejects `--none`.",
	"0.53.0": "A stale binary accepts an outcome with no account and shows no debt.",
	"0.54.0": "A stale binary still offers the verb and writes an event nothing renders.",
	"0.55.0": "A stale binary answers a bare `show` with whatever its role's old default was.",
	"0.56.0": "A stale binary takes --view and does not answer `show <name>`.",
	"0.57.0": "A stale binary cannot answer `show report` and has no operator friction read.",
	"0.58.0": "A stale binary answers `show ledger` and `show archive` and takes no --format.",
	"0.59.0": "A stale binary answers `show citation-ledger` and has no `show evidence`.",
	"0.60.0": "A stale binary takes `--trust`, has no `--confidence`, and accepts a verification of nothing.",
	"0.61.0": "A stale binary tells a seat to run `show --view board`.",
	"0.62.0": "A stale binary answers `show --view changelog` and `--view citation-ledger`.",
	"0.63.0": "A stale binary accepts a mint at a location no reader can find.",
	"0.64.0": "A stale binary takes `--id NOPE` and answers a parse error in prose under --json.",
	"0.65.0": "A stale binary offers `--existence` and renders a leaf-check note no seat can set.",
	"0.66.0": "A stale binary silently builds a blackboard wherever a relative path lands.",
	"0.67.0": "A stale binary has no `declare`, takes no `--binds`, and renders neither.",
	"0.68.0": "A stale binary lays down a CHANGELOG husk nothing fills.",
	"0.69.0": "A stale binary harvests from envelopes and reports a plausible count.",
	"0.70.0": "A stale binary reports a lossy replay as a clean run.",
	"0.71.0": "A stale binary stamps the round its seat id looks like.",
	// ONE VERSION, TWO CAPABILITY CHANGES, AND THAT IS THE HONEST ENTRY. #441 and #445 each
	// bumped Version to 0.72.0 on its own branch, each added its own 0.72.0 line here, and both
	// merged — so git produced a map with a duplicate key and main stopped compiling. Both
	// features really did ship under 0.72.0, so a binary below it lacks BOTH; splitting them
	// across 0.72.0 and 0.73.0 would invent a release that never existed. Merged rather than
	// renumbered for that reason.
	"0.72.0": "A stale binary has no `merge inquiry-support`, so red cannot vote that the report still carries a line of inquiry — and it does NOT refuse `verdict --as PASS` over an unvoted line, so a run passes with its own account of what it investigated unchecked. It also overwrites an open run's marker without a word, and its dashboard watchers outlive a run nothing captured.",
}

// CapabilityDelta is what a binary at this version cannot do, or "" if the version is unknown.
func CapabilityDelta(version string) string { return capabilityDeltas[version] }

// DeltasBetween returns every capability delta a binary at `running` is missing relative to
// `required`, oldest first — the answer to the only question a refused preflight raises.
//
// Returns nothing when running >= required: a NEWER binary than the manifest expects is a
// different problem (it writes events under a contract the plugin has not seen) and listing
// deltas it already has would be noise.
func DeltasBetween(running, required string) []string {
	if running == "" || required == "" || compareVersions(running, required) >= 0 {
		return nil
	}
	var vs []string
	for v := range capabilityDeltas {
		if compareVersions(v, running) > 0 && compareVersions(v, required) <= 0 {
			vs = append(vs, v)
		}
	}
	sort.Slice(vs, func(i, j int) bool { return compareVersions(vs[i], vs[j]) < 0 })
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = fmt.Sprintf("%s: %s", v, capabilityDeltas[v])
	}
	return out
}

// compareVersions orders dotted numeric versions. NOT a general semver implementation — this
// tree's versions are plain MAJOR.MINOR.PATCH with no pre-release or build metadata, and a
// parser that accepted more than the tree produces would be untested surface.
func compareVersions(a, b string) int {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var x, y int
		if i < len(pa) {
			x, _ = strconv.Atoi(pa[i])
		}
		if i < len(pb) {
			y, _ = strconv.Atoi(pb[i])
		}
		if x != y {
			if x < y {
				return -1
			}
			return 1
		}
	}
	return 0
}
