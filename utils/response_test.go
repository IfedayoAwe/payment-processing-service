package utils

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	data := map[string]string{"key": "value"}
	err := Success(c, data, "Success message")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Success message")
	assert.Contains(t, rec.Body.String(), "value")
}

func TestCreated(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	data := map[string]string{"id": "123"}
	err := Created(c, data, "Resource created")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "Resource created")
}

func TestHandleError(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name       string
		err        error
		statusCode int
	}{
		{
			name:       "NotFound",
			err:        NotFoundErr("not found"),
			statusCode: http.StatusNotFound,
		},
		{
			name:       "BadRequest",
			err:        BadRequestErr("bad request"),
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Conflict",
			err:        DuplicateKeyErr("duplicate key"),
			statusCode: http.StatusConflict,
		},
		{
			name:       "InternalError",
			err:        ServerErr(errors.New("internal error")),
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := HandleError(c, tt.err)
			assert.NoError(t, err)
			assert.Equal(t, tt.statusCode, rec.Code)
		})
	}
}
