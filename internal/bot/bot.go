package bot

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/afeedhshaji/signal-llm-bot/internal/bot/message"
	"github.com/afeedhshaji/signal-llm-bot/internal/deduper"
	"github.com/afeedhshaji/signal-llm-bot/internal/igdownloader"
	"github.com/afeedhshaji/signal-llm-bot/internal/signal"
)

type LLM interface {
	Ask(prompt string) (string, error)
}

type Bot struct {
	SignalClient *signal.SignalClient
	GeminiClient LLM
	PollInterval time.Duration
	Deduper      *deduper.Deduper
	BotNumber    string
	BotUUID      string
	IgnoreSelf   bool
}

func NewBot(signalClient *signal.SignalClient, llm LLM, pollInterval time.Duration,
	deduper *deduper.Deduper, botNumber string) *Bot {
	return &Bot{
		SignalClient: signalClient,
		GeminiClient: llm,
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
		evb, _ := json.Marshal(ev)
		hash := sha1.Sum(evb)
		hashStr := hex.EncodeToString(hash[:])
		if b.Deduper.Seen(hashStr) {
			log.Printf("skipping duplicate (hash=%s)", hashStr)
			continue
		}

		msg := message.SimpleExtract(&ev, b.BotNumber, b.BotUUID)
		msg.EventHash = hashStr
		msg.RawEvent = &ev

		if !msg.BotMentioned {
			continue
		}

		log.Printf("Mentioned in %s -> %q", message.TargetLabel(msg), msg.CleanText)

		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(msg.CleanText)), "/help") {
			b.handleHelpCommand(msg)
			return
		}

		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(msg.CleanText)), "/download") {
			var instagramURL string

			commandText := strings.TrimSpace(msg.CleanText[9:])
			if commandText != "" {
				instagramURL = igdownloader.ExtractInstagramURL(commandText)
			}

			if instagramURL == "" {
				instagramURL = igdownloader.ExtractInstagramURL(msg.RawText)
			}

			if instagramURL == "" && msg.Quote != nil && msg.Quote.Text != "" {
				instagramURL = igdownloader.ExtractInstagramURL(msg.Quote.Text)
			}

			if instagramURL != "" {
				b.handleInstagramDownload(msg, instagramURL)
				return
			}

			usage := "To download an Instagram video:\\n• Reply to a message containing an Instagram URL with '@bot /download'\\n• Or use '@bot /download <instagram_url>'"
			b.sendResponse(msg, usage)
			return
		}

		prompt := msg.CleanText
		if msg.Quote != nil && msg.Quote.Text != "" {
			prompt = "Context (replying to): \"" + msg.Quote.Text + "\"\n\nUser message: " + msg.CleanText
			log.Printf("Including reply context from %s: %q", msg.Quote.Author, msg.Quote.Text)
		}

		response, err := b.GeminiClient.Ask(prompt)
		if err != nil {
			log.Printf("Error generating LLM response: %v", err)
			b.sendErrorResponse(msg)
			return
		}

		b.sendResponse(msg, response)
	}
}

// sendResponse sends a text response to the appropriate chat
func (b *Bot) sendResponse(msg message.Message, response string) {
	if msg.GroupID != "" {
		publicID, err := b.SignalClient.GetGroupPublicID(msg.GroupID)
		if err != nil {
			log.Printf("Error getting public group ID: %v", err)
			return
		}
		err = b.SignalClient.SendMessage(publicID, response)
		if err != nil {
			log.Printf("Error sending message to group: %v", err)
		}
	} else if msg.SourceNumber != "" {
		if err := b.SignalClient.SendMessage(msg.SourceNumber, response); err != nil {
			log.Printf("Error sending message to user: %v", err)
		}
	} else if msg.SourceUUID != "" {
		if err := b.SignalClient.SendMessage(msg.SourceUUID, response); err != nil {
			log.Printf("Error sending message to user-uuid: %v", err)
		}
	}
}

// sendErrorResponse sends a generic error message to the chat
func (b *Bot) sendErrorResponse(msg message.Message) {
	generic := "An error occurred while processing your request. Please try again later."
	b.sendResponse(msg, generic)
}

// sendFile sends a file to the appropriate chat
func (b *Bot) sendFile(msg message.Message, filePath, caption string) {
	if msg.GroupID != "" {
		publicID, err := b.SignalClient.GetGroupPublicID(msg.GroupID)
		if err != nil {
			log.Printf("Error getting public group ID: %v", err)
			return
		}
		err = b.SignalClient.SendFile(publicID, filePath, caption)
		if err != nil {
			log.Printf("Error sending file to group: %v", err)
		}
	} else if msg.SourceNumber != "" {
		if err := b.SignalClient.SendFile(msg.SourceNumber, filePath, caption); err != nil {
			log.Printf("Error sending file to user: %v", err)
		}
	} else if msg.SourceUUID != "" {
		if err := b.SignalClient.SendFile(msg.SourceUUID, filePath, caption); err != nil {
			log.Printf("Error sending file to user-uuid: %v", err)
		}
	}
}

// handleInstagramDownload processes Instagram video download requests
func (b *Bot) handleInstagramDownload(msg message.Message, instagramURL string) {
	log.Printf("Processing Instagram download request for: %s", instagramURL)

	b.sendResponse(msg, "Downloading Instagram video... This may take a moment.")

	result := igdownloader.DownloadInstagramVideo(instagramURL)

	if !result.Success {
		log.Printf("Instagram download failed: %v", result.Error)
		b.sendResponse(msg, "Failed to download Instagram video. Please check the URL and try again.")
		return
	}

	b.sendFile(msg, result.VideoFile, "")
}

// handleHelpCommand sends a help message with all available commands
func (b *Bot) handleHelpCommand(msg message.Message) {
	helpText := `🤖 *Signal Bot Commands*

*Available Commands:*
• /download - Download an Instagram video
  • Reply to a message containing an Instagram URL with '@bot /download'
  • Or use '@bot /download <instagram_url>'

• /help - Show this help message

*General Usage:*
• Mention @bot in any message to chat with the AI
• The bot responds to your questions and conversations
• When you reply to a message, the bot includes that context in its response
`
	b.sendResponse(msg, helpText)
}
