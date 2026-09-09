package telecli

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

// openRead opens the store for reading. Every verb but backfill goes through here, so a missing
// store is one message rather than six.
func (e *Env) openRead() (*sql.DB, error) {
	if e.Store == "" {
		return nil, fmt.Errorf("no catalogue path: neither --store nor a home directory could be resolved")
	}
	return catalogue.OpenRead(e.Store)
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
