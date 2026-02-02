package message

import (
	"regexp"
	"strings"

	"github.com/afeedhshaji/signal-llm-bot/internal/signal"
)

type Message struct {
	SourceNumber string
	SourceUUID   string
	GroupID      string
	RawText      string
	CleanText    string
	Mentions     []signal.Mention
	BotMentioned bool
	Quote        *signal.Quote
	EventHash    string
	RawEvent     *signal.Envelope
}

// SimpleExtract extracts message information from a signal envelope
func SimpleExtract(envelope *signal.Envelope, botNumber, botUUID string) Message {
	var m Message

	m.SourceNumber = envelope.SourceNumber
	m.SourceUUID = envelope.SourceUUID
	if envelope.DataMessage != nil {
		dm := envelope.DataMessage
		m.RawText = dm.Message
		m.CleanText = strings.TrimSpace(dm.Message)
		if dm.GroupInfo != nil {
			m.GroupID = dm.GroupInfo.GroupID
		}
		if len(dm.Mentions) > 0 {
			m.Mentions = dm.Mentions
			m.CleanText = RemoveMentionsFromText(m.RawText, m.Mentions)
			for _, men := range m.Mentions {
				if men.Number != "" && NormalizePhone(men.Number) == NormalizePhone(botNumber) {
					m.BotMentioned = true
					break
				}
				if men.UUID != "" && botUUID != "" && men.UUID == botUUID {
					m.BotMentioned = true
					break
				}
			}
		} else {
			if strings.Contains(strings.ToLower(m.CleanText), strings.ToLower(botNumber)) {
				m.BotMentioned = true
				m.CleanText = strings.ReplaceAll(m.CleanText, botNumber, "")
				m.CleanText = strings.TrimSpace(m.CleanText)
			}
		}
		if dm.Quote != nil && dm.Quote.Text != "" {
			q := &signal.Quote{
				ID:     dm.Quote.ID,
				Author: dm.Quote.Author,
				Text:   dm.Quote.Text,
			}
			if q.Author == "" && dm.Quote.AuthorUUID != "" {
				q.Author = dm.Quote.AuthorUUID
			}
			m.Quote = q
		}
	}
	return m
}

// RemoveMentionsFromText removes mentions from the message text
func RemoveMentionsFromText(s string, mentions []signal.Mention) string {
	if s == "" || len(mentions) == 0 {
		return strings.TrimSpace(s)
	}
	runes := []rune(s)
	for i := 0; i < len(mentions); i++ {
		for j := i; j > 0 && mentions[j-1].Start < mentions[j].Start; j-- {
			mentions[j], mentions[j-1] = mentions[j-1], mentions[j]
		}
	}
	for _, mm := range mentions {
		start := mm.Start
		if start < 0 {
			start = 0
		}
		end := start + mm.Length
		if start >= len(runes) {
			continue
		}
		if end > len(runes) {
			end = len(runes)
		}
		runes = append(runes[:start], runes[end:]...)
	}
	out := strings.TrimSpace(string(runes))
	out = regexp.MustCompile(`\s+`).ReplaceAllString(out, " ")
	return out
}

// LooksLikePhone checks if a string looks like a phone number
func LooksLikePhone(s string) bool {
	clean := strings.ReplaceAll(strings.TrimSpace(s), " ", "")
	return regexp.MustCompile(`^\+?\d+$`).MatchString(clean)
}

// NormalizePhone normalizes a phone number by removing spaces
func NormalizePhone(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), " ", "")
}

// TargetLabel returns a label for the message target
func TargetLabel(m Message) string {
	if m.GroupID != "" {
		return "group " + m.GroupID
	}
	if m.SourceNumber != "" {
		return "user " + m.SourceNumber
	}
	if m.SourceUUID != "" {
		return "user-uuid " + m.SourceUUID
	}
	return "unknown"
}
