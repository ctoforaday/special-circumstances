package migrate

import (
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// planIDPhrases is every phrase after which a stored plan's `why` line prints an events.id, one
// space before it: one or more whole tokens, a single space between each. These describe ARCHIVED
// bytes, so they are migrate's own table rather than the live verb's: a later rewording is a row
// added here for the records printed after it, the old row kept for the records printed before. The
// format strings each names (TestEveryIDTheVerbPrintsIsTranslated renders every one):
//   - head: `never sat — audits the report at head %d`, `has sat through no sitting — … — audits the
//     report at head %d`, `head %d moved past its last sitting's pin %d`; from epoch 42 `… was served
//     the whole report, behind head %d` and `… at its pin %d, behind head %d` (retirement.go, the lens
//     fold's readiness);
//   - pin: `sat at the head (pin %d)`, `retired (pin %d)` and the lines above; from epoch 42 `served
//     the report at the head (pin %d)` (retirement.go);
//   - opened at: `its sitting opened at %d %s, so it counts neither as sat nor toward retirement`
//     (retirement.go, unsatSitting), in the lens fold's `has sat through no sitting` and `head moved`
//     lines. Two tokens: a bare `at` would read the count in `it retires at %d` as an id;
//   - edit, event: `changed at edit %d … its last check is event %d` (passblockers.go, reopenedFact);
//   - contradiction: `contradiction %s` among a lens's debts (retirement.go, owedBlocker) and
//     `contradiction %s` of a contradiction blocker (passblockers.go, contradictionFact).
var planIDPhrases = []string{"head", "pin", "opened at", "edit", "event", "contradiction"}

// afterIDPhrase reports whether the numeral at line[i:] follows one of planIDPhrases and one space,
// the phrase beginning at a token boundary: `ahead 5` and `spin 5` are not `head 5` and `pin 5`.
func afterIDPhrase(line string, i int) bool {
	for _, p := range planIDPhrases {
		start := i - len(p) - 1
		if start < 0 || line[start:i] != p+" " {
			continue
		}
		if start == 0 || !isWordByte(line[start-1]) {
			return true
		}
	}
	return false
}

// remandDirectionPrefix is what an archived plan's remand line prints before the bench ruling's
// direction (record.RemandDirectionPrefix, compared by TestMigrateKnowsTheVerbsRemandPrefix). The
// direction is printed last on its line and is the bench's own words, so the rest of the line after
// it is seat prose.
const remandDirectionPrefix = "on the ruling's direction: "

// TranslatePlanIDs carries the events.ids a stored plan names through newID — the events.id each
// source event that wrote one landed at — and re-keys and re-prints the plan when any moved. A plan
// whose ids all kept their numbers is left untouched, its bytes the bytes stored. The ids are the
// plan's head, each stale area's pin, a contradiction blocker's subject, and the numbers its `why`
// lines print after an id phrase (planIDPhrases). Seat words are carried as written: a Go-quoted
// span (a claim) and a remand line's tail after remandDirectionPrefix (the bench's direction), and
// RemandOwed's direction is not read. An id that wrote no event, and a moved id's number printed
// outside every id phrase, are refusals: the first has nothing to name, and the second cannot be
// told from a count.
func TranslatePlanIDs(pl *recordpb.Plan, newID map[int64]int64) error {
	p, _, err := record.ReadPlan([]byte(pl.GetPrinted()))
	if err != nil {
		return fmt.Errorf("this stored plan does not read as a plan: %v", err)
	}
	changed := false
	to := func(n int64) (int64, error) {
		if n == 0 {
			return 0, nil
		}
		m, landed := newID[n]
		if !landed {
			return 0, fmt.Errorf("this stored plan names event %d, which wrote no event on the migrated record before it", n)
		}
		if m != n {
			changed = true
		}
		return m, nil
	}
	if p.Head, err = to(p.Head); err != nil {
		return err
	}
	for i := range p.StaleAreas {
		if p.StaleAreas[i].Pin, err = to(p.StaleAreas[i].Pin); err != nil {
			return err
		}
	}
	for i, b := range p.Blockers {
		if b.Kind != record.BlockerContradiction {
			continue
		}
		n, perr := strconv.ParseInt(b.Subject, 10, 64)
		if perr != nil {
			return fmt.Errorf("this stored plan's contradiction blocker names %q, which is not an events.id", b.Subject)
		}
		m, err := to(n)
		if err != nil {
			return err
		}
		p.Blockers[i].Subject = strconv.FormatInt(m, 10)
	}
	for i, line := range p.Why {
		out, err := translateWhyLine(line, newID, to)
		if err != nil {
			return err
		}
		p.Why[i] = out
	}
	if !changed {
		return nil
	}
	p = p.WithKey()
	pl.Printed = proto.String(p.Printed())
	return nil
}

// isWordByte is the class a numeral is bounded by: a numeral is a maximal run of these that is all
// digits, so the hex inside a content id (`G-7b637cca`) and a seat id's ordinal (`red-lens-logic-2`)
// are never numerals.
func isWordByte(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' || c == '-'
}

func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// translateWhyLine scans line left to right: a Go-quoted span is skipped whole, the first
// remandDirectionPrefix outside one ends the scan with the rest carried as written, a numeral one
// space after an id phrase (afterIDPhrase) goes through to, and any other numeral that is a moved id
// is a refusal.
func translateWhyLine(line string, newID map[int64]int64, to func(int64) (int64, error)) (string, error) {
	var b strings.Builder
	for i := 0; i < len(line); {
		switch c := line[i]; {
		case c == '"':
			q, err := strconv.QuotedPrefix(line[i:])
			if err != nil {
				return "", fmt.Errorf("this stored plan's line %q holds a quote that does not close — the translation cannot tell the seat's words from the verb's", line)
			}
			b.WriteString(q)
			i += len(q)
		case strings.HasPrefix(line[i:], remandDirectionPrefix):
			b.WriteString(line[i:])
			return b.String(), nil
		case isWordByte(c):
			j := i
			for j < len(line) && isWordByte(line[j]) {
				j++
			}
			tok := line[i:j]
			if allDigits(tok) {
				n, perr := strconv.ParseInt(tok, 10, 64)
				phrase := afterIDPhrase(line, i)
				switch {
				case perr != nil:
					// Too long to be an events.id: a count or a word, carried.
				case phrase:
					m, err := to(n)
					if err != nil {
						return "", err
					}
					tok = strconv.FormatInt(m, 10)
				default:
					if m, landed := newID[n]; landed && m != n {
						return "", fmt.Errorf("this stored plan's line %q prints %d, a moved event's number, outside every id phrase; the translation cannot tell whether it names that event", line, n)
					}
				}
			}
			b.WriteString(tok)
			i = j
		default:
			b.WriteByte(c)
			i++
		}
	}
	return b.String(), nil
}
