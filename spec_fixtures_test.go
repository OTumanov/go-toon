package toon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type fixtureBundle struct {
	Version  string `json:"version"`
	Category string `json:"category"`
	Tests    []struct {
		Name     string          `json:"name"`
		Input    json.RawMessage `json:"input"`
		Expected json.RawMessage `json:"expected"`
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

func readFixtureBundle(t *testing.T, rel string) fixtureBundle {
	t.Helper()

	specPath := os.Getenv("TOON_SPEC_PATH")
	if specPath == "" {
		t.Skip("TOON_SPEC_PATH is not set")
	}

	p := filepath.Join(specPath, "tests", "fixtures", rel)
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", p, err)
	}

	var b fixtureBundle
	if err := json.Unmarshal(raw, &b); err != nil {
		t.Fatalf("invalid fixture json %s: %v", p, err)
	}
	return b
}

// TestSpecFixturesSupportedSubset executes currently supported behavior against
// official fixture inputs. This is intentionally a small, explicit subset that
// can be expanded as feature parity grows.
func TestSpecFixturesSupportedSubset(t *testing.T) {
	b := readFixtureBundle(t, filepath.Join("decode", "objects.json"))

	var found bool
	for _, tc := range b.Tests {
		if tc.Name != "parses empty nested object header" {
			continue
		}
		found = true

		var in string
		if err := json.Unmarshal(tc.Input, &in); err != nil {
			t.Fatalf("decode fixture input unmarshal failed: %v", err)
		}

		// "user:" is an officially documented object-header form.
		// In go-toon current parser, the equivalent supported behavior is
		// successful header parsing and struct decode with zero fields.
		var dst struct{}
		if err := Unmarshal([]byte(in), &dst); err != nil {
			t.Fatalf("expected supported subset case to decode, got error: %v", err)
		}
	}

	if !found {
		t.Fatal("supported subset case not found in official fixtures")
	}
}

func TestSpecFixturesSubsetSummary(t *testing.T) {
	specPath := os.Getenv("TOON_SPEC_PATH")
	if specPath == "" {
		t.Skip("TOON_SPEC_PATH is not set")
	}

	encode := readFixtureBundle(t, filepath.Join("encode", "objects.json"))
	decode := readFixtureBundle(t, filepath.Join("decode", "objects.json"))

	// Keep this as a CI-visible checkpoint to track gradual expansion.
	t.Logf(
		"official fixture coverage checkpoint: encode(objects)=%d decode(objects)=%d supported_subset=%d",
		len(encode.Tests), len(decode.Tests), 1,
	)

	if len(encode.Tests) == 0 || len(decode.Tests) == 0 {
		t.Fatal(fmt.Errorf("unexpected empty official fixture bundle"))
	}
}
