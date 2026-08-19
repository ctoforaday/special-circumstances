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
	// LEGACY VOCABULARY — read only by the retired-type arms §II.6 deletes in PR3.
	"as":       "the legacy landing key for a ruling; available.go:196 reads it beside `verdict` as a pre-collapse fallback. MotionRule.ruling replaces it",
	"evidence": "the legacy `dispute` prose key, read only at view.go:519 and assemble.go:885 — both legacy arms. Motion.basis replaces it",
	"ruling":   "the legacy petition-rule/avenue-rule key, read at view.go:567 and capture.go:983. MotionRule's per-subject ruling oneof replaces it",

	"contests_ruling": "the legacy spelling of an appeal — blue pursuing against a ruling. MotionAppeal replaces it, and it was the one legacy field with no counterpart at all (compat.go:105)",
	"petitioner":      "legacy petition vocabulary; the filer is the seat id on the ENVELOPE, not a payload field",
	"response":        "the legacy dispute-respond key; MotionRule's per-subject ruling oneof replaces it",

	// COLLAPSED INTO A FIELD THAT ALREADY HELD THE SAME FACT. Each was a second name for
	// something the schema carried once, and the flag audit that found the duplicate words found
	// these under them.
	"deadlocked": "one half of a two-boolean enum for HOW a non-pass run ended; every reader took the pair apart in a first-match switch. Outcome.ended replaces both",
	"exhausted":  "the other half of that pair; Outcome.ended replaces both",
	"fix_old":    "held the same span as Mint.location, checked by a second matcher — a gap's location and the span its proposal replaces were never two facts",
	"notes":      "SpotCheck's prose when a sample WAS taken, beside `reason` for when none was; one channel split by which branch wrote it. SpotCheck.reason replaces both",
	"reference":  "a free-text third spelling of a source on the one verb that reads one red found itself. Verify.url + Verify.title replace it, as Cite already named sources",
	"claim":      "on a CITE it re-quoted the sentence `location` already held to place the anchor. Verify.claim is a different field and still exists",

	// NEVER A PAYLOAD KEY.
	"script": "a FILENAME in the proof cache (proof.go:191,225), not a recorded field — the proof's script lives on disk and the record carries its sha256",
	"exit":   "a field of proof.Result (proof.go:64), the in-process execution struct — never written to a shard",
	"output": "a field of proof.Result (proof.go:63); the record carries recorded_output/observed_output on Reproduce instead",
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

// AND THE EXCUSES MUST NOT ROT. An entry in notFields naming a key the census never had is dead
// weight that reads as coverage — the same failure the descriptions and requiredness gates catch
// from the other direction.
func TestNoExcuseNamesAKeyTheCensusNeverHad(t *testing.T) {
	census := map[string]bool{}
	for _, k := range loadKeyCensus(t) {
		census[k] = true
	}
	var stale []string
	for k := range notFields {
		if !census[k] {
			stale = append(stale, k)
		}
	}
	sort.Strings(stale)
	for _, k := range stale {
		t.Errorf("notFields excuses %q, which is not in the frozen census — delete it", k)
	}
}
