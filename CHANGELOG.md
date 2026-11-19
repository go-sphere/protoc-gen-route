# Changelog

## Unreleased

## v0.0.3 (2026-11-19)

### Added

- Support `PackageDesc.StdContextType` so generated files use the `protogen`
  qualified `context.Context` identifier (falling back to `"context.Context"`
  when unset in custom templates).
- Record plugin version in the generated file header (`Config.PluginVersion`).

### Fixed

- Emit service `// Deprecated: Do not use.` comments only after confirming the
  service has matching methods, using nil-safe `service.Proto.GetOptions().GetDeprecated()`.
- Decode UTF-8 runes properly in `pascalCase` (`utf8.DecodeRuneInString`).
- Validate `OptionsKey` (disallowing `/`, `\`, `.`, whitespace, and invalid
  identifier suffixes) and validate Go identifiers / import paths in
  `ParseGoIdent` and `Config.Validate`.
- Eliminate the duplicate `hasOptionsRule` scan by collecting matching
  services and methods in a single pass.
- Render per-RPC template sections in declaration order and quote operation
  constants with `strconv.Quote`.

## v0.0.2 (2026-09-19)

### Changed

- Refactor generator into reusable `Config` and `Generator` with immutable
  parsed template.
- Escape generated extra-data map keys and values via `strconv.Quote`.
- Update `buf` test fixtures and `options` dependency to `v0.0.2`.
