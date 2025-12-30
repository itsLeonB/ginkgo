package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// NewCorsMiddleware creates a CORS middleware for Gin with the provided configuration.
// If corsConfig is nil, default settings are used (via cors.Default()).
// The middleware validates the configuration and logs a fatal error if invalid.
// Returns a Gin HandlerFunc to handle CORS according to the specified config.
func (mp *MiddlewareProvider) NewCorsMiddleware(corsConfig *cors.Config) gin.HandlerFunc {
	if corsConfig == nil {
		mp.logger.Warn("CORS configuration is nil, using default settings")
		return cors.Default()
	}

	if err := corsConfig.Validate(); err != nil {
		mp.logger.Fatalf("invalid CORS configuration: %s", err.Error())
	}

	return cors.New(*corsConfig)
}
