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
| `reading.<goos>-<goarch>.golden` | what the reader produces today on that target, byte for byte, including where that is wrong: one set per target (gblock, 2026-11-10), and every case holds the same targets, `linux-amd64` always |

## The rules the gates enforce

- **Rights are a closed set.** `corpus.Rights` lists them. A page whose licence cannot be named in
  that set is not committed; it goes in `references.json` with the reason, and `REFERENCES.md` is
  generated from that.
- **Nothing is unaccounted for.** An empty corpus fails, a directory missing one of its three files or its linux-amd64 golden
  fails, and any other file under `corpus/` fails. A glob that matches nothing otherwise reads
  exactly like a clean board.
- **The goldens pin a BUILD on a TARGET, not a truth** — tesseract, leptonica, the traineddata, the target and
  the PDFium render. Bumping a pin moves these files, and reading that diff is the review the bump
  deserves. `expect` is what says which direction is an improvement.

## Adding a page

Download the source yourself, then:

```
go run ./devcmd/corpusadd -file <the file> -url <where it came from> \
  -page <n> -slug <name> -publisher … -date … -rights … -rights-evidence … -defect … -why …
```

It refuses before it writes anything if the licence is not in the closed set or a field is missing.
Then fill in `expect` from the PAGE IMAGE — not from the reading, which is the thing being judged —
and regenerate the README, this target's golden and, on linux-amd64, `STATUS.md`. Every other target's
golden comes from the `corpus-readings-<target>` artifact native-run uploads, committed after its
diff is read:

```
NATIVE_HELPER_DIR=<helper-dir> go test -count=1 \
  ./internal/fetchcache/ -run TestCorpusGoldens -update
UPDATE_GOLDENS=1 go test -count=1 ./internal/tessocr/ -run 'TestEveryCorpusREADME|TestREFERENCES'
```
