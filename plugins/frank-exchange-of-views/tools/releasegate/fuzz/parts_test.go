package fuzz

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// THE REPORT IS THE LANES' PARTS, AND THE SYNTHESIZER WORKS ON THEM THROUGH ITS ACTS. Each lane the
// engine dispatches writes its part to the path its seat names and records it; the synthesizer frames
// the joined parts with the title and the open questions by edits. A drawn run records two or three
// parts: the third lane withholds its part now and then.
//
// THE PART'S REFUSALS ARE DRIVEN BY CONSTRUCTION. partsSeed's run drives every one on every seat that
// can reach it, and a drawn run takes each arm at partsPct: a lane asks before its file is written
// (no file), with a title in its part (a surface that is not a part's), and again once recorded (a
// second part); once recorded, a lane types each act that changes the report — edit, cite, compute
// and retire — and each is refused as blue's, because a lane's surface does not carry it and the
// report's first change would close the door on the lanes still writing; the frontier reads the
// report before any lane has recorded (no part); the synthesizer types `ingest`, which is the
// lane's (no lane); and a lane that withheld its part asks once the synthesizer has changed the
// report (after the change). Each is the refusal's own code and text, or the run
// fails at the act; the sweep gates each count at zero.

// partsSeed is the seed whose run is FORCED through every part refusal above. It is none of the forced
// seeds before it: the withheld part changes the report every one of them reads.
const partsSeed = 10

// partsPct is how often a drawn run takes each part arm where it can.
const partsPct = 15

// withheldLane is the lane that withholds its part, on partsSeed's run and on a drawn run that
// records two parts.
const withheldLane = seatclass.LaneSeatPrefix + "3"

// The part refusals, by what each refuses, as the sweep tallies them.
const (
	partNotALane   = "not a lane"
	partSecond     = "a second part"
	partAfter      = "after the report changed"
	partNoFile     = "no file"
	partSurface    = "a surface not a part's"
	partNoneToRead = "no part to read"
	partLaneEdit   = "a lane's edit"
	partLaneCite   = "a lane's cite"
	partLaneComp   = "a lane's compute"
	partLaneRetire = "a lane's retire"
)

// laneRefusedActs is each act that changes the report, by the refusal kind a lane typing it counts
// under. One kind per act, so the forced run fails on any one of them the drive stopped reaching.
var laneRefusedActs = []struct{ kind, verb string }{
	{partLaneEdit, "edit"}, {partLaneCite, "cite"}, {partLaneComp, "compute"}, {partLaneRetire, "retire"},
}

// partRefusals is every kind above, in the order the sweep reports them.
var partRefusals = []string{partNotALane, partSecond, partAfter, partNoFile, partSurface, partNoneToRead,
	partLaneEdit, partLaneCite, partLaneComp, partLaneRetire}

// The report every run's synthesis makes of its parts: the title, the parts, the open questions —
// seedReport, whatever order the lanes record in.
const (
	seedTitle = "# § fuzz"
	seedOpen  = "## Open questions\n\nWhether the trend outlasts the sample."
)

// lanePartText is lane i's part: a sentence of its own, under no heading of its own.
func lanePartText(lane string) string {
	switch lane {
	case seatclass.LaneSeatPrefix + "1":
		return "A § fuzz sentence to anchor findings.\n"
	case seatclass.LaneSeatPrefix + "2":
		return "The cost is rising over time across the whole sampled corpus.\n"
	}
	return "Across the sampled corpus the rise holds in every year measured.\n"
}

// partsArm is whether this run takes a part arm now: always on the forced seed, and by the draw
// otherwise.
func (r *runner) partsArm() bool { return r.forceParts || r.coin(partsPct) }

// partFile is where seat's part waits for ingest: the path the tool builds from the seat.
func (r *runner) partFile(seat string) string {
	return filepath.Join(r.runDir, "blue", "candidates", seat+".md")
}

func (r *runner) writePart(seat, text string) {
	if err := writeSetup(r.partFile(seat), []byte(text)); err != nil {
		r.noteActErr("writing " + seat + "'s part: " + err.Error())
	}
}

// refusePart drives an act to the part refusal of kind: refused with code, saying want. Anything else
// fails the run at the act.
func (r *runner) refusePart(kind string, code feov.Code, want string, args ...string) {
	args = append(args, "--json")
	out, err := r.exec(args...)
	argv := strings.Join(args, " ")
	if err == nil {
		r.noteActErr(argv + ": admitted, and the drive drives it to be refused (" + kind + "): " + firstLine(out))
		return
	}
	env, ok := refusalEnvelopeIn(out)
	switch {
	case !ok:
		r.noteActErr(argv + ": refused under --json with no refusal envelope: " + firstLine(out))
	case code != "" && env.Code != string(code):
		r.noteActErr(argv + ": refused with code " + env.Code + ", and the drive drives a " + string(code) + " refusal (" + kind + "): " + firstLine(env.Error))
	case !strings.Contains(env.Error, want):
		r.noteActErr(argv + ": refused, and not as " + kind + " (" + want + "): " + firstLine(env.Error))
	default:
		r.mu.Lock()
		if r.partRefused == nil {
			r.partRefused = map[string]int{}
		}
		r.partRefused[kind]++
		r.mu.Unlock()
	}
}

// frontierReadsNoPart is the frontier reading the report before any lane has recorded its part:
// there is none to read.
func (r *runner) frontierReadsNoPart(seatID string) {
	if !r.partsArm() {
		return
	}
	r.refusePart(partNoneToRead, "", reportproj.NoPart, "show", "report", "--seat-id", seatID)
}

// withholds is whether lane records no part this run: the forced run's third lane, and a drawn run's
// one time in five.
func (r *runner) withholds(lane string) bool {
	if lane != withheldLane {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.withheld == nil {
		r.withheld = map[string]bool{}
		r.withheld[lane] = r.forceParts || r.rng.Intn(100) < 20
	}
	return r.withheld[lane]
}

// lanePart is a lane's sitting: its part written to the path its seat names, and recorded.
func (r *runner) lanePart(seatID string) {
	if r.withholds(seatID) {
		return
	}
	part := lanePartText(seatID)
	r.laneResearches(seatID)
	if r.partsArm() {
		r.refusePart(partNoFile, feov.NotFound, "there is no part at "+r.partFile(seatID), "ingest", "--seat-id", seatID)
	}
	if r.partsArm() {
		r.writePart(seatID, "# Lane "+strings.TrimPrefix(seatID, seatclass.LaneSeatPrefix)+": § fuzz\n\n"+part)
		r.refusePart(partSurface, feov.Validation, "is the synthesizer's to write, not a lane's", "ingest", "--seat-id", seatID)
	}
	r.writePart(seatID, part)
	if _, ok := r.admit("ingest", "--seat-id", seatID); !ok {
		return
	}
	r.mu.Lock()
	r.partsRecorded++
	r.mu.Unlock()
	if _, err := os.Stat(r.partFile(seatID)); !os.IsNotExist(err) {
		r.noteActErr(seatID + "'s part was recorded and its file still stands: " + r.partFile(seatID))
	}
	if r.partsArm() {
		r.refusePart(partSecond, feov.Conflict, "has already recorded its part of the report", "ingest", "--seat-id", seatID)
	}
	if r.partsArm() {
		r.laneActsOnItsPart(seatID, strings.TrimSuffix(part, "\n"))
	}
	r.laneReadsTheReport(seatID)
}

// laneResearches is the rest of a lane's surface before its part, drawn as the part arms are: it
// reads its work and the inquiries, proposes an approach of its own and records that it took it up,
// logs a note, and fetches a source and searches it by its sha256. Each is admitted.
func (r *runner) laneResearches(seatID string) {
	if !r.partsArm() {
		return
	}
	for _, v := range [][]string{{"show"}, {"show", "work"}, {"show", "inquiries"}} {
		r.admit(append(v, "--seat-id", seatID)...)
	}
	out, ok := r.admit("inquiry", "propose", "--seat-id", seatID, "--question", "fuzz: does "+seatID+"'s slice settle it",
		"--hypothesis", "the slice's sources agree")
	if ok {
		if id := inquiryIDOf(out); id != "" {
			r.noteInquiry(id)
			r.admit("inquiry", "record", "--seat-id", seatID, "--id", id, "--as", "underway", "--reason", "fuzz: the lane takes it up")
		} else {
			r.noteInquiryIDErr(seatID + "'s propose named no inquiry id: " + firstLine(out))
		}
	}
	r.admit("log", "--seat-id", seatID, "--type", "friction", "--reason", "fuzz: a lane's note for the operator")
	out, ok = r.admit("fetch", "--seat-id", seatID, "--url", sourceURL("/"+seatID), "--ocr=false", "--json")
	var sum struct {
		Sha256 string `json:"sha256"`
	}
	if !ok || json.Unmarshal([]byte(out), &sum) != nil || sum.Sha256 == "" {
		return
	}
	r.admit("show", "source", "--seat-id", seatID, "--sha", sum.Sha256, "--match", ".")
}

// laneReadsTheReport is a lane reading the projections once its part is recorded, drawn as the
// part arms are. Each is admitted: a read changes nothing, so it closes no door on another lane.
func (r *runner) laneReadsTheReport(seatID string) {
	if !r.partsArm() {
		return
	}
	for _, v := range []string{"report", "board", "changes", "evidence", "findings"} {
		r.admit("show", v, "--seat-id", seatID)
	}
}

// laneActsOnItsPart is a lane, its part recorded, typing each act that changes the report on a
// sentence of its own part. Each is refused as blue's verb and records nothing; the lanes still
// writing record their parts after it, which partsRecorded counts.
func (r *runner) laneActsOnItsPart(seatID, sentence string) {
	for _, a := range laneRefusedActs {
		r.refusePart(a.kind, feov.RoleViolation, fmt.Sprintf("%q is not on your surface — it is the blue seat's verb", a.verb),
			a.verb, "--seat-id", seatID, "--quote", sentence, "--reason", "fuzz: a lane acting on its own part")
	}
}

// synthesize is the synthesizer's sitting on the lanes' parts: a part of its own refused, then the
// title before the first part and the open questions after the last, each an edit quoting the text
// beside it. A sitting that finds the report framed — the synthesis sat again — frames nothing.
// Once it has changed the report, the lane that withheld its part asks, and is refused.
func (r *runner) synthesize(seatID string) {
	if r.partsArm() {
		r.writePart(seatID, lanePartText(seatID))
		r.refusePart(partNotALane, feov.RoleViolation, `"ingest" is not on your surface — it is the lane seat's verb`, "ingest", "--seat-id", seatID)
		_ = os.Remove(r.partFile(seatID))
	}
	text, err := reportproj.RenderFromRecord(r.run())
	if err != nil {
		r.noteActErr("the synthesizer could not read the lanes' parts: " + err.Error())
		return
	}
	if strings.HasPrefix(text, seedTitle) {
		return
	}
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	first, last := lines[0], lines[len(lines)-1]
	if _, ok := r.admit("edit", "--seat-id", seatID, "--key", "frame-title", "--quote", first,
		"--new", seedTitle+"\n\n"+first, "--reason", "fuzz: the title before the lanes' parts"); !ok {
		return
	}
	if _, ok := r.admit("edit", "--seat-id", seatID, "--key", "frame-open", "--quote", last,
		"--new", last+"\n\n"+seedOpen, "--reason", "fuzz: the open questions after the lanes' parts"); !ok {
		return
	}
	if r.withholds(withheldLane) {
		r.writePart(withheldLane, lanePartText(withheldLane))
		r.refusePart(partAfter, feov.Conflict, "the report has already been changed", "ingest", "--seat-id", withheldLane)
	}
}

// partsMissed is, on partsSeed's run, each part refusal the run drove none of.
func (r *runner) partsMissed() []string {
	if !r.forceParts {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	var missed []string
	for _, kind := range partRefusals {
		if r.partRefused[kind] == 0 {
			missed = append(missed, kind)
		}
	}
	return missed
}

// stagePart records text as blue-lane-1's part, through the lane's own ingest, for a test that drives
// the engine through a script of its own rather than debate.js's lanes; blue-synthesize registers as
// the seat that works on the report from there.
func (r *runner) stagePart(text string) error {
	lane := seatclass.LaneSeatPrefix + "1"
	if err := writeSetup(r.partFile(lane), []byte(text)); err != nil {
		return err
	}
	r.register(seatclass.Lane.Role, lane)
	if _, err := r.do("ingest", lane).result(); err != nil {
		return err
	}
	r.register("blue", "blue-synthesize")
	return nil
}
