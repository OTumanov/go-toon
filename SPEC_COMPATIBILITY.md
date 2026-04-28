# TOON Spec Compatibility

This document tracks `go-toon` compatibility work against the official TOON spec fixtures:

- Official fixtures: <https://github.com/toon-format/spec/tree/main/tests>
- Fixture schema and structure: <https://github.com/toon-format/spec/blob/main/tests/README.md>

## Current approach

`go-toon` validates spec integration in CI via:

- fixture ingestion checks (files readable, schema-shaped content present),
- explicit supported-subset behavioral checks,
- incremental expansion of covered cases.

The objective is to evolve coverage transparently without claiming full conformance before it is actually implemented.

## Current checkpoint

- Fixture source: `toon-format/spec/tests` (live checkout in CI workflow)
- Automated fixture ingestion: enabled
- Behavioral subset checks: enabled (initial decode subset)
- Known-gap assertions: enabled (explicit unsupported cases are tracked in tests)
- Tracked subset cases (iteration #6): 46 total
  - supported: 38
  - known gaps: 8

## Expansion plan

1. Decode path:
   - add more `decode/objects` compatible cases,
   - then add `decode/primitives` compatible cases.
2. Encode path:
   - add `encode/objects` cases where output model aligns.
3. Error handling:
   - map official `shouldError` fixture groups to current parser constraints.
4. Coverage reporting:
   - keep a stable “supported subset” list in test code and expose periodic summary in CI logs.
5. Gap migration:
   - when a known-gap case starts passing, move it from `known_gap` to `supported`.

## Current known gaps (encode/objects)

The current 8 tracked known gaps are all in `encode/objects` and share the same root cause:
`go-toon` currently emits a header/body representation for structs, while the official fixture cases expect line-oriented object output.

Tracked cases:

1. `preserves key order in objects`
2. `encodes null values in objects`
3. `quotes string value with colon`
4. `quotes string value with comma`
5. `quotes string value with embedded quotes`
6. `quotes key with colon`
7. `quotes key with spaces`
8. `encodes deeply nested objects`

### Recommended closure order

1. **Output mode abstraction**
   - introduce an internal object-line writer for struct encode path (without removing current fast header/body path).
2. **String escaping/quoting parity**
   - implement quoting and escaping rules needed by object-line fixture outputs.
3. **Key quoting parity**
   - add key quoting for reserved characters and spaces.
4. **Nested object emission**
   - add indentation-aware nested object output.
5. **Gap migration**
   - move each case from `known_gap` to `supported` as soon as exact fixture output matches.

## Notes

- `go-toon` currently prioritizes a performance-oriented format path.
- Some fixtures from the language-agnostic suite target broader syntax and may require intentional design decisions before adoption.
