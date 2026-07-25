package core

import (
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("gocord", c)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}
		userID, ok := sess.Values["user_id"].(int64)
		if !ok || userID == 0 {
			return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}
		c.Set("user_id", userID)
		return next(c)
	}
}
