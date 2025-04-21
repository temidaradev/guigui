// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"image"
	"log/slog"

	"github.com/hajimehoshi/guigui"
)

type DropdownList[T comparable] struct {
	guigui.DefaultWidget

	textButton TextButton
	popupMenu  PopupMenu[T]

	onValueChanged func(index int)
}

func (d *DropdownList[T]) SetOnValueChanged(f func(index int)) {
	d.onValueChanged = f
}

func (d *DropdownList[T]) updateButtonImage(context *guigui.Context) {
	img, err := theResourceImages.Get("unfold_more", context.ColorMode())
	if err != nil {
		slog.Error(err.Error())
		return
	}
	d.textButton.SetImage(img)
}

func (d *DropdownList[T]) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
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

func (d *DropdownList[T]) updateText() {
	if item, ok := d.popupMenu.SelectedItem(); ok {
		d.textButton.SetText(item.Text)
	} else {
		d.textButton.SetText("")
	}
}

func (d *DropdownList[T]) SetItems(items []PopupMenuItem[T]) {
	d.popupMenu.SetItems(items)
	d.updateText()
}

func (d *DropdownList[T]) SetItemsByStrings(items []string) {
	d.popupMenu.SetItemsByStrings(items)
	d.updateText()
}

func (d *DropdownList[T]) SelectedItem() (PopupMenuItem[T], bool) {
	return d.popupMenu.SelectedItem()
}

func (d *DropdownList[T]) ItemByIndex(index int) (PopupMenuItem[T], bool) {
	return d.popupMenu.ItemByIndex(index)
}

func (d *DropdownList[T]) SelectedItemIndex() int {
	return d.popupMenu.SelectedItemIndex()
}

func (d *DropdownList[T]) SetSelectedItemIndex(index int) {
	d.popupMenu.SetSelectedItemIndex(index)
	d.updateText()
}

func (d *DropdownList[T]) SetSelectedItemByTag(tag T) {
	d.popupMenu.SetSelectedItemByTag(tag)
	d.updateText()
}

func (d *DropdownList[T]) DefaultSize(context *guigui.Context) image.Point {
	// The button image affects the size.
	d.updateButtonImage(context)
	return context.Size(&d.textButton)
}
