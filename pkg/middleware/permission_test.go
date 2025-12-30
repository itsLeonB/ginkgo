package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewPermissionMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := ezutil.NewSimpleLogger("test", true, 0)
	mp := NewMiddlewareProvider(logger)

	permissionMap := map[string][]string{
		"admin": {"read", "write"},
		"user":  {"read"},
	}

	// Corrected signature: NewPermissionMiddleware(roleContextKey, requiredPermission, permissionMap)
	mw := mp.NewPermissionMiddleware("role", "write", permissionMap)

	t.Run("has permission", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", nil)
		c.Set("role", "admin")

		mw(c)

		assert.False(t, c.IsAborted())
	})

	t.Run("no permission", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", nil)
		c.Set("role", "user")

		mw(c)

		assert.True(t, c.IsAborted())
		assert.NotEmpty(t, c.Errors)
	})

	t.Run("unknown role", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", nil)
		c.Set("role", "guest")

		mw(c)

		assert.True(t, c.IsAborted())
	})

	t.Run("missing role", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", nil)

		mw(c)

		assert.True(t, c.IsAborted())
	})
}
