package catalogue

import (
	"context"
	"strings"
	"testing"
	"time"
)

// THE IDS THIS TOOL PRINTS MUST BE IDS IT ACCEPTS. `agents` and `find` print Short; a lookup that
// took only the full id answered "may predate capture" about a session the store held.
func TestResolveSessionTakesTheFullIDOrAUniquePrefix(t *testing.T) {
	db := store(t)
	at := time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)
	for _, sid := range []string{"5627f39a-fa4c", "5627f39b-0000", "5627f39b-0000-tail", "c2a2ecb6-8430"} {
		if err := RegisterSession(db, sid, "/p/-w", "", at); err != nil {
			t.Fatal(err)
		}
	}
	ctx := context.Background()
	for _, tc := range []struct{ in, want, errHas string }{
		{"5627f39a-fa4c", "5627f39a-fa4c", ""},
		{"5627f39a", "5627f39a-fa4c", ""}, // Short's form
		{"c2a2", "c2a2ecb6-8430", ""},
		// A full id that is also a prefix of another is still that session, not an ambiguity.
		{"5627f39b-0000", "5627f39b-0000", ""},
		{"5627f39", "", "prefix of 3 sessions"},
		{"5627f39b", "", "prefix of 2 sessions"},
		{"ffffffff", "", "no session ffffffff in the store, by full id or prefix"},
		{"", "", "empty session id"},
		// substr, not LIKE: a wildcard in the input is a character, never a pattern.
		{"%", "", "no session % in the store"},
	} {
		got, err := ResolveSession(ctx, db, tc.in)
		if tc.errHas == "" {
			if err != nil || got != tc.want {
				t.Errorf("ResolveSession(%q) = %q, %v; want %q", tc.in, got, err, tc.want)
			}
			continue
		}
		if err == nil || !strings.Contains(err.Error(), tc.errHas) {
			t.Errorf("ResolveSession(%q) = %q, %v; want an error containing %q", tc.in, got, err, tc.errHas)
		}
	}
}
