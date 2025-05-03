// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/layout"
)

type NumberInputs struct {
	guigui.DefaultWidget

	numberInputForm  basicwidget.Form
	numberInput1Text basicwidget.Text
	numberInput1     basicwidget.NumberInput[int]
	numberInput2Text basicwidget.Text
	numberInput2     basicwidget.NumberInput[int]

	configForm     basicwidget.Form
	editableText   basicwidget.Text
	editableToggle basicwidget.Toggle
	enabledText    basicwidget.Text
	enabledToggle  basicwidget.Toggle

	model *Model
}

func (n *NumberInputs) SetModel(model *Model) {
	n.model = model
}

func (n *NumberInputs) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	u := basicwidget.UnitSize(context)

	// Number Inputs
	width := 12 * u

	n.numberInput1Text.SetText("Number Field")
	n.numberInput1.SetOnValueChanged(func(value int) {
		n.model.NumberInputs().SetNumberFieldValue1(value)
	})
	n.numberInput1.SetValue(n.model.NumberInputs().NumberFieldValue1())
	n.numberInput1.SetEditable(n.model.NumberInputs().Editable())
	context.SetEnabled(&n.numberInput1, n.model.NumberInputs().Enabled())
	context.SetSize(&n.numberInput1, image.Pt(width, guigui.DefaultSize))

	n.numberInput2Text.SetText("Number Field w/ Range and Step")
	n.numberInput2.SetOnValueChanged(func(value int) {
		n.model.NumberInputs().SetNumberFieldValue2(value)
	})
	n.numberInput2.SetMinimumValue(-100)
	n.numberInput2.SetMaximumValue(100)
	n.numberInput2.SetStep(5)
	n.numberInput2.SetValue(n.model.NumberInputs().NumberFieldValue2())
	n.numberInput2.SetEditable(n.model.NumberInputs().Editable())
	context.SetEnabled(&n.numberInput2, n.model.NumberInputs().Enabled())
	context.SetSize(&n.numberInput2, image.Pt(width, guigui.DefaultSize))

	n.numberInputForm.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &n.numberInput1Text,
			SecondaryWidget: &n.numberInput1,
		},
		{
			PrimaryWidget:   &n.numberInput2Text,
			SecondaryWidget: &n.numberInput2,
		},
	})

	// Configurations
	n.editableText.SetText("Editable")
	n.editableToggle.SetOnValueChanged(func(value bool) {
		n.model.NumberInputs().SetEditable(value)
	})
	n.editableToggle.SetValue(n.model.NumberInputs().Editable())

	n.enabledText.SetText("Enabled")
	n.enabledToggle.SetOnValueChanged(func(value bool) {
		n.model.NumberInputs().SetEnabled(value)
	})
	n.enabledToggle.SetValue(n.model.NumberInputs().Enabled())

	n.configForm.SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &n.editableText,
			SecondaryWidget: &n.editableToggle,
		},
		{
			PrimaryWidget:   &n.enabledText,
			SecondaryWidget: &n.enabledToggle,
		},
	})

	gl := layout.GridLayout{
		Bounds: context.Bounds(n).Inset(u / 2),
		Heights: []layout.Size{
			layout.FixedSize(n.numberInputForm.DefaultSize(context).Y),
			layout.FlexibleSize(1),
			layout.FixedSize(n.configForm.DefaultSize(context).Y),
		},
		RowGap: u / 2,
	}
	appender.AppendChildWidgetWithBounds(&n.numberInputForm, gl.CellBounds(0, 0))
	appender.AppendChildWidgetWithBounds(&n.configForm, gl.CellBounds(0, 2))

	return nil
}
