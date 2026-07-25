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

	servers := api.Group("/servers/:server", core.AuthMiddleware)

	messages := servers.Group("/messages")
	messages.GET("/:offset", h.GetMessages)
	messages.DELETE("/:message", h.DeleteMessage)
	messages.POST("/", h.PostMessage)

	reactions := servers.Group("/react")
	reactions.POST("/:message", h.PostReaction)
	reactions.DELETE("/:message", h.DeleteReaction)
}
