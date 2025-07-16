package handlers

import (
	"context"
	"errors"
	"github.com/kxddry/lectura/shared/entities/frontend"
	"github.com/kxddry/lectura/shared/entities/uploaded"
	"github.com/kxddry/lectura/shared/utils/logger/handlers/sl"
	"github.com/kxddry/lectura/shared/utils/storage"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"time"
)

type Storage interface {
	ListFiles(ctx context.Context, user_id uint) ([]frontend.File, error)
}

type FileStorage interface {
	GetPresignedURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error)
}

func ListFiles(ctx context.Context, log *slog.Logger, st Storage, fs FileStorage, bucket string, expiry time.Duration) echo.HandlerFunc {
	const op = "handler.ListFiles"
	log = log.With(slog.String("op", op))

	return func(c echo.Context) error {
		uid := c.Get("uid")
		if uid == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}
		if v, ok := uid.(uint); !ok || v == 0 {
			log.Debug("uid is not uint", "uid", uid)
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		out, err := st.ListFiles(ctx, uid.(uint))
		if err != nil {
			if errors.Is(err, storage.ErrNoFiles) {
				return c.JSON(http.StatusOK, []frontend.File{})
			}
			log.Error("list files", sl.Err(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to list files", err)
		}

		for i, file := range out {
			file.URL, err = fs.GetPresignedURL(ctx, bucket, file.UUID+uploaded.AllowedMimeTypes[file.MimeType], expiry)
			if err != nil {
				log.Error("failed to get file URL", "file", file, sl.Err(err))
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to get file URL", err)
			}
			out[i] = file
		}

		return c.JSON(http.StatusOK, out)
	}
}
