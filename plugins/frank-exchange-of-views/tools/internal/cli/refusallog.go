package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// noteRefusal writes the refusal the tool just gave a seat to the log, as the TOOL's entry.
//
// THE SEAT DOES NOT HAVE TO. On universe-m13 every seat met refusals — `--key` on a verb that has
// none, `--acceptance-check` for the flag `--check`, a mint past its budget — and not one reached
// the log: each seat classed its own as a mistake, which is what the log's help then told it. The
// guess is the operator's signal either way: a flag a seat reached for is a flag the surface made
// it expect. The tool knows every refusal it issues, so the fact is recorded where it happens and
// the seat's entry is left to say only what the tool cannot know — what it expected, and why.
//
// NAMES, NEVER VALUES. The flags given are listed by name; their values are prose a seat wrote, and
// the entry is about the call's shape.
//
// Best-effort and silent: the refusal is what the seat must see, and a failure to record it must
// not replace it. A refusal to the operator, or on a run that did not resolve, records nothing.
func noteRefusal(cmd *cobra.Command, err error) {
	if err == nil || cmd == nil || record.IsToolLogged(err) {
		return
	}
	sc := seat.Of(cmd)
	if sc.SeatID == "" || sc.SeatID == record.OperatorRole {
		return
	}
	run, rerr := sc.Run()
	if rerr != nil {
		return
	}
	// A SITTING TO ATTRIBUTE IT TO. A seat that never registered has none, and the record refuses an
	// append from it; a run with an outcome is over, and its record is closed to seats — a refusal
	// then belongs to no sitting and would land after the documents the run assembled.
	if n, err := record.SittingsOf(run, sc.SeatID); err != nil || n == 0 || record.RecordedOutcome(run) != "" {
		return
	}
	var given []string
	cmd.Flags().Visit(func(f *pflag.Flag) {
		switch f.Name {
		case flags.Run, flags.SeatID, flags.JSON:
		default:
			given = append(given, "--"+f.Name)
		}
	})
	sort.Strings(given)
	with := ""
	if len(given) > 0 {
		with = " with " + strings.Join(given, " ")
	}
	// THE ANNOTATION LAYER IS NOT TEXT. A refusal may quote an anchor it could not resolve, and a live
	// `<!--fx:…-->` in a log entry is a marker in a document that renders it.
	why, _, _ := strings.Cut(strings.TrimSpace(anchortext.Visible(err.Error())), "\n")
	if r := []rune(why); len(r) > 300 {
		why = string(r[:300]) + "…"
	}
	what := "a verb this surface does not have"
	if cmd != cmd.Root() {
		what = "`" + strings.TrimPrefix(cmd.CommandPath(), cmd.Root().Name()+" ") + "`"
	}
	lt, src := recordpb.LogType_LOG_TYPE_REFUSAL, recordpb.LogSource_LOG_SOURCE_TOOL
	_, _ = record.Append(record.Identity{Run: run, SeatID: sc.SeatID}, &recordpb.Log{
		Text: proto.String(fmt.Sprintf("refused %s%s: %s", what, with, why)),
		Type: &lt, Source: &src,
	})
}

// persistentOnly keeps the root's own flags from an argv, with their values, so a refusal raised
// before cobra dispatches can still resolve the run and the seat. Everything else — the unknown
// verb, its flags — is dropped, because the root would refuse to parse it.
func persistentOnly(root *cobra.Command, argv []string) []string {
	var out []string
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		if !strings.HasPrefix(a, "--") {
			continue
		}
		name, _, inline := strings.Cut(strings.TrimPrefix(a, "--"), "=")
		f := root.PersistentFlags().Lookup(name)
		if f == nil {
			continue
		}
		out = append(out, a)
		if !inline && f.NoOptDefVal == "" && i+1 < len(argv) {
			out = append(out, argv[i+1])
			i++
		}
	}
	return out
}

// refuseUnknownCommandNoted is refuseUnknownCommandFirst with its refusal logged. A VERB THAT DOES
// NOT EXIST IS THE REFUSAL THE LOG MOST WANTS — the act a seat reached for and the surface lacked —
// and it is answered before cobra parses a flag, so the run and seat are read off the argv here.
// One function for the binary and the test harness, so the harness cannot measure a binary that
// logs differently from the one that ships.
func refuseUnknownCommandNoted(root *cobra.Command, argv []string, seatID string) error {
	err := refuseUnknownCommandFirst(root, argv, seatID)
	if err != nil && len(argv) > 1 {
		_ = root.ParseFlags(persistentOnly(root, argv[1:]))
		noteRefusal(root, err)
	}
	return err
}
