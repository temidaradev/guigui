// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Guigui Authors

package basicwidget

import (
	"image"
	"image/color"
	"log/slog"

	"github.com/hajimehoshi/guigui"
)

type DropdownListItem[T comparable] struct {
	Text      string
	TextColor color.Color
	Header    bool
	Disabled  bool
	Border    bool
	ID        T
}

type DropdownList[T comparable] struct {
	guigui.DefaultWidget

	button    Button
	popupMenu PopupMenu[T]

	onItemSelected func(index int)
}

func (d *DropdownList[T]) SetOnItemSelected(f func(index int)) {
	d.onItemSelected = f
}

func (d *DropdownList[T]) updateButtonImage(context *guigui.Context) {
	img, err := theResourceImages.Get("unfold_more", context.ColorMode())
	if err != nil {
		slog.Error(err.Error())
		return
	}
	d.button.SetIcon(img)
}

func (d *DropdownList[T]) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	d.updateButtonImage(context)
	d.updateText()

	d.button.SetOnDown(func() {
		d.popupMenu.Open(context)
	})
	d.button.setKeepPressed(d.popupMenu.IsOpen())
	d.button.SetIconAlign(IconAlignEnd)

	appender.AppendChildWidgetWithPosition(&d.button, context.Position(d))

	d.popupMenu.SetOnItemSelected(func(index int) {
		if d.onItemSelected != nil {
			d.onItemSelected(index)
		}
	})
	if !d.popupMenu.IsOpen() {
		d.popupMenu.SetCheckmarkIndex(d.SelectedItemIndex())
	}

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
		d.button.SetText(item.Text)
	} else {
		d.button.SetText("")
	}
}

func (d *DropdownList[T]) SetItems(items []DropdownListItem[T]) {
	var popupMenuItems []PopupMenuItem[T]
	for _, item := range items {
		popupMenuItems = append(popupMenuItems, PopupMenuItem[T](item))
	}
	d.popupMenu.SetItems(popupMenuItems)
	d.updateText()
}

func (d *DropdownList[T]) SetItemsByStrings(items []string) {
	d.popupMenu.SetItemsByStrings(items)
	d.updateText()
}

func (d *DropdownList[T]) SelectedItem() (DropdownListItem[T], bool) {
	item, ok := d.popupMenu.SelectedItem()
	if !ok {
		return DropdownListItem[T]{}, false
	}
	return DropdownListItem[T](item), true
}

func (d *DropdownList[T]) ItemByIndex(index int) (DropdownListItem[T], bool) {
	item, ok := d.popupMenu.ItemByIndex(index)
	if !ok {
		return DropdownListItem[T]{}, false
	}
	return DropdownListItem[T](item), true
}

func (d *DropdownList[T]) SelectedItemIndex() int {
	return d.popupMenu.SelectedItemIndex()
}

func (d *DropdownList[T]) SelectItemByIndex(index int) {
	d.popupMenu.SelectItemByIndex(index)
	d.updateText()
}

func (d *DropdownList[T]) SelectItemByID(id T) {
	d.popupMenu.SelectItemByID(id)
	d.updateText()
}

func (d *DropdownList[T]) DefaultSize(context *guigui.Context) image.Point {
	// The button image affects the size.
	d.updateButtonImage(context)
	return context.Size(&d.button)
}

func (d *DropdownList[T]) IsPopupOpen() bool {
	return d.popupMenu.IsOpen()
}
