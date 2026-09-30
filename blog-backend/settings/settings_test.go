package settings

import (
	"path/filepath"
	"strings"
	"testing"
)

func palette(color string) map[string]string {
	colors := map[string]string{}
	for _, k := range baseKeys {
		colors[k] = color
	}
	return colors
}

func TestLoadMissingIsEmpty(t *testing.T) {
	s, err := NewStore(filepath.Join(t.TempDir(), "settings.json")).Load()
	if err != nil || s.Highlight != nil {
		t.Fatalf("Load = %+v, %v", s, err)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "settings.json"))
	want := Settings{Highlight: &Highlight{Preset: "seoulism", Colors: palette("#AbCdEf"),
		Tokens: map[string]TokenStyle{"type": {Color: "#efeeea", Italic: true}, "keyword": {Bold: true}}}}
	if err := store.Save(want); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil || got.Highlight.Preset != "seoulism" || got.Highlight.Colors["base0F"] != "#AbCdEf" ||
		got.Highlight.Tokens["type"] != (TokenStyle{Color: "#efeeea", Italic: true}) || !got.Highlight.Tokens["keyword"].Bold {
		t.Fatalf("Load = %+v, %v", got.Highlight, err)
	}
}

func TestValidate(t *testing.T) {
	bad := map[string]Settings{
		"preset":        {Highlight: &Highlight{Preset: "Bad Preset!", Colors: palette("#000000")}},
		"color":         {Highlight: &Highlight{Preset: "x", Colors: palette("red")}},
		"css injection": {Highlight: &Highlight{Preset: "x", Colors: palette("#000000; background:url(x)")}},
		"missing key":   {Highlight: &Highlight{Preset: "x", Colors: map[string]string{"base00": "#000000"}}},
	}
	for name, s := range bad {
		if err := s.Validate(); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}

	extra := palette("#000000")
	delete(extra, "base0F")
	extra["base10"] = "#000000"
	err := Settings{Highlight: &Highlight{Preset: "x", Colors: extra}}.Validate()
	if err == nil || !strings.Contains(err.Error(), "base0F") {
		t.Errorf("unknown key: got %v", err)
	}
}
