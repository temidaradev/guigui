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
		w, h := context.Size(r)
		w -= 2 * basicwidget.UnitSize(context)
		h -= 4 * basicwidget.UnitSize(context)
		context.SetSize(&r.counterText, w, h)

		r.counterText.SetSelectable(context, true)
		r.counterText.SetBold(context, true)
		r.counterText.SetHorizontalAlign(context, basicwidget.HorizontalAlignCenter)
		r.counterText.SetVerticalAlign(context, basicwidget.VerticalAlignMiddle)
		r.counterText.SetScale(context, 4)
		r.counterText.SetText(context, fmt.Sprintf("%d", r.counter))

		p := context.Position(r)
		p.X += basicwidget.UnitSize(context)
		p.Y += basicwidget.UnitSize(context)
		context.SetPosition(&r.counterText, p)
		appender.AppendChildWidget(&r.counterText)
	}

	r.resetButton.SetText(context, "Reset")
	context.SetSize(&r.resetButton, 6*basicwidget.UnitSize(context), guigui.AutoSize)
	r.resetButton.SetOnUp(func() {
		r.counter = 0
	})
	if r.counter == 0 {
		context.Disable(&r.resetButton)
	} else {
		context.Enable(&r.resetButton)
	}
	{
		p := context.Position(r)
		_, h := context.Size(r)
		p.X += basicwidget.UnitSize(context)
		p.Y += h - 2*basicwidget.UnitSize(context)
		context.SetPosition(&r.resetButton, p)
		appender.AppendChildWidget(&r.resetButton)
	}

	r.incButton.SetText(context, "Increment")
	context.SetSize(&r.incButton, 6*basicwidget.UnitSize(context), guigui.AutoSize)
	r.incButton.SetOnUp(func() {
		r.counter++
	})
	{
		p := context.Position(r)
		w, h := context.Size(r)
		p.X += w - 7*basicwidget.UnitSize(context)
		p.Y += h - 2*basicwidget.UnitSize(context)
		context.SetPosition(&r.incButton, p)
		appender.AppendChildWidget(&r.incButton)
	}

	r.decButton.SetText(context, "Decrement")
	context.SetSize(&r.decButton, 6*basicwidget.UnitSize(context), guigui.AutoSize)
	r.decButton.SetOnUp(func() {
		r.counter--
	})
	{
		p := context.Position(r)
		w, h := context.Size(r)
		p.X += w - int(13.5*float64(basicwidget.UnitSize(context)))
		p.Y += h - 2*basicwidget.UnitSize(context)
		context.SetPosition(&r.decButton, p)
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
