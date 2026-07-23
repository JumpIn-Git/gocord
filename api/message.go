package api

import (
	"database/sql"
	"encoding/json"
	"gocord/db/query"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// /api/{server}/messages/{offset}, GET; limit of 20
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	UserID := r.Context().Value("user_id").(int64)
	ServerID := r.Context().Value("server_id").(int64)
	Offset, err := strconv.ParseInt(chi.URLParam(r, "offset"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if ok, err := h.Q.UserInServer(r.Context(), query.UserInServerParams{
		UserID:   UserID,
		ServerID: ServerID,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !ok {
		http.Error(w, "user not in server", http.StatusForbidden)
		return
	}

	messages, err := h.Q.GetServerMessages(r.Context(), query.GetServerMessagesParams{
		ServerID: ServerID,
		Limit:    20,
		Offset:   Offset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(messages) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(messages)
}

// /api/{server}/messages/, POST
func (h *Handler) PostMessage(w http.ResponseWriter, r *http.Request) {
	UserID := r.Context().Value("user_id").(int64)
	ServerID := r.Context().Value("server_id").(int64)
	if ok, err := h.Q.UserInServer(r.Context(), query.UserInServerParams{
		UserID:   UserID,
		ServerID: ServerID,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !ok {
		http.Error(w, "user not in server", http.StatusForbidden)
		return
	}

	var req struct {
		Content string `json:"content"`
		ReplyTo *int64 `json:"reply_to,omitempty,string"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Q.CreateMessage(r.Context(), query.CreateMessageParams{
		ID:       h.Flake.Generate().Int64(),
		ServerID: ServerID,
		UserID:   UserID,
		Content:  req.Content,
		ReplyTo:  sql.NullInt64{Int64: *req.ReplyTo, Valid: req.ReplyTo != nil},
		IsReply:  req.ReplyTo != nil,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// /api/{server}/messages/{message}, DELETE
func (h *Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	ServerID := r.Context().Value("server_id").(int64)
	UserID := r.Context().Value("user_id").(int64)
	MessageID, err := strconv.ParseInt(chi.URLParam(r, "message"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if ok, err := h.Q.UserInServer(r.Context(), query.UserInServerParams{
		UserID:   UserID,
		ServerID: ServerID,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !ok {
		http.Error(w, "user not in server", http.StatusForbidden)
		return
	}

	if n, err := h.Q.DeleteMessage(r.Context(), MessageID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if n == 0 {
		http.Error(w, "you are not authorized or message not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}
