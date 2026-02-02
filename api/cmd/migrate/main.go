package main

import (
	"os"

	"github.com/aatuh/recsys-suite/api/migrations"

	toolkitmigrate "github.com/aatuh/api-toolkit/contrib/v2/cmd/migrate"
)

// Build with: go build -o bin/migrate ./cmd/migrate
// Usage:
//
//	DATABASE_URL=postgres://... bin/migrate up
//	DATABASE_URL=postgres://... bin/migrate down
//	DATABASE_URL=postgres://... bin/migrate status
//	DATABASE_URL=postgres://... bin/migrate preflight
func main() {
	if len(os.Args) > 1 && os.Args[1] == "preflight" {
		runPreflight(os.Args[2:])
		return
	}
	toolkitmigrate.Run(toolkitmigrate.Config{
		Embedded: migrations.Migrations,
	})
}
