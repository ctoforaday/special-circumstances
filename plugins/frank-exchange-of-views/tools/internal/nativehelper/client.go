// Package nativehelper runs a native helper executable (feov-ocr, feov-embed) as a child process
// and speaks the frame protocol to it: locate, spawn, handshake, one request at a time under a
// deadline, and a named error for every way that can fail.
//
// A helper is built from native/<helper>/ with native/frame/'s codec. The client accepts a helper
// only when its handshake reports the source digest compiled into this binary (digest_gen.go),
// so a helper built from other sources is refused, never run.
package nativehelper

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"time"
)

const (
	// EnvDir names the one directory a helper is looked up in. `none` means no helper. Unset,
	// the helper is looked up beside the running executable. The name has no FEOV_ prefix: the
	// agent sessions' environment sweep unsets every FEOV_ variable before a test.
	EnvDir = "NATIVE_HELPER_DIR"
	// EnvCapture names the file a helper points fd 1 and fd 2 at during a request. The client
	// owns it, so a request that crashes the helper still leaves its stderr readable.
	EnvCapture = "NATIVE_HELPER_CAPTURE"
	// DefaultDeadline bounds one request, the handshake included, from the moment its frame is
	// written. Every request the client sends is sized to finish in a quarter of it.
	DefaultDeadline = 120 * time.Second
	// closeGrace is how long Close waits for the helper to exit on stdin EOF before killing it.
	closeGrace = 2 * time.Second
)

// Pins is what this binary was built to accept from one helper: its source digest, and the
// sha256 of each extra asset its ASSETS file names, per GOARCH.
type Pins struct {
	Digest string
	Assets []AssetPin
}

// AssetPin is one ASSETS line: the installed file, and its sha256 per GOARCH.
type AssetPin struct {
	File   string
	SHA256 map[string]string
}

// EnginePins is feov-ocr's engine identity, read by nativedigest from native/feov-ocr/PINS.txt.
type EnginePins struct {
	Tesseract      string
	Leptonica      string
	EngTraineddata string
}

// Response is an op's answer: its status, its result fields, and the helper's stderr captured
// during the request.
type Response struct {
	Status Status
	Result [][]byte
	Stderr []byte
}

// Client is one helper, spawned on first use and re-spawned after a failed request. Requests are
// serialised; the deadline does not count the wait for the lock.
type Client struct {
	name     string
	pins     Pins
	deadline time.Duration
	exeDir   func() (string, error)

	mu      sync.Mutex
	proc    *process
	latched error
}

// HelperDigest returns the compiled-in source digest for helper, or empty if unknown.
func HelperDigest(helper string) string {
	return helpers[helper].Digest
}

// HelperAssets returns the asset sha256s for helper on the current runtime.GOARCH.
func HelperAssets(helper string) []string {
	pins, ok := helpers[helper]
	if !ok {
		return nil
	}
	var res []string
	for _, a := range pins.Assets {
		if sha := a.SHA256[runtime.GOARCH]; sha != "" {
			res = append(res, sha)
		}
	}
	return res
}

// IsAbsent reports whether err means no helper came up (helper-absent,
// helper-digest-mismatch, or a failure before the handshake completed).
func IsAbsent(err error) bool {
	var ne *Error
	if errors.As(err, &ne) {
		return ne.Kind == ErrAbsent || ne.Kind == ErrDigestMismatch || ne.Handshake
	}
	return false
}

// Helpers returns the sorted names of helpers this binary carries pins for.
func Helpers() []string {
	names := make([]string, 0, len(helpers))
	for name := range helpers {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// New is the client for a helper this binary has pins for.
func New(helper string) (*Client, error) {
	pins, ok := helpers[helper]
	if !ok {
		return nil, fmt.Errorf("nativehelper: this binary carries no pins for a helper named %q", helper)
	}
	return newClient(helper, pins), nil
}

// NewBeside returns a client that locates the helper beside recordBin instead of beside the
// current process's executable.
func NewBeside(helper, recordBin string) (*Client, error) {
	pins, ok := helpers[helper]
	if !ok {
		return nil, fmt.Errorf("nativehelper: this binary carries no pins for a helper named %q", helper)
	}
	c := newClient(helper, pins)
	c.exeDir = func() (string, error) {
		abs := recordBin
		if a, err := filepath.Abs(recordBin); err == nil {
			abs = a
		}
		if p, err := filepath.EvalSymlinks(abs); err == nil {
			return filepath.Dir(p), nil
		}
		return filepath.Dir(abs), nil
	}
	return c, nil
}

func newClient(name string, pins Pins) *Client {
	return &Client{name: name, pins: pins, deadline: DefaultDeadline, exeDir: executableDir}
}

func executableDir() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", err
	}
	if p, err = filepath.EvalSymlinks(p); err != nil {
		return "", err
	}
	return filepath.Dir(p), nil
}

// Locate is the path the helper is looked up at: under NATIVE_HELPER_DIR when it is set, and only
// there; otherwise beside the running executable. It is ErrAbsent when no executable is there,
// and when NATIVE_HELPER_DIR is set to none or to the empty string.
func (c *Client) Locate() (string, error) {
	exe := c.name
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	if dir, set := os.LookupEnv(EnvDir); set {
		switch dir {
		case "none":
			return "", &Error{Kind: ErrAbsent, Helper: c.name, Detail: EnvDir + "=none runs without a helper"}
		case "":
			return "", &Error{Kind: ErrAbsent, Helper: c.name,
				Detail: EnvDir + " is set to the empty string; set it to the directory holding the helper, or to none"}
		}
		p := filepath.Join(dir, exe)
		if why := notExecutable(p); why != "" {
			return "", &Error{Kind: ErrAbsent, Helper: c.name, Path: p,
				Detail: why + "; " + EnvDir + " is set, and it is the only directory looked in"}
		}
		return p, nil
	}
	dir, err := c.exeDir()
	if err != nil {
		return "", &Error{Kind: ErrAbsent, Helper: c.name, Detail: "cannot find the running executable: " + err.Error()}
	}
	p := filepath.Join(dir, exe)
	if why := notExecutable(p); why != "" {
		return "", &Error{Kind: ErrAbsent, Helper: c.name, Path: p, Detail: why}
	}
	return p, nil
}

// notExecutable says why p is not an executable file, or is empty when it is one. On Windows a
// regular file under the .exe name is executable; elsewhere it needs an execute bit.
func notExecutable(p string) string {
	fi, err := os.Stat(p)
	switch {
	case err != nil:
		return "no such file"
	case !fi.Mode().IsRegular():
		return "not a regular file"
	case runtime.GOOS != "windows" && fi.Mode().Perm()&0o111 == 0:
		return "a file without an execute bit"
	}
	return ""
}

// Start spawns the helper and runs the handshake, unless it is already up. A handshake that
// fails latches its error: every later call returns it, and nothing is spawned again.
func (c *Client) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.start()
}

func (c *Client) start() error {
	if c.latched != nil {
		return c.latched
	}
	path, err := c.Locate()
	if err != nil {
		if c.proc != nil {
			_ = c.proc.close(c.name)
			c.proc = nil
		}
		return err
	}
	if c.proc != nil {
		if c.proc.path == path {
			return nil
		}
		_ = c.proc.close(c.name)
		c.proc = nil
	}
	p, err := spawn(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Error{Kind: ErrAbsent, Helper: c.name, Path: path, Detail: err.Error()}
		}
		c.latched = &Error{Kind: ErrCrashed, Helper: c.name, Path: path, Op: OpHandshake, Exit: "did not start", Detail: err.Error(), Handshake: true}
		return c.latched
	}
	if err := c.handshake(p); err != nil {
		p.kill()
		var e *Error
		if errors.As(err, &e) {
			e.Handshake = true
		}
		c.latched = err
		return err
	}
	c.proc = p
	return nil
}

func (c *Client) handshake(p *process) error {
	f, err := c.exchange(p, OpHandshake, nil)
	if err != nil {
		return err
	}
	if Status(f.Fields[0][0]) != StatusOK {
		return &Error{Kind: ErrProtocol, Helper: c.name, Path: p.path, Op: OpHandshake,
			Detail: "handshake status " + Status(f.Fields[0][0]).String(), Stderr: f.Fields[1]}
	}
	result := f.Fields[2:]
	if len(result) != 1+len(c.pins.Assets) {
		return &Error{Kind: ErrProtocol, Helper: c.name, Path: p.path, Op: OpHandshake,
			Detail: fmt.Sprintf("handshake carries %d fields, want the digest and %d asset sha256", len(result), len(c.pins.Assets))}
	}
	if got := string(result[0]); got != c.pins.Digest {
		return &Error{Kind: ErrDigestMismatch, Helper: c.name, Path: p.path, What: "source digest", Want: c.pins.Digest, Got: got}
	}
	for i, a := range c.pins.Assets {
		want := a.SHA256[runtime.GOARCH]
		if got := string(result[1+i]); got != want || want == "" {
			return &Error{Kind: ErrDigestMismatch, Helper: c.name, Path: p.path, What: a.File + " sha256 (" + runtime.GOARCH + ")", Want: want, Got: got}
		}
	}
	return nil
}

// Do sends one request and returns its response. A helper failure kills the helper and returns
// one of the five errors; the request is not retried, and the next request spawns a fresh
// helper. An op that ran and refused returns its response and an *OpError.
func (c *Client) Do(op Op, fields ...[]byte) (Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.start(); err != nil {
		return Response{}, err
	}
	f, err := c.exchange(c.proc, op, fields)
	if err != nil {
		c.proc = nil
		return Response{}, err
	}
	r := Response{Status: Status(f.Fields[0][0]), Stderr: f.Fields[1], Result: f.Fields[2:]}
	if r.Status != StatusOK {
		msg := ""
		if len(r.Result) > 0 {
			msg = string(r.Result[0])
		}
		return r, &OpError{Helper: c.name, Op: op, Status: r.Status, Message: msg, Stderr: r.Stderr}
	}
	return r, nil
}

// Close ends the helper: stdin closes, the helper has two seconds to exit, then it is killed. It
// returns an error when the helper did not exit cleanly on its own.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	p := c.proc
	c.proc = nil
	if p == nil {
		return nil
	}
	return p.close(c.name)
}

// Running reports whether a helper process is up for this client.
func (c *Client) Running() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.proc != nil
}

// KillProcessForTest kills the running helper process without clearing c.proc, simulating a crash.
func (c *Client) KillProcessForTest() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.proc != nil {
		c.proc.kill()
	}
}

// exchange writes one request and reads its response under the deadline, and checks the frame:
// the protocol version, the echoed id and op, the status and stderr fields. Any failure kills p.
func (c *Client) exchange(p *process, op Op, fields [][]byte) (Frame, error) {
	p.id++
	b, err := EncodeFrame(Frame{Version: ProtocolVersion, Op: op, ID: p.id, Fields: fields})
	if err != nil {
		p.kill()
		return Frame{}, &Error{Kind: ErrProtocol, Helper: c.name, Path: p.path, Op: op, Detail: "request refused before sending: " + err.Error()}
	}
	type result struct {
		f     Frame
		err   error
		wrote bool
	}
	ch := make(chan result, 1)
	start := time.Now()
	timer := time.NewTimer(c.deadline)
	defer timer.Stop()
	go func() {
		if _, err := p.stdin.Write(b); err != nil {
			ch <- result{err: err}
			return
		}
		f, err := ReadFrame(p.frames)
		ch <- result{f: f, err: err, wrote: true}
	}()
	var r result
	select {
	case r = <-ch:
	case <-timer.C:
		p.kill()
		<-ch
		return Frame{}, &Error{Kind: ErrDeadline, Helper: c.name, Path: p.path, Op: op, Elapsed: time.Since(start)}
	}
	elapsed := time.Since(start)
	if r.err != nil {
		if errors.Is(r.err, errFrame) {
			p.kill()
			return Frame{}, &Error{Kind: ErrProtocol, Helper: c.name, Path: p.path, Op: op, Elapsed: elapsed, Detail: r.err.Error()}
		}
		exit, stderr := p.reap()
		detail := "the helper's frame stream ended"
		if !r.wrote {
			detail = "the request could not be written: " + r.err.Error()
		} else if r.err != io.EOF {
			detail = "reading the response: " + r.err.Error()
		}
		return Frame{}, &Error{Kind: ErrCrashed, Helper: c.name, Path: p.path, Op: op, Elapsed: elapsed, Exit: exit, Stderr: stderr, Detail: detail}
	}
	f := r.f
	bad := ""
	switch {
	case f.Version != ProtocolVersion:
		bad = fmt.Sprintf("protocol version %d, this binary speaks %d", f.Version, ProtocolVersion)
	case f.ID != p.id:
		bad = fmt.Sprintf("response id %d to request %d", f.ID, p.id)
	case f.Op != op:
		bad = fmt.Sprintf("response op %s to a %s request", f.Op, op)
	case len(f.Fields) < 2 || len(f.Fields[0]) != 1:
		bad = "response without its status and stderr fields"
	case Status(f.Fields[0][0]) > StatusUnknownOp:
		bad = "response status " + Status(f.Fields[0][0]).String() + ", which the protocol does not define"
	}
	if bad != "" {
		p.kill()
		return Frame{}, &Error{Kind: ErrProtocol, Helper: c.name, Path: p.path, Op: op, Elapsed: elapsed, Detail: bad}
	}
	return f, nil
}

// process is one running helper.
type process struct {
	path    string
	cmd     *exec.Cmd
	stdin   *os.File
	frames  *os.File
	stderr  *tail
	dir     string
	done    chan struct{}
	waitErr error
	id      uint32
}

func spawn(path string) (*process, error) {
	dir, err := os.MkdirTemp("", "nativehelper-")
	if err != nil {
		return nil, err
	}
	inR, inW, err := os.Pipe()
	if err != nil {
		os.RemoveAll(dir)
		return nil, err
	}
	outR, outW, err := os.Pipe()
	if err != nil {
		inR.Close()
		inW.Close()
		os.RemoveAll(dir)
		return nil, err
	}
	p := &process{path: path, stdin: inW, frames: outR, stderr: &tail{}, dir: dir, done: make(chan struct{})}
	p.cmd = exec.Command(path)
	p.cmd.Stdin = inR
	p.cmd.Stdout = outW
	p.cmd.Stderr = p.stderr
	p.cmd.Env = append(os.Environ(), EnvCapture+"="+filepath.Join(dir, "capture"))
	p.cmd.WaitDelay = closeGrace
	err = p.cmd.Start()
	inR.Close()
	outW.Close()
	if err != nil {
		inW.Close()
		outR.Close()
		os.RemoveAll(dir)
		return nil, err
	}
	go func() {
		p.waitErr = p.cmd.Wait()
		close(p.done)
	}()
	return p, nil
}

// reap waits for an exiting helper (killing it after the grace period) and returns its exit
// status and its stderr: the capture file of the request it died in, then what it wrote to the
// stderr pipe, each bounded by BoundStderr. The pipe comes last and is bounded on its own, so a
// request that flooded its capture cannot push out the helper's last words.
func (p *process) reap() (string, []byte) {
	select {
	case <-p.done:
	case <-time.After(closeGrace):
		_ = p.cmd.Process.Kill()
		<-p.done
	}
	var b bytes.Buffer
	if capture, dropped, err := readTail(filepath.Join(p.dir, "capture")); err == nil {
		b.Write(BoundStderr(capture, dropped))
	}
	b.Write(p.stderr.Bytes())
	p.release()
	return p.cmd.ProcessState.String(), b.Bytes()
}

// readTail reads the last StderrKeep bytes of the file at path, and how many bytes precede them.
func readTail(path string) ([]byte, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return nil, 0, err
	}
	skip := max(fi.Size()-StderrKeep, 0)
	if _, err := f.Seek(skip, io.SeekStart); err != nil {
		return nil, 0, err
	}
	b, err := io.ReadAll(io.LimitReader(f, StderrKeep))
	return b, skip, err
}

func (p *process) kill() {
	_ = p.cmd.Process.Kill()
	<-p.done
	p.release()
}

func (p *process) release() {
	p.stdin.Close()
	p.frames.Close()
	os.RemoveAll(p.dir)
}

func (p *process) close(name string) error {
	p.stdin.Close()
	select {
	case <-p.done:
		p.release()
		if p.waitErr != nil {
			return fmt.Errorf("nativehelper: %s exited on stdin EOF with %v", name, p.waitErr)
		}
		return nil
	case <-time.After(closeGrace):
		p.kill()
		return fmt.Errorf("nativehelper: %s did not exit within %s of stdin EOF; killed", name, closeGrace)
	}
}

// tail keeps the last StderrKeep bytes written to it, and counts the bytes it dropped.
type tail struct {
	mu      sync.Mutex
	buf     []byte
	dropped int64
}

func (t *tail) Write(b []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buf = append(t.buf, b...)
	if len(t.buf) > StderrKeep {
		t.dropped += int64(len(t.buf) - StderrKeep)
		t.buf = append([]byte(nil), t.buf[len(t.buf)-StderrKeep:]...)
	}
	return len(b), nil
}

// Bytes is what the tail kept, bounded by BoundStderr.
func (t *tail) Bytes() []byte {
	t.mu.Lock()
	defer t.mu.Unlock()
	return BoundStderr(append([]byte(nil), t.buf...), t.dropped)
}
