package handlers

import (
	"context"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/kxddry/go-utils/pkg/logger/handlers/sl"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"github.com/kxddry/lectura/uploader/internal/config"
	"github.com/kxddry/lectura/uploader/internal/entities"
	"github.com/kxddry/lectura/uploader/pkg/helpers/converter"
	"github.com/labstack/echo/v4"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
)

type KafkaWriter interface {
	Write(context.Context, uploaded.KafkaRecord) error
}

// Client is the interface for S3.
// Client must be able to upload files to S3 or similar storage systems.
type Client interface {
	Uploader
}

type Uploader interface {
	Upload(ctx context.Context, file entities.File, bucket string) error
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

func UploadHandler(ctx context.Context, log *slog.Logger, w KafkaWriter, cli Client, cfg config.Config) echo.HandlerFunc {
	// logging
	const op = "handlers.uploadHandler"
	log = log.With(slog.String("op", op))

	return func(c echo.Context) error {

		uid := c.Get("uid")
		if uidInt, ok := uid.(int64); !ok || uidInt == 0 {
			return echo.NewHTTPError(http.StatusUnauthorized, "uid missing")
		}

		fileHeader, err := c.FormFile("file")
		// failed to get file
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to get file", err)
		}

		// validate weight
		if fileHeader.Size > maxUploadSize {
			return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "File too big. Max allowed: 1G")
		}

		file, err := fileHeader.Open()
		defer file.Close()
		// failed to open file
		if err != nil {
			log.Error(err.Error())
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to open file", err)
		}

		mtype, err := mimetype.DetectReader(file)

		// failed to detect mimetype
		if err != nil {
			log.Error("failed to detect mimetype", sl.Err(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to detect mimetype", err)
		}

		// check mimetype
		ext, ok := allowedMimeTypes[mtype.String()]
		if !ok {
			return echo.NewHTTPError(http.StatusUnsupportedMediaType, "unsupported media type", mtype.String())
		}

		// go back to the start of the file
		if _, err = file.Seek(0, io.SeekStart); err != nil {
			log.Error("failed to seek file", sl.Err(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error", err)
		}

		filename := fileHeader.Filename
		withoutExt := filename[:len(filename)-len(filepath.Ext(filename))]
		// generated UUIDv4 file name for storage
		fileID := uuid.New().String()
		fc := entities.File{
			UUID:      fileID,
			Extension: ext,
			Data:      file,
			Size:      fileHeader.Size,
			Type:      mtype.String(),
		}

		var wavSent bool

		// Convert file to WAV
		if ext != ".wav" {
			wavFC, err := converter.ConvertToWav(fc)

			if err != nil {
				log.Error("failed to convert file", sl.Err(err))
				return c.String(http.StatusBadRequest, "failed to convert file, your file is broken: "+err.Error())
			}
			defer wavFC.Data.Close()

			err = cli.Upload(ctx, wavFC, cfg.S3Storage.BucketName)

			if err != nil {
				log.Error("failed to upload converted wav file", sl.Err(err))
				return c.String(http.StatusInternalServerError, "Failed uploading converted wav file: "+err.Error())
			}

			wavSent = true
			log.Info("Uploaded converted .wav", slog.String("fileID", fileID))
		}

		if !wavSent && ext != ".wav" {
			return c.String(http.StatusInternalServerError, "Failed to send wav file.")
		}

		if !wavSent {
			err = cli.Upload(ctx, fc, cfg.S3Storage.BucketName)
			if err != nil {
				log.Error("failed to upload original file", sl.Err(err))
				return c.String(http.StatusInternalServerError, "Failed to upload file: "+err.Error())
			}
			log.Info("Uploaded file", slog.String("fileID", fileID))
		}
		out := uploaded.KafkaRecord{
			UUID: fileID,
			Update: struct {
				UserID      int64  `json:"user_id"`
				OGFileName  string `json:"og_file_name"`
				OGExtension string `json:"og_extension"`
				Status      int    `json:"status"`
			}{
				UserID:      uid.(int64),
				OGFileName:  withoutExt,
				OGExtension: ext,
				Status:      0,
			},
		}

		if err := w.Write(ctx, out); err != nil {
			log.Error("failed to send to kafka", sl.Err(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "Internal server error")
		}

		log.Info("message sent to kafka")

		return c.String(http.StatusOK, "uploaded successfully")
	}
}
