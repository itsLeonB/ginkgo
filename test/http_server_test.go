package ginkgo_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ginkgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogger implements ezutil.Logger for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Debugf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Info(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Warn(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Warnf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Error(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Fatal(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Fatalf(format string, args ...interface{}) {
	m.Called(format, args)
}

func TestNewHttpServer(t *testing.T) {
	tests := []struct {
		name         string
		srv          *http.Server
		timeout      time.Duration
		logger       ezutil.Logger
		shutdownFunc func() error
		skipTest     bool
	}{
		{
			name: "valid parameters",
			srv: &http.Server{
				Addr: ":8080",
			},
			timeout: 5 * time.Second,
			logger:  &MockLogger{},
			shutdownFunc: func() error {
				return nil
			},
			skipTest: false,
		},
		{
			name: "valid parameters with nil shutdown func",
			srv: &http.Server{
				Addr: ":8080",
			},
			timeout:      5 * time.Second,
			logger:       &MockLogger{},
			shutdownFunc: nil,
			skipTest:     false,
		},
		{
			name:         "nil server",
			srv:          nil,
			timeout:      5 * time.Second,
			logger:       &MockLogger{},
			shutdownFunc: func() error { return nil },
			skipTest:     true, // Skip because log.Fatal terminates the program
		},
		{
			name: "zero timeout",
			srv: &http.Server{
				Addr: ":8080",
			},
			timeout:      0,
			logger:       &MockLogger{},
			shutdownFunc: func() error { return nil },
			skipTest:     true, // Skip because log.Fatal terminates the program
		},
		{
			name: "nil logger",
			srv: &http.Server{
				Addr: ":8080",
			},
			timeout:      5 * time.Second,
			logger:       nil,
			shutdownFunc: func() error { return nil },
			skipTest:     true, // Skip because log.Fatal terminates the program
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipTest {
				t.Skip("Skipping test that would cause log.Fatal")
				return
			}

			mockLogger := tt.logger.(*MockLogger)
			if tt.shutdownFunc == nil {
				mockLogger.On("Warn", mock.AnythingOfType("string"))
			}

			httpServer := ginkgo.NewHttpServer(tt.srv, tt.timeout, tt.logger, tt.shutdownFunc)

			assert.NotNil(t, httpServer)

			if tt.shutdownFunc == nil {
				mockLogger.AssertExpectations(t)
			}
		})
	}
}
