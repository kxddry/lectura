package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"net/http"
)

func JWTFromCookie(secret []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("access_token")
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing access token")
			}

			tokenStr := cookie.Value

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid signing method")
				}
				return secret, nil
			})

			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				c.Set("uid", claims["uid"])
			}
			return next(c)
		}
	}
}
