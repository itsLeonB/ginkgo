package ginkgo

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
	"github.com/itsLeonB/ungerr"
	"github.com/rotisserie/eris"
)

// errorMiddleware handles errors and panics in Gin handlers
type errorMiddleware struct {
	logger ezutil.Logger
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
		if appErr, ok := err.Err.(ungerr.AppError); ok {
			ctx.AbortWithStatusJSON(appErr.HttpStatus(), NewErrorResponse(appErr))
			return
		}

		// Handle other errors
		appError := em.constructAppError(err, ctx)
		ctx.AbortWithStatusJSON(appError.HttpStatus(), NewErrorResponse(appError))
	}
}

// constructAppError converts various error types into AppError
func (em *errorMiddleware) constructAppError(err *gin.Error, ctx *gin.Context) ungerr.AppError {
	// First, try to unwrap with eris to get the original error
	originalErr := eris.Unwrap(err.Err)
	if originalErr == nil {
		// No eris wrapping found - this means the error wasn't properly wrapped
		// Log the location where the error was added to Gin context
		return em.logUnwrappedError(err, ctx)
	}

	// Handle known error types from eris-wrapped errors
	switch originalErr := originalErr.(type) {
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

		// This is an eris-wrapped error but not a known type
		// Log with full stack trace and mask from user
		return em.logAndMaskError(err.Err)
	}
}

// logUnwrappedError handles errors that weren't properly wrapped with eris
func (em *errorMiddleware) logUnwrappedError(err *gin.Error, ctx *gin.Context) ungerr.AppError {
	// This function helps you identify where errors are being added without proper wrapping
	em.logger.Error("UNWRAPPED ERROR DETECTED - Please add eris.Wrap() or return ungerr.AppError")
	em.logger.Errorf("Error type: %T", err.Err)
	em.logger.Errorf("Error message: %s", err.Err.Error())
	em.logger.Errorf("Request: %s %s", ctx.Request.Method, ctx.Request.URL.Path)
	em.logger.Errorf("Handler: %s", ctx.HandlerName())

	// Try to extract some context about where this might have come from
	if ctx.Request != nil {
		em.logger.Errorf("User-Agent: %s", ctx.Request.UserAgent())
		em.logger.Errorf("Remote-Addr: %s", ctx.Request.RemoteAddr)
	}

	// Log the Gin error metadata which might contain handler info
	if err.Meta != nil {
		em.logger.Errorf("Error metadata: %+v", err.Meta)
	}

	em.logger.Error("Stack trace from Gin error location:")
	em.logger.Errorf("%+v", err.Err)

	// Return a masked error to the user
	return ungerr.InternalServerError()
}

// logAndMaskError handles eris-wrapped errors that need to be masked from users
func (em *errorMiddleware) logAndMaskError(err error) ungerr.AppError {
	em.logger.Errorf("Unhandled eris-wrapped error of type: %T", err)
	em.logger.Error("Full stack trace:")
	em.logger.Error(eris.ToString(err, true))

	return ungerr.InternalServerError()
}

// handlePanic recovers from panics and converts them to structured errors
func (em *errorMiddleware) handlePanic(r interface{}, ctx *gin.Context) {
	// Log the panic with full stack trace
	em.logger.Error("PANIC RECOVERED in handler")
	em.logger.Errorf("Request: %s %s", ctx.Request.Method, ctx.Request.URL.Path)
	em.logger.Errorf("Handler: %s", ctx.HandlerName())
	em.logger.Errorf("Panic value: %v", r)
	em.logger.Errorf("Panic type: %T", r)

	if ctx.Request != nil {
		em.logger.Errorf("User-Agent: %s", ctx.Request.UserAgent())
		em.logger.Errorf("Remote-Addr: %s", ctx.Request.RemoteAddr)
		if ctx.Request.Body != nil {
			// Don't log the full body, just indicate if it exists
			em.logger.Error("Request has body: true")
		}
	}

	// Print stack trace
	em.logger.Error("Stack trace:")
	em.logger.Error(string(debug.Stack()))

	// Try to convert panic to a meaningful error
	switch panicValue := r.(type) {
	case string:
		// Handle string panics (often from panic("message"))
		if strings.Contains(panicValue, "index out of range") ||
			strings.Contains(panicValue, "slice bounds out of range") {
			em.logger.Error("Array/slice bounds panic detected")
		} else if strings.Contains(panicValue, "nil pointer dereference") {
			em.logger.Error("Nil pointer dereference panic detected")
		} else {
			em.logger.Errorf("String panic: %s", panicValue)
		}

	case runtime.Error:
		// Handle runtime errors (nil pointer, index out of bounds, etc.)
		em.logger.Errorf("Runtime error panic: %v", panicValue)
		switch panicValue.Error() {
		case "runtime error: invalid memory address or nil pointer dereference":
			em.logger.Error("Nil pointer dereference detected")
		case "runtime error: index out of range":
			em.logger.Error("Index out of range detected")
		case "runtime error: slice bounds out of range":
			em.logger.Error("Slice bounds out of range detected")
		}

	default:
		// Unknown panic type
		em.logger.Errorf("Unknown panic type: %T, value: %v", r, r)
	}

	// If the response hasn't been written yet, send error response
	if !ctx.Writer.Written() {
		appError := ungerr.InternalServerError()
		ctx.AbortWithStatusJSON(appError.HttpStatus(), NewErrorResponse(appError))
	} else {
		// Response was already partially written, we can't send JSON anymore
		em.logger.Errorf("Response already written, cannot send error JSON. Status: %d", ctx.Writer.Status())
	}
}
