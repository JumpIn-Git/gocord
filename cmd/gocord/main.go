package main

import (
	"database/sql"
	"gocord/api"
	"gocord/internal/core"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	opt := url.Values{}
	opt.Add("_journal_mode", "WAL")
	opt.Add("_busy_timeout", "5000")
	opt.Add("_synchronous", "NORMAL")
	opt.Add("_foreign_keys", "ON")
	opt.Add("_txlock", "immediate")
	sqlDB, err := sql.Open("sqlite3", "file:gocord.db?"+opt.Encode())
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	secret, err := os.ReadFile("jwtsecret")
	if err != nil {
		panic(err)
	}

	srv, err := core.NewServer(sqlDB, secret, 1)
	if err != nil {
		panic(err)
	}

	store := sessions.NewCookieStore(secret)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.MaxAge = 60 * 60 * 24 // 1 day
	srv.Echo.Use(session.Middleware(store))

	hub := core.NewHub()
	go hub.Run()

	apiHandler := &api.Handler{
		Srv: srv,
	}

	srv.GET("/ws", srv.HandleWebSocket(hub), core.AuthMiddleware)
	apiHandler.Route(srv.Echo)

	log.Println("listening on :8080")
	go func() {
		for {
			time.Sleep(24 * time.Hour) // Dont run it often because we will also check expiry on invite usage
			if err := srv.Q.DeleteExpiredInvites(srv.Context); err != nil {
				srv.Logger.Error(err)
			}
		}
	}()
	srv.Logger.Fatal(srv.Start(":8080"))
}
