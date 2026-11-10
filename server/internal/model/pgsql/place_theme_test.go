package pgsql

import (
	"strings"
	"testing"
)

func TestSeedPlaceThemesHaveUniquePublicCOSIcons(t *testing.T) {
	themes := seedPlaceThemes()
	if len(themes) != 15 {
		t.Fatalf("theme count = %d, want 15", len(themes))
	}
	seen := make(map[string]bool, len(themes))
	for _, theme := range themes {
		if seen[theme.Code] {
			t.Fatalf("duplicate theme code %q", theme.Code)
		}
		seen[theme.Code] = true
		if !strings.HasPrefix(theme.IconURL, routeThemeIconBaseURL+"/") || !strings.HasSuffix(theme.IconURL, ".png") {
			t.Fatalf("theme %s icon url = %q", theme.Code, theme.IconURL)
		}
	}
}
