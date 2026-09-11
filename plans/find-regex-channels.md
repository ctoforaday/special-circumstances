# `find` attributes each hit by WHERE ripgrep matched — byte offset to JSON field

> STATUS: PASSED plan-audit round 9 (2026-09-11); approved by gblock ("implement, PR, merge when
> green"); implemented and independently reviewed. The raw-JSON literal limit is #905. Revision 9 — answers plan-audit round 8 (a transcript cut AFTER
> its match keeps the span bytes and fails the parse: `recordsAt` now counts every `ok=false`
> wanted record as a stated miss). Revision 8 answered round 7 (the span check now runs
> BEFORE the parse, so a transcript truncated mid-match is counted stale instead of a silent `?`;
> filler lengths are multiples of 4). Revision 7 answered round 6 (a third literal-output
> class: `result` snippets now come from the decoded tool output; §V.4 pins the binaries and
> `--store` so no command can open the real store). Revision 6 answered round 5 (a key's path, the
> negative-offset guard, the byte-order-mark test layout, a snippet-shape golden, the refusal help
> paragraph, pinned `windowAt` layouts; and `parseRipgrep` loses its 8 MB line cap). Revision 5
> answered plan-audit round 4 (test precision for
> `windowAt`, structure starts, out-of-range spans and the byte-order mark; the walk is full-line,
> and R5 is regraded on round 4's measurement). Revision 4 answered round 3 (design confirmed sound:
> ripgrep's offsets hold across multibyte UTF-8, CRLF and a final line with no newline; gaps were
> non-string tokens, the stale path's reach, test seams and the `--in` cost). Revision 3 adopted
> round 2's recommendation: attribute each match by its byte offset in the raw line and the JSON
> value that contains it, instead of re-finding matched text in decoded strings. Found while
> implementing #885 (PR #902, merged); gblock ruled "fix it now", before the restart-recovery work.

## I. Summary & Goals

**The defect, reproduced 2026-09-11 on real data** (the #902 binary against a shape-4 store of
this box's corpus, `~/.claude/scratch/restart-probe/shape-885/catalogue.db`):

- `find 'conduct section' --in peer` → two `peer` rows (`67082c17`, `3ac18039`).
- `find --regex 'conduct sect(ion)' --in peer` → "16 transcript(s) contain "conduct sect(ion)", but
  none has a hit in "peer"".
- `find --regex 'conduct sect(ion)'` (no filter) → rows labelled `result`, with the raw JSON line's
  head as SNIPPET.

**Cause.** ripgrep matches the pattern, then `catalogue.DecodeHit` RE-FINDS the term in each decoded
block with `contains`, a case-insensitive literal substring test (`hit.go:119`), and centres the
snippet with `window`, which searches the literal too (`hit.go:130`). A pattern is never a
substring, so no speaker, thinking or tool_use block is found. The fact "where did it match" is
known to ripgrep and thrown away (`parseRipgrep`, `find.go:207-240`, keeps `path:line` only), then
guessed back from text — a pattern standing in for a field.

**Two further defects in the same function, for literals too.**

- The `tool_result` branch (`hit.go:108-109`) returns `result` without checking the term, so a
  `tool_result` block BEFORE the block that holds the term takes the hit (records of that shape in
  10 transcripts on this box, round-1 audit).
- For the same reason, a match in a tool-result record's cwd, uuid or `tool_use_id` is labelled
  `result` (24,568 records match `tool_use_id` alone, round-2 audit). `result` is documented as
  "what a tool returned" (`hit.go:29`); those are not.

**Objective.** Every hit — literal or pattern — is attributed by the JSON value that contains
ripgrep's match, read from the match's byte offset. IN, SNIPPET and `--in` then mean the same for a
pattern as for a literal, whatever the block order and whatever escapes the match touches.

**Success criteria.**

1. `find --regex 'conduct sect(ion)' --in peer` and `find --regex '\<conduct sect' --in peer`
   (ripgrep word-start, which Go's regexp misreads) return the same two peer rows as the literal
   (§V.4).
2. Each row's IN is decided by the value containing the match start (§III.2's table): a speaker,
   `thinking`, `tool_use` anywhere inside a call's arguments, `result` anywhere inside what a tool
   returned, else `?`. SNIPPET is centred on the match.
3. Matches that touch JSON escapes attribute correctly: a match cut through an escape
   (`section\S` → `section\`), one starting mid-escape (`\w+ection\b` → `nSelection` inside
   `\nSelection`), and one spanning an escaped quote in a `tool_use` input (`ENVELOPE — \\"`).
4. A matching record `find` could not read back as ripgrep matched it — its bytes at the offset
   differ, its line is past the end of the file, the file no longer opens, or its line no longer
   parses as a record (cut after the match, or mid-write) — is a STATED miss:
   counted on stderr, never a row's best hit, and a zero-row result with any such record — filtered
   or not — refuses with exit 1 rather than printing "none has a hit" or an empty table.
5. `find` ignores the caller's ripgrep configuration (`--no-config`) and ripgrep's encoding
   sniffing (`--encoding none`), so matches and offsets are the file's raw bytes on every box. Both
   are tested (§V.2) and stated in the help.
6. Literal output changes in exactly four classes. **Unreadable records**: a matched line that
   no longer parses (or changed, or went away) is a stated miss — counted on stderr and never a
   row — where today it can be a `?` row (criterion 4). **Snippet source**: a `result` row shows the
   DECODED tool output around the match (row 5–6's string), no longer a window over the whole raw
   JSON line (`hit.go:108-109`) — so nearly every `result` row reads differently (the old window
   showed JSON framing; escapes and whitespace runs now decode), whatever the field length; no golden fixture has tool output
   (`corpus_test.go:67-77`), so §V.4 shows it on real data. **Attribution**, only where the old one was
   wrong: block order; a match in a tool-result record's ids, cwd or block keys (`tool_use_id`,
   `is_error`, `type`) now `?`; a case-variant of the term in an earlier block no longer steals the
   hit. **Snippet shape**: for a field longer than the window, the cut is made before whitespace is
   flattened (rule 7), so a snippet from text with whitespace runs can hold different words than
   today, and a cut that lands on whitespace drops it (`… session` becomes `…session`).
   Placement keeps `window`'s n/3 lead and end-clamp, and a field that fits the window whole is
   unchanged. Every golden that changes is named, with its class, and re-read — computed by round
   5: `find-channel-peer.golden` (snippet shape, the dropped space); the notification golden does
   not change.

## II. Technical Context

- Module `plugins/gray-area/tools` (Go 1.25.13). Branch `fix/find-regex-channels` from
  `origin/main` af5e8a4d. ripgrep 15.2.0.
- **ripgrep gives the offset.** Measured on `3ac18039`'s transcript:
  `rg --no-config -n -b -o --null --with-filename 'conduct sect(ion)'` prints
  `<path>\0 2430:6471105:conduct section` (no space): with `--only-matching`, `--byte-offset` is the
  ABSOLUTE file offset of the match itself. Line 2430 starts at byte 6468563, so the match is at
  line offset 2542, and `dd` at 6471105 reads `conduct section`. `--fixed-strings` reports offsets
  the same way. Round 3 confirmed the offsets across multibyte UTF-8, CRLF, several matches per
  line and a final line with no newline — and found that ripgrep STRIPS a UTF-8 byte-order mark and
  reports offsets after it unless `--encoding none` (6 vs 9). `--json` carries the same offsets but
  also the whole matched LINE per match; transcript lines run to megabytes and a common term matches
  most of the corpus. Rejected on size.
- **Offset → value is exact.** Prototype `~/.claude/scratch/offset-probe/main.go` walks a line with
  `json.Decoder.Token()` and `InputOffset()`. On real lines of `3ac18039`: the peer match →
  `.message.content` (and a second copy at `.origin.body`); `section\` (cut escape) →
  `.message.content[0].content` and `.toolUseResult.stdout`; line 4035's `ENVELOPE — \"Move` →
  `.content` of a `queue-operation` (a legitimate `?`); `tool_use_id` → a key name.
- **Census of value paths** (`~/.claude/scratch/offset-probe/census/paths.tsv`, strings, every
  `~/.claude/projects/*/*.jsonl`; round 3's `~/.claude/scratch/plan-audit-regex/census2/paths2.tsv`
  adds numbers and booleans): speaker text at `.message.content` (bare string, 1,419) and
  `.message.content[].text` (5,668 assistant, 103 user); reasoning at `.message.content[].thinking`;
  tool arguments under `.message.content[].input` — strings AND 3,400 non-string values (`timeout`
  1,129, `replace_all` 871, `run_in_background` 478, `limit` 301, `offset` 296); tool results at
  `.message.content[].content` (14,204 strings) and `.message.content[].content[].text` (456); a
  structured copy of each result under `.toolUseResult` (objects, and a bare string in 350
  records); queued prompts at `.attachment.prompt` (strings only, 467). The peer body is
  duplicated at `.origin.body` (54); hook output (`.attachment.stdout`, `.attachment.content`),
  reminders (`.attachment.text`), `.lastPrompt`, `queue-operation .content` and every
  id/cwd/branch are unmodelled.
- **Out of scope, stated:** a literal containing a character JSON escapes (`"`, `\`, a newline)
  never matches its escaped form in the raw line. That is ripgrep searching raw bytes, unchanged by
  this plan; it gets a sentence in the help (§III.3), not a fix.
- **Census — `DecodeHit`** (`grep -rn 'DecodeHit(' plugins/gray-area/tools`): `hit.go:70`,
  `telecli/find.go:356` (only production caller, in `recordsAt`), `hit_test.go:107,126`.
- **Census — `contains`, `window`**: only inside `hit.go`. `contains` is deleted (no caller left);
  `window` is replaced by `windowAt` (§III.2); its doc comment (`hit.go:123-129`) goes with it.
- **Census — `parseRipgrep`, `fileLine`, `recordsAt`, `decodeHits`**
  (`grep -rn -E 'parseRipgrep|fileLine|recordsAt|decodeHits' plugins/gray-area/tools`): `find.go`,
  and `golden_test.go:397-425` — `TestInReadsEveryHitAndStatesNoCap` builds `fileLine{path, line}`
  per body line and calls `decodeHits` directly. It moves to the new record type with offsets it
  computes from the body it wrote (§V.2).
- **Census — `?` and `result` definitions** (`grep -n -E '`\?`|^ *\? |result' plugins/gray-area/README.md
  plugins/gray-area/skills/telepathy/SKILL.md plugins/gray-area/tools/internal/telecli/find.go
  plugins/gray-area/tools/internal/catalogue/hit.go plugins/gray-area/tools/internal/telecli/golden_test.go`):
  `find.go:47-51,64-65` (help), `plugins/gray-area/skills/telepathy/SKILL.md:42-49,60`,
  `plugins/gray-area/README.md:112`, `hit.go:29,33-36`, `golden_test.go:341` (comment asserting a
  regex search reports `?`). `?` keeps its meaning as a row (a part of the record nothing here models)
  and gains "a key name" in its examples; an unparsed line is no longer a `?` row but a stated
  miss, so the `ChannelUnknown` comment (`hit.go:33-36`) drops "or the line did not parse at all"
  from what `find` shows while `DecodeHit` still returns `?` for it; `result` narrows to what its words
  already say.

### Alternatives weighed

| Option | Verdict |
|---|---|
| Compile the pattern with Go's `regexp` (revision 1) | Rejected: a second engine; misreads `\<…\>`, `\b{start}`, class intersection silently and cannot see ripgrep's case setting |
| Re-find ripgrep's matched text, JSON-unescaped (revision 2) | Rejected by measurement: cut escapes (137 records for `section\S`), mid-escape starts (714 for `n[A-Z][a-z]+ing\b`), escaped `tool_use` input, and an "undecodable" predicate that misfires both ways |
| `rg --json` submatch offsets | Rejected: the same offsets plus the full line per match — output the size of the corpus on a common term |
| **`--byte-offset --only-matching --null --encoding none`, attribute by the value containing the offset** | **Chosen**: one engine, one attribution path for literal and pattern, no re-finding, output stays match-sized |

## III. Proposed Changes

```
plugins/gray-area/
├── README.md                                   [MODIFY] :112 — `?` examples; `result` is inside a tool's output
├── skills/telepathy/SKILL.md                   [MODIFY] :42-49 IN/`?`/`result`; :60 — IN and --in work for patterns
└── tools/
    ├── internal/catalogue/hit.go               [MODIFY] Span; Hit.Stale; DecodeHit(line, []Span, inSubagent); locate; the value table; windowAt; delete contains/window
    ├── internal/catalogue/hit_test.go          [MODIFY] cases built with spans; every table row and rule (§V.1)
    ├── internal/telecli/find.go                [MODIFY] rg flags; parseRipgrep → per-record matches; recordsAt line offsets and unreached lines; settle(); help
    ├── internal/telecli/find_test.go           [NEW] recordsAt stale reach; settle refusal; parseRipgrep has no line cap; raw-byte offsets through real ripgrep (BOM)
    ├── internal/telecli/root_test.go           [MODIFY] --regex '\<…' --in peer; RIPGREP_CONFIG_PATH
    ├── internal/telecli/golden_test.go         [MODIFY] :341 comment; TestInReadsEveryHitAndStatesNoCap → offsets; a --regex --in case
    └── internal/telecli/testdata/golden/       [NEW] find-regex-in-peer; [MODIFY] help-find; find-channel-peer (snippet shape); any other golden either class of criterion 6 changes (named with its class, re-read)
```

### III.1 What `find` hands the decoder

- `ripgrepArgs(term string, asRegex bool, names []string) []string` (extracted from `RunE`, so a
  test runs real ripgrep with exactly `find`'s flags) returns `--no-config --encoding none
  --line-number --no-heading --with-filename --only-matching --byte-offset --null` (+
  `--fixed-strings` for a literal), `--`, the term and the names. Output:
  `path\0line:offset:match`. `--null` removes the left-to-right `:` split, which reads a Windows
  path's drive letter (`C:\…`) as the separator.
- `parseRipgrep` returns records in order as `ripRecord{path string; line int; matches
  []rawMatch{abs int64; text string}}`, collapsing matches per `(path, line)` as today so HITS still
  counts records. It walks `raw` with `strings.Cut(…, "\n")` rather than a `bufio.Scanner`:
  ripgrep's output is already in memory, and the Scanner's 8 MB line cap (`find.go:217`, its error
  unchecked) silently dropped every result after a longer match — which `--regex '.*'` on a
  megabyte transcript line produces. `fileLine` is replaced by `ripRecord`; `decodeHits(recs []ripRecord, paths, want)`
  loses its `term` parameter (the offsets carry what the term did).
- `recordsAt(path string, recs []ripRecord, inSubagent bool) (hits []catalogue.Hit, stale int)`
  reads with `bufio.Reader.ReadBytes('\n')`, which keeps the delimiter, so every line start's byte
  offset is exact; each match becomes `catalogue.Span{Start: abs - lineStart, End: Start +
  len(text), Text: text}`. `stale` counts the wanted records it could not decode as matched: every
  record when the file does not open, and every wanted line past the end of the file. A record whose
  bytes differ is returned as a hit with `Stale` set and counted too, and so is every wanted record
  `DecodeHit` returns with `ok=false` (recordsAt sets `Stale` on it; today's `h, _ :=` at
  `find.go:356` discards that bit) — a line that no longer parses, including a transcript cut after
  the match but before that line's end.
- `decodeHits(recs, paths, want) (rows []row, matched, stale int)` sums `recordsAt`'s counts.

### III.2 `catalogue.DecodeHit(line string, spans []Span, inSubagent bool) (Hit, bool)`

```go
// Span is one ripgrep match inside one record, in RAW LINE BYTES — the escaped JSON as written,
// not the decoded text — plus the bytes ripgrep reported, so a line that changed since the
// search is detected rather than misread.
type Span struct {
    Start, End int
    Text       string
}

type Hit struct {
    TS      int64
    Channel Channel
    Snippet string
    // Stale: this record cannot be attributed as ripgrep matched it — DecodeHit sets it when the
    // line's bytes at a span are not ripgrep's match (the file changed since the search), and
    // recordsAt sets it on a record that does not parse (cut, mid-write, or never valid).
    Stale   bool
}
```

1. FIRST, before any parse (it needs no JSON): any span with `Start < 0 || Start > End || End >
   len(line)`, or else `line[Start:End] != Text` → `Stale: true`, `?`, snippet from the head of
   the line. A stale record is never attributed. (`Start < 0` is real: a file rewritten so line N
   now begins after ripgrep's absolute offset. Running this first is what catches a TRUNCATED
   transcript: a line cut partway through its match is a torn prefix that would otherwise fail the
   parse and come back as a plain, uncounted `?`.)
2. The line does not parse → `?`, `ok=false`. A line reaches here with its span bytes intact: a
   record cut AFTER its match but before its line's end, one still being written, or one that
   never parsed. None of them can be attributed, and `find` does not pretend to: `recordsAt`
   counts every `ok=false` wanted record in the stated-miss count (§III.1).
3. No spans → `?`, snippet from the head of the line, `ok=true`. `find` never builds one; it is
   defined so the zero input cannot guess a channel.
4. `locate(line)` walks the WHOLE line with `json.Decoder.Token()`/`InputOffset()` and records the
   raw EXTENT `[s,e)` and JSON path (object keys, array indices) of every key and every value —
   strings, numbers, booleans, null, and objects and arrays (from their opening delimiter to the
   end of their closing one). **A key carries the path of the OBJECT THAT HOLDS IT**, never its
   member's: the `"text"` key of block `i` has path `.message.content[i]` (no row → `?`), while a
   key inside a tool call's arguments has path `.message.content[i].input` (row 4). So `find text`
   or `find input` does not attribute every record through the key names that frame each field.
   A token's `s` is the first byte after skipping whitespace, `:` and `,`
   from the offset before `Token()`; `e` is `InputOffset()` after it. There is no early stop: the
   root object closes only at the line's end, and rows 4–5 need the extent ends of the containers
   holding the span. Each span is attributed by the DEEPEST extent containing its `Start` — so a
   `Start` on structure (`{`, or the `:` or `,` between members) belongs to the innermost container
   around it.
5. **The value table** — the ONLY places that get a channel; everything else, including a `Start`
   in structure outside every row below, is `?`:

   | `Start` lies inside the extent of | Channel | Snippet |
   |---|---|---|
   | `.message.content` when it is a string | `SpeakerOf(facts)` | `windowAt` that string |
   | `.message.content[i].text`, block `i` of type `text` | `SpeakerOf(facts)` | `windowAt` that string |
   | `.message.content[i].thinking`, block type `thinking` | `thinking` | `windowAt` that string |
   | `.message.content[i].input`, block type `tool_use` — any key or value within | `tool_use` | `name + " " +` the raw input extent, windowed on `Start` (as today) |
   | `.message.content[i].content`, block type `tool_result` — any key or value within | `result` | `windowAt` the deepest string holding `Start`, else the raw extent windowed |
   | `.toolUseResult` — whatever its shape (object, array, or bare string), any key or value within | `result` | as the row above |
   | `.attachment.prompt` when a string, or `.attachment.prompt[j].text`, attachment type `queued_command` | `SpeakerOf(attachment facts)` | `windowAt` that string |

   Block and attachment types are read from the decoded record (`decodeBlocks`, `rec.Attachment`),
   indexed by the path — never inferred from text. `facts` are built exactly as today.
   `.origin.body` stays `?`: it duplicates `.message.content`, which precedes it in the line, so
   the record is attributed there first (rule 6). Keys of a block itself (`tool_use_id`,
   `is_error`, `type`) lie outside every row → `?`.
6. A record with several spans takes the FIRST span in line order whose channel is not `?`, else
   `?` from the first span. This is today's "first block that holds the term" rule, stated in
   offsets.
7. `windowAt(raw string, s, e, at, n int)` — precondition `s ≤ at < e`, guaranteed because it is
   called only for the string value whose extent is the deepest one containing `Start`; an `at` on
   the opening quote maps to decoded index 0 and one on the closing quote to `len(d)` — decodes
   the JSON string token `raw[s:e]` and centres on the DECODED index of `at`, found by walking the
   raw bytes escape-by-escape (an invalid raw UTF-8 byte counts as 3, the U+FFFD
   `encoding/json` decodes it to): `\"` `\\` `\/` `\b`
   `\f` `\n` `\r` `\t` = 1 byte; `\uXXXX` = the UTF-8 length of that code point, a valid surrogate
   pair = 4, and a lone surrogate = 3 (Go's `encoding/json` decodes it to U+FFFD); an `at` inside an
   escape maps to that escape's decoded start. Placement is `window`'s, on the decoded string `d`
   and index `i`: if the flattened `d` fits in `n`, return it whole (unchanged from today);
   otherwise `start = max(0, i - n/3)`, `end = start + n`, and if `end > len(d)` then `end,
   start = len(d), max(0, len(d)-n)`; both moved back to rune boundaries; `…` prefixed when
   `start > 0` and appended when `end < len(d)`. The cut `d[start:end]` is THEN flattened
   (`strings.Fields` joined by one space), so whitespace before the match cannot move the
   centre. A raw extent's window (and a `?` snippet) places on `Start` in the raw line the same
   way, with no decoding.

### III.3 `find`

- `settle(out, errw io.Writer, rows []row, matched, stale int, term string, only catalogue.Channel,
  showPaths bool) error` holds everything after decoding: when `stale > 0`, one stderr line —
  `telepathy find: <n> matching record(s) could not be read back as ripgrep matched them (the file
  changed or went away since the search, or the record no longer parses); they are not counted in
  any row's IN`; zero rows with
  `stale > 0` — under `--in C` or unfiltered — → the sentinel `errStaleZero` (exit 1; a value, like
  `errNoRipgrep`, so a test asserts it with `errors.Is` rather than by its prose); zero
  rows with `stale == 0` under `--in C` → "none has a hit in C" as today; otherwise `render`.
  `RunE` calls it; a test calls it directly.
- A stale hit is never a row's `best`; a transcript whose every hit is stale yields no row and is
  in the stale count. HITS still counts ripgrep's records, as today.
- Help (`find.go:47-51,64-65,85-93`): the refusal paragraph (`:90-93`, today: ripgrep absent, a
  pattern that does not compile, a corpus it could not finish reading) gains the fourth case — no
  row survives and some matching records could not be read back as ripgrep matched them — and
  names the stderr line that counts them when rows do survive. IN is read from the value ripgrep's match is in, the same for
  `--regex` as for a literal; `result` is a match anywhere inside what a tool returned, `tool_use`
  anywhere inside a call's arguments; `?` examples gain "a key name"; `find` ignores ripgrep
  configuration files and reads transcripts as raw bytes (no encoding detection, a byte-order mark
  is not skipped); a literal is matched against the raw JSON, so a term containing `"` or `\` does not match
  text that holds it (the JSON stores `\"` and `\\`; measured: `ENVELOPE — "Move` 0 records,
  its escaped form 3), while a term containing a line break is refused by ripgrep and the refusal
  relayed. `help-find.golden` regenerated.
- That raw-JSON literal limit is pre-existing and out of scope; gblock approved tracking it, filed
  as #905.

## IV. Risk & Mitigation

| # | Risk | L × I × cost | Mitigation |
|---|---|---|---|
| R1 | A match touching a JSON escape is misattributed | was high (137 / 714 real records under revision 2) → none by construction: attribution never decodes the match | §V.1 cases for cut, mid-escape, escaped-quote `tool_use`, span across two fields |
| R2 | Literal output changes | certain (unreadable records become stated misses; snippet source for nearly every `result` row; attribution where the old one was wrong; snippet shape) × low × low | Criterion 6 names the four classes; each changed golden is named in the PR with its class and re-read; §V.4 compares the literal peer rows and a literal `result` row's snippet before/after |
| R3 | A transcript rewritten, truncated or removed between search and read | low × medium (wrong channel, or a silent zero under `--in`) × low | Bytes compared, unreached lines and unopenable files counted, stated, `--in` refuses (criterion 4); §V.1 and §V.2 drive all three |
| R4 | `--no-config` / `--encoding none` change a caller's results | low × low × none | Only for callers with a ripgrep config, or a transcript with a byte-order mark (none on this box — without the flag every record of such a file would read stale); stated in help; both tested through real ripgrep (§V.2) |
| R5 | Cost of the full-line token walk on the UNCAPPED `--in` path | low × low × low | Round 4 prototyped the whole new `--in` path on the store copy (`find the --in assistant`: 34,447 records, 615,498 matches): 4.6 s including a 1.9 s full walk, against 9.35 s for the #902 binary — the new path does less work per record than the old re-find loop, not more. No early stop is claimed (rule 4). Gate: §V.4 times both searches back to back; the new `--in assistant` time MUST be ≤ 1.5× the #902 time, or the change stops and goes back to gblock with the numbers |

## V. Verification Plan

1. **Unit and race** (re-arms on any change under `plugins/gray-area/tools`):
   `go -C plugins/gray-area/tools vet ./...`;
   `strace -f -e trace=open,openat -o ~/.claude/scratch/restart-probe/regex-trace.txt go -C plugins/gray-area/tools test -count=1 ./...`
   then `grep -c "${XDG_STATE_HOME:-$HOME/.local/state}/special-circumstances" ~/.claude/scratch/restart-probe/regex-trace.txt`
   MUST print 0; `CGO_ENABLED=1 go -C plugins/gray-area/tools test -count=1 -race ./...`.
   `hit_test.go`: every existing case, with its span computed from the built line
   (`strings.Index` of the RAW JSON form in the marshalled line), keeps its channel and snippet
   assertion; and new cases — each names the table row or rule it fails without:
   - a `tool_result` block BEFORE a matching `text` block → the text's speaker (rule 6, row 2);
   - the term in a text block AND a later `tool_use` input, both spans passed → the speaker (rule 6
     "first": fails under "last");
   - the term only in a tool result's string content → `result`; only in a nested
     `…content[j].text` of a `tool_result` → `result` (row 5); only in `.toolUseResult.stdout` →
     `result`; `.toolUseResult` as a bare string → `result` (row 6);
   - only in the cwd, a uuid, or the `tool_use_id` key/value of a tool-result record → `?`;
   - a numeric `tool_use` input value (`"timeout":120000`, span on `120000`) → `tool_use`; a key
     under `input` → `tool_use` (row 4);
   - STRUCTURE STARTS (rule 4's container extents): a span whose `Start` is the `{` opening a
     `tool_use` input, one on the `,` before `"replace_all"` inside it, and one on the `:` after
     that key → `tool_use` each; a span whose `Start` is the `,` between a `tool_result` block's
     `"tool_use_id"` member and its `"content"` member → `?` (fails if container extents are
     dropped, or if the block's own extent were given row 5);
   - `.attachment.prompt` as an array of text blocks, attachment type `queued_command`, origin
     `peer` → `peer` (row 7, array form);
   - a cut escape (span ending on the `\` of `\n` in a text block) → the speaker;
   - a mid-escape start (span beginning on the `n` of `\n` before `Selection`) → the speaker, and
     the snippet holds `Selection`;
   - a `tool_use` input holding `\"` with the span across it → `tool_use`;
   - a span starting in one field and ending in the next → the first field's channel;
   - `.origin.body` only → `?`;
   - KEY PATHS (rule 4): a span on the `"text"` key of a text block → `?`; a span on the
     `"input"` key of a tool_use block → `?` (each fails if a key took its member's path);
   - `windowAt`, each asserting the EXACT snippet string. The surrogate cases use HAND-WRITTEN
     raw lines: `json.Marshal` (which `rec()` uses) writes non-ASCII as raw UTF-8, so a `\u`
     surrogate escape never reaches its output; the whitespace and escaped-quote cases may use
     `rec()`, which does escape `"`, `\` and control characters. Filler is a
     NON-REPEATING ASCII sequence (`fmt.Sprintf("%03d,", k)` for k = 0, 1, …) so any shift of
     `start` or `end` changes the exact snippet:
     - 148 bytes of filler, then 120 bytes of newline/tab/space runs (as JSON escapes and raw
       spaces), then `MARK` and 300 bytes of filler, in a text block → the exact snippet; fails if
       flattening precedes the cut;
     - (filler comes in 4-byte `%03d,` chunks, so every filler length is a multiple of 4) 100
       bytes of filler, 12 escaped surrogate pairs for U+1F600 (12 raw bytes each, 4 decoded), 52
       bytes of filler, `MARK`, 200 bytes of filler → decoded index of `MARK` = 200, `start` = 160
       (ASCII, past the emoji run, which ends at 148), `end` = 280 → the exact snippet; skipping
       the escape mapping puts the index at 296 (a 96-byte shift) and fails;
     - a lone escaped high surrogate U+D800 (6 raw bytes, 3 decoded) ALONE: 100 bytes of filler,
       the surrogate, 52 bytes of filler, `MARK`, 200 bytes of filler → decoded index 155,
       `start` = 115, `end` = 235, both on ASCII → the exact snippet; counting the surrogate as
       6 (raw), 1 or 4 moves both ends and fails;
     - a span starting on the `n` of an escaped newline before `Selection`, with 100 bytes of
       escaped quotes before it → the exact snippet, which holds `Selection`;
   - RULE ORDER: a torn line (a record cut partway through, so it does not parse) whose span runs
     past its end → `Stale`, not `ok=false`; a torn line whose span bytes are intact → `?`,
     `ok=false`, not `Stale` (fails if the parse runs first, or if the check ignores intact bytes);
   - OUT OF RANGE (rule 1): a span with `Start < 0`; one with `Start > End`; one with `Start`
     in range and `End > len(line)` → each `Stale`, `?`, and no panic (each fails against a guard
     missing that one clause);
   - a span whose bytes differ from `Text` → `Stale`, `?`; no spans → `?`, `ok=true`;
   - the unparsed line → `?`, `ok=false`.
2. **Command, golden and seam tests** (re-arm on `telecli` or `hit.go`):
   `go -C plugins/gray-area/tools test -count=1 ./internal/telecli -run 'TestGolden|TestFind|TestInReads|TestRecordsAt|TestSettle|TestParseRipgrep|TestRipgrepOffsets'`:
   - `TestInReadsEveryHitAndStatesNoCap` builds `ripRecord`s whose offsets it computes from the
     body it wrote (line start + `strings.Index(line, "sleeper term")`); its assertions unchanged.
   - `find_test.go` `TestRecordsAtCountsWhatItCannotReadBack`: a temp transcript; records for a
     line whose bytes at the offset differ, a line past the end, a path that does not exist, and a
     record whose absolute offset lies BEFORE its line's start (the file grew ahead of it since the
     search: negative `Start`), a transcript TRUNCATED THROUGH its only match, and one truncated
     AFTER its only match but before that line's end (span bytes intact, the line unparseable) →
     `stale` = 6 (each alone = 1, no panic), and no stale hit is a row's best in `decodeHits`.
     BOTH truncated cases continue through `settle` with `only=peer`: `errors.Is(err,
     errStaleZero)` — the silent zero criterion 4 refuses; the after-the-match case fails if
     `recordsAt` drops `ok`.
   - `find_test.go` `TestParseRipgrepHasNoLineCap`: `parseRipgrep` over output holding one 9 MB
     match line between two ordinary ones → all three records (the old 8 MB Scanner cap dropped the
     long line and everything after it).
   - `find_test.go` `TestSettleRefusesAZeroItCouldNotRead`: `settle` with no rows, `stale=2`,
     `only=peer` → `errors.Is(err, errStaleZero)` and no "none has a hit" on `out`; the same with
     `only=""` → `errStaleZero` and no table on `out`; `stale=0`, `only=peer` → "none has a hit in
     \"peer\"" and nil; rows present and `stale=1` → the stderr line and the rows.
   - `find_test.go` `TestRipgrepOffsetsAreRawBytes` (skips loudly without ripgrep, like the
     others): a temp transcript whose line 1 is the UTF-8 byte-order mark (`EF BB BF`) followed by
     an unrelated record (that line never parses, which is fine — it holds no match), and whose
     line 2 is a user record with `origin.kind` `human` holding `conduct section` in a text block;
     run real ripgrep with `ripgrepArgs("conduct section", false, …)`, parse with `parseRipgrep`,
     read back with `recordsAt` → `stale == 0` and channel `user`. Fails without `--encoding none`
     (round 5 measured ripgrep 15.2 reporting offset 84 against the true 87: stale 1).
   - `root_test.go` through `h.run` (named `TestFindRegexReportsWhoSpoke` and
     `TestFindIgnoresRipgrepConfig`, so the `-run` filter above selects them):
     `--regex '\<conduct sect' --in peer` returns the peer row;
     with `t.Setenv("RIPGREP_CONFIG_PATH", <file containing --ignore-case>)`,
     `find 'CONDUCT SECTION'` prints "no transcript contains" (fails without `--no-config`).
   - new golden `find-regex-in-peer` (`--regex 'conduct sect(ion)' --in peer`: the peer row);
     `help-find` regenerated; every other changed golden listed in the PR body with its reason.
   Then `go -C scripts run ./golden` (uncached, as CI runs it).
3. **Surfaces**: `go -C scripts run ./pluginparity`; `go -C scripts run ./frontmatter`;
   `go -C scripts run ./archaeology -base origin/main`; `go -C scripts run ./rulesweep -base
   origin/main` (the skill changes; the commit carries `Rule-Class` / `Sibling-Sweep`).
4. **Real data**: build the new binary to `~/.claude/scratch/restart-probe/bin/telepathy-regex`
   (`go -C plugins/gray-area/tools build -o … ./cmd/telepathy`). EVERY command below runs as
   `<binary> --store ~/.claude/scratch/restart-probe/shape-885/catalogue.db find …` — the old
   binary is `~/.claude/scratch/restart-probe/bin/telepathy-885`, the new one `telepathy-regex`;
   a bare `find` would open the real store, which this plan never touches:
   - a literal term that lands in a tool result (chosen from `find <term> --paths` output so its
     tool output holds a `\n`): the row's snippet from both binaries, reported side by side — the
     old one a window of raw JSON, the new one the decoded output around the match (criterion 6,
     snippet source);
   - `find --regex 'conduct sect(ion)' --in peer` and `find --regex '\<conduct sect' --in peer`
     MUST each list `67082c17` and `3ac18039` as `peer`, as `find 'conduct section' --in peer`
     does (before and after);
   - `find --regex 'conduct sect(ion)'` MUST NOT label those rows `result`;
   - `find --regex 'section\S'` and `find --regex '\w+ection\b'` MUST print no stale line, and a
     row whose record's match lies in a text block MUST read as its speaker (spot-check three with
     `--paths` against the line);
   - `find tool_use_id`: a row MUST read `result` only where the match lies inside tool output
     (round 7: 158 records in 67 transcripts quote `tool_use_id` in tool output, where `result` is
     right); any `result` row is checked against its line with `--paths` before being called
     either correct or a regression, and none may come from the block's own `tool_use_id` key;
   - timing (R5): `time` `find the` and `find the --in assistant` with
     `~/.claude/scratch/restart-probe/bin/telepathy-885` and the new binary, each run twice back to
     back, reported; the new `--in assistant` time MUST be ≤ 1.5× the old.
5. **Auditor gate**: `/prosthetic-conscience:plan-audit plans/find-regex-channels.md` → PASS.
