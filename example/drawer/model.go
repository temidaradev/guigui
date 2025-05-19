// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Guigui Authors

package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Model struct {
	leftOpening       bool
	leftClosing       bool
	leftClosingCount  int
	rightOpening      bool
	rightClosing      bool
	rightClosingCount int
}

func panelMaxClosingCount() int {
	return ebiten.TPS() / 10
}

func (m *Model) Tick() {
	if m.leftOpening {
		m.leftClosingCount--
		if m.leftClosingCount == 0 {
			m.leftOpening = false
		}
	}
	if m.leftClosing {
		m.leftClosingCount++
		if m.leftClosingCount == panelMaxClosingCount() {
			m.leftClosing = false
		}
	}
	if m.rightOpening {
		m.rightClosingCount--
		if m.rightClosingCount == 0 {
			m.rightOpening = false
		}
	}
	if m.rightClosing {
		m.rightClosingCount++
		if m.rightClosingCount == panelMaxClosingCount() {
			m.rightClosing = false
		}
	}
}

func (m *Model) DefaultPanelWidth(context *guigui.Context) int {
	u := basicwidget.UnitSize(context)
	return 8 * u
}

func (m *Model) IsLeftPanelOpen() bool {
	return m.leftClosingCount == 0 && !m.leftOpening && !m.leftClosing
}

func (m *Model) SetLeftPanelOpen(open bool) {
	if open {
		if m.leftOpening {
			return
		}
		if m.leftClosingCount == 0 {
			return
		}
		m.leftOpening = true
		m.leftClosing = false
		return
	}
	if m.leftClosing {
		return
	}
	if m.leftClosingCount == panelMaxClosingCount() {
		return
	}
	m.leftClosing = true
	m.leftOpening = false
}

func (m *Model) LeftPanelWidth(context *guigui.Context) int {
	fullWidth := m.DefaultPanelWidth(context)
	rate := float64(m.leftClosingCount) / float64(panelMaxClosingCount())
	return int(float64(fullWidth) * (1 - rate))
}

func (m *Model) IsRightPanelOpen() bool {
	return m.rightClosingCount == 0 && !m.rightOpening && !m.rightClosing
}

func (m *Model) SetRightPanelOpen(open bool) {
	if open {
		if m.rightOpening {
			return
		}
		if m.rightClosingCount == 0 {
			return
		}
		m.rightOpening = true
		m.rightClosing = false
		return
	}
	if m.rightClosing {
		return
	}
	if m.rightClosingCount == panelMaxClosingCount() {
		return
	}
	m.rightClosing = true
	m.rightOpening = false
}

func (m *Model) RightPanelWidth(context *guigui.Context) int {
	fullWidth := m.DefaultPanelWidth(context)
	rate := float64(m.rightClosingCount) / float64(panelMaxClosingCount())
	return int(float64(fullWidth) * (1 - rate))
}
