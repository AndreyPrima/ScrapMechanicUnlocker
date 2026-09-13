// Package ui provides the minimal one-action unlock window.
package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	nativedlg "github.com/sqweek/dialog"

	"smunlocker/internal/service"
	"smunlocker/internal/steam"
	"smunlocker/internal/unlock"
)

// NewWindow builds the minimal unlock window: file row, Unlock + Restore row,
// and a single inline status line. No popups except the file picker.
func NewWindow(a fyne.App) fyne.Window {
	w := a.NewWindow("SM Unlocker")
	w.Resize(fyne.NewSize(400, 220))
	w.SetFixedSize(false)

	status := widget.NewLabel("Select your unlock file.")
	status.Wrapping = fyne.TextWrapWord

	entry := widget.NewEntry()
	entry.SetPlaceHolder("No file selected")
	entry.Enable()

	unlockBtn := widget.NewButton(fmt.Sprintf("Unlock all (%d)", len(unlock.SortedOutfitIDs)), nil)
	unlockBtn.Disable()

	restoreBtn := widget.NewButton("Restore .bak", nil)
	restoreBtn.Disable()

	prefs := a.Preferences()
	lastDir := prefs.String("lastDir")

	var pending string

	updateRestore := func() {
		if pending == "" {
			restoreBtn.Disable()
			return
		}
		if _, err := os.Stat(pending + ".bak"); err == nil {
			restoreBtn.Enable()
		} else {
			restoreBtn.Disable()
		}
	}

	refresh := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			pending = ""
			unlockBtn.Disable()
			restoreBtn.Disable()
			return
		}
		r, err := service.Inspect(path)
		if err != nil {
			pending = ""
			unlockBtn.Disable()
			restoreBtn.Disable()
			status.SetText("Error: " + err.Error())
			return
		}
		pending = path
		entry.SetText(path)
		updateRestore()
		if r.Added == 0 {
			unlockBtn.Disable()
			status.SetText(fmt.Sprintf("User %s: Already unlocked (%d/%d).", r.SteamID, r.Before, r.Total))
			return
		}
		unlockBtn.Enable()
		status.SetText(fmt.Sprintf("User %s: %d/%d — %d to add.", r.SteamID, r.Before, r.Total, r.Added))
	}

	entry.OnSubmitted = func(text string) {
		refresh(text)
	}

	browse := widget.NewButton("Browse…", func() {
		start := lastDir
		if start == "" {
			start = steam.DefaultDir()
		}
		// Native system file dialog (GTK FileChooser on Linux,
		// GetOpenFileName on Windows). No extension filter: the game
		// file is literally named "unlock", a filter would hide it.
		path, err := nativedlg.File().
			Title("Select your unlock file").
			SetStartDir(start).
			Load()
		if err != nil {
			if errors.Is(err, nativedlg.ErrCancelled) {
				return // user cancelled, keep current state
			}
			status.SetText("Error: " + err.Error())
			return
		}
		lastDir = dirOf(path)
		prefs.SetString("lastDir", lastDir)
		refresh(path)
	})

	unlockBtn.OnTapped = func() {
		if pending == "" {
			return
		}
		unlockBtn.Disable()
		status.SetText("Working…")
		if _, err := service.Unlock(pending); err != nil {
			status.SetText("Error: " + err.Error())
			if r, err2 := service.Inspect(pending); err2 == nil && r.Added > 0 {
				unlockBtn.Enable()
			}
			updateRestore()
			return
		}
		refresh(pending)
	}

	restoreBtn.OnTapped = func() {
		if pending == "" {
			return
		}
		bak := pending + ".bak"
		data, err := os.ReadFile(bak)
		if err != nil {
			status.SetText("Error: " + err.Error())
			updateRestore()
			return
		}
		perm := os.FileMode(0o644)
		if st, err := os.Stat(pending); err == nil {
			perm = st.Mode().Perm()
		}
		if err := os.WriteFile(pending, data, perm); err != nil {
			status.SetText("Error: " + err.Error())
			updateRestore()
			return
		}
		refresh(pending)
		status.SetText("Restored from unlock.bak. " + status.Text)
	}

	// Auto-preselect: reuse steam.FindUnlockFiles, show most recent if many.
	if hits := steam.FindUnlockFiles(); len(hits) == 1 {
		refresh(hits[0])
	} else if len(hits) > 1 {
		best := hits[0]
		bestStat, _ := os.Stat(best)
		for _, h := range hits[1:] {
			st, err := os.Stat(h)
			if err != nil {
				continue
			}
			if bestStat == nil || st.ModTime().After(bestStat.ModTime()) {
				best = h
				bestStat = st
			}
		}
		refresh(best)
		status.SetText(status.Text + fmt.Sprintf(" Found %d unlock files — showing most recent, Browse to change.", len(hits)))
	}

	w.SetContent(container.NewVBox(
		container.NewBorder(nil, nil, nil, browse, entry),
		container.NewGridWithColumns(2, unlockBtn, restoreBtn),
		status,
	))
	return w
}

func dirOf(path string) string {
	if d := filepath.Dir(path); d != "" {
		return d
	}
	return steam.DefaultDir()
}
