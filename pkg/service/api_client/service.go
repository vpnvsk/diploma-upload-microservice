package api_client

import (
	"github.com/vpnvsk/amunetip-patent-upload/internal/config"
	"github.com/vpnvsk/amunetip-patent-upload/internal/model"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/repository"
	"log/slog"
)

type APIClient struct {
	cfg  *config.Config
	log  *slog.Logger
	repo *repository.KTMineRepository
}

func NewAPIClient(log *slog.Logger, repo *repository.KTMineRepository, cfg *config.Config) *APIClient {
	return &APIClient{
		cfg:  cfg,
		log:  log,
		repo: repo,
	}
}

func (c *APIClient) GetData(input model.UploadInput) error {
	return nil
}
