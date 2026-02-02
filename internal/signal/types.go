package signal

// EnvelopeWrapper matches the Signal REST API response structure
type EnvelopeWrapper struct {
	Envelope Envelope `json:"envelope"`
	Account  string   `json:"account"`
}

type Envelope struct {
	SourceNumber string       `json:"sourceNumber"`
	Source       string       `json:"source"`
	SourceUUID   string       `json:"sourceUuid"`
	Timestamp    int64        `json:"timestamp"`
	DataMessage  *DataMessage `json:"dataMessage"`
}

type DataMessage struct {
	Message   string     `json:"message"`
	Mentions  []Mention  `json:"mentions"`
	GroupInfo *GroupInfo `json:"groupInfo"`
	Quote     *Quote     `json:"quote"`
}

type GroupInfo struct {
	GroupID string `json:"groupId"`
}

type Mention struct {
	Start  int    `json:"start"`
	Length int    `json:"length"`
	Number string `json:"number"`
	UUID   string `json:"uuid"`
}

type Quote struct {
	ID         int64  `json:"id"`
	Author     string `json:"author"`
	AuthorUUID string `json:"authorUuid"`
	Text       string `json:"text"`
}

// QuoteRequest represents a quote to include when sending a message
type QuoteRequest struct {
	ID     int64  `json:"id"`
	Author string `json:"author"`
	Text   string `json:"text"`
}
