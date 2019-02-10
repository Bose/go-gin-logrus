package ginlogrus

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func TestNewBuffer(t *testing.T) {

	c := getTestContext("boo", "bar", true)
	logger := GetCtxLogger(c)
	logger.Info("now")

	buff := NewBuffer(logger)
	logger.Info("hey")
	if strings.Contains(buff.String(), "hey") == false || strings.Contains(buff.String(), "entries") == false {
		t.Errorf("Expected hey and found %v", buff.String())
	}
	if strings.Contains(buff.String(), "now") {
		t.Errorf("didn't expeect 'now' and got %v", buff.String())
	}

	c = getTestContext("boo", "bar", false)
	logger.Info("now")

}

func getTestContext(hdr string, v string, withAggregateLogger bool) *gin.Context {
	buf := new(bytes.Buffer)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/", buf)
	c.Request.Header.Set(hdr, v)

	if withAggregateLogger {
		aggregateLoggingBuff := LogBuffer{}
		aggregateRequestLogger := &logrus.Logger{
			Out:       &aggregateLoggingBuff,
			Formatter: new(logrus.JSONFormatter),
			Hooks:     make(logrus.LevelHooks),
			Level:     logrus.DebugLevel,
		}
		// you have to use this logger for every *logrus.Entry you create
		c.Set("aggregate-logger", aggregateRequestLogger)

	}
	return c
}
