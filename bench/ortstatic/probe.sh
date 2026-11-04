#!/usr/bin/env bash
# probe.sh <target> <native|zig> <ort out dir> <probe src> <out dir> [zig]
# Builds the ortprobe (ONNX Runtime linked) and the same probe with -tags noort, each unstripped and stripped (-s -w),
# for <target>, the way FEOV links its cgo stack: zig cc/c++ with zig's static libc++ (or the host compiler for native).
# Writes <out>/{ortprobe,ortprobe.s,noort,noort.s}[.exe] and <out>/sizes.txt.
set -euo pipefail
TARGET=$1; TC=$2; ORT=$(cd "$3" && pwd); P=$(cd "$4" && pwd); O=$(mkdir -p "$5" && cd "$5" && pwd); ZIG=${6:-${ZIG:-zig}}
GOOS=${TARGET%-*}; GOARCH=${TARGET#*-}; export GOOS GOARCH CGO_ENABLED=1
EXE=; [ "$GOOS" = windows ] && EXE=.exe
export CGO_CFLAGS="-I$ORT/include"
LIBS="-L$ORT/lib -lonnxruntime_static"
EXTLD="-static"; LDX=""
if [ "$TC" = zig ]; then
  case $TARGET in
    linux-amd64) TRIPLE=x86_64-linux-musl ;; linux-arm64) TRIPLE=aarch64-linux-musl ;;
    windows-amd64) TRIPLE=x86_64-windows-gnu ;; windows-arm64) TRIPLE=aarch64-windows-gnu ;;
    darwin-amd64) TRIPLE=x86_64-macos ;; darwin-arm64) TRIPLE=aarch64-macos ;;
  esac
  ZB=$O/zigbin; mkdir -p "$ZB"
  printf '#!/bin/sh\nexec %q cc -target %s "$@"\n' "$ZIG" "$TRIPLE" > "$ZB/cc"
  printf '#!/bin/sh\nexec %q c++ -target %s "$@"\n' "$ZIG" "$TRIPLE" > "$ZB/cxx"
  chmod +x "$ZB/cc" "$ZB/cxx"; export CC=$ZB/cc CXX=$ZB/cxx
  export ZIG_GLOBAL_CACHE_DIR=${ZIG_GLOBAL_CACHE_DIR:-$O/zig-cache}
  LIBS="$LIBS -lc++ ${ORT_EXTRA_LIBS:-}"
  if [ "$GOOS" = darwin ]; then
    # build-cstack.sh's darwin recipe: tapi stubs exporting what Go imports from libresolv, CoreFoundation, Security.
    F=$O/frameworks; GOSRC=$(go env GOROOT)/src; PROC=${TRIPLE%%-*}; [ "$PROC" = aarch64 ] && PROC=arm64
    # $2: extra symbols the C stack imports (abseil cctz's CoreFoundation calls), merged with Go's own.
    syms() { { grep -rhoE "//go:cgo_import_dynamic [^ ]+ [^ ]+ \"[^\"]*$1\"" "$GOSRC" | awk '{print "_" $3}'; printf '%s\n' ${2:-}; } | sort -u | paste -sd, - | sed 's/,/, /g'; }
    CFX="_CFTimeZoneCopyDefault _CFTimeZoneGetName _CFStringGetLength _CFStringGetMaximumSizeForEncoding _CFStringGetCString _CFRelease"
    tbd() { mkdir -p "$(dirname "$1")"; { printf -- "--- !tapi-tbd\ntbd-version: 4\ntargets: [ %s-macos ]\ninstall-name: '%s'\ncurrent-version: 1.0\n" "$PROC" "$2"
            [ -n "${3:-}" ] && printf -- "exports:\n  - targets: [ %s-macos ]\n    symbols: [ %s ]\n" "$PROC" "$3"; printf -- '...\n'; } > "$1"; }
    tbd "$F/lib/libresolv.tbd" /usr/lib/libresolv.9.dylib
    tbd "$F/Frameworks/CoreFoundation.framework/CoreFoundation.tbd" /System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation "$(syms CoreFoundation "$CFX")"
    tbd "$F/Frameworks/Security.framework/Security.tbd" /System/Library/Frameworks/Security.framework/Versions/A/Security "$(syms Security)"
    # abseil cctz calls CoreFoundation (allowlisted); Go names that framework only when its own packages need it.
    LIBS="$LIBS -L$F/lib -F$F/Frameworks -framework CoreFoundation"; EXTLD="-L$F/lib -F$F/Frameworks"; LDX="-w"
  fi
else
  case $GOOS in
    linux) LIBS="$LIBS -lstdc++ -lm -lpthread -ldl" ;;
    darwin) LIBS="$LIBS -lc++ -framework Foundation"; EXTLD="" ;;
    windows) LIBS="$LIBS -lstdc++" ;;
  esac
fi
export CGO_LDFLAGS="$LIBS"
echo "CC=${CC:-default} CGO_LDFLAGS=$CGO_LDFLAGS extldflags=$EXTLD"
b() { # b <out> <tags> <extra ldflags>
  local ld="$LDX $3 -linkmode external"; [ -n "$EXTLD" ] && ld="$ld -extldflags '$EXTLD'"
  echo "go build -tags '$2' -ldflags \"$ld\" -o $1"
  ( cd "$P" && go build -tags "$2" -ldflags "$ld" -o "$O/$1" . ) 2>&1 | grep -v "warning: Using 'dlopen'" | tail -40
  return "${PIPESTATUS[0]}"
}
rc=0
b ortprobe$EXE "" "" || rc=1
b ortprobe.s$EXE "" "-s -w" || rc=1
b noort$EXE noort "" || rc=1
b noort.s$EXE noort "-s -w" || rc=1
( cd "$O" && for f in ortprobe$EXE ortprobe.s$EXE noort$EXE noort.s$EXE; do [ -f "$f" ] && echo "$(wc -c < "$f") $f"; done ) | tee "$O/sizes.txt"
exit $rc
