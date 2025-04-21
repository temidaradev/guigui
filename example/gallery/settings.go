// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"sync"

	"golang.org/x/text/language"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/layout"
)

type Settings struct {
	guigui.DefaultWidget

	form                  basicwidget.Form
	colorModeText         basicwidget.Text
	colorModeDropdownList basicwidget.DropdownList[guigui.ColorMode]
	localeText            basicwidget.Text
	localeDropdownList    basicwidget.DropdownList[language.Tag]
	scaleText             basicwidget.Text
	scaleDropdownList     basicwidget.DropdownList[float64]

	initOnce sync.Once
}

func (s *Settings) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	s.colorModeText.SetText("Color Mode")
	s.colorModeDropdownList.SetItems([]basicwidget.DropdownListItem[guigui.ColorMode]{
		{
			Text: "Light",
			Tag:  guigui.ColorModeLight,
		},
		{
			Text: "Dark",
			Tag:  guigui.ColorModeDark,
		},
	})
	s.colorModeDropdownList.SetOnValueChanged(func(index int) {
		item, ok := s.colorModeDropdownList.ItemByIndex(index)
		if !ok {
			return
		}
		context.SetColorMode(item.Tag)
	})

	s.localeText.SetText("Locale")
	s.localeDropdownList.SetItems([]basicwidget.DropdownListItem[language.Tag]{
		{
			Text: "(Default)",
			Tag:  language.Und,
		},
		{
			Text: "en",
			Tag:  language.English,
		},
		{
			Text: "ja",
			Tag:  language.Japanese,
		},
		{
			Text: "ko",
			Tag:  language.Korean,
		},
		{
			Text: "zh-Hans",
			Tag:  language.SimplifiedChinese,
		},
		{
			Text: "zh-Hant",
			Tag:  language.TraditionalChinese,
		},
	})
	s.localeDropdownList.SetOnValueChanged(func(index int) {
		item, ok := s.localeDropdownList.ItemByIndex(index)
		if !ok {
			return
		}
		context.SetAppLocales([]language.Tag{item.Tag})
	})

	s.scaleText.SetText("Scale")
	s.scaleDropdownList.SetItems([]basicwidget.DropdownListItem[float64]{
		{
			Text: "80%",
			Tag:  0.8,
		},
		{
			Text: "100%",
			Tag:  1,
		},
		{
			Text: "120%",
			Tag:  1.2,
		},
	})
	s.scaleDropdownList.SetOnValueChanged(func(index int) {
		item, ok := s.scaleDropdownList.ItemByIndex(index)
		if !ok {
			return
		}
		context.SetAppScale(item.Tag)
	})

	s.initOnce.Do(func() {
		s.colorModeDropdownList.SelectItemByTag(context.ColorMode())
		s.localeDropdownList.SelectItemByIndex(0)
		s.scaleDropdownList.SelectItemByIndex(1)
	})

	s.form.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &s.colorModeText,
			SecondaryWidget: &s.colorModeDropdownList,
		},
		{
			PrimaryWidget:   &s.localeText,
			SecondaryWidget: &s.localeDropdownList,
		},
		{
			PrimaryWidget:   &s.scaleText,
			SecondaryWidget: &s.scaleDropdownList,
		},
	})

	u := basicwidget.UnitSize(context)
	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(s).Inset(u / 2),
		Heights: []layout.Size{
			layout.MaxContentSize(func(index int) int {
				if index >= 1 {
					return 0
				}
				return s.form.DefaultSize(context).Y
			}),
		},
		RowGap: u / 2,
	}).RepeatingCellBounds() {
		if i >= 1 {
			break
		}
		appender.AppendChildWidgetWithBounds(&s.form, bounds)
	}

	return nil
}
