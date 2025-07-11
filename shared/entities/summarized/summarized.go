package summarized

// BrokerRecord contains the summarized text. It contains the output produced by the summarizer package.
type BrokerRecord struct {
	// Extension is always .txt
	TextName string `json:"text_name"`
	TextID   string `json:"text_id"`
	// FileType is always text/plain
	TextSize int64  `json:"text_size"`
	Language string `json:"language"`
}
