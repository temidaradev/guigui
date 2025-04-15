// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"image"
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
	d.textButton.SetImage(img)
}

func (d *DropdownList) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	d.updateButtonImage(context)
	d.updateText()

	d.textButton.SetOnDown(func() {
		d.popupMenu.Open(context)
	})
	d.textButton.SetForcePressed(context.IsVisible(&d.popupMenu))

	appender.AppendChildWidgetWithPosition(&d.textButton, context.Position(d))

	d.popupMenu.SetOnClosed(func(index int) {
		if d.onValueChanged != nil {
			d.onValueChanged(index)
		}
	})
	d.popupMenu.SetCheckmarkIndex(d.SelectedItemIndex())

	pt := context.Position(d)
	pt.X -= listItemCheckmarkSize(context) + listItemTextAndImagePadding(context)
	pt.X = max(pt.X, 0)
	pt.Y -= listItemPadding(context)
	pt.Y += int((float64(context.Size(d).Y) - LineHeight(context)) / 2)
	pt.Y -= int(float64(d.popupMenu.SelectedItemIndex()) * LineHeight(context))
	pt.Y = max(pt.Y, 0)
	appender.AppendChildWidgetWithPosition(&d.popupMenu, pt)

	return nil
}

func (d *DropdownList) updateText() {
	if item, ok := d.popupMenu.SelectedItem(); ok {
		d.textButton.SetText(item.Text)
	} else {
		d.textButton.SetText("")
	}
}

func (d *DropdownList) SetItemsByStrings(items []string) {
	d.popupMenu.SetItemsByStrings(items)
	d.updateText()
}

func (d *DropdownList) SelectedItemIndex() int {
	return d.popupMenu.SelectedItemIndex()
}

func (d *DropdownList) SetSelectedItemIndex(index int) {
	d.popupMenu.SetSelectedItemIndex(index)
	d.updateText()
}

func (d *DropdownList) DefaultSize(context *guigui.Context) image.Point {
	// The button image affects the size.
	d.updateButtonImage(context)
	return context.Size(&d.textButton)
}
