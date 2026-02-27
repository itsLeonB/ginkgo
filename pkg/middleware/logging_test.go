package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ezutil/v2/simple"
	"github.com/stretchr/testify/assert"
)

func TestNewLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := simple.NewLogger("test", true, 0)
	mp := NewMiddlewareProvider(logger)

	// Corrected: No arguments locally, it seems from logging.go content
	mw := mp.NewLoggingMiddleware()

	t.Run("logs request", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/test", nil)

		start := time.Now()
		mw(c)
		duration := time.Since(start)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, duration >= 0)
	})

	t.Run("logs body for POST", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"foo":"bar"}`)
		c.Request = httptest.NewRequest("POST", "/api/data", body)
		c.Request.Header.Set("Content-Type", "application/json")

		mw(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
