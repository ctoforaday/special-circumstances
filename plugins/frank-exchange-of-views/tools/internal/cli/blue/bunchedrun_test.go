package blue

import (
	"slices"
	"testing"
)

// THE BUNCHED RUN'S TRIGGER, over the three measured shapes and the shapes beside them that apply.
// The first three rows are universe-m19 ev400, universe-m18 ev263 and universe-m18 ev267, their
// prose shortened and their anchors, sentence counts and positions kept.
func TestBunchedRun(t *testing.T) {
	const a, b, c, d, e = "<!--fx:F-9d782c44-->", "<!--fx:F-68c1c930-->", "<!--gap:G-5139f4f0-->", "<!--cite:C-706fda63-->", "<!--fx:F-61449326-->"
	for _, tc := range []struct {
		name, report, old, new string
		run                    []string
		was, now               int
	}{
		{"m19 ev400: one sentence to two, the run after the second",
			"The field lists terms through 88, so it does not show 91" + a + b + d + ". Next.\n",
			"The field lists terms through 88, so it does not show 91" + a + b + d + ".",
			"The field lists terms through 88, so it cannot place 91. The definition does place 91 in it" + a + b + d + ".",
			[]string{"F-9d782c44", "F-68c1c930", "C-706fda63"}, 1, 2},
		{"m18 ev263: the quote and the replacement both stop before the terminator",
			"The answer does not depend on convention" + a + c + b + ". Next.\n",
			"The answer does not depend on convention" + a + c + b,
			"The answer rests on the standard definition. A reader who adopts another should check it" + a + c + b,
			[]string{"F-9d782c44", "G-5139f4f0", "F-68c1c930"}, 1, 2},
		{"m18 ev267: two runs kept in order, the end run after the fourth sentence",
			"Three routes agree" + e + a + ", and a script reproduces each" + b + c + d + ". Next.\n",
			"Three routes agree" + e + a + ", and a script reproduces each" + b + c + d,
			"Three outputs of one program agree" + e + a + ". The Wilson test is separate. The hand product depends on none. Each output repeats" + b + c + d,
			[]string{"F-68c1c930", "G-5139f4f0", "C-706fda63"}, 1, 4},

		{"the run distributed over the new sentences applies",
			"The field stops at 88, so it does not show 91" + a + b + ". Next.\n",
			"The field stops at 88, so it does not show 91" + a + b + ".",
			"The field stops at 88" + a + ". The definition places 91" + b + ".", nil, 0, 0},
		{"the run written on an earlier sentence applies",
			"The field stops at 88, so it does not show 91" + a + b + ". Next.\n",
			"The field stops at 88, so it does not show 91" + a + b + ".",
			"The field stops at 88" + a + b + ". The definition places 91.", nil, 0, 0},
		{"a rewording that adds no sentence applies",
			"The field stops at 88, so it does not show 91" + a + b + ". Next.\n",
			"The field stops at 88, so it does not show 91" + a + b + ".",
			"The field stops at 88 and cannot place 91" + a + b + ".", nil, 0, 0},
		{"sentences added before the run's sentence, kept word for word, apply",
			"The field stops at 88" + a + b + ". Next.\n",
			"The field stops at 88" + a + b + ".",
			"The sequence has a data field. The field stops at 88" + a + b + ".", nil, 0, 0},
		{"one anchor is no run",
			"The field stops at 88, so it does not show 91" + a + ". Next.\n",
			"The field stops at 88, so it does not show 91" + a + ".",
			"The field stops at 88. The definition places 91" + a + ".", nil, 0, 0},
		{"a cut down to the run adds no sentence",
			"The field stops at 88" + a + b + ". Next.\n",
			"The field stops at 88" + a + b + ".", a + b, nil, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, applied, exact, err := validateEdit(tc.report, tc.old, tc.new)
			if err != nil {
				t.Fatalf("the edit is not legal before this rule: %v", err)
			}
			run, was, now := bunchedRun(tc.report, tc.old, applied, exact)
			if !slices.Equal(run, tc.run) || was != tc.was || len(now) != tc.now {
				t.Fatalf("bunchedRun = %v, %d, %d sentence(s) %q\nwant %v, %d, %d", run, was, len(now), now, tc.run, tc.was, tc.now)
			}
		})
	}
}
