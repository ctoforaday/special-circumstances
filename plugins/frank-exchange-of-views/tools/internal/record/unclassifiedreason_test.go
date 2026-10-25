package record

import (
	"database/sql"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// THE WRITE AND THE DATABASE AGREE ON A BLANK REASON. An `unclassified` mint's reason is refused by
// requireUnclassifiedReason and by Mint's (check); each reason here is put to both, the check's own
// expression evaluated by SQLite, and the two answer alike — a vertical tab and a no-break space are
// words to both, space, tab, LF and CR blank to both.
func TestTheWriteAndTheDatabaseAgreeOnABlankReason(t *testing.T) {
	var expr string
	for _, c := range recordsql.MessageChecks((&recordpb.Mint{}).ProtoReflect().Descriptor()) {
		if strings.Contains(c.GetExpr(), "'unclassified'") {
			expr = c.GetExpr()
		}
	}
	if expr == "" {
		t.Fatal("Mint carries no check on an unclassified reason")
	}
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	for _, reason := range []string{"", " ", " \t\n\r ", "\v", " ", " ", " x ", "the nearest class is spec_underspecification"} {
		var dbAccepts bool
		if err := db.QueryRow(`SELECT `+expr+` FROM (SELECT 'unclassified' AS "class", ? AS "mint_reason")`, reason).Scan(&dbAccepts); err != nil {
			t.Fatalf("evaluating the check on %q: %v", reason, err)
		}
		writeAccepts := requireUnclassifiedReason(&recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), Class: recordpb.GapClass_GAP_CLASS_UNCLASSIFIED.Enum(), MintReason: proto.String(reason)}) == nil
		if dbAccepts != writeAccepts {
			t.Errorf("reason %q: the database accepts %v, the write %v", reason, dbAccepts, writeAccepts)
		}
	}
}
