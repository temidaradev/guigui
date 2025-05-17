// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Guigui Authors

package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

type Model struct {
	leftPanelClosed  bool
	rightPanelClosed bool

	leftOpening  bool
	leftClosing  bool
	leftCount    int
	rightOpening bool
	rightClosing bool
	rightCount   int
}

func panelInitCount() int {
	return ebiten.TPS() / 10
}

func (m *Model) Tick() {
	if m.leftCount > 0 {
		m.leftCount--
		if m.leftCount == 0 {
			m.leftOpening = false
			m.leftClosing = false
		}
	}
	if m.rightCount > 0 {
		m.rightCount--
		if m.rightCount == 0 {
			m.rightOpening = false
			m.rightClosing = false
		}
	}
}

func (m *Model) DefaultPanelWidth(context *guigui.Context) int {
	u := basicwidget.UnitSize(context)
	return 8 * u
}

func (m *Model) IsLeftPanelOpen() bool {
	return !m.leftPanelClosed && !m.leftOpening && !m.leftClosing
}

func (m *Model) SetLeftPanelOpen(open bool) {
	if m.leftPanelClosed == !open {
		return
	}
	m.leftPanelClosed = !open
	if open {
		m.leftOpening = true
	} else {
		m.leftClosing = true
	}
	m.leftCount = panelInitCount()
}

func (m *Model) LeftPanelWidth(context *guigui.Context) int {
	fullWidth := m.DefaultPanelWidth(context)
	if m.leftOpening {
		rate := float64(m.leftCount) / float64(panelInitCount())
		return int(float64(fullWidth) * (1 - rate))
	}
	if m.leftClosing {
		rate := float64(m.leftCount) / float64(panelInitCount())
		return int(float64(fullWidth) * rate)
	}
	if m.leftPanelClosed {
		return 0
	}
	return fullWidth
}

func (m *Model) IsRightPanelOpen() bool {
	return !m.rightPanelClosed && !m.rightOpening && !m.rightClosing
}

func (m *Model) SetRightPanelOpen(open bool) {
	if m.rightPanelClosed == !open {
		return
	}
	m.rightPanelClosed = !open
	if open {
		m.rightOpening = true
	} else {
		m.rightClosing = true
	}
	m.rightCount = panelInitCount()
}

func (m *Model) RightPanelWidth(context *guigui.Context) int {
	fullWidth := m.DefaultPanelWidth(context)
	if m.rightOpening {
		rate := float64(m.rightCount) / float64(panelInitCount())
		return int(float64(fullWidth) * (1 - rate))
	}
	if m.rightClosing {
		rate := float64(m.rightCount) / float64(panelInitCount())
		return int(float64(fullWidth) * rate)
	}
	if m.rightPanelClosed {
		return 0
	}
	return fullWidth
}
