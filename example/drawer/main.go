// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Guigui Authors

package main

import (
	"fmt"
	"image"
	"os"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/layout"
)

const dummyText = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

type Root struct {
	guigui.DefaultWidget

	background   basicwidget.Background
	toolbar      Toolbar
	leftPanel    LeftPanel
	contentPanel ContentPanel
	rightPanel   RightPanel

	model Model
}

func (r *Root) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	appender.AppendChildWidgetWithBounds(&r.background, context.Bounds(r))

	r.toolbar.SetModel(&r.model)

	gl := layout.GridLayout{
		Bounds: context.Bounds(r),
		Heights: []layout.Size{
			layout.FixedSize(r.toolbar.DefaultSize(context).Y),
			layout.FlexibleSize(1),
		},
	}
	appender.AppendChildWidgetWithBounds(&r.toolbar, gl.CellBounds(0, 0))

	u := basicwidget.UnitSize(context)
	var leftWidth int
	if r.model.IsLeftPanelOpen() {
		leftWidth = 8 * u
	}
	var rightWidth int
	if r.model.IsRightPanelOpen() {
		rightWidth = 8 * u
	}
	contentGL := layout.GridLayout{
		Bounds: gl.CellBounds(0, 1),
		Widths: []layout.Size{
			layout.FixedSize(leftWidth),
			layout.FlexibleSize(1),
			layout.FixedSize(rightWidth),
		},
	}
	appender.AppendChildWidgetWithBounds(&r.leftPanel, contentGL.CellBounds(0, 0))
	appender.AppendChildWidgetWithBounds(&r.contentPanel, contentGL.CellBounds(1, 0))
	appender.AppendChildWidgetWithBounds(&r.rightPanel, contentGL.CellBounds(2, 0))

	return nil
}

func main() {
	op := &guigui.RunOptions{
		Title:      "Drawers",
		WindowSize: image.Pt(800, 600),
		RunGameOptions: &ebiten.RunGameOptions{
			ApplePressAndHoldEnabled: true,
		},
	}
	if err := guigui.Run(&Root{}, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
