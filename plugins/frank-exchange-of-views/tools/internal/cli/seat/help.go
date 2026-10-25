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
// ("name the nearest class when none fits", "an unruled motion blocks a PASS") or tells it how to succeed ("quote
// enough context that the span matches exactly one place").
//
// That distinction is a judgement, not a gate: no test can tell a useful precondition from a
// redundant warning, and pretending otherwise would be a check that fires on the wrong half.
//
// ONE PAGE MAY CARRY ANOTHER'S TEXT, BY INCLUDING IT. Every document is a named template in one
// set, keyed as its verb is, so `{{template "register" .}}` renders register's whole document into
// the page that names it. A keyed variant that says everything its base says and more — blue's
// register, which alone carries the repair flag — includes the base and adds its own paragraphs.
// Each page still teaches its verb on its own, and the shared text has one source, so the two
// pages cannot drift. The include lands the base's sections in place, so the variant's own text
// continues whichever section the base ends in; an addition that lands in the menu makes the menu
// more than one line, which is a panic below.
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
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
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
// `verdict` lists the PASS blockers by rendering record.PassBlockerGlosses — the list comes from
// the table that governs it, so changing the table changes the sentence. A hand-typed list in the
// file would sit beside a table free to move, which is the two-copies shape this repository keeps
// deleting. A document names {{.PassBlockerKinds}} instead.
//
// missingkey=error, so a placeholder with no value is a PANIC and never renders "<no value>" into
// a menu line where a seat would read it as the tool's considered wording.
var HelpValues = map[string]any{}

type helpDoc struct{ menu, detail string }

// RAW AT INIT, RENDERED AT USE. Rendering here would run before any other package's init had
// contributed to HelpValues — measured: the first version panicked on a placeholder because
// package seat initializes before the package that supplies it. The command tree is built after
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

// helpSet is every document parsed as a template named by its key, so one document can include
// another. Built on first render, under helpMu.
var helpSet *template.Template

// render substitutes the values and documents a document may name, and refuses one it cannot
// resolve.
func render(key string) string {
	if helpSet == nil {
		set := template.New("").Option("missingkey=error")
		for k, src := range helpSrc {
			if _, err := set.New(k).Parse(src); err != nil {
				panic("seat: help/" + k + ".md: " + err.Error())
			}
		}
		helpSet = set
	}
	var b strings.Builder
	if err := helpSet.ExecuteTemplate(&b, key, HelpValues); err != nil {
		panic("seat: help/" + key + ".md: " + err.Error() +
			" — a help document names a value or a document nothing supplies, and rendering it blank would put" +
			" the tool's name on a sentence with a hole in it")
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
			return d, feov.Errorf(feov.Internal, "unknown section %q — the sections are `## menu` and `## detail`", strings.TrimSpace(line))
		}
		if cur != nil {
			*cur += line + "\n"
		}
	}
	d.menu, d.detail = strings.TrimSpace(d.menu), strings.TrimSpace(d.detail)
	if d.menu == "" {
		return d, feov.Errorf(feov.Internal, "no `## menu` section, or it is empty — this is the line a seat reads while deciding, and a blank one is indistinguishable from a verb with nothing to say")
	}
	if d.detail == "" {
		return d, feov.Errorf(feov.Internal, "no `## detail` section, or it is empty — a verb whose long page adds nothing to its menu line teaches a seat that opening the page is not worth doing")
	}
	if strings.Contains(d.menu, "\n") {
		return d, feov.Errorf(feov.Internal, "the menu is %d lines — it is ONE line, printed inside a listing", strings.Count(d.menu, "\n")+1)
	}
	if len(d.menu) > MenuLimit {
		return d, feov.Errorf(feov.Internal, "the menu is %d characters, over the %d limit — cobra prints it in full in the parent listing, so every character is paid by every seat that opens it", len(d.menu), MenuLimit)
	}
	return d, nil
}

// LoadHelp resolves a document and returns its two fields, or an error naming what is wrong. It is
// the testable form of helpFor: the gates live in package cli, where the whole command tree can be
// built and every package's init has supplied its HelpValues.
func LoadHelp(key string) (menu, detail string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = feov.Errorf(feov.Internal, "%v", r)
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
	if _, ok := helpSrc[name]; !ok {
		panic("seat: no help/" + name + ".md — every verb owns a help document, and a verb without one would render a blank entry in the listing that a seat cannot distinguish from a verb with nothing to say")
	}
	d, err := parseHelp(render(name))
	if err != nil {
		panic(fmt.Sprintf("help/%s.md: %v", name, err))
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
