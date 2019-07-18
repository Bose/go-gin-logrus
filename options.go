package ginlogrus

import "github.com/sirupsen/logrus"

// Option - define options for WithTracing()
type Option func(*options)
type options struct {
	aggregateLogging bool
	logLevel logrus.Level
}

// defaultOptions - some defs options to NewJWTCache()
var defaultOptions = options{
	aggregateLogging: false,
	logLevel: logrus.DebugLevel,
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
