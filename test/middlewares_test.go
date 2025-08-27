package ginkgo_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ginkgo"
	"github.com/itsLeonB/ungerr"
	"github.com/stretchr/testify/assert"
)

func TestNewCorsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		corsConfig *cors.Config
		expectFunc bool
	}{
		{
			name:       "nil config uses default",
			corsConfig: nil,
			expectFunc: true,
		},
		{
			name: "valid config",
			corsConfig: &cors.Config{
				AllowOrigins: []string{"http://localhost:3000"},
				AllowMethods: []string{"GET", "POST"},
				AllowHeaders: []string{"Content-Type"},
			},
			expectFunc: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := ginkgo.NewCorsMiddleware(tt.corsConfig)
			assert.NotNil(t, middleware)

			// Test that the middleware function works
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest("OPTIONS", "/", nil)
			ctx.Request.Header.Set("Origin", "http://localhost:3000")

			middleware(ctx)
			// The middleware should set CORS headers
			assert.Contains(t, w.Header(), "Access-Control-Allow-Origin")
		})
	}
}

func TestNewAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authStrategy   string
		tokenCheckFunc func(ctx *gin.Context, token string) (bool, map[string]any, error)
		expectedAbort  bool
		setupRequest   func(*gin.Context)
	}{
		{
			name:         "successful authentication",
			authStrategy: "Bearer",
			tokenCheckFunc: func(ctx *gin.Context, token string) (bool, map[string]any, error) {
				if token == "valid_token" {
					return true, map[string]any{"user_id": "123", "role": "admin"}, nil
				}
				return false, nil, nil
			},
			expectedAbort: false,
			setupRequest: func(ctx *gin.Context) {
				ctx.Request.Header.Set("Authorization", "Bearer valid_token")
			},
		},
		{
			name:         "missing token",
			authStrategy: "Bearer",
			tokenCheckFunc: func(ctx *gin.Context, token string) (bool, map[string]any, error) {
				return false, nil, nil
			},
			expectedAbort: true,
			setupRequest: func(ctx *gin.Context) {
				// No Authorization header
			},
		},
		{
			name:         "invalid token format",
			authStrategy: "Bearer",
			tokenCheckFunc: func(ctx *gin.Context, token string) (bool, map[string]any, error) {
				return false, nil, nil
			},
			expectedAbort: true,
			setupRequest: func(ctx *gin.Context) {
				ctx.Request.Header.Set("Authorization", "InvalidFormat token")
			},
		},
		{
			name:         "token check returns false",
			authStrategy: "Bearer",
			tokenCheckFunc: func(ctx *gin.Context, token string) (bool, map[string]any, error) {
				return false, nil, nil
			},
			expectedAbort: true,
			setupRequest: func(ctx *gin.Context) {
				ctx.Request.Header.Set("Authorization", "Bearer invalid_token")
			},
		},
		{
			name:         "token check returns error",
			authStrategy: "Bearer",
			tokenCheckFunc: func(ctx *gin.Context, token string) (bool, map[string]any, error) {
				return false, nil, ungerr.InternalServerError()
			},
			expectedAbort: true,
			setupRequest: func(ctx *gin.Context) {
				ctx.Request.Header.Set("Authorization", "Bearer error_token")
			},
		},
		{
			name:         "unsupported auth strategy",
			authStrategy: "UnsupportedStrategy",
			tokenCheckFunc: func(ctx *gin.Context, token string) (bool, map[string]any, error) {
				return false, nil, nil
			},
			expectedAbort: true,
			setupRequest: func(ctx *gin.Context) {
				ctx.Request.Header.Set("Authorization", "Bearer valid_token")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := ginkgo.NewAuthMiddleware(tt.authStrategy, tt.tokenCheckFunc)
			assert.NotNil(t, middleware)

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
				// Check that context values were set
				userID, exists := ctx.Get("user_id")
				assert.True(t, exists)
				assert.Equal(t, "123", userID)
				role, exists := ctx.Get("role")
				assert.True(t, exists)
				assert.Equal(t, "admin", role)
			}
		})
	}
}

func TestNewPermissionMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	permissionMap := map[string][]string{
		"admin":  {"read", "write", "delete"},
		"editor": {"read", "write"},
		"viewer": {"read"},
	}

	tests := []struct {
		name               string
		roleContextKey     string
		requiredPermission string
		expectedAbort      bool
		setupContext       func(*gin.Context)
	}{
		{
			name:               "admin has delete permission",
			roleContextKey:     "role",
			requiredPermission: "delete",
			expectedAbort:      false,
			setupContext: func(ctx *gin.Context) {
				ctx.Set("role", "admin")
			},
		},
		{
			name:               "editor has write permission",
			roleContextKey:     "role",
			requiredPermission: "write",
			expectedAbort:      false,
			setupContext: func(ctx *gin.Context) {
				ctx.Set("role", "editor")
			},
		},
		{
			name:               "viewer does not have write permission",
			roleContextKey:     "role",
			requiredPermission: "write",
			expectedAbort:      true,
			setupContext: func(ctx *gin.Context) {
				ctx.Set("role", "viewer")
			},
		},
		{
			name:               "role not found in context",
			roleContextKey:     "role",
			requiredPermission: "read",
			expectedAbort:      true,
			setupContext: func(ctx *gin.Context) {
				// Don't set role in context
			},
		},
		{
			name:               "unknown role",
			roleContextKey:     "role",
			requiredPermission: "read",
			expectedAbort:      true,
			setupContext: func(ctx *gin.Context) {
				ctx.Set("role", "unknown")
			},
		},
		{
			name:               "role is not string type",
			roleContextKey:     "role",
			requiredPermission: "read",
			expectedAbort:      true,
			setupContext: func(ctx *gin.Context) {
				ctx.Set("role", 123) // Set non-string value
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := ginkgo.NewPermissionMiddleware(tt.roleContextKey, tt.requiredPermission, permissionMap)
			assert.NotNil(t, middleware)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest("GET", "/", nil)

			tt.setupContext(ctx)

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
