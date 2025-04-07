// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/hajimehoshi/guigui"
)

type Button struct {
	guigui.DefaultWidget

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

func (b *Button) HandlePointingInput(context *guigui.Context) guigui.HandleInputResult {
	if guigui.IsEnabled(b) && b.isHovered() {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			if b.onDown != nil {
				b.onDown()
			}
			return guigui.HandleInputByWidget(b)
		}
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			if b.onUp != nil {
				b.onUp()
			}
			return guigui.HandleInputByWidget(b)
		}
	}
	return guigui.HandleInputResult{}
}

func (b *Button) Update(context *guigui.Context) error {
	hovered := b.isHovered()
	if b.prevHovered != hovered {
		b.prevHovered = hovered
		guigui.RequestRedraw(b)
	}
	return nil
}

func (b *Button) CursorShape(context *guigui.Context) (ebiten.CursorShapeType, bool) {
	if guigui.IsEnabled(b) && b.isHovered() {
		return ebiten.CursorShapePointer, true
	}
	return 0, true
}

func (b *Button) Draw(context *guigui.Context, dst *ebiten.Image) {
	// TODO: In the dark theme, the color should be different.
	// At least, shadow should be darker.
	// See macOS's buttons.
	cm := context.ColorMode()
	backgroundColor := Color2(cm, ColorTypeBase, 1, 0.3)
	borderColor := Color2(cm, ColorTypeBase, 0.7, 0)
	if b.isActive() {
		backgroundColor = Color2(cm, ColorTypeBase, 0.95, 0.25)
		borderColor = Color2(cm, ColorTypeBase, 0.7, 0)
	} else if guigui.IsEnabled(b) && b.isHovered() {
		backgroundColor = Color2(cm, ColorTypeBase, 0.975, 0.275)
		borderColor = Color2(cm, ColorTypeBase, 0.7, 0)
	} else if !guigui.IsEnabled(b) {
		backgroundColor = Color2(cm, ColorTypeBase, 0.95, 0.25)
		borderColor = Color2(cm, ColorTypeBase, 0.8, 0.1)
	}

	bounds := guigui.Bounds(b)
	r := min(RoundedCornerRadius(context), bounds.Dx()/4, bounds.Dy()/4)
	border := !b.borderInvisible
	if guigui.IsEnabled(b) && b.isHovered() {
		border = true
	}
	if border || b.isActive() {
		bounds := bounds.Inset(int(1 * context.Scale()))
		DrawRoundedRect(context, dst, bounds, backgroundColor, r)
	}

	if border {
		borderType := RoundedRectBorderTypeOutset
		if b.isActive() {
			borderType = RoundedRectBorderTypeInset
		} else if !guigui.IsEnabled(b) {
			borderType = RoundedRectBorderTypeRegular
		}
		DrawRoundedRectBorder(context, dst, bounds, borderColor, r, float32(1*context.Scale()), borderType)
	}
}

func (b *Button) isHovered() bool {
	return guigui.IsWidgetHitAt(b, image.Pt(ebiten.CursorPosition()))
}

func (b *Button) isActive() bool {
	return guigui.IsEnabled(b) && b.isHovered() && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

func defaultButtonSize(context *guigui.Context) (int, int) {
	return 6 * UnitSize(context), UnitSize(context)
}

func (b *Button) SetSize(context *guigui.Context, width, height int) {
	dw, dh := defaultButtonSize(context)
	b.widthMinusDefault = width - dw
	b.heightMinusDefault = height - dh
}

func (b *Button) Size(context *guigui.Context) (int, int) {
	dw, dh := defaultButtonSize(context)
	return b.widthMinusDefault + dw, b.heightMinusDefault + dh
}
