package handlers

import (
	"context"
	"encoding/json"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/kxddry/go-utils/pkg/logger/handlers/sl"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"github.com/segmentio/kafka-go"
	"io"
	"log/slog"
	"net/http"
	"time"
)

var allowedMimeTypes = map[string]string{
	"video/mp4":       ".mp4",
	"video/quicktime": ".mov",
	"video/x-msvideo": ".avi",
	"audio/aac":       ".aac",
	"audio/wav":       ".wav",
	"audio/ogg":       ".ogg",
	"audio/mpeg":      ".mpeg",
	"audio/mp4":       ".mp3",
}

type KafkaInput struct {
	FileName   string `json:"file_name"`
	FileID     string `json:"file_id"`
	FileType   string `json:"file_type"`
	FileSize   int64  `json:"file_size"`
	UploadedAt int64  `json:"uploaded_at"`
	// UserID     string `json:"user_id"` TODO: implement auth
	// Metadata map[string]string `json:"metadata,omitempty"`
}

const maxUploadSize = 1 << 30 // 1 GB

func UploadHandler(ctx context.Context, log *slog.Logger, w *kafka.Writer, mc *minio.Client, bucket string) echo.HandlerFunc {
	const op = "handlers.uploadHandler"
	log = log.With(slog.String("op", op))
	return func(c echo.Context) error {
		fileHeader, err := c.FormFile("file")

		if err != nil {
			return c.String(http.StatusBadRequest, "Failed to get file: "+err.Error())
		}

		if fileHeader.Size > maxUploadSize {
			return c.String(http.StatusRequestEntityTooLarge, "File too big. Max allowed: 2G")
		}

		file, err := fileHeader.Open()
		if err != nil {
			log.Error(err.Error())
			return c.String(http.StatusInternalServerError, "Failed to open file: "+err.Error())
		}

		defer file.Close()

		mtype, err := mimetype.DetectReader(file)

		if err != nil {
			log.Error("failed to detect mimetype", sl.Err(err))
			return c.String(http.StatusInternalServerError, "Failed to detect mimetype: "+err.Error())
		}

		ext, ok := allowedMimeTypes[mtype.String()]
		if !ok {
			return c.String(http.StatusUnsupportedMediaType, "Unsupported media type: "+mtype.String())
		}

		if _, err = file.Seek(0, io.SeekStart); err != nil {
			log.Error("failed to seek file", sl.Err(err))
			return c.String(http.StatusInternalServerError, "Internal server error")
		}

		fileID := uuid.New().String() + ext

		info, err := mc.PutObject(ctx, bucket, fileID, file, fileHeader.Size, minio.PutObjectOptions{ContentType: mtype.String()})
		if err != nil {
			log.Error("failed to upload file", sl.Err(err))
			return c.String(http.StatusInternalServerError, "Failed to upload file.")
		}
		log.Info("Uploaded file", slog.String("file name", fileID), slog.Int64("weight in bytes", info.Size))

		go func() {
			out := KafkaInput{
				FileName:   fileHeader.Filename,
				FileID:     fileID,
				FileType:   mtype.String(),
				FileSize:   info.Size,
				UploadedAt: time.Now().UTC().Unix(),
			}

			msgBytes, _ := json.Marshal(out)

			msg := kafka.Message{
				Key:   []byte(fileHeader.Filename),
				Value: msgBytes,
				Time:  time.Now().UTC(),
			}

			if err := w.WriteMessages(ctx, msg); err != nil {
				// TODO: add some retry logic here
				log.Error("failed to write message", sl.Err(err))
			} else {
				log.Info("message sent to Kafka", slog.String("file name", fileHeader.Filename), slog.Int64("weight in bytes", info.Size))
			}
		}()

		return c.String(http.StatusOK, "uploaded successfully")
	}

}
