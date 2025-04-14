// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"image"

	"github.com/hajimehoshi/guigui"
)

type PopupMenu struct {
	guigui.DefaultWidget

	popup   Popup
	content popupMenuContent

	onClosed func(index int)
}

func (p *PopupMenu) SetOnClosed(f func(index int)) {
	p.onClosed = f
}

func (p *PopupMenu) SetCheckmarkIndex(index int) {
	p.content.setCheckmarkIndex(index)
}

func (p *PopupMenu) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	p.content.popupMenu = p

	// Do not set a text list as a content directly, or the text list size will be modified.
	p.popup.SetContent(&p.content)
	p.popup.SetCloseByClickingOutside(true)
	p.popup.SetOnClosed(func(reason PopupClosedReason) {
		if p.onClosed != nil {
			p.onClosed(p.content.selectedItemIndex())
		}
	})
	p.popup.SetContentBounds(p.content.contentBounds(context))
	appender.AppendChildWidget(&p.popup)

	// Sync the visibility with the popup.
	// TODO: This is tricky. Refactor this. Perhaps, introducing Widget.PassThrough might be a good idea.
	if context.IsVisible(&p.popup) {
		context.Show(p)
	} else {
		context.Hide(p)
	}

	return nil
}

func (p *PopupMenu) ZDelta() int {
	return popupZ
}

func (p *PopupMenu) Open(context *guigui.Context) {
	context.Show(p)
	p.popup.Open(context)
}

func (p *PopupMenu) Close() {
	p.popup.Close()
}

func (p *PopupMenu) SetItemsByStrings(items []string) {
	p.content.setItemsByStrings(items)
}

func (p *PopupMenu) SelectedItem() (TextListItem, bool) {
	return p.content.selectedItem()
}

func (p *PopupMenu) SelectedItemIndex() int {
	return p.content.selectedItemIndex()
}

func (p *PopupMenu) SetSelectedItemIndex(index int) {
	p.content.setSelectedItemIndex(index)
}

type popupMenuContent struct {
	guigui.DefaultWidget

	popupMenu *PopupMenu

	textList TextList
}

func (p *popupMenuContent) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	p.textList.SetStyle(ListStyleMenu)
	p.textList.SetOnItemSelected(func(index int) {
		p.popupMenu.Close()
		if p.popupMenu.onClosed != nil {
			p.popupMenu.onClosed(index)
		}
	})
	pt := context.Position(p)
	context.SetPosition(&p.textList, pt)
	appender.AppendChildWidget(&p.textList)
	return nil
}

func (p *popupMenuContent) contentBounds(context *guigui.Context) image.Rectangle {
	pos := context.Position(p.popupMenu)
	w, h := context.Size(&p.textList)
	if h > 24*UnitSize(context) {
		h = 24 * UnitSize(context)
		context.SetSize(&p.textList, guigui.AutoSize, h)
	}
	r := image.Rectangle{
		Min: pos,
		Max: pos.Add(image.Pt(w, h)),
	}
	aw, ah := context.AppSize()
	if r.Max.X > aw {
		r.Min.X = aw - w
		r.Max.X = aw
	}
	if r.Min.X < 0 {
		r.Min.X = 0
		r.Max.X = w
	}
	if r.Max.Y > ah {
		r.Min.Y = ah - h
		r.Max.Y = ah
	}
	if r.Min.Y < 0 {
		r.Min.Y = 0
		r.Max.Y = h
	}
	return r
}

func (p *popupMenuContent) setCheckmarkIndex(index int) {
	p.textList.SetCheckmarkIndex(index)
}

func (p *popupMenuContent) setItemsByStrings(items []string) {
	p.textList.SetItemsByStrings(items)
}

func (p *popupMenuContent) selectedItem() (TextListItem, bool) {
	return p.textList.SelectedItem()
}

func (p *popupMenuContent) selectedItemIndex() int {
	return p.textList.SelectedItemIndex()
}

func (p *popupMenuContent) setSelectedItemIndex(index int) {
	p.textList.SetSelectedItemIndex(index)
}
