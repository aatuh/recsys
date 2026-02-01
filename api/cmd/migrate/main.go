package main

import (
	"github.com/aatuh/recsys-suite/api/migrations"

	toolkitmigrate "github.com/aatuh/api-toolkit-contrib/cmd/migrate"
)

// Build with: go build -o bin/migrate ./cmd/migrate
// Usage:
//
//	DATABASE_URL=postgres://... bin/migrate up
//	DATABASE_URL=postgres://... bin/migrate down
//	DATABASE_URL=postgres://... bin/migrate status
func main() {
	toolkitmigrate.Run(toolkitmigrate.Config{
		Embedded: migrations.Migrations,
	})
}
