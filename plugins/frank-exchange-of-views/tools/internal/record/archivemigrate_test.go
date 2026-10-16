package record_test

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
)

// The tests in package record that read an archived run read it MIGRATED: an archive's event
// tables are its creating binary's, and migrate is the one thing that brings a record forward.
// Package record cannot import migrate, which imports it; this package can, and hands the
// migration to them before any test runs.
func init() {
	*record.ArchiveMigrator = func(src, dst string) error {
		_, err := migrate.Migrate(src, dst, migrate.Entries(), migrate.Options{})
		return err
	}
}
