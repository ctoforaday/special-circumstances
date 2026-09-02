package fetchcache

// THE SECOND ARTIFACT, AND WHY THE FIRST WAS NOT ENOUGH.
//
// `blue cite` stores sha256(body) — the PDF. The seat read TEXT AN EXTRACTOR PRODUCED. Those
// are different bytes, and only the first was ever kept. So `reproduce` against the source sha
// proves the DOCUMENT is unchanged and says nothing about whether the text a claim came from
// is recoverable: a different extractor, version, page range or OCR setting yields materially
// different text from identical input. "We have the document" is not "we have what was read",
// and only the second supports an audit at the leaf.

// Extractor turns a fetched document into text. It is an interface for the same reason Fetcher
// is: the cache's logic must be testable without a 5 MB WebAssembly module and a compile step,
// and a stub extractor keeps every existing cache test offline and fast.
type Extractor interface {
	// Extract reads body as the given media type. cacheDir is a writable directory the
	// implementation may use for its own compiled-module cache; it is never required to.
	//
	// IT RETURNS NO ERROR, DELIBERATELY. Every outcome — a clean extraction, an encrypted
	// document, a scan with no text layer, a content type nothing extracts — is a FACT ABOUT
	// THE DOCUMENT that belongs on the record, not a failure that should abort the fetch. An
	// error return invites a caller to drop it on the floor with `_`; a field cannot be
	// dropped without deleting a line of code someone has to read.
	Extract(cacheDir, mediaType string, body []byte) Extraction
}

// Extraction is what an Extractor found. The zero value means "nothing was attempted", which
// is what a content type nobody extracts should produce.
type Extraction struct {
	// Text is the extracted, normalized text. Empty means there was none to extract — see
	// Reason, which is then required.
	Text string
	// Title is the document's own title metadata, the first rung of Label's filename chain.
	// Untrusted author text; Label sanitizes it.
	Title string
	// Pages is the document's page count, or 0 where the format has none.
	Pages int
	// Attempted is false when this media type is not one the extractor handles at all. It is
	// the difference between "no text" and "we never looked", and collapsing the two is the
	// [[facts-are-fields]] failure this whole change exists to fix.
	Attempted bool
	// Reason states why Text is empty, whenever Attempted is true and Text is not.
	Reason string
	// ExtractorID is library plus semver (#629 D3) — the key an audit re-runs against.
	ExtractorID string
	// OCRDerived marks text recovered by optical recognition rather than read from a text
	// layer, so its weaker reproducibility is declared rather than discovered.
	OCRDerived bool
}

// DefaultExtractor is the process-wide Extractor, a variable for exactly the reason Default is:
// the command tree is built once, so a test pins behaviour by swapping this and restoring it
// with a defer.
//
// THE REAL ONE IS THE DEFAULT, and that placement is the point. The obvious alternative — a
// no-op default that the binary's main replaces — puts a silent, correct-looking failure one
// forgotten wiring line away: a no-op returns Attempted:false, which records "nobody looked",
// which is exactly what a fetch of a plain HTML page legitimately records. The miss and the
// honest zero would again be the same bytes. So the engine ships wired, and a TEST is what has
// to do the swapping.
var DefaultExtractor Extractor = PDFExtractor{}
