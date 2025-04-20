// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package font

import (
	"cmp"
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

func (f *FaceChooser) faceSources(size float64, weight text.Weight, locales []language.Tag) []*text.GoTextFaceSource {
	type priority struct {
		priority          float64
		localeIndex       int
		registrationIndex int
	}

	priorities := map[*text.GoTextFaceSource]priority{}

	for i, l := range locales {
		for j, fp := range f.priorityFuncs {
			p := min(max(fp.priority(FaceSourceHint{
				Size:   size,
				Weight: weight,
				Locale: l,
			}), 0), 1)
			if priorities[fp.faceSource].priority >= p {
				continue
			}
			priorities[fp.faceSource] = priority{
				priority:          p,
				localeIndex:       i,
				registrationIndex: j,
			}
		}
	}

	var faceSources []*text.GoTextFaceSource
	for _, fp := range f.priorityFuncs {
		faceSources = append(faceSources, fp.faceSource)
	}

	slices.SortStableFunc(faceSources, func(a, b *text.GoTextFaceSource) int {
		ap, aOk := priorities[a]
		bp, bOk := priorities[b]
		if !aOk && !bOk {
			return 0
		}
		if !aOk {
			return 1
		}
		if !bOk {
			return -1
		}
		if ap.priority != bp.priority {
			return -cmp.Compare(ap.priority, bp.priority)
		}
		if ap.localeIndex != bp.localeIndex {
			return cmp.Compare(ap.localeIndex, bp.localeIndex)
		}
		// Prefer later registrations.
		return -cmp.Compare(ap.registrationIndex, bp.registrationIndex)
	})

	return faceSources
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

	faceSources := f.faceSources(size, weight, locales)

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
