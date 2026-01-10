package services

import (
	"database/sql"

	"github.com/IfedayoAwe/payment-processing-service/config"
	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/utils"
)

type Services struct {
	DB      *sql.DB
	Queries gen.Querier
	Config  *config.Config
	Cache   utils.Cache
}

func NewServices(db *sql.DB, queries gen.Querier, cfg *config.Config, cache utils.Cache) *Services {
	return &Services{
		DB:      db,
		Queries: queries,
		Config:  cfg,
		Cache:   cache,
	}
}
