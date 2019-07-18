package ginlogrus

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

// ContextTraceIDField - used to find the trace id in the gin.Context - optional
var ContextTraceIDField string

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
//   6. A []byte for the request header that contains the trace id
//   7. A []byte for "getting" the requestID out of the gin.Context
//   8. A list of possible ginlogrus.Options to apply
func WithTracing(
	logger loggerEntryWithFields,
	useBanner bool,
	timeFormat string,
	utc bool,
	logrusFieldNameForTraceID string,
	traceIDHeader []byte,
	contextTraceIDField []byte,
	opt ...Option) gin.HandlerFunc {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}
	if contextTraceIDField != nil {
		ContextTraceIDField = string(contextTraceIDField)
	}
	return func(c *gin.Context) {
		// var aggregateLoggingBuff strings.Builder
		// var aggregateLoggingBuff logBuffer
		aggregateLoggingBuff := NewLogBuffer(WithBanner(true))
		aggregateRequestLogger := &logrus.Logger{
			Out:       &aggregateLoggingBuff,
			Formatter: new(logrus.JSONFormatter),
			Hooks:     make(logrus.LevelHooks),
			Level:     opts.logLevel,
		}

		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path

		if opts.aggregateLogging {
			// you have to use this logger for every *logrus.Entry you create
			c.Set("aggregate-logger", aggregateRequestLogger)
		}
		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		if utc {
			end = end.UTC()
		}

		var requestID string
		// see if we're using github.com/Bose/go-gin-opentracing which will set a span in "tracing-context"
		if s, foundSpan := c.Get("tracing-context"); foundSpan {
			span := s.(opentracing.Span)
			requestID = fmt.Sprintf("%v", span)
		}
		// check a user defined context field
		if len(requestID) == 0 && contextTraceIDField != nil {
			if id, ok := c.Get(string(ContextTraceIDField)); ok {
				requestID = id.(string)
			}
		}
		// okay.. finally check the request header
		if len(requestID) == 0 && traceIDHeader != nil {
			requestID = c.Request.Header.Get(string(traceIDHeader))
		}

		comment := c.Errors.ByType(gin.ErrorTypePrivate).String()

		fields := logrus.Fields{
			logrusFieldNameForTraceID: requestID,
			"status":                  c.Writer.Status(),
			"method":                  c.Request.Method,
			"path":                    path,
			"ip":                      c.ClientIP(),
			"latency-ms":              float64(latency) / float64(time.Millisecond),
			"user-agent":              c.Request.UserAgent(),
			"time":                    end.Format(timeFormat),
			"comment":                 comment,
		}
		if len(c.Errors) > 0 {
			entry := logger.WithFields(fields)
			// Append error field if this is an erroneous request.
			entry.Error(c.Errors.String())
		} else {
			if gin.Mode() != gin.ReleaseMode && !opts.aggregateLogging {
				entry := logger.WithFields(fields)
				if useBanner {
					entry.Info("[GIN] --------------------------------------------------------------- GinLogrusWithTracing ----------------------------------------------------------------")
				} else {
					entry.Info()
				}
			}
			if opts.aggregateLogging {
				aggregateLoggingBuff.StoreHeader("request-summary-info", fields)
				// if useBanner {
				// 	fields["banner"] = "[GIN] --------------------------------------------------------------- GinLogrusWithTracing ----------------------------------------------------------------"
				// }
				fmt.Printf(aggregateLoggingBuff.String())
			}
		}
	}
}
