package middleware

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kxddry/lectura/shared/entities/auth"
	"github.com/labstack/echo/v4"
	"net/http"
)

type JWTMiddlewareConfig struct {
	PublicKeys map[string]ed25519.PublicKey
	CookieName string
}

func JWTMiddleware(cfg JWTMiddlewareConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(cfg.CookieName)
			if err != nil {
				if errors.Is(err, http.ErrNoCookie) {
					return next(c)
				}
				return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
			}

			tokenStr := cookie.Value

			parsedToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if token.Method.Alg() != jwt.SigningMethodEdDSA.Alg() {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				kid, ok := token.Header["kid"].(string)
				if !ok {
					return nil, errors.New("missing kid in token header")
				}
				key, ok := cfg.PublicKeys[kid]
				if !ok {
					return nil, fmt.Errorf("unknown kid %s", kid)
				}
				return key, nil
			})

			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
			}

			if !parsedToken.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			claims, err := auth.ParseJWT(tokenStr, cfg.PublicKeys[parsedToken.Header["kid"].(string)])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
			}

			c.Set("uid", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("exp", claims.Exp)
			c.Set("app_id", claims.AppID)

			return next(c)
		}
	}
}
