package gui

import (
	"fynescope/settings"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type (
	ScpDarkTheme struct {
		id int
	}
	ScpLightTheme struct {
		id int
	}
)

var (
	_          fyne.Theme = (*ScpDarkTheme)(nil)
	_          fyne.Theme = (*ScpLightTheme)(nil)
	lightTheme ScpLightTheme
	darkTheme  ScpDarkTheme
	themes     [2]fyne.Theme
)

const (
	ColorNameCha              fyne.ThemeColorName = "chaColor"
	ColorNameGeneratorDisp    fyne.ThemeColorName = "generatorDispColor"
	ColorNameSignalBackground fyne.ThemeColorName = "signalBackgroundColor"
	ColorNameDivision         fyne.ThemeColorName = "divisionColor"
)

func init() {
	themes[settings.LightTheme] = lightTheme
	themes[settings.DarkTheme] = darkTheme

}
func Theme(t settings.ThemeType) fyne.Theme {
	return themes[t]
}
func (t ScpDarkTheme) Color(c fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch c {
	case ColorNameGeneratorDisp:
		return color.RGBA{0xff, 0, 0, 0xff}
	case ColorNameCha:
		return color.RGBA{100, 180, 255, 255}
	case theme.ColorNameMenuBackground:
		return color.RGBA{0, 0, 0, 255}
	case ColorNameSignalBackground:
		return color.Black
	case ColorNameDivision:
		return color.RGBA{50, 150, 50, 3}
	case theme.ColorNameBackground:
		return color.RGBA{0, 0, 0, 0}
	case theme.ColorNameForeground:
		return color.White
	case theme.ColorNameButton:
		return color.Alpha16{A: 0x0}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 0x26, G: 0x26, B: 0x26, A: 0xff}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x42}
	case theme.ColorNameError:
		return color.NRGBA{R: 0xf4, G: 0x43, B: 0x36, A: 0xff}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0x21, G: 0x96, B: 0xf3, A: 0x7f}
	case theme.ColorNameHover:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xf}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x19}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0xb2, G: 0xb2, B: 0xb2, A: 0xff}
	case theme.ColorNamePressed:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x66}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0x21, G: 0x96, B: 0xf3, A: 0xff}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x99}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x66}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x19}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0x8f, G: 0x8f, B: 0x8f, A: 0xff}
	case theme.ColorNameOverlayBackground:
		return color.Black
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0x0, G: 0xff, B: 0x0, A: 0xff}
	case theme.ColorNameSeparator:
		return color.Black
	case theme.ColorNameForegroundOnPrimary:
		return color.White
	default:
		return theme.DefaultTheme().Color(c, variant)
	}
}
func (t ScpLightTheme) Color(c fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch c {
	case ColorNameGeneratorDisp:
		return color.RGBA{0xff, 0, 0, 0xff}
	case ColorNameCha:
		return color.RGBA{100, 180, 255, 0x42}
	case theme.ColorNameBackground:
		return color.White
	case ColorNameSignalBackground:
		return color.White
	case ColorNameDivision:
		return color.RGBA{0, 55, 0, 255}
	case theme.ColorNameMenuBackground:
		return color.RGBA{0xff, 0xff, 0xff, 0xff}
	case theme.ColorNameForeground:
		return color.Black
	case theme.ColorNameButton:
		return color.Alpha16{A: 0xf9}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 0x26, G: 0x26, B: 0x26, A: 0xf}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x42}
	case theme.ColorNameError:
		return color.NRGBA{R: 0xf4, G: 0x43, B: 0x36, A: 0xf}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0x21, G: 0x96, B: 0xf3, A: 0x7f}
	case theme.ColorNameHover:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xf}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0xb2, G: 0xb2, B: 0xb2, A: 0xf}
	case theme.ColorNamePressed:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x66}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0x21, G: 0x96, B: 0xf3, A: 0xff}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x99}
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0x0, G: 0x0, B: 0x0, A: 0x66}
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x19}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 0xf0, G: 0xf0, B: 0xf0, A: 0xf0}
	case theme.ColorNameSeparator:
		return color.Black
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0x0, G: 0xff, B: 0x0, A: 0xff}
	case theme.ColorNameForegroundOnPrimary:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xf0}
	default:
		return theme.DefaultTheme().Color(c, variant)
	}
}

func (t ScpLightTheme) Font(s fyne.TextStyle) fyne.Resource {
	if s.Monospace {
		return theme.DefaultTheme().Font(s)
	}
	if s.Bold {
		if s.Italic {
			return theme.DefaultTheme().Font(s)
		}
		return theme.DefaultTheme().Font(s)
	}
	if s.Italic {
		return theme.DefaultTheme().Font(s)
	}
	return theme.DefaultTheme().Font(s)
}
func (t ScpDarkTheme) Font(s fyne.TextStyle) fyne.Resource {
	if s.Monospace {
		return theme.DefaultTheme().Font(s)
	}
	if s.Bold {
		if s.Italic {
			return theme.DefaultTheme().Font(s)
		}
		return theme.DefaultTheme().Font(s)
	}
	if s.Italic {
		return theme.DefaultTheme().Font(s)
	}
	return theme.DefaultTheme().Font(s)
}

func (t ScpDarkTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (t ScpLightTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (t ScpDarkTheme) Size(s fyne.ThemeSizeName) float32 {
	switch s {
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNamePadding:
		return 4
	case theme.SizeNameScrollBar:
		return 16
	case theme.SizeNameScrollBarSmall:
		return 3
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameText:
		return 14
	case theme.SizeNameInputBorder:
		return 2
	default:
		return theme.DefaultTheme().Size(s)
	}
}
func (t ScpLightTheme) Size(s fyne.ThemeSizeName) float32 {
	switch s {
	case theme.SizeNameCaptionText:
		return 11
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNamePadding:
		return 4
	case theme.SizeNameScrollBar:
		return 16
	case theme.SizeNameScrollBarSmall:
		return 3
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameText:
		return 14
	case theme.SizeNameInputBorder:
		return 2
	default:
		return theme.DefaultTheme().Size(s)
	}
}
