package qcli

import (
	"io"
	"os"
)

// CopyFileBits - copy file content from a to b
// differs from CopyFile in:
//   - does not do permissions - new files created with 0644
//   - if src is a symlink, copies content, not link.
//   - does not invoke sh.
func CopyFileBits(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func PathExists(d string) bool {
	_, err := os.Stat(d)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
