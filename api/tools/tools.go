//go:build tools
// +build tools

package tools

import (
	_ "ariga.io/atlas/sql/migrate"
	_ "ariga.io/atlas/sql/postgres"
	_ "github.com/swaggo/swag/cmd/swag"
)
