package migrate_test

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"google.golang.org/protobuf/proto"
)

// replayed appends as Replay does, with record.Migrating set.
func replayed(id record.Identity, body proto.Message) (*record.Event, error) {
	record.Migrating = true
	defer func() { record.Migrating = false }()
	return record.Append(id, body)
}

func resRefusals(res *migrate.Manifest) any {
	if res == nil {
		return nil
	}
	return res.Refusals
}
