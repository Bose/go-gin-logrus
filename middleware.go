package ginlogrus

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	return func(c *gin.Context) {
		var aggregateLoggingBuff strings.Builder
		aggregateRequestLogger := &logrus.Logger{
			Out:       &aggregateLoggingBuff,
			Formatter: new(logrus.JSONFormatter),
			Hooks:     make(logrus.LevelHooks),
			Level:     logrus.DebugLevel,
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
			if gin.Mode() != gin.ReleaseMode && !opts.aggregateLogging {
				if useBanner {
					entry.Info("[GIN] --------------------------------------------------------------- GinLogrusWithTracing ----------------------------------------------------------------")
				} else {
					entry.Info()
				}
			}
			if opts.aggregateLogging {
				entry.Logger = aggregateRequestLogger // which uses aggregateLoggingBuff for it's io.Writer
				if useBanner {
					entry.Info("[GIN] --------------------------------------------------------------- GinLogrusWithTracing ----------------------------------------------------------------")
				} else {
					entry.Info()
				}
				fmt.Printf(aggregateLoggingBuff.String())
			}
		}
	}
}

// SetCtxLogger - set the *logrus.Entry for this request in the gin.Context so it can be used throughout the request
func SetCtxLogger(c *gin.Context, logger *logrus.Entry) *logrus.Entry {
	log, found := c.Get("aggregate-logger")
	if found {
		logger.Logger = log.(*logrus.Logger)
	}
	// make sure the original fields for the request are still added
	logger = logger.WithFields(logrus.Fields{
		"requestID": CxtRequestID(c),
		"method":    c.Request.Method,
		"path":      c.Request.URL.Path})

	c.Set("ctxLogger", logger)
	return logger
}

// GetCtxLogger - get the *logrus.Entry for this request from the gin.Context
func GetCtxLogger(c *gin.Context) *logrus.Entry {
	l, ok := c.Get("ctxLogger")
	if ok {
		return l.(*logrus.Entry)
	}
	logger := logrus.WithFields(logrus.Fields{
		"requestID": CxtRequestID(c),
		"method":    c.Request.Method,
		"path":      c.Request.URL.Path,
	})
	log, found := c.Get("aggregate-logger")
	if found {
		logger.Logger = log.(*logrus.Logger)
	}
	c.Set("ctxLogger", logger)
	return logger
}

// CxtRequestID - set a logrus Field entry with the tracing ID for the request
func CxtRequestID(c *gin.Context) string {
	requestID := c.Request.Header.Get("uber-trace-id")
	if len(requestID) == 0 {
		requestID = uuid.New().String()
		fmt.Println(requestID)
	}
	c.Set("RequestID", requestID)
	return requestID
}

// GetCxtRequestID - dig the request ID out of the *logrus.Entry in the gin.Context
func GetCxtRequestID(c *gin.Context) string {
	l, ok := c.Get("ctxLogger")
	if ok {
		requestID, ok := l.(*logrus.Entry).Data["requestID"].(string)
		if ok {
			return requestID
		}
		return "unknown"
	}
	return "unknown"
}
