# vecembed reference data

The tests embed these files (`go:embed`), so a test binary carries them and needs no Python and no
network.

| file | sha256 | what it is |
|---|---|---|
| `speed.jsonl` | `28f580ca0256d880bda9cd6ddfccc36c87910bfaa2aee14237795dff9957e4ac` | 300 texts, one `{"text": …}` per line: the embedding bake-off's fidelity set (Wikipedia prose, many with `\uXXXX` escapes and LaTeX) |
| `len200.jsonl` | `1d22029f2b8b6f5041bdd689cf04de831a62addc010bfb4d1065e93433fbc177` | 100 chunks of about 200 words: the bake-off's one-thread speed set, the source chunker's window |
| `ref_gte.npy` | `8ff54a03dce2d64556f290ba7abb1044a74a8c6d60e186f7659075466bfe0b0f` | float32 `(300, 384)`, little-endian, C order: FP32 `thenlper/gte-small` embeddings of `speed.jsonl` in line order — `AutoModel` last hidden state, mean-pooled over the attention mask, L2-normalised, input truncated at 512 ids |
| `hf_tokens_400.json` | `15ca857ed0aa107f678d8f488dd82145aab0e2b556f98b59634070321deb0c1b` | `AutoTokenizer.from_pretrained('thenlper/gte-small')(text, truncation=True, max_length=512).input_ids` for `speed.jsonl` (key `speed`, 300) then `len200.jsonl` (key `len200`, 100); transformers 5.22.1, tokenizers 0.23.3 |

`speed.jsonl`, `len200.jsonl` and `ref_gte.npy` are the bake-off's inputs and reference, copied
unchanged; plans/vector-search.md's S5 bar (mean ≥ 0.984, min ≥ 0.966) was measured against them.
`ref_gte.npy` was re-derived for every 30th row with the PyTorch model as described above: cosine
1.0000000 ± 1e-7 on each. `hf_tokens_400.json` was written by a script that loads each line's
`text`, calls the tokenizer as above, and writes one id list per line.
