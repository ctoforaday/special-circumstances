package runlive

import "io"

// readSharingDelete is os.ReadFile through openSharingDelete: the read never stops a writer
// removing or replacing the file under it.
func readSharingDelete(path string) ([]byte, error) {
	f, err := openSharingDelete(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}
