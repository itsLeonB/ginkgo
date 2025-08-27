package ginkgo_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ginkgo"
	"github.com/stretchr/testify/assert"
)

// Since extractToken, extractBearerToken, and validateAndExtractBearerToken are not exported,
// we need to test them through the public interfaces that use them (like the auth middleware)

func TestExtractTokenThroughAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		authStrategy string
		setupRequest func(*gin.Context)
		expectedAbort bool
	}{
		{
			name:         "Bearer strategy with valid token",
			authStrategy: "Bearer",
			setupRequest: func(ctx *gin.Context) {
				ctx.Request.Header.Set("Authorization", "Bearer valid_token_123")
			},
			expectedAbort: false,
		},
		{
			name:         "Bearer strategy with missing token",
			authStrategy: "Bearer",
			setupRequest: func(ctx *gin.Context) {
				// No Authorization header
			},
			expectedAbort: true,
		},
		{
			name:         "Bearer strategy with invalid format - wrong prefix",
			authStrategy: "Bearer",
			setupRequest: func(ctx *gin.Context) {
				ctx.Request.Header.Set("Authorization", "Basic token123")
			},
			expectedAbort: true,
		},
		{
			name:         "Bearer strategy with invalid format - no space",
			authStrategy: "Bearer",
			setupRequest: func(ctx *gin.Context) {
				ctx.Request.Header.Set("Authorization", "Bearertoken123")
			},
			expectedAbort: true,
		},
		{
			name:         "Bearer strategy with invalid format - too many parts",
			authStrategy: "Bearer",
			setupRequest: func(ctx *gin.Context) {
				ctx.Request.Header.Set("Authorization", "Bearer token part extra")
			},
			expectedAbort: true,
		},
		{
			name:         "Bearer strategy with invalid format - only Bearer",
			authStrategy: "Bearer",
			setupRequest: func(ctx *gin.Context) {
				ctx.Request.Header.Set("Authorization", "Bearer")
			},
			expectedAbort: true,
		},
		{
			name:         "Bearer strategy with case sensitive bearer",
			authStrategy: "Bearer",
			setupRequest: func(ctx *gin.Context) {
				ctx.Request.Header.Set("Authorization", "bearer token123")
			},
			expectedAbort: true,
		},
		{
			name:         "Bearer strategy with empty token",
			authStrategy: "Bearer",
			setupRequest: func(ctx *gin.Context) {
				ctx.Request.Header.Set("Authorization", "Bearer ")
			},
			expectedAbort: false, // Empty space after Bearer is actually valid - it's just an empty token
		},
		{
			name:         "unsupported auth strategy",
			authStrategy: "UnsupportedStrategy",
			setupRequest: func(ctx *gin.Context) {
				ctx.Request.Header.Set("Authorization", "Bearer valid_token")
			},
			expectedAbort: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenCheckFunc := func(ctx *gin.Context, token string) (bool, map[string]any, error) {
				// Always return true for valid tokens to test extraction logic
				return true, map[string]any{"user_id": "123"}, nil
			}

			mockLogger := &MockLogger{}
			provider := ginkgo.NewMiddlewareProvider(mockLogger)
			middleware := provider.NewAuthMiddleware(tt.authStrategy, tokenCheckFunc)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest("GET", "/", nil)

			tt.setupRequest(ctx)

			middleware(ctx)

			if tt.expectedAbort {
				assert.True(t, ctx.IsAborted())
				assert.NotEmpty(t, ctx.Errors)
			} else {
				assert.False(t, ctx.IsAborted())
				assert.Empty(t, ctx.Errors)
			}
		})
	}
}

func TestValidateAndExtractBearerTokenEdgeCases(t *testing.T) {
	// Test through auth middleware to cover edge cases
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		authHeader   string
		expectedAbort bool
	}{
		{
			name:         "valid JWT token",
			authHeader:   "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expectedAbort: false,
		},
		{
			name:         "token with special characters",
			authHeader:   "Bearer token-with_special.chars123",
			expectedAbort: false,
		},
		{
			name:         "very long token",
			authHeader:   "Bearer " + string(make([]byte, 1000)),
			expectedAbort: false,
		},
		{
			name:         "empty string",
			authHeader:   "",
			expectedAbort: true,
		},
		{
			name:         "only space",
			authHeader:   " ",
			expectedAbort: true,
		},
		{
			name:         "Bearer with multiple spaces",
			authHeader:   "Bearer  token",
			expectedAbort: true, // Should be invalid due to multiple spaces
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenCheckFunc := func(ctx *gin.Context, token string) (bool, map[string]any, error) {
				return true, map[string]any{"user_id": "123"}, nil
			}

			mockLogger := &MockLogger{}
			provider := ginkgo.NewMiddlewareProvider(mockLogger)
			middleware := provider.NewAuthMiddleware("Bearer", tokenCheckFunc)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest("GET", "/", nil)

			if tt.authHeader != "" {
				ctx.Request.Header.Set("Authorization", tt.authHeader)
			}

			middleware(ctx)

			if tt.expectedAbort {
				assert.True(t, ctx.IsAborted())
			} else {
				assert.False(t, ctx.IsAborted())
			}
		})
	}
}
