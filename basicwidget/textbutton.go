// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Guigui Authors

package basicwidget

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget/internal/draw"
)

type IconAlign int

const (
	IconAlignStart IconAlign = iota
	IconAlignEnd
)

type TextButton struct {
	guigui.DefaultWidget

	button    Button
	text      Text
	icon      Image
	IconAlign IconAlign

	textColor color.Color
}

func (t *TextButton) SetOnDown(f func()) {
	t.button.SetOnDown(f)
}

func (t *TextButton) SetOnUp(f func()) {
	t.button.SetOnUp(f)
}

func (b *TextButton) setOnRepeat(f func()) {
	b.button.setOnRepeat(f)
}

func (t *TextButton) SetText(text string) {
	t.text.SetValue(text)
}

func (t *TextButton) SetTextBold(bold bool) {
	t.text.SetBold(bold)
}

func (t *TextButton) SetIcon(icon *ebiten.Image) {
	t.icon.SetImage(icon)
}

func (t *TextButton) SetIconAlign(align IconAlign) {
	if t.IconAlign == align {
		return
	}
	t.IconAlign = align
	guigui.RequestRedraw(t)
}

func (t *TextButton) SetTextColor(clr color.Color) {
	if draw.EqualColor(t.textColor, clr) {
		return
	}
	t.textColor = clr
	guigui.RequestRedraw(t)
}

func (t *TextButton) setPairedButton(pair *TextButton) {
	t.button.setPairedButton(&pair.button)
}

func (t *TextButton) setKeepPressed(keep bool) {
	t.button.setKeepPressed(keep)
}

func (t *TextButton) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	appender.AppendChildWidgetWithBounds(&t.button, context.Bounds(t))

	s := context.Size(t)

	imgSize := t.iconSize(context)

	tw := t.text.TextSize(context).X
	if t.textColor != nil {
		t.text.SetColor(t.textColor)
	} else {
		t.text.SetColor(draw.TextColor(context.ColorMode(), context.IsEnabled(t)))
	}
	t.text.SetHorizontalAlign(HorizontalAlignCenter)
	t.text.SetVerticalAlign(VerticalAlignMiddle)

	ds := t.defaultSize(context, false)
	textP := context.Position(t)
	if t.icon.HasImage() {
		textP.X += (s.X - ds.X) / 2
		switch t.IconAlign {
		case IconAlignStart:
			textP.X += textButtonEdgeAndImagePadding(context)
			textP.X += imgSize + textButtonTextAndImagePadding(context)
		case IconAlignEnd:
			textP.X += textButtonEdgeAndTextPadding(context)
		}
	} else {
		textP.X += (s.X - tw) / 2
	}
	if t.button.isPressed(context) {
		textP.Y += int(1 * context.Scale())
	}
	appender.AppendChildWidgetWithBounds(&t.text, image.Rectangle{
		Min: textP,
		Max: textP.Add(image.Pt(tw, s.Y)),
	})

	imgP := context.Position(t)
	if t.text.Value() != "" {
		imgP.X += (s.X - ds.X) / 2
		switch t.IconAlign {
		case IconAlignStart:
			imgP.X += textButtonEdgeAndImagePadding(context)
		case IconAlignEnd:
			imgP.X += textButtonEdgeAndTextPadding(context)
			imgP.X += tw + textButtonTextAndImagePadding(context)
		}
	} else {
		imgP.X += (s.X - imgSize) / 2
	}
	imgP.Y += (s.Y - imgSize) / 2
	if t.button.isPressed(context) {
		imgP.Y += int(1 * context.Scale())
	}
	appender.AppendChildWidgetWithBounds(&t.icon, image.Rectangle{
		Min: imgP,
		Max: imgP.Add(image.Pt(imgSize, imgSize)),
	})

	return nil
}

func (t *TextButton) DefaultSize(context *guigui.Context) image.Point {
	return t.defaultSize(context, false)
}

func (t *TextButton) defaultSize(context *guigui.Context, forceBold bool) image.Point {
	dh := defaultButtonSize(context).Y
	var w int
	if forceBold {
		w = t.text.boldTextSize(context).X
	} else {
		w = t.text.TextSize(context).X
	}
	if t.icon.HasImage() {
		w += t.defaultIconSize(context)
		if t.text.Value() != "" {
			w += textButtonTextAndImagePadding(context)
		}
		w += textButtonEdgeAndTextPadding(context)
		w += textButtonEdgeAndImagePadding(context)
		return image.Pt(w, dh)
	}
	return image.Pt(w+UnitSize(context), dh)
}

func (t *TextButton) setSharpenCorners(sharpenCorners draw.SharpenCorners) {
	t.button.setSharpenCorners(sharpenCorners)
}

func textButtonTextAndImagePadding(context *guigui.Context) int {
	return UnitSize(context) / 4
}

func textButtonEdgeAndTextPadding(context *guigui.Context) int {
	return UnitSize(context) / 2
}

func textButtonEdgeAndImagePadding(context *guigui.Context) int {
	return UnitSize(context) / 4
}

func (t *TextButton) defaultIconSize(context *guigui.Context) int {
	return int(LineHeight(context))
}

func (t *TextButton) iconSize(context *guigui.Context) int {
	s := context.Size(t)
	return min(t.defaultIconSize(context), s.X, s.Y)
}

func (t *TextButton) setUseAccentColor(use bool) {
	t.button.setUseAccentColor(use)
}
