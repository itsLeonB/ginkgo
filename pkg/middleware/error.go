package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ginkgo/pkg/response"
	"github.com/itsLeonB/ungerr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type errorMiddleware struct {
	logger ezutil.Logger
	tracer trace.Tracer
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
	m := &errorMiddleware{logger: logger, tracer: otel.GetTracerProvider().Tracer(packageName)}
	return m.handle
}

func appErrorToErrorObject(appError ungerr.AppError) any {
	return response.NewErrorResponse(errorObject{
		Code:   appError.Error(),
		Detail: appError.Details(),
	})
}

func (em *errorMiddleware) handle(ctx *gin.Context) {
	c, span := em.tracer.Start(ctx.Request.Context(), "ErrorMiddleware.handle")
	defer span.End()
	ctx.Request = ctx.Request.WithContext(c)

	defer func() {
		if r := recover(); r != nil {
			em.handlePanic(r, ctx, span)
		}
	}()

	ctx.Next()

	ginErr := ctx.Errors.Last()
	if ginErr == nil {
		return
	}

	err := ginErr.Err
	logCtx := em.logger.WithContext(ctx)

	// Already a well-typed AppError — warn and respond.
	if appError, ok := err.(ungerr.AppError); ok {
		span.RecordError(appError)
		span.SetStatus(codes.Error, "application error")
		logCtx.WithError(appError).Warn("application error")
		ctx.AbortWithStatusJSON(appError.HttpStatus(), appErrorToErrorObject(appError))
		return
	}

	// UnknownError has two distinct log messages depending on whether a cause is present.
	if unknownErr, ok := err.(*ungerr.UnknownError); ok {
		logCtx = logCtx.WithError(unknownErr)
		if cause := ungerr.Unwrap(err); cause != nil {
			span.RecordError(cause)
			span.SetStatus(codes.Error, "wrapped error")
			if appError := em.identifyKnownError(cause); appError != nil {
				span.SetStatus(codes.Error, "identified error")
				logCtx.WithError(appError).Warn("identified wrapped error")
				ctx.AbortWithStatusJSON(appError.HttpStatus(), appErrorToErrorObject(appError))
				return
			}
			logCtx.Error("unhandled error") // only if truly unidentifiable
		} else {
			span.RecordError(err)
			span.SetStatus(codes.Error, "unexpected error")
			logCtx.Error("unexpected error")
		}
		appError := ungerr.InternalServerError()
		ctx.AbortWithStatusJSON(appError.HttpStatus(), appErrorToErrorObject(appError))
		return
	}

	// Try to map remaining known error types (validation, JSON, network, etc.).
	appError := em.identifyKnownError(err)
	if appError != nil {
		logCtx.WithError(appError).Warn("application error")
	} else {
		// Completely unrecognised error — developer forgot to wrap with ungerr.Wrap().
		logCtx.
			WithError(err).
			WithField("handler", ctx.HandlerName()).
			Error("unwrapped error detected — wrap with ungerr.Wrap()")
		appError = ungerr.InternalServerError()
	}

	span.RecordError(appError)
	span.SetStatus(codes.Error, "application error")
	ctx.AbortWithStatusJSON(appError.HttpStatus(), appErrorToErrorObject(appError))
}

func (em *errorMiddleware) identifyKnownError(err error) ungerr.AppError {
	switch e := err.(type) {
	case validator.ValidationErrors:
		msgs := make([]string, len(e))
		for i, ve := range e {
			msgs[i] = ve.Error()
		}
		return ungerr.ValidationError(msgs)

	case *json.SyntaxError:
		return ungerr.BadRequestError("invalid json")

	case *json.UnmarshalTypeError:
		return ungerr.BadRequestError(fmt.Sprintf("invalid value for field %s", e.Field))

	case *net.OpError:
		if e.Timeout() {
			return ungerr.TimeoutError("your connection may be slow, please retry")
		}
		return identifyOtherErrors(e)

	default:
		return identifyOtherErrors(e)
	}
}

func identifyOtherErrors(e error) ungerr.AppError {
	errStr := e.Error()
	if e == io.EOF || errStr == "EOF" {
		return ungerr.BadRequestError("missing request body")
	}
	if strings.Contains(errStr, "connection reset by peer") ||
		strings.Contains(errStr, "broken pipe") {
		return ungerr.BadRequestError("connection error")
	}
	return nil
}

func (em *errorMiddleware) handlePanic(r any, ctx *gin.Context, span trace.Span) {
	em.logger.
		WithContext(ctx.Request.Context()).
		WithFields(map[string]any{
			"handler":     ctx.HandlerName(),
			"panic.type":  fmt.Sprintf("%T", r),
			"panic.value": fmt.Sprintf("%v", r),
			"stack_trace": string(debug.Stack()),
		}).
		Error("panic recovered")

	appError := ungerr.InternalServerError()
	span.RecordError(appError)
	span.SetStatus(codes.Error, "panic recovered")

	if ctx.Writer.Written() {
		em.logger.
			WithContext(ctx.Request.Context()).
			WithField("http.status_code", ctx.Writer.Status()).
			Error("response already written after panic, could not send error JSON")
		return
	}
	ctx.AbortWithStatusJSON(appError.HttpStatus(), appErrorToErrorObject(appError))
}
