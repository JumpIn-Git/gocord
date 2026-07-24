package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
	"gocord/db/query"
)

type AuthResponse struct {
	Token string `json:"token"`
}

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

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}

	id := h.Flake.Generate().Int64()
	if err = h.Q.CreateUser(c.Request().Context(), query.CreateUserParams{
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

	claims := jwt.MapClaims{
		"user_id": strconv.FormatInt(id, 10),
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.Secret)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, AuthResponse{Token: tokenString})
}

func (h *Handler) Login(c echo.Context) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	user, err := h.Q.GetUserByUsername(c.Request().Context(), req.Username)
	if errors.Is(err, sql.ErrNoRows) {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	claims := jwt.MapClaims{
		"user_id": strconv.FormatInt(user.ID, 10),
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.Secret)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, AuthResponse{Token: tokenString})
}
