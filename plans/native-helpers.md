# Native code leaves `feov-record`: the `feov-ocr` and `feov-embed` helpers

> STATUS: proposed 2026-11-04, not started. Forks F1–F5 (§I.4) are open and go to gblock before the auditor gate.
> Ownership after approval (gblock, 2026-11-04): Part 1 (helper mechanics) and Part 3 (`feov-embed`) are built in
> the vector-search session; Part 2 (`feov-ocr`) is handed to the session working on OCR.

## I. Summary & Goals

### 1. Objective

FEOV's C and C++ code moves out of the Go binary into two native helper executables, each shipped per
target, fetched when first needed and verified by hash:

- **`feov-ocr`** carries tesseract 5.5.3 and leptonica 1.87.0 (today's `internal/tessocr` cgo engine, `shim.cpp`, and
  the embedded `eng.traineddata`).
- **`feov-embed`** carries ONNX Runtime 1.31.1 (reduced operator set) and runs gte-small for hybrid search (#1171,
  `plans/vector-search.md`).

Each helper is a **thin C++ kernel**: it decodes a request frame, calls the library, and writes a response frame on
stdout. All logic stays in Go. Rendering, cropping, rotation, table reconstruction and tokenization never move into a
helper. That rule is what keeps the native code small enough that C++ is the right language for it (§II.1).

`feov-record` and every FEOV command become pure Go, built with `CGO_ENABLED=0`.

### 2. Why not leave it as it is (the value case)

The naive alternative is today's shape: link every C/C++ library into `feov-record` with zig, and for ONNX Runtime
load the official DLL on Windows.

| | In-process static, plus a Windows DLL | Helpers (this plan) |
|---|---|---|
| Windows embedding | Official `onnxruntime.dll`, plus `msvcp140`/`vcruntime140` | Static MSVC `/MT` executable, if Part 0 proves it (F3) |
| A tesseract crash | Kills `feov-record` and the seat's command | The helper exits; Go reports a failed page and restarts it |
| Race detector | Covers the stub, never the real engine (`scripts/check/gates.go:167`, a recorded open decision) | Covers everything that ships in Go; the C++ sits behind a process boundary |
| Go link | `-linkmode external` with zig, for every FEOV command (`hooks.yml:1786-1828`) | Plain `go build` |
| Download | The engine rides inside the 41,032,712-byte `feov-record` (v1.73.0, linux_amd64) whether or not a run reads a scan | Fetched on first scan or first search |
| Each new C++ library | Must build under zig for all six targets inside the Go link | Builds per helper, with the native toolchain where zig cannot |

**What it costs:**
- A frame protocol, with a Go client and a C++ codec.
- A third fetch path, beside `fetch-bin.sh` and `sc-doctor`.
- A release job that builds non-Go artifacts and takes one Windows job's output.
- Re-pinned OCR goldens, and a rewritten `docs/c-stack.md`.

The value case holds only if the protocol stays small and the OCR goldens come through unchanged. §V checks both.

The measurements behind it are on #1171 (comment 6621511526), in `~/.claude/scratch/vecplan/ortstatic/REPORT.md`
(the static ONNX Runtime proof) and in `~/.claude/scratch/vecplan/ortq/RESULT.md` (8-bit quality).

### 3. Success criteria

- **S1.** `CGO_ENABLED=0 go build ./cmd/...` builds FEOV for all six targets, and no FEOV Go binary links externally.
- **S2.** Every helper passes `staticproof` on all six targets, against its own allowlist (§III.1.5).
- **S3.** OCR is behaviour-preserving. Each of the 13 corpus goldens
  (`internal/tessocr/testdata/corpus/*/reading.golden`) differs from main on line 1 only, the engine identity, and the
  remaining bytes are identical.
- **S4.** OCR wall time per corpus page through `feov-ocr` is within 5% of the in-process engine on the same runner,
  measured with the two interleaved.
- **S5.** `feov-embed` int8 fidelity against `ref_gte.npy` is mean ≥ 0.984 and min ≥ 0.966 on every target. These are the
  measured values. Retrieval quality, not cosine, is the bar for 8-bit, and it passed on 2026-11-04 (hybrid beats
  BM25 on SciFact +0.048 [0.026, 0.069] and on NFCorpus +0.040 [0.027, 0.054]).
- **S6.** `feov-embed` one-thread speed on 200-word chunks is ≥ 5 chunks/s on linux-amd64, linux-arm64 and darwin-arm64.
  - Measured with the reduced Release build: 7.1 chunks/s on an EPYC 7763.
  - Measured with the `-Os` build: 5.9 chunks/s on Neoverse N2 and 5.9 on M1.
  - darwin-amd64 measured 4.2 chunks/s on an i7-8700B, below the bar. That is stated, not hidden.
- **S7.** A helper that is missing, fails its hash, or crashes produces a loud, named refusal. It never comes back as an
  empty reading or an empty search result (§III.1.3).

### 4. Forks for gblock (resolve before the gate)

**F1: How helpers are provisioned.**
- *Recommended:* `feov-record` fetches a helper on first use, from the release it was built for. The asset hash is
  compiled in.
  - Hooks never wait on a helper download.
  - The engine identity can be computed without starting the helper (§III.1.4).
- *Alternative:* `fetch-bin.sh` fetches the helpers eagerly with the hook binaries. Its all-or-none rule would then make
  every hook's first start wait on tens of MB (`scripts/fetchbingen/fetch-bin.sh:246-256`).
- This revises the earlier ruling, "model as a release asset via fetch-bin.sh" (2026-10-31): the model would be fetched
  the same way as `feov-embed`, not through `fetch-bin.sh`.

**F2: Release cadence.**
- *Recommended:* helpers ship in the FEOV plugin release, built by the same release run. The release job reads no cache
  (`hooks.yml:1681-1700`), so every FEOV release rebuilds both helpers from pinned sources. The agent's estimate is
  15–25 minutes per target, run in parallel.
- *Alternative:* helpers get their own tag namespace, with a committed pin file. Each FEOV release is then faster, but it
  needs a second release kind, and the release job today refuses a tag that names no plugin (`hooks.yml:1718-1729`).

**F3: Windows `feov-embed`.**
- *Static:* built with MSVC and `/MT`, importing OS DLLs only. Part 0 tests whether that is possible.
- *DLL:* the helper loads the official `onnxruntime.dll`, fetched beside it.
- gblock ruled "Windows loads a dll" before the helper design existed. With a helper, the DLL is no longer forced, so
  this is asked again once Part 0 reports.

**F4: Who builds Part 1.**
- *Recommended:* the vector-search session builds Part 1 first. Part 2 (OCR, the other session) and Part 3 (embed, this
  session) then build on it in parallel.
- *Alternative:* the OCR session builds Part 1 inside Part 2, and embed waits for it.

**F5: Universe runs.**
- *Recommended:* `universe.sh build --native` builds both helpers from the checkout and installs them where the fetcher
  looks. A plain `build` has no helpers, so OCR and search refuse loudly. That matches today: a universe has no OCR
  engine at all (census G, `scripts/universe.sh:430-449`).
- What a universe must never do is fetch the *published* helpers, which would test the release instead of the change
  (AGENTS.md, "prove the universe carries your change").

### 5. Non-goals

- A Rust rewrite. §II.1 says why C++.
- Moving any Go logic into a helper.
- GPU execution providers.
- Changing how OCR reads a page. Part 2 is a move, and S3 is its proof.
- Retiring PDFium. It is WebAssembly under wazero, pure Go, and outside the C stack (census J).

## II. Technical Context

### 1. Language: C++, not Rust

- **The libraries are C++, and the existing code is too.** Their APIs are called directly, with no binding crate in
  between. `internal/tessocr/shim.cpp` is most of `feov-ocr` already.
- **No new toolchain.** Linux and macOS use the pinned `zig c++` (`third_party/pins/PINS.txt`). Windows `feov-embed` uses
  MSVC on a Windows runner, which is how ONNX Runtime builds there. Windows `feov-ocr` stays on zig and MinGW, as it is
  today.
- **Rust's safety would protect the wrong layer.** The untrusted input is pixels from fetched documents, which are
  parsed inside tesseract and leptonica, and Rust glue does not make those safer. The C++ glue's own exposure is the
  frame decoder, which is fuzzed and run under AddressSanitizer and UndefinedBehaviorSanitizer in CI (§V).
- **Rust's costs:**
  - A second pinned toolchain.
  - cargo-zigbuild for Linux and macOS, plus MSVC targets for Windows.
  - Binding crates of uneven upkeep: ONNX Runtime's `ort` is maintained, tesseract's crates are thinner.
- **What would reverse this:** a helper growing real logic. The thin-kernel rule (§I.1) is the guard.

### 2. What crosses the boundary today (OCR)

The full table is in census A. The Go-facing contract does not change:
`PageEngine.ReadPage(png []byte, dpi int) (tessocr.PageResult, error)` plus `Identity() string`
(`internal/fetchcache/pageread.go:41-51`).

Underneath it are 7 C operations: `tsv`, `rot_band`, `grid_stats`, `grid_lines`, `repaired_rules`, `ink`, and the
per-page `diag` capture.
- **Every image argument is encoded PNG bytes.** leptonica's `pixReadMem` decodes them and `shim.cpp normalize()`
  normalizes them. The protocol keeps sending PNG bytes, not decoded pixels, because decoding in Go could change bytes,
  and S3 depends on the bytes not changing.
- **Round trips matter, not bytes.** A grid page costs about 8 + N round trips, N being its level cells. Payloads are
  about 350 KB at 300 DPI and 1.4 MB at 600, and a page takes 12–24 s.
- **Concurrency is unchanged:** one engine per process behind a mutex (`pageread.go:61-82`), so one `feov-ocr` child per
  `feov-record` process.

### 3. What crosses the boundary (embedding)

Token ids and an attention mask go in (WordPiece in Go, which matches Hugging Face on 400 of 400 texts); 384 float32
values per text come back, mean-pooled and L2-normalized in the helper. The model is a separate asset,
`gte-small.int8.ort`, 33,844,232 bytes.

### 4. Constraints

- **Tests refuse the network** (`nonet.OnlyLoopback`, `internal/testbuild/testbuild.go:96`), so no test can fetch a
  helper. Tests take a prebuilt helper from a directory named by an environment variable, or skip loudly.
- **The variable's name must not start with `FEOV_`.** AGENTS.md has every agent session run
  `unset $(env | grep -o '^FEOV_[A-Z_]*')` before tests, and a `FEOV_` name would be swept away with the rest.
- **No schema change.** The engine identity is a string value in `Cite.ocr_engine` and in cache receipts, so there is
  no epoch bump and no `feov-migrate` translation (census J, L). Old runs keep their old identity strings, and nothing
  compares a stored `ocr_engine` with the current identity.
- **The release fuzz never exercised OCR.** `release-fuzz-build` is a plain build (census F), so it is unchanged.

## III. Proposed Changes (the spec)

### Part 0: Windows feasibility probe (in flight, scratch only)

ONNX Runtime 1.31.1 reduced, built with MSVC and `--enable_msvc_static_runtime`, linked `/MT` into a probe C++
`embed-probe.exe`. CI runs on branch `bench/embed-engines` (scratch, no PR) and reports:
- the executable's DLL imports;
- `staticproof`'s verdict;
- int8 and f32 fidelity, and one-thread speed.

Output: `~/.claude/scratch/vecplan/msvcprobe/REPORT.md`. This answers F3 and nothing ships from it.

### Part 1: Helper mechanics (owner per F4)

#### 1.1 Layout

```
plugins/frank-exchange-of-views/native/          [NEW] helper sources, outside tools/cmd (census G: every enumerator treats tools/cmd as a shipped Go binary)
  frame/frame.h, frame.cpp                       [NEW] C++ codec for the frame protocol
  frame/fuzz_frame.cpp                           [NEW] libFuzzer target for the decoder
  ocr/   (Part 2)                                [NEW]
  embed/ (Part 3)                                [NEW]
  build.sh                                       [NEW] build one helper for one target (absorbs build-cstack.sh's build and check modes)
plugins/frank-exchange-of-views/tools/internal/nativehelper/   [NEW] Go: locate, verify, spawn, frame client, restart-on-exit
plugins/frank-exchange-of-views/tools/third_party/pins/        [MODIFY] PINS.txt gains ONNX Runtime; build-cstack.sh folds into native/build.sh; staticproof gains per-artifact allowlists
```

#### 1.2 Frame protocol

Each frame is:
- a fixed little-endian header: magic, version, op, request id, payload length;
- the payload: length-prefixed fields.

The first frame each way is a handshake. It carries the protocol version and the helper's build identity; a version
mismatch is a refusal, with no negotiation. A response carries:
- a status;
- the op's result;
- the helper's stderr captured during the request (the `diag` capture moves here, census A,
  `pageread.go:445 engine.log`).

The Go client applies one deadline per request. On EOF, a non-zero exit or a deadline, it kills and restarts the
helper once, then reports the failure as named. No retry loop.

#### 1.3 Absent, unverifiable and crashed helpers (S7)

Each case is a distinct error:
- `helper-absent`: not installed and not fetchable;
- `helper-hash-mismatch`;
- `helper-crashed`, with the exit status and the captured stderr;
- `helper-protocol`.

`fetch`'s `ocr_engine_absent` field (`internal/cli/fetchsummary.go:91`) keeps its name and widens its meaning to "no
working engine". `ocr_reason` names which case it was. These replace `tessocr.ErrNotCompiledIn` and its text, "build
with -tags tessocr" (`pageread.go:88-95`, `tessocr.go:107`).

#### 1.4 Provisioning and identity (per F1/F2)

- **Hashes go in at build time.** The release job writes each helper asset's sha256 into `feov-record` with
  `-ldflags -X`. A dev build has no pinned hash: it uses a helper from the override directory (§II.4), and its identity
  is the sha256 of the file it actually opened.
- **Fetching:**
  - The fetch reads the same release's assets, `<helper>_<os>_<arch>[.exe]` and `SHA256SUMS`.
  - It writes the helper into `<plugin cache>/bin/` beside `feov-record`.
  - It verifies the hash before the file is ever made executable.
  - It takes the same machine-wide lock pattern the release fuzz uses, so two seats don't race the download.
- **The model** `gte-small.int8.ort` is fetched the same way and verified the same way.
- **Identity can be computed without the helper.**
  - `Identity()` gains `+feov-ocr@<sha256[:12]>` and keeps the pure-Go `tessocr@<hash>` part, which hashes the Go source
    that remains.
  - It reads the hash compiled into the binary, so the cache-reuse paths that call it before any read
    (`pageread.go:299`, `readscanned.go:108`) never start or fetch the helper (census J).
  - The source hash no longer covers `shim.cpp` and `shim.h`, which move out. The helper's asset hash covers them
    instead, because the asset is built from them.
- **`setup` preflight** (`internal/setup/run.go:214-232`) reports whether each helper is installed and matches its
  hash, and fetches nothing. A run learns before its first sitting that OCR is unavailable, not halfway through one.

#### 1.5 Release job (`.github/workflows/hooks.yml`)

- **A new job, `native-windows`,** runs on `windows-latest`, plus `windows-11-arm` if that label exists. Part 0 records
  whether it does. It builds Windows `feov-embed` with MSVC using `/Brepro` and no timestamps, so the release's
  idempotent re-run (`hooks.yml:1863-1883`, byte-for-byte `SHA256SUMS`) does not fail with "artifacts DIFFER". It
  uploads the result as a workflow artifact.
- **The `release` job:**
  - It gains `needs: native-windows`.
  - It builds Linux, macOS and Windows `feov-ocr`, and Linux and macOS `feov-embed`, with zig through `native/build.sh`.
  - It downloads the Windows artifacts, runs `staticproof` on every helper, and only then writes `SHA256SUMS`.
  - The Go loop loses its capability-keyed C-stack branch (`hooks.yml:1786-1828`). Every FEOV command takes the plain
    `go build -trimpath -ldflags "-s -w"` path, plus the `-X` hashes for `feov-record`.
- **`sc-doctor`'s release contract test** (`release_contract_test.go:64-74`) parses the release job for the literal
  `for pair in …;` and the job's *first* `-o "…"`. The helper build steps go *after* the Go loop, or use no `-o`, and the
  test is extended so that it also enumerates the helper asset names.
- **CI jobs `ocr-corpus` and `feov-record`** (census F):
  - They build `feov-ocr`, cached on its pinned inputs, and run plain `go test` with the override directory set.
  - The test list that used to come from grepping for the `tessocr` build tag (`hooks.yml:588`) is derived from a
    `nativehelper.Require(t, "feov-ocr")` call instead.
  - The Windows leg now exercises the real client against a real helper, not only the stub.
- **`staticproof`** takes the allowlist per artifact. `feov-ocr` keeps today's lists. `feov-embed`'s Windows list is
  whatever Part 0 measures, each DLL named with the reason it is there.

#### 1.6 Enumerators (census G)

- **Helpers are not Go `cmd/` directories**, so `fetch-bin.sh`, `fetchbingen`, `pluginparity`, `bootstrap-plugins.sh`,
  `marketplacegen`, `versionguard` and `gcpolicy` are unchanged under F1.
- **`sc-doctor`** reports each helper as installed, missing or hash-mismatched. Its fix is to fetch, never to build:
  there is no source-build fallback for C++ (`doctor/main.go:533-535`).
- **`pluginparity`** gains one check: the helpers named in `requirements.json` match the directories under `native/`.
- **`universe.sh`** follows F5.

#### 1.7 Gates and docs

- **`scripts/check/gates.go`:**
  - The race scope's stated limit (`:167-174`) is rewritten. The race detector now covers the Go client; the C++ is not
    raced, and the text says so.
  - The `ocr-corpus` skip text (`:326-331`) names `feov-ocr` and the override directory.
- **Docs:**
  - `docs/c-stack.md` is rewritten around the helpers: zig for Linux, macOS and Windows `feov-ocr`; MSVC for Windows
    `feov-embed`. The MSVC runner is the decision the current doc says a native-only toolchain would need.
  - `docs/local-development.md:302-305` is rewritten.
  - `docs/setup-script.md:73` currently says "nothing needs provisioning", which becomes false.

### Part 2: `feov-ocr` (owner: the OCR session)

**The move, file by file (census B, C):**
- `shim.cpp` and `shim.h` move to `native/ocr/`, with a main loop over the frame codec.
- `eng.traineddata` (4 MB) is embedded in the helper, and its pin and `TestEmbeddedTraineddataPin` move with it.
- `engine_cgo.go` and `engine_stub.go` are deleted. A pure-Go client implements the same exported functions over
  `nativehelper`, so `page.go`, `reconstruct.go`, `lattice.go` and the other pure-Go files do not change.

**Tests (census D):**
- The 14 tagged test files re-gate on `nativehelper.Require`. That includes `blankreceipt_cgo_test.go` and
  `cli/ocrcite_real_test.go`, which the first grep missed.
- `diagcrash_cgo_test.go` is rewritten for the new crash shape, where a crash is a helper exit.
- `libcprobe`, the musl libc probe from #1598, becomes a self-test op in `feov-ocr`, run by CI. It cannot stay a Go cgo
  test.
- `testbuild/tags_*.go` and `engineBuildFlags` are deleted.

**Agent-facing text (census H):**
- `cli/fetch.go:62` says "compiled into this binary". The fix is made there and regenerated with `agentgen`, which
  updates 10 agent definitions.
- `cli/ocr.go:244` help, `pageread.go:92-93` and `seatprobe/{build.go:360,scanned.go:29}` are rewritten. The difftest
  manual golden is regenerated, and every changed line is read.

**Goldens:** the 13 corpus goldens are regenerated, and S3 is checked by diffing everything after line 1.

### Part 3: `feov-embed` (owner: the vector-search session)

**The helper:** `native/embed/` holds the ONNX Runtime reduced operator set (`ops.config` from the scratch proof) and the
Release minimal build: 6,057,720 bytes added on linux-amd64, at full speed. One op, `embed(ids, mask) -> float32[n][384]`,
runs with one intra-op thread.

**The plan:** `plans/vector-search.md` is revised in its own PR on this one. Its llama.cpp and `internal/vecembed` sections
(`:8, :61, :69-120, :329-332, :458-506`) give way to ONNX Runtime 8-bit and the helper.
- Its search, schema, chunking and RRF design stand, along with the rulings on `--quote` and the hybrid.
- Its "approved" status is struck, because the agent that wrote it set that status itself, and it goes back through the
  gate.

### Consumer census

Re-run against `origin/main` at `7b973ec4e`, 2026-11-04. Each group's command re-enumerates it; Appendix A has every
line.

| Group | What | Count | Enumerate |
|---|---|---|---|
| A | The Go↔C boundary | 7 C ops, about 12 call sites per page | `grep -n "C\.tessocr_\|^func " $F/internal/tessocr/engine_cgo.go` |
| B | The `internal/tessocr` split | 30 files | `ls $F/internal/tessocr; grep -ln "go:build" $F/internal/tessocr/*.go` |
| C | Go importers and callers of the engine | 6 | `grep -rln 'internal/tessocr"' $F --include=*.go \| grep -v _test` |
| D | Build-tag test infrastructure | 14 tagged + 3 harness sites | `grep -rln '^//go:build.*tessocr' plugins --include=*.go` |
| E | C-stack tooling | 4 | `ls $F/third_party/pins` |
| F | `hooks.yml` | 3 jobs build the stack, 1 publishes, 2 contract tests parse the workflow | `grep -n "tessocr\|cstack\|linkmode\|for pair in" .github/workflows/hooks.yml` |
| G | Enumerators that treat `tools/cmd` as the binary list | 9 | `grep -rn 'tools/cmd' scripts plugins/*/hooks plugins/prosthetic-conscience/tools/internal/doctor .github \| grep -v _test` |
| H | Agent-facing text | 1 source, 10 generated, 3 help pages | `grep -rn "compiled into this binary\|tags tessocr" plugins/frank-exchange-of-views --include=*.go --include=*.md` |
| I | Gates | 3 | `grep -n "tessocr\|C stack" scripts/check/*.go` |
| J | Engine identity | 1 producer, 4 stores, 3 comparers | `grep -rn "Identity()\|ocr_engine" $F/internal --include=*.go \| grep -v _test` |
| K | Docs and plans | 3 docs, 1 live plan | `grep -rln "tessocr\|build-cstack\|linkmode external" docs plans` |

`F=plugins/frank-exchange-of-views/tools`.

## IV. Risk & Mitigation

| # | Risk | L × I | Mitigation |
|---|---|---|---|
| R1 | The OCR readings change when the code moves (decode, normalization or threading differences) | M × H | PNG bytes still cross the boundary; S3's line-1-only diff on all 13 goldens decides it (Part 2) |
| R2 | The MSVC static build is not possible, or not reproducible | M × M | Part 0 settles possibility before anything ships; `/Brepro` plus a release re-run in CI settles reproducibility (§V.6). If it fails, F3 falls back to the DLL |
| R3 | A protocol defect returns empty results that look like a clean reading or no hits | L × H | The distinct errors of §III.1.3; the fuzzed decoder; a test that kills the helper mid-request and asserts `helper-crashed` |
| R4 | Concurrent seats race the first-use fetch | M × L | The lock in §III.1.4; verify the hash before chmod; write to a temp file and rename |
| R5 | A universe tests the published helper instead of the checkout's | M × M | F5 and `universe.sh build --native`; the grep-the-staged-marketplace rule is extended to `cmp` the helper |
| R6 | The `sc-doctor` contract test matches a helper step's `-o` | H × L | §III.1.5 placement, and the test enumerates helpers |
| R7 | Release time grows by about 25 minutes | H × L | Accepted under F2; the alternative is a second release kind |
| R8 | Part 2 and Part 3 drift apart in protocol use | M × M | One `nativehelper` package and one `frame.h`, both landed in Part 1 before either helper |

## V. Verification Plan (the checklist)

1. **`go -C scripts run ./check`** after `unset $(env | grep -o '^FEOV_[A-Z_]*')`: every local gate at `-count=1`. Read
   the fail count, not the exit code.
2. **Pure Go:** `CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build ./cmd/...` in FEOV's tools module, for all six targets
   (S1).
3. **Frame codec:**
   - The Go fuzz test `go test -fuzz=FuzzFrame -fuzztime=60s ./internal/nativehelper/`.
   - The libFuzzer target `native/frame/fuzz_frame.cpp` under ASan and UBSan in CI.
   - A test that kills the helper mid-request and asserts `helper-crashed` (R3).
4. **Real-data check, OCR (S3, S4):**
   - The 13 corpus scans are read through `feov-ocr` in CI's `ocr-corpus` job.
   - `diff <(tail -n +2 old) <(tail -n +2 new)` is empty for every golden.
   - The interleaved in-process and helper timing table is posted to the PR.
5. **Real-data check, embedding (S5, S6):**
   - Fidelity against `ref_gte.npy` on the 300-text set, per target, in CI.
   - The replay set and BEIR hybrid scores reproduce `ortq/RESULT.md` through the helper.
   - One-thread speed per target.
6. **Release:**
   - A release dry-run on a scratch tag builds every asset and passes `staticproof` on each helper.
   - A re-run of the same tag reports "SHA256SUMS match" (R2).
7. **Running the engine:**
   - `scripts/universe.sh build --native`, then `cmp` the staged helpers against the checkout's build.
   - A smoke run that fetches one scanned PDF and runs one `--search`, with the run's `engine.log` read.
8. **The PR's `release-fuzz` parts and `release-fuzz-tally`**, read on the final head (FEOV's tools module changes).
9. **The auditor gate:** `/prosthetic-conscience:plan-audit` on this plan once F1–F5 are ruled.

## Appendix A: Census detail

Copied from `~/.claude/scratch/vecplan/census/CENSUS.md`, the run against `7b973ec4e`.

### A. The engine boundary: Go <-> C today (protocol sizing) — 1 group, the contract itself

Enumerate: `grep -n "C\.tessocr_\|^func " $F/internal/tessocr/engine_cgo.go; grep -n "^[a-z].*tessocr_" $F/internal/tessocr/shim.h`

C surface (shim.h), every image argument is ENCODED PNG BYTES (leptonica `pixReadMem` decodes, then shim.cpp `normalize()` = pixConvertTo8 + background normalization, binarize at 180 for morphology):

| C call | Go wrapper (engine_cgo.go) | in | out |
|---|---|---|---|
| tessocr_new(traineddata,len) | New() :57 | embedded eng.traineddata (go:embed :40) | handle / NULL |
| tessocr_tsv(e,png,len,psm) | Engine.PageTSV(png, PSM) :82 | PNG + PSM (3 auto, 11 sparse, 10 single-char) | level-5 word TSV string |
| tessocr_rot_band(e,png,len,x,y,w,h) | Engine.RotatedBand :91 | PNG + crop rect | TSV/text string |
| tessocr_grid_stats(png,len,sel,&h,&v,&ix) | DetectGrid(png, GridThresholds) :117 (no Engine) | PNG + SEL | 3 ints, rc stage code |
| tessocr_grid_lines(png,len,sel) | GridLines :187 (no Engine) | PNG + SEL | text "h\|v x y w h" lines -> ParseLattice |
| tessocr_repaired_rules(png,len,ch,cv,sel) | RepairedRules :202 (no Engine) | PNG + DashGap,DashGap,SEL | same lattice text |
| tessocr_ink(png,len,&dark,&total) | InkFraction :133 (no Engine) | PNG | 2 int64, rc stage code |
| tessocr_diag_begin(path)/diag_end | captureDiagnostics(fn) :153 | dup2 of process fd 2 to a temp file, SetCrashOutput | tesseract stderr text for the page |
| tessocr_free/free_text | Close, defers | — | — |

Call sites per page (pure-Go orchestration in page.go, untagged):
- `$F/internal/tessocr/page.go:194` Engine.ReadPage(pagePNG, GridThresholds) — wraps the whole page in captureDiagnostics; result.Diagnostics -> `engine.log` evidence (`$F/internal/fetchcache/pageread.go:445`) — changes: diagnostics must come back per request in the response frame (helper captures its own fd 2), crash = helper exit, not a process-wide trace.
- `page.go:211` InkFraction; `:224` DetectGrid; `:234/:261` RepairedRules (repair verdict); `:241/:274` PageTSV(PSMAuto); `:290` PageTSV on rotate90CW(page) (Go re-encodes PNG, `:623`); `:301` PageTSV(PSMSparseText); `:332` RotatedBand; `:344/:464` GridLines or RepairedRules; `:476-489` readLevels: one PageTSV(PSMSingleChar) PER LEVEL CELL on a Go-cropped PNG (`CropCell`, normalize.go:26). — changes: every one becomes a request frame; a grid page is ~8 + N_cells round-trips each carrying a full or cropped PNG (~350 KB at 300 DPI, ~1.4 MB at 600). Pages cost 12–24 s, so pipe copy is noise; the ROUND-TRIP COUNT (not bytes) is what the protocol must make cheap.
- `$F/internal/fetchcache/pageread.go:41-51` PageEngine interface {ReadPage(png []byte, dpi int) (tessocr.PageResult, error); Identity() string} — unchanged shape; implementation swaps.
- `pageread.go:54` DefaultPageEngine = &TessocrPageEngine{} ; `:61-82` lazy tessocr.New() under a mutex, one engine per process — changes: lazy spawn of feov-ocr child (one per feov-record process), mutex stays (helper is single-threaded per TessBaseAPI).
- `pageread.go:88-95` engineAbsentLoudly wraps ErrNotCompiledIn with "build with -tags tessocr" — changes: new error = helper missing/unfetchable/hash mismatch; sentence and errors.Is target change.

---------------------------------------------------------------------------------------------------
### B. internal/tessocr package split (cgo half leaves; pure half stays) — 30 files

Enumerate: `ls $F/internal/tessocr; grep -ln "go:build" $F/internal/tessocr/*.go`

- `$F/internal/tessocr/engine_cgo.go:1` `//go:build tessocr && cgo`, import "C", go:embed eng.traineddata :40 — DELETED; replaced by a pure-Go client speaking the frame protocol (same exported funcs: New/PageTSV/RotatedBand/DetectGrid/InkFraction/GridLines/RepairedRules/captureDiagnostics).
- `$F/internal/tessocr/engine_stub.go:1` `!tessocr || !cgo` stub returning ErrNotCompiledIn — DELETED (no tag); the "absent" case becomes "helper absent", loud.
- `$F/internal/tessocr/shim.cpp`, `shim.h` — MOVE into feov-ocr's C++ source (main loop + frame codec added).
- `$F/internal/tessocr/eng.traineddata` (4 MB, pinned in PINS.txt:28) — MOVES into the helper (embedded in the C++ binary, or a sibling asset); leaves the Go binary.
- `$F/internal/tessocr/tessocr.go:39-41` tesseractPin/leptonicaPin/engTraineddataPin; `:51` cstackPin typed digest of PINS.txt+build-cstack.sh; `:57` `//go:embed *.go shim.cpp shim.h` engineSource; `:100` Identity() — CHANGES: shim.* no longer in this dir; identity must name the helper (see group J).
- `$F/internal/tessocr/tessocr.go:107` ErrNotCompiledIn and its message "build with -tags tessocr" — renamed/reworded (helper absent).
- `$F/internal/tessocr/tessocr.go:1-17` package doc ("statically linked into this binary", tag split, invocation) — rewrite.
- `$F/internal/tessocr/page.go:1-14,190-205` doc "orchestrates Engine methods that are themselves tag-split" — reword; logic unchanged.
- `page.go`, `reconstruct.go`, `lattice.go`, `levels.go`, `normalize.go`, `repaired.go`, `spread.go`, `reading.go` — unchanged (pure Go, stay in feov-record; still hashed into identity).
- Tests, tagged (all need a real helper instead of the tag): `engine_cgo_test.go:20` TestEmbeddedTraineddataPin (moves to helper build/test), `goldens_cgo_test.go`, `textcells_cgo_test.go`, `diag_cgo_test.go`, `diagcrash_cgo_test.go:24,44` (crash/stderr capture via SetCrashOutput — semantics change entirely: crash = child exit), — re-gate on "helper present" not on a build tag.
- Tests, untagged: `engine_stub_test.go:12` TestStubRefusesLoudly, `stubbuild_test.go:16` TestDefaultBuildCompiles — rewrite to "missing helper refuses loudly".
- `reconstruct_test.go:160` TestIdentityMatchesPins (recomputes cstackPin from PINS.txt+build-cstack.sh), `:193` TestEngineSourceHashTracksTheCode, `reconstruct_crlf_test.go` (CRLF fold because shim.* not LF-pinned) — move/rework with identity.
- `testdata/gen/main.go:15` "go test -tags tessocr -run TestTextCellGoldens -update" — doc command changes.
- `testdata/corpus/README.md:31,49-50` tagged suite invocation (eval build-cstack env + -linkmode external) — rewrite.
- `testdata/corpus/*/reading.golden` line 1 (13 files) carries `tesseract@5.5.3+leptonica@1.87.0+eng@7d4322bd2a77+cstack@484e7942d3c4+tessocr@605c3c75d37a` — ALL regenerate on any identity change; body bytes should be UNCHANGED if PNG bytes still cross the boundary (the proof the move is behaviour-preserving).

---------------------------------------------------------------------------------------------------
### C. Go importers of tessocr / callers of the engine — 6 non-test sites

Enumerate: `grep -rln 'internal/tessocr"' $F --include=*.go | grep -v _test`

- `$F/internal/fetchcache/pageread.go:13` — PageEngine + receipts; reuse check `:299` `r.Engine == DefaultPageEngine.Identity()` BEFORE reading; `:330,:503` stamps Identity — changes: Identity() must stay cheap and infallible (it is called on cache-hit paths that must not spawn/fetch the helper).
- `$F/internal/fetchcache/readscanned.go:14,108,201` — fetch's fused render-and-read; `:108` reuse check compares prev.Engine to Identity() before rendering — same constraint.
- `$F/internal/fetchcache/pagerender.go:21,49` — only `tessocr.RenderDPI` — unchanged.
- `$F/internal/fetchcache/pagelocate.go:144` — copies rec.Engine into loc.Engine for cite — unchanged (string passthrough).
- `$F/internal/cli/fetch.go:17,205` — `s.OCREngineAbsent = errors.Is(rerr, tessocr.ErrNotCompiledIn)` — changes: target error; field `ocr_engine_absent` (fetchsummary.go:91) is seat-facing JSON, its meaning shifts from "binary lacks engine" to "helper not installed/fetch failed"; decide whether a fetch failure is "absent" or a distinct reason.
- `$F/internal/cli/ocr.go:60-72` doc comment "THE ENGINE IS IN THE BINARY NOW", `:226`, `:244` `ocr read` Long help "OCR engine compiled into this binary (tesseract + leptonica, statically linked)" — help text changes -> difftest manual golden.
- `$F/internal/cli/fetch.go:62` fetch Long help "LOCAL OCR ENGINE compiled into this binary" — changes; this sentence is inlined by agentgen into 9 agent definitions (group H).

---------------------------------------------------------------------------------------------------
### D. Build-tag test infrastructure (tessocr tag) — 14 tagged files + 3 harness sites

Enumerate: `grep -rln '^//go:build.*tessocr' plugins --include=*.go; grep -rn 'EngineCompiledIn\|engineBuildFlags' $F --include=*.go`

- `$F/internal/testbuild/tags_tessocr.go:11` `engineBuildFlags = -tags tessocr -ldflags -linkmode external -extldflags "-static"`, EngineCompiledIn=true — DELETED; replaced by "a real helper is available" (path/env), feov-record test builds become plain.
- `$F/internal/testbuild/tags_default.go` EngineCompiledIn=false — DELETED with it.
- `$F/internal/testbuild/testbuild.go:352-353` Binary() appends engineBuildFlags to `go build` — becomes plain go build; tests needing OCR must also locate a feov-ocr.
- `$F/internal/testbuild/testbuild.go:96` Main installs nonet.OnlyLoopback (seatprobe/nonet_test.go:11) — CONSTRAINT: any first-use helper download inside tests is refused off-loopback; tests need a prebuilt helper or a loopback fixture release.
- `$F/internal/seatprobe/separatedbuild_test.go:94` skips ErrEngineRequired when !EngineCompiledIn — re-key on helper presence.
- `$F/internal/seatprobe/scanned.go:29` ErrEngineRequired "binary with the OCR engine compiled in"; `build.go:350-362` reads ocr_engine_absent, error text "(-tags tessocr)" — reword/re-key.
- `$F/internal/seatprobe/fixture_cgo_test.go:1` `tessocr && cgo`, TestScannedFixtureSpanIsFoundOnPage1ByTheRealEngine (CI anchors on the name, hooks.yml:600) — re-gate.
- `$F/internal/fetchcache/enginesuccess_test.go:1` `tessocr && cgo`; `blankreceipt_cgo_test.go:1` `tessocr` (NOT in the raw grep's list); `corpus_cgo_test.go:1,44` `tessocr`, FEOV_OCR_CORPUS, TestCorpusGoldens — re-gate.
- `$F/internal/fetchcache/enginestub_test.go:1` `!tessocr || !cgo` TestAScannedReadOnAStubBuildRefusesLoudly; `$F/internal/cli/enginestub_test.go:1,18` TestFetchOnAStubBuildStatesTheEngineIsAbsent — become "no helper" tests (run everywhere, no tag).
- `$F/internal/cli/ocrcite_real_test.go:1` `tessocr`, TestOCRCiteOnRealScan, FEOV_OCR_E2E-gated, run by hand only (not in any CI step) — re-gate.
- `$F/internal/libcprobe/probe_cgo.go:1`, `probe_cgo_test.go` — `tessocr && cgo && linux`, probes the zig musl libc's bounded scans (#1598) via cgo — CANNOT stay a Go cgo test once feov-record has no cgo: becomes a helper self-test op or a C test in the helper's build.
- `$F/releasegate/fuzz/*` uses testbuild (fuzz_test.go, hookinvocation_test.go, testmain_test.go) — unchanged; fuzz never had the engine (release-fuzz-build is a plain build).

---------------------------------------------------------------------------------------------------
### E. C-stack build tooling (third_party/pins) — 4 artifacts

Enumerate: `ls $F/third_party/pins; grep -rn "build-cstack\|staticproof\|PINS.txt" --include=* -r .github scripts plugins docs | grep -v "^plans/"`

- `$F/third_party/pins/build-cstack.sh:2-12` builds the static tesseract+leptonica prefix per target; `env` mode `:79-123` prints CGO_* + CSTACK_EXTLDFLAGS for a GO cgo link (comments name `go build -tags tessocr`, Go's darwin -lresolv/CoreFoundation/Security tbd stubs) — CHANGES: becomes the helper build (zig c++ links feov-ocr itself); `env`'s cgo contract and the Go-specific darwin stubs (Security/CoreFoundation exist for Go's crypto/x509) disappear or are re-derived for a C++-only link.
- `build-cstack.sh check <binary>` `:53-54` -> `go run staticproof` — unchanged mechanism, now aimed at helpers, not Go binaries.
- `$F/third_party/pins/staticproof/main.go:4,109-124` PE allowlist {kernel32, ntdll, api-ms-win-crt-*}, Mach-O {libSystem.B, libresolv.9, CoreFoundation} — CHANGES: allowlist per helper; an MSVC-built feov-embed (ONNX Runtime) will import other system DLLs (e.g. advapi32/dbghelp/ole32 — unmeasured) and an MSVC /MT CRT, so the PE allowlist becomes per-artifact.
- `$F/third_party/pins/PINS.txt:8,26-28` pins; comment "go:embed'ed at internal/tessocr/eng.traineddata; TestEmbeddedTraineddataPin" — moves with the traineddata; ONNX Runtime pins join (and MSVC toolchain cannot be pinned by sha the same way — zig is a tarball, MSVC is a runner image).
- `.github/actions/go-cache` — no C-stack reference (checked) — unchanged.

---------------------------------------------------------------------------------------------------
### F. CI workflow (.github/workflows/hooks.yml) — 3 jobs build the C stack, 1 publishes, 2 contract tests parse it

Enumerate: `grep -n "^  [a-z0-9_-]*:$\|tessocr\|cstack\|linkmode\|for cmd in\|for pair in\|-o \"" .github/workflows/hooks.yml`

- `hooks.yml:275-328` job `ocr-corpus` (ubuntu only): steps "C stack for the OCR engine (cache)" :301 (actions/cache key `cstack-${{runner.os}}-hashFiles(PINS.txt, build-cstack.sh)`, skipped on push), "(build on miss)" :308, "the corpus, real engine, statically linked" :314 runs `go test -tags tessocr -ldflags '-linkmode external -extldflags "-static"' -run '^TestCorpusGoldens$' ./internal/fetchcache/` — CHANGES: build feov-ocr (cached), plain `go test` pointing at it.
- `hooks.yml:387-` job `feov-record` matrix [ubuntu-latest, windows-latest]; `:533-548` comment (Windows legs prove the STUB compiles); `:550` cache step; `:557` build on miss (Linux only); `:570-614` "the OCR engine's own suite, real engine, statically linked": tagged flags `:577`, engine_tests DERIVED by `grep -l '^//go:build tessocr' internal/fetchcache/*_test.go` `:588`, runs ./internal/tessocr ./internal/libcprobe and narrowed fetchcache/seatprobe, greps for TestScannedFixtureSpanIsFoundOnPage1ByTheRealEngine and TestEveryProbeBoardBuildsThroughASubprocessWithASeparatedRecord — CHANGES: the tag-derived test list loses its key (needs a new derivation, e.g. a helper-required marker); Windows leg can now exercise the real client against a fake helper.
- `hooks.yml:726-823` job `release-fuzz-build`: `go build -o "$out/bin/" ./cmd/...` plain — unchanged (fuzz never ran OCR).
- `hooks.yml:1159-1216` job `fetch-bin-platforms` (ubuntu/macos/windows): fixture plugin with `tools/cmd/alpha,beta`, asset `<n>_<os>_<arch><ext>` + SHA256SUMS over file:// — unchanged unless fetch-bin.sh learns helpers; then a helper fixture joins.
- `hooks.yml:1355-1364` job `scripts` comment: sc-doctor discovers binaries with ReadDir(<plugin>/tools/cmd) — constraint: helper source must NOT live under tools/cmd.
- `hooks.yml:1586-1601` `debate-sim` guard steps (fetch-bin copies, buildid, marketplace pinning, event schema) — unchanged.
- `hooks.yml:1681-1700` job `release` (ubuntu-latest, ONE job, `needs: [test, feov-record, gray-area]`, setup-go cache:false — "A RELEASE READS NO CACHE").
- `hooks.yml:1786-1828` step "Cross-compile": `for cmd in cmd/*/` x `for pair in windows/amd64 windows/arm64 darwin/amd64 darwin/arm64 linux/amd64 linux/arm64`; CAPABILITY-KEYED `if [ -x third_party/pins/build-cstack.sh ]` -> builds stack per target, `go build -tags tessocr -trimpath -ldflags "-s -w -linkmode external -extldflags '$CSTACK_EXTLDFLAGS'"` for EVERY FEOV command (not only feov-record; :1814-1820 darwin feov-sitting-write strerror story, FEOV 1.68.0), then `build-cstack.sh check` each; else plain `GOOS GOARCH go build -trimpath -ldflags "-s -w"`; `cd $DIST && sha256sum * > SHA256SUMS` :1834 — CHANGES: every FEOV cmd takes the plain branch; helper builds are a separate loop (not cmd/*), zig per target on Linux; MSVC windows feov-embed CANNOT run in this ubuntu job.
- `hooks.yml:1863-1883` step "Release": idempotent re-run compares downloaded SHA256SUMS byte-for-byte with what this run built, else fails "artifacts DIFFER"; `gh release create "$TAG" "$DIST"/*` from ONE directory — CHANGES: a windows MSVC job must hand artifacts to this job (upload/download-artifact) BEFORE SHA256SUMS is computed; and MSVC/PE output must be reproducible (/Brepro, no timestamps) or every re-run of a release fails as "differs".
- `.github/workflows/race-nightly.yml`, `nightly-alarm.yml`, `labels.yml` — no C-stack reference (checked) — unchanged.

---------------------------------------------------------------------------------------------------
### G. Binary enumeration / provisioning (tools/cmd as the list of shipped binaries) — 9 enumerators

Enumerate: `grep -rn 'tools/cmd' scripts plugins/*/hooks plugins/prosthetic-conscience/tools/internal/doctor .github --include=* | grep -v _test`

All of these treat "shipped binary" == "directory under plugins/<p>/tools/cmd/, built by `go build ./cmd/<name>`". A C++ helper fits none of them.

- `scripts/fetchbingen/fetch-bin.sh:122-133` up_to_date: every tools/cmd/<n> has bin/<n>[.exe] and .fetch/installed == tag; `:225-239` missing list from tools/cmd; `:246-256` downloads SHA256SUMS, asset `${n}_${os}_${arch}${ext}`, verifies EVERY asset before installing ANY (all or none); `:265` stamps .fetch/installed — CHANGES only if helpers are fetched by fetch-bin.sh (then helpers need their own list, and "all or none" would make every hook wait on a 10–100 MB helper download); if feov-record fetches helpers itself, fetch-bin.sh is unchanged but there are now TWO fetchers with two stamps.
- `plugins/{frank-exchange-of-views,gray-area,prosthetic-conscience}/hooks/fetch-bin.sh` (lines 113-234) — GENERATED copies of the above (`go -C scripts run ./fetchbingen`, gate "Every plugin's fetch-bin.sh matches the one authored copy", hooks.yml:1589).
- `scripts/fetchbingen/main.go:120-138` targets(): plugins with tools/cmd — unchanged.
- `plugins/prosthetic-conscience/tools/internal/doctor/main.go:60` AssetName `<name>_<goos>_<goarch>[.exe]`; `:73-91` binariesOf ReadDir(tools/cmd); `:102` siblingBinaries walks every installed SC plugin's tools/cmd; `:490` fetchRelease via `gh release download --pattern <asset> --pattern SHA256SUMS`; `:533-535` buildFromSource `go build -C tools ./cmd/<name>` fallback; `:550` stampInstalled — CHANGES: doctor cannot see, check, fetch or BUILD a helper (no go build fallback for C++); needs a declared helper list (requirements.json?) and a fetch-only path, or explicitly reports helpers as not its concern.
- `plugins/prosthetic-conscience/tools/internal/doctor/release_contract_test.go:64-74` regex-parses hooks.yml's release job for `for pair in …;` and the `-o "…/<template>"` (`reOutput`), `:241` TestSHA256SUMSCoversEveryContractedName — CHANGES: the release job restructure must keep the literal `for pair in` and the `-o` template in the release job's FIRST -o (it takes `reOutput.FindStringSubmatch(wf[start:])` — a helper build step placed before the Go loop with its own `-o "..."` would be matched instead).
- `plugins/prosthetic-conscience/tools/internal/doctor/fetchbin_contract_test.go:74,218,287` — contract tests over fetch-bin.sh — change iff fetch-bin.sh changes.
- `scripts/pluginparity/main.go:121-143` counts tools/cmd dirs vs docs/setup-script.md ("22 at the time of writing", setup-script.md:148); `:147-172` requirements.json `_hook_binaries` vs tools/cmd (PC requirements.json:65, gray-area :28; FEOV declares none); `:214-` shippedHarnessProblems ("Every directory under plugins/*/tools/cmd/ ships") — unchanged as long as helpers stay out of tools/cmd; a new helper list would want a parity check of its own.
- `scripts/bootstrap-plugins.sh:85-97` cold bootstrap: `go build ./cmd/<name>` for every cmd in every cached plugin — unchanged for feov-record (already plain build -> today NO OCR in bootstrapped sessions); helpers never built here.
- `scripts/universe.sh:430-449` builds every cmd from the CHECKOUT with plain go build into cache + staged marketplace, stamps .fetch/installed (:470-) — CHANGES/trap: today a universe has NO OCR engine (stub); after, a first-use helper fetch would pull the PUBLISHED release's helper, not the checkout's — violates "prove the universe carries your change"; universe.sh needs a "build helpers from checkout" path or an explicit override env.
- `scripts/marketplacegen/main.go:15` "A plugin without tools/cmd has no binaries and no tag" — unchanged.
- `scripts/versionguard/main.go` — tag/manifest agreement, per-binary version constant sweep; no tools/cmd or tag assumption about C (checked) — unchanged; but if helpers get their own tag namespace (`feov-ocr--v…`) the release job's "tag names a plugin with tools/go.mod" check (hooks.yml:1718-1729) refuses it.
- `scripts/internal/gcpolicy/gcpolicy.go:4` every main under tools/cmd calls Short/Long — unchanged (helpers are not Go).
- `$F/internal/setup/run.go:214-232` setup preflight locates feov-record beside itself and checks its schema epoch — candidate consumer: the natural place to preflight (or prefetch) feov-ocr so a run does not discover a missing helper mid-sitting; today unchanged.

---------------------------------------------------------------------------------------------------
### H. Agent-facing text naming the in-process engine — 1 source sentence, 9 generated carriers, 3 help pages

Enumerate: `grep -rn "compiled into this binary\|compiled in\b\|tags tessocr" plugins/frank-exchange-of-views --include=*.go --include=*.md | grep -v _test`

- `$F/internal/cli/fetch.go:62` SOURCE of "read by the LOCAL OCR ENGINE compiled into this binary" — rewrite; regenerate with `go -C scripts run ./agentgen`.
- `plugins/frank-exchange-of-views/agents/{blue-researcher:553, blue-synthesizer:469, blue-lane:265, red-lens-voice:347, red-lens-logic:346, red-lens-architecture:346, red-lens-adversary:347, red-lens-computation:349, red-lens-evidence:348, red-lens-dark-side:346}.md` — GENERATED copies (agentgen -check) — regenerate, never hand-edit.
- `$F/internal/difftest/testdata/manual_runs_every_help_page_live.golden:960` — renders the fetch help — regenerate (UPDATE_GOLDENS=1), read the changed line.
- `$F/internal/cli/ocr.go:244` `ocr read` Long help — rewrite (operator tree, not a seat verb; still in manual golden).
- `$F/internal/fetchcache/pageread.go:92-93` error a seat reads in ocr_reason: "not compiled into this binary … build with -tags tessocr" — rewrite (seat-facing via fetch summary).
- `$F/internal/seatprobe/build.go:360`, `scanned.go:29` — probe-harness error text — rewrite.
- `$F/internal/cli/fetchsummary.go:91,412` `ocr_engine_absent` field + comment — field name stays meaningful ("engine absent"), meaning widens.

---------------------------------------------------------------------------------------------------
### I. Gates and dev tooling that encode the tag — 3 sites

Enumerate: `grep -n "tessocr\|C stack" scripts/check/*.go`

- `scripts/check/gates.go:326-331` gate `ocr-corpus` kindExtern skip text "needs the OCR engine's C stack… run the tagged suite… ANY edit under internal/tessocr re-arms it… Identity() hashes that package's source" — rewrite skip; re-arm set widens to the helper source.
- `scripts/check/gates.go:167-174` raceScope STATED LIMIT: "the engine behind -tags tessocr is not race-instrumented anywhere… musl cgo objects under the race build's linking broke…" — CHANGES (improves): with no cgo, the race gate covers the Go client + pipe protocol; the limit text goes (helper itself is not raced — say so).
- `scripts/deadcode` — runs the default build; libcprobe/engine_cgo never seen — unchanged; the new Go protocol client becomes visible to it (must be reachable from cmd/feov-record).

---------------------------------------------------------------------------------------------------
### J. Engine identity — where it is computed, stored, compared — 1 producer, 4 stores, 3 comparers

Enumerate: `grep -rn "Identity()\|OcrEngine\|ocr_engine\|\.Engine\b" $F/internal --include=*.go | grep -v "s.Identity()\|_test\|record.pb.go\|decode_gen\|hookcmd"`

Producer:
- `$F/internal/tessocr/tessocr.go:100` Identity() = `tesseract@<pin>+leptonica@<pin>+eng@<12 of traineddata sha>+cstack@<cstackPin>+tessocr@<sha256[:12] of package non-test source incl shim.cpp/shim.h>` — CHANGES: shim.* leave the package, so the hash no longer covers the C++; the helper's own identity (its pinned asset sha256, or a self-reported build id) must join the string, e.g. `+feov-ocr@<sha>`. Constraint from group C: Identity() is called on cache-reuse paths before any read, so it must be computable WITHOUT spawning or fetching the helper (derive from the pinned helper hash the fetcher verifies against, not from a handshake).
- `$F/internal/fetchcache/pageread.go:66` TessocrPageEngine.Identity — delegates — swaps.
Stores (all plain strings; no schema change, no epoch bump, no migrate):
- `$F/internal/fetchcache/pageread.go:405-422` ReadingRecord.Engine `json:"engine"` in `<run>/cache/…/pages/reading.json`; per-page receipt `:330` Engine — cache only (not in run-archive) — values change; every cached reading re-reads once.
- `$F/internal/record/recordpb/record.proto:1289-1296` Cite.ocr_engine (field 14) + ocr_text_sha (15), stamped by `$F/internal/cli/blue/cite.go:161` from pagelocate.go:144 — IN THE RECORD and run-archive; old runs keep old strings; nothing compares a stored ocr_engine with the current Identity (record.go:1677-1681 only requires non-empty; evidenceview.go:121,504 displays it) — unchanged code; value grammar changes.
- `$F/internal/record/migrate/remap.go:128` lists Cite.ocr_engine among carried fields — unchanged.
- `$F/internal/tessocr/testdata/corpus/*/reading.golden:1` (13 files) — regenerate.
Comparers:
- `pageread.go:299` page receipt reuse; `readscanned.go:108` whole-reading reuse — unchanged logic.
- `$F/internal/cli/ocr.go:210,297`, `fetchsummary.go:175,416` print the engine string — unchanged.
Note: PDF rendering identity is separate — `$F/internal/fetchcache/pdfextract.go:67-81` extractorIdentity = go-pdfium (WebAssembly via wazero, pure Go) — unaffected; PDFium is not part of the C stack.

---------------------------------------------------------------------------------------------------
### K. Docs and plans describing the in-process engine — 3 docs, 1 live plan, 9 OCR plans (historical-in-effect)

Enumerate: `grep -rln "tessocr\|build-cstack\|linkmode external\|compiled into the binary" docs plans README.md plugins/*/README.md`

Docs (live, must be rewritten):
- `docs/c-stack.md:1-46` whole doc: "feov-record links a static C/C++ stack… behind the tessocr build tag… Go build links with -linkmode external, zig as external linker, for every command"; "Adding a C or C++ library… If it builds only with a platform's native toolchain, the release job needs that platform's runner, and that is a decision" — rewrite: stack builds helpers; MSVC windows runner for feov-embed IS that decision.
- `docs/local-development.md:302-305` "A plain build has no OCR engine. Without -tags tessocr…" — rewrite (plain build has the client; OCR needs a helper).
- `docs/setup-script.md:73` "reads a scan with the OCR engine compiled into the binary, so nothing needs provisioning" — FALSE after; helpers need provisioning (fetch on first use); `:148` "22 hook binaries" count — unchanged if helpers stay out of tools/cmd.
Plans:
- `plans/vector-search.md` — LIVE ("approved 2026-11-03, not started", plans/README.md:22): designs llama.cpp statically linked INTO feov-record via build-cstack.sh/-tags tessocr, internal/vecembed cgo shim, GGUF model beside feov-record provisioned by fetch-bin.sh + sc-doctor (:8,:61,:71,:88-96,:114,:329-332,:458-506) — SUPERSEDED in its native half by this design (ONNX Runtime helper, MSVC windows); must be revised or the new plan must state it replaces those sections.
- `plans/local-ocr.md` — header still "under review" but SHIPPED (code cites it: tessocr.go, page.go, lattice.go, pageread.go, pagerender.go, ocr.go, hooks.yml x2, receipts_test.go, fuzz_test.go, promptsize_test.go) — historical record; do not rewrite; code comments citing "Wave 1" for the tag split need rewording where they describe the build.
- Implemented/measured, verification loops hold tagged commands (historical, leave): `plans/ocr-dashed-rules.md:389-391,408`, `ocr-dpi-derivation.md:236-237`, `ocr-page-corpus.md:263-264`, `ocr-background-norm.md:97-98`, `ocr-cite-derives-page.md:515,868-914` (draft status but implemented), `ocr-text-cell-rows.md:182`, `ocr-resolution-and-antialiasing.md:65,161` (REJECTED), `ocr-declare-the-resolution.md` (NOT SHIPPED), `ocr-level-subcolumns.md`, `ocr-end-to-end-evidence.md` (proposed), `ocr-blocked-page-resilience.md` (superseded).
- `plans/packaging-goreleaser.md` — not started; describes the hand-rolled release loop — a helper/multi-runner release strengthens or kills its case; note only.

---------------------------------------------------------------------------------------------------
### L. Explicitly unchanged (checked)

- Other cgo in the tree: only `internal/tessocr/engine_cgo.go` and `internal/libcprobe/probe_cgo.go` contain `import "C"` (`grep -rln 'import "C"' plugins --include=*.go`). SQLite is modernc (pure Go); PDFium is wazero/WASM.
- gray-area and prosthetic-conscience modules: no C stack, no tag.
- `.qlty`, `.gitattributes`, `.gitignore`: no tessocr/shim/cstack entries (note: shim.cpp/.h are NOT LF-pinned, which is why identity folds CRLF — moving them keeps that need).
- Event schema / record.proto DDL: no field added or changed by the move (identity is a string value) — no epoch bump.
- release-fuzz (4 parts + tally): builds plain, never exercised OCR — unchanged.
