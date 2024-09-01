package api_client

import (
	"github.com/vpnvsk/amunetip-patent-upload/pkg/repository"
	"log/slog"
)

type APIClient struct {
	log  *slog.Logger
	repo *repository.KTMineRepository
}

func NewAPIClient(log *slog.Logger, repo *repository.KTMineRepository) *APIClient {
	return &APIClient{
		log:  log,
		repo: repo,
	}
}
