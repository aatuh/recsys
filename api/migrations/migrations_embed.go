package migrations

import "embed"

// Embed migrations so binaries can run without a filesystem dir.
//
//go:embed *.sql
var Migrations embed.FS
