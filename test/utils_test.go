package ginkgo_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/itsLeonB/ginkgo"
	"github.com/stretchr/testify/assert"
)

func TestGetPathParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		paramKey       string
		paramValue     string
		expectedValue  interface{}
		expectedExists bool
		expectedError  bool
		targetType     string
	}{
		{
			name:           "string parameter exists",
			paramKey:       "id",
			paramValue:     "test123",
			expectedValue:  "test123",
			expectedExists: true,
			expectedError:  false,
			targetType:     "string",
		},
		{
			name:           "int parameter exists",
			paramKey:       "id",
			paramValue:     "42",
			expectedValue:  42,
			expectedExists: true,
			expectedError:  false,
			targetType:     "int",
		},
		{
			name:           "bool parameter exists true",
			paramKey:       "active",
			paramValue:     "true",
			expectedValue:  true,
			expectedExists: true,
			expectedError:  false,
			targetType:     "bool",
		},
		{
			name:           "parameter does not exist",
			paramKey:       "nonexistent",
			paramValue:     "",
			expectedValue:  "",
			expectedExists: false,
			expectedError:  false,
			targetType:     "string",
		},
		{
			name:           "invalid int parameter",
			paramKey:       "id",
			paramValue:     "not_a_number",
			expectedValue:  0,
			expectedExists: false,
			expectedError:  true,
			targetType:     "int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			if tt.name != "parameter does not exist" {
				ctx.Params = gin.Params{{Key: tt.paramKey, Value: tt.paramValue}}
			}

			switch tt.targetType {
			case "string":
				value, exists, err := ginkgo.GetPathParam[string](ctx, tt.paramKey)
				if tt.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedValue, value)
				}
				assert.Equal(t, tt.expectedExists, exists)

			case "int":
				value, exists, err := ginkgo.GetPathParam[int](ctx, tt.paramKey)
				if tt.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedValue, value)
				}
				assert.Equal(t, tt.expectedExists, exists)

			case "bool":
				value, exists, err := ginkgo.GetPathParam[bool](ctx, tt.paramKey)
				if tt.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedValue, value)
				}
				assert.Equal(t, tt.expectedExists, exists)
			}
		})
	}
}

func TestGetRequiredPathParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		paramKey      string
		paramValue    string
		expectedValue interface{}
		expectedError bool
		targetType    string
	}{
		{
			name:          "string parameter exists",
			paramKey:      "id",
			paramValue:    "test123",
			expectedValue: "test123",
			expectedError: false,
			targetType:    "string",
		},
		{
			name:          "parameter missing",
			paramKey:      "missing",
			paramValue:    "",
			expectedValue: "",
			expectedError: true,
			targetType:    "string",
		},
		{
			name:          "invalid int parameter",
			paramKey:      "id",
			paramValue:    "not_a_number",
			expectedValue: 0,
			expectedError: true,
			targetType:    "int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			if tt.paramValue != "" {
				ctx.Params = gin.Params{{Key: tt.paramKey, Value: tt.paramValue}}
			}

			switch tt.targetType {
			case "string":
				value, err := ginkgo.GetRequiredPathParam[string](ctx, tt.paramKey)
				if tt.expectedError {
					assert.Error(t, err)
					if tt.paramValue == "" {
						assert.Contains(t, err.Error(), "missing path param")
					}
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedValue, value)
				}

			case "int":
				value, err := ginkgo.GetRequiredPathParam[int](ctx, tt.paramKey)
				if tt.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedValue, value)
				}
			}
		})
	}
}

func TestBindRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type TestStruct struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	tests := []struct {
		name          string
		requestBody   string
		contentType   string
		bindType      binding.Binding
		expectedValue TestStruct
		expectedError bool
	}{
		{
			name:        "valid JSON binding",
			requestBody: `{"name":"John","email":"john@example.com"}`,
			contentType: "application/json",
			bindType:    binding.JSON,
			expectedValue: TestStruct{
				Name:  "John",
				Email: "john@example.com",
			},
			expectedError: false,
		},
		{
			name:          "invalid JSON binding",
			requestBody:   `{"name":"John","email":}`,
			contentType:   "application/json",
			bindType:      binding.JSON,
			expectedValue: TestStruct{},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", tt.contentType)

			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			result, err := ginkgo.BindRequest[TestStruct](ctx, tt.bindType)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to bind request")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedValue, result)
			}
		})
	}
}

func TestGetFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		contextKey    string
		contextValue  interface{}
		expectedValue interface{}
		expectedError bool
		targetType    string
	}{
		{
			name:          "string value exists",
			contextKey:    "user_id",
			contextValue:  "12345",
			expectedValue: "12345",
			expectedError: false,
			targetType:    "string",
		},
		{
			name:          "key does not exist",
			contextKey:    "nonexistent",
			contextValue:  nil,
			expectedValue: "",
			expectedError: true,
			targetType:    "string",
		},
		{
			name:          "type assertion fails",
			contextKey:    "value",
			contextValue:  "string_value",
			expectedValue: 0,
			expectedError: true,
			targetType:    "int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			if tt.contextValue != nil {
				ctx.Set(tt.contextKey, tt.contextValue)
			}

			switch tt.targetType {
			case "string":
				value, err := ginkgo.GetFromContext[string](ctx, tt.contextKey)
				if tt.expectedError {
					assert.Error(t, err)
					if tt.contextValue == nil {
						assert.Contains(t, err.Error(), "not found in context")
					} else {
						assert.Contains(t, err.Error(), "error asserting value")
					}
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedValue, value)
				}

			case "int":
				value, err := ginkgo.GetFromContext[int](ctx, tt.contextKey)
				if tt.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedValue, value)
				}
			}
		})
	}
}

func TestGetAndParseFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		contextKey    string
		contextValue  interface{}
		expectedValue interface{}
		expectedError bool
		targetType    string
	}{
		{
			name:          "parse string to int",
			contextKey:    "number",
			contextValue:  "42",
			expectedValue: 42,
			expectedError: false,
			targetType:    "int",
		},
		{
			name:          "key does not exist",
			contextKey:    "nonexistent",
			contextValue:  nil,
			expectedValue: 0,
			expectedError: true,
			targetType:    "int",
		},
		{
			name:          "context value is not string",
			contextKey:    "number",
			contextValue:  42,
			expectedValue: 0,
			expectedError: true,
			targetType:    "int",
		},
		{
			name:          "invalid parse",
			contextKey:    "number",
			contextValue:  "not_a_number",
			expectedValue: 0,
			expectedError: true,
			targetType:    "int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			if tt.contextValue != nil {
				ctx.Set(tt.contextKey, tt.contextValue)
			}

			switch tt.targetType {
			case "int":
				value, err := ginkgo.GetAndParseFromContext[int](ctx, tt.contextKey)
				if tt.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expectedValue, value)
				}
			}
		})
	}
}
