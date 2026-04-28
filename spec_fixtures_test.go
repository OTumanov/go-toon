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

type subsetCase struct {
	fixtureFile string
	testName    string
	mode        string // supported | known_gap
}

var trackedSubsetCases = []subsetCase{
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses empty nested object header",
		mode:        "supported",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses objects with primitive values",
		mode:        "supported",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses null values in objects",
		mode:        "supported",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses quoted object value with colon",
		mode:        "supported",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses quoted object value with escaped quotes",
		mode:        "supported",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses quoted key with colon",
		mode:        "supported",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses dotted keys as identifiers",
		mode:        "known_gap",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses deeply nested objects with indentation",
		mode:        "known_gap",
	},
	{
		fixtureFile: filepath.Join("decode", "primitives.json"),
		testName:    "parses positive integer",
		mode:        "known_gap",
	},
	{
		fixtureFile: filepath.Join("decode", "primitives.json"),
		testName:    "parses true",
		mode:        "known_gap",
	},
	{
		fixtureFile: filepath.Join("decode", "primitives.json"),
		testName:    "parses quoted string with escaped quotes",
		mode:        "known_gap",
	},
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
	for _, c := range trackedSubsetCases {
		t.Run(c.testName, func(t *testing.T) {
			b := readFixtureBundle(t, c.fixtureFile)
			tc, ok := findFixtureTestByName(b, c.testName)
			if !ok {
				t.Fatalf("fixture test not found: %s (%s)", c.testName, c.fixtureFile)
			}

			var in string
			if err := json.Unmarshal(tc.Input, &in); err != nil {
				t.Fatalf("decode fixture input unmarshal failed: %v", err)
			}

			switch c.mode {
			case "supported":
				var dst struct {
					ID     int
					Name   string
					Active bool
					Value  string
					Note   string
					OrderID int `toon:"order:id"`
					Text    string
				}
				if err := Unmarshal([]byte(in), &dst); err != nil {
					t.Fatalf("expected supported subset case to decode, got error: %v", err)
				}
			case "known_gap":
				// Keep known gaps explicit and test-visible. A future change that
				// starts passing these should trigger moving the case to supported.
				var dst struct { // representative object target for gap tracking
					ID     int
					Name   string
					Active bool
				}
				err := Unmarshal([]byte(in), &dst)
				if err == nil && (dst.ID != 0 || dst.Name != "" || dst.Active) {
					t.Fatalf("known-gap case unexpectedly behaves as supported; move it to supported list")
				}
			default:
				t.Fatalf("unknown subset case mode: %s", c.mode)
			}
		})
	}
}

func TestSpecFixturesSubsetSummary(t *testing.T) {
	specPath := os.Getenv("TOON_SPEC_PATH")
	if specPath == "" {
		t.Skip("TOON_SPEC_PATH is not set")
	}

	encode := readFixtureBundle(t, filepath.Join("encode", "objects.json"))
	decode := readFixtureBundle(t, filepath.Join("decode", "objects.json"))
	supported, known := subsetCaseCounters()

	// Keep this as a CI-visible checkpoint to track gradual expansion.
	t.Logf(
		"official fixture coverage checkpoint: encode(objects)=%d decode(objects)=%d supported_subset=%d known_gaps=%d",
		len(encode.Tests), len(decode.Tests), supported, known,
	)

	if len(encode.Tests) == 0 || len(decode.Tests) == 0 {
		t.Fatal(fmt.Errorf("unexpected empty official fixture bundle"))
	}
}

func subsetCaseCounters() (supported int, knownGaps int) {
	for _, c := range trackedSubsetCases {
		switch c.mode {
		case "supported":
			supported++
		case "known_gap":
			knownGaps++
		}
	}
	return supported, knownGaps
}

func findFixtureTestByName(b fixtureBundle, name string) (struct {
	Name     string          `json:"name"`
	Input    json.RawMessage `json:"input"`
	Expected json.RawMessage `json:"expected"`
}, bool) {
	for _, tc := range b.Tests {
		if tc.Name == name {
			return tc, true
		}
	}
	return struct {
		Name     string          `json:"name"`
		Input    json.RawMessage `json:"input"`
		Expected json.RawMessage `json:"expected"`
	}{}, false
}
