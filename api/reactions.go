package api

import (
	"gocord/db/query"
	"net/http"

	"github.com/labstack/echo/v4"
)

// /api/{server}/react/{message}
func (h *Handler) PostReaction(c echo.Context) error {
	UserID := c.Get("user_id").(int64)
	var req struct {
		ServerID  int64  `param:"server"`
		MessageID int64  `param:"message"`
		Emoji     string `json:"emoji"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	if ok, err := h.Q.UserInServer(c.Request().Context(), query.UserInServerParams{
		UserID:   UserID,
		ServerID: req.ServerID,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	} else if !ok {
		return echo.NewHTTPError(http.StatusForbidden, "user not in server")
	}
	if err := h.Q.CreateReaction(c.Request().Context(), query.CreateReactionParams{
		MessageID: req.MessageID,
		UserID:    UserID,
		Emoji:     req.Emoji,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, nil)
}

func (h *Handler) DeleteReaction(c echo.Context) error {
	UserID := c.Get("user_id").(int64)
	var req struct {
		ServerID  int64  `param:"server"`
		MessageID int64  `param:"message"`
		Emoji     string `json:"emoji"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}
	if ok, err := h.Q.UserInServer(c.Request().Context(), query.UserInServerParams{
		UserID:   UserID,
		ServerID: req.ServerID,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	} else if !ok {
		return echo.NewHTTPError(http.StatusForbidden, "user not in server")
	}
	if err := h.Q.DeleteReaction(c.Request().Context(), query.DeleteReactionParams{
		MessageID: req.MessageID,
		UserID:    UserID,
		Emoji:     req.Emoji,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, nil)
}
