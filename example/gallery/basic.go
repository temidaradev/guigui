// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/layout"
)

type Basic struct {
	guigui.DefaultWidget

	form           basicwidget.Form
	textButtonText basicwidget.Text
	textButton     basicwidget.TextButton
	toggleText     basicwidget.Text
	toggle         basicwidget.Toggle
	textFieldText  basicwidget.Text
	textField      basicwidget.TextField
	textListText   basicwidget.Text
	textList       basicwidget.TextList[int]
}

func (b *Basic) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	b.textButtonText.SetText("Text Button")
	b.textButton.SetText("Click Me!")
	b.toggleText.SetText("Toggle")
	b.textFieldText.SetText("Text Field")
	b.textField.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
	b.textListText.SetText("Text List")
	b.textList.SetItemsByStrings([]string{"Item 1", "Item 2", "Item 3"})

	b.form.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &b.textButtonText,
			SecondaryWidget: &b.textButton,
		},
		{
			PrimaryWidget:   &b.toggleText,
			SecondaryWidget: &b.toggle,
		},
		{
			PrimaryWidget:   &b.textFieldText,
			SecondaryWidget: &b.textField,
		},
		{
			PrimaryWidget:   &b.textListText,
			SecondaryWidget: &b.textList,
		},
	})

	u := basicwidget.UnitSize(context)
	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(b).Inset(u / 2),
		Heights: []layout.Size{
			layout.MaxContentSize(func(index int) int {
				if index >= 1 {
					return 0
				}
				return b.form.DefaultSize(context).Y
			}),
		},
		RowGap: u / 2,
	}).RepeatingCellBounds() {
		if i >= 1 {
			break
		}
		appender.AppendChildWidgetWithBounds(&b.form, bounds)
	}

	return nil
}
