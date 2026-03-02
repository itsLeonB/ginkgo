package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/itsLeonB/ezutil/v2/simple"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	logger := simple.NewLogger("test", true, 0)

	t.Run("success", func(t *testing.T) {
		srv := &http.Server{}
		timeout := 5 * time.Second
		shutdownFunc := func() error { return nil }

		s := New(srv, timeout, logger, shutdownFunc)

		assert.NotNil(t, s)
		assert.Equal(t, srv, s.srv)
		assert.Equal(t, timeout, s.timeout)
		assert.NotNil(t, s.shutdownFunc)
	})

	t.Run("nil shutdown func", func(t *testing.T) {
		srv := &http.Server{}
		timeout := 5 * time.Second

		s := New(srv, timeout, logger, nil)

		assert.NotNil(t, s)
		assert.Nil(t, s.shutdownFunc)
	})

	// Note: We cannot easily test fatal exits in unit tests without causing the test runner to exit.
	// Typically, we would refactor the constructor to return error or use a mock logger that captures Fatal calls.
	// Since the current implementation calls log.Fatal or logger.Fatal directly, we skip those negative test cases here
	// or would need to run them in a subprocess.
}
