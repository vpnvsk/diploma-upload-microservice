package ktmine_repository

import (
	"github.com/vpnvsk/amunetip-patent-upload/internal/config"
	"log/slog"
)

type KTMineRepository struct {
	log *slog.Logger
	cfg *config.Config
}

func NewKTMineRepository(log *slog.Logger, cfg *config.Config) *KTMineRepository {
	return &KTMineRepository{
		log: log,
		cfg: cfg,
	}
}
