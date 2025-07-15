package grpc

import (
	"context"
	"crypto/ed25519"
	"fmt"
	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/kxddry/lectura/shared/utils/base64"
	ssov2 "github.com/kxddry/sso-protos/v2/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"time"
)

type Client struct {
	api        ssov2.AuthClient
	log        *slog.Logger
	AuthPubkey *ed25519.PublicKey
	AppPubkey  *ed25519.PublicKey
	AppId      int64
	AppName    string
}

func New(ctx context.Context, log *slog.Logger, addr string, timeout time.Duration, retries int, appName string, pubKey ed25519.PublicKey) (*Client, error) {
	const op = "auth.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithBackoff(grpcretry.BackoffLinear(timeout)),
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(retries)),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	cc, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	c := &Client{
		api: ssov2.NewAuthClient(cc),
		log: log,
	}

	c.AppPubkey = &pubKey

	resp, err := c.api.GetPublicKey(ctx, &ssov2.PubkeyRequest{})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	key, err := base64.UnmarshalPubKey(resp.Pubkey)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	c.AuthPubkey = &key
	c.AppName = appName
	c.AppId, err = c.AppID(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return c, nil
}

func (c *Client) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "grpc.IsAdmin"
	resp, err := c.api.IsAdmin(ctx, &ssov2.IsAdminRequest{
		UserId: userID,
	})
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return resp.IsAdmin, nil
}

func (c *Client) AppID(ctx context.Context) (int64, error) {
	const op = "grpc.AppID"

	resp, err := c.api.AppID(ctx, &ssov2.AppRequest{
		Name:   c.AppName,
		Pubkey: base64.MarshalPubKey(*c.AppPubkey),
	})

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return resp.AppId, nil
}

func (c *Client) Login(ctx context.Context, email, pass string, appId int64) (string, error) {
	const op = "grpc.Login"
	resp, err := c.api.Login(ctx, &ssov2.LoginRequest{
		Email:    email,
		Password: pass,
		AppId:    appId,
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return resp.Token, nil
}

func (c *Client) Register(ctx context.Context, email, password string) (int64, error) {
	const op = "grpc.Register"
	resp, err := c.api.Register(ctx, &ssov2.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return resp.UserId, nil
}

func InterceptorLogger(l *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, lvl grpclog.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
