// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/layout"
)

type Buttons struct {
	guigui.DefaultWidget

	form                basicwidget.Form
	textButtonText      basicwidget.Text
	textButton          basicwidget.TextButton
	textImageButtonText basicwidget.Text
	textImageButton     basicwidget.TextButton
	toggleButtonText    basicwidget.Text
	toggleButton        basicwidget.ToggleButton
}

func (b *Buttons) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	b.textButtonText.SetText("Text Button")
	b.textButton.SetText("Button")
	b.textImageButtonText.SetText("Text w/ Image Button")
	b.textImageButton.SetText("Button")
	img, err := theImageCache.Get("check", context.ColorMode())
	if err != nil {
		return err
	}
	b.textImageButton.SetImage(img)
	b.toggleButtonText.SetText("Toggle Button")

	b.form.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &b.textButtonText,
			SecondaryWidget: &b.textButton,
		},
		{
			PrimaryWidget:   &b.textImageButtonText,
			SecondaryWidget: &b.textImageButton,
		},
		{
			PrimaryWidget:   &b.toggleButtonText,
			SecondaryWidget: &b.toggleButton,
		},
	})

	u := basicwidget.UnitSize(context)
	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(b).Inset(u / 2),
		Heights: []layout.Size{
			layout.MaxContentSize(func(index int) int {
				if index >= 1 {
					return 0
				}
				return context.Size(&b.form).Y
			}),
		},
		RowGap: u / 2,
	}).RepeatingCellBounds() {
		if i >= 1 {
			break
		}
		appender.AppendChildWidgetWithBounds(&b.form, bounds)
	}

	return nil
}
