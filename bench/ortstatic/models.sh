#!/usr/bin/env bash
# models.sh <bench dir (bench/)> <work> <out>
# Exports thenlper/gte-small to ONNX (float32 + dynamic int8) with bench/embed/ort/ort.py, converts both to ORT format
# for each target platform (Fixed optimisation, type reduction on), and records sizes and the operator configs.
set -euo pipefail
B=$(cd "$1" && pwd); W=$(mkdir -p "$2" && cd "$2" && pwd); O=$(mkdir -p "$3" && cd "$3" && pwd)
python3 -m venv "$W/venv"; PY=$W/venv/bin/python
"$PY" -m pip install -q --upgrade pip
"$PY" -m pip install -q torch==2.15.0 --index-url https://download.pytorch.org/whl/cpu
"$PY" -m pip install -q transformers==5.22.1 tokenizers==0.23.3 safetensors==0.8.0 huggingface_hub==2.4.0 \
  onnxruntime==1.31.1 onnx==1.23.2 onnxscript==0.7.2 numpy==2.5.4 protobuf==7.36.2 sentencepiece pyyaml tqdm
"$PY" -m pip freeze | grep -iE '^(torch|transformers|onnxruntime|onnx|numpy)=='
HF=$W/gte-small; REV=17e1f347d17fe144873b1201da91788898c639cd; mkdir -p "$HF/1_Pooling"
for f in config.json model.safetensors vocab.txt tokenizer.json tokenizer_config.json special_tokens_map.json modules.json sentence_bert_config.json 1_Pooling/config.json; do
  curl -fsSL --retry 5 -o "$HF/$f" "https://huggingface.co/thenlper/gte-small/resolve/$REV/$f"
done
mkdir -p "$W/onnx"
"$PY" -I "$B/embed/ort/ort.py" "$HF" "$W/onnx" export
cp "$HF/vocab.txt" "$O/vocab.txt"
for p in amd64 arm; do
  "$PY" -I -m onnxruntime.tools.convert_onnx_models_to_ort --optimization_style Fixed --enable_type_reduction \
    --target_platform $p --output_dir "$O/$p" "$W/onnx" > "$O/convert-$p.log" 2>&1 || { cat "$O/convert-$p.log"; exit 1; }
  cp "$O/$p/vocab.txt" /dev/null 2>&1 || cp "$HF/vocab.txt" "$O/$p/vocab.txt"
done
ls -l "$W/onnx" "$O/amd64" "$O/arm" | tee "$O/sizes.txt"
for p in amd64 arm; do
  echo "== diff of generated $p config vs committed (comment lines excluded)"
  diff <(grep -v '^#' "$O/$p/required_operators_and_types.config") <(grep -v '^#' "$B/ortstatic/ops.config") && echo "identical operator set ($p)"
done
