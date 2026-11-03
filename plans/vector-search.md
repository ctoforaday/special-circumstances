# Hybrid lexical + vector search over the report and cached sources (`llama.cpp` + `gte-small-Q8_0`)

> STATUS: approved 2026-11-03 — decision document for [#1171](https://github.com/ctoforaday/special-circumstances/issues/1171). Audited: PASS (two consecutive passes: Round 2 PASS, Round 3 confirming PASS). Implementation follows PR creation and merge.

Ruled by `gblock` across 2026-10-31 and 2026-11-01:
1. **Hybrid in one go**: ship SQLite FTS5 BM25 (`porter unicode61`) and dense vector search together, fused via Reciprocal Rank Fusion ($\text{RRF}, k=60$).
2. **Ranked read is distinct from exhaustive `--quote`**: `--quote` stays exhaustive across every view (returning every literal match with zero truncation). The hybrid search is a ranked top-$K$ read over the report and cached sources whose output always states that it is ranked and cut, how many chunks and sources were searched, which cached sources had no extracted text, and that absence from top-$K$ proves nothing.
3. **Statically linked `llama.cpp` C/C++ stack**: use `third_party/pins/build-cstack.sh` (`zig cc` / `zig c++`, 6 release targets) to compile and statically link `llama.cpp` (`libllama.a` + `libggml*.a`) behind the existing `-tags tessocr` C-stack build gate, rather than maintaining custom pure-Go float32 matrix assembly or shipping unquantized 130 MB models.
4. **Model selection by measured quality/size/speed tradeoffs**: pin `ggml-org/gte-small-Q8_0-GGUF` (`gte-small-q8_0.gguf`, `36,685,056` bytes / 35.0 MiB, `Q8_0` quantized 12-layer BERT bi-encoder, 384 dims) as a checksummed release asset provisioned via `hooks/fetch-bin.sh` and `sc-doctor`.

---

## I. Summary & Goals

### 1. Value Case & Problem Statement

Since [#1171](https://github.com/ctoforaday/special-circumstances/issues/1171) was filed on 2026-09-26, three prerequisites shipped on `main`:
- HTML text extraction (`extractHTML` in `internal/fetchcache/htmlextract.go`, PR #1673).
- Single-source literal and regex search (`feov-record show source --sha <sha|--anchor C-…> --quote|--match`, PR #1662).
- Removal of token-Jaccard `lens near-match` (`internal/record/nearmatch.go`, retired 2026-10-24).

What remains broken is **vocabulary mismatch** and **cross-source discovery**:
- Across October 2026 seat transcripts (`~/.claude/scratch/vecplan/replay/SUMMARY.md`), seats executed **1,494** `show report --match|--quote` calls (345 returning zero hits), **133** `show source --match|--quote` calls, **215** `show report | grep` calls, and **2,452** `grep`/`rg` commands over `<run>/cache/`.
- `show source` requires `--sha` or `--anchor` upfront; a seat cannot search across all fetched sources in `<run>/cache/` in a single act.
- When a seat searches for a concept using words that differ from the source's exact phrasing (55 of the 79 verified failed-then-reworded seat searches in `replay.jsonl` are wording misses, and 17 are regex alternation hunts), literal `--quote` and regex `--match` return 0 hits.

### 2. Why Pure-Go Embedders Failed the Pre-Committed Bar

Before selecting the `llama.cpp` path, we wrote a pre-committed pass bar (`~/.claude/scratch/vecplan/bakeoff/BAR.md`) and benchmarked pure-Go candidates on a 4-vCPU `AMD EPYC 7B12` (`AVX2` + `FMA`):
- **Static token-table lookup (`go-potion` + `minishlab/potion-retrieval-32M`)**: fast (`5,519 chunks/s`), 129 MB FP32 weights, but has no transformer layers (embeds by averaging static token vectors). On BEIR `SciFact`, standalone `potion-retrieval-32M` (`0.633` nDCG@10) loses to stemmed BM25 (`0.682`), and its hybrid RRF gain over BM25 is `+0.018` (`95% CI [-0.005, +0.041]`, crossing zero and failing the quality bar). Furthermore, `go-potion` rewrites its model file on load (`os.Create(modelPath)`).
- **Pure-Go BERT (`rcarmo/gte-go` + `thenlper/gte-small`)**: passes the quality bar (`+0.055` hybrid nDCG@10 on `SciFact`, `+0.042` on `NFCorpus`), but runs FP32 naive 3-nested-loop matrix multiplication in pure Go at **1.23 chunks/s** (`812 ms/chunk`, 4× below the 5.0 chunks/s bar), and its upstream `amd64` AVX2 assembly (`gebpKernelAVX2`) segfaults at `DX=0` due to an uninitialized pointer register while computing $C \mathrel{+}= A \times B$ instead of $C = A \times B$.

### 3. Model Selection & Quality Tradeoffs

We evaluated six embedding models and three SQLite FTS5 BM25 tokenizers across two gold-labeled scientific retrieval benchmarks (**BEIR SciFact**: 300 queries, 5,183 docs; **BEIR NFCorpus**: 323 queries, 3,633 docs) and the **79-row FEOV seat replay dataset** (50 report rows, 29 source rows; 55 wording misses; 200-word source chunks with 40-word stride; RRF fusion $k=60$ with `bm25/porter unicode61`). Raw artifacts and scripts live in `~/.claude/scratch/vecplan/results/`.

| Model / Method | Architecture & Pooling | GGUF `Q8_0` Size | SciFact nDCG@10 (Vec / Hybrid) | SciFact Hybrid $\Delta$ vs BM25 (95% CI) | NFCorpus nDCG@10 (Vec / Hybrid) | NFCorpus Hybrid $\Delta$ vs BM25 (95% CI) | BEIR Mean Hybrid nDCG@10 | Replay All-79 Hybrid (`h@1` / `h@5` / `h@10` / `MRR`) |
|---|---|---|---|---|---|---|---|---|
| `bm25/porter unicode61` | Lexical FTS5 | 0 MiB | — / `0.682` | baseline | — / `0.322` | baseline | `0.5020` | `0.544` / `0.848` / `0.873` / `0.670` |
| **`ggml-org/gte-small-Q8_0-GGUF`** (`thenlper/gte-small`) | 12L BERT, 384d, **Mean pool**, no prefix | **35.0 MiB** | **`0.727` / `0.738`** | **`+0.055 [+0.033, +0.077]`** | **`0.348`** / `0.364` | **`+0.042 [+0.028, +0.056]`** | **`0.5506`** | `0.519` / `0.835` / **`0.937`** / `0.659` |
| `ggml-org/bge-small-en-v1.5-Q8_0-GGUF` (`BAAI/bge-small-en-v1.5`) | 12L BERT, 384d, CLS pool, query prefix req. | **35.0 MiB** | `0.713` / `0.730` | `+0.047 [+0.027, +0.067]` | `0.344` / **`0.369`** | **`+0.047 [+0.033, +0.062]`** | `0.5491` | **`0.570`** / `0.835` / `0.911` / **`0.681`** |
| `all-MiniLM-L6-v2` (Q8\_0) | 6L BERT, 384d, Mean pool | ~23 MiB | `0.645` / `0.719` | `+0.036 [+0.014, +0.058]` | `0.316` / `0.356` | `+0.034 [+0.020, +0.048]` | `0.5373` | `0.443` / `0.848` / `0.949` / `0.604` |
| `snowflake-arctic-embed-xs` | 6L BERT, 384d, CLS pool, prefix req. | ~23 MiB | `0.645` / `0.706` | `+0.024 [+0.002, +0.045]` | `0.309` / `0.347` | `+0.025 [+0.015, +0.036]` | `0.5266` | `0.481` / `0.785` / `0.886` / `0.616` |
| `potion-retrieval-32M` | Static token table, 512d | 129 MB (FP32) | `0.633` / `0.700` | `+0.018 [-0.005, +0.041]` | `0.308` / `0.341` | `+0.019 [+0.008, +0.030]` | `0.5205` | `0.506` / `0.899` / `0.949` / `0.654` |
| `potion-base-8M` | Static token table, 256d | 30 MB (FP32) | `0.506` / `0.619` | `-0.063 [-0.085, -0.042]` | `0.243` / `0.313` | `-0.009 [-0.021, +0.003]` | `0.4660` | `0.406` / `0.886` / `0.938` / `0.595` |

We also evaluated larger official GGUF models (`ggml-org/jina-embeddings-v2-base-en-Q8_0-GGUF` at 139 MiB, `google/embeddinggemma-2` at ~295 MiB Q8\_0 / 300M params, `ggml-org/Nomic-Embed-Text-v2-GGUF` at 488 MiB, and `ggml-org/bge-m3-Q8_0-GGUF` at 605 MiB):
- **Why 110M–560M models (`embeddinggemma-2`, `bge-m3`, `nomic-v2`) are rejected**: they are **4× to 17× larger** (139–605 MiB vs. 35.0 MiB) and **5× to 10× slower** on CPU per chunk (~75–150 ms vs. ~12–18 ms in `llama.cpp` `Q8_0`), pushing a 200-chunk source from ~2.5 s into a 15–30 s wait for marginal retrieval delta on English technical text. In addition, `google/embeddinggemma-2` uses the gated `gemma` license on Hugging Face (requiring authentication tokens, which breaks unauthenticated `PINS.txt` / `fetch-bin.sh` downloads), whereas `gte-small` and `bge-small-en-v1.5` are MIT-licensed.
- **Why `ggml-org/gte-small-Q8_0-GGUF` is chosen over `ggml-org/bge-small-en-v1.5-Q8_0-GGUF`**:
  1. **Only model whose standalone vector 95% CI clears BM25 on both BEIR benchmarks**: `gte-small` standalone beats `bm25/porter` by `+0.045 [+0.013, +0.075]` on `SciFact` and `+0.026 [+0.005, +0.047]` on `NFCorpus`. `bge-small-en-v1.5` standalone on `SciFact` is `+0.030 [-0.003, +0.061]`, whose 95% CI crosses zero.
  2. **Highest BEIR Mean Hybrid nDCG@10 (`0.5506` vs `0.5491`), highest BEIR Recall@100 (`0.950` SciFact, `0.337` NFCorpus vs `0.942` / `0.311`), and highest Replay `Hit@10` (`0.937` vs `0.911`)**: because the ranked read returns the top $K=10$ results, `Hit@10` and `R@100` govern whether the relevant passage reaches the seat at all (`bge-small-en-v1.5` has higher Replay `Hit@1` at `0.570` vs `0.519`, but misses 2 more targets out of top 10).
  3. **Self-contained GGUF metadata ([[facts-are-fields]])**: binary inspection of `gte-small-q8_0.gguf` confirms `general.architecture = "bert"`, `bert.pooling_type = 1` (`LLAMA_POOLING_TYPE_MEAN`), `bert.context_length = 512`, `bert.embedding_length = 384`, `tokenizer.ggml.model = "bert"`, and **symmetric query/document encoding (no magic prompt prefix)**. By contrast, `bge-small-en-v1.5-q8_0.gguf` (`bert.pooling_type = 2`, `LLAMA_POOLING_TYPE_CLS`) requires prepending the Asymmetric Query Prefix `"Represent this sentence for searching relevant passages: "` to queries — a string absent from the GGUF header that must be hardcoded in Go.

### 4. Quantitative Success Criteria

| Gate | Metric | Target |
|---|---|---|
| **G1: Retrieval quality (gold benchmarks)** | Hybrid (`bm25/porter unicode61` + `gte-small-q8_0` RRF $k=60$) nDCG@10 vs `bm25/porter` | 95% paired bootstrap CI lower bound $> 0$ on both `SciFact` (`+0.055 [+0.033, +0.077]`) and `NFCorpus` (`+0.042 [+0.028, +0.056]`) |
| **G2: Embedding fidelity (`Q8_0` CGo vs FP32 reference)** | Cosine similarity between `internal/vecembed` (`llama.cpp` `gte-small-q8_0.gguf`) and `sentence-transformers` FP32 `thenlper/gte-small` over the 150-text replay sample | Mean cosine $\ge 0.990$, 5th-percentile $\ge 0.980$, minimum $\ge 0.950$ |
| **G3: CPU throughput & query latency** | Throughput on 200-word (~290-token) chunks on 4 cores (`AMD EPYC 7B12`); cold single-query latency (mmap + embed 1 query + cosine over 3,000 cached vectors) | Throughput $\ge 40\text{ chunks/s}$ ($\ge 8\times$ the 5.0 chunks/s bar; $\le 5\text{ s}$ per 200-chunk / 1 MB source); cold query $\le 250\text{ ms}$ (bar: $\le 1.5\text{ s}$) |
| **G4: Static binary proof across 6 targets** | `third_party/pins/build-cstack.sh check <binary>` across `linux-{amd64,arm64}`, `windows-{amd64,arm64}`, `darwin-{amd64,arm64}` | 0 forbidden dynamic libraries (`staticproof` green on all 6 release targets, with Windows PE allowlist updated for `llama.cpp` system DLLs `{bcrypt.dll, advapi32.dll, ws2_32.dll}`) |
| **G5: Exhaustive `--quote` preservation** | `TestEverySelectorThatAViewAdvertisesSelects` (`internal/cli/selectorcensus_test.go`) | `--quote` returns 100% of literal matches with 0 truncation on every view |
| **G6: Protocol State Integrity** | `show report --search` does NOT record `ServeReport` event | Lens calling `--search` does not advance reading pin without whole report |

---

## II. Technical Context

### 1. C++ Static Stack (`third_party/pins/` + `internal/vecembed`)

The repository already cross-compiles C++17 `tesseract 5.5.3` and C `leptonica 1.87.0` into fully static release binaries across all 6 targets using `zig 0.17.0` (`zig cc` / `zig c++`) and `cmake 4.4.3` in [`plugins/frank-exchange-of-views/tools/third_party/pins/build-cstack.sh`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/third_party/pins/build-cstack.sh).

We extend that exact mechanism with **two new entries** in [`plugins/frank-exchange-of-views/tools/third_party/pins/PINS.txt`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/third_party/pins/PINS.txt):
1. **`llama.cpp` source tarball**: `ggml-org/llama.cpp` release `v0.8.0` (`https://api.github.com/repos/ggml-org/llama.cpp/tarball/v0.8.0`), built by CMake into static `libllama.a`, `libggml.a`, `libggml-base.a`, and `libggml-cpu.a` in `$PREFIX/lib/` with headers (`llama.h`, `ggml.h`, `ggml-alloc.h`, `ggml-backend.h`, `ggml-cpu.h`, `ggml-opt.h`) installed into `$PREFIX/include/`.
   - CMake flags:
     ```cmake
     -DBUILD_SHARED_LIBS=OFF
     -DLLAMA_BUILD_TESTS=OFF
     -DLLAMA_BUILD_EXAMPLES=OFF
     -DLLAMA_BUILD_TOOLS=OFF
     -DLLAMA_BUILD_SERVER=OFF
     -DLLAMA_BUILD_COMMON=OFF
     -DLLAMA_CURL=OFF
     -DGGML_OPENMP=OFF
     -DGGML_NATIVE=OFF
     -DGGML_LLAMAFILE=OFF
     ```
   - Target SIMD configuration in `build-cstack.sh`:
     - `amd64` (`linux-amd64`, `windows-amd64`, `darwin-amd64`): `-DGGML_AVX=ON -DGGML_AVX2=ON -DGGML_FMA=ON -DGGML_F16C=ON -DGGML_AVX512=OFF` (portable across all modern x86\_64 runners and developer machines, matching our EPYC/GitHub Actions baseline).
     - `arm64` (`linux-arm64`, `windows-arm64`, `darwin-arm64`): ARM NEON + FP16 dotprod enabled natively by `ggml-cpu` when `CMAKE_SYSTEM_PROCESSOR=aarch64`/`arm64`.
2. **Pinned GGUF model**: `gte-small-q8_0.gguf` (`36,685,056` bytes, SHA256 `945a2da32e7bfb2cdb3a3fde50bc40c5d103db84fedc078adf415f3bb10f2ed0`, from `https://huggingface.co/ggml-org/gte-small-Q8_0-GGUF/resolve/1dc93210ef4edf1a611523c8d17544c47564275b/gte-small-q8_0.gguf`).
3. **Staticproof Windows PE Allowlists**:
   - `llama.cpp` links standard Windows OS-provided system DLLs: `bcrypt.dll` (cryptography primitives), `advapi32.dll` (system/token APIs), and `ws2_32.dll` (Winsock symbols present even when server is disabled).
   - In [`plugins/frank-exchange-of-views/tools/third_party/pins/staticproof/main.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/third_party/pins/staticproof/main.go#L109), `peLibAllowed` is extended to permit `{bcrypt.dll, advapi32.dll, ws2_32.dll}` as standard OS-provided libraries, and `staticproof_test.go` updated accordingly.
4. **C-Stack Recipe Digest & Golden Re-stamping**:
   - Editing `PINS.txt` and `build-cstack.sh` updates `cstackRecipeDigest`, which changes `cstackPin` in [`internal/tessocr/tessocr.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr/tessocr.go).
   - The first line of each OCR golden in `internal/tessocr/testdata/corpus/*/reading.golden` records `cstackPin`. Implementing this change requires updating `cstackPin` in `tessocr.go` and re-stamping all 13 corpus goldens via `UPDATE_GOLDENS=1 go test -tags tessocr -count=1 ./internal/tessocr/`.

### 2. Thin CGo Wrapper (`internal/vecembed`)

Following the exact architecture of [`internal/tessocr`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr/tessocr.go), `internal/vecembed` is split into:
- `vecembed.go` (untagged): defines the `Embedder` interface (`Embed(texts []string) ([][]float32, error)`, `ModelID() string`, `Close() error`), `ErrNotCompiledIn`, `ErrModelMissing`, and deterministic L2 normalization + cosine similarity helpers.
- `engine_stub.go` (`//go:build !tessocr || !cgo`): returns `ErrNotCompiledIn` from `New()`, keeping plain `go test ./...` and cross-GOOS `go vet` fast and C-toolchain-free.
- `engine_cgo.go` + `shim.cpp` + `shim.h` (`//go:build tessocr && cgo`): ~90 lines of C/C++ wrapping `llama.cpp`'s C API:
  - `llama_model_load_from_file(model_path, mparams)` with `mparams.use_mmap = true` (zero-copy OS page-cache mapping; `< 2 ms` warm load).
  - `llama_init_from_model(model, cparams)` with `cparams.embeddings = true`, `cparams.n_ctx = 512`, `cparams.n_batch = 512`, `cparams.n_ubatch = 512`, `cparams.pooling_type = LLAMA_POOLING_TYPE_MEAN` (verified against the model's own GGUF header `bert.pooling_type = 1`).
  - `llama_tokenize(vocab, text, text_len, tokens, 512, /*add_special=*/true, /*parse_special=*/false)` — uses `llama.cpp`'s native BERT WordPiece tokenizer directly from the GGUF vocabulary (`30,522` tokens), eliminating `gte-go`'s ASCII-only tokenization bug.
  - Multi-sequence batching via `llama_batch_init(512, 0, n_seq)` + `llama_encode(ctx, batch)` + `llama_get_embeddings_seq(ctx, seq_id)` followed by L2 normalization ($\|\mathbf{v}\|_2 = 1$).

### 3. Model Supply via `fetch-bin.sh`, `sc-doctor`, and `universe.sh`

Per `gblock`'s ruling, `gte-small-q8_0.gguf` (35.0 MiB) is distributed as a **checksummed release asset** rather than embedded via `//go:embed`:
1. **Why a file asset matches `llama.cpp`**: `llama_model_load_from_file` memory-maps the GGUF directly from the OS buffer cache (`use_mmap = true`), so repeated CLI invocations of `feov-record` share the same physical RAM pages with near-zero startup latency. Embedding 35 MiB in Go via `//go:embed` would bloat all 6 platform `feov-record` binaries in every release by 210 MiB and prevent `mmap`.
2. **Where the model file lives**: `${CLAUDE_PLUGIN_ROOT}/bin/gte-small-q8_0.gguf` (beside `feov-record`, resolved via `os.Executable()` or overridden in CI/dev by `FEOV_EMBED_MODEL`, which `build-cstack.sh env` exports pointing to `$PREFIX/share/models/gte-small-q8_0.gguf`).
3. **Provisioning & self-healing**:
   - The release job in [`.github/workflows/hooks.yml`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/.github/workflows/hooks.yml#L2018-L2044) copies the SHA256-verified `gte-small-q8_0.gguf` from `${work}/dl/gte-small-q8_0.gguf` into `${DIST}/gte-small-q8_0.gguf` before running `sha256sum * > SHA256SUMS`.
   - `scripts/fetchbingen/fetch-bin.sh` and `sc-doctor` (`plugins/prosthetic-conscience/tools/internal/doctor/main.go`) verify and install platform-independent data assets listed in `tools/third_party/pins/ASSETS.txt` (`gte-small-q8_0.gguf`) from `$BASE_URL/$tag/$asset` against `SHA256SUMS` into `$ROOT/bin/$asset`.
   - `scripts/universe.sh` copies `gte-small-q8_0.gguf` (from `$FEOV_EMBED_MODEL` or the user's plugin cache) into both the cache root and the staged marketplace's `bin/`.
   - **Refusal contract**: when `gte-small-q8_0.gguf` is missing or its SHA256 disagrees with `PINS.txt`, a hybrid search refuses with an explicit error naming the missing file (`gte-small-q8_0.gguf`), expected SHA256 prefix, and remediation (`run /prosthetic-conscience:doctor --fix`), **never silently degrading to lexical-only BM25**.

### 4. Chunking, Vector Cache, and Hybrid RRF Retrieval

- **Chunking**:
  - **Sources (`<run>/cache/`)**: split extracted text (`SourceTextPath(run, entry, origin)` — `.txt` for PDF/HTML, `.ocr.txt` for OCR) into **200-word chunks with 40-word overlap** (`stride = 160` words), preserving byte offsets `[startByte, endByte)` and (for `ocr_derived` sources) the page number from `LocateSourceSpan`'s page boundary map so every returned chunk carries the exact `source_span` and `page` that `blue cite` and `lens verify` consume.
  - **Report (`blue/report.md`)**: split into paragraphs / logical blocks bounded by section headings (`heading`, line range `start_line..end_line`, and verbatim text). Because a report is typically 50–300 paragraphs (2–15 KB), it is embedded and FTS5-indexed in memory at query time in $< 80\text{ ms}$ with nothing persisted on disk.
- **Content-addressed vector cache for sources (`<run>/cache/<text_sha>.vec`)**:
  - Keyed by `text_sha = sha256(extracted_text_bytes)` + `model_id` (`gte-small-q8_0@945a2da32e7b`) + `chunk_spec` (`w200_s160`).
  - Why `text_sha` and NOT `entry.Sha` (raw body SHA): when `ocr read` or `render-page` re-derives `<sha>.ocr.txt`, `text_sha` changes automatically, making stale vectors impossible without invalidation hooks (`pagerender.go:342`, `readscanned.go:148`).
  - Computed eagerly at the end of `foldExtraction` / `ocr read` (and lazily on first `show source` search if absent, e.g., on a migrated or archived run). At `llama.cpp` `Q8_0` speed ($\ge 40\text{ chunks/s}$), a typical 40-chunk web article embeds in $< 1\text{ s}$ and a 200-chunk paper in $< 4\text{ s}$.
- **Hybrid Fusion (`BM25` + Cosine via RRF)**:
  - **Lexical arm**: ephemeral in-memory SQLite (`:memory:`) ` USING fts5(text, tokenize='porter unicode61')`. Query terms are tokenized into an FTS5 `OR` query of stemmed tokens (`bm25(chunks)` rank).
  - **Vector arm**: exact inner product (cosine similarity on L2-normalized `[]float32` vectors) over all chunks in scope.
  - **Fusion**: Reciprocal Rank Fusion with standard constant $k = 60$:
    $$\text{score}(d) = \frac{1}{60 + \text{rank}_{\text{bm25}}(d)} + \frac{1}{60 + \text{rank}_{\text{vec}}(d)}$$
    where a chunk absent from the BM25 match set contributes only its vector term $\frac{1}{60 + \text{rank}_{\text{vec}}(d)}$.

---

### 5. Explicit Response Schemas for Report and Source Search

To prevent ambiguous parsing by seats and downstream tools, each view and scoping mode has a strictly defined root JSON schema:

#### A. Cross-Source Ranked Hybrid Search (`show source --search <query>`)
Triggered when neither `--sha` nor `--anchor` is passed. Searches across all cached sources in `<run>/cache/`.
Every hit in `hits` carries its source identity (`sha256` and `url`), byte coordinates, and citeable `source_span` so it is immediately citeable via `blue cite --url <url> --source-span "<source_span>"`.
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
      "matched": "exact matched chunk text",
      "sentence": "the full sentence surrounding the match",
      "sentence_start": 105,
      "sentence_end": 182,
      "source_span": "the full sentence surrounding the match",
      "not_citeable": "",
      "start": 120,
      "end": 145
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
    "notice": "Ranked top-K hybrid (BM25 + vector) results; this list is cut and absence proves nothing. For an exhaustive literal match set, use --quote."
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
      "matched": "...",
      "sentence": "...",
      "sentence_start": 105,
      "sentence_end": 182,
      "source_span": "...",
      "not_citeable": "",
      "start": 120,
      "end": 145
    }
  ],
  "search_scope": {
    "ranked_and_cut": true,
    "returned": 5,
    "searched_chunks": 32,
    "notice": "Ranked top-K hybrid (BM25 + vector) results; this list is cut and absence proves nothing. For an exhaustive literal match set, use --quote."
  }
}
```

#### C. Cross-Source Literal Search (`show source --quote <text>`)
Triggered when neither `--sha` nor `--anchor` is passed. Exhaustive literal substring match across all cached sources in `<run>/cache/`.
```json
{
  "hits": [
    {
      "sha256": "4a7b9c1d...",
      "url": "https://example.org/article.html",
      "source_text_origin": "embedded",
      "text_sha": "e3b0c442...",
      "page": 0,
      "matched": "exact literal string",
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
    "criterion": "literal: \"exact literal string\"",
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

#### D. Single-Source Literal Search (`show source --sha <sha|--anchor C-…> --quote <text>`)
Existing `SourceSearch` shape (`internal/fetchcache/sourcesearch.go`), unchanged. Returns every match with `selection.matched = N`.

#### E. Report Ranked Hybrid Search (`show report --search <query>`)
Searches the single report document (`blue/report.md`). Paragraph chunks. Has NO `searched_sources` or `unsearched_sources`:
```json
{
  "chunks": [
    {
      "heading": "## Research areas",
      "start_line": 42,
      "end_line": 56,
      "text": "The verified prime test ...",
      "score": 0.0325,
      "anchor": "C-abc12345"
    }
  ],
  "search_scope": {
    "ranked_and_cut": true,
    "returned": 10,
    "searched_chunks": 48,
    "notice": "Ranked top-K hybrid (BM25 + vector) results; this list is cut and absence proves nothing. For an exhaustive literal match set, use --quote."
  }
}
```

---

## III. Proposed Changes (the spec)

### 1. Surface & Vocabulary Design (Complete the Concept)

Per [`internal/flags/names.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags/names.go#L3-L17) (*"Two meanings for one word is a defect... and two words for one meaning is the same defect"*):
- **`--quote` stays unchanged everywhere**: on every `show` and `inquest` view, and on `show source`, `--quote` continues to select/return **every** literal substring match with zero truncation.
- **`show source` scoping flags (`--sha` and `--anchor`)**:
  - `show source` has always accepted `--sha` or `--anchor` (it has never accepted `--url`).
  - When `--sha` or `--anchor` is passed: scopes the search to that single source.
  - When neither `--sha` nor `--anchor` is passed: executes **cross-source search** across all cached sources in `<run>/cache/`.
  - When neither source scoping nor pattern flags are passed: refuses with an explicit usage message: `"name what to find: --search (ranked hybrid) or --quote (exhaustive literal), and optionally scope to one source with --sha or --anchor"`.
- **`--search` [NEW] on `show report` and `show source`; `--match` [DELETE]**:
  - `gblock` ruled (2026-10-18, `regexes-are-questioned.md`; 2026-10-31, `CHECKPOINT.md`): replace regex `--match` with hybrid search, while keeping `--quote` as the exhaustive literal match set.
  - Why `--match` is deleted rather than kept on `board`/`changes`: in October transcripts, seats ran only 4 content searches on `board` and 7 on `changes` (nearly all ID lookups or literal substring checks that `--quote` already serves). Keeping RE2 `--match` on half the views and hybrid ranked search on the other half under the same flag name violates `flags/names.go` and breaks `TestEverySelectorThatAViewAdvertisesSelects` (`selectorcensus_test.go:45`), which asserts that a selector flag matching `"zzqqxx-matches-no-entry"` returns **0 entries** (a ranked vector search always returns top-$K$ nearest neighbors even when no literal word matches).
  - Therefore:
    1. **Every view** (`show` and `inquest`, except `telemetry`) keeps `--quote` as its exhaustive literal selector (`AddSelectorFlags` registers `--quote`).
    2. **`show report` and `show source`** gain `--search <query>` (`flags.Search = "search"`): ranked top-$K$ hybrid (BM25 + vector) retrieval.
    3. **`flags.Match` (`--match`) is deleted** from `internal/flags/names.go`, `internal/cli/seat/selector.go`, `internal/cli/seat/source.go`, `internal/cli/seat/verbs.go`, and `internal/cli/selectorcensus_test.go`.
- **Mutual Exclusion Contract**:
  - `--search` and `--quote` on `show report` and `show source` are **strictly mutually exclusive**.
  - Passing both on the same command returns `feov.Validation` error:
    `"--search and --quote are mutually exclusive: --quote returns every literal match; --search returns a cut ranked list. Pass one"`
  - On `show report`: `--search` is also refused beside `--anchor` and beside `--assembled`. `--quote` is refused beside `--search` and beside `--anchor`.
- **Protocol State & Pin Protection (`ServeReport`)**:
  - In [`internal/cli/seat/verbs.go:1007-1014`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat/verbs.go#L1007-L1014), the whole-report serving check for lens seats:
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
  - A lens invoking `show report --search` reads a cut ranked list of chunks, not the whole report; setting `serve = false` guarantees that `record.ServeReport` is NOT logged and the lens's reading pin does NOT advance.
- **Terms Registry Update**:
  - Update `internal/terms/terms.json:568` definition of "source search" to:
    `"A source search is a read of cached sources' text for literal text or a ranked hybrid query (BM25 and dense vectors), scoped to one source or across all cached sources in the run; each hit carries the source's sentence, its URL and SHA256, and whether a citation can quote that sentence as printed, and the search records nothing."`
  - Re-run `go -C scripts run ./vocabdoc` to update vocabulary documentation.
- **Release Fuzz Updates**:
  - Update call sites in `releasegate/fuzz/` (`fuzz_test.go:937`, `5171`, `6468`, `7717`, `parts_test.go:213`) from `--match` to `--quote` and `--search`.

---

### 2. Component Breakdown & File Inventory

- **`plugins/frank-exchange-of-views/tools/third_party/pins/`**
  - [`PINS.txt`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/third_party/pins/PINS.txt) **[MODIFY]**: add `llama-b0.8.0.tar.gz` (`https://api.github.com/repos/ggml-org/llama.cpp/tarball/v0.8.0`) and `gte-small-q8_0.gguf` (`945a2da32e7bfb2cdb3a3fde50bc40c5d103db84fedc078adf415f3bb10f2ed0`).
  - `ASSETS.txt` **[NEW]**: single-line manifest (`gte-small-q8_0.gguf`) declaring platform-independent release assets that `hooks.yml` publishes into `SHA256SUMS` and that `fetch-bin.sh` / `sc-doctor` / `universe.sh` install into `${CLAUDE_PLUGIN_ROOT}/bin/`.
  - [`build-cstack.sh`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/third_party/pins/build-cstack.sh) **[MODIFY]**: build `llama.cpp` static archives (`libllama.a`, `libggml.a`, `libggml-base.a`, `libggml-cpu.a`), install headers into `$PREFIX/include`, stage `gte-small-q8_0.gguf` into `$PREFIX/share/models/gte-small-q8_0.gguf`, and export `FEOV_EMBED_MODEL` in `build-cstack.sh env`.
  - [`staticproof/main.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/third_party/pins/staticproof/main.go) & `staticproof_test.go` **[MODIFY]**: extend `peLibAllowed` to accept `{bcrypt.dll, advapi32.dll, ws2_32.dll}` as standard Windows OS libraries.
- **`plugins/frank-exchange-of-views/tools/internal/tessocr/`**
  - [`tessocr.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr/tessocr.go) **[MODIFY]**: update `cstackPin` to match the new `cstackRecipeDigest` resulting from `PINS.txt` and `build-cstack.sh` edits.
  - `testdata/corpus/*/reading.golden` **[MODIFY]**: re-stamp all 13 goldens with `UPDATE_GOLDENS=1 go test -tags tessocr -count=1 ./internal/tessocr/`.
- **`plugins/frank-exchange-of-views/tools/internal/vecembed/`** **[NEW]**
  - `vecembed.go` **[NEW]**: `Embedder` interface, `Default() (Embedder, error)`, model path resolution (`FEOV_EMBED_MODEL` override, else `filepath.Join(filepath.Dir(os.Executable()), "gte-small-q8_0.gguf")`), SHA256 pin verification (`945a2da32e7b…`), L2 normalization, and `Cosine(a, b []float32) float32`.
  - `engine_stub.go` **[NEW]** (`//go:build !tessocr || !cgo`): returns `ErrNotCompiledIn`.
  - `engine_cgo.go`, `shim.h`, `shim.cpp` **[NEW]** (`//go:build tessocr && cgo`): CGo bridge to `libllama.a` + `libggml*.a`.
  - `vecembed_test.go` & `engine_cgo_test.go` **[NEW]**: unit tests, `PINS.txt` hash agreement test (`TestModelPinMatchesPINS`), determinism test, and `Q8_0` vs FP32 reference cosine fidelity test (`TestGteSmallQ8Fidelity`).
- **`plugins/frank-exchange-of-views/tools/internal/hybridsearch/`** **[NEW]**
  - `chunk.go` **[NEW]**: deterministic 200-word / 40-word-stride source chunker preserving byte spans and page numbers, plus report paragraph/section chunker.
  - `cache.go` **[NEW]**: binary read/write of `<run>/cache/<text_sha>.vec` (magic header, model identity hash, chunk count, byte ranges, and `float32` vectors).
  - `rrf.go` **[NEW]**: in-memory SQLite FTS5 (`porter unicode61`) BM25 + dense cosine ranking + RRF ($k=60$) fusion + structured `SearchResult` with `SearchScope` metadata.
  - `hybridsearch_test.go` **[NEW]**: unit and replay benchmark tests verifying BM25, vector, and RRF fusion invariants.
- **`plugins/frank-exchange-of-views/tools/internal/fetchcache/`**
  - [`sourcesearch.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/sourcesearch.go) **[MODIFY]**:
    - Remove regex `match` parameter from `SearchSource(run, e, quote)`.
    - Add `SearchCrossSourcesQuote(run, quote)`: exhaustive literal search returning `CrossSourceSearch`.
    - Add `SearchSourceHybrid(run, e, query, topK)` and `SearchCrossSourcesHybrid(run, query, topK)`: ranked hybrid search returning `SourceSearch` or `CrossSourceSearch` with `SearchScope`.
  - [`fetchcache.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/fetchcache.go), [`pageread.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/pageread.go), [`readscanned.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/readscanned.go) **[MODIFY]**: pre-warm `<text_sha>.vec` when an embedder is compiled in and available.
- **`plugins/frank-exchange-of-views/tools/internal/flags/` & `internal/cli/`**
  - [`internal/flags/names.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags/names.go) **[MODIFY]**: replace `Match = "match"` with `Search = "search"`; update `flags.All()`.
  - [`internal/cli/seat/selector.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat/selector.go) **[MODIFY]**: `AddSelectorFlags` registers `--quote` (literal exhaustive selection); remove RE2 `--match` path; add ranked hybrid `--search` handler for `show report`.
  - [`internal/cli/seat/source.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat/source.go) **[MODIFY]**: register `--search` alongside `--quote`; allow omitting `--sha`/`--anchor` when `--search` or `--quote` is given (cross-source search over all cached sources); refuse bare `show source` with no flags; enforce mutual exclusion between `--search` and `--quote`.
  - [`internal/cli/seat/verbs.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat/verbs.go) **[MODIFY]**: update `report` and `changes` flag help and `reportSelectorBound` for `--quote` and `--search`; suppress `ServeReport` (`serve = false`) when `flags.Search` is given; remove `--match` residual strings.
  - [`internal/cli/selectorcensus_test.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/selectorcensus_test.go) **[MODIFY]**: hold `--quote` across every view in `selectorViews(t)` and add `TestRankedSearchStatesCutAndScope` for `show report --search` and `show source --search`.
- **`plugins/frank-exchange-of-views/tools/internal/terms/`**
  - [`terms.json`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/terms/terms.json) **[MODIFY]**: update definition of "source search"; run `go -C scripts run ./vocabdoc`.
- **`plugins/frank-exchange-of-views/tools/releasegate/fuzz/`**
  - [`fuzz_test.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/releasegate/fuzz/fuzz_test.go) & `parts_test.go` **[MODIFY]**: replace `--match` invocations with `--search` or `--quote`.
- **Release, Provisioning & Agent Surfaces**
  - [`.github/workflows/hooks.yml`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/.github/workflows/hooks.yml) **[MODIFY]**: stage `third_party/pins/ASSETS.txt` files into `${DIST}/` in the release job before `sha256sum * > SHA256SUMS`; run `go test -tags tessocr ./internal/vecembed/ ./internal/hybridsearch/` in the `cstack` CI leg.
  - [`scripts/fetchbingen/fetch-bin.sh`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/fetchbingen/fetch-bin.sh) **[MODIFY]** (and regenerate `plugins/*/hooks/fetch-bin.sh` via `go -C scripts run ./fetchbingen`): check and install `tools/third_party/pins/ASSETS.txt` entries alongside `tools/cmd/*/` binaries.
  - [`plugins/prosthetic-conscience/tools/internal/doctor/main.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/prosthetic-conscience/tools/internal/doctor/main.go) **[MODIFY]**: check and `--fix` plugin data assets from `ASSETS.txt`.
  - [`scripts/universe.sh`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/universe.sh) **[MODIFY]**: mirror `ASSETS.txt` files into cache and staged marketplace `bin/`.
  - [`scripts/agentgen/src/`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/agentgen) **[MODIFY]** (and regenerate `plugins/frank-exchange-of-views/agents/*.md` and `skills/*/SKILL.md` via `go -C scripts run ./agentgen`, plus goldens via `UPDATE_GOLDENS=1`).

---

### 3. Consumer Censuses (Run Against `origin/main` @ `a7485a7ee`)

#### Census A: `flags.Match` and `flags.Quote` Consumers
Command:
```sh
git grep -n "flags\.\(Match\|Quote\)" plugins/frank-exchange-of-views/tools/
```
Output (42 lines across 12 files):
```text
plugins/frank-exchange-of-views/tools/internal/cli/blue/cite.go:176:	flags.Text(c, flags.Quote, flags.DescQuote+". A mis-quote is rejected rather than guessed at")
plugins/frank-exchange-of-views/tools/internal/cli/blue/cite.go:179:	seat.Require(c, flags.Quote, flags.URL, flags.Title)
plugins/frank-exchange-of-views/tools/internal/cli/blue/cite.go:198:	{Flag: flags.Quote, Field: "location"},
plugins/frank-exchange-of-views/tools/internal/cli/blue/compute.go:216:	flags.Text(c, flags.Quote, flags.DescQuoteAlone)
plugins/frank-exchange-of-views/tools/internal/cli/blue/compute.go:218:	seat.Require(c, flags.Quote, flags.Dir)
plugins/frank-exchange-of-views/tools/internal/cli/blue/compute.go:266:	{Flag: flags.Quote, Field: "location"},
plugins/frank-exchange-of-views/tools/internal/cli/blue/edit.go:55:		oldStr := seat.Str(cmd, flags.Quote)
plugins/frank-exchange-of-views/tools/internal/cli/blue/edit.go:238:	flags.Text(c, flags.Quote, flags.DescQuote+". An anchor standing at its end stays where it is unless you also move it")
plugins/frank-exchange-of-views/tools/internal/cli/blue/edit.go:239:	seat.RequiredUnless(c, flags.Quote, "unless --accept")
plugins/frank-exchange-of-views/tools/internal/cli/blue/edit.go:301:		asserted = append(asserted, seat.Asserted{Flag: flags.Quote, Field: "old"}, seat.Asserted{Flag: flags.New, Field: "new"})
plugins/frank-exchange-of-views/tools/internal/cli/blue/retire.go:60:			Claim:        proto.String(seat.Str(cmd, flags.Quote)),
plugins/frank-exchange-of-views/tools/internal/cli/blue/retire.go:79:		claim := seat.Str(cmd, flags.Quote)
plugins/frank-exchange-of-views/tools/internal/cli/blue/retire.go:81:		if !cmd.Flags().Changed(flags.Quote) && len(named.Value()) > 0 {
plugins/frank-exchange-of-views/tools/internal/cli/blue/retire.go:149:		return retireResult{Claim: seat.Str(cmd, flags.Quote), Anchors: body.Anchors, Stayed: stayed}, nil
plugins/frank-exchange-of-views/tools/internal/cli/blue/retire.go:159:	flags.Text(c, flags.Quote, "the claim being removed, verbatim as it stood in the report BEFORE the edit that took it out; optional when --anchor is given")
plugins/frank-exchange-of-views/tools/internal/cli/blue/retire.go:161:	seat.RequiredUnless(c, flags.Quote, "unless --anchor alone retires a citation")
plugins/frank-exchange-of-views/tools/internal/cli/lens/finding.go:141:	flags.Text(c, flags.Quote, flags.DescQuote+". The finding anchor is placed there")
plugins/frank-exchange-of-views/tools/internal/cli/lens/finding.go:142:	seat.RequiredUnless(c, flags.Quote, "unless --about-kind/--about name the subject")
plugins/frank-exchange-of-views/tools/internal/cli/lens/finding.go:154:	{Flag: flags.Quote, Field: "location"},
plugins/frank-exchange-of-views/tools/internal/cli/lens/mint.go:142:				return verifiedFix(p, report, gapID, seat.Str(cmd, flags.Quote), fixNew)
plugins/frank-exchange-of-views/tools/internal/cli/lens/mint.go:208:			return requireNotEstopped(l, seat.Str(cmd, flags.Quote), supersedes.Value())
plugins/frank-exchange-of-views/tools/internal/cli/lens/mint.go:267:	flags.Text(c, flags.Quote, "the ONE sentence the defect lives at — "+flags.DescQuoteAlone+". The tool places the gap's anchor there")
plugins/frank-exchange-of-views/tools/internal/cli/lens/mint.go:268:	seat.RequiredUnless(c, flags.Quote, "unless --about-kind/--about name the subject")
plugins/frank-exchange-of-views/tools/internal/cli/lens/mint.go:394:	{Flag: flags.Quote, Field: "location"},
plugins/frank-exchange-of-views/tools/internal/cli/lens/verify.go:291:	flags.Text(c, flags.Quote, flags.DescQuote+". It is the REPORT's sentence under every --as value, never a span from the source")
plugins/frank-exchange-of-views/tools/internal/cli/motion/verbs.go:148:				if seat.Given(cmd, flags.Quote) {
plugins/frank-exchange-of-views/tools/internal/cli/motion/verbs.go:231:		flags.Text(c, flags.Quote, "on a gap the bench has ruled closed, the other way to give the new ground")
plugins/frank-exchange-of-views/tools/internal/cli/seat/selector.go:175:	flags.Text(c, flags.Match,
plugins/frank-exchange-of-views/tools/internal/cli/seat/selector.go:177:	flags.Text(c, flags.Quote,
plugins/frank-exchange-of-views/tools/internal/cli/seat/selector.go:184:	rx, _ := c.Flags().GetString(flags.Match)
plugins/frank-exchange-of-views/tools/internal/cli/seat/selector.go:185:	lit, _ := c.Flags().GetString(flags.Quote)
plugins/frank-exchange-of-views/tools/internal/cli/seat/source.go:49:	c.Flags().Lookup(flags.Match).Usage = "find every place the source's text matches this `regex` (RE2, case-insensitive by default)"
plugins/frank-exchange-of-views/tools/internal/cli/seat/source.go:50:	c.Flags().Lookup(flags.Quote).Usage = "find every place the source's text holds this `text` LITERALLY"
plugins/frank-exchange-of-views/tools/internal/cli/seat/source.go:74:	match, quote := Str(cmd, flags.Match), Str(cmd, flags.Quote)
plugins/frank-exchange-of-views/tools/internal/cli/seat/verbs.go:690:			for _, name := range []string{flags.Match, flags.Quote} {
plugins/frank-exchange-of-views/tools/internal/cli/seat/verbs.go:1380:	return anchor.Replace(Str(cmd, flags.Quote), func(string, string) string { return "" })
plugins/frank-exchange-of-views/tools/internal/cli/selectorcensus_test.go:45:			for _, flag := range []string{"--" + flags.Match, "--" + flags.Quote} {
plugins/frank-exchange-of-views/tools/internal/cli/selectorcensus_test.go:69:				out, err := read("--json", "--"+flags.Quote, lit)
plugins/frank-exchange-of-views/tools/internal/cli/selectorcensus_test.go:90:			for _, flag := range []string{"--" + flags.Match, "--" + flags.Quote} {
plugins/frank-exchange-of-views/tools/internal/cli/selectorcensus_test.go:103:						if hit, herr := read("--"+flags.Quote, lit); herr != nil || len(arrayEntries(t, hit)) == 0 {
plugins/frank-exchange-of-views/tools/internal/cli/selectorcensus_test.go:152:			if v.Flags().Lookup(flags.Match) == nil && v.Flags().Lookup(flags.Quote) == nil {
plugins/frank-exchange-of-views/tools/internal/cli/selectorcensus_test.go:155:			if v.Flags().Lookup(flags.Match) == nil || v.Flags().Lookup(flags.Quote) == nil {
```
All acting verbs (`cite`, `compute`, `edit`, `retire`, `finding`, `mint`, `verify`, `motion`) use `flags.Quote` only — **NO CHANGE**. The 4 selector/source files (`selector.go`, `source.go`, `verbs.go`, `selectorcensus_test.go`) plus `flags/names.go` and CLI tests are **[MODIFY]**.

#### Census B: `fetchcache` Text Extraction & Search Consumers
Command:
```sh
git grep -n "SourceTextPath\|OCRTextPath\|SearchSource\|foldExtraction\|extractHTML" plugins/frank-exchange-of-views/tools/
```
Output (23 lines across 11 files):
```text
plugins/frank-exchange-of-views/tools/internal/cli/blue/cite.go:269:				flags.SourceSpan, fetchcache.SourceTextPath(run, entry, origin))
plugins/frank-exchange-of-views/tools/internal/cli/fetchsummary.go:166:	s.TextPath = fetchcache.OCRTextPath(run, r.Sha)
plugins/frank-exchange-of-views/tools/internal/cli/ocr.go:299:		TextPath: fetchcache.OCRTextPath(run, sha), TextSha: r.TextSha,
plugins/frank-exchange-of-views/tools/internal/cli/seat/source.go:88:	res, err := fetchcache.SearchSource(run, e, match, quote)
plugins/frank-exchange-of-views/tools/internal/cli/seat/source.go:95:					flags.Quote, flags.Match, fetchcache.SourceTextPath(s.Run(), entry, origin))
plugins/frank-exchange-of-views/tools/internal/fetchcache/extractor.go:91:		return extractHTML(body)
plugins/frank-exchange-of-views/tools/internal/fetchcache/fetchcache.go:597:		if err := foldExtraction(run, &entry, ex); err != nil {
plugins/frank-exchange-of-views/tools/internal/fetchcache/fetchcache.go:773:		if err := foldExtraction(run, &entry, ex); err != nil {
plugins/frank-exchange-of-views/tools/internal/fetchcache/fetchcache.go:796:func foldExtraction(run record.Run, entry *Entry, ex Extraction) error {
plugins/frank-exchange-of-views/tools/internal/fetchcache/htmlextract.go:28:func extractHTML(body []byte) Extraction {
plugins/frank-exchange-of-views/tools/internal/fetchcache/pagelocate.go:177:		ErrSpanAbsent, SourceTextPath(run, e, origin), SpanNormalisation)
plugins/frank-exchange-of-views/tools/internal/fetchcache/pagelocate.go:445:// SourceTextPath is the file a seat reads a cached source's text in, for an entry whose
plugins/frank-exchange-of-views/tools/internal/fetchcache/pagelocate.go:448:func SourceTextPath(run record.Run, e Entry, origin recordpb.SourceTextOrigin) string {
plugins/frank-exchange-of-views/tools/internal/fetchcache/pagelocate.go:451:		return OCRTextPath(run, e.Sha)
plugins/frank-exchange-of-views/tools/internal/fetchcache/pageread.go:477:// OCRTextPath is the assembled reading, beside the extraction TextPath would have written.
plugins/frank-exchange-of-views/tools/internal/fetchcache/pageread.go:480:func OCRTextPath(run record.Run, sha string) string { return Path(run, sha) + ".ocr.txt" }
plugins/frank-exchange-of-views/tools/internal/fetchcache/pageread.go:523:	if err := writeReplacing(OCRTextPath(run, sha), []byte(text)); err != nil {
plugins/frank-exchange-of-views/tools/internal/fetchcache/pagerender.go:342:	if err := os.Remove(OCRTextPath(run, sha)); err != nil && !os.IsNotExist(err) {
plugins/frank-exchange-of-views/tools/internal/fetchcache/readscanned.go:148:	if err := os.Remove(OCRTextPath(run, sha)); err != nil && !os.IsNotExist(err) {
plugins/frank-exchange-of-views/tools/internal/fetchcache/readscanned.go:222:	if err := writeReplacing(OCRTextPath(run, sha), []byte(text)); err != nil {
plugins/frank-exchange-of-views/tools/internal/fetchcache/sourcesearch.go:83:func SearchSource(run record.Run, e Entry, match, quote string) (SourceSearch, error) {
plugins/frank-exchange-of-views/tools/internal/fetchcache/sourcesearch.go:100:	raw, err := os.ReadFile(SourceTextPath(run, e, origin))
plugins/frank-exchange-of-views/tools/internal/fetchcache/sourcesearch_test.go:40:	got, err := SearchSource(run, e, `.{0,50}composite.{0,20}`, "")
```

#### Census C: C-Stack, Release Asset, and `fetch-bin.sh` Consumers
Command:
```sh
git grep -n "build-cstack\|PINS\.txt\|fetchbingen" .github/ plugins/ scripts/
```
Consumers:
- [`.github/workflows/hooks.yml`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/.github/workflows/hooks.yml): lines `347-366` (OCR corpus job), `652-703` (`feov` C-stack test leg), `2018-2044` (release cross-compile + `SHA256SUMS`). **[MODIFY]**
- [`plugins/frank-exchange-of-views/tools/third_party/pins/PINS.txt`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/third_party/pins/PINS.txt) and [`build-cstack.sh`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/third_party/pins/build-cstack.sh). **[MODIFY]**
- [`scripts/fetchbingen/fetch-bin.sh`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/fetchbingen/fetch-bin.sh) and [`scripts/fetchbingen/main.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/fetchbingen/main.go) (generates `plugins/frank-exchange-of-views/hooks/fetch-bin.sh`, `plugins/gray-area/hooks/fetch-bin.sh`, `plugins/prosthetic-conscience/hooks/fetch-bin.sh`). **[MODIFY]**
- [`plugins/prosthetic-conscience/tools/internal/doctor/main.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/prosthetic-conscience/tools/internal/doctor/main.go), [`fetchbin_contract_test.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/prosthetic-conscience/tools/internal/doctor/fetchbin_contract_test.go), [`release_contract_test.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/prosthetic-conscience/tools/internal/doctor/release_contract_test.go). **[MODIFY]**
- [`scripts/universe.sh`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/universe.sh): lines `378-453` (binary & asset staging into cache and staged marketplace). **[MODIFY]**

#### Census D: Agent-Facing Prose & Constitutions
Command:
```sh
git grep -n "lexical match\|--match\|show source" scripts/agentgen/ plugins/frank-exchange-of-views/skills/ plugins/frank-exchange-of-views/agents/
```
Consumers:
- [`scripts/agentgen/src/skills/research-protocol/SKILL.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/agentgen/src/skills/research-protocol/SKILL.md#L50) (and generated `plugins/frank-exchange-of-views/skills/research-protocol/SKILL.md:52`): updates the guidance on when to use exhaustive literal selection vs. ranked hybrid search, stating that ranked results are cut and never prove absence. **[MODIFY]**
- [`plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js#L176): comment-only reference (`// told to use them, and a seat forbidden grep through the shell then has no search at all`); seat prompts obey `TestNoPromptGrowsItsCommandCatalogue` and `TestTheSeatPromptsNameNoVerb` (never naming CLI flags or verbs). **NO CHANGE to prompt strings**.
- `plugins/frank-exchange-of-views/agents/*.md`: inlined `manual` blocks regenerated via `go -C scripts run ./agentgen`. **[MODIFY — generated]**

---

## IV. Risk & Mitigation

| Risk | Likelihood | Impact | Complexity | Mitigation |
|---|---|---|---|---|
| **R1: Seats treat a 0-hit or top-$K$ cut hybrid search as proof of absence** | High | High | Low | Every `--search` response (text and JSON) prints `ranked_and_cut: true`, `returned: K`, `searched_chunks: N`, `unsearched_sources: […]`, and states that absence from top-$K$ proves nothing and `--quote` is the exhaustive literal check. `research-protocol/SKILL.md` states the same duty. |
| **R2: `ggml-cpu` SIMD flags fail to cross-compile on `darwin-{amd64,arm64}` or `windows-arm64` under `zig cc`** | Medium | Medium | Medium | Set `-DGGML_NATIVE=OFF -DGGML_OPENMP=OFF -DGGML_LLAMAFILE=OFF` across all targets; pass explicit AVX2/FMA flags only on `amd64` and rely on standard ARMv8 NEON on `arm64`. Verified across all 6 targets by `build-cstack.sh` + `staticproof` before merging. |
| **R3: Missing `gte-small-q8_0.gguf` on an unprovisioned machine silently degrades search quality** | Medium | High | Low | Hard refusal (`ErrModelMissing`) naming `gte-small-q8_0.gguf`, its pinned SHA256 prefix, and `/prosthetic-conscience:doctor --fix`. Never falls back silently to BM25-only. |
| **R4: Stale vectors after `ocr read` or `render-page` updates `<sha>.ocr.txt`** | Low | High | Low | Vector cache files are addressed by `<text_sha>.vec` (`sha256` of the extracted text bytes), never by `<entry.Sha>.vec`. Any change to `.txt` or `.ocr.txt` produces a new `text_sha` and misses the old `.vec` file automatically. |
| **R5: `Q8_0` quantization drift vs FP32 reference embeddings** | Low | Medium | Low | `TestGteSmallQ8Fidelity` (`internal/vecembed/engine_cgo_test.go`) gates on mean cosine $\ge 0.990$, 5th-percentile $\ge 0.980$, and min $\ge 0.950$ against precomputed FP32 reference vectors on real chunks. |
| **R6: C-stack build time increases in CI** | High | Low | Low | `llama.cpp` with `LLAMA_BUILD_TOOLS=OFF`, `LLAMA_BUILD_SERVER=OFF`, `LLAMA_BUILD_COMMON=OFF` compiles only `ggml` + `llama` (~35 C/C++ translation units, adding ~45 s to a cold `build-cstack.sh` build) and is cached behind the `PINS.txt` + `build-cstack.sh` key on PRs. |
| **R7: Windows `staticproof` fails on `llama.cpp` system DLL imports** | Medium | High | Low | Extend `peLibAllowed` in `plugins/frank-exchange-of-views/tools/third_party/pins/staticproof/main.go` and `staticproof_test.go` to permit `{bcrypt.dll, advapi32.dll, ws2_32.dll}` as standard Windows OS-provided system libraries. |
| **R8: Protocol pin advance on `show report --search`** | High | High | Low | Ensure `serve = false` in `internal/cli/seat/verbs.go:1007-1014` when `flags.Search` is given, suppressing `record.ServeReport`. Tested in `selectorcensus_test.go`. |

---

## V. Verification Plan (the checklist)

> **Standing rule**: Before running any FEOV test or gate from an agent session, run `unset $(env | grep -o '^FEOV_[A-Z_]*')`.

1. **C-Stack Cross-Build & Static Proof (all 6 release targets)**
   - Run `third_party/pins/build-cstack.sh <target> ~/.claude/scratch/cstack-vec` for `linux-amd64`, `linux-arm64`, `windows-amd64`, `windows-arm64`, `darwin-amd64`, and `darwin-arm64`.
   - Cross-compile `feov-record` with `-tags tessocr` for each target and run `third_party/pins/build-cstack.sh check <binary>`.
   - **Pass criterion**: all 6 binaries build and pass `staticproof` with 0 forbidden dynamic libraries (verifying `peLibAllowed` accepts `{bcrypt.dll, advapi32.dll, ws2_32.dll}`).
   - Update `cstackPin` in `internal/tessocr/tessocr.go` and run `UPDATE_GOLDENS=1 go test -tags tessocr -count=1 ./internal/tessocr/` to re-stamp all 13 corpus goldens.
   - *Re-arm when*: `PINS.txt`, `build-cstack.sh`, `staticproof/main.go`, or `internal/vecembed/shim.*` changes.

2. **Embedder Unit, Pin, & Fidelity Suite (`internal/vecembed`)**
   - Run `eval "$(./third_party/pins/build-cstack.sh env linux-amd64 ~/.claude/scratch/cstack-vec)" && go test -tags tessocr -count=1 -v ./internal/vecembed/`.
   - **Pass criteria**:
     - `TestModelPinMatchesPINS`: `gte-small-q8_0.gguf` SHA256 equals `945a2da32e7bfb2cdb3a3fde50bc40c5d103db84fedc078adf415f3bb10f2ed0` in both `vecembed` and `PINS.txt`.
     - `TestGteSmallQ8Fidelity`: mean cosine $\ge 0.990$, 5th-percentile $\ge 0.980$, min $\ge 0.950$ against FP32 `thenlper/gte-small` reference embeddings; non-ASCII text (`naïve`, `𝐶`, `—`) tokenizes cleanly without spuriously splitting into `[UNK]`.
     - `TestEmbedThroughput`: $\ge 40\text{ chunks/s}$ on 200-word chunks on 4 cores; cold single-query latency $\le 250\text{ ms}$.
     - `TestMissingModelRefusesByName`: when `FEOV_EMBED_MODEL` points to a missing or corrupt file, `New()` refuses naming `gte-small-q8_0.gguf` and `/prosthetic-conscience:doctor --fix`.
   - *Re-arm when*: any file under `internal/vecembed/` or `third_party/pins/` changes.

3. **Hybrid Retrieval & Replay Suite (`internal/hybridsearch` & `internal/fetchcache`)**
   - Run `go test -tags tessocr -count=1 -v ./internal/hybridsearch/ ./internal/fetchcache/`.
   - **Pass criteria**:
     - Hybrid RRF ($k=60$) over the checked-in replay fixture achieves `Hit@10` $\ge 0.93$ and beats standalone vector search at `Hit@5` by $\ge 0.20$.
     - Modifying a cached source's `.ocr.txt` changes its `text_sha` and computes fresh vectors without reading the prior `.vec` file.
     - Searching a run where one cached PDF is `image_only` (no `.ocr.txt`) lists that source in `unsearched_sources` alongside the ranked hits from extracted sources.
     - Cross-source search hits carry `sha256` and `url` and cite cleanly via `blue cite --url <url> --source-span "<source_span>"`.
   - *Re-arm when*: any file under `internal/hybridsearch/` or `internal/fetchcache/` changes.

4. **CLI Surface, Selector Census, & Provisioning Contract**
   - Run `go test -tags tessocr -count=1 -v ./internal/cli/... ./internal/flags/... ./internal/terms/...` and `go -C ../../prosthetic-conscience/tools test -count=1 -v ./internal/doctor/...`.
   - **Pass criteria**:
     - `TestEverySelectorThatAViewAdvertisesSelects` passes for `--quote` across every view in `selectorViews(t)`.
     - `TestRankedSearchStatesCutAndScope` verifies `show report --search` and `show source --search` (single-source and cross-source) emit `ranked_and_cut`, `searched_chunks`, `unsearched_sources`, and citeable `source_span` + `page`.
     - `TestSearchDoesNotAdvanceServeReportPin` verifies that `show report --search` does not record `ServeReport` or advance the reading pin for lens seats.
     - `TestSearchAndQuoteMutuallyExclusive` verifies that passing both `--search` and `--quote` refuses with `feov.Validation`.
     - `fetchbin_contract_test.go` and `release_contract_test.go` verify `gte-small-q8_0.gguf` is checksummed in `SHA256SUMS` and installed by `fetch-bin.sh` and `sc-doctor --fix`.
   - *Re-arm when*: any file under `internal/cli/`, `internal/flags/`, `internal/terms/`, `scripts/fetchbingen/`, or `internal/doctor/` changes.

5. **Generators, Repository Gates, & End-to-End Smoke Verification**
   - Regenerate carriers (`go -C scripts run ./fetchbingen`, `go -C scripts run ./agentgen`, `go -C scripts run ./vocabdoc`, `UPDATE_GOLDENS=1`), read every changed golden line and `--help` diff one command at a time, and run `go -C scripts run ./check`.
   - Build an isolated universe (`scripts/universe.sh build`), verify the staged marketplace carries the new `--search` help text and `gte-small-q8_0.gguf`, and drive `feov-record show report --search` and `feov-record show source --search` against an archived real-source run (`run-archive/`).
   - Drive `releasegate/fuzz` tests (`FEOV_RELEASE_GATE=1 go test ./releasegate/fuzz/`) with `--match` retired.
   - **Pass criterion**: `go -C scripts run ./check` exits 0 with 0 failed gates, and the end-to-end CLI invocation against real cached sources returns citeable spans that `blue cite --source-span` accepts.
