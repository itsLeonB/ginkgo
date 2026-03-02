package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ezutil/v2/simple"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestNewRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := simple.NewLogger("test", true, 0)
	mp := NewMiddlewareProvider(logger)

	// Limit 1 req/sec, burst 1
	mw := mp.NewRateLimitMiddleware(rate.Every(time.Second), 1)

	t.Run("allow first request", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.RemoteAddr = "127.0.0.1:1234"

		mw(c)

		assert.False(t, c.IsAborted())
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("rate limit exceeded", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.RemoteAddr = "127.0.0.2:1234" // Different IP

		// First request consumes token
		mw(c)
		assert.False(t, c.IsAborted())

		// Second request should fail immediately
		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest("GET", "/", nil)
		c2.Request.RemoteAddr = "127.0.0.2:1234"

		mw(c2)

		assert.True(t, c2.IsAborted())
		assert.Equal(t, http.StatusTooManyRequests, w2.Code)

		var response map[string]interface{}
		_ = json.Unmarshal(w2.Body.Bytes(), &response)

		errors, ok := response["errors"].([]interface{})
		if ok && len(errors) > 0 {
			errObj := errors[0].(map[string]interface{})
			assert.Equal(t, "Too Many Requests", errObj["code"])
			assert.Equal(t, "rate limit exceeded", errObj["detail"])
		} else {
			assert.Fail(t, "expected errors in response")
		}
	})
}
