// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
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
			context.SetColorMode(guigui.ColorModeLight)
			return
		}
		context.SetColorMode(item.Tag)
	})
	if !s.colorModeDropdownList.IsPopupOpen() {
		s.colorModeDropdownList.SelectItemByTag(context.ColorMode())
	}

	s.localeText.SetText("Locale")
	s.localeDropdownList.SetItems([]basicwidget.DropdownListItem[language.Tag]{
		{
			Text: "(Default)",
			Tag:  language.Und,
		},
		{
			Text: "English",
			Tag:  language.English,
		},
		{
			Text: "Japanese",
			Tag:  language.Japanese,
		},
		{
			Text: "Korean",
			Tag:  language.Korean,
		},
		{
			Text: "Simplified Chinese",
			Tag:  language.SimplifiedChinese,
		},
		{
			Text: "Traditional Chinese",
			Tag:  language.TraditionalChinese,
		},
	})
	s.localeDropdownList.SetOnValueChanged(func(index int) {
		item, ok := s.localeDropdownList.ItemByIndex(index)
		if !ok {
			context.SetAppLocales(nil)
			return
		}
		if item.Tag == language.Und {
			context.SetAppLocales(nil)
			return
		}
		context.SetAppLocales([]language.Tag{item.Tag})
	})
	if !s.localeDropdownList.IsPopupOpen() {
		if locales := context.AppendAppLocales(nil); len(locales) > 0 {
			s.localeDropdownList.SelectItemByTag(locales[0])
		} else {
			s.localeDropdownList.SelectItemByTag(language.Und)
		}
	}

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
			context.SetAppScale(1)
			return
		}
		context.SetAppScale(item.Tag)
	})
	if !s.scaleDropdownList.IsPopupOpen() {
		s.scaleDropdownList.SelectItemByTag(context.AppScale())
	}

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
