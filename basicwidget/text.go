// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Guigui Authors

package basicwidget

import (
	"image"
	"image/color"
	"log/slog"
	"math"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/rivo/uniseg"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget/internal/draw"
	"github.com/hajimehoshi/guigui/basicwidget/internal/textutil"
	"github.com/hajimehoshi/guigui/internal/clipboard"
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

func isMouseButtonRepeating(button ebiten.MouseButton) bool {
	return repeat(inpututil.MouseButtonPressDuration(button))
}

func isKeyRepeating(key ebiten.Key) bool {
	return repeat(inpututil.KeyPressDuration(key))
}

func repeat(duration int) bool {
	if duration == 1 {
		return true
	}
	delay := ebiten.TPS() * 24 / 60
	if duration < delay {
		return false
	}
	return (duration-delay)%4 == 0
}

func findWordBoundaries(text string, idx int) (start, end int) {
	start = idx
	end = idx

	word, _, _ := uniseg.FirstWordInString(text[idx:], -1)
	end += len(word)

	for {
		word, _, _ = uniseg.FirstWordInString(text[start:], -1)
		if start+len(word) < end {
			start += len(word)
			break
		}
		if start == 0 {
			break
		}
		_, l := utf8.DecodeLastRuneInString(text[:start])
		start -= l
	}

	return start, end
}

type Text struct {
	guigui.DefaultWidget

	field       textinput.Field
	nextText    string
	nextTextSet bool

	hAlign      HorizontalAlign
	vAlign      VerticalAlign
	color       color.Color
	transparent float64
	locales     []language.Tag
	scaleMinus1 float64
	bold        bool
	tabular     bool

	selectable       bool
	editable         bool
	multiline        bool
	autoWrap         bool
	keepTailingSpace bool

	selectionDragStartPlus1 int
	selectionDragEndPlus1   int

	// selectionShiftIndexPlus1 is the index (+1) of the selection that is moved by Shift and arrow keys.
	selectionShiftIndexPlus1 int

	dragging    bool
	prevFocused bool

	clickCount         int
	lastClickTick      int64
	lastClickTextIndex int

	cursor textCursor

	tmpClipboard string

	cachedTextSize map[textSizeCacheKey]image.Point
	lastFace       text.Face
	lastScale      float64
	lastWidth      int

	onValueChanged func(text string, committed bool)
	onEnterPressed func(text string)

	tmpLocales []language.Tag
}

type textSizeCacheKey struct {
	autoWrap bool
	bold     bool
}

func (t *Text) SetOnValueChanged(f func(text string, committed bool)) {
	t.onValueChanged = f
}

func (t *Text) SetOnEnterPressed(f func(text string)) {
	t.onEnterPressed = f
}

func (t *Text) resetCachedTextSize() {
	clear(t.cachedTextSize)
}

func (t *Text) resetAutoWrapCachedTextSize() {
	delete(t.cachedTextSize, textSizeCacheKey{autoWrap: true, bold: false})
	delete(t.cachedTextSize, textSizeCacheKey{autoWrap: true, bold: true})
}

func (t *Text) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	if f := t.face(context, false); t.lastFace != f {
		t.lastFace = f
		t.resetCachedTextSize()
	}
	if t.lastScale != context.Scale() {
		t.lastScale = context.Scale()
		t.resetCachedTextSize()
	}
	if t.autoWrap && t.lastWidth != context.Size(t).X {
		t.lastWidth = context.Size(t).X
		t.resetAutoWrapCachedTextSize()
	}

	if context.IsFocusedOrHasFocusedChild(t) {
		if !t.prevFocused {
			t.field.Focus()
			t.cursor.resetCounter()
			start, end := t.field.Selection()
			if start < 0 || end < 0 {
				t.selectAll()
			}
		}
	} else {
		if t.prevFocused {
			t.commit()
		} else if t.nextTextSet {
			t.setText(t.nextText)
		}
	}

	t.prevFocused = context.IsFocusedOrHasFocusedChild(t)

	if t.selectable || t.editable {
		t.cursor.text = t
		b := t.cursorBounds(context)
		appender.AppendChildWidgetWithBounds(&t.cursor, b)
	}

	return nil
}

func (t *Text) SetSelectable(selectable bool) {
	if t.selectable == selectable {
		return
	}
	t.selectable = selectable
	t.selectionDragStartPlus1 = 0
	t.selectionDragEndPlus1 = 0
	t.selectionShiftIndexPlus1 = 0
	if !t.selectable {
		t.setTextAndSelection(t.field.Text(), 0, 0, -1)
	}
	guigui.RequestRedraw(t)
}

func (t *Text) Value() string {
	return t.field.Text()
}

func (t *Text) SetValue(text string) {
	if t.nextTextSet && t.nextText == text {
		return
	}
	if !t.nextTextSet && t.field.Text() == text {
		return
	}

	// When a user is editing, the text should not be changed.
	// Update the actual value later.
	t.nextText = text
	t.nextTextSet = true
	t.resetCachedTextSize()
}

func (t *Text) ForceSetValue(text string) {
	t.setText(text)
}

func (t *Text) setText(text string) {
	start, end := t.field.Selection()
	start = min(start, len(text))
	end = min(end, len(text))
	t.setTextAndSelection(text, start, end, -1)
	t.nextText = ""
	t.nextTextSet = false
}

func (t *Text) selectAll() {
	t.setTextAndSelection(t.field.Text(), 0, len(t.field.Text()), -1)
}

func (t *Text) setSelection(start, end int) {
	t.setTextAndSelection(t.field.Text(), start, end, -1)
}

func (t *Text) setTextAndSelection(text string, start, end int, shiftIndex int) {
	if !t.multiline {
		text, start, end, shiftIndex = replaceNewLinesWithSpace(text, start, end, shiftIndex)
	}

	t.selectionShiftIndexPlus1 = shiftIndex + 1
	if start > end {
		start, end = end, start
	}

	textChanged := t.field.Text() != text
	if s, e := t.field.Selection(); t.field.Text() == text && s == start && e == end {
		return
	}
	t.field.SetTextAndSelection(text, start, end)
	guigui.RequestRedraw(t)
	if textChanged {
		t.resetCachedTextSize()
		if t.onValueChanged != nil {
			t.onValueChanged(t.field.Text(), false)
		}
	}
}

func (t *Text) SetLocales(locales []language.Tag) {
	if slices.Equal(t.locales, locales) {
		return
	}

	t.locales = append([]language.Tag(nil), locales...)
	guigui.RequestRedraw(t)
}

func (t *Text) SetBold(bold bool) {
	if t.bold == bold {
		return
	}

	t.bold = bold
	guigui.RequestRedraw(t)
}

func (t *Text) SetTabular(tabular bool) {
	if t.tabular == tabular {
		return
	}

	t.tabular = tabular
	guigui.RequestRedraw(t)
}

func (t *Text) SetScale(scale float64) {
	if t.scaleMinus1 == scale-1 {
		return
	}

	t.scaleMinus1 = scale - 1
	guigui.RequestRedraw(t)
}

func (t *Text) HorizontalAlign() HorizontalAlign {
	return t.hAlign
}

func (t *Text) SetHorizontalAlign(align HorizontalAlign) {
	if t.hAlign == align {
		return
	}

	t.hAlign = align
	guigui.RequestRedraw(t)
}

func (t *Text) VerticalAlign() VerticalAlign {
	return t.vAlign
}

func (t *Text) SetVerticalAlign(align VerticalAlign) {
	if t.vAlign == align {
		return
	}

	t.vAlign = align
	guigui.RequestRedraw(t)
}

func (t *Text) SetColor(color color.Color) {
	if draw.EqualColor(t.color, color) {
		return
	}

	t.color = color
	guigui.RequestRedraw(t)
}

func (t *Text) SetOpacity(opacity float64) {
	if 1-t.transparent == opacity {
		return
	}

	t.transparent = 1 - opacity
	guigui.RequestRedraw(t)
}

func (t *Text) IsEditable() bool {
	return t.editable
}

func (t *Text) SetEditable(editable bool) {
	if t.editable == editable {
		return
	}

	if editable {
		t.selectionDragStartPlus1 = 0
		t.selectionDragEndPlus1 = 0
		t.selectionShiftIndexPlus1 = 0
	}
	t.editable = editable
	guigui.RequestRedraw(t)
}

func (t *Text) IsMultiline() bool {
	return t.multiline
}

func (t *Text) SetMultiline(multiline bool) {
	if t.multiline == multiline {
		return
	}

	t.multiline = multiline
	guigui.RequestRedraw(t)
}

func (t *Text) SetAutoWrap(autoWrap bool) {
	if t.autoWrap == autoWrap {
		return
	}

	t.autoWrap = autoWrap
	guigui.RequestRedraw(t)
}

func (t *Text) setKeepTailingSpace(keep bool) {
	if t.keepTailingSpace == keep {
		return
	}

	t.keepTailingSpace = keep
	guigui.RequestRedraw(t)
}

func (t *Text) textBounds(context *guigui.Context) image.Rectangle {
	b := context.Bounds(t)

	ts := t.TextSize(context)

	switch t.vAlign {
	case VerticalAlignTop:
		b.Max.Y = b.Min.Y + ts.Y
	case VerticalAlignMiddle:
		h := b.Dy()
		b.Min.Y += (h - ts.Y) / 2
		b.Max.Y = b.Min.Y + ts.Y
	case VerticalAlignBottom:
		b.Min.Y = b.Max.Y - ts.Y
	}

	return b
}

func (t *Text) face(context *guigui.Context, forceBold bool) text.Face {
	size := FontSize(context) * (t.scaleMinus1 + 1)
	weight := text.WeightMedium
	if t.bold || forceBold {
		weight = text.WeightBold
	}

	var liga uint32
	if !t.selectable && !t.editable {
		liga = 1
	}
	var tnum uint32
	if t.tabular {
		tnum = 1
	}

	features := []fontFeature{
		{
			Tag:   text.MustParseTag("liga"),
			Value: liga,
		},
		{
			Tag:   text.MustParseTag("tnum"),
			Value: tnum,
		},
	}

	var lang language.Tag
	if len(t.locales) > 0 {
		lang = t.locales[0]
	} else {
		t.tmpLocales = slices.Delete(t.tmpLocales, 0, len(t.tmpLocales))
		t.tmpLocales = context.AppendLocales(t.tmpLocales)
		if len(t.tmpLocales) > 0 {
			lang = t.tmpLocales[0]
		}
	}
	return fontFace(size, weight, features, lang)
}

func (t *Text) lineHeight(context *guigui.Context) float64 {
	return LineHeight(context) * (t.scaleMinus1 + 1)
}

func (t *Text) HandlePointingInput(context *guigui.Context) guigui.HandleInputResult {
	if !t.selectable && !t.editable {
		return guigui.HandleInputResult{}
	}

	cursorPosition := image.Pt(ebiten.CursorPosition())
	if t.dragging {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			idx := t.textIndexFromPosition(context, cursorPosition, false)
			start, end := idx, idx
			if t.selectionDragStartPlus1-1 >= 0 {
				start = min(start, t.selectionDragStartPlus1-1)
			}
			if t.selectionDragEndPlus1-1 >= 0 {
				end = max(idx, t.selectionDragEndPlus1-1)
			}
			t.setTextAndSelection(t.field.Text(), start, end, -1)
		}
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			t.dragging = false
			t.selectionDragStartPlus1 = 0
			t.selectionDragEndPlus1 = 0
		}
		return guigui.HandleInputByWidget(t)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if context.IsWidgetHitAt(t, cursorPosition) {
			idx := t.textIndexFromPosition(context, cursorPosition, false)

			if ebiten.Tick()-t.lastClickTick < int64(ebiten.TPS()/2) && t.lastClickTextIndex == idx {
				t.clickCount++
			} else {
				t.clickCount = 1
			}

			switch t.clickCount {
			case 1:
				t.dragging = true
				t.selectionDragStartPlus1 = idx + 1
				t.selectionDragEndPlus1 = idx + 1
				if start, end := t.field.Selection(); start != idx || end != idx {
					t.setTextAndSelection(t.field.Text(), idx, idx, -1)
				}
			case 2:
				t.dragging = true
				text := t.field.Text()
				start, end := findWordBoundaries(text, idx)
				t.selectionDragStartPlus1 = start + 1
				t.selectionDragEndPlus1 = end + 1
				t.setTextAndSelection(text, start, end, -1)
			case 3:
				t.selectAll()
			}

			context.SetFocused(t, true)
			t.lastClickTick = ebiten.Tick()
			t.lastClickTextIndex = idx
			return guigui.HandleInputByWidget(t)
		}
		context.SetFocused(t, false)
	}

	if !context.IsFocusedOrHasFocusedChild(t) {
		if t.field.IsFocused() {
			t.field.Blur()
			guigui.RequestRedraw(t)
		}
		return guigui.HandleInputResult{}
	}
	t.field.Focus()

	if !t.editable && !t.selectable {
		return guigui.HandleInputResult{}
	}

	return guigui.HandleInputResult{}
}

func (t *Text) textToDraw(context *guigui.Context, showComposition bool) string {
	if !context.IsFocusedOrHasFocusedChild(t) && t.nextTextSet {
		return t.nextText
	}
	if showComposition {
		return t.field.TextForRendering()
	}
	return t.field.Text()
}

func (t *Text) selectionToDraw(context *guigui.Context) (start, end int, ok bool) {
	s, e := t.field.Selection()
	if !t.editable {
		return s, e, true
	}
	if !context.IsFocusedOrHasFocusedChild(t) {
		return s, e, true
	}
	cs, ce, ok := t.field.CompositionSelection()
	if !ok {
		return s, e, true
	}
	// When cs == ce, the composition already started but any conversion is not done yet.
	// In this case, put the cursor at the end of the composition.
	// TODO: This behavior might be macOS specific. Investigate this.
	if cs == ce {
		return s + ce, s + ce, true
	}
	return 0, 0, false
}

func (t *Text) compositionSelectionToDraw(context *guigui.Context) (uStart, cStart, cEnd, uEnd int, ok bool) {
	if !t.editable {
		return 0, 0, 0, 0, false
	}
	if !context.IsFocusedOrHasFocusedChild(t) {
		return 0, 0, 0, 0, false
	}
	s, _ := t.field.Selection()
	cs, ce, ok := t.field.CompositionSelection()
	if !ok {
		return 0, 0, 0, 0, false
	}
	// When cs == ce, the composition already started but any conversion is not done yet.
	// In this case, assume the entire region is the composition.
	// TODO: This behavior might be macOS specific. Investigate this.
	l := t.field.UncommittedTextLengthInBytes()
	if cs == ce {
		return s, s, s + l, s + l, true
	}
	return s, s + cs, s + ce, s + l, true
}

func (t *Text) HandleButtonInput(context *guigui.Context) guigui.HandleInputResult {
	if !t.selectable && !t.editable {
		return guigui.HandleInputResult{}
	}

	if t.editable {
		origText := t.field.Text()
		start, _ := t.field.Selection()
		var processed bool
		if pos, ok := t.textPosition(context, start, false); ok {
			var err error
			processed, err = t.field.HandleInputWithBounds(image.Rect(int(pos.X), int(pos.Top), int(pos.X+1), int(pos.Bottom)))
			if err != nil {
				slog.Error(err.Error())
				return guigui.AbortHandlingInputByWidget(t)
			}
		}
		if processed {
			guigui.RequestRedraw(t)
			// Reset the cache size before adjust the scroll offset in order to get the correct text size.
			t.resetCachedTextSize()
			if t.field.Text() != origText {
				if t.onValueChanged != nil {
					t.onValueChanged(t.field.Text(), false)
				}
			}
			return guigui.HandleInputByWidget(t)
		}

		// Do not accept key inputs when compositing.
		if _, _, ok := t.field.CompositionSelection(); ok {
			return guigui.HandleInputByWidget(t)
		}

		// For Windows key binds, see:
		// https://support.microsoft.com/en-us/windows/keyboard-shortcuts-in-windows-dcc61a57-8ff0-cffe-9796-cb9706c75eec#textediting

		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyEnter):
			if t.multiline {
				start, end := t.field.Selection()
				text := t.field.Text()[:start] + "\n" + t.field.Text()[end:]
				t.setTextAndSelection(text, start+len("\n"), start+len("\n"), -1)
			}
			if !t.multiline {
				t.commit()
			}
			// TODO: This is not reached on browsers. Fix this.
			if t.onEnterPressed != nil {
				t.onEnterPressed(t.field.Text())
			}
			return guigui.HandleInputByWidget(t)
		case isKeyRepeating(ebiten.KeyBackspace) ||
			useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyH):
			start, end := t.field.Selection()
			if start != end {
				text := t.field.Text()[:start] + t.field.Text()[end:]
				t.setTextAndSelection(text, start, start, -1)
			} else if start > 0 {
				text, pos := textutil.BackspaceOnGraphemes(t.field.Text(), start)
				t.setTextAndSelection(text, pos, pos, -1)
			}
			return guigui.HandleInputByWidget(t)
		case !useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyD) ||
			useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyD):
			// Delete
			start, end := t.field.Selection()
			if start != end {
				text := t.field.Text()[:start] + t.field.Text()[end:]
				t.setTextAndSelection(text, start, start, -1)
			} else if useEmacsKeybind() && end < len(t.field.Text()) {
				text, pos := textutil.DeleteOnGraphemes(t.field.Text(), end)
				t.setTextAndSelection(text, pos, pos, -1)
			}
			return guigui.HandleInputByWidget(t)
		case isKeyRepeating(ebiten.KeyDelete):
			// Delete one cluster
			if _, end := t.field.Selection(); end < len(t.field.Text()) {
				text, pos := textutil.DeleteOnGraphemes(t.field.Text(), end)
				t.setTextAndSelection(text, pos, pos, -1)
			}
			return guigui.HandleInputByWidget(t)
		case !useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyX) ||
			useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyMeta) && isKeyRepeating(ebiten.KeyX):
			// Cut
			start, end := t.field.Selection()
			if start != end {
				if err := clipboard.WriteAll(t.field.Text()[start:end]); err != nil {
					slog.Error(err.Error())
					return guigui.AbortHandlingInputByWidget(t)
				}
				text := t.field.Text()[:start] + t.field.Text()[end:]
				t.setTextAndSelection(text, start, start, -1)
			}
			return guigui.HandleInputByWidget(t)
		case !useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyV) ||
			useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyMeta) && isKeyRepeating(ebiten.KeyV):
			// Paste
			start, end := t.field.Selection()
			ct, err := clipboard.ReadAll()
			if err != nil {
				slog.Error(err.Error())
				return guigui.AbortHandlingInputByWidget(t)
			}
			text := t.field.Text()[:start] + ct + t.field.Text()[end:]
			t.setTextAndSelection(text, start+len(ct), start+len(ct), -1)
			return guigui.HandleInputByWidget(t)
		}
	}

	switch {
	case isKeyRepeating(ebiten.KeyLeft) ||
		useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyB):
		start, end := t.field.Selection()
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			if t.selectionShiftIndexPlus1-1 == end {
				pos := textutil.PrevPositionOnGraphemes(t.field.Text(), end)
				t.setTextAndSelection(t.field.Text(), start, pos, pos)
			} else {
				pos := textutil.PrevPositionOnGraphemes(t.field.Text(), start)
				t.setTextAndSelection(t.field.Text(), pos, end, pos)
			}
		} else {
			if start != end {
				t.setTextAndSelection(t.field.Text(), start, start, -1)
			} else if start > 0 {
				pos := textutil.PrevPositionOnGraphemes(t.field.Text(), start)
				t.setTextAndSelection(t.field.Text(), pos, pos, -1)
			}
		}
		return guigui.HandleInputByWidget(t)
	case isKeyRepeating(ebiten.KeyRight) ||
		useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyF):
		start, end := t.field.Selection()
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			if t.selectionShiftIndexPlus1-1 == start {
				pos := textutil.NextPositionOnGraphemes(t.field.Text(), start)
				t.setTextAndSelection(t.field.Text(), pos, end, pos)
			} else {
				pos := textutil.NextPositionOnGraphemes(t.field.Text(), end)
				t.setTextAndSelection(t.field.Text(), start, pos, pos)
			}
		} else {
			if start != end {
				t.setTextAndSelection(t.field.Text(), end, end, -1)
			} else if start < len(t.field.Text()) {
				pos := textutil.NextPositionOnGraphemes(t.field.Text(), start)
				t.setTextAndSelection(t.field.Text(), pos, pos, -1)
			}
		}
		return guigui.HandleInputByWidget(t)
	case isKeyRepeating(ebiten.KeyUp) ||
		useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyP):
		lh := t.lineHeight(context)
		shift := ebiten.IsKeyPressed(ebiten.KeyShift)
		var moveEnd bool
		start, end := t.field.Selection()
		idx := start
		if shift && t.selectionShiftIndexPlus1-1 == end {
			idx = end
			moveEnd = true
		}
		if pos, ok := t.textPosition(context, idx, false); ok {
			y := (pos.Top+pos.Bottom)/2 - lh
			idx := t.textIndexFromPosition(context, image.Pt(int(pos.X), int(y)), false)
			if shift {
				if moveEnd {
					t.setTextAndSelection(t.field.Text(), start, idx, idx)
				} else {
					t.setTextAndSelection(t.field.Text(), idx, end, idx)
				}
			} else {
				t.setTextAndSelection(t.field.Text(), idx, idx, -1)
			}
		}
		return guigui.HandleInputByWidget(t)
	case isKeyRepeating(ebiten.KeyDown) ||
		useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyN):
		lh := t.lineHeight(context)
		shift := ebiten.IsKeyPressed(ebiten.KeyShift)
		var moveStart bool
		start, end := t.field.Selection()
		idx := end
		if shift && t.selectionShiftIndexPlus1-1 == start {
			idx = start
			moveStart = true
		}
		if pos, ok := t.textPosition(context, idx, false); ok {
			y := (pos.Top+pos.Bottom)/2 + lh
			idx := t.textIndexFromPosition(context, image.Pt(int(pos.X), int(y)), false)
			if shift {
				if moveStart {
					t.setTextAndSelection(t.field.Text(), idx, end, idx)
				} else {
					t.setTextAndSelection(t.field.Text(), start, idx, idx)
				}
			} else {
				t.setTextAndSelection(t.field.Text(), idx, idx, -1)
			}
		}
		return guigui.HandleInputByWidget(t)
	case useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyA):
		idx := 0
		start, end := t.field.Selection()
		if i := strings.LastIndex(t.field.Text()[:start], "\n"); i >= 0 {
			idx = i + 1
		}
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			t.setTextAndSelection(t.field.Text(), idx, end, idx)
		} else {
			t.setTextAndSelection(t.field.Text(), idx, idx, -1)
		}
		return guigui.HandleInputByWidget(t)
	case useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyE):
		idx := len(t.field.Text())
		start, end := t.field.Selection()
		if i := strings.Index(t.field.Text()[end:], "\n"); i >= 0 {
			idx = end + i
		}
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			t.setTextAndSelection(t.field.Text(), start, idx, idx)
		} else {
			t.setTextAndSelection(t.field.Text(), idx, idx, -1)
		}
		return guigui.HandleInputByWidget(t)
	case !useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyA) ||
		useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyMeta) && isKeyRepeating(ebiten.KeyA):
		t.selectAll()
		return guigui.HandleInputByWidget(t)
	case !useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyC) ||
		useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyMeta) && isKeyRepeating(ebiten.KeyC):
		// Copy
		start, end := t.field.Selection()
		if start != end {
			if err := clipboard.WriteAll(t.field.Text()[start:end]); err != nil {
				slog.Error(err.Error())
				return guigui.AbortHandlingInputByWidget(t)
			}
		}
		return guigui.HandleInputByWidget(t)
	case useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyK):
		// 'Kill' the text after the cursor or the selection.
		start, end := t.field.Selection()
		if start == end {
			end = strings.Index(t.field.Text()[start:], "\n")
			if end < 0 {
				end = len(t.field.Text())
			} else if end == 0 {
				end += start + 1
			} else {
				end += start
			}
		}
		t.tmpClipboard = t.field.Text()[start:end]
		text := t.field.Text()[:start] + t.field.Text()[end:]
		t.setTextAndSelection(text, start, start, -1)
		return guigui.HandleInputByWidget(t)
	case useEmacsKeybind() && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyY):
		// 'Yank' the killed text.
		if t.tmpClipboard != "" {
			start, _ := t.field.Selection()
			text := t.field.Text()[:start] + t.tmpClipboard + t.field.Text()[start:]
			t.setTextAndSelection(text, start+len(t.tmpClipboard), start+len(t.tmpClipboard), -1)
		}
		return guigui.HandleInputByWidget(t)
	}

	return guigui.HandleInputResult{}
}

func (t *Text) commit() {
	if t.onValueChanged != nil {
		t.onValueChanged(t.field.Text(), true)
	}
	t.nextText = ""
	t.nextTextSet = false
}

func (t *Text) Draw(context *guigui.Context, dst *ebiten.Image) {
	textBounds := t.textBounds(context)
	if !textBounds.Overlaps(context.VisibleBounds(t)) {
		return
	}

	var textColor color.Color
	if t.color != nil {
		textColor = t.color
	} else {
		textColor = draw.TextColor(context.ColorMode(), context.IsEnabled(t))
	}
	if t.transparent > 0 {
		textColor = draw.ScaleAlpha(textColor, 1-t.transparent)
	}
	face := t.face(context, false)
	op := &textutil.DrawOptions{
		Options: textutil.Options{
			AutoWrap:         t.autoWrap,
			Face:             face,
			LineHeight:       t.lineHeight(context),
			HorizontalAlign:  textutil.HorizontalAlign(t.hAlign),
			VerticalAlign:    textutil.VerticalAlign(t.vAlign),
			KeepTailingSpace: t.keepTailingSpace,
		},
		TextColor: textColor,
	}
	if start, end, ok := t.selectionToDraw(context); ok {
		if context.IsFocusedOrHasFocusedChild(t) {
			op.DrawSelection = true
			op.SelectionStart = start
			op.SelectionEnd = end
			op.SelectionColor = draw.Color(context.ColorMode(), draw.ColorTypeAccent, 0.8)
		} else {
			op.DrawSelection = false
		}
	}
	if uStart, cStart, cEnd, uEnd, ok := t.compositionSelectionToDraw(context); ok {
		op.DrawComposition = true
		op.CompositionStart = uStart
		op.CompositionEnd = uEnd
		op.CompositionActiveStart = cStart
		op.CompositionActiveEnd = cEnd
		op.InactiveCompositionColor = draw.Color(context.ColorMode(), draw.ColorTypeAccent, 0.8)
		op.ActiveCompositionColor = draw.Color(context.ColorMode(), draw.ColorTypeAccent, 0.4)
		op.CompositionBorderWidth = float32(textCursorWidth(context))
	}
	textutil.Draw(textBounds, dst, t.textToDraw(context, true), op)
}

func (t *Text) DefaultSize(context *guigui.Context) image.Point {
	return t.textSize(context, true, false)
}

func (t *Text) TextSize(context *guigui.Context) image.Point {
	return t.textSize(context, false, false)
}

func (t *Text) boldTextSize(context *guigui.Context) image.Point {
	return t.textSize(context, false, true)
}

func (t *Text) textSize(context *guigui.Context, forceUnwrap bool, forceBold bool) image.Point {
	useAutoWrap := t.autoWrap && !forceUnwrap

	key := textSizeCacheKey{
		autoWrap: useAutoWrap,
		bold:     t.bold || forceBold,
	}
	if size, ok := t.cachedTextSize[key]; ok {
		return size
	}

	txt := t.textToDraw(context, true)
	var w, h float64
	if useAutoWrap {
		cw := context.Size(t).X
		w, h = textutil.Measure(cw, txt, true, t.face(context, forceBold), t.lineHeight(context), t.keepTailingSpace)
	} else {
		// context.Size is not available as this causes infinite recursion, and is not needed. Give 0 as a width.
		w, h = textutil.Measure(0, txt, false, t.face(context, forceBold), t.lineHeight(context), t.keepTailingSpace)
	}
	// If width is 0, the text's bounds and visible bounds are empty, and nothing including its cursor is rendered.
	// Force to set a positive number as the width.
	w = max(w, 1)

	if t.cachedTextSize == nil {
		t.cachedTextSize = map[textSizeCacheKey]image.Point{}
	}

	s := image.Pt(int(math.Ceil(w)), int(math.Ceil(h)))
	t.cachedTextSize[key] = s

	return s
}

func (t *Text) CursorShape(context *guigui.Context) (ebiten.CursorShapeType, bool) {
	if t.selectable || t.editable {
		return ebiten.CursorShapeText, true
	}
	return 0, false
}

func (t *Text) cursorPosition(context *guigui.Context) (position textutil.TextPosition, ok bool) {
	if !context.IsFocusedOrHasFocusedChild(t) {
		return textutil.TextPosition{}, false
	}
	if !t.editable {
		return textutil.TextPosition{}, false
	}
	start, end := t.field.Selection()
	if start < 0 {
		return textutil.TextPosition{}, false
	}
	if end < 0 {
		return textutil.TextPosition{}, false
	}

	_, e, ok := t.selectionToDraw(context)
	if !ok {
		return textutil.TextPosition{}, false
	}

	return t.textPosition(context, e, true)
}

func (t *Text) textIndexFromPosition(context *guigui.Context, position image.Point, showComposition bool) int {
	textBounds := t.textBounds(context)
	if position.Y < textBounds.Min.Y {
		return 0
	}
	txt := t.textToDraw(context, showComposition)
	if position.Y >= textBounds.Max.Y {
		return len(txt)
	}
	op := &textutil.Options{
		AutoWrap:         t.autoWrap,
		Face:             t.face(context, false),
		LineHeight:       t.lineHeight(context),
		HorizontalAlign:  textutil.HorizontalAlign(t.hAlign),
		VerticalAlign:    textutil.VerticalAlign(t.vAlign),
		KeepTailingSpace: t.keepTailingSpace,
	}
	position = position.Sub(textBounds.Min)
	idx := textutil.TextIndexFromPosition(textBounds.Dx(), position, txt, op)
	if idx < 0 || idx > len(txt) {
		return -1
	}
	return idx
}

func (t *Text) textPosition(context *guigui.Context, index int, showComposition bool) (position textutil.TextPosition, ok bool) {
	textBounds := t.textBounds(context)
	if !textBounds.Overlaps(context.VisibleBounds(t)) && t.textToDraw(context, showComposition) != "" {
		return textutil.TextPosition{}, false
	}
	txt := t.textToDraw(context, showComposition)
	op := &textutil.Options{
		AutoWrap:         t.autoWrap,
		Face:             t.face(context, false),
		LineHeight:       t.lineHeight(context),
		HorizontalAlign:  textutil.HorizontalAlign(t.hAlign),
		VerticalAlign:    textutil.VerticalAlign(t.vAlign),
		KeepTailingSpace: t.keepTailingSpace,
	}
	pos0, pos1, count := textutil.TextPositionFromIndex(textBounds.Dx(), txt, index, op)
	if count == 0 {
		return textutil.TextPosition{}, false
	}
	pos := pos0
	if count == 2 {
		pos = pos1
	}
	return textutil.TextPosition{
		X:      pos.X + float64(textBounds.Min.X),
		Top:    pos.Top + float64(textBounds.Min.Y),
		Bottom: pos.Bottom + float64(textBounds.Min.Y),
	}, true
}

func textCursorWidth(context *guigui.Context) int {
	return int(2 * context.Scale())
}

func (t *Text) cursorBounds(context *guigui.Context) image.Rectangle {
	pos, ok := t.cursorPosition(context)
	if !ok {
		return image.Rectangle{}
	}
	w := textCursorWidth(context)
	return image.Rect(int(pos.X)-w/2, int(pos.Top), int(pos.X)+w/2, int(pos.Bottom))
}

type textCursor struct {
	guigui.DefaultWidget

	text *Text

	counter   int
	prevShown bool
	prevPos   textutil.TextPosition
	prevOK    bool
}

func (t *textCursor) resetCounter() {
	t.counter = 0
}

func (t *textCursor) Tick(context *guigui.Context) error {
	pos, ok := t.text.cursorPosition(context)
	if t.prevPos != pos {
		t.resetCounter()
	}
	t.prevPos = pos
	t.prevOK = ok

	t.counter++
	if r := t.shouldRenderCursor(context, t.text); t.prevShown != r {
		t.prevShown = r
		// TODO: This is not efficient. Improve this.
		guigui.RequestRedraw(t)
	}
	return nil
}

func (t *textCursor) shouldRenderCursor(context *guigui.Context, text *Text) bool {
	offset := ebiten.TPS() / 2
	if t.counter > offset && (t.counter-offset)%ebiten.TPS() >= ebiten.TPS()/2 {
		return false
	}
	if _, ok := text.cursorPosition(context); !ok {
		return false
	}
	s, e, ok := text.selectionToDraw(context)
	if !ok {
		return false
	}
	if s != e {
		return false
	}
	return true
}

func (t *textCursor) Draw(context *guigui.Context, dst *ebiten.Image) {
	if !t.shouldRenderCursor(context, t.text) {
		return
	}
	b := t.text.cursorBounds(context)
	tb := context.VisibleBounds(t.text)
	tb.Min.X -= textCursorWidth(context) / 2
	tb.Max.X += textCursorWidth(context) / 2
	if !b.Overlaps(tb) {
		return
	}
	vector.DrawFilledRect(dst.SubImage(tb).(*ebiten.Image), float32(b.Min.X), float32(b.Min.Y), float32(b.Dx()), float32(b.Dy()), draw.Color(context.ColorMode(), draw.ColorTypeAccent, 0.4), false)
}

func (t *textCursor) ZDelta() int {
	return 1
}

func (t *textCursor) DefaultSize(context *guigui.Context) image.Point {
	return t.text.cursorBounds(context).Size()
}

func (t *textCursor) PassThrough() bool {
	return true
}

func replaceNewLinesWithSpace(text string, start, end, shiftIndex int) (string, int, int, int) {
	var buf strings.Builder
	for {
		pos, len := textutil.FirstLineBreakPositionAndLen(text)
		if len == 0 {
			buf.WriteString(text)
			break
		}
		buf.WriteString(text[:pos])
		origLen := buf.Len()
		buf.WriteString(" ")
		if diff := len - 1; diff > 0 {
			if origLen < start {
				if start >= origLen+len {
					start -= diff
				} else {
					// This is a very rare case, e.g. the position is in between '\r' and '\n'.
					start = origLen + 1
				}
			}
			if origLen < end {
				if end >= origLen+len {
					end -= diff
				} else {
					end = origLen + 1
				}
			}
			if shiftIndex >= 0 && origLen < shiftIndex {
				if shiftIndex >= origLen+len {
					shiftIndex -= diff
				} else {
					shiftIndex = origLen + 1
				}
			}
		}
		text = text[pos+len:]
	}
	text = buf.String()

	return text, start, end, shiftIndex
}
