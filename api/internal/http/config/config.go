package config

import (
	"strconv"

	"recsys/shared/util"

	"github.com/google/uuid"
)

type Config struct {
	DatabaseURL  string
	DefaultOrgID uuid.UUID
	HalfLifeDays float64 // for popularity decay
}

func Load() (Config, error) {
	var c Config
	c.DatabaseURL = util.MustGetEnv("DATABASE_URL")
	org := util.MustGetEnv("ORG_ID")
	id, err := uuid.Parse(org)
	if err != nil {
		return c, err
	}
	c.DefaultOrgID = id

	hl := util.MustGetEnv("POPULARITY_HALFLIFE_DAYS")
	v, err := strconv.ParseFloat(hl, 64)
	if err != nil {
		panic("POPULARITY_HALFLIFE_DAYS must be a number")
	}
	c.HalfLifeDays = v
	return c, nil
}
