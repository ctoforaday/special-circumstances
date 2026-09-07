package scorecard

import "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"

// famOfEventsT builds an events-only family fixture; nil stays nil where a test means
// "record unreadable".
func famOfEventsT(evs []*record.Event) *record.Family {
	f := record.NewFamily(nil, evs)
	return &f
}
