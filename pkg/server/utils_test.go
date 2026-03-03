package server_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/itsLeonB/ginkgo/pkg/server"
	"github.com/stretchr/testify/assert"
)

func TestGetPathParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid param", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "123"}}

		val, exists, err := server.GetPathParam[int](c, "id")
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, 123, val)
	})

	t.Run("missing param", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		val, exists, err := server.GetPathParam[int](c, "id")
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.Equal(t, 0, val)
	})

	t.Run("invalid type", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "abc"}}

		_, exists, err := server.GetPathParam[int](c, "id")
		assert.Error(t, err)
		assert.True(t, exists)
	})
}

func TestGetRequiredPathParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid param", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "123"}}

		val, err := server.GetRequiredPathParam[int](c, "id")
		assert.NoError(t, err)
		assert.Equal(t, 123, val)
	})

	t.Run("missing param", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		_, err := server.GetRequiredPathParam[int](c, "id")
		assert.Error(t, err)
	})
}

func TestBindJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type TestStruct struct {
		Name string `json:"name"`
	}

	t.Run("valid json", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(`{"name":"test"}`))

		val, err := server.BindJSON[TestStruct](c)
		assert.NoError(t, err)
		assert.Equal(t, "test", val.Name)
	})

	t.Run("invalid json", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(`invalid`))

		_, err := server.BindJSON[TestStruct](c)
		assert.Error(t, err)
	})
}

func TestGetFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid value", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", 123)

		val, err := server.GetFromContext[int](c, "userID")
		assert.NoError(t, err)
		assert.Equal(t, 123, val)
	})

	t.Run("missing value", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		_, err := server.GetFromContext[int](c, "userID")
		assert.Error(t, err)
	})

	t.Run("invalid type", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", "123")

		_, err := server.GetFromContext[int](c, "userID")
		assert.Error(t, err)
	})
}

func TestBindRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type TestStruct struct {
		Name string `json:"name"`
	}

	t.Run("valid json binding", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(`{"name":"test"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		val, err := server.BindRequest[TestStruct](c, binding.JSON)
		assert.NoError(t, err)
		assert.Equal(t, "test", val.Name)
	})

	t.Run("invalid binding", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(`invalid`))

		_, err := server.BindRequest[TestStruct](c, binding.JSON)
		assert.Error(t, err)
	})
}

func TestGetAndParseFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid parse", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("count", "42")

		val, err := server.GetAndParseFromContext[int](c, "count")
		assert.NoError(t, err)
		assert.Equal(t, 42, val)
	})

	t.Run("missing key", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		_, err := server.GetAndParseFromContext[int](c, "count")
		assert.Error(t, err)
	})

	t.Run("invalid parse", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("count", "invalid")

		_, err := server.GetAndParseFromContext[int](c, "count")
		assert.Error(t, err)
	})
}

func TestHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/success", nil)

		handler := server.Handler("TestHandler.success", 200, func(ctx *gin.Context) (any, error) {
			return map[string]string{"message": "success"}, nil
		})

		handler(c)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/error", nil)

		handler := server.Handler("TestHandler.error", 200, func(ctx *gin.Context) (any, error) {
			return nil, assert.AnError
		})

		handler(c)
		assert.Len(t, c.Errors, 1)
	})
}
