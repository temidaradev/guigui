// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package font

import (
	"slices"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/text/language"
)

type FaceSourceHint struct {
	Size   float64
	Weight text.Weight
	Locale language.Tag
}

type faceSourceWithPriorityFunc struct {
	faceSource *text.GoTextFaceSource
	priority   func(hint FaceSourceHint) float64
}

type FaceChooser struct {
	priorityFuncs []faceSourceWithPriorityFunc
	cache         map[faceCacheKey]text.Face
}

type faceCacheKey struct {
	size     float64
	weight   text.Weight
	ligature bool
	locales  string
}

func (f *FaceChooser) Register(faceSource *text.GoTextFaceSource, priority func(hint FaceSourceHint) float64) {
	f.priorityFuncs = append(f.priorityFuncs, faceSourceWithPriorityFunc{
		faceSource: faceSource,
		priority:   priority,
	})
}

func (f *FaceChooser) Face(size float64, weight text.Weight, ligature bool, locales []language.Tag) text.Face {
	var localeStrs []string
	for _, l := range locales {
		localeStrs = append(localeStrs, l.String())
	}

	key := faceCacheKey{
		size:     size,
		weight:   weight,
		ligature: ligature,
		locales:  strings.Join(localeStrs, ","),
	}
	if f, ok := f.cache[key]; ok {
		return f
	}

	fps := append([]faceSourceWithPriorityFunc{}, f.priorityFuncs...)

	var faceSources []*text.GoTextFaceSource
	for _, l := range locales {
		var highestPriority float64
		var index int
		for i, fp := range fps {
			p := min(max(fp.priority(FaceSourceHint{
				Size:   size,
				Weight: weight,
				Locale: l,
			}), 0), 1)
			// If the priority is the same, the later one is used.
			if highestPriority <= p {
				highestPriority = p
				index = i
			}
		}
		// TODO: Now only one face is added for each locale. Add more faces (#68).
		faceSources = append(faceSources, fps[index].faceSource)
		fps = slices.Delete(fps, index, index+1)
		if len(fps) == 0 {
			break
		}
	}

	var fs []text.Face
	var lang language.Tag
	if len(locales) > 0 {
		lang = locales[0]
	}
	for _, faceSource := range faceSources {
		f := &text.GoTextFace{
			Source:   faceSource,
			Size:     size,
			Language: lang,
		}
		f.SetVariation(text.MustParseTag("wght"), float32(weight))
		if !ligature {
			f.SetFeature(text.MustParseTag("liga"), 0)
		}
		fs = append(fs, f)
	}
	mf, err := text.NewMultiFace(fs...)
	if err != nil {
		panic(err)
	}

	if f.cache == nil {
		f.cache = map[faceCacheKey]text.Face{}
	}
	f.cache[key] = mf

	return mf
}
