package ginkgo

import (
	"crypto/subtle"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rotisserie/eris"
)

func extractToken(ctx *gin.Context, authStrategy string) (string, string, error) {
	switch authStrategy {
	case "Bearer":
		token, errMsg := extractBearerToken(ctx)
		return token, errMsg, nil
	default:
		return "", "", eris.Errorf("unsupported auth strategy: %s", authStrategy)
	}
}

func extractBearerToken(ctx *gin.Context) (string, string) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		return "", ErrMissingToken
	}

	isValid, token := validateAndExtractBearerToken(token)
	if !isValid {
		return "", ErrInvalidToken
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
