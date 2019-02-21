package ginlogrus

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestCopyLoggerWithNewBuffer(t *testing.T) {
	l := logrus.WithFields(logrus.Fields{}) // create new buffer for post request logging
	buff := NewBuffer(l)
	buff.Header = map[string]interface{}{}
	buff.Header["testing"] = 123
	buff.Header["deep"] = map[string]string{
		"foo": "bar",
	}
	newLogger, newBuff := CopyLoggerWithNewBuffer(l)
	if newBuff.Header["testing"] != 123 {
		t.Errorf("expected 123 and got %v", newBuff.Header["testing"])
	}
	if newBuff.Header["deep"].(map[string]string)["foo"] != "bar" {
		t.Errorf("expected bar and got %v", newBuff.Header["deep"].(map[string]string)["foo"])
	}
	newLogger.Info("testing")
	fmt.Printf(newBuff.String())

}
