// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Basic struct {
	guigui.DefaultWidget

	form             basicwidget.Form
	textButtonText   basicwidget.Text
	textButton       basicwidget.TextButton
	toggleButtonText basicwidget.Text
	toggleButton     basicwidget.ToggleButton
	textFieldText    basicwidget.Text
	textField        basicwidget.TextField
	textListText     basicwidget.Text
	textList         basicwidget.TextList
}

func (b *Basic) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	b.textButtonText.SetText("Text Button")
	b.textButton.SetText("Click Me!")
	b.toggleButtonText.SetText("Toggle Button")
	b.textFieldText.SetText("Text Field")
	b.textField.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
	b.textListText.SetText("Text List")
	b.textList.SetItemsByStrings([]string{"Item 1", "Item 2", "Item 3"})

	u := float64(basicwidget.UnitSize(context))
	w, _ := context.Size(b)
	context.SetSize(&b.form, w-int(1*u), guigui.DefaultSize)
	b.form.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &b.textButtonText,
			SecondaryWidget: &b.textButton,
		},
		{
			PrimaryWidget:   &b.toggleButtonText,
			SecondaryWidget: &b.toggleButton,
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
	{
		p := context.Position(b).Add(image.Pt(int(0.5*u), int(0.5*u)))
		appender.AppendChildWidgetWithPosition(&b.form, p)
	}

	return nil
}
