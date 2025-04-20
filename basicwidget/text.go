// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"image"
	"image/color"
	"log/slog"
	"runtime"
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

func isKeyRepeating(key ebiten.Key) bool {
	d := inpututil.KeyPressDuration(key)
	if d == 1 {
		return true
	}
	delay := ebiten.TPS() * 24 / 60
	if d < delay {
		return false
	}
	return (d-delay)%4 == 0
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

type TextFilter func(text string, start, end int) (string, int, int)

func DefaultTextColor(context *guigui.Context) color.Color {
	return draw.Color(context.ColorMode(), draw.ColorTypeBase, 0.1)
}

type Text struct {
	guigui.DefaultWidget

	field textinput.Field

	hAlign      HorizontalAlign
	vAlign      VerticalAlign
	color       color.Color
	transparent float64
	locales     []language.Tag
	fullLocales []language.Tag
	scaleMinus1 float64
	bold        bool

	selectable           bool
	editable             bool
	multiline            bool
	autoWrap             bool
	selectionDragStart   int
	selectionShiftIndex  int
	dragging             bool
	toAdjustScrollOffset bool
	prevFocused          bool

	clickCount         int
	lastClickTick      int64
	lastClickTextIndex int

	filter TextFilter

	cursor        textCursor
	scrollOverlay ScrollOverlay

	temporaryClipboard string

	cachedTextSizePlus1         image.Point
	cachedAutoWrapTextSizePlus1 image.Point
	lastFace                    text.Face
	lastAppScale                float64
	lastWidth                   int

	onEnterPressed func(text string)
}

func (t *Text) SetOnEnterPressed(f func(text string)) {
	t.onEnterPressed = f
}

func (t *Text) resetCachedTextSize() {
	t.cachedTextSizePlus1 = image.Point{}
	t.cachedAutoWrapTextSizePlus1 = image.Point{}
}

func (t *Text) resetAutoWrapCachedTextSize() {
	t.cachedAutoWrapTextSizePlus1 = image.Point{}
}

func (t *Text) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	if f := t.face(context); t.lastFace != f {
		t.lastFace = f
		t.resetCachedTextSize()
	}
	if t.lastAppScale != context.AppScale() {
		t.lastAppScale = context.AppScale()
		t.resetCachedTextSize()
	}
	if t.autoWrap && t.lastWidth != context.Size(t).X {
		t.lastWidth = context.Size(t).X
		t.resetAutoWrapCachedTextSize()
	}

	t.scrollOverlay.SetContentSize(context, t.TextSize(context))

	if !t.prevFocused && context.IsFocused(t) {
		t.field.Focus()
		t.cursor.resetCounter()
		start, end := t.field.Selection()
		if start < 0 || end < 0 {
			t.selectAll()
		}
	} else if t.prevFocused && !context.IsFocused(t) {
		t.applyFilter()
	}
	t.prevFocused = context.IsFocused(t)

	if t.toAdjustScrollOffset && !context.VisibleBounds(t).Empty() {
		t.adjustScrollOffset(context)
		t.toAdjustScrollOffset = false
	}

	if t.selectable || t.editable {
		t.cursor.text = t
		p := context.Position(t)
		p.X -= cursorWidth(context)
		appender.AppendChildWidgetWithPosition(&t.cursor, p)
	}

	context.Hide(&t.scrollOverlay)
	appender.AppendChildWidgetWithBounds(&t.scrollOverlay, context.Bounds(t))

	return nil
}

func (t *Text) SetSelectable(selectable bool) {
	if t.selectable == selectable {
		return
	}
	t.selectable = selectable
	t.selectionDragStart = -1
	t.selectionShiftIndex = -1
	if !t.selectable {
		t.setTextAndSelection(t.field.Text(), 0, 0, -1)
	}
	guigui.RequestRedraw(t)
}

func (t *Text) Text() string {
	return t.field.Text()
}

func (t *Text) SetText(text string) {
	start, end := t.field.Selection()
	start = min(start, len(text))
	end = min(end, len(text))
	t.setTextAndSelection(text, start, end, -1)
}

func (t *Text) SetFilter(filter TextFilter) {
	t.filter = filter
	t.applyFilter()
}

func (t *Text) selectAll() {
	t.setTextAndSelection(t.field.Text(), 0, len(t.field.Text()), -1)
}

func (t *Text) setTextAndSelection(text string, start, end int, shiftIndex int) {
	t.selectionShiftIndex = shiftIndex
	if start > end {
		start, end = end, start
	}

	textChanged := t.field.Text() != text
	if s, e := t.field.Selection(); t.field.Text() == text && s == start && e == end {
		return
	}
	t.field.SetTextAndSelection(text, start, end)
	t.toAdjustScrollOffset = true
	guigui.RequestRedraw(t)
	if textChanged {
		t.resetCachedTextSize()
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

func (t *Text) SetScale(scale float64) {
	if t.scaleMinus1 == scale-1 {
		return
	}

	t.scaleMinus1 = scale - 1
	guigui.RequestRedraw(t)
}

func (t *Text) SetHorizontalAlign(align HorizontalAlign) {
	if t.hAlign == align {
		return
	}

	t.hAlign = align
	guigui.RequestRedraw(t)
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

func (t *Text) SetEditable(editable bool) {
	if t.editable == editable {
		return
	}

	if editable {
		t.selectionDragStart = -1
		t.selectionShiftIndex = -1
	}
	t.editable = editable
	guigui.RequestRedraw(t)
}

func (t *Text) SetScrollable(context *guigui.Context, scrollable bool) {
	if scrollable {
		context.Show(&t.scrollOverlay)
	} else {
		context.Hide(&t.scrollOverlay)
	}
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

func (t *Text) textBounds(context *guigui.Context) image.Rectangle {
	offsetX, offsetY := t.scrollOverlay.Offset()

	b := context.Bounds(t)

	ts := t.TextSize(context)
	if b.Dx() < ts.X {
		b.Max.X = b.Min.X + ts.X
	}

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

	b = b.Add(image.Pt(int(offsetX), int(offsetY)))
	return b
}

func (t *Text) face(context *guigui.Context) text.Face {
	size := FontSize(context) * (t.scaleMinus1 + 1)
	weight := text.WeightMedium
	if t.bold {
		weight = text.WeightBold
	}
	t.fullLocales = slices.Delete(t.fullLocales, 0, len(t.fullLocales))
	t.fullLocales = append(t.fullLocales, t.locales...)
	t.fullLocales = context.AppendLocales(t.fullLocales)
	return fontFace(size, weight, true, t.fullLocales)
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
			if idx < t.selectionDragStart {
				t.setTextAndSelection(t.field.Text(), idx, t.selectionDragStart, -1)
			} else {
				t.setTextAndSelection(t.field.Text(), t.selectionDragStart, idx, -1)
			}
		}
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			t.dragging = false
			t.selectionDragStart = -1
		}
		return guigui.HandleInputByWidget(t)
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if cursorPosition.In(context.VisibleBounds(t)) {
			idx := t.textIndexFromPosition(context, cursorPosition, false)

			if ebiten.Tick()-t.lastClickTick < int64(ebiten.TPS()/2) && t.lastClickTextIndex == idx {
				t.clickCount++
			} else {
				t.clickCount = 1
			}

			switch t.clickCount {
			case 1:
				t.dragging = true
				t.selectionDragStart = idx
				if start, end := t.field.Selection(); start != idx || end != idx {
					t.setTextAndSelection(t.field.Text(), idx, idx, -1)
				}
			case 2:
				text := t.field.Text()
				start, end := findWordBoundaries(text, idx)
				// TODO: `selectionDragEnd` needed to emulate Chrome's behavior.
				t.selectionDragStart = start
				t.setTextAndSelection(text, start, end, -1)
			case 3:
				t.selectAll()
			}

			context.Focus(t)
			t.lastClickTick = ebiten.Tick()
			t.lastClickTextIndex = idx
			return guigui.HandleInputByWidget(t)
		}
		context.Blur(t)
	}

	if !context.IsFocused(t) {
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

func (t *Text) adjustScrollOffset(context *guigui.Context) {
	start, end, ok := t.selectionToDraw(context)
	if !ok {
		return
	}

	tb := t.textBounds(context)
	bounds := context.Bounds(t)
	if pos, ok := t.textPosition(context, end, true); ok {
		var dx, dy float64
		if max := float64(bounds.Max.X); pos.X > max {
			dx = max - pos.X
		}
		if max := float64(bounds.Max.Y); pos.Bottom > max {
			dy = max - pos.Bottom
		}
		t.scrollOverlay.SetOffsetByDelta(context, tb.Size(), dx, dy)
	}
	if pos, ok := t.textPosition(context, start, true); ok {
		var dx, dy float64
		if min := float64(bounds.Min.X); pos.X < min {
			dx = min - pos.X
		}
		if min := float64(bounds.Min.Y); pos.Top < min {
			dy = min - pos.Top
		}
		t.scrollOverlay.SetOffsetByDelta(context, tb.Size(), dx, dy)
	}
}

func (t *Text) textToDraw(showComposition bool) string {
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
	if !context.IsFocused(t) {
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
	if !context.IsFocused(t) {
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
	if !context.IsFocused(t) || !context.IsEnabled(t) {
		return guigui.HandleInputResult{}
	}

	if !t.selectable && !t.editable {
		return guigui.HandleInputResult{}
	}

	start, _ := t.field.Selection()
	var processed bool
	if pos, ok := t.textPosition(context, start, false); ok {
		var err error
		processed, err = t.field.HandleInput(int(pos.X), int(pos.Bottom))
		if err != nil {
			slog.Error(err.Error())
			return guigui.AbortHandlingInputByWidget(t)
		}
	}
	if processed {
		guigui.RequestRedraw(t)
		// Reset the cache size before adjust the scroll offset in order to get the correct text size.
		t.resetCachedTextSize()
		t.adjustScrollOffset(context)
		return guigui.HandleInputByWidget(t)
	}

	// Do not accept key inputs when compositing.
	if _, _, ok := t.field.CompositionSelection(); ok {
		return guigui.HandleInputByWidget(t)
	}

	// For Windows key binds, see:
	// https://support.microsoft.com/en-us/windows/keyboard-shortcuts-in-windows-dcc61a57-8ff0-cffe-9796-cb9706c75eec#textediting

	// TODO: Use WebAPI to detect OS is runtime.GOOS == "js"
	isDarwin := runtime.GOOS == "darwin"

	if t.editable {
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyEnter):
			if t.multiline {
				start, end := t.field.Selection()
				text := t.field.Text()[:start] + "\n" + t.field.Text()[end:]
				t.setTextAndSelection(text, start+len("\n"), start+len("\n"), -1)
			}
			t.applyFilter()
			// TODO: This is not reached on browsers. Fix this.
			if t.onEnterPressed != nil {
				t.onEnterPressed(t.field.Text())
			}
		case isKeyRepeating(ebiten.KeyBackspace) ||
			isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyH):
			start, end := t.field.Selection()
			if start != end {
				text := t.field.Text()[:start] + t.field.Text()[end:]
				t.setTextAndSelection(text, start, start, -1)
			} else if start > 0 {
				text, pos := textutil.BackspaceOnGraphemes(t.field.Text(), start)
				t.setTextAndSelection(text, pos, pos, -1)
			}
		case !isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyD) ||
			isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyD):
			// Delete
			start, end := t.field.Selection()
			if start != end {
				text := t.field.Text()[:start] + t.field.Text()[end:]
				t.setTextAndSelection(text, start, start, -1)
			} else if isDarwin && end < len(t.field.Text()) {
				text, pos := textutil.DeleteOnGraphemes(t.field.Text(), end)
				t.setTextAndSelection(text, pos, pos, -1)
			}
		case isKeyRepeating(ebiten.KeyDelete):
			// Delete one cluster
			if _, end := t.field.Selection(); end < len(t.field.Text()) {
				text, pos := textutil.DeleteOnGraphemes(t.field.Text(), end)
				t.setTextAndSelection(text, pos, pos, -1)
			}

		case !isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyX) ||
			isDarwin && ebiten.IsKeyPressed(ebiten.KeyMeta) && isKeyRepeating(ebiten.KeyX):
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
		case !isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyV) ||
			isDarwin && ebiten.IsKeyPressed(ebiten.KeyMeta) && isKeyRepeating(ebiten.KeyV):
			// Paste
			start, end := t.field.Selection()
			ct, err := clipboard.ReadAll()
			if err != nil {
				slog.Error(err.Error())
				return guigui.AbortHandlingInputByWidget(t)
			}
			text := t.field.Text()[:start] + ct + t.field.Text()[end:]
			t.setTextAndSelection(text, start+len(ct), start+len(ct), -1)
		}
	}

	switch {
	case isKeyRepeating(ebiten.KeyLeft) ||
		isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyB):
		start, end := t.field.Selection()
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			if t.selectionShiftIndex == end {
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
	case isKeyRepeating(ebiten.KeyRight) ||
		isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyF):
		start, end := t.field.Selection()
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			if t.selectionShiftIndex == start {
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
	case isKeyRepeating(ebiten.KeyUp) ||
		isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyP):
		lh := t.lineHeight(context)
		shift := ebiten.IsKeyPressed(ebiten.KeyShift)
		var moveEnd bool
		start, end := t.field.Selection()
		idx := start
		if shift && t.selectionShiftIndex == end {
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
	case isKeyRepeating(ebiten.KeyDown) ||
		isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyN):
		lh := t.lineHeight(context)
		shift := ebiten.IsKeyPressed(ebiten.KeyShift)
		var moveStart bool
		start, end := t.field.Selection()
		idx := end
		if shift && t.selectionShiftIndex == start {
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
	case isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyA):
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
	case isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyE):
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
	case !isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyA) ||
		isDarwin && ebiten.IsKeyPressed(ebiten.KeyMeta) && isKeyRepeating(ebiten.KeyA):
		t.selectAll()
	case !isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyC) ||
		isDarwin && ebiten.IsKeyPressed(ebiten.KeyMeta) && isKeyRepeating(ebiten.KeyC):
		// Copy
		start, end := t.field.Selection()
		if start != end {
			if err := clipboard.WriteAll(t.field.Text()[start:end]); err != nil {
				slog.Error(err.Error())
				return guigui.AbortHandlingInputByWidget(t)
			}
		}
	case isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyK):
		// 'Kill' the text after the cursor or the selection.
		start, end := t.field.Selection()
		if start == end {
			end = strings.Index(t.field.Text()[start:], "\n")
			if end < 0 {
				end = len(t.field.Text())
			} else {
				end += start
			}
		}
		t.temporaryClipboard = t.field.Text()[start:end]
		text := t.field.Text()[:start] + t.field.Text()[end:]
		t.setTextAndSelection(text, start, start, -1)
	case isDarwin && ebiten.IsKeyPressed(ebiten.KeyControl) && isKeyRepeating(ebiten.KeyY):
		// 'Yank' the killed text.
		if t.temporaryClipboard != "" {
			start, _ := t.field.Selection()
			text := t.field.Text()[:start] + t.temporaryClipboard + t.field.Text()[start:]
			t.setTextAndSelection(text, start+len(t.temporaryClipboard), start+len(t.temporaryClipboard), -1)
		}
	}

	return guigui.HandleInputByWidget(t)
}

func (t *Text) applyFilter() {
	if t.filter != nil {
		start, end := t.field.Selection()
		text, start, end := t.filter(t.field.Text(), start, end)
		t.setTextAndSelection(text, start, end, -1)
	}
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
		textColor = DefaultTextColor(context)
	}
	if t.transparent > 0 {
		textColor = draw.ScaleAlpha(textColor, 1-t.transparent)
	}
	face := t.face(context)
	op := &textutil.DrawOptions{
		Options: textutil.Options{
			AutoWrap:        t.autoWrap,
			Face:            face,
			LineHeight:      t.lineHeight(context),
			HorizontalAlign: textutil.HorizontalAlign(t.hAlign),
			VerticalAlign:   textutil.VerticalAlign(t.vAlign),
		},
		TextColor: textColor,
	}
	if start, end, ok := t.selectionToDraw(context); ok {
		op.DrawSelection = true
		op.SelectionStart = start
		op.SelectionEnd = end
		op.SelectionColor = draw.Color(context.ColorMode(), draw.ColorTypeAccent, 0.8)
	}
	if uStart, cStart, cEnd, uEnd, ok := t.compositionSelectionToDraw(context); ok {
		op.DrawComposition = true
		op.CompositionStart = uStart
		op.CompositionEnd = uEnd
		op.CompositionActiveStart = cStart
		op.CompositionActiveEnd = cEnd
		op.InactiveCompositionColor = draw.Color(context.ColorMode(), draw.ColorTypeAccent, 0.8)
		op.ActiveCompositionColor = draw.Color(context.ColorMode(), draw.ColorTypeAccent, 0.4)
		op.CompositionBorderWidth = float32(cursorWidth(context))
	}
	textutil.Draw(textBounds, dst, t.textToDraw(true), op)
}

func (t *Text) DefaultSize(context *guigui.Context) image.Point {
	return t.textSize(context, true)
}

func (t *Text) TextSize(context *guigui.Context) image.Point {
	return t.textSize(context, false)
}

func (t *Text) textSize(context *guigui.Context, forceUnwrap bool) image.Point {
	useAutoWrap := t.autoWrap && !forceUnwrap
	if useAutoWrap {
		if t.cachedAutoWrapTextSizePlus1.X > 0 && t.cachedAutoWrapTextSizePlus1.Y > 0 {
			return t.cachedAutoWrapTextSizePlus1.Add(image.Pt(-1, -1))
		}
	} else {
		if t.cachedTextSizePlus1.X > 0 && t.cachedTextSizePlus1.Y > 0 {
			return t.cachedTextSizePlus1.Add(image.Pt(-1, -1))
		}
	}

	txt := t.textToDraw(true)
	var w, h float64
	if useAutoWrap {
		cw := context.Size(t).X
		w, h = textutil.Measure(cw, txt, true, t.face(context), t.lineHeight(context))
	} else {
		// context.Size is not available as this causes infinite recursion, and is not needed. Give 0 as a width.
		w, h = textutil.Measure(0, txt, false, t.face(context), t.lineHeight(context))
	}
	w *= t.scaleMinus1 + 1
	if useAutoWrap {
		t.cachedAutoWrapTextSizePlus1 = image.Pt(int(w)+1, int(h)+1)
	} else {
		t.cachedTextSizePlus1 = image.Pt(int(w)+1, int(h)+1)
	}
	return image.Pt(int(w), int(h))
}

func (t *Text) CursorShape(context *guigui.Context) (ebiten.CursorShapeType, bool) {
	if t.selectable || t.editable {
		return ebiten.CursorShapeText, true
	}
	return 0, false
}

func (t *Text) cursorPosition(context *guigui.Context) (position textutil.TextPosition, ok bool) {
	if !context.IsFocused(t) {
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
	if !textBounds.Overlaps(context.VisibleBounds(t)) {
		return -1
	}
	txt := t.textToDraw(showComposition)
	op := &textutil.Options{
		AutoWrap:        t.autoWrap,
		Face:            t.face(context),
		LineHeight:      t.lineHeight(context),
		HorizontalAlign: textutil.HorizontalAlign(t.hAlign),
		VerticalAlign:   textutil.VerticalAlign(t.vAlign),
	}
	position = position.Sub(textBounds.Min)
	yOffset := textutil.TextPositionYOffset(textBounds.Size(), txt, op)
	position = position.Sub(image.Pt(0, int(yOffset)))
	idx := textutil.TextIndexFromPosition(textBounds.Dx(), position, txt, op)
	if idx < 0 || idx > len(txt) {
		return -1
	}
	return idx
}

func (t *Text) textPosition(context *guigui.Context, index int, showComposition bool) (position textutil.TextPosition, ok bool) {
	textBounds := t.textBounds(context)
	if !textBounds.Overlaps(context.VisibleBounds(t)) {
		return textutil.TextPosition{}, false
	}
	txt := t.textToDraw(showComposition)
	op := &textutil.Options{
		AutoWrap:        t.autoWrap,
		Face:            t.face(context),
		LineHeight:      t.lineHeight(context),
		HorizontalAlign: textutil.HorizontalAlign(t.hAlign),
		VerticalAlign:   textutil.VerticalAlign(t.vAlign),
	}
	pos0, pos1, count := textutil.TextPositionFromIndex(textBounds.Dx(), txt, index, op)
	if count == 0 {
		return textutil.TextPosition{}, false
	}
	pos := pos0
	if count == 2 {
		pos = pos1
	}
	yOffset := textutil.TextPositionYOffset(textBounds.Size(), txt, op)
	return textutil.TextPosition{
		X:      pos.X + float64(textBounds.Min.X),
		Top:    pos.Top + float64(textBounds.Min.Y) + yOffset,
		Bottom: pos.Bottom + float64(textBounds.Min.Y) + yOffset,
	}, true
}

func cursorWidth(context *guigui.Context) int {
	return int(2 * context.Scale())
}

func (t *Text) cursorBounds(context *guigui.Context) image.Rectangle {
	pos, ok := t.cursorPosition(context)
	if !ok {
		return image.Rectangle{}
	}
	w := cursorWidth(context)
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

func (t *textCursor) Update(context *guigui.Context) error {
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
	vector.DrawFilledRect(dst, float32(b.Min.X), float32(b.Min.Y), float32(b.Dx()), float32(b.Dy()), draw.Color(context.ColorMode(), draw.ColorTypeAccent, 0.4), false)
}

func (t *textCursor) ZDelta() int {
	return 1
}

func (t *textCursor) DefaultSize(context *guigui.Context) image.Point {
	return context.Size(t.text).Add(image.Pt(2*cursorWidth(context), 0))
}
