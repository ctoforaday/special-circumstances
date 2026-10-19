package report

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE CHAIR'S READS OF THE REPORT AGAINST THE INQUIRIES, EACH WITH ITS REASON (inquiries.md).
//
// An inquiry review is one event about the report as a whole, not about any one inquiry: the
// chair read the report and says, in its own words, where the body delivers the research each
// inquiry on the record claims. Its reason is that judgment, and it is rendered as written — an
// attestation of a read, never a verdict on an inquiry. A line of research the report falls short
// on is a gap about that inquiry, on the board with the rest.
//
// IN THE DOCUMENT AND NOT IN `show inquiries`. The view a seat reads is view.InquiryBody alone; a
// review's reason runs to thousands of characters, every one of them already the chair's, and
// carrying them on every seat's read of the inquiries would spend that context on each read.
func inquiryReviews(evs []*record.Event, win record.WindowIndex) string {
	var entries []string
	for _, e := range evs {
		r, ok := recordpb.BodyAs[*recordpb.InquiryReview](e)
		if !ok {
			continue
		}
		quoted := strings.Split(strings.TrimSpace(r.GetReason()), "\n")
		for i, line := range quoted {
			quoted[i] = strings.TrimRight("> "+line, " ")
		}
		entries = append(entries, fmt.Sprintf("### Epoch %d — `%s`\n\n%s", win.Of(e).Epoch, e.GetSeatId(), strings.Join(quoted, "\n")))
	}
	if len(entries) == 0 {
		return ""
	}
	return fmt.Sprintf("## Inquiry reviews (%d)\n\n", len(entries)) +
		"_Each is one read of the report against the inquiries on the record, and the reason is the chair's account of that read, in its words: where the body delivers the research each inquiry claims._\n\n" +
		strings.Join(entries, "\n\n") + "\n"
}
