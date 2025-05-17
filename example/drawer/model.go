// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Guigui Authors

package main

type Model struct {
	leftPanelClosed  bool
	rightPanelClosed bool
}

func (m *Model) IsLeftPanelOpen() bool {
	return !m.leftPanelClosed
}

func (m *Model) SetLeftPanelOpen(open bool) {
	m.leftPanelClosed = !open
}

func (m *Model) IsRightPanelOpen() bool {
	return !m.rightPanelClosed
}

func (m *Model) SetRightPanelOpen(open bool) {
	m.rightPanelClosed = !open
}
