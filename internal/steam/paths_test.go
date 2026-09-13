package steam

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestFindUnlockFilesIn(t *testing.T) {
	dir := t.TempDir()
	mk := func(user, name string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Join(dir, user), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, user, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("User_222", "unlock")
	mk("User_111", "unlock")
	mk("User_333", "other")  // no unlock file inside
	mk("NotAUser", "unlock") // wrong prefix, ignored

	got := findUnlockFilesIn([]string{filepath.Join(dir, "does-not-exist"), dir})
	want := []string{
		filepath.Join(dir, "User_111", "unlock"),
		filepath.Join(dir, "User_222", "unlock"),
	}
	if len(got) != len(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
	if !sort.StringsAreSorted(got) {
		t.Fatalf("not sorted: %q", got)
	}
}
