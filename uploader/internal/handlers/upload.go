package handlers

import (
	"context"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/kxddry/go-utils/pkg/logger/handlers/sl"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"github.com/kxddry/lectura/uploader/pkg/helpers/converter"
	"github.com/labstack/echo/v4"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"
)

type KafkaWriter interface {
	Write(context.Context, uploaded.BrokerRecord) error
}

// Client is the interface for S3.
// Client must be able to upload files to S3 or similar storage systems.
type Client interface {
	Uploader
}

type Uploader interface {
	Upload(ctx context.Context, fc uploaded.FileConfig) error
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
		// failed to get file
		if err != nil {
			return c.String(http.StatusBadRequest, "Failed to get file: "+err.Error())
		}

		// validate weight
		if fileHeader.Size > maxUploadSize {
			return c.String(http.StatusRequestEntityTooLarge, "File too big. Max allowed: 1G")
		}

		file, err := fileHeader.Open()
		defer file.Close()
		// failed to open file
		if err != nil {
			log.Error(err.Error())
			return c.String(http.StatusInternalServerError, "Failed to open file: "+err.Error())
		}

		mtype, err := mimetype.DetectReader(file)

		// failed to detect mimetype
		if err != nil {
			log.Error("failed to detect mimetype", sl.Err(err))
			return c.String(http.StatusInternalServerError, "Failed to detect mimetype: "+err.Error())
		}

		// check mimetype
		ext, ok := allowedMimeTypes[mtype.String()]
		if !ok {
			return c.String(http.StatusUnsupportedMediaType, "Unsupported media type: "+mtype.String())
		}

		// go back to the start of the file
		if _, err = file.Seek(0, io.SeekStart); err != nil {
			log.Error("failed to seek file", sl.Err(err))
			return c.String(http.StatusInternalServerError, "Internal server error")
		}

		filename := fileHeader.Filename
		woExt := filename[:len(filename)-len(filepath.Ext(filename))]
		// generated UUIDv4 file name for storage
		fileID := uuid.New().String()
		fc := uploaded.FileConfig{
			Extension: ext,
			FileName:  woExt, // no extension!
			FileID:    fileID,
			File:      file,
			FileSize:  fileHeader.Size,
			Bucket:    bucket,
			FileType:  mtype.String(),
		}

		wavSize := fileHeader.Size
		var wavSent bool

		// Convert file to WAV
		if ext != ".wav" {
			wavFC, err := converter.ConvertToWav(fc)

			if err != nil {
				log.Error("failed to convert file", sl.Err(err))
				return c.String(http.StatusBadRequest, "failed to convert file, your file is broken: "+err.Error())
			}

			defer wavFC.File.Close()

			err = up.Upload(ctx, wavFC)
			if err != nil {
				log.Error("failed to upload converted wav file", sl.Err(err))
				return c.String(http.StatusInternalServerError, "Failed uploading converted wav file: "+err.Error())
			}

			wavSent = true
			wavSize = wavFC.FileSize
			log.Info("Uploaded converted .wav", slog.String("fileID", fileID))
		}

		// uploading here to guarantee .wav file uploaded
		if !wavSent {
			err = up.Upload(ctx, fc)
			if err != nil {
				log.Error("failed to upload original file", sl.Err(err))
				return c.String(http.StatusInternalServerError, "Failed to upload file: "+err.Error())
			}
			log.Info("Uploaded file", slog.String("fileID", fileID))
		}
		out := uploaded.BrokerRecord{
			Extension:  fc.Extension,
			FileName:   fc.FileName,
			FileID:     fc.FileID,
			FileType:   fc.FileType,
			FileSize:   fc.FileSize,
			WavSize:    wavSize,
			UploadedAt: time.Now().UTC().Unix(),
		}

		if err := w.Write(ctx, out); err != nil {
			log.Error("failed to send to kafka", sl.Err(err))
			return c.String(http.StatusInternalServerError, "Internal server error")
		}

		log.Info("message sent to kafka", slog.String("file name", out.FileName),
			slog.String("file id", out.FileID),
			slog.String("file type", out.FileType),
		)

		return c.String(http.StatusOK, "uploaded successfully")
	}
}
