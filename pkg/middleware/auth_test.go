package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ezutil/v2/simple"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := simple.NewLogger("test", true, 0)
	mp := NewMiddlewareProvider(logger)

	t.Run("success", func(t *testing.T) {
		tokenCheckFunc := func(ctx *gin.Context, token string) (bool, map[string]any, error) {
			return true, map[string]any{"userID": "123"}, nil
		}

		mw := mp.NewAuthMiddleware("Bearer", tokenCheckFunc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer valid-token")

		mw(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.False(t, c.IsAborted())
		userID, exists := c.Get("userID")
		assert.True(t, exists)
		assert.Equal(t, "123", userID)
	})

	t.Run("missing token", func(t *testing.T) {
		tokenCheckFunc := func(ctx *gin.Context, token string) (bool, map[string]any, error) {
			return true, nil, nil
		}

		mw := mp.NewAuthMiddleware("Bearer", tokenCheckFunc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		mw(c)

		assert.True(t, c.IsAborted())
		assert.NotEmpty(t, c.Errors)
	})

	t.Run("user not found", func(t *testing.T) {
		tokenCheckFunc := func(ctx *gin.Context, token string) (bool, map[string]any, error) {
			return false, nil, nil
		}

		mw := mp.NewAuthMiddleware("Bearer", tokenCheckFunc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer valid-token")

		mw(c)

		assert.True(t, c.IsAborted())
		assert.NotEmpty(t, c.Errors)
	})

	t.Run("check error", func(t *testing.T) {
		tokenCheckFunc := func(ctx *gin.Context, token string) (bool, map[string]any, error) {
			return false, nil, errors.New("db error")
		}

		mw := mp.NewAuthMiddleware("Bearer", tokenCheckFunc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer valid-token")

		mw(c)

		assert.True(t, c.IsAborted())
		assert.NotEmpty(t, c.Errors)
	})

	t.Run("unsupported strategy", func(t *testing.T) {
		tokenCheckFunc := func(ctx *gin.Context, token string) (bool, map[string]any, error) {
			return true, nil, nil
		}

		mw := mp.NewAuthMiddleware("Basic", tokenCheckFunc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer token")

		mw(c)

		assert.True(t, c.IsAborted())
		assert.NotEmpty(t, c.Errors)
	})
}
