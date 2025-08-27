package ginkgo

import "math"

// QueryOptions represents common pagination query parameters for HTTP requests.
// It includes validation tags to ensure proper values for page and limit parameters.
type QueryOptions struct {
	Page  int `query:"page" binding:"required,min=1"`
	Limit int `query:"limit" binding:"required,min=1"`
}

// Pagination contains metadata about paginated results.
// It provides information about the current page, total pages, and navigation flags.
type Pagination struct {
	TotalData   int  `json:"totalData"`
	CurrentPage int  `json:"currentPage"`
	TotalPages  int  `json:"totalPages"`
	HasNextPage bool `json:"hasNextPage"`
	HasPrevPage bool `json:"hasPrevPage"`
}

// IsZero checks if all pagination fields are at their zero values.
// Returns true if the pagination data is uninitialized or empty.
func (p Pagination) IsZero() bool {
	return p.TotalData == 0 && p.CurrentPage == 0 && p.TotalPages == 0 && !p.HasNextPage && !p.HasPrevPage
}

// JSONResponse represents a standardized HTTP JSON response structure.
// It can include a message, data payload, error information, and pagination metadata.
type JSONResponse struct {
	Message    string     `json:"message"`
	Data       any        `json:"data,omitzero"`
	Errors     error      `json:"errors,omitempty"`
	Pagination Pagination `json:"pagination,omitzero"`
}

// NewResponse creates a basic JSONResponse with the specified message.
// Additional data, errors, or pagination can be added using the With* methods.
func NewResponse(message string) JSONResponse {
	return JSONResponse{
		Message: message,
	}
}

// NewErrorResponse creates a JSONResponse for error cases.
// It sets the message to the error text and populates the Errors field.
func NewErrorResponse(err error) JSONResponse {
	return JSONResponse{
		Message: err.Error(),
		Errors:  err,
	}
}

// WithData adds a data payload to the JSONResponse.
// Returns a new JSONResponse with the Data field populated.
func (jr JSONResponse) WithData(data any) JSONResponse {
	jr.Data = data
	return jr
}

// WithError adds error information to the JSONResponse.
// Returns a new JSONResponse with the Errors field populated.
func (jr JSONResponse) WithError(err error) JSONResponse {
	jr.Errors = err
	return jr
}

// WithPagination calculates and adds pagination metadata to the JSONResponse.
// It computes total pages and next/previous flags based on query options and total data count.
// Returns a new JSONResponse with pagination metadata included.
func (jr JSONResponse) WithPagination(queryOptions QueryOptions, totalData int) JSONResponse {
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
