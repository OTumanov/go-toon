package toon

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type fixtureBundle struct {
	Version  string `json:"version"`
	Category string `json:"category"`
	Tests    []struct {
		Name string `json:"name"`
	} `json:"tests"`
}

// TestSpecFixturesAvailability ensures CI can consume official TOON fixtures.
// This test intentionally validates fixture plumbing only; behavioral conformance
// assertions are tracked separately and can be incrementally added.
func TestSpecFixturesAvailability(t *testing.T) {
	specPath := os.Getenv("TOON_SPEC_PATH")
	if specPath == "" {
		t.Skip("TOON_SPEC_PATH is not set")
	}

	paths := []string{
		filepath.Join(specPath, "tests", "fixtures", "encode", "objects.json"),
		filepath.Join(specPath, "tests", "fixtures", "decode", "objects.json"),
	}

	for _, p := range paths {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("failed to read fixture %s: %v", p, err)
		}

		var b fixtureBundle
		if err := json.Unmarshal(raw, &b); err != nil {
			t.Fatalf("invalid fixture json %s: %v", p, err)
		}
		if b.Version == "" || b.Category == "" || len(b.Tests) == 0 {
			t.Fatalf("fixture missing mandatory fields: %s", p)
		}
	}
}
