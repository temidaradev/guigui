// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Sidebar struct {
	guigui.DefaultWidget

	sidebar        basicwidget.Sidebar
	sidebarContent sidebarContent
}

func (s *Sidebar) SetModel(model *Model) {
	s.sidebarContent.SetModel(model)
}

func (s *Sidebar) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	context.SetSize(&s.sidebarContent, context.Size(s))
	s.sidebar.SetContent(&s.sidebarContent)

	appender.AppendChildWidgetWithBounds(&s.sidebar, context.Bounds(s))

	return nil
}

type sidebarContent struct {
	guigui.DefaultWidget

	list            basicwidget.List
	listItemWidgets []basicwidget.Text

	model *Model
}

func (s *sidebarContent) SetModel(model *Model) {
	s.model = model
}

func (s *sidebarContent) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
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
			text: "Texts",
			tag:  "texts",
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
		context.SetSize(t, image.Pt(guigui.DefaultSize, basicwidget.UnitSize(context)))
		if s.list.SelectedItemIndex() == i {
			t.SetColor(basicwidget.DefaultActiveListItemTextColor(context))
		} else {
			t.SetColor(basicwidget.DefaultTextColor(context))
		}
	}
	s.list.SetItems(listItems)
	s.list.SetSelectedItemByTag(s.model.Mode())
	s.list.SetOnItemSelected(func(index int) {
		item, ok := s.list.ItemByIndex(index)
		if !ok {
			s.model.SetMode("")
			return
		}
		tag, ok := item.Tag.(string)
		if !ok {
			s.model.SetMode("")
			return
		}
		s.model.SetMode(tag)
	})

	appender.AppendChildWidgetWithBounds(&s.list, context.Bounds(s))

	return nil
}
