package ginlogrus

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// Option - define options for WithTracing()
type Option func(*options)
type options struct {
	aggregateLogging bool
	logLevel         logrus.Level
	writer           io.Writer
}

// defaultOptions - some defs options to NewJWTCache()
var defaultOptions = options{
	aggregateLogging: false,
	logLevel:         logrus.DebugLevel,
	writer:           os.Stdout,
}

// WithAggregateLogging - define an Option func for passing in an optional aggregateLogging
func WithAggregateLogging(a bool) Option {
	return func(o *options) {
		o.aggregateLogging = a
	}
}

// WithLogLevel - define an Option func for passing in an optional logLevel
func WithLogLevel(logLevel logrus.Level) Option {
	return func(o *options) {
		o.logLevel = logLevel
	}
}

// WithWriter allows users to define the writer used for middlware aggregagte logging, the default writer is os.Stdout
func WithWriter(w io.Writer) Option {
	return func(o *options) {
		o.writer = w
	}
}
