// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"image"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
)

type NumberInput struct {
	guigui.DefaultWidget

	textInput TextInput

	nextValue string
	value     int64

	onValueChanged func(value int64)
}

func (n *NumberInput) SetOnValueChanged(f func(value int64)) {
	n.onValueChanged = f
}

func (n *NumberInput) Value() int64 {
	return n.value
}

func (n *NumberInput) SetValue(value int64) {
	n.nextValue = strconv.FormatInt(value, 10)
	n.value = value
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
	n.textInput.SetOnValueChanged(func(text string) {
		if text == "" {
			return
		}
		i, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return
		}
		n.value = i
		if n.onValueChanged != nil {
			n.onValueChanged(i)
		}
	})
	appender.AppendChildWidgetWithBounds(&n.textInput, context.Bounds(n))
	// HasFocusedChildWidget works after appending the child widget.
	if n.nextValue != "" && !context.HasFocusedChildWidget(n) {
		n.textInput.SetText(n.nextValue)
		n.nextValue = ""
	}

	return nil
}

func (n *NumberInput) HandleButtonInput(context *guigui.Context) guigui.HandleInputResult {
	if isKeyRepeating(ebiten.KeyUp) {
		n.textInput.SetText(strconv.FormatInt(n.value+1, 10))
		return guigui.HandleInputByWidget(n)
	}
	if isKeyRepeating(ebiten.KeyDown) {
		n.textInput.SetText(strconv.FormatInt(n.value-1, 10))
		return guigui.HandleInputByWidget(n)
	}
	return guigui.HandleInputResult{}
}

func (n *NumberInput) DefaultSize(context *guigui.Context) image.Point {
	return n.textInput.DefaultSize(context)
}
