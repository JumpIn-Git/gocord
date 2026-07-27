package i18n

import (
	"embed"
	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var bundle *i18n.Bundle

//go:embed *.json
var localeFS embed.FS

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	files, err := localeFS.ReadDir(".")
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		_, err := bundle.LoadMessageFileFS(localeFS, file.Name())
		if err != nil {
			panic(err)
		}
	}
}

type ErrorCode string

const (
	ErrInvalidBody              ErrorCode = "ErrInvalidBody"
	ErrUsernamePasswordRequired ErrorCode = "ErrUsernamePasswordRequired"
	ErrUsernamePasswordTooLong  ErrorCode = "ErrUsernamePasswordTooLong"
	ErrUsernameTaken            ErrorCode = "ErrUsernameTaken"
	ErrInvalidCredentials       ErrorCode = "ErrInvalidCredentials"
	ErrAlreadyInServer          ErrorCode = "ErrAlreadyInServer"
	ErrInviteMismatch           ErrorCode = "ErrInviteMismatch"
	ErrInviteExpired            ErrorCode = "ErrInviteExpired"
	ErrInvalidDuration          ErrorCode = "ErrInvalidDuration"
	ErrOwnerOnlyInvite          ErrorCode = "ErrOwnerOnlyInvite"
	ErrUserNotInServer          ErrorCode = "ErrUserNotInServer"
	ErrNotAuthorized            ErrorCode = "ErrNotAuthorized"
	ErrContentRequired          ErrorCode = "ErrContentRequired"
	ErrMessageNotFound          ErrorCode = "ErrMessageNotFound"
	ErrInvalidEmoji             ErrorCode = "ErrInvalidEmoji"
	ErrReactionNotFound         ErrorCode = "ErrReactionNotFound"
	ErrUnauthorized             ErrorCode = "ErrUnauthorized"
)

const defaultLang = "en"

func Msg(c echo.Context, code ErrorCode) string {
	lang := c.QueryParam("lang")
	accept := c.Request().Header.Get("Accept-Language")
	localizer := i18n.NewLocalizer(bundle, lang, accept)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: string(code)})
	if err != nil {
		c.Logger().Error(err)
		return "Unknown error"
	}
	return msg
}
