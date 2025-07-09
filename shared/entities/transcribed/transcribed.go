package transcribed

type BrokerRecord struct {
	ID       string  `json:"id"`
	TextUrl  string  `json:"text_url"`
	Duration float64 `json:"duration"`
	Language string  `json:"language"`
}

type TranscribeRequest struct {
	ID       string `json:"id"`
	AudioURL string `json:"audio_url"`
}

type TranscribeResponse struct {
	ID       string  `json:"id"`
	Text     string  `json:"text"`
	Duration float64 `json:"duration_sec"`
	Language string  `json:"language"`
}
