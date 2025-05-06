// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Guigui Authors

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

	listForm     basicwidget.Form
	textListText basicwidget.Text
	textList     basicwidget.TextList[int]

	configForm    basicwidget.Form
	enabledText   basicwidget.Text
	enabledToggle basicwidget.Toggle

	model *Model
}

func (l *Lists) SetModel(model *Model) {
	l.model = model
}

func (l *Lists) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	// Lists
	l.textListText.SetValue("Text List")
	var items []basicwidget.TextListItem[int]
	for i := 0; i < 100; i++ {
		items = append(items, basicwidget.TextListItem[int]{
			Text: fmt.Sprintf("Item %d", i),
		})
	}
	l.textList.SetItems(items)
	context.SetSize(&l.textList, image.Pt(guigui.DefaultSize, 6*basicwidget.UnitSize(context)))
	context.SetEnabled(&l.textList, l.model.Lists().Enabled())

	l.listForm.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &l.textListText,
			SecondaryWidget: &l.textList,
		},
	})

	// Configurations
	l.enabledText.SetValue("Enabled")
	l.enabledToggle.SetOnValueChanged(func(value bool) {
		l.model.Lists().SetEnabled(value)
	})
	l.enabledToggle.SetValue(l.model.Lists().Enabled())

	l.configForm.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &l.enabledText,
			SecondaryWidget: &l.enabledToggle,
		},
	})

	u := basicwidget.UnitSize(context)
	gl := layout.GridLayout{
		Bounds: context.Bounds(l).Inset(u / 2),
		Heights: []layout.Size{
			layout.FixedSize(l.listForm.DefaultSize(context).Y),
			layout.FlexibleSize(1),
			layout.FixedSize(l.configForm.DefaultSize(context).Y),
		},
		RowGap: u / 2,
	}
	appender.AppendChildWidgetWithBounds(&l.listForm, gl.CellBounds(0, 0))
	appender.AppendChildWidgetWithBounds(&l.configForm, gl.CellBounds(0, 2))

	return nil
}
