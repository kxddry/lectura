package handlers

import (
	"errors"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"strings"
)

func Upload(log *slog.Logger, uploaderURI, cookieName string) echo.HandlerFunc {
	log = log.With(slog.String("op", "handlers.Upload"))
	return func(c echo.Context) error {
		if _, err := c.Cookie(cookieName); err != nil && errors.Is(err, http.ErrNoCookie) {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}

		target := "https://" + strings.TrimRight(uploaderURI, "/") + c.Request().RequestURI
		log.Info("redirect", slog.String("target", target))
		return c.Redirect(http.StatusTemporaryRedirect, target)
	}
}
