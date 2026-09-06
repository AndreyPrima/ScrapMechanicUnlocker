// Package theme provides a black-only Fyne theme.
package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	fynetheme "fyne.io/fyne/v2/theme"
)

var (
	black       = color.Gray{Y: 0x00}
	panel       = color.Gray{Y: 0x0A}
	border      = color.Gray{Y: 0x2E}
	white       = color.Gray{Y: 0xFF}
	transparent = color.NRGBA{}
)

// Black is a monochrome dark-only theme. Color ignores the requested
// variant so an OS light mode can never bleach the UI.
type Black struct{}

func (Black) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch n {
	case fynetheme.ColorNameBackground:
		return black
	case fynetheme.ColorNameMenuBackground,
		fynetheme.ColorNameOverlayBackground,
		fynetheme.ColorNameHeaderBackground,
		fynetheme.ColorNameButton,
		fynetheme.ColorNameInputBackground:
		return panel
	case fynetheme.ColorNameForeground,
		fynetheme.ColorNamePrimary,
		fynetheme.ColorNameFocus,
		fynetheme.ColorNameHover,
		fynetheme.ColorNamePressed,
		fynetheme.ColorNameSelection,
		fynetheme.ColorNameHyperlink:
		return white
	case fynetheme.ColorNameDisabled,
		fynetheme.ColorNameDisabledButton,
		fynetheme.ColorNamePlaceHolder,
		fynetheme.ColorNameSeparator,
		fynetheme.ColorNameScrollBar,
		fynetheme.ColorNameScrollBarBackground,
		fynetheme.ColorNameInputBorder,
		fynetheme.ColorNameInnerWindowBorder,
		fynetheme.ColorNameInnerWindowBorderInactive:
		return border
	case fynetheme.ColorNameShadow:
		return transparent
	case fynetheme.ColorNameError:
		return color.NRGBA{R: 0xFF, G: 0x55, B: 0x55, A: 0xFF}
	case fynetheme.ColorNameSuccess:
		return color.NRGBA{R: 0x55, G: 0xFF, B: 0x55, A: 0xFF}
	case fynetheme.ColorNameWarning:
		return color.NRGBA{R: 0xFF, G: 0xCC, B: 0x44, A: 0xFF}
	case fynetheme.ColorNameForegroundOnPrimary,
		fynetheme.ColorNameForegroundOnError,
		fynetheme.ColorNameForegroundOnSuccess,
		fynetheme.ColorNameForegroundOnWarning:
		return black
	default:
		return fynetheme.DefaultTheme().Color(n, fynetheme.VariantDark)
	}
}

func (Black) Font(s fyne.TextStyle) fyne.Resource {
	return fynetheme.DefaultTheme().Font(s)
}

func (Black) Icon(n fyne.ThemeIconName) fyne.Resource {
	return fynetheme.DefaultTheme().Icon(n)
}

func (Black) Size(n fyne.ThemeSizeName) float32 {
	return fynetheme.DefaultTheme().Size(n)
}
