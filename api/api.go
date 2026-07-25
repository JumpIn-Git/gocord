package api

import (
	"gocord/internal/core"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	Srv *core.Server
}

func (h *Handler) Route(e *echo.Echo) {
	api := e.Group("/api")
	api.POST("/auth/register", h.Register)
	api.POST("/auth/login", h.Login)

	servers := api.Group("/servers/:server")
	servers.Use(core.AuthMiddleware)
	servers.GET("/messages/:offset", h.GetMessages)
	servers.DELETE("/messages/:message", h.DeleteMessage)
	servers.POST("/messages", h.PostMessage)
}
