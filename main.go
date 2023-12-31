package main

import (
	"time"

	"github.com/fabiocorneti/todoist-assistant/internal/config"
	"github.com/fabiocorneti/todoist-assistant/internal/process"
)

func main() {
	cfg := config.GetConfiguration()
	logger := config.GetLogger()

	ticker := time.NewTicker(time.Duration(cfg.UpdateInterval) * time.Minute)
	defer ticker.Stop()

	process.RunProcess(cfg, logger)
	logger.Infof("Waiting %d minutes to perform the next update", cfg.UpdateInterval)
	for range ticker.C {
		process.RunProcess(cfg, logger)
		logger.Infof("Waiting %d minutes to perform the next update", cfg.UpdateInterval)
	}
}
