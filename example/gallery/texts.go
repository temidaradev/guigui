// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"sync"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/layout"
)

type Texts struct {
	guigui.DefaultWidget

	horizontalAlign basicwidget.HorizontalAlign
	verticalAlign   basicwidget.VerticalAlign
	unwrap          bool
	bold            bool
	selectable      bool
	editable        bool

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

	initOnce sync.Once
}

const sampleText = `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
隴西の李徴は博学才穎、天宝の末年、若くして名を虎榜に連ね、ついで江南尉に補せられたが、性、狷介、自ら恃むところ頗る厚く、賤吏に甘んずるを潔しとしなかった。`

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
			t.horizontalAlign = basicwidget.HorizontalAlignStart
			return
		}
		t.horizontalAlign = item.Tag
	})

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
			t.verticalAlign = basicwidget.VerticalAlignTop
			return
		}
		t.verticalAlign = item.Tag
	})

	t.autoWrapText.SetText("Auto Wrap")
	t.autoWrapToggle.SetValue(!t.unwrap)
	t.autoWrapToggle.SetOnValueChanged(func(checked bool) {
		t.unwrap = !checked
	})

	t.boldText.SetText("Bold")
	t.boldToggle.SetValue(t.bold)
	t.boldToggle.SetOnValueChanged(func(checked bool) {
		t.bold = checked
	})

	t.selectableText.SetText("Selectable")
	t.selectableToggle.SetValue(t.selectable)
	t.selectableToggle.SetOnValueChanged(func(checked bool) {
		t.selectable = checked
		if !t.selectable {
			t.editable = false
		}
	})

	t.editableText.SetText("Editable")
	t.editableToggle.SetValue(t.editable)
	t.editableToggle.SetOnValueChanged(func(checked bool) {
		t.editable = checked
		if t.editable {
			t.selectable = true
		}
	})

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
	t.sampleText.SetHorizontalAlign(t.horizontalAlign)
	t.sampleText.SetVerticalAlign(t.verticalAlign)
	t.sampleText.SetAutoWrap(!t.unwrap)
	t.sampleText.SetBold(t.bold)
	t.sampleText.SetSelectable(t.selectable)
	t.sampleText.SetEditable(t.editable)

	t.initOnce.Do(func() {
		t.sampleText.SetText(sampleText)
		t.horizontalAlignDropdownList.SelectItemByIndex(0)
		t.verticalAlignDropdownList.SelectItemByIndex(0)
	})

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
