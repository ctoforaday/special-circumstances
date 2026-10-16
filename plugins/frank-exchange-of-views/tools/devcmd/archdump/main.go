// Command archdump migrates every archived run through this binary and writes each one's
// assembled report set, so two builds of the tool can be compared document by document over real
// records: the same archives dumped by each, then `diff -r` of the two output directories.
//
// For each run-archive tarball it writes <out>/<run>/ holding every document the assembler wrote
// (report.Files(), the index and the site), plus two files a reader counts:
//
//	REFUSALS        one line per event the migration refused; an empty file when it refused none
//	ASSEMBLE_ERROR  present only when the migrated record did not assemble, holding the error
//
// A refusal or an assembly failure is OUTPUT, never a stop: the next archive is still dumped, so
// one bad run cannot hide the rest. The command exits non-zero when it could not read an archive
// or write a file, and then names which.
//
// Run from the tools module; a relative -archives resolves against the repository root:
//
//	go run ./devcmd/archdump -archives run-archive -out ~/.claude/scratch/<task>/before
package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

func main() {
	archives := flag.String("archives", "run-archive", "the directory of run tarballs (*.tar.gz); relative to the repository root unless absolute")
	out := flag.String("out", "", "the directory to write one subdirectory per archive into; refused if it already holds anything")
	flag.Parse()
	if err := run(*archives, *out); err != nil {
		fmt.Fprintln(os.Stderr, "archdump:", err)
		os.Exit(1)
	}
}

func run(archives, out string) error {
	if out == "" {
		return fmt.Errorf("-out is required: the directory the dump is written into")
	}
	if !filepath.IsAbs(archives) {
		root, err := repotree.Root()
		if err != nil {
			return err
		}
		archives = filepath.Join(root, archives)
	}
	tarballs, err := filepath.Glob(filepath.Join(archives, "*.tar.gz"))
	if err != nil {
		return err
	}
	if len(tarballs) == 0 {
		return fmt.Errorf("no *.tar.gz under %s", archives)
	}
	sort.Strings(tarballs)
	if ents, err := os.ReadDir(out); err == nil && len(ents) > 0 {
		return fmt.Errorf("%s already holds %d entries: a dump goes into an empty directory, so two dumps never mix", out, len(ents))
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	work, err := os.MkdirTemp("", "archdump-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)
	for _, tb := range tarballs {
		name := strings.TrimSuffix(filepath.Base(tb), ".tar.gz")
		if err := dump(tb, filepath.Join(work, name), filepath.Join(out, name)); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		fmt.Println(name)
	}
	return nil
}

// dump migrates one tarball and writes its assembled set into dst.
func dump(tarball, work, dst string) error {
	src, to := filepath.Join(work, "src"), filepath.Join(work, "to")
	if err := untar(tarball, src); err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	m, merr := migrate.Migrate(src, to, migrate.Entries(), migrate.Options{})
	var refusals []string
	if m != nil {
		for _, r := range m.Refusals {
			refusals = append(refusals, fmt.Sprintf("%d %s: %s", r.OldID, r.Word, strings.ReplaceAll(r.Err, "\n", " ")))
		}
	}
	if merr != nil && len(refusals) == 0 {
		refusals = append(refusals, "migrate: "+strings.ReplaceAll(merr.Error(), "\n", " "))
	}
	if err := os.WriteFile(filepath.Join(dst, "REFUSALS"), []byte(lines(refusals)), 0o644); err != nil {
		return err
	}
	if err := recordsql.CloseUnder(to); err != nil {
		return err
	}
	if merr != nil {
		return nil
	}
	r, err := record.OpenRun(to)
	if err == nil {
		_, err = report.Assemble(r)
	}
	if cerr := recordsql.CloseUnder(to); cerr != nil && err == nil {
		err = cerr
	}
	if err != nil {
		return os.WriteFile(filepath.Join(dst, "ASSEMBLE_ERROR"), []byte(err.Error()+"\n"), 0o644)
	}
	for _, f := range append(report.Files(), report.FileIndex, "report.html") {
		b, err := os.ReadFile(filepath.Join(to, f))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dst, f), b, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func lines(ls []string) string {
	if len(ls) == 0 {
		return ""
	}
	return strings.Join(ls, "\n") + "\n"
}

// untar unpacks a run tarball into dir, refusing an entry that would land outside it.
func untar(tarball, dir string) error {
	f, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		rel := filepath.Clean(h.Name)
		if strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
			return fmt.Errorf("tar entry escapes the run directory: %q", h.Name)
		}
		p := filepath.Join(dir, rel)
		switch h.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(p, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				return err
			}
			o, err := os.Create(p)
			if err != nil {
				return err
			}
			if _, err := io.Copy(o, tr); err != nil {
				o.Close()
				return err
			}
			if err := o.Close(); err != nil {
				return err
			}
		}
	}
}
