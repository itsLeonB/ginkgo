package ginkgo_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/itsLeonB/ginkgo"
	"github.com/stretchr/testify/assert"
)

func TestQueryOptions(t *testing.T) {
	tests := []struct {
		name  string
		page  int
		limit int
	}{
		{
			name:  "valid query options",
			page:  1,
			limit: 10,
		},
		{
			name:  "valid query options with large values",
			page:  100,
			limit: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryOptions := ginkgo.QueryOptions{
				Page:  tt.page,
				Limit: tt.limit,
			}

			assert.Equal(t, tt.page, queryOptions.Page)
			assert.Equal(t, tt.limit, queryOptions.Limit)

			// Test JSON marshaling/unmarshaling
			jsonData, err := json.Marshal(queryOptions)
			assert.NoError(t, err)

			var unmarshaled ginkgo.QueryOptions
			err = json.Unmarshal(jsonData, &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, queryOptions, unmarshaled)
		})
	}
}

func TestPagination_IsZero(t *testing.T) {
	tests := []struct {
		name       string
		pagination ginkgo.Pagination
		expected   bool
	}{
		{
			name:       "zero pagination",
			pagination: ginkgo.Pagination{},
			expected:   true,
		},
		{
			name: "non-zero total data",
			pagination: ginkgo.Pagination{
				TotalData: 1,
			},
			expected: false,
		},
		{
			name: "non-zero current page",
			pagination: ginkgo.Pagination{
				CurrentPage: 1,
			},
			expected: false,
		},
		{
			name: "has next page true",
			pagination: ginkgo.Pagination{
				HasNextPage: true,
			},
			expected: false,
		},
		{
			name: "fully populated pagination",
			pagination: ginkgo.Pagination{
				TotalData:   100,
				CurrentPage: 2,
				TotalPages:  10,
				HasNextPage: true,
				HasPrevPage: true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pagination.IsZero()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewResponse(t *testing.T) {
	tests := []struct {
		name            string
		message         string
		expectedMessage string
	}{
		{
			name:            "simple message",
			message:         "Success",
			expectedMessage: "Success",
		},
		{
			name:            "empty message",
			message:         "",
			expectedMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := ginkgo.NewResponse(tt.message)

			assert.Equal(t, tt.expectedMessage, response.Message)
			assert.Nil(t, response.Data)
			assert.Nil(t, response.Errors)
			assert.True(t, response.Pagination.IsZero())
		})
	}
}

func TestNewErrorResponse(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedMsg string
	}{
		{
			name:        "simple error",
			err:         errors.New("something went wrong"),
			expectedMsg: "something went wrong",
		},
		{
			name:        "empty error message",
			err:         errors.New(""),
			expectedMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := ginkgo.NewErrorResponse(tt.err)

			assert.Equal(t, tt.expectedMsg, response.Message)
			assert.Nil(t, response.Data)
			assert.Equal(t, tt.err, response.Errors)
			assert.True(t, response.Pagination.IsZero())
		})
	}
}

func TestJSONResponse_WithData(t *testing.T) {
	tests := []struct {
		name         string
		initialMsg   string
		data         interface{}
		expectedData interface{}
	}{
		{
			name:         "string data",
			initialMsg:   "Success",
			data:         "test data",
			expectedData: "test data",
		},
		{
			name:       "struct data",
			initialMsg: "Success",
			data: map[string]interface{}{
				"id":   1,
				"name": "John",
			},
			expectedData: map[string]interface{}{
				"id":   1,
				"name": "John",
			},
		},
		{
			name:         "nil data",
			initialMsg:   "Success",
			data:         nil,
			expectedData: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := ginkgo.NewResponse(tt.initialMsg).WithData(tt.data)

			assert.Equal(t, tt.initialMsg, response.Message)
			assert.Equal(t, tt.expectedData, response.Data)
			assert.Nil(t, response.Errors)
			assert.True(t, response.Pagination.IsZero())
		})
	}
}

func TestJSONResponse_WithError(t *testing.T) {
	tests := []struct {
		name        string
		initialMsg  string
		err         error
		expectedErr error
	}{
		{
			name:        "simple error",
			initialMsg:  "Failed",
			err:         errors.New("validation failed"),
			expectedErr: errors.New("validation failed"),
		},
		{
			name:        "nil error",
			initialMsg:  "Success",
			err:         nil,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := ginkgo.NewResponse(tt.initialMsg).WithError(tt.err)

			assert.Equal(t, tt.initialMsg, response.Message)
			assert.Nil(t, response.Data)
			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr.Error(), response.Errors.Error())
			} else {
				assert.Nil(t, response.Errors)
			}
			assert.True(t, response.Pagination.IsZero())
		})
	}
}

func TestJSONResponse_WithPagination(t *testing.T) {
	tests := []struct {
		name               string
		queryOptions       ginkgo.QueryOptions
		totalData          int
		expectedPagination ginkgo.Pagination
	}{
		{
			name: "first page with next page",
			queryOptions: ginkgo.QueryOptions{
				Page:  1,
				Limit: 10,
			},
			totalData: 25,
			expectedPagination: ginkgo.Pagination{
				TotalData:   25,
				CurrentPage: 1,
				TotalPages:  3,
				HasNextPage: true,
				HasPrevPage: false,
			},
		},
		{
			name: "middle page",
			queryOptions: ginkgo.QueryOptions{
				Page:  2,
				Limit: 10,
			},
			totalData: 25,
			expectedPagination: ginkgo.Pagination{
				TotalData:   25,
				CurrentPage: 2,
				TotalPages:  3,
				HasNextPage: true,
				HasPrevPage: true,
			},
		},
		{
			name: "last page",
			queryOptions: ginkgo.QueryOptions{
				Page:  3,
				Limit: 10,
			},
			totalData: 25,
			expectedPagination: ginkgo.Pagination{
				TotalData:   25,
				CurrentPage: 3,
				TotalPages:  3,
				HasNextPage: false,
				HasPrevPage: true,
			},
		},
		{
			name: "no data",
			queryOptions: ginkgo.QueryOptions{
				Page:  1,
				Limit: 10,
			},
			totalData: 0,
			expectedPagination: ginkgo.Pagination{
				TotalData:   0,
				CurrentPage: 1,
				TotalPages:  0,
				HasNextPage: false,
				HasPrevPage: false,
			},
		},
		{
			name: "zero limit should return original response",
			queryOptions: ginkgo.QueryOptions{
				Page:  1,
				Limit: 0,
			},
			totalData: 25,
			expectedPagination: ginkgo.Pagination{}, // Should remain zero
		},
		{
			name: "negative limit should return original response",
			queryOptions: ginkgo.QueryOptions{
				Page:  1,
				Limit: -1,
			},
			totalData: 25,
			expectedPagination: ginkgo.Pagination{}, // Should remain zero
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := ginkgo.NewResponse("Success").WithPagination(tt.queryOptions, tt.totalData)

			assert.Equal(t, "Success", response.Message)
			assert.Nil(t, response.Data)
			assert.Nil(t, response.Errors)
			assert.Equal(t, tt.expectedPagination, response.Pagination)
			
			if tt.expectedPagination.IsZero() {
				assert.True(t, response.Pagination.IsZero())
			} else {
				assert.False(t, response.Pagination.IsZero())
			}
		})
	}
}

func TestJSONResponse_ChainedMethods(t *testing.T) {
	testData := map[string]interface{}{
		"users": []string{"user1", "user2"},
	}
	testError := errors.New("partial error")
	queryOptions := ginkgo.QueryOptions{Page: 1, Limit: 10}
	totalData := 25

	response := ginkgo.NewResponse("Success").
		WithData(testData).
		WithError(testError).
		WithPagination(queryOptions, totalData)

	assert.Equal(t, "Success", response.Message)
	assert.Equal(t, testData, response.Data)
	assert.Equal(t, testError, response.Errors)

	expectedPagination := ginkgo.Pagination{
		TotalData:   25,
		CurrentPage: 1,
		TotalPages:  3,
		HasNextPage: true,
		HasPrevPage: false,
	}
	assert.Equal(t, expectedPagination, response.Pagination)
}
