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
}

func (s *ScrollablePanel) SetContent(widget guigui.Widget) {
	s.content = widget
}

func (s *ScrollablePanel) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	if s.content == nil {
		return nil
	}

	p := context.Position(s)
	offsetX, offsetY := s.scollOverlay.Offset()
	p = p.Add(image.Pt(int(offsetX), int(offsetY)))
	context.SetPosition(s.content, p)
	w, h := context.Size(s)
	appender.AppendChildWidget(s.content)

	cw, ch := context.Size(s.content)
	s.scollOverlay.SetContentSize(context, cw, ch)
	context.SetPosition(&s.scollOverlay, context.Position(s))
	context.SetSize(&s.scollOverlay, w, h)
	appender.AppendChildWidget(&s.scollOverlay)

	s.border.scrollOverlay = &s.scollOverlay
	context.SetPosition(&s.border, context.Position(s))
	appender.AppendChildWidget(&s.border)

	return nil
}

type scrollablePanelBorder struct {
	guigui.DefaultWidget

	scrollOverlay *ScrollOverlay
}

func (s *scrollablePanelBorder) Draw(context *guigui.Context, dst *ebiten.Image) {
	// Render borders.
	strokeWidth := float32(1 * context.Scale())
	bounds := context.Bounds(s)
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
