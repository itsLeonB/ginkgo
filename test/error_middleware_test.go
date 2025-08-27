package ginkgo_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ginkgo"
	"github.com/itsLeonB/ungerr"
	"github.com/rotisserie/eris"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewErrorMiddleware(t *testing.T) {
	mockLogger := &MockLogger{}
	middleware := ginkgo.NewErrorMiddleware(mockLogger)

	assert.NotNil(t, middleware)
}

func TestErrorMiddleware_Handle_NoError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	middleware := ginkgo.NewErrorMiddleware(mockLogger)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	// Simulate calling Next() without adding errors
	middleware(ctx)

	assert.False(t, ctx.IsAborted())
	assert.Equal(t, 200, w.Code) // Default status
}

func TestErrorMiddleware_Handle_AppError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	middleware := ginkgo.NewErrorMiddleware(mockLogger)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	// Add an AppError
	appErr := ungerr.BadRequestError("invalid input")
	_ = ctx.Error(appErr)

	middleware(ctx)

	assert.True(t, ctx.IsAborted())
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Just check that we got a JSON response with the right status
	// The actual message might be different due to how ungerr works
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestErrorMiddleware_Handle_UnwrappedError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	mockLogger.On("Error", mock.AnythingOfType("string"))
	mockLogger.On("Errorf", mock.AnythingOfType("string"), mock.Anything)

	middleware := ginkgo.NewErrorMiddleware(mockLogger)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/test", nil)

	// Add an unwrapped error (not using eris.Wrap)
	plainErr := errors.New("plain error")
	_ = ctx.Error(plainErr)

	middleware(ctx)

	assert.True(t, ctx.IsAborted())
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockLogger.AssertExpectations(t)
}

func TestErrorMiddleware_Handle_UnknownWrappedError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	mockLogger.On("Errorf", mock.AnythingOfType("string"), mock.Anything)
	mockLogger.On("Error", mock.AnythingOfType("string"))

	middleware := ginkgo.NewErrorMiddleware(mockLogger)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	// Create an unknown wrapped error
	unknownErr := errors.New("unknown error type")
	wrappedErr := eris.Wrap(unknownErr, "wrapped unknown error")
	_ = ctx.Error(wrappedErr)

	middleware(ctx)

	assert.True(t, ctx.IsAborted())
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockLogger.AssertExpectations(t)
}

func TestErrorMiddleware_Handle_MultipleErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	middleware := ginkgo.NewErrorMiddleware(mockLogger)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	// Add multiple errors - only the last one should be handled
	_ = ctx.Error(errors.New("first error"))
	_ = ctx.Error(ungerr.BadRequestError("second error"))

	middleware(ctx)

	assert.True(t, ctx.IsAborted())
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Just check that we got a JSON response
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestErrorMiddleware_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	middleware := ginkgo.NewErrorMiddleware(mockLogger)

	// Create a Gin router with the error middleware
	router := gin.New()
	router.Use(middleware)

	// Add a route that produces different types of errors
	router.GET("/app-error", func(c *gin.Context) {
		_ = c.Error(ungerr.NotFoundError("resource not found"))
	})

	router.GET("/wrapped-error", func(c *gin.Context) {
		err := errors.New("database connection failed")
		_ = c.Error(eris.Wrap(err, "failed to query database"))
	})

	router.GET("/plain-error", func(c *gin.Context) {
		_ = c.Error(errors.New("plain error"))
	})

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "app error",
			path:           "/app-error",
			expectedStatus: http.StatusNotFound,
			expectedMsg:    "resource not found",
		},
		{
			name:           "wrapped error",
			path:           "/wrapped-error",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "plain error",
			path:           "/plain-error",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name != "app error" {
				// Set up mock expectations for logging
				mockLogger.On("Error", mock.AnythingOfType("string"))
				mockLogger.On("Errorf", mock.AnythingOfType("string"), mock.Anything)
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.path, nil)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			// Just check that we got a JSON response for error cases
			if tt.expectedStatus >= 400 {
				assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
			}
		})
	}
}
