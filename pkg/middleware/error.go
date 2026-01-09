package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ginkgo/pkg/response"
	"github.com/itsLeonB/ungerr"
)

// errorMiddleware handles errors and panics in Gin handlers
type errorMiddleware struct {
	logger ezutil.Logger
}

type errorObject struct {
	Code   string `json:"code"`
	Detail any    `json:"detail"`
}

func (eo errorObject) Error() string {
	return fmt.Sprintf("%s: %s", eo.Code, eo.Detail)
}

// NewErrorMiddleware creates an error handling middleware for Gin.
// It should be registered first (outermost) so it can capture errors/panics
// from all subsequent middlewares and handlers, even if they abort.
// This converts them into AppError or validation errors, and sends a structured JSON response
// with the appropriate HTTP status code. Returns a Gin HandlerFunc.
func newErrorMiddleware(logger ezutil.Logger) gin.HandlerFunc {
	middleware := &errorMiddleware{
		logger: logger,
	}
	return middleware.handle
}

func appErrorToErrorObject(appError ungerr.AppError) any {
	return response.NewErrorResponse(errorObject{
		Code:   appError.Error(),
		Detail: appError.Details(),
	})
}

// Handle is the main middleware function that processes errors and panics
func (em *errorMiddleware) handle(ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			em.handlePanic(r, ctx)
		}
	}()

	ctx.Next()

	if err := ctx.Errors.Last(); err != nil {
		// Check if it's already an AppError
		if appError, ok := err.Err.(ungerr.AppError); ok {
			ctx.AbortWithStatusJSON(appError.HttpStatus(), appErrorToErrorObject(appError))
			return
		}

		// Handle other errors
		appError := em.constructAppError(err, ctx)
		ctx.AbortWithStatusJSON(appError.HttpStatus(), appErrorToErrorObject(appError))
	}
}

// constructAppError converts various error types into AppError
func (em *errorMiddleware) constructAppError(err *gin.Error, ctx *gin.Context) ungerr.AppError {
	unknownErr, ok := err.Err.(*ungerr.UnknownError)
	if ok {
		if originalErr := ungerr.Unwrap(err.Err); originalErr != nil {
			em.logger.Errorf("unhandled error of type %T: %s", originalErr, err.Err.Error())
			return ungerr.InternalServerError()
		}

		em.logger.Error("unexpected error:", unknownErr.Error())
		return ungerr.InternalServerError()
	}

	// Handle known error types from eris-wrapped errors
	switch originalErr := err.Err.(type) {
	case validator.ValidationErrors:
		var errors []string
		for _, e := range originalErr {
			errors = append(errors, e.Error())
		}
		return ungerr.ValidationError(errors)

	case *json.SyntaxError:
		return ungerr.BadRequestError("invalid json")

	case *json.UnmarshalTypeError:
		return ungerr.BadRequestError(fmt.Sprintf("invalid value for field %s", originalErr.Field))

	default:
		// Handle common error patterns
		errStr := originalErr.Error()

		// EOF error from json package is unexported
		if originalErr == io.EOF || errStr == "EOF" {
			return ungerr.BadRequestError("missing request body")
		}

		// Check for network-related errors that might be client errors
		if strings.Contains(errStr, "connection reset by peer") ||
			strings.Contains(errStr, "broken pipe") {
			return ungerr.BadRequestError("connection error")
		}

		return em.logUnwrappedError(err, ctx)
	}
}

// logUnwrappedError handles errors that weren't properly wrapped with eris
func (em *errorMiddleware) logUnwrappedError(err *gin.Error, ctx *gin.Context) ungerr.AppError {
	// This function helps you identify where errors are being added without proper wrapping
	em.logger.Errorf(
		"UNWRAPPED ERROR DETECTED - Please wrap with ungerr.Wrap() or return ungerr.AppError\n"+
			"Error type: %T\n"+
			"Error message: %s\n"+
			"Request: %s %s\n"+
			"Handler: %s\n"+
			"Details: %+v",
		err.Err,
		err.Err.Error(),
		ctx.Request.Method,
		ctx.Request.URL.Path,
		ctx.HandlerName(),
		err.Err,
	)

	// Return a masked error to the user
	return ungerr.InternalServerError()
}

// handlePanic recovers from panics and converts them to structured errors
func (em *errorMiddleware) handlePanic(r interface{}, ctx *gin.Context) {
	// Build panic analysis
	var analysisBuilder strings.Builder
	switch panicValue := r.(type) {
	case string:
		if strings.Contains(panicValue, "index out of range") ||
			strings.Contains(panicValue, "slice bounds out of range") {
			analysisBuilder.WriteString("Array/slice bounds panic detected")
		} else if strings.Contains(panicValue, "nil pointer dereference") {
			analysisBuilder.WriteString("Nil pointer dereference panic detected")
		} else {
			fmt.Fprintf(&analysisBuilder, "String panic: %s", panicValue)
		}

	case runtime.Error:
		fmt.Fprintf(&analysisBuilder, "Runtime error panic: %v", panicValue)
		switch panicValue.Error() {
		case "runtime error: invalid memory address or nil pointer dereference":
			analysisBuilder.WriteString(" - Nil pointer dereference detected")
		case "runtime error: index out of range":
			analysisBuilder.WriteString(" - Index out of range detected")
		case "runtime error: slice bounds out of range":
			analysisBuilder.WriteString(" - Slice bounds out of range detected")
		}

	default:
		fmt.Fprintf(&analysisBuilder, "Unknown panic type: %T, value: %v", r, r)
	}

	// Log everything in one call with stack trace
	em.logger.Errorf(
		"PANIC RECOVERED in handler\n"+
			"Request: %s %s\n"+
			"Handler: %s\n"+
			"Panic value: %v\n"+
			"Panic type: %T\n"+
			"Analysis: %s\n"+
			"Stack trace:\n%s",
		ctx.Request.Method,
		ctx.Request.URL.Path,
		ctx.HandlerName(),
		r,
		r,
		analysisBuilder.String(),
		string(debug.Stack()),
	)

	// If the response hasn't been written yet, send error response
	if !ctx.Writer.Written() {
		appError := ungerr.InternalServerError()
		ctx.AbortWithStatusJSON(appError.HttpStatus(), appErrorToErrorObject(appError))
	} else {
		em.logger.Errorf("Response already written, cannot send error JSON. Status: %d", ctx.Writer.Status())
	}
}
