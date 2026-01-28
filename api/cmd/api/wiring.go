package main

import (
	"recsys/internal/services/foosvc"
	"recsys/internal/store"

	"github.com/aatuh/api-toolkit-contrib/adapters/clock"
	"github.com/aatuh/api-toolkit-contrib/adapters/txpostgres"
	"github.com/aatuh/api-toolkit-contrib/adapters/uuid"
	"github.com/aatuh/api-toolkit-contrib/adapters/validation"
	"github.com/aatuh/api-toolkit/ports"
)

type appDeps struct {
	FooService *foosvc.Service
	Validator  ports.Validator
}

func buildAppDeps(log ports.Logger, pool ports.DatabasePool) appDeps {
	tx := txpostgres.New(pool)
	clk := clock.NewSystemClock()
	ids := uuid.NewUUIDGen()
	val := validation.NewBasicValidator()

	fooRepo := store.NewFooRepo(pool)
	fooSvc := foosvc.New(fooRepo, tx, log, clk, ids)

	return appDeps{
		FooService: fooSvc,
		Validator:  val,
	}
}
