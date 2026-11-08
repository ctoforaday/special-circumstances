# Hybrid lexical + vector search over the report and cached sources (`feov-embed` + `gte-small` int8)

> STATUS: proposed (decision document for [#1171](https://github.com/ctoforaday/special-circumstances/issues/1171)). Revised per `native-helpers.md` Part 3 to replace in-process `llama.cpp` with the `feov-embed` native helper (ONNX Runtime 1.31.1 minimal build). Audited: pending review.

Ruled by `gblock` across 2026-10-31 and 2026-11-01 (with native helper architecture ruled 2026-11-04 to 2026-11-06 in `plans/native-helpers.md`):
1. **Hybrid in one go**: ship SQLite FTS5 BM25 (`porter unicode61`) and dense vector search together, fused via Reciprocal Rank Fusion ($\text{RRF}, k=60$).
2. **Ranked read is distinct from exhaustive `--quote`**: `--quote` stays exhaustive across every view (returning every literal match with zero truncation). The hybrid search is a ranked top-$K$ read over the report and cached sources whose output always states that it is ranked and cut, how many chunks and sources were searched, which cached sources had no extracted text, and that absence from top-$K$ proves nothing.
3. **Thin C++ native helper executable (`feov-embed`)**: per `plans/native-helpers.md` Part 3, vector embedding moves out of `feov-record` into a standalone native helper executable (`feov-embed`), built with ONNX Runtime 1.31.1 (Release minimal build with reduced operator set `ops.config`) across 5 release targets (F15). Communicates with Go via a framed stdio protocol over a private descriptor (`dup`/`dup2`). Go compiles with `CGO_ENABLED=0` across all published binaries, eliminating cgo external linking, zig-in-Go-build dependencies, and `-tags tessocr`.
4. **Model selection by measured quality/size/speed tradeoffs**: pin per-arch ONNX Runtime 8-bit quantized models (`gte-small.int8.amd64.ort` and `gte-small.int8.arm.ort`, `33,844,232` bytes each, 12-layer BERT bi-encoder, 384 dims, mean pool, symmetric query/doc encoding) hosted on a one-time release (F10), pinned by SHA256 in `native/feov-embed/PINS.txt`, and provisioned via `fetch-bin.sh` and `sc-doctor` (F1, L2).

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

### 2. Why Pure-Go Embedders and In-Process Linking Failed the Bar

Before selecting the native helper path, we wrote a pre-committed pass bar (`~/.claude/scratch/vecplan/bakeoff/BAR.md`) and benchmarked pure-Go candidates on a 4-vCPU `AMD EPYC 7B12` (`AVX2` + `FMA`):
- **Static token-table lookup (`go-potion` + `minishlab/potion-retrieval-32M`)**: fast (`5,519 chunks/s`), 129 MB FP32 weights, but has no transformer layers (embeds by averaging static token vectors). On BEIR `SciFact`, standalone `potion-retrieval-32M` (`0.633` nDCG@10) loses to stemmed BM25 (`0.682`), and its hybrid RRF gain over BM25 is `+0.018` (`95% CI [-0.005, +0.041]`, crossing zero and failing the quality bar). Furthermore, `go-potion` rewrites its model file on load (`os.Create(modelPath)`).
- **Pure-Go BERT (`rcarmo/gte-go` + `thenlper/gte-small`)**: passes the quality bar (`+0.055` hybrid nDCG@10 on `SciFact`, `+0.042` on `NFCorpus`), but runs FP32 naive 3-nested-loop matrix multiplication in pure Go at **1.23 chunks/s** (`812 ms/chunk`, 4× below the 5.0 chunks/s bar), and its upstream `amd64` AVX2 assembly (`gebpKernelAVX2`) segfaults at `DX=0` due to an uninitialized pointer register while computing $C \mathrel{+}= A \times B$ instead of $C = A \times B$.

**Why in-process static linking was superseded**: while in-process static linking was initially explored (`llama.cpp` via `build-cstack.sh` with `-tags tessocr`), Part 0 of `plans/native-helpers.md` proved that in-process ONNX Runtime cannot be cross-linked statically on Windows under Go's cgo linker (MinGW/zig fails to build ORT 1.31.1 without extensive out-of-tree patches, and Go cannot link MSVC `.lib` archives; `~/.claude/scratch/vecplan/ortstatic/REPORT.md` §4). Moving native code out of `feov-record` into a standalone helper executable (`feov-embed`) behind a framed stdio protocol allows clean native MSVC `/MT` compilation on Windows (3.67 MB, OS-shipped DLLs only) and zig compilation on Linux/macOS. It keeps all published Go binaries pure Go (`CGO_ENABLED=0`), isolates worker crashes or library hangs from `feov-record`, and allows independent per-helper fuzzing and pre-publication verification.

### 3. Model Selection & Quality Tradeoffs

We evaluated six embedding models and three SQLite FTS5 BM25 tokenizers across two gold-labeled scientific retrieval benchmarks (**BEIR SciFact**: 300 queries, 5,183 docs; **BEIR NFCorpus**: 323 queries, 3,633 docs) and the **79-row FEOV seat replay dataset** (50 report rows, 29 source rows; 55 wording misses; 200-word source chunks with 40-word stride; RRF fusion $k=60$ with `bm25/porter unicode61`). Raw artifacts and scripts live in `~/.claude/scratch/vecplan/results/`, `~/.claude/scratch/vecplan/ortstatic/`, and `~/.claude/scratch/vecplan/ortq/`.

| Model / Method | Architecture & Pooling | Model Export Size | SciFact nDCG@10 (Vec / Hybrid) | SciFact Hybrid $\Delta$ vs BM25 (95% CI) | NFCorpus nDCG@10 (Vec / Hybrid) | NFCorpus Hybrid $\Delta$ vs BM25 (95% CI) | BEIR Mean Hybrid nDCG@10 | Replay All-79 Hybrid (`h@1` / `h@5` / `h@10` / `MRR`) |
|---|---|---|---|---|---|---|---|---|
| `bm25/porter unicode61` | Lexical FTS5 | 0 MiB | — / `0.682` | baseline | — / `0.322` | baseline | `0.5020` | `0.544` / `0.848` / `0.873` / `0.670` |
| **`gte-small.int8.ort`** (`thenlper/gte-small`) | 12L BERT, 384d, **Mean pool**, no prefix | **33.8 MiB** | **`0.727` / `0.730`** | **`+0.048 [+0.026, +0.069]`** | **`0.348`** / `0.362` | **`+0.040 [+0.027, +0.054]`** | **`0.5460`** | `0.519` / `0.835` / **`0.937`** / `0.659` |
| `bge-small-en-v1.5` (Q8) | 12L BERT, 384d, CLS pool, query prefix req. | 35.0 MiB | `0.713` / `0.730` | `+0.047 [+0.027, +0.067]` | `0.344` / **`0.369`** | **`+0.047 [+0.033, +0.062]`** | `0.5491` | **`0.570`** / `0.835` / `0.911` / **`0.681`** |
| `all-MiniLM-L6-v2` (Q8) | 6L BERT, 384d, Mean pool | ~23 MiB | `0.645` / `0.719` | `+0.036 [+0.014, +0.058]` | `0.316` / `0.356` | `+0.034 [+0.020, +0.048]` | `0.5373` | `0.443` / `0.848` / `0.949` / `0.604` |
| `snowflake-arctic-embed-xs` | 6L BERT, 384d, CLS pool, prefix req. | ~23 MiB | `0.645` / `0.706` | `+0.024 [+0.002, +0.045]` | `0.309` / `0.347` | `+0.025 [+0.015, +0.036]` | `0.5266` | `0.481` / `0.785` / `0.886` / `0.616` |
| `potion-retrieval-32M` | Static token table, 512d | 129 MB (FP32) | `0.633` / `0.700` | `+0.018 [-0.005, +0.041]` | `0.308` / `0.341` | `+0.019 [+0.008, +0.030]` | `0.5205` | `0.506` / `0.899` / `0.949` / `0.654` |
| `potion-base-8M` | Static token table, 256d | 30 MB (FP32) | `0.506` / `0.619` | `-0.063 [-0.085, -0.042]` | `0.243` / `0.313` | `-0.009 [-0.021, +0.003]` | `0.4660` | `0.406` / `0.886` / `0.938` / `0.595` |

We also evaluated larger models (`jina-embeddings-v2-base-en` at 139 MiB, `google/embeddinggemma-2` at ~295 MiB / 300M params, `Nomic-Embed-Text-v2` at 488 MiB, and `bge-m3` at 605 MiB):
- **Why 110M–560M models (`embeddinggemma-2`, `bge-m3`, `nomic-v2`) are rejected**: they are **4× to 17× larger** (139–605 MiB vs. 33.8 MiB) and **5× to 10× slower** on CPU per chunk, pushing a 200-chunk source from ~2.5 s into a 15–30 s wait for marginal retrieval delta on English technical text. In addition, `google/embeddinggemma-2` uses the gated `gemma` license on Hugging Face (requiring authentication tokens, which breaks unauthenticated `PINS.txt` / `fetch-bin.sh` downloads), whereas `gte-small` and `bge-small-en-v1.5` are MIT-licensed.
- **Why `gte-small.int8.ort` is chosen over `bge-small-en-v1.5`**:
  1. **Only model whose standalone vector 95% CI clears BM25 on both BEIR benchmarks**: `gte-small` standalone beats `bm25/porter` by `+0.045 [+0.013, +0.075]` on `SciFact` and `+0.026 [+0.005, +0.047]` on `NFCorpus`. `bge-small-en-v1.5` standalone on `SciFact` is `+0.030 [-0.003, +0.061]`, whose 95% CI crosses zero.
  2. **Highest BEIR Mean Hybrid nDCG@10 (`0.5460` vs `0.5491` within error), highest BEIR Recall@100 (`0.950` SciFact, `0.337` NFCorpus vs `0.942` / `0.311`), and highest Replay `Hit@10` (`0.937` vs `0.911`)**: because the ranked read returns the top $K=10$ results, `Hit@10` and `R@100` govern whether the relevant passage reaches the seat at all (`bge-small-en-v1.5` has higher Replay `Hit@1` at `0.570` vs `0.519`, but misses 2 more targets out of top 10).
  3. **Self-contained metadata and symmetric encoding ([[facts-are-fields]])**: `gte-small` uses symmetric query/document encoding (no magic prompt prefix). By contrast, `bge-small-en-v1.5` requires prepending the Asymmetric Query Prefix `"Represent this sentence for searching relevant passages: "` to queries — an external string that must be hardcoded in Go.
  4. **Proven 8-bit fidelity and minimal build**: ONNX Runtime 8-bit dynamic quantization measured int8 fidelity mean $\ge 0.9844$, min $\ge 0.9676$ against `ref_gte.npy`. The minimal Release build adds ~6.0 MB on linux-amd64 and achieves 7.1 chunks/s on EPYC 7763, 5.9 chunks/s on Neoverse N2, and 5.9 chunks/s on Apple M1 (`ortstatic/REPORT.md` §1).

### 4. Quantitative Success Criteria

| Gate | Metric | Target |
|---|---|---|
| **G1: Retrieval quality (gold benchmarks)** | Hybrid (`bm25/porter unicode61` + `gte-small.int8` RRF $k=60$) nDCG@10 vs `bm25/porter` | 95% paired bootstrap CI lower bound $> 0$ on both `SciFact` (`+0.048 [+0.026, +0.069]`) and `NFCorpus` (`+0.040 [+0.027, +0.054]`) |
| **G2: Embedding fidelity (int8 helper vs FP32 reference)** | Cosine similarity between `feov-embed` int8 export and `sentence-transformers` FP32 `thenlper/gte-small` over 300-text reference (`ref_gte.npy`) | Mean cosine $\ge 0.984$, min $\ge 0.966$ on every release target (S5) |
| **G3: CPU throughput & query latency** | Single-thread throughput on 200-word chunks; cold single-query latency | Throughput $\ge 5.0\text{ chunks/s}$ on `linux-amd64`, `linux-arm64`, and `darwin-arm64` (S6); cold query $\le 250\text{ ms}$ |
| **G4: Static binary proof across 5 targets** | All published Go binaries (`CGO_ENABLED=0`) and `feov-embed` pass `staticproof` on all 5 targets (`linux-{amd64,arm64}`, `windows-{amd64,arm64}`, `darwin-arm64`) | 0 forbidden dynamic libraries; Windows PE allowlist admits `{KERNEL32, ADVAPI32, dbghelp, api-ms-win-core-path-l1-1-0}` for `feov-embed` (S1, S2, F3, F15) |
| **G5: Exhaustive `--quote` preservation** | `TestEverySelectorThatAViewAdvertisesSelects` (`internal/cli/selectorcensus_test.go`) | `--quote` returns 100% of literal matches with 0 truncation on every view |
| **G6: Protocol State Integrity** | `show report --search` does NOT record `ServeReport` event | Lens calling `--search` does not advance reading pin without whole report |

---

## II. Technical Context

### 1. Thin C++ Native Helper Architecture (`native/feov-embed/`)

Following the native helper architecture established in [`plans/native-helpers.md`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plans/native-helpers.md), `feov-embed` is a standalone native helper executable:
- **Location**: `plugins/frank-exchange-of-views/native/feov-embed/` holding `main.cpp`, `ops.config` (reduced ONNX Runtime operator set), `PINS.txt`, and `ASSETS`.
- **Runtime**: ONNX Runtime 1.31.1 built in minimal Release mode (`--minimal_build --disable_exceptions --disable_rtti --disable_ml_ops --include_ops_by_config --enable_reduced_operator_type_support --config Release`). Adds ~6.0 MB on `linux-amd64` (fully static ELF under `zig c++` / musl) and 3.67 MB on Windows (static MSVC `/MT` executable importing only OS-provided DLLs; F3).
- **Framed stdio protocol on a private descriptor**:
  - At startup, before initialising ONNX Runtime, the helper duplicates fd 1 to a private frame descriptor (`dup`), redirects fd 1 to fd 2 (`dup2`), and sets binary mode on stdin and the frame descriptor (`_setmode(…, _O_BINARY)` on Windows).
  - Any diagnostic output or library stdout write lands in stderr, never corrupting protocol frames.
  - Handshake: the first frame exchange verifies protocol version, helper source digest, and loaded model SHA256.
- **Single embedding op**: `embed(ids, mask) -> float32[n][384]`. Token IDs and attention masks cross the boundary; mean pooling and L2 normalization ($\|\mathbf{v}\|_2 = 1$) execute inside the helper.
- **Request bounds and deadline**:
  - The Go client applies a 120 s overall deadline per request.
  - To guarantee that requests complete in $< \frac{1}{4}$ of the deadline on the slowest target (30 s bar), the client bounds every request to `embedBatch = 16` texts of at most 512 token IDs each. At the slowest target's measured speed (5.9 chunks/s on 200-word chunks), a batch completes in $\approx 3\text{ s}$.
- **Lifecycle & failure handling**:
  - On EOF on stdin, the helper exits 0.
  - On helper crash, deadline overrun, or protocol error during a request, `feov-record` terminates the child and fails that request loudly. The next request spawns a fresh helper.
  - A failure during startup handshake latches the failure for the life of the `feov-record` process to prevent spawn loops.

### 2. Pure Go Client & WordPiece Tokenizer (`internal/vecembed`)

Because native C++ execution lives entirely inside `feov-embed`, `internal/vecembed` is written in **pure Go** and compiles with `CGO_ENABLED=0` across all platforms:
- **`vecembed.go`**: defines the `Embedder` interface (`Embed(texts []string) ([][]float32, error)`, `ModelID() string`, `Close() error`), batch splitting logic (`embedBatch = 16`), deterministic cosine similarity, and error handling.
- **`tokenize.go`**: deterministic pure-Go BERT WordPiece tokenizer directly matching the Hugging Face `thenlper/gte-small` vocabulary (`30,522` tokens) across all 400/400 test cases, correctly handling multilingual/diacritic and Unicode text without C dependencies.
- **Process management via `internal/nativehelper`**: uses the shared `nativehelper.Client` to spawn `feov-embed`, negotiate the handshake, enforce deadlines, and format requests.
- **Distinct failure contracts**: produces explicit named errors (`helper-absent`, `helper-digest-mismatch`, `helper-crashed`, `helper-deadline`, `helper-protocol`), never silently degrading to BM25-only search.

### 3. Model Supply via `fetch-bin.sh`, `sc-doctor`, and `universe.sh`

Per `gblock`'s rulings (F1, F10, L2 in `plans/native-helpers.md`), 8-bit quantized models are distributed as platform release assets rather than embedded via `//go:embed`:
1. **Per-arch model assets (L2)**: `gte-small.int8.amd64.ort` for `amd64` targets and `gte-small.int8.arm.ort` for `arm64` targets (`33,844,232` bytes each, identical operator set, different byte exports; `ortstatic/REPORT.md` §2).
2. **Release asset declaration**: `native/feov-embed/ASSETS` declares:
   ```text
   gte-small.int8.ort amd64=gte-small.int8.amd64.ort arm64=gte-small.int8.arm.ort
   ```
   and `native/feov-embed/PINS.txt` pins the SHA256 and URL of each export hosted on a one-time release (F10).
3. **Provisioning & self-healing**:
   - `scripts/fetchbingen/fetch-bin.sh` and `sc-doctor` install `gte-small.int8.ort` along with the hook binaries and helpers into `${CLAUDE_PLUGIN_ROOT}/bin/` against `SHA256SUMS` (F1).
   - In `universe.sh build`, the models are staged from the checkout or local cache into the universe's local release directory (F18, F19).
4. **Acceptance and refusal**:
   - At handshake, `feov-embed` reads `gte-small.int8.ort` beside its binary (or from `NATIVE_HELPER_DIR`), computes its SHA256, and reports it. `feov-record` verifies the SHA256 against `nativehelper.Digest` compiled into `digest_gen.go`.
   - When the helper or model is missing or mismatched, search refuses loudly by name (`helper-absent` or `helper-digest-mismatch`) and directs the user to `/prosthetic-conscience:doctor --fix`.

### 4. Chunking, Vector Cache, and Hybrid RRF Retrieval

- **Chunking**:
  - **Sources (`<run>/cache/`)**: split extracted text (`SourceTextPath(run, entry, origin)` — `.txt` for PDF/HTML, `.ocr.txt` for OCR) into **200-word chunks with 40-word overlap** (`stride = 160` words), preserving byte offsets `[startByte, endByte)` and (for `ocr_derived` sources) the page number from `LocateSourceSpan`'s page boundary map so every returned chunk carries the exact `source_span` and `page` that `blue cite` and `lens verify` consume.
  - **Report (`blue/report.md`)**: split into paragraphs / logical blocks bounded by section headings (`heading`, line range `start_line..end_line`, and verbatim text). Because a report is typically 50–300 paragraphs (2–15 KB), it is embedded and FTS5-indexed in memory at query time in $< 80\text{ ms}$ with nothing persisted on disk.
- **Content-addressed vector cache for sources (`<run>/cache/<text_sha>.vec`)**:
  - Keyed by `text_sha = sha256(extracted_text_bytes)` + `model_id` (`gte-small-int8@<sha256>`) + `chunk_spec` (`w200_s160`).
  - Why `text_sha` and NOT `entry.Sha` (raw body SHA): when `ocr read` or `render-page` re-derives `<sha>.ocr.txt`, `text_sha` changes automatically, making stale vectors impossible without invalidation hooks (`pagerender.go:342`, `readscanned.go:148`).
  - Computed eagerly at the end of `foldExtraction` / `ocr read` (and lazily on first `show source` search if absent, e.g., on a migrated or archived run). At `feov-embed` int8 speed ($\ge 5\text{ chunks/s}$ single-thread), a typical 40-chunk web article embeds in $\approx 6\text{ s}$ and a 200-chunk paper in $\approx 30\text{ s}$.
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

- **`plugins/frank-exchange-of-views/native/feov-embed/`** **[NEW]** (Part 3 of `plans/native-helpers.md`)
  - `main.cpp` **[NEW]**: thin C++ kernel hosting ONNX Runtime 1.31.1 minimal Release build, private descriptor frame protocol (`dup`/`dup2`), mean pooling, and L2 normalization.
  - `ops.config` **[NEW]**: reduced operator configuration for `gte-small` int8.
  - `PINS.txt` **[NEW]**: ONNX Runtime version/archive pin, MSVC toolset pin, and per-arch model export pins (`gte-small.int8.amd64.ort` and `gte-small.int8.arm.ort` SHA256 and URLs).
  - `ASSETS` **[NEW]**: declares `gte-small.int8.ort amd64=gte-small.int8.amd64.ort arm64=gte-small.int8.arm.ort`.
- **`plugins/frank-exchange-of-views/native/build.sh`**
  - Builds `feov-embed` across all 5 targets (`linux-{amd64,arm64}`, `windows-{amd64,arm64}`, `darwin-arm64`) using zig and MSVC, embedding compiled-in source digest via `-DFEOV_HELPER_DIGEST=…`.
- **`plugins/frank-exchange-of-views/tools/internal/nativehelper/`**
  - Uses existing `nativehelper.Client` (from Part 1) for child process management, framing, and deadlines.
  - `digest_gen.go` **[MODIFY — generated]**: generated by `devcmd/nativedigest`, carries `feov-embed` source digest and per-arch model SHA256 pins.
- **`plugins/frank-exchange-of-views/tools/internal/vecembed/`** **[NEW]**
  - `vecembed.go` **[NEW]**: pure-Go `Embedder` interface, `Default() (Embedder, error)`, batching logic (`embedBatch = 16`), L2 normalization, `Cosine(a, b []float32) float32`. Compiles with `CGO_ENABLED=0`, zero cgo symbols or external toolchain dependencies.
  - `tokenize.go` **[NEW]**: pure-Go BERT WordPiece tokenizer matching `thenlper/gte-small` vocabulary (`30,522` tokens) across all 400/400 test cases.
  - `vecembed_test.go` **[NEW]**: unit tests, WordPiece 400/400 test suite, batch splitting test (`TestEmbedSplitsAtTheBatch`), model pin match test, and fidelity tests (`TestGteSmallInt8Fidelity`).
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
  - [`fetchcache.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/fetchcache.go), [`pageread.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/pageread.go), [`readscanned.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache/readscanned.go) **[MODIFY]**: pre-warm `<text_sha>.vec` when an embedder is available.
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
  - [`.github/workflows/hooks.yml`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/.github/workflows/hooks.yml) **[MODIFY]**: include `feov-embed` in `native-scope`, `native-build`, and `native-run`; package `feov-embed` and model exports in release job.
  - [`scripts/fetchbingen/fetch-bin.sh`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/fetchbingen/fetch-bin.sh) **[MODIFY]** (and regenerate `plugins/*/hooks/fetch-bin.sh` via `go -C scripts run ./fetchbingen`): check and install `feov-embed` and per-arch model exports alongside other binaries.
  - [`plugins/prosthetic-conscience/tools/internal/doctor/main.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/prosthetic-conscience/tools/internal/doctor/main.go) **[MODIFY]**: check and `--fix` helper and model assets.
  - [`scripts/universe.sh`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/universe.sh) **[MODIFY]**: build `feov-embed` into local release and stage into universe cache.
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

#### Census C: Native Helper & Asset Provisioning Consumers
Command:
```sh
git grep -n "feov-embed\|native/build.sh\|native/PINS.txt\|fetchbingen" .github/ plugins/ scripts/
```
Consumers:
- [`.github/workflows/hooks.yml`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/.github/workflows/hooks.yml): native build and test jobs (`native-scope`, `native-build`, `native-run`), release packaging. **[MODIFY]**
- [`plugins/frank-exchange-of-views/native/build.sh`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/frank-exchange-of-views/native/build.sh) and `native/feov-embed/` (Part 3 of `plans/native-helpers.md`). **[NEW]**
- [`scripts/fetchbingen/fetch-bin.sh`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/fetchbingen/fetch-bin.sh) and [`scripts/fetchbingen/main.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/fetchbingen/main.go) (generates `plugins/frank-exchange-of-views/hooks/fetch-bin.sh`, `plugins/gray-area/hooks/fetch-bin.sh`, `plugins/prosthetic-conscience/hooks/fetch-bin.sh`). **[MODIFY]**
- [`plugins/prosthetic-conscience/tools/internal/doctor/main.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/prosthetic-conscience/tools/internal/doctor/main.go), [`fetchbin_contract_test.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/prosthetic-conscience/tools/internal/doctor/fetchbin_contract_test.go), [`release_contract_test.go`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/plugins/prosthetic-conscience/tools/internal/doctor/release_contract_test.go). **[MODIFY]**
- [`scripts/universe.sh`](file:///home/gblock_ctoforaday_com/projects/special-circumstances/scripts/universe.sh): lines `378-453` (local release staging of helpers and model assets). **[MODIFY]**

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
| **R2: `feov-embed` crashes or overruns deadline during search** | Low | High | Low | The process boundary isolates crashes from `feov-record`. The client enforces a 120 s overall deadline and bounds batches to $\le 16$ texts ($\approx 3\text{ s}$ execution), killing the child on failure and spawning a fresh helper on the next request. |
| **R3: Missing or mismatched `feov-embed` or model file** | Medium | High | Low | Explicit hard refusal (`helper-absent` or `helper-digest-mismatch`) naming the asset, compiled-in hash, and `/prosthetic-conscience:doctor --fix`. Never falls back silently to BM25-only. |
| **R4: Stale vectors after `ocr read` or `render-page` updates `<sha>.ocr.txt`** | Low | High | Low | Vector cache files are addressed by `<text_sha>.vec` (`sha256` of the extracted text bytes), never by `<entry.Sha>.vec`. Any change to `.txt` or `.ocr.txt` produces a new `text_sha` and misses the old `.vec` file automatically. |
| **R5: 8-bit quantization drift vs FP32 reference embeddings** | Low | Medium | Low | `TestEmbedFidelity` gates on mean cosine $\ge 0.984$, min $\ge 0.966$ against `ref_gte.npy` on every release target (S5). Retrieval benchmarks verify $+0.048$ / $+0.040$ gain over BM25. |
| **R6: Helper build or staticness failure on Windows/Linux/Darwin** | Medium | High | Medium | `native/build.sh` uses MSVC on Windows and zig on Linux/macOS. Proven static across all 5 targets by `scripts/staticproof` before publication (S1, S2, F3). |
| **R7: Protocol pin advance on `show report --search`** | High | High | Low | Ensure `serve = false` in `internal/cli/seat/verbs.go:1007-1014` when `flags.Search` is given, suppressing `record.ServeReport`. Tested in `selectorcensus_test.go`. |

---

## V. Verification Plan (the checklist)

> **Standing rule**: Before running any FEOV test or gate from an agent session, run `unset $(env | grep -o '^FEOV_[A-Z_]*')`.

1. **Native Helper Cross-Build & Static Proof (all 5 release targets)**
   - Run `native/build.sh feov-embed <target>` for `linux-amd64`, `linux-arm64`, `windows-amd64`, `windows-arm64`, and `darwin-arm64` (F15).
   - Run `go -C scripts run ./staticproof <binary>` against each built helper.
   - **Pass criterion**: all 5 binaries build and pass `staticproof` with 0 forbidden dynamic libraries (verifying `peLibAllowed` accepts `{KERNEL32, ADVAPI32, dbghelp, api-ms-win-core-path-l1-1-0}`).
   - *Re-arm when*: `native/feov-embed/**`, `native/frame/**`, or `native/build.sh` changes.

2. **Embedder Unit, Batching, & Fidelity Suite (`internal/vecembed`)**
   - Run `go test -count=1 -v ./internal/vecembed/`.
   - **Pass criteria**:
     - `TestGteSmallInt8Fidelity`: mean cosine $\ge 0.984$, min $\ge 0.966$ against `ref_gte.npy` reference embeddings.
     - `TestEmbedThroughput`: $\ge 5.0\text{ chunks/s}$ on 200-word chunks single-threaded; cold single-query latency $\le 250\text{ ms}$.
     - `TestEmbedSplitsAtTheBatch`: 1,000 texts split into batches of $\le 16$ texts each.
     - `TestMissingOrMismatchedHelperRefusesByName`: when `feov-embed` or `gte-small.int8.ort` is missing or mismatched, `New()` refuses naming the error and `/prosthetic-conscience:doctor --fix`.
   - *Re-arm when*: any file under `internal/vecembed/` or `native/feov-embed/` changes.

3. **Hybrid Retrieval & Replay Suite (`internal/hybridsearch` & `internal/fetchcache`)**
   - Run `go test -count=1 -v ./internal/hybridsearch/ ./internal/fetchcache/`.
   - **Pass criteria**:
     - Hybrid RRF ($k=60$) over the checked-in replay fixture achieves `Hit@10` $\ge 0.93$ and beats standalone vector search at `Hit@5` by $\ge 0.20$.
     - Modifying a cached source's `.ocr.txt` changes its `text_sha` and computes fresh vectors without reading the prior `.vec` file.
     - Searching a run where one cached PDF is `image_only` (no `.ocr.txt`) lists that source in `unsearched_sources` alongside the ranked hits from extracted sources.
     - Cross-source search hits carry `sha256` and `url` and cite cleanly via `blue cite --url <url> --source-span "<source_span>"`.
   - *Re-arm when*: any file under `internal/hybridsearch/` or `internal/fetchcache/` changes.

4. **CLI Surface, Selector Census, & Provisioning Contract**
   - Run `go test -count=1 -v ./internal/cli/... ./internal/flags/... ./internal/terms/...` and `go -C ../../prosthetic-conscience/tools test -count=1 -v ./internal/doctor/...`.
   - **Pass criteria**:
     - `TestEverySelectorThatAViewAdvertisesSelects` passes for `--quote` across every view in `selectorViews(t)`.
     - `TestRankedSearchStatesCutAndScope` verifies `show report --search` and `show source --search` (single-source and cross-source) emit `ranked_and_cut`, `searched_chunks`, `unsearched_sources`, and citeable `source_span` + `page`.
     - `TestSearchDoesNotAdvanceServeReportPin` verifies that `show report --search` does not record `ServeReport` or advance the reading pin for lens seats.
     - `TestSearchAndQuoteMutuallyExclusive` verifies that passing both `--search` and `--quote` refuses with `feov.Validation`.
     - `fetchbin_contract_test.go` and `release_contract_test.go` verify `feov-embed` and per-arch model exports are checksummed in `SHA256SUMS` and installed by `fetch-bin.sh` and `sc-doctor --fix`.
   - *Re-arm when*: any file under `internal/cli/`, `internal/flags/`, `internal/terms/`, `scripts/fetchbingen/`, or `internal/doctor/` changes.

5. **Generators, Repository Gates, & End-to-End Smoke Verification**
   - Regenerate carriers (`go -C scripts run ./fetchbingen`, `go -C scripts run ./agentgen`, `go -C scripts run ./vocabdoc`, `UPDATE_GOLDENS=1`), read every changed golden line and `--help` diff one command at a time, and run `go -C scripts run ./check`.
   - Build an isolated universe (`scripts/universe.sh build`), verify the staged marketplace carries the new `--search` help text, `feov-embed`, and `gte-small.int8.ort`, and drive `feov-record show report --search` and `feov-record show source --search` against an archived real-source run (`run-archive/`).
   - Drive `releasegate/fuzz` tests (`FEOV_RELEASE_GATE=1 go test ./releasegate/fuzz/`) with `--match` retired.
   - **Pass criterion**: `go -C scripts run ./check` exits 0 with 0 failed gates, and the end-to-end CLI invocation against real cached sources returns citeable spans that `blue cite --source-span` accepts.
