package flags

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// A FLAG STATED TWICE IS READ BY THE PARSER'S OWN GRAMMAR, and only a flag that says it appends is let
// repeat. Each row is a shape of argv pflag reads in its own way: the scan must agree with the parse
// that follows, or it refuses a call that states a flag once, or misses one that states it twice.
func TestFirstRepeatReadsAnArgvAsTheParserDoes(t *testing.T) {
	tree := func() *cobra.Command {
		root := &cobra.Command{Use: "root"}
		root.PersistentFlags().String("inherited", "", "")
		leaf := &cobra.Command{Use: "leaf", Run: func(*cobra.Command, []string) {}}
		leaf.Flags().StringP("single", "s", "", "")
		leaf.Flags().Bool("on", false, "")
		leaf.Flags().Int("count", 0, "")
		leaf.Flags().Var(new(CSV), "list", "")
		leaf.Flags().StringSlice("appends", nil, "")
		leaf.Flags().StringArray("repeats", nil, "")
		root.AddCommand(leaf)
		return leaf
	}
	for _, c := range []struct {
		name        string
		args        []string
		flag        string
		first, then string
	}{
		{"spaced", []string{"--single", "a", "--single", "b"}, "single", "a", "b"},
		{"inline, then spaced", []string{"--single=a", "--single", "b"}, "single", "a", "b"},
		{"the same value twice", []string{"--single", "a", "--single", "a"}, "single", "a", "a"},
		{"a shorthand and the long form", []string{"-s", "a", "--single", "b"}, "single", "a", "b"},
		{"a boolean twice", []string{"--on", "--on"}, "on", "true", "true"},
		{"a boolean stated two ways", []string{"--on", "--on=false"}, "on", "true", "false"},
		{"an int", []string{"--count", "1", "--count", "2"}, "count", "1", "2"},
		{"a parent's persistent flag", []string{"--inherited", "a", "--inherited", "b"}, "inherited", "a", "b"},
		{"a comma-separated list holds one list", []string{"--list", "a", "--list", "b"}, "list", "a", "b"},
		{"past a flag this verb does not have", []string{"--nonesuch", "--single", "a", "--single", "b"}, "single", "a", "b"},
		{"past --help", []string{"--help", "--single", "a", "--single", "b"}, "single", "a", "b"},
		{"one statement of each", []string{"--single", "a", "--on", "--count", "1"}, "", "", ""},
		{"a value that reads as the flag", []string{"--single", "--single"}, "", "", ""},
		{"a list type that appends", []string{"--appends", "a", "--appends", "b"}, "", "", ""},
		{"a list type that repeats", []string{"--repeats", "a", "--repeats", "b"}, "", "", ""},
		{"words past the end of the flags", []string{"--single", "a", "--", "--single", "b"}, "", "", ""},
		{"nothing", nil, "", "", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			leaf := tree()
			got := FirstRepeat(leaf, c.args)
			if c.flag == "" {
				if got != nil {
					t.Fatalf("%q was read as repeating --%s (%q, %q)", c.args, got.Flag.Name, got.First, got.Then)
				}
				return
			}
			if got == nil || got.Flag.Name != c.flag || got.First != c.first || got.Then != c.then {
				t.Fatalf("%q read as %+v, want --%s %q then %q", c.args, got, c.flag, c.first, c.then)
			}
			// THE SCAN SETS NOTHING: the parse that follows owns the flag's state.
			if got.Flag.Changed {
				t.Errorf("the scan marked --%s as set", c.flag)
			}
		})
	}
}

// The refusal holds both values, and only below its first line, which the refusal log keeps.
func TestARepeatNamesBothValuesBelowItsFirstLine(t *testing.T) {
	for _, c := range []struct {
		name string
		reg  func(*cobra.Command)
		do   string
	}{
		{"a single value", func(c *cobra.Command) { c.Flags().String("x", "", "") }, "pass it once"},
		{"a list", func(c *cobra.Command) { c.Flags().Var(new(CSV), "x", "") }, "comma-separated list"},
	} {
		t.Run(c.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "c"}
			c.reg(cmd)
			said := FirstRepeat(cmd, []string{"--x", "first words", "--x", "other words"}).Said()
			head, rest, _ := strings.Cut(said, "\n")
			if !strings.Contains(head, "--x") || !strings.Contains(head, c.do) {
				t.Errorf("the first line does not name the flag and %q: %s", c.do, head)
			}
			for _, value := range []string{`"first words"`, `"other words"`} {
				if strings.Contains(head, value) {
					t.Errorf("the first line carries %s, which the log would keep: %s", value, head)
				}
				if !strings.Contains(rest, value) {
					t.Errorf("the refusal does not name %s: %s", value, said)
				}
			}
		})
	}
}

// AN ID LIST TAKES EACH ITEM AS ITS SHAPE DOES, and stores what the shape stores.
func TestAnIDListHoldsEachItemToItsShape(t *testing.T) {
	findings := IDList(FindingID)
	if err := findings.Set("<!--fx:F-1d8a5455-->, F-00000001"); err != nil {
		t.Fatal(err)
	}
	if got := findings.Value(); len(got) != 2 || got[0] != "F-1d8a5455" || got[1] != "F-00000001" {
		t.Errorf("stored %q, want the bare ids", got)
	}
	for name, c := range map[string]struct {
		list  *CSV
		value string
	}{
		"not an id":                  {IDList(GapID), "banana"},
		"one good, one not":          {IDList(GapID), "G-00000001,banana"},
		"another kind's id":          {IDList(GapID), "F-00000001"},
		"a token the shape does not": {IDList(GapID), "<!--gap:G-00000001-->"},
		"the tagged text":            {IDList(FindingID), "fx:F-00000001"},
		"an empty item":              {IDList(GapID), "G-00000001,,G-00000002"},
	} {
		if err := c.list.Set(c.value); err == nil {
			t.Errorf("%s: %q was accepted", name, c.value)
		}
	}
	empty := IDList(GapID)
	if err := empty.Set(""); err != nil || !empty.Given() || len(empty.Value()) != 0 {
		t.Errorf("the list with no items, stated: err %v, given %v, value %q", err, empty.Given(), empty.Value())
	}
}
