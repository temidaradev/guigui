// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package main

type Model struct {
	mode string
}

func (m *Model) Mode() string {
	if m.mode == "" {
		return "settings"
	}
	return m.mode
}

func (m *Model) SetMode(mode string) {
	m.mode = mode
}
