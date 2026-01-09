package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ezutil/v2"
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
