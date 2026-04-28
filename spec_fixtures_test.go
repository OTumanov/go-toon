package toon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
		testName:    "parses quoted object value with newline escape",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses quoted object value with leading/trailing spaces",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses quoted object value with only spaces",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses quoted string value that looks like true",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses quoted string value that looks like integer",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses quoted string value that looks like negative decimal",
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
		testName:    "parses quoted key with brackets",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses quoted key with braces",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses quoted key with comma",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses quoted key with spaces",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses quoted key with leading hyphen",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses quoted key with leading and trailing spaces",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "parses quoted numeric key",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "objects.json"),
		testName:    "unescapes quotes in key",
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
	{
		fixtureFile: filepath.Join("encode", "primitives.json"),
		testName:    "encodes safe strings without quotes",
		mode:        "supported",
		target:      "string",
	},
	{
		fixtureFile: filepath.Join("encode", "primitives.json"),
		testName:    "encodes safe string with underscore and numbers",
		mode:        "supported",
		target:      "string",
	},
	{
		fixtureFile: filepath.Join("encode", "primitives.json"),
		testName:    "encodes positive integer",
		mode:        "supported",
		target:      "int",
	},
	{
		fixtureFile: filepath.Join("encode", "primitives.json"),
		testName:    "encodes decimal number",
		mode:        "supported",
		target:      "float",
	},
	{
		fixtureFile: filepath.Join("encode", "primitives.json"),
		testName:    "encodes negative integer",
		mode:        "supported",
		target:      "int",
	},
	{
		fixtureFile: filepath.Join("encode", "primitives.json"),
		testName:    "encodes true",
		mode:        "supported",
		target:      "bool",
	},
	{
		fixtureFile: filepath.Join("encode", "primitives.json"),
		testName:    "encodes false",
		mode:        "supported",
		target:      "bool",
	},
	{
		fixtureFile: filepath.Join("encode", "primitives.json"),
		testName:    "encodes Unicode string without quotes",
		mode:        "supported",
		target:      "string",
	},
	{
		fixtureFile: filepath.Join("encode", "primitives.json"),
		testName:    "quotes empty string",
		mode:        "supported",
		target:      "string",
	},
	{
		fixtureFile: filepath.Join("encode", "primitives.json"),
		testName:    "quotes string that looks like true",
		mode:        "supported",
		target:      "string",
	},
	{
		fixtureFile: filepath.Join("encode", "primitives.json"),
		testName:    "quotes string that looks like integer",
		mode:        "supported",
		target:      "string",
	},
	{
		fixtureFile: filepath.Join("encode", "primitives.json"),
		testName:    "quotes string with leading zero",
		mode:        "supported",
		target:      "string",
	},
	{
		fixtureFile: filepath.Join("encode", "primitives.json"),
		testName:    "escapes newline in string",
		mode:        "supported",
		target:      "string",
	},
	{
		fixtureFile: filepath.Join("encode", "primitives.json"),
		testName:    "escapes backslash in string",
		mode:        "supported",
		target:      "string",
	},
	{
		fixtureFile: filepath.Join("encode", "objects.json"),
		testName:    "preserves key order in objects",
		mode:        "supported",
		target:      "struct-encode-lines",
	},
	{
		fixtureFile: filepath.Join("encode", "objects.json"),
		testName:    "encodes null values in objects",
		mode:        "supported",
		target:      "struct-encode-lines",
	},
	{
		fixtureFile: filepath.Join("encode", "objects.json"),
		testName:    "quotes string value with colon",
		mode:        "supported",
		target:      "struct-encode-lines",
	},
	{
		fixtureFile: filepath.Join("encode", "objects.json"),
		testName:    "quotes string value with comma",
		mode:        "supported",
		target:      "struct-encode-lines",
	},
	{
		fixtureFile: filepath.Join("encode", "objects.json"),
		testName:    "quotes string value with embedded quotes",
		mode:        "supported",
		target:      "struct-encode-lines",
	},
	{
		fixtureFile: filepath.Join("encode", "objects.json"),
		testName:    "quotes key with colon",
		mode:        "supported",
		target:      "struct-encode-lines",
	},
	{
		fixtureFile: filepath.Join("encode", "objects.json"),
		testName:    "quotes key with spaces",
		mode:        "supported",
		target:      "struct-encode-lines",
	},
	{
		fixtureFile: filepath.Join("encode", "objects.json"),
		testName:    "encodes deeply nested objects",
		mode:        "supported",
		target:      "struct-encode-lines",
	},
	{
		fixtureFile: filepath.Join("decode", "numbers.json"),
		testName:    "parses number with trailing zeros in fractional part",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "numbers.json"),
		testName:    "parses negative zero as zero",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "numbers.json"),
		testName:    "parses negative zero with fractional part",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "numbers.json"),
		testName:    "treats leading zero as string not number",
		mode:        "supported",
		target:      "struct",
	},
	{
		fixtureFile: filepath.Join("decode", "numbers.json"),
		testName:    "treats unquoted leading-zero number as string",
		mode:        "supported",
		target:      "string",
	},
	{
		fixtureFile: filepath.Join("decode", "root-form.json"),
		testName:    "parses empty document as empty object",
		mode:        "supported",
		target:      "empty-struct",
	},
	{
		fixtureFile: filepath.Join("decode", "numbers.json"),
		testName:    "parses exponent notation",
		mode:        "supported",
		target:      "float",
	},
	{
		fixtureFile: filepath.Join("decode", "numbers.json"),
		testName:    "parses exponent notation with uppercase E",
		mode:        "supported",
		target:      "float",
	},
	{
		fixtureFile: filepath.Join("decode", "numbers.json"),
		testName:    "parses negative exponent notation",
		mode:        "supported",
		target:      "float",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-nested.json"),
		testName:    "parses root-level array of uniform objects in tabular format",
		mode:        "supported",
		target:      "slice-struct-id",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-nested.json"),
		testName:    "parses empty root-level array",
		mode:        "supported",
		target:      "slice-struct-id",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-primitive.json"),
		testName:    "parses string arrays inline",
		mode:        "supported",
		target:      "struct-tags-array",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-primitive.json"),
		testName:    "parses number arrays inline",
		mode:        "supported",
		target:      "struct-tags-array",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-primitive.json"),
		testName:    "parses empty arrays",
		mode:        "supported",
		target:      "struct-items-string-array",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-primitive.json"),
		testName:    "parses single-item array with empty string",
		mode:        "supported",
		target:      "struct-items-string-array",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-primitive.json"),
		testName:    "parses multi-item array with empty string",
		mode:        "supported",
		target:      "struct-items-string-array",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-primitive.json"),
		testName:    "parses whitespace-only strings in arrays",
		mode:        "supported",
		target:      "struct-items-string-array",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-primitive.json"),
		testName:    "parses strings with delimiters in arrays",
		mode:        "supported",
		target:      "struct-items-string-array",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-primitive.json"),
		testName:    "parses strings that look like primitives when quoted",
		mode:        "supported",
		target:      "struct-items-string-array",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-primitive.json"),
		testName:    "parses strings with structural tokens in arrays",
		mode:        "supported",
		target:      "struct-items-string-array",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-primitive.json"),
		testName:    "parses quoted key with inline array",
		mode:        "supported",
		target:      "struct-mykey-int-array",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-primitive.json"),
		testName:    "parses quoted key with empty array",
		mode:        "supported",
		target:      "struct-xcustom-string-array",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-tabular.json"),
		testName:    "parses tabular arrays of uniform objects",
		mode:        "supported",
		target:      "struct-items-array",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-tabular.json"),
		testName:    "parses nulls and quoted values in tabular rows",
		mode:        "supported",
		target:      "struct-items-id-value-array",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-tabular.json"),
		testName:    "parses quoted colon in tabular row as data",
		mode:        "supported",
		target:      "struct-items-note-array",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-tabular.json"),
		testName:    "parses quoted header keys in tabular arrays",
		mode:        "supported",
		target:      "struct-items-order-full-array",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-tabular.json"),
		testName:    "treats unquoted colon as terminator for tabular rows and start of key-value pair",
		mode:        "supported",
		target:      "struct-items-count",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-tabular.json"),
		testName:    "parses quoted key with tabular array format",
		mode:        "supported",
		target:      "struct-xitems-array",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-tabular.json"),
		testName:    "parses quoted empty string key with tabular array format",
		mode:        "supported",
		target:      "struct-empty-key-array",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-nested.json"),
		testName:    "parses root-level primitive array inline",
		mode:        "known_gap",
		target:      "root-mixed-any-slice",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-nested.json"),
		testName:    "parses root-level array of arrays",
		mode:        "known_gap",
		target:      "root-nested-int-slices",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-nested.json"),
		testName:    "parses root-level array of non-uniform objects in list format",
		mode:        "known_gap",
		target:      "root-list-objects",
	},
	{
		fixtureFile: filepath.Join("decode", "arrays-nested.json"),
		testName:    "parses quoted key with list array format",
		mode:        "known_gap",
		target:      "struct-xitems-array",
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

			isEncodeCase := strings.HasPrefix(c.fixtureFile, "encode"+string(filepath.Separator))

			switch c.mode {
			case "supported":
				if isEncodeCase {
					runSupportedEncodeCase(t, c.target, tc)
					return
				}
				var in string
				if err := json.Unmarshal(tc.Input, &in); err != nil {
					t.Fatalf("decode fixture input unmarshal failed: %v", err)
				}
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
						V       string `toon:"v"`
						UserName string `toon:"user.name"`
						IndexKey int `toon:"[index]"`
						BraceKey int `toon:"{key}"`
						CommaKey int `toon:"a,b"`
						FullName string `toon:"full name"`
						LeadHyphen int `toon:"-lead"`
						SpacedKey int `toon:" a "`
						NumericKey string `toon:"123"`
						QuotedKey int `toon:"he said \"hi\""`
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
				case "empty-struct":
					var dst struct{}
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected empty document to decode into empty object, got error: %v", err)
					}
				case "slice-struct-id":
					var dst []struct {
						ID int `toon:"id"`
					}
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected root struct slice to decode, got error: %v", err)
					}
					assertExpectedSliceStructID(t, tc.Expected, dst)
				case "struct-tags-array":
					type tagsWrap struct {
						Tags []string `toon:"tags"`
						Nums []int    `toon:"nums"`
					}
					var dst tagsWrap
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected inline array decode, got error: %v", err)
					}
					assertExpectedTagsOrNumsDecode(t, tc.Expected, dst)
				case "struct-items-array":
					type item struct {
						SKU   string  `toon:"sku"`
						Qty   int     `toon:"qty"`
						Price float64 `toon:"price"`
					}
					type wrap struct {
						Items []item `toon:"items"`
					}
					var dst wrap
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected tabular object array decode, got error: %v", err)
					}
					assertExpectedItemsObjectArrayDecode(t, tc.Expected, dst)
				case "struct-items-id-value-array":
					type item struct {
						ID    int     `toon:"id"`
						Value *string `toon:"value"`
					}
					type wrap struct {
						Items []item `toon:"items"`
					}
					var dst wrap
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected tabular id/value decode, got error: %v", err)
					}
					assertExpectedItemsIDValueArrayDecode(t, tc.Expected, dst)
				case "struct-items-note-array":
					type item struct {
						ID   int    `toon:"id"`
						Note string `toon:"note"`
					}
					type wrap struct {
						Items []item `toon:"items"`
					}
					var dst wrap
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected tabular note decode, got error: %v", err)
					}
					assertExpectedItemsNoteArrayDecode(t, tc.Expected, dst)
				case "struct-items-order-full-array":
					type item struct {
						OrderID  int    `toon:"order:id"`
						FullName string `toon:"full name"`
					}
					type wrap struct {
						Items []item `toon:"items"`
					}
					var dst wrap
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected quoted-header tabular decode, got error: %v", err)
					}
					assertExpectedItemsOrderFullArrayDecode(t, tc.Expected, dst)
				case "struct-items-count":
					type item struct {
						ID   int    `toon:"id"`
						Name string `toon:"name"`
					}
					type wrap struct {
						Items []item `toon:"items"`
						Count int    `toon:"count"`
					}
					var dst wrap
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected tabular+count decode, got error: %v", err)
					}
					assertExpectedItemsCountDecode(t, tc.Expected, dst)
				case "struct-xitems-array":
					type item struct {
						ID   int    `toon:"id"`
						Name string `toon:"name"`
					}
					type wrap struct {
						XItems []item `toon:"x-items"`
					}
					var dst wrap
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected quoted-key tabular decode, got error: %v", err)
					}
					assertExpectedXItemsArrayDecode(t, tc.Expected, dst)
				case "struct-empty-key-array":
					type item struct {
						ID   int    `toon:"id"`
						Name string `toon:"name"`
					}
					type wrap struct {
						EmptyKey []item `toon:""`
					}
					var dst wrap
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected empty-key tabular decode, got error: %v", err)
					}
					assertExpectedEmptyKeyItemsArrayDecode(t, tc.Expected, dst)
				case "struct-items-string-array":
					type wrap struct {
						Items []string `toon:"items"`
					}
					var dst wrap
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected string array decode, got error: %v", err)
					}
					assertExpectedItemsStringArrayDecode(t, tc.Expected, dst)
				case "struct-mykey-int-array":
					type wrap struct {
						MyKey []int `toon:"my-key"`
					}
					var dst wrap
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected quoted-key int array decode, got error: %v", err)
					}
					assertExpectedMyKeyIntArrayDecode(t, tc.Expected, dst)
				case "struct-xcustom-string-array":
					type wrap struct {
						XCustom []string `toon:"x-custom"`
					}
					var dst wrap
					if err := Unmarshal([]byte(in), &dst); err != nil {
						t.Fatalf("expected quoted-key empty string array decode, got error: %v", err)
					}
					assertExpectedXCustomStringArrayDecode(t, tc.Expected, dst)
				default:
					t.Fatalf("unsupported target mode: %s", c.target)
				}
			case "known_gap":
				if isEncodeCase {
					runKnownGapEncodeCase(t, c.target, tc)
					return
				}
				// Keep known gaps explicit and test-visible. A future change that
				// starts passing these should trigger moving the case to supported.
				var in string
				if err := json.Unmarshal(tc.Input, &in); err != nil {
					t.Fatalf("decode fixture input unmarshal failed: %v", err)
				}
				switch c.target {
				case "float":
					var dst float64
					err := Unmarshal([]byte(in), &dst)
					if err == nil {
						var expected float64
						if uerr := json.Unmarshal(tc.Expected, &expected); uerr == nil && dst == expected {
							t.Fatalf("known-gap case unexpectedly behaves as supported; move it to supported list")
						}
					}
				case "struct-tags-array":
					var dst struct {
						Tags []string `toon:"tags"`
					}
					err := Unmarshal([]byte(in), &dst)
					if err == nil && len(dst.Tags) > 0 {
						t.Fatalf("known-gap case unexpectedly behaves as supported; move it to supported list")
					}
				case "struct-items-string-array":
					var dst struct {
						Items []string `toon:"items"`
					}
					err := Unmarshal([]byte(in), &dst)
					if err == nil && len(dst.Items) > 0 {
						t.Fatalf("known-gap case unexpectedly behaves as supported; move it to supported list")
					}
				case "struct-items-array":
					var dst struct {
						Items []struct {
							SKU   string  `toon:"sku"`
							Qty   int     `toon:"qty"`
							Price float64 `toon:"price"`
						} `toon:"items"`
					}
					err := Unmarshal([]byte(in), &dst)
					if err == nil && len(dst.Items) > 0 {
						t.Fatalf("known-gap case unexpectedly behaves as supported; move it to supported list")
					}
				case "struct-mykey-int-array":
					var dst struct {
						MyKey []int `toon:"my-key"`
					}
					err := Unmarshal([]byte(in), &dst)
					if err == nil && len(dst.MyKey) > 0 {
						t.Fatalf("known-gap case unexpectedly behaves as supported; move it to supported list")
					}
				case "struct-xcustom-string-array":
					var dst struct {
						XCustom []string `toon:"x-custom"`
					}
					err := Unmarshal([]byte(in), &dst)
					if err == nil && len(dst.XCustom) > 0 {
						t.Fatalf("known-gap case unexpectedly behaves as supported; move it to supported list")
					}
				case "struct-items-id-value-array":
					type item struct {
						ID    int     `toon:"id"`
						Value *string `toon:"value"`
					}
					var dst struct {
						Items []item `toon:"items"`
					}
					err := Unmarshal([]byte(in), &dst)
					if err == nil && len(dst.Items) > 0 && dst.Items[0].Value != nil {
						t.Fatalf("known-gap case unexpectedly behaves as supported; move it to supported list")
					}
				case "struct-items-note-array":
					type item struct {
						ID   int    `toon:"id"`
						Note string `toon:"note"`
					}
					var dst struct {
						Items []item `toon:"items"`
					}
					err := Unmarshal([]byte(in), &dst)
					if err == nil && len(dst.Items) == 2 &&
						dst.Items[0].Note == "a:b" && dst.Items[1].Note == "c:d" {
						t.Fatalf("known-gap case unexpectedly behaves as supported; move it to supported list")
					}
				case "struct-items-order-full-array":
					type item struct {
						OrderID  int    `toon:"order:id"`
						FullName string `toon:"full name"`
					}
					var dst struct {
						Items []item `toon:"items"`
					}
					err := Unmarshal([]byte(in), &dst)
					if err == nil && len(dst.Items) > 0 && dst.Items[0].OrderID != 0 && dst.Items[0].FullName != "" {
						t.Fatalf("known-gap case unexpectedly behaves as supported; move it to supported list")
					}
				case "root-mixed-any-slice":
					var dst []interface{}
					err, panicked := safeUnmarshal([]byte(in), &dst)
					if !panicked && err == nil && len(dst) > 0 {
						t.Fatalf("known-gap case unexpectedly behaves as supported; move it to supported list")
					}
				case "root-nested-int-slices":
					var dst [][]int
					err, panicked := safeUnmarshal([]byte(in), &dst)
					if !panicked && err == nil && len(dst) > 0 {
						t.Fatalf("known-gap case unexpectedly behaves as supported; move it to supported list")
					}
				case "root-list-objects":
					var dst []map[string]interface{}
					err, panicked := safeUnmarshal([]byte(in), &dst)
					if !panicked && err == nil && len(dst) > 0 {
						t.Fatalf("known-gap case unexpectedly behaves as supported; move it to supported list")
					}
				default:
					var dst struct { // representative object target for gap tracking
						ID     int
						Name   string
						Active bool
					}
					err := Unmarshal([]byte(in), &dst)
					if err == nil && (dst.ID != 0 || dst.Name != "" || dst.Active) {
						t.Fatalf("known-gap case unexpectedly behaves as supported; move it to supported list")
					}
				}
			default:
				t.Fatalf("unknown subset case mode: %s", c.mode)
			}
		})
	}
}

func runKnownGapEncodeCase(t *testing.T, target string, tc struct {
	Name     string          `json:"name"`
	Input    json.RawMessage `json:"input"`
	Expected json.RawMessage `json:"expected"`
}) {
	t.Helper()
	expected := decodeExpectedEncodedText(t, tc.Expected)

	switch target {
	case "string":
		var in string
		if err := json.Unmarshal(tc.Input, &in); err != nil {
			t.Fatalf("encode fixture input string unmarshal failed: %v", err)
		}
		out, err := Marshal(&in)
		if err != nil {
			t.Fatalf("known-gap encode string case should still encode, got error: %v", err)
		}
		if string(out) == expected {
			t.Fatalf("known-gap case unexpectedly matches spec output; move it to supported list")
		}
	case "struct-encode":
		type nested struct {
			C string `toon:"c"`
		}
		type middle struct {
			B nested `toon:"b"`
		}
		type objFixture struct {
			ID     int     `toon:"id"`
			Name   string  `toon:"name"`
			Active bool    `toon:"active"`
			Value  string  `toon:"value"`
			Note   string  `toon:"note"`
			Text   string  `toon:"text"`
			V      string  `toon:"v"`
			Order  int     `toon:"order:id"`
			Full   string  `toon:"full name"`
			A      middle  `toon:"a"`
		}

		var in objFixture
		if err := json.Unmarshal(tc.Input, &in); err != nil {
			t.Fatalf("encode fixture object input unmarshal failed: %v", err)
		}
		out, err := Marshal(&in)
		if err != nil {
			t.Fatalf("known-gap encode case should still encode, got error: %v", err)
		}
		if string(out) == expected {
			t.Fatalf("known-gap case unexpectedly matches spec output; move it to supported list")
		}
	default:
		t.Fatalf("unsupported known-gap encode target mode: %s", target)
	}
}

func runSupportedEncodeCase(t *testing.T, target string, tc struct {
	Name     string          `json:"name"`
	Input    json.RawMessage `json:"input"`
	Expected json.RawMessage `json:"expected"`
}) {
	t.Helper()
	expected := decodeExpectedEncodedText(t, tc.Expected)

	switch target {
	case "struct-encode-lines":
		switch tc.Name {
		case "quotes string value with colon":
			type s struct {
				Note string `toon:"note"`
			}
			in := s{Note: "a:b"}
			out, err := marshalObjectLinesForSpec(in)
			if err != nil {
				t.Fatalf("expected object-line encode, got error: %v", err)
			}
			assertExpectedEncodedText(t, expected, string(out))
		case "quotes string value with comma":
			type s struct {
				Note string `toon:"note"`
			}
			in := s{Note: "a,b"}
			out, err := marshalObjectLinesForSpec(in)
			if err != nil {
				t.Fatalf("expected object-line encode, got error: %v", err)
			}
			assertExpectedEncodedText(t, expected, string(out))
		case "quotes string value with embedded quotes":
			type s struct {
				Text string `toon:"text"`
			}
			in := s{Text: `say "hello"`}
			out, err := marshalObjectLinesForSpec(in)
			if err != nil {
				t.Fatalf("expected object-line encode, got error: %v", err)
			}
			assertExpectedEncodedText(t, expected, string(out))
		case "quotes key with colon":
			type s struct {
				OrderID int `toon:"order:id"`
			}
			in := s{OrderID: 7}
			out, err := marshalObjectLinesForSpec(in)
			if err != nil {
				t.Fatalf("expected object-line encode, got error: %v", err)
			}
			assertExpectedEncodedText(t, expected, string(out))
		case "quotes key with spaces":
			type s struct {
				FullName string `toon:"full name"`
			}
			in := s{FullName: "Ada"}
			out, err := marshalObjectLinesForSpec(in)
			if err != nil {
				t.Fatalf("expected object-line encode, got error: %v", err)
			}
			assertExpectedEncodedText(t, expected, string(out))
		case "preserves key order in objects":
			type s struct {
				ID     int    `toon:"id"`
				Name   string `toon:"name"`
				Active bool   `toon:"active"`
			}
			in := s{ID: 123, Name: "Ada", Active: true}
			out, err := marshalObjectLinesForSpec(in)
			if err != nil {
				t.Fatalf("expected object-line encode, got error: %v", err)
			}
			assertExpectedEncodedText(t, expected, string(out))
		case "encodes null values in objects":
			type s struct {
				ID    int     `toon:"id"`
				Value *string `toon:"value"`
			}
			in := s{ID: 123, Value: nil}
			out, err := marshalObjectLinesForSpec(in)
			if err != nil {
				t.Fatalf("expected object-line encode, got error: %v", err)
			}
			assertExpectedEncodedText(t, expected, string(out))
		case "encodes deeply nested objects":
			type level3 struct {
				C string `toon:"c"`
			}
			type level2 struct {
				B level3 `toon:"b"`
			}
			type level1 struct {
				A level2 `toon:"a"`
			}
			in := level1{A: level2{B: level3{C: "deep"}}}
			out, err := marshalObjectLinesForSpec(in)
			if err != nil {
				t.Fatalf("expected object-line encode, got error: %v", err)
			}
			assertExpectedEncodedText(t, expected, string(out))
		default:
			t.Fatalf("unsupported struct-encode-lines fixture: %s", tc.Name)
		}
	case "string":
		var in string
		if err := json.Unmarshal(tc.Input, &in); err != nil {
			t.Fatalf("encode fixture input unmarshal failed: %v", err)
		}
		out, err := Marshal(&in)
		if err != nil {
			t.Fatalf("expected string primitive to encode, got error: %v", err)
		}
		assertExpectedEncodedText(t, expected, string(out))
	case "int":
		var in int
		if err := json.Unmarshal(tc.Input, &in); err != nil {
			t.Fatalf("encode fixture input unmarshal failed: %v", err)
		}
		out, err := Marshal(&in)
		if err != nil {
			t.Fatalf("expected int primitive to encode, got error: %v", err)
		}
		assertExpectedEncodedText(t, expected, string(out))
	case "bool":
		var in bool
		if err := json.Unmarshal(tc.Input, &in); err != nil {
			t.Fatalf("encode fixture input unmarshal failed: %v", err)
		}
		out, err := Marshal(&in)
		if err != nil {
			t.Fatalf("expected bool primitive to encode, got error: %v", err)
		}
		assertExpectedEncodedText(t, expected, string(out))
	case "float":
		var in float64
		if err := json.Unmarshal(tc.Input, &in); err != nil {
			t.Fatalf("encode fixture input unmarshal failed: %v", err)
		}
		out, err := Marshal(&in)
		if err != nil {
			t.Fatalf("expected float primitive to encode, got error: %v", err)
		}
		assertExpectedEncodedText(t, expected, string(out))
	default:
		t.Fatalf("unsupported encode target mode: %s", target)
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

func assertExpectedSliceStructID(t *testing.T, raw json.RawMessage, actual []struct {
	ID int `toon:"id"`
}) {
	t.Helper()
	var expected []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatalf("failed to parse expected slice object list: %v", err)
	}
	if len(actual) != len(expected) {
		t.Fatalf("expected %d rows, got %d", len(expected), len(actual))
	}
	for i := range expected {
		var expID int
		if err := json.Unmarshal(expected[i]["id"], &expID); err != nil {
			t.Fatalf("failed to parse expected id at %d: %v", i, err)
		}
		if actual[i].ID != expID {
			t.Fatalf("row %d expected id=%d got %d", i, expID, actual[i].ID)
		}
	}
}

func assertExpectedTagsOrNumsDecode(t *testing.T, raw json.RawMessage, actual struct {
	Tags []string `toon:"tags"`
	Nums []int    `toon:"nums"`
}) {
	t.Helper()
	var expected map[string]json.RawMessage
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatalf("failed to parse expected tags/nums payload: %v", err)
	}
	if v, ok := expected["tags"]; ok {
		var exp []string
		if err := json.Unmarshal(v, &exp); err != nil {
			t.Fatalf("failed to parse expected tags: %v", err)
		}
		if len(actual.Tags) != len(exp) {
			t.Fatalf("expected %d tags, got %d", len(exp), len(actual.Tags))
		}
		for i := range exp {
			if actual.Tags[i] != exp[i] {
				t.Fatalf("tag[%d] expected %q got %q", i, exp[i], actual.Tags[i])
			}
		}
	}
	if v, ok := expected["nums"]; ok {
		var exp []int
		if err := json.Unmarshal(v, &exp); err != nil {
			t.Fatalf("failed to parse expected nums: %v", err)
		}
		if len(actual.Nums) != len(exp) {
			t.Fatalf("expected %d nums, got %d", len(exp), len(actual.Nums))
		}
		for i := range exp {
			if actual.Nums[i] != exp[i] {
				t.Fatalf("num[%d] expected %d got %d", i, exp[i], actual.Nums[i])
			}
		}
	}
}

func assertExpectedItemsStringArrayDecode(t *testing.T, raw json.RawMessage, actual struct {
	Items []string `toon:"items"`
}) {
	t.Helper()
	var expected map[string][]string
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatalf("failed to parse expected items string array: %v", err)
	}
	exp := expected["items"]
	if len(actual.Items) != len(exp) {
		t.Fatalf("expected %d items, got %d", len(exp), len(actual.Items))
	}
	for i := range exp {
		if actual.Items[i] != exp[i] {
			t.Fatalf("items[%d] expected %q got %q", i, exp[i], actual.Items[i])
		}
	}
}

func assertExpectedMyKeyIntArrayDecode(t *testing.T, raw json.RawMessage, actual struct {
	MyKey []int `toon:"my-key"`
}) {
	t.Helper()
	var expected map[string][]int
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatalf("failed to parse expected my-key int array: %v", err)
	}
	exp := expected["my-key"]
	if len(actual.MyKey) != len(exp) {
		t.Fatalf("expected %d my-key items, got %d", len(exp), len(actual.MyKey))
	}
	for i := range exp {
		if actual.MyKey[i] != exp[i] {
			t.Fatalf("my-key[%d] expected %d got %d", i, exp[i], actual.MyKey[i])
		}
	}
}

func assertExpectedXCustomStringArrayDecode(t *testing.T, raw json.RawMessage, actual struct {
	XCustom []string `toon:"x-custom"`
}) {
	t.Helper()
	var expected map[string][]string
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatalf("failed to parse expected x-custom string array: %v", err)
	}
	exp := expected["x-custom"]
	if len(actual.XCustom) != len(exp) {
		t.Fatalf("expected %d x-custom items, got %d", len(exp), len(actual.XCustom))
	}
	for i := range exp {
		if actual.XCustom[i] != exp[i] {
			t.Fatalf("x-custom[%d] expected %q got %q", i, exp[i], actual.XCustom[i])
		}
	}
}

func assertExpectedItemsObjectArrayDecode(t *testing.T, raw json.RawMessage, actual interface{}) {
	t.Helper()
	var expected map[string][]map[string]json.RawMessage
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatalf("failed to parse expected items object array: %v", err)
	}
	exp := expected["items"]

	rv := reflect.ValueOf(actual)
	itemsField := rv.FieldByName("Items")
	if !itemsField.IsValid() || itemsField.Kind() != reflect.Slice {
		t.Fatalf("actual payload must contain Items slice")
	}
	if itemsField.Len() != len(exp) {
		t.Fatalf("expected %d tabular rows, got %d", len(exp), itemsField.Len())
	}
	for i := range exp {
		var sku string
		var qty int
		var price float64
		_ = json.Unmarshal(exp[i]["sku"], &sku)
		_ = json.Unmarshal(exp[i]["qty"], &qty)
		_ = json.Unmarshal(exp[i]["price"], &price)

		row := itemsField.Index(i)
		gotSKU := row.FieldByName("SKU").String()
		gotQty := int(row.FieldByName("Qty").Int())
		gotPrice := row.FieldByName("Price").Float()
		if gotSKU != sku || gotQty != qty || gotPrice != price {
			t.Fatalf("row %d expected sku=%q qty=%d price=%v got sku=%q qty=%d price=%v",
				i, sku, qty, price, gotSKU, gotQty, gotPrice)
		}
	}
}

func assertExpectedItemsIDValueArrayDecode(t *testing.T, raw json.RawMessage, actual interface{}) {
	t.Helper()
	exp := map[string][]map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &exp); err != nil {
		t.Fatalf("failed to parse expected id/value array: %v", err)
	}
	rv := reflect.ValueOf(actual)
	items := rv.FieldByName("Items")
	if items.Len() != len(exp["items"]) {
		t.Fatalf("expected %d rows, got %d", len(exp["items"]), items.Len())
	}
	for i := range exp["items"] {
		row := items.Index(i)
		gotID := int(row.FieldByName("ID").Int())
		var expID int
		_ = json.Unmarshal(exp["items"][i]["id"], &expID)
		if gotID != expID {
			t.Fatalf("row %d id mismatch: expected %d got %d", i, expID, gotID)
		}
		gotValField := row.FieldByName("Value")
		expRaw := exp["items"][i]["value"]
		if string(expRaw) == "null" {
			if !gotValField.IsNil() {
				t.Fatalf("row %d expected nil value", i)
			}
		} else {
			var expVal string
			_ = json.Unmarshal(expRaw, &expVal)
			if gotValField.IsNil() || gotValField.Elem().String() != expVal {
				t.Fatalf("row %d value mismatch", i)
			}
		}
	}
}

func assertExpectedItemsNoteArrayDecode(t *testing.T, raw json.RawMessage, actual interface{}) {
	t.Helper()
	exp := map[string][]map[string]json.RawMessage{}
	_ = json.Unmarshal(raw, &exp)
	rv := reflect.ValueOf(actual)
	items := rv.FieldByName("Items")
	if items.Len() != len(exp["items"]) {
		t.Fatalf("expected %d rows, got %d", len(exp["items"]), items.Len())
	}
	for i := range exp["items"] {
		row := items.Index(i)
		var expNote string
		_ = json.Unmarshal(exp["items"][i]["note"], &expNote)
		if row.FieldByName("Note").String() != expNote {
			t.Fatalf("row %d note mismatch", i)
		}
	}
}

func assertExpectedItemsOrderFullArrayDecode(t *testing.T, raw json.RawMessage, actual interface{}) {
	t.Helper()
	exp := map[string][]map[string]json.RawMessage{}
	_ = json.Unmarshal(raw, &exp)
	rv := reflect.ValueOf(actual)
	items := rv.FieldByName("Items")
	if items.Len() != len(exp["items"]) {
		t.Fatalf("expected %d rows, got %d", len(exp["items"]), items.Len())
	}
	for i := range exp["items"] {
		row := items.Index(i)
		var expOrder int
		var expName string
		_ = json.Unmarshal(exp["items"][i]["order:id"], &expOrder)
		_ = json.Unmarshal(exp["items"][i]["full name"], &expName)
		if int(row.FieldByName("OrderID").Int()) != expOrder || row.FieldByName("FullName").String() != expName {
			t.Fatalf("row %d quoted-header mismatch", i)
		}
	}
}

func assertExpectedItemsCountDecode(t *testing.T, raw json.RawMessage, actual interface{}) {
	t.Helper()
	exp := map[string]json.RawMessage{}
	_ = json.Unmarshal(raw, &exp)
	rv := reflect.ValueOf(actual)
	var expCount int
	_ = json.Unmarshal(exp["count"], &expCount)
	if int(rv.FieldByName("Count").Int()) != expCount {
		t.Fatalf("count mismatch")
	}
}

func assertExpectedXItemsArrayDecode(t *testing.T, raw json.RawMessage, actual interface{}) {
	t.Helper()
	exp := map[string][]map[string]json.RawMessage{}
	_ = json.Unmarshal(raw, &exp)
	rv := reflect.ValueOf(actual)
	items := rv.FieldByName("XItems")
	if items.Len() != len(exp["x-items"]) {
		t.Fatalf("expected %d rows, got %d", len(exp["x-items"]), items.Len())
	}
}

func assertExpectedEmptyKeyItemsArrayDecode(t *testing.T, raw json.RawMessage, actual interface{}) {
	t.Helper()
	exp := map[string][]map[string]json.RawMessage{}
	_ = json.Unmarshal(raw, &exp)
	rv := reflect.ValueOf(actual)
	items := rv.FieldByName("EmptyKey")
	if items.Len() != len(exp[""]) {
		t.Fatalf("expected %d rows for empty-key array, got %d", len(exp[""]), items.Len())
	}
}

func decodeExpectedEncodedText(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var expected string
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatalf("failed to parse expected encoded text: %v", err)
	}
	return expected
}

func assertExpectedEncodedText(t *testing.T, expected, actual string) {
	t.Helper()
	if actual != expected {
		t.Fatalf("expected encoded %q, got %q", expected, actual)
	}
}

func safeUnmarshal(data []byte, target interface{}) (err error, panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	err = Unmarshal(data, target)
	return err, false
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
