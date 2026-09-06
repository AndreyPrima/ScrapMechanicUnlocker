package unlock

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	ErrTooSmall   = errors.New("file too small")
	ErrBadMagic   = errors.New("not an unlock file")
	ErrTruncated  = errors.New("unlock file is truncated")
	steamIDRegexp = regexp.MustCompile(`User_(\d+)`)
)

// UUIDToBytes encodes a hyphenated UUID string as 16 raw bytes.
func UUIDToBytes(uuidStr string) ([]byte, error) {
	clean := strings.ReplaceAll(uuidStr, "-", "")
	return hex.DecodeString(clean)
}

// sortedUnique returns a deduped + sorted copy.
func sortedUnique(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

func encodeIDs(sorted []string) (uuidBytes []byte, err error) {
	uuidBytes = make([]byte, 0, len(sorted)*16)
	for _, u := range sorted {
		b, err := UUIDToBytes(u)
		if err != nil {
			return nil, fmt.Errorf("bad uuid %q: %w", u, err)
		}
		if len(b) != 16 {
			return nil, fmt.Errorf("bad uuid %q: want 16 bytes, got %d", u, len(b))
		}
		uuidBytes = append(uuidBytes, b...)
	}
	return uuidBytes, nil
}

// GenerateUnlockFile builds unlock-file bytes for the given SteamID,
// unlocking every outfit in SortedOutfitIDs.
// Format: BE magic=1 | BE crc32(LE steamID + sorted uuid bytes) | BE count | uuid bytes.
func GenerateUnlockFile(steamID64 string) ([]byte, error) {
	return GenerateUnlockFileWithIDs(steamID64, SortedOutfitIDs)
}

// GenerateUnlockFileWithIDs builds unlock-file bytes for an explicit ID set
// (deduped + sorted). Kept for tests and byte-parity checks.
func GenerateUnlockFileWithIDs(steamID64 string, outfitUUIDs []string) ([]byte, error) {
	id, err := strconv.ParseUint(strings.TrimSpace(steamID64), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("bad steam id %q: %w", steamID64, err)
	}
	steamBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(steamBytes, id)

	sorted := sortedUnique(outfitUUIDs)
	uuidBytes, err := encodeIDs(sorted)
	if err != nil {
		return nil, err
	}

	crcData := make([]byte, 0, len(steamBytes)+len(uuidBytes))
	crcData = append(crcData, steamBytes...)
	crcData = append(crcData, uuidBytes...)
	crc := crc32.ChecksumIEEE(crcData)

	out := make([]byte, 12+len(uuidBytes))
	binary.BigEndian.PutUint32(out[0:4], 1)
	binary.BigEndian.PutUint32(out[4:8], crc)
	binary.BigEndian.PutUint32(out[8:12], uint32(len(sorted)))
	copy(out[12:], uuidBytes)
	return out, nil
}

// formatUUID renders 16 raw bytes as lowercase 8-4-4-4-12 hex.
func formatUUID(b []byte) string {
	h := hex.EncodeToString(b)
	return h[:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:]
}

// ReadUnlockFile returns UUID strings stored in an unlock file.
// Like the original program it validates magic but not CRC.
func ReadUnlockFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) < 12 {
		return nil, ErrTooSmall
	}
	if binary.BigEndian.Uint32(data[:4]) != 1 {
		return nil, ErrBadMagic
	}
	count := int(binary.BigEndian.Uint32(data[8:12]))
	if len(data) < 12+count*16 {
		return nil, ErrTruncated
	}
	out := make([]string, 0, count)
	for i := 0; i < count; i++ {
		off := 12 + i*16
		out = append(out, formatUUID(data[off:off+16]))
	}
	return out, nil
}

// ExtractSteamIDFromPath returns <steamid> from .../User_<steamid>/... or "".
func ExtractSteamIDFromPath(path string) string {
	m := steamIDRegexp.FindStringSubmatch(path)
	if m == nil {
		return ""
	}
	return m[1]
}
