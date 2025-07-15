package handlers

import (
	"context"
	auth2 "github.com/kxddry/lectura/shared/entities/auth"
	"github.com/kxddry/lectura/shared/entities/frontend"
	"github.com/kxddry/lectura/shared/utils/logger/handlers/sl"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"net/http"
)

type LoginClient interface {
	Login(ctx context.Context, email, pass string, appId int64) (string, error)
}

func Login(ctx context.Context, log *slog.Logger, auth LoginClient, appId int64, secure bool, cookieName string) echo.HandlerFunc {
	const op = "handlers.Login"
	log = log.With("op", op)

	return func(c echo.Context) error {

		uid := c.Get("uid")
		if v, ok := uid.(string); ok && v != "" {
			return c.Redirect(http.StatusFound, "/")
		}

		var req frontend.LoginRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if req.Email == "" || req.Password == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "email or password is empty")
		}

		tokenStr, err := auth.Login(ctx, req.Email, req.Password, appId)
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.Unauthenticated {
				log.Info("unauthenticated login", sl.Err(err))
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
			}

			log.Error("failed to login", sl.Err(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}

		jwt, err := auth2.ParseJWTUnverified(tokenStr)
		if err != nil {
			log.Error("failed to parse token", sl.Err(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "invalid token")
		}

		c.SetCookie(&http.Cookie{
			Name:     cookieName,
			Value:    tokenStr,
			Expires:  jwt.Exp,
			Path:     "/",
			Secure:   secure,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		return c.Redirect(http.StatusFound, "/")
	}
}
