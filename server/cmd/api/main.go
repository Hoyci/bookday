package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hoyci/bookday/internal/catalog"
	"github.com/hoyci/bookday/internal/config"
	"github.com/hoyci/bookday/internal/infra/database/pg"
	"github.com/hoyci/bookday/internal/infra/logger"
	"github.com/hoyci/bookday/internal/order"
)

func main() {
	cfg := config.GetConfig()

	appLogger := logger.NewLogger(cfg)
	appLogger.Info("starting bookday application", "app_name", cfg.AppName, "env", cfg.Environment)

	db, err := pg.NewConnection(cfg)
	if err != nil {
		appLogger.Fatal("could not connect to the database", "error", err)
	}
	appLogger.Info("database connection established")

	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Fatal("could not get underlying sql.DB from gorm", "error", err)
	}
	defer sqlDB.Close()

	catalogRepo := catalog.NewGORMRepository(db)
	catalogSvc := catalog.NewService(catalogRepo, appLogger)
	catalogHandler := catalog.NewHTTPHandler(catalogSvc)

	orderRepo := order.NewGORMRepository(db)
	orderSvc := order.NewService(orderRepo, catalogRepo, appLogger)
	orderHandler := order.NewHTTPHandler(orderSvc)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	catalogHandler.RegisterRoutes(router)
	orderHandler.RegisterRoutes(router)

	listenAddr := fmt.Sprintf(":%d", cfg.Port)
	appLogger.Info("server is starting", "address", listenAddr)
	if err := http.ListenAndServe(listenAddr, router); err != nil {
		appLogger.Fatal("failed to start server", "error", err)
	}
}
