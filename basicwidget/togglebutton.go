// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget/internal/draw"
)

type ToggleButton struct {
	guigui.DefaultWidget

	pressed      bool
	value        bool
	onceRendered bool
	prevHovered  bool

	count int

	onValueChanged func(value bool)
}

func (t *ToggleButton) SetOnValueChanged(f func(value bool)) {
	t.onValueChanged = f
}

func (t *ToggleButton) Value() bool {
	return t.value
}

func (t *ToggleButton) SetValue(context *guigui.Context, value bool) {
	if t.value == value {
		return
	}

	t.value = value
	if t.onceRendered {
		t.count = toggleButtonMaxCount() - t.count
	}
	context.RequestRedraw(t)

	if t.onValueChanged != nil {
		t.onValueChanged(value)
	}
}

func toggleButtonMaxCount() int {
	return ebiten.TPS() / 12
}

func (t *ToggleButton) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	hovered := t.isHovered(context)
	if t.prevHovered != hovered {
		t.prevHovered = hovered
		context.RequestRedraw(t)
	}
	return nil
}

func (t *ToggleButton) HandlePointingInput(context *guigui.Context) guigui.HandleInputResult {
	if context.IsEnabled(t) && t.isHovered(context) && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		t.pressed = true
		t.SetValue(context, !t.value)
		return guigui.HandleInputByWidget(t)
	}
	if !context.IsEnabled(t) || !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		t.pressed = false
	}
	return guigui.HandleInputResult{}
}

func (t *ToggleButton) Update(context *guigui.Context) error {
	if t.count > 0 {
		t.count--
		context.RequestRedraw(t)
	}
	return nil
}

func (t *ToggleButton) CursorShape(context *guigui.Context) (ebiten.CursorShapeType, bool) {
	if t.canPress(context) || t.pressed {
		return ebiten.CursorShapePointer, true
	}
	return 0, true
}

func (t *ToggleButton) Draw(context *guigui.Context, dst *ebiten.Image) {
	rate := 1 - float64(t.count)/float64(toggleButtonMaxCount())

	bounds := context.Bounds(t)

	cm := context.ColorMode()
	backgroundColor := draw.Color(context.ColorMode(), draw.ColorTypeBase, 0.8)
	thumbColor := draw.Color2(cm, draw.ColorTypeBase, 1, 0.6)
	borderColor := draw.Color2(cm, draw.ColorTypeBase, 0.7, 0)
	if t.isActive(context) {
		thumbColor = draw.Color2(cm, draw.ColorTypeBase, 0.95, 0.55)
		borderColor = draw.Color2(cm, draw.ColorTypeBase, 0.7, 0)
	} else if t.canPress(context) {
		thumbColor = draw.Color2(cm, draw.ColorTypeBase, 0.975, 0.575)
		borderColor = draw.Color2(cm, draw.ColorTypeBase, 0.7, 0)
	} else if !context.IsEnabled(t) {
		thumbColor = draw.Color2(cm, draw.ColorTypeBase, 0.95, 0.55)
		borderColor = draw.Color2(cm, draw.ColorTypeBase, 0.8, 0.1)
	}

	// Background
	bgColorOff := backgroundColor
	bgColorOn := draw.Color(context.ColorMode(), draw.ColorTypeAccent, 0.5)
	var bgColor color.Color
	if t.value {
		bgColor = draw.MixColor(bgColorOff, bgColorOn, rate)
	} else {
		bgColor = draw.MixColor(bgColorOn, bgColorOff, rate)
	}
	r := bounds.Dy() / 2
	draw.DrawRoundedRect(context, dst, bounds, bgColor, r)

	// Border (upper)
	b := bounds
	b.Max.Y = b.Min.Y + b.Dy()/2
	draw.DrawRoundedRectBorder(context, dst.SubImage(b).(*ebiten.Image), bounds, borderColor, r, float32(1*context.Scale()), draw.RoundedRectBorderTypeInset)

	// Thumb
	cxOff := float64(bounds.Min.X) + float64(r)
	cxOn := float64(bounds.Max.X) - float64(r)
	var cx int
	if t.value {
		cx = int((1-rate)*cxOff + rate*cxOn)
	} else {
		cx = int((1-rate)*cxOn + rate*cxOff)
	}
	cy := bounds.Min.Y + r
	draw.DrawRoundedRect(context, dst, image.Rect(cx-r, cy-r, cx+r, cy+r), thumbColor, r)
	draw.DrawRoundedRectBorder(context, dst, image.Rect(cx-r, cy-r, cx+r, cy+r), borderColor, r, float32(1*context.Scale()), draw.RoundedRectBorderTypeOutset)

	// Border (lower)
	b = bounds
	b.Min.Y = b.Max.Y - b.Dy()/2
	draw.DrawRoundedRectBorder(context, dst.SubImage(b).(*ebiten.Image), bounds, borderColor, r, float32(1*context.Scale()), draw.RoundedRectBorderTypeInset)

	t.onceRendered = true
}

func (t *ToggleButton) canPress(context *guigui.Context) bool {
	return context.IsEnabled(t) && t.isHovered(context) && !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

func (t *ToggleButton) isHovered(context *guigui.Context) bool {
	return context.IsWidgetHitAt(t, image.Pt(ebiten.CursorPosition()))
}

func (t *ToggleButton) isActive(context *guigui.Context) bool {
	return context.IsEnabled(t) && t.isHovered(context) && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && t.pressed
}

func (t *ToggleButton) DefaultSize(context *guigui.Context) (int, int) {
	return int(LineHeight(context) * 1.75), int(LineHeight(context))
}
