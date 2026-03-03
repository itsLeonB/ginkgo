package main

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ginkgo/pkg/server"
	"github.com/itsLeonB/ungerr"
)

func handleSuccess() gin.HandlerFunc {
	return server.Handler("handleSuccess", http.StatusOK, func(ctx *gin.Context) (any, error) {
		return "success", nil
	})
}

func handleError() gin.HandlerFunc {
	return server.Handler("handleError", http.StatusOK, func(ctx *gin.Context) (any, error) {
		return nil, ungerr.Unknown("this error should be handled as InternalServerError")
	})
}

func handleWrappedError() gin.HandlerFunc {
	return server.Handler("handleWrappedError", http.StatusOK, func(ctx *gin.Context) (any, error) {
		return nil, ungerr.Wrap(http.ErrNoCookie, "no cookie")
	})
}

func handleUnwrappedError() gin.HandlerFunc {
	return server.Handler("handleUnwrappedError", http.StatusOK, func(ctx *gin.Context) (any, error) {
		return nil, http.ErrNoCookie
	})
}

func handleAppError() gin.HandlerFunc {
	return server.Handler("handleAppError", http.StatusOK, func(ctx *gin.Context) (any, error) {
		return nil, ungerr.BadRequestError("error should be returned")
	})
}

func handleKnownError() gin.HandlerFunc {
	return server.Handler("handleKnownError", http.StatusOK, func(ctx *gin.Context) (any, error) {
		return nil, &json.SyntaxError{}
	})
}

func handlePanic() gin.HandlerFunc {
	return server.Handler("handlePanic", http.StatusOK, func(ctx *gin.Context) (any, error) {
		panic("panicking")
	})
}
