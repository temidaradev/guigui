// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Guigui Authors

package colormode

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func systemColorMode() ColorMode {
	mode := checkGSettings()
	if mode != Unknown {
		return mode
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Unknown
	}
	mode = checkGTKSettingsFile(filepath.Join(homeDir+".config", "gtk-4.0", "settings.ini"))
	if mode != Unknown {
		return mode
	}

	mode = checkGTKSettingsFile(filepath.Join(homeDir+".config", "gtk-3.0", "settings.ini"))
	if mode != Unknown {
		return mode
	}

	return Unknown
}

func checkGSettings() ColorMode {
	out, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme").Output()
	if err != nil {
		return Unknown
	}

	value := strings.TrimSpace(string(out))
	value = strings.Trim(value, "'")

	switch value {
	case "prefer-dark":
		return Dark
	case "default", "prefer-light":
		return Light
	default:
		return Unknown
	}
}

func checkGTKSettingsFile(path string) ColorMode {
	data, err := os.ReadFile(path)
	if err != nil {
		return Unknown
	}

	content := strings.ToLower(string(data))
	if strings.Contains(content, "gtk-application-prefer-dark-theme=true") {
		return Dark
	}
	return Light
}
