package api

import (
	"context"
	"gocord/db/query"
	"net/http"
	"strconv"

	"github.com/bwmarrin/snowflake"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type Handler struct {
	Q         *query.Queries
	Flake     *snowflake.Node
	TokenAuth *jwtauth.JWTAuth
	*chi.Mux
}

func (h *Handler) Route() {
	h.Mux.Route("/api", func(r chi.Router) {
		r.Post("/auth/register", h.Register)
		r.Post("/auth/login", h.Login)

	})
	h.Mux.Route("/api/servers/{server}", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(h.TokenAuth))
			r.Use(h.AuthMiddleware)
			r.Get("/messages/{offset}", h.GetMessages)
			r.Delete("/messages/{message}", h.DeleteMessage)
			r.Post("/messages", h.PostMessage)
		})
	})
}

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // adapted from jwauth.Authenticator
		token, claims, err := jwtauth.FromContext(r.Context())
		if err != nil || token == nil || claims == nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		userIDStr, ok := claims["user_id"].(string)
		if !ok || userIDStr == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		parsedUserID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ServerID, err := strconv.ParseInt(chi.URLParam(r, "server"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", parsedUserID)
		ctx = context.WithValue(ctx, "server_id", ServerID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
