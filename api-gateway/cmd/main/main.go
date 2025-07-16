package main

import (
	"context"
	"github.com/kxddry/lectura/api-gateway/internal/config"
	"github.com/kxddry/lectura/api-gateway/internal/handlers"
	"github.com/kxddry/lectura/shared/clients/sso/grpc"
	config2 "github.com/kxddry/lectura/shared/utils/config"
	"github.com/kxddry/lectura/shared/utils/ed25519"
	"github.com/kxddry/lectura/shared/utils/logger"
	middleware2 "github.com/kxddry/lectura/shared/utils/middleware"
	"github.com/kxddry/lectura/shared/utils/s3"
	"github.com/kxddry/lectura/shared/utils/storage/postgres"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const cookieName = "access_token"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var cfg config.Config
	config2.MustParseConfig(&cfg)

	log := logger.SetupLogger(cfg.Env)

	pubKey, err := ed25519.LoadPublicKey(cfg.PubkeyPath)
	if err != nil {
		log.Error("Failed to load public key", "err", err)
		os.Exit(1)
	}

	bucket := os.Getenv("BUCKET")
	if bucket == "" {
		log.Warn("Bucket env variable not set, setting BUCKET=input")
		bucket = "input"
	}

	sql, err := postgres.New(cfg.Storage)
	if err != nil {
		panic(err)
	}

	cli, err := s3.NewClient(cfg.S3Storage)
	if err != nil {
		panic(err)
	}

	pubKeyMap, err := ed25519.LoadPublicKeys(cfg.PublicKeys)
	if err != nil {
		log.Error("Failed to load public keys", "err", err)
		os.Exit(1)
	}

	auth, err := grpc.New(ctx, log, cfg.Services.Auth.Address, cfg.Auth.Timeout, cfg.Auth.Retries, cfg.App.Name, pubKey)
	if err != nil {
		panic(err)
	}

	pubkey, keyId, err := auth.GetPublicKey(ctx)
	if err != nil {
		panic(err)
	}

	pubKeyMap[keyId] = *pubkey

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderCookie},
		AllowMethods:     []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	}))
	e.Use(middleware2.JWTMiddleware(middleware2.JWTMiddlewareConfig{
		PublicKeys: pubKeyMap,
		CookieName: "access_token",
	}))

	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "true",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
	}))

	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(cfg.RateLimit))))

	e.POST("/api/v1/login", handlers.Login(ctx, log, auth, auth.AppId, true, cookieName, *auth.AuthPubkey))
	e.POST("/api/v1/register", handlers.Register(ctx, log, auth, auth.AppId, true, cookieName, *auth.AuthPubkey))
	e.GET("/api/v1/files", handlers.ListFiles(ctx, log, sql, cli, bucket, cfg.Expiry))
	e.GET("/api/v1/verify-token", func(c echo.Context) error {
		if uid := c.Get("uid"); uid != nil {
			return c.JSON(http.StatusOK, uid)
		}
		return c.NoContent(http.StatusUnauthorized)
	})

	e.POST("/api/v1/logout", func(c echo.Context) error {
		c.SetCookie(&http.Cookie{
			Name:     cookieName,
			Value:    "",
			Expires:  time.Unix(0, 0),
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		return c.NoContent(http.StatusOK)
	})

	e.Logger.Fatal(e.Start(cfg.Server.Address))
	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("signal received, shutting down gracefully")
}
