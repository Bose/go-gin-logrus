package main

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

func test() {
	var stringsBuilder strings.Builder
	log := &logrus.Logger{
		Out:       &stringsBuilder,
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}
	logger := logrus.WithFields(logrus.Fields{
		"path":   "/test",
		"status": 200})
	logger.Logger = log
	logger.Info("strings.Builder Info")
	logger.Debug("strings.Builder Debug")
	logger.Error("strings.Builder Error")
	fmt.Printf(stringsBuilder.String())
}
