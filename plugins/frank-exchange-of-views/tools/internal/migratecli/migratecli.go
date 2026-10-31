// Package migratecli is feov-migrate: the operator's replay-as-migration, an old record's events
// re-driven, in order, through the current write path into a fresh sibling run.
//
// It is its own binary because it is the one program that links internal/record/migrate, the
// package that reads the shapes every other reader refuses. cmd/feov-migrate is its only importer
// outside tests, and scripts/pluginparity fails any other shipped binary whose import graph holds
// either package. It is an operator's tool and has no seat: no --seat-id, no --run, no surface a
// dispatch selects.
package migratecli

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/buildid"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// Name is the binary's name, as its usage line and its refusals print it. A constant rather than
// argv[0]: the help golden is captured from a binary built under a temporary path, and a name
// read from the filesystem would make it machine-dependent.
const Name = "feov-migrate"

// The command's flags. --from is the old run, --to the fresh sibling it writes, --accept-loss a
// registry-documented untranslatable word the operator accepts BY NAME (it lands in the manifest).
const (
	FlagFrom       = "from"
	FlagTo         = "to"
	FlagAcceptLoss = "accept-loss"
)

// New builds the command.
func New() *cobra.Command {
	var from, to string
	var acceptLoss []string
	var jsonOut bool
	c := &cobra.Command{
		Use:   Name + " --from <oldRunDir> --to <newRunDir>",
		Short: "replay an old record through the current write path into a fresh run (the source is never modified)",
		Long: Name + " reads an old run's record — including a vocabulary the record tool's readers refuse — translates " +
			"each event through the authored registry, and replays them in order through the current write path under " +
			"the ORIGINAL clock. Every act is validated exactly as a seat's would be, so the output is a record " +
			"feov-record of the same release could have written; the inputs channel copies verbatim, the computation " +
			"store is rebuilt for the computations the record holds, and inputs/" + migration.ManifestName + " records " +
			"what was translated, refused, and accepted as loss, and each stored artifact the source lacked. A refusal " +
			"exits non-zero and is the tool's OUTPUT — there is no flag that skips validation.\n\n" +
			"IT WRITES THE EPOCH OF ITS RELEASE: this binary writes event-schema epoch " + fmt.Sprint(record.EventSchema) +
			", the epoch of the feov-record it ships beside, and a migrated run is read with that feov-record. It is an " +
			"operator's tool: it takes no seat id, and no seat's surface carries it.\n\n" +
			"THE PER-TURN MEASUREMENTS ARE BROUGHT ACROSS, not replayed: they are measurements taken from outside, " +
			"from transcripts that are not part of a run, so the record holds the only copy of them and the manifest " +
			"states how many came over. The manifest also states the epoch the source declared beside the one this " +
			"binary wrote, so the move the migration exists to make is a fact you can read rather than infer.\n\n" +
			"THE REPORT'S TEXT IS BROUGHT INTO THE ORDER IT SHIPS IN (" + report.TextOrder + "), where that moves whole sections and " +
			"nothing else: each part of the report is reordered and every act replayed over the parts, and the result is kept when no act is refused and " +
			"the acts give the same text section for section. Otherwise the record is replayed as written and its text keeps its " +
			"author's order. The manifest's report_order says which: " + strings.Join([]string{migrate.OrderNoReport, migrate.OrderInOrder, migrate.OrderReordered, migrate.OrderAsWritten}, ", ") + ".\n\n" +
			"THE RUN IS BUILT BESIDE --to AND MOVED INTO PLACE WHEN IT IS COMPLETE. It is built in a directory named " +
			".<name>.migrating-<random> next to --to, its files are synced, and it is then moved onto --to, so an " +
			"interrupted migration leaves --to as it was (on Windows, a --to that was an empty directory may be absent). " +
			"Its directory entries are synced where the filesystem allows — best effort: a directory that could not be " +
			"synced is a warning on stderr, and on Windows none are. A leftover .<name>.migrating-<random> directory beside " +
			"--to is what an interrupted migration left behind; nothing reads it and nothing removes it, and it is safe to " +
			"delete by hand when no migration is running.\n\n" +
			"--to must be absent, or an empty plain directory — not a link, a junction or a mount point — in a parent you " +
			"can write, with a base name short enough to take the .migrating-<random> suffix. A separated record root (" +
			record.RecordRootEnv + " set) is refused. An existing empty --to is REPLACED, not filled: its mode, owner and " +
			"attributes become the new directory's. On Unix a shell sitting in it is left in the old directory until it " +
			"changes directory; on Windows a directory in use cannot be removed, so the move fails after the build and the " +
			"built run is kept beside --to.",
		Version:       buildid.Revision(),
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if from == "" || to == "" {
				return feov.Errorf(feov.MissingField, "--from and --to are both required — the source run, and the fresh directory the migration writes")
			}
			accept := map[string]bool{}
			for _, w := range acceptLoss {
				accept[w] = true
			}
			m, err := migrate.Migrate(from, to, migrate.Entries(), migrate.Options{AcceptLoss: accept})
			if m != nil {
				reportMigration(cmd, jsonOut, m)
			}
			return err
		},
	}
	c.CompletionOptions.DisableDefaultCmd = true
	c.Flags().StringVar(&from, FlagFrom, "", "the old run directory (its records/record.db file set is read from a copy)")
	c.Flags().StringVar(&to, FlagTo, "", "the run directory to publish — absent, or an empty plain directory (not a link, junction or mount point) in a writable parent; an empty one is replaced")
	c.Flags().StringSliceVar(&acceptLoss, FlagAcceptLoss, nil,
		"accept a registry-documented loss for exactly this word (repeatable); the acceptance lands in the manifest")
	c.Flags().BoolVar(&jsonOut, flags.JSON, false, "print the manifest as JSON on stdout, and a refusal as the JSON refusal envelope, instead of text")
	return c
}

// Main runs the command over argv and returns the process's exit code: 0, or 2 for a refusal —
// which is written as the JSON refusal envelope on stdout under --json, and as `feov-migrate:
// <diagnosis>` on stderr otherwise.
func Main(argv []string, stdout, stderr io.Writer) int {
	c := New()
	c.SetArgs(argv)
	c.SetOut(stdout)
	c.SetErr(stderr)
	err := c.Execute()
	if err == nil {
		return 0
	}
	if jsonOut, _ := c.Flags().GetBool(flags.JSON); jsonOut {
		_ = json.NewEncoder(stdout).Encode(feov.RefusalEnvelope{Code: feov.CodeOf(err), Error: err.Error()})
	} else {
		fmt.Fprintf(stderr, "%s: %v\n", Name, err)
	}
	return 2
}

// reportMigration writes the manifest to stdout — as JSON under --json, as text otherwise — and each
// of the publish's warnings to stderr in both modes, since the manifest on disk cannot carry them.
func reportMigration(cmd *cobra.Command, jsonOut bool, m *migration.Manifest) {
	if jsonOut {
		_ = json.NewEncoder(cmd.OutOrStdout()).Encode(m)
	} else {
		printManifest(cmd.OutOrStdout(), m)
	}
	for _, w := range m.Warnings {
		fmt.Fprintf(cmd.ErrOrStderr(), "warning: %s\n", w)
	}
}

func printManifest(out io.Writer, m *migration.Manifest) {
	in, outN := 0, 0
	for _, n := range m.In {
		in += n
	}
	for _, n := range m.Out {
		outN += n
	}
	fmt.Fprintf(out, "migrated %s\n  %d event(s) in -> %d out; manifest at inputs/%s\n", m.SourcePath, in, outN, migration.ManifestName)
	words := make([]string, 0, len(m.In))
	for w := range m.In {
		words = append(words, w)
	}
	sort.Strings(words)
	for _, w := range words {
		if m.Out[w] != m.In[w] {
			fmt.Fprintf(out, "  %s: %d -> %d\n", w, m.In[w], m.Out[w])
		}
	}
	switch m.ReportOrder {
	case migrate.OrderReordered:
		fmt.Fprintln(out, "  the report's text is brought into the order it ships in: whole sections moved, every act replayed over it")
	case migrate.OrderAsWritten:
		fmt.Fprintln(out, "  the report's text is OUT OF THE ORDER IT SHIPS IN and is kept as written: its acts, replayed over its parts in order, do not give this text in order")
	}
	for w, why := range m.Accepted {
		fmt.Fprintf(out, "  accepted loss of %q: %s\n", w, why)
	}
	if len(m.Refusals) > 0 {
		fmt.Fprintf(out, "  %d REFUSAL(S) — each is in the manifest, and each is a decision to make, not a row to ignore\n", len(m.Refusals))
	}
}
