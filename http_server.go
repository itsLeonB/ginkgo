package ginkgo

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/itsLeonB/ezutil"
)

type HttpServer struct {
	srv          *http.Server
	timeout      time.Duration
	logger       ezutil.Logger
	shutdownFunc func() error
}

func NewHttpServer(srv *http.Server, timeout time.Duration, logger ezutil.Logger, shutdownFunc func() error) *HttpServer {
	if srv == nil {
		log.Fatal("http.Server cannot be nil")
	}
	if timeout <= 0 {
		log.Fatal("timeout must be > 0")
	}
	if logger == nil {
		log.Fatal("logger cannot be nil")
	}
	if shutdownFunc == nil {
		logger.Warn("shutdownFunc is nil, continuing...")
	}

	return &HttpServer{srv, timeout, logger, shutdownFunc}
}

// ServeGracefully starts the HTTP server and handles graceful shutdown
func (hs *HttpServer) ServeGracefully() {
	go func() {
		hs.logger.Infof("starting server on: %s", hs.srv.Addr)
		if err := hs.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			hs.logger.Fatalf("error server listen and serve: %s", err.Error())
		}
	}()

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	<-exit
	hs.logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), hs.timeout)
	defer cancel()

	if err := hs.srv.Shutdown(ctx); err != nil {
		hs.logger.Fatalf("error shutting down: %s", err.Error())
	}

	if err := hs.shutdownFunc(); err != nil {
		hs.logger.Errorf("error in terminating resources: %s", err.Error())
	}

	hs.logger.Info("server successfully shutdown")
}
