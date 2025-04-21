// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/layout"
)

type Texts struct {
	guigui.DefaultWidget

	form                        basicwidget.Form
	horizontalAlignText         basicwidget.Text
	horizontalAlignDropdownList basicwidget.DropdownList[basicwidget.HorizontalAlign]
	verticalAlignText           basicwidget.Text
	verticalAlignDropdownList   basicwidget.DropdownList[basicwidget.VerticalAlign]
	autoWrapText                basicwidget.Text
	autoWrapToggle              basicwidget.Toggle
	boldText                    basicwidget.Text
	boldToggle                  basicwidget.Toggle
	selectableText              basicwidget.Text
	selectableToggle            basicwidget.Toggle
	editableText                basicwidget.Text
	editableToggle              basicwidget.Toggle
	sampleText                  basicwidget.Text

	model *Model
}

func (t *Texts) SetModel(model *Model) {
	t.model = model
}

func (t *Texts) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	t.horizontalAlignText.SetText("Horizontal Align")
	t.horizontalAlignDropdownList.SetItems([]basicwidget.DropdownListItem[basicwidget.HorizontalAlign]{
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
	t.horizontalAlignDropdownList.SetOnValueChanged(func(index int) {
		item, ok := t.horizontalAlignDropdownList.ItemByIndex(index)
		if !ok {
			t.model.Texts().SetHorizontalAlign(basicwidget.HorizontalAlignStart)
			return
		}
		t.model.Texts().SetHorizontalAlign(item.Tag)
	})
	t.horizontalAlignDropdownList.SelectItemByTag(t.model.Texts().HorizontalAlign())

	t.verticalAlignText.SetText("Vertical Align")
	t.verticalAlignDropdownList.SetItems([]basicwidget.DropdownListItem[basicwidget.VerticalAlign]{
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
	t.verticalAlignDropdownList.SetOnValueChanged(func(index int) {
		item, ok := t.verticalAlignDropdownList.ItemByIndex(index)
		if !ok {
			t.model.Texts().SetVerticalAlign(basicwidget.VerticalAlignTop)
			return
		}
		t.model.Texts().SetVerticalAlign(item.Tag)
	})
	t.verticalAlignDropdownList.SelectItemByTag(t.model.Texts().VerticalAlign())

	t.autoWrapText.SetText("Auto Wrap")
	t.autoWrapToggle.SetOnValueChanged(func(value bool) {
		t.model.Texts().SetAutoWrap(value)
	})
	t.autoWrapToggle.SetValue(t.model.Texts().AutoWrap())

	t.boldText.SetText("Bold")
	t.boldToggle.SetOnValueChanged(func(value bool) {
		t.model.Texts().SetBold(value)
	})
	t.boldToggle.SetValue(t.model.Texts().Bold())

	t.selectableText.SetText("Selectable")
	t.selectableToggle.SetOnValueChanged(func(checked bool) {
		t.model.Texts().SetSelectable(checked)
	})
	t.selectableToggle.SetValue(t.model.Texts().Selectable())

	t.editableText.SetText("Editable")
	t.editableToggle.SetOnValueChanged(func(value bool) {
		t.model.Texts().SetEditable(value)
	})
	t.editableToggle.SetValue(t.model.Texts().Editable())

	t.form.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &t.horizontalAlignText,
			SecondaryWidget: &t.horizontalAlignDropdownList,
		},
		{
			PrimaryWidget:   &t.verticalAlignText,
			SecondaryWidget: &t.verticalAlignDropdownList,
		},
		{
			PrimaryWidget:   &t.autoWrapText,
			SecondaryWidget: &t.autoWrapToggle,
		},
		{
			PrimaryWidget:   &t.boldText,
			SecondaryWidget: &t.boldToggle,
		},
		{
			PrimaryWidget:   &t.selectableText,
			SecondaryWidget: &t.selectableToggle,
		},
		{
			PrimaryWidget:   &t.editableText,
			SecondaryWidget: &t.editableToggle,
		},
	})

	t.sampleText.SetMultiline(true)
	t.sampleText.SetHorizontalAlign(t.model.Texts().HorizontalAlign())
	t.sampleText.SetVerticalAlign(t.model.Texts().VerticalAlign())
	t.sampleText.SetAutoWrap(t.model.Texts().AutoWrap())
	t.sampleText.SetBold(t.model.Texts().Bold())
	t.sampleText.SetSelectable(t.model.Texts().Selectable())
	t.sampleText.SetEditable(t.model.Texts().Editable())
	t.sampleText.SetOnValueChanged(func(text string) {
		t.model.Texts().SetText(text)
	})
	if !context.HasFocusedChildWidget(&t.sampleText) {
		t.sampleText.SetText(t.model.Texts().Text())
	}

	u := basicwidget.UnitSize(context)
	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(t).Inset(u / 2),
		Heights: []layout.Size{
			layout.FixedSize(t.form.DefaultSize(context).Y),
			layout.FlexibleSize(1),
		},
		RowGap: u / 2,
	}).CellBounds() {
		switch i {
		case 0:
			appender.AppendChildWidgetWithBounds(&t.form, bounds)
		case 1:
			appender.AppendChildWidgetWithBounds(&t.sampleText, bounds)
		}
	}

	return nil
}
