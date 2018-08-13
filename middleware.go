package ginlogrus

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

type loggerEntryWithFields interface {
	WithFields(fields logrus.Fields) *logrus.Entry
}

// WithTracing returns a gin.HandlerFunc (middleware) that logs requests using logrus.
//
// Requests with errors are logged using logrus.Error().
// Requests without errors are logged using logrus.Info().
//
// It receives:
//   1. A logrus.Entry with fields
//   2. A boolean stating whether to use a BANNER in the log entry
//   3. A time package format string (e.g. time.RFC3339).
//   4. A boolean stating whether to use UTC time zone or local.
//   5. A string to use for Trace ID the Logrus log field.
//   t. A []byte for the request header that contains the trace id
//   5. A []byte for "getting" the requestID out of the gin.Context
func WithTracing(
	logger loggerEntryWithFields,
	useBanner bool,
	timeFormat string,
	utc bool,
	logrusFieldNameForTraceID string,
	traceIDHeader []byte,
	contextTraceIDField []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		if utc {
			end = end.UTC()
		}
		late := fmt.Sprintf("%13v", latency)

		var requestID string
		// see if we're using github.com/Bose/go-gin-opentracing which will set a span in "tracing-context"
		if s, foundSpan := c.Get("tracing-context"); foundSpan {
			span := s.(opentracing.Span)
			requestID = fmt.Sprintf("%v", span)
		}
		// check a user defined context field
		if len(requestID) == 0 && contextTraceIDField != nil {
			if id, ok := c.Get(string(contextTraceIDField)); ok {
				requestID = id.(string)
			}
		}
		// okay.. finally check the request header
		if len(requestID) == 0 && traceIDHeader != nil {
			requestID = c.Request.Header.Get(string(traceIDHeader))
		}

		comment := c.Errors.ByType(gin.ErrorTypePrivate).String()

		entry := logger.WithFields(logrus.Fields{
			logrusFieldNameForTraceID: requestID,
			"status":                  c.Writer.Status(),
			"method":                  c.Request.Method,
			"path":                    path,
			"ip":                      c.ClientIP(),
			"latency":                 late,
			"user-agent":              c.Request.UserAgent(),
			"time":                    end.Format(timeFormat),
			"comment":                 comment,
		})

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			entry.Error(c.Errors.String())
		} else {
			if gin.Mode() != gin.ReleaseMode {
				if useBanner {
					entry.Info("[GIN] --------------------------------------------------------------- GinLogrusWithTracing ----------------------------------------------------------------")
				} else {
					entry.Info()
				}
			}
		}
	}
}
