package report

import (
	"strings"
	"testing"

	"github.com/dop251/goja"
	"golang.org/x/net/html"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// sitePageDOM is the least of a browser site.js reaches at load and on a click, standing in for the
// page's elements: two articles, a heading in the second, and what the script reads of the window.
// The test reads back the hash the script writes and where it scrolled.
const sitePageDOM = `
var scrolledTop = false, scrolledTo = null, hash = null, listeners = {};
function el(props) {
  var e = { dataset: {}, hidden: false, children: [], value: '',
    classList: { toggle: function () {}, remove: function () {}, add: function () {} },
    setAttribute: function () {}, addEventListener: function () {},
    querySelectorAll: function () { return []; },
    getBoundingClientRect: function () { return { top: 0 }; },
    scrollIntoView: function () { scrolledTo = this.id; } };
  for (var k in props) { e[k] = props[k]; }
  return e;
}
var articles = [el({ id: 'doc-report', dataset: { doc: 'report.md' } }), el({ id: 'doc-docket', dataset: { doc: 'docket.md' }, hidden: true })];
var byId = { 'doc-report': articles[0], 'doc-docket': articles[1], 'g-1': el({ id: 'g-1' }), filter: el({}) };
var document = {
  body: el({}),
  querySelectorAll: function (q) { return q === 'main article' ? articles : []; },
  getElementById: function (id) { return byId[id] || null; },
  addEventListener: function (t, f) { (listeners[t] = listeners[t] || []).push(f); },
  createElement: function () { return el({}); }
};
var window = { addEventListener: function () {}, scrollTo: function () { scrolledTop = true; } };
var history = { replaceState: function (a, b, h) { hash = h; } };
var location = { hash: '' };
function requestAnimationFrame() {}
function click(doc, href) {
  var a = { dataset: { doc: doc }, getAttribute: function () { return href; } };
  var e = { target: { closest: function (sel) { return sel === 'a.idref' ? a : null; } }, preventDefault: function () {} };
  listeners.click.forEach(function (f) { f(e); });
}
`

// A LINK TO A DOCUMENT OPENS ITS TAB AT THE TOP. The renderer writes it as a link to the document's
// own article — an empty fragment is no URL to the allowlist — and site.js takes that for the tab
// with no place in it, as it took the empty fragment before: the page scrolls to the top and the
// address names the document alone. A link to a place in a document still scrolls there.
func TestALinkToADocumentOpensItsTabAtTheTop(t *testing.T) {
	page := RenderSite(Ship("# T", []Doc{
		{File: FileReport, Nav: "Report", Body: "See [the board](docket.md) and [a gap](docket.md#g-1).\n"},
		{File: FileDocket, Nav: "Board", Body: "## The board\n"},
	}), record.Family{})
	type link struct{ doc, href string }
	var links []link
	for _, tok := range tokens(mainOf(page)) {
		if tok.Type != html.StartTagToken || tok.Data != "a" {
			continue
		}
		var l link
		for _, a := range tok.Attr {
			switch a.Key {
			case "data-doc":
				l.doc = a.Val
			case "href":
				l.href = a.Val
			}
		}
		links = append(links, l)
	}
	if want := []link{{"docket.md", "#" + docElement(FileDocket)}, {"docket.md", "#g-1"}}; len(links) != 2 || links[0] != want[0] || links[1] != want[1] {
		t.Fatalf("the page's links to the board are %+v, want %+v", links, want)
	}
	for _, c := range []struct {
		link
		hash, scrolledTo string
		top              bool
	}{
		{links[0], "#docket.md", "", true},
		{links[1], "#docket.md/g-1", "g-1", false},
	} {
		vm := goja.New()
		if _, err := vm.RunString(sitePageDOM); err != nil {
			t.Fatal(err)
		}
		if _, err := vm.RunString(strings.Replace(siteJS, "MERMAID_CDN_URL", mermaidCDN, 1)); err != nil {
			t.Fatalf("site.js: %v", err)
		}
		if _, err := vm.RunString("click(" + jsString(c.doc) + ", " + jsString(c.href) + ")"); err != nil {
			t.Fatalf("click: %v", err)
		}
		hash, _ := vm.Get("hash").Export().(string)
		to, _ := vm.Get("scrolledTo").Export().(string)
		top := vm.Get("scrolledTop").ToBoolean()
		shown := !vm.Get("articles").ToObject(vm).Get("1").ToObject(vm).Get("hidden").ToBoolean()
		if hash != c.hash || to != c.scrolledTo || top != c.top || !shown {
			t.Errorf("a click on %+v: address %q, scrolled to %q, to the top %v, board shown %v; want %q, %q, %v, true", c.link, hash, to, top, shown, c.hash, c.scrolledTo, c.top)
		}
	}
}

// jsString is s as a JavaScript string literal.
func jsString(s string) string {
	return "'" + strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(s) + "'"
}
