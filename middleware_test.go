package ginlogrus

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/matryer/is"
	"github.com/sirupsen/logrus"
)

func performRequest(method, target string, router *gin.Engine) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	return w
}

func TestNoLogMessageWithEmptyAggregateEntries(t *testing.T) {
	is := is.New(t)
	buff := ""
	getHandler := func(c *gin.Context) {
		SetCtxLoggerHeader(c, "EmptyEntries", "Nothing should be printed")

		logger := GetCtxLogger(c)
		logger.Info("test-entry-1")
		logger.Info("test-entry-2")
		c.JSON(200, "Hello world!")
	}
	gin.SetMode(gin.DebugMode)
	gin.DisableConsoleColor()

	l := bytes.NewBufferString(buff)
	r := gin.Default()
	r.Use(WithTracing(logrus.StandardLogger(),
		false,
		time.RFC3339,
		true,
		"requestID",
		[]byte("uber-trace-id"), // where jaeger might have put the trace id
		[]byte("RequestID"),     // where the trace ID might already be populated in the headers
		WithAggregateLogging(true),
		WithEmptyAggregateEntries(false),
		WithLogLevel(logrus.WarnLevel),
		WithWriter(l)))
	r.GET("/", getHandler)
	w := performRequest("GET", "/", r)
	is.Equal(200, w.Code)
	t.Log("this is the buffer: ", l)
	is.True(len(l.String()) == 0)
}

func TestLogMessageWithEmptyAggregateEntriesAboveLogLevel(t *testing.T) {
	is := is.New(t)
	buff := ""
	getHandler := func(c *gin.Context) {
		SetCtxLoggerHeader(c, "AggregateEntries", "Shouldnt have messages below WARN")

		logger := GetCtxLogger(c)
		logger.Info("test-entry-1")
		logger.Info("test-entry-2")
		logger.Error("error-entry-1")
		c.JSON(200, "Hello world!")
	}
	gin.SetMode(gin.DebugMode)
	gin.DisableConsoleColor()

	l := bytes.NewBufferString(buff)
	r := gin.Default()
	r.Use(WithTracing(logrus.StandardLogger(),
		false,
		time.RFC3339,
		true,
		"requestID",
		[]byte("uber-trace-id"), // where jaeger might have put the trace id
		[]byte("RequestID"),     // where the trace ID might already be populated in the headers
		WithAggregateLogging(true),
		WithEmptyAggregateEntries(false),
		WithLogLevel(logrus.WarnLevel),
		WithWriter(l)))
	r.GET("/", getHandler)
	w := performRequest("GET", "/", r)
	is.Equal(200, w.Code)
	t.Log("this is the buffer: ", l)
	is.True(len(l.String()) > 0)
	is.True(!strings.Contains(l.String(), "test-entry-1"))
	is.True(!strings.Contains(l.String(), "test-entry-2"))
	is.True(strings.Contains(l.String(), "error-entry-1"))
}

func TestBanner(t *testing.T) {
	is := is.New(t)
	buff := ""
	getHandler := func(c *gin.Context) {
		SetCtxLoggerHeader(c, "new-header-index-name", "this is how you set new header level data")

		logger := GetCtxLogger(c)
		logger.Info("test-entry-1")
		logger.Info("test-entry-2")
		c.JSON(200, "Hello world!")
	}
	gin.SetMode(gin.DebugMode)
	gin.DisableConsoleColor()

	l := bytes.NewBufferString(buff)
	r := gin.Default()
	r.Use(WithTracing(logrus.StandardLogger(),
		false,
		time.RFC3339,
		true,
		"requestID",
		[]byte("uber-trace-id"), // where jaeger might have put the trace id
		[]byte("RequestID"),     // where the trace ID might already be populated in the headers
		WithAggregateLogging(true),
		WithWriter(l)))
	r.GET("/", getHandler)
	w := performRequest("GET", "/", r)
	is.Equal(200, w.Code)
	t.Log("this is the buffer: ", l)
	is.True(!strings.Contains(l.String(), "GinLogrusWithTracing"))

	buff = ""
	l = bytes.NewBufferString(buff)
	r = gin.New()
	r.Use(WithTracing(logrus.StandardLogger(),
		true,
		time.RFC3339,
		true,
		"requestID",
		[]byte("uber-trace-id"), // where jaeger might have put the trace id
		[]byte("RequestID"),     // where the trace ID might already be populated in the headers
		WithAggregateLogging(true),
		WithWriter(l)))
	r.GET("/", getHandler)
	w = performRequest("GET", "/", r)
	is.Equal(200, w.Code)
	t.Log("this is the buffer: ", l)
	is.True(strings.Contains(l.String(), "GinLogrusWithTracing"))

	customBanner := "---- custom banner ----"
	buff = ""
	l = bytes.NewBufferString(buff)
	r = gin.New()
	r.Use(WithTracing(logrus.StandardLogger(),
		true,
		time.RFC3339,
		true,
		"requestID",
		[]byte("uber-trace-id"), // where jaeger might have put the trace id
		[]byte("RequestID"),     // where the trace ID might already be populated in the headers
		WithLogCustomBanner(customBanner),
		WithAggregateLogging(true),
		WithWriter(l)))
	r.GET("/", getHandler)
	w = performRequest("GET", "/", r)
	is.Equal(200, w.Code)
	t.Log("this is the buffer: ", l)
	is.True(strings.Contains(l.String(), customBanner))

}

func TestLogMessageWithProductionLevelReducedLogging(t *testing.T) {
	is := is.New(t)
	buff := ""
	getHandler := func(c *gin.Context) {
		SetCtxLoggerHeader(c, "ReducedLogging", "Shouldn't have messages with a 2xx response")

		logger := GetCtxLogger(c)
		logger.Info("test-entry-1")
		logger.Info("test-entry-2")
		logger.Error("error-entry-1")
		c.JSON(200, "Hello world!")
	}
	failHandler := func(c *gin.Context) {
		SetCtxLoggerHeader(c, "ReducedLogging", "Shouldn't have messages with a 2xx response")
		logger := GetCtxLogger(c)
		logger.Info("test-entry-1")
		logger.Info("test-entry-2")
		logger.Error("error-entry-1")
		c.JSON(401, "Hello fail!")
	}
	gin.SetMode(gin.DebugMode)
	gin.DisableConsoleColor()

	l := bytes.NewBufferString(buff)
	r := gin.Default()
	r.Use(WithTracing(logrus.StandardLogger(),
		false,
		time.RFC3339,
		true,
		"requestID",
		[]byte("uber-trace-id"), // where jaeger might have put the trace id
		[]byte("RequestID"),     // where the trace ID might already be populated in the headers
		WithAggregateLogging(true),
		WithWriter(l),
		WithReducedLoggingFunc(productionLoggingTestFunc),
	))
	r.GET("/", getHandler)
	r.GET("/fail", failHandler)
	w := performRequest("GET", "/", r)
	is.Equal(200, w.Code)
	t.Log("this is the buffer: ", l)
	// Beacuase the request is a 2xx we will not have any log entries including possible errors
	is.True(len(l.String()) == 0)
	is.True(!strings.Contains(l.String(), "test-entry-1"))
	is.True(!strings.Contains(l.String(), "test-entry-2"))
	is.True(!strings.Contains(l.String(), "error-entry-1"))

	w = performRequest("GET", "/fail", r)
	is.Equal(401, w.Code)
	t.Log("this is the buffer: ", l)
	// Beacuase the request is a 401 we will have all log entries including info logs
	is.True(len(l.String()) > 0)
	is.True(strings.Contains(l.String(), "test-entry-1"))
	is.True(strings.Contains(l.String(), "test-entry-2"))
	is.True(strings.Contains(l.String(), "error-entry-1"))

}

func TestLogMessageWithProductionReducedLoggingWarnLevel(t *testing.T) {
	is := is.New(t)
	buff := ""
	getHandler := func(c *gin.Context) {
		SetCtxLoggerHeader(c, "ReducedLogging", "Shouldn't have messages with a 2xx response")

		logger := GetCtxLogger(c)
		logger.Info("test-entry-1")
		logger.Info("test-entry-2")
		logger.Error("error-entry-1")
		c.JSON(200, "Hello world!")
	}
	failHandler := func(c *gin.Context) {
		SetCtxLoggerHeader(c, "ReducedLogging", "Shouldn't have messages with a 2xx response")
		logger := GetCtxLogger(c)
		logger.Info("test-entry-1")
		logger.Info("test-entry-2")
		logger.Error("error-entry-1")
		c.JSON(401, "Hello fail!")
	}
	gin.SetMode(gin.DebugMode)
	gin.DisableConsoleColor()

	l := bytes.NewBufferString(buff)
	r := gin.Default()
	r.Use(WithTracing(logrus.StandardLogger(),
		false,
		time.RFC3339,
		true,
		"requestID",
		[]byte("uber-trace-id"), // where jaeger might have put the trace id
		[]byte("RequestID"),     // where the trace ID might already be populated in the headers
		WithAggregateLogging(true),
		WithWriter(l),
		WithLogLevel(logrus.WarnLevel),
		WithReducedLoggingFunc(productionLoggingTestFunc),
	))
	r.GET("/", getHandler)
	r.GET("/fail", failHandler)
	w := performRequest("GET", "/", r)
	is.Equal(200, w.Code)
	t.Log("this is the buffer: ", l)
	// Beacuase the request is a 2xx we will not have any log entries including possible errors
	is.True(len(l.String()) == 0)
	is.True(!strings.Contains(l.String(), "test-entry-1"))
	is.True(!strings.Contains(l.String(), "test-entry-2"))
	is.True(!strings.Contains(l.String(), "error-entry-1"))

	w = performRequest("GET", "/fail", r)
	is.Equal(401, w.Code)
	t.Log("this is the buffer: ", l)
	// Beacuase the request is a 401 we will have log entries but because we have our log level at WARN we will not have info logs
	is.True(len(l.String()) > 0)
	is.True(!strings.Contains(l.String(), "test-entry-1"))
	is.True(!strings.Contains(l.String(), "test-entry-2"))
	is.True(strings.Contains(l.String(), "error-entry-1"))

}

// Same production logging function that will only log on statusCodes in a certain range
func productionLoggingTestFunc(c *gin.Context) bool {
	statusCode := c.Writer.Status()
	if statusCode < http.StatusOK || statusCode >= http.StatusMultipleChoices {
		return true
	}
	return false
}
