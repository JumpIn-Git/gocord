package core

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/jwtauth/v5"
)

func GetUserIDFromRequest(r *http.Request) (int64, error, int) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return 0, errors.New("Failed to authenticate"), http.StatusUnauthorized
	}
	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return 0, errors.New("Failed to parse user_id"), http.StatusBadRequest
	}
	parsedUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return 0, errors.New("Failed to parse user_id"), http.StatusInternalServerError
	}
	return parsedUserID, nil, 0
}
