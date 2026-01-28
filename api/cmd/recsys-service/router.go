package main

import (
	"recsys/internal/http/handlers"

	"github.com/aatuh/api-toolkit/ports"
)

func mountAppRoutes(r ports.HTTPRouter, log ports.Logger, deps appDeps) {
	recsysHandler := handlers.NewRecsysHandler(deps.RecsysService, log, deps.Validator)
	recsysHandler.RegisterRoutes(r)
}
