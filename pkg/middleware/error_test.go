package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ungerr"
	"github.com/stretchr/testify/assert"
)

func TestNewErrorMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := ezutil.NewSimpleLogger("test", true, 0)
	mp := NewMiddlewareProvider(logger)
	mw := mp.NewErrorMiddleware()

	t.Run("app error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		// Setup handlers properly to simulate middleware chain
		c.Request = httptest.NewRequest("GET", "/", nil)

		// We manually execute the middleware logic sequence since gin.CreateTestContext doesn't easily chaining without router
		// Better approach: use a router
		r := gin.New()
		r.Use(mw)
		r.GET("/", func(c *gin.Context) {
			_ = c.Error(ungerr.NotFoundError("resource not found"))
		})

		r.ServeHTTP(w, c.Request)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Not Found")
	})

	t.Run("raw error conversion", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		r := gin.New()
		r.Use(mw)
		r.GET("/", func(c *gin.Context) {
			_ = c.Error(errors.New("something broke"))
		})

		r.ServeHTTP(w, c.Request)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal Server Error")
	})

	t.Run("panic recovery", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		r := gin.New()
		r.Use(mw)
		r.GET("/", func(c *gin.Context) {
			panic("oops")
		})

		r.ServeHTTP(w, c.Request)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal Server Error")
	})
}
