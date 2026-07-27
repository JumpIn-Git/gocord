package api

import (
	"database/sql"
	"gocord/db/query"
	"gocord/internal/core"
	"gocord/internal/i18n"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// /api/{server}/join
func (h *Handler) JoinServer(c echo.Context) error {
	UserID := c.Get("user_id").(int64)
	var req struct {
		Server int64  `param:"server"`
		Invite string `json:"invite"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, i18n.Msg(c, i18n.ErrInvalidBody))
	}

	if h.Srv.Hub.IsUserInServer(UserID, req.Server) {
		return echo.NewHTTPError(http.StatusForbidden, i18n.Msg(c, i18n.ErrAlreadyInServer))
	}

	invite, err := h.Srv.Q.GetInvite(c.Request().Context(), req.Invite)
	if err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	}
	if invite.ServerID != req.Server {
		return echo.NewHTTPError(http.StatusBadRequest, i18n.Msg(c, i18n.ErrInviteMismatch))
	}
	if time.Now().Compare(*invite.ExpiresAt) < 0 {
		if err := h.Srv.Q.DeleteInvite(c.Request().Context(), req.Invite); err != nil {
			h.Srv.Logger.Error(err)
		}
		return echo.NewHTTPError(http.StatusBadRequest, i18n.Msg(c, i18n.ErrInviteExpired))
	}
	if err := h.Srv.Q.CreateMembership(c.Request().Context(), query.CreateMembershipParams{
		UserID:   UserID,
		ServerID: req.Server,
	}); err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	}
	h.Srv.Hub.Joined <- core.UserJoined{
		UserID: UserID,
		Server: req.Server,
	}
	return c.JSON(http.StatusOK, nil)
}

func (h *Handler) LeaveServer(c echo.Context) error {
	UserID := c.Get("user_id").(int64)
	var req struct {
		Server int64 `param:"server"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, i18n.Msg(c, i18n.ErrInvalidBody))
	}

	if err := h.Srv.Q.LeaveServer(c.Request().Context(), query.LeaveServerParams{
		UserID:   UserID,
		ServerID: req.Server,
	}); err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, i18n.Msg(c, i18n.ErrUserNotInServer))
		}
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	}
	h.Srv.Hub.Left <- core.UserLeft{
		UserID: UserID,
		Server: req.Server,
	}
	return c.JSON(http.StatusOK, nil)
}
