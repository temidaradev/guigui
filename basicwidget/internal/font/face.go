// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package font

import (
	"cmp"
	"slices"

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
	features string
	locales  string
}

func (f *FaceChooser) Register(faceSource *text.GoTextFaceSource, priority func(hint FaceSourceHint) float64) {
	f.priorityFuncs = append(f.priorityFuncs, faceSourceWithPriorityFunc{
		faceSource: faceSource,
		priority:   priority,
	})
	clear(f.cache)
}

func (f *FaceChooser) faceSources(size float64, weight text.Weight, locales []language.Tag) []*text.GoTextFaceSource {
	type priority struct {
		priority    float64
		localeIndex int
	}

	priorities := map[*text.GoTextFaceSource]priority{}

	for i, l := range locales {
		for _, fp := range f.priorityFuncs {
			p := min(max(fp.priority(FaceSourceHint{
				Size:   size,
				Weight: weight,
				Locale: l,
			}), 0), 1)
			if _, ok := priorities[fp.faceSource]; ok {
				if priorities[fp.faceSource].priority >= p {
					continue
				}
			}
			priorities[fp.faceSource] = priority{
				priority:    p,
				localeIndex: i,
			}
		}
	}

	var faceSources []*text.GoTextFaceSource
	for i := len(f.priorityFuncs) - 1; i >= 0; i-- {
		faceSources = append(faceSources, f.priorityFuncs[i].faceSource)
	}

	slices.SortStableFunc(faceSources, func(a, b *text.GoTextFaceSource) int {
		ap := priorities[a]
		bp := priorities[b]
		if ap.priority != bp.priority {
			return -cmp.Compare(ap.priority, bp.priority)
		}
		return cmp.Compare(ap.localeIndex, bp.localeIndex)
	})

	return faceSources
}

func (f *FaceChooser) Face(size float64, weight text.Weight, features []FontFeature, locales []language.Tag) text.Face {
	// 7 is for a long locale length like 'zh-Hans'. 1 is for a comma.
	localeStr := make([]byte, 0, len(locales)*(7+1))
	for _, l := range locales {
		localeStr = append(localeStr, l.String()...)
		localeStr = append(localeStr, ',')
	}

	key := faceCacheKey{
		size:     size,
		weight:   weight,
		features: string(serializeFontFeatures(features)),
		locales:  string(localeStr),
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
		for _, ff := range features {
			f.SetFeature(ff.Tag, ff.Value)
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

type FontFeature struct {
	Tag   text.Tag
	Value uint32
}

func serializeFontFeatures(features []FontFeature) []byte {
	if len(features) == 0 {
		return nil
	}

	b := make([]byte, 0, len(features)*8)
	for _, ff := range features {
		b = append(b, byte(ff.Tag), byte(ff.Tag>>8), byte(ff.Tag>>16), byte(ff.Tag>>24))
		b = append(b, byte(ff.Value), byte(ff.Value>>8), byte(ff.Value>>16), byte(ff.Value>>24))
	}
	return b
}
