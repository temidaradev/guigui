// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/hajimehoshi/guigui"
)

type hoverOverlay struct {
	guigui.DefaultWidget

	hovering      bool
	pressingLeft  bool
	pressingRight bool

	onDown func(mouseButton ebiten.MouseButton, cursorPosition image.Point)
	onUp   func(mouseButton ebiten.MouseButton, cursorPosition image.Point)
}

func (m *hoverOverlay) SetOnDown(f func(mouseButton ebiten.MouseButton, cursorPosition image.Point)) {
	m.onDown = f
}

func (m *hoverOverlay) SetOnUp(f func(mouseButton ebiten.MouseButton, cursorPosition image.Point)) {
	m.onUp = f
}

func (m *hoverOverlay) HandleInput(context *guigui.Context) guigui.HandleInputResult {
	x, y := ebiten.CursorPosition()
	m.setHovering(image.Pt(x, y).In(guigui.VisibleBounds(m)) && guigui.IsVisible(m))

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		if !image.Pt(ebiten.CursorPosition()).In(guigui.VisibleBounds(m)) {
			return guigui.HandleInputResult{}
		}
		if guigui.IsEnabled(m) {
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
				m.setPressing(true, ebiten.MouseButtonLeft)
			}
			if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
				m.setPressing(true, ebiten.MouseButtonRight)
			}
		}
		guigui.Focus(m)
		return guigui.HandleInputByWidget(m)
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && m.pressingLeft ||
		inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) && m.pressingRight {
		if m.pressingLeft {
			m.setPressing(false, ebiten.MouseButtonLeft)
		}
		if m.pressingRight {
			m.setPressing(false, ebiten.MouseButtonRight)
		}
		if !image.Pt(ebiten.CursorPosition()).In(guigui.VisibleBounds(m)) {
			return guigui.HandleInputResult{}
		}
		if guigui.IsEnabled(m) {
			return guigui.HandleInputByWidget(m)
		}
	}

	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		m.setPressing(false, ebiten.MouseButtonLeft)
	}
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		m.setPressing(false, ebiten.MouseButtonRight)
	}

	return guigui.HandleInputResult{}
}

func (m *hoverOverlay) Update(context *guigui.Context) error {
	if !guigui.IsVisible(m) {
		m.setHovering(false)
	}
	return nil
}

func (m *hoverOverlay) setPressing(pressing bool, mouseButton ebiten.MouseButton) {
	switch mouseButton {
	case ebiten.MouseButtonLeft:
		if m.pressingLeft == pressing {
			return
		}
	case ebiten.MouseButtonRight:
		if m.pressingRight == pressing {
			return
		}
	}

	if guigui.IsEnabled(m) {
		if p := image.Pt(ebiten.CursorPosition()); p.In(guigui.VisibleBounds(guigui.Parent(m))) {
			if pressing {
				if m.onDown != nil {
					m.onDown(mouseButton, p)
				}
			} else {
				if m.onUp != nil {
					m.onUp(mouseButton, p)
				}
			}
		}
	}

	switch mouseButton {
	case ebiten.MouseButtonLeft:
		m.pressingLeft = pressing
	case ebiten.MouseButtonRight:
		m.pressingRight = pressing
	}
	guigui.RequestRedraw(m)
}

func (m *hoverOverlay) setHovering(hovering bool) {
	if m.hovering == hovering {
		return
	}
	m.hovering = hovering
	guigui.RequestRedraw(m)
}

func (m *hoverOverlay) IsPressing() bool {
	return m.pressingLeft || m.pressingRight
}

func (m *hoverOverlay) IsHovering() bool {
	return m.hovering
}
