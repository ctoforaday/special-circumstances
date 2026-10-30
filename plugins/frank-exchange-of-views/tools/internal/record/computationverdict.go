package record

import (
	"database/sql"
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// A COMPUTATION'S VERDICT HAS TWO OWNERS (gblock, 2026-10-12, of what the record now calls a
// computation): "the anchor is about the quote, the proof is about the proof itself".
//
//   - THE RE-RUN belongs to the BUNDLE. The bundle's standing re-run is the latest `reproduce` of
//     its sha that RAN it, and it stands for every anchor on that bundle whatever it showed: the
//     bytes that ran are the same bytes whichever sentence they back, and a sha names bytes that
//     never change, so `reproduce` runs a bundle once and a changed bundle is a new computation.
//   - SOUND belongs to the ANCHOR: does the program and its result support what THAT sentence says.
//     It is read off the anchor's standing checks (anchorchecks.go), so a standing `unsound` from
//     one lens stays visible beside another lens's `sound` until the anchor is answered — by its own
//     seat's later verdict, or by an edit that reopens the sentence.
//
// An anchor is VERIFIED when its bundle reproduced AND its own sentence is judged sound.

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

// RerunNone is the bundle state of a computation nobody re-ran — the word every reader prints for it.
const RerunNone = "not_rerun"

// bundleRerun is a bundle's standing re-run: the latest `reproduce` of its sha that RAN it.
type bundleRerun struct {
	rerun recordpb.Rerun
	ev    *Event
	drift string
}

// bundleRerunsOf is every bundle's standing re-run, keyed by bundle_sha. A `reproduce` that judged
// its anchor without running the bundle (`ran` = false) is not a re-run.
func bundleRerunsOf(evs []*Event) map[string]*bundleRerun {
	out := map[string]*bundleRerun{}
	for _, e := range evs {
		r, ok := recordpb.BodyAs[*recordpb.Reproduce](e)
		if !ok || r.GetBundleSha() == "" || !r.GetRan() {
			continue
		}
		out[r.GetBundleSha()] = &bundleRerun{rerun: r.GetRerun(), ev: e, drift: r.GetDrift()}
	}
	return out
}

// validateReproduce is the write's check of a reproduce: the anchor names a computation, the sha is
// that computation's, the act says what it compared, a judgement made without running the bundle
// stands on a re-run the record holds and states that re-run's comparison, and a bundle no checker
// can run is never judged sound.
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
	if !b.GetRan() {
		// A JUDGEMENT THAT RAN NOTHING STANDS ON THE BUNDLE'S RE-RUN, and states what that re-run
		// showed.
		var standing sql.NullString
		ran, err := queryRowAt(q, []any{&standing}, `SELECT "rerun" FROM "reproduce" WHERE "bundle_sha" = ? AND "ran" = 1 ORDER BY "event_id" DESC LIMIT 1`, sha)
		if err != nil {
			return err
		}
		if !ran {
			return fmt.Errorf("reproduce of %s judged its sentence without running its bundle, and nothing on the record ever ran that bundle — a judgement stands on a re-run, so this one re-runs it", b.GetComputationId())
		}
		if standing.String != recordpb.Word(b.GetRerun()) {
			return fmt.Errorf("reproduce of %s ran nothing and records %s, but the bundle's standing re-run is %s — an act that ran nothing states the comparison of the re-run it stands on", b.GetComputationId(), recordpb.Word(b.GetRerun()), standing.String)
		}
	}
	if b.GetRerun() == recordpb.Rerun_RERUN_CALLS_THE_TOOL {
		return refuseSoundOverTheTool(b)
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
	// rerun is the bundle's standing re-run, nil when nobody re-ran the bundle.
	rerun *bundleRerun
	// soundness is one of the Soundness* words.
	soundness string
	// judgements are the anchor's standing soundness verdicts, one per seat, in record order.
	judgements []anchorCheck
}

// reproduced is whether the bundle's standing re-run wrote the recorded output.
func (v computationVerdict) reproduced() bool {
	return v.rerun != nil && v.rerun.rerun == recordpb.Rerun_RERUN_REPRODUCED
}

// verified is the ruling's conjunction: the bundle reproduced and this sentence is judged sound.
func (v computationVerdict) verified() bool { return v.reproduced() && v.soundness == SoundnessSound }

// rerunWord is the bundle state as every reader prints it.
func (v computationVerdict) rerunWord() string {
	if v.rerun == nil {
		return RerunNone
	}
	return recordpb.Word(v.rerun.rerun)
}

// computationVerdictOf is the anchor's standing verdict, read from the ledger and the bundle re-runs.
func computationVerdictOf(s *anchorStanding, reruns map[string]*bundleRerun) computationVerdict {
	v := computationVerdict{soundness: SoundnessNotJudged}
	if s == nil {
		return v
	}
	v.rerun = reruns[s.sha]
	// A RETIRED ANCHOR KEEPS THE VERDICTS IT LEFT WITH. Nothing is owed on a sentence the report no
	// longer carries, which is what `stands` answers for the work list; what its sentence was judged
	// to be is still the record's to show.
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
