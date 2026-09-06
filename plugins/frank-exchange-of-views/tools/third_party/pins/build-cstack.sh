#!/bin/bash
# build-cstack.sh — the static tesseract+leptonica C stack, from PINS.txt, for one target.
#
# Merged from the two Wave 0 spike drivers (~/scratch-tessspike/build.sh, linux/musl;
# ~/scratch-crossbuild/build-cross.sh, windows/darwin) with every recorded gotcha applied;
# each fix cites the failure it prevents. The build host is linux/x86_64 — every target,
# including darwin, cross-builds from it with zig (no macOS SDK; two .tbd stubs below).
#
# Usage:
#   build-cstack.sh <target> <workdir> [start-phase]   build into <workdir>/<target>/prefix
#   build-cstack.sh env <target> <workdir>             print the cgo env for that prefix
#   build-cstack.sh check <binary>                     assert per-format staticness (§V.3)
#
#   target:      linux-amd64 | linux-arm64 | windows-amd64 | windows-arm64 |
#                darwin-amd64 | darwin-arm64
#   start-phase: zlib (default) | libpng | leptonica | tesseract — resume by skipping
#                completed phases.
#
# The whole build runs under an exclusive flock on <workdir>: this box has had duplicated
# background builds racing one prefix, and the loser's half-written libraries link into a
# binary that fails somewhere far from the cause. Waiting is correct; racing is not.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PINS="$SCRIPT_DIR/PINS.txt"
J="${CSTACK_JOBS:-$(nproc)}"

die() { echo "build-cstack: $*" >&2; exit 1; }

# ---- pin manifest access ------------------------------------------------------------
# PINS.txt is the record; this script consumes it rather than restating any URL or hash.
pin_field() { # pin_field <file> <n: 1=sha 3=url>
	awk -v f="$2" -v n="$1" '$0 !~ /^#/ && $2 == f { print $n }' "$PINS"
}

fetch_pinned() { # fetch_pinned <file> <destdir> — download if absent, verify ALWAYS
	local file="$1" dest="$2/$1" sha url
	sha="$(pin_field 1 "$file")" ; url="$(pin_field 3 "$file")"
	[ -n "$sha" ] && [ -n "$url" ] || die "$file is not pinned in PINS.txt"
	if [ ! -f "$dest" ]; then
		echo "fetching $file"
		curl -fsSL -o "$dest.part" "$url" && mv "$dest.part" "$dest"
	fi
	echo "$sha  $dest" | sha256sum -c --quiet - || die "$file does not match its pin"
}

# ---- staticness proof (plan §V.3) ---------------------------------------------------
# One Go checker for all three formats: debug/elf, debug/pe and debug/macho parse the
# actual import records, where the spike's tooling could not — GNU objdump has no
# pei-aarch64 backend, so windows/arm64 fell back to a string-scan there. A parse can
# refuse; a string-scan can only hope. Requires a go tool on PATH.
if [ "${1:-}" = "check" ]; then
	[ $# -eq 2 ] || die "usage: build-cstack.sh check <binary>"
	exec go run "$SCRIPT_DIR/staticproof" "$2"
fi

# ---- target table -------------------------------------------------------------------
[ $# -ge 2 ] || die "usage: build-cstack.sh [env|check] <target> <workdir> [start-phase]"
MODE="build"
if [ "$1" = "env" ]; then MODE="env"; shift; fi
TARGET="$1"; WORK="$2"; START="${3:-zlib}"

case "$TARGET" in
	# CMAKE_SYSTEM_PROCESSOR must be "aarch64", case-sensitively, for the arm targets:
	# tesseract's NEON gate is `MATCHES "arm64|aarch64.*"`, and "ARM64" skips
	# dotproductneon.cpp while simddetect.cpp (compiled __aarch64__) still calls it —
	# an undefined DotProductNEON at link, far from the cause.
	linux-amd64)   TRIPLE=x86_64-linux-musl;  SYSNAME=Linux;   PROC=x86_64  ;;
	linux-arm64)   TRIPLE=aarch64-linux-musl; SYSNAME=Linux;   PROC=aarch64 ;;
	windows-amd64) TRIPLE=x86_64-windows-gnu; SYSNAME=Windows; PROC=AMD64   ;;
	windows-arm64) TRIPLE=aarch64-windows-gnu; SYSNAME=Windows; PROC=aarch64 ;;
	darwin-amd64)  TRIPLE=x86_64-macos;       SYSNAME=Darwin;  PROC=x86_64  ;;
	darwin-arm64)  TRIPLE=aarch64-macos;      SYSNAME=Darwin;  PROC=arm64   ;;
	*) die "unknown target $TARGET" ;;
esac

T="$WORK/$TARGET"
PREFIX="$T/prefix"

if [ "$MODE" = "env" ]; then
	# The engine build contract, both halves: this env, AND `-tags tessocr` on the go
	# command (the engine is opt-in behind that tag; a default build gets the stub and
	# needs none of this). internal/tessocr carries no -I/-L of its own because the
	# prefix lives outside the repo; these variables are how a builder points at it.
	# -lc++ is zig's bundled STATIC libc++ — the whole "-lstdc++ handled by zig" story.
	echo "export CC=$T/bin/zigcc"
	echo "export CXX=$T/bin/zigcxx"
	echo "export CGO_ENABLED=1"
	echo "export CGO_CFLAGS=\"-I$PREFIX/include\""
	echo "export CGO_CXXFLAGS=\"-I$PREFIX/include\""
	echo "export CGO_LDFLAGS=\"-L$PREFIX/lib -ltesseract -lleptonica -lpng16 -lz -lc++\""
	case "$SYSNAME" in
	Darwin)
		# Fully static linking does not exist on macOS (no static libSystem). Go also
		# hard-adds -lresolv and -framework CoreFoundation to darwin external links;
		# the .tbd stubs in the prefix satisfy them at link time, the OS at run time.
		# -w because Go runs dsymutil after a darwin link and this host has none.
		echo "# go build -tags tessocr -ldflags '-w -linkmode external -extldflags -F$PREFIX/Frameworks'"
		;;
	*)
		echo "# go build -tags tessocr -ldflags '-linkmode external -extldflags \"-static\"'"
		;;
	esac
	exit 0
fi

# ---- lock, tools, sources -----------------------------------------------------------
mkdir -p "$WORK" "$T"/{bin,build} "$WORK/src" "$WORK/tools"
exec 9>"$WORK/.build.lock"
if ! flock -n 9; then
	echo "build-cstack: waiting for the lock on $WORK (another build is running)"
	flock 9
fi

# The pin constants in the Go engine and the tarball versions in PINS.txt are two
# carriers of one fact that cannot be generated from each other, so the drift is gated
# here, at the moment a stack is built: a stack whose versions the shipped Identity()
# would misreport must not exist.
TESSOCR_GO="$SCRIPT_DIR/../../internal/tessocr/tessocr.go"
for lib in tesseract leptonica; do
	v="$(awk -v l="$lib" '$0 !~ /^#/ && $2 ~ "^"l"-" { sub("^"l"-", "", $2); sub("\\.tar\\.gz$", "", $2); print $2 }' "$PINS")"
	grep -q "${lib}Pin = \"$v\"" "$TESSOCR_GO" \
		|| die "PINS.txt pins $lib $v but internal/tessocr/tessocr.go's ${lib}Pin disagrees — move both in one change"
done

# Toolchain: the pinned zig and cmake are linux-x86_64 host binaries. On another host,
# point ZIG/CMAKE at your own installs of the same versions.
if [ -z "${ZIG:-}" ]; then
	[ "$(uname -sm)" = "Linux x86_64" ] || die "pinned zig/cmake are linux-x86_64; set ZIG and CMAKE"
	if [ ! -x "$WORK/tools/zig/zig" ]; then
		fetch_pinned zig-0.16.0.tar.xz "$WORK/src"
		mkdir -p "$WORK/tools/zig"
		tar -C "$WORK/tools/zig" --strip-components=1 -xJf "$WORK/src/zig-0.16.0.tar.xz"
	fi
	ZIG="$WORK/tools/zig/zig"
fi
if [ -z "${CMAKE:-}" ]; then
	if [ ! -x "$WORK/tools/cmake/bin/cmake" ]; then
		fetch_pinned cmake-4.4.3.tar.gz "$WORK/src"
		mkdir -p "$WORK/tools/cmake"
		tar -C "$WORK/tools/cmake" --strip-components=1 -xzf "$WORK/src/cmake-4.4.3.tar.gz"
	fi
	CMAKE="$WORK/tools/cmake/bin/cmake"
fi

for f in zlib-1.3.2.tar.gz libpng-1.6.58.tar.gz leptonica-1.87.0.tar.gz tesseract-5.5.3.tar.gz; do
	fetch_pinned "$f" "$WORK/src"
done

# Per-target compiler wrappers. zigar/zigranlib matter for the COFF and Mach-O archives;
# GNU ar on the host writes archives zig's linker then refuses.
for tool in cc c++ ar ranlib; do
	name="zig${tool/c++/cxx}"
	printf '#!/bin/sh\nexec %q %s -target %s "$@"\n' "$ZIG" "$tool" "$TRIPLE" > "$T/bin/$name"
	# ar/ranlib take no -target
	case "$tool" in ar|ranlib) printf '#!/bin/sh\nexec %q %s "$@"\n' "$ZIG" "$tool" > "$T/bin/$name" ;; esac
	chmod +x "$T/bin/$name"
done
export CC="$T/bin/zigcc" CXX="$T/bin/zigcxx"
AR="$T/bin/zigar" RANLIB="$T/bin/zigranlib"

CROSS=(
	-DCMAKE_SYSTEM_NAME="$SYSNAME" -DCMAKE_SYSTEM_PROCESSOR="$PROC"
	-DCMAKE_C_COMPILER="$CC" -DCMAKE_CXX_COMPILER="$CXX"
	-DCMAKE_AR="$AR" -DCMAKE_RANLIB="$RANLIB"
	-DCMAKE_C_COMPILER_AR="$AR" -DCMAKE_CXX_COMPILER_AR="$AR"
	-DCMAKE_C_COMPILER_RANLIB="$RANLIB" -DCMAKE_CXX_COMPILER_RANLIB="$RANLIB"
	-DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX="$PREFIX"
	-DCMAKE_FIND_ROOT_PATH="$PREFIX"
	-DCMAKE_FIND_ROOT_PATH_MODE_PROGRAM=NEVER
	-DCMAKE_FIND_ROOT_PATH_MODE_LIBRARY=ONLY
	-DCMAKE_FIND_ROOT_PATH_MODE_INCLUDE=ONLY
	# zig bundles Apple's SDK zlib.h (any-darwin-any) and its -isystem dirs shadow
	# cmake's; imported-target includes become -I so the prefix headers win everywhere.
	# Without this, libpng dies with "The include path of <zlib.h> is incorrect".
	-DCMAKE_NO_SYSTEM_FROM_IMPORTED=ON
	# mingw headers hide M_PI under strict-ANSI C++17 (baseapi.cpp:2280); harmless on
	# the other targets, so it is set once for all of them.
	-DCMAKE_C_FLAGS=-D_USE_MATH_DEFINES -DCMAKE_CXX_FLAGS=-D_USE_MATH_DEFINES
)

phase() { echo "=== PHASE $1 ($TARGET) : $(date -u +%H:%M:%S) ==="; }
reached=false
runp() { [ "$1" = "$START" ] && reached=true; $reached; }
cd "$T/build"

# NOTE: cmake's IPO probe fails noisily with zig on the cross targets (wmem*/frexpf
# undefined in the LTO test link). Harmless — LTO is off — so do not chase those errors
# in the logs; they fail nothing.

if runp zlib; then
phase zlib
# zlib via cmake on EVERY target, not just cross: its ./configure refuses cross/mingw
# setups, and one recipe beats two. Its CMakeLists predates cmake 4; the policy floor
# bridges that.
rm -rf zlib && mkdir zlib
tar -C zlib --strip-components=1 -xzf "$WORK/src/zlib-1.3.2.tar.gz"
"$CMAKE" -S zlib -B zlib/b "${CROSS[@]}" \
	-DCMAKE_POLICY_VERSION_MINIMUM=3.5 \
	-DZLIB_BUILD_TESTING=OFF -DZLIB_BUILD_SHARED=OFF -DZLIB_BUILD_STATIC=ON \
	-DZLIB_BUILD_MINIZIP=OFF -DZLIB_INSTALL=ON
"$CMAKE" --build zlib/b -j"$J"
"$CMAKE" --install zlib/b
fi

if runp libpng; then
phase libpng
rm -rf libpng && mkdir libpng
tar -C libpng --strip-components=1 -xzf "$WORK/src/libpng-1.6.58.tar.gz"
"$CMAKE" -S libpng -B libpng/b "${CROSS[@]}" \
	-DPNG_SHARED=OFF -DPNG_STATIC=ON -DPNG_TESTS=OFF -DPNG_TOOLS=OFF \
	-DPNG_FRAMEWORK=OFF -DPNG_DEBUG=OFF \
	-DZLIB_ROOT="$PREFIX" -DZLIB_USE_STATIC_LIBS=ON
"$CMAKE" --build libpng/b -j"$J"
"$CMAKE" --install libpng/b
fi

if runp leptonica; then
phase leptonica
rm -rf leptonica && mkdir leptonica
tar -C leptonica --strip-components=1 -xzf "$WORK/src/leptonica-1.87.0.tar.gz"
# SW_BUILD=OFF everywhere: leptonica defaults it ON for Windows and then dies in
# find_package(SW); the linux recipe simply never hit that branch.
"$CMAKE" -S leptonica -B leptonica/b "${CROSS[@]}" \
	-DCMAKE_PREFIX_PATH="$PREFIX" \
	-DBUILD_SHARED_LIBS=OFF -DBUILD_PROG=OFF \
	-DENABLE_PNG=ON -DENABLE_ZLIB=ON \
	-DENABLE_GIF=OFF -DENABLE_JPEG=OFF -DENABLE_TIFF=OFF \
	-DENABLE_WEBP=OFF -DENABLE_OPENJPEG=OFF -DSW_BUILD=OFF \
	-DZLIB_ROOT="$PREFIX" -DPNG_ROOT="$PREFIX" -DZLIB_USE_STATIC_LIBS=ON
"$CMAKE" --build leptonica/b -j"$J"
"$CMAKE" --install leptonica/b
fi

if runp tesseract; then
phase tesseract
rm -rf tesseract && mkdir tesseract
tar -C tesseract --strip-components=1 -xzf "$WORK/src/tesseract-5.5.3.tar.gz"
# zig's mingw import libs are lowercase and its linker is case-sensitive; tesseract's
# CLI links Ws2_32 and fails with "unable to find dynamic system library 'Ws2_32'".
sed -i 's/set(LIB_Ws2_32 Ws2_32)/set(LIB_Ws2_32 ws2_32)/' tesseract/CMakeLists.txt
"$CMAKE" -S tesseract -B tesseract/b "${CROSS[@]}" \
	-DCMAKE_PREFIX_PATH="$PREFIX" \
	-DBUILD_SHARED_LIBS=OFF \
	-DBUILD_TRAINING_TOOLS=OFF -DBUILD_TESTS=OFF \
	-DDISABLE_CURL=ON -DDISABLE_ARCHIVE=ON -DDISABLE_TIFF=ON \
	-DGRAPHICS_DISABLED=ON -DDISABLED_LEGACY_ENGINE=ON \
	-DOPENMP_BUILD=OFF -DSW_BUILD=OFF -DENABLE_LTO=OFF \
	-DUSE_SYSTEM_ICU=OFF -DLEPT_TIFF_RESULT=1
"$CMAKE" --build tesseract/b -j"$J"
"$CMAKE" --install tesseract/b
fi

# ---- post-install fixups the cgo link relies on -------------------------------------
case "$SYSNAME" in
Windows)
	# cmake names the Windows static libs non-standardly; alias them to what cgo's -l
	# flags expect. libleptonica-1.87.0.dll IS a static library — `file` says
	# "current ar archive" — misnamed, not shared.
	cd "$PREFIX/lib"
	alias_lib() { # alias_lib <installed-name> <expected-name>
		if [ -e "$1" ] && [ ! -e "$2" ]; then ln -sf "$1" "$2"; fi
	}
	alias_lib libzs.a libz.a
	alias_lib libzlibstatic.a libz.a
	alias_lib libleptonica-1.87.0.dll libleptonica.a
	alias_lib libtesseract55.a libtesseract.a
	;;
Darwin)
	# Go hard-adds -lresolv and -framework CoreFoundation to darwin external links
	# (net/cgo_unix_cgo_res.go, runtime/cgo) even when nothing uses them. Empty tapi
	# stubs satisfy the linker; the OS provides the real libraries at run time —
	# runtime/cgo's CoreFoundation usage is TARGET_OS_IPHONE-only.
	tbd_target="$PROC-macos"
	mkdir -p "$PREFIX/Frameworks/CoreFoundation.framework"
	cat > "$PREFIX/lib/libresolv.tbd" <<EOF
--- !tapi-tbd
tbd-version: 4
targets: [ $tbd_target ]
install-name: '/usr/lib/libresolv.9.dylib'
current-version: 1.0
...
EOF
	cat > "$PREFIX/Frameworks/CoreFoundation.framework/CoreFoundation.tbd" <<EOF
--- !tapi-tbd
tbd-version: 4
targets: [ $tbd_target ]
install-name: '/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation'
current-version: 1.0
...
EOF
	;;
esac

phase "done"
ls -la "$PREFIX/lib"
echo "build-cstack: $TARGET prefix ready; cgo env: $0 env $TARGET $WORK"
