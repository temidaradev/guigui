// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"log/slog"

	"github.com/hajimehoshi/guigui"
)

type DropdownList struct {
	guigui.DefaultWidget

	textButton   TextButton
	popupMenu    PopupMenu
	onceImageSet bool
}

func (d *DropdownList) SetOnValueChanged(f func(index int)) {
	d.popupMenu.SetOnClosed(func(index int) {
		f(index)
	})
}

func (d *DropdownList) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	d.updateText()
	// TODO: A widget initializer might be better.
	if !d.onceImageSet {
		img, err := theResourceImages.Get("unfold_more", context.ColorMode())
		if err != nil {
			slog.Error(err.Error())
			return
		}
		d.textButton.SetImage(img)
		d.onceImageSet = true
	}

	d.popupMenu.SetHasCheckmark(true)
	d.textButton.SetOnDown(func() {
		pt := guigui.Position(d)
		pt.X -= int(LineHeight(context)) + listItemTextAndImagePadding(context)
		pt.X = max(pt.X, 0)
		pt.Y -= listItemPadding(context)
		_, y := d.Size(context)
		pt.Y += int((float64(y) - LineHeight(context)) / 2)
		pt.Y -= int(float64(d.popupMenu.SelectedItemIndex()) * LineHeight(context))
		pt.Y = max(pt.Y, 0)
		// TODO: Chaning the position here might be too late here.
		// A glitch is visible when the dropdown list is reopened.
		guigui.SetPosition(&d.popupMenu, pt)
		d.popupMenu.Open(context)
	})

	guigui.SetPosition(&d.textButton, guigui.Position(d))
	appender.AppendChildWidget(&d.textButton)
	appender.AppendChildWidget(&d.popupMenu)
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

func (d *DropdownList) Size(context *guigui.Context) (int, int) {
	w, h := d.textButton.Size(context)
	// TODO: This is a little hacky. Refactor this.
	if !d.onceImageSet {
		w += int(LineHeight(context)) + textButtonTextAndImagePadding(context)
	}
	return w, h
}
