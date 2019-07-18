package ginlogrus

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func TestWithAggregateLogging(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{name: "true", want: true},
		{name: "false", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultOptions
			f := WithAggregateLogging(tt.want)
			f(&opts)
			if opts.aggregateLogging != tt.want {
				t.Errorf("WithAggregateLogging() = %v, want %v", opts.aggregateLogging, tt.want)
			}
		})
	}
}

func TestWithLogLevel(t *testing.T) {
	tests := []struct {
		name string
		want logrus.Level
	}{
		{name: "info", want: logrus.InfoLevel},
		{name: "debug", want: logrus.DebugLevel},
		{name: "error", want: logrus.ErrorLevel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultOptions
			f := WithLogLevel(tt.want)
			f(&opts)
			if opts.logLevel != tt.want {
				t.Errorf("WithLogLevel() = %v, want %v", opts.aggregateLogging, tt.want)
			}
		})
	}
}

func TestWithLogLevelAggregateLogging(t *testing.T) {
	tests := []struct {
		name string
		aggregate bool
		level logrus.Level
	}{
		{name: "info and true", aggregate: true, level: logrus.InfoLevel},
		{name: "debug and true",  aggregate: true, level: logrus.DebugLevel},
		{name: "error and false", level: logrus.ErrorLevel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultOptions
			f := WithLogLevelAggregateLogging(tt.aggregate, tt.level)
			f(&opts)
			if opts.logLevel != tt.level || opts.aggregateLogging != tt.aggregate {
				t.Errorf("WithLogLevelAggregateLogging() = %v, want %v", opts, tt)
			}
		})
	}
}