// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package textutil

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type DrawTextOptions struct {
	Face            text.Face
	LineHeight      float64
	HorizontalAlign HorizontalAlign
	VerticalAlign   VerticalAlign
	TextColor       color.Color

	DrawSelection  bool
	SelectionStart int
	SelectionEnd   int
	SelectionColor color.Color

	DrawComposition          bool
	CompositionStart         int
	CompositionEnd           int
	CompositionActiveStart   int
	CompositionActiveEnd     int
	InactiveCompositionColor color.Color
	ActiveCompositionColor   color.Color
	CompositionBorderWidth   float32
}

func DrawText(bounds image.Rectangle, dst *ebiten.Image, str string, options *DrawTextOptions) {
	if options.DrawSelection {
		var tailIndices []int
		for i, r := range str[options.SelectionStart:options.SelectionEnd] {
			if r != '\n' {
				continue
			}
			tailIndices = append(tailIndices, options.SelectionStart+i)
		}
		tailIndices = append(tailIndices, options.SelectionEnd)

		headIdx := options.SelectionStart
		for _, idx := range tailIndices {
			x0, top0, bottom0, ok0 := TextPosition(bounds.Dx(), str, headIdx, options.Face, options.LineHeight, options.HorizontalAlign, options.VerticalAlign)
			x1, _, _, ok1 := TextPosition(bounds.Dx(), str, idx, options.Face, options.LineHeight, options.HorizontalAlign, options.VerticalAlign)
			if ok0 && ok1 {
				x := float32(x0) + float32(bounds.Min.X)
				y := float32(top0) + float32(bounds.Min.Y)
				width := float32(x1 - x0)
				height := float32(bottom0 - top0)
				vector.DrawFilledRect(dst, x, y, width, height, options.SelectionColor, false)
			}
			headIdx = idx + 1
		}
	}

	if options.DrawComposition {
		// Assume that the composition is always in the same line.
		/*if strings.Contains(txt[uStart:uEnd], "\n") {
			slog.Error("composition text must not contain '\\n'")
		}*/
		{
			x0, _, bottom0, ok0 := TextPosition(bounds.Dx(), str, options.CompositionStart, options.Face, options.LineHeight, options.HorizontalAlign, options.VerticalAlign)
			x1, _, _, ok1 := TextPosition(bounds.Dx(), str, options.CompositionEnd, options.Face, options.LineHeight, options.HorizontalAlign, options.VerticalAlign)
			if ok0 && ok1 {
				x := float32(x0) + float32(bounds.Min.X)
				y := float32(bottom0) + float32(bounds.Min.Y) - options.CompositionBorderWidth
				w := float32(x1 - x0)
				h := options.CompositionBorderWidth
				vector.DrawFilledRect(dst, x, y, w, h, options.InactiveCompositionColor, false)
			}
		}
		{
			x0, _, bottom0, ok0 := TextPosition(bounds.Dx(), str, options.CompositionActiveStart, options.Face, options.LineHeight, options.HorizontalAlign, options.VerticalAlign)
			x1, _, _, ok1 := TextPosition(bounds.Dx(), str, options.CompositionActiveEnd, options.Face, options.LineHeight, options.HorizontalAlign, options.VerticalAlign)
			if ok0 && ok1 {
				x := float32(x0) + float32(bounds.Min.X)
				y := float32(bottom0) + float32(bounds.Min.Y) - options.CompositionBorderWidth
				w := float32(x1 - x0)
				h := options.CompositionBorderWidth
				vector.DrawFilledRect(dst, x, y, w, h, options.ActiveCompositionColor, false)
			}
		}
	}

	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(bounds.Min.X), float64(bounds.Min.Y))
	op.ColorScale.ScaleWithColor(options.TextColor)
	if dst.Bounds() != bounds {
		dst = dst.SubImage(bounds).(*ebiten.Image)
	}

	op.LineSpacing = options.LineHeight

	switch options.HorizontalAlign {
	case HorizontalAlignStart:
		op.PrimaryAlign = text.AlignStart
	case HorizontalAlignCenter:
		op.GeoM.Translate(float64(bounds.Dx())/2, 0)
		op.PrimaryAlign = text.AlignCenter
	case HorizontalAlignEnd:
		op.GeoM.Translate(float64(bounds.Dx()), 0)
		op.PrimaryAlign = text.AlignEnd
	}

	c := lineCount(str)
	if c == 0 {
		return
	}
	height := options.LineHeight * float64(c)

	m := options.Face.Metrics()
	padding := (options.LineHeight - (m.HAscent + m.HDescent)) / 2
	op.GeoM.Translate(0, padding)
	switch options.VerticalAlign {
	case VerticalAlignTop:
	case VerticalAlignMiddle:
		op.GeoM.Translate(0, (float64(bounds.Dy())-height)/2)
	case VerticalAlignBottom:
		op.GeoM.Translate(0, float64(bounds.Dy())-height)
	}

	for _, line := range lines(str) {
		text.Draw(dst, line, options.Face, op)
		op.GeoM.Translate(0, options.LineHeight)
	}
}
