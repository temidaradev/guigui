// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"sync"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Sidebar struct {
	guigui.DefaultWidget

	sidebar        basicwidget.Sidebar
	sidebarContent sidebarContent
}

func sidebarWidth(context *guigui.Context) int {
	return 8 * basicwidget.UnitSize(context)
}

func (s *Sidebar) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	_, h := s.Size(context)
	s.sidebar.SetSize(context, sidebarWidth(context), h)
	s.sidebar.SetContent(&s.sidebarContent)
	guigui.SetPosition(&s.sidebar, guigui.Position(s))
	appender.AppendChildWidget(&s.sidebar)

	return nil
}

func (s *Sidebar) Size(context *guigui.Context) (int, int) {
	_, h := guigui.Parent(s).Size(context)
	return sidebarWidth(context), h
}

func (s *Sidebar) SelectedItemTag() string {
	return s.sidebarContent.SelectedItemTag()
}

func (s *Sidebar) SetSelectedItemIndex(index int) {
	s.sidebarContent.SetSelectedItemIndex(index)
}

type sidebarContent struct {
	guigui.DefaultWidget

	list            basicwidget.List
	listItemWidgets []basicwidget.Text

	initOnce sync.Once
}

func (s *sidebarContent) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {

	s.list.SetStyle(basicwidget.ListStyleSidebar)

	type item struct {
		text string
		tag  string
	}
	items := []item{
		{
			text: "Settings",
			tag:  "settings",
		},
		{
			text: "Basic",
			tag:  "basic",
		},
		{
			text: "Buttons",
			tag:  "buttons",
		},
		{
			text: "Lists",
			tag:  "lists",
		},
		{
			text: "Popups",
			tag:  "popups",
		},
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

	_, h := s.Size(context)
	s.list.SetWidth(sidebarWidth(context))
	s.list.SetHeight(h)
	guigui.SetPosition(&s.list, guigui.Position(s))
	appender.AppendChildWidget(&s.list)

	return nil
}

func (s *sidebarContent) SelectedItemTag() string {
	item, ok := s.list.SelectedItem()
	if !ok {
		return ""
	}
	return item.Tag.(string)
}

func (s *sidebarContent) SetSelectedItemIndex(index int) {
	s.list.SetSelectedItemIndex(index)
}
