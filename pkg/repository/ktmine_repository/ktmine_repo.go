package ktmine_repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vpnvsk/amunetip-patent-upload/internal/config"
	"github.com/vpnvsk/amunetip-patent-upload/internal/model"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type KTMineRepository struct {
	log    *slog.Logger
	cfg    *config.Config
	client *http.Client
}

func NewKTMineRepository(log *slog.Logger, cfg *config.Config) *KTMineRepository {
	tr := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     100,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   1000 * time.Second,
	}
	return &KTMineRepository{
		log:    log,
		cfg:    cfg,
		client: client,
	}
}

func (r *KTMineRepository) GetFilteredData(filters model.FilterInterface) (*[]byte, error) {
	op := "repository.GetFilteredData"
	log := r.log.With(slog.String("op", op))

	requestBody, err := json.Marshal(filters)
	if err != nil {
		log.Error("error marshaling request body", err)
		return nil, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/search", r.cfg.KTMineURL), bytes.NewBuffer(requestBody))
	if err != nil {
		log.Error("Error creating request:", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		log.Error("Error making POST request:", err)
		return nil, err
	}
	if resp.StatusCode > 299 || resp.StatusCode < 200 {
		log.Error("Error making POST request: status code", resp.StatusCode)
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &body, nil
}
