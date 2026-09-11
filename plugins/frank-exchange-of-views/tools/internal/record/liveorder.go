package record

import (
	"fmt"
	"sort"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// Keys is every struck act's key, sorted — the Go side of the question the `struck` view answers.
func (x StruckIndex) Keys() []string {
	out := make([]string, 0, len(x.byKey))
	for k := range x.byKey {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// StruckKeys is the SQL side: the keys of the acts the `struck` view names, sorted. measured is
// false for a record older than the correction table — it can hold no correction, and the caller
// must say "not measured" rather than read an empty list as agreement.
func StruckKeys(run Run) (keys []string, measured bool, err error) {
	return correctionQuery(run, `SELECT e."key" FROM "struck" s JOIN "events" e ON e."id" = s."event_id" ORDER BY e."key"`)
}

// LiveKeys is the SQL order of the acts that stand: every row of live_event by pos, as event keys
// ("" for an event with none) — the order Live gives in Go, answered by the view instead.
func LiveKeys(run Run) (keys []string, measured bool, err error) {
	return correctionQuery(run, `SELECT COALESCE(e."key", '') FROM "live_event" l JOIN "events" e ON e."id" = l."event_id"
	  ORDER BY l."pos", l."event_id"`)
}

func correctionQuery(run Run, q string) ([]string, bool, error) {
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return nil, false, err
	}
	has, err := recordsql.HasTable(db, recordpb.Word(recordpb.EventType_EVENT_TYPE_CORRECTION))
	if err != nil || !has {
		return nil, false, err
	}
	rows, err := db.Query(q)
	if err != nil {
		return nil, false, fmt.Errorf("record: asking the correction views: %w", err)
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, false, err
		}
		out = append(out, k)
	}
	return out, true, rows.Err()
}
