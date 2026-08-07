package postgres

import (
	"anti-scam-trainer/backend/internal/core/config"

	"github.com/go-pg/pg"
)

func Connect(cfg config.Config) *pg.DB {
	return pg.Connect(&pg.Options{
		Addr:     cfg.DatabaseAddress,
		User:     cfg.DatabaseUser,
		Password: cfg.DatabasePassword,
		Database: cfg.DatabaseName,
	})
}
