#!/usr/bin/env bash
# Builds every arm on the runner. Usage: build.sh <bench/embed dir> <work dir>
# Third-party code (OpenBLAS, llama.cpp, model weights) is fetched into <work>, never into the checkout.
set -euo pipefail
S=$(cd "$1" && pwd); W=$2; mkdir -p "$W"/{src,bin,models,py}; W=$(cd "$W" && pwd)
J=$(nproc)
step() { echo "::group::$*"; }
end() { echo "::endgroup::"; }

step "python venv"
python3 -m venv "$W/venv"; PY=$W/venv/bin/python
"$PY" -m pip install -q --upgrade pip
"$PY" -m pip install -q torch==2.15.0 --index-url https://download.pytorch.org/whl/cpu
"$PY" -m pip install -q transformers==5.22.1 tokenizers==0.23.3 safetensors==0.8.0 huggingface_hub==2.4.0 \
  onnxruntime==1.31.1 onnx==1.23.2 onnxscript==0.7.2 numpy==2.5.4 protobuf==7.36.2 sentencepiece pyyaml tqdm
"$PY" -m pip freeze | grep -iE '^(torch|transformers|onnxruntime|onnx|numpy|safetensors|tokenizers)=='
end

step "fetch thenlper/gte-small"
HF=$W/models/gte-small; REV=17e1f347d17fe144873b1201da91788898c639cd; mkdir -p "$HF/1_Pooling"
for f in config.json model.safetensors vocab.txt tokenizer.json tokenizer_config.json special_tokens_map.json modules.json sentence_bert_config.json 1_Pooling/config.json; do
  curl -fsSL --retry 5 -o "$HF/$f" "https://huggingface.co/thenlper/gte-small/resolve/$REV/$f"
done
ls -la "$HF"
end

step "gte-go model"
"$PY" -I "$S/gte-go/convert_model.py" "$HF" "$W/models/gte-small.gtemodel"
ls -la "$W/models/gte-small.gtemodel"
end

step "OpenBLAS 0.3.34 (static, DYNAMIC_ARCH)"
cd "$W/src"
curl -fsSL --retry 5 -o ob.tar.gz https://github.com/OpenMathLib/OpenBLAS/releases/download/v0.3.34/OpenBLAS-0.3.34.tar.gz
echo "cd7e129868320cc2d033afa920e31202dfe0b8066a5b66661900ccc0f197dfed  ob.tar.gz" | sha256sum -c -
tar xzf ob.tar.gz; cd OpenBLAS-0.3.34
OBF="USE_OPENMP=0 NO_LAPACK=1 NO_FORTRAN=1 DYNAMIC_ARCH=1 NO_SHARED=1 USE_THREAD=1 NUM_THREADS=64"
make -j"$J" $OBF libs > "$W/openblas-make.log" 2>&1 || { tail -50 "$W/openblas-make.log"; exit 1; }
make $OBF PREFIX="$W/openblas" install > "$W/openblas-install.log" 2>&1 || { tail -50 "$W/openblas-install.log"; exit 1; }
ls -la "$W/openblas/lib"
cc -O2 "$S/blasinfo.c" -o "$W/bin/blasinfo" "$W/openblas/lib/libopenblas.a" -lpthread -lm
"$W/bin/blasinfo"
end

step "gte-go probes"
cd "$S/probe-go"
export GOFLAGS=-mod=mod
CGO_ENABLED=0 go build -o "$W/bin/gtego_asm" .
CGO_ENABLED=1 CGO_CFLAGS="-I$W/openblas/include" \
  CGO_LDFLAGS="-L$W/openblas/lib -Wl,-Bstatic -lopenblas -Wl,-Bdynamic -lm -lpthread" \
  go build -o "$W/bin/gtego_blas" .
if [ -d "$S/gte-simd" ]; then
  cd "$S/probe-simd"; CGO_ENABLED=0 GOEXPERIMENT=simd go build -o "$W/bin/gtego_simd" .
  cd "$S/gte-simd"; CGO_ENABLED=0 GOEXPERIMENT=simd go build -o "$W/bin/simdwidth" ./cmd/simdwidth
  "$W/bin/simdwidth"
fi
for b in "$W"/bin/gtego_*; do go version -m "$b" | grep -E 'CGO_ENABLED|GOEXPERIMENT|GOARCH|GOAMD64|GOARM64' | tr '\n' ' '; echo " <- $b"; done
end

step "llama.cpp v0.8.0"
cd "$W/src"
git -c advice.detachedHead=false clone -q --depth 1 --branch v0.8.0 https://github.com/ggml-org/llama.cpp
test "$(git -C llama.cpp rev-parse HEAD)" = 8642541b255bfb5dc39cd351a178be09013a55d6
curl -fsSL --retry 5 -o "$W/src/json.hpp" https://raw.githubusercontent.com/nlohmann/json/v3.12.0/single_include/nlohmann/json.hpp
echo "aaf127c04cb31c406e5b04a63f1ae89369fccde6d8fa7cdda1ed4f32dfc5de63  $W/src/json.hpp" | sha256sum -c -
COMMON=(-DCMAKE_BUILD_TYPE=Release -DBUILD_SHARED_LIBS=OFF -DGGML_NATIVE=ON -DGGML_OPENMP=OFF -DGGML_LLAMAFILE=ON
  -DLLAMA_CURL=OFF -DLLAMA_OPENSSL=OFF -DLLAMA_BUILD_TESTS=OFF -DLLAMA_BUILD_EXAMPLES=OFF -DLLAMA_BUILD_TOOLS=OFF
  -DLLAMA_BUILD_SERVER=OFF -DGGML_BUILD_TESTS=OFF -DGGML_BUILD_EXAMPLES=OFF)
cmake -S llama.cpp -B llama-noblas "${COMMON[@]}" > "$W/llama-cmake-noblas.log" 2>&1 || { cat "$W/llama-cmake-noblas.log"; exit 1; }
grep -E 'CPU backend|Adding CPU|ARM feature|x86 detected|GGML_SYSTEM_ARCH' "$W/llama-cmake-noblas.log" || true
cmake --build llama-noblas --target llama ggml -j"$J" > "$W/llama-build-noblas.log" 2>&1 || { tail -60 "$W/llama-build-noblas.log"; exit 1; }
cmake -S llama.cpp -B llama-blas "${COMMON[@]}" -DGGML_BLAS=ON -DGGML_BLAS_VENDOR=OpenBLAS \
  -DBLAS_INCLUDE_DIRS="$W/openblas/include" -DBLAS_LIBRARIES="$W/openblas/lib/libopenblas.a" > "$W/llama-cmake-blas.log" 2>&1 || { cat "$W/llama-cmake-blas.log"; exit 1; }
grep -E 'BLAS|Adding CPU' "$W/llama-cmake-blas.log" || true
cmake --build llama-blas --target llama ggml ggml-blas -j"$J" > "$W/llama-build-blas.log" 2>&1 || { tail -60 "$W/llama-build-blas.log"; exit 1; }
INC=(-I"$W/src" -Illama.cpp/include -Illama.cpp/ggml/include)
g++ -O3 -std=c++17 -static "${INC[@]}" "$S/llama/harness_batched.cpp" \
  llama-noblas/src/libllama.a llama-noblas/ggml/src/libggml.a llama-noblas/ggml/src/libggml-cpu.a llama-noblas/ggml/src/libggml-base.a \
  -lpthread -ldl -o "$W/bin/harness_noblas" 2>&1 | grep -v -e 'dlopen' -e 'dl_load_library' || true
g++ -O3 -std=c++17 -static "${INC[@]}" "$S/llama/harness_batched.cpp" \
  llama-blas/src/libllama.a llama-blas/ggml/src/libggml.a llama-blas/ggml/src/ggml-blas/libggml-blas.a llama-blas/ggml/src/libggml-cpu.a llama-blas/ggml/src/libggml-base.a \
  "$W/openblas/lib/libopenblas.a" -lpthread -ldl -lm -o "$W/bin/harness_blas" 2>&1 | grep -v -e 'dlopen' -e 'dl_load_library' || true
test -x "$W/bin/harness_noblas"; test -x "$W/bin/harness_blas"
end

step "GGUF models"
curl -fsSL --retry 5 -o "$W/models/gte-small-q8_0.gguf" \
  https://huggingface.co/ggml-org/gte-small-Q8_0-GGUF/resolve/1dc93210ef4edf1a611523c8d17544c47564275b/gte-small-q8_0.gguf
echo "945a2da32e7bfb2cdb3a3fde50bc40c5d103db84fedc078adf415f3bb10f2ed0  $W/models/gte-small-q8_0.gguf" | sha256sum -c -
"$PY" llama.cpp/convert_hf_to_gguf.py "$HF" --outtype f16 --outfile "$W/models/gte-small-f16.gguf" > "$W/convert-f16.log" 2>&1 || { tail -40 "$W/convert-f16.log"; exit 1; }
tail -3 "$W/convert-f16.log"; ls -la "$W/models"
end

step "ONNX export"
"$PY" -I "$S/ort/ort.py" "$HF" "$W/models" export
end
echo BUILD-OK
