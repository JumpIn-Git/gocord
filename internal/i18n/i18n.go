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

type ErrorCode int

const (
	ErrInvalidBody ErrorCode = iota + 1
	ErrUsernamePasswordRequired
	ErrUsernamePasswordTooLong
	ErrUsernameTaken
	ErrInvalidCredentials
	ErrAlreadyInServer
	ErrInviteMismatch
	ErrInviteExpired
	ErrInvalidDuration
	ErrOwnerOnlyInvite
	ErrUserNotInServer
	ErrNotAuthorized
	ErrContentRequired
	ErrMessageNotFound
	ErrInvalidEmoji
	ErrReactionNotFound
	ErrUnauthorized
)

var codeNames = map[ErrorCode]string{
	ErrInvalidBody:              "ErrInvalidBody",
	ErrUsernamePasswordRequired: "ErrUsernamePasswordRequired",
	ErrUsernamePasswordTooLong:  "ErrUsernamePasswordTooLong",
	ErrUsernameTaken:            "ErrUsernameTaken",
	ErrInvalidCredentials:       "ErrInvalidCredentials",
	ErrAlreadyInServer:          "ErrAlreadyInServer",
	ErrInviteMismatch:           "ErrInviteMismatch",
	ErrInviteExpired:            "ErrInviteExpired",
	ErrInvalidDuration:          "ErrInvalidDuration",
	ErrOwnerOnlyInvite:          "ErrOwnerOnlyInvite",
	ErrUserNotInServer:          "ErrUserNotInServer",
	ErrNotAuthorized:            "ErrNotAuthorized",
	ErrContentRequired:          "ErrContentRequired",
	ErrMessageNotFound:          "ErrMessageNotFound",
	ErrInvalidEmoji:             "ErrInvalidEmoji",
	ErrReactionNotFound:         "ErrReactionNotFound",
	ErrUnauthorized:             "ErrUnauthorized",
}

const defaultLang = "en"

var supported = map[string]bool{
	"en": true,
	"nl": true,
}

func (c ErrorCode) String() string {
	return codeNames[c]
}

func Msg(c echo.Context, code ErrorCode) string {
	lang := c.QueryParam("lang")
	accept := c.Request().Header.Get("Accept-Language")
	localizer := i18n.NewLocalizer(bundle, lang, accept)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: code.String()})
	if err != nil {
		c.Logger().Error(err)
		return "Unknown error"
	}
	return msg
}
