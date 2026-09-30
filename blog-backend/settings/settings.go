// Package settings stores site-wide settings that the admin can change from
// the dashboard, as a JSON file next to the content.
package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

// baseKeys are the 16 color slots of a base16 palette.
var baseKeys = []string{
	"base00", "base01", "base02", "base03", "base04", "base05", "base06", "base07",
	"base08", "base09", "base0A", "base0B", "base0C", "base0D", "base0E", "base0F",
}

// tokenNames are the tokens a palette may style individually; they match
// TOKENS in blog-frontend/src/highlight/theme.js.
var tokenNames = map[string]bool{
	"comment": true, "keyword": true, "type": true, "built_in": true,
	"function": true, "title": true, "string": true, "number": true,
	"literal": true, "variable": true, "attr": true, "meta": true,
	"section": true, "operator": true, "punctuation": true,
}

var (
	hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	presetID = regexp.MustCompile(`^[a-z0-9-]{1,64}$`)
)

// Highlight is the code highlighting palette.
type Highlight struct {
	// Preset is the id of the palette the colors started from, or "custom".
	Preset string `json:"preset"`
	// Colors maps every base16 slot (base00..base0F) to a #rrggbb color.
	Colors map[string]string `json:"colors"`
	// Tokens optionally styles single tokens beyond the base16 slots.
	Tokens map[string]TokenStyle `json:"tokens,omitempty"`
}

// TokenStyle overrides how one token is drawn. An empty Color keeps the
// token's base16 slot color.
type TokenStyle struct {
	Color  string `json:"color,omitempty"`
	Bold   bool   `json:"bold,omitempty"`
	Italic bool   `json:"italic,omitempty"`
}

// Settings is the whole settings document. A nil field means "use the
// frontend default".
type Settings struct {
	Highlight *Highlight `json:"highlight"`
}

// Validate checks that the settings are well-formed.
func (s Settings) Validate() error {
	h := s.Highlight
	if h == nil {
		return nil
	}
	if !presetID.MatchString(h.Preset) {
		return fmt.Errorf("invalid highlight preset %q", h.Preset)
	}
	if len(h.Colors) != len(baseKeys) {
		return fmt.Errorf("highlight colors must have exactly the keys %v", baseKeys)
	}
	for _, key := range baseKeys {
		color, ok := h.Colors[key]
		if !ok {
			return fmt.Errorf("highlight color %s is missing", key)
		}
		if !hexColor.MatchString(color) {
			return fmt.Errorf("highlight color %s must look like #rrggbb, got %q", key, color)
		}
	}
	for name, style := range h.Tokens {
		if !tokenNames[name] {
			return fmt.Errorf("unknown highlight token %q", name)
		}
		if style.Color != "" && !hexColor.MatchString(style.Color) {
			return fmt.Errorf("highlight token %s color must look like #rrggbb, got %q", name, style.Color)
		}
	}
	return nil
}

// Store reads and writes the settings file.
type Store struct {
	path string
	mu   sync.RWMutex
}

// NewStore returns a Store backed by the JSON file at path.
func NewStore(path string) *Store {
	return &Store{path: path}
}

// Load returns the saved settings, or empty settings if none were saved yet.
func (s *Store) Load() (Settings, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var settings Settings
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return settings, nil
	}
	if err != nil {
		return settings, err
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		return settings, fmt.Errorf("parse %s: %w", s.path, err)
	}
	return settings, nil
}

// Save validates and writes settings. The file is replaced atomically.
func (s *Store) Save(settings Settings) error {
	if err := settings.Validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".settings-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // no-op once renamed
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), s.path)
}
