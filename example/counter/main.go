// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/layout"
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

	r.counterText.SetSelectable(true)
	r.counterText.SetBold(true)
	r.counterText.SetHorizontalAlign(basicwidget.HorizontalAlignCenter)
	r.counterText.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
	r.counterText.SetScale(4)
	r.counterText.SetText(fmt.Sprintf("%d", r.counter))

	r.resetButton.SetText("Reset")
	r.resetButton.SetOnUp(func() {
		r.counter = 0
	})
	context.SetEnabled(&r.resetButton, r.counter != 0)

	r.incButton.SetText("Increment")
	r.incButton.SetOnUp(func() {
		r.counter++
	})

	r.decButton.SetText("Decrement")
	r.decButton.SetOnUp(func() {
		r.counter--
	})

	u := basicwidget.UnitSize(context)
	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(r).Inset(u),
		Heights: []layout.Size{
			layout.FlexibleSize(1),
			layout.FixedSize(u),
		},
		RowGap: u,
	}).CellBounds() {
		switch i {
		case 0:
			appender.AppendChildWidgetWithBounds(&r.counterText, bounds)
		case 1:
			for i, bounds := range (layout.GridLayout{
				Bounds: bounds,
				Widths: []layout.Size{
					layout.FixedSize(6 * u),
					layout.FlexibleSize(1),
					layout.FixedSize(6 * u),
					layout.FixedSize(6 * u),
				},
				ColumnGap: u / 2,
			}).CellBounds() {
				switch i {
				case 0:
					appender.AppendChildWidgetWithBounds(&r.resetButton, bounds)
				case 2:
					appender.AppendChildWidgetWithBounds(&r.incButton, bounds)
				case 3:
					appender.AppendChildWidgetWithBounds(&r.decButton, bounds)
				}
			}
		}
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
