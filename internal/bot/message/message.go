package message

import (
	"regexp"
	"strings"
)

type Mention struct {
	Start  int
	Length int
	Number string
	UUID   string
}

type Message struct {
	SourceNumber string
	SourceUUID   string
	GroupID      string
	RawText      string
	CleanText    string
	Mentions     []Mention
	BotMentioned bool
	EventHash    string
	RawEvent     map[string]interface{}
}

func SimpleExtract(ev map[string]interface{}, botNumber, botUUID string) Message {
	var m Message
	env, _ := ev["envelope"].(map[string]interface{})
	if env == nil {
		if dm, ok := ev["dataMessage"].(map[string]interface{}); ok {
			if msg, ok := dm["message"].(string); ok {
				m.RawText = msg
				m.CleanText = strings.TrimSpace(msg)
			}
		}
		return m
	}
	if sn, ok := env["sourceNumber"].(string); ok && LooksLikePhone(sn) {
		m.SourceNumber = sn
	}
	if src, ok := env["source"].(string); ok {
		if LooksLikePhone(src) && m.SourceNumber == "" {
			m.SourceNumber = src
		} else if !LooksLikePhone(src) {
			m.SourceUUID = src
		}
	}
	if su, ok := env["sourceUuid"].(string); ok && m.SourceUUID == "" {
		m.SourceUUID = su
	}
	if dm, ok := env["dataMessage"].(map[string]interface{}); ok {
		if gi, ok := dm["groupInfo"].(map[string]interface{}); ok {
			if gid, ok := gi["groupId"].(string); ok {
				m.GroupID = gid
			}
		}
		if raw, ok := dm["message"].(string); ok && raw != "<nil>" {
			m.RawText = raw
			if mr, ok := dm["mentions"].([]interface{}); ok && len(mr) > 0 {
				for _, it := range mr {
					if mm, ok := it.(map[string]interface{}); ok {
						men := Mention{}
						if s, ok := mm["start"].(float64); ok {
							men.Start = int(s)
						}
						if l, ok := mm["length"].(float64); ok {
							men.Length = int(l)
						}
						if num, ok := mm["number"].(string); ok {
							men.Number = num
						}
						if uuid, ok := mm["uuid"].(string); ok {
							men.UUID = uuid
						}
						m.Mentions = append(m.Mentions, men)
					}
				}
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
				m.CleanText = strings.TrimSpace(m.RawText)
				if strings.Contains(strings.ToLower(m.CleanText), strings.ToLower(botNumber)) {
					m.BotMentioned = true
					m.CleanText = strings.ReplaceAll(m.CleanText, botNumber, "")
					m.CleanText = strings.TrimSpace(m.CleanText)
				}
			}
		}
	}
	return m
}

func RemoveMentionsFromText(s string, mentions []Mention) string {
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

func LooksLikePhone(s string) bool {
	clean := strings.ReplaceAll(strings.TrimSpace(s), " ", "")
	return regexp.MustCompile(`^\+?\d+$`).MatchString(clean)
}

func NormalizePhone(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), " ", "")
}

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
