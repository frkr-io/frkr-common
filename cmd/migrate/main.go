package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/frkr-io/frkr-common/migrate"
)

func main() {
	var dbURL = flag.String("db-url", "", "Database connection URL (required)")
	var migrationsPath = flag.String("migrations-path", "./migrations", "Path to migrations directory")
	flag.Parse()

	if *dbURL == "" {
		fmt.Fprintf(os.Stderr, "Error: --db-url is required\n")
		flag.Usage()
		os.Exit(1)
	}

	if err := migrate.RunMigrations(*dbURL, *migrationsPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Migrations completed successfully")
}


