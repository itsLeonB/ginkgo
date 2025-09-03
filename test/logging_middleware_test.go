package ginkgo_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ginkgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMiddlewareProvider_NewLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		method       string
		path         string
		query        string
		statusCode   int
		hasError     bool
		errorMsg     string
		setupMock    func(*MockLogger)
	}{
		{
			name:       "successful GET request",
			method:     "GET",
			path:       "/api/users",
			query:      "",
			statusCode: 200,
			hasError:   false,
			setupMock: func(m *MockLogger) {
				m.On("Infof", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name:       "successful POST with query params",
			method:     "POST",
			path:       "/api/users",
			query:      "include=profile",
			statusCode: 201,
			hasError:   false,
			setupMock: func(m *MockLogger) {
				m.On("Infof", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name:       "client error without error message",
			method:     "GET",
			path:       "/api/users/999",
			query:      "",
			statusCode: 404,
			hasError:   false,
			setupMock: func(m *MockLogger) {
				m.On("Errorf", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name:       "server error with error message",
			method:     "POST",
			path:       "/api/users",
			query:      "",
			statusCode: 500,
			hasError:   true,
			errorMsg:   "database connection failed",
			setupMock: func(m *MockLogger) {
				m.On("Errorf", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name:       "OPTIONS request skipped",
			method:     "OPTIONS",
			path:       "/api/users",
			query:      "",
			statusCode: 200,
			hasError:   false,
			setupMock: func(m *MockLogger) {
				// No logging expected for OPTIONS
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			tt.setupMock(mockLogger)

			provider := ginkgo.NewMiddlewareProvider(mockLogger)
			middleware := provider.NewLoggingMiddleware()
			assert.NotNil(t, middleware)

			w := httptest.NewRecorder()
			ctx, engine := gin.CreateTestContext(w)
			
			url := tt.path
			if tt.query != "" {
				url += "?" + tt.query
			}
			ctx.Request = httptest.NewRequest(tt.method, url, nil)

			// Set up handler chain
			engine.Use(middleware)
			engine.Any("/*path", func(c *gin.Context) {
				c.Status(tt.statusCode)
				if tt.hasError {
					c.Error(errors.New(tt.errorMsg))
				}
			})

			engine.ServeHTTP(w, ctx.Request)

			mockLogger.AssertExpectations(t)
		})
	}
}
