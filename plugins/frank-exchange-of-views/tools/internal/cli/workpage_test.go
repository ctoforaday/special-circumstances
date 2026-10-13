package cli

import (
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// workPageOf is the `show work` page of one tree, as its `--help` prints it.
func workPageOf(t *testing.T, root *cobra.Command) string {
	t.Helper()
	c, _, err := root.Find([]string{"show", "work"})
	if err != nil || c.Name() != "work" {
		t.Fatalf("no `show work` on this tree: %v", err)
	}
	return c.Long
}

// `show work` TELLS EACH SEAT ABOUT ITS OWN LIST.
//
// record.WorkOfSeat builds the list per role: the mint budget, the regrades owed and whether blue
// answered are a lens's, `engaged` is blue's response sitting's, and an item another seat clears is
// the chair's. A paragraph about a field the reader's list never holds cannot be acted on, and the
// page is read at every sitting's opening act — it reaches a seat's constitution through this same
// tree — so each paragraph is on the page of the roles below and on no other seat's.
func TestShowWorkCarriesTheParagraphsOfTheRoleItIsRenderedFor(t *testing.T) {
	every := []string{"lens", "blue", "chair", "bench"}
	paragraphs := []struct {
		opens string
		roles []string
	}{
		{"ON THE CHAIR'S LIST a blocking item holds a PASS", []string{"chair"}},
		{"FOR EACH GAP YOU MINTED AND LEFT OPEN", []string{"lens"}},
		{"A LENS'S LIST STATES ITS MINT BUDGET", []string{"lens"}},
		{"AN ACCEPTED GRADE MOTION WAITS FOR THE MINTING LENS'S REGRADE", []string{"lens"}},
		{"WHERE THE LENS PRESCRIBED EXACT TEXT", []string{"lens", "blue"}},
		{"`engaged` IS BLUE'S RESPONSE SITTING'S", []string{"blue"}},
		// Every seat's: what the list is, what does not block, what an open gap carries, what is
		// estopped.
		{"EVERYTHING OPEN TO YOU, in one list", every},
		{"An item with `blocks: false` is work nobody will refuse you for skipping", every},
		{"`open` holds OPEN gaps only", every},
		{"`estopped` IS WHAT YOU MAY NOT RE-RAISE", every},
	}
	for _, role := range every {
		page := workPageOf(t, NewRootFor(record.SampleSeatOf(role)))
		for _, p := range paragraphs {
			want := slices.Contains(p.roles, role)
			if got := strings.Contains(page, p.opens); got != want {
				t.Errorf("the %s seat's `show work` page holds the paragraph opening %q: %v, want %v", role, p.opens, got, want)
			}
		}
	}

	// A TREE BUILT FOR NO ROLE HOLDS EVERY PARAGRAPH: it is the whole page, which the operator
	// reads to see what any seat is told.
	whole := &cobra.Command{Use: "feov-record"}
	whole.AddCommand(seat.Show(""))
	page := workPageOf(t, whole)
	for _, p := range paragraphs {
		if !strings.Contains(page, p.opens) {
			t.Errorf("the page built for no role lacks the paragraph opening %q", p.opens)
		}
	}
}
