// Command fakehelper is a native helper written in Go on the Go codec, so the client's tests run
// without a C++ toolchain. Each test builds it with the values it needs:
//
//	-ldflags "-X main.digest=<d> -X main.protocol=<n> -X main.mode=<m> -X main.assets=<a,b>"
//
// assets is the asset sha256 its handshake reports after the digest, comma-separated.
//
// Like native/frame, it captures each request's diagnostics in the file NATIVE_HELPER_CAPTURE
// names (a temporary file when unset), truncated at the request's start, and answers with that
// file's content bounded by nativehelper.BoundStderr. Its own frame errors go to the stderr pipe.
//
// mode is the misbehaviour, on selftest-stdout and selftest-sleep (the handshake is always good):
//
//	""             none
//	"crash"        write to the capture, then exit 3
//	"flood"        write 1 MiB to the capture, then answer normally
//	"flood-crash"  write 1 MiB to the capture, then a last line to the stderr pipe, then exit 5
//	"garbage"      answer with bytes that are not a frame
//	"wrong-id"     answer with another request id
//	"no-fields"    answer with no fields
//	"one-field"    answer with the status field alone
//	"empty-status" answer with an empty status field
//	"status-7"     answer with status byte 7
//	"ignore-eof"   keep running after stdin closes
//
// and "crash-on-helper-op" exits 3 on any op from nativehelper.FirstHelperOp on, after a good
// handshake: a helper that comes up and then dies on a page.
//
// FAKEHELPER_START_LOG, when set, names a file it appends its pid to at each start.
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
)

var (
	digest   = "fake"
	protocol = "" // empty: nativehelper.ProtocolVersion. -X sets only a constant-initialised string.
	mode     = ""
	assets   = ""
)

// floodLine is the last line of a flood; the tests look for it.
const floodLine = "fakehelper: the flood's last line\n"

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
	capture, err := openCapture()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fakehelper: opening the capture file:", err)
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
		if err := capture.Truncate(0); err != nil {
			fmt.Fprintln(os.Stderr, "fakehelper: truncating the capture file:", err)
			os.Exit(4)
		}
		if _, err := capture.Seek(0, io.SeekStart); err != nil {
			fmt.Fprintln(os.Stderr, "fakehelper: rewinding the capture file:", err)
			os.Exit(4)
		}
		resp := answer(req, capture)
		if resp == nil {
			continue
		}
		if err := nativehelper.WriteFrame(os.Stdout, nativehelper.Frame{Version: uint16(v), Op: req.Op, ID: req.ID + idShift(req.Op), Fields: resp}); err != nil {
			fmt.Fprintln(os.Stderr, "fakehelper: writing a frame:", err)
			os.Exit(5)
		}
	}
}

func openCapture() (*os.File, error) {
	if p := os.Getenv(nativehelper.EnvCapture); p != "" {
		return os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	}
	return os.CreateTemp("", "fakehelper-capture-")
}

func idShift(op nativehelper.Op) uint32 {
	if mode == "wrong-id" && op != nativehelper.OpHandshake {
		return 1
	}
	return 0
}

// answer runs one request with diag as its stderr and returns the response's fields, or nil
// when the mode writes something other than a frame.
func answer(req nativehelper.Frame, diag *os.File) [][]byte {
	status, result := byte(nativehelper.StatusOK), [][]byte{}
	if mode == "crash-on-helper-op" && req.Op >= nativehelper.FirstHelperOp {
		fmt.Fprintln(diag, "fakehelper: crashing on a helper op as built to")
		os.Exit(3)
	}
	misbehave := req.Op == nativehelper.OpSelftestStdout || req.Op == nativehelper.OpSelftestSleep
	if misbehave {
		switch mode {
		case "crash":
			fmt.Fprintln(diag, "fakehelper: crashing as built to")
			os.Exit(3)
		case "flood", "flood-crash":
			line := strings.Repeat("x", 1023) + "\n"
			diag.WriteString(strings.Repeat(line, 1024))
			diag.WriteString(floodLine)
			if mode == "flood-crash" {
				fmt.Fprintln(os.Stderr, "fakehelper: writing a frame failed, as built to")
				os.Exit(5)
			}
		case "garbage":
			os.Stdout.Write([]byte("this is not a frame header"))
			return nil
		case "no-fields":
			return [][]byte{}
		case "one-field":
			return [][]byte{{byte(nativehelper.StatusOK)}}
		case "empty-status":
			return [][]byte{{}, {}}
		case "status-7":
			return [][]byte{{7}, {}}
		}
	}
	switch req.Op {
	case nativehelper.OpHandshake:
		result = [][]byte{[]byte(digest)}
		if assets != "" {
			for _, a := range strings.Split(assets, ",") {
				result = append(result, []byte(a))
			}
		}
	case nativehelper.OpSelftestStdout:
		fmt.Fprint(diag, "printf\ncout\nfwrite\nwrite1\n")
		all := make([]byte, 256)
		for i := range all {
			all[i] = byte(i)
		}
		result = [][]byte{all}
	case nativehelper.OpSelftestSleep:
		if len(req.Fields) != 1 || len(req.Fields[0]) != 4 {
			status, result = byte(nativehelper.StatusOpError), [][]byte{[]byte("selftest-sleep takes one 4-byte field")}
			break
		}
		time.Sleep(time.Duration(binary.LittleEndian.Uint32(req.Fields[0])) * time.Millisecond)
	default:
		status, result = byte(nativehelper.StatusUnknownOp), [][]byte{[]byte(req.Op.String())}
	}
	return append([][]byte{{status}, captured(diag)}, result...)
}

// captured is the request's capture, bounded as native/frame bounds it.
func captured(diag *os.File) []byte {
	if _, err := diag.Seek(0, io.SeekStart); err != nil {
		return []byte("fakehelper: rewinding the capture file: " + err.Error())
	}
	var b bytes.Buffer
	if _, err := io.Copy(&b, diag); err != nil {
		return []byte("fakehelper: reading the capture file: " + err.Error())
	}
	return nativehelper.BoundStderr(b.Bytes(), 0)
}
