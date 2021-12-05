package session

import (
	"encoding/gob"
	"errors"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const (
	sessionKey = "session"
	UserIdKey  = "userId"
)

func Middleware() echo.MiddlewareFunc {
	return session.Middleware(sessions.NewCookieStore([]byte("secret")))
}

func Get(key string, c echo.Context) (interface{}, error) {
	sess, err := session.Get(sessionKey, c)
	if err != nil {
		return nil, err
	}

	value, ok := sess.Values[key]
	if !ok {
		return nil, errors.New("session value not found")
	}

	return value, nil
}

func Set(key string, value interface{}, c echo.Context) error {
	gob.Register(value)

	sess, _ := session.Get(sessionKey, c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24, // 1 day
		HttpOnly: true,
	}
	sess.Values[key] = value
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return nil
}
