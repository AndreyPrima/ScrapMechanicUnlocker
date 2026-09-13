package service

import (
	"fmt"
	"os"
	"path/filepath"

	"smunlocker/internal/unlock"
)

// Result describes one inspect or unlock operation.
type Result struct {
	Path    string
	SteamID string
	Before  int
	After   int
	Added   int
	Total   int
}

// Inspect reads path and reports current/missing counts without writing.
func Inspect(path string) (Result, error) {
	var r Result
	r.Path = path
	r.Total = len(unlock.SortedOutfitIDs)
	current, err := unlock.ReadUnlockFile(path)
	if err != nil {
		return r, err
	}
	r.Before = len(current)
	steamID := unlock.ExtractSteamIDFromPath(path)
	if steamID == "" {
		return r, fmt.Errorf("could not determine Steam ID from path (expected .../User_<id>/...): %s", path)
	}
	r.SteamID = steamID
	have := make(map[string]struct{}, len(current))
	for _, c := range current {
		have[c] = struct{}{}
	}
	missing := 0
	for _, id := range unlock.SortedOutfitIDs {
		if _, ok := have[id]; !ok {
			missing++
		}
	}
	r.Added = missing
	r.After = r.Total
	return r, nil
}

// Unlock rewrites path with every outfit, keeping a .bak of the original
// (created once, never overwritten) and writing atomically via temp+rename.
func Unlock(path string) (Result, error) {
	r, err := Inspect(path)
	if err != nil {
		return r, err
	}
	if r.Added == 0 {
		r.After = r.Before
		return r, nil
	}
	content, err := unlock.GenerateUnlockFile(r.SteamID)
	if err != nil {
		return r, err
	}
	perm := fileMode(path)
	if err := writeBackupOnce(path, perm); err != nil {
		return r, err
	}
	if err := writeAtomic(path, content, perm); err != nil {
		return r, err
	}
	return r, nil
}

// fileMode returns the permission bits of path, falling back to 0644.
func fileMode(path string) os.FileMode {
	if st, err := os.Stat(path); err == nil {
		return st.Mode().Perm()
	}
	return 0o644
}

// writeBackupOnce copies path to path+".bak" unless the backup exists.
func writeBackupOnce(path string, perm os.FileMode) error {
	bak := path + ".bak"
	if _, err := os.Stat(bak); err == nil {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return os.WriteFile(bak, data, perm)
}

// writeAtomic writes via temp file in the same dir + rename + fsync.
func writeAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".unlock-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	success := false
	defer func() {
		if !success {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		// Windows refuses to rename over an existing file: remove + retry.
		if rmErr := os.Remove(path); rmErr != nil {
			return err
		}
		if err := os.Rename(tmpName, path); err != nil {
			return err
		}
	}
	success = true
	return nil
}
