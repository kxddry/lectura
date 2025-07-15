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
)

type RegisterClient interface {
	Register(ctx context.Context, email, password string) (int64, error)
	Login(ctx context.Context, email, password string, appId int64) (string, error)
}

func Register(ctx context.Context, log *slog.Logger, auth RegisterClient, appId int64, secure bool, cookieName string, authPubKey ed25519.PublicKey) echo.HandlerFunc {
	const op = "handlers.Register"

	log = log.With(slog.String("operation", op))

	return func(c echo.Context) error {
		email := c.FormValue("email")
		password := c.FormValue("password")

		if email == "" || password == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "email and password are required")
		}

		_, err := auth.Register(ctx, email, password)
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.AlreadyExists {
				log.Info("user already exists", slog.String("email", email))
				return echo.NewHTTPError(http.StatusConflict, "email already exists")
			}

			if ok && st.Code() == codes.InvalidArgument {
				log.Info("invalid credentials", sl.Err(err))
				return echo.NewHTTPError(http.StatusBadRequest, getDesc(err))
			}

			log.Error("failed to register", slog.String("email", email), sl.Err(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}

		// registration successful, log the user in
		tokenStr, err := auth.Login(ctx, email, password, appId)
		if err != nil {
			log.Error("failed to login after registration", sl.Err(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "failed login after registration")
		}

		jwtParsed, err := auth2.ParseJWT(tokenStr, authPubKey)
		if err != nil {
			log.Error("failed to parse token after register", sl.Err(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to parse token after register")
		}

		c.SetCookie(&http.Cookie{
			Name:     cookieName,
			Value:    tokenStr,
			Expires:  jwtParsed.Exp,
			Path:     "/",
			Secure:   secure,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		return c.Redirect(http.StatusFound, "/")
	}
}
