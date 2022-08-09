package logger

import (
	"fmt"

	"github.com/form3tech-oss/http-message-signing-proxy/config"
	log "github.com/sirupsen/logrus"
)

func Configure(cfg config.LogConfig) error {
	level, err := log.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("failed to set log level: %w", err)
	}
	log.SetLevel(level)

	switch cfg.Format {
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	case "text":
		log.SetFormatter(&log.TextFormatter{})
	default:
		return fmt.Errorf("invalid log format '%s', allowed values are [json, text]", cfg.Format)
	}

	return nil
}
