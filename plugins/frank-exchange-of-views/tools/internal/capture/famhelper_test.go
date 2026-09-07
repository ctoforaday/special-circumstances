package capture

import "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"

// famOfBoardT bridges a fixture board to the family pointer scorecards read; nil stays nil.
func famOfBoardT(b *record.Board) *record.Family {
	if b == nil {
		return nil
	}
	f := record.FamilyOfBoard(b)
	return &f
}
