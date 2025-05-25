// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Guigui Authors

package basicwidget

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"slices"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/text/language"
)

//go:generate go run gen.go

//go:embed InterVariable.ttf.gz
var interVariableTTFGz []byte

var theDefaultFaceSource FaceSourceEntry

type UnicodeRange struct {
	Min rune
	Max rune
}

type FaceSourceEntry struct {
	FaceSource    *text.GoTextFaceSource
	UnicodeRanges []UnicodeRange
}

var (
	theFaceCache         map[faceCacheKey]text.Face
	theFaceSourceEntries []FaceSourceEntry
)

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
	e := FaceSourceEntry{
		FaceSource: f,
	}
	theDefaultFaceSource = e
	theFaceSourceEntries = []FaceSourceEntry{e}
}

var (
	tagLiga = text.MustParseTag("liga")
	tagTnum = text.MustParseTag("tnum")
)

func fontFace(size float64, weight text.Weight, liga bool, tnum bool, lang language.Tag) text.Face {
	key := faceCacheKey{
		size:   size,
		weight: weight,
		liga:   liga,
		tnum:   tnum,
		lang:   lang,
	}
	if f, ok := theFaceCache[key]; ok {
		return f
	}

	var fs []text.Face
	for _, entry := range theFaceSourceEntries {
		gtf := &text.GoTextFace{
			Source:   entry.FaceSource,
			Size:     size,
			Language: lang,
		}
		gtf.SetVariation(text.MustParseTag("wght"), float32(weight))
		if liga {
			gtf.SetFeature(tagLiga, 1)
		} else {
			gtf.SetFeature(tagLiga, 0)
		}
		if tnum {
			gtf.SetFeature(tagTnum, 1)
		} else {
			gtf.SetFeature(tagTnum, 0)
		}

		var f text.Face
		if len(entry.UnicodeRanges) > 0 {
			lf := text.NewLimitedFace(gtf)
			for _, r := range entry.UnicodeRanges {
				lf.AddUnicodeRange(r.Min, r.Max)
			}
			f = lf
		} else {
			f = gtf
		}
		fs = append(fs, f)
	}
	mf, err := text.NewMultiFace(fs...)
	if err != nil {
		panic(err)
	}

	if theFaceCache == nil {
		theFaceCache = map[faceCacheKey]text.Face{}
	}
	theFaceCache[key] = mf

	return mf
}

func DefaultFaceSourceEntry() FaceSourceEntry {
	return theDefaultFaceSource
}

func areFaceSourceEntriesEqual(a, b []FaceSourceEntry) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].FaceSource != b[i].FaceSource {
			return false
		}
		if !slices.Equal(a[i].UnicodeRanges, b[i].UnicodeRanges) {
			return false
		}
	}
	return true
}

func SetFaceSources(entries []FaceSourceEntry) {
	if len(entries) == 0 {
		entries = []FaceSourceEntry{theDefaultFaceSource}
	}
	if areFaceSourceEntriesEqual(theFaceSourceEntries, entries) {
		return
	}
	theFaceSourceEntries = slices.Clone(entries)
	clear(theFaceCache)
}

type faceCacheKey struct {
	size   float64
	weight text.Weight
	liga   bool
	tnum   bool
	lang   language.Tag
}
