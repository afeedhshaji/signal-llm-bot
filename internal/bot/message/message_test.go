package message

import (
	"testing"
)

func TestSimpleExtract_WithQuote(t *testing.T) {
	botNumber := "+1234567890"
	botUUID := "test-bot-uuid"

	// Simulate a Signal event with a quote/reply
	event := map[string]interface{}{
		"envelope": map[string]interface{}{
			"sourceNumber": "+9876543210",
			"sourceUuid":   "sender-uuid",
			"dataMessage": map[string]interface{}{
				"message": "@bot help me",
				"quote": map[string]interface{}{
					"id":     float64(12345),
					"author": "+1111111111",
					"text":   "What is the weather like?",
				},
				"mentions": []interface{}{
					map[string]interface{}{
						"start":  float64(0),
						"length": float64(4),
						"number": botNumber,
					},
				},
			},
		},
	}

	msg := SimpleExtract(event, botNumber, botUUID)

	if msg.Quote == nil {
		t.Fatal("Expected Quote to be extracted, but got nil")
	}

	if msg.Quote.ID != 12345 {
		t.Errorf("Expected Quote.ID to be 12345, got %d", msg.Quote.ID)
	}

	if msg.Quote.Author != "+1111111111" {
		t.Errorf("Expected Quote.Author to be +1111111111, got %s", msg.Quote.Author)
	}

	if msg.Quote.Text != "What is the weather like?" {
		t.Errorf("Expected Quote.Text to be 'What is the weather like?', got %s", msg.Quote.Text)
	}

	if !msg.BotMentioned {
		t.Error("Expected BotMentioned to be true")
	}
}

func TestSimpleExtract_WithoutQuote(t *testing.T) {
	botNumber := "+1234567890"
	botUUID := "test-bot-uuid"

	// Simulate a Signal event without a quote
	event := map[string]interface{}{
		"envelope": map[string]interface{}{
			"sourceNumber": "+9876543210",
			"dataMessage": map[string]interface{}{
				"message": "@bot hello",
				"mentions": []interface{}{
					map[string]interface{}{
						"start":  float64(0),
						"length": float64(4),
						"number": botNumber,
					},
				},
			},
		},
	}

	msg := SimpleExtract(event, botNumber, botUUID)

	if msg.Quote != nil {
		t.Errorf("Expected Quote to be nil, got %+v", msg.Quote)
	}

	if !msg.BotMentioned {
		t.Error("Expected BotMentioned to be true")
	}
}

func TestSimpleExtract_QuoteWithAuthorUuid(t *testing.T) {
	botNumber := "+1234567890"
	botUUID := "test-bot-uuid"

	// Simulate a Signal event with a quote using authorUuid
	event := map[string]interface{}{
		"envelope": map[string]interface{}{
			"sourceNumber": "+9876543210",
			"dataMessage": map[string]interface{}{
				"message": "@bot respond",
				"quote": map[string]interface{}{
					"id":         float64(12345),
					"authorUuid": "author-uuid-123",
					"text":       "Original message",
				},
				"mentions": []interface{}{
					map[string]interface{}{
						"start":  float64(0),
						"length": float64(4),
						"number": botNumber,
					},
				},
			},
		},
	}

	msg := SimpleExtract(event, botNumber, botUUID)

	if msg.Quote == nil {
		t.Fatal("Expected Quote to be extracted, but got nil")
	}

	if msg.Quote.Author != "author-uuid-123" {
		t.Errorf("Expected Quote.Author to be 'author-uuid-123', got %s", msg.Quote.Author)
	}
}

func TestSimpleExtract_QuoteWithEmptyText(t *testing.T) {
	botNumber := "+1234567890"
	botUUID := "test-bot-uuid"

	// Simulate a Signal event with a quote but empty text (should not set Quote)
	event := map[string]interface{}{
		"envelope": map[string]interface{}{
			"sourceNumber": "+9876543210",
			"dataMessage": map[string]interface{}{
				"message": "@bot test",
				"quote": map[string]interface{}{
					"id":     float64(12345),
					"author": "+1111111111",
					"text":   "",
				},
				"mentions": []interface{}{
					map[string]interface{}{
						"start":  float64(0),
						"length": float64(4),
						"number": botNumber,
					},
				},
			},
		},
	}

	msg := SimpleExtract(event, botNumber, botUUID)

	if msg.Quote != nil {
		t.Errorf("Expected Quote to be nil for empty text, got %+v", msg.Quote)
	}
}
