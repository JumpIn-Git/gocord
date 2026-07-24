package core

import (
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, err := echo.ContextGet[*jwt.Token](c, "user")
		if err != nil {
			return echo.ErrUnauthorized.WithInternal(err)
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "failed to cast claims as jwt.MapClaims")
		}

		userIDStr, ok := claims["user_id"].(string)
		if !ok || userIDStr == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}
		parsedUserID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
		}

		c.Set("user_id", parsedUserID)
		return next(c)
	}
}
