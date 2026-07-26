package core

import (
	"context"

	"gocord/db/query"

	"github.com/bwmarrin/snowflake"
	"github.com/labstack/echo/v4"
)

type Server struct {
	*echo.Echo
	Q     *query.Queries
	Flake *snowflake.Node
	context.Context
	Hub *Hub
}
