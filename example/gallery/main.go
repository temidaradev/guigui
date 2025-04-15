// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"fmt"
	"image"
	"os"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	_ "github.com/hajimehoshi/guigui/basicwidget/cjkfont"
)

type Root struct {
	guigui.RootWidget

	background basicwidget.Background
	sidebar    Sidebar
	settings   Settings
	basic      Basic
	buttons    Buttons
	lists      Lists
	popups     Popups
}

func (r *Root) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	appender.AppendChildWidgetWithBounds(&r.background, context.Bounds(r))

	rw, rh := context.Size(r)
	sw := 8 * basicwidget.UnitSize(context)
	appender.AppendChildWidgetWithBounds(&r.sidebar, image.Rectangle{
		Min: context.Position(r),
		Max: context.Position(r).Add(image.Pt(sw, rh)),
	})
	p := context.Position(r)
	p.X += sw
	pw := rw - sw
	contentBounds := image.Rectangle{
		Min: p,
		Max: p.Add(image.Pt(pw, rh)),
	}

	switch r.sidebar.SelectedItemTag() {
	case "settings":
		appender.AppendChildWidgetWithBounds(&r.settings, contentBounds)
	case "basic":
		appender.AppendChildWidgetWithBounds(&r.basic, contentBounds)
	case "buttons":
		appender.AppendChildWidgetWithBounds(&r.buttons, contentBounds)
	case "lists":
		appender.AppendChildWidgetWithBounds(&r.lists, contentBounds)
	case "popups":
		appender.AppendChildWidgetWithBounds(&r.popups, contentBounds)
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
