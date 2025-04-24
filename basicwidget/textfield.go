// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget/internal/draw"
)

type TextField struct {
	guigui.DefaultWidget

	text  Text
	focus textFieldFocus

	readonly bool

	prevFocused bool
}

func (t *TextField) SetOnEnterPressed(f func(text string)) {
	t.text.SetOnEnterPressed(f)
}

func (t *TextField) SetOnValueChanged(f func(text string)) {
	t.text.SetOnValueChanged(f)
}

func (t *TextField) Text() string {
	return t.text.Text()
}

func (t *TextField) SetText(text string) {
	t.text.SetText(text)
}

func (t *TextField) SetMultiline(multiline bool) {
	t.text.SetMultiline(multiline)
}

func (t *TextField) SetHorizontalAlign(halign HorizontalAlign) {
	t.text.SetHorizontalAlign(halign)
}

func (t *TextField) SetVerticalAlign(valign VerticalAlign) {
	t.text.SetVerticalAlign(valign)
}

func (t *TextField) SetEditable(editable bool) {
	t.text.SetEditable(editable)
	t.readonly = !editable
}

func (t *TextField) SelectAll() {
	t.text.selectAll()
}

func (t *TextField) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	if t.prevFocused != context.HasFocusedChildWidget(t) {
		t.prevFocused = context.HasFocusedChildWidget(t)
		guigui.RequestRedraw(t)
	}
	if context.IsFocused(t) {
		context.SetFocused(&t.text, true)
		guigui.RequestRedraw(t)
	}

	t.text.SetEditable(true)
	b := context.Bounds(t)
	b.Min.X += UnitSize(context) / 2
	b.Max.X -= UnitSize(context) / 2
	// TODO: Consider multiline.
	if !t.text.IsMultiline() {
		t.text.SetVerticalAlign(VerticalAlignMiddle)
	}
	appender.AppendChildWidgetWithBounds(&t.text, b)

	if context.HasFocusedChildWidget(t) {
		t.focus.textField = t
		w := textFieldFocusBorderWidth(context)
		p := context.Position(t).Add(image.Pt(-w, -w))
		appender.AppendChildWidgetWithPosition(&t.focus, p)
	}

	return nil
}

func (t *TextField) HandlePointingInput(context *guigui.Context) guigui.HandleInputResult {
	if context.IsWidgetHitAt(t, image.Pt(ebiten.CursorPosition())) {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			context.SetFocused(&t.text, true)
			t.text.selectAll()
			return guigui.HandleInputByWidget(t)
		}
	}
	return guigui.HandleInputResult{}
}

func (t *TextField) Draw(context *guigui.Context, dst *ebiten.Image) {
	bounds := context.Bounds(t)
	draw.DrawRoundedRect(context, dst, bounds, draw.Color2(context.ColorMode(), draw.ColorTypeBase, 1, 0.3), RoundedCornerRadius(context))
	clr1, clr2 := draw.BorderColors(context.ColorMode(), draw.RoundedRectBorderTypeInset, false)
	draw.DrawRoundedRectBorder(context, dst, bounds, clr1, clr2, RoundedCornerRadius(context), float32(1*context.Scale()), draw.RoundedRectBorderTypeInset)
}

func (t *TextField) DefaultSize(context *guigui.Context) image.Point {
	// TODO: Increase the height for multiple lines.
	return image.Pt(6*UnitSize(context), UnitSize(context))
}

func textFieldFocusBorderWidth(context *guigui.Context) int {
	return int(4 * context.Scale())
}

type textFieldFocus struct {
	guigui.DefaultWidget

	textField *TextField
}

func (t *textFieldFocus) Draw(context *guigui.Context, dst *ebiten.Image) {
	bounds := context.Bounds(t.textField)
	w := textFieldFocusBorderWidth(context)
	clr := draw.Color(context.ColorMode(), draw.ColorTypeAccent, 0.8)
	bounds = bounds.Inset(-w)
	draw.DrawRoundedRectBorder(context, dst, bounds, clr, clr, w+RoundedCornerRadius(context), float32(w), draw.RoundedRectBorderTypeRegular)
}

func (t *textFieldFocus) ZDelta() int {
	return 1
}

func (t *textFieldFocus) DefaultSize(context *guigui.Context) image.Point {
	return context.Size(t.textField).Add(image.Pt(2*textFieldFocusBorderWidth(context), 2*textFieldFocusBorderWidth(context)))
}

func (t *textFieldFocus) PassThrough() bool {
	return true
}
