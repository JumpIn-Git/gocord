package api

import (
	"gocord/db/query"

	"github.com/labstack/echo/v4"
)

// /api/{server}/join
func (h *Handler) JoinServer(c echo.Context) {
	UserID := c.Get("user_id").(int64)
	var req struct {
		Server int64  `param:"server"`
		Invite string `json:"invite"`
	}
	if err := c.Bind(&req); err != nil {
		return
	}
	if ok, err := h.Srv.Q.UserInServer(c.Request().Context(), query.UserInServerParams{
		UserID:   UserID,
		ServerID: req.Server,
	}); err != nil {
		c.JSON(500, err)
		return
	}
}
