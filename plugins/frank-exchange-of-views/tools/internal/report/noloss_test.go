package report

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// EVERY LINE OF THE REPORT'S TEXT IS IN THE report.md THAT SHIPS, through the whole assembly: the
// parting of the text among the rows, the composition, the token strip and the weave. Each case
// is a shape a report can take that one of those steps could mistake for something to leave out.
func TestNoLineOfTheReportsTextIsLostInAssembly(t *testing.T) {
	const fx = "<!--fx:F-00000aaa-->"
	cases := map[string]string{
		"no headings at all":              "one paragraph.\n\nand another.\n",
		"text before the title":           "prose before the title.\n# T\n**Verdict:** a preamble line\n\n## TL;DR\nx.\n",
		"a lifted heading twice":          "# T\n\n## Open questions\nfirst.\n\n## Open questions\nsecond.\n\n##  open questions (in full)\nthird.\n",
		"a lifted heading, other case":    "# T\n\n## ANALYSIS\nupper.\n\n## tl;dr\nlower.\n",
		"an anchored heading":             "# T\n\n## The Catechism" + fx + "\nanchored heading.\n",
		"a section named as the body's row": "# T\n\n## TL;DR\nx.\n\n## " + bodyHeading + "\na section of that name.\n\n## Technical foundations\ny.\n",
		"an empty section":                "# T\n\n## Analysis\n\n## Open questions\n\n## Appendix\n",
		"a heading in a fence":            "# T\n\n## Analysis\n```\n## Risk matrix\n# not a title\n```\nafter the fence.\n",
		"a tilde fence in a tick fence":   "# T\n\n## Analysis\n```\n~~~\n## Open questions\n```\nafter.\n",
		"a heading in a list":             "# T\n\n## Analysis\n- item\n  ## Nested heading\n  under it.\n",
		"an indented code heading":        "# T\n\n## Analysis\n\n    ## TL;DR\n    indented code.\n\nafter.\n",
		"CRLF":                            "# T\r\n\r\n## TL;DR\r\nx.\r\n\r\n## Appendix\r\ny.\r\n",
		"a heading the tool composes":     "# T\n\n## Alternatives considered\nweighed.\n\n## Risk Matrix\na risk.\n\n## Additional findings\nmore.\n\n## Bibliography\na reference.\n",
		"sections the old list dropped":   "# T\n\n## Verdict\nv.\n\n## Footnotes\nf.\n\n## Expansions\ne.\n\n## Blue team report\nb.\n\n## Claims withdrawn\nc.\n\n## How this run was conducted\nh.\n",
		"a title late in the text":        "## Analysis\nbody.\n# Late title\nafter the title.\n",
		"two titles":                      "# First\n\n## Analysis\nbody.\n\n# Second\n\nafter the second.\n",
		"a table whose rows hold a token": "# T\n\n## Analysis\n\n| n | divides |\n|---|---|" + fx + "\n| ✓ | ✗ |" + fx + "\n\n---" + fx + "\n\n```\n}" + fx + "\n```\n",
		"an unclosed fence":               "# T\n\n## Analysis\n```\ncode\n## Open questions\nnever closed.\n",
	}
	for name, blue := range cases {
		t.Run(name, func(t *testing.T) {
			runDir := newRun(t)
			run := runtest.Open(t, runDir)
			id := record.Identity{Run: run, SeatID: "blue-synthesize"}
			if _, _, err := record.RegisterSeat(id, "", ""); err != nil {
				t.Fatal(err)
			}
			if _, err := record.Append(id, &recordpb.BaseIngest{Text: proto.String(blue)}); err != nil {
				t.Fatal(err)
			}
			docs, err := AssembleAll(run)
			if err != nil {
				t.Fatal(err)
			}
			shipped := render(docs[0], Title(run), docs)
			have := map[string]int{}
			for _, ln := range strings.Split(shipped, "\n") {
				have[strings.TrimSpace(ln)]++
			}
			want := map[string]int{}
			for _, ln := range strings.Split(blue, "\n") {
				if ln = strings.TrimSpace(anchor.StripAssembled(ln)); ln != "" {
					want[ln]++
				}
			}
			for ln, n := range want {
				if have[ln] < n {
					t.Errorf("the line %q is in the report's text %d time(s) and in report.md %d:\n%s", ln, n, have[ln], shipped)
				}
			}
		})
	}
}
