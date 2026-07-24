package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"gocord/db/query"
)

func (h *Handler) GetMessages(c echo.Context) error {
	UserID := c.Get("user_id").(int64)
	ServerID := c.Get("server_id").(int64)
	Offset, err := strconv.ParseInt(c.Param("offset"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if ok, err := h.Q.UserInServer(c.Request().Context(), query.UserInServerParams{
		UserID:   UserID,
		ServerID: ServerID,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	} else if !ok {
		return echo.NewHTTPError(http.StatusForbidden, "user not in server")
	}

	messages, err := h.Q.GetServerMessages(c.Request().Context(), query.GetServerMessagesParams{
		ServerID: ServerID,
		Limit:    20,
		Offset:   Offset,
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
	ServerID := c.Get("server_id").(int64)
	if ok, err := h.Q.UserInServer(c.Request().Context(), query.UserInServerParams{
		UserID:   UserID,
		ServerID: ServerID,
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
		ServerID: ServerID,
		UserID:   UserID,
		Content:  req.Content,
		ReplyTo:  sql.NullInt64{Int64: *req.ReplyTo, Valid: req.ReplyTo != nil},
		IsReply:  req.ReplyTo != nil,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) DeleteMessage(c echo.Context) error {
	ServerID := c.Get("server_id").(int64)
	UserID := c.Get("user_id").(int64)
	MessageID, err := strconv.ParseInt(c.Param("message"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if ok, err := h.Q.UserInServer(c.Request().Context(), query.UserInServerParams{
		UserID:   UserID,
		ServerID: ServerID,
	}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	} else if !ok {
		return echo.NewHTTPError(http.StatusForbidden, "user not in server")
	}

	if n, err := h.Q.DeleteMessage(c.Request().Context(), MessageID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	} else if n == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "you are not authorized or message not found")
	}
	return c.NoContent(http.StatusOK)
}
