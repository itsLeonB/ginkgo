package middleware

import (
	"encoding/json"
	"fmt"
	"io"
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
			em.logger.WithContext(ctx).WithError(appError).Warn("application error")
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
	logCtx := em.logger.WithContext(ctx)

	unknownErr, ok := err.Err.(*ungerr.UnknownError)
	if ok {
		logCtx = logCtx.WithError(unknownErr)
		if originalErr := ungerr.Unwrap(err.Err); originalErr != nil {
			logCtx.Error("unhandled error")
			return ungerr.InternalServerError()
		}

		logCtx.Error("unexpected error")
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

		logCtx.
			WithError(originalErr).
			WithField("handler", ctx.HandlerName()).
			Error("unwrapped error detected — wrap with ungerr.Wrap()")

		return ungerr.InternalServerError()
	}
}

// handlePanic recovers from panics and converts them to structured errors
func (em *errorMiddleware) handlePanic(r any, ctx *gin.Context) {
	em.logger.
		WithContext(ctx.Request.Context()).
		WithFields(map[string]any{
			"handler":     ctx.HandlerName(),
			"panic.type":  fmt.Sprintf("%T", r),
			"panic.value": fmt.Sprintf("%v", r),
			"stack_trace": string(debug.Stack()),
		}).
		Error("panic recovered")

	if !ctx.Writer.Written() {
		appError := ungerr.InternalServerError()
		ctx.AbortWithStatusJSON(appError.HttpStatus(), appErrorToErrorObject(appError))
	} else {
		em.logger.
			WithContext(ctx.Request.Context()).
			WithField("http.status_code", ctx.Writer.Status()).
			Error("response already written after panic, could not send error JSON")
	}
}
