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

type Button struct {
	guigui.DefaultWidget

	pressed            bool
	forcePressed       bool
	widthMinusDefault  int
	heightMinusDefault int
	borderInvisible    bool
	prevHovered        bool

	onDown func()
	onUp   func()
}

func (b *Button) SetOnDown(f func()) {
	b.onDown = f
}

func (b *Button) SetOnUp(f func()) {
	b.onUp = f
}

func (b *Button) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	hovered := b.isHovered()
	if b.prevHovered != hovered {
		b.prevHovered = hovered
		guigui.RequestRedraw(b)
	}
	return nil
}

func (b *Button) HandlePointingInput(context *guigui.Context) guigui.HandleInputResult {
	if guigui.IsEnabled(b) && b.isHovered() {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			b.pressed = true
			if b.onDown != nil {
				b.onDown()
			}
			return guigui.HandleInputByWidget(b)
		}
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && b.pressed {
			b.pressed = false
			if b.onUp != nil {
				b.onUp()
			}
			return guigui.HandleInputByWidget(b)
		}
	}
	if !guigui.IsEnabled(b) || !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		b.pressed = false
	}
	return guigui.HandleInputResult{}
}

func (b *Button) CursorShape(context *guigui.Context) (ebiten.CursorShapeType, bool) {
	if b.canPress() || b.pressed {
		return ebiten.CursorShapePointer, true
	}
	return 0, true
}

func (b *Button) Draw(context *guigui.Context, dst *ebiten.Image) {
	// TODO: In the dark theme, the color should be different.
	// At least, shadow should be darker.
	// See macOS's buttons.
	cm := context.ColorMode()
	backgroundColor := draw.Color2(cm, draw.ColorTypeBase, 1, 0.3)
	borderColor := draw.Color2(cm, draw.ColorTypeBase, 0.7, 0)
	if b.isActive() || b.forcePressed {
		backgroundColor = draw.Color2(cm, draw.ColorTypeBase, 0.95, 0.25)
		borderColor = draw.Color2(cm, draw.ColorTypeBase, 0.7, 0)
	} else if b.canPress() {
		backgroundColor = draw.Color2(cm, draw.ColorTypeBase, 0.975, 0.275)
		borderColor = draw.Color2(cm, draw.ColorTypeBase, 0.7, 0)
	} else if !guigui.IsEnabled(b) {
		backgroundColor = draw.Color2(cm, draw.ColorTypeBase, 0.95, 0.25)
		borderColor = draw.Color2(cm, draw.ColorTypeBase, 0.8, 0.1)
	}

	bounds := guigui.Bounds(b)
	r := min(RoundedCornerRadius(context), bounds.Dx()/4, bounds.Dy()/4)
	border := !b.borderInvisible
	if guigui.IsEnabled(b) && b.isHovered() || b.forcePressed {
		border = true
	}
	if border || b.isActive() || b.forcePressed {
		bounds := bounds.Inset(int(1 * context.Scale()))
		draw.DrawRoundedRect(context, dst, bounds, backgroundColor, r)
	}

	if border {
		borderType := draw.RoundedRectBorderTypeOutset
		if b.isActive() || b.forcePressed {
			borderType = draw.RoundedRectBorderTypeInset
		} else if !guigui.IsEnabled(b) {
			borderType = draw.RoundedRectBorderTypeRegular
		}
		draw.DrawRoundedRectBorder(context, dst, bounds, borderColor, r, float32(1*context.Scale()), borderType)
	}
}

func (b *Button) canPress() bool {
	return guigui.IsEnabled(b) && b.isHovered() && !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

func (b *Button) isHovered() bool {
	return guigui.IsWidgetHitAt(b, image.Pt(ebiten.CursorPosition()))
}

func (b *Button) isActive() bool {
	return guigui.IsEnabled(b) && b.isHovered() && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && b.pressed
}

func (b *Button) SetForcePressed(pressed bool) {
	b.forcePressed = pressed
}

func defaultButtonSize(context *guigui.Context) (int, int) {
	return 6 * UnitSize(context), UnitSize(context)
}

func (b *Button) SetSize(context *guigui.Context, width, height int) {
	dw, dh := defaultButtonSize(context)
	b.widthMinusDefault = width - dw
	b.heightMinusDefault = height - dh
}

func (b *Button) DefaultSize(context *guigui.Context) (int, int) {
	dw, dh := defaultButtonSize(context)
	return b.widthMinusDefault + dw, b.heightMinusDefault + dh
}
