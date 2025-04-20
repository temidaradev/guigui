// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Hajime Hoshi

package font_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/guigui/basicwidget/internal/font"
	"golang.org/x/text/language"
)

func TestFaceSources(t *testing.T) {
	faceSources := map[string]*text.GoTextFaceSource{
		"en":        {},
		"fr":        {},
		"ja":        {},
		"zh-Hans":   {},
		"zh-Hant":   {},
		"fallback1": {},
		"fallback2": {},
	}

	var fc font.FaceChooser
	fc.Register(faceSources["en"], func(hint font.FaceSourceHint) float64 {
		if hint.Locale == language.English {
			return 1
		}
		return 0
	})
	fc.Register(faceSources["fr"], func(hint font.FaceSourceHint) float64 {
		if hint.Locale == language.French {
			return 1
		}
		return 0
	})
	fc.Register(faceSources["ja"], func(hint font.FaceSourceHint) float64 {
		if hint.Locale == language.Japanese {
			return 1
		}
		return 0
	})
	fc.Register(faceSources["zh-Hans"], func(hint font.FaceSourceHint) float64 {
		if hint.Locale == language.SimplifiedChinese {
			return 1
		}
		if hint.Locale == language.Chinese {
			return 2.0 / 3.0
		}
		if hint.Locale == language.TraditionalChinese {
			return 1.0 / 3.0
		}
		return 0
	})
	fc.Register(faceSources["zh-Hant"], func(hint font.FaceSourceHint) float64 {
		if hint.Locale == language.TraditionalChinese {
			return 1
		}
		if hint.Locale == language.Chinese {
			return 2.0 / 3.0
		}
		if hint.Locale == language.SimplifiedChinese {
			return 1.0 / 3.0
		}
		return 0
	})
	fc.Register(faceSources["fallback1"], func(hint font.FaceSourceHint) float64 {
		return 0.1
	})
	// If the priority is the same, the later registered face source is preferred.
	fc.Register(faceSources["fallback2"], func(hint font.FaceSourceHint) float64 {
		return 0.1
	})

	testCases := []struct {
		Locales []language.Tag
		Out     []string
	}{
		{
			Locales: nil,
			Out:     []string{"fallback2", "fallback1", "zh-Hant", "zh-Hans", "ja", "fr", "en"},
		},
		{
			Locales: []language.Tag{
				language.English,
			},
			Out: []string{"en", "fallback2", "fallback1", "zh-Hant", "zh-Hans", "ja", "fr"},
		},
		{
			Locales: []language.Tag{
				language.SimplifiedChinese,
			},
			Out: []string{"zh-Hans", "zh-Hant", "fallback2", "fallback1", "ja", "fr", "en"},
		},
		{
			Locales: []language.Tag{
				language.Chinese,
			},
			Out: []string{"zh-Hant", "zh-Hans", "fallback2", "fallback1", "ja", "fr", "en"},
		},
		{
			Locales: []language.Tag{
				language.SimplifiedChinese,
				language.Japanese,
			},
			Out: []string{"zh-Hans", "ja", "zh-Hant", "fallback2", "fallback1", "fr", "en"},
		},
		{
			Locales: []language.Tag{
				language.English,
				language.French,
				language.Japanese,
				language.SimplifiedChinese,
				language.TraditionalChinese},
			Out: []string{"en", "fr", "ja", "zh-Hans", "zh-Hant", "fallback2", "fallback1"},
		},
		{
			Locales: []language.Tag{
				language.TraditionalChinese,
				language.SimplifiedChinese,
				language.English,
				language.Japanese,
				language.French},
			Out: []string{"zh-Hant", "zh-Hans", "en", "ja", "fr", "fallback2", "fallback1"},
		},
	}
	for _, tc := range testCases {
		var localeStrs []string
		for _, l := range tc.Locales {
			localeStrs = append(localeStrs, l.String())
		}
		t.Run(strings.Join(localeStrs, ","), func(t *testing.T) {
			var got []string
			for _, fs := range fc.FaceSources(1, text.WeightNormal, tc.Locales) {
				for name, fs2 := range faceSources {
					if fs == fs2 {
						got = append(got, name)
						break
					}
				}
			}
			want := tc.Out
			if !slices.Equal(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		})
	}

}
