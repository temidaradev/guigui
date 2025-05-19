// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Guigui Authors

package main

import (
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type ContentPanel struct {
	guigui.DefaultWidget

	panel   basicwidget.Panel
	content contentPanelContent
}

func (c *ContentPanel) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	context.SetSize(&c.content, context.Size(c))
	c.panel.SetContent(&c.content)

	appender.AppendChildWidgetWithBounds(&c.panel, context.Bounds(c))
	return nil
}

type contentPanelContent struct {
	guigui.DefaultWidget

	text basicwidget.Text
}

func (c *contentPanelContent) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	c.text.SetValue("Content panel: " + dummyText)
	c.text.SetAutoWrap(true)
	c.text.SetSelectable(true)
	u := basicwidget.UnitSize(context)
	appender.AppendChildWidgetWithBounds(&c.text, context.Bounds(c).Inset(u/2))
	return nil
}
