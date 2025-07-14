package entities

type TranscribeResponse struct {
	Text     string `json:"text"`
	Language string `json:"language"`
}
