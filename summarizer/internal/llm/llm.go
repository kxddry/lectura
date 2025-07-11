package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kxddry/lectura/summarizer/internal/config"
	"github.com/kxddry/lectura/summarizer/internal/entities"
	"net/http"
)

type OpenAI struct {
	Cfg *config.Config
}

func (o OpenAI) SendMessage(msg []byte) (entities.ChatResponse, error) {
	if o.Cfg == nil {
		return entities.ChatResponse{}, errors.New("config is nil")
	}
	const op = "llm.SendMessage"
	reqBody := entities.ChatRequest{
		Model: o.Cfg.Summarizer.Model,
		Messages: []entities.ChatMessage{
			{Role: "system", Content: []byte(o.Cfg.Summarizer.Prompt)},
			{Role: "user", Content: msg},
		},
		Stream: false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return entities.ChatResponse{}, fmt.Errorf("%s error encoding: %w", op, err)
	}

	req, err := http.NewRequest("POST", o.Cfg.Summarizer.BaseUrl, bytes.NewBuffer(body))
	if err != nil {
		return entities.ChatResponse{}, fmt.Errorf("%s error creating request: %w", op, err)
	}

	req.Header.Set("Content-Type", "application/json")

	if o.Cfg.Summarizer.ApiKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.Cfg.Summarizer.ApiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return entities.ChatResponse{}, fmt.Errorf("%s request failed: %w", op, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return entities.ChatResponse{}, fmt.Errorf("%s: request failed with status code %d", op, resp.StatusCode)
	}

	var chatResp entities.ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return entities.ChatResponse{}, fmt.Errorf("%s error decoding response: %w", op, err)
	}
	if len(chatResp.Choices) == 0 {
		return entities.ChatResponse{}, fmt.Errorf("%s: empty response", op)
	}
	return chatResp, nil
}
