// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package main

import (
	"fmt"
	"image"
	"os"
	"slices"
	"strings"

	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	_ "github.com/hajimehoshi/guigui/basicwidget/cjkfont"
)

var theCurrentID int

func nextTaskID() int {
	theCurrentID++
	return theCurrentID
}

type Task struct {
	ID   int
	Text string
}

func NewTask(text string) Task {
	return Task{
		ID:   nextTaskID(),
		Text: text,
	}
}

type Root struct {
	guigui.RootWidget

	background        basicwidget.Background
	createButton      basicwidget.TextButton
	textField         basicwidget.TextField
	tasksPanel        basicwidget.ScrollablePanel
	tasksPanelContent tasksPanelContent

	tasks []Task
}

func (r *Root) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	w, h := context.Size(r)
	context.SetSize(&r.background, w, h)
	appender.AppendChildWidget(&r.background)

	u := float64(basicwidget.UnitSize(context))

	context.SetSize(&r.textField, w-int(6.5*u), int(u))
	r.textField.SetOnEnterPressed(func(text string) {
		r.tryCreateTask()
	})
	{
		context.SetPosition(&r.textField, context.Position(r).Add(image.Pt(int(0.5*u), int(0.5*u))))
		appender.AppendChildWidget(&r.textField)
	}

	r.createButton.SetText("Create")
	context.SetSize(&r.createButton, int(5*u), guigui.DefaultSize)
	r.createButton.SetOnUp(func() {
		r.tryCreateTask()
	})
	if r.canCreateTask() {
		context.Enable(&r.createButton)
	} else {
		context.Disable(&r.createButton)
	}
	{
		p := context.Position(r)
		p.X += w - int(0.5*u) - int(5*u)
		p.Y += int(0.5 * u)
		context.SetPosition(&r.createButton, p)
		appender.AppendChildWidget(&r.createButton)
	}

	context.SetSize(&r.tasksPanel, w, h-int(2*u))
	r.tasksPanelContent.SetTasks(context, r.tasks)
	r.tasksPanelContent.SetOnDeleted(func(id int) {
		r.tasks = slices.DeleteFunc(r.tasks, func(t Task) bool {
			return t.ID == id
		})
	})
	r.tasksPanel.SetContent(&r.tasksPanelContent)
	context.SetPosition(&r.tasksPanel, context.Position(r).Add(image.Pt(0, int(2*u))))
	context.SetSize(&r.tasksPanelContent, w, guigui.DefaultSize)
	appender.AppendChildWidget(&r.tasksPanel)

	return nil
}

func (r *Root) canCreateTask() bool {
	str := r.textField.Text()
	str = strings.TrimSpace(str)
	return str != ""
}

func (r *Root) tryCreateTask() {
	str := r.textField.Text()
	str = strings.TrimSpace(str)
	if str != "" {
		r.tasks = slices.Insert(r.tasks, 0, NewTask(str))
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
	u := float64(basicwidget.UnitSize(context))

	p := context.Position(t)
	t.doneButton.SetText("Done")
	context.SetSize(&t.doneButton, int(3*u), guigui.DefaultSize)
	t.doneButton.SetOnUp(func() {
		if t.onDoneButtonPressed != nil {
			t.onDoneButtonPressed()
		}
	})
	context.SetPosition(&t.doneButton, p)
	appender.AppendChildWidget(&t.doneButton)

	w, h := context.Size(t)
	context.SetSize(&t.text, w-int(4.5*u), h)
	t.text.SetVerticalAlign(basicwidget.VerticalAlignMiddle)
	context.SetPosition(&t.text, image.Pt(p.X+int(3.5*u), p.Y))
	appender.AppendChildWidget(&t.text)
	return nil
}

func (t *taskWidget) DefaultSize(context *guigui.Context) (int, int) {
	_, h := context.Size(&t.doneButton)
	return 6 * int(basicwidget.UnitSize(context)), h
}

type tasksPanelContent struct {
	guigui.DefaultWidget

	taskWidgets []taskWidget

	onDeleted func(id int)
}

func (t *tasksPanelContent) SetOnDeleted(f func(id int)) {
	t.onDeleted = f
}

func (t *tasksPanelContent) SetTasks(context *guigui.Context, tasks []Task) {
	if len(tasks) != len(t.taskWidgets) {
		if len(tasks) > len(t.taskWidgets) {
			t.taskWidgets = slices.Grow(t.taskWidgets, len(tasks)-len(t.taskWidgets))
			t.taskWidgets = t.taskWidgets[:len(tasks)]
		} else {
			t.taskWidgets = slices.Delete(t.taskWidgets, len(tasks), len(t.taskWidgets))
		}
	}
	for i, task := range tasks {
		t.taskWidgets[i].SetOnDoneButtonPressed(func() {
			if t.onDeleted != nil {
				t.onDeleted(task.ID)
			}
		})
		t.taskWidgets[i].SetText(task.Text)
	}
}

func (t *tasksPanelContent) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error {
	u := float64(basicwidget.UnitSize(context))

	p := context.Position(t)
	x := p.X + int(0.5*u)
	y := p.Y
	for i := range t.taskWidgets {
		// Do not take a variable for the task widget in the for-range loop,
		// since the pointer value of the task widget matters.
		if i > 0 {
			y += int(u / 4)
		}
		context.SetPosition(&t.taskWidgets[i], image.Pt(x, y))
		w, _ := context.Size(t)
		context.SetSize(&t.taskWidgets[i], w, int(u))
		appender.AppendChildWidget(&t.taskWidgets[i])
		y += int(u)
	}

	return nil
}

func (t *tasksPanelContent) DefaultSize(context *guigui.Context) (int, int) {
	u := basicwidget.UnitSize(context)
	var h int
	for i := range t.taskWidgets {
		_, th := context.Size(&t.taskWidgets[i])
		h += th
		h += int(u / 4)
	}
	return 6 * int(u), h
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
