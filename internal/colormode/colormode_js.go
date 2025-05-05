// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Guigui Authors

package colormode

import "syscall/js"

var (
	matchMedia = js.Global().Get("window").Get("matchMedia")
)

func systemColorMode() ColorMode {
	if !matchMedia.Truthy() {
		return Unknown
	}
	media := matchMedia.Invoke("(prefers-color-scheme: dark)")
	if media.Get("matches").Bool() {
		return Dark
	}
	return Light
}
