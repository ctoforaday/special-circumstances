package reportproj

import (
	"strings"
	"testing"
)

// THE SPLICE NEVER REMOVES A TOKEN (P12b, "The splice and the list"). A token inside the located span
// that the replacement does not carry is kept, in order, at the end of the replacement; the edit's
// removed_anchors list is what takes a token out, by a remove after the splice. So a token the seat's
// quote includes and one abutting the span outside it leave the same way, and removeMut's loud failure
// on an absent token stays correct.
func TestTheSpliceKeepsEveryTokenTheReplacementDoesNotCarry(t *testing.T) {
	const f1, f2 = "<!--gap:G-00000001-->", "<!--gap:G-00000002-->"
	report := "One. Two" + f1 + f2 + ". Three.\n"
	start := strings.Index(report, "Two")
	end := start + len("Two"+f1+f2)
	for _, tc := range []struct{ name, new, want string }{
		{"a cut keeps both, in order", "", f1 + f2},
		{"a replacement carrying one keeps the other after it", "Deux" + f2, "Deux" + f2 + f1},
		{"a replacement carrying both changes nothing about them", "Deux" + f1 + f2, "Deux" + f1 + f2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ApplySplice(report, start, end, tc.new); !strings.Contains(got, tc.want) || strings.Count(got, "<!--") != 2 {
				t.Errorf("ApplySplice: %q does not hold %q, each token once", got, tc.want)
			}
		})
	}
	// REPLAY KEEPS THEM TOO: the same transform, located by the same rule.
	got, err := Render(report, []Op{{Old: "Two" + f1 + f2, New: ""}})
	if err != nil || !strings.Contains(got, f1+f2) {
		t.Errorf("replay of a cut dropped a token its replacement did not carry: %q, %v", got, err)
	}
	// AND THE REMOVES COME AFTER: each listed token once, in list order, by RemoveAnchorAt.
	out, err := RemoveListed(got, []string{"G-00000002", "G-00000001"})
	if err != nil || out != "One. Three.\n" {
		t.Errorf("RemoveListed after the splice:\n got %q (%v)\nwant %q", out, err, "One. Three.\n")
	}
	if _, err := RemoveListed(out, []string{"G-00000001"}); err == nil {
		t.Errorf("a remove of a token already gone was not refused")
	}
}
