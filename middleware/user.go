package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const UserIDKey = "user_id"

func UserIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID := c.Request().Header.Get("X-User-ID")
			if userID == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "X-User-ID header is required")
			}

			c.Set(UserIDKey, userID)
			return next(c)
		}
	}
}

func GetUserID(c echo.Context) string {
	userID, ok := c.Get(UserIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}
