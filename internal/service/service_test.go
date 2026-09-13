package service

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"smunlocker/internal/unlock"
)

func writeSeedFile(t *testing.T, dir, steamID string, ids []string) string {
	t.Helper()
	data, err := unlock.GenerateUnlockFileWithIDs(steamID, ids)
	if err != nil {
		t.Fatal(err)
	}
	userDir := filepath.Join(dir, "User_"+steamID)
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fp := filepath.Join(userDir, "unlock")
	if err := os.WriteFile(fp, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return fp
}

func TestInspectCounts(t *testing.T) {
	dir := t.TempDir()
	fp := writeSeedFile(t, dir, "76561198000000000", unlock.SortedOutfitIDs[:3])
	r, err := Inspect(fp)
	if err != nil {
		t.Fatal(err)
	}
	if r.Before != 3 || r.Total != 252 || r.Added != 249 {
		t.Fatalf("bad inspect: %+v", r)
	}
	if r.SteamID != "76561198000000000" {
		t.Fatalf("bad steam id %q", r.SteamID)
	}
}

func TestUnlockCreatesBakAndIsAtomic(t *testing.T) {
	dir := t.TempDir()
	fp := writeSeedFile(t, dir, "76561198000000000", unlock.SortedOutfitIDs[:3])
	orig, err := os.ReadFile(fp)
	if err != nil {
		t.Fatal(err)
	}
	r, err := Unlock(fp)
	if err != nil {
		t.Fatal(err)
	}
	if r.Before != 3 || r.After != 252 || r.Added != 249 {
		t.Fatalf("bad result: %+v", r)
	}
	bak, err := os.ReadFile(fp + ".bak")
	if err != nil {
		t.Fatalf("missing .bak: %v", err)
	}
	if string(bak) != string(orig) {
		t.Fatal(".bak does not match original")
	}
	// No temp files left behind.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if len(e.Name()) > 8 && e.Name()[:8] == ".unlock-" {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}
	// User dir itself must not contain temp files either.
	userEntries, err := os.ReadDir(filepath.Join(dir, "User_76561198000000000"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range userEntries {
		if len(e.Name()) > 8 && e.Name()[:8] == ".unlock-" {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}
}

func TestUnlockIdempotent(t *testing.T) {
	dir := t.TempDir()
	fp := writeSeedFile(t, dir, "76561198000000000", unlock.SortedOutfitIDs[:3])
	if _, err := Unlock(fp); err != nil {
		t.Fatal(err)
	}
	bak1, err := os.ReadFile(fp + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	r, err := Unlock(fp)
	if err != nil {
		t.Fatal(err)
	}
	if r.Added != 0 || r.Before != 252 {
		t.Fatalf("second unlock should be no-op: %+v", r)
	}
	bak2, err := os.ReadFile(fp + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	if string(bak1) != string(bak2) {
		t.Fatal(".bak was overwritten on second run")
	}
}

func TestInspectMissingSteamID(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "unlock")
	data, err := unlock.GenerateUnlockFile("76561198000000000")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fp, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Inspect(fp); err == nil {
		t.Fatal("want error for path without User_<id>")
	}
}

func TestUnlockNoOpPreservesMtime(t *testing.T) {
	dir := t.TempDir()
	fp := writeSeedFile(t, dir, "76561198000000000", unlock.SortedOutfitIDs)
	stamp := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(fp, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(fp)
	if err != nil {
		t.Fatal(err)
	}
	r, err := Unlock(fp)
	if err != nil {
		t.Fatal(err)
	}
	if r.Added != 0 || r.After != r.Before {
		t.Fatalf("no-op should report After==Before: %+v", r)
	}
	after, err := os.ReadFile(fp)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("no-op unlock rewrote the file")
	}
	st, err := os.Stat(fp)
	if err != nil {
		t.Fatal(err)
	}
	if !st.ModTime().Equal(stamp) {
		t.Fatalf("mtime changed: %v != %v", st.ModTime(), stamp)
	}
	if _, err := os.Stat(fp + ".bak"); !os.IsNotExist(err) {
		t.Fatal("no-op unlock should not create .bak")
	}
}

func TestUnlockPreservesPerms(t *testing.T) {
	dir := t.TempDir()
	fp := writeSeedFile(t, dir, "76561198000000000", unlock.SortedOutfitIDs[:3])
	if err := os.Chmod(fp, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Unlock(fp); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{fp, fp + ".bak"} {
		st, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		if st.Mode().Perm() != 0o600 {
			t.Fatalf("%s mode = %o, want 600", p, st.Mode().Perm())
		}
	}
}
