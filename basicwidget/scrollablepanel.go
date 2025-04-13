// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget/internal/draw"
)

type ScrollablePanel struct {
	guigui.DefaultWidget

	content      guigui.Widget
	scollOverlay ScrollOverlay
	border       scrollablePanelBorder

	widthMinusDefault  int
	heightMinusDefault int
}

func (s *ScrollablePanel) SetContent(widget guigui.Widget) {
	s.content = widget
}

func (s *ScrollablePanel) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	if s.content == nil {
		return nil
	}

	p := guigui.Position(s)
	offsetX, offsetY := s.scollOverlay.Offset()
	p = p.Add(image.Pt(int(offsetX), int(offsetY)))
	guigui.SetPosition(s.content, p)
	appender.AppendChildWidget(s.content)

	s.scollOverlay.SetContentSize(guigui.Size(s.content))
	guigui.SetPosition(&s.scollOverlay, guigui.Position(s))
	appender.AppendChildWidget(&s.scollOverlay)

	s.border.scrollOverlay = &s.scollOverlay
	guigui.SetPosition(&s.border, guigui.Position(s))
	appender.AppendChildWidget(&s.border)

	return nil
}

func defaultScrollablePanelSize(context *guigui.Context) (int, int) {
	return 6 * UnitSize(context), 6 * UnitSize(context)
}

func (s *ScrollablePanel) DefaultSize(context *guigui.Context) (int, int) {
	dw, dh := defaultScrollablePanelSize(context)
	return s.widthMinusDefault + dw, s.heightMinusDefault + dh
}

func (s *ScrollablePanel) SetSize(context *guigui.Context, width, height int) {
	dw, dh := defaultScrollablePanelSize(context)
	s.widthMinusDefault = width - dw
	s.heightMinusDefault = height - dh
}

type scrollablePanelBorder struct {
	guigui.DefaultWidget

	scrollOverlay *ScrollOverlay
}

func (s *scrollablePanelBorder) Draw(context *guigui.Context, dst *ebiten.Image) {
	// Render borders.
	strokeWidth := float32(1 * context.Scale())
	bounds := guigui.Bounds(s)
	x0 := float32(bounds.Min.X)
	x1 := float32(bounds.Max.X)
	y0 := float32(bounds.Min.Y)
	y1 := float32(bounds.Max.Y)
	offsetX, offsetY := s.scrollOverlay.Offset()
	if offsetX < 0 {
		vector.StrokeLine(dst, x0+strokeWidth/2, y0, x0+strokeWidth/2, y1, strokeWidth, draw.Color(context.ColorMode(), draw.ColorTypeBase, 0.85), false)
	}
	if offsetY < 0 {
		vector.StrokeLine(dst, x0, y0+strokeWidth/2, x1, y0+strokeWidth/2, strokeWidth, draw.Color(context.ColorMode(), draw.ColorTypeBase, 0.85), false)
	}
}
