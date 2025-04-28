// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/layout"
)

type TextInputs struct {
	guigui.DefaultWidget

	textInputForm       basicwidget.Form
	singleLineText      basicwidget.Text
	singleLineTextInput basicwidget.TextInput
	multilineText       basicwidget.Text
	multilineTextInput  basicwidget.TextInput
	numberInput1Text    basicwidget.Text
	numberInput1        basicwidget.NumberInput
	numberInput2Text    basicwidget.Text
	numberInput2        basicwidget.NumberInput

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

func (t *TextInputs) SetModel(model *Model) {
	t.model = model
}

func (t *TextInputs) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	u := basicwidget.UnitSize(context)

	// Text Inputs
	width := 12 * u

	t.singleLineText.SetText("Single Line")
	t.singleLineTextInput.SetOnValueChanged(func(text string) {
		t.model.TextInputs().SetSingleLineText(text)
	})
	t.singleLineTextInput.SetText(t.model.TextInputs().SingleLineText())
	t.singleLineTextInput.SetHorizontalAlign(t.model.TextInputs().HorizontalAlign())
	t.singleLineTextInput.SetVerticalAlign(t.model.TextInputs().VerticalAlign())
	context.SetEnabled(&t.singleLineTextInput, t.model.TextInputs().Enabled())
	context.SetSize(&t.singleLineTextInput, image.Pt(width, guigui.DefaultSize))

	t.multilineText.SetText("Multiline")
	t.multilineTextInput.SetOnValueChanged(func(text string) {
		t.model.TextInputs().SetMultilineText(text)
	})
	t.multilineTextInput.SetText(t.model.TextInputs().MultilineText())
	t.multilineTextInput.SetMultiline(true)
	t.multilineTextInput.SetHorizontalAlign(t.model.TextInputs().HorizontalAlign())
	t.multilineTextInput.SetVerticalAlign(t.model.TextInputs().VerticalAlign())
	t.multilineTextInput.SetAutoWrap(t.model.TextInputs().AutoWrap())
	context.SetEnabled(&t.multilineTextInput, t.model.TextInputs().Enabled())
	context.SetSize(&t.multilineTextInput, image.Pt(width, 4*u))

	t.numberInput1Text.SetText("Number Field")
	t.numberInput1.SetOnValueChanged(func(value int64) {
		t.model.TextInputs().SetNumberFieldValue1(value)
	})
	t.numberInput1.SetValue(t.model.TextInputs().NumberFieldValue1())
	context.SetEnabled(&t.numberInput1, t.model.TextInputs().Enabled())
	context.SetSize(&t.numberInput1, image.Pt(width, guigui.DefaultSize))

	t.numberInput2Text.SetText("Number Field w/ Range and Step")
	t.numberInput2.SetOnValueChanged(func(value int64) {
		t.model.TextInputs().SetNumberFieldValue2(value)
	})
	t.numberInput2.SetMinimumValue(-100)
	t.numberInput2.SetMaximumValue(100)
	t.numberInput2.SetStep(5)
	t.numberInput2.SetValue(t.model.TextInputs().NumberFieldValue2())
	context.SetEnabled(&t.numberInput2, t.model.TextInputs().Enabled())
	context.SetSize(&t.numberInput2, image.Pt(width, guigui.DefaultSize))

	t.textInputForm.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &t.singleLineText,
			SecondaryWidget: &t.singleLineTextInput,
		},
		{
			PrimaryWidget:   &t.multilineText,
			SecondaryWidget: &t.multilineTextInput,
		},
		{
			PrimaryWidget:   &t.numberInput1Text,
			SecondaryWidget: &t.numberInput1,
		},
		{
			PrimaryWidget:   &t.numberInput2Text,
			SecondaryWidget: &t.numberInput2,
		},
	})

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
			t.model.TextInputs().SetHorizontalAlign(basicwidget.HorizontalAlignStart)
			return
		}
		t.model.TextInputs().SetHorizontalAlign(item.Tag)
	})
	t.horizontalAlignSegmentedControl.SelectItemByTag(t.model.TextInputs().HorizontalAlign())

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
			t.model.TextInputs().SetVerticalAlign(basicwidget.VerticalAlignTop)
			return
		}
		t.model.TextInputs().SetVerticalAlign(item.Tag)
	})
	t.verticalAlignSegmentedControl.SelectItemByTag(t.model.TextInputs().VerticalAlign())

	t.autoWrapText.SetText("Auto Wrap")
	t.autoWrapToggle.SetOnValueChanged(func(value bool) {
		t.model.TextInputs().SetAutoWrap(value)
	})
	t.autoWrapToggle.SetValue(t.model.TextInputs().AutoWrap())

	t.enabledText.SetText("Enabled")
	t.enabledToggle.SetOnValueChanged(func(value bool) {
		t.model.TextInputs().SetEnabled(value)
	})
	t.enabledToggle.SetValue(t.model.TextInputs().Enabled())

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

	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(t).Inset(u / 2),
		Heights: []layout.Size{
			layout.FixedSize(t.textInputForm.DefaultSize(context).Y),
			layout.FlexibleSize(1),
			layout.FixedSize(t.configForm.DefaultSize(context).Y),
		},
		RowGap: u / 2,
	}).CellBounds() {
		switch i {
		case 0:
			appender.AppendChildWidgetWithBounds(&t.textInputForm, bounds)
		case 2:
			appender.AppendChildWidgetWithBounds(&t.configForm, bounds)
		}
	}
	return nil
}
