package handlers

import (
	"context"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/kxddry/go-utils/pkg/logger/handlers/sl"
	"github.com/labstack/echo/v4"
	"io"
	"log/slog"
	"net/http"
	"time"
	"uploader/internal/broker"
	"uploader/internal/storage"
	"uploader/pkg/helpers/converter"
)

type KafkaWriter interface {
	Write(context.Context, broker.BrokerRecord) error
}

type Client interface {
	LinkGetter
	Uploader
}

type LinkGetter interface {
	GetLink() string
}
type Uploader interface {
	Upload(ctx context.Context, fc storage.FileConfig) (url string, size int64, err error)
}

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

const maxUploadSize = 1 << 30 // 1 GB

func UploadHandler(ctx context.Context, log *slog.Logger, w KafkaWriter, up Client, bucket string) echo.HandlerFunc {

	// logging
	const op = "handlers.uploadHandler"
	log = log.With(slog.String("op", op))

	return func(c echo.Context) error {
		fileHeader, err := c.FormFile("file")
		if err != nil {
			return c.String(http.StatusBadRequest, "Failed to get file: "+err.Error())
		}

		// validate weight
		if fileHeader.Size > maxUploadSize {
			return c.String(http.StatusRequestEntityTooLarge, "File too big. Max allowed: 1G")
		}

		file, err := fileHeader.Open()
		defer file.Close()

		if err != nil {
			log.Error(err.Error())
			return c.String(http.StatusInternalServerError, "Failed to open file: "+err.Error())
		}

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

		fileID := uuid.New().String()
		fc := storage.FileConfig{
			Filename: fileID + ext,
			FileID:   fileID,
			File:     file,
			Size:     fileHeader.Size,
			Bucket:   bucket,
			MType:    mtype.String(),
		}

		s3url, size, err := up.Upload(ctx, fc)
		if err != nil {
			log.Error("failed to upload original file", sl.Err(err))
			return c.String(http.StatusInternalServerError, "Failed to upload file: "+err.Error())
		}
		log.Info("Uploaded file", slog.String("fileID", fileID))

		wavUrl := s3url
		wavSize := fileHeader.Size

		// Convert file to WAV
		if ext != ".wav" {
			wavFC, err := converter.ConvertToWav(fc)
			if err != nil {
				log.Error("failed to convert file", sl.Err(err))
				return c.String(http.StatusInternalServerError, "Failed to convert file: "+err.Error())
			}

			defer wavFC.File.Close()

			wavUrl, wavSize, err = up.Upload(ctx, wavFC)
			if err != nil {
				log.Error("failed to upload converted wav file", sl.Err(err))
				return c.String(http.StatusInternalServerError, "Internal server error. Try sending a WAV file.")
			}
			log.Info("Uploaded converted .wav", slog.String("fileID", fileID))
		}

		out := broker.BrokerRecord{
			FileName:   fileHeader.Filename,
			FileID:     fc.FileID,
			FileType:   fc.MType,
			FileSize:   size,
			WavSize:    wavSize,
			UploadedAt: time.Now().UTC().Unix(),
			S3URL:      s3url,
			WavURL:     wavUrl,
		}

		if err := w.Write(ctx, out); err != nil {
			log.Error("failed to send to kafka", sl.Err(err))
			return c.String(http.StatusInternalServerError, "Internal server error")
		}
		log.Info("message sent to kafka", slog.String("file id", fileID))

		return c.String(http.StatusOK, "uploaded successfully")
	}
}
