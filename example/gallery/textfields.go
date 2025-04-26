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

	textFieldForm       basicwidget.Form
	singleLineText      basicwidget.Text
	singleLineTextField basicwidget.TextField
	multilineText       basicwidget.Text
	multilineTextField  basicwidget.TextField

	configForm                      basicwidget.Form
	horizontalAlignText             basicwidget.Text
	horizontalAlignSegmentedControl basicwidget.SegmentedControl[basicwidget.HorizontalAlign]
	verticalAlignText               basicwidget.Text
	verticalAlignSegmentedControl   basicwidget.SegmentedControl[basicwidget.VerticalAlign]
	autoWrapText                    basicwidget.Text
	autoWrapToggle                  basicwidget.Toggle
	enabledText                     basicwidget.Text
	enabledToggle                   basicwidget.Toggle

	model *Model
}

func (t *TextFields) SetModel(model *Model) {
	t.model = model
}

func (t *TextFields) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	u := basicwidget.UnitSize(context)

	// Text Fields
	width := 12 * u

	t.singleLineText.SetText("Single Line")
	t.singleLineTextField.SetOnValueChanged(func(text string) {
		t.model.TextFields().SetSingleLineText(text)
	})
	t.singleLineTextField.SetText(t.model.TextFields().SingleLineText())
	t.singleLineTextField.SetHorizontalAlign(t.model.TextFields().HorizontalAlign())
	t.singleLineTextField.SetVerticalAlign(t.model.TextFields().VerticalAlign())
	context.SetEnabled(&t.singleLineTextField, t.model.TextFields().Enabled())
	context.SetSize(&t.singleLineTextField, image.Pt(width, guigui.DefaultSize))

	t.multilineText.SetText("Multiline")
	t.multilineTextField.SetOnValueChanged(func(text string) {
		t.model.TextFields().SetMultilineText(text)
	})
	t.multilineTextField.SetText(t.model.TextFields().MultilineText())
	t.multilineTextField.SetMultiline(true)
	t.multilineTextField.SetHorizontalAlign(t.model.TextFields().HorizontalAlign())
	t.multilineTextField.SetVerticalAlign(t.model.TextFields().VerticalAlign())
	t.multilineTextField.SetAutoWrap(t.model.TextFields().AutoWrap())
	context.SetEnabled(&t.multilineTextField, t.model.TextFields().Enabled())
	context.SetSize(&t.multilineTextField, image.Pt(width, 4*u))

	// Configurations
	t.horizontalAlignText.SetText("Horizontal Align")
	t.horizontalAlignSegmentedControl.SetItems([]basicwidget.SegmentedControlItem[basicwidget.HorizontalAlign]{
		{
			Text: "Start",
			Tag:  basicwidget.HorizontalAlignStart,
		},
		{
			Text: "Center",
			Tag:  basicwidget.HorizontalAlignCenter,
		},
		{
			Text: "End",
			Tag:  basicwidget.HorizontalAlignEnd,
		},
	})
	t.horizontalAlignSegmentedControl.SetOnItemSelected(func(index int) {
		item, ok := t.horizontalAlignSegmentedControl.ItemByIndex(index)
		if !ok {
			t.model.TextFields().SetHorizontalAlign(basicwidget.HorizontalAlignStart)
			return
		}
		t.model.TextFields().SetHorizontalAlign(item.Tag)
	})
	t.horizontalAlignSegmentedControl.SelectItemByTag(t.model.TextFields().HorizontalAlign())

	t.verticalAlignText.SetText("Vertical Align")
	t.verticalAlignSegmentedControl.SetItems([]basicwidget.SegmentedControlItem[basicwidget.VerticalAlign]{
		{
			Text: "Top",
			Tag:  basicwidget.VerticalAlignTop,
		},
		{
			Text: "Middle",
			Tag:  basicwidget.VerticalAlignMiddle,
		},
		{
			Text: "Bottom",
			Tag:  basicwidget.VerticalAlignBottom,
		},
	})
	t.verticalAlignSegmentedControl.SetOnItemSelected(func(index int) {
		item, ok := t.verticalAlignSegmentedControl.ItemByIndex(index)
		if !ok {
			t.model.TextFields().SetVerticalAlign(basicwidget.VerticalAlignTop)
			return
		}
		t.model.TextFields().SetVerticalAlign(item.Tag)
	})
	t.verticalAlignSegmentedControl.SelectItemByTag(t.model.TextFields().VerticalAlign())

	t.autoWrapText.SetText("Auto Wrap")
	t.autoWrapToggle.SetOnValueChanged(func(value bool) {
		t.model.TextFields().SetAutoWrap(value)
	})
	t.autoWrapToggle.SetValue(t.model.TextFields().AutoWrap())

	t.enabledText.SetText("Enabled")
	t.enabledToggle.SetOnValueChanged(func(value bool) {
		t.model.TextFields().SetEnabled(value)
	})
	t.enabledToggle.SetValue(t.model.TextFields().Enabled())

	t.configForm.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &t.horizontalAlignText,
			SecondaryWidget: &t.horizontalAlignSegmentedControl,
		},
		{
			PrimaryWidget:   &t.verticalAlignText,
			SecondaryWidget: &t.verticalAlignSegmentedControl,
		},
		{
			PrimaryWidget:   &t.autoWrapText,
			SecondaryWidget: &t.autoWrapToggle,
		},
		{
			PrimaryWidget:   &t.enabledText,
			SecondaryWidget: &t.enabledToggle,
		},
	})

	t.textFieldForm.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &t.singleLineText,
			SecondaryWidget: &t.singleLineTextField,
		},
		{
			PrimaryWidget:   &t.multilineText,
			SecondaryWidget: &t.multilineTextField,
		},
	})

	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(t).Inset(u / 2),
		Heights: []layout.Size{
			layout.FixedSize(t.textFieldForm.DefaultSize(context).Y),
			layout.FlexibleSize(1),
			layout.FixedSize(t.configForm.DefaultSize(context).Y),
		},
		RowGap: u / 2,
	}).CellBounds() {
		switch i {
		case 0:
			appender.AppendChildWidgetWithBounds(&t.textFieldForm, bounds)
		case 2:
			appender.AppendChildWidgetWithBounds(&t.configForm, bounds)
		}
	}
	return nil
}
