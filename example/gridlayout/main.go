// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"fmt"
	"image"
	"os"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	_ "github.com/hajimehoshi/guigui/basicwidget/cjkfont"
	"github.com/hajimehoshi/guigui/layout"
)

type Root struct {
	guigui.RootWidget

	fill bool
	gap  bool

	configForm       basicwidget.Form
	fillText         basicwidget.Text
	fillToggleButton basicwidget.ToggleButton
	gapText          basicwidget.Text
	gapToggleButton  basicwidget.ToggleButton

	background basicwidget.Background
	buttons    [16]guigui.Widget
}

func (r *Root) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	appender.AppendChildWidgetWithBounds(&r.background, context.Bounds(r))

	r.fillText.SetText("Fill Widgets into Grid Cells")
	r.fillToggleButton.SetValue(r.fill)
	r.fillToggleButton.SetOnValueChanged(func(value bool) {
		r.fill = value
	})
	r.gapText.SetText("Use Gap")
	r.gapToggleButton.SetValue(r.gap)
	r.gapToggleButton.SetOnValueChanged(func(value bool) {
		r.gap = value
	})
	formItems := []*basicwidget.FormItem{
		{
			PrimaryWidget:   &r.fillText,
			SecondaryWidget: &r.fillToggleButton,
		},
		{
			PrimaryWidget:   &r.gapText,
			SecondaryWidget: &r.gapToggleButton,
		},
	}
	r.configForm.SetItems(formItems)

	u := basicwidget.UnitSize(context)
	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(r).Inset(int(u / 2)),
		Heights: []layout.Size{
			layout.MaxContentSize(func(index int) int {
				if index == 0 {
					return context.Size(&r.configForm).Y
				}
				return 0
			}),
			layout.FractionSize(1),
		},
		RowGap: int(u / 2),
	}).CellBounds() {
		if i == 0 {
			appender.AppendChildWidgetWithBounds(&r.configForm, bounds)
			continue
		}
		g := layout.GridLayout{
			Bounds: bounds,
			Widths: []layout.Size{
				layout.MaxContentSize(func(index int) int {
					return r.buttons[index].DefaultSize(context).X
				}),
				layout.FixedSize(200),
				layout.FractionSize(1),
				layout.FractionSize(2),
			},
			Heights: []layout.Size{
				layout.MaxContentSize(func(index int) int {
					return r.buttons[index].DefaultSize(context).Y
				}),
				layout.FixedSize(100),
				layout.FractionSize(1),
				layout.FractionSize(2),
			},
		}
		if r.gap {
			g.ColumnGap = int(u / 2)
			g.RowGap = int(u / 2)
		}
		for i := range r.buttons {
			if r.buttons[i] == nil {
				r.buttons[i] = &basicwidget.TextButton{}
			}
			t := r.buttons[i].(*basicwidget.TextButton)
			t.SetText(fmt.Sprintf("Button %d", i))
		}
		for i, bounds := range g.CellBounds() {
			widget := r.buttons[i]
			if r.fill {
				appender.AppendChildWidgetWithBounds(widget, bounds)
			} else {
				pt := bounds.Min
				s := widget.DefaultSize(context)
				pt.X += (bounds.Dx() - s.X) / 2
				pt.Y += (bounds.Dy() - s.Y) / 2
				appender.AppendChildWidgetWithBounds(widget, image.Rectangle{
					Min: pt,
					Max: pt.Add(s),
				})
			}
		}
	}
	return nil
}

func main() {
	op := &guigui.RunOptions{
		Title: "Grid Layout",
	}
	if err := guigui.Run(&Root{}, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
