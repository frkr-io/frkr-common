package paths

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MigrationsPath returns the absolute path to frkr-common/migrations directory.
// Uses Go module resolution (go list -m) to find the module location, which
// respects replace directives and works with both local development and published versions.
func MigrationsPath() (string, error) {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/frkr-io/frkr-common")
	output, err := cmd.Output()
	if err == nil {
		moduleDir := strings.TrimSpace(string(output))
		migrationsPath := filepath.Join(moduleDir, "migrations")
		if _, err := os.Stat(migrationsPath); err == nil {
			return filepath.Abs(migrationsPath)
		}
	}

	repoRoot, _ := filepath.Abs("../")
	localPath := filepath.Join(repoRoot, "frkr-common", "migrations")
	if _, err := os.Stat(localPath); err == nil {
		return filepath.Abs(localPath)
	}

	return "", fmt.Errorf("migrations not found: frkr-common module not found via 'go list -m' and local path %s does not exist", localPath)
}
