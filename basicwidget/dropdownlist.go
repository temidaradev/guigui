// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"log/slog"

	"github.com/hajimehoshi/guigui"
)

type DropdownList struct {
	guigui.DefaultWidget

	textButton TextButton
	popupMenu  PopupMenu

	onValueChanged func(index int)
}

func (d *DropdownList) SetOnValueChanged(f func(index int)) {
	d.onValueChanged = f
}

func (d *DropdownList) updateButtonImage(context *guigui.Context) {
	img, err := theResourceImages.Get("unfold_more", context.ColorMode())
	if err != nil {
		slog.Error(err.Error())
		return
	}
	d.textButton.SetImage(context, img)
}

func (d *DropdownList) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	d.updateButtonImage(context)
	d.updateText(context)

	d.textButton.SetOnDown(func() {
		d.popupMenu.Open(context)
	})
	d.textButton.SetForcePressed(context.IsVisible(&d.popupMenu))

	context.SetPosition(&d.textButton, context.Position(d))
	appender.AppendChildWidget(&d.textButton)

	d.popupMenu.SetOnClosed(func(index int) {
		if d.onValueChanged != nil {
			d.onValueChanged(index)
		}
	})
	d.popupMenu.SetCheckmarkIndex(context, d.SelectedItemIndex())

	pt := context.Position(d)
	pt.X -= listItemCheckmarkSize(context) + listItemTextAndImagePadding(context)
	pt.X = max(pt.X, 0)
	pt.Y -= listItemPadding(context)
	_, y := context.Size(d)
	pt.Y += int((float64(y) - LineHeight(context)) / 2)
	pt.Y -= int(float64(d.popupMenu.SelectedItemIndex()) * LineHeight(context))
	pt.Y = max(pt.Y, 0)
	context.SetPosition(&d.popupMenu, pt)
	appender.AppendChildWidget(&d.popupMenu)

	return nil
}

func (d *DropdownList) updateText(context *guigui.Context) {
	if item, ok := d.popupMenu.SelectedItem(); ok {
		d.textButton.SetText(context, item.Text)
	} else {
		d.textButton.SetText(context, "")
	}
}

func (d *DropdownList) SetItemsByStrings(context *guigui.Context, items []string) {
	d.popupMenu.SetItemsByStrings(context, items)
	d.updateText(context)
}

func (d *DropdownList) SelectedItemIndex() int {
	return d.popupMenu.SelectedItemIndex()
}

func (d *DropdownList) SetSelectedItemIndex(context *guigui.Context, index int) {
	d.popupMenu.SetSelectedItemIndex(context, index)
	d.updateText(context)
}

func (d *DropdownList) DefaultSize(context *guigui.Context) (int, int) {
	// The button image affects the size.
	d.updateButtonImage(context)
	return context.Size(&d.textButton)
}
