# frkr-common

Shared Go library for the Traffic Mirroring Platform.

## Purpose

This repository provides shared functionality used across all frkr Go services, including:
- Plugin interfaces (Auth, Encryption)
- Open-source plugin implementations
- Protobuf bindings
- Common utilities and error types
- Database migrations

## Structure

```
frkr-common/
├── internal/
│   ├── plugins/          # Plugin interfaces
│   │   ├── auth.go
│   │   └── encryption.go
│   ├── auth/             # Open auth implementations
│   │   ├── basic.go
│   │   └── oidc.go
│   ├── encryption/       # Open encryption implementations
│   │   └── k8s.go
│   ├── errors/           # Common error types
│   └── models/           # Shared data models
├── migrations/            # Database migrations
├── go.mod
├── go.sum
└── README.md
```

## Dependencies

- `frkr-proto` - Protobuf definitions (Git submodule)
- `github.com/golang-migrate/migrate/v4` - Database migrations
- `github.com/golang-migrate/migrate/v4/database/cockroachdb` - CockroachDB driver

## Usage

```go
import (
    "github.com/frkr-io/frkr-common/auth"
    "github.com/frkr-io/frkr-common/encryption"
)
```

## Building

See [BUILD.md](BUILD.md) for detailed build instructions and important notes about:
- Plugins package location (public, not internal)
- Migration package compatibility
- Usage in other repositories

**Quick Start**:
```bash
go mod tidy
go build ./...
go test ./...
```

## License

Apache 2.0

