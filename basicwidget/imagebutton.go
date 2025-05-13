// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Guigui Authors

package basicwidget

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/guigui"
)

type ImageButton struct {
	guigui.DefaultWidget

	button Button
	image  Image
}

func (i *ImageButton) SetOnDown(f func()) {
	i.button.SetOnDown(f)
}

func (i *ImageButton) SetOnUp(f func()) {
	i.button.SetOnUp(f)
}

func (i *ImageButton) SetImage(image *ebiten.Image) {
	i.image.SetImage(image)
}

func (i *ImageButton) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	appender.AppendChildWidgetWithBounds(&i.button, context.Bounds(i))

	b := context.Bounds(i)
	if i.button.isPressed(context) {
		b = b.Add(image.Pt(0, int(1*context.Scale())))
	}
	appender.AppendChildWidgetWithBounds(&i.image, b)

	return nil
}
