// Command fx is the program the compute tests run: one binary the test builds, so every arm runs on
// every platform without an interpreter. Its first argument picks what it does.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	args := os.Args[1:]
	switch args[0] {
	case "print":
		fmt.Println(strings.Join(args[1:], " "))
	case "random":
		fmt.Println(time.Now().UnixNano())
	case "exitcount":
		// Reads a counter outside the copies and exits with its parity, so two runs exit apart.
		n := 0
		if b, err := os.ReadFile(args[1]); err == nil {
			n, _ = strconv.Atoi(strings.TrimSpace(string(b)))
		}
		os.WriteFile(args[1], []byte(strconv.Itoa(n+1)), 0o644)
		fmt.Println("counted")
		os.Exit(n % 2)
	case "out":
		os.WriteFile(filepath.Join("out", args[1]), []byte(args[2]), 0o644)
		fmt.Println("wrote", args[1])
	case "outrandom":
		os.WriteFile(filepath.Join("out", "r.txt"), []byte(strconv.FormatInt(time.Now().UnixNano(), 10)), 0o644)
		fmt.Println("wrote r.txt")
	case "read":
		b, err := os.ReadFile(args[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Print(string(b))
	case "where":
		// As a program printing its own file and working directory and then raising: getcwd is the
		// system call, which resolves every symlink, as an interpreter's does.
		wd, err := syscall.Getwd()
		if err != nil {
			panic(err)
		}
		fmt.Println("cwd:", wd)
		fmt.Println("file:", filepath.Join(wd, "data.txt"))
		fmt.Fprintln(os.Stderr, "raised: the claim does not hold")
		os.Exit(1)
	case "stderr":
		fmt.Fprintln(os.Stderr, strings.Join(args[1:], " "))
	case "call":
		if m := os.Getenv("FEOV_COMPUTE_RERUN"); m != "" {
			f, _ := os.OpenFile(m, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
			fmt.Fprintln(f, "feov-record show board --json")
			f.Close()
		}
		fmt.Println("0")
	case "readonly":
		os.Mkdir("ro", 0o755)
		os.WriteFile(filepath.Join("ro", "f"), []byte("x"), 0o444)
		os.Chmod("ro", 0o555)
		fmt.Println("ok")
	case "env":
		for _, k := range []string{"SOURCE_DATE_EPOCH", "PYTHONHASHSEED", "TZ", "LC_ALL", "PYTHONUTF8", "PWD"} {
			fmt.Printf("%s=%s\n", k, os.Getenv(k))
		}
		rc, _ := os.ReadFile(os.Getenv("MATPLOTLIBRC"))
		fmt.Printf("MATPLOTLIBRC holds %q\n", rc)
		fmt.Printf("checker marker set: %v\n", os.Getenv("FEOV_COMPUTE_RERUN") != "")
	case "sleep":
		d, _ := time.ParseDuration(args[1])
		time.Sleep(d)
	case "fork":
		self, _ := os.Executable()
		c := exec.Command(self, "sleep", "20s")
		c.Stdout = os.Stdout
		c.Start()
		fmt.Println("done")
	default:
		fmt.Fprintln(os.Stderr, "fx: unknown mode", args[0])
		os.Exit(2)
	}
}
