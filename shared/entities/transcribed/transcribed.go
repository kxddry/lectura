package transcribed

type BrokerRecord struct {
	// Extension is always .txt
	UserID   int64  `json:"user_id"`
	TextName string `json:"text_name"`
	TextID   string `json:"text_id"`
	// FileType is always text/plain
	TextSize int64  `json:"text_size"`
	Language string `json:"language"`
}

type TranscribeResponse struct {
	Text     string `json:"text"`
	Language string `json:"language"`
}

type TextRecord struct {
	UserID   int64
	Language string
	TextID   string
}
