package ginlogrus

import (
	"bytes"
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
}
