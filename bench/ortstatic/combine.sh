#!/usr/bin/env bash
# combine.sh <ort build dir (the <config> dir)> <src dir> <out dir> [ar]
# Merges every static archive the ORT runtime needs into <out>/lib/libonnxruntime_static.a and copies the C API
# headers to <out>/include. Host-only tools (protoc, full protobuf, upb, flatc's library) are left out.
set -euo pipefail
B=$1; SRC=$2; O=$3; AR=${4:-ar}
mkdir -p "$O/lib" "$O/include"
cp "$SRC/include/onnxruntime/core/session/onnxruntime_c_api.h" "$SRC/include/onnxruntime/core/session/onnxruntime_ep_c_api.h" "$O/include/" 2>/dev/null || cp "$SRC/include/onnxruntime/core/session/onnxruntime_c_api.h" "$O/include/"
for h in "$SRC"/include/onnxruntime/core/session/onnxruntime_*.h; do cp "$h" "$O/include/"; done
mapfile -t libs < <(find "$B" \( -name '*.a' -o -name '*.lib' \) \
  ! -name 'libprotoc.a' ! -name 'libprotobuf.a' ! -name 'libupb*.a' ! -name 'libflatbuffers.a' \
  ! -name 'protoc*.lib' ! -name 'libprotobuf.lib' ! -name 'libprotobufd.lib' ! -name 'flatbuffers.lib' ! -name 'upb*.lib' \
  ! -name 'libonnxruntime_static.a' ! -path '*gtest*' ! -path '*gmock*' ! -path '*benchmark*' ! -name '*test*' | sort)
printf '%s\n' "${libs[@]}" > "$O/archives.txt"
OUT="$O/lib/libonnxruntime_static.a"; rm -f "$OUT"
case "$(uname -s)" in
Darwin) if [ "$AR" = ar ]; then libtool -static -o "$OUT" "${libs[@]}"; else { echo "CREATE $OUT"; for l in "${libs[@]}"; do echo "ADDLIB $l"; done; echo SAVE; echo END; } | $AR -M; fi ;;
*) { echo "CREATE $OUT"; for l in "${libs[@]}"; do echo "ADDLIB $l"; done; echo SAVE; echo END; } | $AR -M ;;
esac
ls -l "$OUT"
