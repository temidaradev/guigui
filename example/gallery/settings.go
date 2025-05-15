// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Guigui Authors

package main

import (
	"image"

	"golang.org/x/text/language"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/layout"
)

type Settings struct {
	guigui.DefaultWidget

	form                      basicwidget.Form
	colorModeText             basicwidget.Text
	colorModeSegmentedControl basicwidget.SegmentedControl[string]
	localeText                textWithSubText
	localeDropdownList        basicwidget.DropdownList[language.Tag]
	scaleText                 basicwidget.Text
	scaleSegmentedControl     basicwidget.SegmentedControl[float64]
}

var hongKongChinese = language.MustParse("zh-HK")

func (s *Settings) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	lightModeImg, err := theImageCache.GetMonochrome("light_mode", context.ColorMode())
	if err != nil {
		return err
	}
	darkModeImg, err := theImageCache.GetMonochrome("dark_mode", context.ColorMode())
	if err != nil {
		return err
	}

	s.colorModeText.SetValue("Color mode")
	s.colorModeSegmentedControl.SetItems([]basicwidget.SegmentedControlItem[string]{
		{
			Text: "Auto",
			ID:   "",
		},
		{
			Icon: lightModeImg,
			ID:   "light",
		},
		{
			Icon: darkModeImg,
			ID:   "dark",
		},
	})
	s.colorModeSegmentedControl.SetOnItemSelected(func(index int) {
		item, ok := s.colorModeSegmentedControl.ItemByIndex(index)
		if !ok {
			context.SetColorMode(guigui.ColorModeLight)
			return
		}
		switch item.ID {
		case "light":
			context.SetColorMode(guigui.ColorModeLight)
		case "dark":
			context.SetColorMode(guigui.ColorModeDark)
		default:
			context.UseAutoColorMode()
		}
	})
	if context.IsAutoColorModeUsed() {
		s.colorModeSegmentedControl.SelectItemByID("")
	} else {
		switch context.ColorMode() {
		case guigui.ColorModeLight:
			s.colorModeSegmentedControl.SelectItemByID("light")
		case guigui.ColorModeDark:
			s.colorModeSegmentedControl.SelectItemByID("dark")
		default:
			s.colorModeSegmentedControl.SelectItemByID("")
		}
	}

	s.localeText.text.SetValue("Locale")
	s.localeText.subText.SetValue("The locale affects the glyphs for Chinese characters.")

	s.localeDropdownList.SetItems([]basicwidget.DropdownListItem[language.Tag]{
		{
			Text: "(Default)",
			ID:   language.Und,
		},
		{
			Text: "English",
			ID:   language.English,
		},
		{
			Text: "Japanese",
			ID:   language.Japanese,
		},
		{
			Text: "Korean",
			ID:   language.Korean,
		},
		{
			Text: "Simplified Chinese",
			ID:   language.SimplifiedChinese,
		},
		{
			Text: "Traditional Chinese",
			ID:   language.TraditionalChinese,
		},
		{
			Text: "Hong Kong Chinese",
			ID:   hongKongChinese,
		},
	})
	s.localeDropdownList.SetOnItemSelected(func(index int) {
		item, ok := s.localeDropdownList.ItemByIndex(index)
		if !ok {
			context.SetAppLocales(nil)
			return
		}
		if item.ID == language.Und {
			context.SetAppLocales(nil)
			return
		}
		context.SetAppLocales([]language.Tag{item.ID})
	})
	if !s.localeDropdownList.IsPopupOpen() {
		if locales := context.AppendAppLocales(nil); len(locales) > 0 {
			s.localeDropdownList.SelectItemByID(locales[0])
		} else {
			s.localeDropdownList.SelectItemByID(language.Und)
		}
	}

	s.scaleText.SetValue("Scale")
	s.scaleSegmentedControl.SetItems([]basicwidget.SegmentedControlItem[float64]{
		{
			Text: "80%",
			ID:   0.8,
		},
		{
			Text: "100%",
			ID:   1,
		},
		{
			Text: "120%",
			ID:   1.2,
		},
	})
	s.scaleSegmentedControl.SetOnItemSelected(func(index int) {
		item, ok := s.scaleSegmentedControl.ItemByIndex(index)
		if !ok {
			context.SetAppScale(1)
			return
		}
		context.SetAppScale(item.ID)
	})
	s.scaleSegmentedControl.SelectItemByID(context.AppScale())

	s.form.SetItems([]basicwidget.FormItem{
		{
			PrimaryWidget:   &s.colorModeText,
			SecondaryWidget: &s.colorModeSegmentedControl,
		},
		{
			PrimaryWidget:   &s.localeText,
			SecondaryWidget: &s.localeDropdownList,
		},
		{
			PrimaryWidget:   &s.scaleText,
			SecondaryWidget: &s.scaleSegmentedControl,
		},
	})

	u := basicwidget.UnitSize(context)
	gl := layout.GridLayout{
		Bounds: context.Bounds(s).Inset(u / 2),
		Heights: []layout.Size{
			layout.LazySize(func(row int) layout.Size {
				if row >= 1 {
					return layout.FixedSize(0)
				}
				return layout.FixedSize(s.form.DefaultSize(context).Y)
			}),
		},
		RowGap: u / 2,
	}
	appender.AppendChildWidgetWithBounds(&s.form, gl.CellBounds(0, 0))

	return nil
}

type textWithSubText struct {
	guigui.DefaultWidget

	text    basicwidget.Text
	subText basicwidget.Text
}

func (t *textWithSubText) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	pt := context.Position(t)
	appender.AppendChildWidgetWithPosition(&t.text, pt)

	pt.Y += context.Size(&t.text).Y
	t.subText.SetScale(0.875)
	t.subText.SetMultiline(true)
	t.subText.SetAutoWrap(true)
	t.subText.SetOpacity(0.675)
	appender.AppendChildWidgetWithPosition(&t.subText, pt)

	return nil
}

func (t *textWithSubText) DefaultSize(context *guigui.Context) image.Point {
	s1 := t.text.DefaultSize(context)
	s2 := t.subText.DefaultSize(context)
	return image.Pt(max(s1.X, s2.X), s1.Y+s2.Y)
}
