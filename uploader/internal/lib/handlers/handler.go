package handlers

import (
	"context"
	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"io"
	"log/slog"
	"net/http"
	"uploader/internal/lib/logger/handlers/sl"
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

const maxUploadSize = 1 << 31 // 2 GB

func UploadHandler(ctx context.Context, log *slog.Logger, mc *minio.Client, bucket string) echo.HandlerFunc {
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
		return c.String(http.StatusOK, "uploaded successfully")
	}

}
