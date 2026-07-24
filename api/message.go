package api

import (
	"database/sql"
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
		return err
	}
	if err := c.Bind(&req); err != nil {
		return err
	}

	if ok, err := h.Q.UserInServer(c.Request().Context(), query.UserInServerParams{
		UserID:   UserID,
		ServerID: req.Server,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	} else if !ok {
		return echo.NewHTTPError(http.StatusForbidden, "user not in server")
	}

	messages, err := h.Q.GetServerMessages(c.Request().Context(), query.GetServerMessagesParams{
		ServerID: req.Server,
		Limit:    20,
		Offset:   req.Offset,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if len(messages) == 0 {
		return c.NoContent(http.StatusNoContent)
	}
	return c.JSON(http.StatusOK, messages)
}

func (h *Handler) PostMessage(c echo.Context) error {
	UserID := c.Get("user_id").(int64)

	var params struct {
		Server int64 `param:"server"`
	}
	if err := c.Bind(&params); err != nil {
		return err
	}

	if ok, err := h.Q.UserInServer(c.Request().Context(), query.UserInServerParams{
		UserID:   UserID,
		ServerID: params.Server,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	} else if !ok {
		return echo.NewHTTPError(http.StatusForbidden, "user not in server")
	}

	var req struct {
		Content string `json:"content"`
		ReplyTo *int64 `json:"reply_to,omitempty,string"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.Q.CreateMessage(c.Request().Context(), query.CreateMessageParams{
		ID:       h.Flake.Generate().Int64(),
		ServerID: params.Server,
		UserID:   UserID,
		Content:  req.Content,
		ReplyTo:  sql.NullInt64{Int64: *req.ReplyTo, Valid: req.ReplyTo != nil},
		IsReply:  req.ReplyTo != nil,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	h.Hub.Broadcast <- core.Event{
		Type:     "new_msg",
		ServerID: params.Server,
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
		return err
	}

	if ok, err := h.Q.UserInServer(c.Request().Context(), query.UserInServerParams{
		UserID:   UserID,
		ServerID: req.Server,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	} else if !ok {
		return echo.NewHTTPError(http.StatusForbidden, "user not in server")
	}

	if n, err := h.Q.DeleteMessage(c.Request().Context(), req.MessageID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	} else if n == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "you are not authorized or message not found")
	}
	h.Hub.Broadcast <- core.Event{
		Type:     "del_msg",
		ServerID: req.Server,
		Payload:  req.MessageID,
	}
	return c.NoContent(http.StatusOK)
}
