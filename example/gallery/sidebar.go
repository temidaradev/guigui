// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"
	"sync"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Sidebar struct {
	guigui.DefaultWidget

	sidebar         basicwidget.Sidebar
	list            basicwidget.List
	listItemWidgets []basicwidget.Text

	initOnce sync.Once
}

func sidebarWidth(context *guigui.Context) int {
	return 8 * basicwidget.UnitSize(context)
}

func (s *Sidebar) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	_, h := s.Size(context)
	s.sidebar.SetSize(context, sidebarWidth(context), h)
	s.sidebar.SetContent(context, func(context *guigui.Context, childAppender *basicwidget.ContainerChildWidgetAppender, offsetX, offsetY float64) {
		s.list.SetWidth(sidebarWidth(context))
		s.list.SetHeight(h)
		guigui.SetPosition(&s.list, guigui.Position(s).Add(image.Pt(int(offsetX), int(offsetY))))
		childAppender.AppendChildWidget(&s.list)
	})
	guigui.SetPosition(&s.sidebar, guigui.Position(s))
	appender.AppendChildWidget(&s.sidebar)

	s.list.SetStyle(basicwidget.ListStyleSidebar)

	type item struct {
		text string
		tag  string
	}
	items := []item{
		{"Settings", "settings"},
		{"Basic", "basic"},
		{"Buttons", "buttons"},
		{"Lists", "lists"},
		{"Popups", "popups"},
	}

	if len(s.listItemWidgets) == 0 {
		s.listItemWidgets = make([]basicwidget.Text, len(items))
	}
	listItems := make([]basicwidget.ListItem, len(items))
	for i, item := range items {
		t := &s.listItemWidgets[i]
		t.SetText(item.text)
		listItems[i] = basicwidget.ListItem{
			Content:    t,
			Selectable: true,
			Tag:        item.tag,
		}
	}
	for i := range s.listItemWidgets {
		t := &s.listItemWidgets[i]
		t.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
		t.SetHeight(basicwidget.UnitSize(context))
		if s.list.SelectedItemIndex() == i {
			t.SetColor(basicwidget.DefaultActiveListItemTextColor(context))
		} else {
			t.SetColor(basicwidget.DefaultTextColor(context))
		}
	}
	s.list.SetItems(listItems)

	s.initOnce.Do(func() {
		s.list.SetSelectedItemIndex(0)
	})

	return nil
}

func (s *Sidebar) Size(context *guigui.Context) (int, int) {
	_, h := guigui.Parent(s).Size(context)
	return sidebarWidth(context), h
}

func (s *Sidebar) SelectedItemTag() string {
	item, ok := s.list.SelectedItem()
	if !ok {
		return ""
	}
	return item.Tag.(string)
}

func (s *Sidebar) SetSelectedItemIndex(index int) {
	s.list.SetSelectedItemIndex(index)
}
