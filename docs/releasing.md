# Releasing Varnqor

## Current release target

The first automated target is a portable Windows AMD64 executable.

Wails builds the frontend, generates missing Windows build assets when needed, compiles the Go application, and writes the executable to `build/bin`.

## CI smoke build

Every push to `main` now includes a Windows Wails build and executes:

```text
varnqor-ci.exe version
```

This catches desktop-only compilation regressions that backend unit tests cannot see.

## Manual release artifact

Run the `Release Windows Alpha` workflow with `workflow_dispatch`. The workflow builds and uploads a versioned Windows artifact without creating a GitHub Release.

## Tagged prerelease

When the project is ready for public distribution, push a semantic prerelease tag such as:

```text
v0.1.0-alpha.1
```

The workflow will build `varnqor-windows-amd64.exe`, upload the artifact, and create a GitHub prerelease with generated notes.

Do not tag a release merely because the workflow exists. Before the first public tag, verify the generated executable on a clean Windows machine and replace the generated/default application icon with the approved Varnqor brand asset.
