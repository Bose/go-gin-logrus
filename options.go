package ginlogrus

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Option - define options for WithTracing()
type Option func(*options)

type ReducedLoggingFunc func(c *gin.Context) bool

type options struct {
	aggregateLogging      bool
	logLevel              logrus.Level
	emptyAggregateEntries bool
	reducedLoggingFunc    ReducedLoggingFunc
	writer                io.Writer
	banner                string
}

// defaultOptions - some defs options to NewJWTCache()
var defaultOptions = options{
	aggregateLogging:      false,
	logLevel:              logrus.DebugLevel,
	emptyAggregateEntries: true,
	reducedLoggingFunc:    func(c *gin.Context) bool { return true },
	writer:                os.Stdout,
	banner:                DefaultBanner,
}

// WithAggregateLogging - define an Option func for passing in an optional aggregateLogging
func WithAggregateLogging(a bool) Option {
	return func(o *options) {
		o.aggregateLogging = a
	}
}

// WithEmptyAggregateEntries - define an Option func for printing aggregate logs with empty entries
func WithEmptyAggregateEntries(a bool) Option {
	return func(o *options) {
		o.emptyAggregateEntries = a
	}
}

// WithEmptyAggregateEntries - define an Option func for printing aggregate logs with empty entries
func WithReducedLoggingFunc(a ReducedLoggingFunc) Option {
	return func(o *options) {
		o.reducedLoggingFunc = a
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

// WithLogCustomBanner allows users to define their own custom banner.  There is some overlap with this name and the LogBufferOption.CustomBanner and yes,
// they are related options, but I didn't want to make a breaking API change to support this new option... so we'll have to live with a bit of confusion/overlap in option names
func WithLogCustomBanner(b string) Option {
	return func(o *options) {
		o.banner = b
	}
}
