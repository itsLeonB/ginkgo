package ginkgo

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (mp *MiddlewareProvider) NewLoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodOptions {
			ctx.Next()
			return
		}

		start := time.Now()
		path := ctx.Request.URL.Path
		method := ctx.Request.Method

		// Build full path with query string for logging
		fullPath := path
		if rawQuery := ctx.Request.URL.RawQuery; rawQuery != "" {
			fullPath = path + "?" + rawQuery
		}

		// Process request
		ctx.Next()

		// Calculate duration
		elapsed := time.Since(start)
		statusCode := ctx.Writer.Status()
		clientIP := ctx.ClientIP()

		// Log based on status code (similar to gRPC error handling)
		if statusCode >= 400 {
			errorMsg := ""
			if len(ctx.Errors) > 0 {
				errorMsg = ctx.Errors.String()
			}

			if errorMsg != "" {
				mp.logger.Errorf(
					"[HTTP] method=%s path=%s status=%d duration=%s client_ip=%s error=%s",
					method,
					fullPath,
					statusCode,
					elapsed,
					clientIP,
					errorMsg,
				)
			} else {
				mp.logger.Errorf(
					"[HTTP] method=%s path=%s status=%d duration=%s client_ip=%s",
					method,
					fullPath,
					statusCode,
					elapsed,
					clientIP,
				)
			}
		} else {
			mp.logger.Infof(
				"[HTTP] method=%s path=%s status=%d duration=%s client_ip=%s",
				method,
				fullPath,
				statusCode,
				elapsed,
				clientIP,
			)
		}
	}
}
