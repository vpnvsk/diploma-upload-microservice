package broker_client

import (
	"context"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/repository"
	"log/slog"
)

type BrokerClient struct {
	log  *slog.Logger
	repo repository.BrokerRepository
}

func NewBrokerClient(log *slog.Logger, repo repository.BrokerRepository) BrokerClient {
	return BrokerClient{
		log:  log,
		repo: repo,
	}
}

func (s BrokerClient) ListenPatentUpload(ctx context.Context, handler func(context.Context, []byte) ([]byte, error)) {
	if err := s.repo.ListenAndPublish(ctx, handler); err != nil {
		panic("failed to consume data")
	}
}
