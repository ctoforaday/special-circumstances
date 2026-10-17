package migrate_test

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
)

func resRefusals(res *migrate.Manifest) any {
	if res == nil {
		return nil
	}
	return res.Refusals
}
