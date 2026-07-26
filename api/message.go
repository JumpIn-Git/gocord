package api

import (
	"net/http"

	"gocord/db/query"
	"gocord/internal/core"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetMessages(c echo.Context) error {
	UserID := c.Get("user_id").(int64)
	var req struct {
		Server int64 `param:"server"`
		Offset int64 `param:"offset"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	if ok, err := h.Srv.Q.UserInServer(c.Request().Context(), query.UserInServerParams{
		UserID:   UserID,
		ServerID: req.Server,
	}); err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	} else if !ok {
		return echo.NewHTTPError(http.StatusForbidden, "user not in server")
	}

	messages, err := h.Srv.Q.GetServerMessages(c.Request().Context(), query.GetServerMessagesParams{
		ServerID: req.Server,
		Limit:    20,
		Offset:   req.Offset,
	})
	if err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	}
	if len(messages) == 0 {
		return c.NoContent(http.StatusNoContent)
	}
	return c.JSON(http.StatusOK, messages)
}

func (h *Handler) PostMessage(c echo.Context) error {
	UserID := c.Get("user_id").(int64)
	var req struct {
		Server  int64  `param:"server"`
		Content string `json:"content"`
		ReplyTo *int64 `json:"reply_to,omitempty,string"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	if ok, err := h.Srv.Q.UserInServer(c.Request().Context(), query.UserInServerParams{
		UserID:   UserID,
		ServerID: req.Server,
	}); err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	} else if !ok {
		return echo.NewHTTPError(http.StatusForbidden, "user not in server")
	}

	if err := h.Srv.Q.CreateMessage(c.Request().Context(), query.CreateMessageParams{
		ID:       h.Srv.Flake.Generate().Int64(),
		ServerID: req.Server,
		UserID:   UserID,
		Content:  req.Content,
		ReplyTo:  req.ReplyTo,
		IsReply:  req.ReplyTo != nil,
	}); err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	}

	h.Srv.Hub.Broadcast <- core.Event{
		Type:     "new_msg",
		ServerID: req.Server,
		Payload:  req,
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) DeleteMessage(c echo.Context) error {
	UserID := c.Get("user_id").(int64)
	var req struct {
		Server    int64 `param:"server"`
		MessageID int64 `param:"message"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	if ok, err := h.Srv.Q.UserInServer(c.Request().Context(), query.UserInServerParams{
		UserID:   UserID,
		ServerID: req.Server,
	}); err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	} else if !ok {
		return echo.NewHTTPError(http.StatusForbidden, "user not in server")
	}

	if n, err := h.Srv.Q.DeleteMessage(c.Request().Context(), query.DeleteMessageParams{
		ID:     req.MessageID,
		UserID: UserID,
	}); err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	} else if n == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "you are not authorized or message not found")
	}
	h.Srv.Hub.Broadcast <- core.Event{
		Type:     "del_msg",
		ServerID: req.Server,
		Payload:  req.MessageID,
	}
	return c.NoContent(http.StatusOK)
}
