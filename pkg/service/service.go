package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/vpnvsk/amunetip-patent-upload/internal/config"
	"github.com/vpnvsk/amunetip-patent-upload/internal/model"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/repository"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/service/api_client"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/service/broker_client"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/service/db_client"
	"log/slog"
	"sync"
)

type Service struct {
	log *slog.Logger
	APIClientInterface
	DBClient
	BrokerClient
}

func NewService(log *slog.Logger, repo *repository.Repository, cfg *config.Config) *Service {
	return &Service{
		log:                log,
		APIClientInterface: api_client.NewAPIClient(log, repo.KTMineRepositoryInterface, cfg),
		DBClient:           db_client.NewDBClient(log, repo.DBRepository),
		BrokerClient:       broker_client.NewBrokerClient(log, repo.BrokerRepository),
	}
}

type APIClientInterface interface {
	GetData(input model.UploadInput) error
	FilterPatents(ctx context.Context, req model.Filters) (*model.FilteredPatentsResponse, error)
	GetStatistics(ctx context.Context, parsedFilters []model.SingleParsedFilter) (*map[string]interface{}, int, error)
	ParseFilters(filters model.Filters) ([]model.SingleParsedFilter, error)
	GetFilteredChunkFullPatentRaw(ctx context.Context, parsedFilters []model.SingleParsedFilter, offset int, limit int) (*[]byte, error)
	ParseFullPatent(patent interface{}) model.FilteredFullPatent
}

type DBClient interface {
	HandleSavePatents(ctx context.Context, patents []model.FilteredFullPatent, transactionId, bundleId uuid.UUID) error
}

type BrokerClient interface {
	ListenPatentUpload(ctx context.Context, handler func(context.Context, []byte) ([]byte, error))
}

func (s Service) UploadPatentHandler(ctx context.Context, payload []byte) ([]byte, error) {
	var parsedPayload model.UploadPatentPayload
	if err := json.Unmarshal(payload, &parsedPayload); err != nil {
		return nil, fmt.Errorf("failed to parse body: %w", err)
	}
	convertedFilters, err := s.APIClientInterface.ParseFilters(parsedPayload.Filters)
	if err != nil {
		return nil, fmt.Errorf("failed to convert filters: %w", err)
	}
	_, totalPatents, err := s.APIClientInterface.GetStatistics(ctx, convertedFilters)
	if err != nil {
		return nil, err
	}

	parsedResponse := make([]model.FilteredFullPatent, 0, totalPatents)
	var mu sync.Mutex
	const fetchWorkers = 8
	const parseWorkers = 2

	itemChan := make(chan int)
	fetchedChan := make(chan *[]byte)
	errCh := make(chan error, 1)

	var wgFetch sync.WaitGroup
	wgFetch.Add(fetchWorkers)
	for i := 0; i < fetchWorkers; i++ {
		go func() {
			defer wgFetch.Done()
			for offset := range itemChan {
				data, err := s.APIClientInterface.GetFilteredChunkFullPatentRaw(ctx, convertedFilters, offset*20, 20)
				if err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}
				fetchedChan <- data
			}
		}()
	}

	var wgParse sync.WaitGroup
	wgParse.Add(parseWorkers)
	for i := 0; i < parseWorkers; i++ {
		go func() {
			defer wgParse.Done()
			for rawData := range fetchedChan {
				parsed, err := s.parse(rawData)
				if err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}
				mu.Lock()
				parsedResponse = append(parsedResponse, parsed...)
				mu.Unlock()
			}
		}()
	}
	go func() {
		defer close(itemChan)
		for i := 0; i < totalPatents; i += 20 {
			select {
			case itemChan <- i:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		wgFetch.Wait()
		close(fetchedChan)
	}()
	wgParse.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
		err = s.DBClient.HandleSavePatents(ctx, parsedResponse, parsedPayload.TransactionId, parsedPayload.BundleId)
		if err != nil {
			return nil, fmt.Errorf("failed to save data: %s", err)
		}
		response := model.AnalyzePatentsOutput{TransactionId: parsedPayload.TransactionId, BundleId: parsedPayload.BundleId}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to convert response: %s", err)
		}

		return jsonResponse, err
	}
}

func (s Service) parse(patents *[]byte) ([]model.FilteredFullPatent, error) {
	var data map[string]interface{}
	err := json.Unmarshal(*patents, &data)
	if err != nil {
		return nil, err
	}
	response, ok := data["response"].(map[string]interface{})
	if !ok {
		return nil, errors.New("can't parse response body")
	}
	items, ok := response["items"].([]interface{})
	parsedResponse := make([]model.FilteredFullPatent, 0, 20)
	for _, patent := range items {
		parsedPatent := s.APIClientInterface.ParseFullPatent(patent)
		parsedResponse = append(parsedResponse, parsedPatent)
	}
	return parsedResponse, nil
}
