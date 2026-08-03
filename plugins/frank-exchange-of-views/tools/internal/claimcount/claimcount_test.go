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
		{"one cited sentence", "The sky is blue<!--cite:c-a-->.", 1},
		{"two cited sentences", "The sky is blue<!--cite:c-a-->. Water is wet<!--cite:c-b-->.", 2},
		{
			"a multi-anchor sentence still counts once",
			"Both hold<!--cite:c-a--><!--cite:c-b--> together.",
			1,
		},
		{
			"a finding anchor is not a claim and never counts",
			"A finding sits here<!--fx:f-a--> but nothing cites it.",
			0,
		},
		{
			"footnote-definition lines are not claims",
			"Claim one<!--cite:c-a-->.\n\n[^a]: https://example.com\n[^b]: https://other.example",
			1,
		},
		{
			"headings are excluded even when they carry an anchor",
			"# Title<!--cite:c-a-->\n\nBody claim<!--cite:c-b-->.",
			1,
		},
		{
			"a cited list counts one per line",
			"- first<!--cite:c-a-->\n- second<!--cite:c-b-->\n- third<!--cite:c-c-->",
			3,
		},
		{
			"a claim spanning two lines counts once",
			"This claim wraps across\ntwo physical lines<!--cite:c-a--> before it ends.",
			1,
		},
		{
			"anchors inside a code fence are literals, not claims",
			"Real claim<!--cite:c-a-->.\n\n```\nexample<!--cite:c-b--> in code\nanother<!--cite:c-c-->\n```\n",
			1,
		},
		{
			"tilde fences are excluded too",
			"Real claim<!--cite:c-a-->.\n\n~~~\nexample<!--cite:c-b-->\n~~~\n",
			1,
		},
		{"empty input is zero", "", 0},
		{
			"a line break bounds the unit without punctuation",
			"first<!--cite:c-a-->\nsecond<!--cite:c-b-->",
			2,
		},
		{
			"punctuation bounds two claims on one line",
			"first<!--cite:c-a-->! second<!--cite:c-b-->?",
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
	full := "Alpha<!--cite:c-a-->. Beta<!--cite:c-b-->. Gamma<!--cite:c-c-->."
	minusOne := "Alpha<!--cite:c-a-->. Gamma<!--cite:c-c-->."
	if a, b := Count(full), Count(minusOne); a-b != 1 {
		t.Fatalf("removing one cited claim changed the count by %d (%d -> %d), want exactly 1", a-b, a, b)
	}
}
