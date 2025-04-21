// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget/internal/draw"
)

type TextButton struct {
	guigui.DefaultWidget

	button Button
	text   Text
	image  Image

	textColor color.Color
}

func (t *TextButton) SetOnDown(f func()) {
	t.button.SetOnDown(f)
}

func (t *TextButton) SetOnUp(f func()) {
	t.button.SetOnUp(f)
}

func (t *TextButton) SetText(text string) {
	t.text.SetText(text)
}

func (t *TextButton) SetImage(image *ebiten.Image) {
	t.image.SetImage(image)
}

func (t *TextButton) SetTextColor(clr color.Color) {
	if draw.EqualColor(t.textColor, clr) {
		return
	}
	t.textColor = clr
	guigui.RequestRedraw(t)
}

func (t *TextButton) SetForcePressed(forcePressed bool) {
	t.button.SetForcePressed(forcePressed)
}

func (t *TextButton) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	appender.AppendChildWidgetWithBounds(&t.button, context.Bounds(t))

	s := context.Size(t)

	imgSize := textButtonImageSize(context)

	tw := t.text.TextSize(context).X
	if !context.IsEnabled(&t.button) {
		t.text.SetColor(draw.Color(context.ColorMode(), draw.ColorTypeBase, 0.5))
	} else {
		t.text.SetColor(t.textColor)
	}
	t.text.SetHorizontalAlign(HorizontalAlignCenter)
	t.text.SetVerticalAlign(VerticalAlignMiddle)
	textP := context.Position(t)
	if t.image.HasImage() {
		textP.X += (s.X - tw + UnitSize(context)/4) / 2
		textP.X -= (textButtonTextAndImagePadding(context) + imgSize) / 2
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
	imgP.X = textP.X + tw + textButtonTextAndImagePadding(context)
	imgP.Y += (s.Y - imgSize) / 2
	if t.button.isPressed(context) {
		imgP.Y += int(1 * context.Scale())
	}
	appender.AppendChildWidgetWithBounds(&t.image, image.Rectangle{
		Min: imgP,
		Max: imgP.Add(image.Pt(imgSize, imgSize)),
	})

	return nil
}

func (t *TextButton) DefaultSize(context *guigui.Context) image.Point {
	dh := defaultButtonSize(context).Y
	w := t.text.TextSize(context).X
	if t.image.HasImage() {
		imgSize := textButtonImageSize(context)
		return image.Pt(w+textButtonTextAndImagePadding(context)+imgSize+UnitSize(context)*3/4, dh)
	}
	return image.Pt(w+UnitSize(context), dh)
}

func textButtonImageSize(context *guigui.Context) int {
	return int(LineHeight(context))
}

func textButtonTextAndImagePadding(context *guigui.Context) int {
	return UnitSize(context) / 4
}
