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

type DrawOptions struct {
	Options

	TextColor color.Color

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

func Draw(bounds image.Rectangle, dst *ebiten.Image, str string, options *DrawOptions) {
	if options.DrawSelection {
		for pos, line := range lines(str) {
			end := pos + len(line) - tailingLineBreakLen(line)
			if options.SelectionStart > end {
				continue
			}
			start := pos
			if options.SelectionEnd < start {
				break
			}
			start = max(start, options.SelectionStart)
			end = min(end, options.SelectionEnd)
			pos0, ok0 := TextPositionFromIndex(bounds.Dx(), str, start, &options.Options)
			pos1, ok1 := TextPositionFromIndex(bounds.Dx(), str, end, &options.Options)
			if ok0 && ok1 {
				x := float32(pos0.X) + float32(bounds.Min.X)
				y := float32(pos0.Top) + float32(bounds.Min.Y)
				width := float32(pos1.X - pos0.X)
				height := float32(pos0.Top - pos0.Bottom)
				vector.DrawFilledRect(dst, x, y, width, height, options.SelectionColor, false)
			}
		}
	}

	if options.DrawComposition {
		// TODO: Support multiple lines.
		{
			pos0, ok0 := TextPositionFromIndex(bounds.Dx(), str, options.CompositionStart, &options.Options)
			pos1, ok1 := TextPositionFromIndex(bounds.Dx(), str, options.CompositionEnd, &options.Options)
			if ok0 && ok1 {
				x := float32(pos0.X) + float32(bounds.Min.X)
				y := float32(pos0.Bottom) + float32(bounds.Min.Y) - options.CompositionBorderWidth
				w := float32(pos1.X - pos0.X)
				h := options.CompositionBorderWidth
				vector.DrawFilledRect(dst, x, y, w, h, options.InactiveCompositionColor, false)
			}
		}
		{
			pos0, ok0 := TextPositionFromIndex(bounds.Dx(), str, options.CompositionActiveStart, &options.Options)
			pos1, ok1 := TextPositionFromIndex(bounds.Dx(), str, options.CompositionActiveEnd, &options.Options)
			if ok0 && ok1 {
				x := float32(pos0.X) + float32(bounds.Min.X)
				y := float32(pos0.Bottom) + float32(bounds.Min.Y) - options.CompositionBorderWidth
				w := float32(pos1.X - pos0.X)
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
