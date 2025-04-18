// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/hajimehoshi/guigui/basicwidget/internal/textutil"
)

type HorizontalAlign int

const (
	HorizontalAlignStart  HorizontalAlign = HorizontalAlign(textutil.HorizontalAlignStart)
	HorizontalAlignCenter HorizontalAlign = HorizontalAlign(textutil.HorizontalAlignCenter)
	HorizontalAlignEnd    HorizontalAlign = HorizontalAlign(textutil.HorizontalAlignEnd)
)

type VerticalAlign int

const (
	VerticalAlignTop    VerticalAlign = VerticalAlign(textutil.VerticalAlignTop)
	VerticalAlignMiddle VerticalAlign = VerticalAlign(textutil.VerticalAlignMiddle)
	VerticalAlignBottom VerticalAlign = VerticalAlign(textutil.VerticalAlignBottom)
)

func drawText(bounds image.Rectangle, dst *ebiten.Image, str string, face text.Face, lineHeight float64, hAlign HorizontalAlign, vAlign VerticalAlign, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(bounds.Min.X), float64(bounds.Min.Y))
	op.ColorScale.ScaleWithColor(clr)
	if dst.Bounds() != bounds {
		dst = dst.SubImage(bounds).(*ebiten.Image)
	}

	op.LineSpacing = lineHeight

	switch hAlign {
	case HorizontalAlignStart:
		op.PrimaryAlign = text.AlignStart
	case HorizontalAlignCenter:
		op.GeoM.Translate(float64(bounds.Dx())/2, 0)
		op.PrimaryAlign = text.AlignCenter
	case HorizontalAlignEnd:
		op.GeoM.Translate(float64(bounds.Dx()), 0)
		op.PrimaryAlign = text.AlignEnd
	}

	m := face.Metrics()
	padding := (lineHeight - (m.HAscent + m.HDescent)) / 2

	switch vAlign {
	case VerticalAlignTop:
		op.GeoM.Translate(0, padding)
		op.SecondaryAlign = text.AlignStart
	case VerticalAlignMiddle:
		op.GeoM.Translate(0, float64(bounds.Dy())/2)
		op.SecondaryAlign = text.AlignCenter
	case VerticalAlignBottom:
		op.GeoM.Translate(0, float64(bounds.Dy())-padding)
		op.SecondaryAlign = text.AlignEnd
	}

	text.Draw(dst, str, face, op)
}
