package main

import (
	"database/sql"
	"fmt"
	"gocord/api"
	"gocord/internal/core"
	"log"
	"os"
	"time"

	migrations "gocord/db"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/pressly/goose/v3"
	"github.com/tursodatabase/go-libsql"
)

func main() {
	var db *sql.DB
	primaryURL := os.Getenv("TURSO_DATABASE_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")
	if primaryURL == "" || authToken == "" {
		log.Println("Missing TURSO credentials, using local file")
		var err error
		db, err = sql.Open("libsql", "file:local.db")
		if err != nil {
			panic(err)
		}
	} else {
		conn, err := libsql.NewEmbeddedReplicaConnector(
			"local.db",
			primaryURL,
			libsql.WithAuthToken(authToken),
			libsql.WithReadYourWrites(true),
		)
		if err != nil {
			panic(err)
		}
		db = sql.OpenDB(conn)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	goose.SetBaseFS(migrations.Migrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		panic(err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		panic(err)
	}

	srv, err := core.NewServer(db, 1)
	if err != nil {
		panic(err)
	}

	secret := []byte(os.Getenv("AUTH"))
	enc := []byte(os.Getenv("ENC"))
	store := sessions.NewCookieStore(secret, enc)
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
	fmt.Printf("Listining on %s", port)
	go func() {
		for {
			time.Sleep(24 * time.Hour) // Dont run it often because we will also check expiry on invite usage
			if err := srv.Q.DeleteExpiredInvites(srv.Context); err != nil {
				srv.Logger.Error(err)
			}
		}
	}()
	srv.Logger.Fatal(srv.Start(":" + port))
}
