// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

//go:build !darwin && !js

package basicwidget

func useEmacsKeybind() bool {
	return false
}
