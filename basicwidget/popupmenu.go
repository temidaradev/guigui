// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"image"

	"github.com/hajimehoshi/guigui"
)

type PopupMenu struct {
	guigui.DefaultWidget

	textList TextList
	popup    Popup

	onClosed func(index int)
}

func (p *PopupMenu) SetOnClosed(f func(index int)) {
	p.onClosed = f
}

func (p *PopupMenu) SetCheckmarkIndex(index int) {
	p.textList.SetCheckmarkIndex(index)
}

func (p *PopupMenu) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	p.textList.SetStyle(ListStyleMenu)
	p.textList.list.SetOnItemSelected(func(index int) {
		p.popup.Close()
		if p.onClosed != nil {
			p.onClosed(index)
		}
	})

	p.popup.SetContent(&p.textList)
	p.popup.SetCloseByClickingOutside(true)
	p.popup.SetOnClosed(func(reason PopupClosedReason) {
		if p.onClosed != nil {
			p.onClosed(p.textList.SelectedItemIndex())
		}
	})
	bounds := p.contentBounds(context)
	context.SetSize(&p.textList, bounds.Dx(), bounds.Dy())
	appender.AppendChildWidgetWithBounds(&p.popup, bounds)

	// Sync the visibility with the popup.
	// TODO: This is tricky. Refactor this. Perhaps, introducing Widget.PassThrough might be a good idea.
	if context.IsVisible(&p.popup) {
		context.Show(p)
	} else {
		context.Hide(p)
	}

	return nil
}

func (p *PopupMenu) contentBounds(context *guigui.Context) image.Rectangle {
	pos := context.Position(p)
	// textList's size is updated at Build so do not call guigui.Size to determine the content size.
	// Call DefaultSize instead.
	w, h := p.textList.DefaultSize(context)
	if h > 24*UnitSize(context) {
		h = 24 * UnitSize(context)
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

func (p *PopupMenu) Open(context *guigui.Context) {
	context.Show(p)
	p.popup.Open(context)
}

func (p *PopupMenu) Close() {
	p.popup.Close()
}

func (p *PopupMenu) SetItemsByStrings(items []string) {
	p.textList.SetItemsByStrings(items)
}

func (p *PopupMenu) SelectedItem() (TextListItem, bool) {
	return p.textList.SelectedItem()
}

func (p *PopupMenu) SelectedItemIndex() int {
	return p.textList.SelectedItemIndex()
}

func (p *PopupMenu) SetSelectedItemIndex(index int) {
	p.textList.SetSelectedItemIndex(index)
}
