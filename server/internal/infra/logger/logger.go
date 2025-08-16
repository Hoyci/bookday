package logger

import (
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/hoyci/bookday/internal/config"
)

func NewLogger(cfg *config.Config) *log.Logger {
	logger := log.NewWithOptions(
		os.Stdout,
		log.Options{
			TimeFormat:      time.Kitchen,
			Formatter:       log.JSONFormatter,
			ReportTimestamp: true,
		},
	)

	if cfg.Debug {
		logger.SetLevel(log.DebugLevel)
		logger.SetReportCaller(true)
	}

	if cfg.Environment == "development" {
		logger.SetFormatter(log.TextFormatter)
	}

	return logger
}
