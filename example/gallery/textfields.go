// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/layout"
)

type TextFields struct {
	guigui.DefaultWidget

	form                 basicwidget.Form
	startAlignText       basicwidget.Text
	startAlignTextField  basicwidget.TextField
	centerAlignText      basicwidget.Text
	centerAlignTextField basicwidget.TextField
	endAlignText         basicwidget.Text
	endAlignTextField    basicwidget.TextField
	multilineText        basicwidget.Text
	multilineTextField   basicwidget.TextField

	model *Model
}

func (t *TextFields) SetModel(model *Model) {
	t.model = model
}

func (t *TextFields) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	u := basicwidget.UnitSize(context)
	width := 12 * u

	t.startAlignText.SetText("Horizontal Align - Start")
	t.startAlignTextField.SetOnValueChanged(func(text string) {
		t.model.TextFields().SetHorizontalAlignStartText(text)
	})
	t.startAlignTextField.SetText(t.model.TextFields().HorizontalAlignStartText())
	t.startAlignTextField.SetHorizontalAlign(basicwidget.HorizontalAlignStart)
	context.SetSize(&t.startAlignTextField, image.Pt(width, guigui.DefaultSize))

	t.centerAlignText.SetText("Horizontal Align - Center")
	t.centerAlignTextField.SetOnValueChanged(func(text string) {
		t.model.TextFields().SetHorizontalAlignCenterText(text)
	})
	t.centerAlignTextField.SetText(t.model.TextFields().HorizontalAlignCenterText())
	t.centerAlignTextField.SetHorizontalAlign(basicwidget.HorizontalAlignCenter)
	context.SetSize(&t.centerAlignTextField, image.Pt(width, guigui.DefaultSize))

	t.endAlignText.SetText("Horizontal Align - End")
	t.endAlignTextField.SetOnValueChanged(func(text string) {
		t.model.TextFields().SetHorizontalAlignEndText(text)
	})
	t.endAlignTextField.SetText(t.model.TextFields().HorizontalAlignEndText())
	t.endAlignTextField.SetHorizontalAlign(basicwidget.HorizontalAlignEnd)
	context.SetSize(&t.endAlignTextField, image.Pt(width, guigui.DefaultSize))

	t.multilineText.SetText("Multiline")
	t.multilineTextField.SetOnValueChanged(func(text string) {
		t.model.TextFields().SetMultilineText(text)
	})
	t.multilineTextField.SetText(t.model.TextFields().MultilineText())
	t.multilineTextField.SetMultiline(true)
	context.SetSize(&t.multilineTextField, image.Pt(width, 4*u))

	t.form.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &t.startAlignText,
			SecondaryWidget: &t.startAlignTextField,
		},
		{
			PrimaryWidget:   &t.centerAlignText,
			SecondaryWidget: &t.centerAlignTextField,
		},
		{
			PrimaryWidget:   &t.endAlignText,
			SecondaryWidget: &t.endAlignTextField,
		},
		{
			PrimaryWidget:   &t.multilineText,
			SecondaryWidget: &t.multilineTextField,
		},
	})

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
