# Go Module Setup for Local Development

## Problem

When developing locally, we want to:
1. Try importing from GitHub first (as if published)
2. Fallback to local path if GitHub import fails

## Solution

Go doesn't support automatic fallback, but we can use `replace` directives in `go.mod` for local development.

## Pattern

### For Local Development

Each repo that uses `frkr-common` should have a `replace` directive:

```go
// go.mod
module github.com/frkr-io/frkr-tools

require (
    github.com/frkr-io/frkr-common v0.0.0
)

replace github.com/frkr-io/frkr-common => ../frkr-common
```

### For Published Versions

When publishing to GitHub, remove the `replace` directive and use actual versions:

```go
// go.mod (published)
module github.com/frkr-io/frkr-tools

require (
    github.com/frkr-io/frkr-common v0.1.0
)

// No replace directive - uses GitHub version
```

## Directory Structure

```
frkr-io/
├── frkr-common/          # Shared library
├── frkr-operator/        # Uses frkr-common
├── frkr-tools/            # Uses frkr-common
└── frkr-ingest-gateway/  # Uses frkr-common
```

Each repo uses `replace` to point to local `../frkr-common`.

## Workflow

1. **Local Development**: Use `replace` directives
2. **CI/CD**: Can use `replace` or actual versions (depending on setup)
3. **Published**: Remove `replace`, use GitHub versions

## Example: frkr-tools/go.mod

```go
module github.com/frkr-io/frkr-tools

go 1.18

require (
    github.com/frkr-io/frkr-common v0.0.0
    github.com/spf13/cobra v1.8.0
)

// Local development - points to sibling directory
replace github.com/frkr-io/frkr-common => ../frkr-common
```

## Note

Go doesn't support automatic fallback, but this pattern works well:
- Local: `replace` directive
- Published: Remove `replace`, use GitHub versions
- CI: Can use either approach

