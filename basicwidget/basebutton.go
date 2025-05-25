// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Guigui Authors

package basicwidget

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget/internal/draw"
)

type baseButton struct {
	guigui.DefaultWidget

	pressed         bool
	keepPressed     bool
	useAccentColor  bool
	borderInvisible bool
	prevHovered     bool
	sharpenCorners  draw.SharpenCorners
	pairedButton    *baseButton

	onDown   func()
	onUp     func()
	onRepeat func()
}

func (b *baseButton) SetOnDown(f func()) {
	b.onDown = f
}

func (b *baseButton) SetOnUp(f func()) {
	b.onUp = f
}

func (b *baseButton) setOnRepeat(f func()) {
	b.onRepeat = f
}

func (b *baseButton) setPairedButton(pair *baseButton) {
	b.pairedButton = pair
}

func (b *baseButton) setPressed(pressed bool) {
	if b.pressed == pressed {
		return
	}
	b.pressed = pressed
	guigui.RequestRedraw(b)
}

func (b *baseButton) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	// TODO: Do not call isHovered in Build (#52).
	hovered := b.isHovered(context)
	if b.prevHovered != hovered {
		b.prevHovered = hovered
		guigui.RequestRedraw(b)
	}
	return nil
}

func (b *baseButton) HandlePointingInput(context *guigui.Context) guigui.HandleInputResult {
	if b.isHovered(context) && !b.keepPressed {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			context.SetFocused(b, true)
			b.setPressed(true)
			if b.onDown != nil {
				b.onDown()
			}
			if isMouseButtonRepeating(ebiten.MouseButtonLeft) {
				if b.onRepeat != nil {
					b.onRepeat()
				}
			}
			return guigui.HandleInputByWidget(b)
		}
		if (b.pressed || b.pairedButton != nil && b.pairedButton.pressed) && isMouseButtonRepeating(ebiten.MouseButtonLeft) {
			if b.onRepeat != nil {
				b.onRepeat()
			}
			return guigui.HandleInputByWidget(b)
		}
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && b.pressed {
			b.setPressed(false)
			if b.onUp != nil {
				b.onUp()
			}
			guigui.RequestRedraw(b)
			return guigui.HandleInputByWidget(b)
		}
	}
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		b.setPressed(false)
	}
	return guigui.HandleInputResult{}
}

func (b *baseButton) CursorShape(context *guigui.Context) (ebiten.CursorShapeType, bool) {
	if (b.canPress(context) || b.pressed || b.pairedButton != nil && b.pairedButton.pressed) && !b.keepPressed {
		return ebiten.CursorShapePointer, true
	}
	return 0, true
}

func (b *baseButton) radius(context *guigui.Context) int {
	bounds := context.Bounds(b)
	return min(RoundedCornerRadius(context), bounds.Dx()/4, bounds.Dy()/4)
}

func (b *baseButton) Draw(context *guigui.Context, dst *ebiten.Image) {
	cm := context.ColorMode()
	backgroundColor := draw.ControlColor(context.ColorMode(), context.IsEnabled(b))
	if context.IsEnabled(b) {
		if b.isPressed(context) {
			if b.useAccentColor {
				backgroundColor = draw.Color2(cm, draw.ColorTypeAccent, 0.875, 0.5)
			} else {
				backgroundColor = draw.Color2(cm, draw.ColorTypeBase, 0.95, 0.25)
			}
		} else if b.canPress(context) {
			backgroundColor = draw.Color2(cm, draw.ColorTypeBase, 0.975, 0.275)
		}
	}

	r := b.radius(context)
	border := !b.borderInvisible
	if context.IsEnabled(b) && (b.isHovered(context) || b.keepPressed) {
		border = true
	}
	bounds := context.Bounds(b)
	if border || b.isPressed(context) {
		draw.DrawRoundedRectWithSharpenCorners(context, dst, bounds, backgroundColor, r, b.sharpenCorners)
	}

	if border {
		borderType := draw.RoundedRectBorderTypeOutset
		if b.isPressed(context) {
			borderType = draw.RoundedRectBorderTypeInset
		}
		clr1, clr2 := draw.BorderColors(context.ColorMode(), borderType, b.useAccentColor && b.isPressed(context) && context.IsEnabled(b))
		draw.DrawRoundedRectBorderWithSharpenCorners(context, dst, bounds, clr1, clr2, r, float32(1*context.Scale()), borderType, b.sharpenCorners)
	}
}

func (b *baseButton) canPress(context *guigui.Context) bool {
	return context.IsEnabled(b) && b.isHovered(context) && !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && !b.keepPressed
}

func (b *baseButton) isHovered(context *guigui.Context) bool {
	return context.IsWidgetHitAtCursor(b)
}

func (b *baseButton) isActive(context *guigui.Context) bool {
	return context.IsEnabled(b) && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && b.isHovered(context) && (b.pressed || b.pairedButton != nil && b.pairedButton.pressed)
}

func (b *baseButton) isPressed(context *guigui.Context) bool {
	return context.IsEnabled(b) && b.isActive(context) || b.keepPressed
}

func (b *baseButton) setKeepPressed(keep bool) {
	if b.keepPressed == keep {
		return
	}
	b.keepPressed = keep
	guigui.RequestRedraw(b)
}

func defaultButtonSize(context *guigui.Context) image.Point {
	return image.Pt(6*UnitSize(context), UnitSize(context))
}

func (b *baseButton) DefaultSize(context *guigui.Context) image.Point {
	return defaultButtonSize(context)
}

func (b *baseButton) setSharpenCorners(sharpenCorners draw.SharpenCorners) {
	if b.sharpenCorners == sharpenCorners {
		return
	}
	b.sharpenCorners = sharpenCorners
	guigui.RequestRedraw(b)
}

func (b *baseButton) setUseAccentColor(use bool) {
	if b.useAccentColor == use {
		return
	}
	b.useAccentColor = use
	guigui.RequestRedraw(b)
}
