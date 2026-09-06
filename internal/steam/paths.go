package steam

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// Scrap Mechanic Steam AppID, used for Proton prefix lookup on Linux.
const AppID = "387990"

var (
	homeOnce      = sync.OnceValues(func() (string, error) { return os.UserHomeDir() })
	userDirsCache = sync.OnceValue(buildUserDirs)
)

func homeDir() string {
	home, _ := homeOnce()
	return home
}

// UserDirs returns candidate .../Scrap Mechanic/User dirs, existing first.
//
// Windows: %APPDATA%/Axolot Games/Scrap Mechanic/User
// Linux (Proton): <steam>/steamapps/compatdata/387990/pfx/drive_c/users/
// steamuser/AppData/Roaming/Axolot Games/Scrap Mechanic/User
func UserDirs() []string {
	return userDirsCache()
}

func buildUserDirs() []string {
	var candidates []string
	if runtime.GOOS == "windows" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			candidates = append(candidates,
				filepath.Join(appdata, "Axolot Games", "Scrap Mechanic", "User"))
		}
		return candidates
	}
	home := homeDir()
	roots := []string{
		filepath.Join(home, ".steam", "steam"),
		filepath.Join(home, ".local", "share", "Steam"),
		filepath.Join(home, ".steam", "root"),
		filepath.Join(home, "snap", "steam", "common", ".local", "share", "Steam"),
		filepath.Join(home, ".var", "app", "com.valvesoftware.Steam", "data", "Steam"),
	}
	if compat := os.Getenv("STEAM_COMPAT_DATA_PATH"); compat != "" {
		// $STEAM_COMPAT_DATA_PATH/compatdata/<id> -> steam root guess
		roots = append([]string{filepath.Join(compat, "..", "..")}, roots...)
	}
	seen := map[string]bool{}
	for _, root := range roots {
		if seen[root] {
			continue
		}
		seen[root] = true
		candidates = append(candidates, filepath.Join(root,
			"steamapps", "compatdata", AppID, "pfx", "drive_c",
			"users", "steamuser", "AppData", "Roaming",
			"Axolot Games", "Scrap Mechanic", "User"))
	}
	candidates = append(candidates,
		filepath.Join(home, ".local", "share", "Axolot Games", "Scrap Mechanic", "User"))
	var existing, rest []string
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			existing = append(existing, c)
		} else {
			rest = append(rest, c)
		}
	}
	return append(existing, rest...)
}

// DefaultDir is the best starting dir for the file dialog.
// It never creates bogus dirs on Linux; on Windows it may create %APPDATA% dir.
func DefaultDir() string {
	for _, d := range UserDirs() {
		if st, err := os.Stat(d); err == nil && st.IsDir() {
			return d
		}
	}
	if runtime.GOOS == "windows" {
		if dirs := UserDirs(); len(dirs) > 0 {
			_ = os.MkdirAll(dirs[0], 0o755)
			return dirs[0]
		}
	}
	if home := homeDir(); home != "" {
		return home
	}
	return "."
}
