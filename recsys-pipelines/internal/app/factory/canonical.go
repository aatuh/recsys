package factory

import (
	canon "github.com/aatuh/recsys-suite/recsys-pipelines/internal/adapters/datasource/files"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/app/config"
	"github.com/aatuh/recsys-suite/recsys-pipelines/internal/ports/datasource"
)

func BuildCanonicalStore(cfg config.EnvConfig) datasource.CanonicalStore {
	return canon.NewFSCanonicalStore(cfg.CanonicalDir)
}
