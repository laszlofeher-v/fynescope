package tastybutton

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type Style int

const (
	Green Style = iota
	Red
	Orange
)

type TastyButton struct {
	widget.BaseWidget
	Text     string
	Style    Style
	OnTapped func()
	hovered  bool
	pressed  bool
}

func NewTastyButton(text string, style Style, tapped func()) *TastyButton {
	b := &TastyButton{
		Text:     text,
		Style:    style,
		OnTapped: tapped,
	}
	b.ExtendBaseWidget(b)
	return b
}

type tastyButtonRenderer struct {
	button     *TastyButton
	background *canvas.Rectangle
	shadow     *canvas.Rectangle
	label      *canvas.Text
	objects    []fyne.CanvasObject
}

func (r *tastyButtonRenderer) Destroy() {}

func (r *tastyButtonRenderer) Layout(size fyne.Size) {
	shadowHeight := float32(4)
	if r.button.pressed {
		r.shadow.Hide()
		r.background.Resize(fyne.NewSize(size.Width, size.Height-shadowHeight))
		r.background.Move(fyne.NewPos(0, shadowHeight))

		r.label.Resize(fyne.NewSize(size.Width, r.label.MinSize().Height))
		r.label.Move(fyne.NewPos(0, shadowHeight+(size.Height-shadowHeight)/2-r.label.MinSize().Height/2))
	} else {
		r.shadow.Show()
		r.shadow.Resize(fyne.NewSize(size.Width, size.Height))
		r.shadow.Move(fyne.NewPos(0, 0))

		r.background.Resize(fyne.NewSize(size.Width, size.Height-shadowHeight))
		r.background.Move(fyne.NewPos(0, 0))

		r.label.Resize(fyne.NewSize(size.Width, r.label.MinSize().Height))
		r.label.Move(fyne.NewPos(0, (size.Height-shadowHeight)/2-r.label.MinSize().Height/2))
	}
}

func (r *tastyButtonRenderer) MinSize() fyne.Size {
	txtMin := r.label.MinSize()
	return fyne.NewSize(txtMin.Width+30, txtMin.Height+16)
}

func (r *tastyButtonRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *tastyButtonRenderer) Refresh() {
	r.label.Text = r.button.Text

	var bgCol, shadowCol color.Color
	if r.button.Style == Green {
		if r.button.pressed {
			bgCol = color.NRGBA{R: 20, G: 140, B: 80, A: 255}
		} else if r.button.hovered {
			bgCol = color.NRGBA{R: 40, G: 200, B: 120, A: 255}
		} else {
			bgCol = color.NRGBA{R: 30, G: 180, B: 100, A: 255}
		}
		shadowCol = color.NRGBA{R: 15, G: 100, B: 55, A: 255}
	} else if r.button.Style == Orange {
		if r.button.pressed {
			bgCol = color.NRGBA{R: 200, G: 100, B: 0, A: 255}
		} else if r.button.hovered {
			bgCol = color.NRGBA{R: 255, G: 150, B: 30, A: 255}
		} else {
			bgCol = color.NRGBA{R: 230, G: 120, B: 10, A: 255}
		}
		shadowCol = color.NRGBA{R: 150, G: 70, B: 5, A: 255}
	} else {
		if r.button.pressed {
			bgCol = color.NRGBA{R: 160, G: 30, B: 30, A: 255}
		} else if r.button.hovered {
			bgCol = color.NRGBA{R: 220, G: 50, B: 50, A: 255}
		} else {
			bgCol = color.NRGBA{R: 190, G: 40, B: 40, A: 255}
		}
		shadowCol = color.NRGBA{R: 110, G: 20, B: 20, A: 255}
	}

	r.background.FillColor = bgCol
	r.shadow.FillColor = shadowCol
	r.label.Color = color.White

	r.background.Refresh()
	fyne.Do(r.shadow.Refresh)
	fyne.Do(r.label.Refresh)
}

func (b *TastyButton) CreateRenderer() fyne.WidgetRenderer {
	label := canvas.NewText(b.Text, color.White)
	label.Alignment = fyne.TextAlignCenter
	label.TextStyle = fyne.TextStyle{Bold: true}

	background := canvas.NewRectangle(color.Black)
	background.CornerRadius = 4

	shadow := canvas.NewRectangle(color.Black)
	shadow.CornerRadius = 4

	r := &tastyButtonRenderer{
		button:     b,
		background: background,
		shadow:     shadow,
		label:      label,
		objects:    []fyne.CanvasObject{shadow, background, label},
	}
	r.Refresh()
	return r
}

func (b *TastyButton) Tapped(ev *fyne.PointEvent) {
	if b.OnTapped != nil {
		b.OnTapped()
	}
}

func (b *TastyButton) MouseIn(ev *desktop.MouseEvent) {
	b.hovered = true
	b.Refresh()
}

func (b *TastyButton) MouseOut() {
	b.hovered = false
	b.pressed = false
	b.Refresh()
}

func (b *TastyButton) MouseMoved(ev *desktop.MouseEvent) {}

func (b *TastyButton) MouseDown(ev *desktop.MouseEvent) {
	if ev.Button == desktop.MouseButtonPrimary {
		b.pressed = true
		b.Refresh()
	}
}

func (b *TastyButton) MouseUp(ev *desktop.MouseEvent) {
	if ev.Button == desktop.MouseButtonPrimary {
		b.pressed = false
		b.Refresh()
	}
}
