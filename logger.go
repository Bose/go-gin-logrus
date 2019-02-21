package ginlogrus

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mitchellh/copystructure"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

// SetCtxLoggerHeader - if aggregate logging, add header info... otherwise just info log the data passed
func SetCtxLoggerHeader(c *gin.Context, name string, data interface{}) {
	logger := GetCtxLogger(c)
	_, found := c.Get("aggregate-logger")
	if found {
		logger.Logger.Out.(*LogBuffer).Header[name] = data
	}
	if !found {
		logger.Infof("%s: %v", name, data)
	}
}

// SetCtxLogger - used when you want to set the *logrus.Entry with new logrus.WithFields{} for this request in the gin.Context so it can be used going forward for the request
func SetCtxLogger(c *gin.Context, logger *logrus.Entry) *logrus.Entry {
	log, found := c.Get("aggregate-logger")
	if found {
		logger.Logger = log.(*logrus.Logger)
		logger = logger.WithFields(logrus.Fields{}) // no need to add additional fields when aggregate logging
	}
	if !found {
		// not aggregate logging, so make sure  to add some needed fields
		logger = logger.WithFields(logrus.Fields{
			"requestID": CxtRequestID(c),
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path})
	}
	c.Set("ctxLogger", logger)
	return logger
}

// GetCtxLogger - get the *logrus.Entry for this request from the gin.Context
func GetCtxLogger(c *gin.Context) *logrus.Entry {
	l, ok := c.Get("ctxLogger")
	if ok {
		return l.(*logrus.Entry)
	}
	var logger *logrus.Entry
	log, found := c.Get("aggregate-logger")
	if found {
		logger = logrus.WithFields(logrus.Fields{})
		logger.Logger = log.(*logrus.Logger)
	}
	if !found {
		// not aggregate logging, so make sure  to add some needed fields
		logger = logrus.WithFields(logrus.Fields{
			"requestID": CxtRequestID(c),
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path,
		})
	}
	c.Set("ctxLogger", logger)
	return logger
}

// CxtRequestID - if not already set, then add logrus Field to the entry with the tracing ID for the request.
// then return the trace/request id
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

// NewBuffer - create a new aggregate logging buffer for the *logrus.Entry , which can be flushed by the consumer
// how-to, when to use this:
// 		the request level log entry is written when the request is over, so you need this thing to
// 		write go routine logs that complete AFTER the request is completed.
//      careful: the loggers will share a ref to the same Header (writes to one will affect the other)
// example:
// go func() {
// 		buff := NewBuffer(logger) // logger is an existing *logrus.Entry
// 		// do somem work here and write some logs via the logger.  Like logger.Info("hi mom! I'm a go routine that finished after the request")
// 		fmt.Printf(buff.String()) // this will write the aggregated buffered logs to stdout
// }()
//
func NewBuffer(l *logrus.Entry) *LogBuffer {
	buff := LogBuffer{}
	if l, ok := l.Logger.Out.(*LogBuffer); ok {
		buff.Header = l.Header
	}
	// buff.Header = l.Logger.Out.(*ginlogrus.LogBuffer).Header
	l.Logger = &logrus.Logger{
		Out:       &buff,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}
	return &buff
}

// CopyLoggerWithNewBuffer - copies info out of an existing logger and creates a new logger and aggregate logging buffer
// this is NOT a concurrent safe operation.  The logger's header (map) is iterated over, so you cannot be writing to the
// logger's header at the same time
//
// you would use this when you want to keep the Headers separate.
func CopyLoggerWithNewBuffer(logger *logrus.Entry) (*logrus.Entry, *LogBuffer) {
	newLogger := logrus.WithFields(logrus.Fields{}) // create new buffer for post request logging
	buff := NewBuffer(newLogger)
	buff.AddBanner = true
	if l, ok := logger.Logger.Out.(*LogBuffer); ok {
		dup, err := copystructure.Copy(l.Header)
		if err != nil {
			buff.Header = map[string]interface{}{}
		} else {
			buff.Header = dup.(map[string]interface{})
		}
	} else {
		buff.Header = map[string]interface{}{}
	}
	return newLogger, buff
}
