package record

import (
	"database/sql"
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/compute"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// A COMPUTATION'S VERDICT HAS THREE OWNERS (gblock, 2026-10-12, of what the record now calls a
// computation: "the anchor is about the quote, the proof is about the proof itself"; and the
// re-run ruling of plans/compute-and-show.md §III.1.4).
//
//   - THE RE-RUN belongs to the BUNDLE. A bundle is addressed by its sha and never changes, so it is
//     executed once: the latest `reproduce` that RAN it and names its execution (`rerun_sha`) is the
//     bundle's standing re-run, and a later `reproduce` of any act on the bundle executes nothing.
//   - THE COMPARISON belongs to the ACT. Two acts on one bundle can have recorded different results —
//     a `variant` program records what its first run wrote — so each act is compared against the
//     standing re-run with its OWN recorded output, exit and out/, and its verdict is its own latest
//     reproduce's. An act on a bundle with a standing re-run that nobody compared is `not_compared`, never another
//     act's verdict.
//   - SOUND belongs to the ANCHOR: does the program and its result support what THAT sentence says.
//     It is read off the anchor's standing checks (anchorchecks.go), so a standing `unsound` from
//     one lens stays visible beside another lens's `sound` until the anchor is answered — by its own
//     seat's later verdict, or by an edit that reopens the sentence.
//
// An anchor is VERIFIED when its own comparison reproduced AND its own sentence is judged sound.

// The standing state of a computation anchor's sentence, as every reader spells it.
const (
	// SoundnessSound: a standing check judged this sentence sound, and none judged it unsound.
	SoundnessSound = "sound"
	// SoundnessUnsound: a standing check judged this sentence unsound.
	SoundnessUnsound = "unsound"
	// SoundnessNotJudged: no check of this sentence has been made.
	SoundnessNotJudged = "not_judged"
	// SoundnessReopened: an edit changed this sentence after it was judged, and no judgement of the
	// sentence as it stands has been made since.
	SoundnessReopened = "reopened"
)

// The states of an act no comparison of its own has reached, as every reader spells them.
const (
	// RerunNone: no checker executed the act's bundle.
	RerunNone = "not_rerun"
	// RerunNotCompared: the bundle has a standing re-run, and nobody compared this act against it.
	RerunNotCompared = "not_compared"
)

// standingRerunsOf is every bundle's standing re-run, keyed by bundle_sha: the latest `reproduce`
// that ran it and names its execution by rerun_sha. A `reproduce` that compared without running the
// bundle (`ran` = false) executed nothing, and a bundle no reproduce ran has none.
func standingRerunsOf(evs []*Event) map[string]*Event {
	out := map[string]*Event{}
	for _, e := range evs {
		r, ok := recordpb.BodyAs[*recordpb.Reproduce](e)
		if !ok || r.GetBundleSha() == "" || !r.GetRan() || r.GetRerunSha() == "" {
			continue
		}
		out[r.GetBundleSha()] = e
	}
	return out
}

// actVerdictsOf is each act's comparison, keyed by computation_id: its latest `reproduce`, which
// compared that act's own recorded result against an execution of its bundle.
func actVerdictsOf(evs []*Event) map[string]*Event {
	out := map[string]*Event{}
	for _, e := range evs {
		if r, ok := recordpb.BodyAs[*recordpb.Reproduce](e); ok && r.GetComputationId() != "" {
			out[r.GetComputationId()] = e
		}
	}
	return out
}

// validateReproduce is every writer's check of a reproduce's shape: the anchor names a computation,
// the sha is that computation's, the act says what it compared, and a bundle no checker can run is
// never judged sound. Which execution the act compared against is the live write's act judgement
// (requireNamedExecution).
func validateReproduce(q recordsql.Querier, b *recordpb.Reproduce) error {
	sha, err := requireComputationAnchor(q, b.GetComputationId(), "--anchor")
	if err != nil {
		return err
	}
	if b.GetBundleSha() != sha {
		return fmt.Errorf("reproduce names computation %s and carries the sha %q, which is not that computation's bundle (%s) — the sha is the anchor's, derived by the tool, never stated", b.GetComputationId(), b.GetBundleSha(), sha)
	}
	if b.GetRerun() == recordpb.Rerun_RERUN_UNSPECIFIED {
		return fmt.Errorf("reproduce of %s records no comparison: a reproduce says what its bundle's re-run showed against this act's recorded result", b.GetComputationId())
	}
	if b.GetRerun() == recordpb.Rerun_RERUN_CALLS_THE_TOOL {
		return refuseSoundOverTheTool(b)
	}
	return nil
}

// requireNamedExecution is a reproduce's act judgement (judgeAct): the act names the execution it
// compared against by its rerun_sha — its own where it ran, else the bundle's standing re-run — and
// the tool names it before it writes. A migrated reproduce whose re-run's output the archive holds
// only in part names none, and the replay writes it as the archive holds it.
//
// AN ACT THAT RAN NOTHING NAMES THE STANDING RE-RUN, and it must still be the standing one: two
// checkers can execute one bundle at once, both executions are recorded, and the later stands, so an
// act that read the earlier and writes after the later compared against a re-run that no longer
// stands. Against a standing execution that called the tool there is no output to compare, so the act
// records `calls_the_tool` too, and no act invents a call its standing execution did not make.
func requireNamedExecution(q recordsql.Querier, b *recordpb.Reproduce) error {
	if b.GetRerunSha() == "" {
		return fmt.Errorf("reproduce of %s names no execution — the tool names the re-run the act compared against (rerun_sha), its own where it ran, else the bundle's standing one, before it writes, and this write skipped that", b.GetComputationId())
	}
	if b.GetRan() {
		return nil
	}
	var standingSha, standingRerun sql.NullString
	live, err := queryRowAt(q, []any{&standingSha, &standingRerun}, `SELECT "rerun_sha", "rerun" FROM "reproduce" WHERE "bundle_sha" = ? AND "ran" = 1 AND "rerun_sha" IS NOT NULL ORDER BY "event_id" DESC LIMIT 1`, b.GetBundleSha())
	if err != nil {
		return err
	}
	if !live {
		return fmt.Errorf("the act on %s ran nothing and names the re-run %s, which no act on the record executed — an act that ran nothing compares against the re-run standing for its bundle", b.GetComputationId(), compute.Short(b.GetRerunSha()))
	}
	if b.GetRerunSha() != standingSha.String {
		return fmt.Errorf("the act on %s ran nothing and compared against the re-run %s, but the re-run standing for its bundle is now %s — another execution was recorded after this act read the record, so its comparison is against a re-run that no longer stands; call again to compare against the standing one", b.GetComputationId(), compute.Short(b.GetRerunSha()), compute.Short(standingSha.String))
	}
	calls := recordpb.Word(recordpb.Rerun_RERUN_CALLS_THE_TOOL)
	if (standingRerun.String == calls) != (b.GetRerun() == recordpb.Rerun_RERUN_CALLS_THE_TOOL) {
		return fmt.Errorf("the act on %s ran nothing and records %s, but the re-run it compared against is %s — an execution that called the tool has no output to compare, and is the only one an act records `%s` against", b.GetComputationId(), recordpb.Word(b.GetRerun()), standingRerun.String, calls)
	}
	return nil
}

// refuseSoundOverTheTool refuses a `sound` verdict on a bundle whose re-run called the tool.
func refuseSoundOverTheTool(b *recordpb.Reproduce) error {
	if b.GetSoundness() == recordpb.Soundness_SOUNDNESS_SOUND {
		return fmt.Errorf("computation %s's bundle calls feov-record, which refuses every call made inside a re-run, so no checker can run it to its output — that is `--as unsound`, never sound", b.GetComputationId())
	}
	return nil
}

// computationVerdict is one computation anchor's standing verdict.
type computationVerdict struct {
	// comparison is this act's latest reproduce, nil when nobody compared it.
	comparison *Event
	// standing is the bundle's standing re-run, nil when nobody ran it.
	standing *Event
	// soundness is one of the Soundness* words.
	soundness string
	// judgements are the anchor's standing soundness verdicts, one per seat, in record order.
	judgements []anchorCheck
}

// compared is this act's comparison, nil when it has none.
func (v computationVerdict) compared() *recordpb.Reproduce {
	if v.comparison == nil {
		return nil
	}
	r, _ := recordpb.BodyAs[*recordpb.Reproduce](v.comparison)
	return r
}

// reproduced is whether this act's own comparison found the re-run wrote its recorded result.
func (v computationVerdict) reproduced() bool {
	r := v.compared()
	return r != nil && r.GetRerun() == recordpb.Rerun_RERUN_REPRODUCED
}

// verified is the ruling's conjunction: this act reproduced and its sentence is judged sound.
func (v computationVerdict) verified() bool { return v.reproduced() && v.soundness == SoundnessSound }

// rerunWord is the act's state as every reader prints it: its comparison, or not_compared where its
// bundle has a standing re-run and nobody compared it, or not_rerun where nobody executed the bundle.
func (v computationVerdict) rerunWord() string {
	if r := v.compared(); r != nil {
		return recordpb.Word(r.GetRerun())
	}
	if v.standing != nil {
		return RerunNotCompared
	}
	return RerunNone
}

// computationVerdictOf is one act's standing verdict — computation id on bundle sha — read from the
// ledger (s, nil where the anchor has none), the bundles' standing re-runs and the acts' comparisons.
func computationVerdictOf(id, sha string, s *anchorStanding, reruns map[string]*Event, acts map[string]*Event) computationVerdict {
	v := computationVerdict{soundness: SoundnessNotJudged, comparison: acts[id], standing: reruns[sha]}
	if s == nil {
		return v
	}
	// A RETIRED ANCHOR KEEPS THE VERDICTS IT LEFT WITH. Nothing is owed on a sentence the report no
	// longer carries, which is what `stands` answers for the work list; what its sentence was judged
	// to be is still the record's, and its row read at its id (EvidenceAt) shows it. No document does:
	// evidence.md and the evidence table's lists hold the live anchors.
	if s.retired {
		kept := *s
		kept.retired = false
		s = &kept
	}
	v.judgements = s.standing()
	switch {
	case s.holds():
		v.soundness = SoundnessSound
	case len(v.judgements) > 0:
		v.soundness = SoundnessUnsound
	case s.edit != 0 && s.state() == stateUnchecked && s.checkedBefore(s.edit):
		v.soundness = SoundnessReopened
	}
	return v
}
