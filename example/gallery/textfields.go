// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/layout"
)

type TextFields struct {
	guigui.DefaultWidget

	form            basicwidget.Form
	startAlignText  basicwidget.Text
	startAlign      basicwidget.TextField
	centerAlignText basicwidget.Text
	centerAlign     basicwidget.TextField
	endAlignText    basicwidget.Text
	endAlign        basicwidget.TextField

	model *Model
}

func (t *TextFields) SetModel(model *Model) {
	t.model = model
}

func (t *TextFields) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	t.startAlignText.SetText("Horizontal Align - Start")
	t.startAlign.SetOnValueChanged(func(text string) {
		t.model.TextFields().SetHorizontalAlignStartText(text)
	})
	t.startAlign.SetText(t.model.TextFields().HorizontalAlignStartText())
	t.startAlign.SetHorizontalAlign(basicwidget.HorizontalAlignStart)

	t.centerAlignText.SetText("Horizontal Align - Center")
	t.centerAlign.SetOnValueChanged(func(text string) {
		t.model.TextFields().SetHorizontalAlignCenterText(text)
	})
	t.centerAlign.SetText(t.model.TextFields().HorizontalAlignCenterText())
	t.centerAlign.SetHorizontalAlign(basicwidget.HorizontalAlignCenter)

	t.endAlignText.SetText("Horizontal Align - End")
	t.endAlign.SetOnValueChanged(func(text string) {
		t.model.TextFields().SetHorizontalAlignEndText(text)
	})
	t.endAlign.SetText(t.model.TextFields().HorizontalAlignEndText())
	t.endAlign.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)

	t.form.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &t.startAlignText,
			SecondaryWidget: &t.startAlign,
		},
		{
			PrimaryWidget:   &t.centerAlignText,
			SecondaryWidget: &t.centerAlign,
		},
		{
			PrimaryWidget:   &t.endAlignText,
			SecondaryWidget: &t.endAlign,
		},
	})

	u := basicwidget.UnitSize(context)
	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(t).Inset(u / 2),
		Heights: []layout.Size{
			layout.FixedSize(t.form.DefaultSize(context).Y),
		},
		RowGap: u / 2,
	}).CellBounds() {
		switch i {
		case 0:
			appender.AppendChildWidgetWithBounds(&t.form, bounds)
		}
	}
	return nil
}
