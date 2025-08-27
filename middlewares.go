package ginkgo

import (
	"log"
	"slices"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ungerr"
	"github.com/rotisserie/eris"
)

type MiddlewareProvider struct {
	logger ezutil.Logger
}

func NewMiddlewareProvider(logger ezutil.Logger) *MiddlewareProvider {
	if logger == nil {
		log.Fatal("logger cannot be nil")
	}
	return &MiddlewareProvider{logger}
}

func (mp *MiddlewareProvider) NewErrorMiddleware() gin.HandlerFunc {
	return newErrorMiddleware(mp.logger)
}

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

// NewAuthMiddleware creates an authentication middleware for Gin.
// It extracts a token using the given strategy (e.g., "header" or "cookie") via internal.ExtractToken,
// calls tokenCheckFunc to validate the token and retrieve user data,
// stores user data in the Gin context, and aborts the request on errors.
// Returns a Gin HandlerFunc for authentication handling.
func (mp *MiddlewareProvider) NewAuthMiddleware(
	authStrategy string,
	tokenCheckFunc func(ctx *gin.Context, token string) (bool, map[string]any, error),
) gin.HandlerFunc {
	if tokenCheckFunc == nil {
		mp.logger.Fatalf("tokenCheckFunc cannot be nil")
	}

	return func(ctx *gin.Context) {
		token, errMsg, err := extractToken(ctx, authStrategy)
		if err != nil {
			_ = ctx.Error(eris.Wrap(err, "error extracting token"))
			ctx.Abort()
			return
		}
		if errMsg != "" {
			_ = ctx.Error(ungerr.UnauthorizedError(errMsg))
			ctx.Abort()
			return
		}

		exists, data, err := tokenCheckFunc(ctx, token)
		if err != nil {
			_ = ctx.Error(err)
			ctx.Abort()
			return
		}
		if !exists {
			_ = ctx.Error(ungerr.UnauthorizedError(ErrUserNotFound))
			ctx.Abort()
			return
		}

		for key, val := range data {
			ctx.Set(key, val)
		}

		ctx.Next()
	}
}

// NewPermissionMiddleware creates a permission-checking middleware for Gin.
// It retrieves the user role from context using the provided roleContextKey,
// checks if the role exists in permissionMap and includes the requiredPermission,
// and aborts the request with a ForbiddenError if permission is missing.
// Returns a Gin HandlerFunc for permission enforcement.
func (mp *MiddlewareProvider) NewPermissionMiddleware(
	roleContextKey string,
	requiredPermission string,
	permissionMap map[string][]string,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role := ctx.GetString(roleContextKey)
		if role == "" {
			_ = ctx.Error(eris.Errorf("role not found in context or invalid type"))
			ctx.Abort()
			return
		}

		permissions, ok := permissionMap[role]
		if !ok {
			_ = ctx.Error(eris.Errorf("unknown role: %s", role))
			ctx.Abort()
			return
		}

		if !slices.Contains(permissions, requiredPermission) {
			_ = ctx.Error(ungerr.ForbiddenError(ErrNoPermission))
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
