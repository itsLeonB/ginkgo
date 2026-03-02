package middleware

import (
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/itsLeonB/ezutil/v2/simple"
	"github.com/stretchr/testify/assert"
)

func TestNewCorsMiddleware(t *testing.T) {
	logger := simple.NewLogger("test", true, 0)
	mp := NewMiddlewareProvider(logger)

	t.Run("nil config uses default", func(t *testing.T) {
		middleware := mp.NewCorsMiddleware(nil)
		assert.NotNil(t, middleware)
	})

	t.Run("valid config", func(t *testing.T) {
		config := &cors.Config{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST"},
		}
		middleware := mp.NewCorsMiddleware(config)
		assert.NotNil(t, middleware)
	})
}
