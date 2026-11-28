package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migration"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// TRANSLATION 7 — A SUPERSEDE LINK (EventSchema 82, plans/findings-fold.md §III.B8). The record holds no
// link between gaps: a gap's wording is its minter's to restate on its open gap (Amend), under the
// gap's own id, and a different defect is a gap of its own. An archived mint that superseded a gap is
// one of the two:
//
//   - an AMEND of the chain's first gap, where the link is the seat's own (the seats as the migrated
//     record holds them), the ancestor is open at the successor's mint, and the successor keeps the
//     ancestor's class, subject (`about`), location and check kind — the four fields an amend cannot
//     change. The amend carries the successor's wording, the bytes the seat wrote, with a Regrade
//     where its grades differ; every later reference to the successor's id names the first gap
//     (remap.rekey), and a successor token beside the first gap's collapses to one;
//   - its OWN GAP otherwise, the link dropped: across seats, a closed ancestor, a moved fixed field,
//     or an ancestor the migrated record does not hold.
//
// The ancestor is read off the SOURCE stream (supersededBy), since no translated mint carries it, and
// classified against the record the replay has written so far, where a lost act or a dropped bench
// act may leave it otherwise than the source did. A list-era mint names its ancestors as a list; its
// first is the link and the rest are not carried (quadratic-formula R2-7 keeps R1-4).

// supersedesDrop is why a mint's supersedes is not carried as a field.
const supersedesDrop = "a gap names no gap it replaces: the link is Translation 7's, read off the source stream — an amend of the chain's first gap where the seat restated its own open gap's wording, else the successor stands as its own gap"

// supersededBy is the source pre-pass: each archived mint that names an ancestor, by its source id,
// and the first ancestor it names as the source spells it.
func supersededBy(evs []OldEvent) map[int64]string {
	out := map[int64]string{}
	for _, e := range evs {
		if e.Word != "mint" {
			continue
		}
		if a, _ := e.Fields["supersedes"].(string); a != "" {
			out[e.ID] = a
			continue
		}
		if as := e.Lists["supersedes"]; len(as) > 0 && as[0] != "" {
			out[e.ID] = as[0]
		}
	}
	return out
}

// Link outcomes (migration.SupersedeLink.Outcome).
const (
	linkAmended        = "amended"
	linkCrossSeat      = "cross_seat"
	linkClosedAncestor = "closed_ancestor"
	linkMovedField     = "moved_field"
	linkAbsentAncestor = "absent_ancestor"
)

// ancestorRow is what the classification reads of the ancestor on the record as written so far.
type ancestorRow struct {
	open                                 bool
	mintedBy                             string
	class, aboutKind, aboutRef, location string
	checkKind                            string
	grades                               [4]sql.NullString
}

// supersedeLink classifies the link from successor mint m, written by seatID, to ancestor (as the
// migrated record spells it) on the record q holds, and for an amended link builds the acts it
// becomes: the Amend, and the Regrade where the grades differ (nil where they do not).
func supersedeLink(q recordsql.Querier, m *recordpb.Mint, seatID, ancestor string) (migration.SupersedeLink, *recordpb.Amend, *recordpb.Regrade, error) {
	link := migration.SupersedeLink{Successor: m.GetGapId(), Ancestor: ancestor}
	var a ancestorRow
	var aboutKind, aboutRef, location sql.NullString
	err := q.QueryRow(`SELECT g."open", g."minted_by", g."class", g."about_kind", g."about_ref", g."location", g."check_kind",
	    g."current_severity", g."current_likelihood", g."current_impact", g."current_complexity_cost"
	  FROM "gap" g WHERE g."gap_id" = ?`, ancestor).
		Scan(&a.open, &a.mintedBy, &a.class, &aboutKind, &aboutRef, &location, &a.checkKind,
			&a.grades[0], &a.grades[1], &a.grades[2], &a.grades[3])
	if errors.Is(err, sql.ErrNoRows) {
		link.Outcome = linkAbsentAncestor
		return link, nil, nil, nil
	}
	if err != nil {
		return link, nil, nil, fmt.Errorf("reading %s, which %s supersedes: %w", ancestor, m.GetGapId(), err)
	}
	a.aboutKind, a.aboutRef, a.location = aboutKind.String, aboutRef.String, location.String
	switch {
	case a.mintedBy != seatID:
		link.Outcome = linkCrossSeat
		return link, nil, nil, nil
	case !a.open:
		link.Outcome = linkClosedAncestor
		return link, nil, nil, nil
	}
	if link.Moved = movedFields(a, m); len(link.Moved) > 0 {
		link.Outcome = linkMovedField
		return link, nil, nil, nil
	}
	link.Outcome = linkAmended
	basis := m.GetMintReason()
	if strings.TrimSpace(basis) == "" {
		basis = m.GetProblem()
	}
	amend := &recordpb.Amend{
		GapId:           proto.String(ancestor),
		Problem:         m.Problem,
		RequiredFix:     m.RequiredFix,
		AcceptanceCheck: m.AcceptanceCheck,
		FixNew:          m.FixNew,
		FixBasis:        m.FixBasis,
		Basis:           proto.String(basis),
	}
	regrade := &recordpb.Regrade{GapId: proto.String(ancestor), Basis: proto.String(basis)}
	moved := false
	for k, g := range []*recordpb.Grade{m.Severity, m.Likelihood, m.Impact, m.ComplexityCost} {
		if g == nil || a.grades[k].Valid && a.grades[k].String == recordpb.Word(*g) {
			continue
		}
		moved = true
		v := g.Enum()
		switch k {
		case 0:
			regrade.Severity = v
		case 1:
			regrade.Likelihood = v
		case 2:
			regrade.Impact = v
		case 3:
			regrade.ComplexityCost = v
		}
	}
	if !moved {
		regrade = nil
	}
	link.Regraded = regrade != nil
	return link, amend, regrade, nil
}

// movedFields names the fixed fields successor m holds otherwise than its ancestor: the class, the
// subject (about kind and ref), the location with its anchor tokens taken out, and the check kind.
func movedFields(a ancestorRow, m *recordpb.Mint) []string {
	var moved []string
	if a.class != recordpb.Word(m.GetClass()) {
		moved = append(moved, "class")
	}
	if a.aboutKind != aboutWord(m) || a.aboutRef != m.GetAboutRef() {
		moved = append(moved, "about")
	}
	if untokened(a.location) != untokened(m.GetLocation()) {
		moved = append(moved, "location")
	}
	if a.checkKind != recordpb.Word(m.GetCheckKind()) {
		moved = append(moved, "check_kind")
	}
	return moved
}

// aboutWord is m's about kind as the record stores it, "" where it names none.
func aboutWord(m *recordpb.Mint) string {
	if m.AboutKind == nil {
		return ""
	}
	return recordpb.Word(m.GetAboutKind())
}

// untokened is a location with every anchor token taken out and its ends trimmed: a seat's quote of
// the same sentence carries whichever tokens stood in it when it quoted.
func untokened(s string) string {
	return strings.TrimSpace(anchor.Replace(s, func(string, string) string { return "" }))
}

// actOnAmendedAncestor reports whether body, translated from src, is a close or a regrade that names,
// as the source spells it, a gap an amended link restated (amended holds each such first gap) — the
// gap itself, not a successor re-keyed to it. Asked after the link's amend is written, so only an act
// after the successor's mint is one.
func actOnAmendedAncestor(rm *remap, src OldEvent, body proto.Message, amended map[string]bool) bool {
	switch body.(type) {
	case *recordpb.Close, *recordpb.Regrade:
	default:
		return false
	}
	raw, _ := src.Fields["gap_id"].(string)
	if raw == "" {
		return false
	}
	id := rm.respelled(raw)
	if _, successor := rm.rekeyed[id]; successor {
		return false
	}
	return amended[id]
}
