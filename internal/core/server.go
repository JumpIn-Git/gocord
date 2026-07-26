package core

import (
	"context"
	"database/sql"

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

func NewServer(db *sql.DB, flakeInt int64, echo *echo.Echo) (*Server, error) {
	flake, err := snowflake.NewNode(flakeInt)
	if err != nil {
		return nil, err
	}
	return &Server{
		Echo:    echo,
		Q:       query.New(db),
		Flake:   flake,
		Context: context.Background(),
		Hub:     NewHub(),
	}, nil
}
