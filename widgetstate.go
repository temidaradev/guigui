// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type bounds3D struct {
	bounds image.Rectangle
	z      int
}

type widgetsAndBounds struct {
	bounds3Ds map[Widget]bounds3D
}

func (w *widgetsAndBounds) reset() {
	clear(w.bounds3Ds)
}

func (w *widgetsAndBounds) append(widget Widget, bounds image.Rectangle) {
	if w.bounds3Ds == nil {
		w.bounds3Ds = map[Widget]bounds3D{}
	}
	w.bounds3Ds[widget] = bounds3D{
		bounds: bounds,
		z:      widget.Z(),
	}
}

func (w *widgetsAndBounds) equals(currentWidgets []Widget) bool {
	if len(w.bounds3Ds) != len(currentWidgets) {
		return false
	}
	for _, widget := range currentWidgets {
		b, ok := w.bounds3Ds[widget]
		if !ok {
			return false
		}
		if b.bounds != Bounds(widget) {
			return false
		}
		if b.z != widget.Z() {
			return false
		}
	}
	return true
}

func (w *widgetsAndBounds) redrawIfDifferentParentZ(app *app) {
	for widget, bounds3D := range w.bounds3Ds {
		if isDifferentParentZ(widget) {
			app.requestRedraw(bounds3D.bounds)
			RequestRedraw(widget)
		}
	}
}

type widgetState struct {
	root bool

	position    image.Point
	widthPlus1  int
	heightPlus1 int

	parent   Widget
	children []Widget
	prev     widgetsAndBounds

	hidden       bool
	disabled     bool
	transparency float64

	offscreen *ebiten.Image
}

func Position(widget Widget) image.Point {
	return widget.widgetState().position
}

func SetPosition(widget Widget, position image.Point) {
	widget.widgetState().position = position
	// Rerendering happens at (*.app).requestRedrawIfTreeChanged if necessary.
}

const AutoSize = -1

func SetSize(widget Widget, width, height int) {
	widget.widgetState().widthPlus1 = width + 1
	widget.widgetState().heightPlus1 = height + 1
}

func Size(widget Widget) (int, int) {
	widgetState := widget.widgetState()
	dw, dh := widget.DefaultSize(&theApp.context)
	var w, h int
	if widgetState.widthPlus1 == 0 {
		w = dw
	} else {
		w = widgetState.widthPlus1 - 1
	}
	if widgetState.heightPlus1 == 0 {
		h = dh
	} else {
		h = widgetState.heightPlus1 - 1
	}
	return w, h
}

func Bounds(widget Widget) image.Rectangle {
	widgetState := widget.widgetState()
	width, height := Size(widget)
	return image.Rectangle{
		Min: widgetState.position,
		Max: widgetState.position.Add(image.Point{width, height}),
	}
}

func VisibleBounds(widget Widget) image.Rectangle {
	parent := widget.widgetState().parent
	if parent == nil {
		return theApp.bounds()
	}
	if isDifferentParentZ(widget) {
		return Bounds(widget)
	}
	return VisibleBounds(parent).Intersect(Bounds(widget))
}

func (w *widgetState) isInTree() bool {
	p := w
	for ; p.parent != nil; p = p.parent.widgetState() {
	}
	return p.root
}

func Show(widget Widget) {
	widgetState := widget.widgetState()
	if !widgetState.hidden {
		return
	}
	widgetState.hidden = false
	RequestRedraw(widget)
}

func Hide(widget Widget) {
	widgetState := widget.widgetState()
	if widgetState.hidden {
		return
	}
	widgetState.hidden = true
	Blur(widget)
	RequestRedraw(widget)
}

func IsVisible(widget Widget) bool {
	return widget.widgetState().isVisible()
}

func (w *widgetState) isVisible() bool {
	if w.parent != nil {
		return !w.hidden && IsVisible(w.parent)
	}
	return !w.hidden
}

func Enable(widget Widget) {
	widgetState := widget.widgetState()
	if !widgetState.disabled {
		return
	}
	widgetState.disabled = false
	RequestRedraw(widget)
}

func Disable(widget Widget) {
	widgetState := widget.widgetState()
	if widgetState.disabled {
		return
	}
	widgetState.disabled = true
	Blur(widget)
	RequestRedraw(widget)
}

func IsEnabled(widget Widget) bool {
	return widget.widgetState().isEnabled()
}

func (w *widgetState) isEnabled() bool {
	if w.parent != nil {
		return !w.disabled && IsEnabled(w.parent)
	}
	return !w.disabled
}

func Focus(widget Widget) {
	widgetState := widget.widgetState()
	if !widgetState.isVisible() {
		return
	}
	if !widgetState.isEnabled() {
		return
	}

	if !widgetState.isInTree() {
		return
	}
	if theApp.focusedWidget == widget {
		return
	}

	var oldWidget Widget
	if theApp.focusedWidget != nil {
		oldWidget = theApp.focusedWidget
	}

	theApp.focusedWidget = widget
	RequestRedraw(theApp.focusedWidget)
	if oldWidget != nil {
		RequestRedraw(oldWidget)
	}
}

func Blur(widget Widget) {
	widgetState := widget.widgetState()
	if !widgetState.isInTree() {
		return
	}
	if theApp.focusedWidget != widget {
		return
	}
	theApp.focusedWidget = nil
	RequestRedraw(widget)
}

func IsFocused(widget Widget) bool {
	widgetState := widget.widgetState()
	return widgetState.isInTree() && theApp.focusedWidget == widget && widgetState.isVisible()
}

func HasFocusedChildWidget(widget Widget) bool {
	widgetState := widget.widgetState()
	if IsFocused(widget) {
		return true
	}
	for _, child := range widgetState.children {
		if HasFocusedChildWidget(child) {
			return true
		}
	}
	return false
}

func Opacity(widget Widget) float64 {
	return widget.widgetState().opacity()
}

func (w *widgetState) opacity() float64 {
	return 1 - w.transparency
}

func SetOpacity(widget Widget, opacity float64) {
	if opacity < 0 {
		opacity = 0
	}
	if opacity > 1 {
		opacity = 1
	}
	widgetState := widget.widgetState()
	if widgetState.transparency == 1-opacity {
		return
	}
	widgetState.transparency = 1 - opacity
	RequestRedraw(widget)
}

func RequestRedraw(widget Widget) {
	theApp.requestRedrawWidget(widget)
}

func (w *widgetState) ensureOffscreen(bounds image.Rectangle) *ebiten.Image {
	if w.offscreen != nil {
		if !bounds.In(w.offscreen.Bounds()) {
			w.offscreen.Deallocate()
			w.offscreen = nil
		}
	}
	if w.offscreen == nil {
		w.offscreen = ebiten.NewImageWithOptions(bounds, nil)
	}
	return w.offscreen.SubImage(bounds).(*ebiten.Image)
}

func traverseWidget(widget Widget, f func(widget Widget)) {
	f(widget)
	for _, child := range widget.widgetState().children {
		traverseWidget(child, f)
	}
}

func isDifferentParentZ(widget Widget) bool {
	parent := widget.widgetState().parent
	if parent == nil {
		return false
	}
	return widget.Z() != parent.Z()
}

func IsWidgetHitAt(widget Widget, point image.Point) bool {
	return theApp.isWidgetHitAt(widget, point)
}
