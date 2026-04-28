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
- Tracked subset cases (iteration #5): 38 total
  - supported: 38
  - known gaps: 0

## Expansion plan

1. Decode path:
   - add more `decode/objects` compatible cases,
   - then add `decode/primitives` compatible cases.
2. Encode path:
   - add `encode/objects` cases where output model aligns,
   - then `encode/primitives`.
3. Error handling:
   - map official `shouldError` fixture groups to current parser constraints.
4. Coverage reporting:
   - keep a stable “supported subset” list in test code and expose periodic summary in CI logs.
5. Gap migration:
   - when a known-gap case starts passing, move it from `known_gap` to `supported`.

## Notes

- `go-toon` currently prioritizes a performance-oriented format path.
- Some fixtures from the language-agnostic suite target broader syntax and may require intentional design decisions before adoption.
