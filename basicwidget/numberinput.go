// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"image"
	"math"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget/internal/draw"
)

type NumberInput struct {
	guigui.DefaultWidget

	textInput  TextInput
	upButton   TextButton
	downButton TextButton

	value      int64
	min        int64
	minSet     bool
	max        int64
	maxSet     bool
	stepMinus1 int64

	onValueChanged func(value int64)
}

func (n *NumberInput) SetOnValueChanged(f func(value int64)) {
	n.onValueChanged = f
}

func (n *NumberInput) Value() int64 {
	return n.value
}

func (n *NumberInput) SetValue(value int64) {
	value = min(max(value, n.MinimumValue()), n.MaximumValue())
	if n.value == value {
		return
	}
	n.value = value
	if n.onValueChanged != nil {
		n.onValueChanged(value)
	}
}

func (n *NumberInput) MinimumValue() int64 {
	if n.minSet {
		return n.min
	}
	return math.MinInt64
}

func (n *NumberInput) SetMinimumValue(minimum int64) {
	n.min = minimum
	n.minSet = true
	n.SetValue(n.value)
}

func (n *NumberInput) MaximumValue() int64 {
	if n.maxSet {
		return n.max
	}
	return math.MaxInt64
}

func (n *NumberInput) SetMaximumValue(maximum int64) {
	n.max = maximum
	n.maxSet = true
	n.SetValue(n.value)
}

func (n *NumberInput) SetStep(step int64) {
	n.stepMinus1 = step - 1
}

func (n *NumberInput) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	n.textInput.SetFilter(func(text string, start, end int) (string, int, int) {
		for len(text) > 0 {
			if _, err := strconv.ParseInt(text, 10, 64); err == nil {
				return text, start, end
			}
			text = text[:len(text)-1]
			start = min(start, len(text))
			end = min(end, len(text))
		}
		return "0", min(start, 1), min(end, 1)
	})
	n.textInput.SetHorizontalAlign(HorizontalAlignEnd)
	n.textInput.SetNumber(true)
	n.textInput.setPaddingRight(UnitSize(context) / 2)
	n.textInput.SetOnValueChanged(func(text string) {
		if text == "" {
			return
		}
		i, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return
		}
		n.SetValue(i)
		if n.onValueChanged != nil {
			n.onValueChanged(i)
		}
	})
	appender.AppendChildWidgetWithBounds(&n.textInput, context.Bounds(n))
	// HasFocusedChildWidget works after appending the child widget.
	if !context.HasFocusedChildWidget(n) {
		n.textInput.SetText(strconv.FormatInt(n.value, 10))
	}

	imgUp, err := theResourceImages.Get("keyboard_arrow_up", context.ColorMode())
	if err != nil {
		return err
	}
	imgDown, err := theResourceImages.Get("keyboard_arrow_down", context.ColorMode())
	if err != nil {
		return err
	}

	n.upButton.SetImage(imgUp)
	n.upButton.setSharpenCorners(draw.SharpenCorners{
		LowerLeft:  true,
		LowerRight: true,
	})
	n.upButton.setOnRepeat(func() {
		n.increment()
	})

	b := context.Bounds(n)
	appender.AppendChildWidgetWithBounds(&n.upButton, image.Rectangle{
		Min: image.Point{
			X: b.Max.X - UnitSize(context)*3/4,
			Y: b.Min.Y,
		},
		Max: image.Point{
			X: b.Max.X,
			Y: b.Min.Y + b.Dy()/2,
		},
	})

	n.downButton.SetImage(imgDown)
	n.downButton.setSharpenCorners(draw.SharpenCorners{
		UpperLeft:  true,
		UpperRight: true,
	})
	n.downButton.setOnRepeat(func() {
		n.decrement()
	})

	appender.AppendChildWidgetWithBounds(&n.downButton, image.Rectangle{
		Min: image.Point{
			X: b.Max.X - UnitSize(context)*3/4,
			Y: b.Min.Y + b.Dy()/2,
		},
		Max: b.Max,
	})

	return nil
}

func (n *NumberInput) HandleButtonInput(context *guigui.Context) guigui.HandleInputResult {
	if isKeyRepeating(ebiten.KeyUp) {
		n.increment()
		return guigui.HandleInputByWidget(n)
	}
	if isKeyRepeating(ebiten.KeyDown) {
		n.decrement()
		return guigui.HandleInputByWidget(n)
	}
	return guigui.HandleInputResult{}
}

func (n *NumberInput) increment() {
	step := n.stepMinus1 + 1
	n.SetValue(n.value + step)
	n.textInput.SetText(strconv.FormatInt(n.value, 10))
}

func (n *NumberInput) decrement() {
	step := n.stepMinus1 + 1
	n.SetValue(n.value - step)
	n.textInput.SetText(strconv.FormatInt(n.value, 10))
}

func (n *NumberInput) DefaultSize(context *guigui.Context) image.Point {
	return n.textInput.DefaultSize(context)
}
