package core

import (
	"context"
	"database/sql"
	"gocord/db/query"

	"github.com/bwmarrin/snowflake"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	*chi.Mux
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
		Mux:     chi.NewRouter(),
		Q:       query.New(db),
		Flake:   flake,
		Context: context.Background(),
		secret:  secret,
	}, nil
}
