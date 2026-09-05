package toolchain

// THE SUITE IS ONE BOX, AND ONE BOX HAS ONE OBLIGATION PER TOOL.
//
// Each plugin owns its manifest and declares what IT needs. Two plugins may name the same
// external tool for different reasons and at different strengths — prosthetic-conscience wants
// `jq` for diagnostics and key-scoped settings reads (optional there), frank-exchange-of-views
// needs it to filter a record projection in one call rather than three (required there). Both
// statements are true and neither is the answer to "is this box ready", which is a question
// about the box: the strictest declared need is the obligation, because satisfying it satisfies
// every other declarer and nothing weaker does.
//
// # The defect this closes
//
// sc-doctor's verdict was computed from ITS OWN plugin's manifest alone. Sibling manifests were
// read, probed and printed in their own tables, and then never consulted — so `required` in any
// plugin other than the one shipping the doctor could not produce BLOCKED. Not a jq problem:
// frank-exchange-of-views declares `node` required and the debate engine refuses dispatch
// without it, gray-area declares `git` required, and neither could move the verdict off READY.
// The table said ✗ and the verdict said fine, which is the two-readers disagreement this suite
// exists to refuse — and the more dangerous half was that a plugin author could DECLARE a hard
// requirement and be silently ignored.

// tierRank orders the tiers by strictness. An unrecognised tier ranks LOWEST rather than
// highest: a typo in a manifest must not silently escalate a box to BLOCKED, and the tier
// vocabulary is checked where manifests are authored.
func tierRank(tier string) int {
	switch tier {
	case "required":
		return 3
	case "recommended":
		return 2
	case "optional":
		return 1
	default:
		return 0
	}
}

// Stricter reports whether tier a is a stronger obligation than tier b.
func Stricter(a, b string) bool { return tierRank(a) > tierRank(b) }

// MergeStrictest folds per-plugin probe results into ONE status per tool — the whole suite's
// obligation for this box. Groups are given in precedence order, own manifest first.
//
// For a tool declared more than once the STRICTEST tier wins, and the winner's own probe result
// travels with it (the probe already ran; re-asking would be a second answer to a settled
// question). Two properties are merged across ALL declarers rather than taken from the winner,
// because taking either from one declarer loses a claim the others made:
//
//   - NotApplicable holds only if EVERY declarer says the tool is out of scope here. One plugin
//     saying "absent by design in the cloud" cannot excuse another that needs it there.
//   - TooOld and VersionUnmeasured hold if ANY declarer found them. A minimum one plugin
//     declared is a fact about the binary on this box, not about who asked.
//
// Order within the result follows first appearance, so the own-manifest tools stay in their
// authored order and siblings append — a deterministic table, not a map walk.
func MergeStrictest(groups ...[]Status) []Status {
	var out []Status
	at := map[string]int{}
	for _, group := range groups {
		for _, st := range group {
			key := probeKey(st.Tool)
			i, seen := at[key]
			if !seen {
				at[key] = len(out)
				out = append(out, st)
				continue
			}
			merged := out[i]
			if Stricter(st.Tier, merged.Tier) {
				// The stricter declarer's whole Status becomes the carrier, so the reported
				// purpose and install advice are the ones belonging to the need that binds.
				st.NotApplicable = merged.NotApplicable && st.NotApplicable
				st.TooOld = merged.TooOld || st.TooOld
				if st.VersionUnmeasured == "" {
					st.VersionUnmeasured = merged.VersionUnmeasured
				}
				out[i] = st
				continue
			}
			merged.NotApplicable = merged.NotApplicable && st.NotApplicable
			merged.TooOld = merged.TooOld || st.TooOld
			if merged.VersionUnmeasured == "" {
				merged.VersionUnmeasured = st.VersionUnmeasured
			}
			out[i] = merged
		}
	}
	return out
}

// probeKey is the identity a tool is deduplicated by: the executable actually looked up, which
// is what ProbeInDir keys presence off. Two manifests naming one binary under different `name`
// values are still one obligation on this box.
func probeKey(t Tool) string {
	if f := firstField(t.CheckCmd); f != "" {
		return f
	}
	return t.Name
}
