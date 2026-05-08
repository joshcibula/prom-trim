package rules

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// BackupOptions configures backup behaviour.
type BackupOptions struct {
	// Suffix is appended to the original filename, e.g. ".bak" or a timestamp.
	// If empty a timestamp suffix is used.
	Suffix string
}

// BackupFile creates a copy of src next to the original file before it is
// overwritten. It returns the path of the backup file.
func BackupFile(src string, opts BackupOptions) (string, error) {
	suffix := opts.Suffix
	if suffix == "" {
		suffix = "." + time.Now().UTC().Format("20060102T150405Z") + ".bak"
	}

	dir := filepath.Dir(src)
	base := filepath.Base(src)
	dst := filepath.Join(dir, base+suffix)

	if err := copyFile(src, dst); err != nil {
		return "", fmt.Errorf("backup %s -> %s: %w", src, dst, err)
	}
	return dst, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
