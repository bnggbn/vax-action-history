
-- 20251228 --
# CHANGELOG.md

## [Unreleased]

### Added
- Pure Go implementation of VAX L0
- Generic Schema Generator
- Support for nested structs and arrays
- Comprehensive validation rules (email, url, min/max, etc.)

### Changed
- Migrated from CGO to Pure Go (see rationale in docs)
- Set up WSL2 development environment

### Fixed
- Git line ending issues (added .gitattributes)

-- 20251229 --

### Added
- **internal**
- **schemas package** (`internal/schema/profile.go`)
  - `UpdateProfileDTO` struct with validation tags
  - `GetUpdateProfileSchema()`: returns generated JSON schema for profile update action
- **schema package** (`internal/schema/generator.go`)
  - Generic `Generate[T any]()` function using reflection
  - Struct → JSON Schema conversion (draft-07)
  - Nested struct and array support
  - Validation tag parsing: `min`, `max`, `email`, `url`, `uuid`, `datetime`, `gte`, `lte`, `oneof`, `nullable`
  - Required field detection via `validate:"required"` or non-pointer + no `omitempty`
  - `ToJSON()` helper for pretty-printed output

- **SDTOFactory refactor**
- **consumer package** (`internal/SDTOFactory/consumer/consumer.go`)
  - Schema-based validation engine
  - `validateString()`: enum check, length boundary validation
  - `validateNumber()`: min/max boundary using `big.Rat` for precision
  - `compareNumber()`: arbitrary precision numeric comparison

- **api**
- **api package** (`internal/api/schema_handler.go`)
  - HTTP handler for schema retrieval
  - `HandleGetSchema()`: returns JSON schema by action name query param

-- 20251230 --
### Changed

- **internal**
  - Removed `schema/` package, replaced by `bulider/` switch in SDTOFactory.
  - Removed `schemas/` package.
  - Removed `api/` package.

-- 20260106 --

### Changed
- **SDTOFactory refactor to fluent builder pattern**
  - Moved from `internal/SDTOFactory/` to `pkg/vax/sdto/` (public API)
  - Unified architecture: `FieldSpec` + `FluentAction` + `SchemaBuilder` in single package
  - String length validation now uses numeric parsing (`strconv.Atoi`)
  - Removed split between `consumer/`, `builder/`, `constraint/`, `constructor/`

- **SAE structure**
  - Renamed field: `Payload` → `SDTO` for semantic clarity

### Added
- **Go SDK** (`pkg/vax/sdto/`)
  - `FieldSpec`: validation rule definition (type, min, max, enum)
  - `FluentAction`: fluent builder with chained `.Set()` calls and instant validation
  - `SchemaBuilder`: schema definition builder for providers
  - `ParseSchema()`: cross-service `map[string]any` → `map[string]FieldSpec` conversion
  - Integration tests: Builder → FluentAction → SAE end-to-end flow


### Removed
- `internal/SDTOFactory/{builder,consumer,constraint,constructor}/`
- `vax-demo/` stub
- Unused dependencies in `go.sum`
