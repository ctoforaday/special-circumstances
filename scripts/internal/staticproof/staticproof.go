// Package staticproof asserts the contract every published binary meets: the per-format
// staticness below, and, for a native helper, a size bound (helperMaxBytes).
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

// helperMaxBytes is each native helper's size bound: about 1.25 times its largest stripped build on
// any published target. A helper linked with its debug information is several times its stripped
// size (feov-embed linux-amd64: 178,839,192 B with DWARF, 6,366,560 B stripped; feov-ocr
// linux-amd64: 46,530,584 B and 9,285,672 B), so a link that stops stripping fails here, on the
// native-build leg that made it and again on the release's dist.
var helperMaxBytes = map[string]int64{
	"feov-embed": 8 << 20,
	"feov-ocr":   12 << 20,
}

// Check verifies that the binary at path satisfies the published-binary contract for the given
// artifact: a helper's size bound, then the staticness contract of its format.
func Check(artifact, path string) error {
	if err := checkSize(artifact, path); err != nil {
		return err
	}
	if f, err := elf.Open(path); err == nil {
		defer f.Close()
		return checkELF(path, f)
	}
	if f, err := pe.Open(path); err == nil {
		defer f.Close()
		return checkPE(artifact, path, f)
	}
	if f, err := macho.Open(path); err == nil {
		defer f.Close()
		return checkMachO(artifact, path, f)
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
