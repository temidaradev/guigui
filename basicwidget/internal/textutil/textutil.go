// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package textutil

import (
	"fmt"
	"image"
	"iter"
	"strings"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/rivo/uniseg"
)

type Options struct {
	AutoWrap        bool
	Face            text.Face
	LineHeight      float64
	HorizontalAlign HorizontalAlign
	VerticalAlign   VerticalAlign
}

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

func lines(width int, str string, autoWrap bool, face text.Face) iter.Seq2[int, string] {
	return func(yield func(pos int, s string) bool) {
		origStr := str
		// Use []byte instead of string for efficient concatenations.
		var line []byte
		var pos int
		state := -1
		for len(str) > 0 {
			segment, nextStr, mustBreak, nextState := uniseg.FirstLineSegmentInString(str, state)
			if len(line) > 0 && autoWrap {
				l := string(line) + string(segment)
				// TODO: Consider a line alignment and/or editable/selectable states when calculating the width.
				if text.Advance(l[:len(l)-tailingLineBreakLen(l)], face) > float64(width) {
					if !yield(pos, string(line)) {
						return
					}
					pos += len(line)
					line = line[:0]
				}
			}
			line = append(line, segment...)
			if mustBreak {
				if !yield(pos, string(line)) {
					return
				}
				pos += len(line)
				line = line[:0]
			}
			str = nextStr
			state = nextState
		}

		if len(line) > 0 {
			if !yield(pos, string(line)) {
				return
			}
			pos += len(line)
		}

		// If the string ends with a line break, or an empty line, add an extra empty line.
		if tailingLineBreakLen(origStr) > 0 || origStr == "" {
			if !yield(pos, "") {
				return
			}
		}
	}
}

func oneLineLeft(width int, line string, face text.Face, hAlign HorizontalAlign) float64 {
	w := text.Advance(line[:len(line)-tailingLineBreakLen(line)], face)
	switch hAlign {
	case HorizontalAlignStart:
		return 0
	case HorizontalAlignCenter:
		return (float64(width) - w) / 2
	case HorizontalAlignEnd:
		return float64(width) - w
	default:
		panic(fmt.Sprintf("textutil: invalid HorizontalAlign: %d", hAlign))
	}
}

func TextIndexFromPosition(width int, position image.Point, str string, options *Options) int {
	// Determine the line first.
	padding := textPadding(options.Face, options.LineHeight)
	n := int((float64(position.Y) + padding) / options.LineHeight)

	var pos int
	var line string
	var lineIndex int
	for p, l := range lines(width, str, options.AutoWrap, options.Face) {
		line = l
		pos = p
		if lineIndex >= n {
			break
		}
		lineIndex++
	}

	// Deterine the line index.
	left := oneLineLeft(width, line, options.Face, options.HorizontalAlign)
	var prevA float64
	var clusterFound bool
	for _, c := range visibleCulsters(line, options.Face) {
		a := text.Advance(line[:c.EndIndexInBytes], options.Face)
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

type TextPosition struct {
	X      float64
	Top    float64
	Bottom float64
}

func TextPositionFromIndex(width int, str string, index int, options *Options) (position0, position1 TextPosition, count int) {
	if index < 0 || index > len(str) {
		return TextPosition{}, TextPosition{}, 0
	}

	var y, y0, y1 float64
	var indexInLine0, indexInLine1 int
	var line0, line1 string
	var found0, found1 bool
	for p, l := range lines(width, str, options.AutoWrap, options.Face) {
		// When auto wrap is on, there can be two positions:
		// one in the tail of the previous line and one in the head of the next line.
		if tailingLineBreakLen(l) == 0 && index == p+len(l) {
			found0 = true
			line0 = l
			indexInLine0 = index - p
			y0 = y
		} else if p <= index && index < p+len(l) {
			found1 = true
			line1 = l
			indexInLine1 = index - p
			y1 = y
			break
		}
		y += options.LineHeight
	}

	if !found0 && !found1 {
		return TextPosition{}, TextPosition{}, 0
	}

	paddingY := textPadding(options.Face, options.LineHeight)

	var pos0, pos1 TextPosition
	if found0 {
		x0 := oneLineLeft(width, line0, options.Face, options.HorizontalAlign)
		x0 += text.Advance(line0[:indexInLine0], options.Face)
		pos0 = TextPosition{
			X:      x0,
			Top:    y0 + paddingY,
			Bottom: y0 + options.LineHeight + paddingY,
		}
	}
	if found1 {
		x1 := oneLineLeft(width, line1, options.Face, options.HorizontalAlign)
		x1 += text.Advance(line1[:indexInLine1], options.Face)
		pos1 = TextPosition{
			X:      x1,
			Top:    y1 + paddingY,
			Bottom: y1 + options.LineHeight - paddingY,
		}
	}
	if found0 && !found1 {
		return pos0, TextPosition{}, 1
	}
	if found1 && !found0 {
		return pos1, TextPosition{}, 1
	}
	return pos0, pos1, 2
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

func trimTailingLineBreak(str string) string {
	for {
		c := tailingLineBreakLen(str)
		if c == 0 {
			break
		}
		str = str[:len(str)-c]
	}
	return str
}

func lineCount(width int, str string, autoWrap bool, face text.Face) int {
	var count int
	for range lines(width, str, autoWrap, face) {
		count++
	}
	return count
}

func Measure(width int, str string, autoWrap bool, face text.Face, lineHeight float64) (float64, float64) {
	var maxWidth, height float64
	for _, line := range lines(width, str, autoWrap, face) {
		line = trimTailingLineBreak(line)
		maxWidth = max(maxWidth, text.Advance(line, face))
		// The text is already shifted by (lineHeight - (m.HAscent + m.Descent)) / 2.
		// Thus, just counting the line number is enough.
		height += lineHeight
	}
	return maxWidth, height
}

func textPadding(face text.Face, lineHeight float64) float64 {
	m := face.Metrics()
	padding := (lineHeight - (m.HAscent + m.HDescent)) / 2
	return padding
}

func TextPositionYOffset(size image.Point, str string, options *Options) float64 {
	c := lineCount(size.X, str, options.AutoWrap, options.Face)
	textHeight := options.LineHeight * float64(c)
	yOffset := textPadding(options.Face, options.LineHeight)
	switch options.VerticalAlign {
	case VerticalAlignTop:
	case VerticalAlignMiddle:
		yOffset += (float64(size.Y) - textHeight) / 2
	case VerticalAlignBottom:
		yOffset += float64(size.Y) - textHeight
	}
	return yOffset
}
