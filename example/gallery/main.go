// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	_ "github.com/hajimehoshi/guigui/basicwidget/cjkfont"
	"github.com/hajimehoshi/guigui/layout"
)

type Root struct {
	guigui.RootWidget

	background basicwidget.Background
	sidebar    Sidebar
	settings   Settings
	basic      Basic
	buttons    Buttons
	texts      Texts
	textFields TextFields
	lists      Lists
	popups     Popups

	model Model
}

func (r *Root) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	appender.AppendChildWidgetWithBounds(&r.background, context.Bounds(r))

	r.texts.SetModel(&r.model)
	r.textFields.SetModel(&r.model)
	r.sidebar.SetModel(&r.model)

	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(r),
		Widths: []layout.Size{
			layout.FixedSize(8 * basicwidget.UnitSize(context)),
			layout.FlexibleSize(1),
		},
	}).CellBounds() {
		switch i {
		case 0:
			appender.AppendChildWidgetWithBounds(&r.sidebar, bounds)
		case 1:
			switch r.model.Mode() {
			case "settings":
				appender.AppendChildWidgetWithBounds(&r.settings, bounds)
			case "basic":
				appender.AppendChildWidgetWithBounds(&r.basic, bounds)
			case "buttons":
				appender.AppendChildWidgetWithBounds(&r.buttons, bounds)
			case "texts":
				appender.AppendChildWidgetWithBounds(&r.texts, bounds)
			case "textfields":
				appender.AppendChildWidgetWithBounds(&r.textFields, bounds)
			case "lists":
				appender.AppendChildWidgetWithBounds(&r.lists, bounds)
			case "popups":
				appender.AppendChildWidgetWithBounds(&r.popups, bounds)
			}
		}
	}

	return nil
}

func main() {
	op := &guigui.RunOptions{
		Title: "Component Gallery",
	}
	if err := guigui.Run(&Root{}, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
