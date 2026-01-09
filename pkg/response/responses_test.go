package response

import (
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success with data", func(t *testing.T) {
		data := gin.H{"foo": "bar"}
		resp := NewResponse(data)
		assert.Equal(t, data, resp.Data)
		assert.Empty(t, resp.Errors)
		assert.True(t, resp.Pagination.IsZero())
	})
}

func TestNewErrorResponse(t *testing.T) {
	err := errors.New("something went wrong")
	resp := NewErrorResponse(err)

	assert.Len(t, resp.Errors, 1)
	assert.Equal(t, err, resp.Errors[0])
	assert.Nil(t, resp.Data)
}

func TestPagination(t *testing.T) {
	t.Run("WithPagination", func(t *testing.T) {
		jr := NewResponse(nil)
		opts := QueryOptions{Page: 2, Limit: 10}
		totalData := 25

		jr = jr.WithPagination(opts, totalData)

		assert.False(t, jr.Pagination.IsZero())
		assert.Equal(t, totalData, jr.Pagination.TotalData)
		assert.Equal(t, 2, jr.Pagination.CurrentPage)
		assert.Equal(t, 3, jr.Pagination.TotalPages)
		assert.True(t, jr.Pagination.HasNextPage)
		assert.True(t, jr.Pagination.HasPrevPage)
	})

	t.Run("WithPagination zero limit", func(t *testing.T) {
		jr := NewResponse(nil)
		opts := QueryOptions{Page: 1, Limit: 0}

		jr = jr.WithPagination(opts, 100)

		assert.True(t, jr.Pagination.IsZero())
	})

	t.Run("WithPagination last page", func(t *testing.T) {
		jr := NewResponse(nil)
		opts := QueryOptions{Page: 3, Limit: 10}
		totalData := 25

		jr = jr.WithPagination(opts, totalData)

		assert.False(t, jr.Pagination.IsZero())
		assert.False(t, jr.Pagination.HasNextPage)
		assert.True(t, jr.Pagination.HasPrevPage)
	})

	t.Run("IsZero", func(t *testing.T) {
		p := Pagination{}
		assert.True(t, p.IsZero())

		p.TotalData = 1
		assert.False(t, p.IsZero())
	})
}
