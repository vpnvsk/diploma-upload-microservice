package db_client

import (
	"context"
	"github.com/google/uuid"
	"github.com/vpnvsk/amunetip-patent-upload/internal/model"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/repository"
	"log/slog"
)

type DBClient struct {
	log  *slog.Logger
	repo repository.DBRepository
}

func NewDBClient(log *slog.Logger, repo repository.DBRepository) *DBClient {
	return &DBClient{
		log:  log,
		repo: repo,
	}
}

func (s *DBClient) HandleSavePatents(ctx context.Context, patents []model.FilteredFullPatent, transactionId, bundleId uuid.UUID) error {
	if err := s.repo.SavePatents(ctx, patents, transactionId, bundleId); err != nil {
		return err
	}
	return nil
}
