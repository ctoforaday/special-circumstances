package migrate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
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
//     sha256(source hash ‖ archived id).
//
// Nothing here reads the record back: an id is a function of the source alone, so two migrations
// of one source produce the same ids wherever they are written (the determinism test holds that),
// and Replay fills the table from the whole stream before it writes any of it.
type remap struct {
	sourceHash string
	ids        map[string]string // archived id -> the id the migrated record carries
	words      *regexp.Regexp    // every archived id in ids, at word boundaries
	wordsOf    int               // len(ids) when words was built
	// prose is every archived id rewritten inside a seat's wording, in the order rewritten. A
	// word in a seat's prose that equals an archived id — "Q1 revenue", "the G2 summit" — is
	// rewritten as a reference; nothing can tell the two apart, so each rewrite is kept where a
	// reader can judge it (the manifest's per-field count, and the archive's golden list).
	prose []migration.ProseRewrite
	// archivedNotes is whether the source predates the computation (epoch 49), whose seats' wording
	// names a computation's note `[^P<n>]` (notes).
	archivedNotes bool
	// dropped is every citation label an archived reading of a source its lens found minted, in
	// both the archived and the migrated spelling (dropLabels), and every name an archived finding
	// carries (dropFindings). No reading places a citation and no finding is written, so each one's
	// token is removed from every field and its id from every id list; in a seat's wording a bare
	// label or id is left as written.
	dropped map[string]bool
	// rekeyed is every successor id Translation 7 wrote as an amend, to the id of the chain's first
	// gap, both as the migrated record spells them (rekey). Every id field and token naming the
	// successor after its mint names the first gap.
	rekeyed map[string]string
	firsts  map[string]bool // the first gaps rekeyed names
}

// rekey records that the migrated id succ names the gap first from here on (Translation 7).
func (r *remap) rekey(succ, first string) {
	if r.rekeyed == nil {
		r.rekeyed, r.firsts = map[string]string{}, map[string]bool{}
	}
	r.rekeyed[succ], r.firsts[first] = first, true
}

// current is the id the migrated record names gap id by from here on: the first gap of the chain a
// re-keyed successor joined, else id.
func (r *remap) current(id string) string {
	if first, ok := r.rekeyed[id]; ok {
		return first
	}
	return id
}

// rekeyedTokens is v with every token naming a re-keyed successor naming its first gap, and a token
// that then repeats the token before it, with only whitespace between, taken out: a successor's
// marker beside its first gap's in one sentence is one marker.
func (r *remap) rekeyedTokens(v string) string {
	if len(r.rekeyed) == 0 {
		return v
	}
	v = anchor.Replace(v, func(tok, id string) string {
		if first, ok := r.rekeyed[id]; ok {
			return anchor.Token(first)
		}
		return tok
	})
	var b strings.Builder
	last, prevEnd, prevID := 0, -1, ""
	anchor.Each(v, func(start, end int, id string) {
		if id == prevID && r.firsts[id] && strings.TrimSpace(v[prevEnd:start]) == "" {
			b.WriteString(v[last:start])
			last = end
		}
		prevEnd, prevID = end, id
	})
	if last == 0 {
		return v
	}
	b.WriteString(v[last:])
	return b.String()
}

// around is v from proseContext bytes before [from, to) to proseContext bytes after, widened to
// whole characters.
func around(v string, from, to int) string {
	a, z := max(0, from-proseContext), min(len(v), to+proseContext)
	for a > 0 && !utf8.RuneStart(v[a]) {
		a--
	}
	for z < len(v) && !utf8.RuneStart(v[z]) {
		z++
	}
	return v[a:z]
}

// proseContext is how many bytes of the source text either side of a rewritten id the context keeps.
const proseContext = 40

func newRemap(sourceHash string) *remap {
	return &remap{sourceHash: sourceHash, ids: map[string]string{}}
}

var (
	numberedLens = regexp.MustCompile(`^red-lens-r\d+-L(\d+)$`)
	namedLens    = regexp.MustCompile(`^red-lens-r\d+-([a-z]+(?:-[a-z]+)*)$`)
	roundedSeat  = regexp.MustCompile(`^(red-merge|red-chair|blue-respond|judge)-r\d+$`)
	// currentID is an id already in the one shape, which a source written at this epoch carries.
	currentID = regexp.MustCompile(`^(?:` + anchor.IDPattern("citation", "computation", "docketing", "gap", "inquiry", "motion", "position", "question part") + `|` + findingID + `)$`)
	// hexID is an archived id whose eight hex are kept: an era's lower-case id, and a proof's `P-`
	// id, which a source before epoch 49 carries and which is a computation's `K-` id with the same
	// hex.
	hexID = regexp.MustCompile(`^(?:[fcp]|P)-([0-9a-f]{8})$`)
	// archivedToken is an anchor token as an archived record spells it, its tag and its id. A
	// comment that only looks like one — a seat's prose showing the form with its id elided — is
	// not a token and is left as written.
	archivedToken = regexp.MustCompile(`<!--(fx|cite|proof|gap):([fcp]-[0-9a-f]+|P-[0-9a-f]{8}|G[0-9]+|` + findingID + `)-->`)
	wordEnded     = regexp.MustCompile(`^\w(.*\w)?$`)
	tagKind       = map[string]string{"cite": "citation", "proof": "computation", "gap": "gap"}
	// archivedNote is a computation's woven footnote label as a source before epoch 49 spelled it,
	// `[^P3]`, in a seat's wording: the report numbers a computation's note `[^K3]`.
	archivedNote = regexp.MustCompile(`\[\^P([0-9]+)\]`)
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
//     identities, and change in no other way: a rewritten Computation.argv breaks its sha, and
//     "Q2 earnings" in a source's title is the subject's.
//
// mints names the id fields whose value brings an id into being, and the kind it is an id of.
var (
	idFields = fieldSet(`Mint.gap_id Mint.about_ref
		Close.gap_id Closing.gap_id Regrade.gap_id Amend.gap_id Amend.answers SpotCheck.ids
		Anchor.id Anchor.sentence_of Cite.label Verify.anchor Computation.computation_id
		Computation.answers Computation.cites Reproduce.computation_id Inquiry.inquiry_id BlueEdit.answers BlueEdit.reopened BlueEdit.removed_anchors Retire.anchors Retire.answers Cite.answers
		Motion.motion_id DocketMotion.gap_id
		InquiryCheck.inquiry_id MotionRule.motion_id Dispatch.gap_ids
		Declare.gap_ids Close.anchors Stale.gap_id Tombstone.gap_id Tombstone.docketing_id
		Docketing.docketing_id Docketing.gap_id Docketing.motion_id DocketRuling.docketing_id
		Position.position_id PositionMotion.position_id PositionHearing.gap_ids
		QuestionPart.part_id QuestionPartCheck.part_id
		Regrade.accepts Regrade.answers Close.accepts Inquiry.answers QuestionPart.answers`)
	argumentFields = fieldSet(`Mint.problem Mint.required_fix Mint.acceptance_check
		Mint.mint_reason Close.prose Close.anchor_tool Close.anchor_target Closing.text
		Regrade.basis Amend.problem Amend.required_fix Amend.acceptance_check Amend.basis SpotCheck.reason Cite.text
		Verify.text Computation.text Computation.drift Computation.invariant Reproduce.note Inquiry.question Inquiry.hypothesis Inquiry.method
		Inquiry.reason InquiryCheck.reason BlueEdit.text Retire.reason
		Log.text Motion.basis Motion.relief Docketing.basis DocketRuling.principle DocketRuling.tension
		DocketRuling.review_flag DocketRuling.settled DocketRuling.reopens_on DocketRuling.direction PositionHearing.direction MotionRule.opinion MotionRule.relief
		Outcome.prose Outcome.verdict_why Position.text Halt.opinion
		Certify.statement Declare.holding QuestionPart.text QuestionPart.reason
		Closing.problem Closing.required_fix Closing.acceptance_check`)
	dataFields = fieldSet(`Anchor.location ReportPart.text BlueEdit.edit_key
		BlueEdit.new BlueEdit.old BlueEdit.old_paragraphs BlueEdit.new_paragraphs Cast.lane_seat_ids Cast.seat_ids Cite.cite_key Cite.copy_captured_at Cite.copy_url Cite.fetched_at
		Cite.location Cite.ocr_engine Cite.ocr_text_sha Cite.sha256 Cite.source_after Cite.source_before
		Cite.source_matched Cite.source_span Cite.span_text_sha Cite.title Cite.url EnvelopeAdmitted.agent_id
		Close.anchor_seat Close.carried_from
		Dispatch.seat_id PositionHearing.lens Plan.printed Inquiry.supersedes_status Mint.fix_basis Mint.fix_new Closing.location Closing.fix_new
		Amend.fix_new Amend.fix_basis Amend.result_problem Amend.result_required_fix Amend.result_acceptance_check
		Amend.result_fix_new Amend.result_fix_basis
		Mint.location Mint.mint_key Outcome.verdict_basis Computation.location Computation.key
		Computation.bundle_sha Computation.argv Computation.toolchain_path Computation.toolchain_sha256
		Register.agent_id Register.agent_type
		Register.hook_version Register.run_via Register.tool_version
		Reproduce.bundle_sha Reproduce.claim Reproduce.drift Reproduce.rerun_sha Retire.claim
		SittingClose.agent_id
		SittingClose.agent_transcript_path SittingOpen.agent_id
		SittingOpen.agent_type SittingOpen.seat_id SpotCheck.areas Verify.claim Verify.copy_url Verify.page_render_sha
		Verify.reading_render_sha Verify.title Verify.url QuestionPart.supersedes_coverage`)
	mints = map[string]string{"Mint.gap_id": "gap", "Cite.label": "citation",
		"Computation.computation_id": "computation", "Inquiry.inquiry_id": "inquiry", "Motion.motion_id": "motion",
		"Position.position_id": "position", "QuestionPart.part_id": "question part"}
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

// seat translates an archived seat id to its roundless form; an id already in that form passes. A
// translation that lands off the roster is refused: an archived spelling the patterns below miss
// would otherwise pass through unchanged and be attributed to no party.
func (r *remap) seat(old string) (string, error) {
	s, err := roundlessSeat(old)
	if err != nil {
		return "", err
	}
	if s != record.HarnessSeat && seatclass.RoleOfSeat(s) == "" {
		return "", fmt.Errorf("seat %q translates to %q, which is no seat on the roster — no archived spelling in remap.go's seat translation matches it", old, s)
	}
	return s, nil
}

func roundlessSeat(old string) (string, error) {
	if m := numberedLens.FindStringSubmatch(old); m != nil {
		n, _ := strconv.Atoi(m[1])
		area, ok := lensAreaOfNumber(n)
		if !ok {
			return "", fmt.Errorf("seat %q names lens number %d, which no archived roster had — the number-to-area table in remap.go stops at 7", old, n)
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
	r.ids[old] = r.spelling(kind, old)
	return r.ids[old]
}

// spelling is the id of a kind the migrated record would carry for an archived one, remembering
// nothing.
func (r *remap) spelling(kind, old string) string {
	if old == "" || currentID.MatchString(old) {
		return old
	}
	sum := sha256.Sum256([]byte(r.sourceHash + old))
	b := [4]byte(sum[:4])
	if m := hexID.FindStringSubmatch(old); m != nil {
		hex.Decode(b[:], []byte(m[1]))
	}
	return anchor.ID(kind, b)
}

// The finding, as an archived record spells it: its id in the one shape, and its token's tag. No
// live kind holds either; migrate declares them beside the archived bodies it reads.
const (
	findingID  = `F-[0-9a-f]{8}`
	findingTag = "fx"
)

// wholeFindingID is an archived finding id already in the one shape.
var wholeFindingID = regexp.MustCompile(`^` + findingID + `$`)

// dropFindings records every name an archived finding carries to drop (remap.dropped): its id, its
// archived id and its label, each as the source spells it.
func (r *remap) dropFindings(names []string) {
	if r.dropped == nil {
		r.dropped = map[string]bool{}
	}
	for _, n := range names {
		if n == "" {
			continue
		}
		r.dropped[n], r.dropped[r.findingSpelling(n)] = true, true
	}
}

// findingSpelling is the one-shape id an archived finding name is spelled as, by spelling's rule
// under the finding's letter: a reference written after the source respelled it names it so.
func (r *remap) findingSpelling(old string) string {
	if wholeFindingID.MatchString(old) {
		return old
	}
	sum := sha256.Sum256([]byte(r.sourceHash + old))
	b := [4]byte(sum[:4])
	if m := hexID.FindStringSubmatch(old); m != nil {
		hex.Decode(b[:], []byte(m[1]))
	}
	return "F-" + hex.EncodeToString(b[:])
}

// dropLabels records the citation labels to drop (remap.dropped), each in both spellings.
func (r *remap) dropLabels(labels []string) {
	if r.dropped == nil {
		r.dropped = map[string]bool{}
	}
	for _, l := range labels {
		r.dropped[l], r.dropped[r.spelling("citation", l)] = true, true
	}
}

// apply respells one translated body in place, by the three field lists.
func (r *remap) apply(body proto.Message) {
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
			// A DROPPED LABEL LEAVES AN ID LIST, its element removed rather than respelled to "": an
			// empty element would read as an id nobody minted.
			l := v.List()
			var kept []string
			for j := 0; j < l.Len(); j++ {
				if e := l.Get(j).String(); !idFields[name] || !r.dropped[e] {
					kept = append(kept, r.respell(name, e))
				}
			}
			l.Truncate(0)
			for _, e := range kept {
				l.Append(protoreflect.ValueOfString(e))
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
	if argumentFields[field] && r.archivedNotes {
		v = r.notes(field, v)
	}
	return r.respellIDs(field, v)
}

// notes respells each archived computation footnote label in a seat's wording, kept beside the id
// rewrites a reader can judge.
func (r *remap) notes(field, v string) string {
	var b strings.Builder
	last := 0
	for _, loc := range archivedNote.FindAllStringSubmatchIndex(v, -1) {
		old, now := v[loc[0]:loc[1]], "[^K"+v[loc[2]:loc[3]]+"]"
		b.WriteString(v[last:loc[0]])
		b.WriteString(now)
		last = loc[1]
		r.prose = append(r.prose, migration.ProseRewrite{Field: field, Old: old, New: now, Context: around(v, loc[0], loc[1])})
	}
	b.WriteString(v[last:])
	return b.String()
}

// respellIDs is respell's tokens and ids. A dropped label's token is removed from every field
// whatever its list. Only its archived spelling is read: the archives whose labels are in the one
// shape (m27, m29, m35) hold no token of one in any source field (their citations reach a render
// through the label's Anchor row, which replay writes as no event), and
// TestNoMigratedArchiveHoldsALensPlacedToken fails on a migrated field that holds one.
func (r *remap) respellIDs(field, v string) string {
	v = archivedToken.ReplaceAllStringFunc(v, func(tok string) string {
		m := archivedToken.FindStringSubmatch(tok)
		if m[1] == findingTag || r.dropped[m[2]] {
			return ""
		}
		return anchor.Token(r.id(tagKind[m[1]], m[2]))
	})
	v = r.rekeyedTokens(v)
	switch {
	case mints[field] != "":
		return r.id(mints[field], v)
	case idFields[field]:
		if nv, ok := r.ids[v]; ok {
			v = nv
		}
		if first, ok := r.rekeyed[v]; ok {
			return first
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
		var b strings.Builder
		last := 0
		for _, loc := range r.words.FindAllStringIndex(v, -1) {
			old := v[loc[0]:loc[1]]
			b.WriteString(v[last:loc[0]])
			b.WriteString(r.ids[old])
			last = loc[1]
			r.prose = append(r.prose, migration.ProseRewrite{Field: field, Old: old, New: r.ids[old],
				Context: around(v, loc[0], loc[1])})
		}
		b.WriteString(v[last:])
		return b.String()
	}
	return v
}

// respelled is the id the migrated record carries for an archived id the stream minted, else the id
// as given.
func (r *remap) respelled(old string) string {
	if nv, ok := r.ids[old]; ok {
		return nv
	}
	return old
}
