package api

import (
	"gocord/db/query"
	"gocord/internal/core"
	"gocord/internal/i18n"
	"net/http"

	"github.com/forPelevin/gomoji"
	"github.com/labstack/echo/v4"
)

// /api/{server}/react/{message} POST
func (h *Handler) PostReaction(c echo.Context) error {
	UserID := c.Get("user_id").(int64)
	var req struct {
		ServerID  int64  `param:"server"`
		MessageID int64  `param:"message"`
		Emoji     string `json:"emoji"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, i18n.Msg(c, i18n.ErrInvalidBody))
	}

	if _, err := gomoji.GetInfo(req.Emoji); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, i18n.Msg(c, i18n.ErrInvalidEmoji))
	}

	if !h.Srv.Hub.IsUserInServer(UserID, req.ServerID) {
		return echo.NewHTTPError(http.StatusForbidden, i18n.Msg(c, i18n.ErrUserNotInServer))
	}
	if err := h.Srv.Q.CreateReaction(c.Request().Context(), query.CreateReactionParams{
		MessageID: req.MessageID,
		UserID:    UserID,
		Emoji:     req.Emoji,
	}); err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	}
	h.Srv.Hub.Broadcast <- core.Event{
		Type:     "new_reaction",
		ServerID: req.ServerID,
		Payload:  req.Emoji,
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
		return echo.NewHTTPError(http.StatusBadRequest, i18n.Msg(c, i18n.ErrInvalidBody))
	}

	if _, err := gomoji.GetInfo(req.Emoji); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, i18n.Msg(c, i18n.ErrInvalidEmoji))
	}

	if !h.Srv.Hub.IsUserInServer(UserID, req.ServerID) {
		return echo.NewHTTPError(http.StatusForbidden, i18n.Msg(c, i18n.ErrUserNotInServer))
	}
	if n, err := h.Srv.Q.DeleteReaction(c.Request().Context(), query.DeleteReactionParams{
		MessageID: req.MessageID,
		UserID:    UserID,
		Emoji:     req.Emoji,
	}); err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	} else if n == 0 {
		return echo.NewHTTPError(http.StatusNotFound, i18n.Msg(c, i18n.ErrReactionNotFound))
	}
	h.Srv.Hub.Broadcast <- core.Event{
		Type:     "del_reaction",
		ServerID: req.ServerID,
		Payload:  req.Emoji,
	}
	return c.JSON(http.StatusOK, nil)
}
