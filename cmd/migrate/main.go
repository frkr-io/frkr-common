package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/frkr-io/frkr-common/migrate"
)

func main() {
	var dbURL = flag.String("db-url", "", "Database connection URL (required)")
	if err := migrate.RunMigrations(*dbURL); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Migrations completed successfully")
}


