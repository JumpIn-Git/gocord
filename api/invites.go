package api

import (
	"crypto/rand"
	"gocord/db/query"
	"math/big"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// /api/:server/invite POST, make a new invite
func (h *Handler) PostInvite(c echo.Context) error {
	var req struct {
		Server   int64  `param:"server"`
		Duration string `json:"duration"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}

	var duration time.Duration
	switch req.Duration {
	case "30m":
		duration = 30 * time.Minute
	case "1h":
		duration = time.Hour
	case "6h":
		duration = 6 * time.Hour
	case "12h":
		duration = 12 * time.Hour
	case "24h":
		duration = 24 * time.Hour
	case "1d":
		duration = 24 * time.Hour
	case "7d":
		duration = 7 * 24 * time.Hour
	case "never":
		duration = 0
	default:
		return echo.NewHTTPError(400, "invalid duration")
	}

	UserID := c.Get("user_id").(int64)

	if duration == 0 { // only server owner can make invites that never expire
		if ok, err := h.Srv.Q.IsOwner(c.Request().Context(), query.IsOwnerParams{
			ID:    req.Server,
			Owner: UserID,
		}); err != nil {
			h.Srv.Logger.Error(err)
			return err
		} else if !ok {
			return echo.NewHTTPError(403, "only the server owner can make invites that never expire")
		}
	}

	var expiresAt *time.Time
	if duration != 0 { // null if no expiry
		t := time.Now().Add(duration)
		expiresAt = &t
	}
	inviteCode, err := h.MakeInvite()
	if err != nil {
		return err
	}
	if err := h.Srv.Q.MakeInvite(c.Request().Context(), query.MakeInviteParams{
		ID:        inviteCode,
		ServerID:  req.Server,
		UserID:    UserID,
		ExpiresAt: expiresAt,
	}); err != nil {
		h.Srv.Logger.Error(err)
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"code": inviteCode,
	})
}

func (h *Handler) MakeInvite() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 9)
	max := big.NewInt(int64(len(charset)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}
