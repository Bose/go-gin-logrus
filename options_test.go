package ginlogrus

import (
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
