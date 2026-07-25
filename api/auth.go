package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"gocord/db/query"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Register(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "username and password required")
	}
	if len(req.Username) > 32 || len(req.Password) > 32 {
		return echo.NewHTTPError(http.StatusBadRequest, "username and password must be 32 characters or less")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return echo.ErrInternalServerError
	}

	id := h.Srv.Flake.Generate().Int64()
	if err = h.Srv.Q.CreateUser(c.Request().Context(), query.CreateUserParams{
		ID:           id,
		Username:     req.Username,
		Display:      req.Username,
		PasswordHash: string(hash),
	}); err != nil {
		if extErr, ok := errors.AsType[*sqlite3.Error](err); ok {
			if extErr.ExtendedCode == sqlite3.ErrConstraintUnique || extErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
				return echo.NewHTTPError(http.StatusConflict, "Username already taken")
			}
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	sess, err := session.Get("gocord", c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	sess.Values["user_id"] = id
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
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
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	user, err := h.Srv.Q.GetUserByUsername(c.Request().Context(), req.Username)
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	sess, err := session.Get("gocord", c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	sess.Values["user_id"] = user.ID
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]any{
		"user_id":  user.ID,
		"username": user.Username,
	})
}

func (h *Handler) Logout(c echo.Context) error {
	sess, err := session.Get("gocord", c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	sess.Values = make(map[any]any)
	sess.Options.MaxAge = -1 // browser will remove the cookie
	// TODO add cookie_ver to sqlite users to make old cookies useless
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, nil)
}
