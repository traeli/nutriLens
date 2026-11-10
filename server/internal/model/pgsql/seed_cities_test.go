package pgsql

import "testing"

func TestSeedCitiesIncludesPopularDestinations(t *testing.T) {
	cities := seedCities()
	if len(cities) != 16 {
		t.Fatalf("seedCities count = %d, want 16", len(cities))
	}
	seen := make(map[string]bool, len(cities))
	for _, city := range cities {
		if seen[city.Code] {
			t.Fatalf("duplicate city code %s", city.Code)
		}
		seen[city.Code] = true
	}
	for _, code := range []string{"110000", "440100", "510100", "500000", "610100", "460200", "230100", "530100", "350200", "420100"} {
		if !seen[code] {
			t.Fatalf("popular destination %s missing", code)
		}
	}
}
