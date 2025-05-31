package repository

import (
	"context"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/vpnvsk/amunetip-patent-upload/internal/config"
	"github.com/vpnvsk/amunetip-patent-upload/internal/model"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/repository/db_repository"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/repository/ktmine_repository"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/repository/rabbitmq"
	"log/slog"
)

type Repository struct {
	KTMineRepositoryInterface
	DBRepository
	BrokerRepository
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

	brokerConfig := rabbitmq.BrokerConfig{
		URL:           cfg.BrokerURL,
		ConsumeQueue:  cfg.BrokerConsumeQueue,
		PublishQueue:  cfg.BrokerPublishQueue,
		PrefetchCount: cfg.BrokerPrefetchCount,
	}
	return &Repository{
		KTMineRepositoryInterface: ktmine_repository.NewKTMineRepository(log, cfg),
		DBRepository:              db_repository.NewDBRepository(db, log, cfg),
		BrokerRepository:          rabbitmq.NewBrokerRepo(brokerConfig, log),
	}
}

type KTMineRepositoryInterface interface {
	GetFilteredData(ctx context.Context, filters model.FilterInterface) (*[]byte, error)
}

type DBRepository interface {
	SavePatents(ctx context.Context, patents []model.FilteredFullPatent, transactionId, bundleId uuid.UUID) error
}

type BrokerRepository interface {
	ListenAndPublish(ctx context.Context, handler func(context.Context, []byte) ([]byte, error)) error
}
