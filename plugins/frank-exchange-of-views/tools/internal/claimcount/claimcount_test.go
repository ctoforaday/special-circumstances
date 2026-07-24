package claimcount

import "testing"

// The counter is only worth trusting if the RULE is pinned at its boundaries: what
// counts as one claim, what is excluded, and — the property the retire-vs-drop
// detector rests on — that it moves by exactly one when one claim leaves.
func TestCount(t *testing.T) {
	cases := []struct {
		name string
		md   string
		want int
	}{
		{"unfootnoted prose counts zero", "The sky is blue. Water is wet.", 0},
		{"one footnoted sentence", "The sky is blue[^a].", 1},
		{"two footnoted sentences", "The sky is blue[^a]. Water is wet[^b].", 2},
		{
			"a multi-marker sentence still counts once",
			"Both hold[^a][^b] together.",
			1,
		},
		{
			"footnote-definition lines are the bibliography, not claims",
			"Claim one[^a].\n\n[^a]: https://example.com\n[^b]: https://other.example",
			1,
		},
		{
			"headings are excluded even when they carry a marker",
			"# Title[^a]\n\nBody claim[^b].",
			1,
		},
		{
			"a footnoted list counts one per line",
			"- first[^a]\n- second[^b]\n- third[^c]",
			3,
		},
		{
			"a claim spanning two lines counts once",
			"This claim wraps across\ntwo physical lines[^a] before it ends.",
			1,
		},
		{
			"markers inside a code fence are literals, not claims",
			"Real claim[^a].\n\n```\nexample[^b] in code\nanother[^c]\n```\n",
			1,
		},
		{
			"tilde fences are excluded too",
			"Real claim[^a].\n\n~~~\nexample[^b]\n~~~\n",
			1,
		},
		{"empty input is zero", "", 0},
		{
			"a line break bounds the unit without punctuation",
			"first[^a]\nsecond[^b]",
			2,
		},
		{
			"punctuation bounds two claims on one line",
			"first[^a]! second[^b]?",
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
// footnoted claim must lower the count by exactly one — never zero, never two.
func TestCountMonotonicOnClaimRemoval(t *testing.T) {
	full := "Alpha[^a]. Beta[^b]. Gamma[^c].\n\n[^a]: x\n[^b]: y\n[^c]: z"
	minusOne := "Alpha[^a]. Gamma[^c].\n\n[^a]: x\n[^c]: z"
	if a, b := Count(full), Count(minusOne); a-b != 1 {
		t.Fatalf("removing one footnoted claim changed the count by %d (%d -> %d), want exactly 1", a-b, a, b)
	}
}
