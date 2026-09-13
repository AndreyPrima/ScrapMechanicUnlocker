package unlock

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestOutfitCount(t *testing.T) {
	if len(AllOutfits) != 252 {
		t.Fatalf("want 252 outfits, got %d", len(AllOutfits))
	}
}

func TestSortedIDsPrecomputed(t *testing.T) {
	if len(SortedOutfitIDs) != 252 {
		t.Fatalf("want 252 sorted ids, got %d", len(SortedOutfitIDs))
	}
	if !sort.StringsAreSorted(SortedOutfitIDs) {
		t.Fatal("SortedOutfitIDs is not sorted")
	}
	seen := map[string]bool{}
	for _, id := range SortedOutfitIDs {
		if seen[id] {
			t.Fatalf("duplicate id %q", id)
		}
		seen[id] = true
		if _, ok := AllOutfits[id]; !ok {
			t.Fatalf("id %q not in AllOutfits", id)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	data, err := GenerateUnlockFile("76561198000000000")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 12+16*252 {
		t.Fatalf("bad length %d", len(data))
	}
	if binary.BigEndian.Uint32(data[:4]) != 1 {
		t.Fatal("bad magic")
	}
	if binary.BigEndian.Uint32(data[8:12]) != 252 {
		t.Fatal("bad count")
	}
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "User_76561198000000000"), 0o755); err != nil {
		t.Fatal(err)
	}
	fp := filepath.Join(dir, "User_76561198000000000", "unlock")
	if err := os.WriteFile(fp, data, 0o644); err != nil {
		t.Fatal(err)
	}
	back, err := ReadUnlockFile(fp)
	if err != nil {
		t.Fatal(err)
	}
	if len(back) != 252 {
		t.Fatalf("want 252 back, got %d", len(back))
	}
	if got := ExtractSteamIDFromPath(fp); got != "76561198000000000" {
		t.Fatalf("bad steam id %q", got)
	}
}

func TestGoldenParity(t *testing.T) {
	data, err := GenerateUnlockFile("76561198000000000")
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile("testdata/unlock.golden")
	if err != nil {
		t.Skipf("no golden file: %v", err)
	}
	if !bytes.Equal(data, golden) {
		t.Fatal("output differs from testdata/unlock.golden")
	}
}

func TestReadErrors(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x")
	if err := os.WriteFile(p, []byte{0, 1, 2}, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadUnlockFile(p); err != ErrTooSmall {
		t.Fatalf("want ErrTooSmall, got %v", err)
	}
	if err := os.WriteFile(p, make([]byte, 12), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadUnlockFile(p); err != ErrBadMagic {
		t.Fatalf("want ErrBadMagic, got %v", err)
	}
}

func TestBadUUID(t *testing.T) {
	if _, err := GenerateUnlockFileWithIDs("76561198000000000", []string{"not-a-uuid"}); err == nil {
		t.Fatal("want error for bad uuid")
	}
	if _, err := GenerateUnlockFile("not-a-number"); err == nil {
		t.Fatal("want error for bad steam id")
	}
}

func TestExtractEmpty(t *testing.T) {
	if got := ExtractSteamIDFromPath("/tmp/nothing"); got != "" {
		t.Fatalf("want empty, got %q", got)
	}
}

func TestVerifyCRCRoundTrip(t *testing.T) {
	data, err := GenerateUnlockFile("76561198000000000")
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyCRC(data, "76561198000000000"); err != nil {
		t.Fatalf("valid file failed verify: %v", err)
	}
	if err := VerifyCRC(data, "76561198000000001"); !errors.Is(err, ErrCRCMismatch) {
		t.Fatalf("want ErrCRCMismatch for wrong steam id, got %v", err)
	}
	tampered := append([]byte(nil), data...)
	tampered[len(tampered)-1] ^= 0xff
	if err := VerifyCRC(tampered, "76561198000000000"); !errors.Is(err, ErrCRCMismatch) {
		t.Fatalf("want ErrCRCMismatch for tampered file, got %v", err)
	}
}

func TestVerifyFileCRC(t *testing.T) {
	data, err := GenerateUnlockFile("76561198000000000")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	fp := filepath.Join(dir, "unlock")
	if err := os.WriteFile(fp, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := VerifyFileCRC(fp, "76561198000000000"); err != nil {
		t.Fatalf("valid file failed verify: %v", err)
	}
	if err := VerifyFileCRC(fp, "76561198000000001"); !errors.Is(err, ErrCRCMismatch) {
		t.Fatalf("want ErrCRCMismatch, got %v", err)
	}
}

func TestHugeCountTruncated(t *testing.T) {
	data := make([]byte, 12)
	binary.BigEndian.PutUint32(data[0:4], 1)
	binary.BigEndian.PutUint32(data[8:12], 0xFFFFFFFF)
	dir := t.TempDir()
	fp := filepath.Join(dir, "unlock")
	if err := os.WriteFile(fp, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadUnlockFile(fp); err != ErrTruncated {
		t.Fatalf("want ErrTruncated, got %v", err)
	}
	if err := VerifyCRC(data, "76561198000000000"); err != ErrTruncated {
		t.Fatalf("want ErrTruncated, got %v", err)
	}
}
