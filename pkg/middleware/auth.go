package middleware

import (
	"crypto/subtle"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ungerr"
)

// NewAuthMiddleware creates an authentication middleware for Gin.
// It extracts a token using the given strategy (e.g., "Bearer") via extractToken,
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
			_ = ctx.Error(ungerr.Wrap(err, "error extracting token"))
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
			_ = ctx.Error(ungerr.UnauthorizedError("user data not found"))
			ctx.Abort()
			return
		}

		for key, val := range data {
			ctx.Set(key, val)
		}

		ctx.Next()
	}
}

func extractToken(ctx *gin.Context, authStrategy string) (string, string, error) {
	switch authStrategy {
	case "Bearer":
		token, errMsg := extractBearerToken(ctx)
		return token, errMsg, nil
	default:
		return "", "", ungerr.Unknownf("unsupported auth strategy: %s", authStrategy)
	}
}

func extractBearerToken(ctx *gin.Context) (string, string) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		return "", "missing token"
	}

	isValid, token := validateAndExtractBearerToken(token)
	if !isValid {
		return "", "invalid token"
	}

	return token, ""
}

func validateAndExtractBearerToken(bearerToken string) (bool, string) {
	splits := strings.Split(bearerToken, " ")

	if len(splits) != 2 {
		return false, ""
	}

	tokenType := strings.ToLower(splits[0])
	ok := subtle.ConstantTimeCompare([]byte(tokenType), []byte("bearer")) == 1
	if !ok {
		return false, ""
	}

	return true, splits[1]
}
