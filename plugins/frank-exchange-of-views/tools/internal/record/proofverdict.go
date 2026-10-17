package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// A PROOF'S VERDICT HAS TWO OWNERS (gblock, 2026-10-12): "the anchor is about the quote, the proof
// is about the proof itself".
//
//   - REPRODUCED belongs to the SCRIPT. One matching re-run marks every anchor on that script
//     reproduced: the bytes that ran are the same bytes whichever sentence they back. The script's
//     standing re-run is the latest `reproduce` of its sha that ran it.
//   - SOUND belongs to the ANCHOR: does the script and its result support what THAT sentence says.
//     It is read off the anchor's standing checks (anchorchecks.go), so a standing `unsound` from
//     one lens stays visible beside another lens's `sound` until the anchor is answered — by its own
//     seat's later verdict, or by an edit that reopens the sentence.
//
// An anchor is VERIFIED when its script reproduced AND its own sentence is judged sound.

// The standing state of a proof anchor's sentence, as every reader spells it.
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

// RerunNone is the script state of a proof nobody re-ran — the word every reader prints for it.
const RerunNone = "not_rerun"

// scriptRerun is a script's standing re-run: the latest `reproduce` of its sha that RAN it.
type scriptRerun struct {
	rerun              recordpb.Rerun
	ev                 *Event
	recorded, observed string
}

// scriptRerunsOf is every script's standing re-run, keyed by proof_sha. A `reproduce` that judged
// its anchor without running the script (RERUN_NOT_RUN) is not a re-run.
func scriptRerunsOf(evs []*Event) map[string]*scriptRerun {
	out := map[string]*scriptRerun{}
	for _, e := range evs {
		r, ok := recordpb.BodyAs[*recordpb.Reproduce](e)
		if !ok || r.GetProofSha() == "" || !ranTheScript(r.GetRerun()) {
			continue
		}
		out[r.GetProofSha()] = &scriptRerun{rerun: r.GetRerun(), ev: e, recorded: r.GetRecordedOutput(), observed: r.GetObservedOutput()}
	}
	return out
}

// ranTheScript is whether a reproduce of this rerun value executed the script.
func ranTheScript(r recordpb.Rerun) bool {
	return r != recordpb.Rerun_RERUN_NOT_RUN && r != recordpb.Rerun_RERUN_UNSPECIFIED
}

// validateReproduce is the write's check of a reproduce: the anchor names a proof, the sha is that
// proof's, the act says what it did with the script, a judgement made without running the script
// stands on a re-run the record holds, and a script no checker can run is never judged sound.
func validateReproduce(q recordsql.Querier, b *recordpb.Reproduce) error {
	sha, err := requireProofAnchor(q, b.GetProofId(), "--anchor")
	if err != nil {
		return err
	}
	if b.GetProofSha() != sha {
		return fmt.Errorf("reproduce names proof %s and carries the sha %q, which is not that proof's script (%s) — the sha is the anchor's, derived by the tool, never stated", b.GetProofId(), b.GetProofSha(), sha)
	}
	switch b.GetRerun() {
	case recordpb.Rerun_RERUN_UNSPECIFIED:
		return fmt.Errorf("reproduce of %s records nothing about its script: a reproduce says whether it re-ran the script and what that printed, or that it judged the sentence on the script's standing re-run", b.GetProofId())
	case recordpb.Rerun_RERUN_NOT_RUN:
		var ran int
		if _, err := queryRowAt(q, []any{&ran}, `SELECT COUNT(*) FROM "reproduce" WHERE "proof_sha" = ? AND "rerun" != ?`,
			sha, recordpb.Word(recordpb.Rerun_RERUN_NOT_RUN)); err != nil {
			return err
		}
		if ran == 0 {
			return fmt.Errorf("reproduce of %s judged its sentence without running its script, and nothing on the record ever ran that script — a judgement stands on a re-run, so this one re-runs it", b.GetProofId())
		}
	case recordpb.Rerun_RERUN_CALLS_THE_TOOL:
		if b.GetSoundness() == recordpb.Soundness_SOUNDNESS_SOUND {
			return fmt.Errorf("proof %s's script calls feov-record, which refuses every call made inside a re-run, so no checker can run it to its output — that is `--as unsound`, never sound", b.GetProofId())
		}
	}
	return nil
}

// proofVerdict is one proof anchor's standing verdict.
type proofVerdict struct {
	// rerun is the script's standing re-run, nil when nobody re-ran the script.
	rerun *scriptRerun
	// soundness is one of the Soundness* words.
	soundness string
	// judgements are the anchor's standing soundness verdicts, one per seat, in record order.
	judgements []anchorCheck
}

// reproduced is whether the script's standing re-run printed the recorded output.
func (v proofVerdict) reproduced() bool {
	return v.rerun != nil && v.rerun.rerun == recordpb.Rerun_RERUN_REPRODUCED
}

// verified is the ruling's conjunction: the script reproduced and this sentence is judged sound.
func (v proofVerdict) verified() bool { return v.reproduced() && v.soundness == SoundnessSound }

// rerunWord is the script state as every reader prints it.
func (v proofVerdict) rerunWord() string {
	if v.rerun == nil {
		return RerunNone
	}
	return recordpb.Word(v.rerun.rerun)
}

// proofVerdictOf is the anchor's standing verdict, read from the ledger and the script re-runs.
func proofVerdictOf(s *anchorStanding, reruns map[string]*scriptRerun) proofVerdict {
	v := proofVerdict{soundness: SoundnessNotJudged}
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
	case s.reopened():
		v.soundness = SoundnessReopened
	}
	return v
}
