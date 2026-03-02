package middleware

import (
	"testing"

	"github.com/itsLeonB/ezutil/v2/simple"
	"github.com/stretchr/testify/assert"
)

func TestNewMiddlewareProvider(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		logger := simple.NewLogger("test", true, 0)
		mp := NewMiddlewareProvider(logger)
		assert.NotNil(t, mp)
		assert.Equal(t, logger, mp.logger)
	})
}

func TestNewErrorMiddlewareFromProvider(t *testing.T) {
	logger := simple.NewLogger("test", true, 0)
	mp := NewMiddlewareProvider(logger)

	middleware := mp.NewErrorMiddleware()
	assert.NotNil(t, middleware)
}
