// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Root struct {
	guigui.DefaultWidget

	background  basicwidget.Background
	resetButton basicwidget.TextButton
	incButton   basicwidget.TextButton
	decButton   basicwidget.TextButton
	counterText basicwidget.Text

	counter int
}

func (r *Root) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	appender.AppendChildWidget(&r.background)

	{
		w, h := guigui.Size(r)
		w -= 2 * basicwidget.UnitSize(context)
		h -= 4 * basicwidget.UnitSize(context)
		guigui.SetSize(&r.counterText, w, h)

		r.counterText.SetSelectable(true)
		r.counterText.SetBold(true)
		r.counterText.SetHorizontalAlign(basicwidget.HorizontalAlignCenter)
		r.counterText.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
		r.counterText.SetScale(4)
		r.counterText.SetText(fmt.Sprintf("%d", r.counter))

		p := guigui.Position(r)
		p.X += basicwidget.UnitSize(context)
		p.Y += basicwidget.UnitSize(context)
		guigui.SetPosition(&r.counterText, p)
		appender.AppendChildWidget(&r.counterText)
	}

	r.resetButton.SetText("Reset")
	guigui.SetSize(&r.resetButton, 6*basicwidget.UnitSize(context), guigui.AutoSize)
	r.resetButton.SetOnUp(func() {
		r.counter = 0
	})
	if r.counter == 0 {
		guigui.Disable(&r.resetButton)
	} else {
		guigui.Enable(&r.resetButton)
	}
	{
		p := guigui.Position(r)
		_, h := guigui.Size(r)
		p.X += basicwidget.UnitSize(context)
		p.Y += h - 2*basicwidget.UnitSize(context)
		guigui.SetPosition(&r.resetButton, p)
		appender.AppendChildWidget(&r.resetButton)
	}

	r.incButton.SetText("Increment")
	guigui.SetSize(&r.incButton, 6*basicwidget.UnitSize(context), guigui.AutoSize)
	r.incButton.SetOnUp(func() {
		r.counter++
	})
	{
		p := guigui.Position(r)
		w, h := guigui.Size(r)
		p.X += w - 7*basicwidget.UnitSize(context)
		p.Y += h - 2*basicwidget.UnitSize(context)
		guigui.SetPosition(&r.incButton, p)
		appender.AppendChildWidget(&r.incButton)
	}

	r.decButton.SetText("Decrement")
	guigui.SetSize(&r.decButton, 6*basicwidget.UnitSize(context), guigui.AutoSize)
	r.decButton.SetOnUp(func() {
		r.counter--
	})
	{
		p := guigui.Position(r)
		w, h := guigui.Size(r)
		p.X += w - int(13.5*float64(basicwidget.UnitSize(context)))
		p.Y += h - 2*basicwidget.UnitSize(context)
		guigui.SetPosition(&r.decButton, p)
		appender.AppendChildWidget(&r.decButton)
	}

	return nil
}

func main() {
	op := &guigui.RunOptions{
		Title:           "Counter",
		WindowMinWidth:  600,
		WindowMinHeight: 300,
	}
	if err := guigui.Run(&Root{}, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
