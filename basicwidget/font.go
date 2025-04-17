// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"
	"image/color"
	"iter"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/rivo/uniseg"
)

type HorizontalAlign int

const (
	HorizontalAlignStart HorizontalAlign = iota
	HorizontalAlignCenter
	HorizontalAlignEnd
)

type VerticalAlign int

const (
	VerticalAlignTop VerticalAlign = iota
	VerticalAlignMiddle
	VerticalAlignBottom
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

func autoWrapText(width int, txt string, face text.Face) string {
	var lines []string
	var line string
	var word string
	state := -1
	for len(txt) > 0 {
		cluster, nextText, boundaries, nextState := uniseg.StepString(txt, state)
		switch m := boundaries & uniseg.MaskLine; m {
		default:
			word += cluster
		case uniseg.LineCanBreak, uniseg.LineMustBreak:
			if line == "" {
				line += word + cluster
			} else {
				l := line + word + cluster
				if text.Advance(l, face) > float64(width) {
					lines = append(lines, line)
					line = word + cluster
				} else {
					line += word + cluster
				}
			}
			word = ""
			if m == uniseg.LineMustBreak {
				lines = append(lines, line)
				line = ""
			}
		}
		state = nextState
		txt = nextText
	}

	line += word
	if len(line) > 0 {
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func textUpperLeft(bounds image.Rectangle, str string, face text.Face, lineHeight float64, hAlign HorizontalAlign, vAlign VerticalAlign) (float64, float64) {
	w, h := text.Measure(str, face, lineHeight)
	x := float64(bounds.Min.X)
	y := float64(bounds.Min.Y)

	switch hAlign {
	case HorizontalAlignStart:
	case HorizontalAlignCenter:
		x += (float64(bounds.Dx()) - w) / 2
	case HorizontalAlignEnd:
		x += float64(bounds.Dx()) - w
	}

	m := face.Metrics()
	padding := (lineHeight - (m.HAscent + m.HDescent)) / 2

	switch vAlign {
	case VerticalAlignTop:
		y += padding
	case VerticalAlignMiddle:
		y = (float64(bounds.Dy()) - h) / 2
	case VerticalAlignBottom:
		y = float64(bounds.Dy()) - h - padding
	}

	return x, y
}

func textIndexFromPosition(textBounds image.Rectangle, position image.Point, str string, face text.Face, lineHeight float64, hAlign HorizontalAlign, vAlign VerticalAlign) int {
	lines := strings.Split(str, "\n")
	if len(lines) == 0 {
		return 0
	}

	// Determine the line first.
	m := face.Metrics()
	gap := lineHeight - m.HAscent - m.HDescent
	top := float64(textBounds.Min.Y)
	n := int((float64(position.Y) - top + gap/2) / lineHeight)
	if n < 0 {
		n = 0
	}
	if n >= len(lines) {
		n = len(lines) - 1
	}

	var idx int
	for _, l := range lines[:n] {
		idx += len(l) + 1 // 1 is for a new-line character.
	}

	// Deterine the line index.
	line := lines[n]
	left, _ := textUpperLeft(textBounds, line, face, lineHeight, hAlign, vAlign)
	var prevA float64
	var found bool
	for _, c := range visibleCulsters(line, face) {
		a := text.Advance(line[:c.EndIndexInBytes], face)
		if (float64(position.X) - left) < (prevA + (a-prevA)/2) {
			idx += c.StartIndexInBytes
			found = true
			break
		}
		prevA = a
	}
	if !found {
		idx += len(line)
	}

	return idx
}

func textPosition(textBounds image.Rectangle, str string, index int, face text.Face, lineHeight float64, hAlign HorizontalAlign, vAlign VerticalAlign) (x, top, bottom float64, ok bool) {
	if index < 0 || index > len(str) {
		return 0, 0, 0, false
	}

	y := float64(textBounds.Min.Y)

	lines := strings.Split(str, "\n")
	var line string
	for _, l := range lines {
		// +1 is for \n.
		if index < len(l)+1 {
			line = l
			break
		}
		index -= len(l) + 1 // 1 is for a new-line character.
		y += lineHeight
	}

	x, _ = textUpperLeft(textBounds, line, face, lineHeight, hAlign, vAlign)
	x += text.Advance(line[:index], face)

	m := face.Metrics()
	paddingY := (lineHeight - (m.HAscent + m.HDescent)) / 2
	return x, y + paddingY, y + lineHeight - paddingY, true
}

func graphemes(str string) iter.Seq[string] {
	return func(yield func(s string) bool) {
		state := -1
		for len(str) > 0 {
			var cluster string
			cluster, str, _, state = uniseg.StepString(str, state)
			if !yield(cluster) {
				return
			}
		}
	}
}

func visibleCulsters(str string, face text.Face) []text.Glyph {
	return text.AppendGlyphs(nil, str, face, nil)
}

func backspaceOnGraphemes(str string, position int) (string, int) {
	var pos int
	for c := range graphemes(str) {
		startPos := pos
		endPos := pos + len(c)
		if position > endPos {
			pos = endPos
			continue
		}
		return str[:startPos] + str[endPos:], startPos
	}
	return str, position
}

func deleteOnGraphemes(str string, position int) (string, int) {
	var pos int
	for c := range graphemes(str) {
		startPos := pos
		endPos := pos + len(c)
		if position > startPos {
			pos = endPos
			continue
		}
		return str[:startPos] + str[endPos:], startPos
	}
	return str, position
}

func prevPositionOnGraphemes(str string, position int) int {
	var pos int
	for c := range graphemes(str) {
		startPos := pos
		endPos := pos + len(c)
		if position > endPos {
			pos = endPos
			continue
		}
		return startPos
	}
	return position
}

func nextPositionOnGraphemes(str string, position int) int {
	var pos int
	for c := range graphemes(str) {
		startPos := pos
		endPos := pos + len(c)
		if position > startPos {
			pos = endPos
			continue
		}
		return endPos
	}
	return position
}
