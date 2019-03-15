package ginlogrus

import (
	"testing"
)

func TestWithBanner(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{name: "yes", want: true},
		{name: "no", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultLogBufferOptions()
			f := WithBanner(tt.want)
			f(&opts)
			if opts.addBanner != tt.want {
				t.Errorf("WithBanner() = %v, want %v", opts.addBanner, tt.want)
			}
		})
	}
}

func TestWithHeader(t *testing.T) {
	tests := []struct {
		name      string
		wantKey   string
		wantValue bool
	}{
		{name: "yes", wantKey: "yes", wantValue: true},
		{name: "no", wantKey: "now", wantValue: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := defaultLogBufferOptions()
			f := WithHeader(tt.wantKey, tt.wantValue)
			f(&opts)
			if opts.withHeaders[tt.wantKey].(bool) != tt.wantValue {
				t.Errorf("WithBanner() = %v, want %v", opts.withHeaders[tt.wantKey].(bool), tt.wantValue)
			}
		})
	}
}
