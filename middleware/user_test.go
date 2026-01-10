package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestUserIDMiddleware(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test with X-User-ID header
	req.Header.Set("X-User-ID", "user_1")

	handler := UserIDMiddleware()(func(c echo.Context) error {
		userID := GetUserID(c)
		assert.Equal(t, "user_1", userID)
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUserIDMiddleware_NoHeader(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test without X-User-ID header
	handler := UserIDMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	err := handler(c)
	assert.Error(t, err)

	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
}

func TestGetUserID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test when user ID is set
	c.Set(UserIDKey, "user_1")
	assert.Equal(t, "user_1", GetUserID(c))

	// Test when user ID is not set
	c2 := e.NewContext(req, rec)
	assert.Equal(t, "", GetUserID(c2))
}
