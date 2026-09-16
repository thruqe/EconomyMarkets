package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const DefaultSettingsPath = "data/settings.json"

type Settings struct {
	Theme ThemeMode `json:"theme"`
}

func LoadSettings(path string) Settings {
	if path == "" {
		path = DefaultSettingsPath
	}
	s := Settings{Theme: ThemeDark}
	data, err := os.ReadFile(path)
	if err != nil {
		return s
	}
	_ = json.Unmarshal(data, &s)
	if s.Theme != ThemeLight && s.Theme != ThemeDark {
		s.Theme = ThemeDark
	}
	return s
}

func SaveSettings(path string, s Settings) error {
	if path == "" {
		path = DefaultSettingsPath
	}
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
