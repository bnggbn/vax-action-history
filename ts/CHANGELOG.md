# CHANGELOG

## [Unreleased]

### Added
- Pure TypeScript implementation of VAX L0
- JCS (JSON Canonicalization Scheme) implementation with cross-language test vectors

### Changed
- N/A

### Fixed
- N/A

---

## 2026-01-06

### Added
- **SDTO SDK** (`src/sdto/`)
  - `FieldSpec` interface: validation rule definition (type, min, max, enum)
  - `FluentAction` class: fluent builder with chained `.set()` calls and instant validation
  - `SchemaBuilder` class: schema definition builder for providers
  - `parseSchema()`: cross-service deserialization from `Record<string, unknown>` to `Record<string, FieldSpec>`
  - Factory functions: `newAction()`, `newSchemaBuilder()` to match Go naming conventions

### Changed
- **Validation logic**
  - String length validation uses numeric parsing (`parseInt()`)
  - Number validation with min/max boundary checks
  - Enum validation with strict equality

### Notes
- This implementation maintains identical validation behavior with the Go SDK
- All validation errors are accumulated and thrown on `finalize()`
