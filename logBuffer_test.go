package ginlogrus

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestLogBuffer_String(t *testing.T) {

	tests := []struct {
		name     string
		buff     LogBuffer
		write    []byte
		contains string
	}{
		{
			name:     "hey",
			buff:     NewLogBuffer(WithBanner(true), WithHeader("id1", "val1"), WithHeader("id2", "id2")),
			write:    []byte("\"msg\":\"hey-one\""),
			contains: "hey",
		},
		{
			name:     "hey-now",
			buff:     NewLogBuffer(WithHeader("hey", "now")),
			write:    []byte("\"msg\":\"hey-two\""),
			contains: "hey",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.buff.Write(tt.write)
			fmt.Println("buff == ", tt.buff.String())
			if !strings.Contains(tt.buff.String(), tt.contains) {
				t.Errorf("LogBuffer.String() = %v, want %v", tt.buff.String(), tt.contains)
			}
		})
	}
}

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
			opts := defaultLogBufferOptions
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
			opts := defaultLogBufferOptions
			f := WithHeader(tt.wantKey, tt.wantValue)
			f(&opts)
			if opts.withHeaders[tt.wantKey].(bool) != tt.wantValue {
				t.Errorf("WithBanner() = %v, want %v", opts.withHeaders[tt.wantKey].(bool), tt.wantValue)
			}
		})
	}
}

func TestNewLogBuffer(t *testing.T) {
	tests := []struct {
		name string
		opt  []LogBufferOption
		want LogBuffer
	}{
		{
			name: "one",
			opt:  []LogBufferOption{WithBanner(true), WithHeader("1", true)},
			want: LogBuffer{AddBanner: true, header: map[string]interface{}{"1": true}, headerMU: &sync.RWMutex{}},
		},
		{
			name: "two",
			opt:  []LogBufferOption{WithHeader("1", "one"), WithHeader("2", true)},
			want: LogBuffer{AddBanner: false, header: map[string]interface{}{"1": "one", "2": true}, headerMU: &sync.RWMutex{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLogBuffer(tt.opt...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLogBuffer() = %v, want %v", got, tt.want)
			}
		})
	}
}
