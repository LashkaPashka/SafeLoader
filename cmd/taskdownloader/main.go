package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LashkaPashka/TaskDownloader/internal/config"
	getfile "github.com/LashkaPashka/TaskDownloader/internal/http-server/handlers/getFile"
	gettask "github.com/LashkaPashka/TaskDownloader/internal/http-server/handlers/getTask"
	savelisturls "github.com/LashkaPashka/TaskDownloader/internal/http-server/handlers/saveListUrls"
	eventbus "github.com/LashkaPashka/TaskDownloader/internal/lib/eventBus"
	"github.com/LashkaPashka/TaskDownloader/internal/service"
	storage "github.com/LashkaPashka/TaskDownloader/internal/storage/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev = "dev"
	envProd = "prod"
)

func main() {
	// TODO: init config
	cfg := config.MustLoad()

	// TODO: init logger
	logger := setupLogger(cfg.Env)

	logger.Info(
		"starting task-donlowader",
		slog.String("env", cfg.Env),
		slog.String("version", "@1.0.1"),
	)

	// TODO: init router
	router := chi.NewRouter()
	
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	// TODO: Init storage
	storage, err := storage.New(cfg.StoragePath, logger)
	if err != nil {
		logger.Error("Error init storage")
		return
	}

	// TODO: Init eventBus
	eventbus := eventbus.NewEventBus()

	// TODO: Init storage
	service, err := service.New(storage, cfg.LocalPathStoage, eventbus, logger)
	if err != nil {
		logger.Error("Error init service")
		return
	}

	router.Route("/tasks", func(r chi.Router) {
		r.Post("/", savelisturls.New(service, logger))
		r.Get("/", gettask.New(service, logger))
		r.Get("/", getfile.New(service, logger))
	})

	go service.CompleteTask()

	err = service.SearchQueuedAndComplete()
	if err != nil {
		logger.Error("Invalid searchQueuedAndComplete", slog.String("error", err.Error()))
		return
	}

	logger.Info("starting server", slog.String("address", cfg.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr: cfg.Address,
		Handler: router,
		ReadTimeout: cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout: cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("failed to stop server")
		}
	}()
	
	logger.Info("server started")
	
	<-done
	logger.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("failed to stop server")
		return
	}

	logger.Info("server stopped")
}	


func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log  = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}