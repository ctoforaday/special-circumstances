# The page corpus

Scans that BEAT the reader, kept as fixtures. One directory per page.

## What enters

gblock's rule, 2026-09-18: **keep the complex one-shots.** A page enters because it broke
something, or because it is the only specimen of a kind the corpus lacks. A page that repeats a
class already here does not enter — `defect_class` names what each one is a specimen OF, and the
admission argument belongs in review.

Every failing image is an opportunity: the net widens each round, and `STATUS.md` is the meter.

## What a directory holds

| file | what it is |
|---|---|
| `page.pdf.gz` | one page, pixels only, gzipped. The publisher's own image bytes; the text layer is removed so the reader cannot extract its way around them |
| `provenance.json` | where it came from, what we may do with it, and `expect` — what a CORRECT reader would produce |
| `README.md` | GENERATED from the record |
| `reading.golden` | what the reader produces today, byte for byte, including where that is wrong |

## The rules the gates enforce

- **Rights are a closed set.** `corpus.Rights` lists them. A page whose licence cannot be named in
  that set is not committed; it goes in `references.json` with the reason, and `REFERENCES.md` is
  generated from that.
- **Nothing is unaccounted for.** An empty corpus fails, a directory missing one of its four files
  fails, and any other file under `corpus/` fails. A glob that matches nothing otherwise reads
  exactly like a clean board.
- **The goldens pin a BUILD, not a truth** — tesseract 5.5.3, leptonica 1.87.0, the traineddata and
  the PDFium render. Bumping a pin moves these files, and reading that diff is the review the bump
  deserves. `expect` is what says which direction is an improvement.

## Adding a page

Download the source yourself, then:

```
go run ./internal/tessocr/testdata/corpusadd -file <the file> -url <where it came from> \
  -page <n> -slug <name> -publisher … -date … -rights … -rights-evidence … -defect … -why …
```

It refuses before it writes anything if the licence is not in the closed set or a field is missing.
Then fill in `expect` from the PAGE IMAGE — not from the reading, which is the thing being judged —
and regenerate the README, the golden and `STATUS.md`:

```
eval "$(./third_party/pins/build-cstack.sh env linux-amd64 <cstack>)"
go test -tags tessocr -count=1 -ldflags '-linkmode external -extldflags "-static"' \
  ./internal/fetchcache/ -run TestCorpusGoldens -update
UPDATE_GOLDENS=1 go test -count=1 ./internal/tessocr/ -run 'TestEveryCorpusREADME|TestREFERENCES'
```
