package handlers

import (
	"context"
	"errors"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/kxddry/go-utils/pkg/logger/handlers/sl"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"github.com/kxddry/lectura/uploader/internal/entities"
	"github.com/kxddry/lectura/uploader/pkg/helpers/converter"
	"github.com/labstack/echo/v4"
	"gopkg.in/vansante/go-ffprobe.v2"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
)

type KafkaWriter interface {
	Write(context.Context, uploaded.Record) error
}

// Client is the interface for S3.
// Client must be able to upload files to S3 or similar storage systems.
type Client interface {
	Uploader
}

type Uploader interface {
	Upload(ctx context.Context, bucket string, file uploaded.File) error
}

const maxFileDuration = 14400 // 4 hours

func UploadHandler(ctx context.Context, log *slog.Logger, w KafkaWriter, cli Client, bucket, cookieName string) echo.HandlerFunc {
	const op = "handlers.uploadHandler"
	log = log.With(slog.String("op", op))

	return func(c echo.Context) error {

		if _, err := c.Cookie(cookieName); err != nil && errors.Is(err, http.ErrNoCookie) {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		uid := c.Get("uid")
		if uidInt, ok := uid.(uint); !ok || uidInt == 0 {
			return echo.NewHTTPError(http.StatusUnauthorized, "uid missing")
		}

		fileHeader, err := c.FormFile("file")
		// failed to get file
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Failed to get file", err)
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
		ext, ok := uploaded.AllowedMimeTypes[mtype.String()]
		if !ok {
			return echo.NewHTTPError(http.StatusUnsupportedMediaType, "unsupported media type", mtype.String())
		}

		// go back to the start of the file
		if _, err = file.Seek(0, io.SeekStart); err != nil {
			log.Error("failed to seek file", sl.Err(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error", err)
		}

		data, err := ffprobe.ProbeReader(ctx, file)
		if err != nil {
			log.Error("failed to probe", sl.Err(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to probe", err)
		}
		_, _ = file.Seek(0, io.SeekStart)
		dur := data.Format.DurationSeconds

		if dur > maxFileDuration {
			log.Info("audio too long")
			return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "audio too long, max allowed 4 hours")
		}

		filename := fileHeader.Filename
		withoutExt := filename[:len(filename)-len(filepath.Ext(filename))]
		// generated UUIDv4 file name for storage
		fileID := uuid.New().String()
		fc := entities.New(fileID, ext, file, fileHeader.Size, mtype.String())

		var wavSent bool

		// Convert file to WAV
		if ext != ".wav" {
			wavFC, err := converter.ConvertToWav(fc)

			if err != nil {
				log.Error("failed to convert file", sl.Err(err))
				return c.String(http.StatusBadRequest, "failed to convert file, your file is broken: "+err.Error())
			}
			defer wavFC.Close()

			err = cli.Upload(ctx, bucket, wavFC)

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
			err = cli.Upload(ctx, bucket, fc)
			if err != nil {
				log.Error("failed to upload original file", sl.Err(err))
				return c.String(http.StatusInternalServerError, "Failed to upload file: "+err.Error())
			}
			log.Info("Uploaded file", slog.String("fileID", fileID))
		}
		out := uploaded.Record{
			UUID:   fileID,
			Bucket: bucket,
			Update: struct {
				UserID      uint   `json:"user_id"`
				OGFileName  string `json:"og_file_name"`
				OGExtension string `json:"og_extension"`
				Status      int    `json:"status"`
			}{
				UserID:      uid.(uint),
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
