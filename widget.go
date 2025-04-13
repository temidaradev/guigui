// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package guigui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Widget interface {
	Build(context *Context, appender *ChildWidgetAppender) error
	HandlePointingInput(context *Context) HandleInputResult
	HandleButtonInput(context *Context) HandleInputResult
	Update(context *Context) error
	CursorShape(context *Context) (ebiten.CursorShapeType, bool)
	Draw(context *Context, dst *ebiten.Image)
	Z() int
	DefaultSize(context *Context) (int, int)

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

func Parent(widget Widget) Widget {
	return widget.widgetState().parent
}
