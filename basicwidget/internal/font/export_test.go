// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package font

import (
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/text/language"
)

func (f *FaceChooser) FaceSources(size float64, weight text.Weight, locales []language.Tag) []*text.GoTextFaceSource {
	return f.faceSources(size, weight, locales)
}
