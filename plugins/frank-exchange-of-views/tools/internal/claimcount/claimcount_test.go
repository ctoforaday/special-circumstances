package claimcount

import "testing"

// The counter is only worth trusting if the RULE is pinned at its boundaries: what
// counts as one claim, what is excluded, and — the property the retire-vs-drop
// detector rests on — that it moves by exactly one when one claim leaves. The claim unit
// is now the tool-inserted citation anchor "<!--cite:c-…-->", not the old "[^label]".
func TestCount(t *testing.T) {
	cases := []struct {
		name string
		md   string
		want int
	}{
		{"uncited prose counts zero", "The sky is blue. Water is wet.", 0},
		{"one cited sentence", "The sky is blue<!--cite:C-0000000a-->.", 1},
		{"two cited sentences", "The sky is blue<!--cite:C-0000000a-->. Water is wet<!--cite:C-0000000b-->.", 2},
		{
			// Per citation: a merge of two cited sentences into one carries both anchors and
			// must not read as a claim lost.
			"a multi-anchor sentence counts each citation",
			"Both hold<!--cite:C-0000000a--><!--cite:C-0000000b--> together.",
			2,
		},
		{
			// Distinct labels only: the same anchor twice in one sentence is one claim.
			"an anchor repeated in one sentence counts once",
			"Repeating the same anchor<!--cite:C-0000000a--> twice<!--cite:C-0000000a--> in one sentence.\n",
			1,
		},
		{
			"a finding anchor is not a claim and never counts",
			"A finding sits here<!--fx:F-0000000a--> but nothing cites it.",
			0,
		},
		{
			"footnote-definition lines are not claims",
			"Claim one<!--cite:C-0000000a-->.\n\n[^a]: https://example.com\n[^b]: https://other.example",
			1,
		},
		{
			"headings are excluded even when they carry an anchor",
			"# Title<!--cite:C-0000000a-->\n\nBody claim<!--cite:C-0000000b-->.",
			1,
		},
		{
			"a cited list counts one per item",
			"- first<!--cite:C-0000000a-->\n- second<!--cite:C-0000000b-->\n- third<!--cite:C-0000000c-->",
			3,
		},
		{
			"a claim spanning two lines counts once",
			"This claim wraps across\ntwo physical lines<!--cite:C-0000000a--> before it ends.",
			1,
		},
		{
			"anchors inside a code fence are literals, not claims",
			"Real claim<!--cite:C-0000000a-->.\n\n```\nexample<!--cite:C-0000000b--> in code\nanother<!--cite:C-0000000c-->\n```\n",
			1,
		},
		{
			"tilde fences are excluded too",
			"Real claim<!--cite:C-0000000a-->.\n\n~~~\nexample<!--cite:C-0000000b-->\n~~~\n",
			1,
		},
		{"empty input is zero", "", 0},
		{
			"a soft wrap joins one sentence, and each anchor after prose in it counts",
			"first<!--cite:C-0000000a-->\nsecond<!--cite:C-0000000b-->",
			2,
		},
		{
			"punctuation bounds two claims on one line",
			"first<!--cite:C-0000000a-->! second<!--cite:C-0000000b-->?",
			2,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Count(tc.md); got != tc.want {
				t.Errorf("Count() = %d, want %d\n---\n%s\n---", got, tc.want, tc.md)
			}
		})
	}
}

// MONOTONICITY is the load-bearing property: the capture detector reads an
// unaccounted fall in claim_count as a dropped claim, so removing exactly one
// cited claim must lower the count by exactly one — never zero, never two.
func TestCountMonotonicOnClaimRemoval(t *testing.T) {
	full := "Alpha<!--cite:C-0000000a-->. Beta<!--cite:C-0000000b-->. Gamma<!--cite:C-0000000c-->."
	minusOne := "Alpha<!--cite:C-0000000a-->. Gamma<!--cite:C-0000000c-->."
	if a, b := Count(full), Count(minusOne); a-b != 1 {
		t.Fatalf("removing one cited claim changed the count by %d (%d -> %d), want exactly 1", a-b, a, b)
	}
}
