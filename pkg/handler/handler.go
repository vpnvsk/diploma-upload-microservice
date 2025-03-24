package handler

import (
	"github.com/gin-gonic/gin"
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
	mux.HandleFunc("/filter", h.filterPatents)
	mux.HandleFunc("/upload", h.UploadPatents)
	return mux
}

func (h *Handler) Upload(c *gin.Context) {}
