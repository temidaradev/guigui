// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Guigui Authors

package basicwidget

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget/internal/draw"
)

type PanelStyle int

const (
	PanelStyleDefault PanelStyle = iota
	PanelStyleSide
)

type Panel struct {
	guigui.DefaultWidget

	content      guigui.Widget
	scollOverlay ScrollOverlay
	border       panelBorder
	style        PanelStyle
}

func (p *Panel) SetContent(widget guigui.Widget) {
	p.content = widget
}

func (p *Panel) SetStyle(typ PanelStyle) {
	if p.style == typ {
		return
	}
	p.style = typ
	guigui.RequestRedraw(p)
}

func (p *Panel) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	if p.content == nil {
		return nil
	}

	offsetX, offsetY := p.scollOverlay.Offset()
	appender.AppendChildWidgetWithPosition(p.content, context.Position(p).Add(image.Pt(int(offsetX), int(offsetY))))

	p.scollOverlay.SetContentSize(context, context.Size(p.content))
	appender.AppendChildWidgetWithBounds(&p.scollOverlay, context.Bounds(p))

	p.border.scrollOverlay = &p.scollOverlay
	appender.AppendChildWidgetWithBounds(&p.border, context.Bounds(p))

	return nil
}

func (p *Panel) Draw(context *guigui.Context, dst *ebiten.Image) {
	switch p.style {
	case PanelStyleSide:
		dst.Fill(draw.Color(context.ColorMode(), draw.ColorTypeBase, 0.875))
	}
}

type panelBorder struct {
	guigui.DefaultWidget

	scrollOverlay *ScrollOverlay
}

func (p *panelBorder) Draw(context *guigui.Context, dst *ebiten.Image) {
	// Render borders.
	strokeWidth := float32(1 * context.Scale())
	bounds := context.Bounds(p)
	x0 := float32(bounds.Min.X)
	x1 := float32(bounds.Max.X)
	y0 := float32(bounds.Min.Y)
	y1 := float32(bounds.Max.Y)
	offsetX, offsetY := p.scrollOverlay.Offset()
	r := p.scrollOverlay.scrollRange(context)
	clr := draw.Color(context.ColorMode(), draw.ColorTypeBase, 0.85)
	if offsetX < float64(r.Max.X) {
		vector.StrokeLine(dst, x0+strokeWidth/2, y0, x0+strokeWidth/2, y1, strokeWidth, clr, false)
	}
	if offsetY < float64(r.Max.Y) {
		vector.StrokeLine(dst, x0, y0+strokeWidth/2, x1, y0+strokeWidth/2, strokeWidth, clr, false)
	}
	if offsetX > float64(r.Min.X) {
		vector.StrokeLine(dst, x1-strokeWidth/2, y0, x1-strokeWidth/2, y1, strokeWidth, clr, false)
	}
	if offsetY > float64(r.Min.Y) {
		vector.StrokeLine(dst, x0, y1-strokeWidth/2, x1, y1-strokeWidth/2, strokeWidth, clr, false)
	}
}
