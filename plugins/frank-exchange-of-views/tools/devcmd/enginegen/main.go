// Command enginegen writes the seat block of the engine script from the record's roster.
//
// debate.js is authored source with ONE generated region. The script dispatches every seat, so it
// needs each seat's role, tier class, agent configuration and lens area, the lens areas, the
// bench's occasions and the mass mapping's stamp; Go owns every one of those. This command renders
// them between two marker lines, and `-check` fails when the committed block is not what the
// roster renders — a stale block after a roster edit, or a block edited by hand.
//
// It lives in this module and not under scripts/ because it reads internal/seatclass,
// internal/record and internal/flags, which nothing outside the module can import.
//
// Run from the tools module:
//
//	go run ./devcmd/enginegen          write the block
//	go run ./devcmd/enginegen -check   exit 1 when the committed block is stale
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	// THIS COMMAND RENDERS NO VIEW. It links the record for its occasion words and mass mapping
	// stamp, and integration/surface holds every binary that links the record to linking the
	// renderer the record's views place quote gaps through.
	_ "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

const (
	beginMark = "// BEGIN GENERATED SEATS — go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen writes this from the record's roster. DO NOT EDIT BY HAND."
	endMark   = "// END GENERATED SEATS"
)

// source is everything the block states, as Go holds it.
type source struct {
	seats       map[string]seatclass.Seat
	lane        seatclass.Seat
	lanePrefix  string
	classOf     func(base string) string
	areas       []string
	occasions   []string
	massVersion string
}

func fromTheRecord() source {
	return source{
		seats:       seatclass.Seats,
		lane:        seatclass.Lane,
		lanePrefix:  seatclass.LaneSeatPrefix,
		classOf:     seatclass.ClassOf,
		areas:       flags.LensAreas,
		occasions:   record.OccasionWords(),
		massVersion: record.MassMappingVersion,
	}
}

func main() {
	check := flag.Bool("check", false, "verify the committed block matches what the roster renders, instead of writing it")
	flag.Parse()

	path, err := repotree.DebateJS()
	if err != nil {
		fail(err)
	}
	cur, err := os.ReadFile(path)
	if err != nil {
		fail(err)
	}
	block, err := render(fromTheRecord())
	if err != nil {
		fail(err)
	}
	out, err := splice(cur, block)
	if err != nil {
		fail(fmt.Errorf("%s: %w", path, err))
	}
	if bytes.Equal(cur, out) {
		fmt.Println("enginegen: the engine's seat block matches the roster")
		return
	}
	if *check {
		fmt.Fprintf(os.Stderr, "enginegen: the seat block in %s is stale.\n"+
			"It is written from the record's roster (seatclass.Seats and Lane, flags.LensAreas, the Occasion enum, record.MassMappingVersion). "+
			"Edit the roster, never the block.\n"+
			"Regenerate: go -C plugins/frank-exchange-of-views/tools run ./devcmd/enginegen\n", path)
		os.Exit(1)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		fail(err)
	}
	fmt.Printf("enginegen: wrote the seat block in %s\n", path)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "enginegen:", err)
	os.Exit(1)
}

// render is the block's lines, between the markers and not including them.
//
// IT REFUSES WHAT THE ENGINE COULD NOT DISPATCH. A row with no configuration or no tier class
// renders as a seat the script dispatches under nothing, and an empty area list or vocabulary
// renders as a script that seats no lens or convenes no bench; each is a roster defect, and
// writing it out would move the failure into a run.
func render(s source) ([]byte, error) {
	var b strings.Builder
	row := func(seat seatclass.Seat, id string) (string, error) {
		if seat.AgentType == "" {
			return "", fmt.Errorf("the roster seats %s under no agent configuration", id)
		}
		tier := s.classOf(seat.Base)
		if tier == "" {
			return "", fmt.Errorf("the roster gives %s the tier base %q, which has no tier class", id, seat.Base)
		}
		return fmt.Sprintf("role: %s, tier: %s, agentType: %s", js(seat.Role), js(tier), js(seat.AgentType)), nil
	}

	ids := make([]string, 0, len(s.seats))
	for id, seat := range s.seats {
		// The operator acts outside the debate: nothing dispatches it, and it carries no configuration.
		if seat.Role == seatclass.OperatorRole {
			continue
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	lensOf := map[string]string{}
	b.WriteString("const SEATS = {\n")
	for _, id := range ids {
		seat := s.seats[id]
		fields, err := row(seat, id)
		if err != nil {
			return nil, err
		}
		if seat.Area != "" {
			lensOf[seat.Area] = id
			fields += ", area: " + js(seat.Area)
		}
		fmt.Fprintf(&b, "  %s: { %s },\n", js(id), fields)
	}
	b.WriteString("}\n")

	fields, err := row(s.lane, "a lane")
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(&b, "const LANE = { prefix: %s, %s }\n", js(s.lanePrefix), fields)

	if len(s.areas) == 0 {
		return nil, fmt.Errorf("the record declares no lens area")
	}
	for _, a := range s.areas {
		if lensOf[a] == "" {
			return nil, fmt.Errorf("the lens area %s has no seat on the roster", a)
		}
		delete(lensOf, a)
	}
	for a, id := range lensOf {
		return nil, fmt.Errorf("the roster seats %s for the area %s, which the record does not declare", id, a)
	}
	fmt.Fprintf(&b, "const RED_AREAS = [%s]\n", strings.Join(each(s.areas, js), ", "))

	if len(s.occasions) == 0 {
		return nil, fmt.Errorf("the record's Occasion enum holds no word")
	}
	for _, w := range s.occasions {
		if !identifier.MatchString(w) {
			return nil, fmt.Errorf("the occasion %q is not a word the engine can read as a key", w)
		}
	}
	fmt.Fprintf(&b, "const OCCASION = { %s }\n", strings.Join(each(s.occasions, func(w string) string { return w + ": " + js(w) }), ", "))

	if s.massVersion == "" {
		return nil, fmt.Errorf("the record's mass mapping carries no version")
	}
	fmt.Fprintf(&b, "const MASS_MAPPING_VERSION = %s\n", js(s.massVersion))

	if strings.Contains(b.String(), unwritable) {
		return nil, fmt.Errorf("a roster value holds a character a single-quoted script string cannot carry")
	}
	return []byte(b.String()), nil
}

var (
	identifier = regexp.MustCompile(`^[a-z][a-z_]*$`)
	plain      = regexp.MustCompile(`^[A-Za-z0-9 _:.-]+$`)
)

// unwritable stands in for a value js cannot quote; render refuses a block that holds it.
const unwritable = "\x00"

// js is a value as a single-quoted script string. Roster values are ids and words; one that needs
// escaping is refused by render rather than escaped here by a rule nothing else exercises.
func js(v string) string {
	if !plain.MatchString(v) {
		return unwritable
	}
	return "'" + v + "'"
}

func each(in []string, f func(string) string) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

// splice replaces what stands between the markers. A script without both markers is refused: the
// block's place in the script is authored, because the declarations must precede their first use.
func splice(cur, block []byte) ([]byte, error) {
	i := bytes.Index(cur, []byte(beginMark+"\n"))
	if i < 0 {
		return nil, fmt.Errorf("no line %q — the seat block's place in the script is authored, and this command fills it", beginMark)
	}
	from := i + len(beginMark) + 1
	j := bytes.Index(cur[from:], []byte(endMark+"\n"))
	if j < 0 {
		return nil, fmt.Errorf("the seat block opens and never closes — the line %q is missing", endMark)
	}
	out := append([]byte{}, cur[:from]...)
	out = append(out, block...)
	return append(out, cur[from+j:]...), nil
}
