package db_repository

import (
	"github.com/vpnvsk/amunetip-patent-upload/internal/config"
	"log/slog"
)

type DBRepository struct {
	log *slog.Logger
	cfg *config.Config
}

func NewDBRepository(log *slog.Logger, cfg *config.Config) *DBRepository {
	return &DBRepository{
		log: log,
		cfg: cfg,
	}
}
