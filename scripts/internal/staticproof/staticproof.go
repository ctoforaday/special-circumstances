// Package staticproof asserts the contract every published binary meets: the per-format
// staticness below, and, for a native helper, that it is stripped and within a size bound
// (helperMaxBytes).
//
//	linux (ELF):      fully static — no interpreter, no dynamic section at all.
//	windows (PE):     imports only allowed system libraries (KERNEL32.dll, NTDLL.dll,
//	                  and api-ms-win-crt-* api-sets for Go binaries and feov-ocr;
//	                  feov-embed additionally permits advapi32.dll, dbghelp.dll,
//	                  and api-ms-win-core-path-l1-1-0.dll).
//	darwin (Mach-O):  load commands reference only OS-provided libraries — libSystem.B,
//	                  libresolv.9, CoreFoundation, Security.
package staticproof

import (
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"os"
	"strings"
)

// helperMaxBytes is each native helper's size bound: about 1.3 times its largest stripped build on
// any published target (feov-embed: 6,366,304 B on linux-amd64; feov-ocr: 9,762,360 B on
// darwin-arm64). It is the backstop behind the strip checks, not the strip check: an unstripped
// feov-ocr is 5 times its stripped size on linux-amd64 (46,530,584 B against 9,285,672 B) but
// only 1.13 times on darwin-arm64 (10,993,848 B against 9,762,360 B), inside the bound.
var helperMaxBytes = map[string]int64{
	"feov-embed": 8 << 20,
	"feov-ocr":   12 << 20,
}

// Check verifies that the binary at path satisfies the published-binary contract for the given
// artifact: a helper's size bound, then the staticness contract of its format, then a helper's
// strip.
func Check(artifact, path string) error {
	if err := checkSize(artifact, path); err != nil {
		return err
	}
	if f, err := elf.Open(path); err == nil {
		defer f.Close()
		if err := checkELF(path, f); err != nil {
			return err
		}
		return checkELFStripped(artifact, path, f)
	}
	if f, err := pe.Open(path); err == nil {
		defer f.Close()
		return checkPE(artifact, path, f)
	}
	if f, err := macho.Open(path); err == nil {
		defer f.Close()
		if err := checkMachO(artifact, path, f); err != nil {
			return err
		}
		return checkMachOStripped(artifact, path, f)
	}
	return fmt.Errorf("%s: not ELF, PE or Mach-O — nothing to prove", path)
}

// checkSize refuses a native helper larger than its bound; other artifacts carry none.
func checkSize(artifact, path string) error {
	bound, ok := helperMaxBytes[artifact]
	if !ok {
		return nil
	}
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fi.Size() > bound {
		return fmt.Errorf("%s: %s is %d bytes, over its bound of %d — a helper linked with debug information or symbols", path, artifact, fi.Size(), bound)
	}
	return nil
}

// machOMaxLocalSyms is the most local symbols a stripped helper's Mach-O keeps: the stripped
// darwin-arm64 helpers keep 232 (feov-embed) and 253 (feov-ocr), the unstripped ones 47,633 and
// 96,852.
const machOMaxLocalSyms = 2000

// A native helper linked with its symbols or debug information is refused: an ELF that keeps a
// .symtab or any .debug_* section, or a Mach-O that keeps more than machOMaxLocalSyms local
// symbols or a __DWARF segment. A PE needs no rule: neither the zig nor the MSVC link writes COFF symbols, so a
// stripped and an unstripped helper both carry none, and the size bound is its only check. Other
// artifacts are Go commands, stripped by their own build's -s -w.

func checkELFStripped(artifact, path string, f *elf.File) error {
	if _, helper := helperMaxBytes[artifact]; !helper {
		return nil
	}
	for _, s := range f.Sections {
		if s.Name == ".symtab" || strings.HasPrefix(s.Name, ".debug_") {
			return fmt.Errorf("%s: %s keeps its %s section — a helper linked without stripping", path, artifact, s.Name)
		}
	}
	return nil
}

func checkMachOStripped(artifact, path string, f *macho.File) error {
	if _, helper := helperMaxBytes[artifact]; !helper {
		return nil
	}
	if f.Dysymtab == nil {
		return fmt.Errorf("%s: %s has no LC_DYSYMTAB — its local symbols cannot be counted", path, artifact)
	}
	if n := f.Dysymtab.Nlocalsym; n > machOMaxLocalSyms {
		return fmt.Errorf("%s: %s keeps %d local symbols, over %d — a helper linked without stripping", path, artifact, n, machOMaxLocalSyms)
	}
	if f.Segment("__DWARF") != nil {
		return fmt.Errorf("%s: %s keeps its __DWARF segment — a helper linked without stripping", path, artifact)
	}
	return nil
}

func checkELF(path string, f *elf.File) error {
	for _, p := range f.Progs {
		switch p.Type {
		case elf.PT_INTERP:
			return fmt.Errorf("%s: ELF requests an interpreter — not statically linked", path)
		case elf.PT_DYNAMIC:
			libs, _ := f.ImportedLibraries()
			return fmt.Errorf("%s: ELF has a dynamic section (needs %v) — not statically linked", path, libs)
		}
	}
	return nil
}

func checkPE(artifact, path string, f *pe.File) error {
	// debug/pe reads a file with no MZ header as a bare COFF object, so any bytes at all, zeros
	// included, open as a PE that imports nothing; an executable image has an optional header.
	if f.OptionalHeader == nil {
		return fmt.Errorf("%s: not an executable image (a PE with no optional header)", path)
	}
	syms, err := f.ImportedSymbols()
	if err != nil {
		return fmt.Errorf("%s: reading PE imports: %w", path, err)
	}
	libs := map[string]bool{}
	for _, s := range syms {
		if i := strings.LastIndex(s, ":"); i >= 0 {
			libs[strings.ToLower(s[i+1:])] = true
		}
	}
	var bad []string
	for lib := range libs {
		if !PELibAllowed(artifact, lib) {
			bad = append(bad, lib)
		}
	}
	if len(bad) > 0 {
		return fmt.Errorf("%s: PE imports beyond allowed set for %s: %v", path, artifact, bad)
	}
	return nil
}

// PELibAllowed checks if a library import is permitted for the given artifact.
func PELibAllowed(artifact, lib string) bool {
	lib = strings.ToLower(lib)
	if lib == "kernel32.dll" || lib == "ntdll.dll" || strings.HasPrefix(lib, "api-ms-win-crt-") {
		return true
	}
	if artifact == "feov-embed" {
		// advapi32.dll: event tracing
		// dbghelp.dll: stack traces
		// api-ms-win-core-path-l1-1-0.dll: path handling
		switch lib {
		case "advapi32.dll", "dbghelp.dll", "api-ms-win-core-path-l1-1-0.dll":
			return true
		}
	}
	return false
}

// MachOAllowed is the set of OS-provided libraries permitted on Darwin.
var machOAllowed = map[string]bool{
	"/usr/lib/libSystem.B.dylib": true,
	"/usr/lib/libresolv.9.dylib": true,
	"/System/Library/Frameworks/CoreFoundation.framework/Versions/A/CoreFoundation": true,
	"/System/Library/Frameworks/Security.framework/Versions/A/Security":             true, // crypto/x509 links it on darwin
}

// MachOLibAllowed checks if a Mach-O load command library is permitted.
func MachOLibAllowed(artifact, lib string) bool {
	return machOAllowed[lib]
}

func checkMachO(artifact, path string, f *macho.File) error {
	libs, err := f.ImportedLibraries()
	if err != nil {
		return fmt.Errorf("%s: reading Mach-O load commands: %w", path, err)
	}
	var bad []string
	for _, lib := range libs {
		if !MachOLibAllowed(artifact, lib) {
			bad = append(bad, lib)
		}
	}
	if len(bad) > 0 {
		return fmt.Errorf("%s: Mach-O references non-OS libraries: %v", path, bad)
	}
	return nil
}
