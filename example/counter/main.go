// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"fmt"
	"image"
	"os"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Root struct {
	guigui.RootWidget

	background  basicwidget.Background
	resetButton basicwidget.TextButton
	incButton   basicwidget.TextButton
	decButton   basicwidget.TextButton
	counterText basicwidget.Text

	counter int
}

func (r *Root) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	appender.AppendChildWidgetWithBounds(&r.background, context.Bounds(r))

	{
		s := context.Size(r)
		s.X -= 2 * basicwidget.UnitSize(context)
		s.Y -= 4 * basicwidget.UnitSize(context)

		r.counterText.SetSelectable(true)
		r.counterText.SetBold(true)
		r.counterText.SetHorizontalAlign(basicwidget.HorizontalAlignCenter)
		r.counterText.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
		r.counterText.SetScale(4)
		r.counterText.SetText(fmt.Sprintf("%d", r.counter))

		p := context.Position(r)
		p.X += basicwidget.UnitSize(context)
		p.Y += basicwidget.UnitSize(context)
		appender.AppendChildWidgetWithBounds(&r.counterText, image.Rectangle{
			Min: p,
			Max: p.Add(s),
		})
	}

	r.resetButton.SetText("Reset")
	context.SetSize(&r.resetButton, image.Pt(6*basicwidget.UnitSize(context), guigui.DefaultSize))
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
		p.X += basicwidget.UnitSize(context)
		p.Y += context.Size(r).Y - 2*basicwidget.UnitSize(context)
		appender.AppendChildWidgetWithPosition(&r.resetButton, p)
	}

	r.incButton.SetText("Increment")
	context.SetSize(&r.incButton, image.Pt(6*basicwidget.UnitSize(context), guigui.DefaultSize))
	r.incButton.SetOnUp(func() {
		r.counter++
	})
	{
		p := context.Position(r)
		p.X += context.Size(r).X - 7*basicwidget.UnitSize(context)
		p.Y += context.Size(r).Y - 2*basicwidget.UnitSize(context)
		appender.AppendChildWidgetWithPosition(&r.incButton, p)
	}

	r.decButton.SetText("Decrement")
	context.SetSize(&r.decButton, image.Pt(6*basicwidget.UnitSize(context), guigui.DefaultSize))
	r.decButton.SetOnUp(func() {
		r.counter--
	})
	{
		p := context.Position(r)
		p.X += context.Size(r).X - int(13.5*float64(basicwidget.UnitSize(context)))
		p.Y += context.Size(r).Y - 2*basicwidget.UnitSize(context)
		appender.AppendChildWidgetWithPosition(&r.decButton, p)
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
