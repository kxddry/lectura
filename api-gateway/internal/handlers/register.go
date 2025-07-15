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
	"strconv"
)

type RegisterClient interface {
	Register(ctx context.Context, email, password string) (int64, error)
	Login(ctx context.Context, email, password string, appId int64) (string, error)
}

func Register(ctx context.Context, log *slog.Logger, auth RegisterClient, appId int64, secure bool, cookieName string) echo.HandlerFunc {
	const op = "handlers.Register"

	log = log.With(slog.String("operation", op))

	return func(c echo.Context) error {
		var req frontend.RegisterRequest

		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}

		if req.Email == "" || req.Password == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "email and password are required")
		}

		uid, err := auth.Register(ctx, req.Email, req.Password)
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.AlreadyExists {
				log.Info("user already exists", slog.String("email", req.Email))
				return echo.NewHTTPError(http.StatusConflict, "email already exists")
			}

			log.Error("failed to register", slog.String("email", req.Email), sl.Err(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}

		// registration successful, log the user in
		tokenStr, err := auth.Login(ctx, req.Email, req.Password, appId)
		if err != nil {
			log.Error("failed to login after registration", sl.Err(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "failed login after registration")
		}

		jwtParsed, err := auth2.ParseJWTUnverified(tokenStr)
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

		return c.JSON(http.StatusCreated, map[string]string{"message": "registered and logged in", "uid": strconv.FormatInt(uid, 10)})
	}
}
