package reader

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"log/slog"
)

type KafkaInput struct {
	FileName   string `json:"file_name"`   // the original name
	FileID     string `json:"file_id"`     // the new name without the extension
	FileType   string `json:"file_type"`   // mimetype i think
	FileSize   int64  `json:"file_size"`   // in bytes
	UploadedAt int64  `json:"uploaded_at"` // unix time int64
	S3URL      string `json:"s3_url"`      // self-explanatory
	// UserID     string `json:"user_id"` TODO: implement auth
	// Metadata map[string]string `json:"metadata,omitempty"`
}

var possibleMimeTypes = map[string]string{
	"video/mp4":       ".mp4",
	"video/quicktime": ".mov",
	"video/x-msvideo": ".avi",
	"audio/aac":       ".aac",
	"audio/wav":       ".wav",
	"audio/ogg":       ".ogg",
	"audio/mpeg":      ".mpeg",
	"audio/mp4":       ".mp3",
}

func ProcessInput(ctx context.Context, log *slog.Logger, r *kafka.Reader) {
	for {
		select {
		case <-ctx.Done():
			log.Info("worker stopped")
			return
		default:
			m, err := r.ReadMessage(ctx)
			if err != nil {
				log.Error(err.Error())
				continue
			}

			// TODO: Implement auth and change the key to the userid
			// uid := m.Key

			var input KafkaInput
			if err = json.Unmarshal(m.Value, &input); err != nil {
				// TODO: add DLQ
				log.Error(err.Error())
				continue
			}

		}
	}
}
