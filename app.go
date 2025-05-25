// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Guigui Authors

package guigui

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"maps"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/hajimehoshi/oklab"
)

type debugMode struct {
	showRenderingRegions bool
	showInputLogs        bool
	deviceScale          float64
}

var theDebugMode debugMode

func init() {
	for _, token := range strings.Split(os.Getenv("GUIGUI_DEBUG"), ",") {
		switch {
		case token == "showrenderingregions":
			theDebugMode.showRenderingRegions = true
		case token == "showinputlogs":
			theDebugMode.showInputLogs = true
		case strings.HasPrefix(token, "devicescale="):
			f, err := strconv.ParseFloat(token[len("devicescale="):], 64)
			if err != nil {
				slog.Error(err.Error())
			}
			theDebugMode.deviceScale = f
		}
	}
}

type invalidatedRegionsForDebugItem struct {
	region image.Rectangle
	time   int
}

func invalidatedRegionForDebugMaxTime() int {
	return ebiten.TPS() / 5
}

type app struct {
	root       Widget
	context    Context
	visitedZs  map[int]struct{}
	zs         []int
	hitWidgets []Widget

	invalidatedRegions image.Rectangle

	invalidatedRegionsForDebug []invalidatedRegionsForDebugItem

	screenWidth  float64
	screenHeight float64
	deviceScale  float64

	lastScreenWidth  float64
	lastScreenHeight float64

	focusedWidgetState *widgetState

	offscreen   *ebiten.Image
	debugScreen *ebiten.Image
}

type RunOptions struct {
	Title         string
	WindowSize    image.Point
	WindowMinSize image.Point
	WindowMaxSize image.Point
	AppScale      float64

	RunGameOptions *ebiten.RunGameOptions
}

func Run(root Widget, options *RunOptions) error {
	return RunWithCustomFunc(root, options, ebiten.RunGameWithOptions)
}

func RunWithCustomFunc(root Widget, options *RunOptions, f func(game ebiten.Game, options *ebiten.RunGameOptions) error) error {
	if options == nil {
		options = &RunOptions{}
	}

	ebiten.SetWindowTitle(options.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetScreenClearedEveryFrame(false)
	if options.WindowSize.X > 0 && options.WindowSize.Y > 0 {
		ebiten.SetWindowSize(options.WindowSize.X, options.WindowSize.Y)
	}
	minW := -1
	minH := -1
	maxW := -1
	maxH := -1
	if options.WindowMinSize.X > 0 {
		minW = options.WindowMinSize.X
	}
	if options.WindowMinSize.Y > 0 {
		minH = options.WindowMinSize.Y
	}
	if options.WindowMaxSize.X > 0 {
		maxW = options.WindowMaxSize.X
	}
	if options.WindowMaxSize.Y > 0 {
		maxH = options.WindowMaxSize.Y
	}
	ebiten.SetWindowSizeLimits(minW, minH, maxW, maxH)

	a := &app{
		root:        root,
		deviceScale: deviceScaleFactor(),
	}
	a.root.widgetState().root = true
	a.context.app = a
	if options.AppScale > 0 {
		a.context.appScaleMinus1 = options.AppScale - 1
	}

	var eop ebiten.RunGameOptions
	if options.RunGameOptions != nil {
		eop = *options.RunGameOptions
	}
	// Prefer SRGB for consistent result.
	if eop.ColorSpace == ebiten.ColorSpaceDefault {
		eop.ColorSpace = ebiten.ColorSpaceSRGB
	}

	return f(a, &eop)
}

func deviceScaleFactor() float64 {
	if theDebugMode.deviceScale != 0 {
		return theDebugMode.deviceScale
	}
	// Calling ebiten.Monitor() seems pretty expensive. Do not call this often.
	// TODO: Ebitengine should be fixed.
	return ebiten.Monitor().DeviceScaleFactor()
}

func (a *app) bounds() image.Rectangle {
	return image.Rect(0, 0, int(math.Ceil(a.screenWidth)), int(math.Ceil(a.screenHeight)))
}

func (a *app) Update() error {
	if a.focusedWidgetState == nil {
		a.focusedWidgetState = a.root.widgetState()
	}

	rootState := a.root.widgetState()
	rootState.position = image.Point{}

	if s := deviceScaleFactor(); a.deviceScale != s {
		a.deviceScale = s
		a.requestRedraw(a.bounds())
	}

	// Construct the widget tree.
	a.context.inBuild = true
	if err := a.build(); err != nil {
		return err
	}
	a.context.inBuild = false

	// Handle user inputs.
	// TODO: Handle this in Ebitengine's HandleInput in the future (hajimehoshi/ebiten#1704)
	if r := a.handleInputWidget(handleInputTypePointing); r.widget != nil {
		if theDebugMode.showInputLogs {
			slog.Info("pointing input handled", "widget", fmt.Sprintf("%T", r.widget), "aborted", r.aborted)
		}
	}
	if r := a.handleInputWidget(handleInputTypeButton); r.widget != nil {
		if theDebugMode.showInputLogs {
			slog.Info("keyboard input handled", "widget", fmt.Sprintf("%T", r.widget), "aborted", r.aborted)
		}
	}

	// Construct the widget tree again to reflect the latest state.
	a.context.inBuild = true
	if err := a.build(); err != nil {
		return err
	}
	a.context.inBuild = false

	if !a.cursorShape() {
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	}

	// Update
	if err := a.updateWidget(a.root); err != nil {
		return err
	}

	// Invalidate the engire screen if the screen size is changed.
	var invalidated bool
	if a.lastScreenWidth != a.screenWidth {
		invalidated = true
		a.lastScreenWidth = a.screenWidth
	}
	if a.lastScreenHeight != a.screenHeight {
		invalidated = true
		a.lastScreenHeight = a.screenHeight
	}
	if invalidated {
		a.requestRedraw(a.bounds())
	} else {
		// Invalidate regions if a widget's children state is changed.
		// A widget's bounds might be changed in Update, so do this after updating.
		a.requestRedrawIfTreeChanged(a.root)
	}

	a.resetPrevWidgets(a.root)

	// Resolve dirty widgets.
	_ = traverseWidget(a.root, func(widget Widget) error {
		if !widget.widgetState().dirty {
			return nil
		}
		vb := a.context.VisibleBounds(widget)
		if vb.Empty() {
			return nil
		}
		if theDebugMode.showRenderingRegions {
			slog.Info("request redrawing", "requester", fmt.Sprintf("%T", widget), "at", widget.widgetState().dirtyAt, "region", vb)
		}
		a.requestRedrawWidget(widget)
		widget.widgetState().dirty = false
		widget.widgetState().dirtyAt = ""
		return nil
	})

	if theDebugMode.showRenderingRegions {
		// Update the regions in the reversed order to remove items.
		for idx := len(a.invalidatedRegionsForDebug) - 1; idx >= 0; idx-- {
			if a.invalidatedRegionsForDebug[idx].time > 0 {
				a.invalidatedRegionsForDebug[idx].time--
			} else {
				a.invalidatedRegionsForDebug = slices.Delete(a.invalidatedRegionsForDebug, idx, idx+1)
			}
		}

		if !a.invalidatedRegions.Empty() {
			idx := slices.IndexFunc(a.invalidatedRegionsForDebug, func(i invalidatedRegionsForDebugItem) bool {
				return i.region.Eq(a.invalidatedRegions)
			})
			if idx < 0 {
				a.invalidatedRegionsForDebug = append(a.invalidatedRegionsForDebug, invalidatedRegionsForDebugItem{
					region: a.invalidatedRegions,
					time:   invalidatedRegionForDebugMaxTime(),
				})
			} else {
				a.invalidatedRegionsForDebug[idx].time = invalidatedRegionForDebugMaxTime()
			}
		}
	}

	return nil
}

func (a *app) Draw(screen *ebiten.Image) {
	origScreen := screen
	if theDebugMode.showRenderingRegions {
		if a.offscreen != nil {
			if a.offscreen.Bounds().Dx() != screen.Bounds().Dx() || a.offscreen.Bounds().Dy() != screen.Bounds().Dy() {
				a.offscreen.Deallocate()
				a.offscreen = nil
			}
		}
		if a.offscreen == nil {
			a.offscreen = ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
		}
		screen = a.offscreen
	}
	a.drawWidget(screen)
	a.drawDebugIfNeeded(origScreen)
	a.invalidatedRegions = image.Rectangle{}
}

func (a *app) Layout(outsideWidth, outsideHeight int) (int, int) {
	panic("guigui: game.Layout should never be called")
}

func (a *app) LayoutF(outsideWidth, outsideHeight float64) (float64, float64) {
	s := a.deviceScale
	a.screenWidth = outsideWidth * s
	a.screenHeight = outsideHeight * s
	return a.screenWidth, a.screenHeight
}

func (a *app) requestRedraw(region image.Rectangle) {
	a.invalidatedRegions = a.invalidatedRegions.Union(region)
}

func (a *app) requestRedrawWidget(widget Widget) {
	a.requestRedraw(a.context.VisibleBounds(widget))
	for _, child := range widget.widgetState().children {
		a.requestRedrawIfDifferentParentZ(child)
	}
}

func (a *app) requestRedrawIfDifferentParentZ(widget Widget) {
	if widget.ZDelta() != 0 {
		a.requestRedrawWidget(widget)
		return
	}
	for _, child := range widget.widgetState().children {
		a.requestRedrawIfDifferentParentZ(child)
	}
}

func (a *app) build() error {
	clear(a.visitedZs)
	if a.visitedZs == nil {
		a.visitedZs = map[int]struct{}{}
	}

	var appender ChildWidgetAppender
	if err := traverseWidget(a.root, func(widget Widget) error {
		widgetState := widget.widgetState()

		// Reset the current children before building.
		for _, child := range widgetState.children {
			child.widgetState().parent = nil
		}

		if parent := widgetState.parent; parent != nil {
			widgetState.z = parent.widgetState().z + widget.ZDelta()
		} else {
			widgetState.z = 0
		}
		widgetState.hasVisibleBoundsCache = false
		widgetState.visibleBoundsCache = image.Rectangle{}

		widgetState.children = slices.Delete(widgetState.children, 0, len(widgetState.children))
		appender.app = a
		appender.widget = widget
		if err := widget.Build(&a.context, &appender); err != nil {
			return err
		}

		a.visitedZs[widgetState.z] = struct{}{}

		return nil
	}); err != nil {
		return err
	}

	a.zs = slices.Delete(a.zs, 0, len(a.zs))
	a.zs = slices.AppendSeq(a.zs, maps.Keys(a.visitedZs))
	slices.Sort(a.zs)

	a.hitWidgets = slices.Delete(a.hitWidgets, 0, len(a.hitWidgets))
	pt := image.Pt(ebiten.CursorPosition())
	a.hitWidgets = a.appendWidgetsAt(a.hitWidgets, pt, a.root, true)
	slices.SortStableFunc(a.hitWidgets, func(a, b Widget) int {
		return b.widgetState().z - a.widgetState().z
	})

	return nil
}

type handleInputType int

const (
	handleInputTypePointing handleInputType = iota
	handleInputTypeButton
)

func (a *app) handleInputWidget(typ handleInputType) HandleInputResult {
	for i := len(a.zs) - 1; i >= 0; i-- {
		z := a.zs[i]
		if r := a.doHandleInputWidget(typ, a.root, z); r.shouldRaise() {
			return r
		}
	}
	return HandleInputResult{}
}

func (a *app) doHandleInputWidget(typ handleInputType, widget Widget, zToHandle int) HandleInputResult {
	if widget.PassThrough() {
		return HandleInputResult{}
	}

	// Avoid (*Context).IsVisible and (*Context).IsEnabled for performance.
	// These check parent widget states unnecessarily.

	if widget.widgetState().hidden {
		return HandleInputResult{}
	}

	if widget.widgetState().disabled {
		return HandleInputResult{}
	}

	if typ == handleInputTypeButton && !a.context.IsFocusedOrHasFocusedChild(widget) {
		return HandleInputResult{}
	}

	widgetState := widget.widgetState()
	// Iterate the children in the reverse order of rendering.
	for i := len(widgetState.children) - 1; i >= 0; i-- {
		child := widgetState.children[i]
		if r := a.doHandleInputWidget(typ, child, zToHandle); r.shouldRaise() {
			return r
		}
	}

	if zToHandle != widget.widgetState().z {
		return HandleInputResult{}
	}

	switch typ {
	case handleInputTypePointing:
		return widget.HandlePointingInput(&a.context)
	case handleInputTypeButton:
		return widget.HandleButtonInput(&a.context)
	default:
		panic(fmt.Sprintf("guigui: unknown handleInputType: %d", typ))
	}
}

func (a *app) cursorShape() bool {
	var firstZ int
	for i, widget := range a.hitWidgets {
		if i == 0 {
			firstZ = widget.widgetState().z
		}
		if widget.widgetState().z < firstZ {
			break
		}
		if !widget.widgetState().isEnabled() {
			return false
		}
		shape, ok := widget.CursorShape(&a.context)
		if !ok {
			continue
		}
		ebiten.SetCursorShape(shape)
		return true
	}
	return false
}

func (a *app) updateWidget(widget Widget) error {
	widgetState := widget.widgetState()
	if err := widget.Tick(&a.context); err != nil {
		return err
	}

	for _, child := range widgetState.children {
		if err := a.updateWidget(child); err != nil {
			return err
		}
	}

	return nil
}

func (a *app) requestRedrawIfTreeChanged(widget Widget) {
	widgetState := widget.widgetState()
	// If the children and/or children's bounds are changed, request redraw.
	if !widgetState.prev.equals(&a.context, widgetState.children) {
		a.requestRedraw(a.context.VisibleBounds(widget))

		// Widgets with different Z from their parent's Z (e.g. popups) are outside of widget, so redraw the regions explicitly.
		widgetState.prev.redrawIfDifferentParentZ(a)
		for _, child := range widgetState.children {
			if child.ZDelta() != 0 {
				a.requestRedraw(a.context.VisibleBounds(child))
			}
		}
	}
	for _, child := range widgetState.children {
		a.requestRedrawIfTreeChanged(child)
	}
}

func (a *app) resetPrevWidgets(widget Widget) {
	widgetState := widget.widgetState()
	// Reset the states.
	widgetState.prev.reset()
	for _, child := range widgetState.children {
		widgetState.prev.append(&a.context, child)
	}
	for _, child := range widgetState.children {
		a.resetPrevWidgets(child)
	}
}

func (a *app) drawWidget(screen *ebiten.Image) {
	if a.invalidatedRegions.Empty() {
		return
	}
	dst := screen.SubImage(a.invalidatedRegions).(*ebiten.Image)
	for _, z := range a.zs {
		a.doDrawWidget(dst, a.root, z)
	}
}

func (a *app) doDrawWidget(dst *ebiten.Image, widget Widget, zToRender int) {
	// Do not skip this even when visible bounds are empty.
	// A child widget might have a different Z value and different visible bounds.

	widgetState := widget.widgetState()
	if widgetState.hidden {
		return
	}
	if widgetState.opacity() == 0 {
		return
	}

	customDraw := widgetState.customDraw
	useOffscreen := widgetState.opacity() < 1 || customDraw != nil

	vb := a.context.VisibleBounds(widget)
	var origDst *ebiten.Image
	renderCurrent := zToRender == widget.widgetState().z && !vb.Empty()
	if renderCurrent {
		if useOffscreen {
			origDst = dst
			dst = widgetState.ensureOffscreen(dst.Bounds())
			dst.Clear()
		}
		widget.Draw(&a.context, dst.SubImage(vb).(*ebiten.Image))
	}

	for _, child := range widgetState.children {
		a.doDrawWidget(dst, child, zToRender)
	}

	if renderCurrent {
		if useOffscreen {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(dst.Bounds().Min.X), float64(dst.Bounds().Min.Y))
			op.ColorScale.ScaleAlpha(float32(widgetState.opacity()))
			if customDraw != nil {
				customDraw(origDst.SubImage(vb).(*ebiten.Image), dst, op)
			} else {
				origDst.DrawImage(dst, op)
			}
		}
	}
}

func (a *app) drawDebugIfNeeded(screen *ebiten.Image) {
	if !theDebugMode.showRenderingRegions {
		return
	}

	if a.debugScreen != nil {
		if a.debugScreen.Bounds().Dx() != screen.Bounds().Dx() || a.debugScreen.Bounds().Dy() != screen.Bounds().Dy() {
			a.debugScreen.Deallocate()
			a.debugScreen = nil
		}
	}
	if a.debugScreen == nil {
		a.debugScreen = ebiten.NewImage(screen.Bounds().Dx(), screen.Bounds().Dy())
	}

	a.debugScreen.Clear()
	for _, item := range a.invalidatedRegionsForDebug {
		clr := oklab.OklchModel.Convert(color.RGBA{R: 0xff, G: 0x4b, B: 0x00, A: 0xff}).(oklab.Oklch)
		clr.Alpha = float64(item.time) / float64(invalidatedRegionForDebugMaxTime())
		if clr.Alpha > 0 {
			w := float32(4 * a.context.Scale())
			vector.StrokeRect(a.debugScreen, float32(item.region.Min.X)+w/2, float32(item.region.Min.Y)+w/2, float32(item.region.Dx())-w, float32(item.region.Dy())-w, w, clr, false)
		}
	}
	op := &ebiten.DrawImageOptions{}
	op.Blend = ebiten.BlendCopy
	screen.DrawImage(a.offscreen, op)
	screen.DrawImage(a.debugScreen, nil)
}

func (a *app) isWidgetHitAt(widget Widget) bool {
	if !widget.widgetState().isInTree() {
		return false
	}
	// hitWidgets are ordered by descending z values.
	// Always use a fixed set hitWidgets, as the tree might be dynamically changed during Build.
	for _, w := range a.hitWidgets {
		z1 := w.widgetState().z
		z2 := widget.widgetState().z
		if z1 > z2 {
			// w overlaps widget at point.
			return false
		}
		if z1 < z2 {
			// The same z value no longer exists.
			return false
		}
		if w.widgetState() == widget.widgetState() {
			return true
		}
	}
	return false
}

func (a *app) appendWidgetsAt(widgets []Widget, point image.Point, widget Widget, parentHit bool) []Widget {
	// Avoid (*widgetState).isVisible for performance.
	// These check parent widget states unnecessarily.
	if widget.widgetState().hidden {
		return widgets
	}
	if widget.PassThrough() {
		return widgets
	}

	var hit bool
	if parentHit || widget.ZDelta() != 0 {
		hit = point.In(a.context.VisibleBounds(widget))
	}

	children := widget.widgetState().children
	for i := len(children) - 1; i >= 0; i-- {
		child := children[i]
		widgets = a.appendWidgetsAt(widgets, point, child, hit)
	}

	if !hit {
		return widgets
	}

	widgets = append(widgets, widget)
	return widgets
}
