// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package textutil

import (
	"image"
	"iter"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

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

func visibleCulsters(str string, face text.Face) []text.Glyph {
	return text.AppendGlyphs(nil, str, face, nil)
}

func lines(str string) iter.Seq[string] {
	return func(yield func(s string) bool) {
		var line string
		state := -1
		for len(str) > 0 {
			segment, nextStr, mustBreak, nextState := uniseg.FirstLineSegmentInString(str, state)
			line += segment
			if mustBreak {
				if !yield(line) {
					return
				}
				line = ""
			}
			state = nextState
			str = nextStr
		}
		if len(line) > 0 {
			if !yield(line) {
				return
			}
		}
	}
}

func removeSpaceAtLineTail(str string) string {
	return strings.TrimRightFunc(str, unicode.IsSpace)
}

func AutoWrapText(width int, str string, face text.Face) string {
	var lines []string
	var line string
	var word string
	state := -1
	for len(str) > 0 {
		cluster, nextStr, boundaries, nextState := uniseg.StepString(str, state)
		switch m := boundaries & uniseg.MaskLine; m {
		default:
			word += cluster
		case uniseg.LineCanBreak, uniseg.LineMustBreak:
			if line == "" {
				line += word + cluster
			} else {
				if text.Advance(removeSpaceAtLineTail(line+word+cluster), face) > float64(width) {
					lines = append(lines, line)
					line = word + cluster
				} else {
					line += word + cluster
				}
			}
			word = ""
			if m == uniseg.LineMustBreak {
				lines = append(lines, line[:len(line)-len(cluster)])
				line = ""
			}
		}
		state = nextState
		str = nextStr
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

func TextIndexFromPosition(textBounds image.Rectangle, position image.Point, str string, face text.Face, lineHeight float64, hAlign HorizontalAlign, vAlign VerticalAlign) int {
	lines := slices.Collect(lines(str))
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
		idx += len(l)
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
		if uniseg.HasTrailingLineBreakInString(line) {
			_, s := utf8.DecodeLastRuneInString(line)
			idx -= s
		}
	}

	return idx
}

func TextPosition(textBounds image.Rectangle, str string, index int, face text.Face, lineHeight float64, hAlign HorizontalAlign, vAlign VerticalAlign) (x, top, bottom float64, ok bool) {
	if index < 0 || index > len(str) {
		return 0, 0, 0, false
	}

	y := float64(textBounds.Min.Y)

	var line string
	var found bool
	for l := range lines(str) {
		line = l
		if index < len(l) {
			found = true
			break
		}
		index -= len(l)
		y += lineHeight
	}
	// When found is false, the position is in the tail of the last line.
	if !found && len(str) > 0 && !uniseg.HasTrailingLineBreakInString(str) {
		index = len(line)
		y -= lineHeight
	}

	x, _ = textUpperLeft(textBounds, line, face, lineHeight, hAlign, vAlign)
	x += text.Advance(line[:index], face)

	m := face.Metrics()
	paddingY := (lineHeight - (m.HAscent + m.HDescent)) / 2
	return x, y + paddingY, y + lineHeight - paddingY, true
}
