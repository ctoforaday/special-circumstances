package migrate_test

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
)

func resRefusals(res *migration.Manifest) any {
	if res == nil {
		return nil
	}
	return res.Refusals
}
