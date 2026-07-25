# Fluui Release Guide

This document describes the release process for Fluui.

## Prerequisites

- Go 1.25+
- [goreleaser](https://goreleaser.com) installed (`go install github.com/goreleaser/goreleaser@latest`)
- `GITHUB_TOKEN` environment variable set with repo permissions

## Version Scheme

Fluui follows [Semantic Versioning](https://semver.org/):

- **MAJOR** (v1, v2): Breaking API changes
- **MINOR** (v1.1, v1.2): New components, protocols, features (backward compatible)
- **PATCH** (v1.0.1, v1.0.2): Bug fixes only
- **Pre-release**: v1.0.0-beta.1, v1.0.0-rc.1

## Release Steps

### 1. Verify Readiness

```bash
# Ensure all tests pass
GOCACHE=/tmp/go-cache-fluui go test -race -short -count=1 ./... -timeout 300s

# Check for warnings
go vet ./...

# Run benchmarks
make bench
```

### 2. Update CHANGELOG

Update `CHANGELOG.md`:
- Move `[Unreleased]` items to a new versioned section
- Add release date
- Update version constants in `version.go` if needed

### 3. Create Tag

```bash
git add -A
git commit -m "release: v1.0.0-beta.1"
git tag v1.0.0-beta.1
git push origin main --tags
```

### 4. Run Goreleaser

```bash
# Dry run first
goreleaser release --snapshot --clean

# Real release (creates GitHub release)
GITHUB_TOKEN=xxx goreleaser release --clean
```

### 5. Verify

```bash
# In a separate project, verify go get works
go get github.com/topcheer/fluui@v1.0.0-beta.1

# Verify version info
go run -ldflags "-X github.com/topcheer/fluui.Version=v1.0.0-beta.1" ./cmd/demo -- --version
```

## CI Integration

The Makefile includes release targets. For CI (GitHub Actions):

```yaml
# .github/workflows/release.yml
on:
  push:
    tags:
      - 'v*'
jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: go test -race -short ./...
      - uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Post-Release

1. Verify the GitHub Release page shows the new version
2. Verify `go get` pulls the correct version
3. Update `version.go` `Version` back to `"dev"` for next development cycle
4. Add a new `[Unreleased]` section to CHANGELOG.md
