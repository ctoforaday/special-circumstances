package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A mint's problem prints in the report's risk matrix, so a seat id in it is refused at the write,
// naming the term and the field — and nothing reaches the board.
func TestMintRefusesSeatIdInProblem(t *testing.T) {
	run, lens := liveRegistryRun(t, map[string]recordpb.ClassMaterial{"g": recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE})
	m := liveMint("G1", "g", recordpb.Grade_GRADE_MEDIUM)
	m.Problem = proto.String("blue-respond overstates the bound on the error term.")
	_, err := Append(lens, m)
	if err == nil {
		t.Fatal("a problem naming a seat id landed on the board; it prints in the report")
	}
	for _, want := range []string{"--problem", `"blue-respond"`, "risk matrix", "--reason"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %s: %v", want, err)
		}
	}
	if _, ok := familyGaps(t, run)["G1"]; ok {
		t.Error("the refused mint is on the board")
	}
}

// The fix's first sentence prints too, so the refusal covers required_fix as well.
func TestMintRefusesLaneTagInRequiredFix(t *testing.T) {
	_, lens := liveRegistryRun(t, map[string]recordpb.ClassMaterial{"g": recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE})
	m := liveMint("G1", "g", recordpb.Grade_GRADE_MEDIUM)
	m.Problem = proto.String("The count of witnesses is stated without a source.")
	m.RequiredFix = proto.String("Cite the census for the count [minority: lane-2/primary-literature].")
	_, err := Append(lens, m)
	if err == nil {
		t.Fatal("a required fix carrying a lane tag landed on the board; it prints in the report")
	}
	if !strings.Contains(err.Error(), "--fix") || !strings.Contains(err.Error(), "[minority: lane-2/primary-literature]") {
		t.Errorf("the refusal does not name the field and the term: %v", err)
	}
}

// mint_reason is red's argument to the other seats and never reaches the report, so process words in
// it land — while the same words in the problem would not.
func TestMintAcceptsProcessWordsInMintReason(t *testing.T) {
	run, lens := liveRegistryRun(t, map[string]recordpb.ClassMaterial{"g": recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE})
	const reason = "red-lens-evidence found this in sitting 2; blue-respond closed G1's fix without a check [lane-1]"
	m := liveMint("G1", "g", recordpb.Grade_GRADE_MEDIUM)
	m.Problem = proto.String("The error bound is stated without its derivation.")
	m.MintReason = proto.String(reason)
	appendOK(t, lens, m)
	if got := familyGaps(t, run)["G1"].Mint.GetMintReason(); got != reason {
		t.Errorf("the mint reason was not recorded as written: %q", got)
	}
	bad := liveMint("G2", "g", recordpb.Grade_GRADE_MEDIUM)
	bad.Problem = proto.String(reason)
	if _, err := Append(lens, bad); err == nil {
		t.Error("the same words landed as a problem statement; the refusal is not reading the problem")
	}
}

// Migration re-drives an archived mint through this write path, and an archived mint is what a seat
// DID: its process words are kept, not refused.
func TestMigratingReplaySkipsVoiceRefusal(t *testing.T) {
	run, lens := liveRegistryRun(t, map[string]recordpb.ClassMaterial{"g": recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE})
	Migrating = true
	defer func() { Migrating = false }()
	m := liveMint("G1", "g", recordpb.Grade_GRADE_MEDIUM)
	m.ClassMaterial = recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE.Enum()
	m.Problem = proto.String("blue-synthesize merged a claim from blue-lane-2 unsourced.")
	m.RequiredFix = proto.String("Source it [minority: lane-2/primary-literature].")
	appendOK(t, lens, m)
	if got := familyGaps(t, run)["G1"].Mint.GetProblem(); !strings.Contains(got, "blue-synthesize") {
		t.Errorf("the migrated problem was not kept as archived: %q", got)
	}
}

// A --new prescription becomes the report's text when blue accepts it, so migration replays one that
// carries a refused tell as archived — the same exemption the problem and the fix have.
func TestMigratingReplayKeepsAVoicedFixNew(t *testing.T) {
	run, lens := liveRegistryRun(t, map[string]recordpb.ClassMaterial{"g": recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE})
	Migrating = true
	defer func() { Migrating = false }()
	m := liveMint("G1", "g", recordpb.Grade_GRADE_MEDIUM)
	m.ClassMaterial = recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE.Enum()
	m.Problem = proto.String("The error bound is stated without its derivation.")
	m.Location = proto.String("The bound holds.")
	m.FixNew = proto.String("The bound holds, as blue-respond conceded.")
	appendOK(t, lens, m)
	if got := familyGaps(t, run)["G1"].Mint.GetFixNew(); !strings.Contains(got, "blue-respond") {
		t.Errorf("the migrated prescription was not kept as archived: %q", got)
	}
}

// A labelled corroboration's title prints in the source's note and Bibliography entry, and migration replays an archived one
// that carries a refused tell as archived.
func TestMigratingReplayKeepsAVoicedCorroborationTitle(t *testing.T) {
	_, lens := liveRegistryRun(t, map[string]recordpb.ClassMaterial{"g": recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE})
	Migrating = true
	defer func() { Migrating = false }()
	appendOK(t, lens, &recordpb.Verify{
		Independent: proto.Bool(true),
		Url:         proto.String("https://example.org/voiced"),
		Title:       proto.String("Standard found by red-lens-evidence"),
		Label:       proto.String("c-0123456789ab"),
		Claim:       proto.String("The bound holds."),
		Outcome:     recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS.Enum(),
		Confidence:  recordpb.Confidence_CONFIDENCE_HIGH.Enum(),
		Text:        proto.String("the standard states it"),
	})
}
