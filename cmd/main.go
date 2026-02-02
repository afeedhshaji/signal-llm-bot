package main

import (
	"context"
	"log"
	"os"
	sigs "os/signal"
	"syscall"
	"time"

	"github.com/afeedhshaji/signal-llm-bot/config"
	"github.com/afeedhshaji/signal-llm-bot/internal/bot"
	"github.com/afeedhshaji/signal-llm-bot/pkg/deduper"
	"github.com/afeedhshaji/signal-llm-bot/pkg/openrouter"
	signalapi "github.com/afeedhshaji/signal-llm-bot/internal/signal"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}
	log.Printf("Loaded config: %+v", cfg)

	pollInterval, err := time.ParseDuration(cfg.PollInterval)
	if err != nil {
		log.Fatalf("Invalid poll interval: %v", err)
	}
	// parse OpenRouter timeout (fallback to Gemini timeout string if desired)
	openrouterTimeout, err := time.ParseDuration(cfg.OpenRouterTimeout)
	if err != nil {
		log.Fatalf("Invalid OpenRouter timeout: %v", err)
	}
	deduperTTL := 30 * time.Second
	dedup := deduper.New(deduperTTL)

	signalClient := signalapi.NewSignalClient(cfg.SignalAPIURL, cfg.SignalNumber)
	// Build OpenRouter client and wire it into the bot as the LLM
	openrouterEndpoint := "https://openrouter.ai/api/v1/chat/completions"
	openrouterClient := openrouter.New(cfg.OpenRouterAPIKey, openrouterEndpoint, cfg.OpenRouterModel, openrouterTimeout, cfg.SystemPrompt)

	botInstance := bot.NewBot(
		signalClient,
		openrouterClient,
		pollInterval,
		dedup,
		cfg.SignalNumber,
	)

	// Graceful shutdown with context
	ctx, cancel := context.WithCancel(context.Background())
	stop := make(chan os.Signal, 1)
	sigs.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{})

	go func() {
		botInstance.Start(ctx)
		close(done)
	}()

	<-stop
	log.Println("shutdown requested")
	cancel()
	dedup.Stop()
	<-done
	log.Println("exited")
}
