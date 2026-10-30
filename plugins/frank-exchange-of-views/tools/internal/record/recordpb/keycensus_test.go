package recordpb

import (
	"bufio"
	"os"
	"sort"
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// EVERY PRE-MIGRATION PAYLOAD KEY IS ACCOUNTED FOR — §IV.3's mitigation, and it can only be built
// NOW, while `Payload` still exists to be censused. After PR2's bulk edit the evidence is gone.
//
// The risk it answers: 102 keys and ~1,000 call sites. A key that exists on NO message becomes a
// compile error, which is the migration working. A key silently mapped to the WRONG field does
// not, and a key quietly forgotten does not either — both look exactly like success.
//
// testdata/payload-keys.txt is the frozen census, extracted from the pre-change tree.
func loadKeyCensus(t *testing.T) []string {
	t.Helper()
	f, err := os.Open("testdata/payload-keys.txt")
	if err != nil {
		t.Fatalf("cannot read the frozen key census: %v — it is the only evidence of what the "+
			"pre-migration record carried, and it cannot be re-derived after Payload is deleted", err)
	}
	defer f.Close()
	var keys []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if k := strings.TrimSpace(sc.Text()); k != "" {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		t.Fatal("the key census is empty — a no-match must fail, not read as full coverage")
	}
	return keys
}

// notFields are census keys that deliberately became NO proto field. Each needs a reason: an
// unexplained absence is indistinguishable from one nobody noticed, which is the whole defect
// class this migration exists to remove.
var notFields = map[string]string{
	"reproduced": "REPLACED at epoch 28 by Reproduce.rerun: a bool read false both where the script printed something else and where nothing re-ran it; migrate's reproduce entry translates true to reproduced and false to differs",
	"round":      "RETIRED (plans/roundless.md §III.A.2): the envelope stamped a derived count and the record now answers it — events_w.\"epoch\" is the chair-sitting count at the row, events_w.\"sitting\" the seat's own; a migrated record carries neither and every reader derives them",
	// LEGACY VOCABULARY — read only by the retired-type arms §II.6 deletes in PR3.
	"as":       "the legacy landing key for a ruling; available.go:196 reads it beside `verdict` as a pre-collapse fallback. MotionRule.ruling replaces it",
	"evidence": "the legacy `dispute` prose key, read only at view.go:519 and assemble.go:885 — both legacy arms. Motion.basis replaces it",
	"ruling":   "the legacy petition-rule/inquiry-rule key, read at view.go:567 and capture.go:983. MotionRule's per-subject ruling oneof replaces it",

	"contests_ruling": "the legacy spelling of an appeal — blue pursuing against a ruling. MotionAppeal replaces it, and it was the one legacy field with no counterpart at all (compat.go:105)",
	"petitioner":      "legacy petition vocabulary; the filer is the seat id on the ENVELOPE, not a payload field",
	"response":        "the legacy dispute-respond key; MotionRule's per-subject ruling oneof replaces it",
	"finding_id":      "the finding's anchor id, beside the label a gap's found_by named — two names for one finding. Finding.id is the one name, and migrate carries finding_id onto it",

	// COLLAPSED INTO A FIELD THAT ALREADY HELD THE SAME FACT. Each was a second name for
	// something the schema carried once, and the flag audit that found the duplicate words found
	// these under them.
	"deadlocked":  "one half of a two-boolean enum for HOW a non-pass run ended; every reader took the pair apart in a first-match switch. Outcome.ended replaces both",
	"exhausted":   "the other half of that pair; Outcome.ended replaces both",
	"fix_old":     "held the same span as Mint.location, checked by a second matcher — a gap's location and the span its proposal replaces were never two facts",
	"subject":     "the subject of a motion, a ruling or an appeal, beside the arm that states it; Motion.Subject and MotionRule.Subject read the arm, and the record holds no appeal",
	"none":        "SpotCheck's claim that the archive was empty; the board's count answers it, and SpotCheck.reason says why a spot-check named nothing",
	"notes":       "SpotCheck's prose when a sample WAS taken, beside `reason` for when none was; one channel split by which branch wrote it. SpotCheck.reason replaces both",
	"reference":   "a free-text third spelling of a source on the one verb that reads one red found itself. Verify.url + Verify.title replace it, as Cite already named sources",
	"independent": "a verification's mark that red found the source itself, beside the empty anchor that says the same; `corroborate` is the verb, and every reader splits on the anchor",
	"estopped_by": "the prescribing gap of an estoppel refusal, beside a count that reads only Log.type; the entry's text names the gap",
	"rationale":   "the bench's own prose channel on the retired `opinion` body. The disposition is a docket motion's ruling now, and every subject's ruling carries its ruler's argument on MotionRule.opinion (flag `--reason`) — so the fact survives on a field that already held it for the other three subjects, rather than on a fifth name for one concept",

	// A GAP'S CLASS IS A VALUE OF THE CLOSED GapClass SET, whose own definition, neighbour and
	// distinguisher are the class's; a run neither coins one nor carries a coiner's words.
	"class_new":     "a run coining a class, as a mint flag and as its own event; the class set is closed, and migrate drops both",
	"slug":          "a coined class's name on the retired class_new event; Mint.class names a GapClass value",
	"definition":    "a coined class's definition; each GapClass value's (means) is the definition",
	"neighbor":      "a coined class's nearest class; each GapClass value's (neighbor) option names it",
	"distinguisher": "a coined class's distinguisher; each GapClass value's (distinguisher) option carries it",

	// NEVER A PAYLOAD KEY.
	"script": "a FILENAME in the store, not a recorded field — a computation's program lives on disk under computations/<bundle_sha>/dir/, and the record carries its bundle_sha and argv",
	"output": "a field of the in-process execution result; an act's output is in the store, " + actOutputPath + ", and the record carries none",

	// A PROOF IS A COMPUTATION (epoch 49). Each key is migrate's to translate, onto the field named.
	"proof_id":        "Computation.computation_id and Reproduce.computation_id — the anchor id, `K-` and the same hex",
	"proof_key":       "Computation.key",
	"proof_sha":       "Computation.bundle_sha and Reproduce.bundle_sha — the bundle's identity over its argv and files, which migrate computes from the archived script",
	"proof_basis":     "Computation.basis, the Basis enum: reproducible stays reproducible, observed and not_portable are variant with their drift as the invariant",
	"recorded_output": "the act's own output, in the store at " + actOutputPath + " — a reproduce carries no copy of it",
	"observed_output": "Reproduce.drift: a re-run that differed carries what it wrote there",
}

// actOutputPath is where the store holds one act's output.
const actOutputPath = "computations/<bundle_sha>/acts/<K-id>/output"

// AN ACT'S OUTPUT IS IN THE STORE, and the census entry for `output` says where: the record held a
// copy of it on Reproduce once, and a reason that still said so would excuse the key with a fact
// that is no longer true.
func TestTheOutputKeySaysTheStoreHoldsTheActsOutput(t *testing.T) {
	for _, k := range []string{"output", "recorded_output"} {
		if !strings.Contains(notFields[k], actOutputPath) {
			t.Errorf("notFields[%q] = %q, want it to name %s", k, notFields[k], actOutputPath)
		}
	}
}

func TestEveryPreMigrationKeyIsAccountedFor(t *testing.T) {
	fd, err := protoregistry.GlobalFiles.FindFileByPath("record.proto")
	if err != nil {
		t.Fatalf("cannot find the schema: %v", err)
	}

	// field name -> the messages carrying it, walked from the descriptor so the map cannot rot.
	owners := map[string][]string{}
	var walk func(protoreflect.MessageDescriptor, map[string]bool)
	walk = func(md protoreflect.MessageDescriptor, seen map[string]bool) {
		if seen[string(md.FullName())] {
			return
		}
		seen[string(md.FullName())] = true
		for i := 0; i < md.Fields().Len(); i++ {
			f := md.Fields().Get(i)
			owners[string(f.Name())] = append(owners[string(f.Name())], string(md.Name()))
			if f.Kind() == protoreflect.MessageKind && !f.IsMap() {
				walk(f.Message(), seen)
			}
		}
	}
	seen := map[string]bool{}
	for i := 0; i < fd.Messages().Len(); i++ {
		walk(fd.Messages().Get(i), seen)
	}

	var unaccounted []string
	for _, k := range loadKeyCensus(t) {
		if _, ok := owners[k]; ok {
			continue
		}
		if _, excused := notFields[k]; excused {
			continue
		}
		unaccounted = append(unaccounted, k)
	}
	sort.Strings(unaccounted)
	for _, k := range unaccounted {
		t.Errorf("payload key %q from the pre-migration census maps to NO proto field and has no "+
			"stated reason — either add the field, or add it to notFields with why. A key that "+
			"quietly disappears looks exactly like a migration that worked", k)
	}
}

// AND THE EXCUSES MUST NOT ROT. An entry in notFields naming a key the census never had, or a key a
// proto field already accounts for by name, is dead weight that reads as coverage: deleting it
// changes nothing any test reads — the same failure the descriptions and requiredness gates catch
// from the other direction.
func TestNoExcuseNamesAKeyTheCensusNeverHad(t *testing.T) {
	census := map[string]bool{}
	for _, k := range loadKeyCensus(t) {
		census[k] = true
	}
	fields := map[string]bool{}
	fd, err := protoregistry.GlobalFiles.FindFileByPath("record.proto")
	if err != nil {
		t.Fatalf("cannot find the schema: %v", err)
	}
	for i := 0; i < fd.Messages().Len(); i++ {
		md := fd.Messages().Get(i)
		for j := 0; j < md.Fields().Len(); j++ {
			fields[string(md.Fields().Get(j).Name())] = true
		}
	}
	var stale []string
	for k := range notFields {
		if fields[k] {
			t.Errorf("notFields excuses %q, which a field of that name already accounts for — delete it", k)
		}
		if !census[k] {
			stale = append(stale, k)
		}
	}
	sort.Strings(stale)
	for _, k := range stale {
		t.Errorf("notFields excuses %q, which is not in the frozen census — delete it", k)
	}
}
