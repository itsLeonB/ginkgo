package middleware

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ungerr"
	"github.com/rotisserie/eris"
)

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
			_ = ctx.Error(ungerr.ForbiddenError("no permission"))
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
