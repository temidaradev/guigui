// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"image/color"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget/internal/draw"
)

type TextList struct {
	guigui.DefaultWidget

	list                List
	textListItemWidgets []*textListItemWidget
}

/*type TextListCallback struct {
	OnItemSelected    func(index int)
	OnItemEditStarted func(index int, str string) (from int)
	OnItemEditEnded   func(index int, str string)
	OnItemDropped     func(from, to int)
	OnContextMenu     func(index int, x, y int)
}*/

type TextListItem struct {
	Text      string
	DummyText string
	Color     color.Color
	Header    bool
	Disabled  bool
	Border    bool
	Draggable bool
	Tag       any
}

func (t *TextListItem) selectable() bool {
	return !t.Header && !t.Disabled && !t.Border
}

/*func NewTextList(settings *model.Settings, callback *TextListCallback) *TextList {
	t := &TextList{
		settings: settings,
		callback: callback,
	}
	t.list = NewList(settings, &ListCallback{
		OnItemSelected: func(index int) {
			if callback != nil && callback.OnItemSelected != nil {
				callback.OnItemSelected(index)
			}
		},
		OnItemEditStarted: func(index int) {
			if index < 0 || index >= len(t.textListItems) {
				return
			}
			if !t.textListItems[index].selectable() {
				return
			}
			item, ok := t.textListItems[index].listItem.WidgetWithHeight.(*textListTextItem)
			if !ok {
				return
			}
			if callback != nil && callback.OnItemEditStarted != nil {
				item.edit(callback.OnItemEditStarted(index, item.textListItem.Text))
			}
		},
		OnItemDropped: func(from int, to int) {
			if callback != nil && callback.OnItemDropped != nil {
				callback.OnItemDropped(from, to)
			}
		},
		OnContextMenu: func(index int, x, y int) {
			if callback != nil && callback.OnContextMenu != nil {
				callback.OnContextMenu(index, x, y)
			}
		},
	})
	t.AddChild(t.list, &view.IdentityLayouter{})
	return t
}*/

func (t *TextList) SetCheckmarkIndex(context *guigui.Context, index int) {
	t.list.SetCheckmarkIndex(context, index)
}

func (t *TextList) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	w, h := context.Size(t)
	context.SetSize(&t.list, w, h)

	// To use HasFocusedChildWidget correctly, create the tree first.
	context.SetPosition(&t.list, context.Position(t))
	appender.AppendChildWidget(&t.list)

	for i, item := range t.textListItemWidgets {
		item.text.SetBold(context, item.textListItem.Header)
		if t.list.style != ListStyleMenu && context.HasFocusedChildWidget(t) && t.list.SelectedItemIndex() == i && item.selectable() {
			item.text.SetColor(context, DefaultActiveListItemTextColor(context))
		} else if t.list.style == ListStyleMenu && t.list.isHoveringVisible() && t.list.HoveredItemIndex(context) == i && item.selectable() {
			item.text.SetColor(context, DefaultActiveListItemTextColor(context))
		} else if !item.selectable() && !item.textListItem.Header {
			item.text.SetColor(context, DefaultDisabledListItemTextColor(context))
		} else {
			item.text.SetColor(context, item.textListItem.Color)
		}
	}

	return nil
}

func (t *TextList) SelectedItemIndex() int {
	return t.list.SelectedItemIndex()
}

func (t *TextList) SelectedItem() (TextListItem, bool) {
	if t.list.SelectedItemIndex() < 0 || t.list.SelectedItemIndex() >= len(t.textListItemWidgets) {
		return TextListItem{}, false
	}
	return t.textListItemWidgets[t.list.SelectedItemIndex()].textListItem, true
}

func (t *TextList) ItemAt(index int) (TextListItem, bool) {
	if index < 0 || index >= len(t.textListItemWidgets) {
		return TextListItem{}, false
	}
	return t.textListItemWidgets[index].textListItem, true
}

func (t *TextList) SetItemsByStrings(context *guigui.Context, strs []string) {
	items := make([]TextListItem, len(strs))
	for i, str := range strs {
		items[i].Text = str
	}
	t.SetItems(context, items)
}

func (t *TextList) SetItems(context *guigui.Context, items []TextListItem) {
	if cap(t.textListItemWidgets) < len(items) {
		t.textListItemWidgets = append(t.textListItemWidgets, make([]*textListItemWidget, len(items)-cap(t.textListItemWidgets))...)
	}
	t.textListItemWidgets = t.textListItemWidgets[:len(items)]

	listItems := make([]ListItem, len(items))
	for i, item := range items {
		if t.textListItemWidgets[i] == nil {
			t.textListItemWidgets[i] = newTextListItemWidget(context, t, item)
		} else {
			t.textListItemWidgets[i].setTextListItem(context, item)
		}
		listItems[i] = t.textListItemWidgets[i].listItem()
	}
	t.list.SetItems(listItems)
}

func (t *TextList) ItemsCount() int {
	return len(t.textListItemWidgets)
}

func (t *TextList) Tag(index int) any {
	return t.textListItemWidgets[index].textListItem.Tag
}

func (t *TextList) SetSelectedItemIndex(context *guigui.Context, index int) {
	t.list.SetSelectedItemIndex(context, index)
}

func (t *TextList) JumpToItemIndex(index int) {
	t.list.JumpToItemIndex(index)
}

func (t *TextList) SetStyle(context *guigui.Context, style ListStyle) {
	t.list.SetStyle(context, style)
}

func (t *TextList) SetItemString(str string, index int) {
	t.textListItemWidgets[index].textListItem.Text = str
}

func (t *TextList) AppendItem(item TextListItem) {
	t.AddItem(item, len(t.textListItemWidgets))
}

func (t *TextList) AddItem(item TextListItem, index int) {
	t.textListItemWidgets = slices.Insert(t.textListItemWidgets, index, &textListItemWidget{
		textList:     t,
		textListItem: item,
	})
	t.list.AddItem(t.textListItemWidgets[index].listItem(), index)
}

func (t *TextList) RemoveItem(index int) {
	t.textListItemWidgets = slices.Delete(t.textListItemWidgets, index, index+1)
	t.list.RemoveItem(index)
}

func (t *TextList) MoveItem(from, to int) {
	moveItemInSlice(t.textListItemWidgets, from, 1, to)
	t.list.MoveItem(from, to)
}

func (t *TextList) DefaultSize(context *guigui.Context) (int, int) {
	return t.list.DefaultSize(context)
}

type textListItemWidget struct {
	guigui.DefaultWidget

	textList     *TextList
	textListItem TextListItem

	text Text
}

func newTextListItemWidget(context *guigui.Context, textList *TextList, textListItem TextListItem) *textListItemWidget {
	t := &textListItemWidget{
		textList:     textList,
		textListItem: textListItem,
	}
	t.text.SetText(context, t.textString())
	return t
}

func (t *textListItemWidget) setTextListItem(context *guigui.Context, textListItem TextListItem) {
	t.textListItem = textListItem
	t.text.SetText(context, t.textString())
}

func (t *textListItemWidget) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	p := context.Position(t)
	if t.textListItem.Header {
		p.X += UnitSize(context) / 2
		w, h := context.Size(t)
		context.SetSize(&t.text, w-UnitSize(context), h)
	}
	t.text.SetText(context, t.textString())
	t.text.SetVerticalAlign(context, VerticalAlignMiddle)
	context.SetPosition(&t.text, p)
	appender.AppendChildWidget(&t.text)

	return nil
}

func (t *textListItemWidget) textString() string {
	if t.textListItem.DummyText != "" {
		return t.textListItem.DummyText
	}
	return t.textListItem.Text
}

func (t *textListItemWidget) Draw(context *guigui.Context, dst *ebiten.Image) {
	if t.textListItem.Border {
		p := context.Position(t)
		w, h := context.Size(t)
		x0 := float32(p.X)
		x1 := float32(p.X + w)
		y := float32(p.Y) + float32(h)/2
		width := float32(1 * context.Scale())
		vector.StrokeLine(dst, x0, y, x1, y, width, draw.Color(context.ColorMode(), draw.ColorTypeBase, 0.8), false)
		return
	}
	if t.textListItem.Header {
		bounds := context.Bounds(t)
		draw.DrawRoundedRect(context, dst, bounds, draw.Color(context.ColorMode(), draw.ColorTypeBase, 0.6), RoundedCornerRadius(context))
	}
}

func (t *textListItemWidget) DefaultSize(context *guigui.Context) (int, int) {
	w, _ := t.text.TextSize(context)
	if t.textListItem.Border {
		return w, UnitSize(context) / 2
	}
	return w, int(LineHeight(context))
}

/*func (t *textListItemWidget) index() int {
	for i, tt := range t.textList.textListItemWidgets {
		if tt == t {
			return i
		}
	}
	return -1
}*/

func (t *textListItemWidget) selectable() bool {
	return t.textListItem.selectable() && !t.textListItem.Border
}

func (t *textListItemWidget) listItem() ListItem {
	return ListItem{
		Content:    t,
		Selectable: t.selectable(),
		Wide:       t.textListItem.Header,
		Draggable:  t.textListItem.Draggable,
	}
}

/*func (t *textListTextItem) edit(from int) {
	t.label.Hide()
	t0 := t.textListItem.Text[:from]
	var l0 *Label
	if t0 != "" {
		l0 = NewLabel(t.settings)
		l0.SetText(t0)
		t.AddChild(l0, &view.IdentityLayouter{})
	}
	var tf *TextField
	tf = NewTextField(t.settings, &TextFieldCallback{
		OnTextUpdated: func(value string) {
			t.textListItem.Text = t0 + value
		},
		OnTextConfirmed: func(value string) {
			t.textListItem.Text = t0 + value

			t.label.SetText(t.labelText())
			if l0 != nil {
				l0.RemoveSelf()
			}
			tf.RemoveSelf()
			t.label.Show()

			if t.textList.callback != nil && t.textList.callback.OnItemEditEnded != nil {
				t.textList.callback.OnItemEditEnded(t.index(), t0+value)
			}
		},
	})
	tf.SetText(t.textListItem.Text[from:])
	tf.SetHorizontalAlign(HorizontalAlignStart)
	t.AddChild(tf, view.LayoutFunc(func(args view.WidgetArgs) image.Rectangle {
		bounds := args.Bounds
		if l0 != nil {
			bounds.Min.X += l0.Width(args.Scale)
		}
		return bounds
	}))
	tf.SelectAll()
	tf.Focus()
}*/
