package record

import (
	"database/sql"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// A GAP HAS TWO PARTIES: its raiser, the lens that minted it, and its subject's author. authorOf is
// the author, and every party test on the gap path reads it — who files a proposal, who answers,
// whose accept of the raiser's statement concedes, whose closing states a word. The SQL twin is
// recordsql.authorSeatSQL; the two state one rule.
//
// A gap about a gap belongs to that gap's minter: the defect is in its wording or grades, which only
// its minter changes. Every other subject — a sentence, an anchor, a section, an inquiry, the
// question — is report text or a record of blue's, and belongs to blue's responding seat.
// blue-synthesize and frontier sit only before any lens, so no gap names them.
func authorOf(aboutKind recordpb.AboutKind, aboutRef string, minterOf func(gap string) string) string {
	if aboutKind == recordpb.AboutKind_ABOUT_KIND_GAP {
		return minterOf(aboutRef)
	}
	return blueRespondSeat
}

// authorOfGap is authorOf for gap g, read through q: g's subject off its mint, and the minter of the
// gap it is about where it is about one. An unknown gap has no author ("", false).
func authorOfGap(q recordsql.Querier, g string) (string, bool, error) {
	var kind, ref string
	found, err := queryRowAt(q, []any{&kind, &ref},
		`SELECT COALESCE("about_kind", ''), COALESCE("about_ref", '') FROM "mint" WHERE "gap_id" = ?`, g)
	if err != nil || !found {
		return "", false, err
	}
	var minterErr error
	k, _ := AboutKindOf(kind)
	author := authorOf(k, ref, func(gap string) string {
		var seat string
		if _, err := queryRowAt(q, []any{&seat},
			`SELECT e."seat_id" FROM "mint" m JOIN "events" e ON e."id" = m."event_id" WHERE m."gap_id" = ?`, gap); err != nil {
			minterErr = err
		}
		return seat
	})
	return author, true, minterErr
}

// AuthorOfGap is authorOfGap over run's record: the seat whose acts answer gap g and who files on it.
func AuthorOfGap(run Run, g string) (string, error) {
	var out string
	err := readSnapshot(run, func(q recordsql.Querier) error {
		if noRecord(q) {
			return nil
		}
		a, _, err := authorOfGap(q, g)
		out = a
		return err
	})
	return out, err
}

// authorsOn is authorOf over a stream: each gap's author, from the mints written before the read.
// A gap about a gap minted before the gap it names cannot be (the mint refuses an unknown subject).
func authorsOn(evs []*Event) func(gap string) string {
	type subject struct {
		kind recordpb.AboutKind
		ref  string
	}
	about, minter := map[string]subject{}, map[string]string{}
	for _, e := range evs {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok && m.GetGapId() != "" {
			if _, seen := minter[m.GetGapId()]; !seen {
				about[m.GetGapId()] = subject{m.GetAboutKind(), m.GetAboutRef()}
				minter[m.GetGapId()] = e.GetSeatId()
			}
		}
	}
	return func(gap string) string {
		a := about[gap]
		return authorOf(a.kind, a.ref, func(g string) string { return minter[g] })
	}
}

// AuthorsOn is authorsOn for the renderers outside the record: each gap's subject's author, read off
// evs's mints, so a closing renders on its party's side (the author's, or the raiser's).
func AuthorsOn(evs []*Event) func(gap string) string { return authorsOn(evs) }

// ClosingByAuthor reports whether e is a closing its gap's subject's author filed (authorOf), read
// through author; the other party's closing is the raiser's.
func ClosingByAuthor(e *Event, author func(gap string) string) bool {
	c, ok := recordpb.BodyAs[*recordpb.Closing](e)
	return ok && e.GetSeatId() == author(c.GetGapId())
}

// authorAccept is the gap an act's `accepted` takes the raiser's statement of — its `answers` — on
// the bodies the author's --accept G writes; "" where the act accepts nothing.
func authorAccept(body proto.Message) string {
	type accepting interface {
		GetAnswers() string
		GetAccepted() bool
	}
	switch body.(type) {
	case *recordpb.BlueEdit, *recordpb.Retire, *recordpb.Regrade, *recordpb.Amend:
		if a := body.(accepting); a.GetAccepted() {
			return a.GetAnswers()
		}
	}
	return ""
}

// requireAuthorAccept is the author's accept of the raiser's statement (validate): it names the gap
// it answers, its writer is that gap's author (authorOf) — the bench's application in its ruling's
// transaction aside — and no grade-movement docketing stands on the gap — grade movement has no parties, so no party concedes it. Asked of every writer.
func requireAuthorAccept(q recordsql.Querier, seatID string, body proto.Message) error {
	g := authorAccept(body)
	// THE BENCH'S APPLICATION carries the accept's fields — it applies the raiser's statement as the
	// author's accept would — and is the ruling's act, written in its transaction (requireBenchHeld).
	if g == "" || seatclass.RoleOfSeat(seatID) == "bench" {
		return nil
	}
	author, found, err := authorOfGap(q, g)
	if err != nil || !found {
		return err // an unknown gap is the reference check's refusal
	}
	if seatID != author {
		return feov.Errorf(feov.RoleViolation,
			"--accept %s takes its raiser's statement, which is its author's to take: gap %s's author is %s, and you are %s. Nothing was recorded", g, g, author, seatID)
	}
	// A RETIRE TAKES A RETIRE: the raiser's closing states it, of the anchor the gap is about.
	if r, ok := body.(*recordpb.Retire); ok {
		st, _, err := statementAt(q, g)
		if err != nil {
			return err
		}
		if !st.Retire || len(r.GetAnchors()) != 1 || r.GetAnchors()[0] != st.AboutRef {
			return feov.Errorf(feov.Conflict,
				"--accept %s takes the raiser's stated retire of the anchor gap %s is about, and its closing states none: what it states is taken by `edit --accept %s`, or argue it in your closing. Nothing was recorded", g, g, g)
		}
	}
	// A GAP'S WORDING OR GRADES TAKE THE STATEMENT EXACTLY: the raiser's closing on a gap about the
	// author's gap states them, and the author's amend or regrade accepting it carries those bytes.
	switch b := body.(type) {
	case *recordpb.Amend:
		st, _, err := statementAt(q, g)
		if err != nil {
			return err
		}
		if w := st.Wording; w == nil || !sameStrPtr(b.Problem, w.Problem) || !sameStrPtr(b.RequiredFix, w.RequiredFix) || !sameStrPtr(b.AcceptanceCheck, w.AcceptanceCheck) || b.FixNew != nil {
			return feov.Errorf(feov.Conflict,
				"--accept %s takes the wording its raiser's closing states for your gap, exactly — and the amend is not that wording, or the closing states none. Nothing was recorded", g)
		}
	case *recordpb.Regrade:
		st, _, err := statementAt(q, g)
		if err != nil {
			return err
		}
		if st.Grades == nil || !sameAxes(b, st.Grades) {
			return feov.Errorf(feov.Conflict,
				"--accept %s takes the grades its raiser's closing states for your gap, exactly — and the regrade is not those grades, or the closing states none. Nothing was recorded", g)
		}
	}
	var cause sql.NullString
	if _, err := queryRowAt(q, []any{&cause},
		`SELECT d."cause" FROM "gap_state" s JOIN "docketing" d ON d."docketing_id" = s."standing_docketing" WHERE s."gap_id" = ?`, g); err != nil {
		return err
	}
	if cause.String == recordpb.Word(recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT) {
		return feov.Errorf(feov.Conflict,
			"gap %s is docketed for grade movement: the bench keeps or restores its grades, and no party concedes that question, so your accept of its raiser's statement waits until it is ruled. Nothing was recorded", g)
	}
	return nil
}
