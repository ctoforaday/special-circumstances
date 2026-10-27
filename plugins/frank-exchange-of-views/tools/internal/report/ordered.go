package report

import (
	"slices"
	"strings"
)

// Ordered is text with its sections in the order it ships in: the title line first, each section
// whole at its row's place, and sections of one rank in the order written. Text that is in order
// is returned as it is; in text that is not, one blank line parts each section from the next.
// `migrate` brings a record written before the write held the order forward with it.
func Ordered(text string) string {
	parsed := ParseText(text)
	if len(orderClaims(parsed)) == 0 {
		return text
	}
	all, _, _ := stands(parsed)
	chunks, ranks := [][]string{nil}, []int{0} // the first is the title line's
	for i, ln := range strings.Split(text, "\n") {
		to := len(chunks) - 1
		if len(all) > 0 && all[0].at == i {
			if to = 0; all[0].rank != 0 {
				chunks, ranks = append(chunks, nil), append(ranks, all[0].rank)
				to = len(chunks) - 1
			}
			all = all[1:]
		}
		chunks[to] = append(chunks[to], ln)
	}
	var out []string
	for rank := range slices.Max(ranks) + 1 {
		for i, c := range chunks {
			if s := strings.Trim(strings.Join(c, "\n"), "\n"); ranks[i] == rank && s != "" {
				out = append(out, s)
			}
		}
	}
	return strings.Join(out, "\n\n") + text[len(strings.TrimRight(text, "\n")):]
}
