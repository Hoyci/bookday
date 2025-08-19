package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hoyci/bookday/internal/config"
	"github.com/hoyci/bookday/internal/infra/database/pg"
	"github.com/hoyci/bookday/internal/infra/geocoder"
	"github.com/hoyci/bookday/internal/infra/logger"
	"github.com/hoyci/bookday/internal/order"
	"github.com/hoyci/bookday/internal/routing"
)

func main() {
	cfg := config.GetConfig()
	appLogger := logger.NewLogger(cfg)
	appLogger.Info("starting bookday routing worker", "app_name", cfg.AppName)

	db, err := pg.NewConnection(cfg)
	if err != nil {
		appLogger.Fatal("could not connect to the database", "error", err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	orderRepo := order.NewGORMRepository(db)
	nominatimClient := geocoder.NewNominatimClient(cfg.AppName, "v1.0")
	routingRepo := routing.NewGORMRepository(db)
	routingSvc := routing.NewService(routingRepo, orderRepo, nominatimClient, appLogger)

	// c := cron.New(cron.WithSeconds())

	// _, err = c.AddFunc("0 9 9 * * *", func() {
	appLogger.Info("cron job triggered: starting route generation")

	now := time.Now()
	cutoffTime := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())

	if err := routingSvc.GenerateRoutes(context.Background(), cutoffTime); err != nil {
		appLogger.Error("route generation job failed", "error", err)
	} else {
		appLogger.Info("route generation job completed successfully")
	}
	// })
	// if err != nil {
	// appLogger.Fatal("could not add cron job", "error", err)
	// }

	// c.Start()
	appLogger.Info("cron scheduler started. waiting for jobs...")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	appLogger.Info("shutting down routing worker")
	// ctx := c.Stop()
	// <-ctx.Done()
}
