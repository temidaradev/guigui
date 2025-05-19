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

type Button struct {
	guigui.DefaultWidget

	button    baseButton
	content   guigui.Widget
	text      Text
	icon      Image
	iconAlign IconAlign

	textColor color.Color
}

func (t *Button) SetOnDown(f func()) {
	t.button.SetOnDown(f)
}

func (t *Button) SetOnUp(f func()) {
	t.button.SetOnUp(f)
}

func (b *Button) setOnRepeat(f func()) {
	b.button.setOnRepeat(f)
}

func (t *Button) SetContent(content guigui.Widget) {
	t.content = content
}

func (t *Button) SetText(text string) {
	t.text.SetValue(text)
}

func (t *Button) SetTextBold(bold bool) {
	t.text.SetBold(bold)
}

func (t *Button) SetIcon(icon *ebiten.Image) {
	t.icon.SetImage(icon)
}

func (t *Button) SetIconAlign(align IconAlign) {
	if t.iconAlign == align {
		return
	}
	t.iconAlign = align
	guigui.RequestRedraw(t)
}

func (t *Button) SetTextColor(clr color.Color) {
	if draw.EqualColor(t.textColor, clr) {
		return
	}
	t.textColor = clr
	guigui.RequestRedraw(t)
}

func (t *Button) setPairedButton(pair *Button) {
	t.button.setPairedButton(&pair.button)
}

func (t *Button) setKeepPressed(keep bool) {
	t.button.setKeepPressed(keep)
}

func (t *Button) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	appender.AppendChildWidgetWithBounds(&t.button, context.Bounds(t))

	if t.content != nil {
		r := t.button.radius(context)
		contentP := context.Position(t).Add(image.Pt(r, r))
		contentSize := t.contentSize(context)
		if t.button.isPressed(context) {
			contentP.Y += int(1 * context.Scale())
		}
		appender.AppendChildWidgetWithBounds(t.content, image.Rectangle{
			Min: contentP,
			Max: contentP.Add(contentSize),
		})
	}

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
		switch t.iconAlign {
		case IconAlignStart:
			textP.X += buttonEdgeAndImagePadding(context)
			textP.X += imgSize.X + buttonTextAndImagePadding(context)
		case IconAlignEnd:
			textP.X += buttonEdgeAndTextPadding(context)
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
		switch t.iconAlign {
		case IconAlignStart:
			imgP.X += buttonEdgeAndImagePadding(context)
		case IconAlignEnd:
			imgP.X += buttonEdgeAndTextPadding(context)
			imgP.X += tw + buttonTextAndImagePadding(context)
		}
	} else {
		imgP.X += (s.X - imgSize.X) / 2
	}
	imgP.Y += (s.Y - imgSize.Y) / 2
	if t.button.isPressed(context) {
		imgP.Y += int(1 * context.Scale())
	}
	appender.AppendChildWidgetWithBounds(&t.icon, image.Rectangle{
		Min: imgP,
		Max: imgP.Add(imgSize),
	})

	return nil
}

func (t *Button) DefaultSize(context *guigui.Context) image.Point {
	return t.defaultSize(context, false)
}

func (t *Button) defaultSize(context *guigui.Context, forceBold bool) image.Point {
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
			w += buttonTextAndImagePadding(context)
		}
		w += buttonEdgeAndTextPadding(context)
		w += buttonEdgeAndImagePadding(context)
		return image.Pt(w, dh)
	}
	return image.Pt(w+UnitSize(context), dh)
}

func (t *Button) setSharpenCorners(sharpenCorners draw.SharpenCorners) {
	t.button.setSharpenCorners(sharpenCorners)
}

func buttonTextAndImagePadding(context *guigui.Context) int {
	return UnitSize(context) / 4
}

func buttonEdgeAndTextPadding(context *guigui.Context) int {
	return UnitSize(context) / 2
}

func buttonEdgeAndImagePadding(context *guigui.Context) int {
	return UnitSize(context) / 4
}

func (t *Button) defaultIconSize(context *guigui.Context) int {
	return int(LineHeight(context))
}

func (t *Button) iconSize(context *guigui.Context) image.Point {
	s := context.Size(t)
	if t.text.Value() != "" {
		s := min(t.defaultIconSize(context), s.X, s.Y)
		return image.Pt(s, s)
	}
	r := t.button.radius(context)
	w := max(0, s.X-2*r)
	h := max(int(LineHeight(context)), s.Y-2*r)
	return image.Pt(w, h)
}

func (t *Button) contentSize(context *guigui.Context) image.Point {
	s := context.Size(t)
	r := t.button.radius(context)
	w := max(0, s.X-2*r)
	h := max(0, s.Y-2*r)
	return image.Pt(w, h)
}

func (t *Button) setUseAccentColor(use bool) {
	t.button.setUseAccentColor(use)
}
