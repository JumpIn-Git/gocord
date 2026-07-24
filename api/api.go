package api

import (
	"net/http"
	"strconv"

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
	servers.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			serverIDStr := c.Param("server")
			if serverIDStr == "" {
				return echo.NewHTTPError(http.StatusBadRequest, "server id required")
			}
			serverID, err := strconv.ParseInt(serverIDStr, 10, 64)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid server id")
			}

			c.Set("server_id", serverID)
			return next(c)
		}
	})
	servers.GET("/messages/:offset", h.GetMessages)
	servers.DELETE("/messages/:message", h.DeleteMessage)
	servers.POST("/messages", h.PostMessage)
}
