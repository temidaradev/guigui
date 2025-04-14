// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"fmt"
	"image"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/hajimehoshi/guigui/internal/locale"
	"golang.org/x/text/language"
)

type ColorMode int

var defaultColorMode ColorMode = ColorModeLight

func init() {
	// TODO: Consider the system color mode.
	switch mode := os.Getenv("GUIGUI_COLOR_MODE"); mode {
	case "light", "":
		defaultColorMode = ColorModeLight
	case "dark":
		defaultColorMode = ColorModeDark
	default:
		slog.Warn(fmt.Sprintf("invalid GUIGUI_COLOR_MODE: %s", mode))
	}
}

var envLocales []language.Tag

func init() {
	if locales := os.Getenv("GUIGUI_LOCALES"); locales != "" {
		for _, tag := range strings.Split(os.Getenv("GUIGUI_LOCALES"), ",") {
			l, err := language.Parse(strings.TrimSpace(tag))
			if err != nil {
				slog.Warn(fmt.Sprintf("invalid GUIGUI_LOCALES: %s", tag))
				continue
			}
			envLocales = append(envLocales, l)
		}
	}
}

var systemLocales []language.Tag

func init() {
	ls, err := locale.Locales()
	if err != nil {
		slog.Error(err.Error())
		return
	}
	systemLocales = ls
}

const (
	ColorModeLight ColorMode = iota
	ColorModeDark
)

type Context struct {
	app *app

	deviceScale    float64
	appScaleMinus1 float64
	colorMode      ColorMode
	hasColorMode   bool
	locales        []language.Tag
}

func (c *Context) Scale() float64 {
	return c.deviceScale * c.AppScale()
}

func (c *Context) DeviceScale() float64 {
	return c.deviceScale
}

func (c *Context) setDeviceScale(deviceScale float64) {
	if c.deviceScale == deviceScale {
		return
	}
	c.deviceScale = deviceScale
	c.app.requestRedraw(c.app.bounds())
}

func (c *Context) AppScale() float64 {
	return c.appScaleMinus1 + 1
}

func (c *Context) SetAppScale(scale float64) {
	if c.appScaleMinus1 == scale-1 {
		return
	}
	c.appScaleMinus1 = scale - 1
	c.app.requestRedraw(c.app.bounds())
}

func (c *Context) ColorMode() ColorMode {
	if c.hasColorMode {
		return c.colorMode
	}
	return defaultColorMode
}

func (c *Context) SetColorMode(mode ColorMode) {
	if c.hasColorMode && mode == c.colorMode {
		return
	}

	c.colorMode = mode
	c.hasColorMode = true
	c.app.requestRedraw(c.app.bounds())
}

func (c *Context) ResetColorMode() {
	c.hasColorMode = false
}

func (c *Context) AppendLocales(locales []language.Tag) []language.Tag {
	origLen := len(locales)
	// App locales
	for _, l := range c.locales {
		if slices.Contains(locales[origLen:], l) {
			continue
		}
		locales = append(locales, l)
	}
	// Env locales
	for _, l := range envLocales {
		if slices.Contains(locales[origLen:], l) {
			continue
		}
		locales = append(locales, l)
	}
	// System locales
	for _, l := range systemLocales {
		if slices.Contains(locales[origLen:], l) {
			continue
		}
		locales = append(locales, l)
	}
	return locales
}

func (c *Context) AppendAppLocales(locales []language.Tag) []language.Tag {
	origLen := len(locales)
	for _, l := range c.locales {
		if slices.Contains(locales[origLen:], l) {
			continue
		}
		locales = append(locales, l)
	}
	return locales
}

func (c *Context) SetAppLocales(locales []language.Tag) {
	if slices.Equal(c.locales, locales) {
		return
	}

	c.locales = append([]language.Tag(nil), locales...)
	c.app.requestRedraw(c.app.bounds())
}

func (c *Context) AppSize() (int, int) {
	return c.app.bounds().Dx(), c.app.bounds().Dy()
}

func (c *Context) Position(widget Widget) image.Point {
	return widget.widgetState().position
}

func (c *Context) SetPosition(widget Widget, position image.Point) {
	widget.widgetState().position = position
	// Rerendering happens at (*.app).requestRedrawIfTreeChanged if necessary.
}

const DefaultSize = -1

func (c *Context) SetSize(widget Widget, width, height int) {
	widget.widgetState().widthPlus1 = width + 1
	widget.widgetState().heightPlus1 = height + 1
}

func (c *Context) Size(widget Widget) (int, int) {
	widgetState := widget.widgetState()
	dw, dh := widget.DefaultSize(c)
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

func (c *Context) Bounds(widget Widget) image.Rectangle {
	widgetState := widget.widgetState()
	width, height := c.Size(widget)
	return image.Rectangle{
		Min: widgetState.position,
		Max: widgetState.position.Add(image.Point{width, height}),
	}
}

func (c *Context) VisibleBounds(widget Widget) image.Rectangle {
	parent := widget.widgetState().parent
	if parent == nil {
		return c.app.bounds()
	}
	if widget.ZDelta() != 0 {
		return c.Bounds(widget)
	}
	return c.VisibleBounds(parent).Intersect(c.Bounds(widget))
}

func (c *Context) Show(widget Widget) {
	widgetState := widget.widgetState()
	if !widgetState.hidden {
		return
	}
	widgetState.hidden = false
	RequestRedraw(widget)
}

func (c *Context) Hide(widget Widget) {
	widgetState := widget.widgetState()
	if widgetState.hidden {
		return
	}
	widgetState.hidden = true
	c.Blur(widget)
	RequestRedraw(widget)
}

func (c *Context) IsVisible(widget Widget) bool {
	return widget.widgetState().isVisible()
}

func (c *Context) Enable(widget Widget) {
	widgetState := widget.widgetState()
	if !widgetState.disabled {
		return
	}
	widgetState.disabled = false
	RequestRedraw(widget)
}

func (c *Context) Disable(widget Widget) {
	widgetState := widget.widgetState()
	if widgetState.disabled {
		return
	}
	widgetState.disabled = true
	c.Blur(widget)
	RequestRedraw(widget)
}

func (c *Context) IsEnabled(widget Widget) bool {
	return widget.widgetState().isEnabled()
}

func (c *Context) Focus(widget Widget) {
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
	if c.app.focusedWidget == widget {
		return
	}

	var oldWidget Widget
	if c.app.focusedWidget != nil {
		oldWidget = c.app.focusedWidget
	}

	c.app.focusedWidget = widget
	RequestRedraw(c.app.focusedWidget)
	if oldWidget != nil {
		RequestRedraw(oldWidget)
	}
}

func (c *Context) Blur(widget Widget) {
	widgetState := widget.widgetState()
	if !widgetState.isInTree() {
		return
	}
	if c.app.focusedWidget != widget {
		return
	}
	c.app.focusedWidget = nil
	RequestRedraw(widget)
}

func (c *Context) IsFocused(widget Widget) bool {
	widgetState := widget.widgetState()
	return widgetState.isInTree() && c.app.focusedWidget == widget && widgetState.isVisible()
}

func (c *Context) HasFocusedChildWidget(widget Widget) bool {
	widgetState := widget.widgetState()
	if c.IsFocused(widget) {
		return true
	}
	for _, child := range widgetState.children {
		if c.HasFocusedChildWidget(child) {
			return true
		}
	}
	return false
}

func (c *Context) Opacity(widget Widget) float64 {
	return widget.widgetState().opacity()
}

func (c *Context) SetOpacity(widget Widget, opacity float64) {
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

func (c *Context) IsWidgetHitAt(widget Widget, point image.Point) bool {
	return c.app.isWidgetHitAt(widget, point)
}
