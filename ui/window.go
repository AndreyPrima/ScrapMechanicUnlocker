// Package ui provides the minimal one-action unlock window.
package ui

import (
	"errors"
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	nativedlg "github.com/sqweek/dialog"

	"smunlocker/internal/service"
	"smunlocker/internal/steam"
	"smunlocker/internal/unlock"
)

// NewWindow builds the minimal unlock window: file row, one Unlock button,
// and a single inline status line. No popups except the file picker.
func NewWindow(a fyne.App) fyne.Window {
	w := a.NewWindow("SM Unlocker")
	w.Resize(fyne.NewSize(400, 220))
	w.SetFixedSize(false)

	status := widget.NewLabel("Select your unlock file.")
	status.Wrapping = fyne.TextWrapWord

	entry := widget.NewEntry()
	entry.SetPlaceHolder("No file selected")
	entry.Disable()

	unlockBtn := widget.NewButton(fmt.Sprintf("Unlock all (%d)", len(unlock.SortedOutfitIDs)), nil)
	unlockBtn.Disable()

	var pending string
	var lastDir string

	refresh := func(path string) {
		r, err := service.Inspect(path)
		if err != nil {
			pending = ""
			unlockBtn.Disable()
			status.SetText("Error: " + err.Error())
			return
		}
		pending = path
		entry.SetText(path)
		if r.Added == 0 {
			unlockBtn.Disable()
			status.SetText(fmt.Sprintf("Already unlocked (%d/%d).", r.Before, r.Total))
			return
		}
		unlockBtn.Enable()
		status.SetText(fmt.Sprintf("%d/%d unlocked — %d to add.", r.Before, r.Total, r.Added))
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
		refresh(path)
	})

	unlockBtn.OnTapped = func() {
		if pending == "" {
			return
		}
		unlockBtn.Disable()
		status.SetText("Working…")
		r, err := service.Unlock(pending)
		if err != nil {
			status.SetText("Error: " + err.Error())
			return
		}
		status.SetText(fmt.Sprintf("Done: %d → %d (+%d). Backup: unlock.bak", r.Before, r.After, r.Added))
	}

	w.SetContent(container.NewVBox(
		container.NewBorder(nil, nil, nil, browse, entry),
		unlockBtn,
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
