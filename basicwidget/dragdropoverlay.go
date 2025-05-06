// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Guigui Authors

package basicwidget

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/hajimehoshi/guigui"
)

type dragDropOverlay[T any] struct {
	guigui.DefaultWidget

	object    T
	objectSet bool

	onDropped func(object T)
}

func (d *dragDropOverlay[T]) SetOnDropped(f func(object T)) {
	d.onDropped = f
}

func (d *dragDropOverlay[T]) IsDragging() bool {
	return d.objectSet
}

func (d *dragDropOverlay[T]) Start(object T) {
	d.object = object
	d.objectSet = true
}

func (d *dragDropOverlay[T]) HandlePointingInput(context *guigui.Context) guigui.HandleInputResult {
	if d.objectSet {
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			if image.Pt(ebiten.CursorPosition()).In(context.VisibleBounds(d)) {
				if d.onDropped != nil {
					d.onDropped(d.object)
				}
			}
			var zero T
			d.object = zero
			d.objectSet = false
			return guigui.HandleInputResult{}
		}
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			var zero T
			d.object = zero
			d.objectSet = false
		}
		return guigui.HandleInputResult{}
	}

	return guigui.HandleInputResult{}
}
