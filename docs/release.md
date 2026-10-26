# Release process

Varnqor 1.0 uses CI as a release gate. A version is not considered releasable until the same commit passes all required checks.

## Required gates

1. `go mod tidy` produces no module diff.
2. `go test ./internal/...` passes, including provider parser and v1 migration/aggregation tests.
3. `go vet ./internal/...` passes.
4. React/TypeScript production build passes.
5. Wails produces a real Windows executable.
6. The built executable passes CLI smoke checks for version, help, providers, Doctor, and local config.
7. A SHA-256 checksum is generated for the exact executable.

## Stable release

Stable releases use semantic tags such as:

```text
v1.0.0
```

The tag-driven release workflow builds `Varnqor.exe`, generates `Varnqor.exe.sha256`, uploads the CI artifact, and publishes both files to the GitHub release.

## Version sources

- `version.go` is the source used for local/development builds unless overridden by linker flags.
- `wails.json` carries desktop product metadata.
- tagged release builds override the runtime version with the tag name.

These values should agree before a stable tag is cut.

## Signing

The open-source 1.0 pipeline does not currently sign the Windows binary. Windows SmartScreen may therefore display a reputation warning. Do not describe the binary as signed until a real code-signing certificate is added to the release pipeline.

## Provider release discipline

A provider must not be promoted as fully supported merely because it is detected. Release notes and the Agents UI must describe capabilities separately: usage, tool metadata, Waste analysis, and limit sources.
