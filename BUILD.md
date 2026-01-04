# Building frkr-common

## Prerequisites

- Go 1.18+ (uses golang-migrate v4.17.0 for compatibility)

## Important Notes

### Plugins Package Location

**CRITICAL**: The plugins package is in the **public** location, not internal:

- ✅ Correct: `github.com/frkr-io/frkr-common/plugins`
- ❌ Wrong: `github.com/frkr-io/frkr-common/internal/plugins`

This was moved to allow other services (frkr-ingest-gateway, etc.) to import the plugin interfaces.

### Migration Package

The migration runner uses `golang-migrate` v4.17.0, which is compatible with Go 1.18. If you need a newer version, you may need to upgrade Go.

## Build Steps

### Step 1: Verify Structure

Ensure the plugins package is in the correct location:

```bash
ls plugins/auth.go plugins/encryption.go
```

Should exist at root level, not in `internal/`.

### Step 2: Build

```bash
cd frkr-common
go mod tidy
go build ./...
```

### Step 3: Run Tests

```bash
go test ./...
```

## Dependencies

- `github.com/golang-migrate/migrate/v4@v4.17.0` - Database migrations
- `github.com/lib/pq` - PostgreSQL/CockroachDB driver
- `golang.org/x/oauth2` - OAuth2/OIDC support

## Troubleshooting

### Error: "use of internal package not allowed"

**Solution**: Ensure you're importing from `plugins`, not `internal/plugins`:
```go
import "github.com/frkr-io/frkr-common/plugins"  // ✅ Correct
```

### Error: "undefined: atomic.Bool"

**Solution**: This means golang-migrate was upgraded to a version requiring Go 1.24+. Downgrade:
```bash
go get github.com/golang-migrate/migrate/v4@v4.17.0
go mod tidy
```

### Tests Require Database

Some migration tests may require a running CockroachDB instance. These tests will be skipped or fail gracefully if the database is not available.

## Usage in Other Repositories

When using frkr-common in other repositories:

1. Add to `go.mod`:
   ```go
   require github.com/frkr-io/frkr-common v0.0.0
   ```

2. Add replace directive (for local development):
   ```go
   replace github.com/frkr-io/frkr-common => ../frkr-common
   ```

3. Import:
   ```go
   import "github.com/frkr-io/frkr-common/plugins"
   import "github.com/frkr-io/frkr-common/migrate"
   ```

