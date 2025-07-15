package handlers

import (
	"context"
	"crypto/ed25519"
	auth2 "github.com/kxddry/lectura/shared/entities/auth"
	"github.com/kxddry/lectura/shared/utils/logger/handlers/sl"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"net/http"
	"strings"
)

type LoginClient interface {
	Login(ctx context.Context, email, pass string, appId int64) (string, error)
}

func Login(ctx context.Context, log *slog.Logger, auth LoginClient, appId int64, secure bool, cookieName string, authPubkey ed25519.PublicKey) echo.HandlerFunc {
	const op = "handlers.Login"
	log = log.With("op", op)

	return func(c echo.Context) error {
		uid := c.Get("uid")
		if v, ok := uid.(uint); ok && v != 0 {
			return c.Redirect(http.StatusFound, "/")
		}

		email := c.FormValue("email")
		password := c.FormValue("password")

		if email == "" || password == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "email or password is empty")
		}

		tokenStr, err := auth.Login(ctx, email, password, appId)
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.Unauthenticated {
				log.Info("unauthenticated login", sl.Err(err))
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
			}
			if ok && st.Code() == codes.InvalidArgument {
				log.Info("invalid credentials", sl.Err(err))
				return echo.NewHTTPError(http.StatusBadRequest, getDesc(err))
			}

			log.Error("failed to login", sl.Err(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}

		jwt, err := auth2.ParseJWT(tokenStr, authPubkey)
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

func getDesc(err error) string {
	if err == nil {
		return ""
	}
	str := err.Error()
	a := strings.Split(str, "desc = ")
	return a[len(a)-1]
}
