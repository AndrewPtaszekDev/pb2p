package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// writeFileAtomic writes data to path so that a reader observes either the old
// contents or the complete new contents, never a partially written file.
//
// The data is written to a temp file in the same directory, flushed to stable
// storage, then renamed over path. rename(2) is atomic within a filesystem,
// and the temp file lives beside the target so it is always the same
// filesystem. If the file already exists its mode is preserved; otherwise mode
// is used (minus the umask).
//
// With exclusive set, the write fails with fs.ErrExist instead of overwriting
// an existing file.
func writeFileAtomic(path string, data []byte, mode fs.FileMode, exclusive bool) error {
	dir := filepath.Dir(path)

	if exclusive {
		// CreateTemp below always creates a fresh temp file, so an existing
		// target has to be detected here and reported rather than renamed over.
		if _, err := os.Stat(path); err == nil {
			return fs.ErrExist
		}
	}

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		if exclusive && errors.Is(err, fs.ErrExist) {
			return fs.ErrExist
		}
		return fmt.Errorf("creating temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()

	// Remove the temp file on any failure path; a successful rename makes this
	// a no-op.
	defer func() {
		if tmpName != "" {
			_ = os.Remove(tmpName)
		}
	}()

	if err := writeAndSync(tmp, data); err != nil {
		return fmt.Errorf("writing temp file %s: %w", tmpName, err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file %s: %w", tmpName, err)
	}

	// Preserve the mode of an existing file so atomic replacement doesn't
	// silently change permissions; CreateTemp starts at 0600.
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}

	if err := os.Chmod(tmpName, mode); err != nil {
		return fmt.Errorf("setting mode on temp file %s: %w", tmpName, err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("renaming temp file %s to %s: %w", tmpName, path, err)
	}

	// The temp file no longer exists, so skip the deferred cleanup.
	tmpName = ""

	// fsync the directory so the rename itself survives a crash.
	if dirFile, err := os.Open(dir); err == nil {
		_ = dirFile.Sync()
		_ = dirFile.Close()
	}

	return nil
}

func writeAndSync(f *os.File, data []byte) error {
	if _, err := f.Write(data); err != nil {
		return err
	}
	return f.Sync()
}
