# Releasing Varnqor

## Current release target

The first automated target is a portable Windows AMD64 executable.

## Continuous development artifact

Every push to `main` builds the real Wails Windows application, runs CLI smoke checks, generates a SHA-256 checksum, and uploads the executable plus checksum as an Actions artifact retained for seven days.

These are development builds, not signed releases.

## Manual release artifact

Run the `Release Windows Alpha` workflow with `workflow_dispatch`. The workflow builds and uploads a versioned Windows artifact without creating a GitHub Release.

## Tagged prerelease

When the project is ready for public distribution, push a semantic prerelease tag such as:

```text
v0.1.0-alpha.1
```

The workflow will build `varnqor-windows-amd64.exe`, generate `varnqor-windows-amd64.exe.sha256`, upload both, and create a GitHub prerelease with generated notes.

## Integrity verification

On Windows PowerShell:

```powershell
(Get-FileHash .\varnqor-windows-amd64.exe -Algorithm SHA256).Hash.ToLower()
```

Compare that value with the first field in the distributed `.sha256` file.

## Release gates

Do not tag a release merely because the workflow exists. Before the first public tag:

1. verify the generated executable launches on a clean Windows machine;
2. replace the generated/default application icon with the approved Varnqor brand asset;
3. review release notes and privacy claims;
4. note that the current alpha executable is unsigned unless code signing is added.
