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
	target      string // struct | int | bool | string | float | null-int
}

var trackedSubsetCases = []subsetCase{
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses empty nested object header",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses objects with primitive values",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses null values in objects",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses quoted object value with colon",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses quoted object value with escaped quotes",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses quoted key with colon",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses dotted keys as identifiers",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses deeply nested objects with indentation",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "primitives.json"),
		testName:    "parses positive integer",
		mode:        "supported",
		target:      "int",
	},
	{
		fixtureFile: filepath.Join("decode", "primitives.json"),
		testName:    "parses true",
		mode:        "supported",
		target:      "bool",
	},
	{
		fixtureFile: filepath.Join("decode", "primitives.json"),
		testName:    "parses false",
		mode:        "supported",
		target:      "bool",
	},
	{
		fixtureFile: filepath.Join("decode", "primitives.json"),
		testName:    "parses decimal number",
		mode:        "supported",
		target:      "float",
	},
	{
		fixtureFile: filepath.Join("decode", "primitives.json"),
		testName:    "parses negative integer",
		mode:        "supported",
		target:      "int",
	},
	{
		fixtureFile: filepath.Join("decode", "primitives.json"),
		testName:    "parses null",
		mode:        "supported",
		target:      "null-int",
	},
	{
		fixtureFile: filepath.Join("decode", "primitives.json"),
		testName:    "parses quoted string with escaped quotes",
		mode:        "supported",
		target:      "string",
	},
	{
		fixtureFile: filepath.Join("decode", "primitives.json"),
		testName:    "parses safe unquoted string",
		mode:        "supported",
		target:      "string",
	},
	{
		fixtureFile: filepath.Join("decode", "primitives.json"),
		testName:    "parses quoted string with newline escape",
		mode:        "supported",
		target:      "string",
	},
	{
		fixtureFile: filepath.Join("decode", "primitives.json"),
		testName:    "parses quoted string with tab escape",
		mode:        "supported",
		target:      "string",
	},
	{
		fixtureFile: filepath.Join("decode", "primitives.json"),
		testName:    "parses quoted string with carriage return escape",
		mode:        "supported",
		target:      "string",
	},
	{
		fixtureFile: filepath.Join("decode", "primitives.json"),
		testName:    "parses quoted string with backslash escape",
		mode:        "supported",
		target:      "string",
	},
	{
		fixtureFile: filepath.Join("decode", "primitives.json"),
		testName:    "respects ambiguity quoting for integer",
		mode:        "supported",
		target:      "string",
	},
	{
		fixtureFile: filepath.Join("decode", "primitives.json"),
		testName:    "respects ambiguity quoting for true",
		mode:        "supported",
		target:      "string",
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
				switch c.target {
				case "struct":
					var dst struct {
						ID      int
						Name    string
						Active  bool
						Value   string
						Note    string
						OrderID int `toon:"order:id"`
						Text    string
						UserName string `toon:"user.name"`
						A        struct {
							B struct {
								C string `toon:"c"`
							} `toon:"b"`
						} `toon:"a"`
					}
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected supported subset case to decode, got error: %v", err)
					}
				case "int":
					var dst int
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected int primitive to decode, got error: %v", err)
					}
					assertExpectedInt(t, tc.Expected, dst)
				case "bool":
					var dst bool
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected bool primitive to decode, got error: %v", err)
					}
					assertExpectedBool(t, tc.Expected, dst)
				case "string":
					var dst string
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected string primitive to decode, got error: %v", err)
					}
					assertExpectedString(t, tc.Expected, dst)
				case "float":
					var dst float64
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected float primitive to decode, got error: %v", err)
					}
					assertExpectedFloat(t, tc.Expected, dst)
				case "null-int":
					dst := 777
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected null primitive to decode, got error: %v", err)
					}
					if dst != 0 {
						t.Fatalf("expected null to set zero-value, got %d", dst)
					}
				default:
					t.Fatalf("unsupported target mode: %s", c.target)
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

func assertExpectedInt(t *testing.T, raw json.RawMessage, actual int) {
	t.Helper()
	var expected int
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatalf("failed to parse expected int: %v", err)
	}
	if actual != expected {
		t.Fatalf("expected %d, got %d", expected, actual)
	}
}

func assertExpectedBool(t *testing.T, raw json.RawMessage, actual bool) {
	t.Helper()
	var expected bool
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatalf("failed to parse expected bool: %v", err)
	}
	if actual != expected {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

func assertExpectedString(t *testing.T, raw json.RawMessage, actual string) {
	t.Helper()
	var expected string
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatalf("failed to parse expected string: %v", err)
	}
	if actual != expected {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}

func assertExpectedFloat(t *testing.T, raw json.RawMessage, actual float64) {
	t.Helper()
	var expected float64
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatalf("failed to parse expected float: %v", err)
	}
	if actual != expected {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
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
