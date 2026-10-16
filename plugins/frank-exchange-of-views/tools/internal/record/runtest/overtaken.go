package runtest

import (
	"database/sql"
	"sync"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// Overtaken runs write while another seat's acts commit at the instant the write is about to open
// its transaction (record.WriteBegins), in one commit on that seat's own connection, db — as a seat
// in another process lands its act. The test opens db on the run's record outside the cache
// (recordsql.OpenSeparate), as only a test may. The acts land once, at the first write to begin. A
// verb that checks the record before its write's transaction has already checked by then, and admits
// the write over them; a check made inside the transaction (record.Guard) reads them.
//
// Each act is stamped as its write path would stamp it. A mint carries the class material that path
// would give it.
func Overtaken(t *testing.T, db *sql.DB, acts []*recordpb.Event, write func() error) error {
	t.Helper()
	var once sync.Once
	was := record.WriteBegins
	t.Cleanup(func() { record.WriteBegins = was })
	record.WriteBegins = func() {
		once.Do(func() {
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			for _, ev := range acts {
				ev.Ts = proto.String(record.Stamp())
				if m := ev.GetMint(); m != nil && m.ClassMaterial == nil {
					m.ClassMaterial = recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE.Enum()
				}
				if _, err := recordsql.InsertTx(tx, ev); err != nil {
					t.Fatal(err)
				}
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
		})
	}
	return write()
}
