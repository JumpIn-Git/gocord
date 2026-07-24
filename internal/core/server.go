package core

import (
	"context"
	"database/sql"

	"github.com/bwmarrin/snowflake"
	"github.com/labstack/echo/v4"
	"gocord/db/query"
)

type Server struct {
	*echo.Echo
	Q     *query.Queries
	Flake *snowflake.Node
	context.Context
	secret []byte
}

func NewServer(db *sql.DB, secret []byte) (*Server, error) {
	flake, err := snowflake.NewNode(1)
	if err != nil {
		return nil, err
	}
	return &Server{
		Echo:    echo.New(),
		Q:       query.New(db),
		Flake:   flake,
		Context: context.Background(),
		secret:  secret,
	}, nil
}
