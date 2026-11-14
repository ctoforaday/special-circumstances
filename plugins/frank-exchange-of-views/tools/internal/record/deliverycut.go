package record

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// WHAT A DELIVERY OF REPORT TEXT CARRIES (constitution-gaps P3, §III.3.2).
//
// A seat's delivery is cut ONCE, when its first item is handed, from the report as it then stands;
// its head is the report head at that moment. The writer keeps the cut pieces and hands every later
// one from there, so every piece of one delivery is cut from one text and none is rebuilt from the
// report at a past event. The changes are read from the edits' own stamped paragraphs, never located
// in the report: only the whole report and the gaps' passages are read from report text, and only
// from the report now.
//
//   - A lens with no complete whole delivery, and at its full re-read, is cut the whole report: its
//     sections in order, packed into pieces; their text concatenated is the report at the cut.
//   - Every other lens sitting is cut each head move after its pin, as the edit's stamped paragraphs
//     before and after it (a part's text as the new side), then the passage of each gap it minted,
//     open or closed. With neither, one piece says the report has not changed since its pin.
//   - The chair is cut each stale area's changes it has not borne (area_change), as change pieces.
//   - The bench is cut the passage of each gap an unruled motion about gaps names.

// DeliveryPiece is one item of report text: a label saying what it is, and its text. Text is never
// trimmed; a piece too large for one item is split across consecutive pieces, each labelled as
// continuing.
type DeliveryPiece struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

// DeliveryCut is a seat's delivery as cut: the head it was cut at, whether it is the whole report,
// and its pieces in the order they are handed.
type DeliveryCut struct {
	Head   int64           `json:"head"`
	Whole  bool            `json:"whole"`
	Pieces []DeliveryPiece `json:"pieces"`
}

// CutDelivery cuts seatID's delivery from the record as it stands, each piece small enough that
// fits(label, text) holds. subjects is what the seat's report-text items are about
// (ReportTextSubjects): the chair's stale areas, the bench's gaps; a lens's are cut from the report.
// A seat owed no report text — no report yet, or a role that is handed none — is cut nothing (nil
// pieces).
func CutDelivery(run Run, seatID string, subjects []string, fits func(label, text string) bool) (DeliveryCut, error) {
	var cut DeliveryCut
	role := seatclass.RoleOfSeat(seatID)
	err := readSnapshot(run, func(q recordsql.Querier) error {
		head, err := reportHeadAt(q)
		if err != nil || head == 0 {
			return err
		}
		report, err := RenderAt(q)
		if err != nil {
			return err
		}
		cut.Head = head
		switch role {
		case "lens":
			if cut.Whole, err = NextDeliveryWhole(q, seatID); err != nil {
				return err
			}
			if cut.Whole {
				cut.Pieces = packWhole(report, fits)
				return nil
			}
			pin, err := PinAt(q, seatID)
			if err != nil {
				return err
			}
			changes, err := changeItemsAfter(q, report, pin, head, "")
			if err != nil {
				return err
			}
			passages, err := ownPassages(q, report, seatID)
			if err != nil {
				return err
			}
			cut.Pieces = pieces(changes, passages, fits)
			if len(cut.Pieces) == 0 {
				cut.Pieces = []DeliveryPiece{{Label: "THE REPORT SINCE YOUR PIN", Text: "The report has not changed since your pin, and you hold no gap of your own: nothing of it is new to you this sitting."}}
			}
		case "chair":
			var changes []DeliveryPiece
			for _, area := range subjects {
				cs, err := changeItemsAfter(q, report, -1, head, area)
				if err != nil {
					return err
				}
				changes = append(changes, cs...)
			}
			cut.Pieces = pieces(changes, nil, fits)
		case "bench":
			passages, err := gapsPassages(q, report, subjects)
			if err != nil {
				return err
			}
			cut.Pieces = pieces(nil, passages, fits)
		}
		return nil
	})
	return cut, err
}

// packWhole is the report's sections in order, packed into pieces that fit; a section that does not
// fit alone is split (splitFit). The pieces' text, concatenated, is the report.
func packWhole(report string, fits func(label, text string) bool) []DeliveryPiece {
	const label = "THE WHOLE REPORT"
	var units []string
	for rest := report; rest != ""; {
		i := strings.Index(rest[1:], "\n## ")
		if i < 0 {
			units = append(units, rest)
			break
		}
		units = append(units, rest[:i+2])
		rest = rest[i+2:]
	}
	var out []DeliveryPiece
	cur := ""
	flush := func() {
		if cur != "" {
			out = append(out, DeliveryPiece{Label: label, Text: cur})
			cur = ""
		}
	}
	for _, u := range units {
		if fits(label, cur+u) {
			cur += u
			continue
		}
		flush()
		if fits(label, u) {
			cur = u
			continue
		}
		for _, part := range splitFit(u, func(s string) bool { return fits(label, s) }) {
			out = append(out, DeliveryPiece{Label: label, Text: part})
		}
	}
	flush()
	return out
}

// splitFit splits text into consecutive pieces that each fit, at block boundaries, then sentence
// boundaries, then word boundaries, then characters; the pieces concatenated are text.
func splitFit(text string, fits func(string) bool) []string {
	if fits(text) {
		return []string{text}
	}
	for _, cuts := range [][]int{blockCuts(text), sentenceCuts(text), wordCuts(text)} {
		if len(cuts) == 0 {
			continue
		}
		var out []string
		from, last := 0, 0
		ok := true
		for _, c := range append(cuts, len(text)) {
			if c <= from {
				continue
			}
			if fits(text[from:c]) {
				last = c
				continue
			}
			if last == from {
				ok = false
				break
			}
			out = append(out, text[from:last])
			from = last
			if !fits(text[from:c]) {
				ok = false
				break
			}
			last = c
		}
		if ok {
			if from < len(text) {
				out = append(out, text[from:])
			}
			return out
		}
	}
	// A single word over the budget: cut by characters.
	var out []string
	rs := []rune(text)
	from := 0
	for from < len(rs) {
		to := from + 1
		for to < len(rs) && fits(string(rs[from:to+1])) {
			to++
		}
		out = append(out, string(rs[from:to]))
		from = to
	}
	return out
}

func blockCuts(text string) []int {
	var out []int
	for _, b := range anchor.Blocks(text) {
		if b.Start > 0 {
			out = append(out, b.Start)
		}
	}
	return out
}

func sentenceCuts(text string) []int {
	var out []int
	for _, s := range anchor.Sentences(text) {
		if s[0] > 0 {
			out = append(out, s[0])
		}
	}
	return out
}

func wordCuts(text string) []int {
	var out []int
	for i, r := range text {
		if i > 0 && unicode.IsSpace(r) {
			out = append(out, i)
		}
	}
	return out
}

// change is one head move: the edit's stamped paragraphs before and after it, or a part's text as
// the new side, under the heading of the section its new side now stands in.
type change struct {
	heading, before, after string
}

// changesAfterSQL is the head moves after ?1 and at or before ?2, in record order: each edit's stamped
// paragraphs, and each part's text as the new side. With ?3 set, only the area_change rows of that lens.
const changesAfterSQL = `SELECT h."event_id", h."before", h."after" FROM (
    SELECT ed."event_id", COALESCE(ed."old_paragraphs", '') AS "before", COALESCE(ed."new_paragraphs", '') AS "after" FROM "blue_edit" ed
    UNION ALL
    SELECT p."event_id", '', p."text" FROM "report_part" p) h
  WHERE h."event_id" > ?1 AND h."event_id" <= ?2
    AND (?3 = '' OR h."event_id" IN (SELECT a."change" FROM "area_change" a WHERE a."seat_id" = ?3))
  ORDER BY h."event_id"`

// changeItemsAfter is each head move after pin and at or before head, as change pieces; with area set,
// the area_change rows of that lens (pin is then not read). A move whose two sides differ only in
// anchors is no change.
func changeItemsAfter(q recordsql.Querier, report string, pin, head int64, area string) ([]DeliveryPiece, error) {
	rows, err := q.Query(changesAfterSQL, pin, head, area)
	if err != nil {
		return nil, fmt.Errorf("reading the changes after the pin: %w", err)
	}
	defer rows.Close()
	var out []DeliveryPiece
	for rows.Next() {
		var id int64
		var c change
		if err := rows.Scan(&id, &c.before, &c.after); err != nil {
			return nil, err
		}
		if anchortext.Visible(c.before) == anchortext.Visible(c.after) {
			continue
		}
		c.heading = headingOf(report, c.after)
		out = append(out, changePiece(c, area))
	}
	return out, rows.Err()
}

// headingOf is the heading of the section of report the text stands in, "" where it is not found.
func headingOf(report, text string) string {
	i := strings.Index(report, text)
	if text == "" || i < 0 {
		return ""
	}
	from, level := sectionStart(report, i)
	if level == 0 {
		return ""
	}
	line := report[from:]
	if j := strings.IndexByte(line, '\n'); j >= 0 {
		line = line[:j]
	}
	return line
}

// changePiece renders a change: its section, then its paragraphs before and after; an empty side is
// shown as added or removed.
func changePiece(c change, area string) DeliveryPiece {
	label := "A CHANGE TO THE REPORT"
	if area != "" {
		label = "A CHANGE BEHIND " + area + "'S PIN"
	}
	var b strings.Builder
	if c.heading != "" {
		b.WriteString("In section: " + c.heading + "\n\n")
	}
	switch {
	case c.before == "":
		b.WriteString("ADDED:\n" + c.after)
	case c.after == "":
		b.WriteString("REMOVED:\n" + c.before)
	default:
		b.WriteString("BEFORE:\n" + c.before + "\n\nAFTER:\n" + c.after)
	}
	return DeliveryPiece{Label: label, Text: b.String()}
}

// ownPassagesSQL is the gaps seatID minted, in mint order, with their recorded location.
const ownPassagesSQL = `SELECT m."gap_id", COALESCE(m."location", '') FROM "mint" m JOIN "events" e ON e."id" = m."event_id"
  WHERE e."seat_id" = ?1 ORDER BY m."event_id"`

// ownPassages is the passage of each gap the lens minted, open or closed.
func ownPassages(q recordsql.Querier, report, seatID string) ([]DeliveryPiece, error) {
	return gapPassages(q, report, ownPassagesSQL, seatID)
}

// gapPassageSQL is gap ?1 and its recorded location.
const gapPassageSQL = `SELECT "gap_id", COALESCE("location", '') FROM "mint" WHERE "gap_id" = ?1`

// gapsPassages is the passage of each of gaps, in order.
func gapsPassages(q recordsql.Querier, report string, gaps []string) ([]DeliveryPiece, error) {
	var out []DeliveryPiece
	for _, g := range gaps {
		ps, err := gapPassages(q, report, gapPassageSQL, g)
		if err != nil {
			return nil, err
		}
		out = append(out, ps...)
	}
	return out, nil
}

// latestReopeningSQL is the stamped paragraphs before the latest edit whose reopened lists gap ?1.
const latestReopeningSQL = `SELECT COALESCE(b."old_paragraphs", '') FROM "blue_edit_reopened" r JOIN "blue_edit" b ON b."event_id" = r."event_id"
  WHERE r."value" = ?1 ORDER BY r."event_id" DESC LIMIT 1`

// gapPassages is the passage of each gap the statement names: the gap's section as it stands
// (PassageAround); for a gap whose anchor is gone or bare, the paragraph that held it before the
// latest edit that reopened it; with neither, its recorded location.
func gapPassages(q recordsql.Querier, report, stmt, arg string) ([]DeliveryPiece, error) {
	rows, err := q.Query(stmt, arg)
	if err != nil {
		return nil, fmt.Errorf("reading the gaps whose passages are delivered: %w", err)
	}
	type gap struct{ id, location string }
	var gaps []gap
	for rows.Next() {
		var g gap
		if err := rows.Scan(&g.id, &g.location); err != nil {
			rows.Close()
			return nil, err
		}
		gaps = append(gaps, g)
	}
	rows.Close()
	bare := map[string]bool{}
	for _, id := range claimcount.BareAnchorIDs(report) {
		bare[id] = true
	}
	var out []DeliveryPiece
	for _, g := range gaps {
		tok := anchor.Token(g.id)
		if at := strings.Index(report, tok); at >= 0 && !bare[g.id] {
			out = append(out, DeliveryPiece{Label: "THE PASSAGE OF GAP " + g.id, Text: PassageAround(report, at)})
			continue
		}
		var before string
		if err := q.QueryRow(latestReopeningSQL, g.id).Scan(&before); err == nil && before != "" {
			for _, b := range anchor.Blocks(before) {
				if p := before[b.Start:b.End]; strings.Contains(p, tok) {
					out = append(out, DeliveryPiece{Label: "THE PASSAGE OF GAP " + g.id + ", AS IT STOOD BEFORE THE CHANGE THAT REMOVED ITS TEXT", Text: p})
					break
				}
			}
			continue
		}
		out = append(out, DeliveryPiece{Label: "THE PASSAGE OF GAP " + g.id + ", ITS TEXT GONE FROM THE REPORT; ITS RECORDED LOCATION", Text: g.location})
	}
	return out, nil
}

// pieces is the changes, then each passage no other piece carries whole, each split to fit.
func pieces(changes, passages []DeliveryPiece, fits func(label, text string) bool) []DeliveryPiece {
	var out []DeliveryPiece
	all := append(append([]DeliveryPiece{}, changes...), passages...)
	for i, p := range all {
		if i >= len(changes) && carriedElsewhere(p.Text, all, i) {
			continue
		}
		parts := splitFit(p.Text, func(s string) bool { return fits(p.Label+" (continued)", s) })
		for k, t := range parts {
			label := p.Label
			if k > 0 {
				label += " (continued)"
			}
			out = append(out, DeliveryPiece{Label: label, Text: t})
		}
	}
	return out
}

// carriedElsewhere is whether another piece of the delivery holds every character of text.
func carriedElsewhere(text string, all []DeliveryPiece, self int) bool {
	for j, q := range all {
		if j != self && strings.Contains(q.Text, text) && (j < self || len(q.Text) > len(text)) {
			return true
		}
	}
	return false
}
