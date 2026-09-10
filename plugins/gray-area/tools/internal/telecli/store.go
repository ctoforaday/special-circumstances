package telecli

import (
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

// openRead opens the store for reading. Every verb but backfill goes through here, so a missing
// store is one message rather than six.
//
// A REBUILT STORE IS LOUD UNTIL BACKFILLED. An upgrade that rebuilds the store leaves it empty,
// and an empty store answers `agents`, `touched` and `sql` with plausible zeros — the healthy case
// and the gutted one print the same bytes. So while the store carries a rebuild no backfill has
// followed, every read verb says so on w, once, and then runs anyway: the rows it has are still
// true, and the exit code is the verb's own. A marker that cannot be read is said too, as "not
// measured", rather than taken for a store that needs nothing.
func (e *Env) openRead(w io.Writer) (*sql.DB, error) {
	if e.Store == "" {
		return nil, fmt.Errorf("no catalogue path: neither --store nor a home directory could be resolved")
	}
	db, err := catalogue.OpenRead(e.Store)
	if err != nil {
		return nil, err
	}
	switch at, from, pending, err := catalogue.RebuildPending(db); {
	case err != nil:
		fmt.Fprintf(w, "telepathy: could not read the store's rebuild markers (%v) — whether it needs a backfill is not measured\n", err)
	case pending:
		fmt.Fprintf(w, "telepathy: the store was rebuilt at %s (it was stamped %d) and has not been backfilled since — "+
			"sessions from before then are missing; run `telepathy backfill`\n", at.UTC().Format(time.RFC3339), from)
	}
	return db, nil
}

// ago renders a timestamp against THIS invocation's clock rather than the wall clock, which is
// what lets a golden file exist at all.
func (e *Env) ago(ts int64) string {
	d := e.Now().Sub(time.Unix(ts, 0))
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
