package migrate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// remap is the STATEFUL half of the translation (plans/roundless.md §III.A.5). The registry
// translates one event at a time and cannot know what the run has already minted; this can, and
// it is what turns an archived run's identifiers into the ones a live reader speaks:
//
//   - a seat id loses its epoch and a numbered lens becomes its area: `red-merge-r2` is
//     `red-chair` (its second register is its second sitting, which opens epoch 2);
//     `red-lens-r1-L5` is `red-lens-logic`. Two citation instances in one epoch (`-L1`, `-L2`)
//     were two sittings of the evidence lens, and that is what they become.
//   - every id takes the one shape, <LETTER>-<8 hex> (plans/markers-one-mechanism.md §III.6): an
//     id that held eight hex keeps them under its kind's letter, and every other — a sequence
//     number, a round-scoped gap id, a synthesized motion id — takes the first eight hex of
//     sha256(source hash ‖ archived id). A finding's archived label maps to its finding's id.
//
// Nothing here reads the record back: an id is a function of the source alone, so two migrations
// of one source produce the same ids wherever they are written (the determinism test holds that),
// and Replay fills the table from the whole stream before it writes any of it.
type remap struct {
	sourceHash string
	ids        map[string]string // archived id -> the id the migrated record carries
	words      *regexp.Regexp    // every archived id in ids, at word boundaries
	wordsOf    int               // len(ids) when words was built
}

func newRemap(sourceHash string) *remap {
	return &remap{sourceHash: sourceHash, ids: map[string]string{}}
}

var (
	numberedLens = regexp.MustCompile(`^red-lens-r\d+-L(\d+)$`)
	namedLens    = regexp.MustCompile(`^red-lens-r\d+-([a-z]+(?:-[a-z]+)*)$`)
	roundedSeat  = regexp.MustCompile(`^(red-merge|red-chair|blue-respond|judge)-r\d+$`)
	// currentID is an id already in the one shape, which a source written at this epoch carries.
	currentID = regexp.MustCompile(`^` + anchor.IDPattern("finding", "citation", "proof", "gap", "avenue", "motion") + `$`)
	// hexID is an archived id whose eight hex are kept.
	hexID = regexp.MustCompile(`^[fcp]-([0-9a-f]{8})$`)
	// archivedToken is an anchor token as an archived record spells it, its tag and its id. A
	// comment that only looks like one — a seat's prose showing the form with its id elided — is
	// not a token and is left as written.
	archivedToken = regexp.MustCompile(`<!--(fx|cite|proof|gap):([fcp]-[0-9a-f]+|G[0-9]+)-->`)
	wordEnded     = regexp.MustCompile(`^\w(.*\w)?$`)
	tagKind       = map[string]string{"fx": "finding", "cite": "citation", "proof": "proof", "gap": "gap"}
)

// The three field lists (R-8, R-19) class every string field of every message an event body
// reaches; TestEveryStringFieldIsInOneList fails a field in none or in two, so a field added to
// record.proto is classed before it migrates. In EVERY field an exact anchor token is respelled.
// Beyond that:
//
//   - idFields hold ids: a whole value, or list element, equal to an archived id is replaced.
//   - argumentFields hold a seat's own wording, where an id is a reference to the record: an
//     archived id at word boundaries is replaced.
//   - dataFields hold report text, subject sources, keys, hashes, dates, enumerated values and
//     identities, and change in no other way: a rewritten Proof.script breaks its sha, and
//     "Q2 earnings" in a source's title is the subject's.
//
// mints names the id fields whose value brings an id into being, and the kind it is an id of.
var (
	idFields = fieldSet(`Mint.gap_id Mint.about_ref Mint.supersedes Mint.found_by Mint.distinct_from
		Close.gap_id Close.successor Closing.gap_id Regrade.gap_id SpotCheck.ids Finding.id
		Finding.about_ref Observe.label Anchor.id Cite.label Verify.anchor Verify.label Proof.proof_id
		Proof.answers Proof.cites Avenue.avenue_id BlueEdit.answers BlueEdit.reopened Retire.anchors
		ManifestRow.gap_id Log.estopped_by Motion.motion_id GradeMotion.gap_id DocketMotion.gap_id
		AvenueMotion.avenue_id MotionRule.motion_id MotionAppeal.motion_id Dispatch.gap_ids
		Gate.migration_admitted_gap_ids`)
	argumentFields = fieldSet(`Mint.problem Mint.required_fix Mint.acceptance_check Mint.definition
		Mint.neighbor Mint.distinguisher Mint.mint_reason ClassNew.definition ClassNew.neighbor
		ClassNew.distinguisher Close.prose Close.anchor_tool Close.anchor_target Closing.text
		Regrade.basis SpotCheck.reason Finding.text Observe.text Observe.observation Cite.text
		Verify.text Proof.text Proof.drift Reproduce.note Avenue.line Avenue.hypothesis Avenue.method
		Avenue.reason AvenueReview.reason BlueEdit.text Revision.text Retire.reason ManifestRow.row
		Log.text Motion.basis Motion.relief DocketRuling.principle DocketRuling.tension
		DocketRuling.review_flag DocketRuling.settled DocketRuling.reopens_on MotionRule.opinion
		MotionAppeal.reason Outcome.prose Outcome.verdict_why Position.text Halt.opinion
		Certify.statement Declare.holding Correction.why`)
	dataFields = fieldSet(`Anchor.location Avenue.supersedes_status BaseIngest.text BlueEdit.edit_key
		BlueEdit.new BlueEdit.old Cast.lane_seat_ids Cast.seat_ids Cite.access_date Cite.cite_key
		Cite.location Cite.ocr_engine Cite.ocr_quote Cite.ocr_text_sha Cite.sha256 Cite.title Cite.url
		ClassNew.slug Close.anchor_seat Close.carried_from Correction.corrects Correction.replacement
		Dispatch.seat_id Finding.finding_key Finding.location Mint.class Mint.fix_basis Mint.fix_new
		Mint.location Mint.mint_key Outcome.verdict_basis Proof.location Proof.proof_basis
		Proof.proof_key Proof.proof_sha Proof.script Register.agent_id Register.agent_type
		Register.hook_version Register.repairs_sitting Register.run_via Register.tool_version
		Reproduce.observed_output Reproduce.proof_sha Reproduce.recorded_output Retire.claim
		Retire.removal_basis Retire.superseded_by SittingClose.agent_id
		SittingClose.agent_transcript_path SittingClose.agent_type SittingClose.prompt_id
		SittingClose.session_id SittingLimit.agent_id SittingLimit.agent_type SittingLimit.seat_id
		SittingOpen.agent_id SittingOpen.agent_type SittingOpen.prompt_id SittingOpen.seat_id
		SittingOpen.session_id SpotCheck.areas Verify.access_date Verify.claim Verify.page_render_sha
		Verify.reading_render_sha Verify.title Verify.url`)
	mints = map[string]string{"Mint.gap_id": "gap", "Finding.id": "finding", "Cite.label": "citation",
		"Verify.label": "citation", "Proof.proof_id": "proof", "Avenue.avenue_id": "avenue", "Motion.motion_id": "motion"}
)

func fieldSet(names string) map[string]bool {
	set := map[string]bool{}
	for _, n := range strings.Fields(names) {
		set[n] = true
	}
	return set
}

// lensAreaOfNumber is the roster of the runs in run-archive/ (2026-08-22 .. 2026-09-02, the W2i
// era): roles 1-4 were instances of the citation lens, 5 logic, 6 dark-side, 7 report voice —
// debate.js at 673b7b2b^ line 858-860. A number outside it is a refusal, not a guess.
func lensAreaOfNumber(n int) (string, bool) {
	switch {
	case n >= 1 && n <= 4:
		return "evidence", true
	case n == 5:
		return "logic", true
	case n == 6:
		return "dark-side", true
	case n == 7:
		return "voice", true
	}
	return "", false
}

// seat translates an archived seat id to its roundless form; an id already in that form passes.
func (r *remap) seat(old string) (string, error) {
	if m := numberedLens.FindStringSubmatch(old); m != nil {
		n, _ := strconv.Atoi(m[1])
		area, ok := lensAreaOfNumber(n)
		if !ok {
			return "", fmt.Errorf("migrate: seat %q names lens number %d, which no archived roster had — the number-to-area table in remap.go stops at 7", old, n)
		}
		return "red-lens-" + area, nil
	}
	if m := namedLens.FindStringSubmatch(old); m != nil {
		return "red-lens-" + m[1], nil
	}
	if m := roundedSeat.FindStringSubmatch(old); m != nil {
		if m[1] == "red-merge" {
			return "red-chair", nil
		}
		return m[1], nil
	}
	// THE BENCH WAS FOUR SEAT IDS AND IS ONE. `judge-terminal`, `assemble` and
	// `judge-petition-<petitioner>` had identical surfaces, tier and role, and no refusal turned
	// on which one a seat claimed to be — what differed was the question the engine asked, which
	// now travels on the dispatch rather than in the identity. An archived run carries the old
	// ids on its registers and on every act attributed to them, so they are brought forward here:
	// a record that cannot be read is a record that is gone, and the roster refuses these now.
	//
	// The petitioner is DROPPED rather than preserved. It said WHO FILED, and who filed is on the
	// petition the sitting ruled — recovering it from a seat id was the string-shaped copy.
	if strings.HasPrefix(old, "judge-petition-") || old == "judge-terminal" || old == "assemble" {
		return "judge", nil
	}
	return old, nil
}

// id is the id of a kind the migrated record carries for an archived one, remembered so every
// later reference to the archived id says the new name. An id already in the one shape stands.
func (r *remap) id(kind, old string) string {
	if old == "" || currentID.MatchString(old) {
		return old
	}
	sum := sha256.Sum256([]byte(r.sourceHash + old))
	b := [4]byte(sum[:4])
	if m := hexID.FindStringSubmatch(old); m != nil {
		hex.Decode(b[:], []byte(m[1]))
	}
	r.ids[old] = anchor.ID(kind, b)
	return r.ids[old]
}

// apply respells one translated body in place, by the three field lists. archived is the label its
// archived event carries, which on a finding names the finding's id from here on; a finding
// the archive gave no id takes one from its label.
func (r *remap) apply(body proto.Message, archived any) {
	label, _ := archived.(string)
	if f, ok := body.(*recordpb.Finding); ok && label != "" {
		if f.Id == nil {
			f.Id = &label
		}
		r.ids[label] = r.id("finding", f.GetId())
	}
	r.walk(body.ProtoReflect())
}

// walk respells a message's set fields in schema order, never map order: an id a field mints is
// remembered before the fields after it are read, in every migration alike.
func (r *remap) walk(m protoreflect.Message) {
	for fields, i := m.Descriptor().Fields(), 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		if !m.Has(fd) {
			continue
		}
		name, v := string(m.Descriptor().Name())+"."+string(fd.Name()), m.Get(fd)
		switch {
		case fd.Kind() == protoreflect.MessageKind && fd.IsList():
			for l, j := v.List(), 0; j < l.Len(); j++ {
				r.walk(l.Get(j).Message())
			}
		case fd.Kind() == protoreflect.MessageKind && !fd.IsMap():
			r.walk(v.Message())
		case fd.Kind() == protoreflect.StringKind && fd.IsList():
			for l, j := v.List(), 0; j < l.Len(); j++ {
				l.Set(j, protoreflect.ValueOfString(r.respell(name, l.Get(j).String())))
			}
		case fd.Kind() == protoreflect.StringKind:
			m.Set(fd, protoreflect.ValueOfString(r.respell(name, v.String())))
		}
	}
}

// respell is the value field holds in the migrated record: its anchor tokens respelled, then its
// ids by the field's list. A reference to an id the run never minted is left as it was: the write
// path decides what a dangling reference is worth, and it says so.
func (r *remap) respell(field, v string) string {
	v = archivedToken.ReplaceAllStringFunc(v, func(tok string) string {
		m := archivedToken.FindStringSubmatch(tok)
		return anchor.Token(r.id(tagKind[m[1]], m[2]))
	})
	switch {
	case mints[field] != "":
		return r.id(mints[field], v)
	case idFields[field]:
		if nv, ok := r.ids[v]; ok {
			return nv
		}
	case argumentFields[field] && len(r.ids) > 0:
		if r.wordsOf != len(r.ids) {
			olds := make([]string, 0, len(r.ids))
			for old := range r.ids {
				if wordEnded.MatchString(old) { // only what a word boundary can bound on both sides
					olds = append(olds, regexp.QuoteMeta(old))
				}
			}
			sort.Slice(olds, func(i, j int) bool {
				return len(olds[i]) > len(olds[j]) || len(olds[i]) == len(olds[j]) && olds[i] < olds[j]
			})
			r.words, r.wordsOf = regexp.MustCompile(`\b(?:`+strings.Join(olds, "|")+`)\b`), len(r.ids)
		}
		return r.words.ReplaceAllStringFunc(v, func(old string) string { return r.ids[old] })
	}
	return v
}
