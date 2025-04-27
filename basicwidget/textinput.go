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

type TextInput struct {
	guigui.DefaultWidget

	background    textInputBackground
	text          Text
	frame         textInputFrame
	scrollOverlay ScrollOverlay
	focus         textInputFocus

	prevFocused bool
	prevStart   int
	prevEnd     int

	onTextAndSelectionChanged func(text string, start, end int)
}

func (t *TextInput) SetOnEnterPressed(f func(text string)) {
	t.text.SetOnEnterPressed(f)
}

func (t *TextInput) SetOnValueChanged(f func(text string)) {
	t.text.SetOnValueChanged(f)
}

func (t *TextInput) SetTextAndSelectionChanged(f func(text string, start, end int)) {
	t.onTextAndSelectionChanged = f
}

func (t *TextInput) Text() string {
	return t.text.Text()
}

func (t *TextInput) SetText(text string) {
	t.text.SetText(text)
}

func (t *TextInput) SetMultiline(multiline bool) {
	t.text.SetMultiline(multiline)
}

func (t *TextInput) SetHorizontalAlign(halign HorizontalAlign) {
	t.text.SetHorizontalAlign(halign)
}

func (t *TextInput) SetVerticalAlign(valign VerticalAlign) {
	t.text.SetVerticalAlign(valign)
}

func (t *TextInput) SetAutoWrap(autoWrap bool) {
	t.text.SetAutoWrap(autoWrap)
}

func (t *TextInput) SelectAll() {
	t.text.selectAll()
}

func (t *TextInput) SetFilter(filter TextFilter) {
	t.text.SetFilter(filter)
}

func (t *TextInput) SetNumber(number bool) {
	t.text.SetNumber(number)
}

func textInputPadding(context *guigui.Context) image.Point {
	x := UnitSize(context) / 2
	y := int(float64(UnitSize(context))-LineHeight(context)) / 2
	return image.Pt(x, y)
}

func (t *TextInput) scrollContentSize(context *guigui.Context) image.Point {
	padding := textInputPadding(context)
	return t.text.TextSize(context).Add(image.Pt(2*padding.X, 2*padding.Y))
}

func (t *TextInput) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	if t.prevFocused != context.HasFocusedChildWidget(t) {
		t.prevFocused = context.HasFocusedChildWidget(t)
		guigui.RequestRedraw(t)
	}
	if context.IsFocused(t) {
		context.SetFocused(&t.text, true)
		guigui.RequestRedraw(t)
	}

	padding := textInputPadding(context)

	t.scrollOverlay.SetContentSize(context, t.scrollContentSize(context))

	appender.AppendChildWidgetWithBounds(&t.background, context.Bounds(t))

	t.text.SetEditable(true)
	t.text.SetColor(draw.TextColor(context.ColorMode(), context.IsEnabled(t)))

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
		t.focus.textInput = t
		w := textInputFocusBorderWidth(context)
		p := context.Position(t).Add(image.Pt(-w, -w))
		appender.AppendChildWidgetWithPosition(&t.focus, p)
	}

	return nil
}

func (t *TextInput) adjustScrollOffsetIfNeeded(context *guigui.Context) {
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
	padding := textInputPadding(context)
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

func (t *TextInput) HandlePointingInput(context *guigui.Context) guigui.HandleInputResult {
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

func (t *TextInput) CursorShape(context *guigui.Context) (ebiten.CursorShapeType, bool) {
	return t.text.CursorShape(context)
}

func (t *TextInput) DefaultSize(context *guigui.Context) image.Point {
	if t.text.IsMultiline() {
		return image.Pt(6*UnitSize(context), 4*UnitSize(context))
	}
	return image.Pt(6*UnitSize(context), UnitSize(context))
}

type textInputBackground struct {
	guigui.DefaultWidget
}

func (t *textInputBackground) Draw(context *guigui.Context, dst *ebiten.Image) {
	bounds := context.Bounds(t)
	clr := draw.ControlColor(context.ColorMode(), context.IsEnabled(t))
	draw.DrawRoundedRect(context, dst, bounds, clr, RoundedCornerRadius(context))
}

type textInputFrame struct {
	guigui.DefaultWidget
}

func (t *textInputFrame) Draw(context *guigui.Context, dst *ebiten.Image) {
	bounds := context.Bounds(t)
	clr1, clr2 := draw.BorderColors(context.ColorMode(), draw.RoundedRectBorderTypeInset, false)
	draw.DrawRoundedRectBorder(context, dst, bounds, clr1, clr2, RoundedCornerRadius(context), float32(1*context.Scale()), draw.RoundedRectBorderTypeInset)
}

func (t *textInputFrame) PassThrough() bool {
	return true
}

func textInputFocusBorderWidth(context *guigui.Context) int {
	return int(4 * context.Scale())
}

type textInputFocus struct {
	guigui.DefaultWidget

	textInput *TextInput
}

func (t *textInputFocus) Draw(context *guigui.Context, dst *ebiten.Image) {
	bounds := context.Bounds(t.textInput)
	w := textInputFocusBorderWidth(context)
	clr := draw.Color(context.ColorMode(), draw.ColorTypeAccent, 0.8)
	bounds = bounds.Inset(-w)
	draw.DrawRoundedRectBorder(context, dst, bounds, clr, clr, w+RoundedCornerRadius(context), float32(w), draw.RoundedRectBorderTypeRegular)
}

func (t *textInputFocus) ZDelta() int {
	return 1
}

func (t *textInputFocus) DefaultSize(context *guigui.Context) image.Point {
	return context.Size(t.textInput).Add(image.Pt(2*textInputFocusBorderWidth(context), 2*textInputFocusBorderWidth(context)))
}

func (t *textInputFocus) PassThrough() bool {
	return true
}
