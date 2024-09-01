package service

import (
	"github.com/vpnvsk/amunetip-patent-upload/pkg/repository"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/service/api_client"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/service/db_client"
	"log/slog"
)

type Service struct {
	log *slog.Logger
	APIClient
	DBClient
}

func NewService(log *slog.Logger, repo *repository.Repository) *Service {
	return &Service{
		log:       log,
		APIClient: api_client.NewAPIClient(log, &repo.KTMineRepository),
		DBClient:  db_client.NewDBClient(log, &repo.DBRepository),
	}
}

type APIClient interface {
}

type DBClient interface {
}
