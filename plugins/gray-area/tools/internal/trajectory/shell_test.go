package trajectory

import (
	"regexp"
	"testing"
)

// The measured problem: seats alias the binary to a shell variable, and a naive
// matcher misses those invocations. Both shapes below were taken from a real
// transcript produced by driving a live session, not invented.
func TestFindInvocationsResolvesAliases(t *testing.T) {
	cases := []struct {
		name    string
		command string
		want    []Invocation
	}{
		{
			name:    "direct call",
			command: `./feov-record merge --seat-id red-1`,
			want:    []Invocation{{Binary: "feov-record", Verb: "merge"}},
		},
		{
			name:    "aliased through a variable — the case a grep misses",
			command: `REC=./feov-record ; "$REC" finding --seat-id red-1`,
			want:    []Invocation{{Binary: "feov-record", Verb: "finding", Aliased: true, Via: "REC"}},
		},
		{
			name:    "braced expansion",
			command: `B=/usr/local/bin/feov-record && ${B} close --gap G1`,
			want:    []Invocation{{Binary: "feov-record", Verb: "close", Aliased: true, Via: "B"}},
		},
		{
			name:    "assignment prefixed to the call on one segment",
			command: `FOO=1 feov-record status`,
			want:    []Invocation{{Binary: "feov-record", Verb: "status"}},
		},
		{
			name:    "two calls in one command",
			command: `feov-record a; ./feov-record b`,
			want: []Invocation{
				{Binary: "feov-record", Verb: "a"},
				{Binary: "feov-record", Verb: "b"},
			},
		},
		{
			name:    "windows-style path still names the binary",
			command: `C:\tools\feov-record.exe mint`,
			want:    []Invocation{{Binary: "feov-record", Verb: "mint"}},
		},
		{
			name:    "flags skipped when finding the verb",
			command: `feov-record --seat-id red-1 --run r2 finding`,
			want:    []Invocation{{Binary: "feov-record", Verb: "finding"}},
		},
		{
			name:    "flag with = does not swallow the verb",
			command: `feov-record --seat-id=red-1 finding`,
			want:    []Invocation{{Binary: "feov-record", Verb: "finding"}},
		},
		{
			name:    "downstream of a pipe",
			command: `cat x.json | feov-record ingest`,
			want:    []Invocation{{Binary: "feov-record", Verb: "ingest"}},
		},
		{
			name:    "a different binary is not ours",
			command: `git commit -m "feov-record finding"`,
			want:    nil,
		},
		{
			name:    "unresolvable variable yields nothing rather than a guess",
			command: `"$UNKNOWN" finding`,
			want:    nil,
		},
		{
			name:    "variable pointing elsewhere is not our binary",
			command: `X=/bin/ls ; "$X" finding`,
			want:    nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FindInvocations(tc.command, "feov-record")
			if len(got) != len(tc.want) {
				t.Fatalf("got %d invocations %+v, want %d", len(got), got, len(tc.want))
			}
			for i := range got {
				w := tc.want[i]
				if got[i].Binary != w.Binary || got[i].Verb != w.Verb ||
					got[i].Aliased != w.Aliased || got[i].Via != w.Via {
					t.Errorf("[%d] got %+v, want %+v", i, got[i], w)
				}
				if got[i].Raw == "" {
					t.Errorf("[%d] invocation lost the segment it came from", i)
				}
			}
		})
	}
}

// A naive matcher is what this exists to beat; assert the gap is real so the
// test explains why the shell handling is worth its weight.
func TestAliasedCallIsInvisibleToNaiveMatching(t *testing.T) {
	const cmd = `REC=./feov-record ; "$REC" finding --seat-id red-1`
	naive := regexpMustFind(`feov-record\s+([a-z-]+)`, cmd)
	if naive == "finding" {
		t.Fatal("a naive matcher would have found the verb; the alias case is not what we think")
	}
	got := FindInvocations(cmd, "feov-record")
	if len(got) != 1 || got[0].Verb != "finding" {
		t.Fatalf("resolver failed the case that motivates it: %+v", got)
	}
}

func regexpMustFind(pattern, s string) string {
	m := regexp.MustCompile(pattern).FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}
