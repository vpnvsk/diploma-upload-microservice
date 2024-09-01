package db_client

import (
	"github.com/vpnvsk/amunetip-patent-upload/pkg/repository"
	"log/slog"
)

type DBClient struct {
	log  *slog.Logger
	repo *repository.DBRepository
}

func NewDBClient(log *slog.Logger, repo *repository.DBRepository) *DBClient {
	return &DBClient{
		log:  log,
		repo: repo,
	}
}
