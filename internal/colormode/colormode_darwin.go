// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Guigui Authors

package colormode

import (
	"strings"

	"github.com/ebitengine/purego/objc"
)

var (
	idNSApplication = objc.ID(objc.GetClass("NSApplication"))

	selEffectiveAppearance = objc.RegisterName("effectiveAppearance")
	selName                = objc.RegisterName("name")
	selSharedApplication   = objc.RegisterName("sharedApplication")
	selUTF8String          = objc.RegisterName("UTF8String")
)

func systemColorMode() ColorMode {
	// "effectiveAppearance" works from macOS 10.14. As Go 1.23 supports macOS 11, it's OK to use it.
	//
	// See also:
	// * https://developer.apple.com/documentation/appkit/nsapplication/effectiveappearance?language=objc
	// * https://go.dev/wiki/MinimumRequirements
	objcName := idNSApplication.Send(selSharedApplication).Send(selEffectiveAppearance).Send(selName)
	name := objc.Send[string](objcName, selUTF8String)
	// https://developer.apple.com/documentation/appkit/nsappearance/name-swift.struct?language=objc
	if strings.Contains(name, "Dark") {
		return Dark
	}
	return Light
}
