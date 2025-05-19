// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Guigui Authors

package basicwidget

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget/internal/draw"
)

type List[T comparable] struct {
	guigui.DefaultWidget

	list            baseList[T]
	baseListItems   []baseListItem[T]
	listItems       []ListItem[T]
	listItemWidgets []listItemWidget[T]

	listItemHeightPlus1 int
}

/*type ListCallback struct {
	OnItemSelected    func(index int)
	OnItemEditStarted func(index int, str string) (from int)
	OnItemEditEnded   func(index int, str string)
	OnItemDropped     func(from, to int)
	OnContextMenu     func(index int, x, y int)
}*/

type ListItem[T comparable] struct {
	Text      string
	TextColor color.Color
	Header    bool
	Disabled  bool
	Border    bool
	Movable   bool
	ID        T
}

func (t *ListItem[T]) selectable() bool {
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

func (l *List[T]) SetItemBorderVisible(visible bool) {
	l.list.SetStripeVisible(visible)
}

func (l *List[T]) SetItemHeight(height int) {
	if l.listItemHeightPlus1 == height+1 {
		return
	}
	l.listItemHeightPlus1 = height + 1
	guigui.RequestRedraw(l)
}

func (l *List[T]) SetOnItemSelected(f func(index int)) {
	l.list.SetOnItemSelected(f)
}

func (l *List[T]) SetOnItemsMoved(f func(from, count, to int)) {
	l.list.SetOnItemsMoved(f)
}

func (l *List[T]) SetCheckmarkIndex(index int) {
	l.list.SetCheckmarkIndex(index)
}

func (l *List[T]) updateListItems() {
	l.listItemWidgets = adjustSliceSize(l.listItemWidgets, len(l.listItems))
	l.baseListItems = adjustSliceSize(l.baseListItems, len(l.listItems))

	for i, item := range l.listItems {
		l.listItemWidgets[i].setListItem(item)
		l.baseListItems[i] = l.listItemWidgets[i].listItem()
	}
	l.list.SetItems(l.baseListItems)
}

func (l *List[T]) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	context.SetSize(&l.list, context.Size(l))

	l.updateListItems()

	// To use HasFocusedChildWidget correctly, create the tree first.
	appender.AppendChildWidgetWithPosition(&l.list, context.Position(l))

	focused := context.IsFocusedOrHasFocusedChild(l)
	for i := range l.listItemWidgets {
		item := &l.listItemWidgets[i]
		item.text.SetBold(item.item.Header || l.list.style == ListStyleSidebar && l.SelectedItemIndex() == i)
		switch {
		case l.list.style == ListStyleNormal && focused && l.list.SelectedItemIndex() == i && item.selectable():
			item.text.SetColor(DefaultActiveListItemTextColor(context))
		case l.list.style == ListStyleSidebar && l.list.SelectedItemIndex() == i && item.selectable():
			item.text.SetColor(DefaultActiveListItemTextColor(context))
		case l.list.style == ListStyleMenu && l.list.isHoveringVisible() && l.list.HoveredItemIndex(context) == i && item.selectable():
			item.text.SetColor(DefaultActiveListItemTextColor(context))
		case !item.selectable() && !item.item.Header:
			item.text.SetColor(DefaultDisabledListItemTextColor(context))
		default:
			item.text.SetColor(item.item.TextColor)
		}

		if l.listItemHeightPlus1 > 0 {
			context.SetSize(item, image.Pt(guigui.DefaultSize, l.listItemHeightPlus1-1))
		} else {
			context.SetSize(item, image.Pt(guigui.DefaultSize, guigui.DefaultSize))
		}
	}

	return nil
}

func (l *List[T]) SelectedItemIndex() int {
	return l.list.SelectedItemIndex()
}

func (l *List[T]) SelectedItem() (ListItem[T], bool) {
	if l.list.SelectedItemIndex() < 0 || l.list.SelectedItemIndex() >= len(l.listItemWidgets) {
		return ListItem[T]{}, false
	}
	return l.listItemWidgets[l.list.SelectedItemIndex()].item, true
}

func (l *List[T]) ItemByIndex(index int) (ListItem[T], bool) {
	if index < 0 || index >= len(l.listItemWidgets) {
		return ListItem[T]{}, false
	}
	return l.listItemWidgets[index].item, true
}

func (l *List[T]) SetItemsByStrings(strs []string) {
	items := make([]ListItem[T], len(strs))
	for i, str := range strs {
		items[i].Text = str
	}
	l.SetItems(items)
}

func (l *List[T]) SetItems(items []ListItem[T]) {
	l.listItems = adjustSliceSize(l.listItems, len(items))
	copy(l.listItems, items)

	// Updating list items at Build might be too late, when the text list is not visible like a dropdown menu.
	// Update it here.
	l.updateListItems()
}

func (l *List[T]) ItemsCount() int {
	return len(l.listItemWidgets)
}

func (l *List[T]) ID(index int) any {
	return l.listItemWidgets[index].item.ID
}

func (l *List[T]) SelectItemByIndex(index int) {
	l.list.SelectItemByIndex(index)
}

func (l *List[T]) SelectItemByID(id T) {
	l.list.SelectItemByID(id)
}

func (l *List[T]) JumpToItemIndex(index int) {
	l.list.JumpToItemIndex(index)
}

func (l *List[T]) SetStyle(style ListStyle) {
	l.list.SetStyle(style)
}

func (l *List[T]) SetItemString(str string, index int) {
	l.listItemWidgets[index].item.Text = str
}

func (l *List[T]) DefaultSize(context *guigui.Context) image.Point {
	return l.list.DefaultSize(context)
}

type listItemWidget[T comparable] struct {
	guigui.DefaultWidget

	item ListItem[T]

	// TODO: Allow various widgets.
	text Text
}

func (l *listItemWidget[T]) setListItem(listItem ListItem[T]) {
	l.item = listItem
	l.text.SetValue(listItem.Text)
}

func (l *listItemWidget[T]) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	p := context.Position(l)
	if l.item.Header {
		p.X += UnitSize(context) / 2
		context.SetSize(&l.text, context.Size(l).Add(image.Pt(-UnitSize(context), 0)))
	} else {
		context.SetSize(&l.text, context.Size(l))
	}
	l.text.SetValue(l.item.Text)
	l.text.SetVerticalAlign(VerticalAlignMiddle)
	appender.AppendChildWidgetWithPosition(&l.text, p)

	return nil
}

func (l *listItemWidget[T]) Draw(context *guigui.Context, dst *ebiten.Image) {
	if l.item.Border {
		p := context.Position(l)
		s := context.Size(l)
		x0 := float32(p.X)
		x1 := float32(p.X + s.X)
		y := float32(p.Y) + float32(s.Y)/2
		width := float32(1 * context.Scale())
		vector.StrokeLine(dst, x0, y, x1, y, width, draw.Color(context.ColorMode(), draw.ColorTypeBase, 0.8), false)
		return
	}
	if l.item.Header {
		bounds := context.Bounds(l)
		draw.DrawRoundedRect(context, dst, bounds, draw.Color(context.ColorMode(), draw.ColorTypeBase, 0.6), RoundedCornerRadius(context))
	}
}

func (l *listItemWidget[T]) DefaultSize(context *guigui.Context) image.Point {
	// Assume that every item can use a bold font.
	w := l.text.boldTextSize(context).X
	if l.item.Border {
		return image.Pt(w, UnitSize(context)/2)
	}
	return image.Pt(w, int(LineHeight(context)))
}

/*func (t *textListItemWidget[T]) index() int {
	for i, tt := range t.textList.textListItemWidgets {
		if tt == t {
			return i
		}
	}
	return -1
}*/

func (l *listItemWidget[T]) selectable() bool {
	return l.item.selectable() && !l.item.Border
}

func (l *listItemWidget[T]) listItem() baseListItem[T] {
	return baseListItem[T]{
		Content:    l,
		Selectable: l.selectable(),
		Movable:    l.item.Movable,
		ID:         l.item.ID,
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
	var tf *TextInput
	tf = NewTextInput(t.settings, &TextInputCallback{
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
