package main

import (
	"distributed_storage/internal/config"
	"distributed_storage/internal/http_server/handler"
	"distributed_storage/internal/http_server/midddleware/logger"
	"distributed_storage/internal/storage"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	if err := godotenv.Load("local.env"); err != nil {
		log.Fatalf("Env load err: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config error load %v", err)
	}

	log := setupLogger(cfg.Env)

	store, err := storage.New()
	if err != nil {
		log.Error("Faild to init storage", slog.String("Error", err.Error()))
		os.Exit(1)
	}

	transactionlogger, err := logger.NewTransactionLog(store)
	if err != nil {
		log.Error("Faild to init transaction", slog.String("Error", err.Error()))
		os.Exit(1)
	}
	transactionlogger.Run()

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	// Зарегистрировать обработчики HTTP-запросов Put, Get, Delete в которых указан путь "/v1/{key}"
	router.Put("/v1/{key}", handler.NewPutHandler(transactionlogger, log, store))
	router.Get("/v1/{key}", handler.NewGetHandler(log, store))
	router.Delete("/v1/{key}", handler.NewDeleteHandler(transactionlogger, log, store))

	log.Info("starting server", slog.String("address: ", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServeTLS("dscrt.env", "dskey.env"); err != nil {
		log.Error("faild to start server")
	}

	log.Error("server stoped")
}

func setupLogger(env string) (log *slog.Logger) {
	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return
}
