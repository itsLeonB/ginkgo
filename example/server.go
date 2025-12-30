package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ginkgo/pkg/middleware"
	"github.com/itsLeonB/ginkgo/pkg/server"
	"golang.org/x/time/rate"
)

func setup() *server.Http {
	r := gin.New()
	logger := ezutil.NewSimpleLogger("ginkgo-example", true, 0)
	timeout := 5 * time.Second

	mp := middleware.NewMiddlewareProvider(logger)

	r.Use(mp.NewErrorMiddleware())
	r.Use(mp.NewRateLimitMiddleware(rate.Every(time.Second), 5))

	r.GET("/success", handleSuccess())
	r.GET("/error", handleError())
	r.GET("/wrapped-error", handleWrappedError())
	r.GET("/unwrapped-error", handleUnwrappedError())
	r.GET("/app-error", handleAppError())
	r.GET("/known-error", handleKnownError())
	r.GET("/panic", handlePanic())

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadTimeout:       timeout,
		ReadHeaderTimeout: timeout,
		WriteTimeout:      timeout,
		IdleTimeout:       timeout,
	}

	return server.New(srv, timeout, logger, nil)
}
