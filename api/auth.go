package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"gocord/db/query"
	"net/http"
	"strconv"
	"strings"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type AuthResponse struct {
	Token string `json:"token"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password required", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	id := h.Flake.Generate().Int64()
	if err = h.Q.CreateUser(r.Context(), query.CreateUserParams{
		ID:           id,
		Username:     req.Username,
		Display:      req.Username,
		PasswordHash: string(hash),
	}); err != nil {
		if extErr, ok := errors.AsType[*sqlite3.Error](err); ok {
			if extErr.ExtendedCode == sqlite3.ErrConstraintUnique || extErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
				http.Error(w, "Username already taken", http.StatusConflict)
				return
			}
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, token, err := h.TokenAuth.Encode(map[string]any{
		"user_id": strconv.FormatInt(id, 10),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(AuthResponse{Token: token})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	user, err := h.Q.GetUserByUsername(r.Context(), req.Username)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	_, token, err := h.TokenAuth.Encode(map[string]any{
		"user_id": strconv.FormatInt(user.ID, 10),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(AuthResponse{Token: token})
}
