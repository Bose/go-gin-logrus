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