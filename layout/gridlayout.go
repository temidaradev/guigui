// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package layout

import (
	"image"
	"iter"

	"github.com/hajimehoshi/guigui"
)

type Size struct {
	typ   sizeType
	value int
}

type sizeType int

const (
	sizeTypeDefault sizeType = iota
	sizeTypeFixed
	sizeTypeFraction
)

func DefaultSize() Size {
	return Size{
		typ:   sizeTypeDefault,
		value: 0,
	}
}

func FixedSize(value int) Size {
	return Size{
		typ:   sizeTypeFixed,
		value: value,
	}
}

func FractionSize(value int) Size {
	return Size{
		typ:   sizeTypeFraction,
		value: value,
	}
}

var (
	defaultWidths  = []Size{FractionSize(1)}
	defaultHeights = []Size{FractionSize(1)}
)

type GridLayout struct {
	Bounds    image.Rectangle
	Widths    []Size
	Heights   []Size
	ColumnGap int
	RowGap    int
}

func (g GridLayout) CellBounds(context *guigui.Context, widgets []guigui.Widget) iter.Seq2[guigui.Widget, image.Rectangle] {
	return func(yield func(w guigui.Widget, bounds image.Rectangle) bool) {
		widths := g.Widths
		if len(widths) == 0 {
			widths = defaultWidths
		}
		heights := g.Heights
		if len(heights) == 0 {
			heights = defaultHeights
		}

		widthsInPixels := make([]int, len(widths))
		heightsInPixels := make([]int, len(heights))

		var denomW, denomH int
		restW, restH := g.Bounds.Dx(), g.Bounds.Dy()
		restW -= (len(widths) - 1) * g.ColumnGap
		restH -= (len(heights) - 1) * g.RowGap
		if restW < 0 {
			restW = 0
		}
		if restH < 0 {
			restH = 0
		}

		for i, width := range widths {
			switch width.typ {
			case sizeTypeDefault:
				widthsInPixels[i] = 0
				for j := range (len(widgets)-1)/len(widths) + 1 {
					if j*len(widths)+i >= len(widgets) {
						break
					}
					w, _ := widgets[j*len(widths)+i].DefaultSize(context)
					widthsInPixels[i] = max(widthsInPixels[i], w)
				}
			case sizeTypeFixed:
				widthsInPixels[i] = width.value
			case sizeTypeFraction:
				widthsInPixels[i] = 0
				denomW += width.value
			}
			restW -= widthsInPixels[i]
		}

		if denomW > 0 {
			origRestW := restW
			for i, width := range widths {
				if width.typ != sizeTypeFraction {
					continue
				}
				w := int(float64(origRestW) * float64(width.value) / float64(denomW))
				widthsInPixels[i] = w
				restW -= w
			}
			// TODO: Use a better algorithm to distribute the rest.
			for restW > 0 {
				for i := len(widthsInPixels) - 1; i >= 0; i-- {
					if widths[i].typ != sizeTypeFraction {
						continue
					}
					widthsInPixels[i]++
					restW--
					if restW <= 0 {
						break
					}
				}
				if restW <= 0 {
					break
				}
			}
		}

		for j, height := range heights {
			switch height.typ {
			case sizeTypeDefault:
				heightsInPixels[j] = 0
				for i := range widths {
					if j*len(widths)+i >= len(widgets) {
						break
					}
					_, h := widgets[j*len(widths)+i].DefaultSize(context)
					heightsInPixels[j] = max(heightsInPixels[j], h)
				}
			case sizeTypeFixed:
				heightsInPixels[j] = height.value
			case sizeTypeFraction:
				heightsInPixels[j] = 0
				denomH += height.value
			}
			restH -= heightsInPixels[j]
		}

		if denomH > 0 {
			origRestH := restH
			for j, height := range heights {
				if height.typ != sizeTypeFraction {
					continue
				}
				h := int(float64(origRestH) * float64(height.value) / float64(denomH))
				heightsInPixels[j] = h
				restH -= h
			}
			for restH > 0 {
				for j := len(heightsInPixels) - 1; j >= 0; j-- {
					if heights[j].typ != sizeTypeFraction {
						continue
					}
					heightsInPixels[j]++
					restH--
					if restH <= 0 {
						break
					}
				}
				if restH <= 0 {
					break
				}
			}
		}

		y := g.Bounds.Min.Y
		var widgetIdx int
		for idx := 0; idx < len(widgets); idx += len(widths) * len(heights) {
			for j := 0; j < len(heights); j++ {
				x := g.Bounds.Min.X
				for i := 0; i < len(widths); i++ {
					bounds := image.Rect(x, y, x+widthsInPixels[i], y+heightsInPixels[j])
					if !yield(widgets[widgetIdx], bounds) {
						return
					}
					x += widthsInPixels[i]
					x += g.ColumnGap
					widgetIdx++
					if widgetIdx >= len(widgets) {
						return
					}
				}
				y += heightsInPixels[j]
				y += g.RowGap
			}
		}
	}
}
