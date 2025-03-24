package service

import (
	"github.com/vpnvsk/amunetip-patent-upload/internal/config"
	"github.com/vpnvsk/amunetip-patent-upload/internal/model"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/repository"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/service/api_client"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/service/db_client"
	"log/slog"
)

type Service struct {
	log *slog.Logger
	APIClientInterface
	DBClient
}

func NewService(log *slog.Logger, repo *repository.Repository, cfg *config.Config) *Service {
	return &Service{
		log:                log,
		APIClientInterface: api_client.NewAPIClient(log, repo.KTMineRepositoryInterface, cfg),
		DBClient:           db_client.NewDBClient(log, &repo.DBRepository),
	}
}

type APIClientInterface interface {
	GetData(input model.UploadInput) error
	FilterPatents(req model.Filters) (*model.FilteredPatentsResponse, error)
}

type DBClient interface {
}
