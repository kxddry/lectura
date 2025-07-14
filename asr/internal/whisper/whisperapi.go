package whisper

import (
	"bytes"
	"encoding/json"
	"github.com/kxddry/lectura/asr/internal/entities"
	"io"
	"mime/multipart"
	"net/http"
)

func CallWhisperAPI(apiUrl string, file io.Reader) (*entities.TranscribeResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(part, file); err != nil {
		return nil, err
	}

	_ = writer.Close()

	req, err := http.NewRequest("POST", apiUrl, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result entities.TranscribeResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
