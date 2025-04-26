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

	background    textFieldBackground
	text          Text
	frame         textFieldFrame
	scrollOverlay ScrollOverlay
	focus         textFieldFocus

	prevFocused bool
	prevStart   int
	prevEnd     int

	onTextAndSelectionChanged func(text string, start, end int)
}

func (t *TextField) SetOnEnterPressed(f func(text string)) {
	t.text.SetOnEnterPressed(f)
}

func (t *TextField) SetOnValueChanged(f func(text string)) {
	t.text.SetOnValueChanged(f)
}

func (t *TextField) SetTextAndSelectionChanged(f func(text string, start, end int)) {
	t.onTextAndSelectionChanged = f
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
}

func (t *TextField) SelectAll() {
	t.text.selectAll()
}

func textFieldPadding(context *guigui.Context) image.Point {
	x := UnitSize(context) / 2
	y := int(float64(UnitSize(context))-LineHeight(context)) / 2
	return image.Pt(x, y)
}

func (t *TextField) scrollContentSize(context *guigui.Context) image.Point {
	padding := textFieldPadding(context)
	return t.text.TextSize(context).Add(image.Pt(2*padding.X, 2*padding.Y))
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

	padding := textFieldPadding(context)

	t.scrollOverlay.SetContentSize(context, t.scrollContentSize(context))

	appender.AppendChildWidgetWithBounds(&t.background, context.Bounds(t))

	t.text.SetEditable(true)

	pt := context.Position(t)
	s := t.text.TextSize(context)
	s.X = max(s.X, context.Size(t).X-2*padding.X)
	s.Y = max(s.Y, context.Size(t).Y-2*padding.Y)
	b := image.Rectangle{
		Min: pt,
		Max: pt.Add(s),
	}
	b = b.Add(padding)

	// Set the content size before adjustScrollOffset, as the size affects the adjustment.
	context.SetSize(&t.text, b.Size())
	t.adjustScrollOffsetIfNeeded(context)
	offsetX, offsetY := t.scrollOverlay.Offset()
	b = b.Add(image.Pt(int(offsetX), int(offsetY)))
	appender.AppendChildWidgetWithPosition(&t.text, b.Min)

	appender.AppendChildWidgetWithBounds(&t.frame, context.Bounds(t))

	context.SetVisible(&t.scrollOverlay, t.text.IsMultiline())
	appender.AppendChildWidgetWithBounds(&t.scrollOverlay, context.Bounds(t))

	if context.HasFocusedChildWidget(t) {
		t.focus.textField = t
		w := textFieldFocusBorderWidth(context)
		p := context.Position(t).Add(image.Pt(-w, -w))
		appender.AppendChildWidgetWithPosition(&t.focus, p)
	}

	return nil
}

func (t *TextField) adjustScrollOffsetIfNeeded(context *guigui.Context) {
	start, end, ok := t.text.selectionToDraw(context)
	if !ok {
		return
	}
	if t.prevStart == start && t.prevEnd == end {
		return
	}
	t.prevStart = start
	t.prevEnd = end
	bounds := context.Bounds(t)
	padding := textFieldPadding(context)
	if pos, ok := t.text.textPosition(context, end, true); ok {
		dx := min(float64(bounds.Max.X-padding.X)-pos.X, 0)
		dy := min(float64(bounds.Max.Y-padding.Y)-pos.Bottom, 0)
		t.scrollOverlay.SetOffsetByDelta(context, t.scrollContentSize(context), dx, dy)
	}
	if pos, ok := t.text.textPosition(context, start, true); ok {
		dx := max(float64(bounds.Min.X+padding.X)-pos.X, 0)
		dy := max(float64(bounds.Min.Y+padding.Y)-pos.Top, 0)
		t.scrollOverlay.SetOffsetByDelta(context, t.scrollContentSize(context), dx, dy)
	}
}

func (t *TextField) HandlePointingInput(context *guigui.Context) guigui.HandleInputResult {
	cp := image.Pt(ebiten.CursorPosition())
	if context.IsWidgetHitAt(t, cp) {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			context.SetFocused(&t.text, true)
			idx := t.text.textIndexFromPosition(context, cp, false)
			t.text.setSelection(idx, idx)
			return guigui.HandleInputByWidget(t)
		}
	}
	return guigui.HandleInputResult{}
}

func (t *TextField) CursorShape(context *guigui.Context) (ebiten.CursorShapeType, bool) {
	return t.text.CursorShape(context)
}

func (t *TextField) DefaultSize(context *guigui.Context) image.Point {
	if t.text.IsMultiline() {
		return image.Pt(6*UnitSize(context), 4*UnitSize(context))
	}
	return image.Pt(6*UnitSize(context), UnitSize(context))
}

type textFieldBackground struct {
	guigui.DefaultWidget
}

func (t *textFieldBackground) Draw(context *guigui.Context, dst *ebiten.Image) {
	bounds := context.Bounds(t)
	draw.DrawRoundedRect(context, dst, bounds, draw.Color2(context.ColorMode(), draw.ColorTypeBase, 1, 0.3), RoundedCornerRadius(context))
}

type textFieldFrame struct {
	guigui.DefaultWidget
}

func (t *textFieldFrame) Draw(context *guigui.Context, dst *ebiten.Image) {
	bounds := context.Bounds(t)
	clr1, clr2 := draw.BorderColors(context.ColorMode(), draw.RoundedRectBorderTypeInset, false)
	draw.DrawRoundedRectBorder(context, dst, bounds, clr1, clr2, RoundedCornerRadius(context), float32(1*context.Scale()), draw.RoundedRectBorderTypeInset)
}

func (t *textFieldFrame) PassThrough() bool {
	return true
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
