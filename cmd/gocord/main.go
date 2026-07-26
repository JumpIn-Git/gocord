package main

import (
	"database/sql"
	"gocord/api"
	"gocord/internal/core"
	"os"
	"time"

	migrations "gocord/db"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/pressly/goose/v3"
	"github.com/tursodatabase/go-libsql"
)

func main() {
	e := echo.New()
	logger := e.Logger

	var db *sql.DB
	primaryURL := os.Getenv("TURSO_DATABASE_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")
	if primaryURL == "" || authToken == "" {
		logger.Print("Missing TURSO credentials, using local file")
		var err error
		db, err = sql.Open("libsql", "file:local.db")
		if err != nil {
			logger.Fatal(err)
		}
	} else {
		conn, err := libsql.NewEmbeddedReplicaConnector(
			"local.db",
			primaryURL,
			libsql.WithAuthToken(authToken),
			libsql.WithReadYourWrites(true),
		)
		if err != nil {
			logger.Fatal(err)
		}
		db = sql.OpenDB(conn)
	}

	if err := db.Ping(); err != nil {
		logger.Fatal(err)
	}

	goose.SetBaseFS(migrations.Migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		logger.Fatal(err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		logger.Fatal(err)
	}

	srv, err := core.NewServer(db, 1, e)
	if err != nil {
		logger.Fatal(err)
	}

	secret := os.Getenv("AUTH")
	enc := os.Getenv("ENC")
	if secret == "" || enc == "" {
		logger.Fatal("AUTH and ENC environment variables must be set")
	}
	store := sessions.NewCookieStore([]byte(secret), []byte(enc))
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.MaxAge = 60 * 60 * 24 // 1 day
	srv.Echo.Use(session.Middleware(store))

	go srv.Hub.Run()

	apiHandler := &api.Handler{
		Srv: srv,
	}

	srv.GET("/ws", srv.HandleWebSocket(srv.Hub), core.AuthMiddleware)
	apiHandler.Route(srv.Echo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logger.Infof("Listining on %s", port)
	go func() {
		for {
			time.Sleep(24 * time.Hour) // Dont run it often because we will also check expiry on invite usage
			if err := srv.Q.DeleteExpiredInvites(srv.Context); err != nil {
				logger.Error(err)
			}
		}
	}()
	logger.Fatal(srv.Start(":" + port))
}
