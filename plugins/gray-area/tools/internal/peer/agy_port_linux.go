//go:build linux

package peer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// DaemonAddress finds the HTTP (Connect) address of the language server process pid serves. Its
// candidates are the loopback TCP sockets pid itself holds in LISTEN state: the socket inodes behind
// /proc/<pid>/fd, matched to the rows of /proc/<pid>/net/tcp and tcp6. Each is sent one POST
// carrying no token; the HTTP port answers it 401 with the Connect error "unauthenticated: missing
// CSRF token", and the server's HTTPS port refuses plain HTTP (agy 1.5.18). Exactly one candidate
// must answer so; none or several is an error naming each candidate's answer, never a guess.
func DaemonAddress(client *http.Client, pid int) (string, error) {
	addrs, err := listenAddrs(pid)
	if err != nil {
		return "", err
	}
	if len(addrs) == 0 {
		return "", fmt.Errorf("pid %d holds no loopback TCP socket in LISTEN state", pid)
	}
	// The probes run at once, so the lookup takes one request's time, bounded by client, however
	// many ports the process holds.
	type answer struct {
		status int
		body   []byte
		err    error
	}
	answers := make([]answer, len(addrs))
	var wg sync.WaitGroup
	for i, addr := range addrs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a := &answers[i]
			a.status, a.body, a.err = probe(client, addr)
		}()
	}
	wg.Wait()
	var found []string
	var seen []string
	for i, addr := range addrs {
		status, body, err := answers[i].status, answers[i].body, answers[i].err
		switch {
		case err != nil:
			seen = append(seen, fmt.Sprintf("%s: %v", addr, err))
		case missingToken(status, body):
			found = append(found, addr)
			seen = append(seen, fmt.Sprintf("%s: HTTP %d, the Connect port", addr, status))
		default:
			seen = append(seen, fmt.Sprintf("%s: HTTP %d", addr, status))
		}
	}
	if len(found) != 1 {
		return "", fmt.Errorf("%d of pid %d's listening ports answer as its HTTP port (%s)", len(found), pid, strings.Join(seen, "; "))
	}
	return found[0], nil
}

// DaemonListens reports whether process pid still holds a loopback socket in LISTEN state at addr:
// the check that a token goes to the process it was provisioned for, made just before it is sent.
func DaemonListens(pid int, addr string) (bool, error) {
	addrs, err := listenAddrs(pid)
	if err != nil {
		return false, err
	}
	return slices.Contains(addrs, addr), nil
}

// probe posts an empty Heartbeat with no token to addr.
func probe(client *http.Client, addr string) (int, []byte, error) {
	req, err := http.NewRequest(http.MethodPost,
		"http://"+addr+"/exa.language_server_pb.LanguageServerService/Heartbeat", strings.NewReader("{}"))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return resp.StatusCode, bytes.TrimSpace(b), nil
}

// missingToken reports whether status and body are the Connect CSRF interceptor's answer to a
// request that carried no token.
func missingToken(status int, body []byte) bool {
	if status != http.StatusUnauthorized {
		return false
	}
	var ce struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	return json.Unmarshal(body, &ce) == nil && ce.Code == "unauthenticated" &&
		strings.HasPrefix(ce.Message, "unauthenticated: missing CSRF token")
}

// listenAddrs is the sorted loopback addresses pid holds sockets in LISTEN state on.
func listenAddrs(pid int) ([]string, error) {
	fdDir := fmt.Sprintf("/proc/%d/fd", pid)
	fds, err := os.ReadDir(fdDir)
	if err != nil {
		return nil, err
	}
	inodes := map[string]bool{}
	for _, f := range fds {
		target, err := os.Readlink(filepath.Join(fdDir, f.Name()))
		if err != nil {
			continue // closed meanwhile
		}
		if ino, ok := strings.CutPrefix(target, "socket:["); ok {
			inodes[strings.TrimSuffix(ino, "]")] = true
		}
	}
	set := map[string]bool{}
	for _, name := range []string{"tcp", "tcp6"} {
		raw, err := os.ReadFile(fmt.Sprintf("/proc/%d/net/%s", pid, name))
		if err != nil {
			if name == "tcp6" && os.IsNotExist(err) {
				continue // no IPv6 in this kernel
			}
			return nil, err
		}
		for _, p := range listenRows(raw, inodes) {
			set[p] = true
		}
	}
	addrs := make([]string, 0, len(set))
	for a := range set {
		addrs = append(addrs, a)
	}
	sort.Strings(addrs)
	return addrs, nil
}

// listenRows reads /proc/net/tcp{,6} text for the addresses of loopback sockets in LISTEN state (st
// 0A) whose inode is in inodes. Columns: sl local_address rem_address st tx:rx tr:when retrnsmt uid
// timeout inode.
func listenRows(raw []byte, inodes map[string]bool) []string {
	var addrs []string
	sc := bufio.NewScanner(bytes.NewReader(raw))
	sc.Scan() // the header
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) < 10 || f[3] != "0A" || !inodes[f[9]] {
			continue
		}
		host, port, ok := strings.Cut(f[1], ":")
		ip := loopbackIP(host)
		if !ok || ip == "" {
			continue
		}
		if p, err := strconv.ParseUint(port, 16, 16); err == nil {
			addrs = append(addrs, net.JoinHostPort(ip, strconv.FormatUint(p, 10)))
		}
	}
	return addrs
}

// loopbackIP is the address to dial for a /proc/net address (little-endian hex words) that is
// 127.0.0.1, ::1 or ::ffff:127.0.0.1, and "" for any other.
func loopbackIP(h string) string {
	switch strings.ToUpper(h) {
	case "0100007F", "0000000000000000FFFF00000100007F":
		return "127.0.0.1"
	case "00000000000000000000000001000000":
		return "::1"
	}
	return ""
}
