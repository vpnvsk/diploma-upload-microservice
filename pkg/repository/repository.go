package repository

import (
	"github.com/vpnvsk/amunetip-patent-upload/internal/config"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/repository/db_repository"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/repository/ktmine_repository"
	"log/slog"
)

type Repository struct {
	KTMineRepository
	DBRepository
}

func NewRepository(log *slog.Logger, cfg *config.Config) *Repository {
	return &Repository{
		KTMineRepository: ktmine_repository.NewKTMineRepository(log, cfg),
		DBRepository:     db_repository.NewDBRepository(log, cfg),
	}
}

type KTMineRepository interface {
}

type DBRepository interface {
}
