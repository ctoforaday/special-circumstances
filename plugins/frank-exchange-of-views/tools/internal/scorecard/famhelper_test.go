package scorecard

import "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"

// famOfBoardT bridges a fixture board to the family pointer the rows now read; nil stays nil,
// as "record unreadable" always did.
func famOfBoardT(b *record.Board) *record.Family {
	if b == nil {
		return nil
	}
	f := record.FamilyOfBoard(b)
	return &f
}
