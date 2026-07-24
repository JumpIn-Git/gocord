package api

import (
	"gocord/db/query"
	"gocord/internal/core"

	"github.com/bwmarrin/snowflake"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	Q      *query.Queries
	Flake  *snowflake.Node
	Secret []byte
	Hub    *core.Hub
}

func (h *Handler) Route(e *echo.Echo) {
	api := e.Group("/api")
	api.POST("/auth/register", h.Register)
	api.POST("/auth/login", h.Login)

	servers := api.Group("/servers/:server")
	servers.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: h.Secret,
	}))
	servers.Use(core.AuthMiddleware)
	servers.GET("/messages/:offset", h.GetMessages)
	servers.DELETE("/messages/:message", h.DeleteMessage)
	servers.POST("/messages", h.PostMessage)
}
