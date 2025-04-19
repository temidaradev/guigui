// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package textutil

import (
	"image"
	"iter"
	"strings"
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
				l := line + word + cluster
				l = l[:len(l)-tailingLineBreakLen(l)]
				// TODO: Consider a line alignment and/or editable/selectable states when calculating the width.
				if text.Advance(l, face) > float64(width) {
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

func oneLineLeft(bounds image.Rectangle, line string, face text.Face, hAlign HorizontalAlign) float64 {
	w := text.Advance(line, face)
	x := float64(bounds.Min.X)
	switch hAlign {
	case HorizontalAlignStart:
	case HorizontalAlignCenter:
		x += (float64(bounds.Dx()) - w) / 2
	case HorizontalAlignEnd:
		x += float64(bounds.Dx()) - w
	}
	return x
}

func TextIndexFromPosition(textBounds image.Rectangle, position image.Point, str string, face text.Face, lineHeight float64, hAlign HorizontalAlign, vAlign VerticalAlign) int {
	// Determine the line first.
	m := face.Metrics()
	gap := lineHeight - m.HAscent - m.HDescent
	top := float64(textBounds.Min.Y)
	n := int((float64(position.Y) - top + gap/2) / lineHeight)

	var pos int
	var line string
	var lineFound bool
	var lineIndex int
	for l := range lines(str) {
		line = l
		if lineIndex >= n {
			lineFound = true
			break
		}
		pos += len(l)
		lineIndex++
	}
	// Use the last line.
	if !lineFound {
		pos -= len(line)
	}

	// Deterine the line index.
	left := oneLineLeft(textBounds, line, face, hAlign)
	var prevA float64
	var clusterFound bool
	for _, c := range visibleCulsters(line, face) {
		a := text.Advance(line[:c.EndIndexInBytes], face)
		if (float64(position.X) - left) < (prevA + (a-prevA)/2) {
			pos += c.StartIndexInBytes
			clusterFound = true
			break
		}
		prevA = a
	}
	if !clusterFound {
		pos += len(line)
		pos -= tailingLineBreakLen(line)
	}

	return pos
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

	x = oneLineLeft(textBounds, line, face, hAlign)
	x += text.Advance(line[:index], face)

	m := face.Metrics()
	paddingY := (lineHeight - (m.HAscent + m.HDescent)) / 2
	return x, y + paddingY, y + lineHeight - paddingY, true
}

func tailingLineBreakLen(str string) int {
	if !uniseg.HasTrailingLineBreakInString(str) {
		return 0
	}

	// https://en.wikipedia.org/wiki/Newline#Unicode
	if strings.HasSuffix(str, "\r\n") {
		return 2
	}

	_, s := utf8.DecodeLastRuneInString(str)
	return s
}
