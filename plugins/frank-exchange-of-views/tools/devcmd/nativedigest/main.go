// Command nativedigest writes internal/nativehelper/digest_gen.go from native/: each helper's
// source digest, the sha256 of each asset its ASSETS file names per arch, and feov-ocr's engine
// pins. A binary accepts a helper only when the helper's handshake reports the digest compiled
// in, so the constant must be the sources' and nothing else's.
//
// It lives in this module because it reads internal/nativehelper/nativesrc and finds the
// checkout through internal/repotree.
//
// Run from the tools module:
//
//	go run ./devcmd/nativedigest                 write digest_gen.go
//	go run ./devcmd/nativedigest -check          exit 1 when the committed digest_gen.go is stale
//	go run ./devcmd/nativedigest -print <helper> print one helper's digest, for its build
//	go run ./devcmd/nativedigest -assets <helper> -arch <arch>
//	                                             print `<installed file> <pinned export>` for each
//	                                             ASSETS line, the files a build for that arch stages
//
// -native names another native/ tree, for the tests' copies.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper/nativesrc"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "nativedigest:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("nativedigest", flag.ContinueOnError)
	check := fs.Bool("check", false, "exit 1 when the committed digest_gen.go is not what native/ renders")
	print := fs.String("print", "", "print this helper's source digest and exit")
	assets := fs.String("assets", "", "print this helper's `<installed file> <pinned export>` pairs for -arch and exit")
	arch := fs.String("arch", "", "the GOARCH -assets resolves the exports for")
	native := fs.String("native", "", "the native/ tree to read (default: the checkout's)")
	out := fs.String("out", "", "the file to write or check (default: the checkout's internal/nativehelper/digest_gen.go)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected arguments %q", fs.Args())
	}
	if *native == "" {
		p, err := repotree.Plugin("native")
		if err != nil {
			return err
		}
		*native = p
	}
	if *assets != "" {
		if *arch == "" {
			return fmt.Errorf("-assets needs -arch")
		}
		if err := isHelper(*native, *assets); err != nil {
			return err
		}
		as, err := nativesrc.Assets(*native, *assets)
		if err != nil {
			return err
		}
		for _, a := range as {
			export, ok := a.Export[*arch]
			if !ok {
				return fmt.Errorf("%s/%s/ASSETS names no %s export of %s", *native, *assets, *arch, a.File)
			}
			fmt.Println(a.File, export)
		}
		return nil
	}
	if *arch != "" {
		return fmt.Errorf("-arch is read only with -assets")
	}
	if *print != "" {
		if err := isHelper(*native, *print); err != nil {
			return err
		}
		d, err := nativesrc.Digest(*native, *print)
		if err != nil {
			return err
		}
		fmt.Println(d)
		return nil
	}
	if *out == "" {
		p, err := repotree.Plugin("tools", "internal", "nativehelper", "digest_gen.go")
		if err != nil {
			return err
		}
		*out = p
	}
	want, err := nativesrc.Render(*native)
	if err != nil {
		return err
	}
	if *check {
		got, err := os.ReadFile(*out)
		if err != nil {
			return err
		}
		if !bytes.Equal(bytes.ReplaceAll(got, []byte("\r\n"), []byte("\n")), want) {
			return fmt.Errorf("%s is stale: native/ renders other pins. Run `go run ./devcmd/nativedigest` in plugins/frank-exchange-of-views/tools", *out)
		}
		return nil
	}
	return os.WriteFile(*out, want, 0o644)
}

// isHelper refuses a name that is not a helper under native.
func isHelper(native, name string) error {
	names, err := nativesrc.Helpers(native)
	if err != nil {
		return err
	}
	for _, n := range names {
		if n == name {
			return nil
		}
	}
	return fmt.Errorf("%s holds no helper %q (a helper is a directory holding an ASSETS file)", native, name)
}
