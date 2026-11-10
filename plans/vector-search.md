# Hybrid lexical + vector search over the report and cached sources (`feov-embed` + `gte-small` int8)

> STATUS: proposed (decision document for [#1171](https://github.com/ctoforaday/special-circumstances/issues/1171); awaiting human approval from `gblock`). Revised per `plans/native-helpers.md` Part 3, `REVIEW.md`, and `plan-auditor` findings (2026-11-10). Audited: 2026-11-10 (`plan-auditor-agy` Round 1 `10375d17` FAIL → fixed; Round 2 `5293331a` PASS; Round 3 `b7e94e63` PASS — two consecutive PASSes).

Ruled by `gblock` across 2026-10-31, 2026-11-01, 2026-11-04 to 2026-11-06 (`plans/native-helpers.md`), and 2026-11-10:
1. **Hybrid in one go**: ship SQLite FTS5 BM25 (`porter unicode61`) and dense vector search together, fused via Reciprocal Rank Fusion ($\text{RRF}, k=60$).
2. **Ranked read is distinct from exhaustive `--quote` and `--match`**: `--quote` (literal substring) and `--match` (RE2 regex) stay exhaustive across every view, returning every match with zero truncation. `--search` is a ranked top-$K$ read whose output always states that it is ranked and cut, how many chunks and sources were searched, which cached sources had no extracted text, and that absence from top-$K$ proves nothing.
3. **Thin C++ native helper executable (`feov-embed`)**: per `plans/native-helpers.md` Part 3, vector embedding runs in a standalone native helper executable (`feov-embed`), built with ONNX Runtime 1.31.1 (minimal Release build with reduced operator set `ops.config`) across all 5 release targets (F15). Communicates with Go via a framed stdio protocol over a private descriptor (`dup`/`dup2`). Go compiles with `CGO_ENABLED=0` across all published binaries.
4. **Model selection by measured quality/size/speed tradeoffs**: pin per-arch ONNX Runtime 8-bit quantized models (`gte-small.int8.amd64.ort` and `gte-small.int8.arm.ort`, `33,844,232` bytes each, 12-layer BERT bi-encoder, 384 dims, mean pool, symmetric query/doc encoding) hosted on a one-time release (`model-gte-small-v1`, F10), pinned by SHA256 in `native/feov-embed/PINS.txt`, and provisioned via `fetch-bin.sh` and `sc-doctor` (F1, L2).
5. **Empirical audit of `--match` (ruled 2026-11-10)**: *"If `--search` replaces the use case for match, remove match. But audit it for that — it might be that we want all three universally."* Section III.1 reports the full audit over all 2,252 seat transcripts (`~/.claude/scratch/vecplan/replay/`): `--search` replaces synonym/paraphrase alternations (48.2% of report `--match` calls), but **cannot** replace exhaustive voice/tell absence sweeps (15.9%), multi-ID/formula alternations (10.1% on report; 16.4% across `board`, `findings`, `inquiries`, `work`, `evidence`, `changes`, `avenues`), or regex boundary/character-class checks. Accordingly, **`--match` is retained alongside `--quote`**, and Section III.1 presents the evaluated fork on whether `--search` scopes to the prose corpora (`show report` and `show source`) or all selector-bearing views universally.

---

## I. Summary & Goals

### 1. Value Case & Problem Statement

Since [#1171](https://github.com/ctoforaday/special-circumstances/issues/1171) was filed on 2026-09-26, three prerequisites shipped on `main`:
- HTML text extraction (`extractHTML` in [`internal/fetchcache/htmlextract.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/htmlextract.go), PR #1673).
- Single-source literal and regex search (`feov-record show source --sha <sha|--anchor C-…> --quote|--match`, PR #1662).
- Removal of token-Jaccard `lens near-match` (`internal/record/nearmatch.go`, retired 2026-10-24).

What remains broken is **vocabulary mismatch** and **cross-source discovery**:
- Across October 2026 seat transcripts (`~/.claude/scratch/vecplan/replay/SUMMARY.md` and `acts.jsonl`), seats executed:
  - **1,094** `show report --match` statements (`239` returning zero hits, `649` hit, `6` error, `200` unattributed) and **400** `show report --quote` statements (`106` returning zero hits, `205` hit, `89` unattributed) — **1,494** report selector calls in total (**345** with zero hits), plus **215** `show report | grep` pipelines.
  - **121** `show source --match` statements (`5` zero, `83` hit, `33` unattributed) and **12** `show source --quote` statements (`3` hit, `1` error, `8` unattributed), plus **2,452** `grep`/`rg` commands over `<run>/cache/`.
  - **354** `--match` and **61** `--quote` statements across structured record views (`board`: `117`/`19`, `findings`: `106`/`15`, `inquiries`: `75`/`2`, `work`: `20`/`1`, `evidence`: `17`/`8`, `changes`: `11`/`15`, `avenues`: `8`/`1`).
- `show source` currently requires `--sha` or `--anchor` upfront; a seat cannot search across all fetched sources in `<run>/cache/` in a single act.
- When a seat searches for a concept using words that differ from the document's exact phrasing (55 of the 79 verified failed-then-reworded seat searches in `replay.jsonl` are plain wording misses, and 17 are failed 3+-branch synonym alternations), literal `--quote` and regex `--match` return 0 hits.

### 2. Why Pure-Go Embedders and In-Process Linking Failed the Bar

Before selecting the native helper path, we wrote a pre-committed pass bar (`~/.claude/scratch/vecplan/bakeoff/BAR.md`: B1 `CGO_ENABLED=0/1` build, B2 $\ge 5.0\text{ chunks/s}$ on 200-word chunks on one core and cold query $\le 1.5\text{ s}$, B3 cosine mean $\ge 0.99$ / min $\ge 0.95$ vs FP32 reference, B4 hybrid 95% CI $> 0$ over BM25 on `SciFact` and `NFCorpus` and replay `Hit@5` within `0.05` of BM25) and benchmarked pure-Go candidates:
- **Static token-table lookup (`go-potion` + `minishlab/potion-retrieval-32M`)**: fast (`1,187–1,278 docs/s`), 129 MB FP32 weights, but has no transformer layers (embeds by averaging static token vectors). On BEIR `SciFact`, standalone `potion-retrieval-32M` (`0.633` nDCG@10) loses to stemmed BM25 (`0.682`), and its hybrid RRF gain over BM25 is `+0.018` (`95% CI [-0.005, +0.041]`, crossing zero and failing B4). Furthermore, `go-potion` truncates/rewrites its model file on load (`os.Create(modelPath)`).
- **Pure-Go BERT (`rcarmo/gte-go` + `thenlper/gte-small`)**: passes B4 quality (`+0.055` FP32 hybrid nDCG@10 on `SciFact`, `+0.042` on `NFCorpus`), but out-of-the-box runs naive FP32 3-nested-loop matrix multiplication in pure Go at **1.23 chunks/s** (`812 ms/chunk`), and its upstream AVX2 assembly (`gebpKernelAVX2`) had a shared packing buffer raced by concurrent callers. After fixing the buffer race and AVX2 path in 3 scratch commits (`~/.claude/scratch/vecplan/go/upstream/gte-go`), `gte-go` matched the default implementation (`cosine 0.9999999`) and reached **3.5 chunks/s** on one core — still below the pre-committed B2 bar of **5.0 chunks/s** (`bakeoff/sweep.log`).

**Why in-process static linking was superseded**: Part 0 of [`plans/native-helpers.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plans/native-helpers.md) (`~/.claude/scratch/vecplan/ortstatic/REPORT.md` §4) proved that in-process ONNX Runtime 1.31.1 cannot be cross-linked statically on Windows under Go's cgo linker (MinGW/zig fails on MASM `.asm` kernels and MSVC-only flags even after rebasing MSYS2's 37-file patch set, and Go's MinGW `ld` cannot link MSVC-compiled `.lib` archives due to 58 unresolved MSVC ABI/STL symbols). Moving native execution out of `feov-record` into a standalone helper executable (`feov-embed`) behind a framed stdio protocol allows clean native MSVC `/MT` compilation on Windows (`3.67 MB`, OS-shipped DLLs only) and `zig c++` compilation on Linux/macOS. It keeps all published Go binaries pure Go (`CGO_ENABLED=0`), isolates worker crashes or hangs from `feov-record`, and enables independent helper verification.

### 3. Model Selection & Quality Tradeoffs

We evaluated six embedding models and three SQLite FTS5 BM25 tokenizers across two gold-labeled scientific retrieval benchmarks (**BEIR SciFact**: 300 queries, 5,183 docs; **BEIR NFCorpus**: 323 queries, 3,633 docs) and the **79-row FEOV seat replay dataset** (50 report rows, 29 source rows; 55 plain wording misses; 200-word source chunks with 40-word overlap / 160-word stride; RRF fusion $k=60$ with `bm25/porter unicode61`). Raw artifacts and logs live in `~/.claude/scratch/vecplan/results/beir.log`, `results/replay.log`, `bakeoff/score.json`, `ortq/score.log`, `ortq/RESULT.md`, and `ortstatic/REPORT.md`.

The table below separates the **measured 8-bit ONNX Runtime export (`gte-small.int8.ort`, `ort8`)** from the **FP32 `sentence-transformers` / `ort32` sweep** so that quantized and unquantized figures are never conflated:

| Model / Method (Precision & Source Log) | Architecture & Pooling | Export Size | SciFact nDCG@10 (Vec / Hybrid) | SciFact Hybrid $\Delta$ vs BM25 (95% CI) | NFCorpus nDCG@10 (Vec / Hybrid) | NFCorpus Hybrid $\Delta$ vs BM25 (95% CI) | BEIR Mean Hybrid nDCG@10 | Replay All-79 Hybrid (`h@1` / `h@5` / `h@10` / `MRR`) | Replay Wording-55 Hybrid (`h@1` / `h@5` / `h@10`) |
|---|---|---|---|---|---|---|---|---|---|
| `bm25/porter unicode61` (`beir.log`, `ortq/score.log`) | Lexical FTS5 | 0 MiB | — / `0.682` | baseline | — / `0.322` | baseline | `0.5020` | `0.544` / `0.848` / `0.873` / `0.670` | `0.582` / `0.873` / `0.891` |
| **`gte-small.int8.ort`** (**8-bit ORT `ort8`**, `ortq/score.log`) | 12L BERT, 384d, **Mean pool**, no prefix | **33.8 MiB** (`33,844,232` B) | **`0.721` / `0.731`** | **`+0.048 [+0.026, +0.069]`** | **`0.344` / `0.362`** | **`+0.040 [+0.027, +0.054]`** | **`0.5465`** | `0.494` / **`0.861`** / **`0.924`** / — | `0.509` / **`0.873`** / **`0.927`** |
| `gte-small` (FP32 / `ort32`, `beir.log`, `bakeoff/score.json`) | 12L BERT, 384d, Mean pool, no prefix | 127.0 MiB (`133,181,224` B) | `0.727` / `0.738` | `+0.055 [+0.033, +0.077]` | `0.348` / `0.364` | `+0.042 [+0.028, +0.056]` | `0.5510` | `0.519` / `0.835` / `0.937` / `0.659` | — / `0.873` / — |
| `bge-small-en-v1.5` (FP32 sweep `beir.log`, `replay.log`; ~35 MiB Q8) | 12L BERT, 384d, CLS pool, query prefix req. | 127 MiB FP32 (~35 MiB Q8) | `0.713` / `0.730` | `+0.047 [+0.027, +0.067]` | `0.344` / `0.369` | `+0.047 [+0.033, +0.062]` | `0.5495` | `0.570` / `0.835` / `0.911` / `0.681` | — |
| `all-MiniLM-L6-v2` (FP32 sweep `beir.log`, `replay.log`) | 6L BERT, 384d, Mean pool | 87 MiB FP32 (~23 MiB Q8) | `0.645` / `0.719` | `+0.036 [+0.014, +0.058]` | `0.316` / `0.356` | `+0.034 [+0.020, +0.048]` | `0.5375` | `0.443` / `0.848` / `0.949` / `0.604` | — |
| `snowflake-arctic-embed-xs` (FP32 sweep `beir.log`, `replay.log`) | 6L BERT, 384d, CLS pool, prefix req. | 87 MiB FP32 (~23 MiB Q8) | `0.645` / `0.706` | `+0.024 [+0.002, +0.045]` | `0.309` / `0.347` | `+0.025 [+0.015, +0.036]` | `0.5265` | `0.481` / `0.785` / `0.886` / `0.616` | — |
| `potion-retrieval-32M` (FP32 `beir.log`, `replay.log`) | Static token table, 512d | 129 MB (FP32) | `0.633` / `0.700` | `+0.018 [-0.005, +0.041]` | `0.308` / `0.341` | `+0.019 [+0.008, +0.030]` | `0.5205` | `0.506` / `0.899` / `0.949` / `0.654` | — |
| `potion-base-8M` (FP32 `beir.log`, `replay.log`) | Static token table, 256d | 30 MB (FP32) | `0.506` / `0.619` | `-0.063 [-0.085, -0.042]` | `0.243` / `0.313` | `-0.009 [-0.021, +0.003]` | `0.4660` | `0.406` / `0.886` / `0.938` / `0.595` | — |

We also evaluated larger models (`jina-embeddings-v2-base-en` at 139 MiB, `google/embeddinggemma-2` at ~295 MiB / 300M params, `Nomic-Embed-Text-v2` at 488 MiB, and `bge-m3` at 605 MiB):
- **Why 110M–560M models (`embeddinggemma-2`, `bge-m3`, `nomic-v2`) are rejected**: they are **4× to 17× larger** (139–605 MiB vs. 33.8 MiB) and **5× to 10× slower** on CPU per chunk. In addition, `google/embeddinggemma-2` uses the gated `gemma` license on Hugging Face (requiring authentication tokens, which breaks unauthenticated `PINS.txt` / `fetch-bin.sh` downloads), whereas `gte-small` and `bge-small-en-v1.5` are MIT-licensed.
- **Why `gte-small.int8.ort` (`ort8`) is chosen**:
  1. **Passes the B4 quality bar in 8-bit quantization with 4× smaller model assets (`33.8 MiB` vs `127.0 MiB` FP32)**: `gte-small.int8.ort` hybrid beats `bm25/porter` by **`+0.048 [+0.026, +0.069]`** on `SciFact` (`0.731` vs `0.682`) and **`+0.040 [+0.027, +0.054]`** on `NFCorpus` (`0.362` vs `0.322`), retaining 99.2% of FP32 `gte-small`'s BEIR Mean Hybrid nDCG@10 (`0.5465` vs `0.5510`). In the FP32 sweep, `gte-small` is also the only model whose standalone vector 95% CI clears BM25 on both BEIR benchmarks (`SciFact` `+0.045 [+0.013, +0.075]`, `NFCorpus` `+0.026 [+0.005, +0.047]`; `bge-small-en-v1.5` standalone on `SciFact` is `+0.030 [-0.003, +0.061]`, crossing zero).
  2. **Stronger top-5 and top-10 recall on the 79-row seat replay benchmark**: `gte-small.int8.ort` (`ort8`) achieves **`0.861` `Hit@5` and `0.924` `Hit@10`** on All-79 (`0.873` `Hit@5` and `0.927` `Hit@10` on Wording-55; rescuing 4 of 12 BM25 misses with 3 losses, net +1 over BM25 `0.848` `Hit@5` and +4 over BM25 `0.873` `Hit@10`), whereas `bge-small-en-v1.5` (even in FP32) achieves `0.835` `Hit@5` and `0.911` `Hit@10` (and FP32 `gte-small` achieved `0.835` `Hit@5` and `0.937` `Hit@10`). Because `--search` returns a top-$K=10$ list, `Hit@5` and `Hit@10` govern whether the passage reaches the seat.
  3. **Self-contained metadata and symmetric encoding ([[facts-are-fields]])**: `gte-small` uses symmetric query/document encoding (no magic prompt prefix). By contrast, `bge-small-en-v1.5` requires prepending the Asymmetric Query Prefix `"Represent this sentence for searching relevant passages: "` to queries — an external string that is not recorded in the `.ort` model file.
  4. **Proven 8-bit fidelity and measured single-thread throughput (`ortstatic/REPORT.md` §1, §3, §5)**:
     - Against the FP32 `sentence-transformers` reference (`ref_gte.npy` over 300 texts in `speed.jsonl`), `gte-small.int8.ort` measured cosine similarity **mean `0.984411–0.984460`, min `0.967607–0.969149`** across all targets (`ortstatic/REPORT.md` §1, `ortq/score.log` line 1).
     - In the size-optimised minimal build (`--config MinSizeRel`, `-Os`), 1-thread `int8` throughput on 200-word chunks (`len200.jsonl`) measured **`12.72 chunks/s`** on `linux-amd64` (EPYC 9V45; `6.11 chunks/s` on slower EPYC 7763), **`5.87 chunks/s`** on `linux-arm64` (Neoverse-N2), and **`5.88 chunks/s`** on `darwin-arm64` (Apple M1 Virtual), adding `4.4–5.1 MB` (`ortstatic/REPORT.md` §1 & §3).
     - Because `-Os` (`MinSizeRel`) penalises x64 MLAS kernels by `1.24–2.2×` (`ortstatic/REPORT.md` §5), building with `--config Release` (`minrel`) recovers full unreduced speed on `linux-amd64` (**`7.12 chunks/s`** on EPYC 7763 vs `6.11 chunks/s` for `-Os`, matching the `33.9 MB` full Release build's `7.07 chunks/s`) for `+1.4 MB` (`6,057,720 B` added on `linux-amd64` zig). `native/build.sh` therefore compiles ONNX Runtime with `--config Release` (`minrel`).

### 4. Quantitative Success Criteria

| Gate | Metric | Target |
|---|---|---|
| **G1: Retrieval quality (gold benchmarks)** | Hybrid (`bm25/porter unicode61` + `gte-small.int8.ort` RRF $k=60$) nDCG@10 vs `bm25/porter` | 95% paired bootstrap CI lower bound $> 0$ on both `SciFact` (`+0.048 [+0.026, +0.069]`) and `NFCorpus` (`+0.040 [+0.027, +0.054]`) (`ortq/score.log`) |
| **G2: Embedding fidelity (int8 helper vs FP32 reference, S5)** | Cosine similarity between `feov-embed` (`gte-small.int8.ort`) and FP32 `sentence-transformers` `thenlper/gte-small` over the 300-text reference (`ref_gte.npy`) | Mean cosine $\ge 0.984$, min $\ge 0.966$ executed in `native-run` on all 5 release targets (`TestEmbedFidelity`) |
| **G3: CPU throughput, cold start, & deadline margin (S6 & B2)** | Single-thread throughput on 200-word chunks (`len200.jsonl`); cold 1-query latency; worst-case 16-chunk batch vs 120 s deadline | Throughput $\ge 5.0\text{ chunks/s}$ on `linux-amd64`, `linux-arm64`, and `darwin-arm64` (`TestEmbedSpeed`); cold query $\le 1.5\text{ s}$; 16-chunk batch $< 30\text{ s}$ ($\frac{1}{4}$ of 120 s deadline, `TestEveryOpFitsTheDeadline`) |
| **G4: Static binary proof across 5 targets (S1, S2, F3, F15)** | All published Go binaries (`CGO_ENABLED=0`) and `feov-embed` pass `staticproof` on all 5 targets (`linux-{amd64,arm64}`, `windows-{amd64,arm64}`, `darwin-arm64`) | 0 forbidden dynamic libraries; Windows PE allowlist admits `{KERNEL32, ADVAPI32, dbghelp, api-ms-win-core-path-l1-1-0}` for `feov-embed` |
| **G5: Exhaustive `--quote` and `--match` preservation** | `TestEverySelectorThatAViewAdvertisesSelects` (`internal/cli/selectorcensus_test.go`) | Both `--quote` and `--match` return 100% of matches with 0 truncation across every view in `selectorViews(t)` |
| **G6: Protocol State Integrity** | `show report --search` does NOT record `ServeReport` event | Lens calling `--search` does not advance reading pin without reading the whole report |

---

## II. Technical Context

### 1. Thin C++ Native Helper Architecture (`native/feov-embed/` & `native/build.sh`)

Following [`plans/native-helpers.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plans/native-helpers.md) Part 3, `feov-embed` is a standalone native helper executable:
- **Location**: `plugins/frank-exchange-of-views/native/feov-embed/` holding `main.cpp`, `ops.config` (reduced ONNX Runtime operator set from `ortstatic/REPORT.md` §2), `PINS.txt`, and `ASSETS`.
- **Runtime**: ONNX Runtime 1.31.1 built in minimal `Release` mode (`--minimal_build --disable_exceptions --disable_rtti --disable_ml_ops --include_ops_by_config --enable_reduced_operator_type_support --config Release`). Adds ~6.06 MB on `linux-amd64` (fully static ELF under `zig c++` / musl) and 3.67 MB on Windows (static MSVC `/MT` executable importing only OS-provided DLLs; F3).
- **Cross-Compilation Mechanics Across All 5 Targets (`native/build.sh`, per `ortstatic/REPORT.md` §3 and `plans/native-helpers.md` §II.4)**:
  - **`linux-amd64`**: `zig c++ -target x86_64-linux-musl`, static link (`-static`).
  - **`linux-arm64`**: `zig c++ -target aarch64-linux-musl`, static link (`-static`). Because ONNX Runtime's CMake passes `-march=armv8.2-a+bf16+fp16+i8mm+dotprod` and `cpuinfo` passes `-march=armv8-a` (which `zig c++` rejects when combined with `-mcpu`), the `aarch64` zig wrapper rewrites `-march=armv8.2-a+bf16...` to `-mcpu=generic+v8_2a+bf16+fp16+i8mm+dotprod` and strips `-march=armv8-a`.
  - **`darwin-arm64` (cross-compiled from Linux with `zig c++ -target aarch64-macos-none`)**: uses the `aarch64` `-march` rewrite wrapper and passes `CMAKE_OSX_ARCHITECTURES=arm64`, `Iconv_IS_BUILT_IN=TRUE`, `onnxruntime_DISABLE_DEVICE_DISCOVERY=ON`, stub headers for `CoreFoundation/CFTimeZone.h` and `sys/syscall.h`, a portable clog sink (avoiding `<os/log.h>`), adds `cpuinfo` alongside `onnxruntime_*` build targets, and links against `zig c++`'s shipped `libSystem.tbd` plus two generated `.tbd` stubs for `CoreFoundation` and `Foundation` (verified in `native-run` on `macos-15` where `otool -L` confirms only `/usr/lib/libSystem.B.dylib`, `CoreFoundation`, and `Foundation`).
  - **`windows-amd64` & `windows-arm64`**: built on `windows-2025` using the pinned MSVC toolset (`vcvarsall.bat <arch> -vcvars_ver=<pin>` from `native/PINS.txt`) with `--enable_msvc_static_runtime`, `onnxruntime_DISABLE_DEVICE_DISCOVERY=ON`, `/Brepro`, `/MT`, and `/D_ORT_NO_EXCEPTIONS`.
- **Strict Pinned Build Contract in `native/build.sh` (S8)**:
  - `native/build.sh` reads **only** inputs pinned by URL and SHA256 in `native/PINS.txt` and `native/feov-embed/PINS.txt`.
  - **No unpinned or local-scratch inputs**: `native/build.sh` must NEVER read `$ORT_STATIC_DIR`, must NEVER copy archives from `$HOME/.claude/scratch/...`, and must NEVER run an unpinned `pip install flatbuffers`. Instead, `flatbuffers-25.2.10-py2.py3-none-any.whl` (required by ONNX Runtime's `reduce_op_kernels.py` / `parse_config.py` during the reduced-operator build) is pinned by URL and SHA256 in `native/feov-embed/PINS.txt`, downloaded via `fetch_pinned`, and unpacked into `$WORK/pydeps` on `PYTHONPATH`.
  - **Per-arch model filename mapping & `$OUTDIR` staging**: `native/feov-embed/PINS.txt` and `ASSETS` name the `amd64` export `gte-small.int8.amd64.ort` and the `arm64` export `gte-small.int8.arm.ort` (`--target_platform {amd64,arm}`). `native/build.sh` maps `ARCH=amd64 -> MODEL_ARCH=amd64` (`gte-small.int8.amd64.ort`) and `ARCH=arm64 -> MODEL_ARCH=arm` (`gte-small.int8.arm.ort`). When building `feov-embed`, `native/build.sh` stages into `$OUTDIR`:
    1. `${HELPER}_${OS}_${ARCH}${EXT}` and `${HELPER}${EXT}` (matching `feov-ocr`'s existing `build.sh` contract so `NATIVE_HELPER_DIR` in `native-run` and `feov-record` finds the unsuffixed helper binary directly), and
    2. `gte-small.int8.${MODEL_ARCH}.ort` (for `release`) and `gte-small.int8.ort` (for `NATIVE_HELPER_DIR` in `native-run` and `feov-record`).
    In `hooks.yml`'s `release` job, before `releaseassets -check` runs on `$DIST`, unsuffixed helper binaries (`feov-ocr`, `feov-ocr.exe`, `feov-embed`, `feov-embed.exe`) and `gte-small.int8.ort` are removed from `$DIST`, leaving only `${HELPER}_${OS}_${ARCH}${EXT}` and the declared `ASSETS` exports (`gte-small.int8.amd64.ort` and `gte-small.int8.arm.ort`).
- **Framed stdio protocol on a private descriptor**:
  - At startup, before initialising ONNX Runtime, `feov-embed` duplicates fd 1 to a private frame descriptor (`dup`), redirects fd 1 to fd 2 (`dup2`), and sets binary mode on stdin and the frame descriptor (`_setmode(…, _O_BINARY)` on Windows).
  - Any diagnostic output or library stdout write lands in stderr, never corrupting protocol frames (`TestFrameChannelSurvivesLibraryStdout`).
  - Handshake: the first frame exchange verifies protocol version (`v == 1`), helper source digest (`FEOV_HELPER_DIGEST`), and loaded model SHA256 (`model_sha256`, emitted as the second space-separated token after `FEOV_HELPER_DIGEST` in the `hello` response payload). Any request frame whose version is not `1` or that arrives before a completed handshake is refused.
- **Single embedding op**: `embed(ids, mask) -> float32[n][384]`. Token IDs (`int64[1, seq_len]`) and attention masks (`int64[1, seq_len]`) are fed as 2 input tensors (`input_ids` and `attention_mask`) to the ORT session one text at a time (no cross-text padding variation; the exported `.ort` model wrapper binds `token_type_ids` to zeros internally); mean pooling over masked tokens and L2 normalization ($\|\mathbf{v}\|_2 = 1$) execute inside `feov-embed` (`main.cpp`). Intra-op and inter-op threads are both pinned to `1`.
- **Request bounds and deadline**:
  - The Go client applies a 120 s overall deadline per request (`DefaultDeadline`).
  - To guarantee that every request completes in $< \frac{1}{4}$ of the deadline on the slowest target (`30 s` bar), the client splits inputs into batches of at most `embedBatch = 16` texts of at most 512 token IDs each. At the slowest target's measured speed (`5.87 chunks/s` on 200-word chunks), a 16-chunk batch completes in $\approx 2.7\text{ s}$.
- **Lifecycle & failure handling**:
  - On EOF on stdin, the helper exits 0.
  - On helper crash, deadline overrun, or protocol error during a request, `feov-record` terminates the child and fails that request loudly. The next request spawns a fresh helper.
  - A failure during startup (`helper-absent` or `helper-digest-mismatch`) latches on the client to prevent spawn loops.

### 2. Pure Go Client & WordPiece Tokenizer (`internal/vecembed`)

Because native C++ execution lives entirely inside `feov-embed`, `internal/vecembed` is written in **pure Go** and compiles with `CGO_ENABLED=0` across all platforms:
- **`vecembed.go`**: defines the `Embedder` interface (`Embed(texts []string) ([][]float32, error)`, `ModelID() string`, `Close() error`), batch splitting logic (`embedBatch = 16`), deterministic cosine similarity, and error handling. `ModelID()` returns `"gte-small-int8@" + nativehelper.HelperDigest("feov-embed") + "+target@" + runtime.GOOS + "-" + runtime.GOARCH` (per R9: because 8-bit quantized vectors differ slightly by architecture — mean fidelity `0.984411` on `amd64` vs `0.984464` on `arm64` — `ModelID()` embeds `os-arch` so `<text_sha>.vec` and `report_vecs.bin` cache headers include the target alongside the helper digest and chunk spec).
- **`tokenize.go`**: deterministic pure-Go BERT WordPiece tokenizer (control-char cleaning, CJK splitting, NFD accent stripping, lowercasing, punctuation splitting, WordPiece greedy longest-match-first against the embedded `thenlper/gte-small` `vocab.txt` of `30,522` tokens, `[CLS]`/`[SEP]` wrapping, and 512-token truncation) matching Hugging Face `AutoTokenizer.from_pretrained('thenlper/gte-small')` across **all 400/400 reference texts** (`300` in `speed.jsonl` + `100` in `len200.jsonl`, verified in `TestTokenizerParity400`).
- **Process management via `internal/nativehelper`**: uses `nativehelper.Client` ([`internal/nativehelper/client.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper/client.go)) to spawn `feov-embed`, verify the handshake (where `Client.handshake` checks `1 + len(c.pins.Assets)` fields in the `hello` response against `c.pins.Digest` and `a.SHA256[runtime.GOARCH]` from `digest_gen.go`), enforce deadlines, and exchange frames.
- **Distinct failure contracts**: produces explicit named errors (`helper-absent`, `helper-digest-mismatch`, `helper-crashed`, `helper-deadline`, `helper-protocol`), never silently degrading to BM25-only search.

### 3. Model Supply via `fetch-bin.sh`, `sc-doctor`, and `universe.sh`

Per `gblock`'s rulings (F1, F10, L2 in `plans/native-helpers.md`), 8-bit quantized models are distributed as platform release assets rather than embedded via `//go:embed`:
1. **Per-arch model assets (L2)**: `gte-small.int8.amd64.ort` (`sha256: f9919347c708c08fbdd93df377717bbf282a09ef21423d1a25f9c979bd335320`) for `amd64` targets and `gte-small.int8.arm.ort` (`sha256: 560718670ef2497d047d210a4e08c3bffce7c471a23714526db6b8571154edf7`) for `arm64` targets (`33,844,232` bytes each, identical operator set, different byte exports; `ortstatic/REPORT.md` §2 and `native/feov-embed/PINS.txt`).
2. **Release asset declaration**: `native/feov-embed/ASSETS` declares:
   ```text
   gte-small.int8.ort amd64=gte-small.int8.amd64.ort arm64=gte-small.int8.arm.ort
   ```
   and `native/feov-embed/PINS.txt` pins the SHA256 and URL of each export hosted on the `model-gte-small-v1` release (F10).
3. **Provisioning & self-healing**:
   - `scripts/fetchbingen/fetch-bin.sh` and `sc-doctor` install `gte-small.int8.ort` (fetching the platform's `gte-small.int8.{amd64,arm}.ort` release asset and verifying it against `SHA256SUMS`) into `${CLAUDE_PLUGIN_ROOT}/bin/gte-small.int8.ort` alongside the hook binaries and helpers (F1).
   - In `universe.sh build`, the models are staged into the universe's local release directory (F18, F19).
   - In `scripts/releaseassets/releaseassets.go`, `CheckDist` reads each helper's `ASSETS` exports (`gte-small.int8.amd64.ort`, `gte-small.int8.arm.ort`), verifies that both exports are present in `$DIST` on a FEOV tag, and skips `staticproof.Check` on declared `ASSETS` data files (since `.ort` model files are data assets rather than ELF/PE/Mach-O binaries); `IsNativeSetChange` (`releaseassets.go:234-250`) includes `plugins/frank-exchange-of-views/tools/internal/vecembed/`.
4. **Acceptance and refusal**:
   - At handshake, `feov-embed` reads `gte-small.int8.ort` beside its binary (or from `NATIVE_HELPER_DIR`), computes its SHA256, and reports it in the handshake response (`<helper_digest> <model_sha256>`). `nativehelper.Client.handshake` (`client.go:258-270`) verifies both `fields[0] == c.pins.Digest` (`nativehelper.HelperDigest("feov-embed")`) and `fields[1] == a.SHA256[runtime.GOARCH]` from `c.pins.Assets` in `digest_gen.go`.
   - When the helper or model is missing or mismatched, search refuses loudly by name (`helper-absent` or `helper-digest-mismatch`) and directs the operator to `/prosthetic-conscience:doctor --fix`.

### 4. Chunking, Source & Report Vector Caching, and Hybrid RRF Retrieval

- **Chunking (`internal/hybridsearch/chunk.go`)**:
  - **Sources (`<run>/cache/`, `SourceChunkSpec = "w200_s160"`)**: split extracted text (`SourceTextPath(run, entry, origin)` — `.txt` for PDF/HTML `embedded`, `.ocr.txt` for `ocr_derived`) into **200-word chunks with 40-word overlap** (`stride = 160` words), preserving byte offsets `[startByte, endByte)` and (for `ocr_derived` sources) the page number from `LocateSourceSpan`'s page boundary map so every returned chunk carries the exact coordinates that `blue cite` and `lens verify` consume.
    - **Best-matching sentence selection within a retrieved chunk (`bestChunkSentence` in `internal/fetchcache/sourcesearch.go`)**: because `blue cite --source-span` requires a single citeable sentence (`LocateSourceSpan` rejects multi-sentence spans or spans exceeding the sentence boundary rules), while a 200-word chunk spans multiple sentences, each hybrid `--search` source hit exposes **both**:
      1. `matched`, `start`, `end`, `line`, and `last_line` spanning the full retrieved 200-word chunk (`t.text[c.StartByte:c.EndByte]`), giving the seat the full passage context and exact `--source-lines <line>-<last_line>` coordinates; and
      2. `sentence`, `sentence_start`, `sentence_end`, `source_span`, and `not_citeable` for the **best-matching sentence inside `[c.StartByte, c.EndByte)`**. `bestChunkSentence` enumerates the candidate sentences overlapping `[c.StartByte, c.EndByte)` via `sentenceBounds`, scores each candidate sentence by non-stopword query token overlap (after lowercasing and Porter-equivalent prefix/stem normalization, breaking ties by earliest sentence in the chunk), resolves that sentence's actual OCR page (`locatePage(pages, sentStart)`) and `spanText` so a chunk crossing a page break never clamps a sentence on page $p+1$ to page $p$'s `spanText`, and runs `LocateSourceSpan` on that sentence to populate `source_span` and `not_citeable`.
  - **Report (`blue/report.md`, `ReportChunkSpec = "para_v1"`)**: split into paragraphs / logical blocks bounded by section headings (`heading`, line range `start_line..end_line`).
    - **Anchor comment stripping and multi-anchor extraction**: rendered FEOV reports contain inline HTML comment anchors (`<!--fx:F-...-->`, `<!--cite:C-...-->`, `<!--computation:K-...-->`, `<!--gap:G-...-->`). `ChunkReport` strips all HTML comment anchors (`claimcount.StripAnchors` / `visibleText`) before indexing paragraphs in FTS5 (`porter unicode61`) and before embedding with `feov-embed` so comment markup (`fx`, `cite`, `computation`, `gap`, hex IDs) never pollutes BM25 tokens or dense vectors. `ChunkReport` extracts **all** anchor IDs (`F-`, `C-`, `K-`, `G-`) present on the paragraph into `anchors: []string` (omitted when empty) and sets `anchor` to the first anchor ID on the paragraph for single-anchor convenience.
- **Content-addressed vector cache for sources (`<run>/cache/<text_sha>.vec`)**:
  - Addressed by `text_sha = sha256(extracted_text_bytes)` (where `extracted_text_bytes` are the exact bytes read from `SourceTextPath(run, entry, origin)`).
  - **Binary header includes `modelID` (with target `os-arch`) and `chunkSpec`**: the `.vec` file begins with magic `FEOVVEC2`, `uint16(len(modelID))`, `modelID` (`"gte-small-int8@<helper_digest>+target@<os>-<arch>"`), `uint16(len(chunkSpec))`, `chunkSpec` (`"w200_s160"`), and `uint32(numChunks)`, followed by each chunk's `[startByte, endByte, page]` and `float32[384]` vector. `ReadVecCache(path, wantModelID, wantChunkSpec)` refuses any file whose magic, `modelID` (helper digest or `os-arch` target), or `chunkSpec` disagrees, returning a cache miss so stale offsets from a chunker change or vectors from a different model/helper/architecture are never read.
  - **Why `text_sha` and NOT `entry.Sha` (raw body SHA)**: when `ocr read` or `render-page` re-derives `<sha>.ocr.txt`, `text_sha` changes automatically, making stale vectors impossible without invalidation hooks (`pagerender.go:342`, `readscanned.go:148`).
  - **Pre-warming (`prewarmSourceVectors`)**: called at the end of `foldExtraction` and after `ocr read` / `readScanned` writes `<sha>.ocr.txt`. Reads `SourceTextPath(run, entry, origin)` so `prewarm`'s `text_sha` is always identical to `fullSourceText`'s `text_sha`. Computed lazily on `show source --search` if absent (e.g., on an archived run).
- **Content-addressed paragraph vector cache for the report (`<run>/cache/report_vecs.bin`)**:
  - **Why report caching is required**: at `5.87–7.12 chunks/s` single-threaded, re-embedding all 50–300 paragraphs of `blue/report.md` from scratch on every `show report --search` call would take **8–50 s**. During a debate, `blue edit` modifies only 1–3 paragraphs per sitting while the remaining 95%+ of paragraphs are unchanged.
  - **Cache structure**: `<run>/cache/report_vecs.bin` stores a header (`FEOVRPT1`, `modelID` including `+target@<os>-<arch>`, `ReportChunkSpec = "para_v1"`, `uint32(count)`) followed by entries mapping `para_sha = sha256(visibleText(para))` (`[32]byte`) to `float32[384]`.
  - **Query-time execution**: on `show report --search <query>`, `feov-record` chunks `blue/report.md` into $N$ paragraphs (`ChunkReport`), loads `<run>/cache/report_vecs.bin` (if `modelID` and `ReportChunkSpec` match), collects the $M \le N$ paragraphs whose `para_sha` (`sha256(visibleText)`) is not yet in the map, calls `embedder.Embed` on `[query] + uncached_paragraphs` ($1 + M$ texts), writes the updated map back to `<run>/cache/report_vecs.bin` atomically via `writeReplacing`, and runs hybrid BM25 + cosine RRF over all $N$ paragraphs (`SearchWithVectors`). On a warm report (or after a 1-paragraph edit), only 1–2 texts are embedded ($\approx 50\text{–}300\text{ ms}$).
- **Hybrid Fusion (`BM25` + Cosine via RRF, `internal/hybridsearch/rrf.go`)**:
  - **Lexical arm (`rankBM25`)**: ephemeral in-memory SQLite (`:memory:`) `USING fts5(text, tokenize='porter unicode61')`. Query terms are tokenized into alphanumeric tokens joined by `OR` (`"tok1" OR "tok2"`).
    - **Error handling**: if the query contains 0 alphanumeric tokens, `rankBM25` returns `(nil, nil)` (empty lexical ranks). If SQLite or FTS5 returns any error when opening `:memory:`, creating the FTS5 table, inserting chunks, or executing the `MATCH` query, `rankBM25` returns that error and `Search` fails loudly — never silently turning an FTS5 error into vector-only results labelled as hybrid.
  - **Vector arm**: exact inner product (cosine similarity on L2-normalized `[]float32` vectors) over all chunks in scope.
  - **Fusion**: Reciprocal Rank Fusion with standard constant $k = 60$:
    $$\text{score}(d) = \frac{1}{60 + \text{rank}_{\text{bm25}}(d)} + \frac{1}{60 + \text{rank}_{\text{vec}}(d)}$$
    where a chunk absent from the BM25 match set contributes only its vector term $\frac{1}{60 + \text{rank}_{\text{vec}}(d)}$.

---

### 5. Explicit Response Schemas for Report and Source Search

To prevent ambiguous parsing by seats and downstream tools, each view and scoping mode has a strictly defined root JSON schema. In both `SourceHit` and `CrossSourceHit`, `matched` / `start` / `end` / `line` / `last_line` describe the matched span (the full 200-word chunk for `--search`, or the exact regex/literal match for `--match`/`--quote`), while `sentence` / `sentence_start` / `sentence_end` / `source_span` / `not_citeable` describe the best-matching citeable sentence within that span (and `line` / `last_line` give the 1-based line range for `blue cite --source-lines <line>-<last_line>` when `not_citeable` is non-empty):

#### A. Cross-Source Ranked Hybrid Search (`show source --search <query>`)
Triggered when neither `--sha` nor `--anchor` is passed. Searches across all cached sources in `<run>/cache/`.
Every hit in `hits` carries its source identity (`sha256` and `url`), byte and line coordinates, and citeable `source_span` so it is immediately citeable via `blue cite --url <url> --source-span "<source_span>"` (or `--source-lines <line>-<last_line>`):
```json
{
  "hits": [
    {
      "sha256": "4a7b9c1d...",
      "url": "https://example.org/article.html",
      "source_text_origin": "embedded",
      "text_sha": "e3b0c442...",
      "page": 0,
      "score": 0.0312,
      "matched": "full 200-word retrieved chunk text ...",
      "line": 12,
      "last_line": 28,
      "sentence": "the best-matching sentence inside the chunk",
      "sentence_start": 105,
      "sentence_end": 182,
      "source_span": "the best-matching sentence inside the chunk",
      "not_citeable": "",
      "start": 80,
      "end": 1240
    }
  ],
  "search_scope": {
    "ranked_and_cut": true,
    "returned": 10,
    "searched_chunks": 184,
    "searched_sources": 6,
    "unsearched_sources": [
      {
        "sha256": "f1d2...",
        "url": "https://example.org/scan.pdf",
        "reason": "image_only: no OCR reading yet; run fetch --url <url>"
      }
    ],
    "notice": "Ranked top-K hybrid (BM25 + vector) results; this list is cut and absence proves nothing. For an exhaustive literal match set, use --quote (or --match for a regex)."
  }
}
```

#### B. Single-Source Ranked Hybrid Search (`show source --sha <sha|--anchor C-…> --search <query>`)
Scoped to one cached source. Root retains single-source metadata (`sha256`, `url`, `source_text_origin`, `text_path`):
```json
{
  "sha256": "4a7b9c1d...",
  "url": "https://example.org/article.html",
  "source_text_origin": "embedded",
  "text_path": "/path/to/cache/4a7b9c1d....txt",
  "hits": [
    {
      "page": 0,
      "score": 0.0312,
      "text_sha": "e3b0c442...",
      "matched": "full 200-word retrieved chunk text ...",
      "line": 12,
      "last_line": 28,
      "sentence": "the best-matching sentence inside the chunk",
      "sentence_start": 105,
      "sentence_end": 182,
      "source_span": "the best-matching sentence inside the chunk",
      "not_citeable": "",
      "start": 80,
      "end": 1240
    }
  ],
  "search_scope": {
    "ranked_and_cut": true,
    "returned": 5,
    "searched_chunks": 32,
    "notice": "Ranked top-K hybrid (BM25 + vector) results; this list is cut and absence proves nothing. For an exhaustive literal match set, use --quote (or --match for a regex)."
  }
}
```

#### C. Cross-Source Exhaustive Literal or Regex Search (`show source --quote <text>` or `show source --match <regex>`)
Triggered when neither `--sha` nor `--anchor` is passed. Exhaustive literal substring (`--quote`) or RE2 regex (`--match`) search across all cached sources in `<run>/cache/`. `selection.criterion` uses the same format as single-source `SearchSource` (`"the quote \"...\""` for `--quote`, `"the pattern /.../"` for `--match`):
```json
{
  "hits": [
    {
      "sha256": "4a7b9c1d...",
      "url": "https://example.org/article.html",
      "source_text_origin": "embedded",
      "text_sha": "e3b0c442...",
      "page": 0,
      "matched": "exact matched string",
      "line": 15,
      "last_line": 15,
      "sentence": "the full sentence surrounding the match",
      "sentence_start": 105,
      "sentence_end": 182,
      "source_span": "the full sentence surrounding the match",
      "not_citeable": "",
      "start": 120,
      "end": 140
    }
  ],
  "selection": {
    "criterion": "the quote \"exact matched string\"",
    "matched": 14,
    "searched_sources": 6,
    "unsearched_sources": [
      {
        "sha256": "f1d2...",
        "url": "https://example.org/scan.pdf",
        "reason": "image_only: no OCR reading yet; run fetch --url <url>"
      }
    ]
  }
}
```

#### D. Single-Source Exhaustive Literal or Regex Search (`show source --sha <sha|--anchor C-…> --quote <text>|--match <regex>`)
Existing `SourceSearch` shape ([`internal/fetchcache/sourcesearch.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/sourcesearch.go)), unchanged. Returns every match with `selection.matched = N`.

#### E. Report Ranked Hybrid Search (`show report --search <query>`)
Searches the report document (`blue/report.md`). Paragraph chunks with HTML anchor comments stripped from `text`, `anchor` holding the first anchor ID on the paragraph, and `anchors` listing all anchor IDs (`F-`, `C-`, `K-`, `G-`) on the paragraph. Has NO `searched_sources` or `unsearched_sources`:
```json
{
  "chunks": [
    {
      "heading": "## Research areas",
      "start_line": 42,
      "end_line": 56,
      "text": "The verified prime test ...",
      "score": 0.0325,
      "anchor": "C-abc12345",
      "anchors": ["C-abc12345", "F-def67890"]
    }
  ],
  "search_scope": {
    "ranked_and_cut": true,
    "returned": 10,
    "searched_chunks": 48,
    "notice": "Ranked top-K hybrid (BM25 + vector) results; this list is cut and absence proves nothing. For an exhaustive literal match set, use --quote (or --match for a regex)."
  }
}
```

---

## III. Proposed Changes (the spec)

### 1. Empirical Audit of `--match` Across All Views & Selector Vocabulary Design

Per `gblock`'s 2026-11-10 ruling (*"If `--search` replaces the use case for match, remove match. But audit it for that — it might be that we want all three universally"*), we audited all 2,252 seat transcripts in `~/.claude/scratch/vecplan/replay/` (`searches.jsonl`, `acts.jsonl` parsed with `replay/scripts/parse.py`, `SUMMARY.md`, and `dropped.jsonl`).

#### A. October Call Volume Across All Views (`searches.jsonl` & `acts.jsonl`)

| Surface / View | `--match` statements | `--quote` statements | `--anchor` / `--id` | Bare (no selector) | Notes (`SUMMARY.md` & `acts.jsonl`) |
|---|---:|---:|---:|---:|---|
| `show report` | **1,094** (`239` zero, `649` hit, `6` err, `200` unattr.) | **400** (`106` zero, `205` hit, `89` unattr.) | `52` (`--anchor`) | `3,508` (plus `215` `\| grep`) | 50 of the 79 verified wording-miss rows in `replay.jsonl` |
| `show source` | **121** (`5` zero, `83` hit, `33` unattr.) | **12** (`0` zero, `3` hit, `1` err, `8` unattr.) | — | `0` (plus `2,452` `grep/rg cache`) | 29 of the 79 verified wording-miss rows in `replay.jsonl` |
| `show board` | **117** | **19** | `3` (refused `--id`) | `777` | Structured JSON array of open/closed gaps |
| `show findings` | **106** | **15** | — | `394` | Structured JSON array of lens findings |
| `show inquiries` | **75** | **2** | — | `449` | Structured JSON array of blue inquiries |
| `show work` | **20** | **1** | `3` (refused `--id`) | `854` | Structured JSON of seat's owed work |
| `show evidence` | **17** | **8** | `657` (`--anchor`) | `653` | Structured JSON of citations & computations |
| `show changes` | **11** | **15** | `160` (`--id`) | `109` | Structured JSON of report edits & retires |
| `show avenues` | **8** | **1** | — | `170` | Structured JSON of planned research lanes |

*(Note: an earlier version of this plan claimed only 4 `--match` calls on `board` and 7 on `changes` because a naive regex split on `|` inside `--match "a|b"`, losing every alternation call — the exact parser trap documented in `m11-selector-adoption.md`.)*

#### B. Classification of Real `--match` Patterns on `report` (1,094) and `source` (121)

76.0% of `show report --match` calls (`831 / 1,094`) and 68.6% of `show source --match` calls (`83 / 121`) use regex syntax (`812` and `81` alternations `a|b`, `80` and `16` character classes/boundaries `[a-z]+` / `\b`, `73` and `14` quantifiers `?`/`+`, `19` and `25` escaped metacharacters `Z\[1/7\]` / `3\^90`, `9` wildcard spans `.*`, `5` line anchors `^`/`$`). Grouping every call by **functional intent**:

| Functional Category of `--match` Call | `report` ($N=1{,}094$) | `source` ($N=121$) | Representative Real Seat Patterns (`searches.jsonl`) | Does `--search` Replace It? |
|---|---:|---:|---|---|
| **1. Synonym / morphological / paraphrase alternation** (workaround for lack of stemming & semantic search) | **527** (`48.2%`) | **35** (`28.9%`) | `single-base\|one-base\|single base\|one base` (`blue-researcher`); `looks prime\|Conway\|Rotten\|video\|tag` (`red-lens-evidence`) | **YES — `--search` is strictly better.** Stemmed BM25 (`porter`) + dense `gte-small.int8` RRF replaces hand-guessed synonym lists and rescues the 55 wording misses in `replay.jsonl` (`Hit@10 = 0.924`). |
| **2. Plain literal word, phrase, or number** (no regex metacharacters) | **263** (`24.0%`) | **38** (`31.4%`) | `clear statement`, `closes the gap for this number` (`45` zero hits on `report`); `congruence`, `relatively prime to the number` | **YES — covered by `--quote` (exact literal check) and `--search` (when phrasing differs).** |
| **3. Voice / tell / banned-vocabulary exhaustive sweeps** (`red-lens-voice`, `red-lens-dark-side`, `red-lens-architecture`) | **174** (`15.9%`) | **2** (`1.7%`) | `(recorded script\|script\|sources? reached\|reached\|snippet\|searches\|searched\|Python\|program\|network\|[Rr]erun\|Residual risk\|were checked\|was checked\|not examined\|read only\|not established)` (`zero` is wanted) | **NO — requires exhaustive `--match`.** `0` matches is the proof of absence (`80` rows in `SUMMARY.md:44`), and when a tell occurs, the lens needs **every** matching line across the report. Top-$K$ `--search` always returns $K$ chunks even when 0 tells exist; `--quote` would turn 1 call into 16 separate calls. |
| **4. Multi-identifier / formula / symbol batch alternation** | **111** (`10.1%`) | **44** (`36.4%`) | `Trial division\|Base 3 passes\|P-e6803dfe\|P-c1c0989f\|P-47ce38d8`; `7 × 13\|3\^45\|3\^90`; `3 ?n\|3\^n\|congruence\|pseudoprime` | **NO — requires exhaustive `--match`.** Batches multiple exact anchor IDs, OEIS numbers (`A005935`), or formula notations (`3 ?n\|3\^n`) into one exhaustive read. |
| **5. Structural / word-boundary / character-class regex** | **19** (`1.7%`) | **2** (`1.7%`) | `at least [a-z]+ times` (`red-lens-computation`); `^###.*Summary`; `\b91\b` (`red-lens-evidence`, avoiding `191`/`910`) | **NO — requires `--match`.** Neither literal `--quote` nor ranked `--search` expresses `\b91\b`, `[a-z]+`, or `^###`. |

#### C. Classification of Real `--match` Patterns on Structured Record Views (354 October Calls)

| View | Total `--match` | 1. Multi-ID Alternation (`G-a\|G-b`) | 2. Single ID Lookup (`G-…`) | 3. Multi-Phrase Alternation (`a\|b\|c`) | 4. Regex Metachars (`\b`, `[...]`) | 5. Plain Word/Phrase Literal | Representative Real Patterns (`acts.jsonl`) |
|---|---:|---:|---:|---:|---:|---:|---|
| `board` | **117** | `25` | `24` | `35` | `0` | `33` | `G-bee35a71\|G-0e2b65e9\|G-1189288b`; `two clear statements\|one clear statement` |
| `findings` | **106** | `23` | `8` | `58` | `0` | `17` | `F-3a4a5da0\|F-f22d7716\|F-c1b1e3b9`; `six\|count of the cached\|at least six` |
| `inquiries` | **75** | `3` | `5` | `25` | `1` | `41` | `Q-5f242019\|Q-9bd9eeb9\|Q-31b00376\|Q-4122976f`; `under (Research areas\|Alternatives considered)` |
| `work` | **20** | `4` | `7` | `1` | `0` | `8` | `G-bee35a71\|F-ba76bc46\|F-117a275b`; `Q-\|G-` |
| `evidence` | **17** | `2` | `1` | `4` | `0` | `10` | `P-8933b92c\|P-3dbb0fb4\|P-b45ae2e5\|P-32f6a49d` |
| `changes` | **11** | `1` | `0` | `5` | `0` | `5` | `fx:F-cd03bd17\|open question\|Open questions`; `pass is not proof\|prove nothing\|proves nothing` |
| `avenues` | **8** | `0` | `2` | `3` | `0` | `3` | `trial division\|Fermat\|definition\|reference\|prime` |
| **Total** | **354** | **58** (`16.4%`) | **47** (`13.3%`) | **131** (`37.0%`) | **1** (`0.3%`) | **117** (`33.1%`) | — |

#### D. Audit Verdict & Evaluated Design Fork on Selector Scope

1. **`--match` MUST be kept alongside `--quote` on every view**:
   - Removing `--match` is refuted by the transcript record: **304** October calls on `report` (`27.8%`), **46** on `source` (`38.0%`), and **58** multi-ID batch calls (`16.4%`) plus **131** multi-phrase alternations (`37.0%`) across `board`, `findings`, `inquiries`, `work`, `evidence`, `changes`, and `avenues` rely on exhaustive regex alternation, word boundaries (`\b91\b`), or character classes (`[a-z]+`) where ranked top-$K$ `--search` is a category error (because `0` matches is the proof of absence or every matching entry must be returned uncut) and where `--quote` would require 3–16 separate invocations per check.
   - Keeping `--match` also honors `gblock`'s 2026-10-18 ruling in `regexes-are-questioned.md`: *"User-supplied patterns (`--match`) are fine."*
2. **Where `--search` is registered — Evaluated Design Fork for `gblock`**:
   - **Option S-Prose (Specified in this plan): Keep `--quote` and `--match` on every view (`AddSelectorFlags`); add `--search` on `show report` and `show source` (and cross-source `show source`).**
     - *Evidence for Option S-Prose*:
       1. `report` and `source` are long unstructured prose corpora (hundreds of chunks across multiple external documents) and account for **100% of the 79 verified wording-miss failures** in `replay.jsonl` (`50` on `report`, `29` on `source`).
       2. Structured record views (`board`, `findings`, `inquiries`, `work`, `evidence`, `changes`, `avenues`) are small JSON arrays of record events (typically 5–60 entries per run, median ~15). Across all October single-statement `--match` calls on those structured views (`acts_full.jsonl`), **0 calls failed due to wording mismatch** (the only 2 zero-hit calls on `board` queried a non-existent gap ID `G-c98b23ab`).
       3. Preserves the crisp architectural boundary in [`internal/cli/seat/selector.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat/selector.go#L46): `AddSelectorFlags` (`--match` and `--quote`) are **exhaustive boolean selectors** (`complete: true`, never truncated, 0 matches on non-occurring strings as tested by `TestEverySelectorThatAViewAdvertisesSelects`), whereas `--search` on `report` and `source` is a **ranked, cut top-$K$ retrieval mode** (`ranked_and_cut: true`, `search_scope`).
   - **Option S-Universal (Alternative for `gblock`'s ruling): Register all three (`--quote`, `--match`, `--search`) universally on every selector-bearing view (`AddSelectorFlags` registers all three).**
     - *How Option S-Universal works if selected*: `WriteSelected` in `selector.go` gains a `searchJSONArrays(run, body, query, topK)` branch that serializes each top-level JSON array item, embeds uncached items via `feov-embed` (cached by `sha256(item_bytes)` in `<run>/cache/view_vecs.bin`), ranks items with hybrid BM25 (`porter unicode61`) + cosine RRF ($k=60$), keeps top-$K=10$ per array, and attaches `"search_scope": {"ranked_and_cut": true, "returned": kept, "searched_entries": of, "notice": "..."}`.
     - *Tradeoff against Option S-Universal*: On a 15-entry `board` or `findings` array, cutting to top-$K=10$ risks hiding open gaps or findings a seat needs to account for while adding a `feov-embed` IPC round-trip to a metadata projection that renders in $< 5\text{ ms}$ whole.

#### E. Surface & Mutual Exclusion Rules (under Option S-Prose)
- **`--quote` and `--match` stay unchanged everywhere**: on every `show` and `inquest` view (except `telemetry`), and on `show source`, `--quote` selects every literal substring match and `--match` selects every RE2 regex match with zero truncation.
- **`show source` scoping flags (`--sha` and `--anchor`)**:
  - When `--sha` or `--anchor` is passed: scopes `--search`, `--quote`, or `--match` to that single source.
  - When neither `--sha` nor `--anchor` is passed: executes **cross-source search** (`--search`, `--quote`, or `--match`) across all cached sources in `<run>/cache/`.
  - When neither source scoping nor pattern/query flags (`--search`, `--quote`, `--match`) are passed: refuses with an explicit usage message: `"name what to find: --search (ranked hybrid), --quote (exhaustive literal), or --match (exhaustive regex), and optionally scope to one source with --sha or --anchor"`.
- **Mutual Exclusion Contract**:
  - On `show report` and `show source`, `--search`, `--quote`, and `--match` are **strictly mutually exclusive** (at most one may be passed).
  - Passing `--search` with `--quote` or `--match` returns `feov.Validation`:
    `"--search cannot be combined with --quote or --match: --quote and --match return every matching entry; --search returns a cut ranked list. Pass one"`
  - On `show report`: `--search`, `--quote`, and `--match` are each refused beside `--anchor`, and `--search` is refused beside `--assembled`.
- **Protocol State & Pin Protection (`ServeReport`)**:
  - In [`internal/cli/seat/verbs.go:978-985`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat/verbs.go#L978-L985):
    ```go
    serve := role == "lens" && Str(cmd, flags.Anchor) == "" && !cmd.Flags().Changed(flags.Window)
    if serve {
        sel, serr := SelectorOf(cmd)
        if serr != nil {
            return serr
        }
        serve = !sel.Active() && Str(cmd, flags.Search) == ""
    }
    ```
  - A lens invoking `show report --search` reads a cut ranked list of chunks, not the whole report; `serve = false` guarantees `record.ServeReport` is NOT logged and the lens's reading pin does NOT advance.
- **Terms Registry Update**:
  - Update [`internal/terms/terms.json:568`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/terms/terms.json#L568) definition of `"source search"` to:
    `"A source search is a read of cached sources' text for a regular expression, literal text, or a ranked hybrid query (BM25 and dense vectors), scoped to one source or across all cached sources in the run; each hit carries the source's sentence, its URL and SHA256, and whether a citation can quote that sentence as printed, and the search records nothing."`
  - Re-run `go -C plugins/frank-exchange-of-views/tools run ./devcmd/vocabgen` and `go -C scripts run ./vocabdoc` (which updates [`plugins/frank-exchange-of-views/docs/vocabulary.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/docs/vocabulary.md)).

---

### 2. Component Breakdown & File Inventory

*(Note on PR #1791 branch state: an earlier draft on branch `feat/vector-search-feov-embed` removed `flags.Match` from `internal/flags/names.go`, `internal/cli/seat/selector.go`, `internal/cli/seat/source.go`, `internal/cli/seat/verbs.go`, and `internal/cli/selectorcensus_test.go`; on that branch, `--match` is restored to match `origin/main` alongside `--quote` and `--search`.)*

- **`plugins/frank-exchange-of-views/native/feov-embed/`** **[NEW]** (Part 3 of `plans/native-helpers.md`)
  - `main.cpp` **[NEW]**: thin C++ kernel hosting ONNX Runtime 1.31.1 minimal Release build, private descriptor frame protocol (`dup`/`dup2`), protocol version (`v == 1`) and handshake-first enforcement (reporting `<helper_digest> <model_sha256>`), 1-thread intra/inter-op execution, 2-tensor (`input_ids`, `attention_mask`) per-text session run, mean pooling over attention mask, and L2 normalization.
  - `ops.config` **[NEW]**: reduced operator configuration for `gte-small` int8 (`ortstatic/REPORT.md` §2).
  - `PINS.txt` **[NEW]**: ONNX Runtime 1.31.1 archive pin, pure-Python `flatbuffers-25.2.10-py2.py3-none-any.whl` pin (for `reduce_op_kernels.py`), and per-arch model export pins (`gte-small.int8.amd64.ort` `f9919347...` and `gte-small.int8.arm.ort` `56071867...` on release `model-gte-small-v1`).
  - `ASSETS` **[NEW]**: declares `gte-small.int8.ort amd64=gte-small.int8.amd64.ort arm64=gte-small.int8.arm.ort`.
- **`plugins/frank-exchange-of-views/native/build.sh`** **[MODIFY]**
  - Builds `feov-embed` across all 5 targets (`linux-{amd64,arm64}`, `windows-{amd64,arm64}`, `darwin-arm64`) using `zig c++` (Linux/macOS, with the `aarch64` `-march` rewrite wrapper and `darwin-arm64` CoreFoundation/Foundation `.tbd` and header stubs detailed in §II.1) and MSVC `/MT` (Windows), embedding compiled-in source digest via `-DFEOV_HELPER_DIGEST=…`.
  - Uses **only** pinned inputs (`native/PINS.txt` and `native/feov-embed/PINS.txt`); no `$ORT_STATIC_DIR`, no `~/.claude/scratch/...`, no unpinned `pip install`.
  - Maps `ARCH=amd64 -> MODEL_ARCH=amd64` (`gte-small.int8.amd64.ort`) and `ARCH=arm64 -> MODEL_ARCH=arm` (`gte-small.int8.arm.ort`) when fetching from `PINS.txt`, and stages `${HELPER}_${OS}_${ARCH}${EXT}`, `${HELPER}${EXT}`, `gte-small.int8.${MODEL_ARCH}.ort`, and `gte-small.int8.ort` into `$OUTDIR`.
- **`plugins/frank-exchange-of-views/tools/internal/nativehelper/`** **[MODIFY]**
  - `nativesrc/nativesrc.go`, `nativesrc/nativesrc_test.go`, `client.go`, `digest_gen.go`: `devcmd/nativedigest` generates `feov-embed`'s source digest and `AssetPin` (`gte-small.int8.ort` with `amd64` and `arm64` SHA256s) in `digest_gen.go`; `Client.handshake` verifies both `c.pins.Digest` and `a.SHA256[runtime.GOARCH]`.
- **`plugins/frank-exchange-of-views/tools/internal/vecembed/`** **[NEW]**
  - `vecembed.go` **[NEW]**: pure-Go `Embedder` interface, `Default() (Embedder, error)`, batching logic (`embedBatch = 16`), L2 normalization, `Cosine(a, b []float32) float32`, `ModelID()` returning `"gte-small-int8@" + nativehelper.HelperDigest("feov-embed") + "+target@" + runtime.GOOS + "-" + runtime.GOARCH`. Compiles with `CGO_ENABLED=0`.
  - `tokenize.go` **[NEW]**: pure-Go BERT WordPiece tokenizer matching `thenlper/gte-small` vocabulary (`30,522` tokens) across all 400/400 reference texts (`speed.jsonl` + `len200.jsonl`).
  - `vecembed_test.go` **[NEW]**:
    - `TestTokenizerParity400`: verifies exact token ID parity against all 400 reference texts from Hugging Face `AutoTokenizer.from_pretrained('thenlper/gte-small')`.
    - `TestEmbedSplitsAtTheBatch`: uses a fake helper that records every incoming `embed` request frame's batch length, verifying that 35 texts arrive as `[16, 16, 3]` (never $> 16$), and verified by deleting the batch split in `vecembed.go` to confirm the test fails.
    - `TestEveryOpFitsTheDeadline`: runs a full 16-chunk batch of 200-word chunks against the built helper and asserts elapsed wall clock is $< 30\text{ s}$ ($\frac{1}{4}$ of the 120 s deadline).
    - `TestEmbedFidelity` (S5): embeds the 300 reference texts (`speed.jsonl`) with the built helper and asserts cosine similarity against `ref_gte.npy` has **mean $\ge 0.984$ and min $\ge 0.966$**, logging the exact mean and min in the test output.
    - `TestEmbedSpeed` (S6): embeds the 100 200-word chunks (`len200.jsonl`) single-threaded with the built helper and asserts throughput $\ge 5.0\text{ chunks/s}$ on `linux-amd64`, `linux-arm64`, and `darwin-arm64` (and cold 1-query latency $\le 1.5\text{ s}$), logging the exact chunks/s and cold latency in the test output on all 5 targets (exiting `PASS` rather than `t.Skip` on Windows so `NATIVE_RUN_MUST_PASS` sees `--- PASS: TestEmbedSpeed`).
    - `TestMissingOrMismatchedHelperRefusesByName`: verifies `helper-absent` and `helper-digest-mismatch` refusals.
- **`plugins/frank-exchange-of-views/tools/internal/hybridsearch/`** **[NEW]**
  - `chunk.go` **[NEW]**: deterministic 200-word / 40-word-overlap (`stride = 160`, `SourceChunkSpec = "w200_s160"`) source chunker preserving byte spans and page numbers, plus report paragraph chunker (`ReportChunkSpec = "para_v1"`) that strips HTML comment anchors (`claimcount.StripAnchors`) from `Text` before BM25/embedding and extracts all anchor IDs (`F-`, `C-`, `K-`, `G-`) into `Anchors []string` and `Anchor string`.
  - `cache.go` **[NEW]**: binary read/write of source vector cache `<run>/cache/<text_sha>.vec` (`FEOVVEC2`, checking both `modelID` and `SourceChunkSpec`) and report paragraph vector cache `<run>/cache/report_vecs.bin` (`FEOVRPT1`, checking both `modelID` and `ReportChunkSpec`, keyed by `sha256(visibleText(para))`).
  - `rrf.go` **[NEW]**: in-memory SQLite FTS5 (`porter unicode61`) BM25 (propagating any SQLite/FTS5 error instead of returning silent empty lexical ranks) + dense cosine ranking + RRF ($k=60$) fusion + structured `SearchResult` with `SearchScope` metadata.
  - `hybridsearch_test.go` **[NEW]**: unit tests for chunking, HTML anchor comment stripping and multi-anchor (`F-`, `C-`, `K-`, `G-`) extraction, `.vec` and `report_vecs.bin` cache invalidation on `modelID` or `chunkSpec` change, FTS5 error propagation, and RRF fusion invariants.
- **`plugins/frank-exchange-of-views/tools/internal/fetchcache/`** **[MODIFY]**
  - [`sourcesearch.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/sourcesearch.go) **[MODIFY]**:
    - Retain `SearchSource(run, e, match, quote)` (single-source exhaustive regex and literal search).
    - Add `SearchCrossSources(run, match, quote)`: cross-source exhaustive regex and literal search returning `CrossSourceSearch` (where `CrossSourceHit` includes `Line` and `LastLine` matching `SourceHit`).
    - Add `SearchSourceHybrid(run, e, query, topK)` and `SearchCrossSourcesHybrid(run, query, topK)`: single-source and cross-source ranked hybrid search returning `SourceSearch` or `CrossSourceSearch` with `SearchScope`, where `matched`/`start`/`end`/`line`/`last_line` span the 200-word chunk and `bestChunkSentence` selects the best-matching sentence inside the chunk (resolving its actual OCR page via `locatePage(pages, sentStart)`) for `sentence`, `sentence_start`, `sentence_end`, `source_span`, and `not_citeable`.
    - Add `prewarmSourceVectors(run, entry, origin)` reading `SourceTextPath(run, entry, origin)` so `prewarm`'s `text_sha` always matches `fullSourceText`'s `text_sha`.
  - [`fetchcache.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/fetchcache.go), [`pageread.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/pageread.go), [`readscanned.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/readscanned.go) **[MODIFY]**: pre-warm `<text_sha>.vec` when an embedder is available.
- **`plugins/frank-exchange-of-views/tools/internal/flags/` & `internal/cli/`** **[MODIFY]**
  - [`internal/flags/names.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags/names.go) **[MODIFY]**: add `Search = "search"` while keeping `Match = "match"` and `Quote = "quote"`; update `flags.All()`.
  - [`internal/cli/seat/selector.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat/selector.go) **[MODIFY]**: keep `AddSelectorFlags` (`--match` and `--quote`) unchanged across all views; add ranked hybrid `--search` handler for `show report` (`WriteReportSearch(run, out, reportText, query, topK)`) using `<run>/cache/report_vecs.bin` so unchanged paragraphs are not re-embedded on each call.
  - [`internal/cli/seat/source.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat/source.go) **[MODIFY]**: register `--search` alongside `--match` and `--quote`; allow omitting `--sha`/`--anchor` when `--search`, `--quote`, or `--match` is given (cross-source search over all cached sources); refuse bare `show source` with no flags; enforce mutual exclusion among `--search`, `--quote`, and `--match`.
  - [`internal/cli/seat/verbs.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat/verbs.go), [`internal/cli/seat/help/report.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat/help/report.md), [`internal/cli/seat/help/source.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat/help/source.md), and generated [`internal/cli/seat/help_gen.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat/help_gen.go) **[MODIFY]**: register `--search` on `show report`; update `report.md` and `source.md` help sources and `reportSelectorBound` for `--search`, `--quote`, and `--match` (regenerating `help_gen.go` via `cmdtreegen`); pass `run` to `WriteReportSearch`; suppress `ServeReport` (`serve = false`) when `flags.Search` is given.
  - [`internal/cli/selectorcensus_test.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/selectorcensus_test.go) **[MODIFY]**: keep `TestEverySelectorThatAViewAdvertisesSelects` testing both `--match` and `--quote` across every view in `selectorViews(t)`, and add `TestRankedSearchStatesCutAndScope` for `show report --search` and `show source --search`.
- **`plugins/frank-exchange-of-views/tools/internal/terms/` & `docs/`** **[MODIFY]**
  - [`terms.json`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/terms/terms.json) **[MODIFY]** and generated [`terms/definitions_gen.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/terms/definitions_gen.go) & [`plugins/frank-exchange-of-views/docs/vocabulary.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/docs/vocabulary.md) **[MODIFY — generated]**: update definition of `"source search"`; run `vocabgen` and `vocabdoc`.
- **Plans & Release, Provisioning, & Agent Surfaces** **[MODIFY]**
  - [`plans/README.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plans/README.md) (`:24`) **[MODIFY]**: update `vector-search.md` summary row so both `--quote` and `--match` are stated as preserved (`plans/native-helpers.md` is left untouched to avoid colliding with `39002b05`'s `native/phase-b-fixes` PR).
  - [`scripts/releaseassets/releaseassets.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/releaseassets/releaseassets.go) & [`releaseassets_test.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/releaseassets/releaseassets_test.go) **[MODIFY]**: verify declared `ASSETS` exports (`gte-small.int8.amd64.ort`, `gte-small.int8.arm.ort`) in `CheckDist` on FEOV tags while skipping `staticproof.Check` on data assets, and include `plugins/frank-exchange-of-views/tools/internal/vecembed/` in `IsNativeSetChange`.
  - [`.github/workflows/hooks.yml`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/.github/workflows/hooks.yml) **[MODIFY]**:
    - Include `feov-embed` in `native-build` matrix and `native-run` (`go test` includes `./internal/vecembed/...`, `NATIVE_RUN_MUST_PASS` includes `TestTokenizerParity400`, `TestEmbedFidelity`, `TestEmbedSpeed`, and `TestEveryOpFitsTheDeadline`, and the Windows `servercore:ltsc2025` container step prebuilds and runs `vecembed.test.exe` alongside `tessocr.test.exe`, `fetchcache.test.exe`, and `nativehelper.test.exe`, failing the step with `exit 1` instead of `|| true` if any required test fails).
    - In `feov-record` job (`:746-760`), download `native-*-linux-amd64` / `native-*-windows-amd64` (both `feov-ocr` and `feov-embed` artifacts) into `${{ runner.temp }}/helpers` and `chmod +x` both helpers on Linux.
    - In `release` job (`:1988-2002`), strip unsuffixed helper binaries (`feov-ocr`, `feov-ocr.exe`, `feov-embed`, `feov-embed.exe`) and `gte-small.int8.ort` from `$DIST` before `releaseassets -check` and checksum generation.
  - [`scripts/fetchbingen/main.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/fetchbingen/main.go) **[MODIFY]** (and regenerate `plugins/*/hooks/fetch-bin.sh` via `go -C scripts run ./fetchbingen`): check and install `feov-embed` and per-arch model exports alongside other binaries.
  - [`plugins/prosthetic-conscience/tools/internal/doctor/main.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/prosthetic-conscience/tools/internal/doctor/main.go) **[MODIFY]**: check and `--fix` helper and model assets.
  - [`scripts/universe.sh`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/universe.sh) **[MODIFY]**: build `feov-embed` into local release and stage into universe cache.
  - [`scripts/agentgen/src/skills/research-protocol/SKILL.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/agentgen/src/skills/research-protocol/SKILL.md) **[MODIFY]** (and regenerate `plugins/frank-exchange-of-views/skills/research-protocol/SKILL.md` and `plugins/frank-exchange-of-views/agents/*.md` via `go -C scripts run ./agentgen`, plus [`plugins/frank-exchange-of-views/tools/internal/difftest/testdata/manual_runs_every_help_page_live.golden`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/difftest/testdata/manual_runs_every_help_page_live.golden) via `UPDATE_GOLDENS=1`).

---

### 3. Consumer Censuses (Run Against `origin/main` @ `fe523726b`)

#### Census A: `flags.Match` and `flags.Quote` Consumers
Command:
```sh
git grep -n "flags\.\(Match\|Quote\)" plugins/frank-exchange-of-views/tools/
```
Output (**43 lines across 12 files** on `origin/main` @ `fe523726b`):
- Acting verbs using `flags.Quote` only (`blue/cite.go`: 3, `blue/compute.go`: 3, `blue/edit.go`: 4, `blue/retire.go`: 6, `lens/finding.go`: 3, `lens/mint.go`: 5, `lens/verify.go`: 2, `motion/verbs.go`: 2) — **NO CHANGE**.
- View selector and source search files (`seat/selector.go`: 4, `seat/source.go`: 3, `seat/verbs.go`: 2, `selectorcensus_test.go`: 6) keep `flags.Match` and `flags.Quote` and add `flags.Search` on `report` and `source` — **[MODIFY]**.

#### Census B: `fetchcache` Text Extraction & Search Consumers
Command:
```sh
git grep -n "SourceTextPath\|OCRTextPath\|SearchSource\|foldExtraction\|extractHTML" plugins/frank-exchange-of-views/tools/
```
Output (**54 lines across 20 files** on `origin/main` @ `fe523726b`):
- [`internal/cli/seat/source.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat/source.go) (1 line): routes single-source `--match`/`--quote` to `SearchSource`, cross-source `--match`/`--quote` to `SearchCrossSources`, single-source `--search` to `SearchSourceHybrid`, and cross-source `--search` to `SearchCrossSourcesHybrid`. **[MODIFY]**
- [`internal/fetchcache/fetchcache.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/fetchcache.go) (4 lines), [`pageread.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/pageread.go) (3 lines), [`readscanned.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/readscanned.go) (2 lines): call `prewarmSourceVectors(run, entry, origin)` after writing extracted text. **[MODIFY]**
- [`internal/fetchcache/sourcesearch.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/sourcesearch.go) (3 lines) & `sourcesearch_test.go` (10 lines): keep `SearchSource(run, e, match, quote)` signature intact; add `SearchCrossSources`, `SearchSourceHybrid`, `SearchCrossSourcesHybrid`, and `bestChunkSentence`. **[MODIFY]**
- Unchanged callers/tests (31 lines across 14 files: `cli/blue/cite.go`: 1, `cli/fetchsummary.go`: 1, `cli/ocr.go`: 1, `cli/ocrcite_test.go`: 2, `fetchcache/extractor.go`: 2, `fetchcache/htmlextract.go`: 3, `fetchcache/htmlextract_test.go`: 1, `fetchcache/htmlmarkup_test.go`: 7, `fetchcache/htmlpage_test.go`: 1, `fetchcache/pagelocate.go`: 4, `fetchcache/pageread_test.go`: 2, `fetchcache/pagerender.go`: 1, `fetchcache/readscanned_test.go`: 4, `fetchcache/sourcelines_test.go`: 1) — **NO CHANGE**.

#### Census C: Native Helper & Asset Provisioning Consumers
Command:
```sh
git grep -n "feov-ocr\|native/build.sh\|native/PINS.txt\|fetchbingen" .github/ plugins/ scripts/
```
Output (**170 lines across 72 files** on `origin/main` @ `fe523726b`):
- [`.github/workflows/hooks.yml`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/.github/workflows/hooks.yml) (10 lines): native build and test jobs (`native-scope`, `native-build`, `native-run`, `feov-record`), release packaging (`release`). **[MODIFY]**
- [`plugins/frank-exchange-of-views/native/build.sh`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/native/build.sh) (11 lines) and `native/feov-embed/` (Part 3 of `plans/native-helpers.md`). **[MODIFY / NEW]**
- [`plugins/frank-exchange-of-views/tools/internal/nativehelper/`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper/client.go) (`client.go`: 2, `digest_gen.go`: 2, `nativesrc/nativesrc.go`: 8, `nativesrc/nativesrc_test.go`: 16) and `devcmd/nativedigest/main.go` (1 line): add `feov-embed` helper and `ASSETS` pins to `digest_gen.go` and `nativesrc` tests. **[MODIFY]**
- [`scripts/releaseassets/releaseassets.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/releaseassets/releaseassets.go) (1 line) and `releaseassets_test.go` (8 lines): check `ASSETS` exports in `CheckDist` (skipping `staticproof` on `.ort` data files) and add `internal/vecembed/` to `IsNativeSetChange`. **[MODIFY]**
- [`scripts/internal/staticproof/staticproof.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/internal/staticproof/staticproof.go) (1 line), `allowlist_test.go` (4 lines), `scripts/staticproof/main.go` (1 line): Windows PE allowlist for `feov-embed`. **[MODIFY]**
- [`scripts/fetchbingen/main.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/fetchbingen/main.go) (12 lines), `main_test.go` (1 line), and generated `plugins/{frank-exchange-of-views,gray-area,prosthetic-conscience}/hooks/fetch-bin.sh` (3 lines across 3 files). **[MODIFY]**
- [`plugins/prosthetic-conscience/tools/internal/doctor/fetchbin_contract_test.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/prosthetic-conscience/tools/internal/doctor/fetchbin_contract_test.go) (5 lines) and `main.go` / `release_contract_test.go`. **[MODIFY]**
- [`plugins/frank-exchange-of-views/tools/internal/difftest/testdata/manual_runs_every_help_page_live.golden`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/difftest/testdata/manual_runs_every_help_page_live.golden) (4 lines) and 10 generated `plugins/frank-exchange-of-views/agents/*.md` files (10 lines): regenerated via `agentgen` and `UPDATE_GOLDENS=1`. **[MODIFY — generated]**
- Unchanged `feov-ocr`-specific files (68 lines across 44 files: `native/feov-ocr/{PINS.txt,main.cpp}`: 4, `cli/{enginestub_test.go,fetch.go,fetchocr_test.go,ocr.go,ocrcite_real_test.go}`: 8, `fetchcache/{blankreceipt_test.go,corpus_test.go,enginestub_test.go,enginesuccess_test.go,pageread.go}`: 7, `seatprobe/{build.go,fixture_test.go,scanned.go,separatedbuild_test.go}`: 5, `tessocr/{client.go,corpus/render.go,diag_test.go,diagcrash_test.go,engine_stub_test.go,engine_test.go,goldens_test.go,reconstruct_test.go,stubbuild_test.go,tessocr.go,textcells_test.go}`: 26, 13 `tessocr/testdata/corpus/*/reading.golden` files: 13, `releasegate/fuzz/hookinvocation_test.go`: 1, `prosthetic-conscience/hooks/README.md`: 2, `scripts/check/gates.go`: 3) — **NO CHANGE**.

#### Census D: Agent-Facing Prose & Constitutions
Command:
```sh
git grep -n "lexical match\|--match\|show source" scripts/agentgen/ plugins/frank-exchange-of-views/skills/ plugins/frank-exchange-of-views/agents/
```
Output (**249 lines across 14 files** on `origin/main` @ `fe523726b`):
- [`scripts/agentgen/src/skills/research-protocol/SKILL.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/agentgen/src/skills/research-protocol/SKILL.md#L50) (1 line) and generated `plugins/frank-exchange-of-views/skills/research-protocol/SKILL.md` (1 line): teaches when to use exhaustive selection (`--quote` for literal text, `--match` for regex alternations/tells/boundaries) vs. ranked hybrid search (`--search` for conceptual/wording discovery across the report or all cached sources), stating that ranked results are cut and never prove absence. **[MODIFY]**
- 12 generated `plugins/frank-exchange-of-views/agents/*.md` files (247 lines: `blue-lane.md`: 20, `blue-researcher.md`: 23, `blue-synthesizer.md`: 23, `lead-judge.md`: 17, `red-chair.md`: 17, and 7 `red-lens-*.md` files at 21 each): inlined `manual` blocks regenerated via `go -C scripts run ./agentgen`. **[MODIFY — generated]**

---

## IV. Risk & Mitigation

| Risk | Likelihood | Impact | Complexity | Mitigation |
|---|---|---|---|---|
| **R1: Seats treat a top-$K$ cut hybrid `--search` as proof of absence** | High | High | Low | Every `--search` response (text and JSON) prints `ranked_and_cut: true`, `returned: K`, `searched_chunks: N`, `unsearched_sources: […]`, and states that absence from top-$K$ proves nothing and `--quote` / `--match` are the exhaustive checks. `research-protocol/SKILL.md` states the same duty. |
| **R2: `feov-embed` crashes or overruns deadline during search** | Low | High | Low | The process boundary isolates crashes from `feov-record`. The client enforces a 120 s overall deadline and bounds batches to $\le 16$ texts ($\approx 2.7\text{ s}$ execution), killing the child on failure and spawning a fresh helper on the next request. Verified by `TestEveryOpFitsTheDeadline` and `TestEmbedSplitsAtTheBatch`. |
| **R3: Missing or mismatched `feov-embed` or model file** | Medium | High | Low | Explicit hard refusal (`helper-absent` or `helper-digest-mismatch`) naming the asset, compiled-in hash, and `/prosthetic-conscience:doctor --fix`. Never falls back silently to BM25-only. |
| **R4: Stale vectors after `ocr read`, `render-page`, or chunker parameter edits** | Low | High | Low | Source vector cache files are addressed by `<text_sha>.vec` (`sha256` of `SourceTextPath` bytes) and check both `modelID` and `SourceChunkSpec` (`w200_s160`) in the `FEOVVEC2` header. Report paragraph cache (`report_vecs.bin`) keys paragraphs by `sha256(visibleText(para))` and checks `modelID` and `ReportChunkSpec` (`para_v1`) in the `FEOVRPT1` header. |
| **R5: Repeated `show report --search` re-embeds unchanged report paragraphs** | High | Medium | Low | `<run>/cache/report_vecs.bin` caches paragraph embeddings by `sha256(visibleText(para))`, so `show report --search` embeds only the query plus paragraphs added or edited since the last search (1–3 texts instead of 50–300). |
| **R6: 8-bit quantization drift vs FP32 reference embeddings** | Low | Medium | Low | `TestEmbedFidelity` gates on mean cosine $\ge 0.984$, min $\ge 0.966$ against `ref_gte.npy` across 300 texts on every release target in `native-run` (S5). |
| **R7: Unpinned build inputs or `arm64` model asset name mismatch in `native/build.sh`** | Medium | High | Low | `native/build.sh` reads only pinned URLs/SHA256s from `native/PINS.txt` and `native/feov-embed/PINS.txt` (no `$ORT_STATIC_DIR`, no scratch paths, pinned `flatbuffers` wheel), and maps `arm64 -> gte-small.int8.arm.ort` matching `PINS.txt` and `ASSETS`. |
| **R8: Protocol pin advance on `show report --search`** | High | High | Low | `serve = false` in `internal/cli/seat/verbs.go:978-985` when `flags.Search` is given, suppressing `record.ServeReport`. Tested in `selectorcensus_test.go`. |
| **R9: Retrieved 200-word source chunk is not directly citeable or report HTML anchors pollute BM25/embeddings** | High | High | Low | `bestChunkSentence` selects the best-matching sentence within each retrieved 200-word chunk (resolving its actual OCR page via `locatePage` and checking `LocateSourceSpan`) while preserving `matched`/`start`/`end`/`line`/`last_line` for the chunk. `ChunkReport` strips HTML comment anchors via `claimcount.StripAnchors` before indexing/embedding and extracts all `F-`/`C-`/`K-`/`G-` anchor IDs into `anchors`. |

---

## V. Verification Plan (the checklist)

> **Standing rule**: Before running any FEOV test or gate from an agent session, run `unset $(env | grep -o '^FEOV_[A-Z_]*')`.

1. **Native Helper Cross-Build & Static Proof Across All 5 Release Targets**
   - After rebasing onto `native/phase-b-fixes` (which fixes `releaseassets -scope` to diff `HEAD^1 HEAD` so `native-run` executes `full=true` on PRs), run `native/build.sh feov-embed <target>` across all 5 targets (`linux-amd64`, `linux-arm64`, `windows-amd64`, `windows-arm64`, `darwin-arm64`).
   - Run `go -C scripts run ./staticproof <binary>` against each built helper, and run `go -C scripts test -count=1 ./releaseassets/...`.
   - **Pass criterion**: all 5 binaries build from pinned inputs only (verifying `arm64` fetches `gte-small.int8.arm.ort`), pass `staticproof` with 0 forbidden dynamic libraries, and `releaseassets -check` verifies the target-suffixed helpers and declared `ASSETS` exports (`gte-small.int8.amd64.ort`, `gte-small.int8.arm.ort`).
   - *Re-arm when*: `native/feov-embed/**`, `native/frame/**`, `native/build.sh`, or `scripts/releaseassets/**` changes.

2. **Embedder Unit, Tokenizer Parity, Batching, Deadline, Fidelity (S5), & Speed (S6) Suite (`internal/vecembed`)**
   - Run `go test -count=1 -v ./internal/vecembed/` locally and across all 5 targets in CI `native-run` (verifying `--- PASS` lines appear in every target's log, including `vecembed.test.exe` inside the Windows `servercore:ltsc2025` container with hard failure on non-PASS, and never `"skipped because full=false"`).
   - **Pass criteria**:
     - `TestTokenizerParity400`: 400/400 exact token ID match against Hugging Face `thenlper/gte-small` reference tokenizations (`speed.jsonl` + `len200.jsonl`).
     - `TestEmbedFidelity` (S5): mean cosine $\ge 0.984$, min $\ge 0.966$ against `ref_gte.npy` across all 300 texts on all 5 targets; per-target numbers posted from CI logs into the PR.
     - `TestEmbedSpeed` (S6): single-thread throughput $\ge 5.0\text{ chunks/s}$ on 200-word chunks (`len200.jsonl`) on `linux-amd64`, `linux-arm64`, and `darwin-arm64`; cold 1-query latency $\le 1.5\text{ s}$; per-target numbers posted from CI logs into the PR.
     - `TestEveryOpFitsTheDeadline`: 16-chunk batch of 200-word chunks completes in $< 30\text{ s}$ ($\frac{1}{4}$ of the 120 s deadline) on every target.
     - `TestEmbedSplitsAtTheBatch`: fake helper records incoming request batch lengths for 35 texts as `[16, 16, 3]`; deleting the batch split in `vecembed.go` fails the test.
     - `TestMissingOrMismatchedHelperRefusesByName`: missing or mismatched `feov-embed` or `gte-small.int8.ort` refuses with `helper-absent` or `helper-digest-mismatch` and names `/prosthetic-conscience:doctor --fix`.
   - *Re-arm when*: any file under `internal/vecembed/` or `native/feov-embed/` changes.

3. **Hybrid Retrieval, Vector Cache Key, Report Paragraph Cache, Anchor Stripping, & Best-Sentence Selection Suite (`internal/hybridsearch` & `internal/fetchcache`)**
   - Run `go test -count=1 -v ./internal/hybridsearch/ ./internal/fetchcache/`.
   - **Pass criteria**:
     - `TestVecCacheRejectsChangedModelOrChunkSpec`: changing either `modelID` or `chunkSpec` causes `ReadVecCache` and `ReadReportVecCache` to reject the cached file and recompute vectors.
     - `TestReportSearchCachesUnchangedParagraphs`: running `show report --search` twice after editing 1 paragraph of a 20-paragraph report sends only 2 texts (query + 1 edited paragraph) to the embedder on the second call.
     - `TestChunkReportStripsAnchorsAndExtractsAllTypes`: verifies `ChunkReport` strips `<!--fx:F-...-->`, `<!--cite:C-...-->`, `<!--computation:K-...-->`, and `<!--gap:G-...-->` from `Text` before BM25/embedding and populates `Anchors` with all anchor IDs on the paragraph.
     - `TestRankBM25PropagatesSQLiteErrors`: an FTS5 execution fault returns a non-nil error rather than silently returning vector-only results.
     - `TestPrewarmMatchesFullSourceTextSha`: for both `embedded` (`.txt`) and `ocr_derived` (`.ocr.txt`) sources, `prewarmSourceVectors` writes `<text_sha>.vec` with the exact `text_sha` that `SearchSourceHybrid` reads.
     - `TestHybridSourceHitSelectsCiteableSentenceAndPage`: on a multi-sentence 200-word chunk (including a chunk crossing an OCR page boundary), `SearchSourceHybrid` and `SearchCrossSourcesHybrid` populate `matched`/`start`/`end`/`line`/`last_line` for the chunk and `sentence`/`sentence_start`/`sentence_end`/`source_span` (and the sentence's actual `page`) for the query-matching sentence, which `blue cite --url <url> --source-span "<source_span>"` accepts.
     - Searching a run where one cached PDF is `image_only` (no `.ocr.txt`) lists that source in `unsearched_sources` alongside the ranked hits from extracted sources.
   - *Re-arm when*: any file under `internal/hybridsearch/` or `internal/fetchcache/` changes.

4. **CLI Surface, Selector Census, & Provisioning Contract**
   - Run `go test -count=1 -v ./internal/cli/... ./internal/flags/... ./internal/terms/...` and `go -C ../../prosthetic-conscience/tools test -count=1 -v ./internal/doctor/...`.
   - **Pass criteria**:
     - `TestEverySelectorThatAViewAdvertisesSelects` passes for **both `--match` and `--quote`** across every view in `selectorViews(t)`.
     - `TestRankedSearchStatesCutAndScope` verifies `show report --search` and `show source --search` (single-source and cross-source) emit `ranked_and_cut`, `searched_chunks`, `unsearched_sources`, and citeable `source_span` + `page` + `line`/`last_line`.
     - `TestSearchDoesNotAdvanceServeReportPin` verifies that `show report --search` does not record `ServeReport` or advance the reading pin for lens seats.
     - `TestSearchMutuallyExclusiveWithQuoteAndMatch` verifies that combining `--search` with `--quote` or `--match` refuses with `feov.Validation`.
     - `fetchbin_contract_test.go` and `release_contract_test.go` verify `feov-embed` and per-arch model exports are checksummed in `SHA256SUMS` and installed by `fetch-bin.sh` and `sc-doctor --fix`.
   - *Re-arm when*: any file under `internal/cli/`, `internal/flags/`, `internal/terms/`, `scripts/fetchbingen/`, or `internal/doctor/` changes.

5. **Generators, Repository Gates, & End-to-End Smoke Verification**
   - Regenerate carriers (`go -C scripts run ./fetchbingen`, `go -C scripts run ./agentgen`, `go -C plugins/frank-exchange-of-views/tools run ./devcmd/vocabgen`, `go -C scripts run ./vocabdoc`, `go -C plugins/frank-exchange-of-views/tools run ./devcmd/cmdtreegen`, `UPDATE_GOLDENS=1`), read every changed golden line and `--help` diff one command at a time, and run `go -C scripts run ./check`.
   - Build an isolated universe (`scripts/universe.sh build`), verify the staged marketplace carries the updated help text, `feov-embed`, and `gte-small.int8.ort`, and drive `feov-record show report --search` and `feov-record show source --search` against an archived real-source run (`run-archive/`).
   - **Pass criterion**: `go -C scripts run ./check` exits 0 with 0 failed gates, CI `release-fuzz` (0–3 + tally) and full 5-target `native-run` pass, and an independent review approves before merge.
