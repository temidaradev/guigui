// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 Hajime Hoshi

package basicwidget

import (
	"bytes"
	"compress/gzip"
	_ "embed"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/text/language"

	"github.com/hajimehoshi/guigui/basicwidget/internal/font"
)

//go:generate go run gen.go

//go:embed InterVariable.ttf.gz
var interVariableTTFGz []byte

type FaceSourceQueryResult struct {
	FaceSource *text.GoTextFaceSource
	Priority   float64
}

type FaceSourceHint struct {
	Size   float64
	Weight text.Weight
	Locale language.Tag
}

var theFaceChoooser font.FaceChooser

func RegisterFaceSource(faceSource *text.GoTextFaceSource, priority func(hint FaceSourceHint) float64) {
	theFaceChoooser.Register(faceSource, func(hint font.FaceSourceHint) float64 {
		return priority(FaceSourceHint(hint))
	})
}

func init() {
	r, err := gzip.NewReader(bytes.NewReader(interVariableTTFGz))
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = r.Close()
	}()
	f, err := text.NewGoTextFaceSource(r)
	if err != nil {
		panic(err)
	}
	RegisterFaceSource(f, func(hint FaceSourceHint) float64 {
		script, conf := hint.Locale.Script()
		if script == language.MustParseScript("Latn") || script == language.MustParseScript("Grek") || script == language.MustParseScript("Cyrl") {
			switch conf {
			case language.Exact, language.High:
				return 1
			case language.Low:
				return 0.5
			}
		}
		return 0
	})
}

func fontFace(size float64, weight text.Weight, features []font.FontFeature, locales []language.Tag) text.Face {
	return theFaceChoooser.Face(size, weight, features, locales)
}
