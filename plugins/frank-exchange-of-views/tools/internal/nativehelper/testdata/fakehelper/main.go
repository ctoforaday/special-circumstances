// Command fakehelper is a native helper written in Go on the Go codec, so the client's tests run
// without a C++ toolchain. Each test builds it with the values it needs:
//
//	-ldflags "-X main.digest=<d> -X main.protocol=<n> -X main.mode=<m>"
//
// mode is the misbehaviour: "" (none), "crash" (exit 3 on the first request after the
// handshake, after writing to stderr), "garbage" (answer selftest-stdout with bytes that are not
// a frame), "ignore-eof" (keep running after stdin closes).
// FAKEHELPER_START_LOG, when set, names a file it appends one line to at each start.
package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
)

var (
	digest   = "fake"
	protocol = "" // empty: nativehelper.ProtocolVersion. -X sets only a constant-initialised string.
	mode     = ""
)

func main() {
	if log := os.Getenv("FAKEHELPER_START_LOG"); log != "" {
		f, err := os.OpenFile(log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			fmt.Fprintln(f, os.Getpid())
			f.Close()
		}
	}
	v, err := int(nativehelper.ProtocolVersion), error(nil)
	if protocol != "" {
		v, err = strconv.Atoi(protocol)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "fakehelper: bad protocol version", protocol)
		os.Exit(2)
	}
	for {
		req, err := nativehelper.ReadFrame(os.Stdin)
		if err == io.EOF {
			if mode == "ignore-eof" {
				time.Sleep(time.Hour)
			}
			os.Exit(0)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "fakehelper:", err)
			os.Exit(4)
		}
		status, stderr, result := byte(nativehelper.StatusOK), []byte{}, [][]byte{}
		switch req.Op {
		case nativehelper.OpHandshake:
			result = [][]byte{[]byte(digest)}
		case nativehelper.OpSelftestStdout:
			if mode == "crash" {
				fmt.Fprintln(os.Stderr, "fakehelper: crashing as built to")
				os.Exit(3)
			}
			if mode == "garbage" {
				os.Stdout.Write([]byte("this is not a frame header"))
				continue
			}
			stderr = []byte("printf\ncout\nfwrite\nwrite1\n")
			all := make([]byte, 256)
			for i := range all {
				all[i] = byte(i)
			}
			result = [][]byte{all}
		case nativehelper.OpSelftestSleep:
			if mode == "crash" {
				fmt.Fprintln(os.Stderr, "fakehelper: crashing as built to")
				os.Exit(3)
			}
			if len(req.Fields) != 1 || len(req.Fields[0]) != 4 {
				status, result = byte(nativehelper.StatusOpError), [][]byte{[]byte("selftest-sleep takes one 4-byte field")}
				break
			}
			time.Sleep(time.Duration(binary.LittleEndian.Uint32(req.Fields[0])) * time.Millisecond)
		default:
			status, result = byte(nativehelper.StatusUnknownOp), [][]byte{[]byte(req.Op.String())}
		}
		fields := append([][]byte{{status}, stderr}, result...)
		if err := nativehelper.WriteFrame(os.Stdout, nativehelper.Frame{Version: uint16(v), Op: req.Op, ID: req.ID, Fields: fields}); err != nil {
			os.Exit(5)
		}
	}
}
