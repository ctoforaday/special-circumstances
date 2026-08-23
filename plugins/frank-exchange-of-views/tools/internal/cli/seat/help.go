package seat

// THE HELP IS A DOCUMENT, NOT A GO STRING.
//
// Every verb's help lives in help/<verb>.md and is embedded. Three reasons, in order of weight:
//
//	IT IS PROSE FOR A READER. Written as a string literal it acquired backtick escaping, `+`
//	concatenation across five lines, and a shape nobody could review as writing. `finding` ran to
//	961 characters of that.
//	IT IS WHERE THE PROMPTS' PROSE IS GOING. Moving a paragraph out of debate.js into a markdown
//	file is an edit; moving it into a quoted Go literal is a translation, and translations are
//	where content quietly changes.
//	TWO FIELDS, NAMED. `## menu` is the line a seat reads while DECIDING; `## detail` is the page
//	it reads once it has chosen. One string set into both Short and Long is what made the listing
//	unreadable and the long page redundant.
//
// WHAT THE DETAIL IS NOT FOR: restating what the tool enforces at the write. A typed flag declares
// what it is checked against — flags.GapID().WithCheck(record.GapExists) — and the refusal names the
// verb, the flag, the value, how to list the real ones, and how to create one. A sentence in the
// help saying "this is refused if the id does not exist" arrives at the same moment as that refusal
// and says less. What EARNS its place is anything that changes what a seat does BEFORE the call
// ("coin the class first", "an unruled motion blocks a PASS") or tells it how to succeed ("quote
// enough context that the span matches exactly one place").
//
// That distinction is a judgement, not a gate: no test can tell a useful precondition from a
// redundant warning, and pretending otherwise would be a check that fires on the wrong half.
//
// A FILE LOOKED UP BY NAME IS A PATTERN STANDING IN FOR A SCHEMA, so every seam here is loud. A
// missing file, a missing section, an empty section, or a file no verb claims is a PANIC at
// construction — these commands are built at process start, so the binary fails immediately and
// visibly rather than shipping a blank menu entry. A verb whose help silently rendered empty would
// be indistinguishable from a verb with nothing to say, and the seat reading that listing would
// have no way to tell.

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"sync"
	"text/template"
)

//go:embed help/*.md
var helpFS embed.FS

// MenuLimit is how long a menu line may be.
//
// It is a LIMIT, not a style note. Cobra prints Short IN FULL inside the parent's Available
// Commands block, so every character here is paid by every seat that opens that listing, whether
// or not it cares about this verb. The measured failure is a seat reading the first clause of a
// paragraph and stopping; a line that fits on a line cannot produce it.
const MenuLimit = 120

// HelpValues are the facts a help document may RENDER rather than restate.
//
// `near-match` said "returns the top 5 gaps" by interpolating nearMatchTopN — the number came from
// the constant that governs it, so changing the constant changed the sentence. Moving the help into
// a file would have frozen a hand-typed 5 beside a constant free to move, which is the two-copies
// shape this repository keeps deleting. A document names {{.NearMatchTopN}} instead.
//
// missingkey=error, so a placeholder with no value is a PANIC and never renders "<no value>" into
// a menu line where a seat would read it as the tool's considered wording.
var HelpValues = map[string]any{}

type helpDoc struct{ menu, detail string }

// RAW AT INIT, RENDERED AT USE. Rendering here would run before any other package's init had
// contributed to HelpValues — measured: the first version panicked on {{.NearMatchTopN}} because
// package seat initializes before package merge, which supplies it. The command tree is built after
// every init has run, so resolution happens there.
var helpSrc = mustLoadHelp()

// GUARDED, BECAUSE TREES ARE BUILT CONCURRENTLY. The rendered cache and the claim set are written
// on first use of each verb, and a command tree is constructed per invocation — the tests build one
// per role in parallel and hit `concurrent map read and map write` immediately. The shipped binary
// builds one tree per process today and would not have shown this, which is precisely why it is
// worth fixing rather than noting: the next caller that builds two is a server, and it would find
// out in production.
var (
	helpMu    sync.Mutex
	helpCache = map[string]helpDoc{}
)

func mustLoadHelp() map[string]string {
	out := map[string]string{}
	ents, err := fs.ReadDir(helpFS, "help")
	if err != nil {
		panic("seat: cannot read embedded help: " + err.Error())
	}
	for _, e := range ents {
		name := strings.TrimSuffix(e.Name(), ".md")
		b, err := helpFS.ReadFile("help/" + e.Name())
		if err != nil {
			panic("seat: cannot read help/" + e.Name() + ": " + err.Error())
		}
		out[name] = string(b)
	}
	return out
}

// render substitutes the values a document may name, and refuses one it cannot resolve.
func render(file, src string) string {
	t, err := template.New(file).Option("missingkey=error").Parse(src)
	if err != nil {
		panic("seat: help/" + file + ": " + err.Error())
	}
	var b strings.Builder
	if err := t.Execute(&b, HelpValues); err != nil {
		panic("seat: help/" + file + ": " + err.Error() +
			" — a help document names a value nothing supplies, and rendering it blank would put the tool's" +
			" name on a sentence with a hole in it")
	}
	return b.String()
}

// parseHelp reads the two named sections. Position is not the key: a document that reordered them,
// or grew a third, still resolves — and one missing either is an error rather than a blank.
func parseHelp(src string) (helpDoc, error) {
	var d helpDoc
	var cur *string
	for _, line := range strings.Split(src, "\n") {
		switch strings.TrimSpace(line) {
		case "## menu":
			cur = &d.menu
			continue
		case "## detail":
			cur = &d.detail
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "## ") {
			return d, fmt.Errorf("unknown section %q — the sections are `## menu` and `## detail`", strings.TrimSpace(line))
		}
		if cur != nil {
			*cur += line + "\n"
		}
	}
	d.menu, d.detail = strings.TrimSpace(d.menu), strings.TrimSpace(d.detail)
	if d.menu == "" {
		return d, fmt.Errorf("no `## menu` section, or it is empty — this is the line a seat reads while deciding, and a blank one is indistinguishable from a verb with nothing to say")
	}
	if d.detail == "" {
		return d, fmt.Errorf("no `## detail` section, or it is empty — a verb whose long page adds nothing to its menu line teaches a seat that opening the page is not worth doing")
	}
	if strings.Contains(d.menu, "\n") {
		return d, fmt.Errorf("the menu is %d lines — it is ONE line, printed inside a listing", strings.Count(d.menu, "\n")+1)
	}
	if len(d.menu) > MenuLimit {
		return d, fmt.Errorf("the menu is %d characters, over the %d limit — cobra prints it in full in the parent listing, so every character is paid by every seat that opens it", len(d.menu), MenuLimit)
	}
	return d, nil
}

// LoadHelp resolves a document and returns its two fields, or an error naming what is wrong. It is
// the testable form of helpFor: the gates live in package cli, where the whole command tree can be
// built and every package's init has supplied its HelpValues.
func LoadHelp(key string) (menu, detail string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	d := helpFor(key)
	return d.menu, d.detail, nil
}

// helpFor resolves a verb's document, or panics naming what is missing.
func helpFor(name string) helpDoc {
	helpMu.Lock()
	defer helpMu.Unlock()
	if d, ok := helpCache[name]; ok {
		return d
	}
	src, ok := helpSrc[name]
	if !ok {
		panic("seat: no help/" + name + ".md — every verb owns a help document, and a verb without one would render a blank entry in the listing that a seat cannot distinguish from a verb with nothing to say")
	}
	d, err := parseHelp(render(name+".md", src))
	if err != nil {
		panic(fmt.Sprintf("seat: help/%s.md: %v", name, err))
	}
	helpCache[name] = d
	registeredKeys[name] = true
	return d
}

// registeredKeys records which help key each constructed verb claimed, so a test can check the
// bijection in the direction a panic cannot: an ORPHAN document renders nowhere, and to anyone
// auditing the help it is indistinguishable from a verb whose page was never written.
var registeredKeys = map[string]bool{}

// RegisteredHelpKeys is every key a verb has claimed. It is only meaningful after the command tree
// has been built, which is why the test that reads it builds one first.
func RegisteredHelpKeys() []string {
	helpMu.Lock()
	defer helpMu.Unlock()
	var out []string
	for k := range registeredKeys {
		out = append(out, k)
	}
	return out
}

// HelpSourceFor reports whether a document exists for a key, for the orphan gate.
func HelpSourceFor(key string) (string, bool) { src, ok := helpSrc[key]; return src, ok }

// HelpNames is every verb a help document exists for, so a test can check the bijection against the
// verbs actually registered — a document nobody claims is as much a defect as a verb with none.
func HelpNames() []string {
	var out []string
	for k := range helpSrc {
		out = append(out, k)
	}
	return out
}
