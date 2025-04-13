// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package guigui

import "github.com/hajimehoshi/ebiten/v2"

type DefaultWidget struct {
	widgetState_ widgetState
}

var _ Widget = (*DefaultWidget)(nil)

func (*DefaultWidget) Build(context *Context, appender *ChildWidgetAppender) error {
	return nil
}

func (*DefaultWidget) HandlePointingInput(context *Context) HandleInputResult {
	return HandleInputResult{}
}

func (*DefaultWidget) HandleButtonInput(context *Context) HandleInputResult {
	return HandleInputResult{}
}

func (*DefaultWidget) Update(context *Context) error {
	return nil
}

func (*DefaultWidget) CursorShape(context *Context) (ebiten.CursorShapeType, bool) {
	return 0, false
}

func (*DefaultWidget) Draw(context *Context, dst *ebiten.Image) {
}

func (d *DefaultWidget) Z() int {
	if d.widgetState_.parent == nil {
		return 0
	}
	return d.widgetState_.parent.Z()
}

func (d *DefaultWidget) DefaultSize(context *Context) (int, int) {
	if d.widgetState_.parent == nil {
		return context.app.bounds().Dx(), context.app.bounds().Dy()
	}
	return Size(d.widgetState_.parent)
}

func (d *DefaultWidget) widgetState() *widgetState {
	return &d.widgetState_
}
