// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
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

	u := float64(basicwidget.UnitSize(context))
	w, _ := context.Size(b)
	context.SetSize(&b.form, w-int(1*u), guigui.DefaultSize)
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
	appender.AppendChildWidgetWithPosition(&b.form, context.Position(b).Add(image.Pt(int(0.5*u), int(0.5*u))))

	return nil
}
