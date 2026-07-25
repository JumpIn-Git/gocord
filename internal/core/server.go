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
	secret []byte
	Hub    *Hub
}

func NewServer(db *sql.DB, secret []byte, flakeInt int64) (*Server, error) {
	flake, err := snowflake.NewNode(flakeInt)
	if err != nil {
		return nil, err
	}
	return &Server{
		Echo:    echo.New(),
		Q:       query.New(db),
		Flake:   flake,
		Context: context.Background(),
		secret:  secret,
		Hub:     NewHub(),
	}, nil
}
