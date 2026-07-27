package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"gocord/db/query"
	"gocord/internal/i18n"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Register(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, i18n.Msg(c, i18n.ErrInvalidBody))
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, i18n.Msg(c, i18n.ErrUsernamePasswordRequired))
	}
	if len(req.Username) > 32 || len(req.Password) > 32 {
		return echo.NewHTTPError(http.StatusBadRequest, i18n.Msg(c, i18n.ErrUsernamePasswordTooLong))
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	}

	id := h.Srv.Flake.Generate().Int64()
	if err = h.Srv.Q.CreateUser(c.Request().Context(), query.CreateUserParams{
		ID:           id,
		Username:     req.Username,
		Display:      req.Username,
		PasswordHash: string(hash),
	}); err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "PRIMARY KEY constraint failed") {
			return echo.NewHTTPError(http.StatusConflict, i18n.Msg(c, i18n.ErrUsernameTaken))
		}
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	}

	sess, err := session.Get("gocord", c)
	if err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	}
	sess.Values["user_id"] = id
	sess.Values["session_token"] = h.Srv.Flake.Generate().Int64()
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, map[string]any{
		"user_id":  id,
		"username": req.Username,
	})
}

func (h *Handler) Login(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, i18n.Msg(c, i18n.ErrInvalidBody))
	}

	user, err := h.Srv.Q.GetUserByUsername(c.Request().Context(), req.Username)
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusUnauthorized, i18n.Msg(c, i18n.ErrInvalidCredentials))
	} else if err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, i18n.Msg(c, i18n.ErrInvalidCredentials))
	}

	sess, err := session.Get("gocord", c)
	if err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	}
	sess.Values["user_id"] = user.ID
	sess.Values["session_token"] = h.Srv.Flake.Generate().Int64()
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, map[string]any{
		"user_id":  user.ID,
		"username": user.Username,
	})
}

func (h *Handler) Logout(c echo.Context) error {
	sess, err := session.Get("gocord", c)
	if err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	}
	userID, _ := sess.Values["user_id"].(int64)
	token, _ := sess.Values["session_token"].(int64)

	sess.Values = make(map[any]any)
	sess.Options.MaxAge = -1 // browser will delete the cookie
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		h.Srv.Logger.Error(err)
		return echo.ErrInternalServerError
	}

	if userID != 0 {
		h.Srv.Hub.Disconnect(userID, token)
	}
	return c.JSON(http.StatusOK, nil)
}
