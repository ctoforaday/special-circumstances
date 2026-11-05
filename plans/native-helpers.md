# Native code leaves `feov-record`: the `feov-ocr` and `feov-embed` helpers

> STATUS: proposed 2026-11-04, not started. F1–F10 ruled by gblock 2026-11-04, F11–F12 on 2026-11-05 (§I.4). Revised for
> audit rounds 1 and 2 (`~/.claude/scratch/vecplan/audit/round1/`, `round2/`); next is audit round 3. Line numbers are
> against `origin/main` at `68a735d51` unless a census group says otherwise.
> Ownership (gblock, 2026-11-04, "You'll handle vector here but hand ocr to the agent working on ocr"; F4): the OCR
> session builds Part 1 (helper mechanics) inside Part 2 (`feov-ocr`); the vector-search session builds Part 3
> (`feov-embed`) on top of Part 1 once it lands.
> Precondition (§II.4, §IV R11): F1 needs `fetch-bin.sh` to fetch in an installed plugin. On `origin/main` at
> `68a735d51` it refuses whenever `$ROOT/tools/go.mod` exists (`scripts/fetchbingen/fetch-bin.sh:24-28`, PR #1724),
> which every installed plugin has. That guard is not this plan's change; Part 1's `fetch-bin.sh` work waits for it.

## I. Summary & Goals

### 1. Objective

FEOV's C and C++ code moves out of the Go binary into two native helper executables, each shipped per
target and fetched with the hook binaries by `fetch-bin.sh` (F1). `fetch-bin.sh` verifies each asset against the
release's `SHA256SUMS`; `feov-record` accepts a helper only if the helper's handshake reports the source digest
compiled into `feov-record` (F8, §III.1.4).

- **`feov-ocr`** carries tesseract 5.5.3 and leptonica 1.87.0 (today's `internal/tessocr` cgo engine, `shim.cpp`, and
  the embedded `eng.traineddata`).
- **`feov-embed`** carries ONNX Runtime 1.31.1 (reduced operator set) and runs gte-small for hybrid search (#1171,
  `plans/vector-search.md`).

Each helper is a **thin C++ kernel**: it decodes a request frame, calls the library, and writes a response frame on
a private descriptor. All logic stays in Go. Rendering, cropping, rotation, table reconstruction and tokenization never
move into a helper. That rule is what keeps the native code small enough that C++ is the right language for it (§II.1).

Every published Go asset of every plugin is built with `CGO_ENABLED=0` and proven static in the release job before
`SHA256SUMS` is written (§III.1.5).

### 2. Why not leave it as it is (the value case)

The naive alternative is today's shape: link every C/C++ library into `feov-record` with zig, and for ONNX Runtime
load the official DLL on Windows.

| | In-process static, plus a Windows DLL | Helpers (this plan) |
|---|---|---|
| Windows embedding | Official `onnxruntime.dll`, plus `msvcp140`/`vcruntime140` | Static MSVC `/MT` executable, 3.67 MB, OS DLLs only (Part 0, F3) |
| A tesseract crash | Kills `feov-record` and the seat's command | The helper exits; Go reports that page as failed and the next request starts a fresh helper |
| Race detector | Covers the stub, never the real engine (`scripts/check/gates.go:167-174`, a recorded open decision) | Covers everything that ships in Go; the C++ sits behind a process boundary |
| Go link | `-linkmode external` with zig, for every FEOV command (`hooks.yml:1799-1817`) | `CGO_ENABLED=0 go build` |
| Download | The engine rides inside the 41,032,712-byte `feov-record` (v1.73.0, linux_amd64) | Separate assets, fetched with the hook binaries (F1); `feov-record` shrinks by the engine's share (unmeasured) |
| Each new C++ library | Must build under zig for all six targets inside the Go link | Builds per helper, with the native toolchain where zig cannot |
| Pre-publication execution | Darwin and Windows engines ship never executed | Every helper asset runs on real data on its own OS and arch before publication (F7) |

**What it costs:**
- A frame protocol, with a Go client and a C++ codec.
- A second asset list in `fetch-bin.sh` and `sc-doctor`, beside `tools/cmd`.
- Four CI jobs (`native-scope`, `native-build`, `native-run`, `native-fuzz`), a release job that waits on them for a FEOV
  tag, and a helper build on every event that runs the engine tests (F11).
- Re-pinned OCR goldens, and a rewritten `docs/c-stack.md`.

The value case holds only if the protocol stays small and the OCR goldens come through unchanged. §V checks both.

The measurements behind it are on #1171 (comment 6621511526), in `~/.claude/scratch/vecplan/ortstatic/REPORT.md`
(the static ONNX Runtime proof) and in `~/.claude/scratch/vecplan/ortq/RESULT.md` (8-bit quality).

### 3. Success criteria

- **S1.** Every Go asset the release job publishes, for every plugin and all six targets, is built with
  `CGO_ENABLED=0` and passes `staticproof` in that job before `SHA256SUMS` is written. A cgo-off build of
  `feov-record` at `cffb72ab1` passes today's allowlists on linux/amd64, darwin/arm64 and windows/amd64
  (`~/.claude/scratch/vecplan/cgo0/out.txt`); a plain host build on ubuntu is dynamically linked
  (`audit/round1/feov-record-plain`).
- **S2.** Every helper asset passes `staticproof` against its own allowlist on all six targets (§III.1.5).
- **S3.** OCR is behaviour-preserving on linux-amd64, the target the goldens are pinned on. Each of the 13 corpus
  goldens (`internal/tessocr/testdata/corpus/*/reading.golden`) differs from main on line 1 only, the engine identity,
  and the remaining bytes are identical. Other targets are held to the bar in §III.1.5 `native-run`.
- **S4.** OCR wall time per corpus page through `feov-ocr` is within 5% of the in-process engine on the same runner,
  measured with the two interleaved.
- **S5.** `feov-embed` int8 fidelity against `ref_gte.npy` is mean ≥ 0.984 and min ≥ 0.966 on every target, measured on
  the helper and model export that target publishes. These are the measured values. Retrieval quality, not cosine, is
  the bar for 8-bit, and it passed on 2026-11-04 (hybrid beats BM25 on SciFact +0.048 [0.026, 0.069] and on NFCorpus
  +0.040 [0.027, 0.054]).
- **S6.** `feov-embed` one-thread speed on 200-word chunks is ≥ 5 chunks/s on linux-amd64, linux-arm64 and darwin-arm64.
  - Measured with the reduced Release build: 7.1 chunks/s on an EPYC 7763.
  - Measured with the `-Os` build: 5.9 chunks/s on Neoverse N2 and 5.9 on M1.
  - darwin-amd64 measured 4.2 chunks/s on an i7-8700B, below the bar. That is stated, not hidden.
- **S7.** A helper that is missing, reports another source digest, crashes, overruns its deadline or breaks the
  protocol produces a loud, named refusal. It never comes back as an empty reading or an empty search result
  (§III.1.3).
- **S8.** No build path silently refuses a matching helper in `bin/`, and none silently runs a helper built from other
  sources (§III.1.4 table).
- **S9.** No universe build state leads `fetch-bin.sh` or `sc-doctor` to install a published asset into the universe
  (F5, §III.1.6).

### 4. Rulings

Rulings are gblock's, 2026-11-04 (F1–F10) and 2026-11-05 (F11–F12), picked from options the agent wrote; the option
wording is the agent's, not a quote.

- **F1, provisioning: `fetch-bin.sh`, eagerly.** The helpers and the model are fetched with the hook binaries, by the
  same all-or-none install (`scripts/fetchbingen/fetch-bin.sh:252-286`). A first hook start therefore
  waits on them too: the model alone is 33,844,232 bytes. This keeps the 2026-10-31 choice of shipping the model as a
  release asset through `fetch-bin.sh`. The recommended alternative, a fetch by `feov-record` on first use, was not
  taken.
- **F2, release cadence: the same FEOV release.** Helpers are built for every FEOV release tag. The release reads no
  cache (`hooks.yml:1696-1700`), so every FEOV release rebuilds both helpers from pinned sources: 15–25 minutes per
  target by the agent's estimate, run in parallel.
- **F3, Windows `feov-embed`: static, with the allowlist widened for it.** Part 0 proved it on both arches: built
  with MSVC and `/MT`, the helper imports only KERNEL32, ADVAPI32, dbghelp and `api-ms-win-core-path-l1-1-0`, all
  shipped with Windows. `staticproof`'s PE allowlist becomes per-artifact and admits those three extras for `feov-embed`
  only, each with its reason. This replaces the earlier "Windows loads a dll" ruling, which predated the helper design.
- **F4, Part 1: built by the OCR session inside Part 2.** OCR proves the mechanics with working code first, and
  Part 3 (`feov-embed`) waits for Part 1 to land.
- **F5, universe runs: `universe.sh build --native`.** It builds both helpers from the checkout and installs them where
  `fetch-bin.sh` would. A plain `build` has no helpers, so OCR and search refuse loudly. That matches today, where a
  universe has no OCR engine at all (census G, `scripts/universe.sh:437-453`). A universe never fetches the *published*
  helpers (AGENTS.md, "prove the universe carries your change").
- **F6, release shape: a per-target native matrix.** Helper build jobs run per target on Linux, macOS and Windows
  runners and hand their artifacts to the release job.
- **F7, execution before publish: yes.** Per-target run jobs execute each helper on real data on its own OS and arch
  before the release job publishes it.
- **F8, helper identity: a source digest.** The identity is a hash of the pins, the `native/` sources and the build
  script, in the pattern of today's `cstack@`.
- **F9, `setup` preflight on a missing or mismatched helper: proceed.** `setup` states which helpers are unavailable,
  and OCR and search refuse per call.
- **F10, model hosting: a one-time release on this repository.** The exports are not reproducible across hosts (int8
  fidelity 0.984411–0.984460 across CI exports, `ortstatic/REPORT.md` §8), so the release job cannot re-export them per
  release without breaking the idempotent re-run (`hooks.yml:1860-1870`). One release holds only the two exports. It is
  created by hand with `gh release create --latest=false`, so it never shows as the repository's latest, and it is never
  re-tagged. Its tag does not match the workflow's `*--v*` trigger. Each export's sha256 and URL are pinned in
  `PINS.txt`, and the release job downloads them from there.
- **F11, helper builds per event: build or restore on every event that runs the engine tests.** That covers pull
  requests, pushes, the nightly and dispatch, from a cache seeded on main and rebuilt on a miss (§III.1.5). Every pull
  request keeps today's real-engine coverage.
- **F12, the release job's new steps: first executed on a real FEOV tag.** No dry run. A failure there is loud, and
  release failures are fixed as they arise. This is stated as residue in §V.6.
- **F13, the helper cache on a push to main: restored (gblock, 2026-11-05).** A main push builds only the test targets
  (linux-amd64, windows-amd64), and it restores main's helper cache instead of rebuilding them. A tag still reads no
  cache (F2). This departs from `.github/actions/go-cache`'s push rule for this one artifact, and the cache step says so.
- **F14, unbounded page area: accepted as residue (gblock, 2026-11-05).** A page far larger than the corpus's largest
  fails alone and loudly as `helper-deadline`; the rest of the document reads (§III Part 2).

Decisions the lead took on the round-1 design gaps, reversible and disclosed to gblock (not gblock's rulings):

- **L1 (gap 2).** The helper's handshake reports its source digest. `feov-record` accepts a helper only if that digest
  equals its own compiled-in digest, on every build path, release or dev. The digest is a committed generated constant
  with a test that recomputes it from `native/` and `PINS.txt` (the `TestIdentityMatchesPins` pattern; FEOV's tools
  module cannot `go:embed` files outside `tools/`). `Identity()` uses the compiled digest and never spawns. Release
  asset sha256 stays `fetch-bin.sh`'s `SHA256SUMS` check; there are no `-X` hashes.
- **L2 (gap 8).** The model ships per arch, `gte-small.int8.amd64.ort` and `gte-small.int8.arm.ort`, each measured.

### 5. Non-goals

- A Rust rewrite. §II.1 says why C++.
- Moving any Go logic into a helper.
- GPU execution providers.
- Changing how OCR reads a page. Part 2 is a move, and S3 is its proof.
- Retiring PDFium. It is WebAssembly under wazero, pure Go, and outside the C stack (census J).
- Narrowing `fetch-bin.sh`'s source-tree guard (PR #1724). It is a precondition, owned elsewhere.

## II. Technical Context

### 1. Language: C++, not Rust

- **The libraries are C++, and the existing code is too.** Their APIs are called directly, with no binding crate in
  between. `internal/tessocr/shim.cpp` is most of `feov-ocr` already.
- **No new shipped toolchain.** Linux and macOS use the pinned `zig c++` (`third_party/pins/PINS.txt`). Windows
  `feov-embed` uses MSVC on a Windows runner, which is how ONNX Runtime builds there. Windows `feov-ocr` stays on zig and
  MinGW, as it is today. The fuzz build uses the runner's clang, which ships nothing (§III.1.5 `native-fuzz`).
- **Rust's safety would protect the wrong layer.** The untrusted input is pixels from fetched documents, which are
  parsed inside tesseract and leptonica, and Rust glue does not make those safer. The C++ glue's own exposure is the
  frame decoder, which is fuzzed under AddressSanitizer and UndefinedBehaviorSanitizer in CI (§V.3).
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
values per text come back, mean-pooled and L2-normalized in the helper. The model is a separate asset per arch (L2):
`gte-small.int8.amd64.ort` for amd64 targets and `gte-small.int8.arm.ort` for arm64 targets, 33,844,232 bytes each,
same operator set, different bytes (`ortstatic/REPORT.md` §2). Every arm64 figure in S5 and S6 was measured on the arm
export.

### 4. Constraints

- **Tests refuse the network** (`nonet.OnlyLoopback`, `internal/testbuild/testbuild.go:100`), so no test can fetch a
  helper. Tests take a prebuilt helper from the directory `NATIVE_HELPER_DIR` names (§III.1.4).
- **The variable's name must not start with `FEOV_`.** AGENTS.md has every agent session run
  `unset $(env | grep -o '^FEOV_[A-Z_]*')` before tests, and a `FEOV_` name would be swept away with the rest.
- **No schema change.** The engine identity is a string value in `Cite.ocr_engine` and in cache receipts, so there is
  no epoch bump and no `feov-migrate` translation (census J, L). Old runs keep their old identity strings, and nothing
  compares a stored `ocr_engine` with the current identity.
- **The release fuzz never exercised OCR.** `release-fuzz-build` is a plain build (census F), so it is unchanged.
- **Precondition: `fetch-bin.sh` must fetch in an installed plugin.** On `68a735d51` its guard
  (`scripts/fetchbingen/fetch-bin.sh:24-28`, plugin copies `:25-29`) exits 0 whenever `$ROOT/tools/go.mod` exists, so it
  never fetches for a consumer and F1 cannot work. The `fetch-bin-platforms` fixture (`hooks.yml:1182`) has no
  `tools/go.mod`, so CI does not see it. Part 1's `fetch-bin.sh` and `sc-doctor` work starts once that guard is
  narrowed; the auditor reported it to gblock.
- **Runner images.** `windows-latest` moves its MSVC toolset with image updates. The Windows builders pin the image
  label (`windows-2025`, `windows-11-arm`; Part 0 ran on `windows-latest` and `windows-11-arm`) and select the MSVC toolset from `PINS.txt` with
  `vcvarsall -vcvars_ver=<pin>`, failing when the image no longer carries it. An image update therefore fails at the
  build, loudly, never as an idempotent re-run's "artifacts DIFFER".

## III. Proposed Changes (the spec)

### Part 0: Windows feasibility probe (done, scratch only)

ONNX Runtime 1.31.1 reduced, built with MSVC 14.51 and `--enable_msvc_static_runtime`
(plus `onnxruntime_DISABLE_DEVICE_DISCOVERY=ON`, which drops `dxgi` and `setupapi`), and linked `/MT` into a probe
`embed-probe.exe`. Runs 41798616860, 41803563965 and 41808776479 of `ortmsvc.yml` on the scratch branch
`bench/embed-engines`; run 3 is the source of the import list. Both runner labels exist and ran it.

| | windows-amd64 (`windows-latest`) | windows-arm64 (`windows-11-arm`) |
|---|---|---|
| DLL imports | KERNEL32, ADVAPI32, dbghelp, api-ms-win-core-path-l1-1-0 | the same four |
| Helper size | 3,665,408 B | 3,670,528 B |
| int8 fidelity, mean / min | 0.984411 / 0.968026 (amd64 export) | 0.984464 / 0.968446 (arm export) |
| f32 fidelity | 1.000000 | 1.000000 |
| 1 thread, int8 / f32 chunks/s | 6.0–7.7 / 4.4–5.6 (EPYC 7763 and 9V74, host-dependent) | 6.4 / 2.75 (Cobalt 100, stable) |
| ONNX Runtime build | 22–31 min | 25–28 min |

**Residue:**
- The probe ran only on the build runner images, never on a clean Windows install without Visual Studio.
  `native-run` adds that for windows-amd64 (§III.1.5).
- The objects that import ADVAPI32 (event tracing), dbghelp (stack traces) and PathCch were identified from symbol
  names, not traced.

### Part 1: Helper mechanics (owner: the OCR session, inside Part 2; F4)

#### 1.1 Layout

```
plugins/frank-exchange-of-views/native/          [NEW] helper sources, outside tools/cmd (census G: every enumerator treats tools/cmd as a shipped Go binary)
  frame/frame.h, frame.cpp                       [NEW] C++ codec for the frame protocol, and the startup that claims the frame descriptor (§1.2)
  frame/fuzz_frame.cpp                           [NEW] libFuzzer target for the decoder
  feov-ocr/ASSETS   (Part 2)                     [NEW] empty: the file marks the directory as a helper
  feov-ocr/…        (Part 2)                     [NEW] shim.cpp, shim.h, main.cpp, the libc self-test
  feov-embed/ASSETS (Part 3)                     [NEW] names the per-arch model export (§1.4)
  feov-embed/…      (Part 3)                     [NEW] main.cpp, ops.config
  build.sh                                       [NEW] build one helper for one target; absorbs build-cstack.sh's build mode
plugins/frank-exchange-of-views/tools/internal/nativehelper/   [NEW] Go: locate, handshake, spawn, frame client, deadline, Require for tests
plugins/frank-exchange-of-views/tools/internal/nativehelper/digest_gen.go   [NEW, GENERATED] each helper's source digest and per-arch ASSETS sha256
plugins/frank-exchange-of-views/tools/internal/nativehelper/testdata/fakehelper/   [NEW] a Go helper on the Go codec; each test builds it with the digest, protocol version and misbehaviour it needs (no zig in the feov-record job)
plugins/frank-exchange-of-views/tools/devcmd/nativedigest/     [NEW] writes digest_gen.go; `-print <helper>` for build.sh and CI; `-check`
plugins/frank-exchange-of-views/tools/third_party/pins/        [MODIFY] PINS.txt gains ONNX Runtime, the MSVC toolset and (Part 3) the two model exports; build-cstack.sh is deleted into native/build.sh; staticproof leaves for scripts/
scripts/internal/staticproof/, scripts/staticproof/            [NEW, moved] the staticness check as a package with per-artifact allowlists, and its command
scripts/releaseassets/                                         [NEW] the release job's asset check and the native jobs' matrix (§1.5)
```

**A helper is a directory under `native/` that holds an `ASSETS` file**, named for the binary it builds. `frame/` holds
none and is not a helper. Every enumerator (`fetch-bin.sh`, `sc-doctor`, `releaseassets`, `nativedigest`) uses this
one predicate.

#### 1.2 Frame protocol

Each frame is:
- a fixed little-endian header: magic, version, op, request id, payload length;
- the payload: length-prefixed fields.

**The frame channel is a private descriptor, and only the codec writes it.** At startup, before any library is
initialised, the helper duplicates fd 1 to a new descriptor for frames (`dup`), then points fd 1 at fd 2 (`dup2`), so
a write to stdout by tesseract, leptonica or ONNX Runtime lands in stderr, never in a frame. On Windows the helper also
sets binary mode on stdin and on the frame descriptor (`_setmode(…, _O_BINARY)`). `_dup2(2, 1)` closes fd 1's original
OS handle, so frames are written with `WriteFile` on `_get_osfhandle(<frame descriptor>)`, the handle the duplicate
owns, never on a `GetStdHandle` value read before the redirect. No CRT text-mode translation touches `0x0A` or `0x1A`.

The first frame each way is a handshake. It carries the protocol version and the helper's source digest (F8, L1);
`feov-embed`'s also carries the sha256 of the model it loaded. A version mismatch is a refusal, with no negotiation. A
response carries:
- a status;
- the op's result;
- the helper's stderr captured during the request (the `diag` capture moves here, census A,
  `pageread.go:445 engine.log`). For the request, the capture points fd 1 and fd 2 together at the capture file, then
  restores both: fd 2 to stderr, and fd 1 to fd 2. A library's stdout write during a request lands in that request's
  capture. Redirecting fd 2 alone would leave fd 1 on the original stderr, outside the capture.

Every helper also answers the **self-test ops** `selftest-stdout` (writes to stdout through `printf`, `std::cout`,
`fwrite(stdout)` and `write(1)`, then returns a payload of all 256 byte values including `0x0D 0x0A` and `0x1A`) and
`selftest-sleep <ms>`. `feov-ocr` adds `selftest-traineddata` (the sha256 of its embedded `eng.traineddata`) and, on
Linux targets, `selftest-libc` (the bounded-scan probes of `internal/libcprobe/probe_cgo.go`, #1598, in C++).

**Lifecycle:**
- **Deadline.** The Go client applies one deadline per request, 120 s, the handshake included. It starts when the
  request frame is written; time spent waiting for the process-wide mutex is not counted. **Every request the client can
  send stays under a quarter of the deadline, 30 s, on the slowest target.** The client bounds each request's size, and
  `native-run` sends the largest request of every op on every target (§V.4):
  - `feov-ocr`: no op receives more than one whole page, and the client renders at most `MaxRenderDPI`, 600
    (`internal/fetchcache/pagerender.go:48`). The largest request is the largest corpus page at 600 DPI, given whole to
    each op. A 600-DPI page takes 12–24 s in total over about 8 + N requests on the measured hosts.
  - `feov-embed`: the client sends at most `embedBatch` texts per request, each at most 512 token ids, the model's
    position limit. Indexing is many requests, never one. `embedBatch` is 16, the lead's choice. At darwin-amd64's
    measured 4.2 chunks/s on 200-word chunks that is about 4 s, but a 512-id text costs more than a 200-word one, so
    the test decides, not this estimate.
  - The handshake: spawn to handshake reply. For `feov-embed` that includes loading and hashing the 33.8 MB model.
  - Residue: page area is not bounded. A page far larger than the corpus's largest, such as a large-format drawing, can
    still overrun at 600 DPI. That page fails with `helper-deadline`, loudly and for that page only.
- **Failure of a request.** On EOF, a non-zero exit, a protocol error or the deadline, the client kills the helper and
  reports *that request* as failed with its named error. The request is not retried: a page that crashed tesseract
  crashes it again, and `fetch` already reports a failed page per page.
- **Restart.** The next request spawns one fresh helper. A spawn whose handshake fails latches its error for the life of
  the `feov-record` process, so a broken helper costs one spawn per process, not one per request. No retry loop.
- **Shutdown.** The helper exits 0 on EOF on stdin; it relies on no signal (Windows has no SIGPIPE). The client closes
  stdin at `Close()`, waits 2 s, then kills.

#### 1.3 Absent, mismatched, crashed and stalled helpers (S7)

Each case is a distinct error:
- `helper-absent`: no executable at the looked-up path (§1.4), with the path named;
- `helper-digest-mismatch`: the handshake's digest differs from the compiled-in one, both named, or `feov-embed`'s model
  sha256 differs from the pin for this arch. The per-arch model pins are compiled in like the digest: `nativedigest`
  reads each `ASSETS` file's per-arch sha256 from `PINS.txt` into `digest_gen.go`, and `feov-record` picks the one for
  `runtime.GOARCH`;
- `helper-crashed`, with the exit status and the captured stderr;
- `helper-deadline`, with the op and the elapsed time;
- `helper-protocol`.

`fetch`'s `ocr_engine_absent` field (`internal/cli/fetchsummary.go:91`) keeps its name and means "no helper came up".
It is set by `helper-absent`, `helper-digest-mismatch`, and the latched handshake failure (a crash, deadline or protocol
error before the handshake reply). A `helper-crashed`, `helper-deadline` or `helper-protocol` on a request after a good
handshake fails that page only and leaves the field false. `ocr_reason` names the case either way. These replace `tessocr.ErrNotCompiledIn` and its text, "build
with -tags tessocr" (`pageread.go:85-95`, `tessocr.go:107-111`).

#### 1.4 Provisioning and identity (F1, F2, F8, L1, L2)

- **`fetch-bin.sh` learns a second list** (precondition §II.4). Today it enumerates `tools/cmd/*` (`fetch-bin.sh:143-154`
  in `up_to_date`, `:246-262` in `fetch`). It also enumerates the helpers (§1.1 predicate), which are present in the
  plugin cache because the whole plugin directory installs.
  - It globs `native/*/ASSETS` and keeps a match only if `[ -f ]` holds. An unmatched sh glob yields the pattern itself,
    so without the test a plugin with no `native/` (gray-area, prosthetic-conscience) would list a helper named `*`.
  - A helper's binary asset is `<helper>_<os>_<arch>[.exe]`, like the Go binaries.
  - Each non-empty `ASSETS` line names one extra installed file and, as separate fields, the release asset for each
    arch: `gte-small.int8.ort amd64=gte-small.int8.amd64.ort arm64=gte-small.int8.arm.ort`. `fetch-bin.sh` reads the
    field for its own `$arch` and fails, naming the file and arch, when there is none.
  - Every asset is in the same `SHA256SUMS`, and is verified before any is installed (the existing all-or-none rule).
  - Everything lands in `<plugin cache>/bin/` and is stamped by the existing `.fetch/installed`. `up_to_date` requires the
    helpers and the `ASSETS` files to be present, as it does each `tools/cmd` binary.
  - The edit is made in `scripts/fetchbingen/fetch-bin.sh` and regenerated into the three plugin copies.
- **`fetch-bin.sh` and `sc-doctor` never fetch into a source-built root.** `universe.sh build` writes
  `.fetch/source-built` into every root it populates. `fetch-bin.sh`'s `hook`, `ensure` and `fetch`, and `sc-doctor
  -fix`'s fetch, exit before any network access when it is present, and `ensure` and `-fix` say so in one line naming
  `universe.sh build --native`. If the narrowed #1724 guard already refuses a universe root, Part 1 uses that field and
  adds no marker: one concept, one mechanism.
- **Source digest (F8, L1).** For each helper, `devcmd/nativedigest` hashes, in a fixed order with CRLF folded to LF
  (as `cstackPin` does today): `native/<helper>/**`, `native/frame/**`, `native/build.sh` and `PINS.txt`. It writes
  every helper's digest, and each `ASSETS` file's per-arch sha256 (§1.3), into `internal/nativehelper/digest_gen.go`;
  `TestNativeDigestMatchesSources` recomputes both and fails on drift, with the value to write.
  - `native/build.sh` runs `nativedigest -check`, then `nativedigest -print <helper>`, and compiles the result into the
    helper (`-DFEOV_HELPER_DIGEST=…`). A helper built from a tree whose committed constant is stale does not build.
  - The digest covers the sources, so it does not depend on the build host; the build host's `__FILE__` paths no longer
    reach the identity.
- **Lookup and acceptance, one rule for every build.** If `NATIVE_HELPER_DIR` is set, the helper is looked up there and
  only there (no fallback to `bin/`, so a mis-set override is `helper-absent`, never a silent swap); `none` means no
  helper. Otherwise it is `bin/<helper>[.exe]` beside `os.Executable()`. Whatever is found is accepted only if its
  handshake digest equals the compiled-in one. The digest is a staleness check, not authenticity: authenticity of a
  release asset is `fetch-bin.sh`'s `SHA256SUMS` check, and a party that can set the environment can already replace
  `bin/`.
- **Every build path, under that rule (S8):**

  | Path | Where the helper comes from | How it is checked | Release binary honours the override? |
  |---|---|---|---|
  | Release install | `fetch-bin.sh` into `bin/` | `SHA256SUMS` at fetch; digest at handshake | yes, same rule |
  | `universe.sh build --native` | built from the checkout into `bin/` of every staged root | digest at handshake equals the checkout's constant | — |
  | `universe.sh build` (plain) | none; `.fetch/source-built` stops any fetch | `helper-absent`, loud | — |
  | `bootstrap-plugins.sh` (`:85-97`) | none at bootstrap; it writes no stamp, so the next `ensure` installs the tag's whole set, helpers included | `SHA256SUMS`; the cache's sources are the tag's, so the digests agree | — |
  | `sc-doctor -fix`, fetch or `buildFromSource` (`doctor/main.go:533-538`) | Go binaries fetched or built from the cache's sources; helpers fetched only, never built | `SHA256SUMS`; digests agree as above; `stampInstalled` (`:550`) stamps only when every Go binary, helper and `ASSETS` file is present | — |
  | `testbuild.Binary` (`testbuild.go:322-353`) | `NATIVE_HELPER_DIR`, passed through to the built binary | digest at handshake | — |
  | Plain `go build` in a checkout | `bin/` beside it, if a developer put one there | digest at handshake | — |

- **Identity can be computed without the helper.**
  - `Identity()` becomes `tesseract@<pin>+leptonica@<pin>+eng@<pin>+feov-ocr@<digest[:12]>+tessocr@<hash>`. The
    `feov-ocr@` part replaces `cstack@`, whose inputs it covers. `tessocr@` keeps hashing the Go source that remains.
  - It reads compiled constants only, so the cache-reuse paths that call it before any read (`pageread.go:299`,
    `readscanned.go:108`) never start the helper (census J), and it returns the same string whether or not a helper is
    installed. A cached reading made by the same sources is reused without one.
- **`setup` preflight (F9)** (`internal/setup/run.go:214-232`) spawns each helper once and runs the handshake. It looks
  the helpers up by the rule above, applied to `recordBin`, the `feov-record` the seats run (`cfg.BinDir`'s, else
  setup's own, `:214-222`), not to setup's own `os.Executable()`. The digest it compares is setup's compiled one, which
  is the seats' binary in the documented flow; with `--bin-dir` naming another build that is residue. It never refuses
  the run for a helper: it prints one line per helper, `feov-ocr: available` or `feov-ocr: <error name>, <detail>`,
  and the run proceeds. OCR and search refuse per call with the §1.3 errors.

#### 1.5 Build, run and release jobs (`.github/workflows/hooks.yml`; F6, F7)

- **`scripts/releaseassets`** holds the facts the jobs share:
  - the target set, read from the release job's single `for pair in …;` line (it fails unless there is exactly one in
    that job);
  - per target, the runner (`ubuntu-24.04`, `ubuntu-24.04-arm`, `macos-15`, `macos-15-intel`, `windows-2025`,
    `windows-11-arm`); a target with no runner is an error;
  - per helper and target, the builder: zig on `ubuntu-24.04` for every `feov-ocr` target and the Linux and macOS
    `feov-embed` targets (one Linux host builds all six today, `docs/c-stack.md`), MSVC on the target's own Windows
    runner for Windows `feov-embed`.
  - `releaseassets -matrix` prints the `native-run` matrix, every target, as JSON from the helpers present (§1.1), so a
    new helper joins every job by existing.
  - `releaseassets -scope` decides, for one event, which `native-build` legs run and whether `native-run` runs. Its
    inputs are the event name, the ref and, on a pull request, the merge commit's changed files. It prints the needed
    legs as the `native-build` matrix, plus `full`, and refuses an empty matrix.
    - The **test targets** are derived from `hooks.yml`: the target of every `native-<helper>-<target>` artifact that a
      job other than `native-run` and `release` downloads. Today that is `linux-amd64` (`ocr-corpus` and `feov-record`'s
      Linux leg) and `windows-amd64` (its Windows leg). A new consumer joins by being written.
    - The **native set** of paths is `nativedigest`'s own input list, `devcmd/nativedigest/**`, `internal/nativehelper/**`,
      the directory of every package that imports `internal/nativehelper` (from `go list`, so `internal/tessocr` and Part
      3's embed client join by importing it), `scripts/releaseassets/**`, `scripts/internal/staticproof/**` and `hooks.yml`.

    | Event (every one in `hooks.yml`'s `on:`) | `native-build` legs | Helper cache | `native-run` / `native-fuzz` |
    |---|---|---|---|
    | pull request touching the native set | every helper × target | restore main's | run / run |
    | any other pull request | the test targets | restore main's | skip / skip |
    | push to main | the test targets | restore main's (F13) | skip / skip |
    | push of a FEOV tag | every helper × target | none (F2) | run / skip |
    | push of another plugin's tag | the test targets | none | skip / skip |
    | `schedule`, `workflow_dispatch` on main | every helper × target | restore; save on a miss | run / run |

- **`native-scope`** (`ubuntu-24.04`) has no `if:` and runs on every event, following the repository's "THE JOB ALWAYS
  RUNS AND DECIDES FOR ITSELF" (`hooks.yml:740-745`). It checks out with `fetch-depth: 2`, refuses a pull request whose
  HEAD is not the merge commit (as `release-fuzz-build`'s scope step does, `:771-790`), runs `releaseassets -scope`,
  prints the decision, and outputs the matrix and `full`.
- **`native-build`** (needs `native-scope`, matrix from its output, no `if:`) builds or restores one helper for one target
  with `native/build.sh` and uploads it as `native-<helper>-<target>`. A `feov-embed` leg also downloads its arch's model
  export from the `PINS.txt` URL, checks its sha256, and adds it to the artifact under its installed name.
  - **The cache follows `.github/actions/go-cache`'s rule** (`action.yml:8-13`) with one departure (F13). It is restored on
    every event but a tag push, which touches no cache (F2); a push to main restores main's entries. Only the nightly or a
    dispatch on `refs/heads/main` saves it. That is how a main-seeded entry reaches pull requests: a pull request
    restores entries saved on its base branch and saves none, because an entry saved on its merge ref is readable by no
    other pull request. A pull request whose digest main has not built yet rebuilds on the miss. That happens on a
    native change, or before the next nightly after one merges, and a dispatch on main seeds the entry without waiting.
  - The key is `native-<helper>-<target>-<builder runner>-<$ImageVersion>-<digest>`, the digest from
    `nativedigest -print`. A change to `shim.cpp`, `frame.cpp`, `build.sh` or a pin misses, and so does a builder image
    update.
  - A tag push rebuilds the helpers it needs from source. A main push rebuilds only on a cache miss.
  - Windows builds use `/Brepro` and no timestamps.
  - It runs `staticproof` on its output before uploading.
- **No producer skip turns a consumer into a skipped check.** `native-scope` and `native-build` carry no job-level `if:`,
  so a failure is a red check, never a skip. The test targets are in the matrix on every event, so every consumer's
  artifact exists. A missing artifact fails `download-artifact`, and a test run without a helper fails `Require` (below).
- **`native-run`** (matrix from `-matrix`, one leg per target on that target's own runner; needs `native-scope` and
  `native-build`) runs on every event. Each leg's first step reads `full`; when it is false the rest skip and the leg
  says so, as the `release-fuzz` parts do (`hooks.yml:866-894`). Otherwise it downloads that target's artifacts and
  runs FEOV's `go test` with `CGO_ENABLED=0`, as every published Go asset is built, for the packages that call
  `nativehelper.Require`, with `NATIVE_HELPER_DIR` pointing at them. It fails on any of:
  - the handshake digest or model sha256 differing from the compiled constants;
  - `selftest-stdout`: the frame not byte-identical, or the stdout writes absent from the request's captured stderr;
  - helper exit on stdin EOF taking more than 2 s;
  - `selftest-traineddata` differing from `PINS.txt`; `selftest-libc` failing (Linux targets);
  - the 13 corpus pages: on linux-amd64, any golden byte after line 1 differing (S3); on the other five, the same
    comparison, and a difference fails the job (R9);
  - any op's largest client request (§1.2, Deadline) taking over 30 s, a quarter of the deadline;
  - `feov-embed` (Part 3): fidelity on the 300-text set below S5; speed below S6 on the three targets S6 names, and
    speed reported for the other three.
  - windows-amd64 runs the same Windows binaries a second time inside a `mcr.microsoft.com/windows/servercore:ltsc2025`
    container, which has neither Go nor Visual Studio. On the host the leg builds each `Require` package's test binary
    (`go test -c`, `CGO_ENABLED=0`) and `./cmd/...` into the directory `FEOV_TESTBUILD_PREBUILT` names
    (`internal/testbuild/testbuild.go:280`), so `testbuild` builds nothing inside. The container mounts those, the
    helpers and the checkout, and runs each test binary from its package directory, as the `release-fuzz` parts run
    theirs (`hooks.yml:829-832`). windows-arm64 has no clean-image mechanism the agent knows of; Part 3 checks for an
    arm64 servercore image first, and states the residue if there is none.
- **`native-fuzz`** (`ubuntu-24.04`, needs `native-scope`, no `if:`) runs its steps when the table says so. It installs
  `libclang-rt-18-dev`, which the image's `clang-18 -fsanitize=fuzzer` needs for libFuzzer, then builds
  `native/frame/fuzz_frame.cpp` with `frame.cpp` under `-fsanitize=fuzzer,address,undefined` and runs it for 10
  minutes. zig ships neither libFuzzer nor the sanitizer runtimes, and this build publishes nothing, so the runner's
  compiler is acceptable. It covers linux-amd64; the other targets compile the same `frame.cpp`.
- **The `release` job** (`hooks.yml:1672-1872`):
  - `needs: [test, feov-record, gray-area, native-build, native-run]` and
    `if: ${{ !cancelled() && !failure() && startsWith(github.ref, 'refs/tags/') && contains(github.ref, '--v') }}`
    (`hooks.yml:1682-1683`). Every tag waits for the test targets' helper builds, which the engine tests need (F11). A
    FEOV tag also waits for every `native-run` leg. A failure in any needed job stops the release.
  - The Go loop (`hooks.yml:1790-1822`) loses its capability-keyed C-stack branch (`:1799-1817`). Every command of every
    plugin takes one line, `CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -trimpath -ldflags "-s -w" -o
    "${DIST}/${name}_${os}_${arch}${ext}" "./cmd/${name}"`. The literal `-o` template and the single `for pair in`
    (`:1792`) stay in this job, so `sc-doctor`'s contract test still reads them.
  - For a FEOV tag it downloads every `native-build` artifact into `$DIST`, renaming each model export to its release
    asset name from `ASSETS` and keeping one copy per name. The legs that fetched the exports checked them against
    `PINS.txt` (F10).
  - It runs `releaseassets -check` on `$DIST`, then `sha256sum * > SHA256SUMS` (`:1825`), then publishes. A dynamic
    linux/amd64 Go binary, a missing helper or a helper beyond its allowlist fails here, before any byte is published.
- **`sc-doctor`'s release contract test** (`doctor/release_contract_test.go:64-74, 82, 110-114`) matches `for pair in` in
  the whole file and `-o` after `\n  release:\n`. It is narrowed to search for both inside the release job's span only,
  so no other job's line can stand in for the release's (the native jobs use matrices, not loops, but nothing should
  depend on that). It also asserts that the release job runs `releaseassets -check` before `sha256sum`.
- **CI jobs that run FEOV tests over a `Require` package** (sweep: `grep -rln "nativehelper.Require" $F` for the
  packages, then every `go test` over them in `.github/workflows/`, crossed with every event in `on:`):
  - `ocr-corpus` (`hooks.yml:275-328`) and the `feov-record` Linux and Windows legs (`:387-`) run on every event. They
    need `native-build`, keep no job-level `if:`, download `native-*-<their target>` into one directory and set
    `NATIVE_HELPER_DIR` to it. Plain `go test` links in seconds, so `ocr-corpus`'s C-stack steps (`:301-313`), the
    `feov-record` job's C-stack steps and comment (`:532-561`), its separate OCR step (`:570-616`), its tag-derived test list (`:588-590`) and its `PASS|SKIP` grep
    (`:612-616`) are deleted: the engine tests run in the default leg (`:657`).
  - `native-run` sets it per target.
  - `race-nightly.yml:80` races the shipped graph with `NATIVE_HELPER_DIR=none`; the C++ is not raced (§1.7).
  - **In CI, `Require` never skips for want of a helper.** With `GITHUB_ACTIONS` set (by the runner, so it cannot be
    mistyped) and `NATIVE_HELPER_DIR` unset, `Require` fails the test. `none` is the only way to run without a helper,
    and it skips loudly. A typo in either the variable or its value therefore fails the job. Outside CI an unset
    variable skips loudly, as `go -C scripts run ./check` does.
- **`staticproof`** moves to `scripts/internal/staticproof` with a `scripts/staticproof` command, because every
  plugin's release now runs it. It takes the allowlist per artifact: Go assets and `feov-ocr` keep today's lists;
  `feov-embed`'s Windows list adds `advapi32.dll` (event tracing), `dbghelp.dll` (stack traces) and
  `api-ms-win-core-path-l1-1-0.dll` (path handling) to the shared base, each with its reason in the list (F3). Its
  `allowlist_test.go` moves with it.

#### 1.6 Enumerators (census G)

- **`fetch-bin.sh` and `fetchbingen`** change as in §1.4. Their contract tests change with them:
  `doctor/fetchbin_contract_test.go` (`TestFetchBinInstallsEveryMissingBinaryFromThePinnedRelease` `:255`,
  `TestFetchBinInstallsNothingWhenOneDigestIsWrong` `:271`, `TestFetchBinNamesAnAssetTheReleaseDoesNotCarry` `:284`)
  gain a helper and an `ASSETS` line, and the `fetch-bin-platforms` job's fixture plugin (`hooks.yml:1159-1216`, `:1182`)
  gains `native/fake-helper/ASSETS`, its assets, and a `.fetch/source-built` root.
- **`sc-doctor`** enumerates the helpers and their `ASSETS` the same way (`binariesOf`, `main.go:73-91`), reports each as
  installed or missing, and fetches them; it never builds one (`main.go:533-538` stays Go-only). It holds neither FEOV's
  digest nor a frame client, so it checks no digest. `feov-record`'s handshake and `setup` preflight do.
- **`pluginparity`** needs no helper check: `releaseassets -scope` and `-matrix` generate the build and run matrices from
  the helpers present, and `releaseassets -check` refuses a release missing one.
- **`bootstrap-plugins.sh`** (`:85-97`) builds Go commands from the cache and no helpers, and prints one line saying the
  next session start fetches them.
- **`universe.sh`** follows F5. `build` writes `.fetch/source-built`; `build --native` also runs `native/build.sh` for the
  host target and copies both helpers and the host arch's model export into `bin/` of every root it populates. Its
  post-build assertion (`universe.sh:498-`) is extended to run `fetch-bin.sh fetch` on each staged root and assert
  `bin/` is unchanged.
- **Unchanged:** `marketplacegen`, `versionguard` and `gcpolicy`, because helpers are not Go `cmd/` directories.

#### 1.7 Gates and docs

- **`scripts/check/gates.go`:**
  - The race scope's stated limit (`:167-174`) is rewritten. The race detector covers the Go client; the C++ is not
    raced, and the text says so.
  - The `ocr-corpus` skip text (`:323-325`) names `feov-ocr`, `NATIVE_HELPER_DIR` and `native/build.sh`, says "every
    golden" rather than a count (it says "all eight" over 13 today), and states the re-arm set: `internal/tessocr/**`,
    `native/feov-ocr/**`, `native/frame/**`, `native/build.sh` and `PINS.txt`, the inputs of `Identity()`.
- **Docs:**
  - `docs/c-stack.md` is rewritten around the helpers: zig for Linux, macOS and Windows `feov-ocr`; MSVC for Windows
    `feov-embed`. The MSVC runner is the decision the current doc says a native-only toolchain would need.
  - `docs/local-development.md:302-305` is rewritten, and its universe section documents `build --native` and
    `.fetch/source-built`.
  - `docs/setup-script.md:73` says "nothing needs provisioning", which becomes false.
  - `AGENTS.md`, "AFTER `universe.sh build`, YOU MUST prove the universe carries your change" (`:91`), gains: for a
    native change, build with `--native` and `cmp` each staged helper against `native/build.sh`'s output (R5).

### Part 2: `feov-ocr` (owner: the OCR session)

**The move, file by file (census B, C):**
- `shim.cpp` and `shim.h` move to `native/feov-ocr/`, with a main loop over the frame codec.
- `eng.traineddata` (4 MB) is embedded in the helper. Its pin stays in `PINS.txt`; `TestEmbeddedTraineddataPin` becomes
  a Go test in `internal/tessocr` that calls `selftest-traineddata` through `Require`, so it reads the bytes the shipped
  helper carries, on every target in `native-run`.
- `engine_cgo.go` and `engine_stub.go` are deleted. A pure-Go client implements the same exported functions over
  `nativehelper`, so `page.go`, `reconstruct.go`, `lattice.go` and the other pure-Go files do not change.
- `internal/libcprobe` is deleted; its probes become `selftest-libc`.

**Tests (census D):**
- The 14 tagged test files re-gate on `nativehelper.Require`. That includes `blankreceipt_cgo_test.go` and
  `cli/ocrcite_real_test.go`.
- `diagcrash_cgo_test.go` is rewritten for the new crash shape, where a crash is a helper exit.
- `reconstruct_test.go`'s `TestIdentityMatchesPins` (`:157-174`) and the CRLF fold (`:231`) give way to
  `TestNativeDigestMatchesSources`.
- `testbuild/tags_*.go` and `engineBuildFlags` are deleted; `testbuild.Binary` (`:353`) builds plain and passes
  `NATIVE_HELPER_DIR` through.
- The untagged stub tests (`tessocr/engine_stub_test.go`, `tessocr/stubbuild_test.go`, `fetchcache/enginestub_test.go:25-30`,
  `cli/enginestub_test.go:14-34`) become "no helper refuses loudly" tests, asserting `helper-absent` and its sentence.
  `cli/fetchocr_test.go:121,126` pins the stub's error text in an assertion and is rewritten with it.

**Agent-facing text and comments (census H, re-run at `68a735d51`, 191 lines in
`~/.claude/scratch/vecplan/census/H-r3-68a735d51.txt`, each read for its meaning):**
- `cli/fetch.go:62` says "compiled into this binary". The fix is made there and regenerated with `agentgen`, which
  updates 10 agent definitions; the difftest manual golden renders it three times (`:960, :1538, :2100`).
- `cli/ocr.go:244` help (golden `:2142`), `cli/ocr.go:66-70, :226`, `cli/fetch.go:41, :177`, `pageread.go:17, :36, :40,
  :56, :85-95`, `seatprobe/build.go:368`, `seatprobe/scanned.go:29`, `seatprobe/separatedbuild_test.go:93-94`,
  `cli/ocr_test.go:179`, `tessocr/tessocr.go:1-17, :32-46, :92, :107-111`, `tessocr/page.go:7, :190`,
  `goldens_cgo_test.go:32`, `fetchcache/corpus_cgo_test.go:44`, `testdata/corpus/README.md:31, :49-50`,
  `testdata/gen/main.go:13`, `testdata/gen/pages.go:15`, `testbuild/testbuild.go:352`, `PINS.txt:1-14`, and
  `hooks.yml:285, :532-549`, `race-nightly.yml:27` are rewritten to name the helper. Every changed golden line is read.
- Same words, other sense, unchanged: `record/cast.go:43`, `report/site.go:119`, `seatprobe/surfacecoverage_test.go:157`,
  `setup/run_test.go:115`, `cli/cli_test.go:134`, `difftest/golden_test.go:88`, `integration/surface/hookgraph_test.go:74`.
- Still true after the move, unchanged: the pure-Go tests and files that say they run "without the C stack"
  (`lattice_test.go:9`, `levels_test.go:9`, `page.go:449, :509`, `page_test.go:14`, `reconstruct.go:7`, `spread_test.go:13`,
  `fetchcache/pageread_test.go:17`, whose fake engine still needs no helper), and `staticproof`'s "not statically linked" refusals, which move with it.

**Goldens:** the 13 corpus goldens are regenerated, and S3 is checked by diffing everything after line 1.

### Part 3: `feov-embed` (owner: the vector-search session; starts once Part 1 lands)

**The helper:** `native/feov-embed/` holds the ONNX Runtime reduced operator set (`ops.config` from the scratch proof)
and the Release minimal build: 6,057,720 bytes added on linux-amd64, at full speed. One op,
`embed(ids, mask) -> float32[n][384]`, runs with one intra-op thread. The Go client splits any set of texts into
requests of at most `embedBatch` texts of at most 512 ids each (§1.2, Deadline). It loads `bin/gte-small.int8.ort`, the export for
its arch (L2), and refuses at handshake when the file's sha256 is not that arch's pin.

**The model:** the two exports S5 and S6 were measured with are pinned by sha256 and URL in `PINS.txt`; where they are
hosted is F10. `native-run` measures S5 on each target with the export that target publishes.

**The plan:** `plans/vector-search.md` is revised in its own PR on this one. Its llama.cpp and `internal/vecembed`
sections (`:8, :61, :69-120, :329-332, :458-506`) give way to ONNX Runtime 8-bit and the helper.
- Its search, schema, chunking and RRF design stand, along with the rulings on `--quote` and the hybrid.
- Its "approved" status is struck, because the agent that wrote it set that status itself, and it goes back through the
  gate. The same PR rewrites its `plans/README.md` row (`:22`), which names llama.cpp and the GGUF model.

### Consumer census

Groups A–F and I–K are the run against `origin/main` at `7b973ec4e`, 2026-11-04. Groups G and H, and every line number
cited in §I–§V, are re-derived against `origin/main` at `68a735d51`. Each group's command re-enumerates it;
Appendix A has every line.

| Group | What | Count | Enumerate |
|---|---|---|---|
| A | The Go↔C boundary | 7 C ops, about 12 call sites per page | `grep -n "C\.tessocr_\|^func " $F/internal/tessocr/engine_cgo.go` |
| B | The `internal/tessocr` split | 30 files | `ls $F/internal/tessocr; grep -ln "go:build" $F/internal/tessocr/*.go` |
| C | Go importers and callers of the engine | 4 importing files; 2 more carry its identity or text | `grep -rln 'internal/tessocr"' $F --include=*.go \| grep -v _test` |
| D | Build-tag test infrastructure | 14 tagged + 3 harness sites | `grep -rln '^//go:build.*tessocr' plugins --include=*.go` |
| E | C-stack tooling | 4 | `ls $F/third_party/pins` |
| F | `hooks.yml` | 3 jobs build the stack, 1 publishes, 2 contract tests parse the workflow; 4 jobs run tests over a `Require` package | `grep -n "tessocr\|cstack\|linkmode\|for pair in\|go test" .github/workflows/*.yml` |
| G | Enumerators that treat `tools/cmd` as the binary list | 9 | `grep -rn 'tools/cmd' scripts plugins/*/hooks plugins/prosthetic-conscience/tools/internal/doctor .github \| grep -v _test` |
| H | Text naming the in-process engine | 191 lines: 1 seat-facing source, 10 generated agents, 4 golden lines, the rest help, comments and tests | `grep -rn "statically linked\|compiled into\|C stack\|tags tessocr\|build-cstack\|compiled in\b\|ErrNotCompiledIn\|EngineCompiledIn\|IN THE BINARY\|linkmode external" plugins/frank-exchange-of-views docs AGENTS.md scripts .github` |
| I | Gates | 3 | `grep -n "tessocr\|C stack" scripts/check/*.go` |
| J | Engine identity | 1 producer, 4 stores, 3 comparers | `grep -rn "Identity()\|ocr_engine" $F/internal --include=*.go \| grep -v _test` |
| K | Docs and plans | 3 docs, 1 live plan, 1 rule in AGENTS.md | `grep -rln "tessocr\|build-cstack\|linkmode external" docs plans` |

`F=plugins/frank-exchange-of-views/tools`.

## IV. Risk & Mitigation

| # | Risk | L × I | Mitigation |
|---|---|---|---|
| R1 | The OCR readings change when the code moves (decode, normalization or threading differences) | M × H | PNG bytes still cross the boundary; S3's line-1-only diff on all 13 goldens decides it (Part 2) |
| R2 | The MSVC or zig helper build is not reproducible, or needs something absent from a clean Windows | M × M | `/Brepro`, the pinned image and toolset (§II.4); `native-build` re-run on one commit compares every artifact's sha256 (§V.6); the servercore leg of `native-run` |
| R3 | A protocol defect, or a library writing to stdout, returns results that look like a clean reading or no hits | L × H | The private frame descriptor and binary mode (§1.2); the distinct errors of §1.3; the fuzzed decoder; `selftest-stdout` on all six targets |
| R4 | The eager fetch makes a first hook start slower (the model alone is 33.8 MB) | H × L | Accepted under F1; `fetch-bin.sh`'s existing all-or-none install and stamp, unchanged in shape |
| R5 | A universe tests the published helper instead of the checkout's | M × M | F5; `.fetch/source-built`; the AGENTS.md rule extended to `cmp` the staged helpers |
| R6 | The `sc-doctor` contract test reads a line from another job | M × L | The test's search narrowed to the release job (§1.5) |
| R7 | Release time grows by about 25 minutes for a FEOV tag, and every tag (any plugin's) waits on an uncached build of the test targets' helpers, about 30 minutes once Windows `feov-embed` exists | H × L | Accepted under F2 and F11. Main pushes and pull requests restore main's entries (F13) |
| R8 | Part 2 and Part 3 drift apart in protocol use | M × M | One `nativehelper` package and one `frame.h`, both landed in Part 1 before either helper |
| R9 | `feov-ocr` reads a corpus page differently on another target (SIMD paths in tesseract's LSTM) | M × M | `native-run` compares every target with the goldens. A target that differs blocks the Part 2 PR; per-target goldens, or a weaker bar for that target, go to gblock with the diff. Today such a difference ships unseen |
| R10 | The 120 s deadline fires on a legitimate request on a slow target | L × M | The client bounds every request's size, and `native-run` sends each op's largest request on every target with a 4× margin (§III.1.2, §V.4). A page far larger than the corpus's largest can still overrun, loudly and for that page only |
| R11 | Part 1 lands before #1724's guard is narrowed, so consumers never fetch a helper | M × H | Stated precondition (§II.4); Part 1's `fetch-bin.sh` PR does not merge until a test fixture with `tools/go.mod` fetches |
| R12 | The one-time model release is deleted or re-tagged | L × H | Its tag matches no workflow trigger; each export's sha256 is in `PINS.txt`, so a changed byte fails the release job's download loudly |

## V. Verification Plan (the checklist)

Each item names what re-arms it. Tests that need a helper run through `nativehelper.Require` with `NATIVE_HELPER_DIR`,
the lookup the shipped binary uses. Tests that need a misbehaving helper, or one with another digest, build the Go fake
helper (§1.1). Gap numbers are round 1's unless marked r2.

1. **`go -C scripts run ./check`** after `unset $(env | grep -o '^FEOV_[A-Z_]*')`: every local gate at `-count=1`. Read
   the fail count, not the exit code. Re-armed by any change.
2. **Static release assets (gap 1, S1, S2).** Re-armed by `scripts/releaseassets/**`, `scripts/internal/staticproof/**`,
   the release job.
   - `TestADynamicELFIsRefused` in `scripts/releaseassets`: builds a program importing `net` with `CGO_ENABLED=1` and
     the host C compiler (failing, not skipping, under `GITHUB_ACTIONS`), puts it in a fixture `$DIST`, and asserts
     `-check` fails naming it. This is the defect `audit/round1/feov-record-plain` reproduced.
   - `TestEveryExecutableIsProven`: a fixture `$DIST` of three Go assets and one helper; delete the `staticproof` call
     for one kind and the test fails.
   - `TestAFEOVDistMissingOneHelperIsRefused` and `TestADistWithoutHelpersNeedsNone` (gray-area's shape), so the check
     neither passes a FEOV release missing `feov-ocr_windows_arm64.exe` nor demands helpers of another plugin.
   - `TestTheEmbedAllowlistIsItsOwn`: a fixture PE importing `advapi32.dll` passes as `feov-embed` and fails as a Go
     asset or `feov-ocr`.
   - On the Part 1 PR: every plugin's Go assets are built for all six targets with the release job's own build line into
     a scratch `$DIST`, FEOV's beside the PR's `native-build` artifacts, and `releaseassets -check` runs over each. Read
     one "static" per asset. This is S1 on the PR, not only a compile.
   - `sc-doctor`'s contract test asserts `releaseassets -check` precedes `sha256sum` in the release job.
   - The first release of each plugin after Part 1: read the `releaseassets -check` lines, one "static" per asset.
3. **Frame codec and lifecycle (gaps 6, 9, the restart NOTE).** Re-armed by `native/frame/**`, `internal/nativehelper/**`.
   - `go test -fuzz=FuzzFrame -fuzztime=60s ./internal/nativehelper/`.
   - `native-fuzz`: libFuzzer on `fuzz_frame.cpp` under ASan and UBSan, 10 minutes, zero findings. Its first run proves
     `libclang-rt-18-dev` is installed.
   - `TestFrameChannelSurvivesLibraryStdout`: `selftest-stdout` returns all 256 byte values byte-identical, and the
     stdout writes appear in that request's captured stderr field. Run on all six targets in `native-run`. It fails on
     Linux without the fd move or without the paired fd 1 capture. It fails on Windows without binary mode, or with
     frames written to a `GetStdHandle` value read before `_dup2`.
   - `TestACrashedRequestIsNotRetried`: a request that aborts the helper returns `helper-crashed` with its stderr, is
     not resent, and the next request spawns once.
   - `TestAPageCrashIsNotAnAbsentEngine`: through `fetch`, a helper that handshakes and then aborts on one page reports
     that page `helper-crashed` with `ocr_engine_absent` false. A missing helper reports it true. A rule that set the
     field on any helper error fails the first case.
   - `TestAFailedHandshakeLatches`: a helper with another protocol version is spawned once per process, however many
     requests follow.
   - `TestHelperExitsOnStdinEOF`, all six targets.
4. **Deadline and request size (gap 10; r2 gap 2).** Re-armed by the deadline constant, `embedBatch`, `MaxRenderDPI`,
   `internal/nativehelper/**`, `native/feov-ocr/**`, `native/feov-embed/**`, the corpus.
   - `TestADeadlineKillsAndNames`: `selftest-sleep` past a test deadline returns `helper-deadline` and the next request
     succeeds.
   - `TestEveryOpFitsTheDeadline` in `native-run`, every target, each request under 30 s. It times spawn to handshake
     reply, then every op in the Go client's op table at its largest client request:
     - each `feov-ocr` op gets the largest corpus page at 600 DPI whole, with `rot_band` given the whole-page rectangle;
     - `embed` gets `embedBatch` texts of 512 ids.
     It walks the op table and fails on an op with no largest-request case, so a new op cannot pass by being absent. The
     self-test ops are excluded by name. The darwin-amd64 leg, the slowest measured target, is where `embedBatch` meets
     the bound.
   - `TestEmbedSplitsAtTheBatch` (Part 3), with a fake helper recording each request: 1,000 texts go out as
     ⌈1000/`embedBatch`⌉ requests, none over `embedBatch` texts or 512 ids. Deletion check: remove the split and it fails.
5. **Identity and lookup (gap 2, S8).** Re-armed by `native/**`, `PINS.txt`, `digest_gen.go`, `internal/nativehelper/**`.
   - `TestNativeDigestMatchesSources`: deletion check, change one comment in `native/frame/frame.cpp` and it fails. It
     also recomputes the per-arch model pins from `PINS.txt`.
   - The CRLF fold, on a real CRLF checkout. `.gitattributes` pins no `.cpp`, `.h` or `.txt`, so a Windows checkout of
     `native/` is CRLF. `TestNativeDigestMatchesSources` runs in `feov-record`'s Windows leg, and `native/build.sh`'s
     `nativedigest -check` runs on the MSVC builder. Without the fold, both fail. `TestNativeDigestFoldsCRLF` is the
     unit case. The Windows `native-run` legs compare the handshake with the committed constant and never exercise the
     fold, so they are not this check.
   - `TestAHelperFromOtherSourcesIsRefused`: a `testbuild.Binary` `feov-record` with `NATIVE_HELPER_DIR` at the fake
     helper built with another digest reports `helper-digest-mismatch`, both digests named.
   - `TestTheOverrideHasNoFallback`: `NATIVE_HELPER_DIR` at an empty directory while `bin/` beside the binary holds a
     matching fake helper gives `helper-absent` naming the override.
   - `TestIdentityNeverSpawns`: `Identity()` with `NATIVE_HELPER_DIR` at a fake helper that records each start returns
     the compiled string and starts nothing; with `none` it returns the same string.
   - `TestReleaseLayoutFindsItsHelper`: a built `feov-record` in a temp `bin/` beside a matching helper, override unset,
     reads a page.
6. **Release graph and reproducibility (gaps 5, 7, R2; F12).** Re-armed by `hooks.yml`, `native/**`, `PINS.txt`.
   - actionlint through CI's `qlty` (a local `qlty` does not run it).
   - On the Part 1 PR: run `native-build` once restoring from cache and once uncached on the same commit, and compare
     every artifact's sha256 (zig and MSVC both). A cached artifact that an image update has made stale shows here, as
     well as at a tag's uncached build.
   - On the Part 1 PR: a commit touching only `native/frame/frame.cpp` shows a `native-build` cache miss in its log.
   - The first FEOV release after Part 2: its `release` job started after every `native-run` leg passed; a re-run of the
     same tag reports "SHA256SUMS match".
   - The next gray-area or prosthetic-conscience release after Part 1: `native-scope` chose only the test targets,
     `native-run`'s legs skipped and said so, and the release published.
   - Residue (F12): the release job's new steps first execute on a real FEOV tag. Those steps are the artifact download,
     the export renaming, and `releaseassets -check` over a full FEOV `$DIST`. A failure there is loud and is fixed when
     it happens.
7. **Universe and provisioning (gap 3, S9).** Re-armed by `scripts/fetchbingen/fetch-bin.sh`, `doctor/**`,
   `scripts/universe.sh`.
   - `TestFetchBinNeverFetchesIntoASourceBuiltRoot` in `doctor/fetchbin_contract_test.go`: a root with no
     `tools/go.mod`, a helper missing and a `file://` release holding it. First, without `.fetch/source-built`, `fetch`
     installs the helper. That proves nothing else stops it, #1724's guard included. Then, with the marker, `hook`,
     `ensure` and `fetch` each leave `bin/` and `.fetch/installed` unchanged. Deletion check: remove the marker test
     from `fetch-bin.sh` and the second half fails.
   - `TestFixDoesNotFetchIntoASourceBuiltRoot`, through `fixWith`.
   - `TestStampWaitsForEveryHelper`: `stampInstalled` writes no stamp while one helper or `ASSETS` file is missing.
   - `TestFetchBinPicksTheExportForItsArch`: an `ASSETS` line with `amd64=` and `arm64=` fields installs the right one;
     a line without the host arch fails naming it.
   - `TestEnsureIsSilentWhenTheInstalledTagMatches` (`fetchbin_contract_test.go:537`) keeps its fixture plugin
     helper-less, so the unmatched `native/*/ASSETS` glob is exercised. Deletion check: drop the `[ -f ]` test and it
     fails.
   - `fetch-bin-platforms` (Linux, macOS, Windows) runs the fixture plugin with a fake helper, an `ASSETS` line and a
     source-built root.
   - `scripts/universe.sh build`, then its post-build assertion runs `fetch-bin.sh fetch` on every staged root and `bin/`
     is unchanged.
   - `scripts/universe.sh build --native`, then `cmp` each staged helper against `native/build.sh`'s output, and
     `feov-record setup` in the universe prints `feov-ocr: available`.
8. **Test gating in CI, and a helper on every event (gap 4; r2 gap 1).** Re-armed by
   `internal/nativehelper/require.go`, `scripts/releaseassets/**`, `.github/workflows/*.yml`.
   - `TestRequireFailsInCIWithoutAHelper`: re-runs the test binary with `GITHUB_ACTIONS=true` and `NATIVE_HELPER_DIR`
     unset and asserts a failure, not a skip; with `none`, a skip; with a mistyped value, a failure.
   - `TestScopeBuildsEveryConsumersTarget` in `scripts/releaseassets`: drives `-scope` over the repository's own
     `hooks.yml` for every event its `on:` lists, read from the file. Pushes are split into main, a FEOV tag and another
     plugin's tag, and pull requests into those with and without a native-set change. For each, every target that a job
     needing `native-build` downloads must be in the matrix, and the matrix must not be empty. Deletion checks: replace
     the test-target derivation with the fixed list `linux-amd64`, `windows-amd64`, then run a fixture copy whose
     `feov-record` Windows leg downloads `windows-arm64`, and the no-native-change pull request case fails. Add an event
     to the fixture's `on:` that the table lacks, and it fails.
   - `TestNoNativeJobSkipsAtTheJobLevel`: in `hooks.yml`, `native-scope`, `native-build`, `native-run`, `native-fuzz`
     and every job that needs `native-build` other than `release` have no job-level `if:`. Deletion check: a fixture
     with `if: github.event_name == 'pull_request'` on `native-build` fails it.
   - The shipped path, read on the first run of each event after Part 1 merges: a pull request, a push to main, the
     nightly, and a dispatch on main. For each, read `native-scope`'s decision line and `native-build`'s legs, and the
     engine tests' `--- PASS` lines in `ocr-corpus` and both `feov-record` legs. On a pull request after a nightly,
     `native-build` reports a cache hit on main's entry.
   - The census of F (jobs running `go test` over a `Require` package) is re-run on the Part 1 PR's final head; each job
     sets `NATIVE_HELPER_DIR`, and the PR's first CI run is the proof, since a job that does not is red.
9. **`setup` preflight (gap 11, F9).** Re-armed by `internal/setup/**`, `internal/nativehelper/**`.
   - `TestSetupProceedsAndNamesAMissingHelper`: `testbuild.Binary` `feov-record` in a temp `bin/` with no helpers; `setup`
     creates the run and prints `feov-ocr: helper-absent` and `feov-embed: helper-absent`.
   - `TestSetupNamesAMismatchedHelper`: the same with a fake helper of another digest prints `helper-digest-mismatch`.
   - `TestSetupLooksBesideTheSeatsBinary`: `setup --bin-dir` names a directory holding a matching fake helper, and
     setup's own directory holds none, so it prints `available`. Deletion check: look beside `os.Executable()` and it
     prints `helper-absent`.
10. **Real-data check, OCR (S3, S4, R9).** Re-armed by `native/feov-ocr/**`, `internal/tessocr/**`, the corpus.
    - The 13 corpus scans are read through `feov-ocr` in `ocr-corpus` and in `native-run` on all six targets.
    - `diff <(tail -n +2 old) <(tail -n +2 new)` is empty for every golden on linux-amd64; the other targets per R9.
    - `TestEmbeddedTraineddataPin` through `selftest-traineddata` on all six; `selftest-libc` on both Linux targets.
    - The interleaved in-process and helper timing table is posted to the PR.
11. **Real-data check, embedding (S5, S6, gap 8).** Re-armed by `native/feov-embed/**`, the model pins.
    - `native-run` per target: the handshake's model sha256 equals the pin for that arch
      (`TestEmbedLoadsTheExportForItsArch`); fidelity against `ref_gte.npy` on the 300-text set; one-thread speed.
    - The replay set and BEIR hybrid scores reproduce `ortq/RESULT.md` through the helper.
    - The windows-amd64 servercore leg runs the fidelity set with no Visual Studio and no Go present. Its first run
      proves the host-built test binaries and `FEOV_TESTBUILD_PREBUILT` carry everything the container runs.
12. **Census H (gap 12).** Re-armed by any text naming the engine. After Part 2, the census H command returns only the
    "other sense" and "still true" lines listed in Part 2; any other hit fails the PR's review.
13. **Running the engine:** a smoke run in a `universe.sh build --native` universe that fetches one scanned PDF and runs
    one `--search`, with the run's `engine.log` read.
14. **The PR's `release-fuzz` parts and `release-fuzz-tally`**, read on the final head (FEOV's tools module changes).
15. **The auditor gate:** `/prosthetic-conscience:plan-audit` on this plan.

## Appendix A: Census detail

Copied from `~/.claude/scratch/vecplan/census/CENSUS.md`, the run against `7b973ec4e`. Groups G and H are re-run against
`68a735d51` (marked), and the line numbers of every group are re-checked there; the full H listing is
`~/.claude/scratch/vecplan/census/H-r3-68a735d51.txt`.

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
- `pageread.go:85-95` engineAbsentLoudly wraps ErrNotCompiledIn with "build with -tags tessocr" — changes: the §III.1.3 errors; sentence and errors.Is target change.

---------------------------------------------------------------------------------------------------
### B. internal/tessocr package split (cgo half leaves; pure half stays) — 30 files

Enumerate: `ls $F/internal/tessocr; grep -ln "go:build" $F/internal/tessocr/*.go`

- `$F/internal/tessocr/engine_cgo.go:1` `//go:build tessocr && cgo`, import "C", go:embed eng.traineddata :40 — DELETED; replaced by a pure-Go client speaking the frame protocol (same exported funcs: New/PageTSV/RotatedBand/DetectGrid/InkFraction/GridLines/RepairedRules/captureDiagnostics).
- `$F/internal/tessocr/engine_stub.go:1` `!tessocr || !cgo` stub returning ErrNotCompiledIn — DELETED (no tag); the "absent" case becomes `helper-absent`, loud.
- `$F/internal/tessocr/shim.cpp`, `shim.h` — MOVE into `native/feov-ocr/` (main loop + frame codec added).
- `$F/internal/tessocr/eng.traineddata` (4 MB, pinned in PINS.txt:28) — MOVES into the helper (embedded in the C++ binary); leaves the Go binary.
- `$F/internal/tessocr/tessocr.go:39-41` tesseractPin/leptonicaPin/engTraineddataPin; `:51` cstackPin typed digest of PINS.txt+build-cstack.sh; `:57` `//go:embed *.go shim.cpp shim.h` engineSource; `:100` Identity() — CHANGES: shim.* no longer in this dir; `cstack@` gives way to `feov-ocr@<source digest>` (§III.1.4).
- `$F/internal/tessocr/tessocr.go:107-111` ErrNotCompiledIn and its message "build with -tags tessocr" — replaced by the §III.1.3 errors.
- `$F/internal/tessocr/tessocr.go:1-17` package doc ("statically linked into this binary", tag split, invocation) — rewrite.
- `$F/internal/tessocr/page.go:1-14,190-205` doc "orchestrates Engine methods that are themselves tag-split" — reword; logic unchanged.
- `page.go`, `reconstruct.go`, `lattice.go`, `levels.go`, `normalize.go`, `repaired.go`, `spread.go`, `reading.go` — unchanged (pure Go, stay in feov-record; still hashed into identity).
- Tests, tagged (all need a real helper instead of the tag): `engine_cgo_test.go:20` TestEmbeddedTraineddataPin (becomes a Go test through `selftest-traineddata`), `goldens_cgo_test.go`, `textcells_cgo_test.go`, `diag_cgo_test.go`, `diagcrash_cgo_test.go:24,44` (crash/stderr capture via SetCrashOutput — semantics change entirely: crash = child exit) — re-gate on `Require`.
- Tests, untagged: `engine_stub_test.go:12` TestStubRefusesLoudly, `stubbuild_test.go:16` TestDefaultBuildCompiles — rewrite to "missing helper refuses loudly".
- `reconstruct_test.go:157-174` TestIdentityMatchesPins (recomputes cstackPin from PINS.txt+build-cstack.sh), `:193` TestEngineSourceHashTracksTheCode, `:231` and `reconstruct_crlf_test.go` (CRLF fold because shim.* not LF-pinned) — TestIdentityMatchesPins and the fold give way to TestNativeDigestMatchesSources and TestNativeDigestFoldsCRLF; TestEngineSourceHashTracksTheCode stays for the Go half.
- `testdata/gen/main.go:13`, `testdata/gen/pages.go:15` "go test -tags tessocr -run TestTextCellGoldens -update" — doc command changes.
- `testdata/corpus/README.md:31,49-50` tagged suite invocation (eval build-cstack env + -linkmode external) — rewrite.
- `testdata/corpus/*/reading.golden` line 1 (13 files) carries `tesseract@5.5.3+leptonica@1.87.0+eng@7d4322bd2a77+cstack@484e7942d3c4+tessocr@605c3c75d37a` — ALL regenerate on any identity change; body bytes should be UNCHANGED if PNG bytes still cross the boundary (the proof the move is behaviour-preserving).

---------------------------------------------------------------------------------------------------
### C. Go importers of tessocr / callers of the engine — 4 importing files; 2 more carry its identity or text

Enumerate: `grep -rln 'internal/tessocr"' $F --include=*.go | grep -v _test` (the 4 importers); the 2 others by reading the importers' callers.

- `$F/internal/fetchcache/pageread.go:13` — PageEngine + receipts; reuse check `:299` `r.Engine == DefaultPageEngine.Identity()` BEFORE reading; `:330,:503` stamps Identity — changes: Identity() must stay cheap and infallible (it is called on cache-hit paths that must not spawn/fetch the helper).
- `$F/internal/fetchcache/readscanned.go:14,108,201` — fetch's fused render-and-read; `:108` reuse check compares prev.Engine to Identity() before rendering — same constraint.
- `$F/internal/fetchcache/pagerender.go:21,49` — only `tessocr.RenderDPI` — unchanged.
- `$F/internal/cli/fetch.go:17,205` — `s.OCREngineAbsent = errors.Is(rerr, tessocr.ErrNotCompiledIn)` — changes: target errors are the §III.1.3 set; field `ocr_engine_absent` (fetchsummary.go:91) is seat-facing JSON, its meaning widens to "no working engine", `ocr_reason` names the case.
- Not importers: `$F/internal/fetchcache/pagelocate.go:144` — copies rec.Engine into loc.Engine for cite — unchanged (string passthrough).
- Not importers: `$F/internal/cli/ocr.go:66-72` doc comment "THE ENGINE IS IN THE BINARY NOW", `:226`, `:244` `ocr read` Long help "OCR engine compiled into this binary (tesseract + leptonica, statically linked)" — help text changes -> difftest manual golden `:2138`.
- `$F/internal/cli/fetch.go:62` fetch Long help "LOCAL OCR ENGINE compiled into this binary" — changes; this sentence is inlined by agentgen into 10 agent definitions (group H).

---------------------------------------------------------------------------------------------------
### D. Build-tag test infrastructure (tessocr tag) — 14 tagged files + 3 harness sites

Enumerate: `grep -rln '^//go:build.*tessocr' plugins --include=*.go; grep -rn 'EngineCompiledIn\|engineBuildFlags' $F --include=*.go`

- `$F/internal/testbuild/tags_tessocr.go:11` `engineBuildFlags = -tags tessocr -ldflags -linkmode external -extldflags "-static"`, EngineCompiledIn=true — DELETED; replaced by `nativehelper.Require` and `NATIVE_HELPER_DIR`; feov-record test builds become plain.
- `$F/internal/testbuild/tags_default.go` EngineCompiledIn=false — DELETED with it.
- `$F/internal/testbuild/testbuild.go:352-353` Binary() appends engineBuildFlags to `go build` — becomes plain go build; passes `NATIVE_HELPER_DIR` through.
- `$F/internal/testbuild/testbuild.go:100` Main installs nonet.OnlyLoopback (seatprobe/nonet_test.go:11) — CONSTRAINT: any helper download inside tests is refused off-loopback; tests need a prebuilt helper.
- `$F/internal/seatprobe/separatedbuild_test.go:93-94` skips ErrEngineRequired when !EngineCompiledIn — re-key on `Require`.
- `$F/internal/seatprobe/scanned.go:29` ErrEngineRequired "binary with the OCR engine compiled in"; `build.go:348-368` reads ocr_engine_absent, error text "(-tags tessocr)" — reword/re-key.
- `$F/internal/seatprobe/fixture_cgo_test.go:1` `tessocr && cgo`, TestScannedFixtureSpanIsFoundOnPage1ByTheRealEngine (CI anchors on the name, hooks.yml:600) — re-gate; the CI anchor goes with the deleted OCR step (§III.1.5).
- `$F/internal/fetchcache/enginesuccess_test.go:1` `tessocr && cgo`; `blankreceipt_cgo_test.go:1` `tessocr`; `corpus_cgo_test.go:1,44` `tessocr`, FEOV_OCR_CORPUS, TestCorpusGoldens — re-gate.
- `$F/internal/fetchcache/enginestub_test.go:1,25-30` `!tessocr || !cgo` TestAScannedReadOnAStubBuildRefusesLoudly; `$F/internal/cli/enginestub_test.go:1,14-34` TestFetchOnAStubBuildStatesTheEngineIsAbsent; `$F/internal/cli/fetchocr_test.go:121,126` pins the stub's error text — become "no helper" tests (run everywhere, no tag).
- `$F/internal/cli/ocrcite_real_test.go:1` `tessocr`, TestOCRCiteOnRealScan, FEOV_OCR_E2E-gated, run by hand only (not in any CI step) — re-gate.
- `$F/internal/libcprobe/probe_cgo.go:1`, `probe_cgo_test.go` — `tessocr && cgo && linux`, probes the zig musl libc's bounded scans (#1598) via cgo — DELETED; becomes `selftest-libc` in `feov-ocr`, run on both Linux targets by `native-run`.
- `$F/releasegate/fuzz/*` uses testbuild (fuzz_test.go, hookinvocation_test.go, testmain_test.go) — unchanged; fuzz never had the engine (release-fuzz-build is a plain build).

---------------------------------------------------------------------------------------------------
### E. C-stack build tooling (third_party/pins) — 4 artifacts

Enumerate: `ls $F/third_party/pins; grep -rn "build-cstack\|staticproof\|PINS.txt" --include=* -r .github scripts plugins docs | grep -v "^plans/"`

- `$F/third_party/pins/build-cstack.sh:2-12` builds the static tesseract+leptonica prefix per target; `env` mode `:79-123` prints CGO_* + CSTACK_EXTLDFLAGS for a GO cgo link (comments name `go build -tags tessocr`, Go's darwin -lresolv/CoreFoundation/Security tbd stubs) — DELETED into `native/build.sh` (zig c++ links feov-ocr itself); `env`'s cgo contract goes, and the darwin stubs are re-derived for a C++-only link.
- `build-cstack.sh check <binary>` `:53-54` -> `go run staticproof` — becomes `go -C scripts run ./staticproof -artifact <name> <binary>`, aimed at helpers and every Go asset.
- `$F/third_party/pins/staticproof/main.go:4,109-124` PE allowlist {kernel32, ntdll, api-ms-win-crt-*}, Mach-O {libSystem.B, libresolv.9, CoreFoundation, Security} — MOVES to `scripts/internal/staticproof`; allowlist per artifact (§III.1.5).
- `$F/third_party/pins/PINS.txt:1-14,26-28` pins; comment "go:embed'ed at internal/tessocr/eng.traineddata; TestEmbeddedTraineddataPin" — the traineddata pin stays and its comment names the helper; ONNX Runtime, the MSVC toolset (`-vcvars_ver`) and the two model exports join.
- `.github/actions/go-cache` — no C-stack reference (checked) — unchanged.

---------------------------------------------------------------------------------------------------
### F. CI workflow (.github/workflows/hooks.yml) — 3 jobs build the C stack, 1 publishes, 2 contract tests parse it

Enumerate: `grep -n "^  [a-z0-9_-]*:$\|tessocr\|cstack\|linkmode\|for cmd in\|for pair in\|-o \"\|go test" .github/workflows/*.yml`

- `hooks.yml:275-328` job `ocr-corpus` (ubuntu only): steps "C stack for the OCR engine (cache)" :301 (actions/cache key `cstack-${{runner.os}}-hashFiles(PINS.txt, build-cstack.sh)`, skipped on push), "(build on miss)" :308, "the corpus, real engine, statically linked" :314 runs `go test -tags tessocr -ldflags '-linkmode external -extldflags "-static"' -run '^TestCorpusGoldens$' ./internal/fetchcache/` — CHANGES: needs `native-build`, plain `go test` with `NATIVE_HELPER_DIR`.
- `hooks.yml:387-` job `feov-record` matrix [ubuntu-latest, windows-latest]; `:532-549` comment (Windows legs prove the STUB compiles); `:550` cache step; `:557` build on miss (Linux only); `:570-616` "the OCR engine's own suite, real engine, statically linked": tagged flags `:577`, engine_tests DERIVED by `grep -l '^//go:build tessocr' internal/fetchcache/*_test.go` `:588-590`, runs ./internal/tessocr ./internal/libcprobe and narrowed fetchcache/seatprobe, `PASS|SKIP` acceptance `:612-616` — CHANGES: the step is deleted; both legs set `NATIVE_HELPER_DIR` and the default leg `:657` runs the engine tests.
- `hooks.yml:711` `feov-record` race run-handle test — no `Require` package — unchanged.
- `hooks.yml:726-823` job `release-fuzz-build`: `go build -o "$out/bin/" ./cmd/...` plain — unchanged (fuzz never ran OCR).
- `hooks.yml:1159-1216` job `fetch-bin-platforms` (ubuntu/macos/windows): fixture plugin with `tools/cmd/alpha,beta` (`:1182`), asset `<n>_<os>_<arch><ext>` + SHA256SUMS over file:// — CHANGES: a helper fixture, an `ASSETS` line and a source-built root join.
- `hooks.yml:1358-1362` job `scripts` comment: sc-doctor discovers binaries with ReadDir(<plugin>/tools/cmd) — constraint: helper source must NOT live under tools/cmd.
- `hooks.yml:1583-1594` `debate-sim` guard steps (fetch-bin copies, buildid, marketplace pinning, event schema) — unchanged.
- `hooks.yml:1672-1700` job `release` (ubuntu-latest, ONE job, `if:` `:1682`, `needs: [test, feov-record, gray-area]` `:1683`, setup-go cache:false — "A RELEASE READS NO CACHE") — CHANGES: needs the native jobs, `!cancelled() && !failure()` (§III.1.5).
- `hooks.yml:1777-1825` step "Cross-compile": `for cmd in cmd/*/` x `for pair in windows/amd64 windows/arm64 darwin/amd64 darwin/arm64 linux/amd64 linux/arm64` `:1792`; CAPABILITY-KEYED `if [ -x third_party/pins/build-cstack.sh ]` `:1799` -> builds stack per target, `go build -tags tessocr -trimpath -ldflags "-s -w -linkmode external -extldflags '$CSTACK_EXTLDFLAGS'"` `:1815` for EVERY FEOV command, then `build-cstack.sh check` each `:1817`; else plain `GOOS GOARCH go build -trimpath -ldflags "-s -w"` `:1819` (cgo NOT disabled: linux/amd64 is the host target, so it links dynamically for every plugin); `cd $DIST && sha256sum * > SHA256SUMS` :1825 — CHANGES: one `CGO_ENABLED=0` line for every plugin; `releaseassets -check` before `sha256sum`.
- `hooks.yml:1854-1872` step "Release": idempotent re-run compares downloaded SHA256SUMS byte-for-byte with what this run built, else fails "artifacts DIFFER"; `gh release create "$TAG" "$DIST"/*` from ONE directory — CHANGES: the native artifacts are downloaded into `$DIST` before `releaseassets -check`; MSVC/PE and zig output must be reproducible (/Brepro, pinned toolset), checked by §V.6.
- `.github/workflows/race-nightly.yml:27` comment ("The OCR engine's C stack"), `:80` `go test -race` over the shipped graph — CHANGES: `NATIVE_HELPER_DIR=none`; comment rewritten.
- `nightly-alarm.yml`, `labels.yml` — no C-stack reference (checked) — unchanged.

---------------------------------------------------------------------------------------------------
### G. Binary enumeration / provisioning (tools/cmd as the list of shipped binaries) — 9 enumerators [re-run at `68a735d51`]

Enumerate: `grep -rn 'tools/cmd' scripts plugins/*/hooks plugins/prosthetic-conscience/tools/internal/doctor .github --include=* | grep -v _test`

All of these treat "shipped binary" == "directory under plugins/<p>/tools/cmd/, built by `go build ./cmd/<name>`". A C++ helper fits none of them.

- `scripts/fetchbingen/fetch-bin.sh:24-28` source-tree guard: exits 0 when `$ROOT/tools/go.mod` exists (PR #1724) — every installed plugin has one, so on `68a735d51` nothing is ever fetched for a consumer — PRECONDITION (§II.4), not this plan's change.
- `scripts/fetchbingen/fetch-bin.sh:143-154` up_to_date: every tools/cmd/<n> has bin/<n>[.exe] and .fetch/installed == tag; `:246-262` missing list from tools/cmd; `:267-285` downloads SHA256SUMS, asset `${n}_${os}_${arch}${ext}`, verifies EVERY asset before installing ANY (all or none); `:286` stamps .fetch/installed — CHANGES: helpers and `ASSETS` join both lists (F1); `.fetch/source-built` refusal (§III.1.4).
- `plugins/{frank-exchange-of-views,gray-area,prosthetic-conscience}/hooks/fetch-bin.sh` (the same lines +1) — GENERATED copies of the above (`go -C scripts run ./fetchbingen`, gate "Every plugin's fetch-bin.sh matches the one authored copy", hooks.yml:1586).
- `scripts/fetchbingen/main.go:120-138` targets(): plugins with tools/cmd — unchanged.
- `plugins/prosthetic-conscience/tools/internal/doctor/main.go:60` AssetName `<name>_<goos>_<goarch>[.exe]`; `:73-91` binariesOf ReadDir(tools/cmd); `:102` siblingBinaries walks every installed SC plugin's tools/cmd; `:490` fetchRelease via `gh release download --pattern <asset> --pattern SHA256SUMS`; `:533-538` buildFromSource `go build -C tools ./cmd/<name>` fallback; `:550-570` stampInstalled; `:580-608` fixWith — CHANGES: binariesOf lists helpers and `ASSETS`; helpers are fetch-only; the source-built refusal.
- `plugins/prosthetic-conscience/tools/internal/doctor/release_contract_test.go:64-74` `rePairs`/`reOutput`; `:82` rePairs over the WHOLE file; `:110-114` reOutput after `\n  release:\n`; `:241` TestSHA256SUMSCoversEveryContractedName — CHANGES: both searches narrowed to the release job's span; asserts `releaseassets -check` precedes `sha256sum`.
- `plugins/prosthetic-conscience/tools/internal/doctor/fetchbin_contract_test.go:255,271,284` — contract tests over fetch-bin.sh — gain a helper, an `ASSETS` line and the source-built case.
- `scripts/pluginparity/main.go:121-143` counts tools/cmd dirs vs docs/setup-script.md ("22 at the time of writing", setup-script.md:148); `:147-172` requirements.json `_hook_binaries` vs tools/cmd (PC requirements.json:65, gray-area :28; FEOV declares none); `:214-` shippedHarnessProblems ("Every directory under plugins/*/tools/cmd/ ships") — unchanged; helpers stay out of tools/cmd, and `releaseassets` generates and checks the helper set.
- `scripts/bootstrap-plugins.sh:85-97` cold bootstrap: `go build ./cmd/<name>` for every cmd in every cached plugin, no stamp written — unchanged for Go; helpers never built here; the next `ensure` installs the tag's set.
- `scripts/universe.sh:437-453` builds every cmd from the CHECKOUT with plain go build into cache, `:459-481` mirrors into the staged marketplace, `:482-496` stamps .fetch/installed, `:498-` asserts `ensure` does nothing — CHANGES: `build` writes `.fetch/source-built`; `build --native` builds and stages helpers; the assertion also runs `fetch` (F5, §III.1.6).
- `scripts/marketplacegen/main.go:15` "A plugin without tools/cmd has no binaries and no tag" — unchanged.
- `scripts/versionguard/main.go` — tag/manifest agreement, per-binary version constant sweep — unchanged; helpers take no tag namespace of their own (the release job's "tag names a plugin with tools/go.mod" check, hooks.yml:1709-1720, would refuse one).
- `shared/gcpolicy/gcpolicy.go:4` every main under tools/cmd calls Short/Long — unchanged (helpers are not Go).
- `$F/internal/setup/run.go:214-232` setup preflight locates feov-record beside itself and checks its schema epoch — CHANGES: also handshakes each helper and prints its state; never refuses for it (F9).

---------------------------------------------------------------------------------------------------
### H. Text naming the in-process engine — 191 lines [re-run at `68a735d51`]

Enumerate: `grep -rn "statically linked\|compiled into\|C stack\|tags tessocr\|build-cstack\|compiled in\b\|ErrNotCompiledIn\|EngineCompiledIn\|IN THE BINARY\|linkmode external" plugins/frank-exchange-of-views docs AGENTS.md scripts .github` (case-insensitive; full listing `~/.claude/scratch/vecplan/census/H-r3-68a735d51.txt`). The disposition of every line is in Part 2's "Agent-facing text and comments"; the seat-facing and generated subset:

- `$F/internal/cli/fetch.go:62` SOURCE of "read by the LOCAL OCR ENGINE compiled into this binary" — rewrite; regenerate with `go -C scripts run ./agentgen`.
- `plugins/frank-exchange-of-views/agents/{blue-researcher, blue-synthesizer, blue-lane, red-lens-voice, red-lens-logic, red-lens-architecture, red-lens-adversary, red-lens-computation, red-lens-evidence, red-lens-dark-side}.md` — 10 GENERATED copies (agentgen -check) — regenerate, never hand-edit.
- `$F/internal/difftest/testdata/manual_runs_every_help_page_live.golden:960,1538,2100` (the fetch help, three renderings) and `:2142` (`ocr read`) — regenerate (UPDATE_GOLDENS=1), read each changed line.
- `$F/internal/cli/ocr.go:244` `ocr read` Long help — rewrite (operator tree, not a seat verb; still in manual golden).
- `$F/internal/fetchcache/pageread.go:92-93` error a seat reads in ocr_reason: "not compiled into this binary … build with -tags tessocr" — rewrite (seat-facing via fetch summary).
- `$F/internal/cli/fetchocr_test.go:121,126` — a stub error text pinned in an assertion — rewrite with the error.
- `$F/internal/seatprobe/build.go:368`, `scanned.go:29` — probe-harness error text — rewrite.
- `$F/internal/cli/fetchsummary.go:91,414` `ocr_engine_absent` field + `ocr_reason` — field name stays meaningful ("engine absent"), meaning widens.
- Not matched by the grep, same concept: `AGENTS.md:91` (universe rule) and `docs/local-development.md`'s universe section — extended for `--native` (§III.1.7).
- Other sense, unchanged: `record/cast.go:43`, `report/site.go:119`, `seatprobe/surfacecoverage_test.go:157`, `setup/run_test.go:115`, `cli/cli_test.go:134`, `difftest/golden_test.go:88`, `integration/surface/hookgraph_test.go:74`.

---------------------------------------------------------------------------------------------------
### I. Gates and dev tooling that encode the tag — 3 sites

Enumerate: `grep -n "tessocr\|C stack" scripts/check/*.go`

- `scripts/check/gates.go:323-325` gate `ocr-corpus` kindExtern skip text "needs the OCR engine's C stack… run the tagged suite… ANY edit under internal/tessocr re-arms it… so all eight fail together" (there are 13 goldens) — rewrite skip; re-arm set widens to the helper's digest inputs.
- `scripts/check/gates.go:167-174` raceScope STATED LIMIT: "the engine behind -tags tessocr is not race-instrumented anywhere… musl cgo objects under the race build's linking broke…" — CHANGES (improves): with no cgo, the race gate covers the Go client + pipe protocol; the limit text goes (helper itself is not raced — say so).
- `scripts/deadcode` — runs the default build; libcprobe/engine_cgo never seen — unchanged; the new Go protocol client becomes visible to it (must be reachable from cmd/feov-record).

---------------------------------------------------------------------------------------------------
### J. Engine identity — where it is computed, stored, compared — 1 producer, 4 stores, 3 comparers

Enumerate: `grep -rn "Identity()\|OcrEngine\|ocr_engine\|\.Engine\b" $F/internal --include=*.go | grep -v "s.Identity()\|_test\|record.pb.go\|decode_gen\|hookcmd"`

Producer:
- `$F/internal/tessocr/tessocr.go:100` Identity() = `tesseract@<pin>+leptonica@<pin>+eng@<12 of traineddata sha>+cstack@<cstackPin>+tessocr@<sha256[:12] of package non-test source incl shim.cpp/shim.h>` — CHANGES: `cstack@` gives way to `feov-ocr@<source digest[:12]>` from `digest_gen.go` (F8, L1); computed from compiled constants, so it never spawns or fetches the helper.
- `$F/internal/fetchcache/pageread.go:66` TessocrPageEngine.Identity — delegates — swaps.
Stores (all plain strings; no schema change, no epoch bump, no migrate):
- `$F/internal/fetchcache/pageread.go:405-422` ReadingRecord.Engine `json:"engine"` in `<run>/cache/…/pages/reading.json`; per-page receipt `:330` Engine — cache only (not in run-archive) — values change; every cached reading re-reads once.
- `$F/internal/record/recordpb/record.proto:1320-1328` Cite.ocr_engine (field 14) + ocr_text_sha (15), stamped by `$F/internal/cli/blue/cite.go:157` from pagelocate.go:144 — IN THE RECORD and run-archive; old runs keep old strings; nothing compares a stored ocr_engine with the current Identity (record.go:1617-1621 only requires non-empty; evidenceview.go:119,443 displays it) — unchanged code; value grammar changes.
- `$F/internal/record/migrate/remap.go:133` lists Cite.ocr_engine among carried fields — unchanged.
- `$F/internal/tessocr/testdata/corpus/*/reading.golden:1` (13 files) — regenerate.
Comparers:
- `pageread.go:299` page receipt reuse; `readscanned.go:108` whole-reading reuse — unchanged logic.
- `$F/internal/cli/ocr.go:210,297`, `fetchsummary.go:175,416` print the engine string — unchanged.
Note: PDF rendering identity is separate — `$F/internal/fetchcache/pdfextract.go:67-81` extractorIdentity = go-pdfium (WebAssembly via wazero, pure Go) — unaffected; PDFium is not part of the C stack.

---------------------------------------------------------------------------------------------------
### K. Docs and plans describing the in-process engine — 3 docs, 1 live plan, 1 AGENTS.md rule, 9 OCR plans (historical-in-effect)

Enumerate: `grep -rln "tessocr\|build-cstack\|linkmode external\|compiled into the binary" docs plans README.md plugins/*/README.md`

Docs (live, must be rewritten):
- `docs/c-stack.md:1-46` whole doc: "feov-record links a static C/C++ stack… behind the tessocr build tag… Go build links with -linkmode external, zig as external linker, for every command"; "Adding a C or C++ library… If it builds only with a platform's native toolchain, the release job needs that platform's runner, and that is a decision" — rewrite: stack builds helpers; MSVC windows runner for feov-embed IS that decision.
- `docs/local-development.md:302-305` "A plain build has no OCR engine. Without -tags tessocr…" — rewrite (plain build has the client; OCR needs a helper); its universe section gains `build --native` and `.fetch/source-built`.
- `docs/setup-script.md:73` "reads a scan with the OCR engine compiled into the binary, so nothing needs provisioning" — FALSE after; helpers are provisioned with the hook binaries; `:148` "22 hook binaries" count — unchanged, helpers stay out of tools/cmd.
- `AGENTS.md:91` "AFTER `universe.sh build`, YOU MUST prove the universe carries your change" — extended for `--native` (R5).
Plans:
- `plans/vector-search.md` — LIVE ("approved 2026-11-03, not started", plans/README.md:22): designs llama.cpp statically linked INTO feov-record via build-cstack.sh/-tags tessocr, internal/vecembed cgo shim, GGUF model beside feov-record provisioned by fetch-bin.sh + sc-doctor (:8,:61,:71,:88-96,:114,:329-332,:458-506) — SUPERSEDED in its native half by this design; revised in its own PR (Part 3), which also rewrites its plans/README.md row.
- `plans/local-ocr.md` — header still "under review" but SHIPPED (code cites it: tessocr.go, page.go, lattice.go, pageread.go, pagerender.go, ocr.go, hooks.yml x2, receipts_test.go, fuzz_test.go, promptsize_test.go) — historical record; do not rewrite; code comments citing "Wave 1" for the tag split need rewording where they describe the build.
- Implemented/measured, verification loops hold tagged commands (historical, leave): `plans/ocr-dashed-rules.md:389-391,408`, `ocr-dpi-derivation.md:236-237`, `ocr-page-corpus.md:263-264`, `ocr-background-norm.md:97-98`, `ocr-cite-derives-page.md:515,868-914` (draft status but implemented), `ocr-text-cell-rows.md:182`, `ocr-resolution-and-antialiasing.md:65,161` (REJECTED), `ocr-declare-the-resolution.md` (NOT SHIPPED), `ocr-level-subcolumns.md`, `ocr-end-to-end-evidence.md` (proposed), `ocr-blocked-page-resilience.md` (superseded).
- `plans/packaging-goreleaser.md` — not started; describes the hand-rolled release loop — a helper/multi-runner release strengthens or kills its case; note only.

---------------------------------------------------------------------------------------------------
### L. Explicitly unchanged (checked)

- Other cgo in the tree: only `internal/tessocr/engine_cgo.go` and `internal/libcprobe/probe_cgo.go` contain `import "C"` (`grep -rln 'import "C"' plugins --include=*.go`). SQLite is modernc (pure Go); PDFium is wazero/WASM.
- gray-area and prosthetic-conscience modules: no C stack, no tag. The release's plain branch leaves cgo on for linux/amd64, the host target, so any of their commands that imports `net` or `os/user` publishes dynamically linked (not measured per command); S1 covers them.
- `.qlty`, `.gitattributes`, `.gitignore`: no tessocr/shim/cstack entries (note: shim.cpp/.h are NOT LF-pinned, which is why identity folds CRLF — the source digest keeps the fold).
- Event schema / record.proto DDL: no field added or changed by the move (identity is a string value) — no epoch bump.
- release-fuzz (4 parts + tally): builds plain, never exercised OCR — unchanged.
