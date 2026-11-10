# The C stack, and why native helpers are built with zig

FEOV's native helpers (`feov-ocr`, and `feov-embed`) run as standalone executables beside the pure-Go binaries. `feov-ocr` links a static C/C++ stack: tesseract and leptonica (with zlib and libpng) for local OCR. Every published target gets self-contained helper binaries with no runtime library to install.

| Piece | Where |
|---|---|
| Pinned sources and toolchain (zig, cmake), each by sha256 | `plugins/frank-exchange-of-views/native/PINS.txt` and `native/<helper>/PINS.txt` |
| Build for one target; check output staticness | `plugins/frank-exchange-of-views/native/build.sh` |
| Staticness check (ELF, PE and Mach-O import records) | `scripts/staticproof/` and `scripts/internal/staticproof/` |
| Helper builds and CI verification | `.github/workflows/hooks.yml` (`native-scope`, `native-build`, `native-run`, `native-fuzz`) |

## How a target is built

`native/build.sh <helper> <target> [outdir] [workdir]` compiles the helper using the pinned zig toolchain (`zig cc` / `zig c++ -target <triple>`):

| Target | Triple |
|---|---|
| linux-amd64 / linux-arm64 | `x86_64-linux-musl` / `aarch64-linux-musl` |
| windows-amd64 / windows-arm64 | `x86_64-windows-gnu` / `aarch64-windows-gnu` |
| darwin-arm64 | `aarch64-macos` |

Every published helper is proven static by `staticproof` against its target's allowlist. The Go binaries in all plugins compile with `CGO_ENABLED=0` and are also checked by `staticproof` before release.

The same sources read a page differently on different targets, so a reading belongs to its target. `tessocr.Identity()` ends with `+target@<goos>-<goarch>`, so a reading cached on one target is read again on another. The corpus holds one golden set per target (gblock, 2026-11-10): `reading.<goos>-<goarch>.golden` beside each page. `native-run` compares each target with its own set and uploads its readings as `corpus-readings-<target>`; a new or moved set is committed from that artifact after its diff is read.

## Why zig, and not plain clang or gcc

`zig cc` and `zig c++` are clang and lld. Zig contributes the rest of what a cross-build needs, in one pinned download:

- **A libc for every target.** musl for Linux, mingw-w64 for Windows, and stub libraries for macOS. `zig c++ -target aarch64-windows-gnu` works from a Linux host with nothing else installed.
- **The C++ runtime and compiler support.** libc++, libunwind and compiler-rt, built for each target on demand. That is what makes the result fully static.
- **macOS without Apple's SDK.** Zig links against `.tbd` stubs, which describe a system library's exported symbols (`native/build.sh` writes the ones it needs: libresolv, CoreFoundation, Security). Apple's SDK licence permits its use only on Apple hardware, so a Linux host cannot use it.

Plain clang can target all triples too, but needs a sysroot for each one that we would fetch, pin and keep consistent ourselves:

- **Windows:** mingw-w64, or Microsoft's SDK, whose licence restricts redistribution.
- **macOS:** Apple's SDK, which runs into the licence above. osxcross, the usual workaround, still needs the real SDK.
- **Linux arm64:** a cross glibc or musl.

GCC needs a separate cross-toolchain built per target, and cannot target macOS at all.

**The cost:** zig is pre-1.0, and its toolchain has quirks. Each workaround in `native/build.sh` is commented with the failure it prevents. Read those comments before changing a flag. Bump zig's pin in `PINS.txt` deliberately, and rebuild all targets when you do.

## Adding a C or C++ library

A library joins a native helper only if it builds under the helper's toolchain for every published target and the resulting helper passes `staticproof`.
