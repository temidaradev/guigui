// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"fmt"
	"image"
	"os"
	"slices"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	_ "github.com/hajimehoshi/guigui/basicwidget/cjkfont"
	"github.com/hajimehoshi/guigui/layout"
)

type Root struct {
	guigui.RootWidget

	background        basicwidget.Background
	createButton      basicwidget.TextButton
	textField         basicwidget.TextField
	tasksPanel        basicwidget.ScrollablePanel
	tasksPanelContent tasksPanelContent

	model Model
}

func (r *Root) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	appender.AppendChildWidgetWithBounds(&r.background, context.Bounds(r))

	r.textField.SetOnEnterPressed(func(text string) {
		r.tryCreateTask(text)
	})

	r.createButton.SetText("Create")
	r.createButton.SetOnUp(func() {
		r.tryCreateTask(r.textField.Text())
	})
	if r.model.CanAddTask(r.textField.Text()) {
		context.Enable(&r.createButton)
	} else {
		context.Disable(&r.createButton)
	}

	r.tasksPanelContent.SetModel(&r.model)
	r.tasksPanelContent.SetOnDeleted(func(id int) {
		r.model.DeleteTaskByID(id)
	})
	r.tasksPanel.SetContent(&r.tasksPanelContent)

	u := basicwidget.UnitSize(context)
	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(r).Inset(u / 2),
		Heights: []layout.Size{
			layout.FixedSize(u),
			layout.FlexibleSize(1),
		},
		RowGap: u / 2,
	}).CellBounds() {
		switch i {
		case 0:
			for i, bounds := range (layout.GridLayout{
				Bounds: bounds,
				Widths: []layout.Size{
					layout.FlexibleSize(1),
					layout.FixedSize(5 * u),
				},
				ColumnGap: u / 2,
			}).CellBounds() {
				switch i {
				case 0:
					appender.AppendChildWidgetWithBounds(&r.textField, bounds)
				case 1:
					appender.AppendChildWidgetWithBounds(&r.createButton, bounds)
				}
			}
		case 1:
			context.SetSize(&r.tasksPanelContent, image.Pt(bounds.Dx(), guigui.DefaultSize))
			appender.AppendChildWidgetWithBounds(&r.tasksPanel, bounds)
		}
	}

	return nil
}

func (r *Root) tryCreateTask(text string) {
	if r.model.TryAddTask(text) {
		r.textField.SetText("")
	}
}

type taskWidget struct {
	guigui.DefaultWidget

	doneButton basicwidget.TextButton
	text       basicwidget.Text

	onDoneButtonPressed func()
}

func (t *taskWidget) SetOnDoneButtonPressed(f func()) {
	t.onDoneButtonPressed = f
}

func (t *taskWidget) SetText(text string) {
	t.text.SetText(text)
}

func (t *taskWidget) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	t.doneButton.SetText("Done")
	t.doneButton.SetOnUp(func() {
		if t.onDoneButtonPressed != nil {
			t.onDoneButtonPressed()
		}
	})

	t.text.SetVerticalAlign(basicwidget.VerticalAlignMiddle)

	u := basicwidget.UnitSize(context)
	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(t),
		Widths: []layout.Size{
			layout.FixedSize(3 * u),
			layout.FlexibleSize(1),
		},
		ColumnGap: u / 2,
	}).CellBounds() {
		switch i {
		case 0:
			appender.AppendChildWidgetWithBounds(&t.doneButton, bounds)
		case 1:
			appender.AppendChildWidgetWithBounds(&t.text, bounds)
		}
	}

	return nil
}

func (t *taskWidget) DefaultSize(context *guigui.Context) image.Point {
	return image.Pt(6*int(basicwidget.UnitSize(context)), context.Size(&t.doneButton).Y)
}

type tasksPanelContent struct {
	guigui.DefaultWidget

	taskWidgets []taskWidget

	onDeleted func(id int)

	model *Model
}

func (t *tasksPanelContent) SetOnDeleted(f func(id int)) {
	t.onDeleted = f
}

func (t *tasksPanelContent) SetModel(model *Model) {
	t.model = model
}

func (t *tasksPanelContent) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	if t.model.TaskCount() > len(t.taskWidgets) {
		t.taskWidgets = slices.Grow(t.taskWidgets, t.model.TaskCount()-len(t.taskWidgets))[:t.model.TaskCount()]
	} else {
		t.taskWidgets = slices.Delete(t.taskWidgets, t.model.TaskCount(), len(t.taskWidgets))
	}
	for i := range t.model.TaskCount() {
		task := t.model.TaskByIndex(i)
		t.taskWidgets[i].SetOnDoneButtonPressed(func() {
			if t.onDeleted != nil {
				t.onDeleted(task.ID)
			}
		})
		t.taskWidgets[i].SetText(task.Text)
	}

	u := basicwidget.UnitSize(context)

	for i, bounds := range (layout.GridLayout{
		Bounds: context.Bounds(t),
		Heights: []layout.Size{
			layout.MaxContentSize(func(index int) int {
				if index >= len(t.taskWidgets) {
					return 0
				}
				return context.Size(&t.taskWidgets[index]).Y
			}),
		},
		RowGap: u / 4,
	}).RepeatingCellBounds() {
		if i >= len(t.taskWidgets) {
			break
		}
		appender.AppendChildWidgetWithBounds(&t.taskWidgets[i], bounds)
	}

	return nil
}

func (t *tasksPanelContent) DefaultSize(context *guigui.Context) image.Point {
	u := basicwidget.UnitSize(context)
	var h int
	for i := range t.taskWidgets {
		h += context.Size(&t.taskWidgets[i]).Y
		h += int(u / 4)
	}
	return image.Pt(6*int(u), h)
}

func main() {
	op := &guigui.RunOptions{
		Title:           "TODO",
		WindowMinWidth:  320,
		WindowMinHeight: 240,
	}
	if err := guigui.Run(&Root{}, op); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
