package ginlogrus

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

func SetCtxLoggerHeader(c *gin.Context, name string, data interface{}) {
	logger := GetCtxLogger(c)
	logger.Logger.Out.(*LogBuffer).Header[name] = data
}

// SetCtxLogger - set the *logrus.Entry for this request in the gin.Context so it can be used throughout the request
func SetCtxLogger(c *gin.Context, logger *logrus.Entry) *logrus.Entry {
	log, found := c.Get("aggregate-logger")
	if found {
		logger.Logger = log.(*logrus.Logger)
	}
	logger = logger.WithFields(logrus.Fields{})
	c.Set("ctxLogger", logger)
	return logger
}

// GetCtxLogger - get the *logrus.Entry for this request from the gin.Context
func GetCtxLogger(c *gin.Context) *logrus.Entry {
	l, ok := c.Get("ctxLogger")
	if ok {
		return l.(*logrus.Entry)
	}
	logger := logrus.WithFields(logrus.Fields{})
	log, found := c.Get("aggregate-logger")
	if found {
		logger.Logger = log.(*logrus.Logger)
	}
	c.Set("ctxLogger", logger)
	return logger
}

// CxtRequestID - set a logrus Field entry with the tracing ID for the request
func CxtRequestID(c *gin.Context) string {
	// already setup, so we're done
	if id, found := c.Get("RequestID"); found == true {
		return id.(string)
	}

	// see if we're using github.com/Bose/go-gin-opentracing which will set a span in "tracing-context"
	if s, foundSpan := c.Get("tracing-context"); foundSpan {
		span := s.(opentracing.Span)
		requestID := fmt.Sprintf("%v", span)
		c.Set("RequestID", requestID)
		return requestID
	}

	// some other process might have stuck it in a header
	if len(ContextTraceIDField) != 0 {
		if s, ok := c.Get(ContextTraceIDField); ok {
			span := s.(opentracing.Span)
			requestID := fmt.Sprintf("%v", span)
			c.Set("RequestID", requestID)
			return requestID
		}
	}

	if requestID := c.Request.Header.Get("uber-trace-id"); len(requestID) != 0 {
		c.Set("RequestID", requestID)
		return requestID
	}

	// finally, just create a fake request id...
	requestID := uuid.New().String()
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
