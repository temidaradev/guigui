// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/layout"
)

type Lists struct {
	guigui.DefaultWidget

	form         basicwidget.Form
	textListText basicwidget.Text
	textList     basicwidget.TextList[int]
}

func (l *Lists) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	l.textListText.SetText("Text List")
	var items []basicwidget.TextListItem[int]
	for i := 0; i < 100; i++ {
		items = append(items, basicwidget.TextListItem[int]{
			Text: fmt.Sprintf("Item %d", i),
		})
	}
	l.textList.SetItems(items)
	context.SetSize(&l.textList, image.Pt(guigui.DefaultSize, 6*basicwidget.UnitSize(context)))

	l.form.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &l.textListText,
			SecondaryWidget: &l.textList,
		},
	})

	u := basicwidget.UnitSize(context)
	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(l).Inset(u / 2),
		Heights: []layout.Size{
			layout.LazySize(func(row int) layout.Size {
				if row >= 1 {
					return layout.FixedSize(0)
				}
				return layout.FixedSize(l.form.DefaultSize(context).Y)
			}),
		},
		RowGap: u / 2,
	}).RepeatingCellBounds() {
		if i >= 1 {
			break
		}
		appender.AppendChildWidgetWithBounds(&l.form, bounds)
	}

	return nil
}
