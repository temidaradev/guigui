// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"slices"
)

type tagger[Tag comparable] interface {
	tag() Tag
}

type abstractList[Tag comparable, Item tagger[Tag]] struct {
	items           []Item
	selectedIndices []int

	onItemSelected func(index int)
}

func (a *abstractList[Tag, Item]) SetOnItemSelected(f func(index int)) {
	a.onItemSelected = f
}

func (a *abstractList[Tag, Item]) SetItems(items []Item) {
	a.items = adjustSliceSize(items, len(items))
	copy(a.items, items)
}

func (a *abstractList[Tag, Item]) ItemCount() int {
	return len(a.items)
}

func (c *abstractList[Tag, Item]) ItemByIndex(index int) (Item, bool) {
	if index < 0 || index >= len(c.items) {
		var item Item
		return item, false
	}
	return c.items[index], true
}

func (c *abstractList[Tag, Item]) SelectItemByIndex(index int) bool {
	if index < 0 || index >= len(c.items) {
		if len(c.selectedIndices) == 0 {
			return false
		}
		c.selectedIndices = c.selectedIndices[:0]
		return true
	}

	if len(c.selectedIndices) == 1 && c.selectedIndices[0] == index {
		return false
	}

	selected := slices.Contains(c.selectedIndices, index)
	c.selectedIndices = adjustSliceSize(c.selectedIndices, 1)
	c.selectedIndices[0] = index
	if !selected {
		if c.onItemSelected != nil {
			c.onItemSelected(index)
		}
	}
	return true
}

func (c *abstractList[Tag, Item]) SelectItemByTag(tag Tag) bool {
	idx := slices.IndexFunc(c.items, func(item Item) bool {
		return item.tag() == tag
	})
	return c.SelectItemByIndex(idx)
}

func (c *abstractList[Tag, Item]) SelectedItem() (Item, bool) {
	if len(c.selectedIndices) == 0 {
		var item Item
		return item, false
	}
	return c.items[c.selectedIndices[0]], true
}

func (c *abstractList[Tag, Item]) SelectedItemIndex() int {
	if len(c.selectedIndices) == 0 {
		return -1
	}
	return c.selectedIndices[0]
}
