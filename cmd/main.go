package main

import (
	"context"
	"fmt"
	"github.com/vpnvsk/amunetip-patent-upload/internal"
	"github.com/vpnvsk/amunetip-patent-upload/internal/config"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/handler"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/repository"
	"github.com/vpnvsk/amunetip-patent-upload/pkg/service"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.LoadConfig()
	log := setUpLogger(cfg.ENV)
	repo := repository.NewRepository(log, cfg)
	serv := service.NewService(log, repo, cfg)
	handl := handler.NewHandler(log, serv)
	srv := new(internal.Server)
	go func() {
		if err := srv.Run("8080", handl.InitRoutes()); err != nil {
			errorMessage := fmt.Sprintf("error while running server %s", err.Error())
			panic(errorMessage)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	if err := srv.ShutDown(context.Background()); err != nil {
		log.Error("error while shutting down: %s", err.Error())
	}
}
func setUpLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case "prod":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return log
}
