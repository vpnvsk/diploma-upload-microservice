package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/service"
	"log/slog"
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

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	api := router.Group("/api")
	{
		auth := api.Group("/upload")
		{
			auth.POST("/", h.Upload)
		}
	}
	return router
}

func (h *Handler) Upload(c *gin.Context) {}
