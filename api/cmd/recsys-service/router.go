package main

import (
	"recsys/internal/http/handlers"

	"github.com/aatuh/api-toolkit/ports"
)

func mountAppRoutes(r ports.HTTPRouter, log ports.Logger, deps appDeps) {
	recsysHandler := handlers.NewRecsysHandler(
		deps.RecsysService,
		log,
		deps.Validator,
		handlers.WithOverloadRetryAfter(deps.OverloadRetryAfter),
		handlers.WithExposureLogger(deps.ExposureLogger, deps.ExposureHasher),
		handlers.WithExperimentAssigner(deps.ExperimentAssigner),
		handlers.WithExplainControls(deps.ExplainMaxItems, deps.ExplainRequireAdmin, deps.AdminRole),
	)
	recsysHandler.RegisterRoutes(r)

	adminHandler := handlers.NewAdminHandler(deps.AdminService, log, deps.Validator)
	r.Mount("/", adminHandler.Routes())
}
