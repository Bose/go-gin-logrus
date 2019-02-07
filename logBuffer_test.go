package ginlogrus

import (
	"fmt"
	"strings"
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
			name:     "one",
			buff:     LogBuffer{},
			write:    []byte("hey"),
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
