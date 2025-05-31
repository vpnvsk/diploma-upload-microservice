package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/vpnvsk/amunetip-patent-upload/internal/logger"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/service"
	"log/slog"
	"net/http"
)

type Handler struct {
	log     *slog.Logger
	service *service.Service
}

func NewHandler(log *slog.Logger, service *service.Service) *Handler {
	return &Handler{
		log:     log,
		service: service,
	}
}

func (h *Handler) InitRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/filter", logger.LoggingMiddleware(h.log, http.HandlerFunc(h.filterPatents)))
	mux.Handle("/upload", logger.LoggingMiddleware(h.log, http.HandlerFunc(h.UploadPatents)))
	return mux
}

func (h *Handler) HandlePatentUpload(ctx context.Context) {
	h.service.BrokerClient.ListenPatentUpload(ctx, h.service.UploadPatentHandler)
}

func (h *Handler) Upload(c *gin.Context) {}
