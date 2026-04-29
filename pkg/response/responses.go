package response

import "math"

// QueryOptions represents common pagination query parameters for HTTP requests.
type QueryOptions struct {
	Page  int `form:"page" binding:"required,min=1"`
	Limit int `form:"limit" binding:"required,min=1"`
}

// Pagination contains metadata about paginated results.
type Pagination struct {
	TotalData   int  `json:"totalData"`
	CurrentPage int  `json:"currentPage"`
	TotalPages  int  `json:"totalPages"`
	HasNextPage bool `json:"hasNextPage"`
	HasPrevPage bool `json:"hasPrevPage"`
}

// IsZero checks if all pagination fields are at their zero values.
func (p Pagination) IsZero() bool {
	return p.TotalData == 0 && p.CurrentPage == 0 && p.TotalPages == 0 && !p.HasNextPage && !p.HasPrevPage
}

// JSONResponse represents a standardized HTTP JSON response structure.
type JSONResponse[T any] struct {
	Data       T          `json:"data,omitzero"`
	Errors     []error    `json:"errors,omitempty"`
	Pagination Pagination `json:"pagination,omitzero"`
}

// NewResponse creates a JSONResponse with the specified data.
func NewResponse[T any](data T) JSONResponse[T] {
	return JSONResponse[T]{Data: data}
}

// NewErrorResponse creates a JSONResponse for error cases.
func NewErrorResponse(err ...error) JSONResponse[any] {
	return JSONResponse[any]{Errors: err}
}

// WithPagination calculates and adds pagination metadata to the JSONResponse.
func (jr JSONResponse[T]) WithPagination(queryOptions QueryOptions, totalData int) JSONResponse[T] {
	if queryOptions.Limit <= 0 {
		return jr
	}

	totalPages := int(math.Ceil(float64(totalData) / float64(queryOptions.Limit)))

	jr.Pagination = Pagination{
		TotalData:   totalData,
		CurrentPage: queryOptions.Page,
		TotalPages:  totalPages,
		HasNextPage: queryOptions.Page < totalPages,
		HasPrevPage: queryOptions.Page > 1,
	}

	return jr
}
