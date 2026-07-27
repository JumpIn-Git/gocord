package core

import (
	"gocord/internal/i18n"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("gocord", c)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, i18n.Msg(c, i18n.ErrUnauthorized))
		}
		userID, ok := sess.Values["user_id"].(int64)
		if !ok || userID == 0 {
			return echo.NewHTTPError(http.StatusUnauthorized, i18n.Msg(c, i18n.ErrUnauthorized))
		}
		c.Set("user_id", userID)
		return next(c)
	}
}
