// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Guigui Authors

package guigui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type Widget interface {
	Build(context *Context, appender *ChildWidgetAppender) error
	HandlePointingInput(context *Context) HandleInputResult
	HandleButtonInput(context *Context) HandleInputResult
	Tick(context *Context) error
	CursorShape(context *Context) (ebiten.CursorShapeType, bool)
	Draw(context *Context, dst *ebiten.Image)
	ZDelta() int
	DefaultSize(context *Context) image.Point
	PassThrough() bool

	widgetState() *widgetState
}

type HandleInputResult struct {
	widget  Widget
	aborted bool
}

func HandleInputByWidget(widget Widget) HandleInputResult {
	return HandleInputResult{
		widget: widget,
	}
}

func AbortHandlingInputByWidget(widget Widget) HandleInputResult {
	return HandleInputResult{
		aborted: true,
		widget:  widget,
	}
}

func (r *HandleInputResult) shouldRaise() bool {
	return r.widget != nil || r.aborted
}
