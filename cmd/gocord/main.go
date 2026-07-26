package main

import (
	"context"
	"database/sql"
	"gocord/api"
	"gocord/db/query"
	"gocord/internal/core"
	"os"
	"time"

	migrations "gocord/db"

	"github.com/bwmarrin/snowflake"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/pressly/goose/v3"
	"github.com/tursodatabase/go-libsql"
)

func main() {
	e := echo.New()

	var db *sql.DB
	primaryURL := os.Getenv("TURSO_DATABASE_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")
	if primaryURL == "" || authToken == "" {
		e.Logger.Print("Missing TURSO credentials, using local file")
		var err error
		db, err = sql.Open("libsql", "file:local.db")
		if err != nil {
			e.Logger.Fatal(err)
		}
	} else {
		conn, err := libsql.NewEmbeddedReplicaConnector(
			"local.db",
			primaryURL,
			libsql.WithAuthToken(authToken),
			libsql.WithReadYourWrites(true),
		)
		if err != nil {
			e.Logger.Fatal(err)
		}
		db = sql.OpenDB(conn)
	}

	if err := db.Ping(); err != nil {
		e.Logger.Fatal(err)
	}

	goose.SetBaseFS(migrations.Migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		e.Logger.Fatal(err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		e.Logger.Fatal(err)
	}

	flake, err := snowflake.NewNode(1)
	if err != nil {
		e.Logger.Fatal(err)
	}
	srv := &core.Server{
		Echo:    e,
		Q:       query.New(db),
		Flake:   flake,
		Context: context.Background(),
		Hub:     core.NewHub(),
	}

	auth := os.Getenv("AUTH")
	enc := os.Getenv("ENC")
	if auth == "" || enc == "" {
		e.Logger.Fatal("AUTH and ENC environment variables must be set")
	}
	store := sessions.NewCookieStore([]byte(auth), []byte(enc))
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		MaxAge:   60 * 60 * 24, // 1 day
	}
	srv.Echo.Use(session.Middleware(store))

	go srv.Hub.Run()

	apiHandler := &api.Handler{
		Srv: srv,
	}

	srv.GET("/ws", srv.HandleWebSocket, core.AuthMiddleware)
	apiHandler.Route(srv.Echo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Infof("Listining on %s", port)
	go func() {
		for {
			time.Sleep(24 * time.Hour) // Dont run it often because we will also check expiry on invite usage
			if err := srv.Q.DeleteExpiredInvites(srv.Context); err != nil {
				e.Logger.Error(err)
			}
		}
	}()
	e.Logger.Fatal(srv.Start(":" + port))
}
