// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package basicwidget

import (
	"image"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget/internal/draw"
	"github.com/hajimehoshi/guigui/layout"
)

type SegmentedControlItem[T comparable] struct {
	Text     string
	Disabled bool
	Tag      T
}

func (s SegmentedControlItem[T]) tag() T {
	return s.Tag
}

type SegmentedControl[T comparable] struct {
	guigui.DefaultWidget

	abstractList abstractList[T, SegmentedControlItem[T]]
	textButtons  []TextButton
}

func (s *SegmentedControl[T]) SetOnItemSelected(f func(index int)) {
	s.abstractList.SetOnItemSelected(f)
}

func (s *SegmentedControl[T]) SetItems(items []SegmentedControlItem[T]) {
	s.abstractList.SetItems(items)
}

func (s *SegmentedControl[T]) SelectedItem() (SegmentedControlItem[T], bool) {
	return s.abstractList.SelectedItem()
}

func (s *SegmentedControl[T]) SelectedItemIndex() int {
	return s.abstractList.SelectedItemIndex()
}

func (s *SegmentedControl[T]) ItemByIndex(index int) (SegmentedControlItem[T], bool) {
	return s.abstractList.ItemByIndex(index)
}

func (s *SegmentedControl[T]) SelectItemByIndex(index int) {
	if s.abstractList.SelectItemByIndex(index) {
		guigui.RequestRedraw(s)
	}
}

func (s *SegmentedControl[T]) SelectItemByTag(tag T) {
	if s.abstractList.SelectItemByTag(tag) {
		guigui.RequestRedraw(s)
	}
}

func (s *SegmentedControl[T]) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	s.textButtons = adjustSliceSize(s.textButtons, s.abstractList.ItemCount())

	widths := make([]layout.Size, s.abstractList.ItemCount())
	for i := range s.abstractList.ItemCount() {
		item, _ := s.abstractList.ItemByIndex(i)
		s.textButtons[i].SetText(item.Text)
		s.textButtons[i].SetTextBold(s.abstractList.SelectedItemIndex() == i)
		s.textButtons[i].setUseAccentColor(true)
		if s.abstractList.ItemCount() > 1 {
			switch i {
			case 0:
				s.textButtons[i].setSharpenCorners(draw.SharpenCorners{
					UpperRight: true,
					LowerRight: true,
				})
			case s.abstractList.ItemCount() - 1:
				s.textButtons[i].setSharpenCorners(draw.SharpenCorners{
					UpperLeft: true,
					LowerLeft: true,
				})
			default:
				s.textButtons[i].setSharpenCorners(draw.SharpenCorners{
					UpperLeft:  true,
					LowerLeft:  true,
					UpperRight: true,
					LowerRight: true,
				})
			}
		}
		context.SetEnabled(&s.textButtons[i], !item.Disabled)
		s.textButtons[i].setKeepPressed(s.abstractList.SelectedItemIndex() == i)
		s.textButtons[i].SetOnDown(func() {
			s.SelectItemByIndex(i)
		})
		widths[i] = layout.FlexibleSize(1)
	}

	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(s),
		Widths: widths,
	}).CellBounds() {
		appender.AppendChildWidgetWithBounds(&s.textButtons[i], bounds)
	}
	return nil
}

func (s *SegmentedControl[T]) DefaultSize(context *guigui.Context) image.Point {
	var w, h int
	for i := range s.abstractList.ItemCount() {
		var t TextButton
		item, _ := s.abstractList.ItemByIndex(i)
		t.SetText(item.Text)
		t.SetTextBold(true)
		w = max(w, t.DefaultSize(context).X)
		h = max(h, t.DefaultSize(context).Y)
	}
	return image.Pt(w*len(s.textButtons), h)
}
