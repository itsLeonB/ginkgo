package ginkgo_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/itsLeonB/ginkgo"
	"github.com/rotisserie/eris"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestErrorMiddleware_Handle_JSONSyntaxError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	middleware := ginkgo.NewErrorMiddleware(mockLogger)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	// Create a JSON syntax error
	jsonErr := &json.SyntaxError{Offset: 10}
	wrappedErr := eris.Wrap(jsonErr, "json parsing failed")
	_ = ctx.Error(wrappedErr)

	middleware(ctx)

	assert.True(t, ctx.IsAborted())
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestErrorMiddleware_Handle_JSONUnmarshalTypeError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	middleware := ginkgo.NewErrorMiddleware(mockLogger)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	// Create a JSON unmarshal type error
	jsonErr := &json.UnmarshalTypeError{Field: "age", Type: nil}
	wrappedErr := eris.Wrap(jsonErr, "json unmarshal failed")
	_ = ctx.Error(wrappedErr)

	middleware(ctx)

	assert.True(t, ctx.IsAborted())
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestErrorMiddleware_Handle_EOFError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	middleware := ginkgo.NewErrorMiddleware(mockLogger)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	// Create an EOF error
	wrappedErr := eris.Wrap(io.EOF, "request body error")
	_ = ctx.Error(wrappedErr)

	middleware(ctx)

	assert.True(t, ctx.IsAborted())
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestErrorMiddleware_Handle_EOFStringError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	middleware := ginkgo.NewErrorMiddleware(mockLogger)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	// Create an error with "EOF" string
	eofErr := errors.New("EOF")
	wrappedErr := eris.Wrap(eofErr, "request body error")
	_ = ctx.Error(wrappedErr)

	middleware(ctx)

	assert.True(t, ctx.IsAborted())
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestErrorMiddleware_Handle_ConnectionResetError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	middleware := ginkgo.NewErrorMiddleware(mockLogger)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	// Create a connection reset error
	connErr := errors.New("connection reset by peer")
	wrappedErr := eris.Wrap(connErr, "network error")
	_ = ctx.Error(wrappedErr)

	middleware(ctx)

	assert.True(t, ctx.IsAborted())
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestErrorMiddleware_Handle_BrokenPipeError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	middleware := ginkgo.NewErrorMiddleware(mockLogger)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	// Create a broken pipe error
	brokenPipeErr := errors.New("broken pipe")
	wrappedErr := eris.Wrap(brokenPipeErr, "network error")
	_ = ctx.Error(wrappedErr)

	middleware(ctx)

	assert.True(t, ctx.IsAborted())
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestErrorMiddleware_Handle_ValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	middleware := ginkgo.NewErrorMiddleware(mockLogger)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("GET", "/", nil)

	// Create validation errors - this is complex to mock, so we'll create a simple case
	// In practice, this would come from gin's binding validation
	validationErr := validator.ValidationErrors{}
	wrappedErr := eris.Wrap(validationErr, "validation failed")
	_ = ctx.Error(wrappedErr)

	middleware(ctx)

	assert.True(t, ctx.IsAborted())
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code) // Validation errors typically return 422
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestErrorMiddleware_constructAppError_Coverage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	mockLogger.On("Error", mock.AnythingOfType("string"))
	mockLogger.On("Errorf", mock.AnythingOfType("string"), mock.Anything)

	middleware := ginkgo.NewErrorMiddleware(mockLogger)

	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "JSON syntax error",
			err:            eris.Wrap(&json.SyntaxError{Offset: 5}, "json error"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "JSON unmarshal type error",
			err:            eris.Wrap(&json.UnmarshalTypeError{Field: "test"}, "unmarshal error"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "EOF error",
			err:            eris.Wrap(io.EOF, "eof error"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "EOF string error",
			err:            eris.Wrap(errors.New("EOF"), "eof string error"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "connection reset error",
			err:            eris.Wrap(errors.New("connection reset by peer"), "conn error"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "broken pipe error",
			err:            eris.Wrap(errors.New("broken pipe"), "pipe error"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unknown wrapped error",
			err:            eris.Wrap(errors.New("unknown error"), "wrapped unknown"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest("GET", "/", nil)

			_ = ctx.Error(tt.err)

			middleware(ctx)

			assert.True(t, ctx.IsAborted())
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}

	mockLogger.AssertExpectations(t)
}

func TestErrorMiddleware_logUnwrappedError_Coverage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLogger := &MockLogger{}
	// Set up expectations for all the logging calls
	mockLogger.On("Error", mock.AnythingOfType("string"))
	mockLogger.On("Errorf", mock.AnythingOfType("string"), mock.Anything)

	middleware := ginkgo.NewErrorMiddleware(mockLogger)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest("POST", "/test/path", nil)
	ctx.Request.Header.Set("User-Agent", "test-agent")
	ctx.Request.RemoteAddr = "127.0.0.1:12345"

	// Create an unwrapped error with metadata
	plainErr := errors.New("plain error with metadata")
	ginErr := ctx.Error(plainErr)
	ginErr.Meta = map[string]interface{}{"handler": "test_handler"}

	middleware(ctx)

	assert.True(t, ctx.IsAborted())
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockLogger.AssertExpectations(t)
}
