package bot

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"log"
	"time"

	"github.com/afeedhshaji/signal-llm-bot/internal/bot/message"
	"github.com/afeedhshaji/signal-llm-bot/internal/deduper"
	"github.com/afeedhshaji/signal-llm-bot/internal/gemini"
	"github.com/afeedhshaji/signal-llm-bot/internal/signal"
)

type Bot struct {
	SignalClient *signal.SignalClient
	GeminiClient *gemini.Client
	PollInterval time.Duration
	Deduper      *deduper.Deduper
	BotNumber    string
	BotUUID      string
	IgnoreSelf   bool
}

func NewBot(signalClient *signal.SignalClient, geminiClient *gemini.Client, pollInterval time.Duration,
	deduper *deduper.Deduper, botNumber string) *Bot {
	return &Bot{
		SignalClient: signalClient,
		GeminiClient: geminiClient,
		PollInterval: pollInterval,
		Deduper:      deduper,
		BotNumber:    botNumber,
	}
}

// Start begins the bot's polling loop and stops when context is cancelled
func (b *Bot) Start(ctx context.Context) {
	ticker := time.NewTicker(b.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.handleMessages()
		case <-ctx.Done():
			log.Println("bot: context cancelled, stopping polling loop")
			return
		}
	}
}

// handleMessages fetches and processes new messages
func (b *Bot) handleMessages() {
	events, err := b.SignalClient.ReceiveEvents()
	if err != nil {
		log.Printf("Error receiving events: %v", err)
		return
	}

	for _, ev := range events {
		// Deduplication
		evb, _ := json.Marshal(ev)
		hash := sha1.Sum(evb)
		hashStr := hex.EncodeToString(hash[:])
		if b.Deduper.Seen(hashStr) {
			log.Printf("skipping duplicate (hash=%s)", hashStr)
			continue
		}

		msg := message.SimpleExtract(ev, b.BotNumber, b.BotUUID)
		msg.EventHash = hashStr
		msg.RawEvent = ev

		// Only act when bot is mentioned
		if !msg.BotMentioned {
			continue
		}

		log.Printf("Mentioned in %s -> %q", message.TargetLabel(msg), msg.CleanText)

		response, err := b.GeminiClient.Ask(msg.CleanText)
		if err != nil {
			log.Printf("Error generating Gemini response: %v", err)
			continue
		}

		// If groupID is present, get public group ID
		if msg.GroupID != "" {
			publicID, err := b.SignalClient.GetGroupPublicID(msg.GroupID)
			if err != nil {
				log.Printf("Error getting public group ID: %v", err)
				continue
			}
			err = b.SignalClient.SendMessage(publicID, response)
			if err != nil {
				log.Printf("Error sending message to group: %v", err)
			}
			continue
		}
		// TODO: Handle user messags
	}
}
