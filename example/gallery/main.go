// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	_ "github.com/hajimehoshi/guigui/basicwidget/cjkfont"
)

type Root struct {
	guigui.DefaultWidget

	background basicwidget.Background
	sidebar    Sidebar
	settings   Settings
	basic      Basic
	buttons    Buttons
	lists      Lists
	popups     Popups
}

func (r *Root) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	appender.AppendChildWidget(&r.background)
	appender.AppendChildWidget(&r.sidebar)

	sw := 8 * basicwidget.UnitSize(context)
	rw, rh := context.Size(r)
	context.SetSize(&r.sidebar, sw, rh)
	context.SetPosition(&r.sidebar, context.Position(r))
	p := context.Position(r)
	p.X += sw
	pw := rw - sw
	context.SetPosition(&r.settings, p)
	context.SetSize(&r.settings, pw, rh)
	context.SetPosition(&r.basic, p)
	context.SetSize(&r.basic, pw, rh)
	context.SetPosition(&r.buttons, p)
	context.SetSize(&r.buttons, pw, rh)
	context.SetPosition(&r.lists, p)
	context.SetSize(&r.lists, pw, rh)
	context.SetPosition(&r.popups, p)
	context.SetSize(&r.popups, pw, rh)

	switch r.sidebar.SelectedItemTag() {
	case "settings":
		appender.AppendChildWidget(&r.settings)
	case "basic":
		appender.AppendChildWidget(&r.basic)
	case "buttons":
		appender.AppendChildWidget(&r.buttons)
	case "lists":
		appender.AppendChildWidget(&r.lists)
	case "popups":
		appender.AppendChildWidget(&r.popups)
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
