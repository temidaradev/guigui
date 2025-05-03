// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"image"
	"math/big"
	"strconv"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget/internal/draw"
)

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type NumberInput[T Integer] struct {
	guigui.DefaultWidget

	textInput  TextInput
	upButton   TextButton
	downButton TextButton

	value   T
	min     T
	minSet  bool
	max     T
	maxSet  bool
	step    T
	stepSet bool

	onValueChanged func(value T)
}

func (n *NumberInput[T]) IsEditable() bool {
	return n.textInput.IsEditable()
}

func (n *NumberInput[T]) SetEditable(editable bool) {
	n.textInput.SetEditable(editable)
}

func (n *NumberInput[T]) SetOnValueChanged(f func(value T)) {
	n.onValueChanged = f
}

func (n *NumberInput[T]) Value() T {
	return n.value
}

func (n *NumberInput[T]) SetValue(value T) {
	value = min(max(value, n.MinimumValue()), n.MaximumValue())
	if n.value == value {
		return
	}
	n.value = value
	if isSigned[T]() {
		n.textInput.SetText(strconv.FormatInt(int64(n.value), 10))
	} else {
		n.textInput.SetText(strconv.FormatUint(uint64(n.value), 10))
	}
	if n.onValueChanged != nil {
		n.onValueChanged(value)
	}
}

func (n *NumberInput[T]) MinimumValue() T {
	if n.minSet {
		return n.min
	}
	return minInteger[T]()
}

func (n *NumberInput[T]) SetMinimumValue(minimum T) {
	n.min = minimum
	n.minSet = true
	n.SetValue(n.value)
}

func (n *NumberInput[T]) MaximumValue() T {
	if n.maxSet {
		return n.max
	}
	return maxInteger[T]()
}

func (n *NumberInput[T]) SetMaximumValue(maximum T) {
	n.max = maximum
	n.maxSet = true
	n.SetValue(n.value)
}

func (n *NumberInput[T]) SetStep(step T) {
	n.step = step
	n.stepSet = true
}

func (n *NumberInput[T]) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	n.textInput.SetHorizontalAlign(HorizontalAlignEnd)
	n.textInput.SetNumber(true)
	n.textInput.setPaddingRight(UnitSize(context) / 2)
	n.textInput.SetOnValueChanged(func(text string, committed bool) {
		if !committed {
			return
		}
		var i big.Int
		if _, ok := i.SetString(text, 10); !ok {
			return
		}
		var v T
		if isSigned[T]() {
			var min big.Int
			min.SetInt64(int64(n.MinimumValue()))
			var max big.Int
			max.SetInt64(int64(n.MaximumValue()))
			if i.Cmp(&min) < 0 {
				v = T(n.MinimumValue())
			} else if i.Cmp(&max) > 0 {
				v = T(n.MaximumValue())
			} else {
				v = T(i.Int64())
			}
		} else {
			var min big.Int
			min.SetUint64(uint64(n.MinimumValue()))
			var max big.Int
			max.SetUint64(uint64(n.MaximumValue()))
			if i.Cmp(&min) < 0 {
				v = T(n.MinimumValue())
			} else if i.Cmp(&max) > 0 {
				v = T(n.MaximumValue())
			} else {
				v = T(i.Uint64())
			}
		}
		n.SetValue(v)
		if n.onValueChanged != nil {
			n.onValueChanged(v)
		}
	})
	appender.AppendChildWidgetWithBounds(&n.textInput, context.Bounds(n))
	// HasFocusedChildWidget works after appending the child widget.
	if !context.HasFocusedChildWidget(n) {
		if isSigned[T]() {
			n.textInput.SetText(strconv.FormatInt(int64(n.value), 10))
		} else {
			n.textInput.SetText(strconv.FormatUint(uint64(n.value), 10))
		}
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
	n.upButton.setPairedButton(&n.downButton)
	n.upButton.setOnRepeat(func() {
		n.increment()
	})
	context.SetEnabled(&n.upButton, n.IsEditable() && n.value < n.MaximumValue())

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
	n.downButton.setPairedButton(&n.upButton)
	n.downButton.setOnRepeat(func() {
		n.decrement()
	})
	context.SetEnabled(&n.downButton, n.IsEditable() && n.value > n.MinimumValue())

	appender.AppendChildWidgetWithBounds(&n.downButton, image.Rectangle{
		Min: image.Point{
			X: b.Max.X - UnitSize(context)*3/4,
			Y: b.Min.Y + b.Dy()/2,
		},
		Max: b.Max,
	})

	return nil
}

func (n *NumberInput[T]) HandleButtonInput(context *guigui.Context) guigui.HandleInputResult {
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

func (n *NumberInput[T]) increment() {
	if !n.IsEditable() {
		return
	}
	var step T = 1
	if n.stepSet {
		step = n.step
	}
	n.SetValue(min(increment(n.value, step), n.MaximumValue()))
}

func (n *NumberInput[T]) decrement() {
	if !n.IsEditable() {
		return
	}
	var step T = 1
	if n.stepSet {
		step = n.step
	}
	n.SetValue(max(decrement(n.value, step), n.MinimumValue()))
}

func (n *NumberInput[T]) DefaultSize(context *guigui.Context) image.Point {
	return n.textInput.DefaultSize(context)
}

func isSigned[T Integer]() bool {
	var zero T
	zero--
	return zero < 0
}

func maxInteger[T Integer]() T {
	if isSigned[T]() {
		var zero T
		return 1<<(unsafe.Sizeof(zero)*8-1) - 1
	}
	return ^T(0)
}

func minInteger[T Integer]() T {
	if isSigned[T]() {
		var zero T
		return 1 << (unsafe.Sizeof(zero)*8 - 1)
	}
	return 0
}

func increment[T Integer](value T, step T) T {
	if value+step < value {
		return maxInteger[T]()
	}
	return value + step
}

func decrement[T Integer](value T, step T) T {
	if value-step > value {
		return minInteger[T]()
	}
	return value - step
}
