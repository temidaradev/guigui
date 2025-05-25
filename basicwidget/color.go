// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Guigui Authors

package basicwidget

import (
	"image/color"

	"github.com/hajimehoshi/guigui/basicwidget/internal/draw"
)

func ScaleAlpha(clr color.Color, alpha float64) color.Color {
	return draw.ScaleAlpha(clr, alpha)
}
