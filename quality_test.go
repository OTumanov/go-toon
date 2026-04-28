package toon

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/quick"
)

func TestUnmarshalInvalidInputs(t *testing.T) {
	t.Parallel()

	type row struct {
		name string
		data string
		kind string
	}

	cases := []row{
		{name: "invalid root int token", data: "abc", kind: "int"},
		{name: "invalid root bool token", data: "truthy", kind: "bool"},
		{name: "array size mismatch in list format", data: "items[1]:\n  - 1\n  - 2", kind: "slice"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var err error
			switch tc.kind {
			case "int":
				var dst int
				err = Unmarshal([]byte(tc.data), &dst)
			case "bool":
				var dst bool
				err = Unmarshal([]byte(tc.data), &dst)
			case "slice":
				var dst []struct {
					ID   int    `toon:"id"`
					Name string `toon:"name"`
				}
				err = Unmarshal([]byte(tc.data), &dst)
			default:
				t.Fatalf("unknown kind %q", tc.kind)
			}
			if err == nil {
				t.Fatalf("expected error for malformed input, got nil")
			}
		})
	}
}

type quickRoundTrip struct {
	ID     int
	Name   string
	Active bool
}

func TestQuickRoundTripMarshalUnmarshal(t *testing.T) {
	t.Parallel()

	cfg := &quick.Config{MaxCount: 100}
	prop := func(in quickRoundTrip) bool {
		in.Name = sanitizeASCII(in.Name)
		if in.ID > 1_000_000 {
			in.ID = 1_000_000
		}
		if in.ID < -1_000_000 {
			in.ID = -1_000_000
		}

		data, err := Marshal(&in)
		if err != nil {
			return false
		}

		var out quickRoundTrip
		if err := Unmarshal(data, &out); err != nil {
			return false
		}
		return reflect.DeepEqual(out, in)
	}

	if err := quick.Check(prop, cfg); err != nil {
		t.Fatalf("round-trip property failed: %v", err)
	}
}

func TestOfficialFixtureBundlesPresence(t *testing.T) {
	t.Parallel()

	specPath := os.Getenv("TOON_SPEC_PATH")
	if specPath == "" {
		t.Skip("TOON_SPEC_PATH is not set")
	}

	files := []string{
		filepath.Join("encode", "whitespace.json"),
		filepath.Join("encode", "primitives.json"),
		filepath.Join("encode", "objects.json"),
		filepath.Join("encode", "key-folding.json"),
		filepath.Join("encode", "delimiters.json"),
		filepath.Join("encode", "arrays-tabular.json"),
		filepath.Join("encode", "arrays-primitive.json"),
		filepath.Join("encode", "arrays-objects.json"),
		filepath.Join("encode", "arrays-nested.json"),
		filepath.Join("decode", "whitespace.json"),
		filepath.Join("decode", "validation-errors.json"),
		filepath.Join("decode", "root-form.json"),
		filepath.Join("decode", "primitives.json"),
		filepath.Join("decode", "path-expansion.json"),
		filepath.Join("decode", "objects.json"),
		filepath.Join("decode", "numbers.json"),
		filepath.Join("decode", "indentation-errors.json"),
		filepath.Join("decode", "delimiters.json"),
		filepath.Join("decode", "blank-lines.json"),
		filepath.Join("decode", "arrays-tabular.json"),
		filepath.Join("decode", "arrays-primitive.json"),
		filepath.Join("decode", "arrays-nested.json"),
	}

	for _, rel := range files {
		rel := rel
		t.Run(rel, func(t *testing.T) {
			p := filepath.Join(specPath, "tests", "fixtures", rel)
			raw, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("failed to read fixture bundle: %v", err)
			}
			var b fixtureBundle
			if err := json.Unmarshal(raw, &b); err != nil {
				t.Fatalf("invalid fixture bundle json: %v", err)
			}
			if b.Version == "" || b.Category == "" {
				t.Fatalf("fixture bundle missing metadata")
			}
			if len(b.Tests) == 0 {
				t.Fatalf("fixture bundle has no tests")
			}
			for i, tc := range b.Tests {
				if tc.Name == "" {
					t.Fatalf("test[%d] has empty name", i)
				}
				if len(tc.Input) == 0 {
					t.Fatalf("test[%d] has empty input", i)
				}
				if len(tc.Expected) == 0 {
					t.Fatalf("test[%d] has empty expected", i)
				}
			}
		})
	}
}

func sanitizeASCII(s string) string {
	out := make([]byte, 0, 24)
	for i := 0; i < len(s) && len(out) < 24; i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			out = append(out, c)
			continue
		}
		if c >= 'A' && c <= 'Z' {
			out = append(out, c)
			continue
		}
		if c >= '0' && c <= '9' {
			out = append(out, c)
			continue
		}
		if c == '-' || c == '_' || c == ' ' {
			out = append(out, c)
		}
	}
	if len(out) == 0 {
		return "x"
	}
	return string(out)
}

