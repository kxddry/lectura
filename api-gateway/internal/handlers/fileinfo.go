package handlers

import (
	"context"
	"github.com/kxddry/lectura/shared/utils/logger/handlers/sl"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
)

type InfoStorage interface {
	GetFileData(ctx context.Context, uuid string, uid uint) (string, error)
}

func FileInfo(ctx context.Context, log *slog.Logger, st InfoStorage) echo.HandlerFunc {
	const op = "handlers.FileInfo"
	log = log.With("op", op)

	return func(c echo.Context) error {
		uid, ok := c.Get("uid").(uint)
		if !ok {
			log.Info("unauthorized", slog.Int("uid", int(uid)))
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		uuid := c.Param("uuid")
		if uuid == "" {
			log.Debug("empty uuid")
			return c.String(http.StatusBadRequest, "")
		}

		data, err := st.GetFileData(ctx, uuid, uid)
		if err != nil {
			log.Error("error getting file data", sl.Err(err), data)
			return c.String(http.StatusInternalServerError, data)
		}
		return c.String(http.StatusOK, data)
	}
}
