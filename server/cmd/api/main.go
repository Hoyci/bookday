package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hoyci/bookday/internal/auth"
	"github.com/hoyci/bookday/internal/catalog"
	"github.com/hoyci/bookday/internal/config"
	"github.com/hoyci/bookday/internal/infra/database/pg"
	"github.com/hoyci/bookday/internal/infra/logger"
	appMiddleware "github.com/hoyci/bookday/internal/middleware"
	"github.com/hoyci/bookday/internal/order"
	"github.com/hoyci/bookday/pkg/jwt"
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

	authRepo := auth.NewGORMRepository(db)
	catalogRepo := catalog.NewGORMRepository(db)
	orderRepo := order.NewGORMRepository(db)

	jwtSvc := jwt.NewService(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, "bookday-server-api", int(cfg.JWTAccessExpMinutes), int(cfg.JWTRefreshExpHours))
	authSvc := auth.NewService(authRepo, appLogger, jwtSvc)
	orderSvc := order.NewService(orderRepo, catalogRepo, authRepo, appLogger)
	catalogSvc := catalog.NewService(catalogRepo, appLogger)

	authHandler := auth.NewHTTPHandler(authSvc)
	orderHandler := order.NewHTTPHandler(orderSvc)
	catalogHandler := catalog.NewHTTPHandler(catalogSvc)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	authMiddleware := appMiddleware.NewAuthenticator(jwtSvc)

	router.Group(func(r chi.Router) {
		authHandler.RegisterRoutes(r)
	})

	router.Group(func(r chi.Router) {
		r.Use(authMiddleware.AuthMiddleware)
		r.Use(appMiddleware.RequireRole("CUSTOMER"))

		orderHandler.RegisterRoutes(r)
	})

	catalogHandler.RegisterRoutes(router)

	listenAddr := fmt.Sprintf(":%d", cfg.Port)
	appLogger.Info("server is starting", "address", listenAddr)
	if err := http.ListenAndServe(listenAddr, router); err != nil {
		appLogger.Fatal("failed to start server", "error", err)
	}
}
