package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SignalAPIURL      string
	SignalNumber      string
	BotName           string
	PollInterval      string
	GoogleAPIKey      string
	GeminiModel       string
	GeminiTimeout     string
	SystemPrompt      string
	OpenRouterAPIKey  string
	OpenRouterModel   string
	OpenRouterTimeout string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
		return nil, err
	}

	return &Config{
		SignalAPIURL:      getEnv("SIGNAL_API_URL", "http://localhost:8089"),
		SignalNumber:      getEnv("SIGNAL_NUMBER", ""),
		BotName:           getEnv("BOT_NAME", ""),
		PollInterval:      getEnv("POLL_INTERVAL", "5s"),
		GoogleAPIKey:      getEnv("GOOGLE_API_KEY", ""),
		GeminiModel:       getEnv("GEMINI_MODEL", "gemini-2.0-flash"),
		GeminiTimeout:     getEnv("GEMINI_TIMEOUT", "120s"),
		SystemPrompt:      getEnv("SYSTEM_PROMPT", "You are a helpful assistant."),
		OpenRouterAPIKey:  getEnv("OPENROUTER_API_KEY", ""),
		OpenRouterModel:   getEnv("OPENROUTER_MODEL", "xiaomi/mimo-v2-flash:free"),
		OpenRouterTimeout: getEnv("OPENROUTER_TIMEOUT", "120s"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
