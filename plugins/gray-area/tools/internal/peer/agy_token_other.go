//go:build !unix

package peer

import "errors"

var errNoDaemonToken = errors.New("the daemon token file is read and written on Unix only")

// ReadDaemonToken is unavailable here: the file's owner and mode cannot be checked.
func ReadDaemonToken(string) (string, error) { return "", errNoDaemonToken }

// WriteDaemonToken is unavailable here.
func WriteDaemonToken(string) error { return errNoDaemonToken }
