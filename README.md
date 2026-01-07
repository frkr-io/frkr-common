# frkr-common

Shared Go library for frkr services and tools.

## Overview

`frkr-common` provides shared functionality used across all frkr Go services, including:

- **Plugin interfaces** for authentication and encryption
- **Common data models** for streams, tenants, and messages
- **Database utilities** for stream and tenant management
- **Database migrations** for schema management
- **Message definitions** for ingest and streaming protocols

## Dependencies

This repository uses [`frkr-proto`](https://github.com/frkr-io/frkr-proto) as a Git submodule for Protocol Buffer definitions.

### Initializing the Submodule

When cloning this repository, initialize the submodule:

```bash
git submodule update --init --recursive
```

Or clone with submodules:

```bash
git clone --recurse-submodules https://github.com/frkr-io/frkr-common.git
```

### Updating the Submodule

To update to the latest version of `frkr-proto`:

```bash
git submodule update --remote proto
```

## Installation

```bash
go get github.com/frkr-io/frkr-common
```

## Building

Build the migrate tool:

```bash
make build
```

The binary will be created in the `bin/` directory as `bin/migrate`.

To clean build artifacts:

```bash
make clean
```

## Quick Start

### Using Database Utilities

```go
import (
    "database/sql"
    "github.com/frkr-io/frkr-common/db"
    "github.com/frkr-io/frkr-common/models"
    _ "github.com/lib/pq"
)

// Connect to database
db, err := sql.Open("postgres", "postgres://user@localhost/dbname?sslmode=disable")
if err != nil {
    log.Fatal(err)
}

// Create or get a tenant
tenant, err := db.CreateOrGetTenant(db, "my-tenant")
if err != nil {
    log.Fatal(err)
}

// Create a stream
stream, err := db.CreateStream(db, tenant.ID, "my-api", "My API stream", 7)
if err != nil {
    log.Fatal(err)
}
```

### Using Authentication

```go
import "github.com/frkr-io/frkr-common/auth"

// Validate Basic Auth header
username, password, ok := auth.ValidateBasicAuth(authHeader)
if !ok {
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
}
```

### Running Migrations

**Using the migrate tool:**

```bash
make build
./bin/migrate \
  --db-url="postgres://user@localhost/dbname?sslmode=disable" \
  --migrations-path="./migrations"
```

**Or programmatically:**

```go
import "github.com/frkr-io/frkr-common/migrate"

err := migrate.RunMigrations(
    "postgres://user@localhost/dbname?sslmode=disable",
    "/path/to/migrations",
)
if err != nil {
    log.Fatal(err)
}
```

## Package Reference

- **`auth`** - Authentication utilities (Basic Auth validation)
- **`db`** - Database operations (streams, tenants)
- **`gateway`** - Common gateway configuration, broker integration, and health checks
- **`messages`** - Message type definitions for ingest and streaming
- **`migrate`** - Database migration runner
- **`models`** - Common data models (Stream, Tenant)
- **`plugins`** - Plugin interfaces for auth and encryption
- **`util`** - Shared utility functions (validation, passwords)

## Requirements

- Go 1.21 or later
- PostgreSQL-compatible database (for migrations and database operations)
- Kafka-compatible message broker (for message streaming)

## License

Apache 2.0
