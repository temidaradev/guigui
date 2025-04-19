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

func lines(str string) iter.Seq2[int, string] {
	// TODO: Integrate lines and lines2.
	return lines2(0, str, nil, false)
}

// TODO: Remove this. Auto-wrapping should be done in this package internally.
func AutoWrapText(width int, str string, face text.Face) string {
	var lines []string
	for _, line := range lines2(width, str, face, true) {
		line = line[:len(line)-tailingLineBreakLen(line)]
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func lines2(width int, str string, face text.Face, autoWrap bool) iter.Seq2[int, string] {
	return func(yield func(pos int, s string) bool) {
		var line string
		var word string
		var pos int
		state := -1
		for len(str) > 0 {
			cluster, nextStr, boundaries, nextState := uniseg.StepString(str, state)
			switch m := boundaries & uniseg.MaskLine; m {
			default:
				word += cluster
			case uniseg.LineCanBreak, uniseg.LineMustBreak:
				if line == "" || !autoWrap {
					line += word + cluster
				} else {
					l := line + word + cluster
					// TODO: Consider a line alignment and/or editable/selectable states when calculating the width.
					if text.Advance(l[:len(l)-tailingLineBreakLen(l)], face) > float64(width) {
						if !yield(pos, line) {
							return
						}
						pos += len(line)
						line = word + cluster
					} else {
						line += word + cluster
					}
				}
				word = ""
				if m == uniseg.LineMustBreak {
					if !yield(pos, line) {
						return
					}
					pos += len(line)
					line = ""
				}
			}
			state = nextState
			str = nextStr
		}

		line += word
		if len(line) > 0 {
			if !yield(pos, line) {
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
	m := options.Face.Metrics()
	gap := options.LineHeight - m.HAscent - m.HDescent
	var top float64
	n := int((float64(position.Y) - top + gap/2) / options.LineHeight)

	var pos int
	var line string
	var lineIndex int
	for p, l := range lines(str) {
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

func TextPosition(width int, str string, index int, options *Options) (x, top, bottom float64, ok bool) {
	if index < 0 || index > len(str) {
		return 0, 0, 0, false
	}

	var y float64

	var indexInLine int
	var line string
	var found bool
	for p, l := range lines(str) {
		line = l
		if p <= index && index < p+len(l) {
			found = true
			indexInLine = index - p
			break
		}
		y += options.LineHeight
	}
	// When found is false, the position is in the tail of the last line.
	if !found && len(str) > 0 && !uniseg.HasTrailingLineBreakInString(str) {
		indexInLine = len(line)
		y -= options.LineHeight
	}

	x = oneLineLeft(width, line, options.Face, options.HorizontalAlign)
	x += text.Advance(line[:indexInLine], options.Face)

	m := options.Face.Metrics()
	paddingY := (options.LineHeight - (m.HAscent + m.HDescent)) / 2
	return x, y + paddingY, y + options.LineHeight - paddingY, true
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

func lineCount(str string) int {
	var count int
	for range lines(str) {
		count++
	}
	return count
}
