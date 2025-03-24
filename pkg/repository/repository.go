package repository

import (
	_ "github.com/lib/pq"
	"github.com/vpnvsk/amunetip-patent-upload/internal/config"
	"github.com/vpnvsk/amunetip-patent-upload/internal/model"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/repository/db_repository"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/repository/ktmine_repository"
	"log/slog"
)

type Repository struct {
	KTMineRepositoryInterface
	DBRepository
}

func NewRepository(log *slog.Logger, cfg *config.Config) *Repository {
	db, err := db_repository.NewPostgresDb(db_repository.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		Username: cfg.DBUsername,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.SSLMode,
	})
	if err != nil {
		panic(err)
	}
	return &Repository{
		KTMineRepositoryInterface: ktmine_repository.NewKTMineRepository(log, cfg),
		DBRepository:              db_repository.NewDBRepository(db, log, cfg),
	}
}

type KTMineRepositoryInterface interface {
	GetFilteredData(filters model.FilterInterface) (*[]byte, error)
}

type DBRepository interface {
	SavePatents(data *model.ParsedPatentDataDB) error
}
