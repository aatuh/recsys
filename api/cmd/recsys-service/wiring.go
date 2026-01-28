package main

import (
	"recsys/internal/services/recsysvc"

	"github.com/aatuh/api-toolkit-contrib/adapters/validation"
	"github.com/aatuh/api-toolkit/ports"
)

type appDeps struct {
	RecsysService *recsysvc.Service
	Validator     ports.Validator
}

func buildAppDeps(log ports.Logger, pool ports.DatabasePool) appDeps {
	_ = log
	_ = pool
	recSvc := recsysvc.New(recsysvc.NewNoopEngine())

	return appDeps{
		RecsysService: recSvc,
		Validator:     validation.NewBasicValidator(),
	}
}
