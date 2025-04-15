// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Popups struct {
	guigui.DefaultWidget

	forms                              [2]basicwidget.Form
	blurBackgroundText                 basicwidget.Text
	blurBackgroundToggleButton         basicwidget.ToggleButton
	closeByClickingOutsideText         basicwidget.Text
	closeByClickingOutsideToggleButton basicwidget.ToggleButton
	showButton                         basicwidget.TextButton

	contextMenuPopupText          basicwidget.Text
	contextMenuPopupClickHereText basicwidget.Text

	simplePopup        basicwidget.Popup
	simplePopupContent simplePopupContent

	contextMenuPopup basicwidget.PopupMenu
}

func (p *Popups) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	p.blurBackgroundText.SetText("Blur Background")
	p.closeByClickingOutsideText.SetText("Close by Clicking Outside")
	p.showButton.SetText("Show")
	p.showButton.SetOnUp(func() {
		p.simplePopup.Open(context)
	})

	u := float64(basicwidget.UnitSize(context))

	w, _ := context.Size(p)
	context.SetSize(&p.forms[0], w-int(1*u), guigui.DefaultSize)
	p.forms[0].SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &p.blurBackgroundText,
			SecondaryWidget: &p.blurBackgroundToggleButton,
		},
		{
			PrimaryWidget:   &p.closeByClickingOutsideText,
			SecondaryWidget: &p.closeByClickingOutsideToggleButton,
		},
		{
			SecondaryWidget: &p.showButton,
		},
	})
	pt := context.Position(p).Add(image.Pt(int(0.5*u), int(0.5*u)))
	appender.AppendChildWidgetWithPosition(&p.forms[0], pt)

	p.contextMenuPopupText.SetText("Context Menu")
	p.contextMenuPopupClickHereText.SetText("Click Here by the Right Button")

	context.SetSize(&p.forms[1], w-int(1*u), guigui.DefaultSize)
	p.forms[1].SetItems([]*basicwidget.FormItem{
		{
			PrimaryWidget:   &p.contextMenuPopupText,
			SecondaryWidget: &p.contextMenuPopupClickHereText,
		},
	})
	_, h := context.Size(&p.forms[0])
	pt.Y += h + int(0.5*u)
	appender.AppendChildWidgetWithPosition(&p.forms[1], pt)

	p.simplePopupContent.popup = &p.simplePopup
	p.simplePopup.SetContent(&p.simplePopupContent)
	contentWidth := int(12 * u)
	contentHeight := int(6 * u)
	bounds := context.Bounds(&p.simplePopup)
	contentPosition := image.Point{
		X: bounds.Min.X + (bounds.Dx()-contentWidth)/2,
		Y: bounds.Min.Y + (bounds.Dy()-contentHeight)/2,
	}
	context.SetSize(&p.simplePopupContent, contentWidth, contentHeight)
	p.simplePopup.SetBackgroundBlurred(p.blurBackgroundToggleButton.Value())
	p.simplePopup.SetCloseByClickingOutside(p.closeByClickingOutsideToggleButton.Value())
	p.simplePopup.SetAnimationDuringFade(true)
	appender.AppendChildWidgetWithBounds(&p.simplePopup, image.Rectangle{
		Min: contentPosition,
		Max: contentPosition.Add(image.Pt(contentWidth, contentHeight)),
	})

	p.contextMenuPopup.SetItemsByStrings([]string{"Item 1", "Item 2", "Item 3"})
	// A context menu's position is updated at HandlePointingInput.
	appender.AppendChildWidget(&p.contextMenuPopup)

	return nil
}

func (p *Popups) HandlePointingInput(context *guigui.Context) guigui.HandleInputResult {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		pt := image.Pt(ebiten.CursorPosition())
		if pt.In(context.VisibleBounds(&p.contextMenuPopupClickHereText)) {
			context.SetPosition(&p.contextMenuPopup, pt)
			p.contextMenuPopup.Open(context)
		}
	}
	return guigui.HandleInputResult{}
}

type simplePopupContent struct {
	guigui.DefaultWidget

	popup *basicwidget.Popup

	titleText   basicwidget.Text
	closeButton basicwidget.TextButton
}

func (s *simplePopupContent) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	u := float64(basicwidget.UnitSize(context))

	s.titleText.SetText("Hello!")
	s.titleText.SetBold(true)
	pt := s.popup.ContentBounds(context).Min.Add(image.Pt(int(0.5*u), int(0.5*u)))
	appender.AppendChildWidgetWithPosition(&s.titleText, pt)

	s.closeButton.SetText("Close")
	s.closeButton.SetOnUp(func() {
		s.popup.Close()
	})
	w, h := context.Size(&s.closeButton)
	pt = s.popup.ContentBounds(context).Max.Add(image.Pt(-int(0.5*u)-w, -int(0.5*u)-h))
	appender.AppendChildWidgetWithPosition(&s.closeButton, pt)

	return nil
}
