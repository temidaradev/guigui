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

func (s *Sidebar) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	context.SetSize(&s.sidebar, sidebarWidth(context), guigui.AutoSize)
	s.sidebar.SetContent(&s.sidebarContent)
	context.SetPosition(&s.sidebar, context.Position(s))
	appender.AppendChildWidget(&s.sidebar)

	return nil
}

func (s *Sidebar) DefaultSize(context *guigui.Context) (int, int) {
	_, h := context.Size(guigui.Parent(s))
	return sidebarWidth(context), h
}

func (s *Sidebar) SelectedItemTag() string {
	return s.sidebarContent.SelectedItemTag()
}

func (s *Sidebar) SetSelectedItemIndex(context *guigui.Context, index int) {
	s.sidebarContent.SetSelectedItemIndex(context, index)
}

type sidebarContent struct {
	guigui.DefaultWidget

	list            basicwidget.List
	listItemWidgets []basicwidget.Text

	initOnce sync.Once
}

func (s *sidebarContent) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	s.list.SetStyle(context, basicwidget.ListStyleSidebar)

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
		t.SetText(context, item.text)
		listItems[i] = basicwidget.ListItem{
			Content:    t,
			Selectable: true,
			Tag:        item.tag,
		}
	}
	for i := range s.listItemWidgets {
		t := &s.listItemWidgets[i]
		t.SetVerticalAlign(context, basicwidget.VerticalAlignMiddle)
		context.SetSize(t, guigui.AutoSize, basicwidget.UnitSize(context))
		if s.list.SelectedItemIndex() == i {
			t.SetColor(context, basicwidget.DefaultActiveListItemTextColor(context))
		} else {
			t.SetColor(context, basicwidget.DefaultTextColor(context))
		}
	}
	s.list.SetItems(listItems)

	s.initOnce.Do(func() {
		s.list.SetSelectedItemIndex(context, 0)
	})

	_, h := context.Size(s)
	context.SetSize(&s.list, sidebarWidth(context), h)
	context.SetPosition(&s.list, context.Position(s))
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

func (s *sidebarContent) SetSelectedItemIndex(context *guigui.Context, index int) {
	s.list.SetSelectedItemIndex(context, index)
}
