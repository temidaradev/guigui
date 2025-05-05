// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Guigui Authors

package main

import (
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/layout"
)

type Buttons struct {
	guigui.DefaultWidget

	buttonsForm           basicwidget.Form
	textButtonText        basicwidget.Text
	textButton            basicwidget.TextButton
	textIconButtonText    basicwidget.Text
	textIconButton        basicwidget.TextButton
	segmentedControlHText basicwidget.Text
	segmentedControlH     basicwidget.SegmentedControl[int]
	segmentedControlVText basicwidget.Text
	segmentedControlV     basicwidget.SegmentedControl[int]
	toggleText            basicwidget.Text
	toggle                basicwidget.Toggle

	configForm    basicwidget.Form
	enabledText   basicwidget.Text
	enabledToggle basicwidget.Toggle

	model *Model
}

func (b *Buttons) SetModel(model *Model) {
	b.model = model
}

func (b *Buttons) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	b.textButtonText.SetValue("Text Button")
	b.textButton.SetText("Button")
	context.SetEnabled(&b.textButton, b.model.Buttons().Enabled())

	b.textIconButtonText.SetValue("Text w/ Icon Button")
	b.textIconButton.SetText("Button")
	img, err := theImageCache.Get("check", context.ColorMode())
	if err != nil {
		return err
	}
	b.textIconButton.SetIcon(img)
	context.SetEnabled(&b.textIconButton, b.model.Buttons().Enabled())

	b.segmentedControlHText.SetValue("Segmented Control (Horizontal)")
	b.segmentedControlH.SetItems([]basicwidget.SegmentedControlItem[int]{
		{
			Text: "One",
		},
		{
			Text: "Two",
		},
		{
			Text: "Three",
		},
	})
	b.segmentedControlH.SetDirection(basicwidget.SegmentedControlDirectionHorizontal)
	context.SetEnabled(&b.segmentedControlH, b.model.Buttons().Enabled())

	b.segmentedControlVText.SetValue("Segmented Control (Vertical)")
	b.segmentedControlV.SetItems([]basicwidget.SegmentedControlItem[int]{
		{
			Text: "One",
		},
		{
			Text: "Two",
		},
		{
			Text: "Three",
		},
	})
	b.segmentedControlV.SetDirection(basicwidget.SegmentedControlDirectionVertical)
	context.SetEnabled(&b.segmentedControlV, b.model.Buttons().Enabled())

	b.toggleText.SetValue("Toggle")
	context.SetEnabled(&b.toggle, b.model.Buttons().Enabled())

	b.buttonsForm.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &b.textButtonText,
			SecondaryWidget: &b.textButton,
		},
		{
			PrimaryWidget:   &b.textIconButtonText,
			SecondaryWidget: &b.textIconButton,
		},
		{
			PrimaryWidget:   &b.segmentedControlHText,
			SecondaryWidget: &b.segmentedControlH,
		},
		{
			PrimaryWidget:   &b.segmentedControlVText,
			SecondaryWidget: &b.segmentedControlV,
		},
		{
			PrimaryWidget:   &b.toggleText,
			SecondaryWidget: &b.toggle,
		},
	})

	b.enabledText.SetValue("Enabled")
	b.enabledToggle.SetOnValueChanged(func(enabled bool) {
		b.model.Buttons().SetEnabled(enabled)
	})
	b.enabledToggle.SetValue(b.model.Buttons().Enabled())

	b.configForm.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &b.enabledText,
			SecondaryWidget: &b.enabledToggle,
		},
	})

	u := basicwidget.UnitSize(context)
	gl := layout.GridLayout{
		Bounds: context.Bounds(b).Inset(u / 2),
		Heights: []layout.Size{
			layout.FixedSize(b.buttonsForm.DefaultSize(context).Y),
			layout.FlexibleSize(1),
			layout.FixedSize(b.configForm.DefaultSize(context).Y),
		},
		RowGap: u / 2,
	}
	appender.AppendChildWidgetWithBounds(&b.buttonsForm, gl.CellBounds(0, 0))
	appender.AppendChildWidgetWithBounds(&b.configForm, gl.CellBounds(0, 2))

	return nil
}
