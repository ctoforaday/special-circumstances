# Hybrid lexical + vector search over the report and cached sources (`feov-embed` + `gte-small` int8)

> STATUS: approved by `gblock` 2026-11-10, with `--search` on `show report` and `show source` only (Option S-Prose, §III.1.C); not started (decision document for [#1171](https://github.com/ctoforaday/special-circumstances/issues/1171)). Revised per `plans/native-helpers.md` Part 3, `REVIEW.md`, `plancheck/CHECK.md`, and `plancheck/CHECK2.md` (2026-11-10). Audited: 2026-11-10 (`plan-auditor-agy` — Antigravity port of the shipped `plan-auditor` — Round 1 `10375d17` FAIL → fixed; Rounds 2–5 `5293331a`, `b7e94e63`, `0d359c43`, `7b80c3b0` PASS, followed by local factual text corrections from `CHECK2.md`).

Ruled by `gblock` across 2026-10-31, 2026-11-01, 2026-11-04 to 2026-11-06 (`plans/native-helpers.md`), and 2026-11-10:
1. **Hybrid in one go**: ship SQLite FTS5 BM25 (`porter unicode61`) and dense vector search together, fused via Reciprocal Rank Fusion ($\text{RRF}, k=60$).
2. **Ranked read is distinct from exhaustive `--quote`**: `--quote` (literal substring) stays exhaustive across every view, returning every match with zero truncation (ruled 2026-10-31). `--search` is a ranked top-$K$ read whose output always states that it is ranked and cut, how many chunks and sources were searched, which cached sources had no extracted text, and that absence from top-$K$ proves nothing.
3. **Thin C++ native helper executable (`feov-embed`)**: per `plans/native-helpers.md` Part 3, vector embedding runs in a standalone native helper executable (`feov-embed`), built with ONNX Runtime 1.31.1 (minimal Release build with reduced operator set `ops.config`) across all 5 release targets (F15). Communicates with Go via a framed stdio protocol over a private descriptor (`dup`/`dup2`). Go compiles with `CGO_ENABLED=0` across all published binaries.
4. **Model selection delegated by `gblock` (2026-11-01: *"Can you choose a model based on quality tradeoffs?"*) and asset provisioning ruled (F1, F10; per-arch exports are the lead's decision L2, disclosed to `gblock`)**: the agent selected per-arch ONNX Runtime 8-bit quantized `gte-small` exports (`gte-small.int8.amd64.ort` and `gte-small.int8.arm.ort`, `33,844,232` bytes each, 12-layer BERT bi-encoder, 384 dims, mean pool, symmetric query/doc encoding) based on the measured BEIR and seat replay quality/size/speed tradeoffs in §I.3; per `gblock`'s rulings (F1, F10, L2), the two exports are hosted on a one-time release (`model-gte-small-v1`), pinned by SHA256 in `native/feov-embed/PINS.txt`, and provisioned via `fetch-bin.sh` and `sc-doctor`.
5. **Empirical audit of `--match` (ruled 2026-11-10)**: *"If `--search` replaces the use case for match, remove match. But audit it for that — it might be that we want all three universally."* Section III.1 reports the empirical audit over the October seat transcripts (`~/.claude/scratch/vecplan/replay/` and `plancheck/chk/match-sample-60.txt`): `--search` replaces roughly a quarter of report `--match` calls (synonym/paraphrase hunts for a concept), but **cannot** replace exhaustive voice/tell absence sweeps (`201` calls by the voice, dark-side, and architecture lenses plus blue self-checks, and `190` of the `812` report alternations that returned zero as proof of absence), multi-item/ID alternations (`14` multi-ID calls on `report`), or regex boundary/character-class checks (`84` on report, `14` on source), nor structured-view absence checks ($\le 55$ invocations, `37` on `inquiries`, plus `show inquiries --match … returns zero` cited inside `--verified-with` evidence). Accordingly, **this plan proposes retaining `--match` alongside `--quote`**, and Section III.1 presents the evaluated fork on whether `--search` scopes to the prose corpora (`show report` and `show source`) or all selector-bearing views universally.

---

## I. Summary & Goals

### 1. Value Case & Problem Statement

Since [#1171](https://github.com/ctoforaday/special-circumstances/issues/1171) was filed on 2026-09-26, three prerequisites shipped on `main`:
- HTML text extraction (`extractHTML` in [`internal/fetchcache/htmlextract.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/htmlextract.go), PR #1673).
- Single-source literal and regex search (`feov-record show source --sha <sha|--anchor C-…> --quote|--match`, PR #1662).
- Removal of token-Jaccard `lens near-match` (`internal/record/nearmatch.go`, retired 2026-10-24).

What remains broken is **vocabulary mismatch** and **cross-source discovery**:
- Across October 2026 seat transcripts (`~/.claude/scratch/vecplan/replay/SUMMARY.md` and `searches.jsonl`), seats executed:
  - **1,094** `show report --match` statements (`239` returning zero hits, `649` hit, `6` error, `200` unattributed) and **400** `show report --quote` statements (`106` returning zero hits, `205` hit, `89` unattributed) — **1,494** report selector calls in total (**345** with zero hits), plus **215** `show report | grep` pipelines.
  - **121** `show source --match` statements (`5` zero, `83` hit, `33` unattributed) and **12** `show source --quote` statements (`3` hit, `1` error, `8` unattributed), plus **2,452** `grep`/`rg` commands over `<run>/cache/`.
  - **$\le 55$** actual `--match` invocations on structured record views after stripping `--verified-with`/`--note` prose strings (`inquiries`: `37`, `board`: `7`, `evidence`: `6`, `avenues`: `2`, `findings`: `2`, `work`: `1`; `plancheck/chk2/run/sv.py`), plus `--verified-with` evidence strings where seats cite `show inquiries --match … returns zero` as closing proof of absence.
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
- **Why 110M–560M models (`embeddinggemma-2`, `bge-m3`, `nomic-v2`) are rejected**: they are **4× to 17× larger** (139–605 MiB vs. 33.8 MiB for every consumer download). In addition, `google/embeddinggemma-2` uses the gated `gemma` license on Hugging Face (requiring authentication tokens, which breaks unauthenticated `PINS.txt` / `fetch-bin.sh` downloads), whereas `gte-small` and `bge-small-en-v1.5` are MIT-licensed.
- **Why `gte-small.int8.ort` (`ort8`) is chosen**:
  1. **Passes the B4 quality bar in 8-bit quantization with 4× smaller model assets (`33.8 MiB` vs `127.0 MiB` FP32)**: `gte-small.int8.ort` hybrid beats `bm25/porter` by **`+0.048 [+0.026, +0.069]`** on `SciFact` (`0.731` vs `0.682`) and **`+0.040 [+0.027, +0.054]`** on `NFCorpus` (`0.362` vs `0.322`), retaining 99.2% of FP32 `gte-small`'s BEIR Mean Hybrid nDCG@10 (`0.5465` vs `0.5510`). In the FP32 sweep, `gte-small` is also the only model whose standalone vector 95% CI clears BM25 on both BEIR benchmarks (`SciFact` `+0.045 [+0.013, +0.075]`, `NFCorpus` `+0.026 [+0.005, +0.047]`; `bge-small-en-v1.5` standalone on `SciFact` is `+0.030 [-0.003, +0.061]`, crossing zero).
  2. **Stronger top-5 and top-10 recall on the 79-row seat replay benchmark**: `gte-small.int8.ort` (`ort8`) achieves **`0.861` `Hit@5` and `0.924` `Hit@10`** on All-79 (`0.873` `Hit@5` and `0.927` `Hit@10` on Wording-55; rescuing 4 of 12 BM25 misses with 3 losses, net +1 over BM25 `0.848` `Hit@5` and +4 over BM25 `0.873` `Hit@10`), whereas `bge-small-en-v1.5` (even in FP32) achieves `0.835` `Hit@5` and `0.911` `Hit@10` (and FP32 `gte-small` achieved `0.835` `Hit@5` and `0.937` `Hit@10`). Because `--search` returns a top-$K=10$ list, `Hit@5` and `Hit@10` govern whether the passage reaches the seat.
  3. **Self-contained metadata and symmetric encoding ([[facts-are-fields]])**: `gte-small` uses symmetric query/document encoding (no magic prompt prefix). By contrast, `bge-small-en-v1.5` requires prepending the Asymmetric Query Prefix `"Represent this sentence for searching relevant passages: "` to queries — an external string that is not recorded in the `.ort` model file.
  4. **Proven 8-bit fidelity and measured single-thread throughput (`ortstatic/REPORT.md` §1, §3, §5)**:
     - Against the FP32 `sentence-transformers` reference (`ref_gte.npy` over 300 texts in `speed.jsonl`), `gte-small.int8.ort` measured cosine similarity **mean `0.984411–0.984464`, min `0.967607–0.969149`** across all targets (`ortstatic/REPORT.md` §1, `native-helpers.md:263`, `ortq/score.log` line 1).
     - In the size-optimised minimal build (`--config MinSizeRel`, `-Os`), 1-thread `int8` throughput on 200-word chunks (`len200.jsonl`) measured **`12.72 chunks/s`** on `linux-amd64` (EPYC 9V45; `6.11 chunks/s` on slower EPYC 7763), **`5.87 chunks/s`** on `linux-arm64` (Neoverse-N2), and **`5.88 chunks/s`** on `darwin-arm64` (Apple M1 Virtual), adding `4.4–5.1 MB` (`ortstatic/REPORT.md` §1 & §3).
     - Because `-Os` (`MinSizeRel`) penalises x64 MLAS kernels by `1.24–2.2×` (`ortstatic/REPORT.md` §5), building with `--config Release` (`minrel`) recovers full unreduced speed on `linux-amd64` (**`7.12 chunks/s`** on EPYC 7763 vs `6.11 chunks/s` for `-Os`, matching the `33.9 MB` full Release build's `7.07 chunks/s`) for `+1.4 MB` (`6,057,720 B` added on `linux-amd64` zig). `native/build.sh` therefore compiles ONNX Runtime with `--config Release` (`minrel`).

### 4. Quantitative Success Criteria

| Gate | Metric | Target |
|---|---|---|
| **G1: Retrieval quality (gold benchmarks & seat replay)** | Hybrid (`bm25/porter unicode61` + `gte-small.int8.ort` RRF $k=60$) nDCG@10 vs `bm25/porter`, reproduced through `feov-embed` | 95% paired bootstrap CI lower bound $> 0$ on both `SciFact` (`+0.048 [+0.026, +0.069]`) and `NFCorpus` (`+0.040 [+0.027, +0.054]`), and replay hybrid `Hit@5 = 0.861` / `Hit@10 = 0.924` reproduced through the helper (`ortq/RESULT.md`, `native-helpers.md` §V.11) |
| **G2: Embedding fidelity & per-arch export verification (S5)** | Handshake `model_sha256` matches arch pin (`TestEmbedLoadsTheExportForItsArch`); cosine similarity between `feov-embed` (`gte-small.int8.ort`) and FP32 `sentence-transformers` `thenlper/gte-small` over the 300-text reference (`ref_gte.npy`) | Mean cosine $\ge 0.984$, min $\ge 0.966$ executed in `native-run` (`FEOV_NATIVE_MEASURE=1`) on all 5 release targets (`TestEmbedFidelityMeetsS5`) |
| **G3: CPU throughput, spawn-to-handshake, & worst-case deadline margin (S6 & §V.4)** | Single-thread throughput on 200-word chunks (`len200.jsonl`); spawn-to-handshake latency and worst-case `embedBatch = 16` texts $\times 512$ token IDs vs 120 s deadline | Throughput $\ge 5.0\text{ chunks/s}$ on `linux-amd64`, `linux-arm64`, and `darwin-arm64` (`TestEmbedSpeedMeetsS6`); spawn-to-handshake and `16 × 512`-id `embed` request each $< 30\text{ s}$ ($\frac{1}{4}$ of 120 s deadline, `TestEveryOpFitsTheDeadline`, gated on `FEOV_NATIVE_MEASURE=1`) |
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
  - Handshake: the first frame exchange verifies protocol version (`v == 1`), helper source digest (`FEOV_HELPER_DIGEST`, `result[0]`), and loaded model SHA256 (`model_sha256`, emitted as the separate frame field `result[1]` after `FEOV_HELPER_DIGEST` in the `hello` response frame, matching `Client.handshake`'s `result := f.Fields[2:]` check in `internal/nativehelper/client.go:258-270`). Any request frame whose version is not `1` or that arrives before a completed handshake is refused.
- **Single embedding op**: `embed(ids, mask) -> float32[n][384]`. Token IDs (`int64[1, seq_len]`) and attention masks (`int64[1, seq_len]`) are fed as 2 input tensors (`input_ids` and `attention_mask`) to the ORT session one text at a time (no cross-text padding variation; the exported `.ort` model wrapper binds `token_type_ids` to zeros internally); mean pooling over masked tokens and L2 normalization ($\|\mathbf{v}\|_2 = 1$) execute inside `feov-embed` (`main.cpp`). Intra-op and inter-op threads are both pinned to `1`.
- **Request bounds and deadline**:
  - The Go client applies a 120 s overall deadline per request (`DefaultDeadline`).
  - To guarantee that every request completes in $< \frac{1}{4}$ of the deadline on the slowest target (`30 s` bar), the client splits inputs into batches of at most `embedBatch = 16` texts of at most `512` token IDs each. `TestEveryOpFitsTheDeadline` (`native-helpers.md` §V.4) times both spawn-to-handshake (including hashing the 33.8 MiB model and creating the ORT session) and the worst-case `embed` request of `embedBatch = 16` texts $\times 512$ token IDs, asserting each completes in $< 30\text{ s}$ on every target.
- **Lifecycle & failure handling**:
  - On EOF on stdin, the helper exits 0.
  - On helper crash, deadline overrun, or protocol error during a request, `feov-record` terminates the child and fails that request loudly. The next request spawns a fresh helper.
  - A failure during startup (`helper-absent` or `helper-digest-mismatch`) latches on the client to prevent spawn loops.

### 2. Pure Go Client & WordPiece Tokenizer (`internal/vecembed`)

Because native C++ execution lives entirely inside `feov-embed`, `internal/vecembed` is written in **pure Go** and compiles with `CGO_ENABLED=0` across all platforms:
- **`vecembed.go`**: defines the `Embedder` interface (`Embed(texts []string) ([][]float32, error)`, `ModelID() string`, `Close() error`), batch splitting logic (`embedBatch = 16`, max `512` token IDs per text), deterministic cosine similarity, and error handling. `ModelID()` returns `"gte-small-int8@" + nativehelper.HelperDigest("feov-embed") + "+target@" + runtime.GOOS + "-" + runtime.GOARCH` (per R9: because 8-bit quantized vectors differ slightly by architecture — mean fidelity `0.984411` on `amd64` vs `0.984464` on `arm64` — `ModelID()` embeds `os-arch` so `<text_sha>.vec` and `report_vecs.bin` cache headers include the target alongside the helper digest and chunk spec).
- **`tokenize.go`**: deterministic pure-Go BERT WordPiece tokenizer (control-char cleaning, CJK splitting, NFD accent stripping, lowercasing, punctuation splitting, WordPiece greedy longest-match-first against the embedded `thenlper/gte-small` `vocab.txt` of `30,522` tokens, `[CLS]`/`[SEP]` wrapping, and 512-token truncation) matching Hugging Face `AutoTokenizer.from_pretrained('thenlper/gte-small')` across **all 400/400 reference texts** (`300` in `internal/vecembed/testdata/speed.jsonl` + `100` in `internal/vecembed/testdata/len200.jsonl`, verified in `TestTokenizerParity400` against `internal/vecembed/testdata/hf_tokens_400.json`).
- **Process management via `internal/nativehelper`**: uses `nativehelper.Client` ([`internal/nativehelper/client.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper/client.go)) to spawn `feov-embed`, verify the handshake (where `Client.handshake` checks `1 + len(c.pins.Assets)` fields in the `hello` response against `c.pins.Digest` and `a.SHA256[runtime.GOARCH]` from `digest_gen.go`), enforce deadlines, and exchange frames.
- **Distinct failure contracts**: produces explicit named errors (`helper-absent`, `helper-digest-mismatch`, `helper-crashed`, `helper-deadline`, `helper-protocol`), never silently degrading to BM25-only search.

### 3. Model Supply via `fetch-bin.sh`, `sc-doctor`, and `universe.sh`

Per `gblock`'s rulings F1 and F10 and the lead's disclosed decision L2 (`plans/native-helpers.md`), 8-bit quantized models are distributed as platform release assets rather than embedded via `//go:embed`:
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
   - At handshake, `feov-embed` reads `gte-small.int8.ort` beside its binary (or from `NATIVE_HELPER_DIR`), computes its SHA256, and reports it in the handshake response frame as a separate frame field (`result[0] = <helper_digest>`, `result[1] = <model_sha256>`). `nativehelper.Client.handshake` (`client.go:258-270`) verifies both `result[0] == c.pins.Digest` (`nativehelper.HelperDigest("feov-embed")`) and `result[1] == a.SHA256[runtime.GOARCH]` from `c.pins.Assets` in `digest_gen.go`.
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
  - **Why report caching is required**: at the `5.87–7.12 chunks/s` single-threaded rate measured on 200-word chunks, re-embedding all 50–300 paragraphs of `blue/report.md` from scratch on every `show report --search` call would cost **8–50 s** of pure embedding work alone, whereas a typical `blue edit` modifies only a few paragraphs while the rest of the report is unchanged.
  - **Cache structure**: `<run>/cache/report_vecs.bin` stores a header (`FEOVRPT1`, `modelID` including `+target@<os>-<arch>`, `ReportChunkSpec = "para_v1"`, `uint32(count)`) followed by entries mapping `para_sha = sha256(visibleText(para))` (`[32]byte`) to `float32[384]`.
  - **Query-time execution**: on `show report --search <query>`, `feov-record` chunks `blue/report.md` into $N$ paragraphs (`ChunkReport`), loads `<run>/cache/report_vecs.bin` (if `modelID` and `ReportChunkSpec` match), collects the $M \le N$ paragraphs whose `para_sha` (`sha256(visibleText)`) is not yet in the map, calls `embedder.Embed` on `[query] + uncached_paragraphs` ($1 + M$ texts), writes the updated map back to `<run>/cache/report_vecs.bin` atomically via `writeReplacing`, and runs hybrid BM25 + cosine RRF over all $N$ paragraphs (`SearchWithVectors`). On a warm report (or after a 1-paragraph edit), `feov-record` still pays the one-shot per-invocation helper cold start (spawning `feov-embed`, hashing the 33.8 MiB `.ort` model at handshake, and creating the ORT session, gated at $\le 1.5\text{ s}$ for a 1-query cold start in `TestEmbedSpeedMeetsS6` and $< 30\text{ s}$ in `TestEveryOpFitsTheDeadline`) plus embedding only 1–2 texts (at the `5.87–12.72 chunks/s` rate measured on 200-word chunks, 1 text is `79–170 ms` and 2 texts are `157–341 ms`, rather than `8–50 s` to re-embed all 50–300 paragraphs).
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

Per `gblock`'s 2026-11-10 ruling (*"If `--search` replaces the use case for match, remove match. But audit it for that — it might be that we want all three universally"*), we audited the October 2026 seat transcripts in `~/.claude/scratch/vecplan/replay/` (`searches.jsonl`, `acts.jsonl`, `SUMMARY.md`), `plancheck/chk/match-sample-60.txt`, and `plancheck/chk2/run/sv.py`.

#### A. October Call Volume (`searches.jsonl` & `acts.jsonl`)

- **`show report`**: **1,094** `--match` statements (`239` returning zero hits, `649` hit, `6` error, `200` unattributed; `812` of the `1,094` use `|` alternation, and **`190` of those `812` alternations returned zero**), **400** `--quote` statements (`106` zero, `205` hit, `89` unattributed), `52` `--anchor` calls, and `3,508` bare calls (plus `215` `| grep` pipelines).
- **`show source`**: **121** `--match` statements (`5` zero, `83` hit, `33` unattributed; `81` use `|` alternation), **12** `--quote` statements (`0` zero, `3` hit, `1` error, `8` unattributed), and `2,452` `grep`/`rg` commands over `<run>/cache/`.
- **Structured record views (`board`, `findings`, `inquiries`, `work`, `evidence`, `changes`, `avenues`)**: after stripping `--verified-with` and `--note` prose strings (`plancheck/chk2/run/sv.py`), actual October `--match` invocations on structured views number **$\le 55$** (`inquiries`: `37`, `board`: `7`, `evidence`: `6`, `avenues`: `2`, `findings`: `2`, `work`: `1`, with `0` multi-ID alternations). Crucially, seats invoke `show inquiries --match …` (and `show board --match …`, `show findings --match …`) and cite `"show inquiries --match … returns zero"` inside `--verified-with` strings as the formal verification proof that no open inquiry, gap, or finding matches a pattern.

#### B. Mechanical & Intent Breakdown of `--match` on `report` ($N=1{,}094$) and `source` ($N=121$)

First, classifying every `searches.jsonl` `--match` pattern mechanically (`plancheck/CHECK.md` & `CHECK2.md` §`--match` evidence):

| Mechanical Pattern Class (`searches.jsonl`) | `report` ($N=1{,}094$) | `source` ($N=121$) | Representative Real Seat Patterns (`searches.jsonl`) |
|---|---:|---:|---|
| **1. Multi-ID alternation** (2+ anchor/record IDs `P-…\|P-…`, `K-…`) | **14** (`1.3%`) | **0** (`0.0%`) | `Trial division\|Base 3 passes\|P-e6803dfe\|P-c1c0989f\|P-47ce38d8`; `K-bc8ac2b6\|bc8ac2b6` |
| **2. Structural regex** (`\b`, `^`, `$`, `[...]`, `\d`) | **84** (`7.7%`) | **14** (`11.6%`) | `\b(read\|fetched\|footnotes?\|recomputed\|as summaries\|given with)\b`; `at least [a-z]+ times`; `\b91\b` |
| **3. Alternations by `red-lens-voice`, `red-lens-dark-side`, & `red-lens-architecture`** | **201** (`18.4%`) | **10** (`8.3%`) | `search\|retraction\|verified\|checked against\|cached\|sitting\|epoch\|this run\|the debate\|lens` (`zero` is wanted) |
| **4. Other phrase/number/formula alternations** (by `blue-*`, `red-lens-evidence`, `red-lens-logic`, `red-lens-computation`, `red-chair`, `lead-judge`) | **527** (`48.2%`) | **60** (`49.6%`) | Mix of synonym hunts, blue self-audits for voice/tells, and multi-item/formula lookups (`59` phrase alternations + `1` numbers/formula alternation on `source`; sampled below) |
| **5. Quantifier, wildcard, or escape only** (`?`, `+`, `.*`, `\^`, `\[`, without `\|`) | **10** (`0.9%`) | **0** (`0.0%`) | `Z\[1/7\]`; `3\^90` |
| **6. Plain literal word, phrase, or number** (no regex metacharacters) | **258** (`23.6%`) | **37** (`30.6%`) | `clear statement`, `closes the gap for this number` (`45` zero hits on `report`); `congruence` |

Second, because Mechanical Class 4 (`527` other alternations on `report`) groups by seat role rather than intent, inspecting a 60-pattern random sample of the `812` report alternations (`~/.claude/scratch/vecplan/plancheck/chk/match-sample-60.txt`) shows that report alternations split roughly into **three equal intent shares (~1/3 each)**:
1. **~21 / 60 (~35% of alternations, or ~25% of all report `--match` calls) are synonym/paraphrase hunts for a single concept** (e.g., `any other ring|every ring|all rings|other rings`, `single-base|one-base|single base|one base`, `settles nothing|shows only that`, `wider ring|any other ring|other number system|other rings`). Here `--search` is a direct improvement when the seat wants to find where a concept is discussed (`Hit@10 = 0.924` on the 79-row replay set) — though even within this group, some calls (e.g., `no statement about 91|says nothing about 91`, `no published source|not be retrieved`) are checking whether a claim is **absent** (`zero` returned).
2. **~19 / 60 (~32% of alternations) are voice / tell / banned-vocabulary absence sweeps** — not only by `red-lens-voice`, `red-lens-dark-side`, and `red-lens-architecture`, but also by **`blue-researcher` and `blue-synthesizer` auditing their own draft** for banned process/meta-words before submitting (e.g., sample rows 6, 10, 12, 15, 36, 56: `web search|search results|search covered|at the leaf|literature search|recorded program|Python program|was read to|one search|one query`, returning `zero`). Across all `812` report alternations, **`190` returned zero hits**. Ranked top-$K$ `--search` can **never** return zero on a non-empty report and therefore cannot serve as an absence check.
3. **~20 / 60 (~33% of alternations) are multi-item / multi-section / multi-number batch lookups** checking several distinct items, headings, or numbers in one call (e.g., sample rows 0, 2, 5, 27, 28, 35, 44: `A002808|list 91|Does the full|arithmetic above does not depend|A000040`, `Future research|Research areas|Alternatives`, `program|bundle|three|six|seven|nine`, `Fermat|35|71`).

#### C. Audit Verdict & Evaluated Design Fork on Selector Scope

1. **`--match` MUST be kept alongside `--quote` on every view**:
   - The transcript evidence strongly refutes removing `--match`:
     - **`190` of the `812` report alternations (`23.4%`) returned zero**, and roughly a third of all report alternations (`201` by the voice/dark-side/architecture lenses plus blue's own self-audit sweeps) are exhaustive banned-vocabulary / tell sweeps where `0` matches is the required proof of absence and where any match must be returned uncut.
     - Another third of report alternations batch multiple distinct IDs (`14` multi-ID calls on `report`), section headings, or distinct items into a single exhaustive read (`--quote` would turn 1 call into 4–16 calls), and **`84` report calls + `14` source calls** use structural regexes (`\b`, `^`, `$`, `[a-z]+`, `\d`) that neither `--quote` nor `--search` can express.
     - On structured views (`inquiries`, `board`, `findings`, `work`, `evidence`, `changes`, `avenues`), seats use `--match` (`\le 55` calls, `37` on `inquiries`) for exhaustive absence checks and cite `show inquiries --match … returns zero` inside `--verified-with` evidence.
2. **Where `--search` is registered — Evaluated Design Fork for `gblock`**:
   - **Option S-Prose (ruled by `gblock` 2026-11-10; the option wording is the agent's): Keep `--quote` and `--match` on every view (`AddSelectorFlags`); add `--search` on `show report` and `show source` (single-source and cross-source).**
     - *Rationale for Option S-Prose*:
       1. `report` and `source` are long unstructured prose corpora (hundreds of chunks across multiple external documents) where a seat does not know the author's wording or which cached source holds a passage.
       2. Structured record views (`board`, `findings`, `inquiries`, `work`, `evidence`, `changes`, `avenues`) are small JSON arrays of record events (estimated at a few dozen entries per run) that a seat can read whole from the local SQLite database without spawning a helper or filter exhaustively with `--quote` / `--match`.
       3. Preserves the crisp contract in [`internal/cli/seat/selector.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat/selector.go#L46): `AddSelectorFlags` (`--match` and `--quote`) are **exhaustive boolean selectors** (`complete: true`, never truncated, 0 matches on non-occurring strings as tested by `TestEverySelectorThatAViewAdvertisesSelects`), whereas `--search` on `report` and `source` is a **ranked, cut top-$K$ retrieval mode** (`ranked_and_cut: true`, `search_scope`).
     - *Stated limit of the replay evidence*: `replay.jsonl` (79 rows: 50 report, 29 source) only sampled `report`, `source`, and `grep` searches (`SUMMARY.md` step 2), so it does not empirically measure whether seats ever experience vocabulary mismatch on structured views.
   - **Option S-Universal (Alternative for `gblock`'s ruling): Register all three (`--quote`, `--match`, `--search`) universally on every selector-bearing view (`AddSelectorFlags` registers all three).**
     - *How Option S-Universal works if selected*: `WriteSelected` in `selector.go` gains a `searchJSONArrays(run, body, query, topK)` branch that serializes each top-level JSON array item, embeds uncached items via `feov-embed` (cached by `sha256(item_bytes)` in `<run>/cache/view_vecs.bin` with `ModelID()` including `+target@<os>-<arch>`), ranks items with hybrid BM25 (`porter unicode61`) + cosine RRF ($k=60$), keeps top-$K=10$ per array, and attaches `"search_scope": {"ranked_and_cut": true, "returned": kept, "searched_entries": of, "notice": "..."}`.
     - *Tradeoff against Option S-Universal*: On a `board` or `findings` array with $> 10$ entries, cutting to top-$K=10$ risks hiding open gaps or findings a seat needs to account for while adding a `feov-embed` helper cold-start + IPC round-trip to an in-process SQLite metadata projection.

#### D. Surface & Mutual Exclusion Rules (under Option S-Prose)
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
  - `main.cpp` **[NEW]**: thin C++ kernel hosting ONNX Runtime 1.31.1 minimal Release build, private descriptor frame protocol (`dup`/`dup2`), protocol version (`v == 1`) and handshake-first enforcement (reporting `FEOV_HELPER_DIGEST` and `model_sha256` as separate response frame fields `result[0]` and `result[1]`), 1-thread intra/inter-op execution, 2-tensor (`input_ids`, `attention_mask`) per-text session run, mean pooling over attention mask, and L2 normalization.
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
  - `vecembed.go` **[NEW]**: pure-Go `Embedder` interface, `Default() (Embedder, error)`, batching logic (`embedBatch = 16`, max `512` token IDs per text), L2 normalization, `Cosine(a, b []float32) float32`, `ModelID()` returning `"gte-small-int8@" + nativehelper.HelperDigest("feov-embed") + "+target@" + runtime.GOOS + "-" + runtime.GOARCH`. Compiles with `CGO_ENABLED=0`.
  - `tokenize.go` **[NEW]**: pure-Go BERT WordPiece tokenizer matching `thenlper/gte-small` vocabulary (`30,522` tokens) across all 400/400 reference texts (`testdata/speed.jsonl` + `testdata/len200.jsonl`, compared against `testdata/hf_tokens_400.json`).
  - `vecembed_test.go` **[NEW]** (carrying `plans/native-helpers.md` §V.4 and §V.11 in full):
    - `TestTokenizerParity400`: verifies exact token ID parity against all 400 reference texts from Hugging Face `AutoTokenizer.from_pretrained('thenlper/gte-small')` (`testdata/speed.jsonl` + `testdata/len200.jsonl` vs `testdata/hf_tokens_400.json`).
    - `TestEmbedSplitsAtTheBatch` (`native-helpers.md` §V.4): uses a fake helper that records every incoming `embed` request frame's batch length and per-text token ID length, verifying that **1,000 texts** (including over-length inputs exceeding 512 tokens) go out as $\lceil 1000 / \text{embedBatch} \rceil = 63$ requests (`62 × 16 + 1 × 8`), with **no request exceeding `embedBatch = 16` texts or `512` token IDs**. Deletion check: remove the batch split or 512-id truncation in `vecembed.go` and confirm the test fails.
    - `TestEmbedLoadsTheExportForItsArch` (`native-helpers.md` §V.11): in `native-run` on every target, verifies that the helper handshake's reported `model_sha256` equals the compiled-in `AssetPin.SHA256[runtime.GOARCH]` pin for that architecture (`amd64` vs `arm64`).
    - `TestEveryOpFitsTheDeadline` (`native-helpers.md` §V.4, gated on `FEOV_NATIVE_MEASURE=1` and listed in `NATIVE_RUN_MUST_PASS` so it runs on `native-run`'s dedicated runners and never on shared-CPU `feov-record` legs): times **spawn to handshake reply** (including hashing the 33.8 MiB `.ort` model and session initialization), then walks the Go client's op table and executes the largest client request for `embed` — **`embedBatch = 16` texts of `512` token IDs each** — asserting both spawn-to-handshake and the `16 × 512`-id request complete in $< 30\text{ s}$ ($\frac{1}{4}$ of the 120 s deadline), and failing if any non-selftest op in the client's op table has no largest-request case.
    - `TestEmbedFidelityMeetsS5` (`native-helpers.md` §V.11, gated on `FEOV_NATIVE_MEASURE=1` and listed in `NATIVE_RUN_MUST_PASS`): embeds the 300 reference texts (`testdata/speed.jsonl`) with the built helper and asserts cosine similarity against `testdata/ref_gte.npy` has **mean $\ge 0.984$ and min $\ge 0.966$**, logging the exact mean and min in the test output across all 5 targets (including inside the Windows `servercore:ltsc2025` container with no Visual Studio and no Go present).
    - `TestEmbedSpeedMeetsS6` (`native-helpers.md` §V.11, gated on `FEOV_NATIVE_MEASURE=1` and listed in `NATIVE_RUN_MUST_PASS`): embeds the 100 200-word chunks (`testdata/len200.jsonl`) single-threaded with the built helper, asserts throughput $\ge 5.0\text{ chunks/s}$ on `linux-amd64`, `linux-arm64`, and `darwin-arm64` (and cold 1-query latency $\le 1.5\text{ s}$), and reports speed on the two Windows targets (`PASS` on all 5 targets, never `t.Skip` when `FEOV_NATIVE_MEASURE=1`).
    - `TestMissingOrMismatchedHelperRefusesByName`: verifies `helper-absent` and `helper-digest-mismatch` refusals.
    - `TestModelIDIncludesTargetOSArch`: verifies `ModelID()` contains `"+target@" + runtime.GOOS + "-" + runtime.GOARCH` and that `ReadVecCache` / `ReadReportVecCache` reject a cache file written with a different target's `ModelID()`. Deletion check: drop the `+target@` suffix from `ModelID()` and the test fails.
- **`plugins/frank-exchange-of-views/tools/internal/hybridsearch/`** **[NEW]**
  - `chunk.go` **[NEW]**: deterministic 200-word / 40-word-overlap (`stride = 160`, `SourceChunkSpec = "w200_s160"`) source chunker preserving byte spans and page numbers, plus report paragraph chunker (`ReportChunkSpec = "para_v1"`) that strips HTML comment anchors (`claimcount.StripAnchors`) from `Text` before BM25/embedding and extracts all anchor IDs (`F-`, `C-`, `K-`, `G-`) into `Anchors []string` and `Anchor string`.
  - `cache.go` **[NEW]**: binary read/write of source vector cache `<run>/cache/<text_sha>.vec` (`FEOVVEC2`, checking both `modelID` and `SourceChunkSpec`) and report paragraph vector cache `<run>/cache/report_vecs.bin` (`FEOVRPT1`, checking both `modelID` and `ReportChunkSpec`, keyed by `sha256(visibleText(para))`).
  - `rrf.go` **[NEW]**: in-memory SQLite FTS5 (`porter unicode61`) BM25 (propagating any SQLite/FTS5 error instead of returning silent empty lexical ranks) + dense cosine ranking + RRF ($k=60$) fusion + structured `SearchResult` with `SearchScope` metadata.
  - `hybridsearch_test.go` **[NEW]**: unit tests for chunking, HTML anchor comment stripping and multi-anchor (`F-`, `C-`, `K-`, `G-`) extraction, `.vec` and `report_vecs.bin` cache invalidation on `modelID` (including `+target@<os>-<arch>`) or `chunkSpec` change, FTS5 error propagation, and RRF fusion invariants.
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
  - [`scripts/releaseassets/releaseassets.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/releaseassets/releaseassets.go), [`scripts/releaseassets/main.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/releaseassets/main.go), & [`releaseassets_test.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/releaseassets/releaseassets_test.go) **[MODIFY]**: verify declared `ASSETS` exports (`gte-small.int8.amd64.ort`, `gte-small.int8.arm.ort`) in `CheckDist` on FEOV tags while skipping `staticproof.Check` on data assets; include `plugins/frank-exchange-of-views/tools/internal/vecembed/` in `IsNativeSetChange`; and ensure `releaseassets -scope` diffs `HEAD^1 HEAD` (once `native/phase-b-fixes` merges, or carried directly) so `native-run` executes with `full=true` on PRs touching the native set.
  - [`.github/workflows/hooks.yml`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/.github/workflows/hooks.yml) **[MODIFY]**:
    - Include `feov-embed` in `native-build` matrix and `native-run` (`go test` includes `./internal/vecembed/...` with `FEOV_NATIVE_MEASURE=1`, `NATIVE_RUN_MUST_PASS` includes `TestTokenizerParity400`, `TestEmbedLoadsTheExportForItsArch`, `TestEmbedFidelityMeetsS5`, `TestEmbedSpeedMeetsS6`, and `TestEveryOpFitsTheDeadline`, and the Windows `servercore:ltsc2025` container step prebuilds and runs `vecembed.test.exe` alongside `tessocr.test.exe`, `fetchcache.test.exe`, and `nativehelper.test.exe` with `-e FEOV_NATIVE_MEASURE=1`, failing the step with `exit 1` instead of `|| true` if any required test fails).
    - In `feov-record` job (`:746-760`), download `native-*-linux-amd64` / `native-*-windows-amd64` (both `feov-ocr` and `feov-embed` artifacts) into `${{ runner.temp }}/helpers` and `chmod +x` both helpers on Linux (`FEOV_NATIVE_MEASURE` stays unset here so S5/S6/deadline measurement tests do not run on shared-CPU `feov-record` legs).
    - In `release` job (`:1988-2002`), strip unsuffixed helper binaries (`feov-ocr`, `feov-ocr.exe`, `feov-embed`, `feov-embed.exe`) and `gte-small.int8.ort` from `$DIST` before `releaseassets -check` and checksum generation.
  - [`scripts/fetchbingen/main.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/fetchbingen/main.go) **[MODIFY]** (and regenerate `plugins/*/hooks/fetch-bin.sh` via `go -C scripts run ./fetchbingen`): check and install `feov-embed` and per-arch model exports alongside other binaries.
  - [`plugins/prosthetic-conscience/tools/internal/doctor/main.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/prosthetic-conscience/tools/internal/doctor/main.go) **[MODIFY]**: check and `--fix` helper and model assets.
  - [`scripts/universe.sh`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/universe.sh) **[MODIFY]**: build `feov-embed` into local release and stage into universe cache.
  - [`scripts/agentgen/src/skills/research-protocol/SKILL.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/agentgen/src/skills/research-protocol/SKILL.md) **[MODIFY]** (and regenerate `plugins/frank-exchange-of-views/skills/research-protocol/SKILL.md` and `plugins/frank-exchange-of-views/agents/*.md` via `go -C scripts run ./agentgen`, plus [`plugins/frank-exchange-of-views/tools/internal/difftest/testdata/manual_runs_every_help_page_live.golden`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/difftest/testdata/manual_runs_every_help_page_live.golden) via `UPDATE_GOLDENS=1`).

---

### 3. Consumer Censuses (Run Against `origin/main` @ `a4bc81c79`)

#### Census A: `flags.Match` and `flags.Quote` Consumers
Command:
```sh
git grep -n "flags\.\(Match\|Quote\)" plugins/frank-exchange-of-views/tools/
```
Output (**43 lines across 12 files** on `origin/main` @ `a4bc81c79`):
- Acting verbs using `flags.Quote` only (`blue/cite.go`: 3, `blue/compute.go`: 3, `blue/edit.go`: 4, `blue/retire.go`: 6, `lens/finding.go`: 3, `lens/mint.go`: 5, `lens/verify.go`: 2, `motion/verbs.go`: 2) — **NO CHANGE**.
- View selector and source search files (`seat/selector.go`: 4, `seat/source.go`: 3, `seat/verbs.go`: 2, `selectorcensus_test.go`: 6) keep `flags.Match` and `flags.Quote` and add `flags.Search` on `report` and `source` — **[MODIFY]**.

#### Census B: `fetchcache` Text Extraction & Search Consumers
Command:
```sh
git grep -n "SourceTextPath\|OCRTextPath\|SearchSource\|foldExtraction\|extractHTML" plugins/frank-exchange-of-views/tools/
```
Output (**54 lines across 20 files** on `origin/main` @ `a4bc81c79`):
- [`internal/cli/seat/source.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat/source.go) (1 line): routes single-source `--match`/`--quote` to `SearchSource`, cross-source `--match`/`--quote` to `SearchCrossSources`, single-source `--search` to `SearchSourceHybrid`, and cross-source `--search` to `SearchCrossSourcesHybrid`. **[MODIFY]**
- [`internal/fetchcache/fetchcache.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/fetchcache.go) (4 lines), [`pageread.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/pageread.go) (3 lines), [`readscanned.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/readscanned.go) (2 lines): call `prewarmSourceVectors(run, entry, origin)` after writing extracted text. **[MODIFY]**
- [`internal/fetchcache/sourcesearch.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/sourcesearch.go) (3 lines) & `sourcesearch_test.go` (10 lines): keep `SearchSource(run, e, match, quote)` signature intact; add `SearchCrossSources`, `SearchSourceHybrid`, `SearchCrossSourcesHybrid`, and `bestChunkSentence`. **[MODIFY]**
- Unchanged callers/tests (31 lines across 14 files: `cli/blue/cite.go`: 1, `cli/fetchsummary.go`: 1, `cli/ocr.go`: 1, `cli/ocrcite_test.go`: 2, `fetchcache/extractor.go`: 2, `fetchcache/htmlextract.go`: 3, `fetchcache/htmlextract_test.go`: 1, `fetchcache/htmlmarkup_test.go`: 7, `fetchcache/htmlpage_test.go`: 1, `fetchcache/pagelocate.go`: 4, `fetchcache/pageread_test.go`: 2, `fetchcache/pagerender.go`: 1, `fetchcache/readscanned_test.go`: 4, `fetchcache/sourcelines_test.go`: 1) — **NO CHANGE**.

#### Census C: Native Helper & Asset Provisioning Consumers
Command:
```sh
git grep -n "feov-ocr\|native/build.sh\|native/PINS.txt\|fetchbingen" .github/ plugins/ scripts/
```
Output (**170 lines across 72 files** on `origin/main` @ `a4bc81c79`):
- [`.github/workflows/hooks.yml`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/.github/workflows/hooks.yml) (10 lines): native build and test jobs (`native-scope`, `native-build`, `native-run`, `feov-record`), release packaging (`release`). **[MODIFY]**
- [`plugins/frank-exchange-of-views/native/build.sh`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/native/build.sh) (11 lines) and `native/feov-embed/` (Part 3 of `plans/native-helpers.md`). **[MODIFY / NEW]**
- [`plugins/frank-exchange-of-views/tools/internal/nativehelper/`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper/client.go) (`client.go`: 2, `digest_gen.go`: 2, `nativesrc/nativesrc.go`: 8, `nativesrc/nativesrc_test.go`: 16) and `devcmd/nativedigest/main.go` (1 line): add `feov-embed` helper and `ASSETS` pins to `digest_gen.go` and `nativesrc` tests. **[MODIFY]**
- [`scripts/releaseassets/releaseassets.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/releaseassets/releaseassets.go) (1 line) and `releaseassets_test.go` (8 lines): check `ASSETS` exports in `CheckDist` (skipping `staticproof` on `.ort` data files) and add `internal/vecembed/` to `IsNativeSetChange`. **[MODIFY]**
- [`scripts/internal/staticproof/staticproof.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/internal/staticproof/staticproof.go) (1 line), `allowlist_test.go` (4 lines), `scripts/staticproof/main.go` (1 line): Windows PE allowlist for `feov-embed`. **[MODIFY]**
- [`scripts/fetchbingen/main.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/fetchbingen/main.go) (12 lines), `main_test.go` (1 line), and generated `plugins/{frank-exchange-of-views,gray-area,prosthetic-conscience}/hooks/fetch-bin.sh` (3 lines across 3 files). **[MODIFY]**
- [`plugins/prosthetic-conscience/tools/internal/doctor/fetchbin_contract_test.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/prosthetic-conscience/tools/internal/doctor/fetchbin_contract_test.go) (5 lines) and `main.go` / `release_contract_test.go`. **[MODIFY]**
- [`plugins/frank-exchange-of-views/tools/internal/difftest/testdata/manual_runs_every_help_page_live.golden`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/difftest/testdata/manual_runs_every_help_page_live.golden) (4 lines) and 10 generated `plugins/frank-exchange-of-views/agents/*.md` files (10 lines): regenerated via `agentgen` and `UPDATE_GOLDENS=1`. **[MODIFY — generated]**
- Unchanged `feov-ocr`-specific files (70 lines across 44 files: `native/feov-ocr/{PINS.txt,main.cpp}`: 4, `cli/{enginestub_test.go,fetch.go,fetchocr_test.go,ocr.go,ocrcite_real_test.go}`: 8, `fetchcache/{blankreceipt_test.go,corpus_test.go,enginestub_test.go,enginesuccess_test.go,pageread.go}`: 7, `seatprobe/{build.go,fixture_test.go,scanned.go,separatedbuild_test.go}`: 5, `tessocr/{client.go,corpus/render.go,diag_test.go,diagcrash_test.go,engine_stub_test.go,engine_test.go,goldens_test.go,reconstruct_test.go,stubbuild_test.go,tessocr.go,textcells_test.go}`: 27, 13 `tessocr/testdata/corpus/*/reading.golden` files: 13, `releasegate/fuzz/hookinvocation_test.go`: 1, `prosthetic-conscience/hooks/README.md`: 2, `scripts/check/gates.go`: 3) — **NO CHANGE**.

#### Census D: Agent-Facing Prose & Constitutions
Command:
```sh
git grep -n "lexical match\|--match\|show source" scripts/agentgen/ plugins/frank-exchange-of-views/skills/ plugins/frank-exchange-of-views/agents/
```
Output (**249 lines across 14 files** on `origin/main` @ `a4bc81c79`):
- [`scripts/agentgen/src/skills/research-protocol/SKILL.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/agentgen/src/skills/research-protocol/SKILL.md#L50) (1 line) and generated `plugins/frank-exchange-of-views/skills/research-protocol/SKILL.md` (1 line): teaches when to use exhaustive selection (`--quote` for literal text, `--match` for regex alternations/tells/boundaries) vs. ranked hybrid search (`--search` for conceptual/wording discovery across the report or all cached sources), stating that ranked results are cut and never prove absence. **[MODIFY]**
- 12 generated `plugins/frank-exchange-of-views/agents/*.md` files (247 lines: `blue-lane.md`: 20, `blue-researcher.md`: 23, `blue-synthesizer.md`: 23, `lead-judge.md`: 17, `red-chair.md`: 17, and 7 `red-lens-*.md` files at 21 each): inlined `manual` blocks regenerated via `go -C scripts run ./agentgen`. **[MODIFY — generated]**

---

## IV. Risk & Mitigation

| Risk | Likelihood | Impact | Complexity | Mitigation |
|---|---|---|---|---|
| **R1: Seats treat a top-$K$ cut hybrid `--search` as proof of absence** | High | High | Low | Every `--search` response (text and JSON) prints `ranked_and_cut: true`, `returned: K`, `searched_chunks: N`, `unsearched_sources: […]`, and states that absence from top-$K$ proves nothing and `--quote` / `--match` are the exhaustive checks. `research-protocol/SKILL.md` states the same duty. |
| **R2: `feov-embed` crashes or overruns deadline during search** | Low | High | Low | The process boundary isolates crashes from `feov-record`. The client enforces a 120 s overall deadline and bounds batches to $\le 16$ texts of $\le 512$ token IDs each, killing the child on failure and spawning a fresh helper on the next request. Verified by `TestEveryOpFitsTheDeadline` (`16 × 512` IDs + spawn-to-handshake $< 30\text{ s}$) and `TestEmbedSplitsAtTheBatch` (1,000 texts). |
| **R3: Missing or mismatched `feov-embed` or model file** | Medium | High | Low | Explicit hard refusal (`helper-absent` or `helper-digest-mismatch`) naming the asset, compiled-in hash, and `/prosthetic-conscience:doctor --fix`. Never falls back silently to BM25-only. |
| **R4: Stale vectors after `ocr read`, `render-page`, chunker parameter edits, or across `amd64`/`arm64` targets** | Low | High | Low | Source vector cache files are addressed by `<text_sha>.vec` (`sha256` of `SourceTextPath` bytes) and check both `modelID` (`"gte-small-int8@<helper_digest>+target@<os>-<arch>"`) and `SourceChunkSpec` (`w200_s160`) in the `FEOVVEC2` header. Report paragraph cache (`report_vecs.bin`) keys paragraphs by `sha256(visibleText(para))` and checks `modelID` (with `+target@<os>-<arch>`) and `ReportChunkSpec` (`para_v1`) in the `FEOVRPT1` header. |
| **R5: Repeated `show report --search` re-embeds unchanged report paragraphs** | High | Medium | Low | `<run>/cache/report_vecs.bin` caches paragraph embeddings by `sha256(visibleText(para))`, so `show report --search` embeds only the query plus paragraphs added or edited since the last search (1–3 texts instead of 50–300). |
| **R6: 8-bit quantization drift vs FP32 reference embeddings** | Low | Medium | Low | `TestEmbedFidelityMeetsS5` gates on mean cosine $\ge 0.984$, min $\ge 0.966$ against `ref_gte.npy` across 300 texts on every release target in `native-run` (S5). |
| **R7: Unpinned build inputs or `arm64` model asset name mismatch in `native/build.sh`** | Medium | High | Low | `native/build.sh` reads only pinned URLs/SHA256s from `native/PINS.txt` and `native/feov-embed/PINS.txt` (no `$ORT_STATIC_DIR`, no scratch paths, pinned `flatbuffers` wheel), and maps `arm64 -> gte-small.int8.arm.ort` matching `PINS.txt` and `ASSETS`. |
| **R8: Protocol pin advance on `show report --search`** | High | High | Low | `serve = false` in `internal/cli/seat/verbs.go:978-985` when `flags.Search` is given, suppressing `record.ServeReport`. Tested in `selectorcensus_test.go`. |
| **R9: Retrieved 200-word source chunk is not directly citeable or report HTML anchors pollute BM25/embeddings** | High | High | Low | `bestChunkSentence` selects the best-matching sentence within each retrieved 200-word chunk (resolving its actual OCR page via `locatePage` and checking `LocateSourceSpan`) while preserving `matched`/`start`/`end`/`line`/`last_line` for the chunk. `ChunkReport` strips HTML comment anchors via `claimcount.StripAnchors` before indexing/embedding and extracts all `F-`/`C-`/`K-`/`G-` anchor IDs into `anchors`. |

---

## V. Verification Plan (the checklist)

> **Standing rule**: Before running any FEOV test or gate from an agent session, run `unset $(env | grep -o '^FEOV_[A-Z_]*')`.

1. **Native Helper Cross-Build & Static Proof Across All 5 Release Targets**
   - Ensure `scripts/releaseassets` diffs `HEAD^1 HEAD` (fixing the `git diff` error drop at `scripts/releaseassets/main.go:98` from `native/phase-b-fixes` if not yet merged on `main`) so `native-scope` emits `full=true` on PRs touching the native set, and run `native/build.sh feov-embed <target>` across all 5 targets (`linux-amd64`, `linux-arm64`, `windows-amd64`, `windows-arm64`, `darwin-arm64`).
   - Run `go -C scripts run ./staticproof -artifact feov-embed <binary>` against each built helper, and run `go -C scripts test -count=1 ./releaseassets/...`.
   - **Pass criterion**: all 5 binaries build from pinned inputs only (verifying `arm64` fetches `gte-small.int8.arm.ort`), pass `staticproof` with 0 forbidden dynamic libraries, and `releaseassets -check` verifies the target-suffixed helpers and declared `ASSETS` exports (`gte-small.int8.amd64.ort`, `gte-small.int8.arm.ort`).
   - *Re-arm when*: `native/feov-embed/**`, `native/frame/**`, `native/build.sh`, or `scripts/releaseassets/**` changes.

2. **Embedder Unit, Tokenizer Parity, Batching, Deadline (§V.4), & Real-Data Embedding Contract (§V.11, S5, S6) (`internal/vecembed`)**
   - Run `FEOV_NATIVE_MEASURE=1 go test -count=1 -v ./internal/vecembed/` locally and across all 5 targets in CI `native-run` (verifying `--- PASS` lines appear for every `NATIVE_RUN_MUST_PASS` test in every target's log, including `vecembed.test.exe` inside the Windows `servercore:ltsc2025` container with no Visual Studio and no Go present, and never `"skipped because full=false"`). On `feov-record` CI legs, `FEOV_NATIVE_MEASURE` is unset so the timing/fidelity measurement tests skip there and never run on shared CPUs.
   - **Pass criteria (`plans/native-helpers.md` §V.4 and §V.11)**:
     - `TestTokenizerParity400`: 400/400 exact token ID match against Hugging Face `thenlper/gte-small` reference tokenizations (`testdata/speed.jsonl` + `testdata/len200.jsonl` vs `testdata/hf_tokens_400.json`).
     - `TestEmbedLoadsTheExportForItsArch` (§V.11): in `native-run` per target, the handshake's `model_sha256` equals the compiled-in `AssetPin` SHA256 for that architecture (`amd64` vs `arm64`).
     - `TestEmbedFidelityMeetsS5` (S5, §V.11, gated on `FEOV_NATIVE_MEASURE=1`, in `NATIVE_RUN_MUST_PASS`): mean cosine $\ge 0.984$, min $\ge 0.966$ against `testdata/ref_gte.npy` across all 300 texts on all 5 targets (and inside the `windows-amd64` `servercore:ltsc2025` container); per-target numbers posted from CI logs into the PR.
     - `TestEmbedSpeedMeetsS6` (S6, §V.11, gated on `FEOV_NATIVE_MEASURE=1`, in `NATIVE_RUN_MUST_PASS`): single-thread throughput $\ge 5.0\text{ chunks/s}$ on 200-word chunks (`testdata/len200.jsonl`) on `linux-amd64`, `linux-arm64`, and `darwin-arm64` (and cold 1-query latency $\le 1.5\text{ s}$), and speed reported on the two Windows targets (`PASS` on all 5 targets); per-target numbers posted from CI logs into the PR.
     - `TestEveryOpFitsTheDeadline` (§V.4, gated on `FEOV_NATIVE_MEASURE=1`, in `NATIVE_RUN_MUST_PASS`): times **spawn to handshake reply** (including hashing the 33.8 MiB model and session creation) and walks the Go client's op table to run the largest client request for `embed` — **`embedBatch = 16` texts of `512` token IDs each** — asserting each completes in $< 30\text{ s}$ ($\frac{1}{4}$ of the 120 s deadline) on every target, and failing if any non-selftest op in the op table lacks a largest-request case.
     - `TestEmbedSplitsAtTheBatch` (§V.4): fake helper records each incoming request; **1,000 texts** (including inputs over 512 tokens) go out as $\lceil 1000 / \text{embedBatch} \rceil = 63$ requests, with **none over `embedBatch = 16` texts or `512` token IDs**. Deletion check: remove the batch split or 512-id cap in `vecembed.go` and confirm the test fails.
     - `TestMissingOrMismatchedHelperRefusesByName`: missing or mismatched `feov-embed` or `gte-small.int8.ort` refuses with `helper-absent` or `helper-digest-mismatch` and names `/prosthetic-conscience:doctor --fix`.
     - `TestModelIDIncludesTargetOSArch`: verifies `ModelID()` includes `"+target@" + runtime.GOOS + "-" + runtime.GOARCH`; deletion check: remove the `+target@` suffix from `ModelID()` and the test fails.
     - **Replay and BEIR hybrid score reproduction through `feov-embed` (§V.11)**: run the 79-row seat replay benchmark and BEIR `SciFact` / `NFCorpus` hybrid evaluations driving `feov-embed` + `gte-small.int8.ort`, verifying they reproduce `ortq/RESULT.md` (`SciFact` `0.731` / `+0.048`, `NFCorpus` `0.362` / `+0.040`, Replay All-79 `Hit@5 = 0.861` / `Hit@10 = 0.924`, Wording-55 `Hit@5 = 0.873` / `Hit@10 = 0.927`).
   - *Re-arm when*: any file under `internal/vecembed/` or `native/feov-embed/` changes.

3. **Hybrid Retrieval, Vector Cache Key (with Target `os-arch`), Report Paragraph Cache, Anchor Stripping, & Best-Sentence Selection Suite (`internal/hybridsearch` & `internal/fetchcache`)**
   - Run `go test -count=1 -v ./internal/hybridsearch/ ./internal/fetchcache/`.
   - **Pass criteria**:
     - `TestVecCacheRejectsChangedModelTargetOrChunkSpec`: changing `modelID` (including changing the `+target@<os>-<arch>` suffix, e.g., `linux-amd64` vs `linux-arm64`) or `chunkSpec` causes `ReadVecCache` and `ReadReportVecCache` to reject the cached file and recompute vectors. Deletion check: strip the `+target@<os>-<arch>` component when building the cache key and the cross-target rejection case fails.
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
   - Build an isolated universe (`scripts/universe.sh build`), verify via `cmp` that the staged marketplace carries the checkout's `feov-embed`, `gte-small.int8.ort`, and updated help text, and run a smoke debate (`§V.13`) that fetches a scanned PDF and runs `--search` with the run's `engine.log` read.
   - **Pass criterion**: `go -C scripts run ./check` exits 0 with 0 failed gates, CI `release-fuzz` (0–3 + tally) and full 5-target `native-run` pass, and an independent review approves before merge.
